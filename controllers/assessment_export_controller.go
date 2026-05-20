package controllers

import (
	"be-Clever School/requests"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AssessmentExportController struct {
	svc services.AssessmentExportService
}

func NewAssessmentExportController() *AssessmentExportController {
	return &AssessmentExportController{
		svc: services.NewAssessmentExportService(),
	}
}

// ExportExcel xuất file Excel cho assessment scores
// GET /api/manage/assessments/export-excel?course_ids=1,2,3,4&assessment_id=xxx
func (ctl *AssessmentExportController) ExportExcel(c *gin.Context) {
	var req requests.AssessmentExportExcelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "Invalid parameters",
			"error":   err.Error(),
		})
		return
	}

	if req.CourseIDs == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "course_ids is required",
		})
		return
	}

	if req.AssessmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "assessment_id is required and must be greater than 0",
		})
		return
	}

	err := ctl.svc.ExportExcel(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ImportExcel nhập file Excel đã điền dữ liệu
// POST /api/manage/assessments/import-excel
// Form data: file (multipart/form-data)
func (ctl *AssessmentExportController) ImportExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, err, "File is required", http.StatusBadRequest)
		return
	}

	// Kiểm tra file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		utils.Respond(c, nil, nil, "File must be .xlsx format", http.StatusBadRequest)
		return
	}

	// Mở file
	src, err := file.Open()
	if err != nil {
		utils.Respond(c, nil, err, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer src.Close()

	// Gọi service để xử lý import
	result, err := ctl.svc.ImportExcel(c, src)
	if err != nil {
		utils.Respond(c, nil, err, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.Respond(c, result, nil, "")
}

