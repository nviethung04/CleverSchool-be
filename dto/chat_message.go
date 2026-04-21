package dto

import "time"

// Request DTOs
type SendChatMessageRequest struct {
	Content  string  `json:"content" binding:"omitempty,max=2000"`
	MediaIDs []int64 `json:"media_ids,omitempty"` // IDs của media từ bảng medias
}

// For multipart form data with files
type SendChatMessageWithFilesRequest struct {
	Content  string  `form:"content" binding:"omitempty,max=2000"`
	MediaIDs []int64 `form:"media_ids,omitempty"` // IDs của media từ bảng medias
}

type GetChatMessagesRequest struct {
	Page  int `form:"page" binding:"min=1"`
	Limit int `form:"limit" binding:"min=1,max=100"`
}

// Response DTOs
type ChatMessageResponse struct {
	ID               uint64                    `json:"id"`
	CourseID         uint64                    `json:"course_id"`
	UserID           uint64                    `json:"user_id"`
	RecipientID      *uint64                   `json:"recipient_id"`
	Content          *string                   `json:"content"`
	MessageType      string                    `json:"message_type"`
	IsPinned         bool                      `json:"is_pinned"`
	IsEdited         bool                      `json:"is_edited"`
	EditedAt         *time.Time                `json:"edited_at"`
	ReplyToMessageID *uint64                   `json:"reply_to_message_id"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	User             ChatUserResponse          `json:"user"`
	Recipient        *ChatUserResponse         `json:"recipient,omitempty"`
	ReplyToMessage   *ChatMessageResponse      `json:"reply_to_message,omitempty"`
	Files            []ChatMessageFileResponse `json:"files,omitempty"`  // Deprecated, sử dụng medias thay thế
	Medias           []ChatMediaResponse       `json:"medias,omitempty"` // Media files từ bảng medias
}

type ChatUserResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Avatar   string `json:"avatar"`
}

type ChatMediaResponse struct {
	ID            int64   `json:"id"`
	FileName      string  `json:"file_name"`
	FilePath      string  `json:"file_path"`
	FullPath      string  `json:"full_path"`
	FileType      *string `json:"file_type"`
	FileSize      *int64  `json:"file_size"`
	FileExtension *string `json:"file_extension"`
	DiskName      *string `json:"disk_name"`
	StaticURL     *string `json:"static_url"`
	SortOrder     int     `json:"sort_order"`
}

type UploadedMediaResponse struct {
	ID            int64   `json:"id"`
	FileName      string  `json:"file_name"`
	FilePath      string  `json:"file_path"`
	FullPath      string  `json:"full_path"`
	FileType      *string `json:"file_type"`
	FileSize      *int64  `json:"file_size"`
	FileExtension *string `json:"file_extension"`
	DiskName      *string `json:"disk_name"`
	StaticURL     *string `json:"static_url"`
}

type ChatMessageFileResponse struct {
	ID       uint64 `json:"id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileType string `json:"file_type"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

type ChatMessagesListResponse struct {
	Messages   []ChatMessageResponse `json:"messages"`
	Pagination PaginationResponse    `json:"pagination"`
}

type PaginationResponse struct {
	Page      int  `json:"page"`
	Limit     int  `json:"limit"`
	Total     int  `json:"total"`
	TotalPage int  `json:"total_page"`
	HasMore   bool `json:"has_more"`
}

// Reply-specific DTOs
type SendReplyMessageRequest struct {
	Content          string  `json:"content" binding:"omitempty,max=2000"` // Cho phép rỗng nếu có media_ids
	ReplyToMessageID uint64  `json:"reply_to_message_id"`                  // Không cần required vì được set từ URL param
	MediaIDs         []int64 `json:"media_ids,omitempty"`                  // IDs của media từ bảng medias
}

// For multipart form data with files for replies
type SendReplyMessageWithFilesRequest struct {
	Content          string  `form:"content" binding:"omitempty,max=2000"`
	ReplyToMessageID uint64  `form:"reply_to_message_id"` // Được set từ URL param
	MediaIDs         []int64 `form:"media_ids,omitempty"` // IDs của media từ bảng medias
}

type ReplyMessageResponse struct {
	ID               uint64                    `json:"id"`
	Content          string                    `json:"content"`
	UserID           uint64                    `json:"user_id"`
	UserName         string                    `json:"user_name"`
	ReplyToMessageID uint64                    `json:"reply_to_message_id"`
	CreatedAt        string                    `json:"created_at"`
	UpdatedAt        string                    `json:"updated_at"`
	User             ChatUserResponse          `json:"user"`
	Files            []ChatMessageFileResponse `json:"files,omitempty"`  // Deprecated
	Medias           []ChatMediaResponse       `json:"medias,omitempty"` // Media files từ bảng medias
}

type MessageWithRepliesResponse struct {
	*ChatMessageResponse
	Replies    []ReplyMessageResponse `json:"replies"`
	ReplyCount int                    `json:"reply_count"`
}

// Enhanced ChatMessageResponse to include reply info
type EnhancedChatMessageResponse struct {
	ID               uint64                    `json:"id"`
	CourseID         uint64                    `json:"course_id"`
	UserID           uint64                    `json:"user_id"`
	Content          *string                   `json:"content"`
	MessageType      string                    `json:"message_type"`
	IsPinned         bool                      `json:"is_pinned"`
	IsEdited         bool                      `json:"is_edited"`
	EditedAt         *time.Time                `json:"edited_at"`
	ReplyToMessageID *uint64                   `json:"reply_to_message_id,omitempty"`
	ReplyToMessage   *ChatMessageResponse      `json:"reply_to_message,omitempty"`
	ReplyCount       int                       `json:"reply_count"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	User             ChatUserResponse          `json:"user"`
	Files            []ChatMessageFileResponse `json:"files,omitempty"`  // Deprecated
	Medias           []ChatMediaResponse       `json:"medias,omitempty"` // Media files từ bảng medias
}

// Success response for actions
type ChatActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// Recent sender response
type RecentSenderResponse struct {
	Sender      ChatUserResponse    `json:"sender"`
	LastMessage ChatMessageResponse `json:"last_message"`
}
