package services

import (
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WarningService interface {
	GetActiveUsersCount(minutes int) (*prot.ActiveUsersCountResponse, error)
	GetActiveUsersCountByTimeRange(startTime, endTime time.Time) (*prot.ActiveUsersCountResponse, error)
	GetActiveUsersWithPaging(req *requests.GetActiveUsersRequest, c *gin.Context) (*prot.ActiveUsersListResponse, error)
	GetActiveUsersWithPagingByTimeRange(startTime, endTime time.Time, page, limit int, c *gin.Context) (*prot.ActiveUsersListResponse, error)
	GetFailedLoginsCount(req *requests.GetFailedLoginRequest) (*prot.FailedLoginsCountResponse, error)
	GetFailedLoginsCountByTimeRange(startTime, endTime time.Time, userID int64) (*prot.FailedLoginsCountResponse, error)
	GetFailedLoginsWithPaging(req *requests.GetFailedLoginRequest) (*prot.FailedLoginsListResponse, error)
	GetFailedLoginsWithPagingByTimeRange(startTime, endTime time.Time, userID int64, page, limit int) (*prot.FailedLoginsListResponse, error)
}

type warningService struct {
	activityLogRepo repositories.MonthlyActivityLogRepository
}

func NewWarningService(activityLogRepo repositories.MonthlyActivityLogRepository) WarningService {
	return &warningService{
		activityLogRepo: activityLogRepo,
	}
}

// API 1: Chỉ đếm số lượng (nhanh)
func (s *warningService) GetActiveUsersCount(minutes int) (*prot.ActiveUsersCountResponse, error) {
	// Default: không filter (0 = không filter)
	count, err := s.activityLogRepo.GetActiveUsersCount(minutes, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	return &prot.ActiveUsersCountResponse{
		TotalActiveUsers: count,
		TimeRange:        s.getTimeRangeLabel(minutes),
		Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 1b: Đếm số lượng theo khoảng thời gian cụ thể
func (s *warningService) GetActiveUsersCountByTimeRange(startTime, endTime time.Time) (*prot.ActiveUsersCountResponse, error) {
	// Default: không filter (0 = không filter)
	count, err := s.activityLogRepo.GetActiveUsersCountByTimeRange(startTime, endTime, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	timeRangeLabel := fmt.Sprintf("from %s to %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	return &prot.ActiveUsersCountResponse{
		TotalActiveUsers: count,
		TimeRange:        timeRangeLabel,
		Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 2: Lấy danh sách chi tiết với phân trang
func (s *warningService) GetActiveUsersWithPaging(req *requests.GetActiveUsersRequest, c *gin.Context) (*prot.ActiveUsersListResponse, error) {
	users, total, err := s.activityLogRepo.GetActiveUsersWithPaging(req, c)
	if err != nil {
		return nil, err
	}

	// Lấy thông tin thời gian hoạt động cuối cùng
	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, uint(user.ID))
	}

	lastActivityTimes, err := s.activityLogRepo.GetUsersLastActivityTime(userIDs)
	if err != nil {
		// Nếu lỗi, vẫn trả về data nhưng không có thông tin thời gian
		lastActivityTimes = make(map[uint]time.Time)
	}

	// Convert models.User to ActiveUserInfo với thông tin thời gian
	var activeUsers []*prot.ActiveUserInfo
	for _, user := range users {
		userID := uint(user.ID)
		lastActivity, exists := lastActivityTimes[userID]

		var onlineDuration string
		var lastActivityStr string

		if exists {
			// Tính thời gian online
			duration := time.Since(lastActivity)
			onlineDuration = s.formatDuration(duration)
			lastActivityStr = lastActivity.Format("2006-01-02 15:04:05")
		} else {
			onlineDuration = "Không xác định"
			lastActivityStr = "Không xác định"
		}

		activeUsers = append(activeUsers, &prot.ActiveUserInfo{
			Id:               int32(userID),
			Name:             user.Name,
			Email:            user.Email,
			Username:         user.Username,
			OnlineDuration:   onlineDuration,
			LastActivityTime: lastActivityStr,
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	if req.Limit == 0 {
		totalPages = 1
	}

	return &prot.ActiveUsersListResponse{
		Users:            activeUsers,
		TotalActiveUsers: total,
		Page:             int32(req.Page),
		Limit:            int32(req.Limit),
		TotalPages:       int32(totalPages),
		TimeRange:        s.getTimeRangeLabel(req.Minutes),
		Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *warningService) getTimeRangeLabel(minutes int) string {
	if minutes <= 15 {
		return "currently_online"
	} else if minutes <= 60 {
		return "last_hour"
	} else if minutes <= 1440 {
		return "today"
	} else {
		return "recent_days"
	}
}

// Format duration thành chuỗi dễ đọc bằng tiếng Việt
func (s *warningService) formatDuration(duration time.Duration) string {
	if duration < 0 {
		return "Vừa mới"
	}

	totalSeconds := int(duration.Seconds())

	if totalSeconds < 60 {
		return "Vừa mới"
	}

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	if minutes < 60 {
		if seconds > 0 {
			return fmt.Sprintf("%d phút %d giây trước", minutes, seconds)
		}
		return fmt.Sprintf("%d phút trước", minutes)
	}

	hours := minutes / 60
	minutes = minutes % 60

	if hours < 24 {
		if minutes > 0 {
			return fmt.Sprintf("%d giờ %d phút trước", hours, minutes)
		}
		return fmt.Sprintf("%d giờ trước", hours)
	}

	days := hours / 24
	hours = hours % 24

	if days < 7 {
		if hours > 0 {
			return fmt.Sprintf("%d ngày %d giờ trước", days, hours)
		}
		return fmt.Sprintf("%d ngày trước", days)
	}

	return fmt.Sprintf("%d tuần trước", days/7)
}

// API 3: Đếm số lượng IP có failed login attempts
func (s *warningService) GetFailedLoginsCount(req *requests.GetFailedLoginRequest) (*prot.FailedLoginsCountResponse, error) {
	count, err := s.activityLogRepo.GetFailedLoginsCount(req)
	if err != nil {
		return nil, err
	}

	return &prot.FailedLoginsCountResponse{
		TotalFailedLogins: count, // Thực tế là số lượng IP có failed login
		TimeRange:         s.getTimeRangeLabel(req.Minutes),
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 3b: Đếm số lượng IP có failed login attempts theo khoảng thời gian cụ thể
func (s *warningService) GetFailedLoginsCountByTimeRange(startTime, endTime time.Time, userID int64) (*prot.FailedLoginsCountResponse, error) {
	count, err := s.activityLogRepo.GetFailedLoginsCountByTimeRange(startTime, endTime, userID)
	if err != nil {
		return nil, err
	}

	timeRangeLabel := fmt.Sprintf("from %s to %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	return &prot.FailedLoginsCountResponse{
		TotalFailedLogins: count, // Thực tế là số lượng IP có failed login
		TimeRange:         timeRangeLabel,
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 4: Lấy danh sách chi tiết failed login attempts với phân trang
func (s *warningService) GetFailedLoginsWithPaging(req *requests.GetFailedLoginRequest) (*prot.FailedLoginsListResponse, error) {
	failedLogins, total, err := s.activityLogRepo.GetFailedLoginsWithPaging(req)
	if err != nil {
		return nil, err
	}

	// Convert models.ActivityLog to FailedLoginInfo
	var failedLoginInfos []*prot.FailedLoginInfo
	for _, log := range failedLogins {
		// Helper function to convert JSON to string
		jsonToString := func(jsonData []byte) string {
			if len(jsonData) == 0 {
				return ""
			}
			return string(jsonData)
		}

		var userId int32
		if log.UserID != nil {
			userId = int32(*log.UserID)
		}

		var roleId int32
		if log.RoleID != nil {
			roleId = int32(*log.RoleID)
		}

		failedLoginInfos = append(failedLoginInfos, &prot.FailedLoginInfo{
			UserId:       userId,                         // *int từ activity log
			RoleId:       roleId,                         // *int từ activity log
			ClientIp:     log.ClientIP,                   // IP address
			QueryParams:  jsonToString(log.QueryParams),  // Convert JSON to string
			RequestBody:  jsonToString(log.RequestBody),  // Convert JSON to string
			ResponseBody: jsonToString(log.ResponseBody), // Convert JSON to string
			CreatedAt:    log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Calculate total pages
	totalPages := 1
	if req.Limit > 0 {
		totalPages = int((total + int64(req.Limit) - 1) / int64(req.Limit))
	}

	return &prot.FailedLoginsListResponse{
		FailedLogins:      failedLoginInfos,
		TotalFailedLogins: total,
		Page:              int32(req.Page),
		Limit:             int32(req.Limit),
		TotalPages:        int32(totalPages),
		TimeRange:         s.getTimeRangeLabel(req.Minutes),
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 2b: Lấy danh sách active users với phân trang theo khoảng thời gian cụ thể
func (s *warningService) GetActiveUsersWithPagingByTimeRange(startTime, endTime time.Time, page, limit int, c *gin.Context) (*prot.ActiveUsersListResponse, error) {
	// Lấy filter từ query params nếu có
	var schoolID, courseID, programID int64
	if c != nil {
		if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
			if id, err := strconv.ParseInt(schoolIDStr, 10, 64); err == nil {
				schoolID = id
			}
		}
		if courseIDStr := c.Query("course_id"); courseIDStr != "" {
			if id, err := strconv.ParseInt(courseIDStr, 10, 64); err == nil {
				courseID = id
			}
		}
		if programIDStr := c.Query("program_id"); programIDStr != "" {
			if id, err := strconv.ParseInt(programIDStr, 10, 64); err == nil {
				programID = id
			}
		}
	}

	// Get active users by time range
	activeUserIDs, total, err := s.activityLogRepo.GetActiveUsersByTimeRange(startTime, endTime, page, limit, schoolID, courseID, programID)
	if err != nil {
		return nil, err
	}

	// Get user details for each active user
	var activeUsers []*prot.ActiveUserInfo
	for _, userID := range activeUserIDs {
		user, err := s.activityLogRepo.GetUserByID(userID)
		if err != nil {
			continue // Skip if user not found
		}

		// Get user's activity info in the time range
		lastActivity, err := s.activityLogRepo.GetUserLastActivityInTimeRange(userID, startTime, endTime)
		var onlineDuration, lastActivityStr string
		if err == nil && lastActivity != nil {
			duration := endTime.Sub(*lastActivity)
			onlineDuration = s.formatDuration(duration)
			lastActivityStr = lastActivity.Format("2006-01-02 15:04:05")
		} else {
			onlineDuration = "Không xác định"
			lastActivityStr = "Không xác định"
		}

		activeUsers = append(activeUsers, &prot.ActiveUserInfo{
			Id:               int32(userID),
			Name:             user.Name,
			Email:            user.Email,
			Username:         user.Username,
			OnlineDuration:   onlineDuration,
			LastActivityTime: lastActivityStr,
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if limit == 0 {
		totalPages = 1
	}

	timeRangeLabel := fmt.Sprintf("from %s to %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	return &prot.ActiveUsersListResponse{
		Users:            activeUsers,
		TotalActiveUsers: total,
		Page:             int32(page),
		Limit:            int32(limit),
		TotalPages:       int32(totalPages),
		TimeRange:        timeRangeLabel,
		Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// API 4b: Lấy danh sách failed login attempts với phân trang theo khoảng thời gian cụ thể
func (s *warningService) GetFailedLoginsWithPagingByTimeRange(startTime, endTime time.Time, userID int64, page, limit int) (*prot.FailedLoginsListResponse, error) {
	failedLogins, total, err := s.activityLogRepo.GetFailedLoginsWithPagingByTimeRange(startTime, endTime, userID, page, limit)
	if err != nil {
		return nil, err
	}

	// Convert models.ActivityLog to FailedLoginInfo
	var failedLoginInfos []*prot.FailedLoginInfo
	for _, log := range failedLogins {
		// Helper function to convert JSON to string
		jsonToString := func(jsonData []byte) string {
			if len(jsonData) == 0 {
				return ""
			}
			return string(jsonData)
		}

		var userId int32
		if log.UserID != nil {
			userId = int32(*log.UserID)
		}

		var roleId int32
		if log.RoleID != nil {
			roleId = int32(*log.RoleID)
		}

		failedLoginInfos = append(failedLoginInfos, &prot.FailedLoginInfo{
			UserId:       userId,                         // *int từ activity log
			RoleId:       roleId,                         // *int từ activity log
			ClientIp:     log.ClientIP,                   // IP address
			QueryParams:  jsonToString(log.QueryParams),  // Convert JSON to string
			RequestBody:  jsonToString(log.RequestBody),  // Convert JSON to string
			ResponseBody: jsonToString(log.ResponseBody), // Convert JSON to string
			CreatedAt:    log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	// Calculate total pages
	totalPages := 1
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	timeRangeLabel := fmt.Sprintf("from %s to %s",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"))

	return &prot.FailedLoginsListResponse{
		FailedLogins:      failedLoginInfos,
		TotalFailedLogins: total,
		Page:              int32(page),
		Limit:             int32(limit),
		TotalPages:        int32(totalPages),
		TimeRange:         timeRangeLabel,
		Timestamp:         time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
