package services

import (
	"be-lms/prot"
	"be-lms/requests"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type LessonScheduleCopyService interface {
	CopyLessonSchedules(c *gin.Context, req *requests.CopyLessonScheduleRequest) (*prot.CopyLessonScheduleResponse, error)
}

type lessonScheduleCopyService struct {
	sharedService LessonScheduleCopySharedService
}

func NewLessonScheduleCopyService() LessonScheduleCopyService {
	return &lessonScheduleCopyService{
		sharedService: NewLessonScheduleCopySharedService(),
	}
}

func (s *lessonScheduleCopyService) CopyLessonSchedules(c *gin.Context, req *requests.CopyLessonScheduleRequest) (*prot.CopyLessonScheduleResponse, error) {
	// Kiểm tra quyền truy cập (có thể thêm logic kiểm tra role nếu cần)
	userID := utils.GetCurrentUserId(c)
	if userID <= 0 {
		return &prot.CopyLessonScheduleResponse{
			Success: false,
			Message: "Không có quyền truy cập",
		}, nil
	}

	// Gọi shared service để thực hiện copy
	result, err := s.sharedService.CopyLessonSchedulesBetweenCourses(req.SourceCourseID, req.TargetCourseID)
	if err != nil {
		return &prot.CopyLessonScheduleResponse{
			Success: false,
			Message: "Lỗi hệ thống: " + err.Error(),
		}, err
	}

	return &prot.CopyLessonScheduleResponse{
		Success:              result.Success,
		Message:              result.Message,
		CopiedSchedulesCount: result.CopiedSchedulesCount,
	}, nil
}
