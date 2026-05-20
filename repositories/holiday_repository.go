package repositories

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
)

type HolidayRepository interface {
	base.BaseRepositoryInterface[models.Holiday]
}

type holidayWeekRepository struct {
	*base.BaseRepository[models.Holiday]
}

func NewHolidayRepository() HolidayRepository {
	return &holidayWeekRepository{
		BaseRepository: base.NewBaseRepository[models.Holiday](),
	}
}

