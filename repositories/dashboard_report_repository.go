package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/table_manager"
	"fmt"
	"strings"
	"time"
)

type DashboardReportRepository interface {
	GetDashboardReport(activeStudentTime, activeTeacherTime *int64) (*dto.DashboardReportDTO, error)
}

type dashboardReportRepository struct{}

func NewDashboardReportRepository() DashboardReportRepository {
	return &dashboardReportRepository{}
}

// createActivityLogsUnionQuery tạo UNION query cho các bảng activity_logs theo tháng
func (r *dashboardReportRepository) createActivityLogsUnionQuery(startTime, endTime time.Time) (string, error) {
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

func (r *dashboardReportRepository) GetDashboardReport(activeStudentTime, activeTeacherTime *int64) (*dto.DashboardReportDTO, error) {
	// Đếm số trường từ bảng schools (chỉ lấy những trường chưa bị xóa)
	var totalSchools int64
	if err := db.ReplicaDB.Table("schools").
		Where("deleted_at IS NULL").
		Count(&totalSchools).Error; err != nil {
		return nil, err
	}

	// Đếm số lớp từ bảng classes (chỉ lấy những lớp chưa bị xóa)
	var totalClasses int64
	if err := db.ReplicaDB.Table("classes").
		Where("deleted_at IS NULL").
		Count(&totalClasses).Error; err != nil {
		return nil, err
	}

	// Đếm số người dùng từ bảng users (chỉ lấy những user chưa bị xóa)
	var totalUsers int64
	if err := db.ReplicaDB.Table("users").
		Where("deleted_at IS NULL").
		Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// Xử lý thời gian mặc định cho student active
	var studentTime time.Time
	if activeStudentTime != nil {
		studentTime = time.Unix(*activeStudentTime, 0)
	} else {
		// Mặc định: 15 tháng 9 năm 2025
		studentTime = time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC)
	}

	// Xử lý thời gian mặc định cho teacher active
	var teacherTime time.Time
	if activeTeacherTime != nil {
		teacherTime = time.Unix(*activeTeacherTime, 0)
	} else {
		// Mặc định: 8 tháng 9 năm 2025
		teacherTime = time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC)
	}

	// Đếm số học sinh hoạt động (role_id = 3)
	var studentActive int64
	allActivityLogs, err := r.createActivityLogsUnionQuery(studentTime, time.Now())
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(DISTINCT user_id)
		FROM %s
		WHERE role_id = ? AND created_at > ?
	`, allActivityLogs)

	if err := db.ReplicaDB.Raw(query, 3, studentTime).Scan(&studentActive).Error; err != nil {
		return nil, err
	}

	// Đếm số giáo viên hoạt động (role_id = 2)
	var teacherActive int64
	allActivityLogsTeacher, err := r.createActivityLogsUnionQuery(teacherTime, time.Now())
	if err != nil {
		return nil, err
	}

	teacherQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT user_id)
		FROM %s
		WHERE role_id = ? AND created_at > ?
	`, allActivityLogsTeacher)

	if err := db.ReplicaDB.Raw(teacherQuery, 2, teacherTime).Scan(&teacherActive).Error; err != nil {
		return nil, err
	}

	// Tính thời gian 7 ngày trước hiện tại
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// Đếm số học sinh hoạt động trong 7 ngày gần nhất (role_id = 3)
	var studentActiveWeekly int64
	allActivityLogsWeekly, err := r.createActivityLogsUnionQuery(sevenDaysAgo, time.Now())
	if err != nil {
		return nil, err
	}

	weeklyQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT user_id)
		FROM %s
		WHERE role_id = ? AND created_at > ?
	`, allActivityLogsWeekly)

	if err := db.ReplicaDB.Raw(weeklyQuery, 3, sevenDaysAgo).Scan(&studentActiveWeekly).Error; err != nil {
		return nil, err
	}

	// Đếm số giáo viên hoạt động trong 7 ngày gần nhất (role_id = 2)
	var teacherActiveWeekly int64
	teacherWeeklyQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT user_id)
		FROM %s
		WHERE role_id = ? AND created_at > ?
	`, allActivityLogsWeekly)

	if err := db.ReplicaDB.Raw(teacherWeeklyQuery, 2, sevenDaysAgo).Scan(&teacherActiveWeekly).Error; err != nil {
		return nil, err
	}

	// Đếm số học sinh login thường xuyên (role_id = 3, path = '/api/login', status_code = 200, xuất hiện > 2 lần, created_at > active_student_time)
	var frequentLoginStudents int64
	frequentQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT user_id 
			FROM %s 
			WHERE role_id = ? AND path = ? AND status_code = ? AND created_at > ?
			GROUP BY user_id 
			HAVING COUNT(*) > 2
		) AS frequent_users
	`, allActivityLogs)

	if err := db.ReplicaDB.Raw(frequentQuery, 3, "/api/login", 200, studentTime).Scan(&frequentLoginStudents).Error; err != nil {
		return nil, err
	}

	// Đếm số học sinh login thường xuyên trong 7 ngày gần nhất (role_id = 3, path = '/api/login', status_code = 200, xuất hiện > 2 lần, created_at > sevenDaysAgo)
	var frequentLoginStudentsWeekly int64
	frequentWeeklyQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT user_id 
			FROM %s 
			WHERE role_id = ? AND path = ? AND status_code = ? AND created_at > ?
			GROUP BY user_id 
			HAVING COUNT(*) > 2
		) AS frequent_users_weekly
	`, allActivityLogsWeekly)

	if err := db.ReplicaDB.Raw(frequentWeeklyQuery, 3, "/api/login", 200, sevenDaysAgo).Scan(&frequentLoginStudentsWeekly).Error; err != nil {
		return nil, err
	}

	return &dto.DashboardReportDTO{
		TotalSchools:                totalSchools,
		TotalClasses:                totalClasses,
		TotalUsers:                  totalUsers,
		StudentActive:               studentActive,
		TeacherActive:               teacherActive,
		StudentActiveWeekly:         studentActiveWeekly,
		TeacherActiveWeekly:         teacherActiveWeekly,
		FrequentLoginStudents:       frequentLoginStudents,
		FrequentLoginStudentsWeekly: frequentLoginStudentsWeekly,
	}, nil
}
