package jobs

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/services"
	"errors"
	"time"
)

// SyncCourseFamilySchedulesJob chạy job sync lịch học cho toàn bộ family
func SyncCourseFamilySchedulesJob(courseID int64) error {
	// Defer recover để tránh panic
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in SyncCourseFamilySchedulesJob: %v", r)
		}
	}()

	config.Log.Infof("Starting sync course family schedules job for course_id: %d", courseID)

	// Kiểm tra database connection
	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return errors.New("database not connected")
	}

	if courseID <= 0 {
		return errors.New("course_id phải lớn hơn 0")
	}

	startTime := time.Now()

	// Lấy service để sử dụng logic hiện có
	courseScheduleService := services.NewCourseScheduleService()
	
	// Gọi logic sync family schedules
	_, err := courseScheduleService.SyncFamilySchedules(courseID)
	if err != nil {
		config.Log.Errorf("Error syncing course family schedules: %v", err)
		return err
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished sync course family schedules job for course_id: %d in %v", courseID, duration)
	
	return nil
}
