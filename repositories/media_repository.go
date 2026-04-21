package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"be-lms/config"

	"gorm.io/gorm"
)

type MediaRepository interface {
	Save(media *models.Media) error
	FindById(id int64) (*models.Media, error)
	FindByPath(path string, disk string) (*models.Media, error)
	FindByStaticUrl(url string, disk string) (*models.Media, error)
	FindFolderByFilePathAndParent(filePath string, parentId int64, disk string) (*models.Media, error)
	FindByPathMaster(path string, disk string) (*models.Media, error)
	FindFolderByFilePathAndParentMaster(filePath string, parentId int64, disk string) (*models.Media, error)
	Delete(id int64) error
	DeleteUnscoped(id int64) error
	UpdateOrCreate(media *models.Media) error
	GetFolders(allMedias *[]models.Media) error
	GetFiles() ([]models.Media, error)
	GetFileByFolder(id int64) ([]models.Media, error)
	GetMediaInfo(url string, disk string) models.MediaInfo
	UpdateOrCreateV2(media *models.Media) error
	ClearDuplicateMedia() error
	CleanupOrphanedMedia() error
	PaginateGetFileByFolder(folderId int64, keyword string, page int, perPage int, filterTime, filterType, filterExtension string) ([]models.Media, int64, error)
	ClearMedias() error
}

type mediaRepository struct{}

func NewMediaRepository() MediaRepository {
	return &mediaRepository{}
}

func (r *mediaRepository) Save(media *models.Media) error {
	return db.MasterDB.Create(media).Error
}

func (r *mediaRepository) FindById(id int64) (*models.Media, error) {
	var media models.Media
	if err := db.ReplicaDB.First(&media, id).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindByPath(path string, disk string) (*models.Media, error) {
	var media models.Media
	if err := db.ReplicaDB.
		Where("full_path = ? AND disk_name = ?", path, disk).
		First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindFolderByFilePathAndParent(filePath string, parentId int64, disk string) (*models.Media, error) {
	var media models.Media
	if err := db.ReplicaDB.
		Where("file_path = ? AND parent_id = ? AND type = ? AND disk_name = ?", filePath, parentId, "folder", disk).
		First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindByStaticUrl(url string, disk string) (*models.Media, error) {
	var media models.Media
	if err := db.ReplicaDB.
		Where("static_url = ? AND disk_name = ?", url, disk).
		First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) FindByPathMaster(path string, disk string) (*models.Media, error) {
	var media models.Media
	publicPath := "public/" + path

	if err := db.MasterDB.
		Where("disk_name = ?", disk).
		Where("full_path = ? OR full_path = ?", path, publicPath).
		First(&media).Error; err != nil {
		return nil, err
	}

	return &media, nil
}

func (r *mediaRepository) FindFolderByFilePathAndParentMaster(filePath string, parentId int64, disk string) (*models.Media, error) {
	var media models.Media
	if err := db.MasterDB.
		Where("file_path = ? AND parent_id = ? AND type = ? AND disk_name = ?", filePath, parentId, "folder", disk).
		First(&media).Error; err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *mediaRepository) GetMediaInfo(url string, disk string) models.MediaInfo {
	mediaInfo := models.MediaInfo{
		Path: url,
		Disk: models.Storage,
	}

	if url == "" {
		return mediaInfo
	}

	var media models.Media
	if err := db.ReplicaDB.
		Where("static_url = ? AND disk_name = ?", url, disk).
		First(&media).Error; err != nil {
		return mediaInfo
	}

	mediaInfo.Id = media.ID
	mediaInfo.Path = *media.StaticURL

	return mediaInfo
}

func (r *mediaRepository) UpdateOrCreate(media *models.Media) error {
	var existing models.Media

	err := db.MasterDB.
		Unscoped().
		Where("file_path = ? AND parent_id = ? AND type = ? AND disk_name = ?",
			media.FilePath, media.ParentID, media.Type, media.DiskName).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(media).Error
	} else if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"file_name":      media.FileName,
		"file_type":      media.FileType,
		"file_size":      media.FileSize,
		"file_extension": media.FileExtension,
		"static_url":     media.StaticURL,
		"updated_at":     time.Now(),
		"updated_by":     media.UpdatedBy,
		"deleted_at":     nil,
		"deleted_by":     0,
	}

	return db.MasterDB.
		Unscoped().
		Model(&existing).
		Updates(updates).Error
}

func (r *mediaRepository) UpdateOrCreateV2(media *models.Media) error {
	var existing models.Media

	err := db.MasterDB.
		Unscoped().
		Where("static_url = ?", media.StaticURL).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(media).Error
		}
		return err
	}

	updates := map[string]interface{}{
		"file_name":      media.FileName,
		"file_type":      media.FileType,
		"file_size":      media.FileSize,
		"file_extension": media.FileExtension,
		"file_path":      media.FilePath,
		// "parent_id":     media.ParentID,
		"type":       media.Type,
		"disk_name":  media.DiskName,
		"updated_at": media.UpdatedAt,
		"updated_by": media.UpdatedBy,
		"deleted_at": nil,
		"deleted_by": 0,
	}

	return db.MasterDB.Unscoped().Model(&existing).Updates(updates).Error
}

func (r *mediaRepository) GetFolders(allMedias *[]models.Media) error {
	return db.MasterDB.
		Where("type = ?", "folder").
		Order("file_name ASC").
		Find(allMedias).Error
}

func (r *mediaRepository) GetFiles() ([]models.Media, error) {
	var allMedias []models.Media

	err := db.MasterDB.
		Where("type = ?", "file").
		Order("file_name ASC").
		Find(&allMedias).Error

	return allMedias, err
}

func (r *mediaRepository) GetFileByFolder(folderId int64) ([]models.Media, error) {
	var allMedias []models.Media

	// 1. Lấy tất cả folder id con
	var folderIDs []int64
	err := r.GetAllChildFolderIDs(folderId, &folderIDs)
	if err != nil {
		return nil, err
	}

	// Thêm folder gốc
	folderIDs = append(folderIDs, folderId)

	// 2. Query tất cả file trong các folder đó
	err = db.MasterDB.
		Where("folder_id IN ? AND type = ?", folderIDs, "file").
		Order("folder_id ASC").
		Order("file_name ASC").
		Find(&allMedias).Error

	return allMedias, err
}

func (r *mediaRepository) PaginateGetFileByFolder(folderId int64, keyword string, page int, perPage int, filterTime, filterType, filterExtension string) ([]models.Media, int64, error) {
	var medias []models.Media
	var total int64

	// 1. Lấy tất cả folder id con
	var folderIDs []int64
	err := r.GetAllChildFolderIDs(folderId, &folderIDs)
	if err != nil {
		return nil, 0, err
	}
	// Thêm folder gốc
	folderIDs = append(folderIDs, folderId)

	// 2. Query base
	query := db.MasterDB.
		Model(&models.Media{}).
		Where("folder_id IN ? AND type = ?", folderIDs, "file")

	// 3. Nếu có keyword thì thêm search theo tên
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		query = query.Where("file_name ILIKE ?", likePattern)
	}

	if filterTime != "" {
		times := strings.Split(filterTime, ",")
		if len(times) == 2 {
			start := strings.TrimSpace(times[0])
			end := strings.TrimSpace(times[1])
			query = query.Where("created_at::date BETWEEN ? AND ?", start, end)
		} else {
			query = query.Where("created_at::date = ?", filterTime)
		}
	}

	if filterType != "" {
		query = query.Where("file_type = ?", filterType)
	}

	if filterExtension != "" {
		query = query.Where("file_extension = ?", filterExtension)
	}

	// 4. Đếm tổng số record (cho phân trang)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 5. Thêm limit + offset
	offset := (page - 1) * perPage
	if err := query.
		Order("folder_id ASC").
		Order("file_name ASC").
		Limit(perPage).
		Offset(offset).
		Find(&medias).Error; err != nil {
		return nil, 0, err
	}

	return medias, total, nil
}

func (r *mediaRepository) Delete(id int64) error {
	return db.MasterDB.Where("id = ?", id).Delete(&models.Media{}).Error
}

func (r *mediaRepository) DeleteUnscoped(id int64) error {
	return db.MasterDB.Unscoped().Where("id = ?", id).Delete(&models.Media{}).Error
}

func (r *mediaRepository) GetAllChildFolderIDs(parentID int64, ids *[]int64) error {
	var allFolders []models.Media

	// Lấy toàn bộ folder
	err := db.ReplicaDB.
		Where("type = ?", "folder").
		Find(&allFolders).Error
	if err != nil {
		return err
	}

	// Build map parent -> children
	childrenMap := make(map[int64][]int64)
	for _, f := range allFolders {
		childrenMap[*f.ParentID] = append(childrenMap[*f.ParentID], f.ID)
	}

	// BFS/DFS để duyệt nhanh
	var stack []int64
	stack = append(stack, parentID)

	for len(stack) > 0 {
		// Pop
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, childID := range childrenMap[current] {
			*ids = append(*ids, childID)
			stack = append(stack, childID)
		}
	}

	return nil
}

func (r *mediaRepository) ClearDuplicateMedia() error {
	sql := `
	WITH duplicates AS (
		SELECT id,
			   ROW_NUMBER() OVER (
				   PARTITION BY static_url
				   ORDER BY
					   CASE WHEN deleted_at IS NULL THEN 0 ELSE 1 END,
					   id
			   ) AS rn
		FROM medias
		WHERE static_url IS NOT NULL
	)
	DELETE FROM medias
	WHERE id IN (
		SELECT id FROM duplicates
		WHERE rn > 1
		  AND id IN (SELECT id FROM medias WHERE deleted_at IS NOT NULL)
	);`

	if err := db.MasterDB.Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to clear duplicate media: %w", err)
	}

	return nil
}

func (r *mediaRepository) CleanupOrphanedMedia() error {
	// Step 1: Xóa các bản ghi đã bị soft-delete
	deleteSoftDeletedSQL := `
		DELETE FROM medias
		WHERE deleted_at IS NOT NULL;
	`

	// Step 2: Xóa bản ghi orphan (folder_id hoặc parent_id không tồn tại, trừ khi là 0 hoặc null)
	deleteOrphansSQL := `
		DELETE FROM medias
		WHERE
			(folder_id IS NOT NULL AND folder_id != 0 AND folder_id NOT IN (SELECT id FROM medias)) OR
			(parent_id IS NOT NULL AND parent_id != 0 AND parent_id NOT IN (SELECT id FROM medias));
	`

	tx := db.MasterDB.Begin()

	// Step 1
	if err := tx.Exec(deleteSoftDeletedSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete soft-deleted medias: %w", err)
	}

	// Step 2
	if err := tx.Exec(deleteOrphansSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete orphaned medias: %w", err)
	}

	return tx.Commit().Error
}

func (r *mediaRepository) ClearMedias() error {
	config.Log.Info("clear medias...")
	// delete all medias except the root folder and the public folder
	err := db.MasterDB.Exec("DELETE FROM medias WHERE parent_id = 0 AND id <> 2 AND type = 'folder'").Error
	if err != nil {
		return fmt.Errorf("failed to clear medias: %w", err)
	}

	sqlCleanupOrphans := `
DELETE FROM medias m
WHERE m.id <> 2
  AND (
    (
      m.type = 'folder'
      AND m.parent_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM medias p
        WHERE p.id = m.parent_id
      )
    )
    OR
    (
      m.type = 'file'
      AND m.folder_id IS NOT NULL
      AND NOT EXISTS (
        SELECT 1
        FROM medias f
        WHERE f.id = m.folder_id
      )
    )
  );`

	sqlDeleteEmptyFolders := `
DELETE FROM medias m
WHERE m.type = 'folder'
  AND m.id <> 2
  AND NOT EXISTS (
    SELECT 1
    FROM medias c1
    WHERE c1.folder_id = m.id
      AND c1.type = 'file'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM medias c2
    WHERE c2.parent_id = m.id
      AND c2.type = 'folder'
  );`

	for i := 0; i < 20; i++ {
		if err = db.MasterDB.Exec(sqlCleanupOrphans).Error; err != nil {
			return fmt.Errorf("failed to clear medias: %w", err)
		}
	}

	for i := 0; i < 20; i++ {
		if err = db.MasterDB.Exec(sqlDeleteEmptyFolders).Error; err != nil {
			return fmt.Errorf("failed to clear medias: %w", err)
		}
	}

	return nil
}
