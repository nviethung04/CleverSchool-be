package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/repositories/base"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentService interface {
	GetAll(c *gin.Context) ([]models.Assessment, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Assessment, error)
	Create(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error)
	Update(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Assessment, error)
	AssignRefLesson(c *gin.Context, req *requests.AssessmentRefLessonRequest) error
	CreateWithCriteria(c *gin.Context, req *prot.CreateAssessmentWithCriteriaRequest) (*prot.Assessment, error)
	SaveScoreBulk(c *gin.Context, req *prot.SaveAssessmentScoreBulkRequest) error

	GetPublish(c *gin.Context) (*prot.GetPublishAssessmentResponse, error)
	UpdatePublish(c *gin.Context) (*prot.UpdatePublishAssessmentResponse, error)
}

type assessmentService struct {
	repo             repositories.AssessmentRepository
	criteriaRepo     repositories.AssessmentCriterionRepository
	subcriterionRepo repositories.AssessmentSubcriterionRepository
	scoreRepo        repositories.AssessmentScoreRepository
	groupRepo        repositories.AssessmentCriteriaGroupRepository
	resource         resources.AssessmentResource
}

func NewAssessmentService(repo repositories.AssessmentRepository) AssessmentService {
	return &assessmentService{
		repo:             repo,
		criteriaRepo:     repositories.NewAssessmentCriterionRepository(),
		subcriterionRepo: repositories.NewAssessmentSubcriterionRepository(),
		scoreRepo:        repositories.NewAssessmentScoreRepository(),
		groupRepo:        repositories.NewAssessmentCriteriaGroupRepository(),
		resource:         resources.NewAssessmentResource(),
	}
}

func (s *assessmentService) GetAll(c *gin.Context) ([]models.Assessment, int64, error) {
	var req requests.GetAssessmentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	assessments, total, err := s.repo.GetAllWithPaging(&req, c)

	assessmentIds := make(map[int64]struct{})
	for _, item := range assessments {
		assessmentIds[item.ID] = struct{}{}
	}

	assessmentIdList := make([]int64, 0, len(assessmentIds))
	for id := range assessmentIds {
		assessmentIdList = append(assessmentIdList, id)
	}

	publishes, err := s.repo.FindPublishAssessments(assessmentIdList)

	if err != nil {
		return nil, 0, err
	}

	for i := range assessments {
		var publishCourseIds []int64
		for _, p := range publishes {
			if p.AssessmentId == assessments[i].ID {
				publishCourseIds = append(publishCourseIds, p.CourseId)
			}
		}

		assessments[i].PublishCourseIds = publishCourseIds
	}

	return assessments, total, err
}

func (s *assessmentService) GetByID(c *gin.Context, id int) (*prot.Assessment, error) {
	item, err := s.repo.GetByID(int64(id))
	if err != nil {
		return nil, err
	}

	item.PublishCourseIds = s.repo.GetPublishCourseIds(item.ID)

	return s.resource.FormatAssessment(item), nil
}

func (s *assessmentService) Create(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error) {
	now := time.Now().UTC()
	userID := int64(utils.GetCurrentUserId(c))

	// Nếu truyền copy_from_assessment_id (hoặc assessment_id trong JSON): tạo assessment mới bằng cách copy từ assessment đó
	copyFromID := req.GetCopyFromAssessmentId()
	if copyFromID > 0 {
		return s.createByCopyFromAssessment(c, copyFromID, now, userID)
	}

	entity := s.resource.FormatModelAssessment(req)
	if entity == nil {
		return nil, nil
	}

	entity.CreatedBy = userID
	entity.CreatedAt = now
	entity.UpdatedAt = now

	if err := s.repo.Create(entity); err != nil {
		return nil, err
	}

	// Xử lý gán criteria
	if err := s.syncAssessmentCriteria(entity.ID, req, now, userID); err != nil {
		return nil, err
	}

	entity.PublishCourseIds = s.repo.GetPublishCourseIds(entity.ID)

	return entity, nil
}

// createByCopyFromAssessment tạo assessment mới bằng cách copy từ assessment có id = sourceID (bảng assessments + assessment_ref_criteria).
func (s *assessmentService) createByCopyFromAssessment(c *gin.Context, sourceID int64, now time.Time, userID int64) (*models.Assessment, error) {
	source, err := s.repo.GetByID(sourceID)
	if err != nil || source == nil {
		return nil, fmt.Errorf("không tìm thấy assessment id %d để copy: %w", sourceID, err)
	}

	// 1. Tạo bản ghi mới trong bảng assessments (copy thông tin, không copy id/timestamps/audit)
	entity := &models.Assessment{
		Name:                  source.Name,
		Description:           source.Description,
		Type:                  source.Type,
		ProgramID:             source.ProgramID,
		SubjectID:             source.SubjectID,
		StudyReportCriteriaID: source.StudyReportCriteriaID,
		FileInfos:             source.FileInfos,
		HasFile:               source.HasFile,
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             userID,
		UpdatedBy:             userID,
	}
	if err := s.repo.Create(entity); err != nil {
		return nil, err
	}

	// 2. Sao chép assessment_ref_criteria từ assessment nguồn sang assessment mới
	criteriaIDs, err := s.criteriaRepo.GetRefCriteriaIDsByAssessmentID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("lấy danh sách criteria từ assessment nguồn: %w", err)
	}
	for _, criterionID := range criteriaIDs {
		if criterionID <= 0 {
			continue
		}
		ref := &models.AssessmentRefCriterion{
			AssessmentID:          entity.ID,
			AssessmentCriterionID: criterionID,
			CreatedAt:             now,
			UpdatedAt:             now,
			CreatedBy:             userID,
			UpdatedBy:             userID,
		}
		if err := s.criteriaRepo.CreateRefCriterion(ref); err != nil {
			return nil, fmt.Errorf("tạo assessment_ref_criterion: %w", err)
		}
	}

	entity.PublishCourseIds = s.repo.GetPublishCourseIds(entity.ID)
	return entity, nil
}

func (s *assessmentService) Update(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error) {
	entity := s.resource.FormatModelAssessment(req)
	if entity == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	userID := int64(utils.GetCurrentUserId(c))
	entity.UpdatedBy = userID
	entity.UpdatedAt = now

	if err := s.repo.Update(entity); err != nil {
		return nil, err
	}

	// Xử lý gán criteria với tracking thay đổi
	if err := s.syncAssessmentCriteria(entity.ID, req, now, userID); err != nil {
		return nil, err
	}

	entity.PublishCourseIds = s.repo.GetPublishCourseIds(entity.ID)

	return entity, nil
}

func (s *assessmentService) Delete(c *gin.Context, id int) error {
	return s.repo.Delete(int64(id), int64(utils.GetCurrentUserId(c)))
}

func (s *assessmentService) Restore(c *gin.Context, id int) (*models.Assessment, error) {
	baseRepo := base.NewBaseRepository[*models.Assessment]()
	baseRepo.SetContext(c)
	item, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *item, nil
}

func (s *assessmentService) AssignRefLesson(c *gin.Context, req *requests.AssessmentRefLessonRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if req.AssessmentID <= 0 || req.LessonID <= 0 || req.CourseID <= 0 {
		return fmt.Errorf("assessment_id, lesson_id, course_id phải lớn hơn 0")
	}

	now := time.Now().UTC()
	assignedBy := int64(utils.GetCurrentUserId(c))

	ref := models.AssessmentRefLesson{
		AssessmentId: req.AssessmentID,
		LessonId:     req.LessonID,
		CourseId:     req.CourseID,
		AssignedAt:   &now,
		AssignedBy:   utils.PtrInt64(assignedBy),
	}

	return s.repo.AssignRefLesson(ref)
}

func (s *assessmentService) CreateWithCriteria(c *gin.Context, req *prot.CreateAssessmentWithCriteriaRequest) (*prot.Assessment, error) {
	if req == nil || req.Assessment == nil {
		return nil, fmt.Errorf("assessment data is required")
	}
	if req.LessonId <= 0 && req.Assessment.ProgramId == 0 {
		return nil, fmt.Errorf("lesson_id hoặc program_id là bắt buộc")
	}

	subjectID, programID, err := s.resolveSubjectAndProgram(c, req)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	userID := int64(utils.GetCurrentUserId(c))

	assessment := s.resource.FormatModelAssessment(req.Assessment)
	if assessment == nil {
		return nil, fmt.Errorf("invalid assessment data")
	}

	assessment.CreatedBy = userID
	assessment.CreatedAt = now
	assessment.UpdatedBy = userID
	assessment.UpdatedAt = now
	assessment.ProgramID = programID
	assessment.SubjectID = subjectID
	if req.Assessment.StudyReportCriteriaId > 0 {
		assessment.StudyReportCriteriaID = req.Assessment.StudyReportCriteriaId
	}

	if err := s.repo.Create(assessment); err != nil {
		return nil, err
	}

	if len(req.Criteria) > 0 {
		if err := s.createCriteriaForAssessment(subjectID, assessment.ID, req.Criteria, now, userID); err != nil {
			return nil, err
		}
	}

	if req.LessonId > 0 {
		if err := s.assignAssessmentToLesson(c, assessment.ID, req.LessonId, req.CourseId, now, userID); err != nil {
			return nil, err
		}
	}

	return s.resource.FormatAssessment(assessment), nil
}

func (s *assessmentService) createCriteriaForAssessment(subjectID, assessmentID int64, items []*prot.AssessmentCriterionInput, now time.Time, userID int64) error {
	for _, item := range items {
		hasSubcriteria := len(item.Subcriteria) > 0
		maxScore := item.MaxScore
		if hasSubcriteria {
			maxScore = s.sumSubcriterionInputs(item.Subcriteria)
		}
		criterion := &models.AssessmentCriterion{
			SubjectID:      subjectID,
			Name:           item.Name,
			Description:    item.Description,
			HasSubcriteria: hasSubcriteria,
			MaxScore:       maxScore,
			CreatedAt:      now,
			UpdatedAt:      now,
			CreatedBy:      userID,
			UpdatedBy:      userID,
		}

		if err := s.criteriaRepo.Create(criterion); err != nil {
			return err
		}

		if hasSubcriteria {
			for _, subItem := range item.Subcriteria {
				subcriterion := &models.AssessmentSubcriterion{
					CriterionID: criterion.ID,
					Name:        subItem.Name,
					Description: subItem.Description,
					MaxScore:    subItem.MaxScore,
					CreatedAt:   now,
					UpdatedAt:   now,
					CreatedBy:   userID,
					UpdatedBy:   userID,
				}

				if err := s.subcriterionRepo.Create(subcriterion); err != nil {
					return err
				}
			}
		}

		ref := &models.AssessmentRefCriterion{
			AssessmentID:          assessmentID,
			AssessmentCriterionID: criterion.ID,
			CreatedAt:             now,
			UpdatedAt:             now,
			CreatedBy:             userID,
			UpdatedBy:             userID,
		}

		if err := s.criteriaRepo.CreateRefCriterion(ref); err != nil {
			return err
		}
	}

	return nil
}

func (s *assessmentService) sumSubcriterionInputs(items []*prot.AssessmentSubcriterionInput) float32 {
	var total float32
	for _, item := range items {
		total += item.MaxScore
	}
	return total
}

func (s *assessmentService) assignAssessmentToLesson(c *gin.Context, assessmentID, lessonID, courseID int64, now time.Time, userID int64) error {
	assignedAt := now

	ref := models.AssessmentRefLesson{
		AssessmentId: assessmentID,
		LessonId:     lessonID,
		CourseId:     courseID,
		AssignedAt:   &assignedAt,
		AssignedBy:   utils.PtrInt64(userID),
	}

	return s.repo.AssignRefLesson(ref)
}

func (s *assessmentService) resolveSubjectAndProgram(c *gin.Context, req *prot.CreateAssessmentWithCriteriaRequest) (int64, int64, error) {
	if req.LessonId > 0 {
		var result struct {
			SubjectID int64
			ProgramID int64
		}
		err := db.ReplicaDB.Table("lessons l").
			Joins("JOIN chapters c ON c.id = l.chapter_id").
			Joins("JOIN programs p ON p.id = c.program_id").
			Where("l.id = ? AND l.deleted_at IS NULL", req.LessonId).
			Select("p.subject_id AS subject_id, p.id AS program_id").
			Scan(&result).Error
		if err != nil {
			return 0, 0, err
		}
		if result.SubjectID == 0 || result.ProgramID == 0 {
			return 0, 0, fmt.Errorf("không tìm thấy subject_id/program_id cho lesson_id %d", req.LessonId)
		}
		return result.SubjectID, result.ProgramID, nil
	}

	// No lesson_id, fallback to program_id in request
	if req.Assessment.ProgramId == 0 {
		return 0, 0, fmt.Errorf("program_id is required when lesson_id is not provided")
	}

	var subjectID int64
	err := db.ReplicaDB.Table("programs").
		Where("id = ? AND deleted_at IS NULL", req.Assessment.ProgramId).
		Select("subject_id").
		Scan(&subjectID).Error
	if err != nil {
		return 0, 0, err
	}
	if subjectID == 0 {
		return 0, 0, fmt.Errorf("không tìm thấy subject_id cho program_id %d", req.Assessment.ProgramId)
	}

	return subjectID, req.Assessment.ProgramId, nil
}

func (s *assessmentService) SaveScoreBulk(c *gin.Context, req *prot.SaveAssessmentScoreBulkRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}
	if req.AssessmentId == 0 || req.LessonId == 0 || req.CourseId == 0 {
		return fmt.Errorf("assessment_id, lesson_id, course_id are required")
	}

	userID := int64(utils.GetCurrentUserId(c))
	now := time.Now().UTC()

	// 1. Lấy danh sách criteria IDs thuộc assessment
	validCriteriaIDs, err := s.criteriaRepo.GetRefCriteriaIDsByAssessmentID(req.AssessmentId)
	if err != nil {
		return fmt.Errorf("error getting criteria for assessment %d: %w", req.AssessmentId, err)
	}

	// Tạo map để check nhanh
	validCriteriaMap := make(map[int64]bool)
	for _, id := range validCriteriaIDs {
		validCriteriaMap[id] = true
	}

	// 2. Lấy danh sách subcriteria IDs thuộc các criteria của assessment
	// Map: criterionID -> []subcriteriaIDs
	validSubcriteriaByCriterionMap := make(map[int64]map[int64]bool)
	if len(validCriteriaIDs) > 0 {
		for _, criterionID := range validCriteriaIDs {
			subcriteriaIDs, err := s.subcriterionRepo.GetIDsByCriterionID(criterionID)
			if err != nil {
				return fmt.Errorf("error getting subcriteria for criterion %d: %w", criterionID, err)
			}
			subcriteriaMap := make(map[int64]bool)
			for _, subID := range subcriteriaIDs {
				subcriteriaMap[subID] = true
			}
			validSubcriteriaByCriterionMap[criterionID] = subcriteriaMap
		}
	}

	// 3. Validate tất cả criteria/subcriteria trong request phải thuộc assessment
	for _, student := range req.Students {
		if student.UserId == 0 {
			continue
		}

		for _, criterionInput := range student.Criteria {
			criterionID := criterionInput.CriteriaId

			// Validate criterion
			if !validCriteriaMap[criterionID] {
				return fmt.Errorf("criterion_id %d does not belong to assessment %d", criterionID, req.AssessmentId)
			}

			// Validate subcriteria - phải thuộc criterion đó
			if len(criterionInput.Subcriteria) > 0 {
				subcriteriaMap, ok := validSubcriteriaByCriterionMap[criterionID]
				if !ok {
					// Criterion không có subcriteria nhưng request có subcriteria
					return fmt.Errorf("criterion_id %d does not have subcriteria, but subcriteria were provided", criterionID)
				}

				for _, subInput := range criterionInput.Subcriteria {
					subcriterionID := subInput.SubcriteriaId
					if !subcriteriaMap[subcriterionID] {
						return fmt.Errorf("subcriteria_id %d does not belong to criterion_id %d in assessment %d", subcriterionID, criterionID, req.AssessmentId)
					}
				}
			}
		}
	}

	// Group students theo (assessment_id, student_id, course_id) - chỉ lấy student cuối cùng nếu trùng
	studentMap := make(map[int64]*prot.AssessmentScoreStudentInput)
	for _, student := range req.Students {
		if student.UserId == 0 {
			continue
		}
		// Nếu đã có student với cùng user_id, ghi đè bằng student mới (lấy student cuối cùng)
		studentMap[student.UserId] = student
	}

	// Xử lý từng học sinh (đã distinct)
	for _, student := range studentMap {

		var totalScore float64
		var scoreDetails []*models.AssessmentScoreDetail

		// Xử lý từng criteria
		for _, criterionInput := range student.Criteria {
			criterionID := criterionInput.CriteriaId
			var criterionScore float64

			// Nếu có subcriteria, tính điểm từ subcriteria
			if len(criterionInput.Subcriteria) > 0 {
				for _, subInput := range criterionInput.Subcriteria {
					score := float64(subInput.Score)
					detail := &models.AssessmentScoreDetail{
						AssessmentScoreID: 0, // Sẽ set sau khi tạo AssessmentScore
						CriterionID:       criterionID,
						SubcriterionID:    &subInput.SubcriteriaId,
						Score:             score,
						CreatedAt:         now,
						CreatedBy:         userID,
						UpdatedAt:         now,
						UpdatedBy:         userID,
					}
					scoreDetails = append(scoreDetails, detail)
					criterionScore += score
				}
			} else {
				// Không có subcriteria, dùng điểm criteria
				score := float64(criterionInput.Score)
				criterionScore = score
				detail := &models.AssessmentScoreDetail{
					AssessmentScoreID: 0,
					CriterionID:       criterionID,
					SubcriterionID:    nil,
					Score:             score,
					CreatedAt:         now,
					CreatedBy:         userID,
					UpdatedAt:         now,
					UpdatedBy:         userID,
				}
				scoreDetails = append(scoreDetails, detail)
			}

			totalScore += criterionScore
		}

		// Xóa các assessment_scores cũ có cùng (assessment_id, student_id, course_id)
		if err := s.scoreRepo.DeleteOldScores(req.AssessmentId, student.UserId, req.CourseId, userID); err != nil {
			return fmt.Errorf("error deleting old scores for student %d: %w", student.UserId, err)
		}

		// Tạo AssessmentScore
		assessmentScore := &models.AssessmentScore{
			AssessmentID: req.AssessmentId,
			StudentID:    student.UserId,
			LessonID:     req.LessonId,
			CourseID:     req.CourseId,
			TotalScore:   totalScore,
			StatusScored: 1, // Đã chấm điểm
			IsLate:       false,
			CreatedAt:    now,
			CreatedBy:    userID,
			UpdatedAt:    now,
			UpdatedBy:    userID,
			FileInfos:    convertAssessmentFileInfos(student.FileInfos),
		}

		if err := s.scoreRepo.Create(assessmentScore); err != nil {
			return fmt.Errorf("error creating assessment score for student %d: %w", student.UserId, err)
		}

		// Cập nhật AssessmentScoreID cho các detail
		for _, detail := range scoreDetails {
			detail.AssessmentScoreID = assessmentScore.ID
		}

		// Tạo score details
		if len(scoreDetails) > 0 {
			if err := s.scoreRepo.CreateDetails(scoreDetails); err != nil {
				return fmt.Errorf("error creating score details for student %d: %w", student.UserId, err)
			}
		}
	}

	return nil
}

func convertAssessmentFileInfos(infos []*prot.AssessmentFileInfo) models.MediaInfos {
	if len(infos) == 0 {
		return nil
	}

	result := make(models.MediaInfos, 0, len(infos))
	for _, info := range infos {
		result = append(result, models.MediaDetail{
			Id:   info.Id,
			Disk: info.Disk,
			Path: info.Path,
		})
	}

	return result
}

func (s *assessmentService) syncAssessmentCriteria(assessmentID int64, req *prot.AssessmentRequest, now time.Time, userID int64) error {
	var newCriteriaIDs []int64

	// Xác định danh sách criteria mới
	if req.AssessmentCriteriaGroupId > 0 {
		// Lấy criteria từ group
		criteriaIDs, err := s.groupRepo.GetCriteriaIDsByGroupID(req.AssessmentCriteriaGroupId)
		if err != nil {
			return fmt.Errorf("error getting criteria from group %d: %w", req.AssessmentCriteriaGroupId, err)
		}
		newCriteriaIDs = criteriaIDs
	} else if len(req.AssessmentCriteriaIds) > 0 {
		// Dùng danh sách criteria IDs được truyền vào
		newCriteriaIDs = req.AssessmentCriteriaIds
	} else {
		// assessment_criteria_group_id rỗng và không có AssessmentCriteriaIds: coi là danh sách rỗng, xóa mềm toàn bộ ref hiện tại
		newCriteriaIDs = []int64{}
	}

	// Lấy danh sách criteria hiện tại của assessment
	currentCriteriaIDs, err := s.criteriaRepo.GetRefCriteriaIDsByAssessmentID(assessmentID)
	if err != nil {
		return fmt.Errorf("error getting current criteria for assessment %d: %w", assessmentID, err)
	}

	// Tạo map để dễ so sánh
	currentMap := make(map[int64]bool)
	for _, id := range currentCriteriaIDs {
		currentMap[id] = true
	}

	newMap := make(map[int64]bool)
	for _, id := range newCriteriaIDs {
		if id > 0 {
			newMap[id] = true
		}
	}

	// Tìm các criteria cần thêm mới
	for _, criterionID := range newCriteriaIDs {
		if criterionID <= 0 {
			continue
		}
		if !currentMap[criterionID] {
			// Kiểm tra xem có bị xóa mềm không, nếu có thì restore
			var existingRef models.AssessmentRefCriterion
			err := db.MasterDB.Unscoped().
				Where("assessment_id = ? AND assessment_criterion_id = ?", assessmentID, criterionID).
				First(&existingRef).Error

			if err == nil && existingRef.DeletedAt != nil {
				// Đã tồn tại nhưng bị xóa mềm, restore
				if err := s.criteriaRepo.RestoreRefCriterion(assessmentID, criterionID, userID); err != nil {
					return fmt.Errorf("error restoring criterion %d for assessment %d: %w", criterionID, assessmentID, err)
				}
			} else if err != nil {
				// Chưa tồn tại, tạo mới
				ref := &models.AssessmentRefCriterion{
					AssessmentID:          assessmentID,
					AssessmentCriterionID: criterionID,
					CreatedBy:             userID,
					CreatedAt:             now,
					UpdatedBy:             userID,
					UpdatedAt:             now,
				}

				if err := s.criteriaRepo.CreateRefCriterion(ref); err != nil {
					// Bỏ qua lỗi duplicate (nếu đã tồn tại)
					if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "UNIQUE") {
						return fmt.Errorf("error assigning criterion %d to assessment %d: %w", criterionID, assessmentID, err)
					}
				}
			}
			// Nếu đã tồn tại và không bị xóa, không cần làm gì
		}
	}

	// Tìm các criteria cần xóa (soft delete)
	for _, criterionID := range currentCriteriaIDs {
		if !newMap[criterionID] {
			// Criteria này không còn trong danh sách mới, xóa mềm
			var ref models.AssessmentRefCriterion
			if err := db.MasterDB.
				Where("assessment_id = ? AND assessment_criterion_id = ? AND deleted_at IS NULL", assessmentID, criterionID).
				First(&ref).Error; err == nil {
				ref.DeletedAt = &now
				ref.DeletedBy = userID
				ref.UpdatedBy = userID
				ref.UpdatedAt = now
				if err := db.MasterDB.Save(&ref).Error; err != nil {
					return fmt.Errorf("error soft deleting criterion %d for assessment %d: %w", criterionID, assessmentID, err)
				}
			}
		}
	}

	return nil
}

func (s *assessmentService) GetPublish(c *gin.Context) (*prot.GetPublishAssessmentResponse, error) {
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

	return &prot.GetPublishAssessmentResponse{
		CourseId:     int64(courseId),
		AssessmentId: int64(assessmentId),
		Publish:      publish,
	}, nil
}

func (s *assessmentService) UpdatePublish(c *gin.Context) (*prot.UpdatePublishAssessmentResponse, error) {
	req, err, _ := utils.GetBody[*prot.UpdatePublishAssessmentRequest](c, func() *prot.UpdatePublishAssessmentRequest {
		return &prot.UpdatePublishAssessmentRequest{}
	})

	if err != nil {
		return nil, err
	}

	if req.AssessmentId == 0 {
		return nil, fmt.Errorf("assessment_id are required")
	}

	s.repo.UpdatePublish(req.AssessmentId, req.CourseId, req.PublishCourseIds, req.Publish)

	return &prot.UpdatePublishAssessmentResponse{
		AssessmentId:     req.AssessmentId,
		PublishCourseIds: req.PublishCourseIds,
	}, nil
}
