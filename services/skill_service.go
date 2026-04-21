package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"

	"github.com/gin-gonic/gin"
)

type SkillService interface {
	GetAll(c *gin.Context) ([]models.Skill, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Skill, error)
	Create(c *gin.Context, req *prot.SkillRequest) (*models.Skill, error)
	Update(c *gin.Context, req *prot.SkillRequest) (*models.Skill, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Skill, error)
}

type skillService struct {
	repo repositories.SkillRepository
}

func NewSkillService(repo repositories.SkillRepository) SkillService {
	return &skillService{repo: repo}
}

func (s *skillService) GetAll(c *gin.Context) ([]models.Skill, int64, error) {
	allowedFilters := []string{"type", "parent_id"} // ví dụ các filter có thể dùng
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	skills, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return skills, rows, nil
}

func (s *skillService) GetByID(c *gin.Context, id int) (*prot.Skill, error) {
	skill, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	skillResource := resources.NewSkillResource()
	formattedSkill := skillResource.FormatSkill(skill)

	return formattedSkill, nil
}

func (s *skillService) Create(c *gin.Context, req *prot.SkillRequest) (*models.Skill, error) {
	skillResource := resources.NewSkillResource()
	skill := skillResource.FormatModelSkill(req)

	s.repo.SetContext(c)

	err := s.repo.Create(skill)
	if err != nil {
		return nil, err
	}

	id := int(skill.ID)
	newSkill, _ := s.repo.FindNewByID(id)

	return newSkill, nil
}

func (s *skillService) Update(c *gin.Context, req *prot.SkillRequest) (*models.Skill, error) {
	skillResource := resources.NewSkillResource()
	skill := skillResource.FormatModelSkill(req)

	s.repo.SetContext(c)

	err := s.repo.Update(skill)
	if err != nil {
		return nil, err
	}

	id := int(skill.ID)
	updatedSkill, _ := s.repo.FindNewByID(id)

	return updatedSkill, nil
}

func (s *skillService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *skillService) Restore(c *gin.Context, id int) (*models.Skill, error) {
	s.repo.SetContext(c)
	skill, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return skill, nil
}
