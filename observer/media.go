package observer

import (
	"be-cleverschool/models"

	"gorm.io/gorm"
)

type MediaObserver struct{}

func (u *MediaObserver) BeforeCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

func (u *MediaObserver) AfterCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

func (u *MediaObserver) BeforeUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

func (u *MediaObserver) AfterUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

func (u *MediaObserver) BeforeDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

func (u *MediaObserver) AfterDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Media)
	if ok {
	}
	return nil
}

