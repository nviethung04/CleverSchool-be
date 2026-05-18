package jobs

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"time"
)

// SyncTeacherClassesJob sync user_classes cho giáo viên từ user_courses
func SyncTeacherClassesJob() {
	// Defer recover để tránh panic
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in SyncTeacherClassesJob: %v", r)
		}
	}()

	config.Log.Info("Starting sync teacher classes job")

	// Kiểm tra database connection
	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return
	}

	startTime := time.Now()

	// Sync teacher classes
	err := syncTeacherClasses()
	if err != nil {
		config.Log.Errorf("Error syncing teacher classes: %v", err)
		return
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished sync teacher classes job in %v", duration)
}

// syncTeacherClasses thực hiện logic sync teacher classes
func syncTeacherClasses() error {
	// Bước 1: Lấy tất cả user_courses có role_id = 2 (giáo viên)
	type TeacherCourse struct {
		UserID   int64 `gorm:"column:user_id"`
		CourseID int64 `gorm:"column:course_id"`
	}

	var teacherCourses []TeacherCourse
	err := db.MasterDB.Table("user_courses uc").
		Select("DISTINCT uc.user_id, uc.course_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = uc.user_id").
		Where("urr.role_id = ?", 2).
		Find(&teacherCourses).Error

	if err != nil {
		config.Log.Errorf("Error getting teacher courses: %v", err)
		return err
	}

	config.Log.Infof("Found %d teacher-course pairs to process", len(teacherCourses))

	totalInserted := 0

	// Bước 2: Với mỗi cặp user_id và course_id, lấy danh sách class_id thuộc course đó
	for _, tc := range teacherCourses {
		// Lấy danh sách class_id thuộc course này
		// Điều kiện: class có ít nhất 1 học sinh (role_id = 3) thuộc cả class và course đó
		// Logic: Tìm các class_id mà có ít nhất 1 user có:
		//   - role_id = 3 (học sinh)
		//   - user_id trong user_classes với class_id đó
		//   - user_id trong user_courses với course_id đó
		var classIDs []int64
		err := db.MasterDB.Table("user_classes ucl").
			Select("DISTINCT ucl.class_id").
			Joins("JOIN user_ref_roles urr ON urr.user_id = ucl.user_id AND urr.role_id = ?", 3).
			Joins("JOIN user_courses uco ON uco.user_id = ucl.user_id AND uco.course_id = ?", tc.CourseID).
			Where("ucl.class_id IS NOT NULL").
			Group("ucl.class_id").
			Having("COUNT(DISTINCT ucl.user_id) > 0").
			Pluck("ucl.class_id", &classIDs).Error

		if err != nil {
			config.Log.Errorf("Error getting class IDs for course %d: %v", tc.CourseID, err)
			continue
		}

		if len(classIDs) == 0 {
			continue
		}

		// Bước 3: Thêm vào user_classes các cặp user_id và class_id (nếu chưa tồn tại)
		for _, classID := range classIDs {
			// Kiểm tra xem đã tồn tại chưa
			var count int64
			err := db.MasterDB.Table("user_classes").
				Where("user_id = ? AND class_id = ?", tc.UserID, classID).
				Count(&count).Error

			if err != nil {
				config.Log.Errorf("Error checking existing user_class for user %d, class %d: %v", tc.UserID, classID, err)
				continue
			}

			// Nếu chưa tồn tại thì thêm mới
			if count == 0 {
				now := time.Now()
				err := db.MasterDB.Table("user_classes").Create(map[string]interface{}{
					"user_id":    tc.UserID,
					"class_id":   classID,
					"is_current": true,
					"start_time": now,
				}).Error

				if err != nil {
					config.Log.Errorf("Error inserting user_class for user %d, class %d: %v", tc.UserID, classID, err)
					continue
				}

				totalInserted++
			}
		}
	}

	config.Log.Infof("Successfully inserted %d new user_class records", totalInserted)
	return nil
}
