package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type DepartmentController struct {
	*GenericController[models.Department, prot.Department, *prot.Department]
}

func NewDepartmentController(service services.DepartmentService) *DepartmentController {
	departmentResource := resources.NewDepartmentResource()
	departmentResourceAdapter := NewDepartmentResourceAdapter(departmentResource)

	genericController := NewGenericController(
		service,
		departmentResourceAdapter,
		func() *prot.Department {
			return &prot.Department{}
		},
		func(departments []*prot.Department, totalCount uint64) interface{} {
			return &prot.DepartmentListResponse{
				Departments:       departments,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &DepartmentController{
		GenericController: genericController,
	}
}

