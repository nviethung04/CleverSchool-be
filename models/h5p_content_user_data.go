package models

import "time"

type H5pContentUserData struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContentId    string    `gorm:"column:content_id;primaryKey"`
	ContextId    string    `gorm:"column:context_id"`
	DataType     string    `gorm:"column:data_type"`
	Invalidate   bool      `gorm:"column:invalidate"`
	Preload      bool      `gorm:"column:preload"`
	SubContentId string    `gorm:"column:sub_content_id"`
	UserState    string    `gorm:"column:user_state;type:jsonb"`
	UserId       string    `gorm:"column:user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
