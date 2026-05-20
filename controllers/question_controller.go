package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/dto"
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"
	"strconv"

	"be-cleverschool/middleware"

	"github.com/gin-gonic/gin"
)

type QuestionController struct {
	service services.QuestionService
	*GenericController[models.Question, prot.Question, *prot.Question]
}

func NewQuestionController(service services.QuestionService) *QuestionController {
	questionResource := resources.NewQuestionResource()
	questionResourceAdapter := NewQuestionResourceAdapter(questionResource)

	genericController := NewGenericController(
		service,
		questionResourceAdapter,
		func() *prot.Question {
			return &prot.Question{}
		},
		func(questions []*prot.Question, totalCount uint64) interface{} {
			return &prot.QuestionsResponse{
				Questions:  questions,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &QuestionController{
		GenericController: genericController,
		service:           service,
	}
}

func (qc *QuestionController) GetAll(c *gin.Context) {
	cloneQuestions, totalCount, hasCloned := qc.service.GetCloned(c)

	if hasCloned {
		list := &prot.QuestionsResponse{
			Questions:  cloneQuestions,
			TotalCount: int64(totalCount),
		}
		utils.Respond(c, list, nil, "")
		return
	}

	questions, totalCount, err := qc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	var questionPtrs []*models.Question
	for i := range questions {
		questionPtrs = append(questionPtrs, &questions[i])
	}

	questionResource := resources.NewQuestionResource()
	formattedQuestion := questionResource.FormatQuestions(questionPtrs)

	list := &prot.QuestionsResponse{
		Questions:  formattedQuestion,
		TotalCount: int64(totalCount),
	}

	utils.Respond(c, list, err, "")
}

func (qc *QuestionController) Update(c *gin.Context) {
	// Kiểm tra nếu có query parameters đặc biệt, xử lý riêng
	homeworkId, _ := strconv.Atoi(c.Query("homework_id"))
	examId, _ := strconv.Atoi(c.Query("exam_id"))
	lessonPlanPartId, _ := strconv.Atoi(c.Query("lesson_plan_part_id"))
	levelTestId, _ := strconv.Atoi(c.Query("level_test_id"))
	contestRoundId, _ := strconv.Atoi(c.Query("contest_round_id"))

	// Nếu có assignment ID, xử lý cloned question
	if homeworkId != 0 || examId != 0 || lessonPlanPartId != 0 || levelTestId != 0 || contestRoundId != 0 {
		var assignmentID int
		var assignmentType string

		switch {
		case homeworkId != 0:
			assignmentID = homeworkId
			assignmentType = models.ClonedQuestionTypeHomework
		case examId != 0:
			assignmentID = examId
			assignmentType = models.ClonedQuestionTypeExam
		case lessonPlanPartId != 0:
			assignmentID = lessonPlanPartId
			assignmentType = models.ClonedQuestionTypeLessonPlanPart
		case levelTestId != 0:
			assignmentID = levelTestId
			assignmentType = models.ClonedQuestionTypeLevelTest
		case contestRoundId != 0:
			assignmentID = contestRoundId
			assignmentType = models.ClonedQuestionTypeContestRound
		}

		// Lấy ID từ path
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid ID"), "messages.id_invalid", 400)
			return
		}

		// Lấy request body
		req, err, message := utils.GetBody[*prot.Question](c, func() *prot.Question {
			return &prot.Question{}
		})
		if err != nil {
			utils.Respond(c, nil, err, message)
			return
		}

		// Update cloned question
		question, err := qc.service.UpdateCloned(c, id, assignmentID, assignmentType, req)
		if err != nil {
			utils.Respond(c, nil, err, "")
			return
		}

		utils.Respond(c, question, err, "")
		return
	}

	// Nếu không có assignment ID, sử dụng logic từ GenericController
	qc.GenericController.Update(c)
}

func (qc *QuestionController) SyncKeywords(c *gin.Context) {
	err := qc.service.SyncKeywords(c, 0)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}
	utils.Respond(c, nil, nil, "")
}

func (qc *QuestionController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := qc.service.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (qc *QuestionController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	middleware.SafeGo(func() {
		err := qc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("Question import failed", "error", err)
		} else {
			config.Log.Info("Question import finished successfully")
		}
	})

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}

