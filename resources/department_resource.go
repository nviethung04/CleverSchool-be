package resources

import (
	"be-lms/models"
	"be-lms/prot"
)

type DepartmentResource interface {
	FormatDepartment(department *models.Department) *prot.Department
	FormatDepartments(departments []*models.Department) []*prot.Department
	FormatModelDepartment(department *prot.Department) *models.Department
}

type DepartmentResourceImpl struct{}

func NewDepartmentResource() DepartmentResource {
	return &DepartmentResourceImpl{}
}

func (r *DepartmentResourceImpl) FormatDepartment(department *models.Department) *prot.Department {
	if department == nil {
		return nil
	}

	return &prot.Department{
		Id:          department.ID,
		Name:        department.Name,
		Description: department.Description,
		Status:      department.Status,
		CreatedAt:   department.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   department.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *DepartmentResourceImpl) FormatDepartments(departments []*models.Department) []*prot.Department {
	result := make([]*prot.Department, 0, len(departments))
	for _, u := range departments {
		if formatted := r.FormatDepartment(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *DepartmentResourceImpl) FormatModelDepartment(department *prot.Department) *models.Department {
	if department == nil {
		return nil
	}

	return &models.Department{
		ID:          department.Id,
		Name:        department.Name,
		Description: department.Description,
		Status:      department.Status,
	}
}
