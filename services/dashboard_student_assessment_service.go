package services

import (
	"fmt"

	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardStudentAssessmentService interface {
	GetStudentAssessments(c *gin.Context, req *requests.DashboardStudentAssessmentRequest) (*prot.DashboardStudentAssessmentListResponse, error)
}

type dashboardStudentAssessmentService struct {
	assessmentRepo repositories.AssessmentRepository
	refRepo        repositories.DashboardStudentAssessmentRepository
	scoreRepo      repositories.AssessmentScoreRepository
	resource       resources.AssessmentResource
}

func NewDashboardStudentAssessmentService() DashboardStudentAssessmentService {
	return &dashboardStudentAssessmentService{
		assessmentRepo: repositories.NewAssessmentRepository(),
		refRepo:        repositories.NewDashboardStudentAssessmentRepository(),
		scoreRepo:      repositories.NewAssessmentScoreRepository(),
		resource:       resources.NewAssessmentResource(),
	}
}

func (s *dashboardStudentAssessmentService) GetStudentAssessments(c *gin.Context, req *requests.DashboardStudentAssessmentRequest) (*prot.DashboardStudentAssessmentListResponse, error) {
	// Validate course_id
	if req.CourseID <= 0 {
		return nil, fmt.Errorf("course_id is required and must be greater than 0")
	}

	// Validate assessment_id: nếu không truyền assessment_id thì trả về rỗng
	if req.AssessmentID <= 0 {
		return &prot.DashboardStudentAssessmentListResponse{
			Assessments: []*prot.DashboardStudentAssessmentItem{},
			Total:       0,
		}, nil
	}

	// Lấy student_id từ context
	studentID := int64(utils.GetCurrentUserId(c))
	if studentID <= 0 {
		return nil, fmt.Errorf("student_id is required")
	}

	// 1. Lấy danh sách assessment_ref_lessons có assigned_at != null
	assessmentID := &req.AssessmentID
	var assessmentType *string
	if req.Type != "" {
		assessmentType = &req.Type
	}
	refs, err := s.refRepo.GetAssessmentsByCourseID(req.CourseID, assessmentID, assessmentType)
	if err != nil {
		return nil, fmt.Errorf("error getting assessments: %w", err)
	}

	assessmentRepo := repositories.NewAssessmentRepository()
	publish := assessmentRepo.GetPublish(req.CourseID, req.AssessmentID)

	// 2. Lấy assessment detail cho mỗi assessment và check score
	items := make([]*prot.DashboardStudentAssessmentItem, 0, len(refs))
	for _, ref := range refs {
		// Skip assessment not publish
		if !publish {
			continue
		}
		// Lấy assessment detail
		assessmentModel, err := s.assessmentRepo.GetByID(ref.AssessmentId)
		if err != nil {
			// Nếu không tìm thấy assessment, bỏ qua
			continue
		}

		// Clone assessment để có thể set điểm riêng
		assessmentProto := s.resource.FormatAssessment(assessmentModel)

		item := &prot.DashboardStudentAssessmentItem{
			Assessment: assessmentProto,
			IsScored:   false, // Mặc định là false
		}

		// Check xem student đã có score chưa
		score, err := s.scoreRepo.GetScoreByStudentAndAssessment(studentID, ref.AssessmentId, req.CourseID)
		if err == nil && score != nil {
			item.IsScored = true
			item.TotalScore = float32(score.TotalScore)

			// Convert file_infos thành URLs
			for _, info := range score.FileInfos {
				url := utils.StaticURL(info.Path, info.Disk)
				item.Files = append(item.Files, url)
			}

			// Lấy score details để map điểm vào criteria/subcriteria
			scoreDetails, err := s.scoreRepo.GetScoreDetailsByScoreID(score.ID)
			if err == nil {
				// Tạo map để lookup nhanh: (criterion_id, subcriterion_id) -> score
				scoreMap := make(map[int64]map[int64]float64) // criterion_id -> (subcriterion_id -> score)
				criterionScoreMap := make(map[int64]float64)  // criterion_id -> score (khi không có subcriteria)

				for _, detail := range scoreDetails {
					if detail.SubcriterionID != nil {
						// Có subcriteria
						if scoreMap[detail.CriterionID] == nil {
							scoreMap[detail.CriterionID] = make(map[int64]float64)
						}
						scoreMap[detail.CriterionID][*detail.SubcriterionID] = detail.Score
					} else {
						// Không có subcriteria, dùng điểm của criterion
						criterionScoreMap[detail.CriterionID] = detail.Score
					}
				}

				// Map điểm vào criteria và subcriteria trong assessment
				for _, criterion := range item.Assessment.Criteria {
					// Check xem có subcriteria không
					if len(criterion.Subcriteria) > 0 {
						// Có subcriteria, map điểm vào từng subcriterion
						subcriterionScoreMap := scoreMap[criterion.Id]
						if subcriterionScoreMap != nil {
							for _, subcriterion := range criterion.Subcriteria {
								if score, ok := subcriterionScoreMap[subcriterion.Id]; ok {
									subcriterion.Score = float32(score)
								}
							}
							// Tính tổng điểm của criterion từ subcriteria
							var criterionTotal float64
							for _, subcriterion := range criterion.Subcriteria {
								criterionTotal += float64(subcriterion.Score)
							}
							criterion.Score = float32(criterionTotal)
						}
					} else {
						// Không có subcriteria, dùng điểm của criterion
						if score, ok := criterionScoreMap[criterion.Id]; ok {
							criterion.Score = float32(score)
						}
					}
				}
			}
		}

		items = append(items, item)
	}

	return &prot.DashboardStudentAssessmentListResponse{
		Assessments: items,
		Total:       int64(len(items)),
	}, nil
}
