package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
)

type SettingRepository interface {
	base.BaseRepositoryInterface[models.Setting]
	FindByKey(key string) (*models.Setting, error)
}

type settingRepository struct {
	*base.BaseRepository[models.Setting]
}

func NewSettingRepository() SettingRepository {
	return &settingRepository{
		BaseRepository: base.NewBaseRepository[models.Setting](),
	}
}

func (r *settingRepository) FindByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	err := db.ReplicaDB.Where(`"key" = ?`, key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
