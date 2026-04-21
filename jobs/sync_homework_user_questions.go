package jobs

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories"
	"time"

	"gorm.io/gorm"
)

// SyncHomeworkUserQuestionsJob đồng bộ dữ liệu từ các bảng homework_question_user_*
// sang bảng homework_user_questions theo rule:
// - Bộ (homework_id, user_id, question_id, lesson_id) là duy nhất
// - Nếu đã tồn tại trong homework_user_questions thì bỏ qua
// - Mặc định:
//   + ratio_score = 100
//   + is_all_correct = true
//   + star = 5
//   + number_options = 0
//   + number_time_sent = 1
//   + weight = 1
//   + created_at = created_at của record nguồn
// - Riêng bảng homework_question_user_manual_scoring:
//   + Nếu is_scored = true => ratio_score = score * 10
//   + Nếu is_scored = false => ratio_score = -1
// - Có thể truyền startDate, endDate để chỉ sync các record có created_at trong khoảng đó
//   (nếu nil sẽ bỏ qua điều kiện tương ứng)
func SyncHomeworkUserQuestionsJob(startDate, endDate *time.Time) (totalProcessed int64, totalInserted int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in SyncHomeworkUserQuestionsJob: %v", r)
		}
	}()

	config.Log.Info("Starting SyncHomeworkUserQuestionsJob")

	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return 0, 0, nil
	}

	startTime := time.Now()

	repo := repositories.NewHomeworkUserQuestionRepository()

	// Đồng bộ theo từng bảng
	if p, i, e := syncFromHomeworkQuestionUsers(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_users: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserFillInBlanks(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_fill_in_blanks: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserGroups(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_groups: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserLabelings(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_labelings: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserMatchings(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_matchings: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserPositions(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_positions: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	if p, i, e := syncFromHomeworkQuestionUserManualScoring(repo, startDate, endDate); e != nil {
		config.Log.Errorf("Error syncing from homework_question_user_manual_scoring: %v", e)
		return totalProcessed, totalInserted, e
	} else {
		totalProcessed += p
		totalInserted += i
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished SyncHomeworkUserQuestionsJob in %v, processed=%d, inserted=%d", duration, totalProcessed, totalInserted)

	return totalProcessed, totalInserted, nil
}

// --- Helpers cho từng bảng nguồn ---

func defaultHomeworkUserQuestionRecord(
	homeworkID, userID, questionID, lessonID int64,
	createdAt time.Time,
	ratioScore float64,
) *models.HomeworkUserQuestion {
	return &models.HomeworkUserQuestion{
		HomeworkID:     homeworkID,
		UserID:         userID,
		QuestionID:     questionID,
		LessonID:       lessonID,
		RatioScore:     ratioScore,
		IsAllCorrect:   true,
		Star:           5,
		NumberOptions:  0,
		NumberTimeSent: 1,
		Weight:         1,
		CreatedAt:      createdAt,
	}
}

// applyDateRangeFilter thêm điều kiện created_at theo startDate/endDate nếu có
func applyDateRangeFilter(q *gorm.DB, startDate, endDate *time.Time) *gorm.DB {
	if startDate != nil {
		q = q.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("created_at <= ?", *endDate)
	}
	return q
}

func syncFromHomeworkQuestionUsers(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUser
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserFillInBlanks(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserFillInBlank
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserGroups(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserGroup
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserLabelings(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserLabeling
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserMatchings(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserMatching
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserPositions(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserPosition
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			row.CreatedAt,
			100,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}

func syncFromHomeworkQuestionUserManualScoring(repo repositories.HomeworkUserQuestionRepository, startDate, endDate *time.Time) (int64, int64, error) {
	var rows []models.HomeworkQuestionUserManualScoring
	q := db.MasterDB
	q = applyDateRangeFilter(q, startDate, endDate)
	if err := q.Find(&rows).Error; err != nil {
		return 0, 0, err
	}

	var processed, inserted int64
	for _, row := range rows {
		processed++

		count, err := repo.CountHomeworkUserQuestions(row.HomeworkID, row.UserID, row.QuestionID, row.LessonID)
		if err != nil {
			return processed, inserted, err
		}
		if count > 0 {
			continue
		}

		var createdAt time.Time
		if row.CreatedAt != nil {
			createdAt = *row.CreatedAt
		} else {
			createdAt = time.Now()
		}

		var ratioScore float64
		if row.IsScored {
			if row.Score != nil {
				ratioScore = (*row.Score) * 10
			} else {
				ratioScore = 0
			}
		} else {
			ratioScore = -1
		}

		record := defaultHomeworkUserQuestionRecord(
			row.HomeworkID,
			row.UserID,
			row.QuestionID,
			row.LessonID,
			createdAt,
			ratioScore,
		)

		if err := repo.SaveHomeworkUserQuestion(record, nil); err != nil {
			return processed, inserted, err
		}
		inserted++
	}

	return processed, inserted, nil
}


