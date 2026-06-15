package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LessonPlanController struct {
	*GenericController[models.LessonPlan, prot.LessonPlan, *prot.LessonPlanRequest]
	svc services.LessonPlanService
}

func NewLessonPlanController(service services.LessonPlanService) *LessonPlanController {
	lessonPlanResource := resources.NewLessonPlanResource()
	lessonPlanResourceAdapter := NewLessonPlanResourceAdapter(lessonPlanResource)

	genericController := NewGenericController(
		service,
		lessonPlanResourceAdapter,
		func() *prot.LessonPlanRequest {
			return &prot.LessonPlanRequest{}
		},
		func(lessonPlans []*prot.LessonPlan, totalCount uint64) interface{} {
			return &prot.LessonPlanListResponse{
				LessonPlans: lessonPlans,
				Total:       int64(totalCount),
			}
		},
	)

	ctl := &LessonPlanController{
		GenericController: genericController,
		svc:               service,
	}

	ctl.GenericController.WithRespondListHook(func(c *gin.Context, items []models.LessonPlan, totalCount int64, err error) {
		ctl.RespondList(c, items, totalCount, err)
	})

	return ctl
}

func (ctl *LessonPlanController) Complete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	req, err, message := utils.GetBody[*prot.LessonPlanCompleteRequest](c, func() *prot.LessonPlanCompleteRequest {
		return &prot.LessonPlanCompleteRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	err = ctl.svc.Complete(c, id, req)

	utils.Respond(c, &prot.CompleteResponse{
		Id: int64(id),
	}, err, "")
}

func (ctl *LessonPlanController) RespondList(c *gin.Context, items []models.LessonPlan, totalCount int64, err error) {
	itemPtrs := make([]*models.LessonPlan, 0, len(items))
	for i := range items {
		itemPtrs = append(itemPtrs, &items[i])
	}

	formattedItems := ctl.resource.FormatItems(itemPtrs)
	response := ctl.responseFactory(formattedItems, uint64(totalCount))

	utils.Respond(c, response, err, "")
}
