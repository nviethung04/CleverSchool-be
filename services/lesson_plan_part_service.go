package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/repositories/base"
	"be-Clever School/requests"
	"be-Clever School/resources"

	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type LessonPlanPartService interface {
	GetAll(c *gin.Context) ([]models.LessonPlanPart, int64, error)
	GetByID(c *gin.Context, id int) (*prot.LessonPlanPart, error)
	Create(c *gin.Context, req *prot.LessonPlanPartRequest) (*models.LessonPlanPart, error)
	Update(c *gin.Context, req *prot.LessonPlanPartRequest) (*models.LessonPlanPart, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.LessonPlanPart, error)
}

type lessonPlanPartService struct {
	repo repositories.LessonPlanPartRepository
}

func NewLessonPlanPartService(r repositories.LessonPlanPartRepository) LessonPlanPartService {
	return &lessonPlanPartService{repo: r}
}

func (s *lessonPlanPartService) GetAll(c *gin.Context) ([]models.LessonPlanPart, int64, error) {
	var req requests.GetLessonPlanPartRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	parts, total, err := s.repo.GetAllWithPaging(&req, c)
	if err != nil {
		return nil, 0, err
	}

	return parts, total, nil
}

func (s *lessonPlanPartService) GetByID(c *gin.Context, id int) (*prot.LessonPlanPart, error) {
	part, err := s.repo.GetByID(int64(id), c)
	if err != nil {
		return nil, err
	}
	return modelToProtoLessonPlanPart(part), nil
}

func (s *lessonPlanPartService) Create(c *gin.Context, req *prot.LessonPlanPartRequest) (*models.LessonPlanPart, error) {
	lessonPlanPartResource := resources.NewLessonPlanPartResource()
	lessonPlanPart := lessonPlanPartResource.FormatModelLessonPlanPart(req)

	err := s.repo.Create(lessonPlanPart)
	if err != nil {
		return nil, err
	}
	return lessonPlanPart, nil
}

func (s *lessonPlanPartService) Update(c *gin.Context, req *prot.LessonPlanPartRequest) (*models.LessonPlanPart, error) {
	lessonPlanPartResource := resources.NewLessonPlanPartResource()
	lessonPlanPart := lessonPlanPartResource.FormatModelLessonPlanPart(req)

	err := s.repo.Update(lessonPlanPart)
	if err != nil {
		return nil, err
	}
	return lessonPlanPart, nil
}

func (s *lessonPlanPartService) Delete(c *gin.Context, id int) error {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil
	}
	return s.repo.Delete(int64(id), userID)
}

func (s *lessonPlanPartService) Restore(c *gin.Context, id int) (*models.LessonPlanPart, error) {
	baseRepo := base.NewBaseRepository[*models.LessonPlanPart]()
	baseRepo.SetContext(c)
	lessonPlan, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}
	return *lessonPlan, nil
}

func modelToProtoLessonPlanPart(part *models.LessonPlanPart) *prot.LessonPlanPart {
	isProgram := part.CourseID == 0
	return &prot.LessonPlanPart{
		Id:           part.ID,
		LessonPlanId: part.LessonPlanID,
		CourseId:	  part.CourseID,
		Title:        part.Title,
		Tag:          part.Tag,
		CoverImage:   utils.StaticURL(part.CoverImageInfo.Path, models.Storage),
		SortPosition: int32(part.SortPosition),
		Time:         part.Time,
		IsClasswork:  part.IsClasswork,
		FileType:     part.FileType,
		Link:         utils.StaticURL(part.LinkInfo.Path, models.Storage),
		LinkType:     part.LinkType,
		GuideTeacher: part.GuideTeacher,
		GuideStudent: part.GuideStudent,
		CreatedAt:    part.CreatedAt.Unix(),
		CreatedBy:    part.CreatedBy,
		UpdatedAt:    part.UpdatedAt.Unix(),
		UpdatedBy:    part.UpdatedBy,
		File:         part.File,
		IsProgram: isProgram,
	}
}
