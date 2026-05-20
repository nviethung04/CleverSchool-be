package dto

import (
	"be-cleverschool/models"
	"time"
)

type FeedbackResponse struct {
	ID        int64     `json:"id"`
	UserID    *int64    `json:"user_id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	FileInfo  models.MediaInfo `json:"file_info"`
	RoleID    *int64    `json:"role_id"`
	Status    *int64    `json:"status"`
	Response  string    `json:"response"`
	Note      string    `json:"note"`
	Type      *int64    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy int64     `json:"created_by"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy int64     `json:"updated_by"`
	
	// Thông tin từ bảng users
	Username string `json:"username"`
	Name     string `json:"name"`
	
	// Thông tin từ bảng roles
	RoleName string `json:"role_name"`
} 
