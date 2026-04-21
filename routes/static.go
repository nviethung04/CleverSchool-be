package routes

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

func InitStatic(router *gin.Engine) {
	// router.Static("/public", "./public") // không dùng nếu muốn custom xử lý
	// router.Static("/node_modules", "./node_modules")

	router.GET("/public/*filepath", StreamPublicFile)
	// router.GET("/scorm-static/*filepath", StreamScormFile) // Commented out to avoid conflict

	// Static files for templates
	router.Static("/assets", "./templates")
}

func StreamPublicFile(c *gin.Context) {
	relativePath := c.Param("filepath")
	relativePath = strings.TrimPrefix(relativePath, "/")
	fullPath := filepath.Join("public", filepath.Clean(relativePath))

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil || fileInfo.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}

	// Check nếu là ảnh và có tham số w hoặc h
	ext := strings.ToLower(filepath.Ext(fullPath))
	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
	widthStr := c.Query("w")
	heightStr := c.Query("h")

	if isImage && (widthStr != "" || heightStr != "") {
		srcImage, err := imaging.Open(fullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load image"})
			return
		}

		// Parse width & height
		var width, height int
		if widthStr != "" {
			width, _ = strconv.Atoi(widthStr)
		}
		if heightStr != "" {
			height, _ = strconv.Atoi(heightStr)
		}

		// Resize ảnh
		dstImage := imaging.Resize(srcImage, width, height, imaging.Lanczos)

		// Set headers
		c.Header("Content-Type", "image/png")
		c.Header("Cache-Control", "public, max-age=86400")

		// Encode và ghi ra response
		err = imaging.Encode(c.Writer, dstImage, imaging.PNG)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode image"})
		}
		return
	}

	// Nếu không resize, thì stream file gốc
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Last-Modified", fileInfo.ModTime().UTC().Format(time.RFC1123))

	http.ServeContent(c.Writer, c.Request, filepath.Base(fullPath), fileInfo.ModTime(), file)
}

func StreamScormFile(c *gin.Context) {
	relativePath := c.Param("filepath")
	relativePath = strings.TrimPrefix(relativePath, "/")
	fullPath := filepath.Join("scorm", filepath.Clean(relativePath))

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SCORM file not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open SCORM file"})
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil || fileInfo.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get SCORM file info"})
		return
	}

	// Get content type
	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Last-Modified", fileInfo.ModTime().UTC().Format(time.RFC1123))

	http.ServeContent(c.Writer, c.Request, filepath.Base(fullPath), fileInfo.ModTime(), file)
}
