package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
)

type ContestResultRepository interface {
	GetContestRoundAnswers(contestRoundId int64, userId int64) (interface{}, error)
	GetContestRoundAnswersByStudent(contestRoundId int64) (interface{}, error)
	GetContestRoundResults(contestRoundId int64) (interface{}, error)
	GetContestRoundLeaderboard(contestRoundId int64) (interface{}, error)
}

type contestResultRepository struct{}

func NewContestResultRepository() ContestResultRepository {
	return &contestResultRepository{}
}

func (r *contestResultRepository) GetContestRoundAnswers(contestRoundId int64, userId int64) (interface{}, error) {
	// Get all answers for a specific user in a contest round
	var answers struct {
		MultipleChoice []models.ContestRoundQuestionUserMultipleChoice
		FillInBlank    []models.ContestRoundQuestionUserFillInBlank
		Ordering       []models.ContestRoundQuestionUserPosition
		Matching       []models.ContestRoundQuestionUserMatching
		Labeling       []models.ContestRoundQuestionUserLabeling
		// Category       []models.ContestRoundQuestionUserCategory // TODO: Implement when model is available
		ManualScoring []models.ContestRoundQuestionUserManualScoring
	}

	// Get multiple choice answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.MultipleChoice)

	// Get fill in blank answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.FillInBlank)

	// Get ordering answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.Ordering)

	// Get matching answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.Matching)

	// Get labeling answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.Labeling)

	// Get category answers - TODO: Implement when model is available
	// db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
	//	Find(&answers.Category)

	// Get manual scoring answers
	db.ReplicaDB.Where("contest_round_id = ? AND user_id = ?", contestRoundId, userId).
		Find(&answers.ManualScoring)

	return answers, nil
}

func (r *contestResultRepository) GetContestRoundAnswersByStudent(contestRoundId int64) (interface{}, error) {
	// Get answers for current student (from context)
	// This would get the current user from context and call GetContestRoundAnswers
	// For now, return a placeholder
	return map[string]interface{}{
		"message": "Get answers by student - implementation needed",
	}, nil
}

func (r *contestResultRepository) GetContestRoundResults(contestRoundId int64) (interface{}, error) {
	// Get all results for a contest round
	var results []models.ContestRoundUser

	err := db.ReplicaDB.
		Where("contest_round_id = ?", contestRoundId).
		Order("score DESC, time ASC").
		Find(&results).Error

	return results, err
}

func (r *contestResultRepository) GetContestRoundLeaderboard(contestRoundId int64) (interface{}, error) {
	// Get leaderboard for a contest round
	var leaderboard []struct {
		UserID   int64   `json:"user_id"`
		UserName string  `json:"user_name"`
		Score    float64 `json:"score"`
		Ratio    float64 `json:"ratio"`
		Time     int64   `json:"time"`
		Rank     int     `json:"rank"`
	}

	err := db.ReplicaDB.
		Table("contest_round_users cru").
		Select("cru.user_id, u.name as user_name, cru.score, cru.ratio, cru.time").
		Joins("JOIN users u ON u.id = cru.user_id").
		Where("cru.contest_round_id = ? AND u.deleted_at IS NULL", contestRoundId).
		Order("cru.score DESC, cru.time ASC").
		Find(&leaderboard).Error

	// Add rank
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	return leaderboard, err
}
