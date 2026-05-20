package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

type TeachingPlanService interface {
	GetAll(c *gin.Context) ([]models.TeachingPlan, int64, error)
	GetByID(c *gin.Context, id int) (*prot.TeachingPlan, error)
	Create(c *gin.Context, req *prot.TeachingPlanRequest) (*models.TeachingPlan, error)
	Update(c *gin.Context, req *prot.TeachingPlanRequest) (*models.TeachingPlan, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.TeachingPlan, error)
	Approve(c *gin.Context, req *prot.TeachingPlanApproveRequest) (*models.TeachingPlan, error)
}

type teachingPlanService struct {
	repo repositories.TeachingPlanRepository
}

func NewTeachingPlanService(repo repositories.TeachingPlanRepository) TeachingPlanService {
	return &teachingPlanService{repo: repo}
}

func (s *teachingPlanService) GetAll(c *gin.Context) ([]models.TeachingPlan, int64, error) {
	allowedFilters := []string{"status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"title", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Lessons",
	})

	teachingPlans, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return teachingPlans, rows, nil
}

func (s *teachingPlanService) GetByID(c *gin.Context, id int) (*prot.TeachingPlan, error) {
	s.repo.SetPreload([]string{
		"Lessons",
	})
	teachingPlan, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	teachingPlanResource := resources.NewTeachingPlanResource()
	formattedTeachingPlan := teachingPlanResource.FormatTeachingPlan(teachingPlan)

	return formattedTeachingPlan, nil
}

func (s *teachingPlanService) Create(c *gin.Context, req *prot.TeachingPlanRequest) (*models.TeachingPlan, error) {
	teachingPlanResource := resources.NewTeachingPlanResource()
	teachingPlan := teachingPlanResource.FormatModelTeachingPlan(req)

	s.repo.SetContext(c)

	err := s.repo.Create(teachingPlan)
	if err != nil {
		return nil, err
	}

	s.StoreLesson(teachingPlan.ID, req.Lesson)

	id := int(teachingPlan.ID)
	s.repo.SetPreload([]string{
		"Lessons",
	})
	newTeachingPlan, _ := s.repo.FindNewByID(id)

	return newTeachingPlan, nil
}

func (s *teachingPlanService) Update(c *gin.Context, req *prot.TeachingPlanRequest) (*models.TeachingPlan, error) {
	teachingPlanResource := resources.NewTeachingPlanResource()
	teachingPlan := teachingPlanResource.FormatModelTeachingPlan(req)

	s.repo.SetContext(c)
	s.repo.SetOmit([]string{
		"approved_by",
		"approved_at",
		"approved_note",
		"status",
	})

	err := s.repo.Update(teachingPlan)
	if err != nil {
		return nil, err
	}

	s.StoreLesson(teachingPlan.ID, req.Lesson)

	id := int(teachingPlan.ID)
	s.repo.SetPreload([]string{
		"Lessons",
	})
	updatedTeachingPlan, _ := s.repo.FindNewByID(id)

	return updatedTeachingPlan, nil
}

func (s *teachingPlanService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *teachingPlanService) Restore(c *gin.Context, id int) (*models.TeachingPlan, error) {
	s.repo.SetContext(c)
	teachingPlan, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return teachingPlan, nil
}

func (s *teachingPlanService) Approve(c *gin.Context, req *prot.TeachingPlanApproveRequest) (*models.TeachingPlan, error) {
	teachingPlan, _ := s.repo.FindByID(int(req.Id))

	if teachingPlan.ID == 0 {
		return nil, errors.New("teaching plan not found")
	}

	if req.Status {
		teachingPlan.ApprovedBy = int64(utils.GetCurrentUserId(c))
		teachingPlan.ApprovedAt = time.Now()
		teachingPlan.ApprovedNote = req.ApprovedNote
		teachingPlan.Status = true
	} else {
		teachingPlan.ApprovedBy = 0
		teachingPlan.ApprovedAt = time.Time{}
		teachingPlan.ApprovedNote = ""
		teachingPlan.Status = false
	}

	s.repo.SetContext(c)
	err := s.repo.Update(teachingPlan)
	if err != nil {
		return nil, err
	}

	id := int(teachingPlan.ID)
	s.repo.SetPreload([]string{
		"Lessons",
	})
	updatedTeachingPlan, _ := s.repo.FindNewByID(id)

	return updatedTeachingPlan, nil
}

func (s *teachingPlanService) StoreLesson(id int64, lesson *prot.TeachingPlanLessonInfo) error {
	var lessonIds []int64

	if lesson != nil && lesson.Id > 0 {
		lessonIds = append(lessonIds, lesson.Id)
	}

	s.repo.UpdateLessons(id, lessonIds)

	return nil
}

