package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/table_manager"
	"encoding/json"
	"fmt"
	"strings"

	"time"

	"gorm.io/gorm/clause"
)

type HistoryUseRepository interface {
	SaveWeeklyUsage(targetDate time.Time) error
	SaveMonthlyUsage(targetDate time.Time) error
	AverageUsed(start, end time.Time, schoolId, courseId, programId, roleId int64) (models.AverageUsed, error)
	ActivityIds(start, end time.Time) (models.ActivityIds, error)
}

type historyUseRepository struct{}

func NewHistoryUseRepository() HistoryUseRepository {
	return &historyUseRepository{}
}

func (r *historyUseRepository) SaveWeeklyUsage(targetDate time.Time) error {
	// Lấy ngày đầu tuần (thứ 2)
	weekday := int(targetDate.Weekday())
	if weekday == 0 {
		weekday = 7 // Chủ nhật = 0 -> 7
	}
	weekStart := targetDate.AddDate(0, 0, -(weekday - 1))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())

	weekEnd := weekStart.AddDate(0, 0, 7)

	// Tính thời gian sử dụng theo vai trò
	type RoleDuration struct {
		RoleID   int
		Duration float64
	}
	var roleDurations []RoleDuration

	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", weekStart, weekEnd)

	var unionQueries []string
	var args []interface{}

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

		subQuery := fmt.Sprintf(`
			SELECT
				r.id AS role_id,
				COUNT(DISTINCT FLOOR(EXTRACT(EPOCH FROM %s.created_at) / 300)) * 5 AS duration
			FROM %s
			JOIN users u ON u.id = %s.user_id
			JOIN user_ref_roles urr ON urr.user_id = u.id
			JOIN roles r ON r.id = urr.role_id
			WHERE %s.created_at >= ? AND %s.created_at < ?
			GROUP BY r.id
		`, tableName, tableName, tableName, tableName, tableName)

		unionQueries = append(unionQueries, subQuery)
		args = append(args, weekStart, weekEnd)
	}

	if len(unionQueries) > 0 {
		mainQuery := fmt.Sprintf(`
			SELECT role_id, SUM(duration) as duration
			FROM (
				%s
			) AS combined
			GROUP BY role_id
		`, strings.Join(unionQueries, " UNION ALL "))

		err := db.ReplicaDB.Raw(mainQuery, args...).Scan(&roleDurations).Error
		if err != nil {
			return fmt.Errorf("cannot calculate weekly duration by role: %w", err)
		}
	}

	usageTime := models.UsageTime{}
	for _, rd := range roleDurations {
		switch rd.RoleID {
		case models.AdminRoleId:
			usageTime.Admin = int64(rd.Duration)
		case models.TeacherRoleId:
			usageTime.Teacher = int64(rd.Duration)
		case models.StudentRoleId:
			usageTime.Student = int64(rd.Duration)
		}
	}
	usageTime.All = usageTime.Admin + usageTime.Teacher + usageTime.Student

	usageTimeJSON, _ := json.Marshal(usageTime)

	// Tính device usage
	type DeviceResult struct {
		Device string
		Count  int64
	}
	var deviceResults []DeviceResult

	// Tạo UNION query cho device từ các bảng theo tháng
	var deviceUnionQueries []string
	var deviceArgs []interface{}

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

		deviceSubQuery := fmt.Sprintf(`
			SELECT
				device,
				COUNT(DISTINCT user_id) AS count
			FROM %s
			WHERE device IS NOT NULL
			AND created_at >= ?
			AND created_at < ?
			GROUP BY device
		`, tableName)

		deviceUnionQueries = append(deviceUnionQueries, deviceSubQuery)
		deviceArgs = append(deviceArgs, weekStart, weekEnd)
	}

	if len(deviceUnionQueries) > 0 {
		deviceMainQuery := fmt.Sprintf(`
			SELECT device, SUM(count) as count
			FROM (
				%s
			) AS combined
			GROUP BY device
		`, strings.Join(deviceUnionQueries, " UNION ALL "))

		err := db.ReplicaDB.Raw(deviceMainQuery, deviceArgs...).Scan(&deviceResults).Error
		if err != nil {
			return fmt.Errorf("cannot calculate weekly device usage: %w", err)
		}
	}

	deviceMap := map[string]int64{}
	for _, d := range deviceResults {
		deviceMap[d.Device] = d.Count
	}
	deviceJSON, _ := json.Marshal(deviceMap)

	averageUsed, _ := r.AverageUsed(weekStart, weekEnd, 0, 0, 0, 0)
	averageUsedJSON, _ := json.Marshal(averageUsed)

	activityIds, _ := r.ActivityIds(weekStart, weekEnd)
	activityIdsJSON, _ := json.Marshal(activityIds)

	// Upsert
	history := &models.HistoryUse{
		Type:        "week",
		Date:        weekStart,
		UsageTime:   usageTimeJSON,
		Device:      deviceJSON,
		AverageUsed: averageUsedJSON,
		ActivityIds: activityIdsJSON,
	}

	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"usage_time", "device", "average_used", "activity_ids", "updated_at"}),
	}).Create(history).Error
}
func (r *historyUseRepository) SaveMonthlyUsage(targetDate time.Time) error {
	// Ngày đầu tháng
	monthStart := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	// Tính thời gian sử dụng theo vai trò
	type RoleDuration struct {
		RoleID   int
		Duration float64
	}
	var roleDurations []RoleDuration

	// Lấy tên bảng của tháng này
	tableName := table_manager.GetMonthlyTableName("activity_logs", monthStart)

	// Kiểm tra bảng có tồn tại không
	var exists bool
	err := db.ReplicaDB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = ?
		)
	`, tableName).Scan(&exists).Error

	if err != nil {
		return fmt.Errorf("cannot check table existence: %w", err)
	}

	if exists {
		query := fmt.Sprintf(`
			SELECT
				urr.role_id AS role_id,
				COUNT(DISTINCT FLOOR(EXTRACT(EPOCH FROM %s.created_at) / 300)) * 5 AS duration
			FROM %s
			JOIN users u ON u.id = %s.user_id
			JOIN user_ref_roles urr ON urr.user_id = u.id
			WHERE %s.created_at >= ? AND %s.created_at < ?
			GROUP BY urr.role_id
		`, tableName, tableName, tableName, tableName, tableName)

		err = db.ReplicaDB.Raw(query, monthStart, monthEnd).Scan(&roleDurations).Error
		if err != nil {
			return fmt.Errorf("cannot calculate monthly duration by role: %w", err)
		}
	}

	usageTime := models.UsageTime{}
	for _, rd := range roleDurations {
		switch rd.RoleID {
		case models.AdminRoleId:
			usageTime.Admin = int64(rd.Duration)
		case models.TeacherRoleId:
			usageTime.Teacher = int64(rd.Duration)
		case models.StudentRoleId:
			usageTime.Student = int64(rd.Duration)
		}
	}
	usageTime.All = usageTime.Admin + usageTime.Teacher + usageTime.Student

	usageTimeJSON, _ := json.Marshal(usageTime)

	// Tính device usage
	type DeviceResult struct {
		Device string
		Count  int64
	}
	var deviceResults []DeviceResult

	if exists {
		deviceQuery := fmt.Sprintf(`
			SELECT
				device,
				COUNT(DISTINCT user_id) AS count
			FROM %s
			WHERE device IS NOT NULL
			AND created_at >= ?
			AND created_at < ?
			GROUP BY device
		`, tableName)

		err = db.ReplicaDB.Raw(deviceQuery, monthStart, monthEnd).Scan(&deviceResults).Error
		if err != nil {
			return fmt.Errorf("cannot calculate monthly device usage: %w", err)
		}
	}

	deviceMap := map[string]int64{}
	for _, d := range deviceResults {
		deviceMap[d.Device] = d.Count
	}
	deviceJSON, _ := json.Marshal(deviceMap)

	averageUsed, _ := r.AverageUsed(monthStart, monthEnd, 0, 0, 0, 0)
	averageUsedJSON, _ := json.Marshal(averageUsed)

	activityIds, _ := r.ActivityIds(monthStart, monthEnd)
	activityIdsJSON, _ := json.Marshal(activityIds)
	// Upsert
	history := &models.HistoryUse{
		Type:        "month",
		Date:        monthStart,
		UsageTime:   usageTimeJSON,
		Device:      deviceJSON,
		AverageUsed: averageUsedJSON,
		ActivityIds: activityIdsJSON,
	}

	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"usage_time", "device", "average_used", "activity_ids", "updated_at"}),
	}).Create(history).Error
}

func (r *historyUseRepository) AverageUsed(start, end time.Time, schoolId, courseId, programId, roleId int64) (models.AverageUsed, error) {
	var averageUsed models.AverageUsed

	joinUser1 := ""
	whereUser1 := ""

	joinUser2 := ""
	whereUser2 := ""

	// --- handle school filter ---
	if schoolId > 0 {
		whereUser1 += fmt.Sprintf(" AND u.school_id = %d", schoolId)
		whereUser2 += fmt.Sprintf(" AND u2.school_id = %d", schoolId)
	}

	// --- handle course/program filter depending on which is used ---
	if courseId > 0 || programId > 0 {
		if !strings.Contains(joinUser1, "user_courses") {
			joinUser1 += " JOIN user_courses uc ON uc.user_id = al.user_id "
			joinUser2 += " JOIN user_courses uc ON uc.user_id = al.user_id "
		}
		if courseId > 0 {
			whereUser1 += fmt.Sprintf(" AND uc.course_id = %d", courseId)
			whereUser2 += fmt.Sprintf(" AND uc.course_id = %d", courseId)
		}
		if programId > 0 {
			if !strings.Contains(joinUser1, "courses") {
				joinUser1 += " JOIN courses co ON co.id = uc.course_id "
				joinUser2 += " JOIN courses co ON co.id = uc.course_id "
			}
			whereUser1 += fmt.Sprintf(" AND co.program_id = %d", programId)
			whereUser2 += fmt.Sprintf(" AND co.program_id = %d", programId)
		}
	}

	// --- handle role filter ---
	if roleId > 0 {
		if !strings.Contains(joinUser1, "user_ref_roles") {
			joinUser1 += " JOIN user_ref_roles urr1 ON urr1.user_id = u.id "
		}
		whereUser1 += fmt.Sprintf(" AND urr1.role_id = %d", roleId)

		if !strings.Contains(joinUser2, "user_ref_roles") {
			joinUser2 += " JOIN user_ref_roles urr2 ON urr2.user_id = u2.id "
		}
		whereUser2 += fmt.Sprintf(" AND urr2.role_id = %d", roleId)
	}

	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", start, end)

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

		unionParts = append(unionParts, fmt.Sprintf("SELECT user_id, created_at FROM %s", tableName))
	}

	// Nếu không có bảng nào tồn tại, trả về giá trị mặc định
	if len(unionParts) == 0 {
		return averageUsed, nil
	}

	// Tạo virtual table từ UNION các bảng tháng
	allActivityLogs := fmt.Sprintf("(%s)", strings.Join(unionParts, " UNION ALL "))

	query := fmt.Sprintf(`
		WITH
		all_logs AS %s,
		user_minutes AS (
			SELECT al.user_id, COUNT(*)/15.0 AS total_minutes
			FROM all_logs al
			JOIN users u ON u.id = al.user_id
			%s
			WHERE al.created_at >= ? AND al.created_at < ?
			%s
			GROUP BY al.user_id
		),
		daily AS (
			SELECT date_trunc('day', al.created_at) AS day,
				COUNT(DISTINCT al.user_id) AS dau
			FROM all_logs al
			JOIN users u ON u.id = al.user_id
			%s
			WHERE al.created_at >= ? AND al.created_at < ?
			%s
			GROUP BY date_trunc('day', al.created_at)
		),
		user_days AS (
			SELECT al.user_id, COUNT(DISTINCT al.created_at::date) AS active_days
			FROM all_logs al
			JOIN users u ON u.id = al.user_id
			%s
			WHERE al.created_at >= ? AND al.created_at < ?
			%s
			GROUP BY al.user_id
		),
		counts AS (
			SELECT
				COUNT(DISTINCT al.user_id) AS total,
				COUNT(DISTINCT CASE WHEN user_log_count.log_count > 150 THEN al.user_id END) AS engaged
			FROM all_logs al
			JOIN (
				SELECT user_id, COUNT(*) AS log_count
				FROM all_logs
				GROUP BY user_id
			) AS user_log_count ON user_log_count.user_id = al.user_id
			LEFT JOIN users u2 ON u2.id = al.user_id
			LEFT JOIN user_ref_roles urr ON urr.user_id = u2.id
			%s
			WHERE al.created_at >= ? AND al.created_at < ?
			AND urr.role_id IN (2,3)
			%s
		)
		SELECT
			(SELECT AVG(total_minutes)::float FROM user_minutes) AS minutes_per_session,
			(SELECT AVG(dau)::float FROM daily) AS daily_active_users,
			(SELECT COALESCE(SUM(CASE WHEN active_days>1 THEN 1 END),0)::float /
					NULLIF(COUNT(*),0) FROM user_days) AS returning_rate,
			(SELECT engaged::float/NULLIF(total,0) FROM counts) AS engagement_rate;
	`,
		allActivityLogs,
		joinUser1, whereUser1,
		joinUser1, whereUser1,
		joinUser1, whereUser1,
		joinUser2, whereUser2,
	)

	err := db.ReplicaDB.Raw(query,
		start, end,
		start, end,
		start, end,
		start, end,
	).Scan(&averageUsed).Error

	return averageUsed, err
}

func (r *historyUseRepository) ActivityIds(start, end time.Time) (models.ActivityIds, error) {
	var ids []int64

	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", start, end)

	var unionQueries []string
	var args []interface{}

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

		subQuery := fmt.Sprintf(`
			SELECT DISTINCT user_id
			FROM %s
			WHERE user_id IS NOT NULL AND created_at >= ? AND created_at <= ?
		`, tableName)

		unionQueries = append(unionQueries, subQuery)
		args = append(args, start, end)
	}

	if len(unionQueries) > 0 {
		mainQuery := fmt.Sprintf(`
			SELECT DISTINCT user_id
			FROM (
				%s
			) AS combined
		`, strings.Join(unionQueries, " UNION ALL "))

		err := db.ReplicaDB.Raw(mainQuery, args...).Pluck("user_id", &ids).Error
		if err != nil {
			config.Log.Error("Failed to get activity ids", "error", err)
			return models.ActivityIds{}, err
		}
	}

	return models.ActivityIds{Ids: ids}, nil
}
