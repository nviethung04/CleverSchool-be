package controllers

import (
	"be-lms/config"
	"be-lms/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AIGradingController struct {
	service services.AIGradingService
}

func NewAIGradingController(service services.AIGradingService) *AIGradingController {
	return &AIGradingController{
		service: service,
	}
}

// GradeWritingRequest - Request body cho chấm writing
type GradeWritingRequest struct {
	EssayText string `json:"essay_text" binding:"required"` // Đoạn văn cần chấm
	Rubric    string `json:"rubric"`                        // Đề bài/tiêu chí chấm
}

// GradeSpeakingRequest - Request body cho chấm speaking
type GradeSpeakingRequest struct {
	AudioURL string `json:"audio_url" binding:"required"` // URL file audio mp3
	Rubric   string `json:"rubric"`                       // Đề bài/tiêu chí chấm
}

// @Summary Grade Writing Essay with AI
// @Description Giáo viên gửi bài writing lên để AI chấm điểm và đưa ra feedback
// @Tags AI Grading
// @Accept json
// @Produce json
// @Param body body GradeWritingRequest true "Writing essay to grade"
// @Success 200 {object} services.WritingGradeResult
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /api/manage/ai-grading/writing [post]
func (c *AIGradingController) GradeWriting(ctx *gin.Context) {
	var req GradeWritingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	config.Log.Infof("🤖 AI grading request for writing essay (length: %d chars)", len(req.EssayText))

	// Gọi AI service để chấm bài
	result, err := c.service.GradeWritingEssay(req.EssayText, req.Rubric)
	if err != nil {
		config.Log.Errorf("❌ AI grading failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to grade essay",
			"details": err.Error(),
		})
		return
	}

	config.Log.Infof("✅ AI grading completed - Overall score: %.1f", result.OverallScore)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// @Summary Grade Speaking Audio with AI
// @Description Giáo viên gửi file audio speaking lên để AI chấm điểm
// @Tags AI Grading
// @Accept json
// @Produce json
// @Param body body GradeSpeakingRequest true "Speaking audio to grade"
// @Success 200 {object} services.SpeakingGradeResult
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Server error"
// @Router /api/manage/ai-grading/speaking [post]
func (c *AIGradingController) GradeSpeaking(ctx *gin.Context) {
	var req GradeSpeakingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	config.Log.Infof("🤖 AI grading request for speaking audio: %s", req.AudioURL)

	// Gọi AI service để chấm bài speaking
	result, err := c.service.GradeSpeakingAudio(req.AudioURL, req.Rubric)
	if err != nil {
		config.Log.Errorf("❌ AI speaking grading failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to grade speaking",
			"details": err.Error(),
		})
		return
	}

	config.Log.Infof("✅ AI speaking grading completed - Overall score: %.1f", result.OverallScore)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
