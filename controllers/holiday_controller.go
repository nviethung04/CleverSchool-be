package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type HolidayController struct {
	*GenericController[models.Holiday, prot.Holiday, *prot.HolidayRequest]
	service services.HolidayService
}

func NewHolidayController(service services.HolidayService) *HolidayController {
	holidayResource := resources.NewHolidayResource()
	holidayResourceAdapter := NewHolidayResourceAdapter(holidayResource)

	genericController := NewGenericController(
		service,
		holidayResourceAdapter,
		func() *prot.HolidayRequest {
			return &prot.HolidayRequest{}
		},
		func(holidays []*prot.Holiday, totalCount uint64) interface{} {
			return &prot.HolidaysResponse{
				Holidays: holidays,
				TotalCount:   int64(totalCount),
			}
		},
	)

	ctl := &HolidayController{
		GenericController: genericController,
		service:           service,
	}

	return ctl
}
