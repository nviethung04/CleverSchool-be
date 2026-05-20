package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"be-Clever School/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// MigrationController quản lý việc migrate / upload file
type MigrationController struct{}

// FileMigrationResult lưu kết quả upload từng file
type FileMigrationResult struct {
	FilePath string `json:"file_path"`
	S3Key    string `json:"s3_key,omitempty"`
	S3URL    string `json:"s3_url,omitempty"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// getPublicDir trả về đường dẫn tuyệt đối đến thư mục public
func (mc *MigrationController) getPublicDir() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "public")
}

// UploadPublicDirectory - Upload toàn bộ thư mục public lên S3
func (mc *MigrationController) UploadPublicDirectory(c *gin.Context) {
	publicDir := mc.getPublicDir()
	s3Disk := config.Disks[config.S3]

	// Load AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(s3Disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s3Disk.Key,
			s3Disk.Secret,
			"",
		)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AWS config error: " + err.Error()})
		return
	}

	client := s3.NewFromConfig(awsCfg)
	var results []FileMigrationResult

	// Walk qua toàn bộ thư mục public
	err = filepath.WalkDir(publicDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		// Lấy relative path
		relPath, err := filepath.Rel(publicDir, path)
		if err != nil {
			return err
		}

		// Prefix "public/" để upload vào folder public trên S3
		s3Key := "public/" + filepath.ToSlash(relPath)

		// Mở file thay vì đọc toàn bộ vào RAM
		file, err := os.Open(path)
		if err != nil {
			results = append(results, FileMigrationResult{
				FilePath: relPath,
				Success:  false,
				Error:    err.Error(),
			})
			return nil
		}
		defer file.Close()

		info, _ := d.Info()

		// Upload lên S3
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(s3Disk.Bucket),
			Key:    aws.String(s3Key),
			Body:   file,
		})

		if err != nil {
			results = append(results, FileMigrationResult{
				FilePath: relPath,
				Success:  false,
				Error:    err.Error(),
			})
		} else {
			results = append(results, FileMigrationResult{
				FilePath: relPath,
				S3Key:    s3Key,
				S3URL:    fmt.Sprintf("%s/%s", strings.TrimRight(s3Disk.URL, "/"), s3Key),
				Success:  true,
				FileSize: info.Size(),
			})
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Uploaded %d files from public/", len(results)),
		"results": results,
	})
}
