package controllers

import (
	"be-lms/prot"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type FlashcardController struct {
	flashcardService *services.FlashcardService
}

func NewFlashcardController() *FlashcardController {
	return &FlashcardController{
		flashcardService: services.NewFlashcardService(),
	}
}

func (c *FlashcardController) CreateVocabulary(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.CreateVocabularyRequest](ctx, func() *prot.CreateVocabularyRequest {
		return &prot.CreateVocabularyRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	vocabulary, err := c.flashcardService.CreateVocabulary(*req, userID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create vocabulary", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusCreated, "Vocabulary created successfully", vocabulary)
}

func (c *FlashcardController) GetVocabulary(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("vocabularyId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid vocabulary ID")
		return
	}

	vocabulary, err := c.flashcardService.GetVocabularyByID(id)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Vocabulary retrieved successfully", vocabulary)
}

func (c *FlashcardController) GetAllVocabularies(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	search := ctx.Query("search")
	difficulty, _ := strconv.Atoi(ctx.DefaultQuery("difficulty", "0"))
	partOfSpeech := ctx.Query("part_of_speech")

	filters := map[string]interface{}{
		"page":           page,
		"limit":          limit,
		"search":         search,
		"difficulty":     difficulty,
		"part_of_speech": partOfSpeech,
	}

	vocabularies, total, err := c.flashcardService.GetAllVocabularies(filters)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get vocabularies", err.Error())
		return
	}

	response := map[string]interface{}{
		"vocabularies": vocabularies,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	utils.JSONResponse(ctx, http.StatusOK, "Vocabularies retrieved successfully", response)
}

func (c *FlashcardController) GetLessonFlashcards(ctx *gin.Context) {
	lessonID, err := strconv.ParseInt(ctx.Param("lessonId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	// Get student ID from token
	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	var studentIDPtr *int64
	if err == nil && userID != 0 {
		studentIDPtr = &userID
	}

	lessonData, err := c.flashcardService.GetLessonVocabularies(lessonID, studentIDPtr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Lesson flashcards retrieved successfully", lessonData)
}

func (c *FlashcardController) AddVocabulariesToLesson(ctx *gin.Context) {
	lessonID, err := strconv.ParseInt(ctx.Param("lessonId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	req, err, message := utils.GetBody[*prot.AddVocabularyToLessonRequest](ctx, func() *prot.AddVocabularyToLessonRequest {
		return &prot.AddVocabularyToLessonRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	if err := c.flashcardService.AddVocabulariesToLesson(lessonID, *req); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to add vocabularies to lesson", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusCreated, "Vocabularies added to lesson successfully", nil)
}

func (c *FlashcardController) SyncLessonVocabularies(ctx *gin.Context) {
	lessonID, err := strconv.ParseInt(ctx.Param("lessonId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	req, err, message := utils.GetBody[*prot.AddVocabularyToLessonRequest](ctx, func() *prot.AddVocabularyToLessonRequest {
		return &prot.AddVocabularyToLessonRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	if err := c.flashcardService.SyncLessonVocabularies(lessonID, req.VocabularyIds); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to sync lesson vocabularies", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Lesson vocabularies synced successfully", nil)
}

func (c *FlashcardController) StartFlashcardSession(ctx *gin.Context) {
	req, err, messageError := utils.GetBody[*prot.StartFlashcardSessionRequest](ctx, func() *prot.StartFlashcardSessionRequest {
		return &prot.StartFlashcardSessionRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, messageError)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	studentID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	forceNew := ctx.Query("force_new") == "true" || ctx.Query("force_new") == "1"

	session, err := c.flashcardService.StartFlashcardSession(studentID, *req, forceNew)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start session", err.Error())
		return
	}

	// Validate session_id exists in response
	if sessionID, exists := session["session_id"]; !exists || sessionID == nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Session ID not found in response")
		return
	}

	// Extract message from session data
	message := "Flashcard session started successfully"
	if sessionMessage, ok := session["message"]; ok {
		if msg, isString := sessionMessage.(string); isString {
			message = msg
		}
	}

	// Log for debugging
	ctx.Header("X-Session-ID", fmt.Sprintf("%v", session["session_id"]))

	utils.JSONResponse(ctx, http.StatusCreated, message, session)
}

func (c *FlashcardController) RecordFlashcardActivity(ctx *gin.Context) {
	sessionIDStr := ctx.Param("sessionId")

	// Check for undefined sessionId
	if sessionIDStr == "undefined" || sessionIDStr == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Session ID is undefined or missing. Please start a new session first.")
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid session ID: %s", sessionIDStr))
		return
	}


	req, err, message := utils.GetBody[*prot.RecordFlashcardActivityRequest](ctx, func() *prot.RecordFlashcardActivityRequest {
		return &prot.RecordFlashcardActivityRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	// Lấy vocabulary_id từ request body
	if req.VocabularyId == 0 {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid vocabulary ID")
		return
	}

	if err := c.flashcardService.RecordFlashcardActivity(sessionID, req.VocabularyId, *req); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to record activity", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusCreated, "Activity recorded successfully", nil)
}

func (c *FlashcardController) CompleteFlashcardSession(ctx *gin.Context) {
	sessionIDStr := ctx.Param("sessionId")

	// Check for undefined sessionId
	if sessionIDStr == "undefined" || sessionIDStr == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Session ID is undefined or missing. Please start a new session first.")
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid session ID: %s", sessionIDStr))
		return
	}

	req, err, message := utils.GetBody[*prot.CompleteFlashcardSessionRequest](ctx, func() *prot.CompleteFlashcardSessionRequest {
		return &prot.CompleteFlashcardSessionRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	// Get user ID from token
	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	result, err := c.flashcardService.CompleteFlashcardSession(sessionID, userID, *req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to complete session", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Session completed successfully", result)
}

func (c *FlashcardController) GetLessonProgress(ctx *gin.Context) {
	lessonID, err := strconv.ParseInt(ctx.Param("lessonId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	progress, err := c.flashcardService.GetLessonProgress(userID, lessonID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get lesson progress", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Lesson progress retrieved successfully", progress)
}

func (c *FlashcardController) GetActiveSession(ctx *gin.Context) {
	lessonID, err := strconv.ParseInt(ctx.Param("lessonId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lesson ID")
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	session, err := c.flashcardService.GetActiveSession(userID, lessonID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, "No active session found", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Active session retrieved successfully", session)
}

func (c *FlashcardController) ResumeFlashcardSession(ctx *gin.Context) {
	sessionIDStr := ctx.Param("sessionId")

	// Check for undefined sessionId
	if sessionIDStr == "undefined" || sessionIDStr == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Session ID is undefined or missing. Please start a new session first.")
		return
	}

	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid session ID: %s", sessionIDStr))
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	sessionData, err := c.flashcardService.ResumeFlashcardSession(userID, sessionID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to resume session", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Session resumed successfully", sessionData)
}

func (c *FlashcardController) UpdateVocabularyProgress(ctx *gin.Context) {
	// Clean the parameter to remove any whitespace/newline characters
	vocabularyIDStr := strings.TrimSpace(ctx.Param("vocabularyId"))
	vocabularyID, err := strconv.ParseInt(vocabularyIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid vocabulary ID")
		return
	}

	tokenStr := ctx.GetHeader("Token")
	studentID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated", err.Error())
		return
	}

	req, err, message := utils.GetBody[*prot.UpdateVocabularyProgressRequest](ctx, func() *prot.UpdateVocabularyProgressRequest {
		return &prot.UpdateVocabularyProgressRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	if err := c.flashcardService.UpdateVocabularyProgress(studentID, vocabularyID, *req); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update progress", err.Error())
		return
	}

	utils.JSONResponse(ctx, http.StatusOK, "Progress updated successfully", nil)
}
