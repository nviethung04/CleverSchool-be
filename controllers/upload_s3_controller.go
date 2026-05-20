package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/repositories"
	"be-Clever School/services"

	"github.com/gin-gonic/gin"
)

type UploadS3Controller struct {
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
	var req struct {
		Key         string `json:"key"`
		URL         string `json:"url"`
		Filename    string `json:"filename"`
		Size        int64  `json:"size"`
		ContentType string `json:"content_type"`
		UserID      int64  `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Kiểm tra xem có phải file zip PowerPoint không
	ext := strings.ToLower(filepath.Ext(req.Filename))
	isPowerPointZip := ext == ".zip" && strings.Contains(req.Key, "public/power_point")

	if ext == ".zip" {
		// Đối với file zip PowerPoint, trả về response ngay và xử lý async
		// để tránh client timeout
		url := ConvertURL(req.URL)
		jobID := services.GenerateJobID(req.Key)

		extractReq := &dto.ExtractRequest{
			Filename:     req.Filename,
			Key:          req.Key,
			URL:          req.URL,
			Size:         req.Size,
			ContentType:  req.ContentType,
			UserID:       req.UserID,
			IsPowerPoint: isPowerPointZip,
		}

		go func() {
			// Tạo context mới cho background processing
			ctx := context.Background()
			newCtx := c.Copy()
			newCtx.Request = newCtx.Request.WithContext(ctx)

			_, err := uc.service.Extract(newCtx, extractReq)
			if err != nil {
				config.Log.Error("Background Extract failed: ", err.Error())
				services.UpdateProgress(jobID, services.StatusFailed, 0, fmt.Sprintf("Extraction failed: %v", err))
			} else {
				mediaRepo := repositories.NewMediaRepository()
				err = mediaRepo.ClearMedias()
				if err != nil {
					config.Log.Error("Failed to clear medias: ", err.Error())
				}

				config.Log.Info("Background Extract completed successfully for: ", req.Filename)
			}
		}()

		// Trả về response ngay để tránh client timeout
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Upload accepted, processing in background",
			"data": gin.H{
				"message":   "File is being processed in background",
				"final_url": url,
				"job_id":    jobID,
			},
		})
		return
	}

	// Đối với file thường, xử lý sync như cũ
	// Tạo request mới với body đã đọc để service có thể đọc lại
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Failed to serialize request"})
		return
	}

	// Tạo context mới với body đã đọc
	newCtx := c.Copy()
	// Clone request và set body mới
	newRequest := newCtx.Request.WithContext(context.Background())
	newRequest.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	newRequest.ContentLength = int64(len(bodyBytes))
	newCtx.Request = newRequest

	resp, err := uc.service.UploadComplete(newCtx)
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
		"title":     "Upload File - Clever School Clever School",
		"apiDomain": cfg.APIDomain,
	})
}

func ConvertURL(raw string) string {
	u := raw

	u = strings.Replace(u, "/public", "", 1)

	u = strings.TrimSuffix(u, ".zip")

	u += "/index.html"

	return u
}

// GetExtractProgress trả về progress của extraction job
func (uc *UploadS3Controller) GetExtractProgress(c *gin.Context) {
	// Lấy job_id từ query parameter thay vì path parameter
	// để tránh vấn đề với ký tự đặc biệt trong job_id
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "job_id query parameter is required"})
		return
	}

	progress, err := services.GetProgress(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    progress,
	})
}
