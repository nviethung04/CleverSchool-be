package controllers

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatMessageController struct {
	service      services.ChatMessageService
	mediaRepo    repositories.MediaRepository
	mediaService services.MediaService
}

func NewChatMessageController(service services.ChatMessageService) *ChatMessageController {
	mediaRepo := repositories.NewMediaRepository()
	mediaService := services.NewMediaService(mediaRepo)
	return &ChatMessageController{
		service:      service,
		mediaRepo:    mediaRepo,
		mediaService: mediaService,
	}
}

func (c *ChatMessageController) GetMessages(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	page := 1
	limit := 20

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := c.service.GetMessages(ctx, uint64(courseId), page, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	chatMessages := resource.FormatChats(response.Messages)

	utils.Respond(ctx, &prot.ChatMessagesListResponse{
		Messages: chatMessages,
		Pagination: &prot.PaginationResponse{
			Page:    int32(response.Pagination.Page),
			Limit:   int32(response.Pagination.Limit),
			Total:   int32(response.Pagination.Total),
			HasMore: response.Pagination.HasMore,
		},
	}, nil, "")
}

func (c *ChatMessageController) DeleteMessage(ctx *gin.Context) {
	messageId, err := strconv.Atoi(ctx.Param("messageId"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userId := utils.GetCurrentUserId(ctx)

	err = c.service.DeleteMessage(ctx, uint64(messageId), uint64(userId))
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.ChatActionResponse{
		Success: true,
		Message: "Message deleted successfully",
	}, nil, "")
}

func (c *ChatMessageController) TogglePinMessage(ctx *gin.Context) {
	messageId, err := strconv.Atoi(ctx.Param("messageId"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userId := utils.GetCurrentUserId(ctx)

	err = c.service.TogglePinMessage(ctx, uint64(messageId), uint64(userId))
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.ChatActionResponse{
		Success: true,
		Message: "Message pin status updated successfully",
	}, nil, "")
}

func (c *ChatMessageController) GetPinnedMessages(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	response, err := c.service.GetPinnedMessages(ctx, uint64(courseId))
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChats(response)

	utils.Respond(ctx, &prot.PinnedMessagesResponse{
		PinnedMessages: responseFormatted,
		Count:          int32(len(responseFormatted)),
	}, nil, "")
}

func (c *ChatMessageController) GetMessageCount(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	count, err := c.service.CountMessages(ctx, uint64(courseId))
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.MessageCountResponse{
		TotalMessages: int32(count),
	}, nil, "")
}

func (c *ChatMessageController) GetRecentMessages(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	limit := 10
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	response, err := c.service.GetRecentMessages(ctx, uint64(courseId), limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChats(response)

	utils.Respond(ctx, &prot.RecentMessagesResponse{
		RecentMessages: responseFormatted,
		Count:          int32(len(responseFormatted)),
	}, nil, "")
}

func (c *ChatMessageController) SendMessageWithMedias(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userId := utils.GetCurrentUserId(ctx)

	// Check if request has files (multipart form) or JSON
	contentType := ctx.GetHeader("Content-Type")
	isMultipart := strings.Contains(contentType, "multipart/form-data")

	var req dto.SendChatMessageRequest
	var uploadedMediaIDs []int64

	if isMultipart {
		// Handle multipart form with files
		req.Content = ctx.PostForm("content")
		
		// Get files from form
		form, err := ctx.MultipartForm()
		if err != nil {
			config.Log.Errorf("Failed to parse multipart form: %v", err)
			utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
			return
		}

		files := form.File["files"]
		if len(files) == 0 {
			files = form.File["file"] // Fallback to "file" field name
		}

		// Upload files and get media IDs
		if len(files) > 0 {
			uploadedMediaIDs, err = c.uploadChatFiles(ctx, files)
			if err != nil {
				config.Log.Errorf("Failed to upload chat files: %v", err)
				utils.Respond(ctx, nil, fmt.Errorf("failed to upload files: %w", err), "")
				return
			}
		}

		// Get media_ids from form if provided (for already uploaded files)
		if mediaIDsStr := ctx.PostForm("media_ids"); mediaIDsStr != "" {
			// Parse comma-separated media IDs
			parts := strings.Split(mediaIDsStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					if id, err := strconv.ParseInt(part, 10, 64); err == nil {
						uploadedMediaIDs = append(uploadedMediaIDs, id)
					}
				}
			}
		}

		req.MediaIDs = uploadedMediaIDs
	} else {
		// Handle JSON request
		if err := ctx.ShouldBindJSON(&req); err != nil {
			config.Log.Errorf("Failed to bind JSON in SendMessageWithMedias: %v", err)
			utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
			return
		}
	}

	// Validate request
	if len(req.Content) == 0 && len(req.MediaIDs) == 0 {
		config.Log.Warnf("Empty message request from user %d in course %d", userId, courseId)
		utils.Respond(ctx, nil, errors.New("message must have either content or media attachments"), "")
		return
	}

	result, err := c.service.SendMessageWithMedias(ctx, uint64(courseId), uint64(userId), req)
	if err != nil {
		config.Log.Errorf("Error in SendMessageWithMedias: %v (courseId: %d, userId: %d, mediaIDs: %v)", err, courseId, userId, req.MediaIDs)
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChat(result)

	utils.Respond(ctx, responseFormatted, nil, "")
}

func (c *ChatMessageController) SendToRecipientMessage(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	recipientId, err := strconv.Atoi(ctx.Param("recipientId"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	senderId := utils.GetCurrentUserId(ctx)

	var req dto.SendChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	result, err := c.service.SendToRecipient(ctx, uint64(courseId), uint64(senderId), uint64(recipientId), req)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChat(result)

	utils.Respond(ctx, responseFormatted, nil, "")
}

func (c *ChatMessageController) GetFromRecipientMessage(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	recipientId, err := strconv.Atoi(ctx.Param("recipientId"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	senderId := utils.GetCurrentUserId(ctx)

	page := 1
	limit := 20

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := c.service.GetMessagesWithRecipient(ctx, uint64(courseId), uint64(senderId), uint64(recipientId), page, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	chatMessages := resource.FormatChats(response.Messages)

	utils.Respond(ctx, &prot.ChatMessagesListResponse{
		Messages: chatMessages,
		Pagination: &prot.PaginationResponse{
			Page:    int32(response.Pagination.Page),
			Limit:   int32(response.Pagination.Limit),
			Total:   int32(response.Pagination.Total),
			HasMore: response.Pagination.HasMore,
		},
	}, nil, "")
}

func (c *ChatMessageController) GetRecentSenders(ctx *gin.Context) {
	courseId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	currentUserId := utils.GetCurrentUserId(ctx)

	limit := 20
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	response, err := c.service.GetRecentSenders(ctx, uint64(courseId), uint64(currentUserId), limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, map[string]interface{}{
		"senders": response,
		"count":   len(response),
	}, nil, "")
}

// uploadChatFiles uploads files and returns media IDs
func (c *ChatMessageController) uploadChatFiles(ctx *gin.Context, files []*multipart.FileHeader) ([]int64, error) {
	var mediaIDs []int64
	path := "chat" // Path for chat files

	// Find or create folder for chat files
	folder, err := c.mediaRepo.FindByPath(path, "public")
	if err != nil || folder == nil {
		// Create folder if not exists
		folderReq := prot.Media{
			FilePath: path,
			ParentId: 0,
		}
		if err := c.mediaService.UploadFolder(ctx, folderReq); err != nil {
			config.Log.Warnf("Failed to create chat folder, continuing anyway: %v", err)
		}
		// Try to find again
		folder, _ = c.mediaRepo.FindByPath(path, "public")
	}

	folderID := int64(0)
	if folder != nil {
		folderID = folder.ID
	}

	// Upload each file
	for _, file := range files {
		// Open the file
		src, err := file.Open()
		if err != nil {
			config.Log.Errorf("Failed to open file %s: %v", file.Filename, err)
			continue
		}
		
		// Read file content
		fileData := make([]byte, file.Size)
		if _, err := src.Read(fileData); err != nil {
			src.Close()
			config.Log.Errorf("Failed to read file %s: %v", file.Filename, err)
			continue
		}
		src.Close()
		
		// Use StoreFileFromBytes to upload
		uploader := utils.NewUploaderWithStorage(path, "file", models.Storage)
		info, err := uploader.StoreFileFromBytes(fileData, file.Filename)
		if err != nil {
			config.Log.Errorf("Failed to upload file %s: %v", file.Filename, err)
			continue
		}

		// Parse file size
		var fileSizeInt64 int64
		if info.FileSize != "" {
			if size, err := strconv.ParseInt(info.FileSize, 10, 64); err == nil {
				fileSizeInt64 = size
			} else {
				// Try to parse human-readable size (e.g., "1.5 MB")
				parts := strings.Fields(info.FileSize)
				if len(parts) == 2 {
					if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
						unit := strings.ToUpper(parts[1])
						switch unit {
						case "KB":
							fileSizeInt64 = int64(val * 1024)
						case "MB":
							fileSizeInt64 = int64(val * 1024 * 1024)
						case "GB":
							fileSizeInt64 = int64(val * 1024 * 1024 * 1024)
						default:
							fileSizeInt64 = int64(val)
						}
					}
				}
			}
		}

		// Strip domain from URL
		stripped := utils.StripDomain(info.Url, models.Storage)
		fileUrl := &stripped

		// Save media to database
		media := &models.Media{
			FolderID:      folderID,
			FileName:      info.FileName,
			FilePath:      info.FilePath,
			FileType:      utils.StringPtr(info.FileMimeType),
			FileSize:      utils.Int64Ptr(fileSizeInt64),
			FileExtension: utils.StringPtr(info.FileExt),
			DiskName:      utils.StringPtr(models.Storage),
			StaticURL:     fileUrl,
			Type:          "file",
		}

		if err := c.mediaRepo.Save(media); err != nil {
			config.Log.Errorf("Failed to save media for file %s: %v", file.Filename, err)
			continue // Skip this file and continue with others
		}

		mediaIDs = append(mediaIDs, media.ID)
		config.Log.Infof("Successfully uploaded chat file: %s (media ID: %d)", file.Filename, media.ID)
	}

	if len(mediaIDs) == 0 && len(files) > 0 {
		return nil, fmt.Errorf("failed to upload any files")
	}

	return mediaIDs, nil
}
