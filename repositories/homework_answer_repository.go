package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ManualAnswer struct {
	ID      string
	Type    string
	Source  *string
	Status  bool
	Content json.RawMessage
	Answer  *string
	FileURL *string
    Score   *float64
}

type HomeworkAnswerOverview struct {
	StudentName             string
	HomeworkName            string
	HomeworkDescription     string
	HomeworkStatus          int32
	HomeworkCoverImage      string
	TotalQuestions          int32
	QuestionsCompleted      int64
	LastQuestionIDCompleted int64
	ManualQuestionsCount    int32
	HomeworkQuestionForm    string
	HomeworkFiles           models.MediaInfos
	IsSubmitted             bool
	IsScored                bool
	Rate                    string
}

type HomeworkAnswerRepository interface {
	GetHomeworkAnswerOverview(homeworkID, userID int64) (HomeworkAnswerOverview, error)
	GetClonedQuestions(homeworkID int64) ([]map[string]interface{}, error)
	GetManualAnswers(homeworkID, userID int64) ([]ManualAnswer, error)
    GetHomeworkRatio(homeworkID, userID int64) (float64, error)
    GetLatestHomeworkComment(homeworkID, userID int64) (string, error)
	GetSubmitFiles(homeworkID, userID int64) (models.MediaInfos, error)
}

type homeworkAnswerRepository struct {
}

func NewHomeworkAnswerRepository() HomeworkAnswerRepository {
	return &homeworkAnswerRepository{}
}

func (r *homeworkAnswerRepository) GetHomeworkAnswerOverview(homeworkID, userID int64) (HomeworkAnswerOverview, error) {
	var overview HomeworkAnswerOverview
	var user models.User
	var hw models.Homework
	if err := db.MasterDB.Model(&models.User{}).Where("id = ?", userID).First(&user).Error; err != nil {
		return overview, err
	}
	if err := db.MasterDB.Model(&models.Homework{}).Where("id = ? AND deleted_at IS NULL", homeworkID).First(&hw).Error; err != nil {
		return overview, err
	}
	overview.StudentName = user.Name
	overview.HomeworkName = hw.Name
	overview.HomeworkDescription = hw.Description
	overview.HomeworkStatus = int32(hw.Status)
	overview.HomeworkCoverImage = hw.CoverImageInfo.Path // trả về string gốc
	overview.HomeworkQuestionForm = hw.QuestionForm
	overview.HomeworkFiles = hw.FileInfos
	// Lấy số câu hỏi từ cloned_questions
	var raw struct{ Questions json.RawMessage }
	if err := db.MasterDB.Table("cloned_questions").Select("questions").Where("assignment_id = ? AND assignment_type = ?", homeworkID, "homework").First(&raw).Error; err == nil {
		var arr []interface{}
		if err := json.Unmarshal(raw.Questions, &arr); err == nil {
			overview.TotalQuestions = int32(len(arr))
		}
	}
	// Lấy số câu đã hoàn thành và id câu cuối cùng đã hoàn thành
	var hwUser models.HomeworkUser
	db.MasterDB.Model(&models.HomeworkUser{}).
		Joins("INNER JOIN homeworks ON homeworks.id = homework_users.homework_id").
		Where("homework_users.homework_id = ? AND homework_users.user_id = ? AND homeworks.deleted_at IS NULL", homeworkID, userID).
		First(&hwUser)

	overview.QuestionsCompleted = hwUser.QuestionsCompleted
	overview.LastQuestionIDCompleted = hwUser.LastQuestionIDCompleted
	overview.IsSubmitted = hwUser.ID > 0
	overview.IsScored = hwUser.StatusScoring == 2
	overview.Rate = hwUser.Rate

	// Lấy số câu hỏi manual đã làm
	var manualCount int64
	db.MasterDB.Table("homework_question_user_manual_scoring").
		Joins("INNER JOIN homeworks ON homeworks.id = homework_question_user_manual_scoring.homework_id").
		Where("homework_question_user_manual_scoring.homework_id = ? AND homework_question_user_manual_scoring.user_id = ? AND homeworks.deleted_at IS NULL", homeworkID, userID).
		Distinct("homework_question_user_manual_scoring.question_id").
		Count(&manualCount)
	overview.ManualQuestionsCount = int32(manualCount)

	return overview, nil
}

func (r *homeworkAnswerRepository) GetClonedQuestions(homeworkID int64) ([]map[string]interface{}, error) {
	var raw struct{ Questions json.RawMessage }
	if err := db.MasterDB.Table("cloned_questions").
		Select("questions").
		Joins("INNER JOIN homeworks ON homeworks.id = cloned_questions.assignment_id").
		Where("cloned_questions.assignment_id = ? AND cloned_questions.assignment_type = ? AND homeworks.deleted_at IS NULL", homeworkID, "homework").
		First(&raw).Error; err != nil {
		return nil, err
	}
	var questions []map[string]interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *homeworkAnswerRepository) GetManualAnswers(homeworkID, userID int64) ([]ManualAnswer, error) {
	var results []ManualAnswer
    type ManualDB struct {
        QuestionID string
        Answer     *string
        MediaInfo  *models.MediaInfo
        Score      sql.NullFloat64
        ScoringAt  *time.Time
    }
	var dbAnswers []ManualDB
    if err := db.MasterDB.Table("homework_question_user_manual_scoring").
        Select("question_id, answer, file_info as media_info, score, scoring_at").
		Joins("INNER JOIN homeworks ON homeworks.id = homework_question_user_manual_scoring.homework_id").
		Where("homework_question_user_manual_scoring.homework_id = ? AND homework_question_user_manual_scoring.user_id = ? AND homeworks.deleted_at IS NULL", homeworkID, userID).
		Find(&dbAnswers).Error; err != nil {
		return nil, err
	}
	questions, err := r.GetClonedQuestions(homeworkID)
	if err != nil {
		return nil, err
	}
	for _, a := range dbAnswers {
		// Tìm thông tin câu hỏi tương ứng trong cloned_questions
		for _, q := range questions {
			// Convert id to string safely
			var questionID string
			if id, ok := q["id"].(string); ok {
				questionID = id
			} else if idFloat, ok := q["id"].(float64); ok {
				questionID = fmt.Sprintf("%.0f", idFloat)
			} else {
				continue // Skip if id is not string or float64
			}

			// Convert type to string safely
			var questionType string
			if qType, ok := q["type"].(string); ok {
				questionType = qType
			} else if qTypeFloat, ok := q["type"].(float64); ok {
				questionType = fmt.Sprintf("%.0f", qTypeFloat)
			} else {
				continue // Skip if type is not string or float64
			}

			if (questionType == "speaking" || questionType == "writing") && questionID == a.QuestionID {
				content, _ := json.Marshal(q["content"])
				var path *string

				if a.MediaInfo != nil && a.MediaInfo.Path != "" {
					p := a.MediaInfo.Path
					path = &p
				}
                var scored bool
                if a.ScoringAt != nil {
                    scored = true
                }
                var scorePtr *float64
                if a.Score.Valid {
                    v := a.Score.Float64
                    scorePtr = &v
                }
                results = append(results, ManualAnswer{
					ID:      a.QuestionID,
					Type:    questionType,
					Source:  nil,
                    Status:  scored,
					Content: content,
					Answer:  a.Answer,
                    FileURL: path,
                    Score:   scorePtr,
				})
				break
			}
		}
	}
	return results, nil
}

func (r *homeworkAnswerRepository) GetHomeworkRatio(homeworkID, userID int64) (float64, error) {
    var ratio float64
    if err := db.MasterDB.Table("homework_users").
        Select("COALESCE(homework_users.ratio,0)").
        Joins("INNER JOIN homeworks ON homeworks.id = homework_users.homework_id").
        Where("homework_users.homework_id = ? AND homework_users.user_id = ? AND homeworks.deleted_at IS NULL", homeworkID, userID).
        Scan(&ratio).Error; err != nil {
        return 0, err
    }
    return ratio, nil
}

func (r *homeworkAnswerRepository) GetLatestHomeworkComment(homeworkID, userID int64) (string, error) {
    var content string
    // Lấy comment mới nhất theo updated_at
    err := db.MasterDB.Table("homework_comments").
        Select("homework_comments.content").
        Joins("INNER JOIN homeworks ON homeworks.id = homework_comments.homework_id").
        Where("homework_comments.homework_id = ? AND homework_comments.student_id = ? AND homeworks.deleted_at IS NULL", homeworkID, userID).
        Order("homework_comments.updated_at DESC").
        Limit(1).
        Scan(&content).Error
    if err != nil {
        return "", err
    }
    return content, nil
}

func (r *homeworkAnswerRepository) GetSubmitFiles(homeworkID, userID int64) (models.MediaInfos, error) {
	var homeworkUser models.HomeworkUser

	db.ReplicaDB.
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		First(&homeworkUser)

	return homeworkUser.FileInfos, nil

}
