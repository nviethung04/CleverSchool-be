package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type TopicService interface {
	GetAll(c *gin.Context) ([]models.Topic, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Topic, error)
	Create(c *gin.Context, req *prot.TopicRequest) (*models.Topic, error)
	Update(c *gin.Context, req *prot.TopicRequest) (*models.Topic, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Topic, error)
}

type topicService struct {
	repo repositories.TopicRepository
}

func NewTopicService(repo repositories.TopicRepository) TopicService {
	return &topicService{repo: repo}
}

func (s *topicService) GetAll(c *gin.Context) ([]models.Topic, int64, error) {
	allowedFilters := []string{"type", "parent_id", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	topics, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return topics, rows, nil
}

func (s *topicService) GetByID(c *gin.Context, id int) (*prot.Topic, error) {
	topic, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	topicResource := resources.NewTopicResource()
	formattedTopic := topicResource.FormatTopic(topic)

	return formattedTopic, nil
}

func (s *topicService) Create(c *gin.Context, req *prot.TopicRequest) (*models.Topic, error) {
	topicResource := resources.NewTopicResource()
	topic := topicResource.FormatModelTopic(req)

	s.repo.SetContext(c)

	err := s.repo.Create(topic)
	if err != nil {
		return nil, err
	}

	id := int(topic.ID)
	newTopic, _ := s.repo.FindNewByID(id)

	return newTopic, nil
}

func (s *topicService) Update(c *gin.Context, req *prot.TopicRequest) (*models.Topic, error) {
	topicResource := resources.NewTopicResource()
	topic := topicResource.FormatModelTopic(req)

	s.repo.SetContext(c)

	err := s.repo.Update(topic)
	if err != nil {
		return nil, err
	}

	id := int(topic.ID)
	updatedTopic, _ := s.repo.FindNewByID(id)

	return updatedTopic, nil
}

func (s *topicService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *topicService) Restore(c *gin.Context, id int) (*models.Topic, error) {
	s.repo.SetContext(c)
	topic, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return topic, nil
}
