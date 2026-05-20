package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GoogleMeetingController struct {
	service services.GoogleMeetingService
}

type UpdateGoogleRecordingRequest struct {
	RecordingURL    string `json:"recording_url" binding:"required"`
	RecordingStatus string `json:"recording_status"`
	DurationMinutes int    `json:"duration_minutes"`
}

func NewGoogleMeetingController() *GoogleMeetingController {
	return &GoogleMeetingController{
		service: services.NewGoogleMeetingService(),
	}
}

// CreateMeeting creates a new Google Meet meeting
func (ctrl *GoogleMeetingController) CreateMeeting(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	var request services.CreateGoogleMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Respond(c, nil, err, "Invalid request data")
		return
	}

	meeting, err := ctrl.service.CreateMeeting(uint(userID), &request)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to create Google Meet")
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// GetMeeting retrieves a specific meeting
func (ctrl *GoogleMeetingController) GetMeeting(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID")
		return
	}

	meeting, err := ctrl.service.GetMeeting(uint(id))
	if err != nil {
		utils.Respond(c, nil, err, "Meeting not found")
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// UpdateMeeting updates an existing meeting
func (ctrl *GoogleMeetingController) UpdateMeeting(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID")
		return
	}

	var request services.UpdateGoogleMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Respond(c, nil, err, "Invalid request data")
		return
	}

	meeting, err := ctrl.service.UpdateMeeting(uint(id), &request)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to update meeting")
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// DeleteMeeting deletes a meeting
func (ctrl *GoogleMeetingController) DeleteMeeting(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID")
		return
	}

	err = ctrl.service.DeleteMeeting(uint(id))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to delete meeting")
		return
	}

	utils.Respond(c, gin.H{
		"message": "Meeting deleted successfully",
	}, nil, "")
}

func (ctrl *GoogleMeetingController) ForceDeleteMeeting(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID")
		return
	}

	deletedBy := utils.GetCurrentUserId(c)
	if err := ctrl.service.ForceDeleteMeeting(uint(id), uint(deletedBy)); err != nil {
		utils.Respond(c, nil, err, "Failed to hard delete meeting")
		return
	}

	utils.Respond(c, &prot.DeleteResponse{
		Id:      int64(id),
		Message: "Meeting hard-deleted successfully",
	}, nil, "")
}

// GetUserMeetings retrieves meetings for current user
func (ctrl *GoogleMeetingController) GetUserMeetings(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	meetings, err := ctrl.service.GetUserMeetings(uint(userID), page, limit)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to retrieve meetings")
		return
	}

	utils.Respond(c, gin.H{
		"meetings": meetings,
		"page":     page,
		"limit":    limit,
	}, nil, "")
}

// CreateCourseIntegration creates a meeting for a specific course
func (ctrl *GoogleMeetingController) CreateCourseIntegration(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)
	courseIdParam := c.Param("id")
	courseId, err := strconv.ParseUint(courseIdParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid course ID")
		return
	}

	var request services.CreateGoogleMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Respond(c, nil, err, "Invalid request data")
		return
	}

	// Set course ID
	courseIdUint := uint(courseId)
	request.CourseID = &courseIdUint

	meeting, err := ctrl.service.CreateMeeting(uint(userID), &request)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to create Google Meet for course")
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// CreateLessonIntegration creates a meeting for a specific lesson
func (ctrl *GoogleMeetingController) CreateLessonIntegration(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)
	lessonIdParam := c.Param("id")
	lessonId, err := strconv.ParseUint(lessonIdParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid lesson ID")
		return
	}

	var request services.CreateGoogleMeetingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Respond(c, nil, err, "Invalid request data")
		return
	}

	// Set lesson ID
	lessonIdUint := uint(lessonId)
	request.LessonID = &lessonIdUint

	meeting, err := ctrl.service.CreateMeeting(uint(userID), &request)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to create Google Meet for lesson")
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// GetCourseMeetings retrieves meetings for a course
func (ctrl *GoogleMeetingController) GetCourseMeetings(c *gin.Context) {
	courseIdParam := c.Param("id")
	courseId, err := strconv.ParseUint(courseIdParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid course ID")
		return
	}

	meetings, err := ctrl.service.GetCourseMeetings(uint(courseId))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to retrieve course meetings")
		return
	}

	utils.Respond(c, gin.H{
		"meetings":  meetings,
		"course_id": courseId,
	}, nil, "")
}

// GetLessonMeetings retrieves meetings for a lesson
func (ctrl *GoogleMeetingController) GetLessonMeetings(c *gin.Context) {
	lessonIdParam := c.Param("id")
	lessonId, err := strconv.ParseUint(lessonIdParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid lesson ID")
		return
	}

	meetings, err := ctrl.service.GetLessonMeetings(uint(lessonId))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to retrieve lesson meetings")
		return
	}

	utils.Respond(c, gin.H{
		"meetings":  meetings,
		"lesson_id": lessonId,
	}, nil, "")
}

// GetUpcomingMeetings retrieves upcoming meetings for current user
func (ctrl *GoogleMeetingController) GetUpcomingMeetings(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	meetings, err := ctrl.service.GetUpcomingMeetings(uint(userID), limit)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to retrieve upcoming meetings")
		return
	}

	utils.Respond(c, gin.H{
		"meetings": meetings,
		"limit":    limit,
	}, nil, "")
}
func (ctrl *GoogleMeetingController) UpdateRecording(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID")
		return
	}

	var req UpdateGoogleRecordingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request data")
		return
	}

	status := req.RecordingStatus
	if status == "" {
		status = "available"
	}

	if err := ctrl.service.UpdateRecording(uint(id), req.RecordingURL, status, req.DurationMinutes); err != nil {
		utils.Respond(c, nil, err, "Failed to update recording")
		return
	}

	utils.Respond(c, &prot.CompleteResponse{
		Id:      int64(id),
		Message: "Recording updated",
	}, nil, "")
}
