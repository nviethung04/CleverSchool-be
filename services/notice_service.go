package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NoticeService interface {
	GetAll(c *gin.Context) ([]models.Notice, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Notice, error)
	Create(c *gin.Context, req *prot.Notice) (*models.Notice, error)
	Update(c *gin.Context, req *prot.Notice) (*models.Notice, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Notice, error)
}

type noticeService struct {
	repo     repositories.NoticeRepository
	resource resources.NoticeResource
}

func NewNoticeService(repo repositories.NoticeRepository) NoticeService {
	return &noticeService{
		repo:     repo,
		resource: resources.NewNoticeResource(),
	}
}

func (s *noticeService) GetAll(c *gin.Context) ([]models.Notice, int64, error) {
	allowedFilters := []string{"status", "type"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{"NotificationLogs"})

	notices, rows, err := s.repo.FindAll()

	if err != nil {
		return nil, 0, err
	}

	return notices, rows, nil
}

func (s *noticeService) GetByID(c *gin.Context, id int) (*prot.Notice, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{"NotificationLogs"})
	notice, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("notice not found")
	}

	return s.resource.FormatNotice(notice), nil
}

func (s *noticeService) Create(c *gin.Context, req *prot.Notice) (*models.Notice, error) {
	notice := s.resource.FormatModelNotice(req)

	s.repo.SetContext(c)
	if err := s.repo.Create(notice); err != nil {
		return nil, fmt.Errorf("failed to create notice: %w", err)
	}

	id := int(notice.ID)
	s.repo.SetPreload([]string{"NotificationLogs"})
	newNotice, _ := s.repo.FindNewByID(id)

	return newNotice, nil
}

func (s *noticeService) Update(c *gin.Context, req *prot.Notice) (*models.Notice, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	s.repo.SetContext(c)

	notice := s.resource.FormatModelNotice(req)
	notice.ID = int64(id)

	if err := s.repo.Update(notice); err != nil {
		return nil, err
	}

	s.repo.SetPreload([]string{"NotificationLogs"})
	updateNotice, _ := s.repo.FindNewByID(id)

	return updateNotice, nil
}

func (s *noticeService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *noticeService) Restore(c *gin.Context, id int) (*models.Notice, error) {
	s.repo.SetContext(c)
	notice, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return notice, nil
}

