package services

import (
	"be-lms/config"
	"be-lms/dto"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

type PowerPointService interface {
	FileExists(filePath string) bool
	GetPowerPointData(c *gin.Context, folder string) (map[string]interface{}, error)
	GetContentType(filename string) string
	ScanByPath(c *gin.Context) error
	// S3 methods
	GetFileFromS3(key string) ([]byte, error)
	FileExistsInS3(key string) bool
	GetS3ObjectInfo(key string) (*S3ObjectInfo, error)
}

type S3ObjectInfo struct {
	Size         int64
	LastModified string
	ContentType  string
}

type powerPointService struct{}

func NewPowerPointService() PowerPointService {
	return &powerPointService{}
}

// Helper function to create S3 client
func (s *powerPointService) getS3Client() (*s3.Client, string, error) {
	s3Disk := config.Disks[config.S3]

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		if s3Disk.Endpoint != "" {
			ep := aws.Endpoint{
				URL:               s3Disk.Endpoint,
				SigningRegion:     s3Disk.Region,
				HostnameImmutable: true,
			}
			return ep, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(s3Disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Disk.Key, s3Disk.Secret, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, "", err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Disk.PathStyle
	})

	return client, s3Disk.Bucket, nil
}

func (s *powerPointService) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (s *powerPointService) GetPowerPointData(c *gin.Context, folder string) (map[string]interface{}, error) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	baseUrl := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	powerPointUrl := fmt.Sprintf("%s/power-point/%s", baseUrl, folder)

	// Get the title from the folder name or use a default
	title := folder
	if title == "" {
		title = "PowerPoint Presentation"
	}

	data := map[string]interface{}{
		"title":           title,
		"folder":          folder,
		"base_url":        baseUrl,
		"power_point_url": powerPointUrl,
		"data_path":       fmt.Sprintf("/power_point/%s/data", folder),
	}

	return data, nil
}

func (s *powerPointService) GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".xml":
		return "application/xml; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func (s *powerPointService) ScanByPath(c *gin.Context) error {
	path := c.Param("path")

	config.Log.Info("ScanByPath on S3: ", path)

	if path == "" {
		return errors.New("missing path parameter")
	}

	client, bucket, err := s.getS3Client()
	if err != nil {
		config.Log.Error("getS3Client: ", err)
		return err
	}

	// Ensure prefix ends with slash
	s3Prefix := fmt.Sprintf("power_point/%s/", strings.Trim(path, "/"))

	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(s3Prefix),
	})

	hasIndex := make(map[string]bool)
	hasData := make(map[string]bool)
	sizeByRoot := make(map[string]int64)

	for paginator.HasMorePages() {
		page, pErr := paginator.NextPage(context.TODO())
		if pErr != nil {
			config.Log.Error("ListObjectsV2: ", pErr)
			return pErr
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if strings.HasSuffix(key, "/index.html") {
				root := strings.TrimSuffix(key, "/index.html")
				hasIndex[root] = true
				sizeByRoot[root] += aws.ToInt64(obj.Size)
			}
			if idx := strings.Index(key, "/data/"); idx != -1 {
				root := key[:idx]
				hasData[root] = true
				sizeByRoot[root] += aws.ToInt64(obj.Size)
			}
		}
	}

	var packages []string
	for root := range hasIndex {
		if hasData[root] {
			packages = append(packages, root)
		}
	}

	config.Log.Info("ScanByPath S3 found packages: ", len(packages))

	uploadSvc := NewPowerPointUploadService()
	for _, fullRoot := range packages {
		rel := strings.TrimPrefix(fullRoot, "power_point/")
		rel = strings.Trim(rel, "/")
		fileSize := sizeByRoot[fullRoot]
		now := time.Now().Format("2006-01-02 15:04:05")

		pp := dto.PowerPoint{
			FolderName: rel,
			FileSize:   fileSize,
			Filename:   "index.html",
			CreatedAt:  now,
			StaticURL:  fmt.Sprintf("/power-point/%s/index.html", rel),
		}

		if _, saveErr := uploadSvc.SaveMediaWithoutMove(pp); saveErr != nil {
			config.Log.Error("SaveMediaWithoutMove: ", saveErr)
			return saveErr
		}
	}

	return nil
}

// S3 Methods
func (s *powerPointService) GetFileFromS3(key string) ([]byte, error) {
	client, bucket, err := s.getS3Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

func (s *powerPointService) FileExistsInS3(key string) bool {
	client, bucket, err := s.getS3Client()
	if err != nil {
		return false
	}

	_, err = client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return err == nil
}

func (s *powerPointService) GetS3ObjectInfo(key string) (*S3ObjectInfo, error) {
	client, bucket, err := s.getS3Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	result, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info from S3: %w", err)
	}

	info := &S3ObjectInfo{
		Size:        aws.ToInt64(result.ContentLength),
		ContentType: aws.ToString(result.ContentType),
	}

	if result.LastModified != nil {
		info.LastModified = result.LastModified.Format("2006-01-02 15:04:05")
	}

	return info, nil
}
