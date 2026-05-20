package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	_ "be-cleverschool/models"
	"encoding/json"
	"fmt"

	_ "gorm.io/gorm"
)

type GetExerciseAnswersRepository interface {
	GetExerciseQuestions(exerciseID, userID int64) ([]dto.QuestionWithScore, error)
	GetMultipleChoiceAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerMultipleChoice, error)
	GetFillInBlankAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerFillInBlank, error)
	GetPositionAnswerExercise(exerciseID, userID, questionID int64) ([]dto.ExerciseAnswerPosition, error)
	GetMatchingAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerMatching, error)
	GetLabelingAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerLabeling, error)
	GetGroupAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerGroup, error)
	GetManualAnswerExercise(exerciseID, questionID, userID int64) (*dto.ExerciseAnswerManual, error)
	GetAllAndTrueAnswers(questionID int64) (allAnswers struct {
		IDs                       []int64
		FileURLs, Kinds, Contents []string
	}, trueAnswers struct {
		IDs                       []int64
		FileURLs, Kinds, Contents []string
	}, err error)
    GetStudentName(userID int64) (string, error)
    GetExerciseUserInfo(exerciseID, userID int64) (duration int64, totalScore float64, submittedAt int64, hasManualScoring bool, ratio float64, err error)
	GetTotalQuestions(exerciseID int64) (int, error)
	GetUnscoredManualQuestions(exerciseID, userID int64) (int, error)
	GetExerciseComment(exerciseID, userID int64) (string, error)
}

type getExerciseAnswersRepository struct {
}

func NewGetExerciseAnswersRepository() GetExerciseAnswersRepository {
	return &getExerciseAnswersRepository{}
}

func (r *getExerciseAnswersRepository) GetExerciseQuestions(exerciseID, userID int64) ([]dto.QuestionWithScore, error) {
	// Sử dụng cloned_questions thay vì exercise_questions
	var questions []dto.QuestionWithScore
	err := db.ReplicaDB.Table("cloned_questions").
		Select("questions.id, questions.content, questions.question_type as type, questions.file_url as file_url, questions.kind as kind, 1.0 as max_score").
		Joins("JOIN questions ON questions.id = ANY(SELECT jsonb_array_elements_text(cloned_questions.questions)::bigint)").
		Where("cloned_questions.assignment_id = ? AND cloned_questions.assignment_type = ?", exerciseID, "exercise").
		Find(&questions).Error
	return questions, err
}

// GetMultipleChoiceAnswerExercise Multiple Choice: Lấy câu trả lời của người dùng và câu trả lời đúng (is_correct = true)
func (r *getExerciseAnswersRepository) GetMultipleChoiceAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerMultipleChoice, error) {
	var result []dto.ExerciseAnswerMultipleChoice

	// Query 2: Lấy tất cả đáp án người dùng đã chọn từ bảng exercise_question_users
	var userAnswers []struct {
		ID            int64
		ExerciseID    int64
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
	if err := db.ReplicaDB.Table("exercise_question_users").
		Select("exercise_question_users.*").
		Where("exercise_question_users.exercise_id = ? AND exercise_question_users.question_id = ? AND exercise_question_users.user_id = ?", exerciseID, questionID, userID).
		Find(&userAnswers).Error; err != nil {
		return nil, err
	}

	// Gom nhóm các đáp án của người dùng theo exercise_id, question_id, user_id
	answerMap := make(map[string]*dto.ExerciseAnswerMultipleChoice)
	for _, userAnswer := range userAnswers {
		key := fmt.Sprintf("%d_%d_%d", userAnswer.ExerciseID, userAnswer.QuestionID, userAnswer.UserID)
		if answer, exists := answerMap[key]; exists {
			// Thêm đáp án vào mảng nếu đã tồn tại
			answer.AnswerIDs = append(answer.AnswerIDs, userAnswer.AnswerID)
			answer.AnswerContents = append(answer.AnswerContents, userAnswer.AnswerContent)
			answer.AnswerFileURLs = append(answer.AnswerFileURLs, userAnswer.FileURL)
			answer.AnswerKinds = append(answer.AnswerKinds, userAnswer.Kind)
		} else {
			// Tạo mới nếu chưa tồn tại
			answerMap[key] = &dto.ExerciseAnswerMultipleChoice{
				ID:                 userAnswer.ID,
				ExerciseID:         userAnswer.ExerciseID,
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

// GetFillInBlankAnswerExercise Fill in Blank: Lấy câu trả lời của người dùng và câu trả lời đúng theo vị trí
func (r *getExerciseAnswersRepository) GetFillInBlankAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerFillInBlank, error) {
	var answers []dto.ExerciseAnswerFillInBlank

	// Lấy đáp án người dùng và đáp án đúng
	if err := db.ReplicaDB.Table("exercise_question_user_fill_in_blanks").
		Select(`
			exercise_question_user_fill_in_blanks.sort_position,
			exercise_question_user_fill_in_blanks.answer,
			exercise_question_user_fill_in_blanks.is_correct,
			exercise_question_user_fill_in_blanks.score
		`).
		Where("exercise_question_user_fill_in_blanks.exercise_id = ? AND exercise_question_user_fill_in_blanks.question_id = ? AND exercise_question_user_fill_in_blanks.user_id = ?", exerciseID, questionID, userID).
		Order("exercise_question_user_fill_in_blanks.sort_position ASC").
		Find(&answers).Error; err != nil {
		return nil, err
	}

	return answers, nil
}

// GetPositionAnswerExercise Ordering/Dragdrop: Lấy câu trả lời của người dùng và câu trả lời đúng theo thứ tự
func (r *getExerciseAnswersRepository) GetPositionAnswerExercise(exerciseID, userID, questionID int64) ([]dto.ExerciseAnswerPosition, error) {
	var result []dto.ExerciseAnswerPosition

	// Lấy tất cả các vị trí từ bảng exercise_question_user_positions
	var userPositions []models.ExerciseQuestionUserPosition
	if err := db.ReplicaDB.Where("exercise_id = ? AND user_id = ? AND question_id = ?", exerciseID, userID, questionID).Find(&userPositions).Error; err != nil {
		return nil, err
	}

	// Xây dựng kết quả
	for _, userPos := range userPositions {
		result = append(result, dto.ExerciseAnswerPosition{
			SortPosition:        int32(userPos.SortPosition),
			AnswerGroupPosition: userPos.AnswerGroupPosition,
			IsCorrect:           userPos.IsCorrect,
			Score:               userPos.Score,
		})
	}

	return result, nil
}

// GetMatchingAnswerExercise Matching: Lấy câu trả lời của người dùng và câu trả lời đúng (first_item_id = second_item_id)
func (r *getExerciseAnswersRepository) GetMatchingAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerMatching, error) {
	var answers []dto.ExerciseAnswerMatching

	err := db.ReplicaDB.Table("exercise_question_user_matchings").
		Select(`
			exercise_question_user_matchings.*
		`).
		Where("exercise_question_user_matchings.exercise_id = ? AND exercise_question_user_matchings.question_id = ? AND exercise_question_user_matchings.user_id = ?", exerciseID, questionID, userID).
		Find(&answers).Error
	return answers, err
}

// GetLabelingAnswerExercise Labeling: Lấy câu trả lời của người dùng và câu trả lời đúng (blank_id = answer_id)
func (r *getExerciseAnswersRepository) GetLabelingAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerLabeling, error) {
	var answers []dto.ExerciseAnswerLabeling
	err := db.ReplicaDB.Table("exercise_question_user_labelings").
		Select(`
			exercise_question_user_labelings.blank_id,
			exercise_question_user_labelings.answer_id,
			exercise_question_user_labelings.is_correct,
			exercise_question_user_labelings.score
		`).
		Where("exercise_question_user_labelings.exercise_id = ? AND exercise_question_user_labelings.question_id = ? AND exercise_question_user_labelings.user_id = ?", exerciseID, questionID, userID).
		Find(&answers).Error
	return answers, err
}

// GetGroupAnswerExercise Category: Lấy câu trả lời của người dùng và câu trả lời đúng (answer_id = group_id)
func (r *getExerciseAnswersRepository) GetGroupAnswerExercise(exerciseID, questionID, userID int64) ([]dto.ExerciseAnswerGroup, error) {
	var answers []dto.ExerciseAnswerGroup

	if err := db.ReplicaDB.Table("exercise_question_user_groups").
		Select(`
			exercise_question_user_groups.*
		`).
		Where("exercise_question_user_groups.exercise_id = ? AND exercise_question_user_groups.question_id = ? AND exercise_question_user_groups.user_id = ?", exerciseID, questionID, userID).
		Find(&answers).Error; err != nil {
		return nil, err
	}

	return answers, nil
}

// GetManualAnswerExercise Manual: Lấy câu trả lời của người dùng và điểm số (nếu đã chấm)
func (r *getExerciseAnswersRepository) GetManualAnswerExercise(exerciseID, questionID, userID int64) (*dto.ExerciseAnswerManual, error) {
	var answer dto.ExerciseAnswerManual
	err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").
		Select("answer, file_info as answer_file_info, score, is_scored").
		Where("exercise_id = ? AND question_id = ? AND user_id = ?", exerciseID, questionID, userID).
		First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetAllAndTrueAnswers trả về all_answer và true_answer cho câu hỏi multiple choice
func (r *getExerciseAnswersRepository) GetAllAndTrueAnswers(questionID int64) (allAnswers struct {
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

func (r *getExerciseAnswersRepository) GetStudentName(userID int64) (string, error) {
    var name string
    err := db.ReplicaDB.Table("users").Select("name").Where("id = ?", userID).Scan(&name).Error
    return name, err
}

func (r *getExerciseAnswersRepository) GetExerciseUserInfo(exerciseID, userID int64) (duration int64, totalScore float64, submittedAt int64, hasManualScoring bool, ratio float64, err error) {
	type result struct {
		Time             int64
		Score            float64
		CreatedAt        float64
		HasManualScoring bool
		Ratio            float64
	}
	var res result
    err = db.ReplicaDB.Table("exercise_users").
		Select("time, score, extract(epoch from created_at) as created_at, has_manual_scoring, COALESCE(ratio,0) as ratio").
        Where("exercise_id = ? AND user_id = ?", exerciseID, userID).
		Scan(&res).Error
	return res.Time, res.Score, int64(res.CreatedAt), res.HasManualScoring, res.Ratio, err
}

func (r *getExerciseAnswersRepository) GetTotalQuestions(exerciseID int64) (int, error) {
	// Sử dụng cloned_questions thay vì exercise_questions
	var raw struct{ Questions json.RawMessage }
	if err := db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", exerciseID, "exercise").
		First(&raw).Error; err != nil {
		return 0, err
	}
	
	var questions []interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return 0, err
	}
	
	return len(questions), nil
}

func (r *getExerciseAnswersRepository) GetUnscoredManualQuestions(exerciseID, userID int64) (int, error) {
	var count int64
	err := db.ReplicaDB.Table("exercise_question_user_manual_scoring").Where("exercise_id = ? AND user_id = ? AND is_scored = false", exerciseID, userID).Count(&count).Error
	return int(count), err
}

func (r *getExerciseAnswersRepository) GetExerciseComment(exerciseID, userID int64) (string, error) {
    var content string
    err := db.ReplicaDB.Table("exercise_comments").Select("content").Where("exercises_id = ? AND student_id = ?", exerciseID, userID).Order("created_at desc").Limit(1).Scan(&content).Error
    return content, err
}

