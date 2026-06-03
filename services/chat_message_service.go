package services

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type ChatMessageService interface {
	SendMessage(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	SendMessageWithFiles(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest, files []*multipart.FileHeader) (*dto.ChatMessageResponse, error)
	SendMessageWithMedias(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	UploadFilesToMedias(ctx context.Context, courseId uint64, userId uint64, files []*multipart.FileHeader) ([]dto.UploadedMediaResponse, error)
	GetMessages(ctx context.Context, courseId uint64, page, limit int) (*dto.ChatMessagesListResponse, error)
	DeleteMessage(ctx context.Context, messageId uint64, userId uint64) error
	TogglePinMessage(ctx context.Context, messageId uint64, userId uint64) error
	GetPinnedMessages(ctx context.Context, courseId uint64) ([]dto.ChatMessageResponse, error)
	CountMessages(ctx context.Context, courseId uint64) (int64, error)
	GetRecentMessages(ctx context.Context, courseId uint64, limit int) ([]dto.ChatMessageResponse, error)
}

type chatMessageService struct {
	repo             repositories.ChatMessageRepository
	userRepo         repositories.UserRepository
	courseRepo       repositories.CourseRepository
	reactionRepo     repositories.ChatMessageReactionRepository
	messageMediaRepo repositories.MessageMediaRepository
	mediaRepo        repositories.MediaRepository
	pubsubService    ChatPubSubService
}

func NewChatMessageService(
	repo repositories.ChatMessageRepository,
	userRepo repositories.UserRepository,
	courseRepo repositories.CourseRepository,
	reactionRepo repositories.ChatMessageReactionRepository,
	messageMediaRepo repositories.MessageMediaRepository,
	mediaRepo repositories.MediaRepository,
) ChatMessageService {
	return &chatMessageService{
		repo:             repo,
		userRepo:         userRepo,
		courseRepo:       courseRepo,
		reactionRepo:     reactionRepo,
		messageMediaRepo: messageMediaRepo,
		mediaRepo:        mediaRepo,
		pubsubService:    NewChatPubSubService(),
	}
}

// determineMessageType determines message type based on content and files
func (s *chatMessageService) determineMessageType(content string, files []*multipart.FileHeader) string {
	// If there's any text content, always return "text" regardless of files
	if strings.TrimSpace(content) != "" {
		return "text"
	}

	// If no text, determine type based on files
	if len(files) == 0 {
		return "text" // Default to text if no content and no files
	}

	if len(files) == 1 {
		// Single file: determine type by extension
		filename := strings.ToLower(files[0].Filename)
		if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") ||
			strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".gif") ||
			strings.HasSuffix(filename, ".webp") {
			return "image"
		}
		if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".avi") ||
			strings.HasSuffix(filename, ".mov") || strings.HasSuffix(filename, ".wmv") {
			return "video"
		}
		if strings.HasSuffix(filename, ".mp3") || strings.HasSuffix(filename, ".wav") ||
			strings.HasSuffix(filename, ".flac") || strings.HasSuffix(filename, ".aac") {
			return "audio"
		}
		return "file" // Other single file types
	}

	// Multiple files: check if all are the same type
	var fileTypes []string
	for _, file := range files {
		filename := strings.ToLower(file.Filename)
		if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") ||
			strings.HasSuffix(filename, ".png") || strings.HasSuffix(filename, ".gif") ||
			strings.HasSuffix(filename, ".webp") {
			fileTypes = append(fileTypes, "image")
		} else if strings.HasSuffix(filename, ".mp4") || strings.HasSuffix(filename, ".avi") ||
			strings.HasSuffix(filename, ".mov") || strings.HasSuffix(filename, ".wmv") {
			fileTypes = append(fileTypes, "video")
		} else if strings.HasSuffix(filename, ".mp3") || strings.HasSuffix(filename, ".wav") ||
			strings.HasSuffix(filename, ".flac") || strings.HasSuffix(filename, ".aac") {
			fileTypes = append(fileTypes, "audio")
		} else {
			fileTypes = append(fileTypes, "file")
		}
	}

	// Check if all files are the same type
	if len(fileTypes) > 0 {
		firstType := fileTypes[0]
		allSameType := true
		for _, t := range fileTypes {
			if t != firstType {
				allSameType = false
				break
			}
		}
		if allSameType {
			return firstType
		}
	}

	return "mixed" // Multiple files of different types
}

// SendMessage xử lý logic gửi tin nhắn
func (s *chatMessageService) SendMessage(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	// 1. Business Logic: Validate user exists
	_, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Business Logic: Skip course validation for now to avoid gin context issue
	// TODO: Implement proper course validation that doesn't require gin context
	// _, err = s.courseRepo.FindByID(int(courseId))
	// if err != nil {
	//     return nil, fmt.Errorf("course not found: %w", err)
	// }

	// 3. Business Logic: Check user permission to send message in this course
	// TODO: Implement course membership check
	// hasPermission := s.checkUserCoursePermission(userId, courseId)
	// if !hasPermission {
	//     return nil, fmt.Errorf("user does not have permission to send message in this course")
	// }

	// 4. Business Logic: Validate content and determine message type
	if len(req.Content) == 0 {
		return nil, fmt.Errorf("message content cannot be empty")
	}
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	// Automatically determine message type - always "text" for simple messages
	messageType := "text"

	// 5. Create message entity
	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType, // Use determined message type
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 6. Delegate to Repository for data persistence
	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// 7. Load complete data with relations for response
	createdMessage, err := s.repo.GetMessageByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created message: %w", err)
	}

	// 8. Transform to DTO
	response := s.mapToResponse(createdMessage)

	// 9. Publish real-time notification via Redis Pub/Sub (was missing -> only messages with medias had events)
	if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
		fmt.Printf("Failed to publish new message notification: %v\n", err)
	}

	return response, nil
}

// SendMessageWithMedias gửi tin nhắn với media từ bảng medias
func (s *chatMessageService) SendMessageWithMedias(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	// 1. Business Logic: Validate user exists
	_, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Business Logic: Skip course validation for now to avoid gin context issue
	// TODO: Implement proper course validation that doesn't require gin context
	// _, err = s.courseRepo.FindByID(int(courseId))
	// if err != nil {
	//     return nil, fmt.Errorf("course not found: %w", err)
	// }

	// 3. Business Logic: Validate content or medias exist
	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("message must have either content or media attachments")
	}
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	// 4. Business Logic: Validate media IDs if provided
	if len(req.MediaIDs) > 0 {
		for _, mediaID := range req.MediaIDs {
			_, err := s.mediaRepo.FindById(mediaID)
			if err != nil {
				return nil, fmt.Errorf("media with ID %d not found: %w", mediaID, err)
			}
		}
	}

	// 5. Business Logic: Determine message type
	messageType := "text"
	if len(req.MediaIDs) > 0 {
		messageType = s.determineMessageTypeFromMedias(ctx, req.MediaIDs)
	}

	// 6. Create message entity
	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType,
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 7. Create message with medias using transaction
	if err := s.repo.CreateMessageWithMedias(ctx, message, req.MediaIDs); err != nil {
		return nil, fmt.Errorf("failed to create message with medias: %w", err)
	}

	// 8. Load complete data with relations for response
	createdMessage, err := s.repo.GetMessageByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created message: %w", err)
	}

	// 9. Transform to DTO
	response := s.mapToResponse(createdMessage)

	// 10. Publish real-time notification via Redis Pub/Sub
	if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish new message notification: %v\n", err)
	}

	return response, nil
}

// SendMessageWithFiles xử lý logic gửi tin nhắn kèm file
func (s *chatMessageService) SendMessageWithFiles(ctx context.Context, courseId uint64, userId uint64, req dto.SendChatMessageRequest, files []*multipart.FileHeader) (*dto.ChatMessageResponse, error) {
	// 1. Business Logic: Validate user exists
	_, err := s.userRepo.FindByID(int(userId))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Business Logic: Skip course validation for now to avoid gin context issue
	// TODO: Implement proper course validation that doesn't require gin context
	// _, err = s.courseRepo.FindByID(int(courseId))
	// if err != nil {
	//     return nil, fmt.Errorf("course not found: %w", err)
	// }

	// 3. Business Logic: Determine message type based on content and files
	messageType := s.determineMessageType(req.Content, files)

	// Validate message type
	validTypes := map[string]bool{
		"text":  true,
		"file":  true,
		"image": true,
		"video": true,
		"audio": true,
		"mixed": true, // For multiple different file types without text
	}
	if !validTypes[messageType] {
		return nil, fmt.Errorf("invalid message type: %s", messageType)
	}

	// 4. Business Logic: Validate content (optional for file messages)
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("message content too long")
	}

	// Validate that we have either content or files
	if len(req.Content) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("message must have either content or files")
	}

	// 5. Create message entity
	message := &models.ChatMessage{
		CourseID:    courseId,
		UserID:      userId,
		Content:     &req.Content,
		MessageType: messageType, // Use the determined message type
		IsPinned:    false,
		IsEdited:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 6. Create message first
	if err := s.repo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// 7.5 (adjust numbering) Load complete data with relations for response
	createdMessage, err := s.repo.GetMessageByID(ctx, message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created message: %w", err)
	}

	// 8. Transform to DTO
	response := s.mapToResponse(createdMessage)

	// 9. Publish real-time notification via Redis Pub/Sub (previously missing)
	if err := s.pubsubService.NotifyNewMessage(courseId, message.ID, userId, response); err != nil {
		fmt.Printf("Failed to publish new message notification: %v\n", err)
	}

	return response, nil
}

// GetMessages xử lý logic lấy danh sách tin nhắn với phân trang
func (s *chatMessageService) GetMessages(ctx context.Context, courseId uint64, page, limit int) (*dto.ChatMessagesListResponse, error) {
	// 1. Business Logic: Validate and sanitize pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 2. Delegate to Repository
	messages, total, err := s.repo.GetMessagesByCourseID(ctx, courseId, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// 3. Business Logic: Transform to DTO
	responseMessages := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responseMessages[i] = *s.mapToResponse(&msg)
	}

	// 4. Business Logic: Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &dto.ChatMessagesListResponse{
		Messages: responseMessages,
		Pagination: dto.PaginationResponse{
			Page:      page,
			Limit:     limit,
			Total:     int(total),
			TotalPage: totalPages,
			HasMore:   page < totalPages,
		},
	}, nil
}

// DeleteMessage xử lý logic xóa tin nhắn với authorization
func (s *chatMessageService) DeleteMessage(ctx context.Context, messageId uint64, userId uint64) error {
	// 1. Business Logic: Validate message exists
	message, err := s.repo.GetMessageByID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	// 2. Business Logic: Authorization - only message owner can delete
	if message.UserID != userId {
		return fmt.Errorf("unauthorized: only message owner can delete this message")
	}

	// 3. Business Logic: Check if message can be deleted (e.g., not too old)
	if time.Since(message.CreatedAt) > 24*time.Hour {
		return fmt.Errorf("cannot delete message older than 24 hours")
	}

	// 4. Delete all reactions for this message first (cascade delete)
	err = s.reactionRepo.DeleteAllReactionsByMessageID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("failed to delete message reactions: %w", err)
	}

	// 5. Delete all replies for this message (cascade delete)
	err = s.repo.DeleteRepliesByMessageID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("failed to delete message replies: %w", err)
	}

	// 6. Delegate to Repository to delete the message
	err = s.repo.DeleteMessage(ctx, messageId)
	if err != nil {
		return err
	}

	// 7. Publish real-time notification via Redis Pub/Sub
	if err := s.pubsubService.NotifyMessageDeleted(message.CourseID, messageId, userId); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish message deleted notification: %v\n", err)
	}

	return nil
}

// TogglePinMessage xử lý logic pin/unpin tin nhắn
func (s *chatMessageService) TogglePinMessage(ctx context.Context, messageId uint64, userId uint64) error {
	// 1. Business Logic: Validate message exists
	message, err := s.repo.GetMessageByID(ctx, messageId)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	// 2. Business Logic: Authorization - check if user can pin messages
	// TODO: Implement role-based permission check
	// hasPermission := s.checkPinPermission(userId, message.CourseID)
	// if !hasPermission {
	//     return fmt.Errorf("unauthorized: user does not have permission to pin messages")
	// }

	// Get current pin status before toggle
	wasPinned := message.IsPinned

	// 3. Delegate to Repository
	err = s.repo.TogglePinMessage(ctx, messageId)
	if err != nil {
		return err
	}

	// 4. Publish real-time notification via Redis Pub/Sub
	newPinStatus := !wasPinned
	if err := s.pubsubService.NotifyMessagePinned(message.CourseID, messageId, userId, newPinStatus); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish message pin notification: %v\n", err)
	}

	return nil
}

// GetPinnedMessages lấy danh sách tin nhắn đã pin
func (s *chatMessageService) GetPinnedMessages(ctx context.Context, courseId uint64) ([]dto.ChatMessageResponse, error) {
	messages, err := s.repo.GetPinnedMessagesByCourseID(ctx, courseId)
	if err != nil {
		return nil, fmt.Errorf("failed to get pinned messages: %w", err)
	}

	responses := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.mapToResponse(&msg)
	}

	return responses, nil
}

// CountMessages đếm tổng số tin nhắn trong khóa học
func (s *chatMessageService) CountMessages(ctx context.Context, courseId uint64) (int64, error) {
	return s.repo.CountMessagesByCourseID(ctx, courseId)
}

// GetRecentMessages lấy tin nhắn gần đây
func (s *chatMessageService) GetRecentMessages(ctx context.Context, courseId uint64, limit int) ([]dto.ChatMessageResponse, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	messages, err := s.repo.GetRecentMessagesByCourseID(ctx, courseId, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}

	responses := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.mapToResponse(&msg)
	}

	return responses, nil
}

// mapToResponse chuyển đổi model thành DTO response
func (s *chatMessageService) mapToResponse(message *models.ChatMessage) *dto.ChatMessageResponse {
	fmt.Printf("🔄 DEBUG mapToResponse - Message ID: %d, User ID: %d\n", message.ID, message.UserID)

	response := &dto.ChatMessageResponse{
		ID:               message.ID,
		CourseID:         message.CourseID,
		UserID:           message.UserID,
		Content:          message.Content,
		MessageType:      message.MessageType,
		IsPinned:         message.IsPinned,
		IsEdited:         message.IsEdited,
		EditedAt:         message.EditedAt,
		ReplyToMessageID: message.ReplyToMessageID,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
		User: dto.ChatUserResponse{
			ID:       uint64(message.User.ID),
			Username: message.User.Username,
			FullName: message.User.Name,                            // Adjusted based on actual User model
			Avatar:   s.generateAvatarURL(message.User.AvatarInfo), // Use helper function with debug
		},
	}

	// Map medias from message_medias relationship
	if len(message.MessageMedias) > 0 {
		response.Medias = make([]dto.ChatMediaResponse, len(message.MessageMedias))
		for i, messageMedia := range message.MessageMedias {
			// Tạo full URL cho static_url
			var fullStaticURL *string
			if messageMedia.Media.StaticURL != nil {
				diskName := "s3" // Default disk
				if messageMedia.Media.DiskName != nil {
					diskName = *messageMedia.Media.DiskName
				}
				fullURL := utils.StaticURL(*messageMedia.Media.StaticURL, diskName)
				fullStaticURL = &fullURL
			}

			response.Medias[i] = dto.ChatMediaResponse{
				ID:            messageMedia.Media.ID,
				FileName:      messageMedia.Media.FileName,
				FilePath:      messageMedia.Media.FilePath,
				FullPath:      messageMedia.Media.FullPath,
				FileType:      messageMedia.Media.FileType,
				FileSize:      messageMedia.Media.FileSize,
				FileExtension: messageMedia.Media.FileExtension,
				DiskName:      messageMedia.Media.DiskName,
				StaticURL:     fullStaticURL,
				SortOrder:     messageMedia.SortOrder,
			}
		}
	}

	return response
}

// determineMessageTypeFromMedias xác định loại tin nhắn dựa trên media IDs
func (s *chatMessageService) determineMessageTypeFromMedias(ctx context.Context, mediaIDs []int64) string {
	if len(mediaIDs) == 0 {
		return "text"
	}

	var fileTypes []string
	for _, mediaID := range mediaIDs {
		media, err := s.mediaRepo.FindById(mediaID)
		if err != nil {
			continue // Skip if media not found
		}

		if media.FileExtension == nil {
			fileTypes = append(fileTypes, "file")
			continue
		}

		ext := strings.ToLower(*media.FileExtension)

		// Images
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		isImage := false
		for _, imgExt := range imageExts {
			if ext == imgExt {
				fileTypes = append(fileTypes, "image")
				isImage = true
				break
			}
		}
		if isImage {
			continue
		}

		// Videos
		videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
		isVideo := false
		for _, vidExt := range videoExts {
			if ext == vidExt {
				fileTypes = append(fileTypes, "video")
				isVideo = true
				break
			}
		}
		if isVideo {
			continue
		}

		// Audio
		audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
		isAudio := false
		for _, audExt := range audioExts {
			if ext == audExt {
				fileTypes = append(fileTypes, "audio")
				isAudio = true
				break
			}
		}
		if isAudio {
			continue
		}

		// Default to file
		fileTypes = append(fileTypes, "file")
	}

	// Check if all files are the same type
	if len(fileTypes) > 0 {
		firstType := fileTypes[0]
		allSameType := true
		for _, t := range fileTypes {
			if t != firstType {
				allSameType = false
				break
			}
		}
		if allSameType {
			return firstType
		}
	}

	return "mixed" // Multiple files of different types
}

// determineMessageTypeFromFiles xác định loại tin nhắn dựa trên file
func (s *chatMessageService) determineMessageTypeFromFiles(files []*multipart.FileHeader) string {
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))

		// Images
		imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
		for _, imgExt := range imageExts {
			if ext == imgExt {
				return "image"
			}
		}

		// Videos
		videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
		for _, vidExt := range videoExts {
			if ext == vidExt {
				return "video"
			}
		}

		// Audio
		audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
		for _, audExt := range audioExts {
			if ext == audExt {
				return "audio"
			}
		}
	}

	// Default to file if not image/video/audio
	return "file"
}

// getFileType xác định loại file dựa trên extension
func (s *chatMessageService) getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".mkv": true}
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true}

	if imageExts[ext] {
		return "image"
	}
	if videoExts[ext] {
		return "video"
	}
	if audioExts[ext] {
		return "audio"
	}

	return "document"
}

// UploadFilesToMedias upload files và lưu vào bảng medias, trả về danh sách media IDs
func (s *chatMessageService) UploadFilesToMedias(ctx context.Context, courseId uint64, userId uint64, files []*multipart.FileHeader) ([]dto.UploadedMediaResponse, error) {
	fmt.Printf("=== DEBUG: UploadFilesToMedias START ===\n")
	fmt.Printf("DEBUG: courseId: %d, userId: %d, files count: %d\n", courseId, userId, len(files))

	if len(files) == 0 {
		return []dto.UploadedMediaResponse{}, nil
	}

	var uploadedMedias []dto.UploadedMediaResponse

	for i, fileHeader := range files {
		fmt.Printf("DEBUG: Processing file %d: %s\n", i, fileHeader.Filename)

		// 1. Upload file to storage và tạo record trong bảng medias
		// Bạn có thể sử dụng service upload file đã có sẵn ở đây
		// Ví dụ: s.mediaService.UploadFile(...)

		// Giả sử bạn có MediaService để handle upload:
		media, err := s.uploadFileToMedia(ctx, courseId, userId, fileHeader)
		if err != nil {
			fmt.Printf("DEBUG: Failed to upload file %s: %v\n", fileHeader.Filename, err)
			return nil, fmt.Errorf("failed to upload file %s: %w", fileHeader.Filename, err)
		}

		// 2. Convert media model to DTO response
		uploadedMedia := dto.UploadedMediaResponse{
			ID:            media.ID,
			FileName:      media.FileName,
			FilePath:      media.FilePath,
			FullPath:      media.FullPath,
			FileType:      media.FileType,
			FileSize:      media.FileSize,
			FileExtension: media.FileExtension,
			DiskName:      media.DiskName,
			StaticURL:     media.StaticURL,
		}

		uploadedMedias = append(uploadedMedias, uploadedMedia)
		fmt.Printf("DEBUG: Successfully uploaded file %s with media ID: %d\n", fileHeader.Filename, media.ID)
	}

	fmt.Printf("DEBUG: UploadFilesToMedias END - uploaded %d files\n", len(uploadedMedias))
	return uploadedMedias, nil
}

// uploadFileToMedia helper method để upload single file vào bảng medias
func (s *chatMessageService) uploadFileToMedia(ctx context.Context, courseId uint64, userId uint64, fileHeader *multipart.FileHeader) (*models.Media, error) {
	fmt.Printf("DEBUG: uploadFileToMedia START for file: %s\n", fileHeader.Filename)

	// TODO: Implement logic upload file to storage và lưu vào bảng medias
	// Đây là placeholder - bạn cần implement logic này dựa trên hệ thống upload hiện có

	// Ví dụ implementation (cần thay đổi theo logic thực tế của bạn):
	// 1. Upload file to S3/local storage
	// 2. Create Media record in database
	// 3. Return Media model

	// Tạm thời return mock data để fix compilation error
	media := &models.Media{
		ID:            123, // This should be actual ID from database
		FileName:      fileHeader.Filename,
		FilePath:      fmt.Sprintf("chat_files/course_%d/%s", courseId, fileHeader.Filename),
		FullPath:      fmt.Sprintf("storage/chat_files/course_%d/%s", courseId, fileHeader.Filename),
		FileType:      &[]string{s.getFileType(fileHeader.Filename)}[0],
		FileSize:      &fileHeader.Size,
		FileExtension: &[]string{filepath.Ext(fileHeader.Filename)}[0],
		DiskName:      &[]string{"s3"}[0],
		StaticURL:     &[]string{fmt.Sprintf("/storage/chat_files/course_%d/%s", courseId, fileHeader.Filename)}[0],
		Type:          "file",
		CreatedBy:     int64(userId),
		UpdatedBy:     int64(userId),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save to database using mediaRepo
	if err := s.mediaRepo.Save(media); err != nil {
		return nil, fmt.Errorf("failed to save media to database: %w", err)
	}

	fmt.Printf("DEBUG: uploadFileToMedia END - media ID: %d\n", media.ID)
	return media, nil
}

// generateAvatarURL generates avatar URL with debug logging
func (s *chatMessageService) generateAvatarURL(avatarInfo models.MediaInfo) string {
	fmt.Printf("🖼️ DEBUG Avatar - Path: '%s', Disk: '%s'\n", avatarInfo.Path, avatarInfo.Disk)

	if avatarInfo.Path == "" {
		fmt.Printf("🖼️ DEBUG Avatar - Empty path, returning empty string\n")
		return ""
	}

	avatarURL := utils.StaticURL(avatarInfo.Path, avatarInfo.Disk)
	fmt.Printf("🖼️ DEBUG Avatar - Generated URL: '%s'\n", avatarURL)

	return avatarURL
}
