package controllers

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/services"
	"be-Clever School/utils"
	"bytes"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MeetingAttendanceController struct {
	service services.MeetingAttendanceService
}

type MeetingJoinResponse struct {
	Message    string                    `json:"message"`
	Meeting    *models.MicrosoftMeeting  `json:"meeting"`
	Attendance *models.MeetingAttendance `json:"attendance"`
	JoinURL    string                    `json:"join_url"`
}

type MeetingAttendancesResponse struct {
	Attendances []*dto.MeetingAttendanceResponse `json:"attendances"`
	Summary     *models.AttendanceSummary        `json:"summary"`
}

type MeetingAttendanceHistoryResponse struct {
	Attendances []*models.MeetingAttendance `json:"attendances"`
	Page        int                         `json:"page"`
	Limit       int                         `json:"limit"`
}

func NewMeetingAttendanceController() *MeetingAttendanceController {
	return &MeetingAttendanceController{
		service: services.NewMeetingAttendanceService(),
	}
}

type JoinMeetingRequest struct {
	MeetingID  uint   `json:"meeting_id"`
	ShortCode  string `json:"short_code"`
	JoinMethod string `json:"join_method"`
	DeviceInfo string `json:"device_info"`
}

// @Summary Join Meeting by Short Code
// @Description Student joins a meeting using short code, attendance is recorded automatically
// @Tags Meeting Attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param code path string true "Meeting short code"
// @Success 200 {object} map[string]interface{} "join success with meeting details"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Meeting not found"
// @Router /microsoft/meetings/join/{code} [post]
func (c *MeetingAttendanceController) JoinByShortCode(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		utils.Respond(ctx, nil, http.ErrNoCookie, "Unauthorized", http.StatusUnauthorized)
		return
	}

	shortCode := ctx.Param("code")
	if shortCode == "" {
		utils.Respond(ctx, nil, http.ErrNoCookie, "Short code is required", http.StatusBadRequest)
		return
	}

	deviceInfo := ctx.GetHeader("User-Agent")
	ipAddress := ctx.ClientIP()

	roleId := utils.GetCurrentRoleId(ctx)
	meeting, attendance, err := c.service.JoinByShortCode(shortCode, uint(userID), roleId, deviceInfo, ipAddress)
	if err != nil {
		config.Log.Errorf("❌ Failed to join meeting: %v", err)
		utils.Respond(ctx, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	config.Log.Infof("✅ User %d joined meeting %s via short code", userID, meeting.Title)

	utils.Respond(ctx, &MeetingJoinResponse{
		Message:    "Successfully joined meeting",
		Meeting:    meeting,
		Attendance: attendance,
		JoinURL:    meeting.JoinURL,
	}, nil, "")
}

// @Summary Join Meeting by ID
// @Description Record attendance when user joins a meeting directly
// @Tags Meeting Attendance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "attendance recorded"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /microsoft/meetings/{id}/join [post]
func (c *MeetingAttendanceController) JoinMeeting(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		utils.Respond(ctx, nil, http.ErrNoCookie, "Unauthorized", http.StatusUnauthorized)
		return
	}

	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	deviceInfo := ctx.GetHeader("User-Agent")
	ipAddress := ctx.ClientIP()

	attendance, err := c.service.JoinMeeting(uint(meetingID), uint(userID), "direct", deviceInfo, ipAddress)
	if err != nil {
		utils.Respond(ctx, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, attendance, nil, "")
}

// @Summary Leave Meeting
// @Description Record when user leaves a meeting
// @Tags Meeting Attendance
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "left meeting"
// @Router /microsoft/meetings/{id}/leave [post]
func (c *MeetingAttendanceController) LeaveMeeting(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		utils.Respond(ctx, nil, http.ErrNoCookie, "Unauthorized", http.StatusUnauthorized)
		return
	}

	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	if err := c.service.LeaveMeeting(uint(meetingID), uint(userID)); err != nil {
		utils.Respond(ctx, nil, err, err.Error(), http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, &models.AttendanceSummary{MeetingID: uint(meetingID)}, nil, "")
}

// @Summary Get Meeting Attendances
// @Description Get all attendances for a meeting (for teachers), excludes teachers
// @Tags Meeting Attendance
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "attendances list with user details"
// @Router /microsoft/meetings/{id}/attendances [get]
func (c *MeetingAttendanceController) GetMeetingAttendances(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	attendances, err := c.service.GetMeetingAttendancesWithDetails(uint(meetingID))
	if err != nil {
		utils.Respond(ctx, nil, err, err.Error(), http.StatusInternalServerError)
		return
	}

	summary, _ := c.service.GetAttendanceSummary(uint(meetingID))

	utils.Respond(ctx, &MeetingAttendancesResponse{
		Attendances: attendances,
		Summary:     summary,
	}, nil, "")
}

// @Summary Get User Attendance History
// @Description Get attendance history for current user
// @Tags Meeting Attendance
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{} "attendance history"
// @Router /microsoft/meetings/my-attendance [get]
func (c *MeetingAttendanceController) GetUserAttendanceHistory(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		utils.Respond(ctx, nil, http.ErrNoCookie, "Unauthorized", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	attendances, err := c.service.GetUserAttendanceHistory(uint(userID), page, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Respond(ctx, &MeetingAttendanceHistoryResponse{
		Attendances: attendances,
		Page:        page,
		Limit:       limit,
	}, nil, "")
}

// @Summary Get Attendance Summary
// @Description Get attendance summary for a meeting
// @Tags Meeting Attendance
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "attendance summary"
// @Router /microsoft/meetings/{id}/summary [get]
func (c *MeetingAttendanceController) GetAttendanceSummary(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	summary, err := c.service.GetAttendanceSummary(uint(meetingID))
	if err != nil {
		utils.Respond(ctx, nil, err, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Respond(ctx, summary, nil, "")
}

// @Summary Export Meeting Attendances to Excel
// @Description Export attendance list to Excel file for a meeting
// @Tags Meeting Attendance
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {file} file "Excel file"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /microsoft/meetings/{id}/attendances/export [get]
func (c *MeetingAttendanceController) ExportAttendancesToExcel(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	buf := bytes.NewBuffer(nil)

	err = c.service.ExportAttendancesToExcel(uint(meetingID), buf)
	if err != nil {
		config.Log.Errorf("Failed to export attendances to Excel: %v", err)
		utils.Respond(ctx, nil, err, "Failed to export attendances", http.StatusInternalServerError)
		return
	}

	// Set response headers for file download
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=attendance_"+strconv.FormatUint(uint64(meetingID), 10)+".xlsx")
	ctx.Writer.WriteHeader(http.StatusOK)
	_, _ = ctx.Writer.Write(buf.Bytes())
}
