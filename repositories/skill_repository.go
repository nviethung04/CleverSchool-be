package repositories

import (
	"be-lms/models"
	"be-lms/repositories/base"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SkillRepository interface {
	base.BaseRepositoryInterface[models.Skill]
	base.BeforeQueryHook
}

type skillRepository struct {
	*base.BaseRepository[models.Skill]
}

func NewSkillRepository() SkillRepository {
	repo := &skillRepository{
		BaseRepository: base.NewBaseRepository[models.Skill](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (s *skillRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	return query
}
