package services

import (
	"fmt"
	"strings"

	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type DashboardAssessmentReportService interface {
	GetAssessmentReport(c *gin.Context, req *requests.AssessmentReportRequest) (*prot.AssessmentReportResponse, error)
}

type dashboardAssessmentReportService struct {
	assessmentRepo          repositories.AssessmentRepository
	reportRepo              repositories.DashboardAssessmentReportRepository
	scoreRepo               repositories.AssessmentScoreRepository
	studyReportCriteriaRepo repositories.StudyReportCriteriaRepository
	resource                resources.AssessmentResource
	studyReportCriteriaRes  resources.StudyReportCriteriaResource
}

func NewDashboardAssessmentReportService() DashboardAssessmentReportService {
	return &dashboardAssessmentReportService{
		assessmentRepo:          repositories.NewAssessmentRepository(),
		reportRepo:              repositories.NewDashboardAssessmentReportRepository(),
		scoreRepo:               repositories.NewAssessmentScoreRepository(),
		studyReportCriteriaRepo: repositories.NewStudyReportCriteriaRepository(),
		resource:                resources.NewAssessmentResource(),
		studyReportCriteriaRes:  resources.NewStudyReportCriteriaResource(),
	}
}

func (s *dashboardAssessmentReportService) GetAssessmentReport(c *gin.Context, req *requests.AssessmentReportRequest) (*prot.AssessmentReportResponse, error) {
	// Validate assessment_id
	if req.AssessmentID <= 0 {
		return nil, fmt.Errorf("assessment_id is required and must be greater than 0")
	}

	// Validate class_id hoặc class_main_id hoặc student_id
	// Nếu không truyền class_id và class_main_id thì vẫn cho phép,
	// miễn là có ít nhất 1 student_id (trường hợp gọi trực tiếp theo danh sách học sinh)
	if req.ClassMainID <= 0 && req.ClassID <= 0 && len(req.StudentID) == 0 {
		return nil, fmt.Errorf("class_id, class_main_id or student_id is required")
	}

	// 1. Lấy assessment detail (chỉ lấy 1 lần)
	assessmentModel, err := s.assessmentRepo.GetByID(req.AssessmentID)
	if err != nil {
		return nil, err
	}

	// 2. Lấy danh sách học sinh trong class
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
	
	// Nếu có class_main_id thì dùng class_main_id, ngược lại dùng class_id
	// Trường hợp không có cả 2 nhưng có student_id thì classID = 0, classMainID = 0
	var classID int64
	if req.ClassMainID > 0 {
		classID = 0 // Bỏ qua class_id khi dùng class_main_id
	} else {
		classID = req.ClassID
	}
	
	students, total, err := s.reportRepo.GetStudentsByClass(classID, req.ClassMainID, req.StudentID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 3. Lấy toàn bộ scores cho danh sách students (để tránh query từng student)
	studentIDs := make([]int64, 0, len(students))
	for _, sItem := range students {
		studentIDs = append(studentIDs, sItem.StudentID)
	}

	scoreList, err := s.scoreRepo.GetScoresByAssessmentAndStudents(req.AssessmentID, studentIDs)
	if err != nil {
		return nil, err
	}

	// Map: student_id -> score mới nhất
	scoreByStudent := make(map[int64]*models.AssessmentScore)
	for i := range scoreList {
		score := &scoreList[i]
		if _, ok := scoreByStudent[score.StudentID]; !ok {
			scoreByStudent[score.StudentID] = score
		}
	}

	// 4. Build response + thống kê
	items := make([]*prot.AssessmentStudentItem, 0, len(students))
	statistic := &prot.AssessmentReportStatistic{}

	for _, sItem := range students {
		// Clone assessment cho mỗi student (để có thể set điểm riêng)
		assessmentProto := s.resource.FormatAssessment(assessmentModel)

		item := &prot.AssessmentStudentItem{
			StudentId:   sItem.StudentID,
			StudentName: sItem.StudentName,
			Assessment:  assessmentProto,
			IsScored:    false, // Mặc định là false
		}

		// Lấy assessment_score đã preload ở trên (không query từng student)
		score, ok := scoreByStudent[sItem.StudentID]
		if ok && score != nil {
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
								var sumPositive float32
								var hasPositive bool

								for _, subcriterion := range criterion.Subcriteria {
									// Nếu chưa có trong DB thì = -1
									if score, ok := subcriterionScoreMap[subcriterion.Id]; ok {
										subcriterion.Score = float32(score)
									} else {
										subcriterion.Score = -1
									}

									// Cộng các score > 0 để tính score cho tiêu chí cha
									if subcriterion.Score > 0 {
										sumPositive += subcriterion.Score
										hasPositive = true
									}
								}

								// Nếu có ít nhất 1 subcriterion có score > 0
								// thì score của tiêu chí cha = tổng các score dương
								if hasPositive {
									criterion.Score = sumPositive
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

		// Tính thống kê theo total_score (chỉ khi get_score=true)
		if req.GetScore {
			if !item.IsScored || item.TotalScore < 0 {
				statistic.Unscored++
			} else {
				ts := item.TotalScore
				switch {
				case ts < 50:
					statistic.From_0To_50++
				case ts >= 50 && ts < 70:
					statistic.From_50To_70++
				case ts >= 70 && ts < 80:
					statistic.From_70To_80++
				case ts >= 80:
					statistic.From_80To_100++
				}
			}
		}

		// Lấy study report criteria + giá trị star/check cho HS này
		// nếu get_study_report_criteria = true và assessment có study_report_criteria_id
		if req.GetStudyReportCriteria && assessmentModel.StudyReportCriteriaID > 0 {
			// 1) Lấy tiêu chí gốc (structure) từ study_report_criterias
			s.studyReportCriteriaRepo.SetContext(c)
			s.studyReportCriteriaRepo.SetPreload([]string{
				"Subject",
				"Skills",
				"Skills.Types",
			})

			criteriaModel, err := s.studyReportCriteriaRepo.FindByID(int(assessmentModel.StudyReportCriteriaID))
			if err == nil && criteriaModel != nil {
				criteriaProto := s.studyReportCriteriaRes.FormatStudyReportCriteria(criteriaModel)

				// 2) Lấy study_report_id tương ứng với (student_id, course_id, subject_id, assessment_id)
				//    để truy ra các giá trị star/is_check trong study_report_ref_skill_types
				var courseID int64
				err := db.ReplicaDB.Table("assessment_ref_lessons arl").
					Select("arl.course_id").
					Joins("JOIN user_courses uc ON uc.course_id = arl.course_id").
					Where("arl.assessment_id = ? AND uc.user_id = ? AND arl.assigned_at IS NOT NULL",
						req.AssessmentID, sItem.StudentID).
					Limit(1).
					Scan(&courseID).Error

				if err == nil && courseID > 0 {
					var studyReportID int64
					err = db.ReplicaDB.Table("study_reports").
						Select("id").
						Where("student_id = ? AND course_id = ? AND subject_id = ? AND assessment_id = ?",
							sItem.StudentID, courseID, assessmentModel.SubjectID, req.AssessmentID).
						Limit(1).
						Scan(&studyReportID).Error

					if err == nil && studyReportID > 0 {
						// 3) Lấy toàn bộ ref (star/is_check) cho report này
						type refRow struct {
							SkillID     int64 `gorm:"column:skill_id"`
							SkillTypeID int64 `gorm:"column:skill_type_id"`
							Star        *int  `gorm:"column:star"`
							IsCheck     bool  `gorm:"column:is_check"`
						}

						var refs []refRow
						if err := db.ReplicaDB.Table("study_report_ref_skill_types").
							Select("skill_id, skill_type_id, star, is_check").
							Where("study_report_id = ?", studyReportID).
							Find(&refs).Error; err == nil && len(refs) > 0 {

							starMap := make(map[int64]int32)
							checkMap := make(map[int64]bool)
							for _, r := range refs {
								if r.Star != nil {
									starMap[r.SkillTypeID] = int32(*r.Star)
								}
								// check dùng riêng map, vì 1 type có thể là check
								if r.IsCheck {
									checkMap[r.SkillTypeID] = true
								}
							}

							// Hàm đệ quy apply star/check vào toàn bộ cây type
							var applyValues func(t *prot.StudyReportSkillType)
							applyValues = func(t *prot.StudyReportSkillType) {
								if v, ok := starMap[t.Id]; ok {
									t.Star = v
								}
								if v, ok := checkMap[t.Id]; ok {
									t.IsCheck = v
								}
								for _, child := range t.NodeTypes {
									applyValues(child)
								}
							}

							// Áp star/check cho tất cả star_skills và check_skills
							for _, sSkill := range criteriaProto.StarSkills {
								for _, t := range sSkill.Types {
									applyValues(t)
								}
							}
							for _, cSkill := range criteriaProto.CheckSkills {
								for _, t := range cSkill.Types {
									applyValues(t)
								}
							}
						}
					}
				}

				item.StudyReportCriteria = criteriaProto
			}
		}

		// Lấy general_comment từ study_reports cho học sinh này
		// Lấy course_id từ assessment_ref_lessons và user_courses
		var courseIDForComment int64
		err := db.ReplicaDB.Table("assessment_ref_lessons arl").
			Select("arl.course_id").
			Joins("JOIN user_courses uc ON uc.course_id = arl.course_id").
			Where("arl.assessment_id = ? AND uc.user_id = ? AND arl.assigned_at IS NOT NULL",
				req.AssessmentID, sItem.StudentID).
			Limit(1).
			Scan(&courseIDForComment).Error

		if err == nil && courseIDForComment > 0 {
			var generalComment string
			err = db.ReplicaDB.Table("study_reports").
				Select("general_comment").
				Where("student_id = ? AND course_id = ? AND subject_id = ? AND assessment_id = ?",
					sItem.StudentID, courseIDForComment, assessmentModel.SubjectID, req.AssessmentID).
				Limit(1).
				Scan(&generalComment).Error

			if err == nil {
				item.GeneralComment = generalComment
			}
		}

		items = append(items, item)
	}

	// 5. Lấy class_name, school_name và school_logo cho report (API mới)
	var className, schoolName, schoolLogo string
	if req.ClassMainID > 0 {
		var logoInfo models.MediaInfo
		className, schoolName, logoInfo, err = s.reportRepo.GetClassMainInfoForReport(req.ClassMainID)
		if err != nil {
			return nil, err
		}
		// Format URL logo nếu có
		if logoInfo.Path != "" {
			schoolLogo = utils.StaticURL(logoInfo.Path, logoInfo.Disk)
		}
	} else if req.ClassID > 0 {
		var logoInfo models.MediaInfo
		className, schoolName, logoInfo, err = s.reportRepo.GetClassInfoForReport(req.ClassID)
		if err != nil {
			return nil, err
		}
		// Format URL logo nếu có
		if logoInfo.Path != "" {
			schoolLogo = utils.StaticURL(logoInfo.Path, logoInfo.Disk)
		}
	}

	// 6. Lấy danh sách giáo viên VN và FR (lọc theo subject_id từ assessment)
	var vnTeacherNames, frTeacherNames []string
	if req.ClassMainID > 0 {
		vnTeacherNames, frTeacherNames, err = s.reportRepo.GetTeachersByClassMain(req.ClassMainID, assessmentModel.SubjectID)
		if err != nil {
			// Nếu có lỗi, để trống
			vnTeacherNames = []string{}
			frTeacherNames = []string{}
		}
	} else if req.ClassID > 0 {
		vnTeacherNames, frTeacherNames, err = s.reportRepo.GetTeachersByClass(req.ClassID, assessmentModel.SubjectID)
		if err != nil {
			// Nếu có lỗi, để trống
			vnTeacherNames = []string{}
			frTeacherNames = []string{}
		}
	}

	// Format danh sách giáo viên thành string (nối bằng dấu phẩy)
	vnTeacher := ""
	if len(vnTeacherNames) > 0 {
		vnTeacher = strings.Join(vnTeacherNames, ", ")
	}
	frTeacher := ""
	if len(frTeacherNames) > 0 {
		frTeacher = strings.Join(frTeacherNames, ", ")
	}

	// 7. Lấy program_name từ assessment và class_main_id (intersection)
	var programName string
	var classMainID int64
	if req.ClassMainID > 0 {
		classMainID = req.ClassMainID
	} else if req.ClassID > 0 {
		// Nếu không có class_main_id, lấy từ class_id
		var classMainIDFromClass int64
		err = s.reportRepo.GetClassMainIDByClassID(req.ClassID, &classMainIDFromClass)
		if err == nil && classMainIDFromClass > 0 {
			classMainID = classMainIDFromClass
		}
	}
	
	if classMainID > 0 {
		programName, err = s.reportRepo.GetProgramNamesByAssessmentAndClassMain(req.AssessmentID, classMainID)
		if err != nil {
			// Nếu có lỗi, để trống
			programName = ""
		}
	}

	return &prot.AssessmentReportResponse{
		Students: items,
		Total:    total,
		Info: &prot.AssessmentReportInfo{
			SchoolName:  schoolName,
			ClassName:   className,
			ProgramName: programName,
			SchoolLogo:  schoolLogo,
			VnTeacher:   vnTeacher,
			FrTeacher:   frTeacher,
		},
		Statistic: statistic,
	}, nil
}


