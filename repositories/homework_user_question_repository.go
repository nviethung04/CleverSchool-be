package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"

	"gorm.io/gorm"
)

// HomeworkUserQuestionRepository chứa toàn bộ query tới bảng homework_user_questions
type HomeworkUserQuestionRepository interface {
	SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error
	CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error)
	GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID int64) (*models.HomeworkUserQuestion, error)
	DeleteByKey(homeworkID, userID, questionID, lessonID int64, tx *gorm.DB) error
	UpdateRatioScoreByKey(homeworkID, userID, questionID, lessonID int64, ratioScore float64, tx *gorm.DB) error
	GetStarByHomeworkUserLesson(homeworkID, userID, lessonID int64) (numerator float64, totalWeight float64, totalStar float64, err error)
	GetLatestByHomeworkUserQuestion(homeworkID, userID, lessonID, questionID int64) (*models.HomeworkUserQuestion, error)
	GetAllCorrectByHomeworkUser(homeworkID, userID int64) ([]models.HomeworkUserQuestion, error)
}

type homeworkUserQuestionRepository struct{}

func NewHomeworkUserQuestionRepository() HomeworkUserQuestionRepository {
	return &homeworkUserQuestionRepository{}
}

func (r *homeworkUserQuestionRepository) SaveHomeworkUserQuestion(record *models.HomeworkUserQuestion, tx *gorm.DB) error {
	// Đảm bảo dùng UTC và set thời gian tạo nếu chưa có
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	} else {
		record.CreatedAt = record.CreatedAt.UTC()
	}
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Create(record).Error
}

func (r *homeworkUserQuestionRepository) CountHomeworkUserQuestions(homeworkID, userID, questionID, lessonID int64) (int64, error) {
	var count int64
	err := db.MasterDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ?", homeworkID, userID, questionID, lessonID).
		Count(&count).Error
	return count, err
}

func (r *homeworkUserQuestionRepository) GetHomeworkUserQuestionByCorrect(homeworkID, userID, questionID, lessonID int64) (*models.HomeworkUserQuestion, error) {
	var records []models.HomeworkUserQuestion
	err := db.MasterDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ? AND is_all_correct = ?",
			homeworkID, userID, questionID, lessonID, true).
		Order("created_at DESC, id DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		// Chưa có bản ghi nào đúng hết
		return nil, nil
	}
	// Lấy bản ghi mới nhất
	return &records[0], nil
}

// DeleteByKey xóa tất cả bản ghi trùng bộ 4 khóa chính (homework_id, user_id, question_id, lesson_id)
func (r *homeworkUserQuestionRepository) DeleteByKey(homeworkID, userID, questionID, lessonID int64, tx *gorm.DB) error {
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ?",
		homeworkID, userID, questionID, lessonID).
		Delete(&models.HomeworkUserQuestion{}).Error
}

// UpdateRatioScoreByKey cập nhật ratio_score cho tất cả bản ghi có bộ 4 khóa trùng
func (r *homeworkUserQuestionRepository) UpdateRatioScoreByKey(homeworkID, userID, questionID, lessonID int64, ratioScore float64, tx *gorm.DB) error {
	if tx == nil {
		tx = db.MasterDB
	}
	return tx.Model(&models.HomeworkUserQuestion{}).
		Where("homework_id = ? AND user_id = ? AND question_id = ? AND lesson_id = ?",
			homeworkID, userID, questionID, lessonID).
		Update("ratio_score", ratioScore).Error
}

// GetStarByHomeworkUserLesson trả về:
// - numerator = SUM(ratio_score * weight)
// - totalWeight = SUM(weight)
// - totalStar = SUM(star)
// Chỉ tính các bản ghi có is_all_correct = true cho bộ 3 (homework_id, user_id, lesson_id).
// Lưu ý: filter ratio_score >= 0 nếu cần sẽ được xử lý ở tầng service khi tính ratio.
func (r *homeworkUserQuestionRepository) GetStarByHomeworkUserLesson(homeworkID, userID, lessonID int64) (float64, float64, float64, error) {
	var result struct {
		Numerator   float64
		TotalWeight float64
		TotalStar   float64
	}

	err := db.ReplicaDB.Table("homework_user_questions").
		Select("COALESCE(SUM(ratio_score * weight), 0) AS numerator, COALESCE(SUM(weight), 0) AS total_weight, COALESCE(SUM(star), 0) AS total_star").
		Where("homework_id = ? AND user_id = ? AND lesson_id = ? AND is_all_correct = ?",
			homeworkID, userID, lessonID, true).
		Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}

	return result.Numerator, result.TotalWeight, result.TotalStar, nil
}

// GetLatestByHomeworkUserQuestion lấy bản ghi mới nhất theo bộ 4 khóa (homework_id, user_id, lesson_id, question_id)
// Chỉ lấy các bản ghi có is_all_correct = true (theo yêu cầu API homework-answers)
func (r *homeworkUserQuestionRepository) GetLatestByHomeworkUserQuestion(homeworkID, userID, lessonID, questionID int64) (*models.HomeworkUserQuestion, error) {
	var record models.HomeworkUserQuestion
	err := db.ReplicaDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND lesson_id = ? AND question_id = ? AND is_all_correct = ?",
			homeworkID, userID, lessonID, questionID, true).
		Order("created_at DESC, id DESC").
		First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetAllCorrectByHomeworkUser lấy tất cả bản ghi có is_all_correct = true và ratio_score > 0
// theo homework_id và user_id (không cần lesson_id)
func (r *homeworkUserQuestionRepository) GetAllCorrectByHomeworkUser(homeworkID, userID int64) ([]models.HomeworkUserQuestion, error) {
	var records []models.HomeworkUserQuestion
	err := db.ReplicaDB.Table("homework_user_questions").
		Where("homework_id = ? AND user_id = ? AND is_all_correct = ? AND ratio_score > 0",
			homeworkID, userID, true).
		Order("question_id, created_at DESC, id DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}
