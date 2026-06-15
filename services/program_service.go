package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProgramService interface {
	GetAll(c *gin.Context) ([]models.Program, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Program, error)
	Create(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error)
	Update(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Program, error)
}

type programService struct {
	repo repositories.ProgramRepository
}

func NewProgramService(repo repositories.ProgramRepository) ProgramService {
	return &programService{repo: repo}
}

func (s *programService) GetAll(c *gin.Context) ([]models.Program, int64, error) {
	allowedFilters := []string{}
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
	s.repo.SetPreload([]string{
		"Courses",
	})

	programs, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return programs, rows, nil
}

func (s *programService) GetByID(c *gin.Context, id int) (*prot.Program, error) {
	s.repo.SetPreload([]string{
		"Courses",
		"Chapters",
		"Chapters.Lessons",
	})
	s.repo.SetContext(c)
	program, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	programResource := resources.NewProgramResource()
	formattedProgram := programResource.FormatProgram(program)

	return formattedProgram, nil
}

func (s *programService) Create(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error) {
	programResource := resources.NewProgramResource()
	program := programResource.FormatModelProgram(req)

	s.repo.SetContext(c)

	err := s.repo.Create(program)
	if err != nil {
		return nil, err
	}

	id := int(program.ID)
	s.repo.SetPreload([]string{
		"Courses",
		"Chapters",
		"Chapters.Lessons",
	})
	newProgram, _ := s.repo.FindNewByID(id)

	return newProgram, nil
}

func (s *programService) Update(c *gin.Context, req *prot.ProgramRequest) (*models.Program, error) {
	s.repo.SetContext(c)

	existing, err := s.repo.FindByID(int(req.Id))
	if err != nil {
		return nil, err
	}

	programResource := resources.NewProgramResource()
	program := programResource.FormatModelProgram(req)

	if req.SubjectId == 0 {
		program.SubjectId = existing.SubjectId
	}
	if req.Image == "" || strings.HasPrefix(req.Image, "blob:") {
		program.ImageInfo = existing.ImageInfo
	}

	err = s.repo.Update(program)
	if err != nil {
		return nil, err
	}

	s.repo.ChapterUpdateSortPosition(program.ID)

	id := int(program.ID)
	s.repo.SetPreload([]string{
		"Courses",
		"Chapters",
		"Chapters.Lessons",
	})
	updatedProgram, _ := s.repo.FindNewByID(id)

	return updatedProgram, nil
}

func (s *programService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)

	err := s.repo.DeleteProgram(id)

	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *programService) Restore(c *gin.Context, id int) (*models.Program, error) {
	s.repo.SetContext(c)
	program, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return program, nil
}
