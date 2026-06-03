package controllers

import (
	"net/http"

	"be-lms/config"
	"be-lms/services"

	"github.com/gin-gonic/gin"
)

type UploadS3Controller struct{
	service services.UploadService
}

func NewUploadS3Controller() *UploadS3Controller {
	service := services.NewUploadService()
	return &UploadS3Controller{
		service: service,
	}
}

func (uc *UploadS3Controller) PresignUpload(c *gin.Context) {
	if uc.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "service not initialized"})
		return
	}
	resp, err := uc.service.PresignUpload(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (uc *UploadS3Controller) UploadComplete(c *gin.Context) {
	resp, err := uc.service.UploadComplete(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (uc *UploadS3Controller) Extract(c *gin.Context) {
	resp, err := uc.service.Extract(c, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (uc *UploadS3Controller) ShowUploadPage(c *gin.Context) {
	cfg := config.LoadConfig()
	c.HTML(http.StatusOK, "upload_s3.html", gin.H{
		"title":     "Upload File - LMS Enspire",
		"apiDomain": cfg.APIDomain,
	})
}
