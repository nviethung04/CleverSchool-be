package services

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/redis"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PowerPointUploadService interface {
	IsValidArchiveFile(filename string) bool
	IsValidFolderName(folderName string) bool
	GenerateFolderName(filename string) string
	FolderExists(folderName string) bool
	UploadAndExtract(file *multipart.FileHeader, folderName string) (*dto.PowerPoint, error)
	UploadAndExtractFromPath(srcArchivePath string, folderName string) (*dto.PowerPoint, error)
	FindPackages(folderName string) ([]dto.PowerPoint, error)
	ListPresentations() ([]dto.PowerPointInfo, error)
	DeletePresentation(folderName string) error
	GetPresentationInfo(folderName string) (*dto.PowerPointInfo, error)
	SaveMedia(file dto.PowerPoint) (*prot.File, error)
	SaveMediaWithoutMove(file dto.PowerPoint) (*prot.File, error)
	// S3 upload methods
	UploadFolderToS3(localFolderPath, s3Prefix string) error
	DeleteLocalFolder(folderPath string) error
	MoveFolderToS3(localFolderPath, s3Prefix string) error
}

type powerPointUploadService struct {
	basePath     string
	mediaService MediaService
}

func NewPowerPointUploadService() PowerPointUploadService {
	mediaRepo := repositories.NewMediaRepository()
	ms := NewMediaService(mediaRepo)

	return &powerPointUploadService{
		basePath:     "power_point",
		mediaService: ms,
	}
}

func (s *powerPointUploadService) IsValidArchiveFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".zip"
}

func (s *powerPointUploadService) IsValidFolderName(folderName string) bool {
	// Allow letters, numbers, hyphens, and underscores only
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, folderName)
	return matched && len(folderName) > 0 && len(folderName) <= 150
}

func (s *powerPointUploadService) GenerateFolderName(filename string) string {
	// Remove extension and replace spaces/special chars with hyphens
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove special characters except hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", name, timestamp)
}

func (s *powerPointUploadService) FolderExists(folderName string) bool {
	folderPath := filepath.Join(s.basePath, folderName)
	_, err := os.Stat(folderPath)
	return !os.IsNotExist(err)
}

func (s *powerPointUploadService) UploadAndExtract(file *multipart.FileHeader, folderName string) (*dto.PowerPoint, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		config.Log.Error("MkdirAll: ", err)
		return nil, fmt.Errorf("failed to create base directory")
	}

	// Create folder for this presentation
	folderPath := filepath.Join(s.basePath, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		config.Log.Error("MkdirAll: ", err)
		return nil, fmt.Errorf("failed to create folder")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		config.Log.Error("Open: ", err)
		return nil, fmt.Errorf("failed to open uploaded file")
	}
	defer src.Close()

	// Create a temporary file to store the uploaded content
	ext := strings.ToLower(filepath.Ext(file.Filename))
	tempFile, err := os.CreateTemp("", "powerpoint-upload-*"+ext)
	if err != nil {
		config.Log.Error("CreateTemp: ", err)
		return nil, fmt.Errorf("failed to create temp file")
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded content to temp file
	if _, err := io.Copy(tempFile, src); err != nil {
		config.Log.Error("Copy: ", err)
		return nil, fmt.Errorf("failed to copy file content")
	}

	// Extract the archive file based on extension
	if ext == ".zip" {
		if err := s.mediaService.ExtractZip(tempFile.Name(), folderPath); err != nil {
			config.Log.Error("extractZip: ", err)
			return nil, fmt.Errorf("failed to extract ZIP file")
		}
	} else if ext == ".rar" {
		if err := s.mediaService.ExtractRar(tempFile.Name(), folderPath); err != nil {
			config.Log.Error("extractRar: ", err)
			return nil, fmt.Errorf("failed to extract RAR file")
		}
	} else {
		config.Log.Error("Unsupported archive format: ")
		return nil, fmt.Errorf("unsupported archive format")
	}

	// Get presentation info
	info, err := s.GetPresentationInfo(folderName)
	if err != nil {
		config.Log.Error("GetPresentationInfo: ", err)
		return nil, fmt.Errorf("failed to get presentation info")
	}

	return &dto.PowerPoint{
		FolderName: folderName,
		FileSize:   file.Size,
		Filename:   file.Filename,
		CreatedAt:  info.CreatedAt,
		StaticURL:  info.StaticURL,
	}, nil
}

func (s *powerPointUploadService) UploadAndExtractFromPath(srcArchivePath string, folderName string) (*dto.PowerPoint, error) {
	// Ensure base and target folder
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		config.Log.Error("MkdirAll:", err)
		return nil, fmt.Errorf("failed to create base directory")
	}
	folderPath := filepath.Join(s.basePath, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		config.Log.Error("MkdirAll:", err)
		return nil, fmt.Errorf("failed to create folder")
	}

	ext := strings.ToLower(filepath.Ext(srcArchivePath))
	switch ext {
	case ".zip":
		if err := s.mediaService.ExtractZip(srcArchivePath, folderPath); err != nil {
			config.Log.Error("extractZip:", err)
			return nil, fmt.Errorf("failed to extract ZIP file")
		}
	case ".rar":
		if err := s.mediaService.ExtractRar(srcArchivePath, folderPath); err != nil {
			config.Log.Error("extractRar:", err)
			return nil, fmt.Errorf("failed to extract RAR file")
		}
	default:
		config.Log.Error("Unsupported archive format:", ext)
		return nil, fmt.Errorf("unsupported archive format")
	}

	info, err := s.GetPresentationInfo(folderName)
	if err != nil {
		config.Log.Error("GetPresentationInfo:", err)
		return nil, fmt.Errorf("failed to get presentation info")
	}

	return &dto.PowerPoint{
		FolderName: folderName,
		FileSize:   info.FileSize,
		Filename:   filepath.Base(srcArchivePath),
		CreatedAt:  info.CreatedAt,
		StaticURL:  info.StaticURL,
	}, nil
}

// FindPackages scans the extracted directory for multiple PowerPoint packages.
// A package is identified as a directory that contains an index.html file and a sibling directory named "data".
func (s *powerPointUploadService) FindPackages(folderName string) ([]dto.PowerPoint, error) {
	rootPath := filepath.Join(s.basePath, folderName)

	// Ensure folder exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("folder not found")
	}

	// Walk and find all package roots
	pkgRelPaths, err := s.scanPackageDirs(rootPath)
	if err != nil {
		return nil, err
	}

	var results []dto.PowerPoint
	for _, rel := range pkgRelPaths {
		abs := rootPath
		relClean := rel
		if relClean != "." && relClean != "" {
			abs = filepath.Join(rootPath, relClean)
		}

		// Compose folder identifier relative to power_point base
		folderIdentifier := folderName
		if relClean != "." && relClean != "" {
			folderIdentifier = filepath.ToSlash(filepath.Join(folderName, relClean))
		}

		size := s.calculateFolderSize(abs)

		stat, statErr := os.Stat(abs)
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		if statErr == nil {
			createdAt = stat.ModTime().Format("2006-01-02 15:04:05")
		}

		staticURL := fmt.Sprintf("/power-point/%s/index.html", folderIdentifier)

		results = append(results, dto.PowerPoint{
			FolderName: folderIdentifier,
			FileSize:   size,
			Filename:   filepath.Base(abs),
			CreatedAt:  createdAt,
			StaticURL:  staticURL,
		})
	}

	return results, nil
}

// scanPackageDirs returns relative directories (to root) that look like a PPT package (index.html + data/)
func (s *powerPointUploadService) scanPackageDirs(root string) ([]string, error) {
	var packages []string
	seen := make(map[string]bool)

	walkErr := filepath.WalkDir(root, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		base := d.Name()
		if strings.HasPrefix(base, ".") || strings.EqualFold(base, "__MACOSX") {
			return nil
		}

		indexPath := filepath.Join(currentPath, "index.html")
		dataDir := filepath.Join(currentPath, "data")

		if fileExists(indexPath) && dirExists(dataDir) {
			rel, _ := filepath.Rel(root, currentPath)
			if rel == "" {
				rel = "."
			}
			if !seen[rel] {
				seen[rel] = true
				packages = append(packages, rel)
			}
			return nil
		}

		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}
	return packages, nil
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (s *powerPointUploadService) ListPresentations() ([]dto.PowerPointInfo, error) {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, err
	}

	var presentations []dto.PowerPointInfo
	for _, entry := range entries {
		if entry.IsDir() {
			info, err := s.GetPresentationInfo(entry.Name())
			if err == nil {
				presentations = append(presentations, *info)
			}
		}
	}

	return presentations, nil
}

func (s *powerPointUploadService) DeletePresentation(folderPath string) error {
	return os.RemoveAll(folderPath)
}

func (s *powerPointUploadService) GetPresentationInfo(folderName string) (*dto.PowerPointInfo, error) {
	folderPath := filepath.Join(s.basePath, folderName)

	// Check if folder exists
	stat, err := os.Stat(folderPath)
	if err != nil {
		return nil, err
	}

	// Get total size
	totalSize := s.calculateFolderSize(folderPath)

	return &dto.PowerPointInfo{
		FolderName: folderName,
		FileSize:   totalSize,
		CreatedAt:  stat.ModTime().Format("2006-01-02 15:04:05"),
		StaticURL:  fmt.Sprintf("/power-point/%s/index.html", folderName),
	}, nil
}

func (s *powerPointUploadService) calculateFolderSize(folderPath string) int64 {
	var size int64
	filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func (s *powerPointUploadService) SaveMediaWithoutMove(file dto.PowerPoint) (*prot.File, error) {
	mediaRepo := repositories.NewMediaRepository()

	// Use power_point as disk and root folder
	diskName := config.Public
	rootPath := "power_point"

	parentZero := int64(0)

	// Ensure root folder exists on this disk
	parentRootFolder, err := mediaRepo.FindFolderByFilePathAndParent("", int64(0), diskName)

	if err == nil && parentRootFolder.ID != 0 {
		parentZero = parentRootFolder.ID
	}

	rootFolder, err := mediaRepo.FindFolderByFilePathAndParent(rootPath, parentZero, diskName)

	if err != nil || rootFolder == nil || rootFolder.ID == 0 {
		// fallback to legacy full_path lookup
		rootFolder, _ = mediaRepo.FindByPathMaster(rootPath, diskName)
		if rootFolder == nil || rootFolder.ID == 0 {
			zero := int64(0)
			candidate := &models.Media{
				ParentID: &parentZero,
				FileName: "",
				FilePath: rootPath,
				FullPath: rootPath,
				FileSize: &zero,
				DiskName: &diskName,
				Type:     "folder",
			}
			if err := mediaRepo.UpdateOrCreate(candidate); err != nil {
				return &prot.File{}, fmt.Errorf("failed to create power_point root folder")
			}
			// re-fetch to ensure we have ID
			rootFolder, _ = mediaRepo.FindFolderByFilePathAndParentMaster(rootPath, parentZero, diskName)
			if rootFolder == nil || rootFolder.ID == 0 {
				rootFolder, _ = mediaRepo.FindByPathMaster(rootPath, diskName)
			}
		}
	}

	if rootFolder == nil || rootFolder.ID == 0 {
		return &prot.File{}, fmt.Errorf("failed to resolve power_point root folder")
	}

	// Build folder hierarchy for folder segments in file.FolderName
	segments := strings.Split(strings.Trim(file.FolderName, "/"), "/")
	current := rootFolder
	currentFullPath := rootFolder.FullPath

	for _, seg := range segments {
		seg = strings.Trim(seg, "/")
		if seg == "" {
			continue
		}
		nextFullPath := strings.Trim(currentFullPath+"/"+seg, "/")
		// Prefer lookup by (file_path,parent_id)
		existing, _ := mediaRepo.FindFolderByFilePathAndParent(seg, current.ID, diskName)
		if existing != nil && existing.ID != 0 {
			current = existing
			currentFullPath = nextFullPath
			continue
		}
		// Fallback by full_path
		existing, _ = mediaRepo.FindByPath(nextFullPath, diskName)
		if existing != nil && existing.ID != 0 {
			current = existing
			currentFullPath = nextFullPath
			continue
		}

		parentID := current.ID
		zero := int64(0)
		folderNode := &models.Media{
			ParentID: &parentID,
			FileName: "",
			FilePath: seg,
			FullPath: nextFullPath,
			FileSize: &zero,
			DiskName: &diskName,
			Type:     "folder",
		}
		if err := mediaRepo.UpdateOrCreate(folderNode); err != nil {
			return &prot.File{}, fmt.Errorf("failed to create folder segment")
		}
		// Re-fetch to get the actual ID
		current, _ = mediaRepo.FindFolderByFilePathAndParentMaster(seg, parentID, diskName)
		if current == nil || current.ID == 0 {
			current, _ = mediaRepo.FindByPathMaster(nextFullPath, diskName)
		}
		if current == nil || current.ID == 0 {
			return &prot.File{}, fmt.Errorf("failed to resolve folder segment")
		}
		currentFullPath = nextFullPath
	}

	// Prepare file media entry under deepest folder
	staticURL := utils.StaticURL(file.StaticURL, config.PowerPoint)
	stripped := utils.StripDomain(staticURL, models.Storage)
	fileUrl := &stripped

	powerPointDisk := config.PowerPoint

	fileName := filepath.Base(currentFullPath)

	fileExt := "ppt"
	fileType := "power_point"

	media := &models.Media{
		FolderID:      current.ID,
		FileName:      fileName,
		FilePath:      currentFullPath,
		FileType:      &fileType,
		FileSize:      &file.FileSize,
		FileExtension: &fileExt,
		DiskName:      &powerPointDisk,
		StaticURL:     fileUrl,
		Type:          "file",
	}

	if err := mediaRepo.UpdateOrCreateV2(media); err != nil {
		return &prot.File{}, fmt.Errorf("failed to create media")
	}

	joinedPath := filepath.Join(media.FilePath, media.FileName)

	redis.DeleteCache("medias:all-files")

	return &prot.File{
		Id:            int32(media.ID),
		FileName:      media.FileName,
		Url:           utils.StaticURL(utils.DerefStr(media.StaticURL), config.PowerPoint),
		FileSize:      strconv.FormatInt(utils.DerefInt64(media.FileSize), 10),
		FileMimeType:  utils.DerefStr(media.FileType),
		FileExtension: utils.DerefStr(media.FileExtension),
		DiskName:      utils.DerefStr(media.DiskName),
		Path:          utils.DerefStr(&joinedPath),
		Type:          media.Type,
	}, nil
}

func (s *powerPointUploadService) SaveMedia(file dto.PowerPoint) (*prot.File, error) {
	// First save media without moving folder
	result, err := s.SaveMediaWithoutMove(file)
	if err != nil {
		return result, err
	}

	// After saving media, move the local folder to S3
	localFolderPath := filepath.Join(s.basePath, file.FolderName)
	s3Prefix := fmt.Sprintf("power_point/%s", file.FolderName)

	// Check if local folder exists before trying to move
	if _, err := os.Stat(localFolderPath); err == nil {
		// Move folder to S3 in background (don't block the response)
		go func() {
			moveErr := s.MoveFolderToS3(localFolderPath, s3Prefix)
			if moveErr != nil {
				config.Log.Error(fmt.Sprintf("Failed to move folder %s to S3: %v", localFolderPath, moveErr))
			} else {
				config.Log.Info(fmt.Sprintf("Successfully moved PowerPoint folder %s to S3", file.FolderName))
			}
		}()
	}

	return result, nil
}

// Helper function to create S3 client
func (s *powerPointUploadService) getS3Client() (*s3.Client, string, error) {
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

// UploadFolderToS3 uploads entire folder to S3
func (s *powerPointUploadService) UploadFolderToS3(localFolderPath, s3Prefix string) error {
	client, bucket, err := s.getS3Client()
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Walk through all files in the folder
	err = filepath.Walk(localFolderPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path from localFolderPath
		relPath, err := filepath.Rel(localFolderPath, filePath)
		if err != nil {
			return err
		}

		// Create S3 key
		s3Key := filepath.ToSlash(filepath.Join(s3Prefix, relPath))

		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Upload to S3
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Key),
			Body:   file,
    		ContentType: aws.String(contentType),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s to S3: %w", filePath, err)
		}

		return nil
	})

	return err
}

// DeleteLocalFolder deletes local folder and all its contents
func (s *powerPointUploadService) DeleteLocalFolder(folderPath string) error {
	return os.RemoveAll(folderPath)
}

// MoveFolderToS3 uploads folder to S3 and then deletes local folder
func (s *powerPointUploadService) MoveFolderToS3(localFolderPath, s3Prefix string) error {
	// First upload to S3
	err := s.UploadFolderToS3(localFolderPath, s3Prefix)
	if err != nil {
		return fmt.Errorf("failed to upload folder to S3: %w", err)
	}

	// Then delete local folder
	// err = s.DeleteLocalFolder(localFolderPath)
	// if err != nil {
	// 	config.Log.Error(fmt.Sprintf("Failed to delete local folder %s after S3 upload: %v", localFolderPath, err))
	// 	// Don't return error here as the upload was successful
	// }

	config.Log.Info(fmt.Sprintf("Successfully moved folder %s to S3 prefix %s", localFolderPath, s3Prefix))
	return nil
}
