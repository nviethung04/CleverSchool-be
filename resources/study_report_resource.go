package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"sort"
	"strings"
)

type StudyReportResource interface {
	FormatStudyReport(report *models.StudyReport) *prot.StudyReport
	FormatStudyReports(reports []*models.StudyReport) []*prot.StudyReport
	FormatModelStudyReport(req *prot.StudyReportRequest) *models.StudyReport
	SplitReportSkills(report *models.StudyReport) ([]*prot.StudyReportStarSkillResult, []*prot.StudyReportCheckSkillResult, int, int32, float64)
	IsComplete(report *models.StudyReport) bool
}

type studyReportResourceImpl struct{}

func NewStudyReportResource() StudyReportResource {
	return &studyReportResourceImpl{}
}

func (r *studyReportResourceImpl) FormatStudyReport(report *models.StudyReport) *prot.StudyReport {
	if report == nil {
		return nil
	}

	starSkills, checkSkills, maxStar, totalStar, avgStar := r.SplitReportSkills(report)

	var assessmentScoreId int64
	if report.AssessmentScoreId != nil {
		assessmentScoreId = *report.AssessmentScoreId
	}

	result := &prot.StudyReport{
		Id:                    report.ID,
		Name:                  report.Name,
		Description:           report.Description,
		SubjectId:             report.SubjectId,
		StudyReportCriteriaId: report.StudyReportCriteriaId,
		AssessmentScoreId:     assessmentScoreId,
		AssessmentId:          report.AssessmentId,
		CourseId:              report.CourseId,
		StudentId:             report.StudentId,
		TeacherId:             report.TeacherId,
		GeneralComment:        report.GeneralComment,
		IsCompleted:           report.IsCompleted,
		StarSkills:            starSkills,
		CheckSkills:           checkSkills,
		MaxStar:               int32(maxStar),
		TotalStar:             totalStar,
		AvgStar:               avgStar,
		CreatedAt:             report.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             report.UpdatedAt.Format("2006-01-02 15:04:05"),
		Publish:               report.Publish,
	}

	if report.TotalStar != nil {
		result.TotalStar = int32(*report.TotalStar)
	}

	if report.AvgStar != nil {
		result.AvgStar = *report.AvgStar
	}

	// Format StudentInfo
	if report.Student.ID != 0 {
		className := ""
		if len(report.Student.Classes) > 0 {
			className = report.Student.Classes[0].Name
		}
		result.StudentInfo = &prot.StudyReportStudentInfo{
			Id:    report.Student.ID,
			Name:  report.Student.Name,
			Class: className,
		}
	}

	// Format TeacherInfo
	if report.Teacher.ID != 0 {
		result.TeacherInfo = &prot.StudyReportTeacherInfo{
			Id:   report.Teacher.ID,
			Name: report.Teacher.Name,
		}
	}

	// Format CourseInfo
	if report.Course.ID != 0 {
		programName := ""
		programId := int64(0)
		if report.Course.Program.ID != 0 {
			programName = report.Course.Program.Name
			programId = report.Course.Program.ID
		}
		result.CourseInfo = &prot.StudyReportCourseInfo{
			Id:        report.Course.ID,
			Name:      report.Course.Name,
			Program:   programName,
			IdProgram: programId,
		}
	}

	// Format Notes from Criteria
	if report.Criteria.ID != 0 {
		result.Notes = parseNotesJSON(report.Criteria.Notes)
	}

	if report.Assessment != nil && report.Assessment.ID != 0 {
		result.Assessment = &prot.StudyReportAssessmentInfo{
			Id:   report.Assessment.ID,
			Name: report.Assessment.Name,
			Type: report.Assessment.Type,
		}
	} else {
		result.Assessment = &prot.StudyReportAssessmentInfo{}
	}

	result.IsCompleted = r.IsComplete(report)

	return result
}

func (r *studyReportResourceImpl) FormatStudyReports(reports []*models.StudyReport) []*prot.StudyReport {
	result := make([]*prot.StudyReport, 0, len(reports))
	for _, report := range reports {
		if formatted := r.FormatStudyReport(report); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *studyReportResourceImpl) FormatModelStudyReport(req *prot.StudyReportRequest) *models.StudyReport {
	if req == nil {
		return nil
	}

	report := &models.StudyReport{
		ID:                    req.Id,
		Name:                  req.Name,
		Description:           req.Description,
		SubjectId:             req.SubjectId,
		StudyReportCriteriaId: req.StudyReportCriteriaId,
		AssessmentScoreId:     &req.AssessmentScoreId,
		CourseId:              req.CourseId,
		StudentId:             req.StudentId,
		TeacherId:             req.TeacherId,
		GeneralComment:        req.GeneralComment,
		IsCompleted:           req.IsCompleted,
		AssessmentId:          req.AssessmentId,
	}

	return report
}

func (r *studyReportResourceImpl) SplitReportSkills(report *models.StudyReport) ([]*prot.StudyReportStarSkillResult, []*prot.StudyReportCheckSkillResult, int, int32, float64) {
	starList := make([]*prot.StudyReportStarSkillResult, 0)
	checkList := make([]*prot.StudyReportCheckSkillResult, 0)

	maxStarVal := 0
	if report.Criteria.ID != 0 && report.Criteria.MaxStar != nil {
		maxStarVal = *report.Criteria.MaxStar
	}

	if report.Criteria.ID == 0 || len(report.Criteria.Skills) == 0 {
		return starList, checkList, maxStarVal, 0, 0
	}

	var leafStarSum int32
	var leafCount int32
	for _, skill := range report.Criteria.Skills {
		if skill.ID == 0 {
			continue
		}

		switch strings.ToLower(skill.Type) {
		case "star":
			starSkill := &prot.StudyReportStarSkillResult{
				Id:     skill.ID,
				NameVn: skill.NameVn,
				NameEn: skill.NameEn,
				Types:  make([]*prot.StudyReportStarTypeResult, 0),
			}

			roots, totalStar, leafCnt := buildStarTypeTreeWithValues(skill.Types, report.Skills, skill.ID)
			starSkill.Types = roots
			starSkill.AverageStar = float32(totalStar)
			if leafCnt > 0 {
				starSkill.AverageStar = float32(totalStar) / float32(leafCnt)
			}

			// accumulate totals (only leaves)
			leafStarSum += int32(totalStar)
			leafCount += int32(leafCnt)

			starList = append(starList, starSkill)

		case "check":
			checkSkill := &prot.StudyReportCheckSkillResult{
				Id:     skill.ID,
				NameVn: skill.NameVn,
				NameEn: skill.NameEn,
				Types:  make([]*prot.StudyReportCheckTypeResult, 0),
			}

			checkSkill.Types = buildCheckTypeTreeWithValues(skill.Types, report.Skills, skill.ID)

			checkList = append(checkList, checkSkill)
		}
	}

	var avgStar float64
	if leafCount > 0 {
		avgStar = float64(leafStarSum) / float64(leafCount)
	}

	return starList, checkList, maxStarVal, leafStarSum, avgStar
}

func getStarValue(refs []models.StudyReportRefSkillType, skillID, typeID int64) int32 {
	for _, ref := range refs {
		if ref.SkillId == skillID && ref.SkillTypeId == typeID {
			if ref.Star != nil {
				return int32(*ref.Star)
			}
			break
		}
	}
	return 0
}

func getCheckValue(refs []models.StudyReportRefSkillType, skillID, typeID int64) bool {
	for _, ref := range refs {
		if ref.SkillId == skillID && ref.SkillTypeId == typeID {
			return ref.IsCheck
		}
	}
	return false
}

func buildStarTypeTreeWithValues(modelTypes []models.StudyReportSkillType, refs []models.StudyReportRefSkillType, skillID int64) ([]*prot.StudyReportStarTypeResult, int, int) {
	type wrapper struct {
		node      *prot.StudyReportStarTypeResult
		sortOrder int
		parentID  *int64
	}

	nodes := make(map[int64]*wrapper, len(modelTypes))
	roots := make([]*prot.StudyReportStarTypeResult, 0)

	sorted := make([]models.StudyReportSkillType, len(modelTypes))
	copy(sorted, modelTypes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].SortOrder == sorted[j].SortOrder {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].SortOrder < sorted[j].SortOrder
	})

	totalStar := 0
	leafCount := 0

	for _, t := range sorted {
		starVal := getStarValue(refs, skillID, t.ID)
		totalStar += int(starVal)

		parentIdVal := int64(0)
		if t.ParentId != nil {
			parentIdVal = *t.ParentId
		}

		w := &wrapper{
			node: &prot.StudyReportStarTypeResult{
				Id:          t.ID,
				NameVn:      t.NameVn,
				NameEn:      t.NameEn,
				Star:        starVal,
				ParentId:    parentIdVal,
				Level:       int32(t.Level),
				NodeTypes:   make([]*prot.StudyReportStarTypeResult, 0),
				AverageStar: 0,
			},
			sortOrder: t.SortOrder,
			parentID:  t.ParentId,
		}
		nodes[t.ID] = w
	}

	for _, t := range sorted {
		w := nodes[t.ID]
		if t.ParentId != nil && nodes[*t.ParentId] != nil {
			parent := nodes[*t.ParentId]
			parent.node.NodeTypes = append(parent.node.NodeTypes, w.node)
		} else {
			roots = append(roots, w.node)
		}
	}

	var dfs func(n *prot.StudyReportStarTypeResult) (sum int, count int)
	dfs = func(n *prot.StudyReportStarTypeResult) (int, int) {
		if len(n.NodeTypes) == 0 {
			n.AverageStar = float32(n.Star)
			return int(n.Star), 1
		}
		n.Star = 0
		childSum := 0
		childCount := 0
		for _, child := range n.NodeTypes {
			cs, cc := dfs(child)
			childSum += cs
			childCount += cc
		}
		if childCount > 0 {
			n.AverageStar = float32(childSum) / float32(childCount)
		} else {
			n.AverageStar = 0
		}
		return childSum, childCount
	}
	totalStar = 0
	leafCount = 0
	for _, r := range roots {
		s, c := dfs(r)
		totalStar += s
		leafCount += c
	}

	return roots, totalStar, leafCount
}

func buildCheckTypeTreeWithValues(modelTypes []models.StudyReportSkillType, refs []models.StudyReportRefSkillType, skillID int64) []*prot.StudyReportCheckTypeResult {
	type wrapper struct {
		node      *prot.StudyReportCheckTypeResult
		sortOrder int
		parentID  *int64
	}

	nodes := make(map[int64]*wrapper, len(modelTypes))
	roots := make([]*prot.StudyReportCheckTypeResult, 0)

	sorted := make([]models.StudyReportSkillType, len(modelTypes))
	copy(sorted, modelTypes)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].SortOrder == sorted[j].SortOrder {
			return sorted[i].ID < sorted[j].ID
		}
		return sorted[i].SortOrder < sorted[j].SortOrder
	})

	for _, t := range sorted {
		isCheck := getCheckValue(refs, skillID, t.ID)
		parentIdVal := int64(0)
		if t.ParentId != nil {
			parentIdVal = *t.ParentId
		}

		w := &wrapper{
			node: &prot.StudyReportCheckTypeResult{
				Id:        t.ID,
				NameVn:    t.NameVn,
				NameEn:    t.NameEn,
				IsCheck:   isCheck,
				ParentId:  parentIdVal,
				Level:     int32(t.Level),
				NodeTypes: make([]*prot.StudyReportCheckTypeResult, 0),
			},
			sortOrder: t.SortOrder,
			parentID:  t.ParentId,
		}
		nodes[t.ID] = w
	}

	for _, t := range sorted {
		w := nodes[t.ID]
		if t.ParentId != nil && nodes[*t.ParentId] != nil {
			parent := nodes[*t.ParentId]
			parent.node.NodeTypes = append(parent.node.NodeTypes, w.node)
		} else {
			roots = append(roots, w.node)
		}
	}

	return roots
}

func (r *studyReportResourceImpl) IsComplete(report *models.StudyReport) bool {
	evaluatedMap := make(map[int64]bool)
	skillStarMap := make(map[int64]int)
	for _, ref := range report.Skills {
		if ref.Star != nil {
			skillStarMap[ref.SkillTypeId] = *ref.Star
			evaluatedMap[ref.SkillTypeId] = true
		}
	}

	for _, skill := range report.Criteria.Skills {
		if skill.Type != models.StudyReportTypeStar {
			continue
		}

		parentChildMap := make(map[int64][]models.StudyReportSkillType)
		for _, skillType := range skill.Types {
			if skillType.ParentId != nil {
				parentID := *skillType.ParentId
				parentChildMap[parentID] = append(parentChildMap[parentID], skillType)
			}
		}

		var checkNode func(nodeID int64) bool
		checkNode = func(nodeID int64) bool {
			if children, hasChildren := parentChildMap[nodeID]; hasChildren {
				for _, child := range children {
					if !checkNode(child.ID) {
						return false
					}
				}
				return true
			} else {
				if !evaluatedMap[nodeID] || (skillStarMap[nodeID] == 0) {
					return false
				}
				return true
			}
		}

		for _, skillType := range skill.Types {
			if skillType.ParentId == nil {
				if !checkNode(skillType.ID) {
					return false
				}
			}
		}
	}

	return true
}

