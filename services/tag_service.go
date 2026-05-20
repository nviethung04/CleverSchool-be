package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type TagService interface {
	GetAll(c *gin.Context) ([]models.Tag, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Tag, error)
	Create(c *gin.Context, req *prot.TagRequest) (*models.Tag, error)
	Update(c *gin.Context, req *prot.TagRequest) (*models.Tag, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Tag, error)
}

type tagService struct {
	repo repositories.TagRepository
}

func NewTagService(repo repositories.TagRepository) TagService {
	return &tagService{repo: repo}
}

func (s *tagService) GetAll(c *gin.Context) ([]models.Tag, int64, error) {
	allowedFilters := []string{"type", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	tags, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return tags, rows, nil
}

func (s *tagService) GetByID(c *gin.Context, id int) (*prot.Tag, error) {
	tag, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	tagResource := resources.NewTagResource()
	formattedTag := tagResource.FormatTag(tag)

	return formattedTag, nil
}

func (s *tagService) Create(c *gin.Context, req *prot.TagRequest) (*models.Tag, error) {
	tagResource := resources.NewTagResource()
	tag := tagResource.FormatModelTag(req)

	s.repo.SetContext(c)

	err := s.repo.Create(tag)
	if err != nil {
		return nil, err
	}

	id := int(tag.ID)
	newTag, _ := s.repo.FindNewByID(id)

	return newTag, nil
}

func (s *tagService) Update(c *gin.Context, req *prot.TagRequest) (*models.Tag, error) {
	tagResource := resources.NewTagResource()
	tag := tagResource.FormatModelTag(req)

	s.repo.SetContext(c)

	err := s.repo.Update(tag)
	if err != nil {
		return nil, err
	}

	id := int(tag.ID)
	updatedTag, _ := s.repo.FindNewByID(id)

	return updatedTag, nil
}

func (s *tagService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *tagService) Restore(c *gin.Context, id int) (*models.Tag, error) {
	s.repo.SetContext(c)
	tag, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
