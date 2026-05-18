package controllers

import (
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type PowerPointController struct {
	svc services.PowerPointService
}

func NewPowerPointController(svc services.PowerPointService) *PowerPointController {
	return &PowerPointController{svc: svc}
}

func (pc *PowerPointController) HandlePowerPoint(c *gin.Context) {
	folder := c.Param("folder")

	folder = pc.FixURL(folder)

	// Get the PowerPoint HTML file path from S3
	s3Key := fmt.Sprintf("power_point/%s/index.html", folder)

	// Check if file exists in S3
	if !pc.svc.FileExistsInS3(s3Key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "PowerPoint file not found"})
		return
	}

	// Read the PowerPoint HTML file from S3
	content, err := pc.svc.GetFileFromS3(s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read PowerPoint file from S3"})
		return
	}

	// Replace relative paths with API paths for data
	htmlContent := string(content)

	// Replace various asset path patterns
	assetBaseURL := "/power-point/" + folder + "/data/"

	// Replace src="data/..." with src="/power-point/{folder}/data/..."
	htmlContent = strings.ReplaceAll(htmlContent, `src="data/`, `src="`+assetBaseURL)

	// Replace href="data/..." with href="/power-point/{folder}/data/..."
	htmlContent = strings.ReplaceAll(htmlContent, `href="data/`, `href="`+assetBaseURL)

	// Replace any remaining relative data/ paths using regex
	re := regexp.MustCompile(`(["\'])data/([^"\']+)`)
	htmlContent = re.ReplaceAllString(htmlContent, `$1`+assetBaseURL+`$2`)

	// Serve the HTML content directly
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func (pc *PowerPointController) ServePowerPointAssets(c *gin.Context) {
	folder := c.Param("folder")
	assetPath := c.Param("filepath")

	// Get the asset file path from S3
	s3Key := fmt.Sprintf("power_point/%s/data/%s", folder, assetPath)

	// Check if file exists in S3
	if !pc.svc.FileExistsInS3(s3Key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset file not found"})
		return
	}

	// Get file info from S3
	objectInfo, err := pc.svc.GetS3ObjectInfo(s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get asset file info"})
		return
	}

	// Get file content from S3
	content, err := pc.svc.GetFileFromS3(s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read asset file from S3"})
		return
	}

	// Set headers
	c.Header("Content-Length", fmt.Sprintf("%d", objectInfo.Size))

	// Get content type
	contentType := pc.svc.GetContentType(assetPath)
	if objectInfo.ContentType != "" {
		contentType = objectInfo.ContentType
	}
	c.Header("Content-Type", contentType)

	// Serve the content
	c.Data(http.StatusOK, contentType, content)
}

// ServeDynamicPowerPoint supports nested folders like /power-point/a/b/c/index.html and /power-point/a/b/c/data/...
func (pc *PowerPointController) ServeDynamicPowerPoint(c *gin.Context) {
	any := c.Param("any")
	any = strings.TrimPrefix(any, "/")

	if strings.HasSuffix(any, "/index.html") || any == "index.html" {
		folder := strings.TrimSuffix(any, "/index.html")
		s3Key := fmt.Sprintf("power_point/%s/index.html", folder)

		if !pc.svc.FileExistsInS3(s3Key) {
			c.JSON(http.StatusNotFound, gin.H{"error": "PowerPoint file not found"})
			return
		}

		content, err := pc.svc.GetFileFromS3(s3Key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read PowerPoint file from S3"})
			return
		}

		htmlContent := string(content)
		assetBaseURL := "/power-point/" + folder + "/data/"
		htmlContent = strings.ReplaceAll(htmlContent, `src="data/`, `src="`+assetBaseURL)
		htmlContent = strings.ReplaceAll(htmlContent, `href="data/`, `href="`+assetBaseURL)
		re := regexp.MustCompile(`(["\'])data/([^"\']+)`)
		htmlContent = re.ReplaceAllString(htmlContent, `$1`+assetBaseURL+`$2`)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
		return
	}

	idx := strings.Index(strings.ToLower(any), "/data/")
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid path"})
		return
	}
	folder := any[:idx]
	subPath := any[idx+len("/data/"):]
	s3Key := fmt.Sprintf("power_point/%s/data/%s", folder, subPath)

	if !pc.svc.FileExistsInS3(s3Key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Asset file not found"})
		return
	}

	// Get file info from S3
	objectInfo, err := pc.svc.GetS3ObjectInfo(s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get asset file info"})
		return
	}

	// Get file content from S3
	content, err := pc.svc.GetFileFromS3(s3Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read asset file from S3"})
		return
	}

	// Set headers
	c.Header("Content-Length", fmt.Sprintf("%d", objectInfo.Size))

	// Get content type
	contentType := pc.svc.GetContentType(subPath)
	if objectInfo.ContentType != "" {
		contentType = objectInfo.ContentType
	}
	c.Header("Content-Type", contentType)

	// Serve the content
	c.Data(http.StatusOK, contentType, content)
}

func (pc *PowerPointController) ScanByPath(c *gin.Context) {
	err := pc.svc.ScanByPath(c)
	utils.Respond(c, nil, err, "")
}

func (pc *PowerPointController)  FixURL(s string) string {
	// Đổi dấu + thành %2B trước
	s = strings.ReplaceAll(s, "+", "%2B")
	// Đổi %20 thành +
	s = strings.ReplaceAll(s, "%20", "+")
	return s
}
