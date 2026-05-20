package controllers

import (
	"be-cleverschool/requests"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WarningController struct {
	warningService services.WarningService
}

func NewWarningController(warningService services.WarningService) *WarningController {
	return &WarningController{
		warningService: warningService,
	}
}

// GET /api/manage/warnings/active-users-count?minutes=15&start_time=1704096600&end_time=1704126300
func (wc *WarningController) GetActiveUsersCount(c *gin.Context) {
	minutesStr := c.DefaultQuery("minutes", "15")
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil || minutes < 0 {
		utils.Respond(c, nil, errors.New("Invalid minutes parameter"), "", http.StatusBadRequest)
		return
	}

	// Nếu minutes = 0, sử dụng thời gian bắt đầu và kết thúc
	if minutes == 0 {
		startTimeStr := c.Query("start_time")
		endTimeStr := c.Query("end_time")

		if startTimeStr == "" || endTimeStr == "" {
			utils.Respond(c, nil, errors.New("start_time and end_time are required when minutes=0"), "", http.StatusBadRequest)
			return
		}

		// Parse Unix timestamp
		startTimestamp, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid start_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		endTimestamp, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid end_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		// Convert Unix timestamp to time.Time
		startTime := time.Unix(startTimestamp, 0)
		endTime := time.Unix(endTimestamp, 0)

		if endTime.Before(startTime) {
			utils.Respond(c, nil, errors.New("end_time must be after start_time"), "", http.StatusBadRequest)
			return
		}

		result, err := wc.warningService.GetActiveUsersCountByTimeRange(startTime, endTime)
		if err != nil {
			utils.Respond(c, nil, err, "", http.StatusInternalServerError)
			return
		}

		utils.Respond(c, result, nil, "")
		return
	}

	// Logic cũ cho minutes > 0
	if minutes > 10080 { // Giới hạn 7 ngày
		minutes = 10080
	}

	result, err := wc.warningService.GetActiveUsersCount(minutes)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, result, nil, "")
}

// GET /api/manage/warnings/active-users?minutes=15&page=1&limit=20&start_time=1704096600&end_time=1704126300
func (wc *WarningController) GetActiveUsers(c *gin.Context) {
	var req requests.GetActiveUsersRequest

	// Bind query parameters theo pattern của dự án
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "", http.StatusBadRequest)
		return
	}

	// Nếu minutes = 0, sử dụng thời gian bắt đầu và kết thúc
	if req.Minutes == 0 {
		if req.StartTime == "" || req.EndTime == "" {
			utils.Respond(c, nil, errors.New("start_time and end_time are required when minutes=0"), "", http.StatusBadRequest)
			return
		}

		// Parse Unix timestamp
		startTimestamp, err := strconv.ParseInt(req.StartTime, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid start_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		endTimestamp, err := strconv.ParseInt(req.EndTime, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid end_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		// Convert Unix timestamp to time.Time
		startTime := time.Unix(startTimestamp, 0)
		endTime := time.Unix(endTimestamp, 0)

		if endTime.Before(startTime) {
			utils.Respond(c, nil, errors.New("end_time must be after start_time"), "", http.StatusBadRequest)
			return
		}

		// Set defaults for pagination
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 20
		}
		if req.Limit > 100 {
			req.Limit = 100
		}

		result, err := wc.warningService.GetActiveUsersWithPagingByTimeRange(startTime, endTime, req.Page, req.Limit, c)
		if err != nil {
			utils.Respond(c, nil, err, "", http.StatusInternalServerError)
			return
		}

		utils.Respond(c, result, nil, "")
		return
	}

	// Set defaults
	if req.Minutes <= 0 {
		req.Minutes = 15
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// Set limits để tránh query quá nặng
	if req.Minutes > 10080 {
		req.Minutes = 10080
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	result, err := wc.warningService.GetActiveUsersWithPaging(&req, c)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, result, nil, "")
}

// GET /api/manage/warnings/failed-logins-count?minutes=60&user_id=123&start_time=1704096600&end_time=1704126300
func (wc *WarningController) GetFailedLoginsCount(c *gin.Context) {
	var req requests.GetFailedLoginRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "", http.StatusBadRequest)
		return
	}

	// Nếu minutes = 0, sử dụng thời gian bắt đầu và kết thúc
	if req.Minutes == 0 {
		startTimeStr := c.Query("start_time")
		endTimeStr := c.Query("end_time")

		if startTimeStr == "" || endTimeStr == "" {
			utils.Respond(c, nil, errors.New("start_time and end_time are required when minutes=0"), "", http.StatusBadRequest)
			return
		}

		// Parse Unix timestamp
		startTimestamp, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid start_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		endTimestamp, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid end_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		// Convert Unix timestamp to time.Time
		startTime := time.Unix(startTimestamp, 0)
		endTime := time.Unix(endTimestamp, 0)

		if endTime.Before(startTime) {
			utils.Respond(c, nil, errors.New("end_time must be after start_time"), "", http.StatusBadRequest)
			return
		}

		result, err := wc.warningService.GetFailedLoginsCountByTimeRange(startTime, endTime, int64(req.UserID))
		if err != nil {
			utils.Respond(c, nil, err, "", http.StatusInternalServerError)
			return
		}

		utils.Respond(c, result, nil, "")
		return
	}

	// Set defaults cho logic cũ
	if req.Minutes <= 0 {
		req.Minutes = 60 // Default 1 hour for failed logins
	}

	// Set limits để tránh query quá nặng
	if req.Minutes > 10080 { // Max 7 days
		req.Minutes = 10080
	}

	result, err := wc.warningService.GetFailedLoginsCount(&req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, result, nil, "")
}

// GET /api/manage/warnings/failed-logins?minutes=60&user_id=123&page=1&limit=20&start_time=1704096600&end_time=1704126300
func (wc *WarningController) GetFailedLogins(c *gin.Context) {
	var req requests.GetFailedLoginRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "", http.StatusBadRequest)
		return
	}

	// Nếu minutes = 0, sử dụng thời gian bắt đầu và kết thúc
	if req.Minutes == 0 {
		if req.StartTime == "" || req.EndTime == "" {
			utils.Respond(c, nil, errors.New("start_time and end_time are required when minutes=0"), "", http.StatusBadRequest)
			return
		}

		// Parse Unix timestamp
		startTimestamp, err := strconv.ParseInt(req.StartTime, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid start_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		endTimestamp, err := strconv.ParseInt(req.EndTime, 10, 64)
		if err != nil {
			utils.Respond(c, nil, errors.New("Invalid end_time format. Use Unix timestamp"), "", http.StatusBadRequest)
			return
		}

		// Convert Unix timestamp to time.Time
		startTime := time.Unix(startTimestamp, 0)
		endTime := time.Unix(endTimestamp, 0)

		if endTime.Before(startTime) {
			utils.Respond(c, nil, errors.New("end_time must be after start_time"), "", http.StatusBadRequest)
			return
		}

		// Set defaults for pagination
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.Limit <= 0 {
			req.Limit = 20
		}
		if req.Limit > 100 {
			req.Limit = 100
		}

		result, err := wc.warningService.GetFailedLoginsWithPagingByTimeRange(startTime, endTime, int64(req.UserID), req.Page, req.Limit)
		if err != nil {
			utils.Respond(c, nil, err, "", http.StatusInternalServerError)
			return
		}

		utils.Respond(c, result, nil, "")
		return
	}

	// Set defaults
	if req.Minutes <= 0 {
		req.Minutes = 60 // Default 1 hour for failed logins
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	// Set limits để tránh query quá nặng
	if req.Minutes > 10080 { // Max 7 days
		req.Minutes = 10080
	}
	if req.Limit > 100 { // Max 100 records per page
		req.Limit = 100
	}

	result, err := wc.warningService.GetFailedLoginsWithPaging(&req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, result, nil, "")
}

