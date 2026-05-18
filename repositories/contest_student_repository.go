package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
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

	// Query với joiner logic mở rộng - học sinh có thể thấy contest round nếu:
	// 1. Được thêm trực tiếp (person joiner)
	// 2. Thuộc lớp được thêm (class joiner)
	// 3. Thuộc trường được thêm (school joiner)
	// 4. Thuộc tỉnh được thêm (province joiner)
	query := db.ReplicaDB.Table("contest_rounds cr").
		Select(`cr.id, cr.name, cr.description, 
			CAST(extract(epoch from cr.start_time) AS BIGINT) as start_time,
			CAST(extract(epoch from cr.end_time) AS BIGINT) as end_time,
			c.name as contest_name`).
		Joins("JOIN contests c ON c.id = cr.contest_id AND c.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = ?", userID).
		Where("cr.deleted_at IS NULL").
		Where(`(
			-- 1. Person joiner: học sinh được thêm trực tiếp
			EXISTS (
				SELECT 1 FROM contest_round_joiner_persons crjp 
				WHERE crjp.contest_round_id = cr.id AND crjp.user_id = ? AND crjp.deleted_at IS NULL
			) OR
			-- 2. Class joiner: học sinh thuộc lớp được thêm
			EXISTS (
				SELECT 1 FROM contest_round_joiner_classes crjc
				JOIN classes cl ON cl.id = crjc.class_id AND cl.deleted_at IS NULL
				JOIN user_classes uc ON uc.class_id = cl.id AND uc.user_id = ? AND uc.is_current = true
				WHERE crjc.contest_round_id = cr.id AND crjc.deleted_at IS NULL
			) OR
			-- 3. School joiner: học sinh thuộc trường được thêm
			EXISTS (
				SELECT 1 FROM contest_round_joiner_schools crjs
				JOIN schools s ON s.id = crjs.school_id AND s.deleted_at IS NULL
				WHERE crjs.contest_round_id = cr.id AND crjs.deleted_at IS NULL
				AND u.school_id = s.id
			)
			/* Province joiner disabled temporarily due to schema mismatch
			OR
			-- 4. Province joiner: học sinh thuộc tỉnh được thêm
			EXISTS (
				SELECT 1 FROM contest_round_joiner_provinces crjpr
				JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL
				JOIN wards w ON w.code = s.ward_code
				JOIN provinces p ON p.code = w.province_code
				WHERE crjpr.contest_round_id = cr.id AND crjpr.deleted_at IS NULL
				AND p.id = crjpr.province_id
			)
			*/
		)`, userID, userID).
		Order("cr.id DESC")

	// Count với joiner logic mở rộng
	countQuery := db.ReplicaDB.Table("contest_rounds cr").
		Joins("JOIN contests c ON c.id = cr.contest_id AND c.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = ?", userID).
		Where("cr.deleted_at IS NULL").
		Where(`(
			-- 1. Person joiner
			EXISTS (
				SELECT 1 FROM contest_round_joiner_persons crjp 
				WHERE crjp.contest_round_id = cr.id AND crjp.user_id = ? AND crjp.deleted_at IS NULL
			) OR
			-- 2. Class joiner
			EXISTS (
				SELECT 1 FROM contest_round_joiner_classes crjc
				JOIN classes cl ON cl.id = crjc.class_id AND cl.deleted_at IS NULL
				JOIN user_classes uc ON uc.class_id = cl.id AND uc.user_id = ? AND uc.is_current = true
				WHERE crjc.contest_round_id = cr.id AND crjc.deleted_at IS NULL
			) OR
			-- 3. School joiner
			EXISTS (
				SELECT 1 FROM contest_round_joiner_schools crjs
				JOIN schools s ON s.id = crjs.school_id AND s.deleted_at IS NULL
				WHERE crjs.contest_round_id = cr.id AND crjs.deleted_at IS NULL
				AND u.school_id = s.id
			)
			/* Province joiner disabled temporarily due to schema mismatch
			OR
			-- 4. Province joiner
			EXISTS (
				SELECT 1 FROM contest_round_joiner_provinces crjpr
				JOIN schools s ON s.id = u.school_id AND s.deleted_at IS NULL
				JOIN wards w ON w.code = s.ward_code
				JOIN provinces p ON p.code = w.province_code
				WHERE crjpr.contest_round_id = cr.id AND crjpr.deleted_at IS NULL
				AND p.id = crjpr.province_id
			)
			*/
		)`, userID, userID)

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
