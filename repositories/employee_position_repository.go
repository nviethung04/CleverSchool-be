package repositories

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
)

type EmployeePositionRepository interface {
	base.BaseRepositoryInterface[models.EmployeePosition]
}

type employeePositionRepository struct {
	*base.BaseRepository[models.EmployeePosition]
}

func NewEmployeePositionRepository() EmployeePositionRepository {
	return &employeePositionRepository{
		BaseRepository: base.NewBaseRepository[models.EmployeePosition](),
	}
}

