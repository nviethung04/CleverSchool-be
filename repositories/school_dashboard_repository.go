package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
)

type SchoolDashboardRepository interface {
	GetAllWithStats() ([]dto.SchoolSummaryResponse, error)
}

type schoolDashboardRepository struct {

}

func NewSchoolDashboardRepository() SchoolDashboardRepository {
	return &schoolDashboardRepository{}
}

func (r *schoolDashboardRepository) GetAllWithStats() ([]dto.SchoolSummaryResponse, error) {
	var schools []dto.SchoolSummaryResponse

	err := db.ReplicaDB.Table("schools s").
		Select(`
			s.id,
			s.name,
			COUNT(DISTINCT CASE WHEN urr.role_id = 3 THEN u.id END) AS student_count,
			COUNT(DISTINCT CASE WHEN urr.role_id = 2 THEN u.id END) AS teacher_count,
			COUNT(DISTINCT c.id) AS classroom_count,
			COUNT(DISTINCT cs.course_id) AS course_count
		`).
		Joins("LEFT JOIN users u ON u.school_id = s.id").
		Joins("LEFT JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("LEFT JOIN classes c ON c.school_id = s.id").
		Joins("LEFT JOIN course_schools cs ON cs.school_id = s.id").
		Group("s.id, s.name").
		Order("s.name").
		Scan(&schools).Error

	return schools, err
}


