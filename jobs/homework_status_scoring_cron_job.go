package jobs

import (
	"be-lms/config"
	"be-lms/database/db"
	"time"

	"github.com/robfig/cron/v3"
)

// StartHomeworkStatusScoringCronJob chạy job sync homework status scoring mỗi ngày
func StartHomeworkStatusScoringCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Chạy mỗi ngày lúc 2:00 AM
	c.AddFunc("0 2 * * *", func() {
		SyncHomeworkStatusScoringJob()
	})

	c.Start()
}

// SyncHomeworkStatusScoringJob sync status scoring cho tất cả homework_users
func SyncHomeworkStatusScoringJob() {
	// Defer recover để tránh panic
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in SyncHomeworkStatusScoringJob: %v", r)
		}
	}()

	config.Log.Info("Starting sync homework status scoring job")

	// Kiểm tra database connection
	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return
	}

	startTime := time.Now()

	// Sync status scoring
	err := syncHomeworkStatusScoring()
	if err != nil {
		config.Log.Errorf("Error syncing homework status scoring: %v", err)
		return
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished sync homework status scoring job in %v", duration)
}

// syncHomeworkStatusScoring thực hiện logic sync status scoring
func syncHomeworkStatusScoring() error {
	// Bước 1: Cập nhật status = 0 (không cần chấm) cho các homework_users
	// không có trong homework_question_user_manual_scoring
	err := updateNoManualScoringStatus()
	if err != nil {
		return err
	}

	// Bước 2: Cập nhật status = 2 (đã chấm xong) cho các homework_users
	// có trong homework_question_user_manual_scoring và tất cả đều is_scored = true
	err = updateCompletedScoringStatus()
	if err != nil {
		return err
	}

	// Bước 3: Cập nhật status = 1 (chưa chấm xong) cho các homework_users
	// có trong homework_question_user_manual_scoring và có ít nhất 1 is_scored = false
	err = updateIncompleteScoringStatus()
	if err != nil {
		return err
	}

	return nil
}

// updateNoManualScoringStatus cập nhật status = 0 cho homework không cần chấm
func updateNoManualScoringStatus() error {
	// Sử dụng GORM subquery để tìm homework_users không có trong homework_question_user_manual_scoring
	subQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = homework_users.homework_id").
		Where("homework_question_user_manual_scoring.user_id = homework_users.user_id")

	result := db.MasterDB.Table("homework_users").
		Where("NOT EXISTS (?)", subQuery).
		Update("status_scoring", 0)

	if result.Error != nil {
		config.Log.Errorf("Error updating no manual scoring status: %v", result.Error)
		return result.Error
	}

	config.Log.Infof("Updated %d homework_users to status_scoring = 0 (no manual scoring)", result.RowsAffected)
	return nil
}

// updateCompletedScoringStatus cập nhật status = 2 cho homework đã chấm xong
func updateCompletedScoringStatus() error {
	// Subquery 1: Kiểm tra có tồn tại trong homework_question_user_manual_scoring
	existsSubQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = homework_users.homework_id").
		Where("homework_question_user_manual_scoring.user_id = homework_users.user_id")

	// Subquery 2: Kiểm tra không có câu hỏi nào chưa chấm (is_scored = false)
	unscoredSubQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = homework_users.homework_id").
		Where("homework_question_user_manual_scoring.user_id = homework_users.user_id").
		Where("homework_question_user_manual_scoring.is_scored = ?", false)

	result := db.MasterDB.Table("homework_users").
		Where("EXISTS (?)", existsSubQuery).
		Where("NOT EXISTS (?)", unscoredSubQuery).
		Update("status_scoring", 2)

	if result.Error != nil {
		config.Log.Errorf("Error updating completed scoring status: %v", result.Error)
		return result.Error
	}

	config.Log.Infof("Updated %d homework_users to status_scoring = 2 (completed scoring)", result.RowsAffected)
	return nil
}

// updateIncompleteScoringStatus cập nhật status = 1 cho homework chưa chấm xong
func updateIncompleteScoringStatus() error {
	// Subquery: Kiểm tra có ít nhất 1 câu hỏi chưa chấm (is_scored = false)
	unscoredSubQuery := db.MasterDB.Table("homework_question_user_manual_scoring").
		Select("1").
		Where("homework_question_user_manual_scoring.homework_id = homework_users.homework_id").
		Where("homework_question_user_manual_scoring.user_id = homework_users.user_id").
		Where("homework_question_user_manual_scoring.is_scored = ?", false)

	result := db.MasterDB.Table("homework_users").
		Where("EXISTS (?)", unscoredSubQuery).
		Update("status_scoring", 1)

	if result.Error != nil {
		config.Log.Errorf("Error updating incomplete scoring status: %v", result.Error)
		return result.Error
	}

	config.Log.Infof("Updated %d homework_users to status_scoring = 1 (incomplete scoring)", result.RowsAffected)
	return nil
}
