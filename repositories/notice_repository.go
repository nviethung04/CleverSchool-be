package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NoticeRepository interface {
	base.BaseRepositoryInterface[models.Notice]
	base.BeforeQueryHook
}

type noticeRepository struct {
	*base.BaseRepository[models.Notice]
}

func NewNoticeRepository() NoticeRepository {
	repo := &noticeRepository{
		BaseRepository: base.NewBaseRepository[models.Notice](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *noticeRepository) BeforeQuery(query *gorm.DB, c *gin.Context) *gorm.DB {
	return query
}
