package controllers

import (
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardSchoolsProtoController struct {
	dashboardSchoolsProtoService services.DashboardSchoolsProtoService
}

func NewDashboardSchoolsProtoController(dashboardSchoolsProtoService services.DashboardSchoolsProtoService) *DashboardSchoolsProtoController {
	return &DashboardSchoolsProtoController{
		dashboardSchoolsProtoService: dashboardSchoolsProtoService,
	}
}

// GetDashboardSchoolsProto lấy danh sách dashboard schools với phân trang và tìm kiếm, trả về protobuf
func (c *DashboardSchoolsProtoController) GetDashboardSchoolsProto(ctx *gin.Context) {
	req := &prot.DashboardSchoolsRequest{}

	// Parse search string from query parameter
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

	// Parse page and limit from query parameters
	pageStr := ctx.Query("page")
	if pageStr != "" {
		if page, err := strconv.ParseInt(pageStr, 10, 32); err == nil {
			req.Page = int32(page)
		}
	}
	limitStr := ctx.Query("limit")
	if limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			req.Limit = int32(limit)
		}
	}

	// Get data from service
	response, err := c.dashboardSchoolsProtoService.GetDashboardSchoolsProto(req)
	if err != nil {
		utils.Respond(ctx, nil, err, "Failed to get dashboard schools", 500)
		return
	}

	// Return protobuf response
	utils.Respond(ctx, response, nil, "success", 200)
}
