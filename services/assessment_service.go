package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"fmt"
	"sort"
	"strconv"
	"time"

	"be-lms/i18n"

	"github.com/gin-gonic/gin"
)

type AssessmentService interface {
	GetAll(c *gin.Context) ([]models.Assessment, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Assessment, error)
	Create(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error)
	Update(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Assessment, error)
	Assigned(c *gin.Context, id int) (*prot.AssignedAssessment, error)
	AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error)
}

type assessmentService struct {
	repo repositories.AssessmentRepository
}

func NewAssessmentService(repo repositories.AssessmentRepository) AssessmentService {
	return &assessmentService{repo: repo}
}

func (s *assessmentService) GetAll(c *gin.Context) ([]models.Assessment, int64, error) {
	allowedFilters := []string{"type", "subject_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, parseErr := strconv.Atoi(limitStr); parseErr == nil && limit > 0 {
			perPage = limit
		}
	}

	s.repo.SetSearch(keyword, []string{"name", "id", "description"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{"Subject"})

	assessments, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return assessments, rows, nil
}

func (s *assessmentService) GetByID(c *gin.Context, id int) (*prot.Assessment, error) {
	s.repo.SetPreload([]string{"Subject"})
	assessment, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	assessmentResource := resources.NewAssessmentResource()
	return assessmentResource.FormatAssessment(assessment), nil
}

func (s *assessmentService) Create(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error) {
	assessmentResource := resources.NewAssessmentResource()
	assessment := assessmentResource.FormatModelAssessment(req)

	s.repo.SetContext(c)

	if err := s.repo.Create(assessment); err != nil {
		return nil, err
	}

	newAssessment, _ := s.repo.FindNewByID(int(assessment.ID))
	return newAssessment, nil
}

func (s *assessmentService) Update(c *gin.Context, req *prot.AssessmentRequest) (*models.Assessment, error) {
	assessmentResource := resources.NewAssessmentResource()
	assessment := assessmentResource.FormatModelAssessment(req)

	s.repo.SetContext(c)

	if err := s.repo.Update(assessment); err != nil {
		return nil, err
	}

	updatedAssessment, _ := s.repo.FindNewByID(int(assessment.ID))
	return updatedAssessment, nil
}

func (s *assessmentService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *assessmentService) Restore(c *gin.Context, id int) (*models.Assessment, error) {
	s.repo.SetContext(c)
	assessment, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}
	return assessment, nil
}

func (s *assessmentService) Assigned(c *gin.Context, id int) (*prot.AssignedAssessment, error) {
	req, err, _ := utils.GetBody[*prot.AssignedAssessment](c, func() *prot.AssignedAssessment {
		return &prot.AssignedAssessment{}
	})
	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}
	if req.CourseId == 0 || req.LessonId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	ref := models.AssessmentRefLesson{
		AssessmentId: int64(id),
		LessonId:     req.LessonId,
		CourseId:     req.CourseId,
	}
	if req.IsAssigned {
		now := time.Now()
		ref.AssignedAt = &now
		ref.AssignedBy = utils.PtrInt64(int64(utils.GetCurrentUserId(c)))
	} else {
		ref.AssignedAt = nil
		ref.AssignedBy = nil
	}

	if err := s.repo.Assigned(ref); err != nil {
		return req, fmt.Errorf(i18n.Localize("messages.no_record_update"))
	}
	return req, nil
}

func (s *assessmentService) AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error) {
	assessment, err := s.repo.AssignedLesson(int64(id))
	if err != nil {
		return nil, err
	}

	var lessons []*prot.AssignedLesson
	for _, lesson := range assessment.Lessons {
		var isAssigned bool
		for _, ref := range assessment.AssessmentRefLessons {
			if ref.LessonId == lesson.ID {
				isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
				if isAssigned {
					break
				}
			}
		}
		lessons = append(lessons, &prot.AssignedLesson{
			Id:          lesson.ID,
			Title:       lesson.Title,
			Description: lesson.Description,
			ObjectTitle: lesson.ObjectTitle,
			IsAssigned:  isAssigned,
		})
	}

	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].IsAssigned != lessons[j].IsAssigned {
			return lessons[i].IsAssigned && !lessons[j].IsAssigned
		}
		return lessons[i].Id < lessons[j].Id
	})
	return &prot.AssignedLessons{Lessons: lessons}, nil
}
