package services

import (
	"be-lms/config"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
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

type uploadService struct {}

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
			Filename:    req.Filename,
			Key:         req.Key,
			URL:         req.URL,
			Size:        req.Size,
			ContentType: req.ContentType,
			UserID:      req.UserID,
			IsPowerPoint: true,
		}
		res, err := s.Extract(c, extractReq)

		if err != nil {
			config.Log.Error("Extract failed: ", req.URL)
		}
		return &dto.CompleteResponse{
			Message:    res.Message,
			FinalUrl: utils.StaticURL(res.FinalUrl, s3DiskName),
		}, err
	}

	diskName := config.Public // disk name, ví dụ "public"
	parentZero := int64(0)

	mediaRepo := repositories.NewMediaRepository()

	// Lấy folder root
	parentRootFolder, err := mediaRepo.FindFolderByFilePathAndParent("", parentZero, diskName)
	if err == nil && parentRootFolder != nil && parentRootFolder.ID != 0 {
		parentZero = parentRootFolder.ID
	}

	current := parentRootFolder
	currentFullPath := parentRootFolder.FullPath

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
		Message: "success",
		FinalUrl: utils.StaticURL(req.URL, s3DiskName),
	},
	nil
}

func (s *uploadService) Extract(c *gin.Context, req *dto.ExtractRequest) (*dto.ExtractResponse, error) {
	if req == nil || req.Key == "" {
		var body dto.ExtractRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			config.Log.Error("Bind JSON error:", err)
			return nil, err
		}
		req = &body
	}

	// Validate required fields
	if req.Key == "" || req.URL == "" || req.Filename == "" {
		config.Log.Error("missing key/url/filename")
		return nil, fmt.Errorf("missing key/url/filename")
	}

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
	current := parentRootFolder

	// Lấy base name và S3 prefix từ req.Key
	baseName := strings.TrimSuffix(filepath.Base(req.Key), ext)
	s3Dir := filepath.Dir(req.Key)
	s3Prefix := filepath.ToSlash(filepath.Join(s3Dir, baseName))

	// Local extract temp dir (unique, tránh đụng file cũ)
	localExtract := filepath.Join(os.TempDir(), baseName)
	_ = os.RemoveAll(localExtract) // xoá nếu tồn tại
	if err := os.MkdirAll(localExtract, os.ModePerm); err != nil {
		return nil, err
	}
	// localExtract, err := os.MkdirTemp("", baseName+"-*")
	// if err != nil {
	// 	config.Log.Error("cannot create local extract dir: " + err.Error())
	// 	return nil, err
	// }
	defer os.RemoveAll(localExtract) // cleanup khi xong

	// Khởi tạo client S3
	client, _, bucket, err := s.NewS3AndPresigner()
	if err != nil {
		config.Log.Error("failed init s3: " + err.Error())
		return nil, err
	}

	// Tải file ZIP/RAR từ S3 về local
	tmpFile, err := os.CreateTemp("", "uploaded-archive-*")
	if err != nil {
		config.Log.Error("cannot create temp file: " + err.Error())
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	getOut, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(req.Key),
	})
	if err != nil {
		config.Log.Error("failed to download from s3: " + err.Error())
		return nil, err
	}
	defer getOut.Body.Close()

	if _, err := io.Copy(tmpFile, getOut.Body); err != nil {
		config.Log.Error("failed to save temp file: " + err.Error())
		return nil, err
	}

	var finalUrl string

	// Giải nén
	mediaSvc := NewMediaService(mediaRepo)
	if ext == ".zip" {
		if err := mediaSvc.ExtractZip(tmpFile.Name(), localExtract); err != nil {
			config.Log.Error("unzip failed: " + err.Error())
			return nil, err
		}
	} else {
		if err := mediaSvc.ExtractRar(tmpFile.Name(), localExtract); err != nil {
			config.Log.Error("unrar failed: " + err.Error())
			return nil, err
		}
	}

	// Upload toàn bộ nội dung localExtract lên S3
	err = filepath.Walk(localExtract, func(p string, d os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(localExtract, p)

		// Bỏ qua __MACOSX và file ẩn macOS
		if strings.HasPrefix(rel, "__MACOSX/") || strings.HasPrefix(filepath.Base(rel), "._") {
			return nil
		}

		s3Key := filepath.ToSlash(filepath.Join(s3Prefix, rel))

		f, err := os.Open(p)
		if err != nil {
			return err
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

		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:             aws.String(bucket),
			Key:                aws.String(s3Key),
			Body:               f,
			ContentType:        aws.String(mimeType),
			ContentDisposition: aws.String("inline"),
		})
		return err
	})
	if err != nil {
		config.Log.Error("upload extracted failed: " + err.Error())
		return nil, err
	}

	// Quét folder localExtract và insert DB
	folderIDMap := map[string]int64{
		"": current.ID,
	}

	fullPath := ""
	parts := strings.Split(s3Dir, "/")
	parentID := current.ID
	publicDisk := config.Public

	config.Log.Info("parts: ", parts)

	for _, part := range parts {
		config.Log.Info("part: ", part)
		if part == "" || part == "public" {
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

	var totalFiles int
	err = filepath.Walk(localExtract, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, _ := filepath.Rel(localExtract, p)
		rel = filepath.ToSlash(rel)

		// Bỏ qua gốc, __MACOSX, file ẩn macOS
		if rel == "" || strings.Contains(info.Name(), "__MACOSX") || strings.HasPrefix(filepath.Base(rel), "._") || strings.Contains(rel, "__MACOSX") {
			return nil
		}

		// Nếu là folder
		if info.IsDir() {
			// Nếu folder là "data" và là PowerPoint → skip
			if strings.ToLower(info.Name()) == "data" && req.IsPowerPoint {
				return nil
				// return filepath.SkipDir
			}

			// 🔥 Tạo fullPath dựa trên rel thay vì current.FullPath

			fullPath := current.FullPath + "/" + filepath.Base(p)
			newFolder, _ := mediaRepo.FindFolderByFilePathAndParentMaster(fullPath, current.ID, publicDisk)
			if newFolder != nil && newFolder.ID > 0 {
				current = newFolder
				folderIDMap[newFolder.FullPath] = newFolder.ID
			} else {
				media := &models.Media{
					ParentID: &current.ID,
					FilePath: filepath.Base(p),
					Type:     "folder",
					FullPath: fullPath,
					DiskName: &publicDisk,
				}
				_ = mediaRepo.UpdateOrCreate(media)
				current = media
				folderIDMap[fullPath] = media.ID
			}
			return nil
		}

		// Nếu là file
		if req.IsPowerPoint {
			base := strings.ToLower(filepath.Base(rel))
			if base != "index.html" {
				return nil
			}
			// Check cùng cấp có "data"
			parentDir := filepath.Join(localExtract, filepath.Dir(rel))
			dataDir := filepath.Join(parentDir, "data")
			if stat, err := os.Stat(dataDir); err != nil || !stat.IsDir() {
				return nil
			} else {
				if finalUrl == "" {
					finalUrl = filepath.ToSlash(filepath.Join(s3Prefix, rel))
				} else {
					finalUrl = ""
				}
			}
		}

		// Insert file vào DB
		fileExt := filepath.Ext(info.Name())
		mimeType := mime.TypeByExtension(fileExt)
		size := info.Size()
		static := filepath.ToSlash(filepath.Join(s3Prefix, rel))

		media := &models.Media{
			FolderID:      current.ID,
			FileName:      info.Name(),
			FilePath:      current.FullPath,
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
		config.Log.Error("scan extracted failed: " + err.Error())
		return nil, err
	}

	return &dto.ExtractResponse{
		Message: "Extract successful",
		Folder:  strings.TrimPrefix(s3Prefix, "public/"),
		Files:   int32(totalFiles),
		BaseKey: req.Key,
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
