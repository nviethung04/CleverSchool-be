package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"fmt"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StudyReportService interface {
	GetAll(c *gin.Context) ([]models.StudyReport, int64, error)
	GetByID(c *gin.Context, id int) (*prot.StudyReport, error)
	Create(c *gin.Context, req *prot.StudyReportRequest) (*models.StudyReport, error)
	Update(c *gin.Context, req *prot.StudyReportRequest) (*models.StudyReport, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.StudyReport, error)

	Evaluate(c *gin.Context) (*prot.StudyReportEvaluateResponse, error)
	GetPublish(c *gin.Context) (*prot.GetPublishStudyReportResponse, error)
	UpdatePublish(c *gin.Context) (*prot.UpdatePublishStudyReportResponse, error)
}

type studyReportService struct {
	repo repositories.StudyReportRepository
}

func NewStudyReportService(repo repositories.StudyReportRepository) StudyReportService {
	return &studyReportService{repo: repo}
}

func (s *studyReportService) GetAll(c *gin.Context) ([]models.StudyReport, int64, error) {
	allowedFilters := []string{"subject_id", "course_id", "student_id", "teacher_id", "type", "is_completed", "assessment_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	// Validate assessment_id: nếu không truyền assessment_id thì trả về rỗng
	if filter["assessment_id"] == nil {
		return []models.StudyReport{}, 0, nil
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "description", "type", "type_value"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
		"Student",
		"Teacher",
		"Subject",
		"Course",
		"Course.Program",
		"Student.Classes",
		"Assessment",
	})

	roleId := utils.GetCurrentRoleId(c)

	if filter["teacher_id"] == nil && roleId == models.TeacherRoleId {
		userId := utils.GetCurrentUserId(c)
		filter["teacher_id"] = userId
	}

	if filter["student_id"] == nil && roleId == models.StudentRoleId {
		userId := utils.GetCurrentUserId(c)
		filter["student_id"] = userId
		// filter["is_completed"] = true
	}

	var items []models.StudyReport
	var rows int64

	if roleId == models.StudentRoleId {
		allItems, err := s.repo.GetAll()
		if err != nil {
			return nil, 0, err
		}
		items = allItems
		rows = int64(len(allItems))
	} else {
		var err error
		items, rows, err = s.repo.FindAll()
		if err != nil {
			return nil, 0, err
		}
	}

	courseIds := make(map[int64]struct{})
	assessmentIds := make(map[int64]struct{})
	for _, item := range items {
		courseIds[item.CourseId] = struct{}{}
		assessmentIds[item.AssessmentId] = struct{}{}
	}

	courseIdList := make([]int64, 0, len(courseIds))
	for id := range courseIds {
		courseIdList = append(courseIdList, id)
	}

	assessmentIdList := make([]int64, 0, len(assessmentIds))
	for id := range assessmentIds {
		assessmentIdList = append(assessmentIdList, id)
	}

	publishes, err := s.repo.FindPublishReports(courseIdList, assessmentIdList)
	if err != nil {
		return nil, 0, err
	}

	publishMap := make(map[string]bool, len(publishes))
	for _, p := range publishes {
		key := fmt.Sprintf("%d:%d", p.CourseId, p.AssessmentId)
		publishMap[key] = true
	}

	for i := range items {
		key := fmt.Sprintf("%d:%d", items[i].CourseId, items[i].AssessmentId)
		if publishMap[key] {
			items[i].Publish = true
		}
	}

	if roleId == models.StudentRoleId {
		filtered := make([]models.StudyReport, 0, len(items))
		studyReportResource := resources.NewStudyReportResource()
		for _, item := range items {
			if item.Publish {
				isCompleted := studyReportResource.IsComplete(&item)

				if isCompleted {
					filtered = append(filtered, item)
				}
			}
		}

		items = filtered
		rows = int64(len(filtered))

		start := (page - 1) * perPage
		end := start + perPage
		if start > len(items) {
			items = []models.StudyReport{}
		} else {
			if end > len(items) {
				end = len(items)
			}
			items = items[start:end]
		}
	}

	return items, rows, nil
}

func (s *studyReportService) GetByID(c *gin.Context, id int) (*prot.StudyReport, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
		"Student",
		"Teacher",
		"Subject",
		"Course",
		"Course.Program",
		"Student.Classes",
		"Assessment",
	})
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	item.Publish = s.repo.GetPublish(item.CourseId, item.AssessmentId)

	resource := resources.NewStudyReportResource()
	return resource.FormatStudyReport(item), nil
}

func (s *studyReportService) Create(c *gin.Context, req *prot.StudyReportRequest) (*models.StudyReport, error) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId == models.TeacherRoleId {
		userId := utils.GetCurrentUserId(c)
		req.TeacherId = int64(userId)
	}

	courseId, err := s.repo.ValidateStudyReportIDs(req.StudentId, req.AssessmentId, req.CourseId)

	if err != nil {
		return nil, err
	}

	req.CourseId = courseId

	if req.CourseId == 0 {
		return nil, fmt.Errorf("Student %d does not belong to the course of assessment %d", req.StudentId, req.AssessmentId)
	}

	existingReport, err := s.repo.FindByStudentCourseSubjectAssessment(
		req.StudentId, req.CourseId, req.SubjectId, req.AssessmentId)

	resource := resources.NewStudyReportResource()
	model := resource.FormatModelStudyReport(req)

	s.repo.SetContext(c)

	if err == nil && existingReport != nil {
		model.ID = existingReport.ID
		if err := s.repo.Update(model); err != nil {
			return nil, err
		}
	} else if err == gorm.ErrRecordNotFound {
		if err := s.repo.Create(model); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if err := s.repo.ReplaceSkillValues(model.ID, model.StudyReportCriteriaId, buildStudyReportSkillValues(req)); err != nil {
		return nil, err
	}

	// Tính và lưu total_star và avg_star vào DB
	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
	})
	tempReport, err := s.repo.FindNewByID(int(model.ID))
	if err == nil && tempReport != nil {
		_, _, _, totalStar, avgStar := resource.SplitReportSkills(tempReport)
		totalStarInt := int(totalStar)
		if err := s.repo.UpdateTotalStarAndAvgStar(model.ID, totalStarInt, avgStar); err != nil {
			// Log error nhưng không return để không làm gián đoạn flow
			config.Log.Errorf("Failed to update total_star and avg_star for study_report %d: %v", model.ID, err)
		}
	}

	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
		"Student",
		"Teacher",
		"Subject",
		"Course",
		"Course.Program",
		"Student.Classes",
		"Assessment",
	})

	newModel, _ := s.repo.FindNewByID(int(model.ID))
	newModel.Publish = s.repo.GetPublish(newModel.CourseId, newModel.AssessmentId)

	isComplete := newModel.IsCompleted
	isCompleteByTeacher := resource.IsComplete(newModel)

	if isComplete != isCompleteByTeacher {
		newModel.IsCompleted = isCompleteByTeacher
		s.repo.UpdateIsCompleted(model.ID, isCompleteByTeacher)
	}

	return newModel, nil
}

func (s *studyReportService) Update(c *gin.Context, req *prot.StudyReportRequest) (*models.StudyReport, error) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId == models.TeacherRoleId {
		userId := utils.GetCurrentUserId(c)
		req.TeacherId = int64(userId)
	}

	courseId, err := s.repo.ValidateStudyReportIDs(req.StudentId, req.AssessmentId, req.CourseId)

	if err != nil {
		return nil, err
	}

	req.CourseId = courseId

	if req.CourseId == 0 {
		return nil, fmt.Errorf("Student %d does not belong to the course of assessment %d", req.StudentId, req.AssessmentId)
	}

	resource := resources.NewStudyReportResource()
	model := resource.FormatModelStudyReport(req)

	s.repo.SetContext(c)

	if err := s.repo.Update(model); err != nil {
		return nil, err
	}

	if err := s.repo.ReplaceSkillValues(model.ID, model.StudyReportCriteriaId, buildStudyReportSkillValues(req)); err != nil {
		return nil, err
	}

	// Tính và lưu total_star và avg_star vào DB
	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
	})
	tempReport, err := s.repo.FindNewByID(int(model.ID))
	if err == nil && tempReport != nil {
		_, _, _, totalStar, avgStar := resource.SplitReportSkills(tempReport)
		totalStarInt := int(totalStar)
		if err := s.repo.UpdateTotalStarAndAvgStar(model.ID, totalStarInt, avgStar); err != nil {
			// Log error nhưng không return để không làm gián đoạn flow
			config.Log.Errorf("Failed to update total_star and avg_star for study_report %d: %v", model.ID, err)
		}
	}

	s.repo.SetPreload([]string{
		"Skills",
		"Criteria",
		"Criteria.Skills",
		"Criteria.Skills.Types",
		"Student",
		"Teacher",
		"Subject",
		"Course",
		"Course.Program",
		"Student.Classes",
		"Assessment",
	})
	updatedModel, _ := s.repo.FindNewByID(int(model.ID))
	updatedModel.Publish = s.repo.GetPublish(updatedModel.CourseId, updatedModel.AssessmentId)

	isComplete := updatedModel.IsCompleted
	isCompleteByTeacher := resource.IsComplete(updatedModel)

	if isComplete != isCompleteByTeacher {
		updatedModel.IsCompleted = isCompleteByTeacher
		s.repo.UpdateIsCompleted(model.ID, isCompleteByTeacher)
	}

	return updatedModel, nil
}

func (s *studyReportService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *studyReportService) Restore(c *gin.Context, id int) (*models.StudyReport, error) {
	s.repo.SetContext(c)
	return s.repo.Restore(id)
}

func buildStudyReportSkillValues(req *prot.StudyReportRequest) []repositories.StudyReportSkillValue {
	values := make([]repositories.StudyReportSkillValue, 0)

	for _, skill := range req.StarSkills {
		for _, tp := range skill.Types {
			flattenStarTypeInputs(&values, skill.Id, tp)
		}
	}

	for _, skill := range req.CheckSkills {
		for _, tp := range skill.Types {
			flattenCheckTypeInputs(&values, skill.Id, tp)
		}
	}

	return values
}

func flattenStarTypeInputs(
	dst *[]repositories.StudyReportSkillValue,
	skillID int64,
	tp *prot.StudyReportStarTypeInput,
) {
	value := repositories.StudyReportSkillValue{
		SkillID:     skillID,
		SkillTypeID: tp.Id,
	}

	val := tp.Star
	value.StarValue = &val

	*dst = append(*dst, value)

	for idx, child := range tp.NodeTypes {
		childSort := idx + 1
		_ = childSort
		flattenStarTypeInputs(dst, skillID, child)
	}
}

func flattenCheckTypeInputs(
	dst *[]repositories.StudyReportSkillValue,
	skillID int64,
	tp *prot.StudyReportCheckTypeInput,
) {
	value := repositories.StudyReportSkillValue{
		SkillID:     skillID,
		SkillTypeID: tp.Id,
	}

	val := tp.IsCheck
	value.CheckValue = &val

	*dst = append(*dst, value)

	for idx, child := range tp.NodeTypes {
		childSort := idx + 1
		_ = childSort
		flattenCheckTypeInputs(dst, skillID, child)
	}
}

func (s *studyReportService) Evaluate(c *gin.Context) (*prot.StudyReportEvaluateResponse, error) {
	courseId, err := strconv.Atoi(c.Param("course_id"))
	if err != nil || courseId <= 0 {
		return nil, fmt.Errorf("invalid course_id")
	}

	teacherId := utils.GetCurrentUserId(c)
	if teacherId <= 0 {
		return nil, fmt.Errorf("teacher_id is required")
	}

	assessmentIdStr := c.Query("assessment_id")
	assessmentId, err := strconv.ParseInt(assessmentIdStr, 10, 64)
	if err != nil || assessmentId <= 0 {
		return nil, fmt.Errorf("assessment_id is required")
	}

	isCompleteParam := c.Query("is_completed")

	data, err := s.repo.GetEvaluateData(int64(courseId), assessmentId)
	if err != nil {
		return nil, err
	}

	students := data.Students
	course := data.Course
	reports := data.Reports

	reportMap := make(map[int64]map[int64]*models.StudyReport)
	for i := range reports {
		studentId := reports[i].StudentId
		reportAssessmentId := reports[i].AssessmentId
		if reportMap[studentId] == nil {
			reportMap[studentId] = make(map[int64]*models.StudyReport)
		}
		reportMap[studentId][reportAssessmentId] = &reports[i]
	}

	items := make([]*prot.StudyReportEvaluateItem, 0)
	studyReportResource := resources.NewStudyReportResource()

	for _, student := range students {
		studentReports := reportMap[student.ID]

		item := &prot.StudyReportEvaluateItem{
			StudentId:   student.ID,
			StudentName: student.Name,
			CourseId:    int64(courseId),
			CourseName:  course.Name,
		}

		report, hasReport := studentReports[assessmentId]

		totalSkills := 0
		ratedSkills := 0

		if hasReport && report != nil {
			starSkills, checkSkills, _, totalStar, avgStar := studyReportResource.SplitReportSkills(report)

			if report.Criteria.ID > 0 {
				for _, skill := range report.Criteria.Skills {
					for _, tp := range skill.Types {
						totalSkills++
						for _, ref := range report.Skills {
							if ref.SkillId == tp.ID {
								ratedSkills++
							}
						}
					}
				}
			}

			item.HasGeneralComment = report.GeneralComment != ""
			item.GeneralComment = report.GeneralComment
			item.IsCompleted = studyReportResource.IsComplete(report)
			item.ReportName = report.Name
			item.ReportId = report.ID
			item.TotalStar = totalStar
			item.StarSkills = starSkills
			item.CheckSkills = checkSkills
			item.AvgStar = avgStar
			item.AssessmentId = report.AssessmentId
		} else {
			item.HasGeneralComment = false
		}

		item.CriteriaCount = int32(totalSkills)
		item.EvaluatedCriteria = int32(ratedSkills)

		if isCompleteParam != "" {
			if isCompleteParam == "true" || isCompleteParam == "1" {
				if item.IsCompleted {
					items = append(items, item)
				}
			} else {
				if !item.IsCompleted {
					items = append(items, item)
				}
			}
		} else {
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ReportId > items[j].ReportId
	})

	publish := s.repo.GetPublish(int64(courseId), assessmentId)

	return &prot.StudyReportEvaluateResponse{
		Items:      items,
		TotalCount: uint64(len(items)),
		Publish:    publish,
	}, nil
}

func (s *studyReportService) GetPublish(c *gin.Context) (*prot.GetPublishStudyReportResponse, error) {
	courseIdStr := c.Query("course_id")
	courseId, err := strconv.Atoi(courseIdStr)
	if err != nil || courseId <= 0 {
		return nil, fmt.Errorf("invalid course_id")
	}

	assessmentIdStr := c.Query("assessment_id")
	assessmentId, err := strconv.Atoi(assessmentIdStr)
	if err != nil || assessmentId <= 0 {
		return nil, fmt.Errorf("assessment_id is required")
	}

	publish := s.repo.GetPublish(int64(courseId), int64(assessmentId))

	return &prot.GetPublishStudyReportResponse{
		CourseId:     int64(courseId),
		AssessmentId: int64(assessmentId),
		Publish:      publish,
	}, nil
}
func (s *studyReportService) UpdatePublish(c *gin.Context) (*prot.UpdatePublishStudyReportResponse, error) {
	req, err, _ := utils.GetBody[*prot.UpdatePublishStudyReportRequest](c, func() *prot.UpdatePublishStudyReportRequest {
		return &prot.UpdatePublishStudyReportRequest{}
	})

	if err != nil {
		return nil, err
	}

	if req.AssessmentId == 0 || req.CourseId == 0 {
		return nil, fmt.Errorf("assessment_id and course_id are required")
	}

	s.repo.UpdatePublish(int64(req.CourseId), int64(req.AssessmentId), req.Publish)

	return &prot.UpdatePublishStudyReportResponse{
		CourseId:     req.CourseId,
		AssessmentId: req.AssessmentId,
		Publish:      req.Publish,
	}, nil
}
