package controllers

import (
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ChatMessageController struct {
	service services.ChatMessageService
}

func NewChatMessageController(service services.ChatMessageService) *ChatMessageController {
	return &ChatMessageController{
		service: service,
	}
}

// GetMessages handles GET /courses/:id/chat/messages
func (c *ChatMessageController) GetMessages(ctx *gin.Context) {
	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
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

	response, err := c.service.GetMessages(ctx.Request.Context(), courseId, page, limit)
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

// DeleteMessage handles DELETE /courses/:id/chat/messages/:messageId
func (c *ChatMessageController) DeleteMessage(ctx *gin.Context) {
	messageIdStr := ctx.Param("messageId")
	messageId, err := strconv.ParseUint(messageIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userIdInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized")
		return
	}

	var userId uint64
	switch v := userIdInterface.(type) {
	case int:
		userId = uint64(v)
	case uint64:
		userId = v
	case int64:
		userId = uint64(v)
	default:
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	err = c.service.DeleteMessage(ctx.Request.Context(), messageId, userId)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.ChatActionResponse{
		Success: true,
		Message: "Message deleted successfully",
	}, nil, "")
}

// TogglePinMessage handles POST /courses/:id/chat/messages/:messageId/pin
func (c *ChatMessageController) TogglePinMessage(ctx *gin.Context) {
	messageIdStr := ctx.Param("messageId")
	messageId, err := strconv.ParseUint(messageIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	userIdInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized")
		return
	}

	var userId uint64
	switch v := userIdInterface.(type) {
	case int:
		userId = uint64(v)
	case uint64:
		userId = v
	case int64:
		userId = uint64(v)
	default:
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	err = c.service.TogglePinMessage(ctx.Request.Context(), messageId, userId)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.ChatActionResponse{
		Success: true,
		Message: "Message pin status updated successfully",
	}, nil, "")
}

// GetPinnedMessages handles GET /courses/:id/chat/messages/pinned
func (c *ChatMessageController) GetPinnedMessages(ctx *gin.Context) {
	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	response, err := c.service.GetPinnedMessages(ctx.Request.Context(), courseId)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChats(response)

	utils.Respond(ctx, &prot.PinnedMessagesResponse{
		PinnedMessages: responseFormatted,
		Count: int32(len(responseFormatted)),
	}, nil, "")
}

// GetMessageCount handles GET /courses/:id/chat/messages/count
func (c *ChatMessageController) GetMessageCount(ctx *gin.Context) {
	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	count, err := c.service.CountMessages(ctx.Request.Context(), courseId)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	utils.Respond(ctx, &prot.MessageCountResponse{
		TotalMessages: int32(count),
	}, nil, "")
}

// GetRecentMessages handles GET /courses/:id/chat/messages/recent
func (c *ChatMessageController) GetRecentMessages(ctx *gin.Context) {
	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
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

	response, err := c.service.GetRecentMessages(ctx.Request.Context(), courseId, limit)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChats(response)

	utils.Respond(ctx, &prot.RecentMessagesResponse{
		RecentMessages: responseFormatted,
		Count: int32(len(responseFormatted)),
	}, nil, "")
}

// handleTextMessage xử lý tin nhắn text thông thường
func (c *ChatMessageController) handleTextMessage(ctx *gin.Context, courseId, userId uint64) {
	var req dto.SendChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Lỗi bind JSON: %v\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	fmt.Printf("DEBUG: Request data - Content: %s\n", req.Content)

	// Validate content for text messages
	if len(strings.TrimSpace(req.Content)) == 0 {
		utils.Respond(ctx, nil, errors.New("Text message content cannot be empty"), "Text message content cannot be empty")
		return
	}

	result, err := c.service.SendMessage(ctx.Request.Context(), courseId, userId, req)
	if err != nil {
		fmt.Printf("DEBUG: Lỗi service.SendMessage: %v\n", err)
		utils.Respond(ctx, nil, err, "")
		return
	}

	fmt.Printf("DEBUG: Gửi tin nhắn thành công với ID: %d\n", result.ID)

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChat(result)

	utils.Respond(ctx, responseFormatted, nil, "")
}

// handleFileUpload xử lý upload file kèm tin nhắn
func (c *ChatMessageController) handleFileUpload(ctx *gin.Context, courseId, userId uint64) {
	fmt.Printf("=== DEBUG: handleFileUpload START ===\n")
	fmt.Printf("DEBUG: courseId: %d, userId: %d\n", courseId, userId)

	var req dto.SendChatMessageWithFilesRequest
	if err := ctx.ShouldBind(&req); err != nil {
		fmt.Printf("DEBUG: Lỗi bind form: %v\n", err)
		fmt.Printf("DEBUG: Error type: %T\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	fmt.Printf("DEBUG: File upload request - Content: '%s'\n", req.Content)

	// Get uploaded files
	form, err := ctx.MultipartForm()
	if err != nil {
		fmt.Printf("DEBUG: Lỗi get multipart form: %v\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	fmt.Printf("DEBUG: Multipart form keys: %+v\n", func() []string {
		keys := make([]string, 0, len(form.File))
		for k := range form.File {
			keys = append(keys, k)
		}
		return keys
	}())

	files := form.File["files"]
	fmt.Printf("DEBUG: Số file với key 'files': %d\n", len(files))

	if len(files) == 0 {
		fmt.Printf("DEBUG: Không tìm thấy files với key 'files', checking all keys\n")
		for key, fileHeaders := range form.File {
			fmt.Printf("DEBUG: Key '%s' có %d files\n", key, len(fileHeaders))
		}
		utils.Respond(ctx, nil, errors.New("No files uploaded"), "No files uploaded")
		return
	}

	fmt.Printf("DEBUG: Số file upload: %d\n", len(files))

	// Validate files
	for i, file := range files {
		fmt.Printf("DEBUG: File %d: name='%s', size=%d\n", i, file.Filename, file.Size)
		if err := c.validateUploadedFile(file); err != nil {
			fmt.Printf("DEBUG: File validation failed: %v\n", err)
			utils.Respond(ctx, nil, err, "")
			return
		}
	}

	// Convert to SendChatMessageRequest
	sendReq := dto.SendChatMessageRequest{
		Content: req.Content,
	}

	fmt.Printf("DEBUG: Calling service.SendMessageWithFiles\n")
	// Call service with files
	result, err := c.service.SendMessageWithFiles(ctx.Request.Context(), courseId, userId, sendReq, files)
	if err != nil {
		fmt.Printf("DEBUG: Lỗi service.SendMessageWithFiles: %v\n", err)
		utils.Respond(ctx, nil, err, "")
		return
	}

	fmt.Printf("DEBUG: Gửi tin nhắn với file thành công với ID: %d\n", result.ID)


	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChat(result)

	utils.Respond(ctx, responseFormatted, nil, "")

	fmt.Printf("=== DEBUG: handleFileUpload END ===\n")
}

// validateUploadedFile kiểm tra file upload
func (c *ChatMessageController) validateUploadedFile(file *multipart.FileHeader) error {
	// Check file size (max 50MB)
	maxSize := int64(50 * 1024 * 1024) // 50MB
	if file.Size > maxSize {
		return errors.New("File size too large (max 50MB)")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		// Images
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true,
		// Videos
		".mp4": true, ".avi": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".mkv": true,
		// Audio
		".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".m4a": true,
		// Documents
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".txt": true, ".rtf": true, ".zip": true, ".rar": true, ".7z": true,
	}

	if !allowedExts[ext] {
		return fmt.Errorf("File type %s not allowed", ext)
	}

	return nil
}

// SendMessageWithMedias handles POST /courses/:id/chat/messages/send-message-with-medias
func (c *ChatMessageController) SendMessageWithMedias(ctx *gin.Context) {
	fmt.Printf("=== DEBUG: SendMessageWithMedias START ===\n")

	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Get user ID from context
	userIdInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized")
		return
	}

	var userId uint64
	switch v := userIdInterface.(type) {
	case int:
		userId = uint64(v)
	case uint64:
		userId = v
	case int64:
		userId = uint64(v)
	default:
		fmt.Printf("DEBUG: userID có type không hỗ trợ: %T, value: %+v\n", userIdInterface, userIdInterface)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Parse request body
	var req dto.SendChatMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Lỗi bind JSON: %v\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	fmt.Printf("DEBUG: SendMessageWithMedias - courseId: %d, userId: %d, content: %s, media_ids: %v\n",
		courseId, userId, req.Content, req.MediaIDs)

	// Call service
	result, err := c.service.SendMessageWithMedias(ctx.Request.Context(), courseId, userId, req)
	if err != nil {
		fmt.Printf("DEBUG: Lỗi service.SendMessageWithMedias: %v\n", err)
		utils.Respond(ctx, nil, err, "")
		return
	}

	fmt.Printf("DEBUG: SendMessageWithMedias thành công với ID: %d\n", result.ID)

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChat(result)

	utils.Respond(ctx, responseFormatted, nil, "")

	fmt.Printf("=== DEBUG: SendMessageWithMedias END ===\n")
}

// UploadFilesToMedias handles POST /courses/:id/chat/messages/upload-files-to-medias
func (c *ChatMessageController) UploadFilesToMedias(ctx *gin.Context) {
	fmt.Printf("=== DEBUG: UploadFilesToMedias START ===\n")

	courseIdStr := ctx.Param("id")
	courseId, err := strconv.ParseUint(courseIdStr, 10, 64)
	if err != nil {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Get user ID from context
	userIdInterface, exists := ctx.Get("userID")
	if !exists {
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.unauthorized")), "messages.unauthorized")
		return
	}

	var userId uint64
	switch v := userIdInterface.(type) {
	case int:
		userId = uint64(v)
	case uint64:
		userId = v
	case int64:
		userId = uint64(v)
	default:
		fmt.Printf("DEBUG: userID có type không hỗ trợ: %T, value: %+v\n", userIdInterface, userIdInterface)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		fmt.Printf("DEBUG: Lỗi parse multipart form: %v\n", err)
		utils.Respond(ctx, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		utils.Respond(ctx, nil, errors.New("No files uploaded"), "No files uploaded")
		return
	}

	fmt.Printf("DEBUG: Số files upload: %d\n", len(files))

	// Validate files
	for i, file := range files {
		fmt.Printf("DEBUG: Validating file %d: %s (size: %d)\n", i, file.Filename, file.Size)
		if err := c.validateUploadedFile(file); err != nil {
			fmt.Printf("DEBUG: File validation failed: %v\n", err)
			utils.Respond(ctx, nil, fmt.Errorf("File %s validation failed: %v", file.Filename, err), "")
			return
		}
	}

	// Call service to upload files to medias
	result, err := c.service.UploadFilesToMedias(ctx.Request.Context(), courseId, userId, files)
	if err != nil {
		fmt.Printf("DEBUG: Lỗi service.UploadFilesToMedias: %v\n", err)
		utils.Respond(ctx, nil, err, "")
		return
	}

	fmt.Printf("DEBUG: UploadFilesToMedias thành công - uploaded %d files\n", len(result))

	resource := resources.NewChatResource()
	responseFormatted := resource.FormatChatMedias(result)

	utils.Respond(ctx, responseFormatted, nil, "")

	fmt.Printf("=== DEBUG: UploadFilesToMedias END ===\n")
}
