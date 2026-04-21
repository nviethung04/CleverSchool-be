package observer

import "gorm.io/gorm"

type Observer interface {
	BeforeCreate(data interface{}, tx *gorm.DB) error
	AfterCreate(data interface{}, tx *gorm.DB) error
	BeforeUpdate(data interface{}, tx *gorm.DB) error
	AfterUpdate(data interface{}, tx *gorm.DB) error
	BeforeDelete(data interface{}, tx *gorm.DB) error
	AfterDelete(data interface{}, tx *gorm.DB) error
}
