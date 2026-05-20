package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
)

type EmployeePositionResource interface {
	FormatEmployeePosition(employeePosition *models.EmployeePosition) *prot.EmployeePosition
	FormatEmployeePositions(employeePositions []*models.EmployeePosition) []*prot.EmployeePosition
	FormatModelEmployeePosition(employeePosition *prot.EmployeePosition) *models.EmployeePosition
}

type EmployeePositionResourceImpl struct{}

func NewEmployeePositionResource() EmployeePositionResource {
	return &EmployeePositionResourceImpl{}
}

func (r *EmployeePositionResourceImpl) FormatEmployeePosition(employeePosition *models.EmployeePosition) *prot.EmployeePosition {
	if employeePosition == nil {
		return nil
	}

	return &prot.EmployeePosition{
		Id:          employeePosition.ID,
		Name:        employeePosition.Name,
		Level:       employeePosition.Level,
		Description: employeePosition.Description,
		Status:      employeePosition.Status,
		CreatedAt:   employeePosition.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   employeePosition.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *EmployeePositionResourceImpl) FormatEmployeePositions(employeePositions []*models.EmployeePosition) []*prot.EmployeePosition {
	result := make([]*prot.EmployeePosition, 0, len(employeePositions))
	for _, u := range employeePositions {
		if formatted := r.FormatEmployeePosition(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *EmployeePositionResourceImpl) FormatModelEmployeePosition(employeePosition *prot.EmployeePosition) *models.EmployeePosition {
	if employeePosition == nil {
		return nil
	}

	return &models.EmployeePosition{
		ID:          employeePosition.Id,
		Name:        employeePosition.Name,
		Level:       employeePosition.Level,
		Description: employeePosition.Description,
		Status:      employeePosition.Status,
	}
}
