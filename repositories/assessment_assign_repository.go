package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"

	"gorm.io/gorm"
)

func (r *assessmentRepository) Assigned(ref models.AssessmentRefLesson) error {
	isAssign := ref.AssignedBy != nil && *ref.AssignedBy > 0

	var existing models.AssessmentRefLesson
	query := db.MasterDB.Where(
		"lesson_id = ? AND assessment_id = ?",
		ref.LessonId, ref.AssessmentId,
	)
	if ref.CourseId == 0 {
		query = query.Where("course_id IS NULL OR course_id = 0")
	} else {
		query = query.Where("course_id = ?", ref.CourseId)
	}
	err := query.First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !isAssign {
			return nil
		}
		create := ref
		if ref.CourseId == 0 {
			return db.MasterDB.Omit("CourseId").Create(&create).Error
		}
		return db.MasterDB.Create(&create).Error
	}
	if err != nil {
		return err
	}

	if !isAssign {
		return db.MasterDB.Exec(
			"UPDATE assessment_ref_lessons SET assigned_by = NULL, assigned_at = NULL WHERE id = ?",
			existing.ID,
		).Error
	}

	return db.MasterDB.Model(&existing).Updates(map[string]interface{}{
		"assigned_by": ref.AssignedBy,
		"assigned_at": ref.AssignedAt,
	}).Error
}

func (r *assessmentRepository) AssignedLesson(id int64) (*models.Assessment, error) {
	var assessment models.Assessment
	err := db.ReplicaDB.
		Preload("Lessons").
		Preload("AssessmentRefLessons").
		Where("id = ?", id).
		First(&assessment).Error
	if err != nil {
		return nil, err
	}
	return &assessment, nil
}
