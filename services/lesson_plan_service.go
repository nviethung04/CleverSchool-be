package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/repositories/base"
	"be-lms/requests"
	"be-lms/resources"

	"be-lms/utils"

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
	return modelToProtoLessonPlan(lp), nil
}

func (s *lessonPlanService) Create(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error) {
	coverImageUrl := utils.StripDomain(req.CoverImage, models.Storage)
	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(coverImageUrl, models.Storage)

	lp := &models.LessonPlan{
		Name:           req.Name,
		Description:    req.Description,
		CoverImageInfo: imageInfo,
		Status:         int(req.Status),
		SortPosition:   int(req.SortPosition),
		TotalTime:      int(req.TotalTime),
		CreatedBy:      1,
	}

	err := s.repo.Create(lp, int(req.LessonId))
	if err != nil {
		return nil, err
	}
	return lp, nil
}

func (s *lessonPlanService) Update(c *gin.Context, req *prot.LessonPlanRequest) (*models.LessonPlan, error) {
	lessonPlanResource := resources.NewLessonPlanResource()
	lessonPlan := lessonPlanResource.FormatModelLessonPlan(req)

	err := s.repo.Update(lessonPlan)
	if err != nil {
		return nil, err
	}

	userId := utils.GetCurrentUserId(c)
	s.repo.Complete(int(req.Id), int64(userId), req.IsComplete)

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

func modelToProtoLessonPlan(lp *models.LessonPlan) *prot.LessonPlan {
	var isComplete bool
	var completeAt int64

	if lp.Complete != nil {
		isComplete = true
		completeAt = lp.Complete.CompletedAt.Unix()
	}

	return &prot.LessonPlan{
		Id:           int64(lp.ID),
		Name:         lp.Name,
		Description:  lp.Description,
		CoverImage:   utils.StaticURL(lp.CoverImageInfo.Path, models.Storage),
		Status:       int32(lp.Status),
		SortPosition: int32(lp.SortPosition),
		TotalTime:    int64(lp.TotalTime),
		Views:        int32(lp.Views),
		CreatedAt:    lp.CreatedAt.Unix(),
		UpdatedAt:    lp.UpdatedAt.Unix(),
		IsComplete:   isComplete,
		CompleteAt:   completeAt,
	}
}

func (s *lessonPlanService) CompleteByIds(c *gin.Context, ids []int) ([]models.LessonPlanComplete, error) {
	return s.repo.CompleteByIds(ids)
}
