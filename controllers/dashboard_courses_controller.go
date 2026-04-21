package controllers

import (
	"be-lms/dto"
	"be-lms/repositories"
	"be-lms/services"
	_ "encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type DashboardCoursesController struct {
	dashboardCoursesService services.DashboardCoursesService
	dashboardCoursesRepo    repositories.DashboardCoursesRepository
}

func NewDashboardCoursesController(dashboardCoursesService services.DashboardCoursesService) *DashboardCoursesController {
	return &DashboardCoursesController{
		dashboardCoursesService: dashboardCoursesService,
		dashboardCoursesRepo:    repositories.NewDashboardCoursesRepository(),
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
		"Course ID", "Course Name", "Object Title", "School ID", "School Name", "Total Students", "Total Teachers", "Teacher IDs", "Teacher Username", "Teacher Name",
		"Active Teachers", "Active Students", "Students Completed Homework",
		"Total Homeworks", "Assigned Homeworks",
		"Students Completed All Homeworks", "Students Doing Homeworks", "Students Not Started Any Homework",
		"Student Active Not Started Any Homework", "Homework Over 50% Student Complete",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for row, course := range courses {
		rowIndex := row + 2 // Start from row 2 (after headers)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), course.CourseID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), course.CourseName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), course.ObjectTitle)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), course.SchoolID)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), course.SchoolName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), course.TotalStudents)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), course.TotalTeachers)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), course.TeacherIDs)
		// Convert TeacherInfos to comma-separated strings
		var teacherUsernames []string
		var teacherNames []string
		for _, teacher := range course.TeacherInfos {
			teacherUsernames = append(teacherUsernames, teacher.Username)
			teacherNames = append(teacherNames, teacher.Name)
		}
		teacherUsernameStr := ""
		teacherNameStr := ""
		if len(teacherUsernames) > 0 {
			teacherUsernameStr = strings.Join(teacherUsernames, ",")
			teacherNameStr = strings.Join(teacherNames, ",")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), teacherUsernameStr)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIndex), teacherNameStr)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), course.ActiveTeachersSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowIndex), course.ActiveStudentsSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowIndex), course.StudentsCompletedHomeworkSelectedWeek)
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowIndex), course.TotalHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", rowIndex), course.AssignedHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("P%d", rowIndex), course.StudentsCompletedAllHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", rowIndex), course.StudentsDoingHomeworks)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", rowIndex), course.StudentsNotStartedAnyHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("S%d", rowIndex), course.StudentActiveNotStartedAnyHomework)
		f.SetCellValue(sheetName, fmt.Sprintf("T%d", rowIndex), course.HomeworkOver50PercentStudentComplete)
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

	// Parse teacher_id from query parameter
	teacherIDStr := ctx.Query("teacher_id")
	teacherID := int64(0)
	if teacherIDStr != "" {
		if id, err := strconv.ParseInt(teacherIDStr, 10, 64); err == nil {
			teacherID = id
		}
	}

	// Parse course_id from query parameter
	courseIDStr := ctx.Query("course_id")
	courseID := int64(0)
	if courseIDStr != "" {
		if id, err := strconv.ParseInt(courseIDStr, 10, 64); err == nil {
			courseID = id
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

	// Get all dashboard courses với filter (không phân trang) - gọi repository trực tiếp
	courses, err := c.dashboardCoursesRepo.GetDashboardCourses(schoolID, teacherID, courseID, startDate, endDate)
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
