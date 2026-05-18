package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type DepartmentRepository interface {
	base.BaseRepositoryInterface[models.Department]
}

type departmentRepository struct {
	*base.BaseRepository[models.Department]
}

func NewDepartmentRepository() DepartmentRepository {
	return &departmentRepository{
		BaseRepository: base.NewBaseRepository[models.Department](),
	}
}
