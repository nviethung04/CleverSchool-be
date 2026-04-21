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
	// Get all results for a contest round với query tối ưu
	var results []struct {
		ID                 int64   `json:"id"`
		ContestRoundID     int64   `json:"contest_round_id"`
		UserID             int64   `json:"user_id"`
		Score              float64 `json:"score"`
		Ratio              float64 `json:"ratio"`
		Time               int64   `json:"time"`
		HasManualScoring   bool    `json:"has_manual_scoring"`
		CreatedAt          string  `json:"created_at"`
		
		// User info
		UserName           string  `json:"user_name"`
		UserCode           string  `json:"user_code"`
		UserPhone          string  `json:"user_phone"`
		UserEmail          string  `json:"user_email"`
		
		// School info
		SchoolName         string  `json:"school_name"`
		
		// Address info
		ProvinceName       string  `json:"province_name"`
		
		// Contest info
		ContestRoundName   string  `json:"contest_round_name"`
		ContestName        string  `json:"contest_name"`
	}

	err := db.ReplicaDB.
		Table("contest_round_users cru").
		Select(`cru.id, cru.contest_round_id, cru.user_id, cru.score, cru.ratio, cru.time, 
			cru.has_manual_scoring, cru.created_at,
			u.name as user_name, u.code as user_code, u.phone_number as user_phone, u.email as user_email,
			s.name as school_name,
			COALESCE(p.name, '') as province_name,
			cr.name as contest_round_name,
			c.name as contest_name`).
		Joins("JOIN users u ON u.id = cru.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN wards w ON w.code = s.ward_code").
		Joins("LEFT JOIN provinces p ON p.code = w.province_code").
		Joins("JOIN contest_rounds cr ON cr.id = cru.contest_round_id AND cr.deleted_at IS NULL").
		Joins("JOIN contests c ON c.id = cr.contest_id AND c.deleted_at IS NULL").
		Where("cru.contest_round_id = ?", contestRoundId).
		Order("cru.score DESC, cru.time ASC").
		Find(&results).Error

	return results, err
}

func (r *contestResultRepository) GetContestRoundLeaderboard(contestRoundId int64) (interface{}, error) {
	// Get leaderboard for a contest round
	var leaderboard []struct {
		Rank       int     `json:"rank"`
		UserID     int64   `json:"user_id"`
		UserName   string  `json:"user_name"`
		UserCode   string  `json:"user_code"`
		SchoolName string  `json:"school_name"`
		Score      float64 `json:"score"`
		Ratio      float64 `json:"ratio"`
		Time       int64   `json:"time"`
	}

	err := db.ReplicaDB.
		Table("contest_round_users cru").
		Select(`cru.user_id, u.name as user_name, u.code as user_code, 
			COALESCE(s.name, '') as school_name, cru.score, cru.ratio, cru.time`).
		Joins("JOIN users u ON u.id = cru.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL").
		Where("cru.contest_round_id = ?", contestRoundId).
		Order("cru.score DESC, cru.time ASC").
		Find(&leaderboard).Error

	// Add rank
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	return leaderboard, err
}
