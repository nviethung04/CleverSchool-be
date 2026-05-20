package observer

import (
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"errors"

	"gorm.io/gorm"
)

type RoleObserver struct{}

func (u *RoleObserver) BeforeCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Role)
	if ok {
	}
	return nil
}

func (u *RoleObserver) AfterCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Role)
	if ok {
	}
	return nil
}

func (u *RoleObserver) BeforeUpdate(data interface{}, tx *gorm.DB) (err error) {
	role, ok := data.(*models.Role)
	if ok && role.ID != 0 {
		var currentRole models.Role
		err := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.Role{}).
			Where("id = ?", role.ID).
			First(&currentRole).Error

		if err != nil {
			return errors.New(i18n.Localize("messages.role_not_found"))
		}

		systemRoles := []string{"admin", "teacher", "student"}

		if currentRole.Name != "" {
			for _, systemRole := range systemRoles {
				if currentRole.Name == systemRole && role.Name != systemRole {
					return errors.New(i18n.Localize("messages.cannot_update_system_role"))
				}
			}
		}
	}
	return nil
}

func (u *RoleObserver) AfterUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Role)
	if ok {
	}
	return nil
}

func (u *RoleObserver) BeforeDelete(data interface{}, tx *gorm.DB) (err error) {
	role, ok := data.(*models.Role)
	if ok && role.ID != 0 {
		var currentRole models.Role
		err := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.Role{}).
			Where("id = ?", role.ID).
			First(&currentRole).Error

		if err != nil {
			return errors.New(i18n.Localize("messages.role_not_found"))
		}

		systemRoles := []string{"admin", "teacher", "student"}

		if currentRole.Name != "" {
			for _, systemRole := range systemRoles {
				if currentRole.Name == systemRole {
					return errors.New(i18n.Localize("messages.cannot_delete_system_role"))
				}
			}
		}
	}
	return nil
}

func (u *RoleObserver) AfterDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Role)
	if ok {
	}
	return nil
}

