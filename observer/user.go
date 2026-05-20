package observer

import (
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"errors"

	"gorm.io/gorm"
)

type UserObserver struct{}

func (u *UserObserver) BeforeCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.User)
	if ok {
	}
	return nil
}

func (u *UserObserver) AfterCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.User)
	if ok {
	}
	return nil
}

func (u *UserObserver) BeforeUpdate(data interface{}, tx *gorm.DB) (err error) {
	user, ok := data.(*models.User)
	if ok && user.ID != 0 {
		var currentUser models.User
		err := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.User{}).
			Where("id = ?", user.ID).
			First(&currentUser).Error

		if err != nil {
			return errors.New(i18n.Localize("messages.user_not_found"))
		}

		systemUsers := []string{"admin", "teacher", "student"}

		if currentUser.Username != "" {
			for _, systemUser := range systemUsers {
				if currentUser.Username == systemUser && user.Username != systemUser {
					return errors.New(i18n.Localize("messages.cannot_update_system_user"))
				}
			}
		}
	}
	return nil
}

func (u *UserObserver) AfterUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.User)
	if ok {
	}
	return nil
}

func (u *UserObserver) BeforeDelete(data interface{}, tx *gorm.DB) (err error) {
	user, ok := data.(*models.User)
	if ok && user.ID != 0 {
		var currentUser models.User
		err := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.User{}).
			Where("id = ?", user.ID).
			First(&currentUser).Error

		if err != nil {
			return errors.New(i18n.Localize("messages.user_not_found"))
		}

		systemUsers := []string{"admin", "teacher", "student"}

		if currentUser.Username != "" {
			for _, systemUser := range systemUsers {
				if currentUser.Username == systemUser {
					return errors.New(i18n.Localize("messages.cannot_delete_system_user"))
				}
			}
		}
	}
	return nil
}

func (u *UserObserver) AfterDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.User)
	if ok {
	}
	return nil
}

