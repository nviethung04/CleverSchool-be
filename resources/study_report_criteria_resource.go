package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"encoding/json"
	"sort"
	"strings"
)

type StudyReportCriteriaResource interface {
	FormatStudyReportCriteria(c *models.StudyReportCriteria) *prot.StudyReportCriteria
	FormatStudyReportCriterias(cs []*models.StudyReportCriteria) []*prot.StudyReportCriteria
	FormatModelStudyReportCriteria(req *prot.StudyReportCriteriaRequest) *models.StudyReportCriteria
}

type StudyReportCriteriaResourceImpl struct{}

func NewStudyReportCriteriaResource() StudyReportCriteriaResource {
	return &StudyReportCriteriaResourceImpl{}
}

func (r *StudyReportCriteriaResourceImpl) FormatStudyReportCriteria(c *models.StudyReportCriteria) *prot.StudyReportCriteria {
	if c == nil {
		return nil
	}

	var starSkills []*prot.StudyReportSkill
	var checkSkills []*prot.StudyReportSkill
	var subject *prot.StudyReportSubjectInfo

	for _, skill := range c.Skills {
		list := &starSkills
		if strings.ToLower(skill.Type) == "check" {
			list = &checkSkills
		}

		sort.Slice(skill.Types, func(i, j int) bool {
			if skill.Types[i].SortOrder == skill.Types[j].SortOrder {
				return skill.Types[i].ID < skill.Types[j].ID
			}
			return skill.Types[i].SortOrder < skill.Types[j].SortOrder
		})

		skillTypes := buildSkillTypeTree(skill.Types)

		*list = append(*list, &prot.StudyReportSkill{
			Id:          int64(skill.ID),
			CriteriaId:  int64(skill.CriteriaId),
			NameVn:      skill.NameVn,
			NameEn:      skill.NameEn,
			Description: skill.Description,
			Types:       skillTypes,
			Type:        skill.Type,
			SortOrder:   int32(skill.SortOrder),
		})
	}

	if c.Subject.ID > 0 {
		subject = &prot.StudyReportSubjectInfo{
			Id:   int64(c.Subject.ID),
			Name: c.Subject.Name,
		}
	}

	return &prot.StudyReportCriteria{
		Id:          c.ID,
		SubjectId:   c.SubjectId,
		Subject:     subject,
		Name:        c.Name,
		Description: c.Description,
		MaxStar:     int32(getIntOrZero(c.MaxStar)),
		Notes:       parseNotesJSON(c.Notes),
		StarSkills:  starSkills,
		CheckSkills: checkSkills,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *StudyReportCriteriaResourceImpl) FormatStudyReportCriterias(cs []*models.StudyReportCriteria) []*prot.StudyReportCriteria {
	result := make([]*prot.StudyReportCriteria, 0, len(cs))
	for _, c := range cs {
		if formatted := r.FormatStudyReportCriteria(c); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *StudyReportCriteriaResourceImpl) FormatModelStudyReportCriteria(req *prot.StudyReportCriteriaRequest) *models.StudyReportCriteria {
	if req == nil {
		return nil
	}

	return &models.StudyReportCriteria{
		ID:          req.Id,
		SubjectId:   req.SubjectId,
		Name:        req.Name,
		Description: req.Description,
		MaxStar:     intPtrOrNil(req.MaxStar),
		Notes:       notesSliceToJSON(req.Notes),
	}
}

func getIntOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func intPtrOrNil(v int32) *int {
	val := int(v)
	if val == 0 {
		return nil
	}
	return &val
}

func convertProtoSkills(skills []*prot.StudyReportSkill) []models.StudyReportSkill {
	result := make([]models.StudyReportSkill, 0, len(skills))
	for idx, skill := range skills {
		sortOrder := idx + 1
		if skill.SortOrder > 0 {
			sortOrder = int(skill.SortOrder)
		}

		modelSkill := models.StudyReportSkill{
			ID:          skill.Id,
			CriteriaId:  skill.CriteriaId,
			NameVn:      skill.NameVn,
			NameEn:      skill.NameEn,
			Description: skill.Description,
			Type:        skill.Type,
			SortOrder:   sortOrder,
		}

		typeModels := make([]models.StudyReportSkillType, 0)
		for typeIdx, skillType := range skill.Types {
			sortType := typeIdx + 1
			if skillType.SortOrder > 0 {
				sortType = int(skillType.SortOrder)
			}
			appendProtoTypeWithChildren(&typeModels, skillType, nil, 1, sortType)
		}

		modelSkill.Types = typeModels
		result = append(result, modelSkill)
	}
	return result
}

func buildSkillTypeTree(modelTypes []models.StudyReportSkillType) []*prot.StudyReportSkillType {
	nodes := make(map[int64]*prot.StudyReportSkillType, len(modelTypes))
	roots := make([]*prot.StudyReportSkillType, 0)

	for _, t := range modelTypes {
		nodes[t.ID] = &prot.StudyReportSkillType{
			Id:        t.ID,
			SkillId:   t.SkillId,
			NameVn:    t.NameVn,
			NameEn:    t.NameEn,
			SortOrder: int32(t.SortOrder),
			Level:     int32(t.Level),
			ParentId: func() int64 {
				if t.ParentId != nil {
					return *t.ParentId
				}
				return 0
			}(),
		}
	}

	for _, t := range modelTypes {
		node := nodes[t.ID]
		if t.ParentId != nil && nodes[*t.ParentId] != nil {
			parent := nodes[*t.ParentId]
			parent.NodeTypes = append(parent.NodeTypes, node)
		} else {
			roots = append(roots, node)
		}
	}

	var sortTree func(list []*prot.StudyReportSkillType)
	sortTree = func(list []*prot.StudyReportSkillType) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].SortOrder == list[j].SortOrder {
				return list[i].Id < list[j].Id
			}
			return list[i].SortOrder < list[j].SortOrder
		})
		for _, n := range list {
			sortTree(n.NodeTypes)
		}
	}
	sortTree(roots)

	return roots
}

func appendProtoTypeWithChildren(dst *[]models.StudyReportSkillType, tp *prot.StudyReportSkillType, parentID *int64, level int16, sortOrder int) {
	model := models.StudyReportSkillType{
		ID:      tp.Id,
		SkillId: tp.SkillId,
		NameVn:  tp.NameVn,
		NameEn:  tp.NameEn,
		SortOrder: func() int {
			if tp.SortOrder > 0 {
				return int(tp.SortOrder)
			}
			return sortOrder
		}(),
		ParentId: parentID,
		Level:    level,
	}
	*dst = append(*dst, model)

	for idx, child := range tp.NodeTypes {
		childSort := idx + 1
		if child.SortOrder > 0 {
			childSort = int(child.SortOrder)
		}
		appendProtoTypeWithChildren(dst, child, &tp.Id, level+1, childSort)
	}
}

func parseNotesJSON(raw string) []*prot.StudyReportNote {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []*prot.StudyReportNote{}
	}

	type noteEntry struct {
		Value       int    `json:"value"`
		Text        string `json:"text"`
		Description string `json:"description"`
	}

	var entries []noteEntry
	if err := json.Unmarshal([]byte(trimmed), &entries); err != nil {
		return []*prot.StudyReportNote{}
	}

	result := make([]*prot.StudyReportNote, 0, len(entries))
	for _, entry := range entries {
		result = append(result, &prot.StudyReportNote{
			Value:       int32(entry.Value),
			Text:        entry.Text,
			Description: entry.Description,
		})
	}
	return result
}

func notesSliceToJSON(notes []*prot.StudyReportNote) string {
	if len(notes) == 0 {
		return "[]"
	}

	type noteEntry struct {
		Value       int32  `json:"value"`
		Text        string `json:"text"`
		Description string `json:"description"`
	}

	entries := make([]noteEntry, 0, len(notes))
	for _, n := range notes {
		entries = append(entries, noteEntry{
			Value:       n.Value,
			Text:        n.Text,
			Description: n.Description,
		})
	}

	bytes, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}
