package routes

import (
	"be-cleverschool/middleware"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func RegisterLogRoute(r *gin.Engine) {
	authGroup := r.Group("/", middleware.BaseMiddleware())
	authGroup.GET("/log", func(c *gin.Context) {
		logPath := c.Query("path")

		// Load AWS config (bắt buộc có AWS_REGION)
		awsCfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Không thể load AWS config",
				"detail": err.Error(),
			})
			return
		}
		s3Client := s3.NewFromConfig(awsCfg)

		bucket := os.Getenv("AWS_BUCKET")
		if bucket == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Chưa cấu hình AWS_BUCKET"})
			return
		}

		// Domain để filter
		domain := os.Getenv("API_DOMAIN")
		if domain == "" {
			domain = c.Request.Host
		}
		domain = ExtractDomainName(domain)

		prefix := os.Getenv("LOG_PREFIX")
		if prefix == "" {
			prefix = "logs/"
		}

		// Filter theo domain
		domainPrefix := prefix + domain

		// Nếu không truyền path → list file theo domain
		if logPath == "" {
			resp, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
				Bucket: aws.String(bucket),
				Prefix: aws.String(domainPrefix),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Không thể list logs từ S3",
					"detail": err.Error(),
				})
				return
			}

			if len(resp.Contents) == 0 {
				c.JSON(http.StatusOK, gin.H{
					"message": "Không tìm thấy file nào",
					"bucket":  bucket,
					"prefix":  domainPrefix,
				})
				return
			}

			var logs []gin.H
			for _, obj := range resp.Contents {
				if filepath.Ext(*obj.Key) == ".log" {
					logs = append(logs, gin.H{
						"key":     *obj.Key,
						"name":    filepath.Base(*obj.Key),
						"size":    obj.Size,
						"modTime": obj.LastModified.Format(time.RFC3339),
					})
				}
			}

			c.JSON(http.StatusOK, logs)
			return
		}

		// Có path → tải file từ S3
		if filepath.Ext(logPath) != ".log" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Đường dẫn không hợp lệ"})
			return
		}

		// key = prefix + domain + "-" + file
		key := filepath.Join(prefix, filepath.Base(logPath))
		if !strings.HasPrefix(key, domainPrefix) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bạn không có quyền xem log của domain khác"})
			return
		}

		obj, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":  "Không tìm thấy file log trên S3",
				"bucket": bucket,
				"key":    key,
				"detail": err.Error(),
			})
			return
		}
		defer obj.Body.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, obj.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể đọc file log từ S3"})
			return
		}

		c.Data(http.StatusOK, "text/plain; charset=utf-8", buf.Bytes())
	})
}

func ExtractDomainName(rawURL string) string {
	u, err := url.Parse(rawURL)
	var host string
	if err != nil || u.Hostname() == "" {
		host = strings.TrimPrefix(rawURL, "http://")
		host = strings.TrimPrefix(host, "https://")
		host = strings.Split(host, "/")[0]
		host = strings.Split(host, ":")[0]
	} else {
		host = u.Hostname()
	}

	if idx := strings.Index(host, "-api"); idx != -1 {
		host = host[:idx]
	} else {
		if dot := strings.Index(host, "."); dot != -1 {
			host = host[:dot]
		}
	}

	return strings.ReplaceAll(host, "-", "_")
}

