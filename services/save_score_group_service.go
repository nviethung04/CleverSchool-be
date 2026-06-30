package services

import (
	"be-lms/database/db"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"fmt"
	"strconv"
)

type SaveScoreGroupService interface {
	SaveScoreGroupExam(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
	SaveScoreGroupHomework(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
	SaveScoreGroupExercise(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error)
}

type saveScoreGroupService struct {
	repo                  repositories.SaveScoreGroupRepository
	correctRepo           repositories.SaveCorrectHomeworkRepository
	clonedQuestionService ClonedQuestionService
}

func NewSaveScoreGroupService(repo repositories.SaveScoreGroupRepository, clonedQuestionService ClonedQuestionService) SaveScoreGroupService {
	return &saveScoreGroupService{
		repo:                  repo,
		correctRepo:           repositories.NewSaveCorrectHomeworkRepository(),
		clonedQuestionService: clonedQuestionService,
	}
}

func (s *saveScoreGroupService) SaveScoreGroupExam(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.ExamId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	categoryQuestion, err := categoryQuestionForScoring(q)
	if err != nil {
		return nil, err
	}

	perGroupScore := 1.0
	numGroups := len(req.Answers)
	totalScore, correctCount, groupResults := buildCategoryGroupScoreResults(req.Answers, categoryQuestion, perGroupScore)

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var records []*models.ExamQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		answerID, err := strconv.ParseInt(answerIDStr, 10, 64)
		if err != nil {
			continue
		}
		correctGroupID := categoryCorrectGroupID(categoryQuestion, answerID, answerIDStr)
		isCorrect := correctGroupID > 0 && groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
		}
		records = append(records, &models.ExamQuestionUserGroup{
			ExamID:     req.ExamId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   answerID,
			GroupID:    groupID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	if err := s.repo.SaveBatchExamQuestionUserGroup(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseGroup{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Groups:       groupResults,
		IsAllCorrect: correctCount == numGroups,
	}, nil
}

func (s *saveScoreGroupService) SaveScoreGroupHomework(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.HomeworkId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	questions, err := s.clonedQuestionService.GetQuestionsMap(req.HomeworkId, "homework")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	categoryQuestion, err := categoryQuestionForScoring(q)
	if err != nil {
		return nil, err
	}

	perGroupScore := 1.0
	numGroups := len(req.Answers)
	totalScore, correctCount, groupResults := buildCategoryGroupScoreResults(req.Answers, categoryQuestion, perGroupScore)

	tx := db.MasterDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	needSave, err := CheckAndCleanExistingAnswersWithTx(tx, "homework_question_user_groups", req.HomeworkId, userID, req.QuestionId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var records []*models.HomeworkQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		answerID, err := strconv.ParseInt(answerIDStr, 10, 64)
		if err != nil {
			continue
		}
		correctGroupID := categoryCorrectGroupID(categoryQuestion, answerID, answerIDStr)
		isCorrect := correctGroupID > 0 && groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
		}
		records = append(records, &models.HomeworkQuestionUserGroup{
			HomeworkID: req.HomeworkId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   answerID,
			GroupID:    groupID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}

	if !needSave {
		tx.Commit()
		return &prot.SaveScoreResponseGroup{
			QuestionId:   req.QuestionId,
			TotalScore:   utils.RoundTo2Decimal(totalScore),
			Groups:       groupResults,
			IsAllCorrect: correctCount == numGroups,
		}, nil
	}

	if correctCount == numGroups {
		err := s.correctRepo.UpsertHomeworkUserOnCorrect(req.HomeworkId, req.LessonId, userID, req.QuestionId, "group")
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := s.repo.SaveBatchHomeworkQuestionUserGroup(records, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &prot.SaveScoreResponseGroup{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Groups:       groupResults,
		IsAllCorrect: correctCount == numGroups,
	}, nil
}

func (s *saveScoreGroupService) SaveScoreGroupExercise(req *prot.SaveScoreGroupRequest, userID int64) (*prot.SaveScoreResponseGroup, error) {
	if req.ExerciseId == 0 || len(req.Answers) == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.input_invalid"))
	}

	questions, err := s.clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	if err != nil {
		return nil, err
	}
	q, ok := questions[fmt.Sprintf("%v", req.QuestionId)]
	if !ok {
		return nil, fmt.Errorf("Không tìm thấy câu hỏi trong cloned_question")
	}

	categoryQuestion, err := categoryQuestionForScoring(q)
	if err != nil {
		return nil, err
	}

	perGroupScore := 1.0
	numGroups := len(req.Answers)
	totalScore, correctCount, groupResults := buildCategoryGroupScoreResults(req.Answers, categoryQuestion, perGroupScore)

	var records []*models.ExerciseQuestionUserGroup
	for answerIDStr, groupID := range req.Answers {
		answerID, err := strconv.ParseInt(answerIDStr, 10, 64)
		if err != nil {
			continue
		}
		correctGroupID := categoryCorrectGroupID(categoryQuestion, answerID, answerIDStr)
		isCorrect := correctGroupID > 0 && groupID == correctGroupID
		score := 0.0
		if isCorrect {
			score = perGroupScore
		}
		records = append(records, &models.ExerciseQuestionUserGroup{
			ExerciseID: req.ExerciseId,
			LessonID:   req.LessonId,
			UserID:     userID,
			QuestionID: req.QuestionId,
			AnswerID:   answerID,
			GroupID:    groupID,
			IsCorrect:  isCorrect,
			Score:      score,
		})
	}
	if err := s.repo.SaveBatchExerciseQuestionUserGroup(records, db.MasterDB); err != nil {
		return nil, err
	}

	return &prot.SaveScoreResponseGroup{
		QuestionId:   req.QuestionId,
		TotalScore:   utils.RoundTo2Decimal(totalScore),
		Groups:       groupResults,
		IsAllCorrect: correctCount == numGroups,
	}, nil
}
