package controllers

import (
	"be-lms/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// ShowUploadPage hiển thị trang upload file
func (c *UploadController) ShowUploadPage(ctx *gin.Context) {
	cfg := config.LoadConfig()
	ctx.HTML(http.StatusOK, "upload.html", gin.H{
		"title":     "Upload File - LMS Enspire",
		"apiDomain": cfg.APIDomain,
	})
}

// GetUploadStats trả về thống kê upload (có thể mở rộng sau)
func (c *UploadController) GetUploadStats(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Upload stats retrieved successfully",
		"data": gin.H{
			"total_uploads": 0,
			"successful":    0,
			"failed":        0,
		},
	})
}
