package examples

import (
	"be-lms/services"
	"fmt"
)

// ExampleUsageOfLessonScheduleCopySharedService ví dụ về cách sử dụng shared service
func ExampleUsageOfLessonScheduleCopySharedService() {
	// Tạo instance của shared service
	copyService := services.NewLessonScheduleCopySharedService()

	// Ví dụ 1: Copy lesson schedules giữa 2 courses
	sourceCourseID := int64(1)
	targetCourseID := int64(2)

	result, err := copyService.CopyLessonSchedulesBetweenCourses(sourceCourseID, targetCourseID)
	if err != nil {
		fmt.Printf("Lỗi: %v\n", err)
		return
	}

	if result.Success {
		fmt.Printf("✅ %s\n", result.Message)
		fmt.Printf("📊 Số lịch học đã copy: %d\n", result.CopiedSchedulesCount)
	} else {
		fmt.Printf("❌ %s\n", result.Message)
	}
}

// ExampleUsageInAnotherService ví dụ sử dụng trong service khác
type AnotherService struct {
	copyService services.LessonScheduleCopySharedService
}

func NewAnotherService() *AnotherService {
	return &AnotherService{
		copyService: services.NewLessonScheduleCopySharedService(),
	}
}

func (s *AnotherService) SomeBusinessLogic(sourceCourseID, targetCourseID int64) error {
	// Thực hiện copy lesson schedules
	result, err := s.copyService.CopyLessonSchedulesBetweenCourses(sourceCourseID, targetCourseID)
	if err != nil {
		return fmt.Errorf("lỗi khi copy lesson schedules: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("copy lesson schedules thất bại: %s", result.Message)
	}

	// Tiếp tục logic nghiệp vụ khác...
	fmt.Printf("Copy thành công %d lịch học\n", result.CopiedSchedulesCount)
	return nil
}

// ExampleUsageInController ví dụ sử dụng trong controller khác
func ExampleUsageInController() {
	// Trong controller, bạn có thể inject shared service
	copyService := services.NewLessonScheduleCopySharedService()

	// Sử dụng trong method của controller
	sourceCourseID := int64(10)
	targetCourseID := int64(20)

	result, err := copyService.CopyLessonSchedulesBetweenCourses(sourceCourseID, targetCourseID)
	if err != nil {
		// Xử lý lỗi
		fmt.Printf("Lỗi hệ thống: %v\n", err)
		return
	}

	// Trả về response
	if result.Success {
		fmt.Printf("API response: %s\n", result.Message)
	} else {
		fmt.Printf("API error: %s\n", result.Message)
	}
}
