package services

import (
	"fmt"

	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type AssessmentStudentService interface {
	GetStudentsWithAssessment(c *gin.Context, req *requests.AssessmentStudentRequest) (*prot.AssessmentStudentListResponse, error)
}

type assessmentStudentService struct {
	assessmentRepo repositories.AssessmentRepository
	studentRepo    repositories.AssessmentStudentRepository
	scoreRepo      repositories.AssessmentScoreRepository
	resource       resources.AssessmentResource
}

func NewAssessmentStudentService() AssessmentStudentService {
	return &assessmentStudentService{
		assessmentRepo: repositories.NewAssessmentRepository(),
		studentRepo:    repositories.NewAssessmentStudentRepository(),
		scoreRepo:      repositories.NewAssessmentScoreRepository(),
		resource:       resources.NewAssessmentResource(),
	}
}

func (s *assessmentStudentService) GetStudentsWithAssessment(c *gin.Context, req *requests.AssessmentStudentRequest) (*prot.AssessmentStudentListResponse, error) {
	// Validate assessment_id
	if req.AssessmentID <= 0 {
		return nil, fmt.Errorf("assessment_id is required and must be greater than 0")
	}

	// Validate course_id
	if req.CourseID <= 0 {
		return nil, fmt.Errorf("course_id is required and must be greater than 0")
	}

	// 1. Lấy assessment detail (chỉ lấy 1 lần)
	assessmentModel, err := s.assessmentRepo.GetByID(req.AssessmentID)
	if err != nil {
		return nil, err
	}

	// 2. Lấy danh sách học sinh trong course
	limit := req.Limit
	if limit <= 0 {
		limit = 0 // 0 = không giới hạn
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := 0
	if limit > 0 {
		offset = (page - 1) * limit
	}

	var studentID *int64
	if req.StudentID > 0 {
		studentID = &req.StudentID
	}
	students, total, err := s.studentRepo.GetStudentsByCourse(req.CourseID, studentID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 3. Build response
	items := make([]*prot.AssessmentStudentItem, 0, len(students))
	for _, sItem := range students {
		// Clone assessment cho mỗi student (để có thể set điểm riêng)
		assessmentProto := s.resource.FormatAssessment(assessmentModel)

		item := &prot.AssessmentStudentItem{
			StudentId:   sItem.StudentID,
			StudentName: sItem.StudentName,
			Assessment:  assessmentProto,
			IsScored:    false, // Mặc định là false
		}

		// Lấy assessment_score để check is_scored (luôn check, không cần get_score=true)
		score, err := s.scoreRepo.GetScoreByStudentAndAssessment(sItem.StudentID, req.AssessmentID, req.CourseID)
		if err == nil && score != nil {
			item.IsScored = true

			// Gán file_infos từ DB (luôn gán, không cần check get_score)
			for _, info := range score.FileInfos {
				item.FileInfos = append(item.FileInfos, &prot.AssessmentFileInfo{
					Id:   info.Id,
					Disk: info.Disk,
					Path: info.Path,
					Url:  utils.StaticURL(info.Path, info.Disk),
				})
			}

			// Nếu get_score=true, lấy thêm điểm và chi tiết điểm
			if req.GetScore {
				item.TotalScore = float32(score.TotalScore)

				// Convert file_infos thành URLs
				for _, info := range score.FileInfos {
					url := utils.StaticURL(info.Path, info.Disk)
					item.Files = append(item.Files, url)
				}

				// Lấy score details
				scoreDetails, err := s.scoreRepo.GetScoreDetailsByScoreID(score.ID)
				if err == nil {
					// Tạo map để tìm điểm nhanh: criterion_id -> score, subcriterion_id -> score
					criterionScoreMap := make(map[int64]float64)
					subcriterionScoreMap := make(map[int64]float64)

					for _, detail := range scoreDetails {
						if detail.SubcriterionID != nil && *detail.SubcriterionID > 0 {
							subcriterionScoreMap[*detail.SubcriterionID] = detail.Score
						} else {
							criterionScoreMap[detail.CriterionID] = detail.Score
						}
					}

					// Map điểm vào criteria và subcriteria
					if item.Assessment != nil && len(item.Assessment.Criteria) > 0 {
						for _, criterion := range item.Assessment.Criteria {
							// Set điểm cho criterion: nếu chưa có trong DB thì = -1
							if score, ok := criterionScoreMap[criterion.Id]; ok {
								criterion.Score = float32(score)
							} else {
								criterion.Score = -1
							}

							// Set điểm cho subcriteria
							if len(criterion.Subcriteria) > 0 {
								for _, subcriterion := range criterion.Subcriteria {
									// Nếu chưa có trong DB thì = -1
									if score, ok := subcriterionScoreMap[subcriterion.Id]; ok {
										subcriterion.Score = float32(score)
									} else {
										subcriterion.Score = -1
									}
								}
							}
						}
					}
				}
			}
		} else if req.GetScore {
			// Học sinh chưa có điểm, set total_score = -1 và tất cả điểm tiêu chí = -1
			item.TotalScore = -1

			// Set điểm -1 cho tất cả criteria và subcriteria
			if item.Assessment != nil && len(item.Assessment.Criteria) > 0 {
				for _, criterion := range item.Assessment.Criteria {
					criterion.Score = -1

					if len(criterion.Subcriteria) > 0 {
						for _, subcriterion := range criterion.Subcriteria {
							subcriterion.Score = -1
						}
					}
				}
			}
		}

		items = append(items, item)
	}

	// Lấy course_info
	courseName, objectTitle, teacherNames, err := s.studentRepo.GetCourseInfo(req.CourseID)
	if err != nil {
		return nil, err
	}

	return &prot.AssessmentStudentListResponse{
		Students: items,
		Total:    total,
		CourseInfo: &prot.AssessmentCourseInfo{
			Name:        courseName,
			ObjectTitle: objectTitle,
			TeacherNames: teacherNames,
		},
	}, nil
}
