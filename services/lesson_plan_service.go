package services

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"be-cleverschool/resources"

	"be-cleverschool/utils"

	"strconv"

	"github.com/gin-gonic/gin"
)

type LessonPlanService interface {
	GetAll(c *gin.Context) ([]models.LessonPlan, int64, error)
	GetByID(c *gin.Context, id int) (*prot.LessonPlan, error)
	Create(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error)
	Update(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.LessonPlan, error)
	Complete(c *gin.Context, id int, req *prot.LessonPlanCompleteRequest) error
}

type lessonPlanService struct {
	repo repositories.LessonPlanRepository
}

func NewLessonPlanService(r repositories.LessonPlanRepository) LessonPlanService {
	return &lessonPlanService{repo: r}
}

func (s *lessonPlanService) GetAll(c *gin.Context) ([]models.LessonPlan, int64, error) {
	var req requests.GetLessonPlanRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	allowedFilters := []string{}
	_, _, _, _, sort, err := utils.ParsePaginationParams(c, allowedFilters)

	if err != nil {
		return nil, 0, err
	}

	lessonPlans, total, err := s.repo.GetAllWithPaging(&req, sort, c)
	if err != nil {
		return nil, 0, err
	}

	return lessonPlans, total, nil
}

func (s *lessonPlanService) GetByID(c *gin.Context, id int) (*prot.LessonPlan, error) {
	lp, err := s.repo.GetByID(id, c)
	if err != nil {
		return nil, err
	}

	resource := resources.NewLessonPlanResource()
	return resource.FormatLessonPlan(lp), nil
}

func (s *lessonPlanService) Create(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error) {
	coverImageUrl := utils.StripDomain(req.CoverImage, models.Storage)
	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(coverImageUrl, models.Storage)

	userId := utils.GetCurrentUserId(c)

	var authorId int64
	if req.AuthorId > 0 {
		authorId = int64(req.AuthorId)
	} else if req.Author != nil && req.Author.Id != 0 {
		authorId = req.Author.Id
	} else {
		authorId = int64(userId)
	}

	lp := &models.LessonPlan{
		Name:           req.Name,
		Description:    req.Description,
		CoverImageInfo: imageInfo,
		Status:         int(req.Status),
		SortPosition:   int(req.SortPosition),
		TotalTime:      int(req.TotalTime),
		CreatedBy:      int64(userId),
		AuthorId:       authorId,
	}

	err := s.repo.Create(lp, int(req.LessonId))
	if err != nil {
		return nil, err
	}

	s.StoreLessons(lp.ID, req.Lessons)

	new, _ := s.repo.GetByID(int(lp.ID), c)
	return new, nil
}

func (s *lessonPlanService) Update(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error) {
	lessonPlanResource := resources.NewLessonPlanResource()
	lessonPlan := lessonPlanResource.FormatModelLessonPlan(req)

	userId := utils.GetCurrentUserId(c)

	if req.AuthorId > 0 {
		lessonPlan.AuthorId = req.AuthorId
	} else if req.Author != nil && req.Author.Id > 0 {
		lessonPlan.AuthorId = req.Author.Id
	} else {
		lessonPlan.AuthorId = int64(userId)
	}

	err := s.repo.Update(lessonPlan)
	if err != nil {
		return nil, err
	}

	s.repo.Complete(int(req.Id), int64(userId), req.IsComplete)
	s.StoreLessons(req.Id, req.Lessons)

	update, _ := s.repo.GetByID(int(req.Id), c)

	return update, nil
}

func (s *lessonPlanService) Delete(c *gin.Context, id int) error {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil
	}
	// Optionally accept lesson_id via query to delete the relation
	lessonID := 0
	if lessonIDStr := c.Query("lesson_id"); lessonIDStr != "" {
		if parsed, convErr := strconv.Atoi(lessonIDStr); convErr == nil {
			lessonID = parsed
		}
	}
	if lessonID > 0 {
		// Xoá record trong bảng trung gian lesson_plan_ref_lessons
		err := s.repo.DeleteLessonPlansLesson(id, lessonID)
		if err != nil {
			return err
		}
	}

	// Xoá lesson plan chính
	return s.repo.Delete(id, userID)
}

func (s *lessonPlanService) Restore(c *gin.Context, id int) (*models.LessonPlan, error) {
	baseRepo := base.NewBaseRepository[*models.LessonPlan]()
	baseRepo.SetContext(c)
	lessonPlan, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}
	return *lessonPlan, nil
}

func (s *lessonPlanService) Complete(c *gin.Context, id int, req *prot.LessonPlanCompleteRequest) error {
	userId := utils.GetCurrentUserId(c)
	return s.repo.Complete(id, int64(userId), req.IsComplete)
}

func (s *lessonPlanService) CompleteByIds(c *gin.Context, ids []int) ([]models.LessonPlanComplete, error) {
	return s.repo.CompleteByIds(ids)
}

func (s *lessonPlanService) StoreLessons(id int64, lessons []*prot.LessonPlanLessonInfo) error {
	var lessonIds []int64

	for _, lesson := range lessons {
		lessonIds = append(lessonIds, lesson.Id)
	}

	s.repo.UpdateLessons(id, lessonIds)

	return nil
}

