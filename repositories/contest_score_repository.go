package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ContestScoreRepository interface {
	SaveContestScoreMultipleChoice(req interface{}) (interface{}, error)
	SaveContestScoreFillInBlank(req interface{}) (interface{}, error)
	SaveContestScoreOrderingDragdrop(req interface{}) (interface{}, error)
	SaveContestScoreMatching(req interface{}) (interface{}, error)
	SaveContestScoreLabeling(req interface{}) (interface{}, error)
	SaveContestScoreCategory(req interface{}) (interface{}, error)
	SaveContestScoreManualScoring(req interface{}) (interface{}, error)
	SubmitContestRound(req interface{}) (interface{}, error)
	SkipContestQuestion(req interface{}) (interface{}, error)
	CheckSubmitContestRound(contestRoundId int64) (interface{}, error)
	UpdateOrCreateContestRoundUser(contestRoundUser *models.ContestRoundUser) error
	GetContestRoundByID(id int64) (*models.ContestRound, error)
	UpdateEvaluate(userId, contestRoundId int64, score float64) error
}

type contestScoreRepository struct{}

func NewContestScoreRepository() ContestScoreRepository {
	return &contestScoreRepository{}
}

// Note: Individual save methods not needed - use SaveScoreBulk API instead
func (r *contestScoreRepository) SaveContestScoreMultipleChoice(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreFillInBlank(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreOrderingDragdrop(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreMatching(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreLabeling(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreCategory(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SaveContestScoreManualScoring(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SubmitContestRound(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) SkipContestQuestion(req interface{}) (interface{}, error) {
	return nil, fmt.Errorf("use SaveScoreBulk API instead")
}

func (r *contestScoreRepository) CheckSubmitContestRound(contestRoundId int64) (interface{}, error) {
	// Check if contest round can be submitted
	// This would check if all questions are answered or skipped
	var totalQuestions int64
	var answeredQuestions int64
	var skippedQuestions int64

	// Count total questions
	db.ReplicaDB.Table("contest_round_questions").
		Where("contest_round_id = ?", contestRoundId).
		Count(&totalQuestions)

	// Count answered questions (example for multiple choice)
	db.ReplicaDB.Table("contest_round_question_user_multiple_choices").
		Where("contest_round_id = ?", contestRoundId).
		Count(&answeredQuestions)

	// Count skipped questions
	db.ReplicaDB.Table("contest_round_question_user_skips").
		Where("contest_round_id = ?", contestRoundId).
		Count(&skippedQuestions)

	canSubmit := (answeredQuestions + skippedQuestions) >= totalQuestions

	return map[string]interface{}{
		"can_submit":         canSubmit,
		"total_questions":    totalQuestions,
		"answered_questions": answeredQuestions,
		"skipped_questions":  skippedQuestions,
	}, nil
}

func (r *contestScoreRepository) UpdateOrCreateContestRoundUser(contestRoundUser *models.ContestRoundUser) error {
	if contestRoundUser == nil {
		return fmt.Errorf("contestRoundUser cannot be nil")
	}

	var existing models.ContestRoundUser
	err := db.ReplicaDB.
		Where("contest_round_id = ? AND user_id = ?",
			contestRoundUser.ContestRoundId, contestRoundUser.UserId).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Chưa có => tạo mới
			if err := db.MasterDB.Create(contestRoundUser).Error; err != nil {
				return fmt.Errorf("failed to create contest round user: %w", err)
			}
			return nil
		}
		// Lỗi khác
		return fmt.Errorf("failed to query contest round user: %w", err)
	}

	// Có rồi => update FileInfos
	err = db.ReplicaDB.Model(&existing).
		Updates(map[string]interface{}{
			"file_infos":  contestRoundUser.FileInfos,
			"has_manual_scoring":  contestRoundUser.HasManualScoring,
			"score":       contestRoundUser.Score,
			"ratio":       contestRoundUser.Ratio,
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update contest round user: %w", err)
	}

	return nil
}

func (r *contestScoreRepository) GetContestRoundByID(id int64) (*models.ContestRound, error) {
	var contestRound models.ContestRound

	query := db.ReplicaDB.Model(&models.ContestRound{})

	err := query.Where("id = ?", id).
		First(&contestRound).Error
	if err != nil {
		return nil, err
	}

	return &contestRound, nil
}

func (r *contestScoreRepository) UpdateEvaluate(userId, contestRoundId int64, score float64) error {
    result := db.MasterDB.Table("contest_round_users").
        Where("user_id = ? AND contest_round_id = ?", userId, contestRoundId).
        Update("ratio", score)

    if result.Error != nil {
		config.Log.Errorf("Error updating evaluate: %v", result.Error)
        return result.Error
    }

    if result.RowsAffected == 0 {
		return errors.New("contest round user not found")
    }

    return nil
}
