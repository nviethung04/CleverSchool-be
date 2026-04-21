package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"

	"gorm.io/gorm"
)

type StudyReportCriteriaRepository interface {
	base.BaseRepositoryInterface[models.StudyReportCriteria]
	ReplaceSkills(criteriaID int64, skills []StudyReportSkillWithTypes) error
}

type studyReportCriteriaRepository struct {
	*base.BaseRepository[models.StudyReportCriteria]
}

func NewStudyReportCriteriaRepository() StudyReportCriteriaRepository {
	return &studyReportCriteriaRepository{
		BaseRepository: base.NewBaseRepository[models.StudyReportCriteria](),
	}
}

type StudyReportSkillWithTypes struct {
	Skill models.StudyReportSkill
	Types []models.StudyReportSkillType
}

func (r *studyReportCriteriaRepository) ReplaceSkills(criteriaID int64, skills []StudyReportSkillWithTypes) error {
	tx := db.MasterDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	keepSkillIDs := make([]int64, 0, len(skills))
	for _, record := range skills {
		if record.Skill.CriteriaId == 0 {
			record.Skill.CriteriaId = criteriaID
		}

		skillID, err := r.upsertSkill(tx, record.Skill)
		if err != nil {
			tx.Rollback()
			return err
		}
		keepSkillIDs = append(keepSkillIDs, skillID)

		if err := r.syncSkillTypes(tx, skillID, record.Types); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := r.deleteRemovedSkills(tx, criteriaID, keepSkillIDs); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *studyReportCriteriaRepository) upsertSkill(tx *gorm.DB, skill models.StudyReportSkill) (int64, error) {
	if skill.ID > 0 {
		data := map[string]interface{}{
			"name_vn":     skill.NameVn,
			"name_en":     skill.NameEn,
			"description": skill.Description,
			"type":        skill.Type,
			"sort_order":  skill.SortOrder,
		}

		if err := tx.Model(&models.StudyReportSkill{}).
			Where("id = ? AND criteria_id = ?", skill.ID, skill.CriteriaId).
			Updates(data).Error; err != nil {
			return 0, err
		}

		return skill.ID, nil
	}

	if err := tx.Create(&skill).Error; err != nil {
		return 0, err
	}

	return skill.ID, nil
}

func (r *studyReportCriteriaRepository) syncSkillTypes(tx *gorm.DB, skillID int64, types []models.StudyReportSkillType) error {
	keepTypeIDs := make([]int64, 0, len(types))
	idMap := make(map[int64]int64)

	for _, t := range types {
		origID := t.ID
		// Khi copy: ID âm = original id để map parent_id (origID = -t.ID), sau khi tạo mới thì idMap[origID] = newID
		if t.ID < 0 {
			origID = -t.ID
			t.ID = 0
		}

		if t.ParentId != nil {
			if mapped, ok := idMap[*t.ParentId]; ok && mapped > 0 {
				t.ParentId = &mapped
			} else {
				t.ParentId = nil
			}
		}

		if t.ID <= 0 {
			t.ID = 0
		}

		t.SkillId = skillID

		if t.ID > 0 {
			data := map[string]interface{}{
				"name_vn":    t.NameVn,
				"name_en":    t.NameEn,
				"parent_id":  t.ParentId,
				"level":      t.Level,
				"sort_order": t.SortOrder,
			}

			if err := tx.Model(&models.StudyReportSkillType{}).
				Where("id = ? AND skill_id = ?", t.ID, skillID).
				Updates(data).Error; err != nil {
				return err
			}

			keepTypeIDs = append(keepTypeIDs, t.ID)
		} else {
			if err := tx.Create(&t).Error; err != nil {
				return err
			}
			keepTypeIDs = append(keepTypeIDs, t.ID)
		}

		if origID != 0 {
			idMap[origID] = t.ID
		}
	}

	query := tx.Where("skill_id = ?", skillID)
	if len(keepTypeIDs) > 0 {
		query = query.Where("id NOT IN ?", keepTypeIDs)
	}

	if err := query.Delete(&models.StudyReportSkillType{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *studyReportCriteriaRepository) deleteRemovedSkills(tx *gorm.DB, criteriaID int64, keepIDs []int64) error {
	query := tx.Model(&models.StudyReportSkill{}).Where("criteria_id = ?", criteriaID)
	if len(keepIDs) > 0 {
		query = query.Where("id NOT IN ?", keepIDs)
	}

	var deleteIDs []int64
	if err := query.Pluck("id", &deleteIDs).Error; err != nil {
		return err
	}

	if len(deleteIDs) == 0 {
		return nil
	}

	if err := tx.Where("skill_id IN ?", deleteIDs).
		Delete(&models.StudyReportSkillType{}).Error; err != nil {
		return err
	}

	return tx.Where("id IN ?", deleteIDs).Delete(&models.StudyReportSkill{}).Error
}
