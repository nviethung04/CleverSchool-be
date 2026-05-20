package services

import (
	"be-Clever School/database/db"
	"fmt"
)

// LessonScheduleCopySharedService interface cho service copy lesson schedules
type LessonScheduleCopySharedService interface {
	CopyLessonSchedulesBetweenCourses(sourceCourseID, targetCourseID int64) (*LessonScheduleCopyResult, error)
}

type lessonScheduleCopySharedService struct{}

// LessonScheduleCopyResult kết quả trả về khi copy lesson schedules
type LessonScheduleCopyResult struct {
	Success              bool   `json:"success"`
	Message              string `json:"message"`
	CopiedSchedulesCount int32  `json:"copied_schedules_count"`
}

// NewLessonScheduleCopySharedService tạo instance mới của service
func NewLessonScheduleCopySharedService() LessonScheduleCopySharedService {
	return &lessonScheduleCopySharedService{}
}

// CopyLessonSchedulesBetweenCourses copy lesson schedules từ course nguồn sang course đích
func (s *lessonScheduleCopySharedService) CopyLessonSchedulesBetweenCourses(sourceCourseID, targetCourseID int64) (*LessonScheduleCopyResult, error) {
	// 1. Kiểm tra xem 2 course có cùng program_id không
	var sourceProgramID, targetProgramID int64
	err := db.ReplicaDB.Table("courses").
		Select("program_id").
		Where("id = ?", sourceCourseID).
		Scan(&sourceProgramID).Error
	if err != nil {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Không tìm thấy course nguồn",
		}, err
	}

	err = db.ReplicaDB.Table("courses").
		Select("program_id").
		Where("id = ?", targetCourseID).
		Scan(&targetProgramID).Error
	if err != nil {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Không tìm thấy course đích",
		}, err
	}

	if sourceProgramID != targetProgramID {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Hai course không cùng program, không thể copy lịch học",
		}, nil
	}

	// 2. Kiểm tra xem course nguồn có tồn tại lịch trong bảng lesson_schedules không
	var scheduleCount int64
	err = db.ReplicaDB.Table("lesson_schedules").
		Where("course_id = ?", sourceCourseID).
		Count(&scheduleCount).Error
	if err != nil {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Lỗi khi kiểm tra lịch học course nguồn",
		}, err
	}

	if scheduleCount == 0 {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Course nguồn không có lịch học nào để copy",
		}, nil
	}

	// 3. Lấy danh sách lịch học từ course nguồn
	var sourceSchedules []map[string]interface{}
	err = db.ReplicaDB.Table("lesson_schedules").
		Where("course_id = ?", sourceCourseID).
		Order("sort_position ASC, scheduled_date ASC").
		Find(&sourceSchedules).Error
	if err != nil {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Lỗi khi lấy danh sách lịch học course nguồn",
		}, err
	}

	// 4. Xóa lịch học cũ của course đích (nếu có)
	err = db.MasterDB.Table("lesson_schedules").Where("course_id = ?", targetCourseID).Delete(nil).Error
	if err != nil {
		return &LessonScheduleCopyResult{
			Success: false,
			Message: "Lỗi khi xóa lịch học cũ của course đích",
		}, err
	}

	// 5. Tạo lịch học mới cho course đích
	var copiedCount int32 = 0
	for _, schedule := range sourceSchedules {
		// Tạo map cho insert
		newSchedule := map[string]interface{}{
			"course_id":      targetCourseID,
			"lesson_id":      schedule["lesson_id"],
			"shift_id":       schedule["shift_id"],
			"scheduled_date": schedule["scheduled_date"],
			"week_id":        schedule["week_id"],
			"sort_position":  schedule["sort_position"],
		}

		err = db.MasterDB.Table("lesson_schedules").Create(newSchedule).Error
		if err != nil {
			return &LessonScheduleCopyResult{
				Success: false,
				Message: fmt.Sprintf("Lỗi khi tạo lịch học cho course đích: %v", err),
			}, err
		}
		copiedCount++
	}

	return &LessonScheduleCopyResult{
		Success:              true,
		Message:              fmt.Sprintf("Copy thành công %d lịch học từ course %d sang course %d", copiedCount, sourceCourseID, targetCourseID),
		CopiedSchedulesCount: copiedCount,
	}, nil
}
