package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"encoding/json"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

type HomeworkUserQuestionService interface {
	SaveHomeworkUserQuestion(homeworkID, userID, questionID, lessonID int64, isAllCorrect bool, tx *gorm.DB) (star int, ratioScore float64, weight float64, numberTimeSent int, err error)
	GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID int64) (*models.HomeworkUserQuestion, error)
	GetWeightForQuestion(homeworkID, questionID int64) (float64, error)
}

type homeworkUserQuestionService struct {
	repo                  repositories.HomeworkUserQuestionRepository
	clonedQuestionService ClonedQuestionService
}

func NewHomeworkUserQuestionService(repo repositories.HomeworkUserQuestionRepository, clonedQuestionService ClonedQuestionService) HomeworkUserQuestionService {
	return &homeworkUserQuestionService{
		repo:                  repo,
		clonedQuestionService: clonedQuestionService,
	}
}

// GetHomeworkUserQuestionByCorrect là hàm dùng chung cho nhiều dạng câu hỏi
func (s *homeworkUserQuestionService) GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID int64) (*models.HomeworkUserQuestion, error) {
	return s.repo.GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID)
}

// GetWeightForQuestion trả về weight theo logic dùng chung (lấy attributes rồi chuẩn hóa weight < 1 thành 1)
func (s *homeworkUserQuestionService) GetWeightForQuestion(homeworkID, questionID int64) (float64, error) {
	weight, err := s.getWeightFromAttributes(homeworkID, questionID)
	if err != nil {
		return 0, err
	}
	if weight < 1 {
		weight = 1
	}
	return weight, nil
}

func (s *homeworkUserQuestionService) getWeightFromAttributes(homeworkID, questionID int64) (float64, error) {
	// Query lại từ database để lấy toàn bộ JSON của question
	var raw struct {
		Questions json.RawMessage `gorm:"column:questions"`
	}
	err := db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", homeworkID, "homework").
		First(&raw).Error
	if err != nil {
		return 0, fmt.Errorf("Lỗi lấy cloned_question: %v", err)
	}

	var questions []map[string]interface{}
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return 0, fmt.Errorf("Lỗi parse questions: %v", err)
	}

	// Tìm question theo ID
	questionIDStr := strconv.FormatInt(questionID, 10)
	for _, q := range questions {
		if qID, ok := q["id"].(string); ok && qID == questionIDStr {
			// Parse attributes
			attributes, ok := q["attributes"].([]interface{})
			if !ok || len(attributes) == 0 {
				return 0, nil
			}

			totalWeight := 0.0
			count := 0
			for _, attrInterface := range attributes {
				attr, ok := attrInterface.(map[string]interface{})
				if !ok {
					continue
				}
				value, ok := attr["value"].(map[string]interface{})
				if !ok {
					continue
				}
				if weight, ok := value["weight"].(float64); ok {
					totalWeight += weight
					count++
				}
			}

			if count == 0 {
				return 0, nil
			}
			return totalWeight / float64(count), nil
		}
	}

	return 0, fmt.Errorf("Không tìm thấy câu hỏi với ID %d", questionID)
}

func (s *homeworkUserQuestionService) SaveHomeworkUserQuestion(homeworkID, userID, questionID, lessonID int64, isAllCorrect bool, tx *gorm.DB) (int, float64, float64, int, error) {
	// Kiểm tra xem đã tồn tại bản ghi is_all_correct = true chưa
	existingCorrectRecord, err := s.repo.GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID)
	if err == nil && existingCorrectRecord != nil {
		// Đã tồn tại bản ghi đúng hết rồi, trả về giá trị cũ, không cần tính toán hay lưu nữa
		return existingCorrectRecord.Star, existingCorrectRecord.RatioScore, existingCorrectRecord.Weight, existingCorrectRecord.NumberTimeSent, nil
	}

	// Đếm số lần đã nộp
	existCount, err := s.repo.CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Lỗi đếm số lần nộp: %w", err)
	}

	numberTimeSent := int(existCount) + 1

	// Tính star: chỉ khi đúng hết
	star := 0
	if isAllCorrect {
		star = 5 - (numberTimeSent - 1)
		if star < 1 {
			star = 1
		}
	}

	// Tính RatioScore: P = 100 - (A - 1) * 5, tối thiểu 50, chỉ khi đúng hết
	var ratioScore float64
	if isAllCorrect {
		ratioScore = 100.0 - float64(numberTimeSent-1)*5.0
		if ratioScore < 50 {
			ratioScore = 50
		}
	}

	// Lấy weight từ attributes (trung bình cộng)
	weight, err := s.getWeightFromAttributes(homeworkID, questionID)
	if err != nil {
		// Nếu không lấy được weight, để mặc định 0
		weight = 0
	}
	// Nếu weight < 1 (bao gồm cả 0) thì chuẩn hóa thành 1
	if weight < 1 {
		weight = 1
	}

	logRecord := &models.HomeworkUserQuestion{
		HomeworkID:     homeworkID,
		LessonID:       lessonID,
		UserID:         userID,
		QuestionID:     questionID,
		RatioScore:     ratioScore,
		IsAllCorrect:   isAllCorrect,
		Star:           star,
		NumberOptions:  0, // Luôn lưu là 0
		NumberTimeSent: numberTimeSent,
		Weight:         weight,
	}

	err = s.repo.SaveHomeworkUserQuestion(logRecord, tx)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return star, ratioScore, weight, numberTimeSent, nil
}

