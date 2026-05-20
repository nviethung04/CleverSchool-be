package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type GoogleMeetingAttendanceController struct {
	service services.GoogleMeetingAttendanceService
	meeting repositories.GoogleMeetingRepository
}

func NewGoogleMeetingAttendanceController() *GoogleMeetingAttendanceController {
	return &GoogleMeetingAttendanceController{
		service: services.NewGoogleMeetingAttendanceService(),
		meeting: repositories.NewGoogleMeetingRepository(),
	}
}

func (c *GoogleMeetingAttendanceController) JoinByShortCode(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	shortCode := ctx.Param("code")
	if shortCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	deviceInfo := ctx.GetHeader("User-Agent")
	ipAddress := ctx.ClientIP()

	roleId := utils.GetCurrentRoleId(ctx)
	meeting, attendance, err := c.service.JoinByShortCode(shortCode, uint(userID), roleId, deviceInfo, ipAddress)
	if err != nil {
		config.Log.Errorf("❌ Failed to join google meeting: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Successfully joined meeting",
		"meeting":    meeting,
		"attendance": attendance,
		"join_url":   meeting.MeetURL,
	})
}

func (c *GoogleMeetingAttendanceController) JoinMeeting(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	deviceInfo := ctx.GetHeader("User-Agent")
	ipAddress := ctx.ClientIP()

	attendance, err := c.service.JoinMeeting(uint(meetingID), uint(userID), "direct", deviceInfo, ipAddress)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Attendance recorded",
		"attendance": attendance,
	})
}

func (c *GoogleMeetingAttendanceController) LeaveMeeting(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	if err := c.service.LeaveMeeting(uint(meetingID), uint(userID)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Left meeting successfully"})
}

func (c *GoogleMeetingAttendanceController) GetMeetingAttendances(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	attendances, err := c.service.GetMeetingAttendances(uint(meetingID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	summary, _ := c.service.GetAttendanceSummary(uint(meetingID))

	ctx.JSON(http.StatusOK, gin.H{
		"attendances": attendances,
		"summary":     summary,
	})
}

func (c *GoogleMeetingAttendanceController) GetAttendanceSummary(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	summary, err := c.service.GetAttendanceSummary(uint(meetingID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"summary": summary})
}

func (c *GoogleMeetingAttendanceController) ExportAttendancesExcel(ctx *gin.Context) {
	meetingID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	meeting, err := c.meeting.GetByID(uint(meetingID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
		return
	}

	attendances, err := c.service.GetMeetingAttendances(uint(meetingID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Attendances"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"Email", "Full name", "Course", "Lesson", "Date", "Time"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	courseName := ""
	if meeting.CourseID != nil && meeting.Course != nil {
		courseName = meeting.Course.Name
	}
	lessonName := ""
	if meeting.LessonID != nil && meeting.Lesson != nil {
		if meeting.Lesson.ObjectTitle != "" {
			lessonName = meeting.Lesson.ObjectTitle
		} else {
			lessonName = meeting.Lesson.Title
		}
	}

	for row, a := range attendances {
		rowIndex := row + 2
		joinedAt := a.JoinedAt

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), a.User.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), a.User.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), courseName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), lessonName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), joinedAt.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), joinedAt.Format("15:04:05"))
	}

	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c:%c", 'A'+i, 'A'+i)
		f.SetColWidth(sheetName, col, col, 22)
	}

	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	baseFileName := fmt.Sprintf("meet_attendance_%d_%s.xlsx", meeting.ID, time.Now().Format("20060102_150405"))
	baseFileName = strings.ReplaceAll(baseFileName, "/", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "\\", "_")
	baseFileName = strings.ReplaceAll(baseFileName, ":", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "*", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "?", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "\"", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "<", "_")
	baseFileName = strings.ReplaceAll(baseFileName, ">", "_")
	baseFileName = strings.ReplaceAll(baseFileName, "|", "_")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", baseFileName))
	ctx.Header("Content-Transfer-Encoding", "binary")

	if err := f.Write(ctx.Writer); err != nil {
		config.Log.Errorf("❌ Failed to export attendances excel: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export excel"})
		return
	}
}

