package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SettingService interface {
	GetAll(c *gin.Context) ([]models.Setting, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Setting, error)
	GetByKey(key string) (*prot.Setting, error)
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
	allowedFilters := []string{"is_active"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, parseErr := strconv.Atoi(limitStr); parseErr == nil && limit > 0 {
			perPage = limit
		}
	}

	if activeStr := c.Query("is_active"); activeStr != "" {
		if active, parseErr := strconv.ParseBool(activeStr); parseErr == nil {
			filter["is_active"] = active
		}
	}

	s.repo.SetSearch(keyword, []string{"name", "key", "id"})
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
		return nil, err
	}
	return resources.NewSettingResource().FormatSetting(setting), nil
}

func (s *settingService) GetByKey(key string) (*prot.Setting, error) {
	setting, err := s.repo.FindByKey(key)
	if err != nil {
		return nil, err
	}
	return resources.NewSettingResource().FormatSetting(setting), nil
}

func (s *settingService) Create(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error) {
	settingResource := resources.NewSettingResource()
	setting := settingResource.FormatModelSetting(req)

	s.repo.SetContext(c)
	if err := s.repo.Create(setting); err != nil {
		return nil, err
	}

	newSetting, _ := s.repo.FindNewByID(int(setting.ID))
	return newSetting, nil
}

func (s *settingService) Update(c *gin.Context, req *prot.SettingRequest) (*models.Setting, error) {
	settingResource := resources.NewSettingResource()
	setting := settingResource.FormatModelSetting(req)

	s.repo.SetContext(c)
	if err := s.repo.Update(setting); err != nil {
		return nil, err
	}

	updatedSetting, _ := s.repo.FindNewByID(int(setting.ID))
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
		return nil, err
	}
	return setting, nil
}
