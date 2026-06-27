package services

import (
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/utils"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type AssessmentScoringService interface {
	SaveScoresBulk(req *requests.AssessmentSaveScoreBulkRequest, actorID int64) (map[string]string, error)
	GetTeacherStudents(req *requests.AssessmentStudentsRequest) (map[string]interface{}, error)
	GetPublish(courseID, assessmentID int64) (map[string]interface{}, error)
	SetPublish(req *requests.AssessmentPublishRequest) error
	StudentSubmit(userID int64, req *requests.AssessmentSubmitRequest) error

	CreateCriteriaGroup(req *requests.CreateGroupWithCriteriaRequest, actorID int64) (map[string]interface{}, error)
	UpdateCriteriaGroup(id int64, req *requests.CreateGroupWithCriteriaRequest, actorID int64) (map[string]interface{}, error)
	ListCriteriaGroups(subjectID *int64, keyword string, page, limit int) (map[string]interface{}, error)
	GetCriteriaGroup(id int64) (map[string]interface{}, error)
	DeleteCriteriaGroup(id int64, actorID int64) error

	ListStudyReportCriterias(subjectID *int64, keyword string, page, limit int) (map[string]interface{}, error)
	GetStudyReportCriteria(id int64) (map[string]interface{}, error)
	CreateStudyReportCriteria(req *requests.StudyReportCriteriaCreateRequest) (map[string]interface{}, error)
	UpdateStudyReportCriteria(id int64, req *requests.StudyReportCriteriaCreateRequest) (map[string]interface{}, error)
	DeleteStudyReportCriteria(id int64) error

	ListStudyReports(req *requests.StudyReportListRequest, viewerUserID int64, isStudent bool) (map[string]interface{}, error)
	GetStudyReportDetail(id int64, viewerUserID int64, isStudent bool) (map[string]interface{}, error)
	CreateStudyReport(req *requests.StudyReportCreateRequest, actorID int64) (map[string]interface{}, error)
	UpdateStudyReport(id int64, req *requests.StudyReportCreateRequest, actorID int64) (map[string]interface{}, error)
	GetStudyReportEvaluates(courseID int64, req *requests.StudyReportEvaluatesRequest) (map[string]interface{}, error)
	SetStudyReportPublish(req *requests.StudyReportPublishRequest) error

	BuildStudentAssessmentItem(assessment *models.Assessment, userID, courseID int64, published bool) map[string]interface{}
	LoadCriteriaWithScores(groupID int64, score *models.AssessmentScore) []map[string]interface{}
}

type assessmentScoringService struct {
	repo repositories.AssessmentScoringRepository
}

func NewAssessmentScoringService(repo repositories.AssessmentScoringRepository) AssessmentScoringService {
	return &assessmentScoringService{repo: repo}
}

func (s *assessmentScoringService) SaveScoresBulk(req *requests.AssessmentSaveScoreBulkRequest, actorID int64) (map[string]string, error) {
	assessment, err := s.repo.GetAssessmentByID(req.AssessmentID)
	if err != nil {
		return nil, err
	}
	for _, st := range req.Students {
		total, details, isScored := s.buildScoreDetails(assessment.AssessmentCriteriaGroupId, st.Criteria)
		fileInfos := parseFileInfos(st.FileInfos)
		score := &models.AssessmentScore{
			UserId:       st.UserID,
			AssessmentId: req.AssessmentID,
			CourseId:     req.CourseID,
			TotalScore:   total,
			FileInfos:    fileInfos,
			IsScored:     isScored,
			UpdatedBy:    actorID,
			CreatedBy:    actorID,
		}
		if err := s.repo.UpsertScore(score, details); err != nil {
			return nil, err
		}
	}
	return map[string]string{"message": "Scores saved successfully"}, nil
}

func (s *assessmentScoringService) buildScoreDetails(groupID int64, criteriaReq []requests.AssessmentSaveScoreCriteriaRequest) (float64, []models.AssessmentScoreDetail, bool) {
	details := make([]models.AssessmentScoreDetail, 0)
	var total float64
	isScored := false
	if groupID <= 0 {
		return 0, details, false
	}
	group, err := s.repo.GetCriteriaGroupWithCriteria(groupID)
	if err != nil {
		return 0, details, false
	}
	reqMap := make(map[int64]requests.AssessmentSaveScoreCriteriaRequest)
	for _, c := range criteriaReq {
		reqMap[c.CriteriaID] = c
	}
	for _, c := range group.Criteria {
		reqC, ok := reqMap[c.ID]
		if !ok {
			continue
		}
		if c.HasSubcriteria && len(c.Subcriteria) > 0 {
			subMap := make(map[int64]float64)
			for _, sub := range reqC.Subcriteria {
				subMap[sub.SubcriteriaID] = sub.Score
			}
			var critTotal float64
			for _, sub := range c.Subcriteria {
				sc := subMap[sub.ID]
				if sc >= 0 {
					isScored = true
					critTotal += sc
				}
				sid := sub.ID
				details = append(details, models.AssessmentScoreDetail{
					CriteriaId:    c.ID,
					SubcriteriaId: &sid,
					Score:         sc,
				})
			}
			total += critTotal
		} else {
			sc := reqC.Score
			if sc >= 0 {
				isScored = true
				total += sc
			}
			details = append(details, models.AssessmentScoreDetail{
				CriteriaId: c.ID,
				Score:      sc,
			})
		}
	}
	return total, details, isScored
}

func parseFileInfos(items []map[string]interface{}) models.MediaInfos {
	result := make(models.MediaInfos, 0, len(items))
	for _, item := range items {
		id := int64(0)
		switch v := item["id"].(type) {
		case float64:
			id = int64(v)
		case int64:
			id = v
		}
		disk, _ := item["disk"].(string)
		path, _ := item["path"].(string)
		if path == "" {
			continue
		}
		result = append(result, models.MediaDetail{Id: id, Disk: disk, Path: path})
	}
	return result
}

func (s *assessmentScoringService) GetTeacherStudents(req *requests.AssessmentStudentsRequest) (map[string]interface{}, error) {
	assessment, err := s.repo.GetAssessmentByID(req.AssessmentID)
	if err != nil {
		return nil, err
	}
	studentIDs, err := s.repo.ListCourseStudentIDs(req.CourseID)
	if err != nil {
		return nil, err
	}
	if req.StudentID != nil && *req.StudentID > 0 {
		filtered := make([]int64, 0, 1)
		for _, id := range studentIDs {
			if id == *req.StudentID {
				filtered = append(filtered, id)
			}
		}
		studentIDs = filtered
	}
	names, _ := s.repo.GetStudentNames(studentIDs)
	scores, _ := s.repo.GetScoresByAssessmentCourse(req.AssessmentID, req.CourseID)
	scoreMap := make(map[int64]*models.AssessmentScore)
	for i := range scores {
		scoreMap[scores[i].UserId] = &scores[i]
	}

	var group *models.AssessmentCriteriaGroup
	if assessment.AssessmentCriteriaGroupId > 0 {
		group, _ = s.repo.GetCriteriaGroupWithCriteria(assessment.AssessmentCriteriaGroupId)
	}

	courseName := ""
	objectTitle := ""
	teacherNames := ""
	var course models.Course
	if err := db.ReplicaDB.Where("id = ?", req.CourseID).First(&course).Error; err == nil {
		courseName = course.Name
		objectTitle = course.ObjectTitle
	}

	students := make([]map[string]interface{}, 0, len(studentIDs))
	for _, sid := range studentIDs {
		sc := scoreMap[sid]
		item := s.formatTeacherStudentItem(assessment, group, sc, sid, names[sid])
		students = append(students, item)
	}

	return map[string]interface{}{
		"@type": "type.googleapis.com/prot.AssessmentStudentListResponse",
		"students": students,
		"total":    strconv.Itoa(len(students)),
		"course_info": map[string]string{
			"name":          courseName,
			"object_title":  objectTitle,
			"teacher_names": teacherNames,
		},
	}, nil
}

func (s *assessmentScoringService) formatTeacherStudentItem(
	assessment *models.Assessment,
	group *models.AssessmentCriteriaGroup,
	score *models.AssessmentScore,
	studentID int64,
	studentName string,
) map[string]interface{} {
	var criteria []map[string]interface{}
	totalScore := 0.0
	isScored := false
	files := []string{}
	var fileInfos []map[string]interface{}

	if score != nil {
		totalScore = score.TotalScore
		isScored = score.IsScored
		for _, fi := range score.FileInfos {
			fileInfos = append(fileInfos, map[string]interface{}{
				"id": fi.Id, "disk": fi.Disk, "path": fi.Path,
			})
			if url := utils.StaticURL(fi.Path, models.Storage); url != "" {
				files = append(files, url)
			}
		}
		criteria = s.LoadCriteriaWithScores(assessment.AssessmentCriteriaGroupId, score)
	} else if group != nil {
		criteria = s.LoadCriteriaWithScores(assessment.AssessmentCriteriaGroupId, nil)
	}

	groupName := ""
	groupIDStr := "0"
	if group != nil {
		groupName = group.Name
		groupIDStr = strconv.FormatInt(group.ID, 10)
	}

	subjectName := ""
	if assessment.Subject != nil {
		subjectName = assessment.Subject.Name
	}

	assessmentFiles := []string{}
	for _, fi := range assessment.FileInfos {
		if url := utils.StaticURL(fi.Path, models.Storage); url != "" {
			assessmentFiles = append(assessmentFiles, url)
		}
	}

	return map[string]interface{}{
		"student_id":   strconv.FormatInt(studentID, 10),
		"student_name": studentName,
		"total_score":  totalScore,
		"files":        files,
		"file_infos":   fileInfos,
		"is_scored":    isScored,
		"assessment": map[string]interface{}{
			"id":                             strconv.FormatInt(assessment.ID, 10),
			"name":                           assessment.Name,
			"description":                    assessment.Description,
			"type":                           assessment.Type,
			"program_id":                     strconv.FormatInt(assessment.ProgramId, 10),
			"created_at":                     assessment.CreatedAt.Format("2006-01-02 15:04:05"),
			"created_by":                     strconv.FormatInt(assessment.CreatedBy, 10),
			"updated_at":                     assessment.UpdatedAt.Format("2006-01-02 15:04:05"),
			"updated_by":                     strconv.FormatInt(assessment.UpdatedBy, 10),
			"study_report_criteria_id":       strconv.FormatInt(assessment.StudyReportCriteriaId, 10),
			"subject_id":                     strconv.FormatInt(assessment.SubjectId, 10),
			"assessment_criteria_group_id":   groupIDStr,
			"assessment_criteria_group_name": groupName,
			"study_report_criteria_name":     "",
			"subject_name":                   subjectName,
			"criteria":                       criteria,
			"files":                          assessmentFiles,
			"has_file":                       len(assessment.FileInfos) > 0,
		},
	}
}

func (s *assessmentScoringService) LoadCriteriaWithScores(groupID int64, score *models.AssessmentScore) []map[string]interface{} {
	if groupID <= 0 {
		return []map[string]interface{}{}
	}
	group, err := s.repo.GetCriteriaGroupWithCriteria(groupID)
	if err != nil {
		return []map[string]interface{}{}
	}
	detailMap := make(map[string]float64)
	if score != nil {
		for _, d := range score.Details {
			key := fmt.Sprintf("%d:%v", d.CriteriaId, d.SubcriteriaId)
			detailMap[key] = d.Score
		}
	}
	result := make([]map[string]interface{}, 0, len(group.Criteria))
	for _, c := range group.Criteria {
		item := map[string]interface{}{
			"id":              strconv.FormatInt(c.ID, 10),
			"subject_id":      strconv.FormatInt(group.SubjectId, 10),
			"name":            c.Name,
			"description":     c.Description,
			"has_subcriteria": c.HasSubcriteria,
			"max_score":       c.MaxScore,
			"created_at":      c.CreatedAt.Format("2006-01-02 15:04:05"),
			"created_by":      strconv.FormatInt(c.CreatedBy, 10),
			"updated_at":      c.UpdatedAt.Format("2006-01-02 15:04:05"),
			"updated_by":      strconv.FormatInt(c.UpdatedBy, 10),
			"score":           -1.0,
			"subcriteria":     []map[string]interface{}{},
		}
		if c.HasSubcriteria {
			subs := make([]map[string]interface{}, 0, len(c.Subcriteria))
			var critScore float64
			for _, sub := range c.Subcriteria {
				key := fmt.Sprintf("%d:%v", c.ID, sub.ID)
				sc := detailMap[key]
				if sc >= 0 {
					critScore += sc
				} else {
					sc = -1
				}
				subs = append(subs, map[string]interface{}{
					"id": strconv.FormatInt(sub.ID, 10), "name": sub.Name,
					"description": sub.Description, "max_score": sub.MaxScore, "score": sc,
				})
			}
			item["subcriteria"] = subs
			item["score"] = critScore
		} else {
			key := fmt.Sprintf("%d:%v", c.ID, nil)
			if sc, ok := detailMap[key]; ok {
				item["score"] = sc
			}
		}
		result = append(result, item)
	}
	return result
}

func (s *assessmentScoringService) BuildStudentAssessmentItem(assessment *models.Assessment, userID, courseID int64, published bool) map[string]interface{} {
	var score *models.AssessmentScore
	isScored := false
	totalScore := 0.0
	files := []string{}
	if published {
		if sc, err := s.repo.GetScore(userID, assessment.ID, courseID); err == nil {
			score = sc
			isScored = sc.IsScored
			totalScore = sc.TotalScore
			for _, fi := range sc.FileInfos {
				if url := utils.StaticURL(fi.Path, models.Storage); url != "" {
					files = append(files, url)
				}
			}
		}
	}
	criteria := []map[string]interface{}{}
	if published && score != nil {
		criteria = s.LoadCriteriaWithScores(assessment.AssessmentCriteriaGroupId, score)
	} else if assessment.AssessmentCriteriaGroupId > 0 {
		criteria = s.LoadCriteriaWithScores(assessment.AssessmentCriteriaGroupId, nil)
		if !published {
			for _, c := range criteria {
				c["score"] = 0.0
				if subs, ok := c["subcriteria"].([]map[string]interface{}); ok {
					for _, sub := range subs {
						sub["score"] = 0.0
					}
				}
			}
			isScored = false
			totalScore = 0
		}
	}

	groupName := ""
	if assessment.AssessmentCriteriaGroupId > 0 {
		if g, err := s.repo.GetCriteriaGroupWithCriteria(assessment.AssessmentCriteriaGroupId); err == nil {
			groupName = g.Name
		}
	}
	subjectName := ""
	if assessment.Subject != nil {
		subjectName = assessment.Subject.Name
	}
	assessmentFiles := []string{}
	for _, fi := range assessment.FileInfos {
		if url := utils.StaticURL(fi.Path, models.Storage); url != "" {
			assessmentFiles = append(assessmentFiles, url)
		}
	}

	return map[string]interface{}{
		"assessment": map[string]interface{}{
			"id":                             strconv.FormatInt(assessment.ID, 10),
			"name":                           assessment.Name,
			"description":                    assessment.Description,
			"type":                           assessment.Type,
			"program_id":                     strconv.FormatInt(assessment.ProgramId, 10),
			"created_at":                     assessment.CreatedAt.Format("2006-01-02 15:04:05"),
			"created_by":                     strconv.FormatInt(assessment.CreatedBy, 10),
			"updated_at":                     assessment.UpdatedAt.Format("2006-01-02 15:04:05"),
			"updated_by":                     strconv.FormatInt(assessment.UpdatedBy, 10),
			"study_report_criteria_id":       strconv.FormatInt(assessment.StudyReportCriteriaId, 10),
			"subject_id":                     strconv.FormatInt(assessment.SubjectId, 10),
			"assessment_criteria_group_id":   strconv.FormatInt(assessment.AssessmentCriteriaGroupId, 10),
			"assessment_criteria_group_name": groupName,
			"study_report_criteria_name":     "",
			"subject_name":                   subjectName,
			"criteria":                       criteria,
			"files":                          assessmentFiles,
			"has_file":                       len(assessment.FileInfos) > 0,
			"is_assigned":                    true,
		},
		"is_scored":    isScored,
		"total_score":  totalScore,
		"files":        files,
	}
}

func (s *assessmentScoringService) GetPublish(courseID, assessmentID int64) (map[string]interface{}, error) {
	pub, err := s.repo.GetAssessmentPublish(courseID, assessmentID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"@type":         "type.googleapis.com/prot.AssessmentPublishResponse",
		"course_id":     strconv.FormatInt(courseID, 10),
		"assessment_id": strconv.FormatInt(assessmentID, 10),
		"publish":       pub,
	}, nil
}

func (s *assessmentScoringService) SetPublish(req *requests.AssessmentPublishRequest) error {
	return s.repo.SetAssessmentPublish(req.CourseID, req.AssessmentID, req.Publish)
}

func (s *assessmentScoringService) StudentSubmit(userID int64, req *requests.AssessmentSubmitRequest) error {
	var assignedCount int64
	if err := db.ReplicaDB.Table("assessment_ref_lessons").
		Where("assessment_id = ? AND course_id = ? AND assigned_by IS NOT NULL AND assigned_by > 0", req.AssessmentID, req.CourseID).
		Count(&assignedCount).Error; err != nil {
		return err
	}
	if assignedCount == 0 {
		return fmt.Errorf(i18n.Localize("messages.no_records_found"))
	}

	fileInfos := parseFileInfos(req.FileInfos)
	score := &models.AssessmentScore{
		UserId:       userID,
		AssessmentId: req.AssessmentID,
		CourseId:     req.CourseID,
		FileInfos:    fileInfos,
		IsScored:     false,
		TotalScore:   0,
		CreatedBy:    userID,
		UpdatedBy:    userID,
	}
	return s.repo.UpsertScore(score, nil)
}

// --- Criteria groups ---

func (s *assessmentScoringService) CreateCriteriaGroup(req *requests.CreateGroupWithCriteriaRequest, actorID int64) (map[string]interface{}, error) {
	group := &models.AssessmentCriteriaGroup{
		Name: req.Group.Name, SubjectId: req.Group.SubjectID, HasFile: req.Group.HasFile,
		CreatedBy: actorID, UpdatedBy: actorID,
	}
	if req.Group.CopyFromAssessmentCriteriaGroupID > 0 {
		src, err := s.repo.GetCriteriaGroupByID(req.Group.CopyFromAssessmentCriteriaGroupID)
		if err == nil {
			criteria := copyCriteriaFromGroup(src)
			if err := s.repo.CreateCriteriaGroup(group, criteria); err != nil {
				return nil, err
			}
			return s.formatCriteriaGroup(group.ID)
		}
	}
	criteria := parseCriteriaFromRequest(req)
	if err := s.repo.CreateCriteriaGroup(group, criteria); err != nil {
		return nil, err
	}
	return s.formatCriteriaGroup(group.ID)
}

func copyCriteriaFromGroup(src *models.AssessmentCriteriaGroup) []models.AssessmentCriterion {
	out := make([]models.AssessmentCriterion, 0, len(src.Criteria))
	for _, c := range src.Criteria {
		nc := models.AssessmentCriterion{
			Name: c.Name, Description: c.Description, MaxScore: c.MaxScore, SortOrder: c.SortOrder,
		}
		for _, sub := range c.Subcriteria {
			nc.Subcriteria = append(nc.Subcriteria, models.AssessmentSubcriterion{
				Name: sub.Name, Description: sub.Description, MaxScore: sub.MaxScore, SortOrder: sub.SortOrder,
			})
		}
		out = append(out, nc)
	}
	return out
}

func parseCriteriaFromRequest(req *requests.CreateGroupWithCriteriaRequest) []models.AssessmentCriterion {
	out := make([]models.AssessmentCriterion, 0, len(req.Criteria))
	for i, c := range req.Criteria {
		nc := models.AssessmentCriterion{
			Name: c.Name, Description: c.Description, MaxScore: c.MaxScore, SortOrder: i,
		}
		for j, sub := range c.Subcriteria {
			nc.Subcriteria = append(nc.Subcriteria, models.AssessmentSubcriterion{
				Name: sub.Name, Description: sub.Description, MaxScore: sub.MaxScore, SortOrder: j,
			})
		}
		out = append(out, nc)
	}
	return out
}

func (s *assessmentScoringService) UpdateCriteriaGroup(id int64, req *requests.CreateGroupWithCriteriaRequest, actorID int64) (map[string]interface{}, error) {
	group, err := s.repo.GetCriteriaGroupByID(id)
	if err != nil {
		return nil, err
	}
	group.Name = req.Group.Name
	group.SubjectId = req.Group.SubjectID
	group.HasFile = req.Group.HasFile
	group.UpdatedBy = actorID
	criteria := parseCriteriaFromRequest(req)
	if err := s.repo.UpdateCriteriaGroup(group, criteria); err != nil {
		return nil, err
	}
	return s.formatCriteriaGroup(id)
}

func (s *assessmentScoringService) formatCriteriaGroup(id int64) (map[string]interface{}, error) {
	group, err := s.repo.GetCriteriaGroupByID(id)
	if err != nil {
		return nil, err
	}
	criteria := make([]map[string]interface{}, 0, len(group.Criteria))
	for _, c := range group.Criteria {
		subs := make([]map[string]interface{}, 0, len(c.Subcriteria))
		for _, sub := range c.Subcriteria {
			subs = append(subs, map[string]interface{}{
				"id": strconv.FormatInt(sub.ID, 10), "name": sub.Name,
				"description": sub.Description, "max_score": sub.MaxScore,
			})
		}
		criteria = append(criteria, map[string]interface{}{
			"id": strconv.FormatInt(c.ID, 10), "name": c.Name, "description": c.Description,
			"has_subcriteria": c.HasSubcriteria, "max_score": c.MaxScore, "subcriteria": subs,
		})
	}
	return map[string]interface{}{
		"@type": "type.googleapis.com/prot.AssessmentCriteriaGroup",
		"id": strconv.FormatInt(group.ID, 10), "name": group.Name,
		"subject_id": strconv.FormatInt(group.SubjectId, 10), "has_file": group.HasFile,
		"created_at": group.CreatedAt.Format("2006-01-02 15:04:05"),
		"created_by": strconv.FormatInt(group.CreatedBy, 10),
		"updated_at": group.UpdatedAt.Format("2006-01-02 15:04:05"),
		"updated_by": strconv.FormatInt(group.UpdatedBy, 10),
		"criteria": criteria,
	}, nil
}

func (s *assessmentScoringService) ListCriteriaGroups(subjectID *int64, keyword string, page, limit int) (map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	groups, total, err := s.repo.ListCriteriaGroups(subjectID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}
	formatted := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		full, _ := s.formatCriteriaGroup(g.ID)
		if full != nil {
			formatted = append(formatted, full)
		}
	}
	return map[string]interface{}{
		"@type": "type.googleapis.com/prot.AssessmentCriteriaGroupsResponse",
		"groups": formatted, "total": strconv.FormatInt(total, 10),
	}, nil
}

func (s *assessmentScoringService) GetCriteriaGroup(id int64) (map[string]interface{}, error) {
	return s.formatCriteriaGroup(id)
}

func (s *assessmentScoringService) DeleteCriteriaGroup(id int64, actorID int64) error {
	return s.repo.DeleteCriteriaGroup(id, actorID)
}

// --- Study report criteria ---

func (s *assessmentScoringService) formatStudyReportCriteria(c *models.StudyReportCriteria) map[string]interface{} {
	var notes, starSkills, checkSkills interface{}
	_ = json.Unmarshal([]byte(c.Notes), &notes)
	_ = json.Unmarshal([]byte(c.StarSkills), &starSkills)
	_ = json.Unmarshal([]byte(c.CheckSkills), &checkSkills)
	return map[string]interface{}{
		"id": strconv.FormatInt(c.ID, 10), "subject_id": strconv.FormatInt(c.SubjectId, 10),
		"name": c.Name, "description": c.Description, "max_star": c.MaxStar,
		"created_at": c.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at": c.UpdatedAt.Format("2006-01-02 15:04:05"),
		"notes": notes, "star_skills": starSkills, "check_skills": checkSkills,
	}
}

func (s *assessmentScoringService) ListStudyReportCriterias(subjectID *int64, keyword string, page, limit int) (map[string]interface{}, error) {
	if limit <= 0 {
		limit = 100
	}
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	list, total, err := s.repo.ListStudyReportCriterias(subjectID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(list))
	for _, c := range list {
		items = append(items, s.formatStudyReportCriteria(&c))
	}
	return map[string]interface{}{
		"study_report_criterias": items,
		"total_count":            strconv.FormatInt(total, 10),
	}, nil
}

func (s *assessmentScoringService) GetStudyReportCriteria(id int64) (map[string]interface{}, error) {
	c, err := s.repo.GetStudyReportCriteriaByID(id)
	if err != nil {
		return nil, err
	}
	return s.formatStudyReportCriteria(c), nil
}

func (s *assessmentScoringService) CreateStudyReportCriteria(req *requests.StudyReportCriteriaCreateRequest) (map[string]interface{}, error) {
	c := &models.StudyReportCriteria{
		SubjectId: req.SubjectID, Name: req.Name, Description: req.Description, MaxStar: req.MaxStar,
		Notes: repositories.MarshalJSONRaw(req.Notes),
		StarSkills: repositories.MarshalJSONRaw(req.StarSkills),
		CheckSkills: repositories.MarshalJSONRaw(req.CheckSkills),
	}
	if err := s.repo.CreateStudyReportCriteria(c); err != nil {
		return nil, err
	}
	return s.formatStudyReportCriteria(c), nil
}

func (s *assessmentScoringService) UpdateStudyReportCriteria(id int64, req *requests.StudyReportCriteriaCreateRequest) (map[string]interface{}, error) {
	c, err := s.repo.GetStudyReportCriteriaByID(id)
	if err != nil {
		return nil, err
	}
	c.SubjectId = req.SubjectID
	c.Name = req.Name
	c.Description = req.Description
	c.MaxStar = req.MaxStar
	c.Notes = repositories.MarshalJSONRaw(req.Notes)
	c.StarSkills = repositories.MarshalJSONRaw(req.StarSkills)
	c.CheckSkills = repositories.MarshalJSONRaw(req.CheckSkills)
	if err := s.repo.UpdateStudyReportCriteria(c); err != nil {
		return nil, err
	}
	return s.formatStudyReportCriteria(c), nil
}

func (s *assessmentScoringService) DeleteStudyReportCriteria(id int64) error {
	return s.repo.DeleteStudyReportCriteria(id)
}

// --- Study reports ---

func (s *assessmentScoringService) formatStudyReport(r *models.StudyReport, includePublish bool, published bool) map[string]interface{} {
	var starSkills, checkSkills interface{}
	_ = json.Unmarshal([]byte(r.StarSkills), &starSkills)
	_ = json.Unmarshal([]byte(r.CheckSkills), &checkSkills)

	item := map[string]interface{}{
		"id":                        strconv.FormatInt(r.ID, 10),
		"name":                      r.Name,
		"description":               r.Description,
		"subject_id":                strconv.FormatInt(r.SubjectId, 10),
		"study_report_criteria_id":  strconv.FormatInt(r.StudyReportCriteriaId, 10),
		"course_id":                 strconv.FormatInt(r.CourseId, 10),
		"student_id":                strconv.FormatInt(r.StudentId, 10),
		"teacher_id":                strconv.FormatInt(r.TeacherId, 10),
		"general_comment":           r.GeneralComment,
		"total_star":                r.TotalStar,
		"avg_star":                  r.AvgStar,
		"type":                      r.Type,
		"type_value":                r.TypeValue,
		"is_completed":              r.IsCompleted,
		"max_star":                  r.MaxStar,
		"created_at":                r.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":                r.UpdatedAt.Format("2006-01-02 15:04:05"),
		"star_skills":               starSkills,
		"check_skills":              checkSkills,
		"notes":                     []interface{}{},
	}
	if !includePublish || published {
		// show comment
	} else {
		item["general_comment"] = ""
	}
	if r.Assessment != nil {
		item["assessment"] = map[string]interface{}{
			"id": r.Assessment.ID, "name": r.Assessment.Name,
		}
	}
	if r.Student != nil {
		item["student_info"] = map[string]string{
			"id": strconv.FormatInt(r.Student.ID, 10), "name": r.Student.Name, "class": "",
		}
	}
	if r.Teacher != nil {
		item["teacher_info"] = map[string]string{
			"id": strconv.FormatInt(r.Teacher.ID, 10), "name": r.Teacher.Name,
		}
	}
	if r.Course != nil {
		item["course_info"] = map[string]interface{}{
			"id": strconv.FormatInt(r.Course.ID, 10), "name": r.Course.Name,
			"program": "", "id_program": strconv.FormatInt(r.Course.ProgramId, 10),
		}
	}
	return item
}

func (s *assessmentScoringService) ListStudyReports(req *requests.StudyReportListRequest, viewerUserID int64, isStudent bool) (map[string]interface{}, error) {
	courseID := int64(0)
	assessmentID := int64(0)
	if req.CourseID != nil {
		courseID = *req.CourseID
	}
	if req.AssessmentID != nil {
		assessmentID = *req.AssessmentID
	}
	studentFilter := int64(0)
	if isStudent {
		studentFilter = viewerUserID
	}
	published := true
	if isStudent && assessmentID > 0 && courseID > 0 {
		pub, _ := s.repo.GetStudyReportPublish(courseID, assessmentID)
		published = pub
	}
	list, err := s.repo.ListStudyReports(courseID, assessmentID, studentFilter)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		items = append(items, s.formatStudyReport(&list[i], isStudent, published))
	}
	return map[string]interface{}{
		"study_reports": items,
		"total_count":   strconv.Itoa(len(items)),
	}, nil
}

func (s *assessmentScoringService) GetStudyReportDetail(id int64, viewerUserID int64, isStudent bool) (map[string]interface{}, error) {
	r, err := s.repo.GetStudyReportByID(id)
	if err != nil {
		return nil, err
	}
	if isStudent && r.StudentId != viewerUserID {
		return nil, errors.New("forbidden")
	}
	published := true
	if isStudent {
		pub, _ := s.repo.GetStudyReportPublish(r.CourseId, r.AssessmentId)
		published = pub
	}
	return s.formatStudyReport(r, isStudent, published), nil
}

func (s *assessmentScoringService) buildStudyReportModel(req *requests.StudyReportCreateRequest, actorID int64) *models.StudyReport {
	totalStar, avgStar := calcStarTotals(req.StarSkills)
	assessment, _ := s.repo.GetAssessmentByID(req.AssessmentID)
	name := ""
	reportType := ""
	typeValue := ""
	maxStar := 5
	if assessment != nil {
		name = assessment.Name
		reportType = assessment.Type
	}
	if src, err := s.repo.GetStudyReportCriteriaByID(req.StudyReportCriteriaID); err == nil {
		maxStar = src.MaxStar
	}
	return &models.StudyReport{
		Name: name, SubjectId: req.SubjectID,
		StudyReportCriteriaId: req.StudyReportCriteriaID,
		AssessmentId: req.AssessmentID, CourseId: req.CourseID,
		StudentId: req.StudentID, TeacherId: actorID,
		GeneralComment: req.GeneralComment, IsCompleted: req.IsCompleted,
		TotalStar: totalStar, AvgStar: avgStar, MaxStar: maxStar,
		Type: reportType, TypeValue: typeValue,
		StarSkills:  repositories.MarshalJSONRaw(req.StarSkills),
		CheckSkills: repositories.MarshalJSONRaw(req.CheckSkills),
	}
}

func calcStarTotals(starSkills []map[string]interface{}) (float64, float64) {
	var sum float64
	var count float64
	var walk func(nodes []interface{})
	walk = func(nodes []interface{}) {
		for _, n := range nodes {
			m, ok := n.(map[string]interface{})
			if !ok {
				continue
			}
			if star, ok := m["star"].(float64); ok && star > 0 {
				sum += star
				count++
			}
			if children, ok := m["node_types"].([]interface{}); ok {
				walk(children)
			}
			if types, ok := m["types"].([]interface{}); ok {
				walk(types)
			}
		}
	}
	raw, _ := json.Marshal(starSkills)
	var arr []interface{}
	_ = json.Unmarshal(raw, &arr)
	walk(arr)
	if count == 0 {
		return 0, 0
	}
	avg := sum / count
	return sum, math.Round(avg*100) / 100
}

func (s *assessmentScoringService) CreateStudyReport(req *requests.StudyReportCreateRequest, actorID int64) (map[string]interface{}, error) {
	if existing, err := s.repo.GetStudyReportByKeys(req.AssessmentID, req.CourseID, req.StudentID); err == nil && existing != nil {
		return s.UpdateStudyReport(existing.ID, req, actorID)
	}
	report := s.buildStudyReportModel(req, actorID)
	if actorID > 0 {
		report.TeacherId = actorID
	}
	if err := s.repo.CreateStudyReport(report); err != nil {
		return nil, err
	}
	created, _ := s.repo.GetStudyReportByID(report.ID)
	return s.formatStudyReport(created, false, true), nil
}

func (s *assessmentScoringService) UpdateStudyReport(id int64, req *requests.StudyReportCreateRequest, actorID int64) (map[string]interface{}, error) {
	report, err := s.repo.GetStudyReportByID(id)
	if err != nil {
		return nil, err
	}
	totalStar, avgStar := calcStarTotals(req.StarSkills)
	report.GeneralComment = req.GeneralComment
	report.IsCompleted = req.IsCompleted
	report.StarSkills = repositories.MarshalJSONRaw(req.StarSkills)
	report.CheckSkills = repositories.MarshalJSONRaw(req.CheckSkills)
	report.TotalStar = totalStar
	report.AvgStar = avgStar
	report.UpdatedAt = time.Now()
	if actorID > 0 {
		report.TeacherId = actorID
	}
	if err := s.repo.UpdateStudyReport(report); err != nil {
		return nil, err
	}
	updated, _ := s.repo.GetStudyReportByID(id)
	return s.formatStudyReport(updated, false, true), nil
}

func (s *assessmentScoringService) GetStudyReportEvaluates(courseID int64, req *requests.StudyReportEvaluatesRequest) (map[string]interface{}, error) {
	assessmentID := int64(0)
	if req.AssessmentID != nil {
		assessmentID = *req.AssessmentID
	}
	studentIDs, err := s.repo.ListCourseStudentIDs(courseID)
	if err != nil {
		return nil, err
	}
	names, _ := s.repo.GetStudentNames(studentIDs)
	reports, _ := s.repo.ListStudyReports(courseID, assessmentID, 0)
	reportMap := make(map[int64]*models.StudyReport)
	for i := range reports {
		reportMap[reports[i].StudentId] = &reports[i]
	}
	var assessment *models.Assessment
	if assessmentID > 0 {
		assessment, _ = s.repo.GetAssessmentByID(assessmentID)
	}
	publish, _ := s.repo.GetStudyReportPublish(courseID, assessmentID)

	items := make([]map[string]interface{}, 0)
	for _, sid := range studentIDs {
		report := reportMap[sid]
		isCompleted := false
		if report != nil {
			isCompleted = report.IsCompleted
		}
		if req.IsCompleted != nil && *req.IsCompleted != isCompleted {
			continue
		}
		item := s.formatEvaluateItem(sid, names[sid], courseID, assessment, report)
		items = append(items, item)
	}
	return map[string]interface{}{
		"@type":       "type.googleapis.com/prot.StudyReportEvaluateResponse",
		"items":       items,
		"publish":     publish,
		"total_count": strconv.Itoa(len(items)),
	}, nil
}

func (s *assessmentScoringService) formatEvaluateItem(studentID int64, studentName string, courseID int64, assessment *models.Assessment, report *models.StudyReport) map[string]interface{} {
	courseName := ""
	var course models.Course
	if err := db.ReplicaDB.Where("id = ?", courseID).First(&course).Error; err == nil {
		courseName = course.Name
	}
	item := map[string]interface{}{
		"student_id":   strconv.FormatInt(studentID, 10),
		"student_name": studentName,
		"course_id":    strconv.FormatInt(courseID, 10),
		"course_name":  courseName,
		"report_id":    "0",
		"report_name":  "",
		"type":         "",
		"type_value":   "",
		"is_completed": false,
		"criteria_count": 0, "evaluated_criteria": 0,
		"has_general_comment": false,
		"general_comment":     "",
	}
	if assessment != nil {
		item["assessment_id"] = strconv.FormatInt(assessment.ID, 10)
		item["type"] = assessment.Type
	}
	if report != nil {
		var starSkills, checkSkills interface{}
		_ = json.Unmarshal([]byte(report.StarSkills), &starSkills)
		_ = json.Unmarshal([]byte(report.CheckSkills), &checkSkills)
		item["report_id"] = strconv.FormatInt(report.ID, 10)
		item["report_name"] = report.Name
		item["is_completed"] = report.IsCompleted
		item["general_comment"] = report.GeneralComment
		item["has_general_comment"] = report.GeneralComment != ""
		item["star_skills"] = starSkills
		item["check_skills"] = checkSkills
		item["total_star"] = report.TotalStar
		item["avg_star"] = report.AvgStar
	}
	return item
}

func (s *assessmentScoringService) SetStudyReportPublish(req *requests.StudyReportPublishRequest) error {
	return s.repo.SetStudyReportPublish(req.CourseID, req.AssessmentID, req.Publish)
}
