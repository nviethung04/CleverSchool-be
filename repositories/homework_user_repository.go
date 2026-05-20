package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type HomeworkUserRepository interface {
	GetByHomeworkAndUser(homeworkID, userID int64) (*models.HomeworkUser, error)
	GetByHomeworkUserAndLesson(homeworkID, userID, lessonID int64) (*models.HomeworkUser, error)
	Create(homeworkUser *models.HomeworkUser) error
	Update(homeworkUser *models.HomeworkUser) error
	UpdateLastQuestionCompleted(homeworkID, userID, questionID int64) error
	SaveHomeworkUser(homeworkUser *models.HomeworkUser) error
	SumScoreFillInBlank(homeworkID, userID int64) (float64, error)
	SumScoreGroup(homeworkID, userID int64) (float64, error)
	SumScoreLabeling(homeworkID, userID int64) (float64, error)
	SumScoreManual(homeworkID, userID int64) (float64, error)
	SumScoreMatching(homeworkID, userID int64) (float64, error)
	SumScorePosition(homeworkID, userID int64) (float64, error)
	SumScoreUser(homeworkID, userID int64) (float64, error)
	UpdateHomeworkUserScore(homeworkID, userID int64, score float64) error
	UpdateHomeworkUserStatusScoring(homeworkID, userID int64) error
	GetCorrectCountByTable(table string, homeworkID, userID int64) (map[int64]int, error)
	GetManualScoringByHomework(homeworkID, userID int64) (map[int64]float64, error)
	GetAnswerCountByTable(table string, homeworkID, userID int64) (map[int64]int, error)
	UpdateHomeworkUserRatio(homeworkID, userID int64, ratio float64) error
	UpdateHomeworkUserExp(homeworkID, userID int64, exp float64) error
	UpdateHomeworkUserStar(homeworkID, userID int64, star int64) error
	UpdateHomeworkUserHasManualScoring(homeworkID, userID int64, hasManualScoring bool) error
	GetQuestionsCompleted(homeworkID, userID int64) (int64, error)
	UpdateOrCreate(homeworkUser *models.HomeworkUser) error
}

type homeworkUserRepository struct {
	db *gorm.DB
}

func NewHomeworkUserRepository() HomeworkUserRepository {
	return &homeworkUserRepository{}
}

func (r *homeworkUserRepository) GetByHomeworkAndUser(homeworkID, userID int64) (*models.HomeworkUser, error) {
	var homeworkUser models.HomeworkUser
	err := db.MasterDB.Where("homework_id = ? AND user_id = ?", homeworkID, userID).First(&homeworkUser).Error
	if err != nil {
		return nil, err
	}
	return &homeworkUser, nil
}

func (r *homeworkUserRepository) GetByHomeworkUserAndLesson(homeworkID, userID, lessonID int64) (*models.HomeworkUser, error) {
	var homeworkUser models.HomeworkUser
	err := db.MasterDB.Where("homework_id = ? AND user_id = ? AND lesson_id = ?", homeworkID, userID, lessonID).First(&homeworkUser).Error
	if err != nil {
		return nil, err
	}
	return &homeworkUser, nil
}

func (r *homeworkUserRepository) Create(homeworkUser *models.HomeworkUser) error {
	return db.MasterDB.Create(homeworkUser).Error
}

func (r *homeworkUserRepository) Update(homeworkUser *models.HomeworkUser) error {
	return db.MasterDB.Save(homeworkUser).Error
}

func (r *homeworkUserRepository) UpdateLastQuestionCompleted(homeworkID, userID, questionID int64) error {
	return db.MasterDB.Model(&models.HomeworkUser{}).
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Update("last_question_id_completed", questionID).Error
}

func (r *homeworkUserRepository) SaveHomeworkUser(homeworkUser *models.HomeworkUser) error {
	// Kiểm tra xem đã có bản ghi cho bài kiểm tra này chưa
	var existingHomeworkUser models.HomeworkUser
	err := db.MasterDB.Where("homework_id = ? AND user_id = ?", homeworkUser.HomeworkID, homeworkUser.UserID).First(&existingHomeworkUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Nếu đã có bản ghi thì cập nhật
	if existingHomeworkUser.ID != 0 {
		return db.MasterDB.Model(&existingHomeworkUser).Updates(map[string]interface{}{
			"score":      homeworkUser.Score,
			"updated_at": time.Now().UTC(),
		}).Error
	}

	// Nếu chưa có bản ghi thì tạo mới
	now := time.Now().UTC()
	homeworkUser.CreatedAt = now
	homeworkUser.UpdatedAt = now
	return db.MasterDB.Create(homeworkUser).Error
}

func (r *homeworkUserRepository) SumScoreFillInBlank(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_fill_in_blanks").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScoreGroup(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_groups").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScoreLabeling(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_labelings").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScoreManual(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScoreMatching(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_matchings").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScorePosition(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_user_positions").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) SumScoreUser(homeworkID, userID int64) (float64, error) {
	var sum float64
	err := db.MasterDB.Table("homework_question_users").
		Select("COALESCE(SUM(score),0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&sum).Error
	return sum, err
}

func (r *homeworkUserRepository) UpdateHomeworkUserScore(homeworkID, userID int64, score float64) error {
	dbConn := db.MasterDB
	var homeworkUserID int64
	err := dbConn.Table("homework_users").
		Select("id").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&homeworkUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if homeworkUserID > 0 {
		return dbConn.Table("homework_users").
			Where("id = ?", homeworkUserID).
			Updates(map[string]interface{}{
				"score": score,
			}).Error
	} else {
		return dbConn.Table("homework_users").
			Create(map[string]interface{}{
				"homework_id": homeworkID,
				"user_id":     userID,
				"score":       score,
				"created_at":  now,
				"updated_at":  now,
			}).Error
	}
}

func (r *homeworkUserRepository) UpdateHomeworkUserStatusScoring(homeworkID, userID int64) error {
	// Kiểm tra xem có câu hỏi nào cần chấm thủ công không
	existsSubQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = ?", homeworkID).
		Where("homework_question_user_manual_scoring.user_id = ?", userID)

	// Kiểm tra xem có câu hỏi nào chưa chấm xong không
	unscoredSubQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = ?", homeworkID).
		Where("homework_question_user_manual_scoring.user_id = ?", userID).
		Where("homework_question_user_manual_scoring.is_scored = ?", false)

	// Xác định status_scoring
	var statusScoring int16 = 0 // Mặc định: không cần chấm thủ công

	// Kiểm tra có tồn tại manual scoring không
	var exists bool
	if err := db.MasterDB.Raw("SELECT EXISTS (?)", existsSubQuery).Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		// Kiểm tra có câu hỏi chưa chấm xong không
		var hasUnscored bool
		if err := db.MasterDB.Raw("SELECT EXISTS (?)", unscoredSubQuery).Scan(&hasUnscored).Error; err != nil {
			return err
		}

		if hasUnscored {
			statusScoring = 1 // Chưa chấm xong
		} else {
			statusScoring = 2 // Đã chấm xong
		}
	}

	// Cập nhật status_scoring
	return db.MasterDB.Table("homework_users").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Updates(map[string]interface{}{
			"status_scoring": statusScoring,
		}).Error
}

//Xử lý chấm điểm theo tỉ lệ

func (r *homeworkUserRepository) GetCorrectCountByTable(table string, homeworkID, userID int64) (map[int64]int, error) {
	result := make(map[int64]int)
	rows, err := db.MasterDB.Table(table).
		Select("question_id, COUNT(*)").
		Where("homework_id = ? AND user_id = ? AND is_correct = true", homeworkID, userID).
		Group("question_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var cnt int
		if err := rows.Scan(&qid, &cnt); err == nil {
			result[qid] = cnt
		}
	}
	return result, nil
}

func (r *homeworkUserRepository) GetAnswerCountByTable(table string, homeworkID, userID int64) (map[int64]int, error) {
	result := make(map[int64]int)
	rows, err := db.MasterDB.Table(table).
		Select("question_id, COUNT(*)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Group("question_id").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var cnt int
		if err := rows.Scan(&qid, &cnt); err == nil {
			result[qid] = cnt
		}
	}
	return result, nil
}

func (r *homeworkUserRepository) GetManualScoringByHomework(homeworkID, userID int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	rows, err := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("question_id, score").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var qid int64
		var score float64
		if err := rows.Scan(&qid, &score); err == nil {
			result[qid] = score
		}
	}
	return result, nil
}

func (r *homeworkUserRepository) UpdateHomeworkUserRatio(homeworkID, userID int64, ratio float64) error {
	dbConn := db.MasterDB
	var homeworkUserID int64
	err := dbConn.Table("homework_users").
		Select("id").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&homeworkUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if homeworkUserID > 0 {
		return dbConn.Table("homework_users").
			Where("id = ?", homeworkUserID).
			Updates(map[string]interface{}{
				"ratio": ratio,
			}).Error
	} else {
		return dbConn.Table("homework_users").
			Create(map[string]interface{}{
				"homework_id": homeworkID,
				"user_id":     userID,
				"ratio":       ratio,
				"created_at":  now,
				"updated_at":  now,
			}).Error
	}
}

func (r *homeworkUserRepository) GetQuestionsCompleted(homeworkID, userID int64) (int64, error) {
	var completed int64
	err := db.MasterDB.Table("homework_users").
		Select("COALESCE(questions_completed,0)").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&completed).Error
	return completed, err
}

func (r *homeworkUserRepository) UpdateHomeworkUserExp(homeworkID, userID int64, exp float64) error {
	dbConn := db.MasterDB
	var homeworkUserID int64
	err := dbConn.Table("homework_users").
		Select("id").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&homeworkUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if homeworkUserID > 0 {
		// Dùng UpdateColumn để không tự động cập nhật updated_at
		return dbConn.Table("homework_users").
			Where("id = ?", homeworkUserID).
			UpdateColumn("exp", exp).Error
	} else {
		return dbConn.Table("homework_users").
			Create(map[string]interface{}{
				"homework_id": homeworkID,
				"user_id":     userID,
				"exp":         exp,
				"created_at":  now,
				"updated_at":  now,
			}).Error
	}
}

func (r *homeworkUserRepository) UpdateHomeworkUserStar(homeworkID, userID int64, star int64) error {
	dbConn := db.MasterDB
	var homeworkUserID int64
	err := dbConn.Table("homework_users").
		Select("id").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&homeworkUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if homeworkUserID > 0 {
		// Dùng UpdateColumn để không tự động cập nhật updated_at
		return dbConn.Table("homework_users").
			Where("id = ?", homeworkUserID).
			UpdateColumn("star", star).Error
	} else {
		return dbConn.Table("homework_users").
			Create(map[string]interface{}{
				"homework_id": homeworkID,
				"user_id":     userID,
				"star":        star,
				"created_at":  now,
				"updated_at":  now,
			}).Error
	}
}

func (r *homeworkUserRepository) UpdateHomeworkUserHasManualScoring(homeworkID, userID int64, hasManualScoring bool) error {
	dbConn := db.MasterDB
	var homeworkUserID int64
	err := dbConn.Table("homework_users").
		Select("id").
		Where("homework_id = ? AND user_id = ?", homeworkID, userID).
		Scan(&homeworkUserID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	now := time.Now().UTC()
	if homeworkUserID > 0 {
		return dbConn.Table("homework_users").
			Where("id = ?", homeworkUserID).
			Updates(map[string]interface{}{
				"has_manual_scoring": hasManualScoring,
				"updated_at":         now,
			}).Error
	} else {
		return dbConn.Table("homework_users").
			Create(map[string]interface{}{
				"homework_id":        homeworkID,
				"user_id":            userID,
				"has_manual_scoring": hasManualScoring,
				"created_at":         now,
				"updated_at":         now,
			}).Error
	}
}

func (r *homeworkUserRepository) UpdateOrCreate(homeworkUser *models.HomeworkUser) error {
	if homeworkUser == nil {
		return fmt.Errorf("homeworkUser cannot be nil")
	}

	var existing models.HomeworkUser
	err := db.MasterDB.
		Where("homework_id = ? AND lesson_id = ? AND user_id = ?",
			homeworkUser.HomeworkID, homeworkUser.LessonID, homeworkUser.UserID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Chưa có => tạo mới
			if err := db.MasterDB.Create(homeworkUser).Error; err != nil {
				return fmt.Errorf("failed to create homework user: %w", err)
			}
			return nil
		}
		// Lỗi khác
		return fmt.Errorf("failed to query homework user: %w", err)
	}

	// Có rồi => update FileInfos
	err = db.MasterDB.Model(&existing).
		Updates(map[string]interface{}{
			"file_infos":         homeworkUser.FileInfos,
			"has_manual_scoring": homeworkUser.HasManualScoring,
			"score":              homeworkUser.Score,
			"ratio":              homeworkUser.Ratio,
			"updated_at":         time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update homework user: %w", err)
	}

	return nil
}

