package controllers

import (
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"
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
