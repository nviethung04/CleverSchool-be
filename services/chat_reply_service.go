package services

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"context"
	"fmt"
	"mime/multipart"
	"time"
)

type ChatReplyService interface {
	SendReply(ctx context.Context, courseID, userID uint64, req dto.SendReplyMessageRequest) (*dto.ReplyMessageResponse, error)
	SendReplyWithFiles(ctx context.Context, courseID, userID uint64, req dto.SendReplyMessageRequest, files []*multipart.FileHeader) (*dto.ReplyMessageResponse, error)
	SendReplyWithMedias(ctx context.Context, courseID, userID uint64, messageID uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error)
	GetRepliesByMessageID(ctx context.Context, messageID uint64, page, limit int) ([]dto.ReplyMessageResponse, int64, error)
	GetMessageWithReplies(ctx context.Context, messageID uint64) (*dto.MessageWithRepliesResponse, error)
	DeleteReply(ctx context.Context, replyID, userID uint64) error
}

type chatReplyService struct {
	messageRepo   repositories.ChatMessageRepository
	userRepo      repositories.UserRepository
	courseRepo    repositories.CourseRepository
	pubsubService ChatPubSubService
}

func NewChatReplyService(
	messageRepo repositories.ChatMessageRepository,
	userRepo repositories.UserRepository,
	courseRepo repositories.CourseRepository,
) ChatReplyService {
	return &chatReplyService{
		messageRepo:   messageRepo,
		userRepo:      userRepo,
		courseRepo:    courseRepo,
		pubsubService: NewChatPubSubService(),
	}
}

func (s *chatReplyService) SendReply(ctx context.Context, courseID, userID uint64, req dto.SendReplyMessageRequest) (*dto.ReplyMessageResponse, error) {
	// 1. Validate user exists
	user, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Validate original message exists
	originalMessage, err := s.messageRepo.GetMessageByID(ctx, req.ReplyToMessageID)
	if err != nil {
		return nil, fmt.Errorf("original message not found: %w", err)
	}

	// 3. Validate course permission
	if originalMessage.CourseID != courseID {
		return nil, fmt.Errorf("message does not belong to this course")
	}

	// 4. Validate content or medias exist
	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("reply must have either content or media attachments")
	}

	// 5. Validate content length
	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("reply content too long")
	}

	// 6. Determine message type
	messageType := "text"
	if len(req.Content) > 0 && len(req.MediaIDs) > 0 {
		messageType = "mixed"
	} else if len(req.MediaIDs) > 0 {
		messageType = "file" // Simplified for now
	}

	// 7. Create reply message
	reply := &models.ChatMessage{
		CourseID:         courseID,
		UserID:           userID,
		Content:          &req.Content,
		MessageType:      messageType,
		ReplyToMessageID: &req.ReplyToMessageID,
		IsPinned:         false,
		IsEdited:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 6. Save reply
	err = s.messageRepo.CreateReply(ctx, reply)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	if err := s.pubsubService.NotifyNewReply(courseID, reply.ID, req.ReplyToMessageID, userID, reply); err != nil {
		fmt.Printf("Failed to publish new reply notification: %v\n", err)
	}

	// 7. Return response
	return &dto.ReplyMessageResponse{
		ID:               reply.ID,
		Content:          req.Content,
		UserID:           userID,
		UserName:         user.Name,
		ReplyToMessageID: req.ReplyToMessageID,
		CreatedAt:        reply.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        reply.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		User: dto.ChatUserResponse{
			ID:       userID,
			Username: user.Username,
			FullName: user.Name,
			Avatar:   utils.StaticURL(reply.User.AvatarInfo.Path, reply.User.AvatarInfo.Disk),
		},
	}, nil
}

func (s *chatReplyService) GetRepliesByMessageID(ctx context.Context, messageID uint64, page, limit int) ([]dto.ReplyMessageResponse, int64, error) {
	// 1. Validate message exists
	_, err := s.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, 0, fmt.Errorf("message not found: %w", err)
	}

	// 2. Get replies
	replies, total, err := s.messageRepo.GetRepliesByMessageID(ctx, messageID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get replies: %w", err)
	}

	// 3. Convert to response
	var replyResponses []dto.ReplyMessageResponse
	for _, reply := range replies {
		replyResponses = append(replyResponses, dto.ReplyMessageResponse{
			ID:               reply.ID,
			Content:          *reply.Content,
			UserID:           reply.UserID,
			UserName:         reply.User.Name,
			ReplyToMessageID: *reply.ReplyToMessageID,
			CreatedAt:        reply.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        reply.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			User: dto.ChatUserResponse{
				ID:       reply.UserID,
				Username: reply.User.Username,
				FullName: reply.User.Name,
				Avatar:   utils.StaticURL(reply.User.AvatarInfo.Path, reply.User.AvatarInfo.Disk),
			},
		})
	}

	return replyResponses, total, nil
}

func (s *chatReplyService) GetMessageWithReplies(ctx context.Context, messageID uint64) (*dto.MessageWithRepliesResponse, error) {
	// 1. Get message with basic info
	message, err := s.messageRepo.GetMessageWithReplies(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// 2. Get replies (first 10)
	replies, total, err := s.messageRepo.GetRepliesByMessageID(ctx, messageID, 1, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}

	// 3. Convert replies to response
	var replyResponses []dto.ReplyMessageResponse
	for _, reply := range replies {
		replyResponses = append(replyResponses, dto.ReplyMessageResponse{
			ID:               reply.ID,
			Content:          *reply.Content,
			UserID:           reply.UserID,
			UserName:         reply.User.Name,
			ReplyToMessageID: *reply.ReplyToMessageID,
			CreatedAt:        reply.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        reply.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			User: dto.ChatUserResponse{
				ID:       reply.UserID,
				Username: reply.User.Username,
				FullName: reply.User.Name,
				Avatar:   utils.StaticURL(reply.User.AvatarInfo.Path, reply.User.AvatarInfo.Disk),
			},
		})
	}

	// 4. Build main message response
	chatMessageResponse := &dto.ChatMessageResponse{
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
			ID:       message.UserID,
			Username: message.User.Username,
			FullName: message.User.Name,
			Avatar:   utils.StaticURL(message.User.AvatarInfo.Path, message.User.AvatarInfo.Disk),
		},
	}

	// Add reply_to_message info if this is a reply
	if message.ReplyToMessageID != nil && message.ReplyToMessage != nil {
		chatMessageResponse.ReplyToMessage = &dto.ChatMessageResponse{
			ID:          message.ReplyToMessage.ID,
			CourseID:    message.ReplyToMessage.CourseID,
			UserID:      message.ReplyToMessage.UserID,
			Content:     message.ReplyToMessage.Content,
			MessageType: message.ReplyToMessage.MessageType,
			CreatedAt:   message.ReplyToMessage.CreatedAt,
			UpdatedAt:   message.ReplyToMessage.UpdatedAt,
			User: dto.ChatUserResponse{
				ID:       message.ReplyToMessage.UserID,
				Username: message.ReplyToMessage.User.Username,
				FullName: message.ReplyToMessage.User.Name,
				Avatar:   utils.StaticURL(message.ReplyToMessage.User.AvatarInfo.Path, message.ReplyToMessage.User.AvatarInfo.Disk),
			},
		}
	}

	response := &dto.MessageWithRepliesResponse{
		ChatMessageResponse: chatMessageResponse,
		Replies:             replyResponses,
		ReplyCount:          int(total),
	}

	return response, nil
}

func (s *chatReplyService) DeleteReply(ctx context.Context, replyID, userID uint64) error {
	// 1. Get reply
	reply, err := s.messageRepo.GetMessageByID(ctx, replyID)
	if err != nil {
		return fmt.Errorf("reply not found: %w", err)
	}

	// 2. Check if it's actually a reply
	if reply.ReplyToMessageID == nil {
		return fmt.Errorf("this is not a reply message")
	}

	// 3. Check ownership
	if reply.UserID != userID {
		return fmt.Errorf("only the reply owner can delete this reply")
	}

	// 4. Delete reply (this will also delete reactions via cascade)
	err = s.messageRepo.DeleteMessage(ctx, replyID)
	if err != nil {
		return fmt.Errorf("failed to delete reply: %w", err)
	}

	return nil
}

func (s *chatReplyService) SendReplyWithFiles(ctx context.Context, courseID, userID uint64, req dto.SendReplyMessageRequest, files []*multipart.FileHeader) (*dto.ReplyMessageResponse, error) {
	// 1. Validate user exists
	user, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Validate original message exists
	originalMessage, err := s.messageRepo.GetMessageByID(ctx, req.ReplyToMessageID)
	if err != nil {
		return nil, fmt.Errorf("original message not found: %w", err)
	}

	// 3. Validate course permission
	if originalMessage.CourseID != courseID {
		return nil, fmt.Errorf("message does not belong to this course")
	}

	// 4. Validate content and files
	if len(req.Content) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("reply must have either content or files")
	}

	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("reply content too long")
	}

	// 5. Determine message type based on content and files
	messageType := s.determineMessageType(req.Content, files)

	// 6. Create reply message entity
	replyMessage := &models.ChatMessage{
		CourseID:         courseID,
		UserID:           userID,
		Content:          &req.Content,
		MessageType:      messageType,
		ReplyToMessageID: &req.ReplyToMessageID,
		IsPinned:         false,
		IsEdited:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 7. Create reply in database (skip file upload for now)
	if err := s.messageRepo.CreateReply(ctx, replyMessage); err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	// 10. Return response
	return &dto.ReplyMessageResponse{
		ID:               replyMessage.ID,
		Content:          req.Content,
		UserID:           userID,
		UserName:         user.Name,
		ReplyToMessageID: req.ReplyToMessageID,
		CreatedAt:        replyMessage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        replyMessage.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		User: dto.ChatUserResponse{
			ID:       userID,
			Username: user.Username,
			FullName: user.Name,
			Avatar:   utils.StaticURL(user.AvatarInfo.Path, user.AvatarInfo.Disk),
		},
	}, nil
}

// Helper methods (borrowed from chat message service)
func (s *chatReplyService) determineMessageType(content string, files []*multipart.FileHeader) string {
	// If there's text content, always return "text"
	if len(content) > 0 {
		return "text"
	}

	// If no text but has files, determine by file types
	if len(files) == 0 {
		return "text" // Default fallback
	}

	// For now, simplified logic - if has files, return "file"
	// TODO: Implement proper file type detection when medias integration is complete
	return "file"
}

// SendReplyWithMedias sends a reply message with media attachments using media IDs
func (s *chatReplyService) SendReplyWithMedias(ctx context.Context, courseID, userID uint64, messageID uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	fmt.Printf("=== DEBUG: SendReplyWithMedias START ===\n")
	fmt.Printf("DEBUG: courseID: %d, userID: %d, messageID: %d, content: %s, media_ids: %v\n",
		courseID, userID, messageID, req.Content, req.MediaIDs)

	// 1. Validate user exists
	_, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 2. Validate parent message exists
	_, err = s.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("parent message not found: %w", err)
	}

	// 3. Validate content or medias exist
	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("reply must have either content or media attachments")
	}

	// 4. Determine message type
	messageType := "text"
	if len(req.Content) > 0 && len(req.MediaIDs) > 0 {
		messageType = "mixed"
	} else if len(req.MediaIDs) > 0 {
		messageType = "file" // Simplified for now
	}

	// 5. Create reply message
	replyMessage := &models.ChatMessage{
		CourseID:         courseID,
		UserID:           userID,
		Content:          &req.Content,
		MessageType:      messageType,
		ReplyToMessageID: &messageID,
		IsPinned:         false,
		IsEdited:         false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// 6. Create reply with medias
	if len(req.MediaIDs) > 0 {
		err = s.messageRepo.CreateMessageWithMedias(ctx, replyMessage, req.MediaIDs)
	} else {
		err = s.messageRepo.CreateReply(ctx, replyMessage)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	// 7. Load the created reply with all relations
	createdReply, err := s.messageRepo.GetMessageByID(ctx, replyMessage.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created reply: %w", err)
	}

	// 8. Map to response DTO (reuse logic from chat message service)
	response := s.mapToResponse(createdReply)

	// 9. Publish real-time notification via Redis Pub/Sub
	if err := s.pubsubService.NotifyNewReply(courseID, createdReply.ID, messageID, userID, response); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to publish new reply notification: %v\n", err)
	}

	fmt.Printf("DEBUG: SendReplyWithMedias SUCCESS - reply ID: %d\n", createdReply.ID)
	fmt.Printf("=== DEBUG: SendReplyWithMedias END ===\n")
	return response, nil
}

// mapToResponse maps ChatMessage model to ChatMessageResponse DTO
func (s *chatReplyService) mapToResponse(message *models.ChatMessage) *dto.ChatMessageResponse {
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
			FullName: message.User.Name, // Adjusted based on actual User model
			// Avatar:   message.User.Avatar, // Add if available
		},
	}

	// Map medias from message_medias relationship
	if len(message.MessageMedias) > 0 {
		response.Medias = make([]dto.ChatMediaResponse, len(message.MessageMedias))
		for i, messageMedia := range message.MessageMedias {
			response.Medias[i] = dto.ChatMediaResponse{
				ID:            messageMedia.Media.ID,
				FileName:      messageMedia.Media.FileName,
				FilePath:      messageMedia.Media.FilePath,
				FullPath:      messageMedia.Media.FullPath,
				FileType:      messageMedia.Media.FileType,
				FileSize:      messageMedia.Media.FileSize,
				FileExtension: messageMedia.Media.FileExtension,
				DiskName:      messageMedia.Media.DiskName,
				StaticURL:     messageMedia.Media.StaticURL,
				SortOrder:     messageMedia.SortOrder,
			}
		}
	}

	// Map reply to message if exists
	if message.ReplyToMessage != nil {
		response.ReplyToMessage = &dto.ChatMessageResponse{
			ID:          message.ReplyToMessage.ID,
			CourseID:    message.ReplyToMessage.CourseID,
			UserID:      message.ReplyToMessage.UserID,
			Content:     message.ReplyToMessage.Content,
			MessageType: message.ReplyToMessage.MessageType,
			IsPinned:    message.ReplyToMessage.IsPinned,
			IsEdited:    message.ReplyToMessage.IsEdited,
			CreatedAt:   message.ReplyToMessage.CreatedAt,
			UpdatedAt:   message.ReplyToMessage.UpdatedAt,
			User: dto.ChatUserResponse{
				ID:       uint64(message.ReplyToMessage.User.ID),
				Username: message.ReplyToMessage.User.Username,
				FullName: message.ReplyToMessage.User.Name,
			},
		}
	}

	return response
}

// generateAvatarURL generates avatar URL with debug logging
func (s *chatReplyService) generateAvatarURL(avatarInfo models.MediaInfo) string {
	fmt.Printf("🖼️ DEBUG Avatar - Path: '%s', Disk: '%s'\n", avatarInfo.Path, avatarInfo.Disk)

	if avatarInfo.Path == "" {
		fmt.Printf("🖼️ DEBUG Avatar - Empty path, returning empty string\n")
		return ""
	}

	avatarURL := utils.StaticURL(avatarInfo.Path, avatarInfo.Disk)
	fmt.Printf("🖼️ DEBUG Avatar - Generated URL: '%s'\n", avatarURL)

	return avatarURL
}
