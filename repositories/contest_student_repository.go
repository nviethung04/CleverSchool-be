package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
)

type ContestStudentRepository interface {
	GetContestRoundsByStudent(userID int64, limit, offset int) ([]dto.GetContestRoundsByStudentDTO, int64, error)
}

type contestStudentRepository struct{}

func NewContestStudentRepository() ContestStudentRepository {
	return &contestStudentRepository{}
}

func (r *contestStudentRepository) GetContestRoundsByStudent(userID int64, limit, offset int) ([]dto.GetContestRoundsByStudentDTO, int64, error) {
	var result []dto.GetContestRoundsByStudentDTO
	var total int64

	// Query với joiner logic đúng theo migration (user_id không phải person_id)
	query := db.ReplicaDB.Table("contest_rounds cr").
		Select(`cr.id, cr.name, cr.description, 
			CAST(extract(epoch from cr.start_time) AS BIGINT) as start_time,
			CAST(extract(epoch from cr.end_time) AS BIGINT) as end_time,
			c.name as contest_name`).
		Joins("JOIN contests c ON c.id = cr.contest_id AND c.deleted_at IS NULL").
		Where("cr.deleted_at IS NULL").
		Where(`EXISTS (
			SELECT 1 FROM contest_round_joiner_persons crjp 
			WHERE crjp.contest_round_id = cr.id AND crjp.user_id = ? AND crjp.deleted_at IS NULL
		)`, userID).
		Order("cr.id DESC")

	// Count với joiner logic
	countQuery := db.ReplicaDB.Table("contest_rounds cr").
		Joins("JOIN contests c ON c.id = cr.contest_id AND c.deleted_at IS NULL").
		Where("cr.deleted_at IS NULL").
		Where(`EXISTS (
			SELECT 1 FROM contest_round_joiner_persons crjp 
			WHERE crjp.contest_round_id = cr.id AND crjp.user_id = ? AND crjp.deleted_at IS NULL
		)`, userID)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Scan(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
