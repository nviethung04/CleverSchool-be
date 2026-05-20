package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
)

type SettingRepository interface {
	base.BaseRepositoryInterface[models.Setting]
	FindByKey(key string) (*models.Setting, error)
	IsKeyExists(key string, id int64) (bool, error)
}

type settingRepository struct {
	*base.BaseRepository[models.Setting]
}

func NewSettingRepository() SettingRepository {
	return &settingRepository{
		BaseRepository: base.NewBaseRepository[models.Setting](),
	}
}

func (s *settingRepository) FindByKey(key string) (*models.Setting, error) {
	var setting models.Setting
	err := db.ReplicaDB.Where("key = ?", key).First(&setting).Error
	return &setting, err
}

func (s *settingRepository) IsKeyExists(key string, id int64) (bool, error) {
	query := db.ReplicaDB.Model(&models.Setting{}).Where("key = ?", key)

	if id > 0 {
		query = query.Where("id <> ?", id)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

