package controllers

import (
	"be-lms/repositories"
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ZoomMeetingController struct {
	zoomMeetingService services.ZoomMeetingService
}

func NewZoomMeetingController() *ZoomMeetingController {
	meetingRepo := repositories.NewZoomMeetingRepository()
	zoomAuthRepo := repositories.NewZoomAuthRepository()
	zoomMeetingService := services.NewZoomMeetingService(meetingRepo, zoomAuthRepo)

	return &ZoomMeetingController{
		zoomMeetingService: zoomMeetingService,
	}
}

// POST /api/zoom/meetings
func (zmc *ZoomMeetingController) CreateMeeting(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req services.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request data", http.StatusBadRequest)
		return
	}

	meeting, err := zmc.zoomMeetingService.CreateMeeting(int64(userID.(int)), &req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meeting, nil, "Meeting created successfully")
}

// GET /api/zoom/meetings/:id
func (zmc *ZoomMeetingController) GetMeeting(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	meeting, err := zmc.zoomMeetingService.GetMeetingByID(id)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusNotFound)
		return
	}

	utils.Respond(c, meeting, nil, "")
}

// GET /api/zoom/meetings
func (zmc *ZoomMeetingController) GetUserMeetings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	limit, _ := strconv.Atoi(limitStr)
	page, _ := strconv.Atoi(pageStr)
	offset := (page - 1) * limit

	meetings, err := zmc.zoomMeetingService.GetUserMeetings(int64(userID.(int)), limit, offset)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"meetings": meetings,
		"page":     page,
		"limit":    limit,
	}, nil, "")
}

// GET /api/courses/:id/zoom-meetings
func (zmc *ZoomMeetingController) GetCourseMeetings(c *gin.Context) {
	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid course ID", http.StatusBadRequest)
		return
	}

	meetings, err := zmc.zoomMeetingService.GetCourseMeetings(courseID)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meetings, nil, "")
}

// GET /api/lessons/:id/zoom-meetings
func (zmc *ZoomMeetingController) GetLessonMeetings(c *gin.Context) {
	lessonIDStr := c.Param("id")
	lessonID, err := strconv.ParseInt(lessonIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	meetings, err := zmc.zoomMeetingService.GetLessonMeetings(lessonID)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meetings, nil, "")
}

// PUT /api/zoom/meetings/:id
func (zmc *ZoomMeetingController) UpdateMeeting(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	var req services.UpdateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request data", http.StatusBadRequest)
		return
	}

	meeting, err := zmc.zoomMeetingService.UpdateMeeting(id, &req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meeting, nil, "Meeting updated successfully")
}

// DELETE /api/zoom/meetings/:id
func (zmc *ZoomMeetingController) DeleteMeeting(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	err = zmc.zoomMeetingService.DeleteMeeting(id)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"message": "Meeting deleted successfully",
	}, nil, "")
}

// POST /api/zoom/meetings/:id/start
func (zmc *ZoomMeetingController) StartMeeting(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid meeting ID", http.StatusBadRequest)
		return
	}

	response, err := zmc.zoomMeetingService.StartMeeting(id)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, response, nil, "Meeting started successfully")
}

// POST /api/courses/:id/zoom-meeting
func (zmc *ZoomMeetingController) CreateCourseMeeting(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	courseIDStr := c.Param("id")
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var req services.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Set course ID
	req.CourseID = &courseID

	meeting, err := zmc.zoomMeetingService.CreateMeeting(int64(userID.(int)), &req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meeting, nil, "Course meeting created successfully")
}

// POST /api/lessons/:id/zoom-meeting
func (zmc *ZoomMeetingController) CreateLessonMeeting(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lessonIDStr := c.Param("id")
	lessonID, err := strconv.ParseInt(lessonIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	var req services.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Set lesson ID
	req.LessonID = &lessonID

	meeting, err := zmc.zoomMeetingService.CreateMeeting(int64(userID.(int)), &req)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, meeting, nil, "Lesson meeting created successfully")
}
