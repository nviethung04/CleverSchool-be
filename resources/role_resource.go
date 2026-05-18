package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
)

type RoleResource interface {
	FormatRole(role *models.Role) *prot.Role
	FormatRoles(roles []*models.Role) []*prot.Role
	FormatModelRole(role *prot.Role) *models.Role
}

type RoleResourceImpl struct{}

func NewRoleResource() RoleResource {
	return &RoleResourceImpl{}
}

func (r *RoleResourceImpl) FormatRole(role *models.Role) *prot.Role {
	if role == nil {
		return nil
	}

	return &prot.Role{
		Id:        int64(role.ID),
		Name:      role.Name,
		DefaultPageView: role.DefaultPageView,
		DefaultPageId: role.DefaultPageID,
		Status:    role.Status,
		CreatedAt: role.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: role.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *RoleResourceImpl) FormatRoles(roles []*models.Role) []*prot.Role {
	result := make([]*prot.Role, 0, len(roles))
	for _, s := range roles {
		if formatted := r.FormatRole(s); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *RoleResourceImpl) FormatModelRole(role *prot.Role) *models.Role {
	if role == nil {
		return nil
	}

	defaultPageView := role.DefaultPageView
	if role.DefaultPageView == "" {
		defaultPageView = models.PageAdmin
	}

	defaultPageId := role.DefaultPageId
	if role.DefaultPageId == 0 {
		defaultPageId = models.AdminRoleId
	}

	return &models.Role{
		ID:     int64(role.Id),
		Name:   role.Name,
		Status: role.Status,
		DefaultPageView: defaultPageView,
		DefaultPageID: defaultPageId,
	}
}
