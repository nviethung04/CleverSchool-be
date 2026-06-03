package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/table_manager"
	"fmt"
	"strings"
	"time"
)

type DashboardReportSchoolsRepository interface {
	CalculateSchoolStatistics(startDate, endDate time.Time) ([]models.DashboardReportSchoolWeeks, error)
	SaveSchoolStatistics(statistics []models.DashboardReportSchoolWeeks) error
	GetExistingReport(schoolID int64, startDate, endDate time.Time) (*models.DashboardReportSchoolWeeks, error)
	DeleteExistingReport(schoolID int64, startDate, endDate time.Time) error
}

type dashboardReportSchoolsRepository struct{}

func NewDashboardReportSchoolsRepository() DashboardReportSchoolsRepository {
	return &dashboardReportSchoolsRepository{}
}

// createActivityLogsUnionQuery tạo UNION query cho các bảng activity_logs theo tháng
func (r *dashboardReportSchoolsRepository) createActivityLogsUnionQuery(startTime, endTime time.Time) (string, error) {
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

func (r *dashboardReportSchoolsRepository) CalculateSchoolStatistics(startDate, endDate time.Time) ([]models.DashboardReportSchoolWeeks, error) {
	// Lấy danh sách tất cả trường học
	var schools []models.School
	if err := db.ReplicaDB.Where("deleted_at IS NULL").Find(&schools).Error; err != nil {
		return nil, err
	}

	var results []models.DashboardReportSchoolWeeks

	// Tạo endDate đến cuối ngày (23:59:59)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// Tạo UNION query cho activity_logs
	allActivityLogs, err := r.createActivityLogsUnionQuery(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Batch processing với batch size 100
	batchSize := 100
	totalSchools := len(schools)
	
	fmt.Printf("🔄 Bắt đầu xử lý %d trường học với batch size %d\n", totalSchools, batchSize)
	
	for i := 0; i < totalSchools; i += batchSize {
		end := i + batchSize
		if end > totalSchools {
			end = totalSchools
		}
		
		batch := schools[i:end]
		fmt.Printf("📦 Xử lý batch %d-%d/%d (%.1f%%)\n", 
			i+1, end, totalSchools, float64(end)/float64(totalSchools)*100)
		
		// Xử lý batch
		batchResults, err := r.processSchoolBatch(batch, startDate, endDate, endOfDay, allActivityLogs)
		if err != nil {
			fmt.Printf("❌ Lỗi batch %d-%d: %v\n", i+1, end, err)
			continue
		}
		
		results = append(results, batchResults...)
		fmt.Printf("✅ Hoàn thành batch %d-%d (%d records)\n", i+1, end, len(batchResults))
	}

	return results, nil
}

func (r *dashboardReportSchoolsRepository) processSchoolBatch(schools []models.School, startDate, endDate, endOfDay time.Time, allActivityLogs string) ([]models.DashboardReportSchoolWeeks, error) {
	var results []models.DashboardReportSchoolWeeks
	
	for _, school := range schools {
		statistics := models.DashboardReportSchoolWeeks{
			SchoolID:  school.ID,
			StartDate: startDate,
			EndDate:   endDate,
		}

		// 1. Đếm tổng số học sinh (role = 3) của trường được tạo trước cuối ngày end_date
		var totalStudents int64
		studentQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			WHERE u.school_id = ? AND u.deleted_at IS NULL AND urr.role_id = 3 AND u.created_at <= ?
		`
		if err := db.ReplicaDB.Raw(studentQuery, school.ID, endOfDay).Scan(&totalStudents).Error; err != nil {
			return nil, err
		}
		statistics.TotalStudents = totalStudents

		// 2. Đếm tổng số giáo viên (role = 2) của trường được tạo trước cuối ngày end_date
		var totalTeachers int64
		teacherQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN courses c ON uc.course_id = c.id
			INNER JOIN course_schools cs ON c.id = cs.course_id
			WHERE cs.school_id = ? AND u.deleted_at IS NULL AND c.deleted_at IS NULL
			AND u.id IN (SELECT user_id FROM user_ref_roles WHERE role_id = 2)
			AND u.created_at <= ?
		`
		if err := db.ReplicaDB.Raw(teacherQuery, school.ID, endOfDay).Scan(&totalTeachers).Error; err != nil {
			return nil, err
		}
		statistics.TotalTeachers = totalTeachers

		// 3. Đếm số học sinh active trong thời gian (role_id = 3)
		var activeStudents int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
			studentQuery := fmt.Sprintf(`
				SELECT COUNT(DISTINCT all_activity_logs.user_id)
				FROM %s
				INNER JOIN users u ON all_activity_logs.user_id = u.id
				WHERE u.school_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(studentQuery, school.ID, 3, startDate, endDate).Scan(&activeStudents).Error; err != nil {
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
				INNER JOIN courses c ON uc.course_id = c.id
				INNER JOIN course_schools cs ON c.id = cs.course_id
				WHERE cs.school_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(teacherQuery, school.ID, 2, startDate, endDate).Scan(&activeTeachers).Error; err != nil {
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
			INNER JOIN homeworks h ON hu.homework_id = h.id
			WHERE u.school_id = ? AND u.deleted_at IS NULL 
			AND hu.questions_completed >= h.total_questions
			AND hu.created_at >= ? AND hu.created_at <= ?
		`

		if err := db.ReplicaDB.Raw(homeworkQuery, school.ID, startDate, endDate).Scan(&studentsCompletedHomework).Error; err != nil {
			return nil, err
		}
		statistics.StudentsCompletedHomework = studentsCompletedHomework

		results = append(results, statistics)
	}

	return results, nil
}

func (r *dashboardReportSchoolsRepository) SaveSchoolStatistics(statistics []models.DashboardReportSchoolWeeks) error {
	// Xóa các báo cáo cũ trong khoảng thời gian này trước khi lưu mới
	if len(statistics) > 0 {
		startDate := statistics[0].StartDate
		endDate := statistics[0].EndDate

		if err := db.MasterDB.Where("start_date = ? AND end_date = ?", startDate, endDate).
			Delete(&models.DashboardReportSchoolWeeks{}).Error; err != nil {
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

func (r *dashboardReportSchoolsRepository) GetExistingReport(schoolID int64, startDate, endDate time.Time) (*models.DashboardReportSchoolWeeks, error) {
	var report models.DashboardReportSchoolWeeks
	err := db.ReplicaDB.Where("school_id = ? AND start_date = ? AND end_date = ?", schoolID, startDate, endDate).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *dashboardReportSchoolsRepository) DeleteExistingReport(schoolID int64, startDate, endDate time.Time) error {
	return db.MasterDB.Where("school_id = ? AND start_date = ? AND end_date = ?", schoolID, startDate, endDate).
		Delete(&models.DashboardReportSchoolWeeks{}).Error
}
