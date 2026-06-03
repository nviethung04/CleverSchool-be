package observer

import (
	"be-lms/models"

	"gorm.io/gorm"
)

type QuestionObserver struct{}

func (u *QuestionObserver) BeforeCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {
		// q.MediaInfo = models.SyncMediaInfoByPath(tx.Session(&gorm.Session{NewDB: true}), q.FileUrl)
	}
	return nil
}

func (u *QuestionObserver) AfterCreate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {

	}
	return nil
}

func (u *QuestionObserver) BeforeUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {
		// q.MediaInfo = models.SyncMediaInfoByPath(tx.Session(&gorm.Session{NewDB: true}), q.FileUrl)
	}
	return nil
}

func (u *QuestionObserver) AfterUpdate(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {
	}
	return nil
}

func (u *QuestionObserver) BeforeDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {
	}
	return nil
}

func (u *QuestionObserver) AfterDelete(data interface{}, tx *gorm.DB) (err error) {
	_, ok := data.(*models.Question)
	if ok {
	}
	return nil
}
