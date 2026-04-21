package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type NoticeController struct {
	*GenericController[models.Notice, prot.Notice, *prot.Notice]
}

func NewNoticeController(service services.NoticeService) *NoticeController {
	noticeResource := resources.NewNoticeResource()
	noticeResourceAdapter := NewNoticeResourceAdapter(noticeResource)

	genericController := NewGenericController(
		service,
		noticeResourceAdapter,
		func() *prot.Notice {
			return &prot.Notice{}
		},
		func(notices []*prot.Notice, totalCount uint64) interface{} {
			return &prot.NoticeListResponse{
				Notices:    notices,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &NoticeController{
		GenericController: genericController,
	}
}
