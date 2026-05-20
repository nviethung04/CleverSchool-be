package controllers

import (
	"be-Clever School/dto"
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type DashboardTeacherHomeworkStudentController struct {
	svc services.DashboardTeacherHomeworkStudentService
}

func NewDashboardTeacherHomeworkStudentController() *DashboardTeacherHomeworkStudentController {
	return &DashboardTeacherHomeworkStudentController{
		svc: services.NewDashboardTeacherHomeworkStudentService(),
	}
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkStudent(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Get homework_id from URL parameter
	homeworkIDStr := c.Param("homework_id")
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Set the homework_id from URL parameter
	req.HomeworkID = homeworkID

	resp, err := ctl.svc.GetStudents(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardTeacherHomeworkStudentController) ExportTeacherHomeworkStudent(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Get homework_id from URL parameter
	homeworkIDStr := c.Param("homework_id")
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Set the homework_id from URL parameter
	req.HomeworkID = homeworkID

	// Set Limit = 0 để lấy tất cả dữ liệu (không phân trang)
	req.Limit = 0
	req.Page = 0

	// Gọi service để lấy dữ liệu
	resp, err := ctl.svc.GetStudents(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get student data",
			"error":   err.Error(),
		})
		return
	}

	// Convert protobuf response to DTO
	var students []dto.DashboardTeacherHomeworkStudent
	for _, student := range resp.Students {
		students = append(students, dto.DashboardTeacherHomeworkStudent{
			StudentID:          student.GetStudentId(),
			StudentName:        student.GetStudentName(),
			TotalQuestions:     student.GetTotalQuestions(),
			QuestionsCompleted: student.GetQuestionsCompleted(),
		})
	}

	// Set response headers for Excel file
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=teacher_homework_student.xlsx")

	// Create Excel file
	err = ctl.createExcelFileForStudent(c, students)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create Excel file",
			"error":   err.Error(),
		})
		return
	}
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkStats(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetStudentStats(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardTeacherHomeworkStudentController) DashboardTeacherHomeworkOverview(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkOverviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	resp, err := ctl.svc.GetHomeworkOverview(c, &req)
	utils.Respond(c, resp, err, "")
}

func (ctl *DashboardTeacherHomeworkStudentController) ExportTeacherHomeworkStats(c *gin.Context) {
	var req requests.DashboardTeacherHomeworkStudentStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	// Set Limit = 0 để lấy tất cả dữ liệu (không phân trang)
	req.Limit = 0
	req.Page = 0

	// Gọi service để lấy dữ liệu
	resp, err := ctl.svc.GetStudentStats(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get student stats data",
			"error":   err.Error(),
		})
		return
	}

	// Convert protobuf response to DTO
	var students []dto.DashboardTeacherHomeworkStudentStats
	for _, student := range resp.Students {
		students = append(students, dto.DashboardTeacherHomeworkStudentStats{
			StudentID:              student.GetStudentId(),
			StudentName:            student.GetStudentName(),
			TotalHomework:          student.GetTotalHomework(),
			TotalAssignedHomeworks: student.GetTotalAssignedHomeworks(),
			InProgressHomework:     student.GetInProgressHomework(),
			CompletedHomework:      student.GetCompletedHomework(),
			NotStartedHomework:     student.GetNotStartedHomework(),
			AverageRatio:           student.GetAverageRatio(),
			CompletedHomeworkRatio: student.GetCompletedHomeworkRatio(),
		})
	}

	// Set response headers for Excel file
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=teacher_homework_stats.xlsx")

	// Create Excel file
	err = ctl.createExcelFile(c, students)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create Excel file",
			"error":   err.Error(),
		})
		return
	}
}

// createExcelFile tạo file Excel từ dữ liệu student stats
func (ctl *DashboardTeacherHomeworkStudentController) createExcelFile(ctx *gin.Context, students []dto.DashboardTeacherHomeworkStudentStats) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set sheet name
	sheetName := "Teacher Homework Stats"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{
		"Student ID",
		"Student Name",
		"Total Homework",
		"Total Assigned Homeworks",
		"In Progress Homework",
		"Completed Homework",
		"Not Started Homework",
		"Average Ratio",
		"Completed Homework Ratio",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for row, student := range students {
		rowIndex := row + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), student.StudentID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), student.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), student.TotalHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), student.TotalAssignedHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), student.InProgressHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), student.CompletedHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), student.NotStartedHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), student.AverageRatio)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), student.CompletedHomeworkRatio)
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c:%c", 'A'+i, 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Write to response
	return f.Write(ctx.Writer)
}

// createExcelFileForStudent tạo file Excel từ dữ liệu student (cho API homework/{homework_id}/export)
func (ctl *DashboardTeacherHomeworkStudentController) createExcelFileForStudent(ctx *gin.Context, students []dto.DashboardTeacherHomeworkStudent) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set sheet name
	sheetName := "Teacher Homework Student"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{
		"Student ID",
		"Student Name",
		"Total Questions",
		"Questions Completed",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for row, student := range students {
		rowIndex := row + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), student.StudentID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), student.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), student.TotalQuestions)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), student.QuestionsCompleted)
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c:%c", 'A'+i, 'A'+i)
		f.SetColWidth(sheetName, col, col, 20)
	}

	// Write to response
	return f.Write(ctx.Writer)
}
