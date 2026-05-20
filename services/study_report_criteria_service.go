package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type StudyReportCriteriaService interface {
	GetAll(c *gin.Context) ([]models.StudyReportCriteria, int64, error)
	GetByID(c *gin.Context, id int) (*prot.StudyReportCriteria, error)
	Create(c *gin.Context, req *prot.StudyReportCriteriaRequest) (*models.StudyReportCriteria, error)
	Update(c *gin.Context, req *prot.StudyReportCriteriaRequest) (*models.StudyReportCriteria, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.StudyReportCriteria, error)
}

type studyReportCriteriaService struct {
	repo repositories.StudyReportCriteriaRepository
}

func NewStudyReportCriteriaService(repo repositories.StudyReportCriteriaRepository) StudyReportCriteriaService {
	return &studyReportCriteriaService{
		repo: repo,
	}
}

func (s *studyReportCriteriaService) GetAll(c *gin.Context) ([]models.StudyReportCriteria, int64, error) {
	allowedFilters := []string{"subject_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "description"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Subject",
		"Skills",
		"Skills.Types",
	})

	items, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return items, rows, nil
}

func (s *studyReportCriteriaService) GetByID(c *gin.Context, id int) (*prot.StudyReportCriteria, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Subject",
		"Skills",
		"Skills.Types",
	})
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	res := resources.NewStudyReportCriteriaResource()
	formatted := res.FormatStudyReportCriteria(item)

	return formatted, nil
}

func (s *studyReportCriteriaService) Create(c *gin.Context, req *prot.StudyReportCriteriaRequest) (*models.StudyReportCriteria, error) {
	s.repo.SetContext(c)

	// Đã truyền copy_from_study_report_criteria_id → chỉ dùng logic copy (bỏ qua logic tạo mới từ request)
	// protojson dùng camelCase; hỗ trợ thêm snake_case từ body
	copyFromID := req.GetCopyFromStudyReportCriteriaId()
	if copyFromID == 0 {
		copyFromID = getCopyFromStudyReportCriteriaIDFromBody(c)
	}
	if copyFromID > 0 {
		return s.createByCopyFromCriteria(c, copyFromID)
	}

	// Logic tạo mới theo request (subject_id, name, star_skills, check_skills, ...)
	res := resources.NewStudyReportCriteriaResource()
	model := res.FormatModelStudyReportCriteria(req)

	now := time.Now()
	model.CreatedAt = now
	model.UpdatedAt = now
	// CreatedBy được base repo BeforeCreate gán từ context; UpdatedBy set sau nếu cần

	if err := s.repo.Create(model); err != nil {
		return nil, err
	}

	if err := s.storeSkills(model.ID, req); err != nil {
		return nil, err
	}

	id := int(model.ID)
	s.repo.SetPreload([]string{
		"Subject",
		"Skills",
		"Skills.Types",
	})
	newModel, _ := s.repo.FindNewByID(id)

	return newModel, nil
}

// getCopyFromStudyReportCriteriaIDFromBody đọc body JSON và lấy copy_from_study_report_criteria_id (snake_case) vì protojson chỉ bind camelCase.
func getCopyFromStudyReportCriteriaIDFromBody(c *gin.Context) int64 {
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return 0
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		return 0
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return 0
	}
	if v, ok := m["copy_from_study_report_criteria_id"]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int:
			return int64(n)
		case int64:
			return n
		}
	}
	return 0
}

// createByCopyFromCriteria tạo study_report_criteria mới: select hàng từ study_report_criterias theo id → tạo hàng mới từ dữ liệu đó, rồi copy skills + skill_types.
func (s *studyReportCriteriaService) createByCopyFromCriteria(c *gin.Context, sourceID int64) (*models.StudyReportCriteria, error) {
	// 1. Select từ bảng study_report_criterias theo id, lấy ra hàng đó
	var sourceRow models.StudyReportCriteria
	if err := db.ReplicaDB.Table("study_report_criterias").Where("id = ?", sourceID).First(&sourceRow).Error; err != nil {
		return nil, fmt.Errorf("không tìm thấy study report criteria id %d để copy: %w", sourceID, err)
	}

	now := time.Now()
	userID := int64(utils.GetCurrentUserId(c))

	// 2. Tạo hàng mới trong study_report_criterias từ dữ liệu hàng vừa select (cùng subject_id, name, description, max_star, notes; audit mới)
	newRow := &models.StudyReportCriteria{
		SubjectId:   sourceRow.SubjectId,
		Name:        sourceRow.Name,
		Description: sourceRow.Description,
		MaxStar:     sourceRow.MaxStar,
		Notes:       sourceRow.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}
	if err := db.MasterDB.Model(&models.StudyReportCriteria{}).
		Select("SubjectId", "Name", "Description", "MaxStar", "Notes", "CreatedAt", "CreatedBy", "UpdatedAt", "UpdatedBy").
		Create(newRow).Error; err != nil {
		return nil, err
	}

	// 3. Lấy skills + types của hàng nguồn để copy sang hàng mới
	var sourceWithSkills models.StudyReportCriteria
	if err := db.ReplicaDB.Preload("Skills").Preload("Skills.Types").Where("id = ?", sourceID).First(&sourceWithSkills).Error; err != nil {
		return newRow, nil // đã tạo xong criteria, bỏ qua copy skills nếu lỗi
	}

	// 4. Copy study_report_skills và study_report_skill_types sang criteria mới (parent_id map sang id mới)
	records := make([]repositories.StudyReportSkillWithTypes, 0, len(sourceWithSkills.Skills))
	for _, sk := range sourceWithSkills.Skills {
		typesCopy := sortSkillTypesForCopy(sk.Types)
		records = append(records, repositories.StudyReportSkillWithTypes{
			Skill: models.StudyReportSkill{
				ID:          0,
				CriteriaId:  newRow.ID,
				NameVn:      sk.NameVn,
				NameEn:      sk.NameEn,
				Description: sk.Description,
				Type:        sk.Type,
				SortOrder:   sk.SortOrder,
			},
			Types: typesCopy,
		})
	}
	if len(records) > 0 {
		if err := s.repo.ReplaceSkills(newRow.ID, records); err != nil {
			return nil, fmt.Errorf("copy study_report_skills và study_report_skill_types: %w", err)
		}
	}

	id := int(newRow.ID)
	s.repo.SetPreload([]string{"Subject", "Skills", "Skills.Types"})
	newModel, _ := s.repo.FindNewByID(id)
	return newModel, nil
}

// sortSkillTypesForCopy sắp xếp types theo level, id (cha trước con) và gán ID = -originalID để repository map parent_id sang id mới.
func sortSkillTypesForCopy(types []models.StudyReportSkillType) []models.StudyReportSkillType {
	if len(types) == 0 {
		return nil
	}
	flat := make([]models.StudyReportSkillType, len(types))
	copy(flat, types)
	sort.Slice(flat, func(i, j int) bool {
		if flat[i].Level != flat[j].Level {
			return flat[i].Level < flat[j].Level
		}
		return flat[i].ID < flat[j].ID
	})
	out := make([]models.StudyReportSkillType, 0, len(flat))
	for _, t := range flat {
		out = append(out, models.StudyReportSkillType{
			ID:        -t.ID, // âm để syncSkillTypes dùng origID cho idMap (parent_id remap)
			SkillId:   t.SkillId,
			NameVn:    t.NameVn,
			NameEn:    t.NameEn,
			SortOrder: t.SortOrder,
			ParentId:  t.ParentId,
			Level:     t.Level,
		})
	}
	return out
}

func (s *studyReportCriteriaService) Update(c *gin.Context, req *prot.StudyReportCriteriaRequest) (*models.StudyReportCriteria, error) {
	res := resources.NewStudyReportCriteriaResource()
	model := res.FormatModelStudyReportCriteria(req)

	s.repo.SetContext(c)

	if err := s.repo.Update(model); err != nil {
		return nil, err
	}

	if err := s.storeSkills(model.ID, req); err != nil {
		return nil, err
	}

	id := int(model.ID)
	s.repo.SetPreload([]string{
		"Subject",
		"Skills",
		"Skills.Types",
	})
	updatedModel, _ := s.repo.FindNewByID(id)

	return updatedModel, nil
}

func (s *studyReportCriteriaService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *studyReportCriteriaService) Restore(c *gin.Context, id int) (*models.StudyReportCriteria, error) {
	s.repo.SetContext(c)
	item, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *studyReportCriteriaService) storeSkills(criteriaID int64, req *prot.StudyReportCriteriaRequest) error {
	merged := make([]*prot.StudyReportSkill, 0)
	tempID := int64(-1)

	for _, sk := range req.StarSkills {
		sk.Type = models.StudyReportTypeStar
		merged = append(merged, sk)
	}

	for _, sk := range req.CheckSkills {
		sk.Type = models.StudyReportTypeCheck
		merged = append(merged, sk)
	}

	records := make([]repositories.StudyReportSkillWithTypes, 0, len(merged))

	for index, skill := range merged {
		sortOrder := index + 1
		if skill.SortOrder > 0 {
			sortOrder = int(skill.SortOrder)
		}

		record := repositories.StudyReportSkillWithTypes{
			Skill: models.StudyReportSkill{
				ID:          skill.Id,
				CriteriaId:  criteriaID,
				NameVn:      skill.NameVn,
				NameEn:      skill.NameEn,
				Description: skill.Description,
				Type:        skill.Type,
				SortOrder:   sortOrder,
			},
		}

		for indexType, tp := range skill.Types {
			sortOrderType := indexType + 1
			if tp.SortOrder > 0 {
				sortOrderType = int(tp.SortOrder)
			}

			appendSkillTypeWithChildren(&record.Types, tp, nil, 1, sortOrderType, &tempID)
		}

		records = append(records, record)
	}

	return s.repo.ReplaceSkills(criteriaID, records)
}

func appendSkillTypeWithChildren(dst *[]models.StudyReportSkillType, tp *prot.StudyReportSkillType, parentID *int64, level int16, sortOrder int, tempID *int64) {
	if tp.Id == 0 && tempID != nil {
		tp.Id = *tempID
		*tempID--
	}
	currentID := tp.Id

	*dst = append(*dst, models.StudyReportSkillType{
		ID:     tp.Id,
		NameVn: tp.NameVn,
		NameEn: tp.NameEn,
		SortOrder: func() int {
			if tp.SortOrder > 0 {
				return int(tp.SortOrder)
			}
			return sortOrder
		}(),
		ParentId: parentID,
		Level:    level,
	})

	for idx, child := range tp.NodeTypes {
		childSort := idx + 1
		if child.SortOrder > 0 {
			childSort = int(child.SortOrder)
		}
		appendSkillTypeWithChildren(dst, child, &currentID, level+1, childSort, tempID)
	}
}

