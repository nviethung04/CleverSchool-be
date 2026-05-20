package services

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"
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
	user, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	originalMessage, err := s.messageRepo.GetByID(ctx, req.ReplyToMessageID)
	if err != nil {
		return nil, fmt.Errorf("original message not found: %w", err)
	}

	if originalMessage.CourseID != courseID {
		return nil, fmt.Errorf("message does not belong to this course")
	}

	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("reply must have either content or media attachments")
	}

	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("reply content too long")
	}

	messageType := "text"
	if len(req.Content) > 0 && len(req.MediaIDs) > 0 {
		messageType = "mixed"
	} else if len(req.MediaIDs) > 0 {
		messageType = "file"
	}

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

	err = s.messageRepo.CreateReply(ctx, reply)
	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewReply: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewReply(courseID, reply.ID, req.ReplyToMessageID, userID, reply); err != nil {
			config.Log.Error("❌ Failed to publish new reply notification: %v", err)
		}
	}()

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
	_, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, 0, fmt.Errorf("message not found: %w", err)
	}

	replies, total, err := s.messageRepo.GetRepliesByID(ctx, messageID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get replies: %w", err)
	}

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
	message, err := s.messageRepo.GetWithReplies(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	replies, total, err := s.messageRepo.GetRepliesByID(ctx, messageID, 1, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get replies: %w", err)
	}

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

func (s *chatReplyService) SendReplyWithFiles(ctx context.Context, courseID, userID uint64, req dto.SendReplyMessageRequest, files []*multipart.FileHeader) (*dto.ReplyMessageResponse, error) {
	user, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	originalMessage, err := s.messageRepo.GetByID(ctx, req.ReplyToMessageID)
	if err != nil {
		return nil, fmt.Errorf("original message not found: %w", err)
	}

	if originalMessage.CourseID != courseID {
		return nil, fmt.Errorf("message does not belong to this course")
	}

	if len(req.Content) == 0 && len(files) == 0 {
		return nil, fmt.Errorf("reply must have either content or files")
	}

	if len(req.Content) > 2000 {
		return nil, fmt.Errorf("reply content too long")
	}

	messageType := s.determineMessageType(req.Content, files)

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

	if err := s.messageRepo.CreateReply(ctx, replyMessage); err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

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

func (s *chatReplyService) determineMessageType(content string, files []*multipart.FileHeader) string {
	if len(content) > 0 {
		return "text"
	}

	if len(files) == 0 {
		return "text"
	}

	return "file"
}

func (s *chatReplyService) SendReplyWithMedias(ctx context.Context, courseID, userID uint64, messageID uint64, req dto.SendChatMessageRequest) (*dto.ChatMessageResponse, error) {
	_, err := s.userRepo.FindByID(int(userID))
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	_, err = s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("parent message not found: %w", err)
	}

	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		return nil, fmt.Errorf("reply must have either content or media attachments")
	}

	messageType := "text"
	if len(req.Content) > 0 && len(req.MediaIDs) > 0 {
		messageType = "mixed"
	} else if len(req.MediaIDs) > 0 {
		messageType = "file"
	}

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

	if len(req.MediaIDs) > 0 {
		err = s.messageRepo.CreateWithMedias(ctx, replyMessage, req.MediaIDs)
	} else {
		err = s.messageRepo.CreateReply(ctx, replyMessage)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create reply: %w", err)
	}

	createdReply, err := s.messageRepo.GetByID(ctx, replyMessage.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created reply: %w", err)
	}

	response := s.mapToResponse(createdReply)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				config.Log.Info("❌ PANIC in NotifyNewReply: %v", r)
			}
		}()

		if err := s.pubsubService.NotifyNewReply(courseID, createdReply.ID, messageID, userID, response); err != nil {
			config.Log.Error("⚠️ Failed to publish new reply notification: %v", err)
		}
	}()

	return response, nil
}

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
			FullName: message.User.Name,
		},
	}

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

func (s *chatReplyService) generateAvatarURL(avatarInfo models.MediaInfo) string {
	if avatarInfo.Path == "" {
		return ""
	}

	avatarURL := utils.StaticURL(avatarInfo.Path, avatarInfo.Disk)

	return avatarURL
}
