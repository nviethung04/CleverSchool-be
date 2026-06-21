package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	_ "be-lms/models"
	"encoding/json"
	"fmt"

	_ "gorm.io/gorm"
)

type GetExamAnswersRepository interface {
	GetExamQuestions(examID, userID int64) ([]dto.QuestionWithScore, error)
	GetMultipleChoiceAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerMultipleChoice, error)
	GetFillInBlankAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerFillInBlank, error)
	GetPositionAnswer(examID, userID, questionID int64) ([]dto.ExamAnswerPosition, error)
	GetMatchingAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerMatching, error)
	GetLabelingAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerLabeling, error)
	GetGroupAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerGroup, error)
	GetManualAnswer(examID, questionID, userID int64) (*dto.ExamAnswerManual, error)
	GetAllAndTrueAnswers(questionID int64) (allAnswers struct {
		IDs                       []int64
		FileURLs, Kinds, Contents []string
	}, trueAnswers struct {
		IDs                       []int64
		FileURLs, Kinds, Contents []string
	}, err error)
	GetStudentName(userID int64) (string, error)
	GetExamUserInfo(examID, userID int64) (duration int64, totalScore float64, submittedAt int64, hasManualScoring bool, ratio float64, err error)
	GetTotalQuestions(examID int64) (int, error)
	GetUnscoredManualQuestions(examID, userID int64) (int, error)
	GetExamComment(examID, userID int64) (string, error)
}

type getExamAnswersRepository struct {
}

func NewGetExamAnswersRepository() GetExamAnswersRepository {
	return &getExamAnswersRepository{}
}

func (r *getExamAnswersRepository) GetExamQuestions(examID, userID int64) ([]dto.QuestionWithScore, error) {
	// Sử dụng cloned_questions thay vì exam_questions
	var questions []dto.QuestionWithScore
	err := db.ReplicaDB.Table("cloned_questions").
		Select("questions.id, questions.content, questions.question_type as type, questions.file_url as file_url, questions.kind as kind, 1.0 as max_score").
		Joins("JOIN questions ON questions.id = ANY(SELECT jsonb_array_elements_text(cloned_questions.questions)::bigint)").
		Where("cloned_questions.assignment_id = ? AND cloned_questions.assignment_type = ?", examID, "exam").
		Find(&questions).Error
	return questions, err
}

// GetMultipleChoiceAnswer Multiple Choice: Lấy câu trả lời của người dùng và câu trả lời đúng (is_correct = true)
func (r *getExamAnswersRepository) GetMultipleChoiceAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerMultipleChoice, error) {
	var result []dto.ExamAnswerMultipleChoice

	// Query 2: Lấy tất cả đáp án người dùng đã chọn từ bảng exam_question_users
	var userAnswers []struct {
		ID            int64
		ExamID        int64
		QuestionID    int64
		UserID        int64
		AnswerID      int64
		AnswerContent string
		FileURL       string
		Kind          string
		IsCorrect     bool
		Score         float64
		CreatedAt     string
		UpdatedAt     string
	}
	if err := db.ReplicaDB.Table("exam_question_users").
		Select("exam_question_users.*").
		Where("exam_question_users.exam_id = ? AND exam_question_users.question_id = ? AND exam_question_users.user_id = ?", examID, questionID, userID).
		Find(&userAnswers).Error; err != nil {
		return nil, err
	}

	// Gom nhóm các đáp án của người dùng theo exam_id, question_id, user_id
	answerMap := make(map[string]*dto.ExamAnswerMultipleChoice)
	for _, userAnswer := range userAnswers {
		key := fmt.Sprintf("%d_%d_%d", userAnswer.ExamID, userAnswer.QuestionID, userAnswer.UserID)
		if answer, exists := answerMap[key]; exists {
			// Thêm đáp án vào mảng nếu đã tồn tại
			answer.AnswerIDs = append(answer.AnswerIDs, userAnswer.AnswerID)
			answer.AnswerContents = append(answer.AnswerContents, userAnswer.AnswerContent)
			answer.AnswerFileURLs = append(answer.AnswerFileURLs, userAnswer.FileURL)
			answer.AnswerKinds = append(answer.AnswerKinds, userAnswer.Kind)
		} else {
			// Tạo mới nếu chưa tồn tại
			answerMap[key] = &dto.ExamAnswerMultipleChoice{
				ID:                 userAnswer.ID,
				ExamID:             userAnswer.ExamID,
				QuestionID:         userAnswer.QuestionID,
				UserID:             userAnswer.UserID,
				AnswerIDs:          []int64{userAnswer.AnswerID},
				AnswerContents:     []string{userAnswer.AnswerContent},
				AnswerFileURLs:     []string{userAnswer.FileURL},
				AnswerKinds:        []string{userAnswer.Kind},
				IsCorrect:          userAnswer.IsCorrect,
				Score:              userAnswer.Score,
				CreatedAt:          userAnswer.CreatedAt,
				UpdatedAt:          userAnswer.UpdatedAt,
				TrueAnswerIDs:      make([]int64, 0),
				TrueAnswers:        make([]string, 0),
				TrueAnswerFileURLs: make([]string, 0),
				TrueAnswerKinds:    make([]string, 0),
				AllAnswerIDs:       make([]int64, 0),
				AllAnswerFileURLs:  make([]string, 0),
				AllAnswerKinds:     make([]string, 0),
				AllAnswerContents:  make([]string, 0),
			}
		}
	}

	// Thêm đáp án đúng vào kết quả
	for _, answer := range answerMap {
		result = append(result, *answer)
	}

	return result, nil
}

// GetFillInBlankAnswer Fill in Blank: Lấy câu trả lời của người dùng và câu trả lời đúng theo vị trí
func (r *getExamAnswersRepository) GetFillInBlankAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerFillInBlank, error) {
	var answers []dto.ExamAnswerFillInBlank

	// Lấy đáp án người dùng và đáp án đúng
	if err := db.ReplicaDB.Table("exam_question_user_fill_in_blanks").
		Select(`
			exam_question_user_fill_in_blanks.sort_position,
			exam_question_user_fill_in_blanks.answer,
			exam_question_user_fill_in_blanks.is_correct,
			exam_question_user_fill_in_blanks.score
		`).
		Where("exam_question_user_fill_in_blanks.exam_id = ? AND exam_question_user_fill_in_blanks.question_id = ? AND exam_question_user_fill_in_blanks.user_id = ?", examID, questionID, userID).
		Order("exam_question_user_fill_in_blanks.sort_position ASC").
		Find(&answers).Error; err != nil {
		return nil, err
	}

	return answers, nil
}

// GetPositionAnswer Ordering/Dragdrop: Lấy câu trả lời của người dùng và câu trả lời đúng theo thứ tự
func (r *getExamAnswersRepository) GetPositionAnswer(examID, userID, questionID int64) ([]dto.ExamAnswerPosition, error) {
	var result []dto.ExamAnswerPosition

	// Lấy tất cả các vị trí từ bảng exam_question_user_positions
	var userPositions []models.ExamQuestionUserPosition
	if err := db.ReplicaDB.Where("exam_id = ? AND user_id = ? AND question_id = ?", examID, userID, questionID).Find(&userPositions).Error; err != nil {
		return nil, err
	}

	// Xây dựng kết quả
	for _, userPos := range userPositions {
		result = append(result, dto.ExamAnswerPosition{
			SortPosition:        int32(userPos.SortPosition),
			AnswerGroupPosition: userPos.AnswerGroupPosition,
			IsCorrect:           userPos.IsCorrect,
			Score:               userPos.Score,
		})
	}

	return result, nil
}

// GetMatchingAnswer Matching: Lấy câu trả lời của người dùng và câu trả lời đúng (first_item_id = second_item_id)
func (r *getExamAnswersRepository) GetMatchingAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerMatching, error) {
	var answers []dto.ExamAnswerMatching

	err := db.ReplicaDB.Table("exam_question_user_matchings").
		Select(`
			exam_question_user_matchings.*
		`).
		Where("exam_question_user_matchings.exam_id = ? AND exam_question_user_matchings.question_id = ? AND exam_question_user_matchings.user_id = ?", examID, questionID, userID).
		Find(&answers).Error
	return answers, err
}

// GetLabelingAnswer Labeling: Lấy câu trả lời của người dùng và câu trả lời đúng (blank_id = answer_id)
func (r *getExamAnswersRepository) GetLabelingAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerLabeling, error) {
	var answers []dto.ExamAnswerLabeling
	err := db.ReplicaDB.Table("exam_question_user_labelings").
		Select(`
			exam_question_user_labelings.blank_id,
			exam_question_user_labelings.answer_id,
			exam_question_user_labelings.is_correct,
			exam_question_user_labelings.score
		`).
		Where("exam_question_user_labelings.exam_id = ? AND exam_question_user_labelings.question_id = ? AND exam_question_user_labelings.user_id = ?", examID, questionID, userID).
		Find(&answers).Error
	return answers, err
}

// GetGroupAnswer Category: Lấy câu trả lời của người dùng và câu trả lời đúng (answer_id = group_id)
func (r *getExamAnswersRepository) GetGroupAnswer(examID, questionID, userID int64) ([]dto.ExamAnswerGroup, error) {
	var answers []dto.ExamAnswerGroup

	if err := db.ReplicaDB.Table("exam_question_user_groups").
		Select(`
			exam_question_user_groups.*
		`).
		Where("exam_question_user_groups.exam_id = ? AND exam_question_user_groups.question_id = ? AND exam_question_user_groups.user_id = ?", examID, questionID, userID).
		Find(&answers).Error; err != nil {
		return nil, err
	}

	return answers, nil
}

// GetManualAnswer Manual: Lấy câu trả lời của người dùng và điểm số (nếu đã chấm)
func (r *getExamAnswersRepository) GetManualAnswer(examID, questionID, userID int64) (*dto.ExamAnswerManual, error) {
	var answer dto.ExamAnswerManual
	err := db.ReplicaDB.Table("exam_question_user_manual_scoring").
		Select("answer, file_info as answer_file_info, score, is_scored").
		Where("exam_id = ? AND question_id = ? AND user_id = ?", examID, questionID, userID).
		First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetAllAndTrueAnswers trả về all_answer và true_answer cho câu hỏi multiple choice
func (r *getExamAnswersRepository) GetAllAndTrueAnswers(questionID int64) (allAnswers struct {
	IDs                       []int64
	FileURLs, Kinds, Contents []string
}, trueAnswers struct {
	IDs                       []int64
	FileURLs, Kinds, Contents []string
}, err error) {
	// Lấy all answers
	var allAns []struct {
		ID      int64
		FileURL string
		Kind    string
		Content string
	}
	err = db.ReplicaDB.Table("answers").
		Select("id, file_url, kind, content").
		Where("question_id = ?", questionID).
		Find(&allAns).Error
	if err != nil {
		return
	}
	for _, a := range allAns {
		allAnswers.IDs = append(allAnswers.IDs, a.ID)
		allAnswers.FileURLs = append(allAnswers.FileURLs, a.FileURL)
		allAnswers.Kinds = append(allAnswers.Kinds, a.Kind)
		allAnswers.Contents = append(allAnswers.Contents, a.Content)
	}

	// Lấy true answers
	var trueAns []struct {
		ID      int64
		FileURL string
		Kind    string
		Content string
	}
	err = db.ReplicaDB.Table("answers").
		Select("id, file_url, kind, content").
		Where("question_id = ? AND is_correct = true", questionID).
		Find(&trueAns).Error
	if err != nil {
		return
	}
	for _, a := range trueAns {
		trueAnswers.IDs = append(trueAnswers.IDs, a.ID)
		trueAnswers.FileURLs = append(trueAnswers.FileURLs, a.FileURL)
		trueAnswers.Kinds = append(trueAnswers.Kinds, a.Kind)
		trueAnswers.Contents = append(trueAnswers.Contents, a.Content)
	}
	return
}

func (r *getExamAnswersRepository) GetStudentName(userID int64) (string, error) {
	var name string
	err := db.ReplicaDB.Table("users").Select("name").Where("id = ?", userID).Scan(&name).Error
	return name, err
}

func (r *getExamAnswersRepository) GetExamUserInfo(examID, userID int64) (duration int64, totalScore float64, submittedAt int64, hasManualScoring bool, ratio float64, err error) {
	type result struct {
		Time             int64
		Score            float64
		CreatedAt        float64
		HasManualScoring bool
		Ratio            float64
	}
	var res result
	err = db.ReplicaDB.Table("exam_users").
		Select("time, score, extract(epoch from created_at) as created_at, has_manual_scoring, COALESCE(ratio,0) as ratio").
		Where("exam_id = ? AND user_id = ?", examID, userID).
		Scan(&res).Error
	return res.Time, res.Score, int64(res.CreatedAt), res.HasManualScoring, res.Ratio, err
}

func (r *getExamAnswersRepository) GetTotalQuestions(examID int64) (int, error) {
	// Sử dụng cloned_questions thay vì exam_questions
	var raw struct{ Questions json.RawMessage }
	if err := db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", examID, "exam").
		First(&raw).Error; err != nil {
		return 0, err
	}
	
	var questions []interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return 0, err
	}
	
	return len(questions), nil
}

func (r *getExamAnswersRepository) GetUnscoredManualQuestions(examID, userID int64) (int, error) {
	var count int64
	err := db.ReplicaDB.Table("exam_question_user_manual_scoring").Where("exam_id = ? AND user_id = ? AND is_scored = false", examID, userID).Count(&count).Error
	return int(count), err
}

func (r *getExamAnswersRepository) GetExamComment(examID, userID int64) (string, error) {
	var content string
	err := db.ReplicaDB.Table("exam_comments").Select("content").Where("exam_id = ? AND student_id = ?", examID, userID).Order("created_at desc").Limit(1).Scan(&content).Error
	return content, err
}
