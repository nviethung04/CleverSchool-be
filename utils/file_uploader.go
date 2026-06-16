package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"be-lms/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Url          string `json:"url"`
	FileName     string `json:"file_name"`
	FileSize     string `json:"file_size"`
	FileMimeType string `json:"file_mime_type"`
	FileExt      string `json:"file_extension"`
	FilePath     string `json:"file_path"`
	UploadDir    string
}

type FileUploader struct {
	Path      string
	Request   string
	Storage   string
	MaxSizeMB int64
	AllowMime map[string][]string
}

func NewFileUploader(path, request string) *FileUploader {
	return &FileUploader{
		Path:      path,
		Request:   request,
		Storage:   "local",
		MaxSizeMB: 512000,
		AllowMime: map[string][]string{
			"h5p":   {"h5p"},
			"image": {"jpeg", "png", "jpg", "gif", "webp", "svg", "bmp"},
			"video": {"avi", "mov", "wmv", "mp4", "3gp", "flv"},
			"audio": {"mp3", "wav", "aac", "ogg", "flac", "m4a"},
			"file":  {"avi", "mov", "wmv", "mp4", "3gp", "flv", "pdf", "txt", "csv", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "zip", "rar", "7z", "webm", "html", "htm"},
		},
	}
}

func NewUploaderWithStorage(path, request, storage string) *FileUploader {
	fu := NewFileUploader(path, request)
	fu.Storage = storage
	return fu
}

func (fu *FileUploader) Store(c *gin.Context) (*FileInfo, error) {
	config.Log.Info("Store...")
	file, err := c.FormFile(fu.Request)
	if err != nil {
		config.Log.Errorf("failed to read uploaded file: %w", err)
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	config.Log.Info("originalName...")
	originalName := filepath.Base(filepath.Clean(file.Filename))
	ext := filepath.Ext(originalName)

	if !fu.validateFile(file) {
		config.Log.Error("file type or size not allowe")
		return nil, fmt.Errorf("file type or size not allowed")
	}

	filename := generateSafeFilename(originalName)

	disk, ok := config.Disks[fu.Storage]
	if !ok {
		return nil, fmt.Errorf("storage disk not found: %s", fu.Storage)
	}

	storagePath := filepath.Join(disk.Root, fu.Path)
	var fileURL, fullPath, fileFullPath string

	switch disk.Driver {
	case "local":
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}

		fullPath = filepath.Join(storagePath, filename)

		if err := saveLocal(file, fullPath); err != nil {
			return nil, fmt.Errorf("failed to save file locally: %w", err)
		}

		fullPath = filepath.Join("public", fullPath)
		relative := filepath.ToSlash(strings.TrimPrefix(fullPath, disk.Root))
		fileFullPath = filename
		fileURL = strings.TrimRight(disk.URL, "/") + "/" + strings.TrimLeft(relative, "/")
	case "scorm":
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}

		fullPath = filepath.Join(storagePath, filename)

		if err := saveLocal(file, fullPath); err != nil {
			return nil, fmt.Errorf("failed to save file locally: %w", err)
		}

		fullPath = filepath.Join("scorm", fullPath)
		relative := filepath.ToSlash(strings.TrimPrefix(fullPath, disk.Root))
		fileFullPath = filename
		fileURL = strings.TrimRight(disk.URL, "/") + "/" + strings.TrimLeft(relative, "/")

	case "s3":
		config.Log.Info("save to s3...")
		key := filepath.ToSlash(filepath.Join(fu.Path, filename))
		fileURL, err = s3Upload(file, key, disk)
		if err != nil {
			config.Log.Errorf("failed to upload file to s3: %w", err)
			return nil, fmt.Errorf("failed to upload file to s3: %w", err)
		}

		fullPath = filepath.Join(storagePath, filename)
		fileFullPath = filename

	default:
		return nil, fmt.Errorf("unsupported storage: %s", fu.Storage)
	}

	return &FileInfo{
		Url:          fileURL,
		FileName:     fileFullPath,
		FileSize:     formatBytes(file.Size),
		FileMimeType: file.Header.Get("Content-Type"),
		FileExt:      ext,
		FilePath:     fu.Path,
		UploadDir:    storagePath,
	}, nil
}

func (fu *FileUploader) StoreFileFromBytes(fileData []byte, originalFilename string) (*FileInfo, error) {
	if len(fileData) == 0 {
		return nil, fmt.Errorf("file data is empty")
	}

	if !fu.validateFileData(fileData) {
		return nil, fmt.Errorf("file type or size not allowed")
	}

	if originalFilename == "" {
		originalFilename = "file"
	}
	originalFilename = filepath.Base(originalFilename)

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d-%s", timestamp, originalFilename)

	disk, ok := config.Disks[fu.Storage]
	if !ok {
		return nil, fmt.Errorf("storage disk not found: %s", fu.Storage)
	}

	var fileURL string

	switch disk.Driver {
	case "local":
		storagePath := filepath.Join(disk.Root, fu.Path)
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		fullPath := filepath.Join(storagePath, filename)
		if err := saveLocalFromBytes(fileData, fullPath); err != nil {
			return nil, fmt.Errorf("failed to save file locally: %w", err)
		}

		fullPath = filepath.Join("public", fullPath)
		relative := strings.TrimPrefix(fullPath, disk.Root)
		fileURL = filepath.ToSlash(filepath.Join(disk.URL, relative))

	case "s3":
		key := filepath.ToSlash(filepath.Join(fu.Path, filename))
		url, err := s3UploadFromBytes(fileData, key, disk)
		if err != nil {
			return nil, fmt.Errorf("failed to upload to s3: %w", err)
		}
		fileURL = url

	default:
		return nil, fmt.Errorf("unsupported storage: %s", fu.Storage)
	}

	return &FileInfo{
		Url:          fileURL,
		FileName:     filename,
		FileSize:     formatBytes(int64(len(fileData))),
		FileMimeType: detectMimeType(fileData),
		FileExt:      filepath.Ext(originalFilename),
		FilePath:     fu.Path,
	}, nil
}

func (fu *FileUploader) DeleteFile(filename string) error {
	disk, ok := config.Disks[fu.Storage]
	if !ok {
		return fmt.Errorf("storage disk not found: %s", fu.Storage)
	}

	switch disk.Driver {
	case "local":
		config.Log.Info("Remove in storage...")
		fullPath := filepath.Join(disk.Root, fu.Path, filename)
		if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil

	case "s3":
		key := filepath.ToSlash(filepath.Join(fu.Path, filename))

		client, err := config.NewS3Client(disk)
		if err != nil {
			return err
		}

		_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(disk.Bucket),
			Key:    aws.String(key),
		})
		return err

	default:
		return fmt.Errorf("unsupported storage: %s", fu.Storage)
	}
}

func (fu *FileUploader) DeleteFolder() error {
	disk, ok := config.Disks[fu.Storage]
	if !ok {
		return fmt.Errorf("storage disk not found: %s", fu.Storage)
	}

	switch disk.Driver {
	case "local":
		fullPath := filepath.Join(disk.Root, fu.Path)

		dirEntries, err := os.ReadDir(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if len(dirEntries) > 0 {
			return fmt.Errorf("cannot delete folder: directory is not empty")
		}

		if err := os.Remove(fullPath); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("delete folder not supported for storage: %s", fu.Storage)
	}
}

func saveLocal(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func saveLocalFromBytes(fileData []byte, fullPath string) error {
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ghi dữ liệu vào file
	_, err = file.Write(fileData)
	return err
}

func s3Upload(file *multipart.FileHeader, key string, disk config.DiskConfig) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return "", err
	}

	client, err := config.NewS3Client(disk)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Upload file lên S3 / R2
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(disk.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(mimeType),
		// Nếu bucket bật "Block all public access" thì bỏ ACL đi
		// ACL: types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to s3: %w", err)
	}

	return fmt.Sprintf("%s/%s", strings.TrimRight(disk.URL, "/"), key), nil
}

func s3UploadFromBytes(fileData []byte, key string, disk config.DiskConfig) (string, error) {
	client, err := config.NewS3Client(disk)
	if err != nil {
		return "", err
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(disk.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileData),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", strings.TrimRight(disk.URL, "/"), key), nil
}

func (fu *FileUploader) validateFile(file *multipart.FileHeader) bool {
	if file.Size <= 0 {
		return false
	}
	sizeMB := file.Size / 1024 / 1024
	if sizeMB > fu.MaxSizeMB {
		return false
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	for _, allowed := range fu.AllowMime {
		for _, e := range allowed {
			if ext == e || strings.Contains(ext, e) {
				return true
			}
		}
	}

	// Fallback when browser omits extension (e.g. blob re-upload, clipboard paste).
	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if strings.HasPrefix(contentType, "image/") {
		return true
	}

	return false
}

func (fu *FileUploader) validateFileData(fileData []byte) bool {
	return len(fileData) <= int(fu.MaxSizeMB)*1024*1024
}

func formatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	}
}

func generateSafeFilename(original string) string {
	safeName := strings.ToLower(strings.ReplaceAll(original, " ", "_"))
	ext := filepath.Ext(safeName)

	if ext == ".zip" {
		return safeName
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d-%s", timestamp, safeName)
}

func detectMimeType(data []byte) string {
	return http.DetectContentType(data)
}
