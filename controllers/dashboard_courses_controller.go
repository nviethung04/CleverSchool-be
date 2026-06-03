package controllers

import (
	"be-lms/dto"
	"be-lms/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type DashboardCoursesController struct {
	dashboardCoursesService services.DashboardCoursesService
}

func NewDashboardCoursesController(dashboardCoursesService services.DashboardCoursesService) *DashboardCoursesController {
	return &DashboardCoursesController{
		dashboardCoursesService: dashboardCoursesService,
	}
}

// GetDashboardCourses lấy danh sách dashboard courses
func (c *DashboardCoursesController) GetDashboardCourses(ctx *gin.Context) {
	// Parse school_id from query parameter
	schoolIDStr := ctx.Query("school_id")
	schoolID := int64(0)

	if schoolIDStr != "" {
		if id, err := strconv.ParseInt(schoolIDStr, 10, 64); err == nil {
			schoolID = id
		}
	}

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

	// Get dashboard courses
	courses, err := c.dashboardCoursesService.GetDashboardCourses(schoolID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get dashboard courses",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    courses,
		"message": "success",
	})
}

// ExportDashboardCourses xuất file Excel dashboard courses
func (c *DashboardCoursesController) ExportDashboardCourses(ctx *gin.Context) {
	// Parse school_id from query parameter
	schoolID := int64(0)
	
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

	// Get dashboard courses data
	courses, err := c.dashboardCoursesService.GetDashboardCourses(schoolID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get dashboard courses data",
			"error":   err.Error(),
		})
		return
	}

	// Set response headers for Excel file
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=dashboard_courses.xlsx")

	// Create Excel file
	err = c.createExcelFile(ctx, courses)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create Excel file",
			"error":   err.Error(),
		})
		return
	}
}

// createExcelFile tạo file Excel từ dữ liệu courses
func (c *DashboardCoursesController) createExcelFile(ctx *gin.Context, courses []dto.DashboardCoursesResponse) error {
	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set sheet name
	sheetName := "Dashboard Courses"
	f.SetSheetName("Sheet1", sheetName)

	// Set headers
	headers := []string{
		"ID", "Course ID", "Course Name", "Object Title", "School ID", "School Name", "Total Students", "Total Teachers", "Teacher IDs", "Teacher Infos",
		"Active Teachers (Sep 8)", "Active Students (Sep 15)", "Students Completed Homework (Sep 15)",
		"Students Completed Homework (Selected Week)", "Active Students (Selected Week)", "Active Teachers (Selected Week)",
		"Total Homeworks", "Assigned Homeworks", "Completed Homeworks",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for row, course := range courses {
		rowIndex := row + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), course.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), course.CourseID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), course.CourseName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), course.ObjectTitle)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), course.SchoolID)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), course.SchoolName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), course.TotalStudents)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), course.TotalTeachers)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), course.TeacherIDs)
		// Convert TeacherInfos to JSON string for Excel
		teacherInfosJSON, _ := json.Marshal(course.TeacherInfos)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIndex), string(teacherInfosJSON))
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), course.ActiveTeachersFromSep8)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIndex), course.ActiveStudentsFromSep15)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIndex), course.StudentsCompletedHomeworkFromSep15)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIndex), course.StudentsCompletedHomeworkSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIndex), course.ActiveStudentsSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIndex), course.ActiveTeachersSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIndex), course.TotalHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", rowIndex), course.AssignedHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("S%d", rowIndex), course.CompletedHomeworks)
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := fmt.Sprintf("%c:%c", 'A'+i, 'A'+i)
		f.SetColWidth(sheetName, col, col, 15)
	}

	// Write to response
	return f.Write(ctx.Writer)
}

// ExportDashboardCoursesAll xuất tất cả dữ liệu dashboard courses ra file Excel (không phân trang)
func (c *DashboardCoursesController) ExportDashboardCoursesAll(ctx *gin.Context) {
	// Parse school_id from query parameter
	schoolIDStr := ctx.Query("school_id")
	schoolID := int64(0)

	if schoolIDStr != "" {
		if id, err := strconv.ParseInt(schoolIDStr, 10, 64); err == nil {
			schoolID = id
		}
	}

    // Parse start_date and end_date from query parameters (Unix timestamp -> YYYY-MM-DD)
    startDateStr := ctx.Query("start_date")
    endDateStr := ctx.Query("end_date")
    var startDate, endDate string
    if startDateStr != "" {
        if ts, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
            startDate = time.Unix(ts, 0).Format("2006-01-02")
        }
    }
    if endDateStr != "" {
        if ts, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
            endDate = time.Unix(ts, 0).Format("2006-01-02")
        }
    }

	// Get all dashboard courses (không phân trang)
	courses, err := c.dashboardCoursesService.GetDashboardCourses(schoolID, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to get dashboard courses for export",
			"error":   err.Error(),
		})
		return
	}

	// Set response headers for Excel file
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=dashboard_courses_all.xlsx")

	// Create Excel file
	err = c.createExcelFile(ctx, courses)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Failed to create Excel file",
			"error":   err.Error(),
		})
		return
	}
}
