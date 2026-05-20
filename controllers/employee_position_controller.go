package controllers

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
)

type EmployeePositionController struct {
	*GenericController[models.EmployeePosition, prot.EmployeePosition, *prot.EmployeePosition]
}

func NewEmployeePositionController(service services.EmployeePositionService) *EmployeePositionController {
	employeePositionResource := resources.NewEmployeePositionResource()
	employeePositionResourceAdapter := NewEmployeePositionResourceAdapter(employeePositionResource)

	genericController := NewGenericController(
		service,
		employeePositionResourceAdapter,
		func() *prot.EmployeePosition {
			return &prot.EmployeePosition{}
		},
		func(employeePositions []*prot.EmployeePosition, totalCount uint64) interface{} {
			return &prot.EmployeePositionListResponse{
				EmployeePositions:       employeePositions,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &EmployeePositionController{
		GenericController: genericController,
	}
}

