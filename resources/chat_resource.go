package resources

import (
	"be-lms/dto"
	"be-lms/prot"
)

type ChatResource interface {
	FormatChat(chat *dto.ChatMessageResponse) *prot.ChatMessageResponse
	FormatChats(chats []dto.ChatMessageResponse) []*prot.ChatMessageResponse
	FormatChatMedia(media *dto.UploadedMediaResponse) *prot.UploadedMediaResponse
	FormatChatMedias(medias []dto.UploadedMediaResponse) []*prot.UploadedMediaResponse
}

type ChatResourceImpl struct{}

func NewChatResource() ChatResource {
	return &ChatResourceImpl{}
}

func (r *ChatResourceImpl) FormatChat(chat *dto.ChatMessageResponse) *prot.ChatMessageResponse {
	if chat == nil {
		return nil
	}

	var user prot.ChatUserResponse
	if chat.User.ID > 0 {
		user = prot.ChatUserResponse{
			Id:       chat.User.ID,
			Username: chat.User.Username,
			FullName: chat.User.FullName,
			Avatar:   chat.User.Avatar,
		}
	}

	var recipient *prot.ChatUserResponse
	if chat.Recipient != nil {
		recipient = &prot.ChatUserResponse{
			Id:       chat.Recipient.ID,
			Username: chat.Recipient.Username,
			FullName: chat.Recipient.FullName,
			Avatar:   chat.Recipient.Avatar,
		}
	}

	var replyTo *prot.ChatMessageResponse
	if chat.ReplyToMessage != nil && chat.ReplyToMessage.User.ID > 0 {
		replyTo = &prot.ChatMessageResponse{
			Id:      chat.ReplyToMessage.ID,
			Content: *chat.ReplyToMessage.Content,
			User: &prot.ChatUserResponse{
				Id:       chat.ReplyToMessage.User.ID,
				Username: chat.ReplyToMessage.User.Username,
				FullName: chat.ReplyToMessage.User.FullName,
				Avatar:   chat.ReplyToMessage.User.Avatar,
			},
			CreatedAt: chat.ReplyToMessage.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	files := make([]*prot.ChatMessageFileResponse, 0, len(chat.Files))
	for _, f := range chat.Files {
		files = append(files, &prot.ChatMessageFileResponse{
			Id:       f.ID,
			FileName: f.FileName,
			FileUrl:  f.FileURL,
			FileType: f.FileType,
			FileSize: f.FileSize,
			MimeType: f.MimeType,
		})
	}

	medias := make([]*prot.ChatMediaResponse, 0, len(chat.Medias))
	for _, m := range chat.Medias {
		medias = append(medias, &prot.ChatMediaResponse{
			Id:            m.ID,
			FileName:      m.FileName,
			FilePath:      m.FilePath,
			FullPath:      m.FullPath,
			FileType:      m.FileType,
			FileSize:      m.FileSize,
			FileExtension: m.FileExtension,
			DiskName:      m.DiskName,
			StaticUrl:     m.StaticURL,
			SortOrder:     int32(m.SortOrder),
		})
	}

	var editedAt string
	if chat.EditedAt != nil {
		editedAt = chat.EditedAt.Format("2006-01-02 15:04:05")
	} else {
		editedAt = ""
	}

	var replyToMessageId uint64
	if chat.ReplyToMessageID != nil {
		replyToMessageId = *chat.ReplyToMessageID
	} else {
		replyToMessageId = 0
	}

	var recipientId uint64
	if chat.RecipientID != nil {
		recipientId = *chat.RecipientID
	}

	var content string
	if chat.Content != nil {
		content = *chat.Content
	}

	return &prot.ChatMessageResponse{
		Id:               chat.ID,
		CourseId:         chat.CourseID,
		UserId:           chat.UserID,
		RecipientId:      recipientId,
		Content:          content,
		MessageType:      chat.MessageType,
		IsPinned:         chat.IsPinned,
		IsEdited:         chat.IsEdited,
		ReplyToMessageId: replyToMessageId,
		EditedAt:         editedAt,
		CreatedAt:        chat.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        chat.UpdatedAt.Format("2006-01-02 15:04:05"),
		User:             &user,
		Recipient:        recipient,
		ReplyToMessage:   replyTo,
		Files:            files,
		Medias:           medias,
	}
}

func (r *ChatResourceImpl) FormatChats(chats []dto.ChatMessageResponse) []*prot.ChatMessageResponse {
	result := make([]*prot.ChatMessageResponse, 0, len(chats))
	for _, u := range chats {
		if formatted := r.FormatChat(&u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ChatResourceImpl) FormatChatMedia(media *dto.UploadedMediaResponse) *prot.UploadedMediaResponse {
	if media == nil {
		return nil
	}

	return &prot.UploadedMediaResponse{
		Id:            media.ID,
		FileName:      media.FileName,
		FilePath:      media.FilePath,
		FileType:      media.FileType,
		FileSize:      media.FileSize,
		FileExtension: media.FileExtension,
		DiskName:      media.DiskName,
		StaticUrl:     media.StaticURL,
	}
}

func (r *ChatResourceImpl) FormatChatMedias(medias []dto.UploadedMediaResponse) []*prot.UploadedMediaResponse {
	result := make([]*prot.UploadedMediaResponse, 0, len(medias))
	for _, u := range medias {
		if formatted := r.FormatChatMedia(&u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}
