package services

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"context"
	"fmt"
	"io"
	"mime"
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

type UploadService interface {
	PresignUpload(c *gin.Context) (*dto.PresignResponse, error)
	UploadComplete(c *gin.Context) (*dto.CompleteResponse, error)
	Extract(c *gin.Context, req *dto.ExtractRequest) (*dto.ExtractResponse, error)
}

type uploadService struct{}

func NewUploadService() UploadService {
	return &uploadService{}
}

func (s *uploadService) PresignUpload(c *gin.Context) (*dto.PresignResponse, error) {
	var req dto.PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		config.Log.Error(err)
		return nil, err
	}

	if req.ExpiresSecond <= 0 {
		req.ExpiresSecond = 15 * 60
	}

	filename := filepath.Base(req.Filename)

	if req.Folder != "" {
		filename = fmt.Sprintf("%s/%s", req.Folder, filename)
	}

	// **Thêm prefix public/** để upload vào folder public trong bucket
	key := fmt.Sprintf("public/%s", filename)

	_, presigner, bucket, err := s.NewS3AndPresigner()
	if err != nil {
		config.Log.Error("failed init s3: " + err.Error())
		return nil, err
	}

	// Build PutObjectInput
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if req.ContentType != "" {
		input.ContentType = aws.String(req.ContentType)
	} else {
		// optional: guess from extension
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".zip":
			input.ContentType = aws.String("application/zip")
		case ".pdf":
			input.ContentType = aws.String("application/pdf")
		case ".png":
			input.ContentType = aws.String("image/png")
		case ".jpg", ".jpeg":
			input.ContentType = aws.String("image/jpeg")
		}
	}

	// Presign PUT
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	presignedReq, err := presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(time.Duration(req.ExpiresSecond)*time.Second))
	if err != nil {
		config.Log.Error("failed presign: " + err.Error())
		return nil, err
	}

	// Construct object URL
	s3Disk := config.Disks[config.S3]
	// Thêm /public/ vào URL nếu cần
	objectURL := fmt.Sprintf("%s/%s", strings.TrimRight(s3Disk.URL, "/"), key)

	return &dto.PresignResponse{
		UploadURL: presignedReq.URL,
		ObjectURL: objectURL,
		Key:       key,
		ExpiresIn: req.ExpiresSecond,
	}, nil
}

func (s *uploadService) UploadComplete(c *gin.Context) (*dto.CompleteResponse, error) {
	// Tạo context mới từ Background() với timeout dài hơn (10 phút)
	// để tránh bị cancel bởi client timeout hoặc middleware
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var req dto.UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// Validate required fields
	if req.Key == "" || req.URL == "" || req.Filename == "" {
		return nil, fmt.Errorf("missing key/url/filename")
	}

	s3DiskName := config.S3

	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext == ".zip" && strings.Contains(req.Key, "public/power_point") {
		extractReq := &dto.ExtractRequest{
			Filename:     req.Filename,
			Key:          req.Key,
			URL:          req.URL,
			Size:         req.Size,
			ContentType:  req.ContentType,
			UserID:       req.UserID,
			IsPowerPoint: true,
		}
		// Truyền context với timeout dài vào Extract
		newCtx := c.Copy()
		newCtx.Request = c.Request.WithContext(ctx)
		res, err := s.Extract(newCtx, extractReq)

		if err != nil {
			config.Log.Error("Extract failed: ", req.URL, " error: ", err.Error())
			return nil, fmt.Errorf("extract failed: %w", err)
		}

		if res == nil {
			config.Log.Error("Extract returned nil response: ", req.URL)
			return nil, fmt.Errorf("extract returned nil response")
		}

		return &dto.CompleteResponse{
			Message:  res.Message,
			FinalUrl: utils.StaticURL(res.FinalUrl, s3DiskName),
		}, nil
	}

	diskName := config.Public // disk name, ví dụ "public"
	parentZero := int64(0)

	mediaRepo := repositories.NewMediaRepository()

	// Lấy folder root
	parentRootFolder, err := mediaRepo.FindFolderByFilePathAndParent("", parentZero, diskName)
	if err == nil && parentRootFolder != nil && parentRootFolder.ID != 0 {
		parentZero = parentRootFolder.ID
	}

	var current *models.Media
	var currentFullPath string
	if parentRootFolder != nil {
		current = parentRootFolder
		currentFullPath = parentRootFolder.FullPath
	} else {
		// Tạo root folder nếu chưa tồn tại
		zero := int64(0)
		rootFolder := &models.Media{
			ParentID: &zero,
			FileName: "",
			FilePath: "",
			FullPath: "",
			FileSize: &zero,
			DiskName: &diskName,
			Type:     "folder",
		}
		mediaRepo.UpdateOrCreate(rootFolder)
		current, _ = mediaRepo.FindFolderByFilePathAndParent("", parentZero, diskName)
		if current == nil {
			current = rootFolder
		}
		currentFullPath = current.FullPath
	}

	// Tạo folder theo key nếu chưa tồn tại
	keyParts := strings.Split(req.Key, "/")
	for index, part := range keyParts {
		if part == "public" || index == len(keyParts)-1 {
			continue
		}
		nextFullPath := strings.Trim(currentFullPath+"/"+part, "/")

		existing, _ := mediaRepo.FindFolderByFilePathAndParent(part, current.ID, diskName)
		if existing != nil && existing.ID != 0 {
			current = existing
			currentFullPath = nextFullPath
			continue
		}

		existing, _ = mediaRepo.FindByPath(nextFullPath, diskName)
		if existing != nil && existing.ID != 0 {
			current = existing
			currentFullPath = nextFullPath
			continue
		}

		zero := int64(0)
		folderNode := &models.Media{
			ParentID: &current.ID,
			FileName: "",
			FilePath: part,
			FullPath: nextFullPath,
			FileSize: &zero,
			DiskName: &diskName,
			Type:     "folder",
		}
		mediaRepo.UpdateOrCreate(folderNode)
		current, _ = mediaRepo.FindFolderByFilePathAndParent(part, current.ID, diskName)
		if current == nil {
			current = folderNode
		}
		currentFullPath = nextFullPath
	}

	fileSize := req.Size
	fileType := req.ContentType
	fullPath := req.Key
	fileName := req.Filename
	filePath := strings.TrimPrefix(filepath.Dir(fullPath), "public/")
	fileExt := filepath.Ext(fileName)
	staticURL := utils.StripDomain(req.URL, s3DiskName)

	media := &models.Media{
		FolderID:      current.ID,
		FileName:      fileName,
		FilePath:      filePath,
		FileType:      &fileType,
		FileSize:      &fileSize,
		FileExtension: &fileExt,
		DiskName:      &s3DiskName,
		StaticURL:     &staticURL,
		Type:          "file",
	}

	mediaRepo.UpdateOrCreateV2(media)

	return &dto.CompleteResponse{
			Message:  "success",
			FinalUrl: utils.StaticURL(req.URL, s3DiskName),
		},
		nil
}

func (s *uploadService) Extract(c *gin.Context, req *dto.ExtractRequest) (*dto.ExtractResponse, error) {
	// Context cho S3/upload: 60 phút để đủ cho zip lớn + nested zip (tránh "context deadline exceeded")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	if req == nil || req.Key == "" {
		var body dto.ExtractRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			config.Log.Error("Bind JSON error:", err)
			return nil, err
		}
		req = &body
	}

	config.Log.Info("Extract request: ", req.Key)
	config.Log.Info("Extract request: ", req.URL)
	config.Log.Info("Extract request: ", req.Filename)
	config.Log.Info("Extract request: ", req.IsPowerPoint)

	// Validate required fields
	if req.Key == "" || req.URL == "" || req.Filename == "" {
		config.Log.Error("missing key/url/filename")
		return nil, fmt.Errorf("missing key/url/filename")
	}

	// Tạo job ID và khởi tạo progress
	jobID := GenerateJobID(req.Key)
	_ = UpdateProgress(jobID, StatusPending, 0, "Starting extraction process...")

	// Chỉ xử lý file ZIP hoặc RAR
	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext != ".zip" && ext != ".rar" {
		config.Log.Error("only zip/rar files are supported")
		return nil, fmt.Errorf("only zip/rar files are supported")
	}

	diskName := "public"
	s3DiskName := config.S3
	mediaRepo := repositories.NewMediaRepository()

	// Lấy folder root trên DB
	parentZero := int64(0)
	parentRootFolder, err := mediaRepo.FindFolderByFilePathAndParentMaster("", parentZero, diskName)
	if err == nil && parentRootFolder != nil && parentRootFolder.ID != 0 {
		parentZero = parentRootFolder.ID
	}

	var current *models.Media
	if parentRootFolder != nil {
		current = parentRootFolder
	} else {
		// Tạo root folder nếu chưa tồn tại
		zero := int64(0)
		rootFolder := &models.Media{
			ParentID: &zero,
			FileName: "",
			FilePath: "",
			FullPath: "",
			FileSize: &zero,
			DiskName: &diskName,
			Type:     "folder",
		}
		mediaRepo.UpdateOrCreate(rootFolder)
		current, _ = mediaRepo.FindFolderByFilePathAndParentMaster("", parentZero, diskName)
		if current == nil {
			current = rootFolder
		}
	}

	// Lấy base name và S3 prefix từ req.Key
	baseName := strings.TrimSuffix(filepath.Base(req.Key), ext)
	s3Dir := filepath.Dir(req.Key)
	s3Prefix := filepath.ToSlash(filepath.Join(s3Dir, baseName))

	// Local extract dir - nếu là public/power_point thì extract vào thư mục power_point
	var localExtract string
	var shouldCleanup bool
	if strings.Contains(req.Key, "public/power_point") {
		// Extract vào thư mục power_point
		localExtract = filepath.Join("power_point", baseName)
		shouldCleanup = false // Không xóa thư mục power_point sau khi xong
	} else {
		// Extract vào temp directory như cũ
		localExtract = filepath.Join(os.TempDir(), baseName)
		shouldCleanup = true           // Xóa temp directory sau khi xong
		_ = os.RemoveAll(localExtract) // xoá temp directory nếu tồn tại
	}

	if err := os.MkdirAll(localExtract, os.ModePerm); err != nil {
		return nil, err
	}

	if shouldCleanup {
		defer os.RemoveAll(localExtract) // cleanup temp directory khi xong
	}

	// Khởi tạo client S3
	_ = UpdateProgress(jobID, StatusDownloading, 10, "Initializing S3 client...")
	client, _, bucket, err := s.NewS3AndPresigner()
	if err != nil {
		_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Failed to initialize S3 client: %v", err))
		config.Log.Error("failed init s3: " + err.Error())
		return nil, err
	}

	// Key cho GetObject: thử key gốc trước; nếu NoSuchKey thì thử chuẩn hóa public/public/ -> public/
	s3GetKey := req.Key
	s3GetKeyNormalized := s3GetKey
	if strings.HasPrefix(s3GetKeyNormalized, "public/public/") {
		s3GetKeyNormalized = "public/" + strings.TrimPrefix(s3GetKeyNormalized, "public/public/")
	}
	config.Log.Info("S3 GetObject: bucket=", bucket, " key=", s3GetKey)

	// Tải file ZIP/RAR từ S3 về local
	_ = UpdateProgress(jobID, StatusDownloading, 20, "Downloading file from S3...")
	tmpFile, err := os.CreateTemp("", "uploaded-archive-*")
	if err != nil {
		_ = UpdateProgress(jobID, StatusFailed, 0, "Cannot create temp file")
		config.Log.Error("cannot create temp file: " + err.Error())
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	getOut, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3GetKey),
	})
	if err != nil && s3GetKeyNormalized != s3GetKey {
		// Thử key đã chuẩn hóa (một số bucket chỉ lưu một lớp public/)
		config.Log.Info("S3 GetObject retry with normalized key: ", s3GetKeyNormalized)
		getOut, err = client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3GetKeyNormalized),
		})
		s3GetKey = s3GetKeyNormalized
	}
	if err != nil {
		msg := fmt.Sprintf("S3 download failed: %v", err)
		_ = UpdateProgress(jobID, StatusFailed, 0, msg)
		config.Log.Error("failed to download from s3: bucket=", bucket, " key=", s3GetKey, " error: ", err.Error())
		return nil, err
	}
	defer getOut.Body.Close()

	// Copy có báo tiến trình (file lớn 1.5GB+ có thể tải vài phút)
	var total int64
	if getOut.ContentLength != nil && *getOut.ContentLength > 0 {
		total = *getOut.ContentLength
	}
	written, err := copyWithProgress(ctx, jobID, tmpFile, getOut.Body, total, 20, 30, StatusDownloading, "Downloading file from S3...")
	if err != nil {
		_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Failed to save temp file: %v", err))
		config.Log.Error("failed to save temp file: " + err.Error())
		return nil, err
	}
	_ = written // used for progress

	_ = UpdateProgress(jobID, StatusDownloading, 30, "File downloaded successfully")

	var finalUrl string

	// Giải nén
	_ = UpdateProgress(jobID, StatusExtracting, 40, "Extracting archive...")
	mediaSvc := NewMediaService(mediaRepo)
	if ext == ".zip" {
		if err := mediaSvc.ExtractZip(tmpFile.Name(), localExtract); err != nil {
			_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Unzip failed: %v", err))
			config.Log.Error("unzip failed: " + err.Error())
			return nil, err
		}
	} else {
		if err := mediaSvc.ExtractRar(tmpFile.Name(), localExtract); err != nil {
			_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Unrar failed: %v", err))
			config.Log.Error("unrar failed: " + err.Error())
			return nil, err
		}
	}
	_ = UpdateProgress(jobID, StatusExtracting, 50, "Archive extracted successfully")

	// Sửa quyền file/thư mục sau giải nén để tránh "permission denied" khi đọc/upload (zip có thể chứa mode hạn chế)
	_ = filepath.Walk(localExtract, func(p string, d os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d == nil {
			return nil
		}
		if d.IsDir() {
			_ = os.Chmod(p, 0755)
		} else {
			_ = os.Chmod(p, 0644)
		}
		return nil
	})

	// Upload toàn bộ nội dung localExtract lên S3
	_ = UpdateProgress(jobID, StatusUploading, 60, "Uploading files to S3...")
	var totalFiles, uploadedFiles int
	// Đếm tổng số file trước
	filepath.Walk(localExtract, func(p string, d os.FileInfo, walkErr error) error {
		if d != nil && !d.IsDir() {
			totalFiles++
		}
		return nil
	})

	err = filepath.Walk(localExtract, func(p string, d os.FileInfo, walkErr error) error {
		if walkErr != nil {
			config.Log.Warn("walk error (upload): ", walkErr.Error(), " path: ", p)
			return nil // Bỏ qua lỗi và tiếp tục
		}

		if d == nil {
			return nil
		}

		rel, _ := filepath.Rel(localExtract, p)
		rel = filepath.ToSlash(rel)

		// Nếu là folder
		if d.IsDir() {
			return nil
		}

		// Bỏ qua __MACOSX và file ẩn macOS
		if strings.HasPrefix(rel, "__MACOSX/") || strings.HasPrefix(filepath.Base(rel), "._") {
			return nil
		}

		// KHÔNG bỏ qua thư mục data khi upload lên S3 - upload tất cả file (kể cả trong data)
		// Chỉ bỏ qua nếu có lỗi permission khi mở file (xử lý ở dưới)

		s3Key := filepath.ToSlash(filepath.Join(s3Prefix, rel))

		if strings.HasPrefix(s3Key, "public/public/") {
			s3Key = "public/" + strings.TrimPrefix(s3Key, "public/public/")
		}

		// Nếu key chứa public/power_point thì bỏ prefix "public/" khỏi s3Key khi upload
		if strings.Contains(req.Key, "public/power_point") {
			s3Key = strings.TrimPrefix(s3Key, "public/")
		} else if strings.Contains(s3Key, "public/Root/power_point") {
			s3Key = strings.TrimPrefix(s3Key, "public/Root/")
		} else if strings.Contains(req.Key, "public/Root") {
			s3Key = strings.Replace(s3Key, "public/Root/", "public/", 1)
		}

		f, err := os.Open(p)
		if err != nil {
			config.Log.Warn("failed to open file: ", err.Error(), " path: ", p)
			return nil // Bỏ qua file này và tiếp tục
		}
		defer f.Close()

		fileExt := filepath.Ext(p)
		mimeType := mime.TypeByExtension(fileExt)
		if mimeType == "" {
			switch fileExt {
			case ".jpg", ".jpeg":
				mimeType = "image/jpeg"
			case ".png":
				mimeType = "image/png"
			case ".gif":
				mimeType = "image/gif"
			default:
				mimeType = "application/octet-stream"
			}
		}

		_, err = client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:             aws.String(bucket),
			Key:                aws.String(s3Key),
			Body:               f,
			ContentType:        aws.String(mimeType),
			ContentDisposition: aws.String("inline"),
		})
		if err == nil {
			uploadedFiles++
			if totalFiles > 0 {
				progress := 60 + (uploadedFiles*20)/totalFiles // 60-80%
				_ = UpdateProgress(jobID, StatusUploading, progress, fmt.Sprintf("Uploading files to S3... (%d/%d)", uploadedFiles, totalFiles))
			}
			// Nếu file là .zip và nằm trong thư mục power_point thì tiếp tục giải nén như upload zip
			if fileExt == ".zip" && strings.Contains(s3Key, "power_point") {
				nestedReq := &dto.ExtractRequest{
					Filename:     filepath.Base(rel),
					Key:          s3Key,
					URL:          utils.StaticURL(s3Key, s3DiskName),
					IsPowerPoint: true,
				}
				if _, nestedErr := s.Extract(c, nestedReq); nestedErr != nil {
					config.Log.Warn("nested zip extract failed: ", s3Key, " ", nestedErr.Error())
				}
			}
		}
		return err
	})
	if err != nil {
		_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Upload failed: %v", err))
		config.Log.Error("upload extracted failed: " + err.Error())
		return nil, err
	}
	_ = UpdateProgress(jobID, StatusUploading, 80, "Files uploaded to S3 successfully")

	// Quét folder localExtract và insert DB
	_ = UpdateProgress(jobID, StatusScanning, 85, "Scanning and indexing files...")
	folderIDMap := map[string]int64{
		"": current.ID,
	}

	fullPath := ""
	parts := strings.Split(s3Dir, "/")
	parentID := current.ID
	publicDisk := config.Public

	for _, part := range parts {
		if part == "" || part == "public" || part == "Root" {
			continue
		}
		if fullPath == "" {
			fullPath = part
		}

		oldFolder, _ := mediaRepo.FindFolderByFilePathAndParentMaster(fullPath, parentID, publicDisk)

		if oldFolder != nil && oldFolder.ID > 0 {
			parentID = oldFolder.ID
			folderIDMap[fullPath] = oldFolder.ID
			current = oldFolder
		} else {
			folder := &models.Media{
				ParentID: &parentID,
				FilePath: part,
				Type:     "folder",
				FullPath: fullPath,
				DiskName: &publicDisk,
			}
			_ = mediaRepo.UpdateOrCreate(folder)
			parentID = folder.ID
			folderIDMap[fullPath] = folder.ID
			current = folder
		}
	}

	// Reset totalFiles để đếm lại số file được insert vào DB
	totalFiles = 0
	err = filepath.Walk(localExtract, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			config.Log.Warn("walk error (scan): ", walkErr.Error(), " path: ", p)
			return nil // Bỏ qua lỗi và tiếp tục
		}

		if info == nil {
			return nil
		}

		rel, _ := filepath.Rel(localExtract, p)
		rel = filepath.ToSlash(rel)

		// Bỏ qua gốc, __MACOSX, file ẩn macOS
		if rel == "" || strings.Contains(info.Name(), "__MACOSX") || strings.HasPrefix(filepath.Base(rel), "._") || strings.Contains(rel, "__MACOSX") {
			return nil
		}

		// KHÔNG bỏ qua file trong thư mục data khi scan để insert vào DB - insert tất cả file

		// Nếu là folder
		if info.IsDir() {
			// Bỏ qua thư mục data - không insert vào DB
			if req.IsPowerPoint {
				parts := strings.Split(rel, "/")
				for _, part := range parts {
					if strings.ToLower(strings.TrimSpace(part)) == "data" || strings.ToLower(strings.TrimSpace(part)) == "data.local-only" {
						return nil // Bỏ qua folder data khi insert vào DB
					}
				}
			}

			// Tạo fullPath từ s3Prefix + rel (bỏ prefix "public/")
			folderFullPath := filepath.ToSlash(filepath.Join(s3Prefix, rel))
			if (strings.Contains(req.Key, "public/") && !req.IsPowerPoint) || strings.Contains(req.Key, "public/power_point") {
				folderFullPath = strings.TrimPrefix(folderFullPath, "public/")
			}

			// Check và tạo từng cấp folder
			parts := strings.Split(folderFullPath, "/")
			parentID := int64(0) // Bắt đầu từ root
			builtFullPath := ""

			for i, part := range parts {
				if part == "" {
					continue
				}
				if builtFullPath == "" {
					builtFullPath = part
				} else {
					builtFullPath = builtFullPath + "/" + part
				}

				if existingID, ok := folderIDMap[builtFullPath]; ok {
					parentID = existingID
					continue
				}

				if strings.Contains(builtFullPath, "public/") && !req.IsPowerPoint {
					builtFullPath = strings.TrimPrefix(builtFullPath, "public/")
				}

				if strings.Contains(builtFullPath, "Root/") && !req.IsPowerPoint {
					builtFullPath = strings.TrimPrefix(builtFullPath, "Root/")
				}

				var folderParentID int64
				if i == 0 {
					folderParentID = 0
					if part == "power_point" {
						powerPointFolder, _ := mediaRepo.FindByPathMaster("power_point", publicDisk)
						if powerPointFolder != nil && powerPointFolder.ID > 0 {
							if powerPointFolder.ParentID == nil || *powerPointFolder.ParentID != 2 {
								correctParentID := int64(2)
								powerPointFolder.ParentID = &correctParentID
								_ = mediaRepo.UpdateOrCreate(powerPointFolder)
							}
							folderIDMap[builtFullPath] = powerPointFolder.ID
							parentID = powerPointFolder.ID
							continue
						}
						folderParentID = 2
					} else if part == "Root" {
						rootFolder, _ := mediaRepo.FindByPathMaster("", publicDisk)
						if rootFolder != nil && rootFolder.ID > 0 {
							if rootFolder.ParentID == nil || *rootFolder.ParentID != 2 {
								correctParentID := int64(2)
								rootFolder.ParentID = &correctParentID
								_ = mediaRepo.UpdateOrCreate(rootFolder)
							}
							folderIDMap[builtFullPath] = rootFolder.ID
							parentID = rootFolder.ID
							continue
						}
						folderParentID = 0
					}
				} else {
					folderParentID = parentID
				}

				// Tìm folder theo full_path trước
				oldFolder, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)

				// Nếu không tìm thấy theo full_path, tìm tất cả folder có cùng file_path (bất kể parent_id)
				// để tìm folder cũ có full_path sai hoặc parent_id sai
				if oldFolder == nil || oldFolder.ID == 0 {
					// Tìm theo file_path và parent_id trước
					oldFolder, _ = mediaRepo.FindFolderByFilePathAndParentMaster(part, folderParentID, publicDisk)

					// Nếu tìm thấy folder nhưng full_path hoặc parent_id sai, cập nhật
					if oldFolder != nil && oldFolder.ID > 0 {
						oldFolder, _ = mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
					}

					// Nếu vẫn không tìm thấy, tìm tất cả folder có cùng file_path (bất kể parent_id)
					if oldFolder == nil || oldFolder.ID == 0 {
						existingFolders := []models.Media{}
						db.MasterDB.Unscoped().Where("file_path = ? AND type = ? AND disk_name = ?",
							part, "folder", publicDisk).Find(&existingFolders)

						// Tìm folder có full_path khác builtFullPath hoặc parent_id sai (folder cũ cần cập nhật)
						for _, existing := range existingFolders {
							if existing.FullPath != builtFullPath || (existing.ParentID == nil || *existing.ParentID != folderParentID) {
								updated, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
								if updated != nil && updated.ID > 0 {
									oldFolder = updated
									break
								}
							}
						}
					}
				}

				if oldFolder != nil && oldFolder.ID > 0 {
					// Đảm bảo parent_id và full_path đúng
					needUpdate := false
					if oldFolder.ParentID == nil || *oldFolder.ParentID != folderParentID {
						needUpdate = true
					}
					if oldFolder.FullPath != builtFullPath {
						needUpdate = true
					}

					if needUpdate {
						if folderParentID == 0 {
							folderParentID = 3
						}

						oldFolder, _ = mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
					}

					if oldFolder != nil && oldFolder.ID > 0 {
						parentID = oldFolder.ID
						folderIDMap[builtFullPath] = oldFolder.ID
						current = oldFolder
					}
				} else {
					// Không tìm thấy folder nào, tạo mới
					folderName := part

					if (folderName == "data" || folderName == "data.local-only") &&
						strings.Contains(builtFullPath, "power_point/") {
						continue
					}

					newFolder := &models.Media{
						ParentID: &folderParentID,
						FilePath: folderName,
						Type:     "folder",
						FullPath: builtFullPath,
						DiskName: &publicDisk,
					}
					if err := db.MasterDB.Create(newFolder).Error; err == nil && newFolder.ID > 0 {
						parentID = newFolder.ID
						folderIDMap[builtFullPath] = newFolder.ID
						current = newFolder
					} else {
						// Nếu Create thất bại, tìm lại
						createdFolder, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
						if createdFolder != nil && createdFolder.ID > 0 {
							parentID = createdFolder.ID
							folderIDMap[builtFullPath] = createdFolder.ID
							current = createdFolder
						}
					}
				}
			}
			return nil
		}

		// Nếu là file
		// Tìm index.html để set finalUrl (không cần kiểm tra thư mục data vì có thể không có hoặc bị bỏ qua)
		if req.IsPowerPoint {
			base := strings.ToLower(filepath.Base(rel))
			if base == "index.html" {
				if finalUrl == "" {
					finalUrl = filepath.ToSlash(filepath.Join(s3Prefix, rel))
					finalUrl = strings.TrimPrefix(finalUrl, "public/")
					config.Log.Info("Found index.html, setting finalUrl: ", finalUrl, " rel: ", rel)
				} else {
					config.Log.Warn("Multiple index.html found, clearing finalUrl. Previous: ", finalUrl, " current: ", rel)
					finalUrl = ""
				}
			}
		}

		// Bỏ qua file trong thư mục data khi insert vào DB (nhưng vẫn upload lên S3)
		// TRỪ index.html - index.html phải luôn được insert vào DB
		if req.IsPowerPoint && !info.IsDir() {
			base := strings.ToLower(filepath.Base(rel))
			// Nếu là index.html, không bỏ qua
			if base != "index.html" {
				parts := strings.Split(rel, "/")
				for _, part := range parts {
					if strings.ToLower(strings.TrimSpace(part)) == "data" || strings.ToLower(strings.TrimSpace(part)) == "data.local-only" {
						return nil // Bỏ qua file trong thư mục data khi insert vào DB (trừ index.html)
					}
				}
			}
		}

		// Insert file vào DB (tất cả file, không chỉ index.html, nhưng bỏ qua file trong data)
		// Tạo fullPath từ s3Prefix + rel (bỏ prefix "public/")
		fileFullPath := filepath.ToSlash(filepath.Join(s3Prefix, rel))
		if strings.Contains(req.Key, "public/power_point") {
			fileFullPath = strings.TrimPrefix(fileFullPath, "public/")
		}

		if strings.Contains(fileFullPath, "Root/") {
			fileFullPath = strings.Replace(fileFullPath, "Root/", "", 1)
		}

		// Tìm folder chứa file này
		fileDir := filepath.Dir(fileFullPath)
		var folderID int64

		// Check folderIDMap trước
		if existingFolderID, ok := folderIDMap[fileDir]; ok {
			folderID = existingFolderID
		} else {
			// Tìm folder trong DB
			parts := strings.Split(fileDir, "/")
			parentID := int64(0)
			builtFullPath := ""

			for i, part := range parts {
				if part == "" {
					continue
				}

				if part == "public" && !req.IsPowerPoint {
					continue
				}

				if builtFullPath == "" {
					builtFullPath = part
				} else {
					builtFullPath = builtFullPath + "/" + part
				}

				if strings.Contains(builtFullPath, "public/") && !req.IsPowerPoint {
					builtFullPath = strings.TrimPrefix(builtFullPath, "public/")
				}

				var folderParentID int64
				if i == 0 {
					folderParentID = 0
					if part == "power_point" {
						// Check power_point có id = 3, parent_id = 2
						powerPointFolder, _ := mediaRepo.FindByPathMaster("power_point", publicDisk)
						if powerPointFolder != nil && powerPointFolder.ID > 0 {
							// Đảm bảo parent_id đúng
							if powerPointFolder.ParentID == nil || *powerPointFolder.ParentID != 2 {
								correctParentID := int64(2)
								powerPointFolder.ParentID = &correctParentID
								_ = mediaRepo.UpdateOrCreate(powerPointFolder)
							}
							folderIDMap[builtFullPath] = powerPointFolder.ID
							parentID = powerPointFolder.ID
							continue
						}
						folderParentID = 2
					}
				} else {
					folderParentID = parentID
				}

				// Tìm folder theo full_path trước
				oldFolder, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)

				// Nếu không tìm thấy theo full_path, tìm tất cả folder có cùng file_path (bất kể parent_id)
				// để tìm folder cũ có full_path sai hoặc parent_id sai
				if oldFolder == nil || oldFolder.ID == 0 {
					// Tìm theo file_path và parent_id trước
					oldFolder, _ = mediaRepo.FindFolderByFilePathAndParentMaster(part, folderParentID, publicDisk)

					// Nếu tìm thấy folder nhưng full_path hoặc parent_id sai, cập nhật
					if oldFolder != nil && oldFolder.ID > 0 {
						if oldFolder.FullPath != builtFullPath || (oldFolder.ParentID == nil || *oldFolder.ParentID != folderParentID) {
							err := db.MasterDB.Unscoped().Where("id = ?", oldFolder.ID).Updates(map[string]interface{}{
								"full_path": builtFullPath,
								"parent_id": folderParentID,
							}).Error
							if err == nil {
								oldFolder, _ = mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
							}
						}
					}

					// Nếu vẫn không tìm thấy, tìm tất cả folder có cùng file_path (bất kể parent_id)
					if oldFolder == nil || oldFolder.ID == 0 {
						existingFolders := []models.Media{}
						db.MasterDB.Unscoped().Where("file_path = ? AND type = ? AND disk_name = ?",
							part, "folder", publicDisk).Find(&existingFolders)

						// Tìm folder có full_path khác builtFullPath hoặc parent_id sai (folder cũ cần cập nhật)
						for _, existing := range existingFolders {
							if existing.FullPath != builtFullPath || (existing.ParentID == nil || *existing.ParentID != folderParentID) {
								// Cập nhật folder cũ với full_path và parent_id đúng
								err := db.MasterDB.Unscoped().Where("id = ?", existing.ID).Updates(map[string]interface{}{
									"full_path": builtFullPath,
									"parent_id": folderParentID,
								}).Error
								if err == nil {
									// Tìm lại sau khi update
									updated, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
									if updated != nil && updated.ID > 0 {
										oldFolder = updated
										break
									}
								}
							}
						}
					}
				} else {
					oldFolder, _ = mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
				}

				if oldFolder != nil && oldFolder.ID > 0 {
					// Đảm bảo parent_id và full_path đúng
					needUpdate := false
					if oldFolder.ParentID == nil || *oldFolder.ParentID != folderParentID {
						needUpdate = true
					}
					if oldFolder.FullPath != builtFullPath {
						needUpdate = true
					}

					if needUpdate {
						if folderParentID == 0 {
							folderParentID = 3
						}

						oldFolder, _ = mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
					}

					if oldFolder != nil && oldFolder.ID > 0 {
						parentID = oldFolder.ID
						folderIDMap[builtFullPath] = oldFolder.ID
					}
				} else {
					folderName := part

					if (folderName == "data" || folderName == "data.local-only") &&
						strings.Contains(builtFullPath, "power_point/") {
						continue
					}

					// Không tìm thấy folder nào, tạo mới
					newFolder := &models.Media{
						ParentID: &folderParentID,
						FilePath: folderName,
						Type:     "folder",
						FullPath: builtFullPath,
						DiskName: &publicDisk,
					}
					if err := db.MasterDB.Create(newFolder).Error; err == nil && newFolder.ID > 0 {
						parentID = newFolder.ID
						folderIDMap[builtFullPath] = newFolder.ID
					} else {
						// Nếu Create thất bại, tìm lại
						createdFolder, _ := mediaRepo.FindByPathMaster(builtFullPath, publicDisk)
						if createdFolder != nil && createdFolder.ID > 0 {
							parentID = createdFolder.ID
							folderIDMap[builtFullPath] = createdFolder.ID
						}
					}
				}
			}
			// Sau khi tạo tất cả folder, parentID là id của folder cuối cùng trong chuỗi fileDir
			// builtFullPath cuối cùng trong vòng lặp phải là fileDir
			// Nếu không tìm thấy fileDir trong folderIDMap, dùng parentID (folder cuối cùng)
			if existingFolderID, ok := folderIDMap[fileDir]; ok {
				folderID = existingFolderID
			} else {
				// builtFullPath cuối cùng trong vòng lặp phải là fileDir
				// Nếu không, tìm lại với fileDir
				if builtFullPath == fileDir {
					folderIDMap[fileDir] = parentID
					folderID = parentID
				} else {
					// Tìm lại folder với fileDir (full_path)
					finalFolder, _ := mediaRepo.FindByPathMaster(fileDir, publicDisk)
					if finalFolder != nil && finalFolder.ID > 0 {
						// Đảm bảo parent_id đúng
						if finalFolder.ParentID == nil || *finalFolder.ParentID != parentID {
							finalFolder.ParentID = &parentID
							_ = mediaRepo.UpdateOrCreate(finalFolder)
						}
						folderIDMap[fileDir] = finalFolder.ID
						folderID = finalFolder.ID
					} else {
						// Fallback: dùng parentID
						folderIDMap[fileDir] = parentID
						folderID = parentID
					}
				}
			}
		}

		fileExt := filepath.Ext(info.Name())
		mimeType := mime.TypeByExtension(fileExt)
		static := filepath.ToSlash(filepath.Join(s3Prefix, rel))
		if strings.HasPrefix(static, "public/public/") {
			static = "public/" + strings.TrimPrefix(static, "public/public/")
		}
		size := info.Size()

		// Nếu key chứa public/power_point thì bỏ prefix "public/" khỏi static URL
		if strings.Contains(req.Key, "public/power_point") {
			static = strings.TrimPrefix(static, "public/")
		}

		if strings.Contains(static, "Root/") {
			static = strings.Replace(static, "Root/", "", 1)
		}

		if strings.Contains(static, "public/power_point") {
			static = strings.Replace(static, "public/", "", 1)
		}

		if strings.Contains(fileDir, "public/") {
			fileDir = strings.Replace(fileDir, "public/", "", 1)
		}

		fileName := s.GetFolderName(static)
		size = s.GetFolderSize(ctx, client, bucket, static)

		if filepath.Ext(fileName) == "" {
			fileName = fileName + fileExt
		}

		if fileName == "data.local-only" ||
			fileExt == ".DS_Store" ||
			(strings.Contains(fileDir, "power_point/") && (strings.Contains(fileDir, "/data") || strings.Contains(fileDir, "/data.local-only"))) {
			return nil
		}

		// Không tạo Media cho file .zip trong power_point vì đã được giải nén đệ quy (nested Extract)
		if fileExt == ".zip" && strings.Contains(fileDir, "power_point/") {
			return nil
		}

		media := &models.Media{
			FolderID:      folderID,
			FileName:      fileName,
			FilePath:      fileDir,
			FileType:      &mimeType,
			FileSize:      &size,
			FileExtension: &fileExt,
			DiskName:      &s3DiskName,
			StaticURL:     &static,
			Type:          "file",
		}
		_ = mediaRepo.UpdateOrCreateV2(media)
		totalFiles++
		return nil
	})
	if err != nil {
		_ = UpdateProgress(jobID, StatusFailed, 0, fmt.Sprintf("Scan failed: %v", err))
		config.Log.Error("scan extracted failed: " + err.Error())
		return nil, err
	}

	// Hoàn thành
	if finalUrl == "" {
		finalUrl = utils.StaticURL(req.URL, s3DiskName)
	}
	_ = UpdateProgress(jobID, StatusCompleted, 100, "Extraction completed successfully")

	// Update final URL trong progress
	prog, _ := GetProgress(jobID)
	if prog != nil {
		prog.FinalURL = finalUrl
		_ = SaveProgress(jobID, prog)
	}

	return &dto.ExtractResponse{
		Message:  "Extract successful",
		Folder:   strings.TrimPrefix(s3Prefix, "public/"),
		Files:    int32(totalFiles),
		BaseKey:  req.Key,
		FinalUrl: finalUrl,
	}, nil
}

func (s *uploadService) NewS3AndPresigner() (*s3.Client, *s3.PresignClient, string, error) {
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
		return nil, nil, "", err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Disk.PathStyle
	})

	presigner := s3.NewPresignClient(client)

	return client, presigner, s3Disk.Bucket, nil
}

func (s *uploadService) GetFolderName(path string) string {
	return filepath.Base(path)
}

func (s *uploadService) GetFolderPrefix(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[:idx+1]
}

func (s *uploadService) GetFolderSize(ctx context.Context, s3Client *s3.Client, bucket, prefix string) int64 {
	folderPrefix := s.GetFolderPrefix(prefix)
	result, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(folderPrefix),
	})
	if err != nil {
		return 0
	}

	var total int64
	for _, obj := range result.Contents {
		total += *obj.Size
	}

	return total
}

// copyWithProgress copy từ src vào dst, cập nhật progress job từ progressStart% đến progressEnd% (cho file lớn 1.5GB+).
func copyWithProgress(ctx context.Context, jobID string, dst io.Writer, src io.Reader, totalBytes int64, progressStart, progressEnd int, status, message string) (int64, error) {
	const chunkReport = 50 << 20 // báo mỗi 50MB
	var written int64
	buf := make([]byte, 32*1024)
	lastReport := int64(0)
	lastPct := -1

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
		}
		nr, errRead := src.Read(buf)
		if nr > 0 {
			nw, errWrite := dst.Write(buf[:nr])
			written += int64(nw)
			if errWrite != nil {
				return written, errWrite
			}
			if nw != nr {
				return written, io.ErrShortWrite
			}
		}
		if errRead == io.EOF {
			break
		}
		if errRead != nil {
			return written, errRead
		}
		// Báo tiến trình: mỗi 50MB hoặc mỗi khi % thay đổi (khi biết total)
		currentPct := int64(0)
		if totalBytes > 0 {
			currentPct = written * 100 / totalBytes
		}
		shouldReport := (totalBytes <= 0 && written-lastReport >= chunkReport) ||
			(totalBytes > 0 && (written-lastReport >= chunkReport || int(currentPct) != lastPct))
		if !shouldReport {
			continue
		}
		lastReport = written
		pct := progressStart
		if totalBytes > 0 {
			pct = progressStart + int((written*int64(progressEnd-progressStart))/totalBytes)
			if pct > progressEnd {
				pct = progressEnd
			}
			lastPct = int(currentPct)
		}
		msg := message
		if totalBytes > 0 {
			msg = fmt.Sprintf("Downloading... %d%% (%s / %s)", pct, formatBytes(written), formatBytes(totalBytes))
		}
		_ = UpdateProgress(jobID, status, pct, msg)
	}
	return written, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

