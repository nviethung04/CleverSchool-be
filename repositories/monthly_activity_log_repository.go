package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"be-lms/table_manager"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MonthlyActivityLogRepository interface {
	Create(activityLog models.ActivityLog) error
	CreateWithTable(activityLog models.ActivityLog, tableName string) error
	EnsureTableExists(tableName string) error
	GetActiveUsersCount(minutes int) (int64, error)
	GetActiveUsersCountByTimeRange(startTime, endTime time.Time) (int64, error)
	GetActiveUsersWithPaging(req *requests.GetActiveUsersRequest, c *gin.Context) ([]models.User, int64, error)
	GetActiveUsersByTimeRange(startTime, endTime time.Time, page, limit int) ([]uint, int64, error)
	GetUserByID(userID uint) (*models.User, error)
	GetUserLastActivityInTimeRange(userID uint, startTime, endTime time.Time) (*time.Time, error)
	GetUsersLastActivityTime(userIDs []uint) (map[uint]time.Time, error)
	GetFailedLoginsCount(req *requests.GetFailedLoginRequest) (int64, error)
	GetFailedLoginsCountByTimeRange(startTime, endTime time.Time, userID int64) (int64, error)
	GetFailedLoginsWithPaging(req *requests.GetFailedLoginRequest) ([]models.ActivityLog, int64, error)
	GetFailedLoginsWithPagingByTimeRange(startTime, endTime time.Time, userID int64, page, limit int) ([]models.ActivityLog, int64, error)
	PermanentlyDeleteOldRecords(now time.Time) error
}

type monthlyActivityLogRepository struct{}

func NewMonthlyActivityLogRepository() MonthlyActivityLogRepository {
	return &monthlyActivityLogRepository{}
}

// Create tạo log entry vào bảng của tháng hiện tại
func (r *monthlyActivityLogRepository) Create(activityLog models.ActivityLog) error {
	tableName := table_manager.GetCurrentMonthTableName("activity_logs")
	return r.CreateWithTable(activityLog, tableName)
}

// CreateWithTable tạo log entry vào bảng cụ thể
func (r *monthlyActivityLogRepository) CreateWithTable(activityLog models.ActivityLog, tableName string) error {
	// Đảm bảo bảng tồn tại
	if err := r.EnsureTableExists(tableName); err != nil {
		return err
	}

	// Detect device
	activityLog.Device = detectDevice(*activityLog.Agent)

	// Sử dụng raw SQL để insert vào bảng cụ thể
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, role_id, method, path, status_code, client_ip, device, latency_ms,
			query_params, request_body, response_body, attribute, session_id, agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)

	return db.MasterDB.Exec(sql,
		activityLog.UserID,
		activityLog.RoleID,
		activityLog.Method,
		activityLog.Path,
		activityLog.StatusCode,
		activityLog.ClientIP,
		activityLog.Device,
		activityLog.LatencyMs,
		activityLog.QueryParams,
		activityLog.RequestBody,
		activityLog.ResponseBody,
		activityLog.Attribute,
		activityLog.SessionID,
		activityLog.Agent,
		activityLog.CreatedAt,
	).Error
}

// EnsureTableExists đảm bảo bảng tồn tại, nếu không thì tạo mới
func (r *monthlyActivityLogRepository) EnsureTableExists(tableName string) error {
	// Kiểm tra xem bảng đã tồn tại chưa
	var exists bool
	err := db.MasterDB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = ?
		)
	`, tableName).Scan(&exists).Error

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Tạo bảng mới với cấu trúc giống bảng gốc
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			user_id BIGINT,
			role_id INTEGER,
			method VARCHAR(10) NOT NULL,
			path TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			client_ip VARCHAR(45),
			device VARCHAR(20),
			latency_ms INTEGER,
			query_params JSONB,
			request_body JSONB,
			response_body JSONB,
			attribute JSONB,
			session_id TEXT,
			agent TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`, tableName)

	if err := db.MasterDB.Exec(createTableSQL).Error; err != nil {
		return err
	}

	// Tạo indexes
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at DESC)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_user_id ON %s(user_id)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_user_created_at ON %s(user_id, created_at DESC)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_device ON %s(device)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_device_created_at ON %s(device, created_at DESC)", tableName, tableName),
	}

	for _, indexSQL := range indexes {
		if err := db.MasterDB.Exec(indexSQL).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetActiveUsersCount đếm số lượng users có hoạt động trong X phút
func (r *monthlyActivityLogRepository) GetActiveUsersCount(minutes int) (int64, error) {
	timeThreshold := time.Now().Add(time.Duration(-minutes) * time.Minute)

	// Lấy danh sách bảng cần query (tháng hiện tại và có thể tháng trước)
	tableNames := table_manager.GetTableNamesForLastMonths("activity_logs", 2)

	var totalCount int64

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

		var count int64
		err = db.ReplicaDB.Raw(fmt.Sprintf(`
			SELECT COUNT(DISTINCT user_id)
			FROM %s
			WHERE user_id IS NOT NULL AND created_at > ?
		`, tableName), timeThreshold).Scan(&count).Error

		if err != nil {
			continue
		}

		totalCount += count
	}

	return totalCount, nil
}

// GetActiveUsersCountByTimeRange đếm số lượng users có hoạt động trong khoảng thời gian cụ thể
func (r *monthlyActivityLogRepository) GetActiveUsersCountByTimeRange(startTime, endTime time.Time) (int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	var totalCount int64

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

		var count int64
		err = db.ReplicaDB.Raw(fmt.Sprintf(`
			SELECT COUNT(DISTINCT user_id)
			FROM %s
			WHERE user_id IS NOT NULL AND created_at >= ? AND created_at <= ?
		`, tableName), startTime, endTime).Scan(&count).Error

		if err != nil {
			continue
		}

		totalCount += count
	}

	return totalCount, nil
}

// GetActiveUsersWithPaging lấy danh sách users có hoạt động với phân trang
func (r *monthlyActivityLogRepository) GetActiveUsersWithPaging(req *requests.GetActiveUsersRequest, c *gin.Context) ([]models.User, int64, error) {
	timeThreshold := time.Now().Add(time.Duration(-req.Minutes) * time.Minute)

	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForLastMonths("activity_logs", 2)

	// Tạo UNION query cho tất cả các bảng
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

		unionQueries = append(unionQueries, fmt.Sprintf(`
			SELECT DISTINCT user_id, MAX(created_at) as last_activity
			FROM %s
			WHERE user_id IS NOT NULL AND created_at > ?
			GROUP BY user_id
		`, tableName))
		args = append(args, timeThreshold)
	}

	if len(unionQueries) == 0 {
		return []models.User{}, 0, nil
	}

	// Tạo query chính với UNION
	mainQuery := strings.Join(unionQueries, " UNION ALL ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT DISTINCT user_id FROM (%s) as combined
		) as distinct_users
	`, mainQuery)

	err := db.ReplicaDB.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	usersQuery := fmt.Sprintf(`
		SELECT u.* FROM users u
		INNER JOIN (
			SELECT DISTINCT user_id FROM (%s) as combined
			ORDER BY user_id
			LIMIT ? OFFSET ?
		) as active_users ON u.id = active_users.user_id
		ORDER BY u.id ASC
	`, mainQuery)

	var users []models.User
	offset := (req.Page - 1) * req.Limit
	allArgs := append(args, req.Limit, offset)

	err = db.ReplicaDB.Raw(usersQuery, allArgs...).Scan(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetActiveUsersByTimeRange lấy danh sách user IDs có hoạt động trong khoảng thời gian cụ thể với phân trang
func (r *monthlyActivityLogRepository) GetActiveUsersByTimeRange(startTime, endTime time.Time, page, limit int) ([]uint, int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	// Tạo UNION query cho tất cả các bảng
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

		unionQueries = append(unionQueries, fmt.Sprintf(`
			SELECT DISTINCT user_id
			FROM %s
			WHERE user_id IS NOT NULL AND created_at BETWEEN ? AND ?
		`, tableName))
		args = append(args, startTime, endTime)
	}

	if len(unionQueries) == 0 {
		return []uint{}, 0, nil
	}

	// Tạo query chính với UNION
	mainQuery := strings.Join(unionQueries, " UNION ALL ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT DISTINCT user_id FROM (%s) as combined
		) as distinct_users
	`, mainQuery)

	err := db.ReplicaDB.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get user IDs with pagination
	usersQuery := fmt.Sprintf(`
		SELECT DISTINCT user_id FROM (%s) as combined
		ORDER BY user_id
		LIMIT ? OFFSET ?
	`, mainQuery)

	var userIDs []uint
	offset := (page - 1) * limit
	allArgs := append(args, limit, offset)

	err = db.ReplicaDB.Raw(usersQuery, allArgs...).Scan(&userIDs).Error
	if err != nil {
		return nil, 0, err
	}

	return userIDs, total, nil
}

// GetUserByID lấy thông tin user theo ID
func (r *monthlyActivityLogRepository) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	err := db.ReplicaDB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserLastActivityInTimeRange lấy thời gian hoạt động cuối cùng của user trong khoảng thời gian
func (r *monthlyActivityLogRepository) GetUserLastActivityInTimeRange(userID uint, startTime, endTime time.Time) (*time.Time, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	var lastActivity time.Time

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

		var activity time.Time
		err = db.ReplicaDB.Raw(fmt.Sprintf(`
			SELECT MAX(created_at)
			FROM %s
			WHERE user_id = ? AND created_at BETWEEN ? AND ?
		`, tableName), userID, startTime, endTime).Scan(&activity).Error

		if err != nil {
			continue
		}

		if !activity.IsZero() && activity.After(lastActivity) {
			lastActivity = activity
		}
	}

	if lastActivity.IsZero() {
		return nil, gorm.ErrRecordNotFound
	}

	return &lastActivity, nil
}

// GetUsersLastActivityTime lấy thời gian hoạt động cuối cùng của các users
func (r *monthlyActivityLogRepository) GetUsersLastActivityTime(userIDs []uint) (map[uint]time.Time, error) {
	if len(userIDs) == 0 {
		return make(map[uint]time.Time), nil
	}

	// Lấy danh sách bảng cần query (6 tháng gần nhất)
	tableNames := table_manager.GetTableNamesForLastMonths("activity_logs", 6)

	// Tạo UNION query cho tất cả các bảng
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

		unionQueries = append(unionQueries, fmt.Sprintf(`
			SELECT user_id, MAX(created_at) as created_at
			FROM %s
			WHERE user_id = ANY(?)
			GROUP BY user_id
		`, tableName))
		args = append(args, userIDs)
	}

	if len(unionQueries) == 0 {
		return make(map[uint]time.Time), nil
	}

	// Tạo query chính với UNION
	mainQuery := strings.Join(unionQueries, " UNION ALL ")

	// Lấy kết quả cuối cùng
	finalQuery := fmt.Sprintf(`
		SELECT user_id, MAX(created_at) as created_at
		FROM (%s) as combined
		GROUP BY user_id
	`, mainQuery)

	type UserActivity struct {
		UserID    uint      `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
	}

	var activities []UserActivity
	err := db.ReplicaDB.Raw(finalQuery, args...).Scan(&activities).Error
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := make(map[uint]time.Time)
	for _, activity := range activities {
		result[activity.UserID] = activity.CreatedAt
	}

	return result, nil
}

// GetFailedLoginsCount đếm số lượng IP có failed login attempts
func (r *monthlyActivityLogRepository) GetFailedLoginsCount(req *requests.GetFailedLoginRequest) (int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForLastMonths("activity_logs", 2)

	var totalCount int64

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

		query := fmt.Sprintf(`
			SELECT COUNT(DISTINCT client_ip)
			FROM %s
			WHERE path = ? AND status_code != ? AND client_ip IS NOT NULL AND client_ip != ''
		`, tableName)

		var args []interface{}
		args = append(args, "/api/login", 200)

		// Áp dụng filter theo thời gian nếu có
		if req.Minutes > 0 {
			timeThreshold := time.Now().Add(time.Duration(-req.Minutes) * time.Minute)
			query += " AND created_at > ?"
			args = append(args, timeThreshold)
		}

		// Filter theo user_id nếu có
		if req.UserID > 0 {
			query += " AND user_id = ?"
			args = append(args, req.UserID)
		}

		var count int64
		err = db.ReplicaDB.Raw(query, args...).Scan(&count).Error
		if err != nil {
			continue
		}

		totalCount += count
	}

	return totalCount, nil
}

// GetFailedLoginsCountByTimeRange đếm số lượng IP có failed login attempts trong khoảng thời gian cụ thể
func (r *monthlyActivityLogRepository) GetFailedLoginsCountByTimeRange(startTime, endTime time.Time, userID int64) (int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	var totalCount int64

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

		query := fmt.Sprintf(`
			SELECT COUNT(DISTINCT client_ip)
			FROM %s
			WHERE path = ? AND status_code != ? AND client_ip IS NOT NULL AND client_ip != ''
			AND created_at >= ? AND created_at <= ?
		`, tableName)

		var args []interface{}
		args = append(args, "/api/login", 200, startTime, endTime)

		// Filter theo user_id nếu có
		if userID > 0 {
			query += " AND user_id = ?"
			args = append(args, userID)
		}

		var count int64
		err = db.ReplicaDB.Raw(query, args...).Scan(&count).Error
		if err != nil {
			continue
		}

		totalCount += count
	}

	return totalCount, nil
}

// GetFailedLoginsWithPaging lấy danh sách failed login attempts với phân trang
func (r *monthlyActivityLogRepository) GetFailedLoginsWithPaging(req *requests.GetFailedLoginRequest) ([]models.ActivityLog, int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForLastMonths("activity_logs", 2)

	// Tạo UNION query cho tất cả các bảng
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

		query := fmt.Sprintf(`
			SELECT id, user_id, role_id, method, path, status_code, client_ip, device,
				latency_ms, query_params, request_body, response_body, attribute,
				session_id, agent, created_at
			FROM %s
			WHERE path = ? AND status_code != ? AND client_ip IS NOT NULL AND client_ip != ''
		`, tableName)

		args = append(args, "/api/login", 200)

		// Áp dụng filter theo thời gian nếu có
		if req.Minutes > 0 {
			timeThreshold := time.Now().Add(time.Duration(-req.Minutes) * time.Minute)
			query += " AND created_at > ?"
			args = append(args, timeThreshold)
		}

		// Filter theo user_id nếu có
		if req.UserID > 0 {
			query += " AND user_id = ?"
			args = append(args, req.UserID)
		}

		unionQueries = append(unionQueries, query)
	}

	if len(unionQueries) == 0 {
		return []models.ActivityLog{}, 0, nil
	}

	// Tạo query chính với UNION
	mainQuery := strings.Join(unionQueries, " UNION ALL ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (%s) as combined
	`, mainQuery)

	err := db.ReplicaDB.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get failed logins with pagination
	failedLoginsQuery := fmt.Sprintf(`
		SELECT * FROM (%s) as combined
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, mainQuery)

	var failedLogins []models.ActivityLog
	offset := (req.Page - 1) * req.Limit
	allArgs := append(args, req.Limit, offset)

	err = db.ReplicaDB.Raw(failedLoginsQuery, allArgs...).Scan(&failedLogins).Error
	if err != nil {
		return nil, 0, err
	}

	return failedLogins, total, nil
}

// GetFailedLoginsWithPagingByTimeRange lấy danh sách failed login attempts trong khoảng thời gian cụ thể với phân trang
func (r *monthlyActivityLogRepository) GetFailedLoginsWithPagingByTimeRange(startTime, endTime time.Time, userID int64, page, limit int) ([]models.ActivityLog, int64, error) {
	// Lấy danh sách bảng cần query
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	// Tạo UNION query cho tất cả các bảng
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

		query := fmt.Sprintf(`
			SELECT id, user_id, role_id, method, path, status_code, client_ip, device,
				latency_ms, query_params, request_body, response_body, attribute,
				session_id, agent, created_at
			FROM %s
			WHERE path = ? AND status_code != ? AND client_ip IS NOT NULL AND client_ip != ''
			AND created_at >= ? AND created_at <= ?
		`, tableName)

		args = append(args, "/api/login", 200, startTime, endTime)

		// Filter theo user_id nếu có
		if userID > 0 {
			query += " AND user_id = ?"
			args = append(args, userID)
		}

		unionQueries = append(unionQueries, query)
	}

	if len(unionQueries) == 0 {
		return []models.ActivityLog{}, 0, nil
	}

	// Tạo query chính với UNION
	mainQuery := strings.Join(unionQueries, " UNION ALL ")

	// Count total
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM (%s) as combined
	`, mainQuery)

	err := db.ReplicaDB.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get failed logins with pagination
	failedLoginsQuery := fmt.Sprintf(`
		SELECT * FROM (%s) as combined
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, mainQuery)

	var failedLogins []models.ActivityLog
	offset := (page - 1) * limit
	allArgs := append(args, limit, offset)

	err = db.ReplicaDB.Raw(failedLoginsQuery, allArgs...).Scan(&failedLogins).Error
	if err != nil {
		return nil, 0, err
	}

	return failedLogins, total, nil
}

// PermanentlyDeleteOldRecords xóa các bản ghi cũ (xóa toàn bộ bảng của tháng cũ)
func (r *monthlyActivityLogRepository) PermanentlyDeleteOldRecords(now time.Time) error {
	// Lấy tên bảng của tháng trước
	oldTableName := table_manager.GetPreviousMonthTableName("activity_logs")

	// Kiểm tra bảng có tồn tại không
	var exists bool
	err := db.MasterDB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = ?
		)
	`, oldTableName).Scan(&exists).Error

	if err != nil {
		return err
	}

	if !exists {
		return nil // Bảng không tồn tại, không cần xóa
	}

	// Xóa bảng
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", oldTableName)
	return db.MasterDB.Exec(dropTableSQL).Error
}

func detectDevice(agent string) string {
	agentLower := strings.ToLower(agent)

	switch {
	case strings.Contains(agentLower, "iphone"):
		return "Mobile"
	case strings.Contains(agentLower, "ipad"):
		return "Tablet"
	case strings.Contains(agentLower, "android") && strings.Contains(agentLower, "mobile"):
		return "Mobile"
	case strings.Contains(agentLower, "android"):
		return "Tablet"
	case strings.Contains(agentLower, "windows nt"):
		return "Desktop"
	case strings.Contains(agentLower, "macintosh"):
		return "Desktop"
	default:
		return "Tablet"
	}
}
