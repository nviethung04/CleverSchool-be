package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
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
