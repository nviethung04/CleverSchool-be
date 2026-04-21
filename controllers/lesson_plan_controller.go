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
	var itemPtrs []*models.LessonPlan

	courseCoverImages := map[int64][]string{
		101: {
			utils.StaticURL("public/lesson plan/1751537880-lesson_plan_1.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751537938-lesson_plan_2.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751537957-lesson_plan_3.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751537974-lesson_plan_4.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751537993-lesson_plan_5.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751538008-lesson_plan_6.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751538024-lesson_plan_7.png", models.Storage),
			utils.StaticURL("public/lesson plan/1751538041-lesson_plan_8.png", models.Storage),
		},
	}

	coverIndexes := make(map[int64]int)

	for i := range items {
		images, ok := courseCoverImages[101]
		if ok && len(images) > 0 {
			idx := coverIndexes[101] % len(images)
			items[i].CoverImageInfo.Path = images[idx]
			coverIndexes[101]++
		} else {
			items[i].CoverImageInfo.Path = ""
		}
		itemPtrs = append(itemPtrs, &items[i])
	}

	formattedItems := ctl.resource.FormatItems(itemPtrs)
	response := ctl.responseFactory(formattedItems, uint64(totalCount))

	utils.Respond(c, response, err, "")
}
