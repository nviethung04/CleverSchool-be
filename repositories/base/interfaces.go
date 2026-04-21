package base

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseRepositoryInterface[T any] interface {
	GetAll() ([]T, error)
	FindAll() ([]T, int64, error)
	FindByID(id int) (*T, error)
	FindNewByID(id int) (*T, error)
	Create(entity *T) error
	Update(entity *T) error
	Delete(id int) error
	Restore(id int) (*T, error)
	SetFilter(filter map[string]interface{})
	SetLimit(limit int)
	SetPage(page int)
	SetSort(sort map[string]string)
	SetPreload(preload []string)
	SetOmit(omit []string)
	SetEnableRedis(enable bool)
	SetRedisTime(time time.Duration)
	SetContext(c *gin.Context)
	SetSearch(value string, fields []string)
	ApplySearch(query *gorm.DB) *gorm.DB
	SetAlias(alias map[string]string)
	SetBeforeQueryHook(hook BeforeQueryHook)
	SetMeili(index string, filterKeys []string)
	PermanentlyDeleteOldRecords(before time.Time) error
	SyncMediaInfo() error
	SyncMediaInfoForRecentlyUpdatedMedia(within time.Duration) error
	SyncMediaInfoByID(id int64, infoFieldName string, dbColumnName string) error
	GetAdminSchoolId(ctx *gin.Context) int
}
