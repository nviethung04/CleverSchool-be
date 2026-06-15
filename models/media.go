package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"be-lms/config"

	"gorm.io/gorm"
)

// Storage is the active media disk (public local or s3/R2). Set via InitStorage after .env load.
var Storage string

func InitStorage() {
	Storage = config.MediaDisk()
}

func StorageDisk() string {
	if Storage != "" {
		return Storage
	}
	return config.MediaDisk()
}

func ResolveStorageDisk(preferred string) string {
	if preferred != "" {
		return preferred
	}
	return StorageDisk()
}

func (Media) TableName() string {
	return "medias"
}

type Media struct {
	ID            int64   `gorm:"primaryKey;autoIncrement,column:id" json:"id"`
	ParentID      *int64  `gorm:"column:parent_id" json:"parent_id"`
	FolderID      int64   `gorm:"column:folder_id" json:"folder_id"`
	FileName      string  `gorm:"column:file_name" json:"file_name"`
	FilePath      string  `gorm:"column:file_path" json:"file_path"`
	FullPath      string  `gorm:"column:full_path" json:"full_path"`
	FileType      *string `gorm:"column:file_type" json:"file_type"`
	FileSize      *int64  `gorm:"column:file_size" json:"file_size"`
	FileExtension *string `gorm:"column:file_extension" json:"file_extension"`
	DiskName      *string `gorm:"column:disk_name" json:"disk_name"`
	StaticURL     *string `gorm:"column:static_url" json:"static_url"`
	Type          string  `gorm:"column:type;default:file" json:"type"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`

	Children []Media `gorm:"-" json:"children,omitempty"`
}

type MediaInfo struct {
	Id   int64  `json:"id"`
	Path string `json:"path"`
	Disk string `json:"disk"`
}

type MediaDetail struct {
	Id   int64  `json:"id"`
	Path string `json:"path"`
	Disk string `json:"disk"`
	Type string `json:"type"`
}

// Scan implements sql.Scanner
func (m *MediaInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan MediaInfo: wrong type %T", value)
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer
func (m MediaInfo) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func SyncMediaInfoByPath(db *gorm.DB, fileUrl string) MediaInfo {
	if fileUrl == "" {
		return MediaInfo{}
	}

	var media Media
	err := db.Where("static_url = ?", fileUrl).First(&media).Error

	if err != nil {
		return MediaInfo{
			Id:   0,
			Path: fileUrl,
			Disk: "",
		}
	}

	return MediaInfo{
		Id:   media.ID,
		Path: *media.StaticURL,
		Disk: *media.DiskName,
	}
}

type MediaInfos []MediaDetail

// Scan implements sql.Scanner for MediaInfos
func (m *MediaInfos) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan MediaInfos: wrong type %T", value)
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for MediaInfos
func (m MediaInfos) Value() (driver.Value, error) {
	return json.Marshal(m)
}
