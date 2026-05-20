package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContestQuestionService interface {
	GetContestRoundQuestions(c *gin.Context, contestRoundId int64) (interface{}, error)
	AssignQuestionsToContestRound(c *gin.Context, contestRoundId int64, req interface{}) error
	RemoveQuestionsFromContestRound(c *gin.Context, contestRoundId int64, req interface{}) error
	UpdateContestRoundQuestion(c *gin.Context, contestRoundId int64, questionId int, req *prot.Question) (*prot.Question, error)
}

type contestQuestionService struct {
	repo         repositories.ContestQuestionRepository
	clonedRepo   repositories.ClonedQuestionRepository
	questionRepo repositories.QuestionRepository
}

func NewContestQuestionService() ContestQuestionService {
	return &contestQuestionService{
		repo:         repositories.NewContestQuestionRepository(),
		clonedRepo:   repositories.NewClonedQuestionRepository(),
		questionRepo: repositories.NewQuestionRepository(),
	}
}

func (s *contestQuestionService) GetContestRoundQuestions(c *gin.Context, contestRoundId int64) (interface{}, error) {
	// Use cloned_questions table like exam/homework
	cloned, err := s.clonedRepo.FindByAssignment(contestRoundId, models.ClonedQuestionTypeContestRound)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []interface{}{}, nil
		}
		return nil, err
	}

	// Parse questions from JSON and format them
	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(cloned.Questions, &rawQuestions); err != nil {
		return nil, err
	}

	questionResource := resources.NewQuestionResource()
	questions := make([]*prot.Question, 0, len(rawQuestions))

	for _, q := range rawQuestions {
		question, err := questionResource.ParseProtQuestionFromJSON(q)
		if err != nil {
			continue // Skip invalid questions
		}

		// Format question for display
		question = questionResource.FormatByRole(question, int64(utils.GetCurrentRoleId(c)))
		question = questionResource.FormatStaticURL(question)
		question = questionResource.FormatMediaUrls(question)

		questions = append(questions, question)
	}

	return questions, nil
}

func (s *contestQuestionService) AssignQuestionsToContestRound(c *gin.Context, contestRoundId int64, req interface{}) error {
	// Parse request to get question IDs
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return errors.New("invalid request format")
	}

	questionIdsInterface, exists := reqMap["question_ids"]
	if !exists {
		return errors.New("question_ids not found in request")
	}

	questionIds, ok := questionIdsInterface.([]interface{})
	if !ok {
		return errors.New("question_ids must be an array")
	}

	// Get full question data from database
	questions := make([]*prot.Question, 0, len(questionIds))
	questionResource := resources.NewQuestionResource()

	// fmt.Printf("🔍 Processing %d question IDs\n", len(questionIds))

	for _, idInterface := range questionIds {
		idFloat, ok := idInterface.(float64)
		if !ok {
			// fmt.Printf("🔍 Skipping invalid ID: %v\n", idInterface)
			continue // Skip invalid IDs
		}
		questionId := int64(idFloat)
		// fmt.Printf("🔍 Processing question ID: %d\n", questionId)

		// Get question from database
		question, err := s.questionRepo.FindByID(int(questionId))
		if err != nil {
			// fmt.Printf("🔍 Question not found: %d, error: %v\n", questionId, err)
			continue // Skip if question not found
		}

		// fmt.Printf("🔍 Found question: ID=%d\n", question.ID)

		// Convert to protobuf format
		protQuestion := questionResource.FormatQuestion(question)
		if protQuestion == nil {
			// fmt.Printf("🔍 Failed to convert question to protobuf: %d\n", questionId)
			continue // Skip if conversion fails
		}

		// fmt.Printf("🔍 Converted to protobuf: ID=%d, Type=%s\n", protQuestion.Id, protQuestion.Type)
		questions = append(questions, protQuestion)
	}

	// fmt.Printf("🔍 Total questions processed: %d\n", len(questions))

	// Create request with full question data
	questionScores := make(map[string]float64)
	for _, q := range questions {
		// Use default score of 1.0 for all questions
		questionScores[strconv.FormatInt(q.Id, 10)] = 1.0
	}

	createReq := &prot.CreateQuestionRelationRequest{
		QuestionScores: questionScores,
		IsReset:        false,
	}

	// Use the same logic as exam but with full question data
	return s.UpdateCloneQuestionWithFullData(c, contestRoundId, models.ClonedQuestionTypeContestRound, questions, createReq)
}

func (s *contestQuestionService) RemoveQuestionsFromContestRound(c *gin.Context, contestRoundId int64, req interface{}) error {
	// Remove all questions from contest round
	return s.clonedRepo.DeleteByAssignment(contestRoundId, models.ClonedQuestionTypeContestRound)
}

func (s *contestQuestionService) UpdateContestRoundQuestion(c *gin.Context, contestRoundId int64, questionId int, req *prot.Question) (*prot.Question, error) {
	// Similar to exam's UpdateCloned method
	clonedData, err := s.clonedRepo.FindByAssignment(contestRoundId, models.ClonedQuestionTypeContestRound)
	if err != nil {
		return nil, err
	}

	// Parse existing questions
	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(clonedData.Questions, &rawQuestions); err != nil {
		return nil, err
	}

	// Find and update the specific question
	questionResource := resources.NewQuestionResource()
	for i, q := range rawQuestions {
		question, err := questionResource.ParseProtQuestionFromJSON(q)
		if err != nil {
			continue
		}

		if int(question.Id) == questionId {
			// Update the question
			updatedQuestionJSON, err := json.Marshal(req)
			if err != nil {
				return nil, err
			}
			rawQuestions[i] = updatedQuestionJSON
			break
		}
	}

	// Save back to cloned_questions
	updatedQuestionsJSON, err := json.Marshal(rawQuestions)
	if err != nil {
		return nil, err
	}

	clonedData.Questions = updatedQuestionsJSON
	clonedData.UpdatedAt = time.Now()
	clonedData.UpdatedBy = int64(utils.GetCurrentUserId(c))

	if err := s.clonedRepo.Update(clonedData); err != nil {
		return nil, err
	}

	return req, nil
}

// Helper method similar to exam's UpdateCloneQuestion
func (s *contestQuestionService) UpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, req *prot.CreateQuestionRelationRequest) error {
	clonedData, err := s.clonedRepo.FindByAssignment(assignmentId, assignmentType)
	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData.ID == 0 {
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, "contest_round_id", models.Question{}, req, &models.ClonedQuestion{})
	} else {
		return s.CreateOrUpdateCloneQuestion(c, assignmentId, assignmentType, "contest_round_id", models.Question{}, req, clonedData)
	}
}

// Helper method to update clone question with full question data
func (s *contestQuestionService) UpdateCloneQuestionWithFullData(c *gin.Context, assignmentId int64, assignmentType string, questions []*prot.Question, req *prot.CreateQuestionRelationRequest) error {
	clonedData, err := s.clonedRepo.FindByAssignment(assignmentId, assignmentType)
	if errors.Is(err, gorm.ErrRecordNotFound) || clonedData.ID == 0 {
		// Create new with full question data
		questionsJSON, err := json.Marshal(questions)
		if err != nil {
			return err
		}

		// fmt.Printf("🔍 Creating new cloned question with %d questions\n", len(questions))
		// fmt.Printf("🔍 Questions JSON: %s\n", string(questionsJSON))

		clonedData = &models.ClonedQuestion{
			AssignmentID:   assignmentId,
			AssignmentType: assignmentType,
			Questions:      questionsJSON,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      int64(utils.GetCurrentUserId(c)),
			UpdatedBy:      int64(utils.GetCurrentUserId(c)),
		}
		return s.clonedRepo.Create(clonedData)
	} else {
		// Update existing with full question data
		questionsJSON, err := json.Marshal(questions)
		if err != nil {
			return err
		}

		// fmt.Printf("🔍 Updating existing cloned question with %d questions\n", len(questions))
		// fmt.Printf("🔍 Questions JSON: %s\n", string(questionsJSON))

		clonedData.Questions = questionsJSON
		clonedData.UpdatedAt = time.Now()
		clonedData.UpdatedBy = int64(utils.GetCurrentUserId(c))
		return s.clonedRepo.Update(clonedData)
	}
}

// Helper method similar to exam's CreateOrUpdateCloneQuestion
func (s *contestQuestionService) CreateOrUpdateCloneQuestion(c *gin.Context, assignmentId int64, assignmentType string, filterKey string, question models.Question, req *prot.CreateQuestionRelationRequest, clonedData *models.ClonedQuestion) error {
	// This is a simplified version - the full implementation would be similar to exam's logic
	// For now, create a basic cloned question record

	questionData := make([]map[string]interface{}, 0)
	for questionID, score := range req.QuestionScores {
		questionData = append(questionData, map[string]interface{}{
			"id":    questionID,
			"score": score,
		})
	}

	questionsJSON, err := json.Marshal(questionData)
	if err != nil {
		return err
	}

	if clonedData.ID == 0 {
		// Create new
		clonedData = &models.ClonedQuestion{
			AssignmentID:   assignmentId,
			AssignmentType: assignmentType,
			Questions:      questionsJSON,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      int64(utils.GetCurrentUserId(c)),
			UpdatedBy:      int64(utils.GetCurrentUserId(c)),
		}
		return s.clonedRepo.Create(clonedData)
	} else {
		// Update existing
		clonedData.Questions = questionsJSON
		clonedData.UpdatedAt = time.Now()
		clonedData.UpdatedBy = int64(utils.GetCurrentUserId(c))
		return s.clonedRepo.Update(clonedData)
	}
}

