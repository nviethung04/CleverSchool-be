package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/table_manager"
	"fmt"
	"strings"
	"sync"
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
	studentTime := time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC)
	if activeStudentTime != nil {
		studentTime = time.Unix(*activeStudentTime, 0)
	}
	teacherTime := time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC)
	if activeTeacherTime != nil {
		teacherTime = time.Unix(*activeTeacherTime, 0)
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	now := time.Now()

	unionStart := studentTime
	if teacherTime.Before(unionStart) {
		unionStart = teacherTime
	}
	if sevenDaysAgo.Before(unionStart) {
		unionStart = sevenDaysAgo
	}
	allActivityLogs, err := r.createActivityLogsUnionQuery(unionStart, now)
	if err != nil {
		return nil, err
	}

	var (
		totalSchools                int64
		totalClasses                int64
		totalUsers                  int64
		studentActive               int64
		teacherActive               int64
		studentActiveWeekly         int64
		teacherActiveWeekly         int64
		frequentLoginStudents       int64
		frequentLoginStudentsWeekly int64
	)
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}

	wg := sync.WaitGroup{}

	wg.Add(3)
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Table("schools").Where("deleted_at IS NULL").Count(&totalSchools).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Table("classes").Where("deleted_at IS NULL").Count(&totalClasses).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Table("users").Where("deleted_at IS NULL").Count(&totalUsers).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	countDistinctSQL := fmt.Sprintf(`
		SELECT COUNT(DISTINCT user_id) FROM %s WHERE role_id = ? AND created_at > ?
	`, allActivityLogs)
	frequentSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT user_id FROM %s
			WHERE role_id = ? AND path = ? AND status_code = ? AND created_at > ?
			GROUP BY user_id HAVING COUNT(*) > 2
		) AS t
	`, allActivityLogs)

	wg.Add(6)
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(countDistinctSQL, 3, studentTime).Scan(&studentActive).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(countDistinctSQL, 2, teacherTime).Scan(&teacherActive).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(countDistinctSQL, 3, sevenDaysAgo).Scan(&studentActiveWeekly).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(countDistinctSQL, 2, sevenDaysAgo).Scan(&teacherActiveWeekly).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(frequentSQL, 3, "/api/login", 200, studentTime).Scan(&frequentLoginStudents).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(db.ReplicaDB.Raw(frequentSQL, 3, "/api/login", 200, sevenDaysAgo).Scan(&frequentLoginStudentsWeekly).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
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

