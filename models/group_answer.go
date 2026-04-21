package models

type GroupAnswer struct {
	ID int64 `gorm:"primaryKey"`

	Content  string
	FileInfo MediaInfo `gorm:"type:jsonb" json:"file_info"`
	Kind     string    `gorm:"type:KIND_ENUM;not null"`

	SortPosition int
}
