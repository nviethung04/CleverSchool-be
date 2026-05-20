package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetExerciseAnswersController struct {
	service services.GetExerciseAnswersService
}

func NewGetExerciseAnswersController(service services.GetExerciseAnswersService) *GetExerciseAnswersController {
	return &GetExerciseAnswersController{service: service}
}

func (c *GetExerciseAnswersController) GetTeacherExerciseAnswers(ctx *gin.Context) {
	exerciseIDStr := ctx.Query("exercise_id")
	userIDStr := ctx.Query("user_id")

	if exerciseIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exercise_id is required"})
		return
	}
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise_id"})
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	req := &prot.GetExerciseAnswersRequest{ // reuse proto
		ExerciseId: exerciseID, // interpreted as exercise_id in service
		UserId:     userID,
	}

	response, err := c.service.GetExerciseAnswers(ctx, req)
	utils.Respond(ctx, response, err, "")
}

func (c *GetExerciseAnswersController) GetStudentExerciseAnswers(ctx *gin.Context) {
	exerciseIDStr := ctx.Query("exercise_id")

	if exerciseIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "exercise_id is required"})
		return
	}

	exerciseID, err := strconv.ParseInt(exerciseIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise_id"})
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	req := &prot.GetExerciseAnswersRequest{
		ExerciseId: exerciseID,
		UserId:     userID,
	}

	response, err := c.service.GetExerciseAnswers(ctx, req)
	utils.Respond(ctx, response, err, "")
}
