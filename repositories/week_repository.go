package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"time"
)

type WeekRepository interface {
	base.BaseRepositoryInterface[models.Week]
	GetByDate(date time.Time) (*models.Week, error)
	FindWeeksByDateRange(start, end time.Time) ([]models.Week, error)
}

type weekRepository struct {
	*base.BaseRepository[models.Week]
}

func NewWeekRepository() WeekRepository {
	return &weekRepository{
		BaseRepository: base.NewBaseRepository[models.Week](),
	}
}

func (r *weekRepository) GetByDate(date time.Time) (*models.Week, error) {
	var week models.Week
	err := db.ReplicaDB.
		Where("start_date <= ? AND end_date >= ?", date, date).
		First(&week).Error

	if err != nil {
		return nil, err
	}
	return &week, nil
}

func (r *weekRepository) FindWeeksByDateRange(start, end time.Time) ([]models.Week, error) {
	var weeks []models.Week

	// Lấy tuần có khả năng giao với range
	err := db.ReplicaDB.
		Where("start_date <= ? AND end_date >= ?", end, start).
		Order("start_date asc").
		Find(&weeks).Error
	if err != nil {
		return nil, err
	}

	var result []models.Week
	for _, w := range weeks {
		result = append(result, w)

		// // Tính overlap
		// overlapStart := maxTime(w.StartDate, start)
		// overlapEnd := minTime(w.EndDate, end)

		// // Nếu overlap hợp lệ
		// if overlapEnd.After(overlapStart) {
		// 	days := int(overlapEnd.Sub(overlapStart).Hours()/24) + 1
		// 	if days >= 4 {
		// 		result = append(result, w)
		// 	}
		// }
	}

	return result, nil
}

// Helper
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

