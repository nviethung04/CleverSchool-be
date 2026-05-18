package controllers

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
	svc services.FeedbackService
	*GenericController[dto.FeedbackResponse, prot.Feedback, *prot.FeedbackRequest]
}

func NewFeedbackController(service services.FeedbackService) *FeedbackController {
	feedbackResource := resources.NewFeedbackResource()
	feedbackResourceAdapter := NewFeedbackResourceAdapter(feedbackResource)

	genericController := NewGenericController(
		service,
		feedbackResourceAdapter,
		func() *prot.FeedbackRequest {
			return &prot.FeedbackRequest{}
		},
		func(feedbacks []*prot.Feedback, totalCount uint64) interface{} {
			return &prot.FeedbackListResponse{
				Feedbacks: feedbacks,
				Total:     int64(totalCount),
			}
		},
	)

	return &FeedbackController{
		GenericController: genericController,
		svc:               service,
	}
}

func (ctl *FeedbackController) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	req, err, message := utils.GetBody[*prot.UpdateFeedbackStatusRequest](c, func() *prot.UpdateFeedbackStatusRequest {
		return &prot.UpdateFeedbackStatusRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	data, err := ctl.svc.UpdateStatus(c, id, &req.Status, req.Response, req.Note, &req.Type)
	utils.Respond(c, data, err, "")
}
