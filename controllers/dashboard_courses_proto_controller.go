package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardCoursesProtoController struct {
	dashboardCoursesProtoService services.DashboardCoursesProtoService
}

func NewDashboardCoursesProtoController(dashboardCoursesProtoService services.DashboardCoursesProtoService) *DashboardCoursesProtoController {
	return &DashboardCoursesProtoController{
		dashboardCoursesProtoService: dashboardCoursesProtoService,
	}
}

// GetDashboardCoursesProto lấy danh sách dashboard courses với protobuf
func (c *DashboardCoursesProtoController) GetDashboardCoursesProto(ctx *gin.Context) {
	// Parse query parameters
	req := &prot.DashboardCoursesRequest{}

	// Parse school_id
	if schoolIDStr := ctx.Query("school_id"); schoolIDStr != "" {
		if schoolID, err := strconv.ParseInt(schoolIDStr, 10, 64); err == nil {
			req.SchoolId = schoolID
		}
	}

	// Parse teacher_id
	if teacherIDStr := ctx.Query("teacher_id"); teacherIDStr != "" {
		if teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64); err == nil {
			req.TeacherId = teacherID
		}
	}

	// Parse course_id
	if courseIDStr := ctx.Query("course_id"); courseIDStr != "" {
		if courseID, err := strconv.ParseInt(courseIDStr, 10, 64); err == nil {
			req.CourseId = courseID
		}
	}

	// Parse search
	req.Search = ctx.Query("search")

	// Parse start_date and end_date (Unix timestamp)
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	
	// Convert Unix timestamp to YYYY-MM-DD format
	if startDateStr != "" {
		if timestamp, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			req.StartDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}
	if endDateStr != "" {
		if timestamp, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
			req.EndDate = time.Unix(timestamp, 0).Format("2006-01-02")
		}
	}

	// Parse pagination
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.ParseInt(pageStr, 10, 32); err == nil {
			req.Page = int32(page)
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			req.Limit = int32(limit)
		}
	}

	// Get data from service
	response, err := c.dashboardCoursesProtoService.GetDashboardCoursesProto(req)
	if err != nil {
		utils.Respond(ctx, nil, err, "Failed to get dashboard courses", 500)
		return
	}

	// Return protobuf response
	utils.Respond(ctx, response, nil, "success", 200)
}

// GetCourseStudentsProto lấy danh sách học sinh trong course kèm thông tin homework
func (c *DashboardCoursesProtoController) GetCourseStudentsProto(ctx *gin.Context) {
	req := &prot.DashboardCourseStudentsRequest{}

	// Parse course_id
	if courseIDStr := ctx.Query("course_id"); courseIDStr != "" {
		if courseID, err := strconv.ParseInt(courseIDStr, 10, 64); err == nil {
			req.CourseId = courseID
		} else {
			utils.Respond(ctx, nil, err, "Invalid course_id", 400)
			return
		}
	} else {
		utils.Respond(ctx, nil, nil, "course_id is required", 400)
		return
	}

	// Parse start_date and end_date
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" {
		utils.Respond(ctx, nil, nil, "start_date is required", 400)
		return
	}
	if endDateStr == "" {
		utils.Respond(ctx, nil, nil, "end_date is required", 400)
		return
	}

	// Try parse Unix timestamp first
	var startDate, endDate string
	if timestamp, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
		startDate = time.Unix(timestamp, 0).Format("2006-01-02")
	} else {
		// Use as-is if not a timestamp (assuming YYYY-MM-DD format)
		startDate = startDateStr
	}

	if timestamp, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
		endDate = time.Unix(timestamp, 0).Format("2006-01-02")
	} else {
		// Use as-is if not a timestamp (assuming YYYY-MM-DD format)
		endDate = endDateStr
	}

	req.StartDate = startDate
	req.EndDate = endDate

	// Get data from service
	response, err := c.dashboardCoursesProtoService.GetCourseStudentsProto(req)
	if err != nil {
		utils.Respond(ctx, nil, err, "Failed to get course students", 500)
		return
	}

	// Return protobuf response
	utils.Respond(ctx, response, nil, "success", 200)
}

// GetCourseHomeworksProto trả về danh sách homework của course
func (c *DashboardCoursesProtoController) GetCourseHomeworksProto(ctx *gin.Context) {
	req := &prot.DashboardCourseHomeworksRequest{}

	// course_id bắt buộc
	courseIDStr := ctx.Query("course_id")
	if courseIDStr == "" {
		utils.Respond(ctx, nil, nil, "course_id is required", 400)
		return
	}
	courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, err, "Invalid course_id", 400)
		return
	}
	req.CourseId = courseID

	// end_date bắt buộc
	endDateStr := ctx.Query("end_date")
	if endDateStr == "" {
		utils.Respond(ctx, nil, nil, "end_date is required", 400)
		return
	}
	if timestamp, err := strconv.ParseInt(endDateStr, 10, 64); err == nil {
		req.EndDate = time.Unix(timestamp, 0).Format("2006-01-02")
	} else {
		req.EndDate = endDateStr
	}

	// start_date tùy chọn
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if timestamp, err := strconv.ParseInt(startDateStr, 10, 64); err == nil {
			req.StartDate = time.Unix(timestamp, 0).Format("2006-01-02")
		} else {
			req.StartDate = startDateStr
		}
	}

	response, err := c.dashboardCoursesProtoService.GetCourseHomeworksProto(req)
	if err != nil {
		utils.Respond(ctx, nil, err, "Failed to get course homeworks", 500)
		return
	}

	utils.Respond(ctx, response, nil, "success", 200)
}