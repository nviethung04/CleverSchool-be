package services

import (
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/redis"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"fmt"
	"time"

	"be-cleverschool/config"
	"github.com/gin-gonic/gin"
)

const (
	settingsCacheKeyPrefix = "settings:"
	settingsCacheTTL       = 24 * time.Hour
)

type SettingService interface {
	GetAll(c *gin.Context) ([]models.Setting, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Setting, error)
	GetByKey(c *gin.Context, key string) (*prot.Setting, error)
	Create(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error)
	Update(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Setting, error)
}

type settingService struct {
	repo repositories.SettingRepository
}

func NewSettingService(repo repositories.SettingRepository) SettingService {
	return &settingService{repo: repo}
}

func (s *settingService) GetAll(c *gin.Context) ([]models.Setting, int64, error) {
	allowedFilters := []string{"key", "status", "is_active"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	settings, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return settings, rows, nil
}

func (s *settingService) GetByID(c *gin.Context, id int) (*prot.Setting, error) {
	setting, err := s.repo.FindByID(id)

	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.id_invalid"))
	}

	settingResource := resources.NewSettingResource()
	formattedSetting := settingResource.FormatSetting(setting)

	return formattedSetting, nil
}

func (s *settingService) GetByKey(c *gin.Context, key string) (*prot.Setting, error) {
	cacheKey := settingsCacheKeyPrefix + key
	cached, err := redis.RememberCache[*prot.Setting](cacheKey, settingsCacheTTL, func() (*prot.Setting, error) {
		config.Log.Info("GetByKey", "key", key)
		setting, err := s.repo.FindByKey(key)
		if err != nil {
			return nil, fmt.Errorf(i18n.Localize("messages.no_records_found"))
		}
		if setting.IsInternal {
			return nil, fmt.Errorf(i18n.Localize("messages.is_internal_setting"))
		}
		settingResource := resources.NewSettingResource()
		return settingResource.FormatSetting(setting), nil
	})

	if err != nil {
		return nil, err
	}
	return cached, nil
}

func (s *settingService) Create(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error) {
	settingResource := resources.NewSettingResource()
	setting := settingResource.FormatModelSetting(req)

	isKeyExists, _ := s.repo.IsKeyExists(req.Key, 0)
	if isKeyExists || req.Key == "" {
		return nil, fmt.Errorf(i18n.Localize("messages.key_invalid"))
	}

	s.repo.SetContext(c)

	err := s.repo.Create(setting)
	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.create_data"))
	}

	id := int(setting.ID)
	newSetting, _ := s.repo.FindNewByID(id)

	return newSetting, nil
}

func (s *settingService) Update(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error) {
	settingResource := resources.NewSettingResource()
	setting := settingResource.FormatModelSetting(req)

	isKeyExists, _ := s.repo.IsKeyExists(req.Key, setting.ID)
	if isKeyExists || req.Key == "" {
		return nil, fmt.Errorf(i18n.Localize("messages.key_invalid"))
	}

	s.repo.SetContext(c)

	err := s.repo.Update(setting)
	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.error_update_data"))
	}

	_ = redis.DeleteCache(settingsCacheKeyPrefix + req.Key)

	id := int(setting.ID)
	updatedSetting, _ := s.repo.FindNewByID(id)

	return updatedSetting, nil
}

func (s *settingService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *settingService) Restore(c *gin.Context, id int) (*models.Setting, error) {
	s.repo.SetContext(c)
	setting, err := s.repo.Restore(id)
	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.error_restore_data"))
	}
	return setting, nil
}

