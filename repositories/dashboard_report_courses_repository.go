package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/table_manager"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DashboardReportCoursesRepository interface {
	CalculateCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error)
	SaveCourseStatistics(statistics []models.DashboardReportCourses) error
	GetExistingReport(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error)
	DeleteExistingReport(courseID int64, startDate, endDate time.Time) error
}

type dashboardReportCoursesRepository struct{}

func NewDashboardReportCoursesRepository() DashboardReportCoursesRepository {
	return &dashboardReportCoursesRepository{}
}

// createActivityLogsUnionQuery tạo UNION query cho các bảng activity_logs theo tháng
func (r *dashboardReportCoursesRepository) createActivityLogsUnionQuery(startTime, endTime time.Time) (string, error) {
	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	// Tạo UNION ALL cho các bảng tháng
	var unionParts []string
	for _, tableName := range tableNames {
		// Kiểm tra bảng có tồn tại không
		var exists bool
		err := db.ReplicaDB.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = ?
			)
		`, tableName).Scan(&exists).Error

		if err != nil || !exists {
			continue
		}

		unionParts = append(unionParts, fmt.Sprintf("SELECT * FROM %s", tableName))
	}

	// Nếu không có bảng nào tồn tại, trả về query rỗng
	if len(unionParts) == 0 {
		return "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs", nil
	}

	// Tạo virtual table từ UNION các bảng tháng
	return fmt.Sprintf("(%s) AS all_activity_logs", strings.Join(unionParts, " UNION ALL ")), nil
}

func (r *dashboardReportCoursesRepository) CalculateCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error) {
	// Lấy danh sách tất cả khóa học
	var courses []models.Course
	if err := db.ReplicaDB.Where("deleted_at IS NULL").Find(&courses).Error; err != nil {
		return nil, err
	}

	var results []models.DashboardReportCourses

	// Tạo endDate đến cuối ngày (23:59:59)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// Tạo UNION query cho activity_logs
	allActivityLogs, err := r.createActivityLogsUnionQuery(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Batch processing với batch size 100
	batchSize := 100
	totalCourses := len(courses)
	
	fmt.Printf("🔄 Bắt đầu xử lý %d khóa học với batch size %d\n", totalCourses, batchSize)
	
	for i := 0; i < totalCourses; i += batchSize {
		end := i + batchSize
		if end > totalCourses {
			end = totalCourses
		}
		
		batch := courses[i:end]
		fmt.Printf("📦 Xử lý batch %d-%d/%d (%.1f%%)\n", 
			i+1, end, totalCourses, float64(end)/float64(totalCourses)*100)
		
		// Xử lý batch
		batchResults, err := r.processCourseBatch(batch, startDate, endDate, endOfDay, allActivityLogs)
		if err != nil {
			fmt.Printf("❌ Lỗi batch %d-%d: %v\n", i+1, end, err)
			continue
		}
		
		results = append(results, batchResults...)
		fmt.Printf("✅ Hoàn thành batch %d-%d (%d records)\n", i+1, end, len(batchResults))
	}

	return results, nil
}

func (r *dashboardReportCoursesRepository) processCourseBatch(courses []models.Course, startDate, endDate, endOfDay time.Time, allActivityLogs string) ([]models.DashboardReportCourses, error) {
	var results []models.DashboardReportCourses
	
	for _, course := range courses {
		statistics := models.DashboardReportCourses{
			CourseID:  course.ID,
			StartDate: startDate,
			EndDate:   endDate,
		}

		// 1. Đếm tổng số học sinh của khóa học (qua user_courses)
		var totalStudents int64
		studentQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			WHERE uc.course_id = ? AND u.deleted_at IS NULL AND urr.role_id = 3 AND u.created_at <= ?
		`
		if err := db.ReplicaDB.Raw(studentQuery, course.ID, endOfDay).Scan(&totalStudents).Error; err != nil {
			return nil, err
		}
		statistics.TotalStudents = totalStudents

		// 2. Đếm tổng số giáo viên của khóa học và lấy danh sách teacher_ids + teacher_infos
		var totalTeachers int64
		var teacherIDs []string
		var teacherInfos []models.DashboardTeacherInfo
		teacherQuery := `
			SELECT u.id, u.username, u.name
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			WHERE uc.course_id = ? AND u.deleted_at IS NULL AND urr.role_id = 2 AND u.created_at <= ?
		`
		
		var teachers []struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
		}
		if err := db.ReplicaDB.Raw(teacherQuery, course.ID, endOfDay).Scan(&teachers).Error; err != nil {
			return nil, err
		}
		
		totalTeachers = int64(len(teachers))
		for _, teacher := range teachers {
			teacherIDs = append(teacherIDs, strconv.FormatInt(teacher.ID, 10))
			teacherInfos = append(teacherInfos, models.DashboardTeacherInfo{
				ID:       teacher.ID,
				Username: teacher.Username,
				Name:     teacher.Name,
			})
		}
		statistics.TotalTeachers = totalTeachers
		statistics.TeacherIDs = strings.Join(teacherIDs, ",")
		statistics.TeacherInfos = teacherInfos

		// 3. Đếm số học sinh active trong thời gian (role_id = 3)
		var activeStudents int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
			studentQuery := fmt.Sprintf(`
				SELECT COUNT(DISTINCT all_activity_logs.user_id)
				FROM %s
				INNER JOIN users u ON all_activity_logs.user_id = u.id
				INNER JOIN user_courses uc ON u.id = uc.user_id
				WHERE uc.course_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(studentQuery, course.ID, 3, startDate, endDate).Scan(&activeStudents).Error; err != nil {
				// Nếu có lỗi, set về 0
				activeStudents = 0
			}
		}
		statistics.ActiveStudents = activeStudents

		// 4. Đếm số giáo viên active trong thời gian (role_id = 2)
		var activeTeachers int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
			teacherQuery := fmt.Sprintf(`
				SELECT COUNT(DISTINCT all_activity_logs.user_id)
				FROM %s
				INNER JOIN users u ON all_activity_logs.user_id = u.id
				INNER JOIN user_courses uc ON u.id = uc.user_id
				WHERE uc.course_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(teacherQuery, course.ID, 2, startDate, endDate).Scan(&activeTeachers).Error; err != nil {
				// Nếu có lỗi, set về 0
				activeTeachers = 0
			}
		}
		statistics.ActiveTeachers = activeTeachers

		// 5. Đếm số học sinh hoàn thành ít nhất 1 bài tập trong thời gian
		var studentsCompletedHomework int64
		homeworkQuery := `
			SELECT COUNT(DISTINCT hu.user_id)
			FROM homework_users hu
			INNER JOIN users u ON hu.user_id = u.id
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN homeworks h ON hu.homework_id = h.id
			WHERE uc.course_id = ? AND u.deleted_at IS NULL 
			AND hu.questions_completed >= h.total_questions
			AND hu.created_at >= ? AND hu.created_at <= ?
		`

		if err := db.ReplicaDB.Raw(homeworkQuery, course.ID, startDate, endDate).Scan(&studentsCompletedHomework).Error; err != nil {
			return nil, err
		}
		statistics.StudentsCompletedHomework = studentsCompletedHomework

		// 6. Tính tổng số homework (không giới hạn thời gian)
		var totalHomeworks int64
		totalHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM homeworks h
			INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id
			WHERE hrl.course_id = ? 
			AND h.deleted_at IS NULL
		`
		if err := db.ReplicaDB.Raw(totalHomeworksQuery, course.ID).Scan(&totalHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			totalHomeworks = 0
		}
		statistics.TotalHomeworks = totalHomeworks

		// 7. Tính số homework đã được giao (assigned_at IS NOT NULL)
		var assignedHomeworks int64
		assignedHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM homeworks h
			INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id
			WHERE hrl.course_id = ?
			AND hrl.assigned_at IS NOT NULL
			AND h.deleted_at IS NULL
		`
		if err := db.ReplicaDB.Raw(assignedHomeworksQuery, course.ID).Scan(&assignedHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			assignedHomeworks = 0
		}
		statistics.AssignedHomeworks = assignedHomeworks

		// 8. Tính số homework đã hoàn thành (có ít nhất 1 user hoàn thành)
		var completedHomeworks int64
		completedHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM homeworks h
			INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id
			INNER JOIN homework_users hu ON h.id = hu.homework_id AND hu.lesson_id = hrl.lesson_id
			WHERE hrl.course_id = ?
			AND h.deleted_at IS NULL
			AND hrl.assigned_at IS NOT NULL
			AND hu.questions_completed >= h.total_questions
		`
		if err := db.ReplicaDB.Raw(completedHomeworksQuery, course.ID).Scan(&completedHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			completedHomeworks = 0
		}
		statistics.CompletedHomeworks = completedHomeworks

		results = append(results, statistics)
	}

	return results, nil
}

func (r *dashboardReportCoursesRepository) SaveCourseStatistics(statistics []models.DashboardReportCourses) error {
	// Xóa các báo cáo cũ trong khoảng thời gian này trước khi lưu mới
	if len(statistics) > 0 {
		startDate := statistics[0].StartDate
		endDate := statistics[0].EndDate

		if err := db.MasterDB.Where("start_date = ? AND end_date = ?", startDate, endDate).
			Delete(&models.DashboardReportCourses{}).Error; err != nil {
			return err
		}
	}

	// Batch save với batch size 100
	batchSize := 100
	totalRecords := len(statistics)
	
	fmt.Printf("💾 Bắt đầu lưu %d records với batch size %d\n", totalRecords, batchSize)
	
	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}
		
		batch := statistics[i:end]
		fmt.Printf("💾 Lưu batch %d-%d/%d (%.1f%%)\n", 
			i+1, end, totalRecords, float64(end)/float64(totalRecords)*100)
		
		if err := db.MasterDB.Create(&batch).Error; err != nil {
			fmt.Printf("❌ Lỗi lưu batch %d-%d: %v\n", i+1, end, err)
			return err
		}
		
		fmt.Printf("✅ Hoàn thành lưu batch %d-%d (%d records)\n", i+1, end, len(batch))
	}

	return nil
}

func (r *dashboardReportCoursesRepository) GetExistingReport(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error) {
	var report models.DashboardReportCourses
	err := db.ReplicaDB.Where("course_id = ? AND start_date = ? AND end_date = ?", courseID, startDate, endDate).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *dashboardReportCoursesRepository) DeleteExistingReport(courseID int64, startDate, endDate time.Time) error {
	return db.MasterDB.Where("course_id = ? AND start_date = ? AND end_date = ?", courseID, startDate, endDate).
		Delete(&models.DashboardReportCourses{}).Error
}
