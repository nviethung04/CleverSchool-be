package services

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"strconv"
)

type DashboardStudentAssessmentService interface {
	GetStudentAssessmentList(userID int64, courseID int64, subjectID *int64, assessmentType string, limit, page int) (*prot.AssessmentsResponse, error)
	GetStudentAssessments(userID int64, courseID int64, subjectID *int64, assessmentType string, assessmentID *int64, limit, page int) (*dto.DashboardStudentAssessmentListResponseDTO, error)
}

type dashboardStudentAssessmentService struct {
	repo        repositories.DashboardStudentAssessmentRepository
	scoringRepo repositories.AssessmentScoringRepository
	scoringSvc  AssessmentScoringService
}

func NewDashboardStudentAssessmentService(
	repo repositories.DashboardStudentAssessmentRepository,
	scoringRepo repositories.AssessmentScoringRepository,
	scoringSvc AssessmentScoringService,
) DashboardStudentAssessmentService {
	return &dashboardStudentAssessmentService{
		repo: repo, scoringRepo: scoringRepo, scoringSvc: scoringSvc,
	}
}

func (s *dashboardStudentAssessmentService) GetStudentAssessmentList(
	userID int64,
	courseID int64,
	subjectID *int64,
	assessmentType string,
	limit, page int,
) (*prot.AssessmentsResponse, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}

	assessments, total, err := s.repo.ListAssessmentsForStudent(userID, courseID, subjectID, assessmentType, nil, limit, offset)
	if err != nil {
		return nil, err
	}

	assessmentResource := resources.NewAssessmentResource()
	ptrs := make([]*models.Assessment, 0, len(assessments))
	for i := range assessments {
		ptrs = append(ptrs, &assessments[i])
	}

	return &prot.AssessmentsResponse{
		Assessments: assessmentResource.FormatAssessments(ptrs),
		Total:       total,
	}, nil
}

func (s *dashboardStudentAssessmentService) GetStudentAssessments(
	userID int64,
	courseID int64,
	subjectID *int64,
	assessmentType string,
	assessmentID *int64,
	limit, page int,
) (*dto.DashboardStudentAssessmentListResponseDTO, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}

	assessments, total, err := s.repo.ListAssessmentsForStudent(userID, courseID, subjectID, assessmentType, assessmentID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]dto.DashboardStudentAssessmentItemDTO, 0, len(assessments))
	for i := range assessments {
		published, _ := s.scoringRepo.GetAssessmentPublish(courseID, assessments[i].ID)
		items = append(items, mapStudentAssessmentItem(s.scoringSvc, &assessments[i], userID, courseID, published))
	}

	return &dto.DashboardStudentAssessmentListResponseDTO{
		Type:        "type.googleapis.com/prot.DashboardStudentAssessmentListResponse",
		Assessments: items,
		Total:       strconv.FormatInt(total, 10),
	}, nil
}

func mapStudentAssessmentItem(svc AssessmentScoringService, assessment *models.Assessment, userID, courseID int64, published bool) dto.DashboardStudentAssessmentItemDTO {
	raw := svc.BuildStudentAssessmentItem(assessment, userID, courseID, published)
	assessmentMap, _ := raw["assessment"].(map[string]interface{})
	item := dto.DashboardStudentAssessmentItemDTO{
		IsScored:   raw["is_scored"].(bool),
		TotalScore: raw["total_score"].(float64),
	}
	if files, ok := raw["files"].([]string); ok {
		item.Files = files
	}
	if assessmentMap != nil {
		item.Assessment = mapToAssessmentDetail(assessmentMap)
	}
	return item
}

func mapToAssessmentDetail(m map[string]interface{}) dto.DashboardStudentAssessmentDetailDTO {
	detail := dto.DashboardStudentAssessmentDetailDTO{
		ID:                          strVal(m["id"]),
		Name:                        strVal(m["name"]),
		Description:                 strVal(m["description"]),
		Type:                        strVal(m["type"]),
		ProgramID:                   strVal(m["program_id"]),
		CreatedAt:                   strVal(m["created_at"]),
		CreatedBy:                   strVal(m["created_by"]),
		UpdatedAt:                   strVal(m["updated_at"]),
		UpdatedBy:                   strVal(m["updated_by"]),
		StudyReportCriteriaID:       strVal(m["study_report_criteria_id"]),
		SubjectID:                   strVal(m["subject_id"]),
		AssessmentCriteriaGroupID:   strVal(m["assessment_criteria_group_id"]),
		AssessmentCriteriaGroupName: strVal(m["assessment_criteria_group_name"]),
		StudyReportCriteriaName:     strVal(m["study_report_criteria_name"]),
		SubjectName:                 strVal(m["subject_name"]),
		IsAssigned:                  boolVal(m["is_assigned"]),
		HasFile:                     boolVal(m["has_file"]),
	}
	if files, ok := m["files"].([]string); ok {
		detail.Files = files
	}
	if criteria, ok := m["criteria"].([]map[string]interface{}); ok {
		detail.Criteria = mapCriteriaList(criteria)
	}
	return detail
}

func mapCriteriaList(list []map[string]interface{}) []dto.DashboardStudentAssessmentCriterionDTO {
	out := make([]dto.DashboardStudentAssessmentCriterionDTO, 0, len(list))
	for _, c := range list {
		crit := dto.DashboardStudentAssessmentCriterionDTO{
			ID: strVal(c["id"]), SubjectID: strVal(c["subject_id"]), Name: strVal(c["name"]),
			Description: strVal(c["description"]), HasSubcriteria: boolVal(c["has_subcriteria"]),
			MaxScore: floatVal(c["max_score"]), CreatedAt: strVal(c["created_at"]),
			CreatedBy: strVal(c["created_by"]), UpdatedAt: strVal(c["updated_at"]),
			UpdatedBy: strVal(c["updated_by"]), Score: floatVal(c["score"]),
		}
		if subs, ok := c["subcriteria"].([]map[string]interface{}); ok {
			for _, sub := range subs {
				crit.Subcriteria = append(crit.Subcriteria, dto.DashboardStudentAssessmentSubcriterionDTO{
					ID: strVal(sub["id"]), Name: strVal(sub["name"]), Description: strVal(sub["description"]),
					MaxScore: floatVal(sub["max_score"]), Score: floatVal(sub["score"]),
				})
			}
		}
		out = append(out, crit)
	}
	return out
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func floatVal(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func boolVal(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
