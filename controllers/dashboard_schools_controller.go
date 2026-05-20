package controllers

import (
	"be-Clever School/dto"
	"be-Clever School/services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type DashboardSchoolsController struct {
	dashboardSchoolsService services.DashboardSchoolsService
}

func NewDashboardSchoolsController(dashboardSchoolsService services.DashboardSchoolsService) *DashboardSchoolsController {
	return &DashboardSchoolsController{
		dashboardSchoolsService: dashboardSchoolsService,
	}
}

// GetDashboardSchools lấy danh sách dashboard schools
func (c *DashboardSchoolsController) GetDashboardSchools(ctx *gin.Context) {
	// Parse start_date and end_date from query parameters (Unix timestamp)
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	
	var startDate, endDate string
	// Convert Unix timestamp to YYYY-MM-DD format
	if startDateStr != "" {
		if timestamp, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			startDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}
	if endDateStr != "" {
		if timestamp, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			endDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}

	// Get dashboard schools với điều kiện cố định
	schools, err := c.dashboardSchoolsService.GetDashboardSchools(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get dashboard schools",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    schools,
		"message": "success",
	})
}

// ExportDashboardSchools xuất file Excel dashboard schools
func (c *DashboardSchoolsController) ExportDashboardSchools(ctx *gin.Context) {
	// Parse start_date and end_date from query parameters (Unix timestamp)
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	
	var startDate, endDate string
	// Convert Unix timestamp to YYYY-MM-DD format
	if startDateStr != "" {
		if timestamp, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			startDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}
	if endDateStr != "" {
		if timestamp, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			endDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}

	// Get dashboard schools data
	schools, err := c.dashboardSchoolsService.GetDashboardSchools(startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get dashboard schools data",
			"error":   err.Error(),
		})
		return
	}

	// Set response headers for Excel file
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=dashboard_schools.xlsx")

	// Create Excel file
	err = c.createExcelFile(ctx, schools)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create Excel file",
			"error":   err.Error(),
		})
		return
	}
}

// createExcelFile tạo file Excel từ dữ liệu schools
func (c *DashboardSchoolsController) createExcelFile(ctx *gin.Context, schools []dto.DashboardSchoolsResponse) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set sheet name
	sheetName := "Dashboard Schools"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{
		"School ID", "School Name", "Total Students", "Total Teachers",
		"Active Teachers (Sep 8)", "Active Students (Sep 15)", "Students Completed Homework (Sep 15)",
		"Students Completed Homework (Selected Week)", "Active Students (Selected Week)", "Active Teachers (Selected Week)",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for row, school := range schools {
		rowIndex := row + 2 // Start from row 2 (after headers)
		
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), school.SchoolID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), school.SchoolName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), school.TotalStudents)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), school.TotalTeachers)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), school.ActiveTeachersFromSep8)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), school.ActiveStudentsFromSep15)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), school.StudentsCompletedHomeworkFromSep15)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), school.StudentsCompletedHomeworkSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), school.ActiveStudentsSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIndex), school.ActiveTeachersSelectedWeek)
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c:%c", 'A'+i, 'A'+i)
		f.SetColWidth(sheetName, col, col, 15)
	}

	// Write to response
	return f.Write(ctx.Writer)
}
