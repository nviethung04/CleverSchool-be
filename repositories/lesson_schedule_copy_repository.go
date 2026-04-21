package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
	"fmt"
)

type LessonScheduleCopyRepository interface {
	CopyLessonSchedules(req *requests.CopyLessonScheduleRequest) (*dto.CopyLessonScheduleResponse, error)
}

type lessonScheduleCopyRepository struct{}

func NewLessonScheduleCopyRepository() LessonScheduleCopyRepository {
	return &lessonScheduleCopyRepository{}
}

func (r *lessonScheduleCopyRepository) CopyLessonSchedules(req *requests.CopyLessonScheduleRequest) (*dto.CopyLessonScheduleResponse, error) {
	// 1. Kiểm tra xem 2 course có cùng program_id không
	var sourceProgramID, targetProgramID int64
	err := db.ReplicaDB.Table("courses").
		Select("program_id").
		Where("id = ?", req.SourceCourseID).
		Scan(&sourceProgramID).Error
	if err != nil {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Không tìm thấy course nguồn",
		}, err
	}

	err = db.ReplicaDB.Table("courses").
		Select("program_id").
		Where("id = ?", req.TargetCourseID).
		Scan(&targetProgramID).Error
	if err != nil {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Không tìm thấy course đích",
		}, err
	}

	if sourceProgramID != targetProgramID {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Hai course không cùng program, không thể copy lịch học",
		}, nil
	}

	// 2. Kiểm tra xem course nguồn có tồn tại lịch trong bảng lesson_schedules không
	var scheduleCount int64
	err = db.ReplicaDB.Table("lesson_schedules").
		Where("course_id = ?", req.SourceCourseID).
		Count(&scheduleCount).Error
	if err != nil {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Lỗi khi kiểm tra lịch học course nguồn",
		}, err
	}

	if scheduleCount == 0 {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Course nguồn không có lịch học nào để copy",
		}, nil
	}

	// 3. Lấy danh sách lịch học từ course nguồn
	var sourceSchedules []map[string]interface{}
	err = db.ReplicaDB.Table("lesson_schedules").
		Where("course_id = ?", req.SourceCourseID).
		Order("sort_position ASC, scheduled_date ASC").
		Find(&sourceSchedules).Error
	if err != nil {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Lỗi khi lấy danh sách lịch học course nguồn",
		}, err
	}

	// 4. Xóa lịch học cũ của course đích (nếu có)
	err = db.MasterDB.Table("lesson_schedules").Where("course_id = ?", req.TargetCourseID).Delete(nil).Error
	if err != nil {
		return &dto.CopyLessonScheduleResponse{
			Success: false,
			Message: "Lỗi khi xóa lịch học cũ của course đích",
		}, err
	}

	// 5. Tạo lịch học mới cho course đích
	var copiedCount int32 = 0
	for _, schedule := range sourceSchedules {
		// Tạo map cho insert
		newSchedule := map[string]interface{}{
			"course_id":      req.TargetCourseID,
			"lesson_id":      schedule["lesson_id"],
			"shift_id":       schedule["shift_id"],
			"scheduled_date": schedule["scheduled_date"],
			"week_id":        schedule["week_id"],
			"sort_position":  schedule["sort_position"],
		}

		err = db.MasterDB.Table("lesson_schedules").Create(newSchedule).Error
		if err != nil {
			return &dto.CopyLessonScheduleResponse{
				Success: false,
				Message: fmt.Sprintf("Lỗi khi tạo lịch học cho course đích: %v", err),
			}, err
		}
		copiedCount++
	}

	return &dto.CopyLessonScheduleResponse{
		Success:              true,
		Message:              fmt.Sprintf("Copy thành công %d lịch học từ course %d sang course %d", copiedCount, req.SourceCourseID, req.TargetCourseID),
		CopiedSchedulesCount: copiedCount,
	}, nil
}
