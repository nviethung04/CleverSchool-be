package controllers

import (
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"be-Clever School/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

type SaveScoreController struct {
	serviceMC                     services.SaveScoreMultipleChoiceService
	serviceFB                     services.SaveScoreFillInBlankService
	serviceP                      services.SaveScorePositionService
	serviceM                      services.SaveScoreMatchingService
	serviceL                      services.SaveScoreLabelingService
	serviceG                      services.SaveScoreGroupService
	serviceB                      services.SaveScoreBulkService
	serviceMS                     services.ManualScoringService
	saveScoreManualScoringService services.SaveScoreManualScoringService
	homeworkUserService           services.HomeworkUserService
	homeworkSkipQuestionService   services.HomeworkSkipQuestionService
	homeworkCalculationService    services.HomeworkCalculationService
	userStarExpService            services.UserStarExpService
}

func NewSaveScoreController(
	serviceMC services.SaveScoreMultipleChoiceService,
	serviceFB services.SaveScoreFillInBlankService,
	serviceP services.SaveScorePositionService,
	serviceM services.SaveScoreMatchingService,
	serviceL services.SaveScoreLabelingService,
	serviceG services.SaveScoreGroupService,
	serviceB services.SaveScoreBulkService,
	serviceMS services.ManualScoringService,
	saveScoreManualScoringService services.SaveScoreManualScoringService,
	homeworkUserService services.HomeworkUserService,
	homeworkSkipQuestionService services.HomeworkSkipQuestionService,
	homeworkCalculationService services.HomeworkCalculationService,
	userStarExpService services.UserStarExpService,
) *SaveScoreController {
	return &SaveScoreController{
		serviceMC:                     serviceMC,
		serviceFB:                     serviceFB,
		serviceP:                      serviceP,
		serviceM:                      serviceM,
		serviceL:                      serviceL,
		serviceG:                      serviceG,
		serviceB:                      serviceB,
		serviceMS:                     serviceMS,
		saveScoreManualScoringService: saveScoreManualScoringService,
		homeworkUserService:           homeworkUserService,
		homeworkSkipQuestionService:   homeworkSkipQuestionService,
		homeworkCalculationService:    homeworkCalculationService,
		userStarExpService:            userStarExpService,
	}
}

// Helper function để check và update did_it_again trong homework_user_skip_questions
// Chỉ gọi khi câu hỏi được làm đúng
func (c *SaveScoreController) checkAndUpdateSkipQuestionIfCorrect(homeworkID, userID, questionID int64, isCorrect bool) {
	if homeworkID > 0 && userID > 0 && questionID > 0 && isCorrect {
		// Chỉ gọi service để update did_it_again khi câu hỏi được làm đúng
		c.homeworkSkipQuestionService.UpdateDidItAgainIfSkipped(homeworkID, userID, questionID)
	}
}

// Helper để lấy question_id từ AnswerRequest (hỗ trợ cả proto và JSON format)
func getQuestionIdFromAnswer(answer *prot.AnswerRequest) int64 {
	// Kiểm tra proto format (oneof) trước
	if answer.GetManualRequest() != nil {
		return answer.GetManualRequest().QuestionId
	}
	if answer.GetGroupRequest() != nil {
		return answer.GetGroupRequest().QuestionId
	}
	if answer.GetLabelingRequest() != nil {
		return answer.GetLabelingRequest().QuestionId
	}
	if answer.GetMatchingRequest() != nil {
		return answer.GetMatchingRequest().QuestionId
	}
	if answer.GetPositionRequest() != nil {
		return answer.GetPositionRequest().QuestionId
	}
	if answer.GetFillInBlankRequest() != nil {
		return answer.GetFillInBlankRequest().QuestionId
	}
	if answer.GetMultipleChoiceRequest() != nil {
		return answer.GetMultipleChoiceRequest().QuestionId
	}
	// Nếu không có oneof, lấy từ JSON format (sử dụng protojson để parse)
	// Note: Proto struct có thể chưa có field QuestionId, cần dùng protojson
	jsonBytes, err := protojson.Marshal(answer)
	if err != nil {
		return 0
	}
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		return 0
	}
	if qid, ok := jsonMap["question_id"].(float64); ok {
		return int64(qid)
	}
	return 0
}

func (c *SaveScoreController) SaveScoreMultipleChoice(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreMultipleChoiceRequest](ctx, func() *prot.SaveScoreMultipleChoiceRequest {
		return &prot.SaveScoreMultipleChoiceRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponse
	if req.ExamId != 0 {
		data, err = c.serviceMC.SaveScoreMultipleChoiceExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceMC.SaveScoreMultipleChoiceHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsCorrect)
		}
	} else if req.LevelTestId != 0 {
		data, err = c.serviceMC.SaveScoreMultipleChoiceLevelTest(req, userID)
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScoreFillInBlank(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreFillInBlankRequest](ctx, func() *prot.SaveScoreFillInBlankRequest {
		return &prot.SaveScoreFillInBlankRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponseFillInBlank
	if req.ExamId != 0 {
		data, err = c.serviceFB.SaveScoreFillInBlankExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceFB.SaveScoreFillInBlankHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsAllCorrect)
		}
	} else if req.LevelTestId != 0 {
		//data, err = c.serviceFB.SaveScoreFillInBlankLevelTest(&req, userID)
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScorePosition(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScorePositionRequest](ctx, func() *prot.SaveScorePositionRequest {
		return &prot.SaveScorePositionRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponsePosition
	if req.ExamId != 0 {
		data, err = c.serviceP.SaveScorePositionExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceP.SaveScorePositionHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsAllCorrect)
		}
	} else if req.LevelTestId != 0 {
		//data, err = c.serviceP.SaveScorePositionLevelTest(&req, userID)
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScoreMatching(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreMatchingRequest](ctx, func() *prot.SaveScoreMatchingRequest {
		return &prot.SaveScoreMatchingRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponseMatching
	if req.ExamId != 0 {
		data, err = c.serviceM.SaveScoreMatchingExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceM.SaveScoreMatchingHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsAllCorrect)
		}
	} else if req.LevelTestId != 0 {
		data, err = c.serviceM.SaveScoreMatchingLevelTest(req, userID)
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScoreLabeling(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreLabelingRequest](ctx, func() *prot.SaveScoreLabelingRequest {
		return &prot.SaveScoreLabelingRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponseLabeling
	if req.ExamId != 0 {
		data, err = c.serviceL.SaveScoreLabelingExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceL.SaveScoreLabelingHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsAllCorrect)
		}
	} else if req.LevelTestId != 0 {
		//data, err = c.serviceL.SaveScoreLabelingLevelTest(&req, userID)
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScoreGroup(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreGroupRequest](ctx, func() *prot.SaveScoreGroupRequest {
		return &prot.SaveScoreGroupRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	var data *prot.SaveScoreResponseGroup
	if req.ExamId != 0 {
		data, err = c.serviceG.SaveScoreGroupExam(req, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceG.SaveScoreGroupHomework(req, userID)
		// Check và update did_it_again chỉ khi câu hỏi được làm đúng
		if err == nil && data != nil {
			c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, data.IsAllCorrect)
		}
	} else {
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	utils.Respond(ctx, data, err, "")
}

func (c *SaveScoreController) SaveScoreBulk(ctx *gin.Context) {
	// Debug: In ra raw body để xem frontend gửi gì
	body, _ := ctx.GetRawData()
	fmt.Printf("🔍 Raw Body: %s\n", string(body))
	fmt.Printf("🔍 Content-Type: %s\n", ctx.GetHeader("Content-Type"))

	// Reset body để GetSaveScoreBulkBody có thể đọc
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Sử dụng helper function riêng cho SaveScoreBulk
	req, err, message := utils.GetSaveScoreBulkBody(ctx)
	if err != nil {
		fmt.Printf("🔍 GetSaveScoreBulkBody Error: %v, Message: %s\n", err, message)
		utils.Respond(ctx, nil, err, message)
		return
	}

	// Debug log
	fmt.Printf("🔍 SaveScoreBulk: ExamId=%d, ExerciseId=%d, HomeworkId=%d, ContestRoundId=%d, LevelTestId=%d, ListAnswers=%v\n",
		req.ExamId, req.ExerciseId, req.HomeworkId, req.ContestRoundId, req.LevelTestId, len(req.ListAnswers))

	// Lấy JSON maps gốc từ context
	var jsonMaps []map[string]interface{}
	if mapsInterface, exists := ctx.Get("save_score_bulk_answer_json_maps"); exists {
		if maps, ok := mapsInterface.([]map[string]interface{}); ok {
			jsonMaps = maps
		}
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	if req.UserId > 0 {
		userID = req.UserId
	}

	var data *prot.SaveScoreBulkResponse
	if req.ExamId != 0 {
		data, err = c.serviceB.SaveScoreBulkExam(req, jsonMaps, userID)
	} else if req.ExerciseId != 0 {
		data, err = c.serviceB.SaveScoreBulkExercise(req, jsonMaps, userID)
	} else if req.HomeworkId != 0 {
		data, err = c.serviceB.SaveScoreBulkHomework(req, jsonMaps, userID)
		// Check và update did_it_again cho từng question chỉ khi làm đúng
		if err == nil && data != nil {
			// Tạo map để check kết quả của từng câu hỏi
			questionResults := make(map[int64]bool)

			// Check kết quả từ các response
			for _, mc := range data.MultipleChoice {
				questionResults[mc.QuestionId] = mc.IsCorrect
			}
			for _, fib := range data.FillInBlank {
				questionResults[fib.QuestionId] = fib.IsAllCorrect
			}
			for _, pos := range data.Ordering {
				questionResults[pos.QuestionId] = pos.IsAllCorrect
			}
			for _, pos := range data.Dragdrop {
				questionResults[pos.QuestionId] = pos.IsAllCorrect
			}
			for _, match := range data.Matching {
				questionResults[match.QuestionId] = match.IsAllCorrect
			}
			for _, label := range data.Labeling {
				questionResults[label.QuestionId] = label.IsAllCorrect
			}
			for _, group := range data.Category {
				questionResults[group.QuestionId] = group.IsAllCorrect
			}

			// Check và update cho từng câu hỏi
			for _, answer := range req.ListAnswers {
				questionId := getQuestionIdFromAnswer(answer)
				if isCorrect, exists := questionResults[questionId]; exists {
					c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, questionId, isCorrect)
				}
			}
		}
	} else if req.ContestRoundId != 0 {
		data, err = c.serviceB.SaveScoreBulkContestRound(req, jsonMaps, userID)
	} else {
		// LevelTest and other types not yet implemented
		utils.Respond(ctx, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	if err != nil {
		fmt.Printf("🔍 Error in SaveScoreBulk: %v\n", err)
	}

	utils.Respond(ctx, data, err, "")
}

// SubmitHomework: tính score/ratio và lưu vào homework_users
func (c *SaveScoreController) SubmitHomework(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SubmitHomeworkRequest](ctx, func() *prot.SubmitHomeworkRequest {
		return &prot.SubmitHomeworkRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	homeworkRepo := repositories.NewHomeworkRepository()
	homework, err := homeworkRepo.GetByID(req.HomeworkId, 0, ctx)
	if err != nil {
		utils.Respond(ctx, nil, err, "Không tìm thấy homework")
		return
	}

	score := 0.0
	ratio := 0.0
	star := 0.0
	exp := 0.0
	hasManualScoring := false
	isCompleted := false

	if homework.QuestionForm == models.QuestionFormQuestionType {
		// Kiểm tra bộ 3 (homework_id, user_id, lesson_id) từ homework_users
		homeworkUserRepo := repositories.NewHomeworkUserRepository()
		oldHomeworkUser, err := homeworkUserRepo.GetByHomeworkUserAndLesson(req.HomeworkId, userID, req.LessonId)
		if err != nil {
			utils.Respond(ctx, nil, err, "Không tìm thấy homework user với lesson_id tương ứng")
			return
		}

		// Lấy giá trị cũ trước khi tính toán mới
		var oldExp float64
		var oldStar int64
		if oldHomeworkUser != nil {
			oldExp = oldHomeworkUser.Exp
			// Lấy star cũ từ homework_users.star
			oldStar = int64(oldHomeworkUser.Star)
		}

		// Tính toán lại exp, star, ratio_score qua service chuyên dụng
		calculationResult, err := c.homeworkCalculationService.CalculateHomeworkMetrics(req.HomeworkId, userID, req.LessonId)
		if err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		// Lưu các giá trị đã tính toán vào database
		score = calculationResult.Score
		if err := c.homeworkUserService.SaveHomeworkScoreService(req.HomeworkId, userID, score); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		ratio = calculationResult.Ratio
		if err := c.homeworkUserService.SaveHomeworkRatioService(req.HomeworkId, userID, ratio); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		exp = calculationResult.Exp
		if err := c.homeworkUserService.SaveHomeworkExpService(req.HomeworkId, userID, exp); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		star = float64(calculationResult.Star)
		if err := c.homeworkUserService.SaveHomeworkStarService(req.HomeworkId, userID, calculationResult.Star); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		isCompleted = calculationResult.IsCompleted

		// Tính change_star và change_exp
		// Nếu chưa có record cũ: change = giá trị mới
		// Nếu đã có record cũ: change = giá trị mới - giá trị cũ (có thể âm nếu mới < cũ)
		var changeStar int64
		var changeExp float64
		if oldHomeworkUser == nil {
			// Chưa có record cũ, change = giá trị mới
			changeStar = calculationResult.Star
			changeExp = exp
		} else {
			// Đã có record cũ, change = giá trị mới - giá trị cũ
			changeStar = calculationResult.Star - oldStar
			changeExp = exp - oldExp
		}

		// Lưu tổng star và exp vào bảng user_star_exp
		note := fmt.Sprintf("homework %d", req.HomeworkId)
		if err := c.userStarExpService.UpdateStarAndExp(userID, changeStar, changeExp, note); err != nil {
			// Log lỗi nhưng không dừng flow
			fmt.Printf("Lỗi khi lưu star và exp: %v\n", err)
		}

		// Kiểm tra xem homework có câu hỏi type speaking hoặc writing không từ cloned_questions
		hasManualScoring, err = c.homeworkUserService.CheckHomeworkHasManualScoringFromClonedQuestions(req.HomeworkId)
		if err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		// Lưu has_manual_scoring vào homework_users
		if err := c.homeworkUserService.SaveHomeworkHasManualScoringService(req.HomeworkId, userID, hasManualScoring); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}

		// Cập nhật status_scoring sau khi tính lại điểm
		if err := c.homeworkUserService.UpdateHomeworkStatusScoringService(req.HomeworkId, userID); err != nil {
			utils.Respond(ctx, nil, err, "")
			return
		}
	} else {
		if req.UserId > 0 {
			userID = req.UserId
		}

		hasManualScoring = true

		mediaRepo := repositories.NewMediaRepository()

		fileInfos := make([]models.MediaDetail, 0)

		for _, file := range req.Files {
			fileUrl := utils.StripDomain(file.Url, models.Storage)
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
			fileInfos = append(fileInfos, models.MediaDetail{
				Path: utils.StripDomain(fileInfo.Path, models.Storage),
				Id:   fileInfo.Id,
				Type: file.Type,
				Disk: fileInfo.Disk,
			})
		}

		homeworkUserRepo := repositories.NewHomeworkUserRepository()
		homeworkUserRepo.UpdateOrCreate(&models.HomeworkUser{
			HomeworkID:       req.HomeworkId,
			UserID:           userID,
			LessonID:         req.LessonId,
			Score:            score,
			Ratio:            ratio,
			FileInfos:        fileInfos,
			HasManualScoring: hasManualScoring,
		})
	}

	files := make([]*prot.SubmitHomeworkFile, 0)

	for _, file := range req.Files {
		files = append(files, &prot.SubmitHomeworkFile{
			Type: file.Type,
			Url:  utils.StaticURL(file.Url, models.Storage),
		})
	}

	// Lấy total_exp, total_star, change_exp, change_star từ bảng user_star_exp (is_current = true)
	var totalExp float64
	var totalStar int64
	var responseChangeExp float64
	var responseChangeStar int64

	if homework.QuestionForm == models.QuestionFormQuestionType {
		userStarExpRepo := repositories.NewUserStarExpRepository()
		currentUserStarExp, err := userStarExpRepo.GetCurrentByUserID(userID)
		if err == nil && currentUserStarExp != nil {
			totalExp = currentUserStarExp.TotalExp
			totalStar = currentUserStarExp.TotalStar
			responseChangeExp = currentUserStarExp.ChangeExp
			responseChangeStar = currentUserStarExp.ChangeStar
		}
	}

	// Tính completion_rate qua service
	completionRate, err := c.homeworkUserService.CalculateCompletionRate(req.HomeworkId, userID)
	if err != nil {
		// Nếu có lỗi, để completion_rate = 0
		completionRate = 0
	}

	response := &prot.SubmitHomeworkResponse{
		HomeworkId:                    req.HomeworkId,
		UserId:                         userID,
		Score:                         score,
		Ratio:                         ratio,
		HasManualScoring:              hasManualScoring,
		HasQuestionsNeedManualGrading: hasManualScoring,
		Files:                         files,
		Star:                          star,
		IsCompleted:                   isCompleted,
		Exp:                           exp,
		TotalExp:                      totalExp,
		TotalStar:                     totalStar,
		ChangeExp:                     responseChangeExp,
		ChangeStar:                    responseChangeStar,
		CompletionRate:                completionRate,
	}

	utils.Respond(ctx, response, nil, "")
}

// SaveManualScoring lưu câu trả lời cần chấm điểm thủ công
func (c *SaveScoreController) SaveManualScoring(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveAnswerManualRequest](ctx, func() *prot.SaveAnswerManualRequest {
		return &prot.SaveAnswerManualRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	// Kiểm tra xem là exam, exercise hay homework
	if req.ExamId > 0 {
		if err := c.serviceMS.SaveExamManualScoringService(*req, userID); err != nil {
			utils.Respond(ctx, &prot.SaveMessage{
				Success: false,
				Message: err.Error(),
			}, nil, "", http.StatusInternalServerError)
			return
		}
	} else if req.ExerciseId > 0 {
		if err := c.serviceMS.SaveExerciseManualScoringService(*req, userID); err != nil {
			utils.Respond(ctx, &prot.SaveMessage{
				Success: false,
				Message: err.Error(),
			}, nil, "", http.StatusInternalServerError)
			return
		}
	} else if req.HomeworkId > 0 {
		if err := c.serviceMS.SaveHomeworkManualScoringService(*req, userID); err != nil {
			utils.Respond(ctx, &prot.SaveMessage{
				Success: false,
				Message: err.Error(),
			}, nil, "", http.StatusInternalServerError)
			return
		}
		// Check và update did_it_again cho manual scoring (luôn coi là đúng vì đã được save)
		c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, req.QuestionId, true)
	} else {
		utils.Respond(ctx, &prot.SaveMessage{
			Success: false,
			Message: "exam_id, exercise_id or homework_id is required",
		}, nil, "")
		return
	}

	utils.Respond(ctx, &prot.SaveMessage{
		Success: true,
		Message: "Saved successfully",
	}, nil, "")
}

func (c *SaveScoreController) SaveScoreManualScoring(ctx *gin.Context) {
	req, err, message := utils.GetBody[*prot.SaveScoreManualRequest](ctx, func() *prot.SaveScoreManualRequest {
		return &prot.SaveScoreManualRequest{}
	})
	if err != nil {
		utils.Respond(ctx, nil, err, message)
		return
	}

	tokenStr := ctx.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	// Kiểm tra xem là exam, exercise hay homework
	if req.ExamId > 0 {
		err = c.saveScoreManualScoringService.SaveScoreManualScoring(req, userID)
	} else if req.ExerciseId > 0 {
		err = c.saveScoreManualScoringService.SaveScoreManualScoringExercise(req, userID)
	} else if req.HomeworkId > 0 {
		err = c.saveScoreManualScoringService.SaveScoreManualScoringHomework(req, userID)
		// Check và update did_it_again cho từng question_id trong score_list (luôn coi là đúng vì đã được save)
		if err == nil {
			for questionIdStr := range req.ScoreList {
				// Convert string to int64
				if questionId, parseErr := strconv.ParseInt(questionIdStr, 10, 64); parseErr == nil {
					c.checkAndUpdateSkipQuestionIfCorrect(req.HomeworkId, userID, questionId, true)
				}
			}
		}
	} else {
		utils.Respond(ctx, nil, fmt.Errorf("exam_id, exercise_id or homework_id is required"), "messages.data_invalid", http.StatusBadRequest)
		return
	}

	if err != nil {
		utils.Respond(ctx, nil, err, "")
		return
	}

	responseData := gin.H{"user_id": req.UserId}
	if req.ExamId > 0 {
		responseData["exam_id"] = req.ExamId
	} else if req.ExerciseId > 0 {
		responseData["exercise_id"] = req.ExerciseId
	} else if req.HomeworkId > 0 {
		responseData["homework_id"] = req.HomeworkId
	}

	utils.Respond(ctx, responseData, nil, "")
}

func (s *SaveScoreController) Evaluate(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.EvaluateRequest](c, func() *prot.EvaluateRequest {
		return &prot.EvaluateRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	data, err := s.serviceB.Evaluate(req)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	utils.Respond(c, data, nil, "")
}

func (s *SaveScoreController) TeacherEvaluate(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.TeacherEvaluateRequest](c, func() *prot.TeacherEvaluateRequest {
		return &prot.TeacherEvaluateRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	data, err := s.homeworkUserService.TeacherEvaluate(req)

	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, data, nil, "")
}
