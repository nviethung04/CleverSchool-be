package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type MicrosoftMeetingController struct {
	service services.MicrosoftMeetingService
}

type SetRecordingShareURLRequest struct {
	ShareURL string `json:"share_url"`
}

func (c *MicrosoftMeetingController) SetRecordingShareURL(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid meeting ID", err.Error())
		return
	}

	var req SetRecordingShareURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	shareURL := strings.TrimSpace(req.ShareURL)

	if err := c.service.SetRecordingShareURL(uint(meetingID), shareURL); err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to save share link", err.Error())
		return
	}

	msg := "Recording share link saved"
	if shareURL == "" {
		msg = "Recording share link cleared"
	}
	c.ResponseSuccess(ctx, gin.H{"message": msg, "recording_share_url": shareURL})
}

func (c *MicrosoftMeetingController) FetchRecording(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid meeting ID", err.Error())
		return
	}

	recordingURL, err := c.service.FetchRecording(uint(meetingID))
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, "Recording not available", err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message":       "Recording fetched",
		"recording_url": recordingURL,
	})
}

func NewMicrosoftMeetingController() *MicrosoftMeetingController {
	return &MicrosoftMeetingController{
		service: services.NewMicrosoftMeetingService(),
	}
}

func (c *MicrosoftMeetingController) isUserInCourse(userID uint, courseID uint) (bool, error) {
	var count int64
	err := db.ReplicaDB.Table("user_courses").
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateMeetingRequest represents the request body for creating a Teams meeting
type CreateMeetingRequest struct {
	Title       string                            `json:"title" binding:"required"`
	Description string                            `json:"description"`
	StartTime   string                            `json:"start_time" binding:"required"`
	EndTime     string                            `json:"end_time" binding:"required"`
	TimeZone    string                            `json:"time_zone"`
	CourseID    *uint                             `json:"course_id"`
	LessonID    *uint                             `json:"lesson_id"`
	Attendees   []models.MicrosoftMeetingAttendee `json:"attendees"`
}

// UpdateMeetingRequest represents the request body for updating a Teams meeting
type UpdateMeetingRequest struct {
	Title       string                            `json:"title"`
	Description string                            `json:"description"`
	StartTime   string                            `json:"start_time"`
	EndTime     string                            `json:"end_time"`
	Attendees   []models.MicrosoftMeetingAttendee `json:"attendees"`
}

// @Summary Create Microsoft Teams Meeting
// @Description Create a new Microsoft Teams meeting
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateMeetingRequest true "Meeting details"
// @Success 201 {object} map[string]interface{} "meeting created"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /microsoft/meetings [post]
func (c *MicrosoftMeetingController) CreateMeeting(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("💥 Panic in CreateMeeting controller: %v", r)
		}
	}()

	config.Log.Infof("🚀 Starting CreateMeeting controller")
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	config.Log.Infof("🔍 UserID: %d", userID)

	permissionInfo, permErr := c.service.CheckMeetingPermissions(uint(userID))
	if permErr != nil {
		c.ResponseError(ctx, http.StatusUnauthorized, "Microsoft account not connected", permErr.Error())
		return
	}
	if permissionInfo == nil || !permissionInfo.HasOnlineMeetingPermission || !permissionInfo.HasCalendarPermission {
		c.ResponseError(ctx, http.StatusForbidden, "Forbidden", "Microsoft account does not have sufficient permissions to create meeting")
		return
	}

	var request CreateMeetingRequest
	config.Log.Infof("📝 Binding JSON request")
	if err := ctx.ShouldBindJSON(&request); err != nil {
		config.Log.Errorf("❌ JSON binding error: %v", err)
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	config.Log.Infof("✅ JSON bound successfully: %+v", request)

	// Parse times
	startTime, err := time.Parse(time.RFC3339, request.StartTime)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid start_time format", "Use RFC3339 format (2006-01-02T15:04:05Z07:00)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, request.EndTime)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid end_time format", "Use RFC3339 format (2006-01-02T15:04:05Z07:00)")
		return
	}

	// Set default timezone
	if request.TimeZone == "" {
		request.TimeZone = "UTC"
	}

	if request.CourseID != nil {
		roleId := utils.GetCurrentRoleId(ctx)
		if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
			ok, err := c.isUserInCourse(uint(userID), *request.CourseID)
			if err != nil {
				c.ResponseError(ctx, http.StatusInternalServerError, "Failed to validate course membership", err.Error())
				return
			}
			if !ok {
				c.ResponseError(ctx, http.StatusForbidden, "Forbidden", "User is not enrolled in this course")
				return
			}
		}
	}

	// Convert to service request
	serviceRequest := &services.CreateMicrosoftMeetingRequest{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		TimeZone:    request.TimeZone,
		CourseID:    request.CourseID,
		LessonID:    request.LessonID,
		Attendees:   request.Attendees,
	}

	config.Log.Infof("🔄 Calling service.CreateMeeting with service: %+v", c.service)
	meeting, err := c.service.CreateMeeting(uint(userID), serviceRequest)
	if err != nil {
		config.Log.Errorf("❌ Failed to create Microsoft Teams meeting: %v", err)
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to create meeting", err.Error())
		return
	}

	config.Log.Infof("✅ Created Microsoft Teams meeting: %s", meeting.JoinURL)

	c.ResponseSuccess(ctx, gin.H{
		"message": "Microsoft Teams meeting created successfully",
		"meeting": meeting,
	})
}

// @Summary Get Meeting by ID
// @Description Get Microsoft Teams meeting details by ID
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "meeting details"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Router /microsoft/meetings/{id} [get]
func (c *MicrosoftMeetingController) GetMeeting(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid meeting ID", nil)
		return
	}

	meeting, err := c.service.GetMeeting(uint(id))
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	if meeting.CourseID != nil {
		roleId := utils.GetCurrentRoleId(ctx)
		if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
			ok, err := c.isUserInCourse(uint(userID), *meeting.CourseID)
			if err != nil {
				c.ResponseError(ctx, http.StatusInternalServerError, "Failed to validate course membership", err.Error())
				return
			}
			if !ok {
				c.ResponseError(ctx, http.StatusForbidden, "Forbidden", "User is not enrolled in this course")
				return
			}
		}
	}

	c.ResponseSuccess(ctx, gin.H{
		"meeting": meeting,
		"user_id": userID,
	})
}

// @Summary Update Meeting
// @Description Update Microsoft Teams meeting details
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Meeting ID"
// @Param request body UpdateMeetingRequest true "Updated meeting details"
// @Success 200 {object} map[string]interface{} "meeting updated"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Router /microsoft/meetings/{id} [put]
func (c *MicrosoftMeetingController) UpdateMeeting(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid meeting ID", nil)
		return
	}

	var request UpdateMeetingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Parse times if provided
	var startTime, endTime time.Time
	if request.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, request.StartTime)
		if err != nil {
			c.ResponseError(ctx, http.StatusBadRequest, "Invalid start_time format", "Use RFC3339 format")
			return
		}
	}

	if request.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, request.EndTime)
		if err != nil {
			c.ResponseError(ctx, http.StatusBadRequest, "Invalid end_time format", "Use RFC3339 format")
			return
		}
	}

	// Convert to service request
	serviceRequest := &services.UpdateMicrosoftMeetingRequest{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Attendees:   request.Attendees,
	}

	meeting, err := c.service.UpdateMeeting(uint(id), serviceRequest)
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to update meeting", err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Meeting updated successfully",
		"meeting": meeting,
	})
}

// @Summary Delete Meeting
// @Description Delete Microsoft Teams meeting
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Param id path int true "Meeting ID"
// @Success 200 {object} map[string]interface{} "meeting deleted"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /microsoft/meetings/{id} [delete]
// DeleteMicrosoftMeeting xóa Microsoft meeting
func (c *MicrosoftMeetingController) DeleteMicrosoftMeeting(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	err = c.service.DeleteMeeting(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Meeting deleted successfully"})
}

// CheckMeetingPermissions kiểm tra quyền tạo meeting
func (c *MicrosoftMeetingController) CheckMeetingPermissions(ctx *gin.Context) {
	fmt.Println("🔍 DEBUG: CheckMeetingPermissions called")

	userID, exists := ctx.Get("userID")
	fmt.Printf("🔍 DEBUG: userID exists: %v, userID: %v, type: %T\n", exists, userID, userID)

	if !exists {
		fmt.Println("🔍 DEBUG: userID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	permissionInfo, err := c.service.CheckMeetingPermissions(uint(userIDInt))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": permissionInfo})
}

// @Summary Get User's Meetings
// @Description Get all Microsoft Teams meetings for the authenticated user
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "user meetings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /microsoft/meetings/user [get]
func (c *MicrosoftMeetingController) GetUserMeetings(ctx *gin.Context) {
	userID := c.GetUserID(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	meetings, err := c.service.GetUserMeetings(userID, page, limit)
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to get meetings", err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"meetings": meetings,
		"page":     page,
		"limit":    limit,
	})
}

// @Summary Get Course Meetings
// @Description Get all Microsoft Teams meetings for a course
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Param course_id path int true "Course ID"
// @Success 200 {object} map[string]interface{} "course meetings"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /microsoft/meetings/course/{course_id} [get]
func (c *MicrosoftMeetingController) GetCourseMeetings(ctx *gin.Context) {
	userID := c.GetUserID(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	courseID, err := strconv.ParseUint(ctx.Param("course_id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid course ID", nil)
		return
	}

	meetings, err := c.service.GetCourseMeetings(uint(courseID))
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to get meetings", err.Error())
		return
	}

	// Enforce that current user must belong to the course, except Admin/School roles
	roleId := utils.GetCurrentRoleId(ctx)
	if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
		ok, err := c.isUserInCourse(uint(userID), uint(courseID))
		if err != nil {
			c.ResponseError(ctx, http.StatusInternalServerError, "Failed to validate course membership", err.Error())
			return
		}
		if !ok {
			c.ResponseError(ctx, http.StatusForbidden, "Forbidden", "User is not enrolled in this course")
			return
		}
	}

	c.ResponseSuccess(ctx, gin.H{
		"meetings": meetings,
	})
}

func (c *MicrosoftMeetingController) CreateLessonIntegration(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	lessonID, err := strconv.ParseUint(ctx.Param("lesson_id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid lesson ID", nil)
		return
	}

	var request CreateMeetingRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Set lesson ID (don't require course ID)
	lessonIDUint := uint(lessonID)
	request.LessonID = &lessonIDUint

	// Parse times
	startTime, err := time.Parse(time.RFC3339, request.StartTime)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid start_time format", "Use RFC3339 format (2006-01-02T15:04:05Z07:00)")
		return
	}

	endTime, err := time.Parse(time.RFC3339, request.EndTime)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid end_time format", "Use RFC3339 format (2006-01-02T15:04:05Z07:00)")
		return
	}

	meeting, err := c.service.CreateMeeting(uint(userID), &services.CreateMicrosoftMeetingRequest{
		Title:       request.Title,
		Description: request.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		TimeZone:    request.TimeZone,
		CourseID:    request.CourseID,
		LessonID:    request.LessonID,
		Attendees:   request.Attendees,
	})

	if err != nil {
		config.Log.Errorf("❌ Failed to create meeting for lesson: %v", err)
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to create meeting", err.Error())
		return
	}

	config.Log.Infof("✅ Meeting created for lesson %d: %s", lessonID, meeting.Title)

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Meeting created successfully",
		"meeting": meeting,
	})
}

func (c *MicrosoftMeetingController) GetLessonMeetings(ctx *gin.Context) {
	lessonID, err := strconv.ParseUint(ctx.Param("lesson_id"), 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid lesson ID", nil)
		return
	}

	meetings, err := c.service.GetLessonMeetings(uint(lessonID))
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to get meetings", err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"meetings":  meetings,
		"lesson_id": lessonID,
	})
}

// JoinByShortCode returns meeting info & join URL by short code
func (c *MicrosoftMeetingController) JoinByShortCode(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	shortCode := ctx.Param("code")
	if shortCode == "" {
		c.ResponseError(ctx, http.StatusBadRequest, "Short code is required", nil)
		return
	}

	meeting, err := c.service.GetByShortCode(shortCode)
	if err != nil {
		c.ResponseError(ctx, http.StatusNotFound, "Meeting not found", err.Error())
		return
	}

	if meeting.CourseID != nil {
		roleId := utils.GetCurrentRoleId(ctx)
		if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
			ok, err := c.isUserInCourse(uint(userID), *meeting.CourseID)
			if err != nil {
				c.ResponseError(ctx, http.StatusInternalServerError, "Failed to validate course membership", err.Error())
				return
			}
			if !ok {
				c.ResponseError(ctx, http.StatusForbidden, "Forbidden", "User is not enrolled in this course")
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "OK",
		"meeting":   meeting,
		"join_url":  meeting.JoinURL,
		"join_web":  meeting.JoinWebURL,
		"shortcode": meeting.ShortCode,
	})
}

// @Summary Get Upcoming Meetings
// @Description Get upcoming Microsoft Teams meetings for the user
// @Tags Microsoft Meetings
// @Security BearerAuth
// @Param limit query int false "Number of meetings" default(5)
// @Success 200 {object} map[string]interface{} "upcoming meetings"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /microsoft/meetings/upcoming [get]
func (c *MicrosoftMeetingController) GetUpcomingMeetings(ctx *gin.Context) {
	userID := c.GetUserID(ctx)
	if userID == 0 {
		c.ResponseError(ctx, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))

	meetings, err := c.service.GetUpcomingMeetings(userID, limit)
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to get meetings", err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"meetings": meetings,
	})
}

// Helper methods
func (c *MicrosoftMeetingController) GetUserID(ctx *gin.Context) uint {
	userIDInterface, exists := ctx.Get("userID")
	if !exists {
		return 0
	}

	switch userID := userIDInterface.(type) {
	case uint:
		return userID
	case int:
		if userID <= 0 {
			return 0
		}
		return uint(userID)
	case int64:
		if userID <= 0 {
			return 0
		}
		return uint(userID)
	case float64:
		return uint(userID)
	case string:
		id, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			return 0
		}
		return uint(id)
	default:
		return 0
	}
}

func (c *MicrosoftMeetingController) ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (c *MicrosoftMeetingController) ResponseError(ctx *gin.Context, statusCode int, message string, details interface{}) {
	response := gin.H{
		"success": false,
		"message": message,
	}

	if details != nil {
		response["details"] = details
	}

	ctx.JSON(statusCode, response)
}

