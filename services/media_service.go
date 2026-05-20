package services

import (
	"archive/zip"
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/redis"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"log"
	"syscall"

	"github.com/gin-gonic/gin"
)

const FileDiskStorage = "public"

type MediaService interface {
	UploadFile(c *gin.Context) (*prot.File, error)
	UploadFolder(c *gin.Context, req prot.Media) error
	DeleteFolder(c *gin.Context, folderId int64) error
	DeleteFolderAndFiles(c *gin.Context, folderId int64) error
	DeleteFile(c *gin.Context, fileId int64) error
	Folders(c *gin.Context, folderId int64) (*prot.MediaTree, error)
	Files(c *gin.Context, folderId int64) (*prot.FileResponse, error)
	UploadFileZip(info *utils.FileInfo, path string) (*prot.File, error)
	ExtractZip(zipPath, extractPath string) error
	ExtractRar(rarPath, extractPath string) error
}

type mediaService struct {
	repo repositories.MediaRepository
}

func NewMediaService(repo repositories.MediaRepository) MediaService {
	return &mediaService{repo: repo}
}

func (s *mediaService) UploadFile(c *gin.Context) (*prot.File, error) {
	config.Log.Info("UploadFile")
	path := c.PostForm("path")
	shipExtract := c.PostForm("skip_extract")

	if path == "" {
		path = "files"
	}

	// B1: Upload file to storage local
	uploader := utils.NewUploaderWithStorage(path, "file", models.Storage)
	info, err := uploader.Store(c)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	// B2: Check .zip extend, ExtractZip
	if strings.EqualFold(filepath.Ext(info.FileName), ".zip") && shipExtract != "true" {
		return s.UploadFileZip(info, path)
	}

	// B3: Parse file size
	var fileSizeInt64 int64
	if info.FileSize != "" {
		fileSizeInt64, err = strconv.ParseInt(info.FileSize, 10, 64) // ignore parse error, default to 0

		if err != nil {
			fileSizeInt64, _ = parseHumanReadableSize(info.FileSize)
		}
	}

	// B4: Find or create folder by path
	folder, err := s.repo.FindByPath(path, FileDiskStorage)
	if err != nil || folder == nil {
		folder, err = CreateFolderByPath(path, s.repo)
		if err != nil {
			return nil, fmt.Errorf("failed to find or create folder: %w", err)
		}
	}

	stripped := utils.StripDomain(info.Url, models.Storage)
	fileUrl := &stripped

	// B5: Save media
	media := &models.Media{
		FolderID:      folder.ID,
		FileName:      info.FileName,
		FilePath:      info.FilePath,
		FileType:      utils.StringPtr(info.FileMimeType),
		FileSize:      utils.Int64Ptr(fileSizeInt64),
		FileExtension: utils.StringPtr(info.FileExt),
		DiskName:      utils.StringPtr(models.Storage),
		StaticURL:     fileUrl,
		Type:          "file",
	}

	if err := s.repo.Save(media); err != nil {
		return nil, fmt.Errorf("failed to save media: %w", err)
	}

	// B6: Delete cache
	redis.DeleteCache(fmt.Sprintf("medias:all-files"))

	// B7: Get path và URL
	filePath := filepath.Join(media.FilePath, media.FileName)
	//if utils.DerefStr(media.DiskName) == models.Storage {
	//	filePath = filepath.Join("public", filePath)
	//}

	url := utils.DerefStr(media.StaticURL)
	if url != "" {
		url = utils.StaticURL(url, models.Storage)
	}

	// B8: return File
	return &prot.File{
		Id:            int32(media.ID),
		FileName:      media.FileName,
		Url:           url,
		FileSize:      strconv.FormatInt(utils.DerefInt64(media.FileSize), 10),
		FileMimeType:  utils.DerefStr(media.FileType),
		FileExtension: utils.DerefStr(media.FileExtension),
		DiskName:      utils.DerefStr(media.DiskName),
		Path:          filePath,
		Type:          media.Type,
	}, nil
}

func (s *mediaService) UploadFolder(c *gin.Context, req prot.Media) error {
	fullPath := req.FilePath
	if req.ParentId != 0 {
		parentMedia, _ := s.repo.FindById(req.ParentId)

		if parentMedia.FullPath != "" {
			fullPath = fmt.Sprintf("%s/%s", parentMedia.FullPath, req.FilePath)
		}
	}
	media := &models.Media{
		ParentID: utils.Int64Ptr(req.ParentId),
		FileName: "",
		FilePath: req.FilePath,
		FullPath: fullPath,
		Type:     "folder",
		FileSize: utils.Int64Ptr(0),
		DiskName: utils.StringPtr(FileDiskStorage),
	}

	err := s.repo.UpdateOrCreate(media)

	return err
}

func (s *mediaService) DeleteFolder(c *gin.Context, folderId int64) error {
	folder, _ := s.repo.FindById(folderId)

	if folder == nil {
		return fmt.Errorf("folder not found")
	}

	allMedias, _ := s.repo.GetFileByFolder(folder.ID)

	if len(allMedias) > 0 {
		return fmt.Errorf("folder not delete")
	}

	uploader := utils.NewUploaderWithStorage(folder.FullPath, "folder", FileDiskStorage)

	err := uploader.DeleteFolder()

	if err != nil {
		return err
	}

	return s.repo.Delete(int64(folder.ID))
}

func (s *mediaService) DeleteFolderAndFiles(c *gin.Context, folderId int64) error {
	folder, _ := s.repo.FindById(folderId)

	if folder == nil {
		return fmt.Errorf("folder not found")
	}

	files, _ := s.repo.GetFileByFolder(folder.ID)

	for _, file := range files {
		s.repo.DeleteUnscoped(file.ID)
	}

	return s.repo.DeleteUnscoped(int64(folder.ID))
}

func (s *mediaService) DeleteFile(c *gin.Context, fileId int64) error {
	file, _ := s.repo.FindById(fileId)
	if file == nil {
		return fmt.Errorf("file not found")
	}

	diskName := utils.DerefStr(file.DiskName)

	if diskName == config.PowerPoint {
		powerPointService := NewPowerPointUploadService()
		err := powerPointService.DeletePresentation(file.FilePath)

		if err != nil {
			return err
		}
	} else {
		uploader := utils.NewUploaderWithStorage(file.FilePath, "file", FileDiskStorage)
		err := uploader.DeleteFile(file.FileName)

		if err != nil {
			return err
		}
	}

	return s.repo.Delete(int64(file.ID))
}

func (s *mediaService) Folders(c *gin.Context, folderId int64) (*prot.MediaTree, error) {
	var allMedias []models.Media
	if err := s.repo.GetFolders(&allMedias); err != nil {
		return nil, err
	}

	tree := BuildMediaTree(allMedias, folderId)

	return tree[0], nil
}

func (s *mediaService) Files(c *gin.Context, folderId int64) (*prot.FileResponse, error) {
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "20")
	keyword := c.DefaultQuery("keyword", "")
	filterTime := c.DefaultQuery("time", "")
	filterType := c.DefaultQuery("type", "")
	filterExtension := c.DefaultQuery("extension", "")

	if keyword == "" {
		keyword = c.DefaultQuery("search", "")
	}

	page, errPage := strconv.Atoi(pageStr)
	if errPage != nil || page <= 0 {
		page = 1
	}

	perPage, errPerPage := strconv.Atoi(perPageStr)
	if errPerPage != nil || perPage <= 0 {
		perPage = 20
	}

	data, count, err := s.repo.PaginateGetFileByFolder(folderId, keyword, page, perPage, filterTime, filterType, filterExtension)

	if err != nil {
		return nil, err
	}

	files := FormatFiles(data)

	reponse := prot.FileResponse{
		Files:      files,
		TotalCount: uint64(count),
	}

	return &reponse, nil
}

func (s *mediaService) UploadFileZip(info *utils.FileInfo, path string) (*prot.File, error) {
	zipPath := filepath.Join(info.UploadDir, info.FileName)
	baseExtractPath := strings.TrimSuffix(info.FileName, ".zip")
	extractPath := filepath.Join(info.UploadDir, baseExtractPath)
	relPath := filepath.Join(path, baseExtractPath)

	// Tránh trùng folder
	suffix := 1
	for {
		checkPath := strings.TrimPrefix(filepath.Join(path, baseExtractPath), "public/")
		existing, _ := s.repo.FindByPath(checkPath, FileDiskStorage)
		if existing != nil && existing.ID != 0 {
			baseExtractPath = fmt.Sprintf("%s_%d", strings.TrimSuffix(info.FileName, ".zip"), suffix)
			extractPath = filepath.Join(info.UploadDir, baseExtractPath)
			relPath = filepath.Join(path, baseExtractPath)
			suffix++
		} else {
			break
		}
	}

	// Giải nén
	if err := s.ExtractZip(zipPath, extractPath); err != nil {
		return nil, fmt.Errorf("unzip failed: %w", err)
	}
	_ = os.Remove(zipPath) // Optional: log lỗi nếu cần

	// Tạo folder DB chính (folder cha)
	parentFolder, err := s.repo.FindByPath(path, FileDiskStorage)
	if err != nil || parentFolder == nil {
		parentFolder, err = CreateFolderByPath(path, s.repo)
		if err != nil {
			return nil, fmt.Errorf("failed to find or create folder: %w", err)
		}
	}

	// Khởi tạo media cho folder chính
	var folderSize int64
	folderName := filepath.Base(extractPath)

	mediaFolder := &models.Media{
		ParentID: &parentFolder.ID,
		FilePath: folderName,
		FullPath: relPath,
		Type:     "folder",
		DiskName: utils.StringPtr(FileDiskStorage),
	}

	if err := s.repo.Save(mediaFolder); err != nil {
		return nil, fmt.Errorf("failed to save folder media: %w", err)
	}

	folderIDMap := map[string]int64{
		"": parentFolder.ID, // Gốc giải nén
	}

	// Duyệt tất cả file con trong thư mục giải nén

	err = filepath.WalkDir(extractPath, func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Bỏ qua file ẩn / hệ thống
		if strings.HasPrefix(d.Name(), ".") || strings.Contains(fullPath, "__MACOSX") {
			return nil
		}

		// Đường dẫn tương đối từ public/
		rel, _ := filepath.Rel("public", fullPath)
		rel = filepath.ToSlash(rel) // Chuẩn hóa cho URL

		parentPath := filepath.ToSlash(filepath.Dir(rel))
		folderID := mediaFolder.ID

		if id, ok := folderIDMap[parentPath]; ok {
			folderID = id
		}

		if d.IsDir() {
			if path != parentPath {
				media := &models.Media{
					ParentID: &folderID,
					FilePath: d.Name(),
					Type:     "folder",
					FullPath: fmt.Sprintf("%s/%s", parentPath, d.Name()),
					DiskName: utils.StringPtr(FileDiskStorage),
				}

				if err := s.repo.Save(media); err != nil {
					config.Log.Errorf("failed to save folder %s: %v", fullPath, err)
					return nil
				}

				// Gán ID thư mục vào map
				folderIDMap[rel] = media.ID
				return nil
			}
		}

		// File: lấy thông tin
		stat, statErr := os.Stat(fullPath)
		if statErr != nil {
			return statErr
		}

		fileSize := stat.Size()
		folderSize += fileSize

		fileExt := filepath.Ext(d.Name())
		mimeType := mime.TypeByExtension(fileExt)

		if path != parentPath {
			url := utils.StringPtr(rel)
			var staticURL *string

			if url != nil {
				staticURL = utils.StringPtr(fmt.Sprintf("public/%s", *url))
			}

			media := &models.Media{
				FolderID:      folderID,
				FileName:      d.Name(),
				FilePath:      parentPath,
				FileType:      utils.StringPtr(mimeType),
				FileSize:      utils.Int64Ptr(fileSize),
				FileExtension: utils.StringPtr(fileExt),
				StaticURL:     staticURL,
				DiskName:      utils.StringPtr(FileDiskStorage),
				Type:          "file",
			}

			if saveErr := s.repo.Save(media); saveErr != nil {
				config.Log.Errorf("failed to save file %s: %v", fullPath, saveErr)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk extracted files failed: %w", err)
	}

	// Build URL & path trả về
	url := utils.StaticURL(utils.DerefStr(mediaFolder.StaticURL), models.Storage)
	if url == "" {
		url = utils.StaticURL(filepath.Join(relPath), models.Storage)
	}

	return &prot.File{
		Id:            int32(mediaFolder.ID),
		FileName:      mediaFolder.FileName,
		Url:           url,
		FileSize:      strconv.FormatInt(utils.DerefInt64(mediaFolder.FileSize), 10),
		DiskName:      utils.DerefStr(mediaFolder.DiskName),
		FileExtension: ".zip",
		FileMimeType:  "application/zip",
		Path:          filepath.Join(mediaFolder.FilePath, mediaFolder.FileName),
		Type:          "folder",
	}, nil
}

func PaginateAndSearchFiles(allMedias []models.Media, keyword string, page int, perPage int) ([]*prot.File, int) {
	var filtered []*prot.File

	normKeyword := utils.NormalizeKeyword(keyword)

	for _, media := range allMedias {
		if media.FolderID > 0 && media.Type == "file" {
			cleanFileName := utils.RemoveAccents(media.FileName)
			cleanFileName = strings.ToLower(cleanFileName)
			cleanFileName = strings.ReplaceAll(cleanFileName, " ", "-")

			if keyword != "" && !strings.Contains(cleanFileName, normKeyword) {
				continue
			}

			// Build path
			filePath := media.FilePath + "/" + media.FileName
			if utils.DerefStr(media.DiskName) == config.PowerPoint {
				rel := filepath.ToSlash(filepath.Join(media.FilePath, media.FileName))
				rel = strings.TrimPrefix(rel, "power_point/")
				filePath = "power-point/" + rel
			} else {
				if utils.DerefStr(media.DiskName) == models.Storage {
					filePath = "public/" + filePath
				}
			}

			url := utils.DerefStr(media.StaticURL)
			if url != "" {
				url = utils.StaticURL(url, *media.DiskName)
			}

			filtered = append(filtered, &prot.File{
				Id:            int32(media.ID),
				FileName:      media.FileName,
				Url:           url,
				Path:          filePath,
				FileSize:      fmt.Sprintf("%d", utils.GetInt64(media.FileSize)),
				FileMimeType:  utils.GetString(media.FileType),
				FileExtension: utils.GetString(media.FileExtension),
				DiskName:      utils.GetString(media.DiskName),
				Type:          media.Type,
			})
		}
	}

	total := len(filtered)

	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

func FormatFiles(allMedias []models.Media) []*prot.File {
	var filtered []*prot.File

	for _, media := range allMedias {
		if media.FolderID > 0 && media.Type == "file" {
			cleanFileName := utils.RemoveAccents(media.FileName)
			cleanFileName = strings.ToLower(cleanFileName)
			cleanFileName = strings.ReplaceAll(cleanFileName, " ", "-")

			// Build path
			filePath := media.FilePath + "/" + media.FileName
			if utils.DerefStr(media.DiskName) == config.PowerPoint {
				rel := filepath.ToSlash(filepath.Join(media.FilePath, media.FileName))
				rel = strings.TrimPrefix(rel, "power_point/")
				filePath = "power-point/" + rel
			} else {
				if utils.DerefStr(media.DiskName) == models.Storage {
					filePath = "public/" + filePath
				}
			}

			url := utils.DerefStr(media.StaticURL)
			if url != "" {
				url = utils.StaticURL(url, *media.DiskName)
			}

			filtered = append(filtered, &prot.File{
				Id:            int32(media.ID),
				FileName:      media.FileName,
				Url:           url,
				Path:          filePath,
				FileSize:      fmt.Sprintf("%d", utils.GetInt64(media.FileSize)),
				FileMimeType:  utils.GetString(media.FileType),
				FileExtension: utils.GetString(media.FileExtension),
				DiskName:      utils.GetString(media.DiskName),
				Type:          media.Type,
			})
		}
	}

	return filtered
}

func PaginateFiles(allMedias []models.Media, folderID int64, page int, perPage int) ([]*prot.File, int) {
	var filtered []*prot.File

	for _, media := range allMedias {
		if media.FolderID > 0 && media.FolderID == folderID && media.Type == "file" {
			filePath := media.FilePath + "/" + media.FileName
			if *media.DiskName == config.PowerPoint {
				rel := filepath.ToSlash(filepath.Join(media.FilePath, media.FileName))
				rel = strings.TrimPrefix(rel, "power_point/")
				filePath = "power-point/" + rel
			} else {
				if utils.DerefStr(media.DiskName) == models.Storage {
					filePath = "public/" + filePath
				}
			}

			url := utils.DerefStr(media.StaticURL)
			if url != "" {
				url = utils.StaticURL(url, models.Storage)
			}

			filtered = append(filtered, &prot.File{
				Id:            int32(media.ID),
				FileName:      media.FileName,
				Url:           url,
				Path:          utils.DerefStr(&filePath),
				FileSize:      fmt.Sprintf("%d", utils.GetInt64(media.FileSize)),
				FileMimeType:  utils.GetString(media.FileType),
				FileExtension: utils.GetString(media.FileExtension),
				DiskName:      utils.GetString(media.DiskName),
				Type:          media.Type,
			})
		}
	}

	total := len(filtered)

	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

func BuildMediaTree(flatList []models.Media, parentID int64) []*prot.MediaTree {
	mediaMap := make(map[int64][]*prot.MediaTree)
	rootParentID := parentID // Lưu parentID ban đầu

	for _, media := range flatList {
		segment := media.FileName
		if segment == "" {
			_, segment = path.Split(strings.TrimSuffix(media.FilePath, "/"))
		}

		node := &prot.MediaTree{
			Id:       media.ID,
			ParentId: utils.GetInt64(media.ParentID),
			Path:     segment,
			DiskName: utils.GetString(media.DiskName),
			Children: []*prot.MediaTree{},
		}

		mediaMap[node.ParentId] = append(mediaMap[node.ParentId], node)
	}

	var build func(parentID int64, parentPath string) []*prot.MediaTree

	build = func(parentID int64, parentPath string) []*prot.MediaTree {
		nodes := mediaMap[parentID]
		for _, node := range nodes {
			if parentPath == "" {
				node.FullPath = strings.Trim(node.Path, "/")
			} else {
				node.FullPath = parentPath + "/" + strings.Trim(node.Path, "/")
			}
			node.Children = build(node.Id, node.FullPath)
		}
		if parentID != rootParentID {
			sort.Slice(nodes, func(i, j int) bool {
				return strings.ToLower(nodes[i].Path) < strings.ToLower(nodes[j].Path)
			})
		}
		return nodes
	}

	roots := build(parentID, "")
	return roots
}

func CreateFolderByPath(path string, repo repositories.MediaRepository) (*models.Media, error) {
	paths := strings.Split(path, "/")
	parentID := int64(0)
	var parentIDPtr *int64 = &parentID
	var currentFolder *models.Media

	paths = append([]string{""}, paths...)

	for _, p := range paths {
		p = strings.TrimSpace(p)

		currentPath := p
		if currentFolder != nil {
			currentPath = strings.Trim(fmt.Sprintf("%s/%s", currentFolder.FullPath, p), "/")
		}

		folder, err := repo.FindByPath(currentPath, FileDiskStorage)
		if err == nil {
			parentIDPtr = &folder.ID
			currentFolder = folder
			continue
		}

		fileSizeInt64 := int64(0)

		newFolder := &models.Media{
			ParentID: parentIDPtr,
			FileName: "",
			FilePath: p,
			FullPath: currentPath,
			FileSize: utils.Int64Ptr(fileSizeInt64),
			DiskName: utils.StringPtr(FileDiskStorage),
			Type:     "folder",
		}

		if err := repo.Save(newFolder); err != nil {
			return nil, fmt.Errorf("failed to create folder %s: %w", currentPath, err)
		}

		parentIDPtr = &newFolder.ID
		currentFolder = newFolder
	}

	if currentFolder == nil {
		return nil, fmt.Errorf("no valid folder created from path: %s", path)
	}

	return currentFolder, nil
}

func (s *mediaService) ExtractZip(zipPath, extractPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create destination folder if it does not exist
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return err
	}

	// Unzip each file
	for _, file := range reader.File {
		// Normalize paths to avoid Zip Slip
		destPath := filepath.Join(extractPath, file.Name)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(extractPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		// KHÔNG bỏ qua file trong thư mục "data" - extract tất cả file để upload lên S3
		// Nếu có lỗi permission thì sẽ được xử lý gracefully ở dưới (dòng 745-772)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				// Log warning nhưng tiếp tục với file khác
				config.Log.Warn("failed to create directory during extraction: ", err.Error(), " path: ", destPath)
				continue
			}
			continue
		}

		// Make sure the parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			// Log warning nhưng tiếp tục với file khác
			config.Log.Warn("failed to create parent directory during extraction: ", err.Error(), " path: ", destPath)
			continue
		}

		// Create files
		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			// Xử lý lỗi permission một cách graceful
			if strings.Contains(err.Error(), "permission denied") {
				config.Log.Warn("permission denied when creating file during extraction, skipping: ", destPath)
				continue
			}
			return err
		}
		srcFile, err := file.Open()
		if err != nil {
			outFile.Close()
			// Xử lý lỗi khi mở file trong zip
			if strings.Contains(err.Error(), "permission denied") {
				config.Log.Warn("permission denied when opening file from zip, skipping: ", file.Name)
				continue
			}
			return err
		}

		if _, err := io.Copy(outFile, srcFile); err != nil {
			outFile.Close()
			srcFile.Close()
			// Xử lý lỗi khi copy file
			if strings.Contains(err.Error(), "permission denied") {
				config.Log.Warn("permission denied when copying file during extraction, skipping: ", destPath)
				continue
			}
			return err
		}

		outFile.Close()
		srcFile.Close()
	}

	// ➕ Cleanup phase: if there is only one sub-directory, move the contents up to the parent
	// entries, err := os.ReadDir(extractPath)
	// if err != nil {
	// 	return err
	// }

	// // Find only one subdirectory
	// if len(entries) == 1 && entries[0].IsDir() {
	// 	subDir := filepath.Join(extractPath, entries[0].Name())

	// 	// Read subdirectory contents
	// 	subEntries, err := os.ReadDir(subDir)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	for _, subEntry := range subEntries {
	// 		oldPath := filepath.Join(subDir, subEntry.Name())
	// 		newPath := filepath.Join(extractPath, subEntry.Name())

	// 		// Move file/folder up one level
	// 		if err := os.Rename(oldPath, newPath); err != nil {
	// 			return fmt.Errorf("failed to move from subdir: %v", err)
	// 		}
	// 	}

	// 	// Delete empty subfolders
	// 	if err := os.Remove(subDir); err != nil {
	// 		return fmt.Errorf("failed to remove empty subdirectory: %v", err)
	// 	}
	// }

	return nil
}

func (s *mediaService) ExtractZipOld(zipPath, extractPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if filepath.Base(extractPath) == "Root" {
		extractPath = filepath.Dir(extractPath)
	}

	// Tạo thư mục đích nếu chưa có
	if err := os.MkdirAll(extractPath, 0755); err != nil {
		return err
	}

	// Giải nén từng file
	for _, file := range reader.File {
		destPath := filepath.Join(extractPath, file.Name)

		// Chặn Zip Slip
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(extractPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, srcFile)

		// Đóng file ngay cả khi lỗi
		outFile.Close()
		srcFile.Close()

		if err != nil {
			return err
		}
	}

	// ======= Cleanup =======
	entries, err := os.ReadDir(extractPath)
	if err != nil {
		return err
	}

	// Nếu chỉ có 1 thư mục con, di chuyển nội dung lên
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(extractPath, entries[0].Name())

		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		for _, subEntry := range subEntries {
			oldPath := filepath.Join(subDir, subEntry.Name())
			newPath := filepath.Join(extractPath, subEntry.Name())

			// Nếu file/folder đích đã tồn tại thì gán tên mới
			if _, err := os.Stat(newPath); err == nil {
				newPath = newPath + "_moved"
			}

			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("failed to move from subdir: %v", err)
			}
		}

		// Xóa subdir
		if err := os.RemoveAll(subDir); err != nil {
			return fmt.Errorf("failed to remove subdirectory: %v", err)
		}
	}

	return nil
}

func (s *mediaService) ExtractRar(rarPath, extractPath string) error {
	// Ensure extract directory exists
	if err := os.MkdirAll(extractPath, 0o755); err != nil {
		return fmt.Errorf("failed to create extract directory: %v", err)
	}

	// Create a temporary directory for extraction near destination (same filesystem)
	parentDir := filepath.Dir(extractPath)
	tempDir, err := os.MkdirTemp(parentDir, ".rar_extract_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Try available tools in order: unar -> unrar -> rar
	type toolAttempt struct {
		name   string
		cmd    *exec.Cmd
		stdout bytes.Buffer
		stderr bytes.Buffer
	}

	var attempts []toolAttempt

	// Try unar first (preferred) - only if available
	if path, err := exec.LookPath("unar"); err == nil {
		// Try unar with different parameter combinations
		params := [][]string{
			{"-f", "-D", "-o", tempDir, rarPath},       // Force overwrite, no directory, output to temp
			{"-f", "-o", tempDir, rarPath},             // Force overwrite, output to temp
			{"-f", "-D", "-o", tempDir, "-q", rarPath}, // Add quiet mode
		}

		var success bool
		for i, param := range params {
			cmd := exec.Command(path, param...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			// unar often returns exit status 1 for minor issues but still extracts files successfully
			if err == nil || (err != nil && hasExtractedFiles(tempDir)) {
				success = true
				break
			}

			if i == len(params)-1 && !success {
				// All attempts failed, add to attempts list for fallback
				cmd = exec.Command(path, "-f", "-o", tempDir, rarPath)
				var finalStdout, finalStderr bytes.Buffer
				cmd.Stdout = &finalStdout
				cmd.Stderr = &finalStderr
				attempts = append(attempts, toolAttempt{name: "unar", cmd: cmd, stdout: finalStdout, stderr: finalStderr})
			}
		}

		// If unar succeeded, move contents from temp to final destination
		if success {
			if err := moveContents(tempDir, extractPath); err != nil {
				return fmt.Errorf("failed to move extracted contents: %v", err)
			}
			if err := os.Remove(rarPath); err != nil {
				log.Printf("warning: failed to delete original RAR file %s: %v", rarPath, err)
			} else {
				log.Printf("successfully deleted original RAR file: %s", rarPath)
			}
			return nil
		}
	}

	// Try unrar with better parameters for server environments (extract directly to destination)
	if path, err := exec.LookPath("unrar"); err == nil {
		// Try different unrar parameter combinations
		unrarParams := [][]string{
			{"x", "-o+", "-y", rarPath, extractPath + "/"},
			{"x", "-o+", "-y", "-r", rarPath, extractPath + "/"},
			{"e", "-o+", "-y", rarPath, extractPath + "/"},
			{"e", "-o+", "-y", "-r", rarPath, extractPath + "/"},
		}

		for _, param := range unrarParams {
			cmd := exec.Command(path, param...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil || hasExtractedFiles(extractPath) {
				// Success: delete original RAR file and return
				if err := os.Remove(rarPath); err != nil {
					log.Printf("warning: failed to delete original RAR file %s: %v", rarPath, err)
				} else {
					log.Printf("successfully deleted original RAR file: %s", rarPath)
				}
				return nil
			}
		}

		// If all unrar attempts failed, store one attempt for detailed error
		cmd := exec.Command(path, "x", "-o+", "-y", rarPath, extractPath+"/")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		attempts = append(attempts, toolAttempt{name: "unrar", cmd: cmd, stdout: stdout, stderr: stderr})
	}

	// Try rar as last resort (extract directly to destination)
	if path, err := exec.LookPath("rar"); err == nil {
		cmd := exec.Command(path, "x", "-o+", "-y", rarPath, extractPath+"/")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		attempts = append(attempts, toolAttempt{name: "rar", cmd: cmd, stdout: stdout, stderr: stderr})
	}

	if len(attempts) == 0 {
		return fmt.Errorf("no extractor found: please install 'unar' or 'unrar' (optional: 'rar')")
	}

	var lastErr error
	for i := range attempts {
		att := &attempts[i]
		if err := att.cmd.Run(); err != nil {
			lastErr = fmt.Errorf("%s failed: %v; stdout: %s; stderr: %s",
				att.name, err, att.stdout.String(), att.stderr.String())
			continue
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return fmt.Errorf("failed to extract RAR file with available tools: %v", lastErr)
	}

	// If we got here via rar/unrar fallback success, files are already in extractPath
	// Delete original RAR file after successful extraction
	if err := os.Remove(rarPath); err != nil {
		log.Printf("warning: failed to delete original RAR file %s: %v", rarPath, err)
	} else {
		log.Printf("successfully deleted original RAR file: %s", rarPath)
	}

	return nil
}

// hasExtractedFiles checks if any files were actually extracted to the directory
func hasExtractedFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

// moveContents moves all contents from src to dst, handling nested directories
func moveContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// If only one directory exists and it has the same name as the destination, move its contents
	if len(entries) == 1 && entries[0].IsDir() {
		singleDir := filepath.Join(src, entries[0].Name())
		subEntries, err := os.ReadDir(singleDir)
		if err != nil {
			return err
		}

		for _, entry := range subEntries {
			oldPath := filepath.Join(singleDir, entry.Name())
			newPath := filepath.Join(dst, entry.Name())

			if err := os.Rename(oldPath, newPath); err != nil {
				// Fallback for cross-device link (EXDEV): copy then remove
				if errors.Is(err, syscall.EXDEV) || strings.Contains(strings.ToLower(err.Error()), "cross-device") {
					if copyErr := copyEntry(oldPath, newPath); copyErr != nil {
						return fmt.Errorf("failed to copy %s: %v", entry.Name(), copyErr)
					}
					if remErr := os.RemoveAll(oldPath); remErr != nil {
						return fmt.Errorf("failed to remove source after copy %s: %v", entry.Name(), remErr)
					}
				} else {
					return fmt.Errorf("failed to move %s: %v", entry.Name(), err)
				}
			}
		}
	} else {
		// Move all files/directories directly
		for _, entry := range entries {
			oldPath := filepath.Join(src, entry.Name())
			newPath := filepath.Join(dst, entry.Name())

			if err := os.Rename(oldPath, newPath); err != nil {
				// Fallback for cross-device link (EXDEV): copy then remove
				if errors.Is(err, syscall.EXDEV) || strings.Contains(strings.ToLower(err.Error()), "cross-device") {
					if copyErr := copyEntry(oldPath, newPath); copyErr != nil {
						return fmt.Errorf("failed to copy %s: %v", entry.Name(), copyErr)
					}
					if remErr := os.RemoveAll(oldPath); remErr != nil {
						return fmt.Errorf("failed to remove source after copy %s: %v", entry.Name(), remErr)
					}
				} else {
					return fmt.Errorf("failed to move %s: %v", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

func copyEntry(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		// Preserve symlink
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	if info.IsDir() {
		return copyDir(src, dst, info)
	}
	return copyFile(src, dst, info)
}

func copyDir(src, dst string, info fs.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if err := copyEntry(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, info fs.FileInfo) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	// Ensure destination dir exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	// Create destination file with same permissions
	dstF, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() { _ = dstF.Close() }()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}
	return nil
}

func parseHumanReadableSize(sizeStr string) (int64, error) {
	units := map[string]float64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	parts := strings.Fields(sizeStr)
	if len(parts) != 2 {
		return 0, errors.New("invalid size format")
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	unit := strings.ToUpper(parts[1])
	multiplier, ok := units[unit]
	if !ok {
		return 0, errors.New("unknown size unit")
	}

	return int64(value * multiplier), nil
}

