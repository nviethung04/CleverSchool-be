package services

import (
	"fmt"

	"be-lms/repositories"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type AssessmentCourseService interface {
	AssignCoursesToAssessment(c *gin.Context, assessmentID, lessonID int64) error
}

type assessmentCourseService struct {
	refRepo repositories.AssessmentRefLessonRepository
}

func NewAssessmentCourseService() AssessmentCourseService {
	return &assessmentCourseService{
		refRepo: repositories.NewAssessmentRefLessonRepository(),
	}
}

func (s *assessmentCourseService) AssignCoursesToAssessment(c *gin.Context, assessmentID, lessonID int64) error {
	// Validate inputs
	if assessmentID <= 0 {
		return fmt.Errorf("assessment_id is required and must be greater than 0")
	}
	if lessonID <= 0 {
		return fmt.Errorf("lesson_id is required and must be greater than 0")
	}

	// Lấy assigned_by từ context
	assignedBy := int64(utils.GetCurrentUserId(c))
	if assignedBy <= 0 {
		return fmt.Errorf("assigned_by is required")
	}

	// 1. Lấy program_id từ lesson_id (join lessons -> chapters -> program_id)
	programID, err := s.refRepo.GetProgramIDByLessonID(lessonID)
	if err != nil {
		return fmt.Errorf("error getting program_id from lesson_id %d: %w", lessonID, err)
	}
	if programID <= 0 {
		return fmt.Errorf("lesson_id %d does not have a valid program_id", lessonID)
	}

	// 2. Lấy tất cả courses thuộc program
	courses, err := s.refRepo.GetCoursesByProgramID(programID)
	if err != nil {
		return fmt.Errorf("error getting courses by program_id %d: %w", programID, err)
	}

	if len(courses) == 0 {
		return fmt.Errorf("no courses found for program_id %d", programID)
	}

	// 3. Lấy danh sách course IDs
	courseIDs := make([]int64, 0, len(courses))
	for _, course := range courses {
		courseIDs = append(courseIDs, course.ID)
	}

	// 4. Gán tất cả courses vào assessment_ref_lessons
	// Nếu record đã tồn tại (cùng bộ 3 assessment_id, lesson_id, course_id) → giữ nguyên
	// Nếu chưa tồn tại → thêm mới
	err = s.refRepo.AssignCoursesToAssessment(assessmentID, lessonID, courseIDs, assignedBy)
	if err != nil {
		return fmt.Errorf("error assigning courses to assessment: %w", err)
	}

	return nil
}

