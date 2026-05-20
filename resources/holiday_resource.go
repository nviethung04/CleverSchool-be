package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"time"
)


type HolidayResource interface {
	FormatHoliday(holiday *models.Holiday) *prot.Holiday
	FormatHolidays(holidays []*models.Holiday) []*prot.Holiday
	FormatModelHoliday(holiday *prot.HolidayRequest) *models.Holiday
	ProtToModel(holiday *prot.Holiday) *models.Holiday
}

type HolidayResourceImpl struct{}

func NewHolidayResource() HolidayResource {
	return &HolidayResourceImpl{}
}

func (r *HolidayResourceImpl) FormatHoliday(holiday *models.Holiday) *prot.Holiday {
	return &prot.Holiday{
		Id:          holiday.ID,
		Name:        holiday.Name,
		Description: holiday.Description,
		StartDate:   holiday.StartDate.Format("2006-01-02"),
		EndDate:     holiday.EndDate.Format("2006-01-02"),
		Type:        holiday.Type,
		Status:      holiday.Status,
		SortOrder:   int32(holiday.SortOrder),
		WeekId:      holiday.WeekId,
		CreatedAt:   holiday.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   holiday.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *HolidayResourceImpl) FormatHolidays(holidays []*models.Holiday) []*prot.Holiday {
	var result []*prot.Holiday
	for _, holiday := range holidays {
		result = append(result, r.FormatHoliday(holiday))
	}
	return result
}

func (r *HolidayResourceImpl) FormatModelHoliday(request *prot.HolidayRequest) *models.Holiday {
	holidayWeek := &models.Holiday{
		ID:          request.Id,
		Name:        request.Name,
		Description: request.Description,
		Type:        request.Type,
		Status:      request.Status,
		SortOrder:   int(request.SortOrder),
	}

	// Parse dates
	if request.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", request.StartDate); err == nil {
			holidayWeek.StartDate = startDate
		}
	}

	if request.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", request.EndDate); err == nil {
			holidayWeek.EndDate = endDate
		}
	}

	return holidayWeek
}

func (r *HolidayResourceImpl) ProtToModel(request *prot.Holiday) *models.Holiday {
	holidayWeek := &models.Holiday{
		ID:          request.Id,
		Name:        request.Name,
		Description: request.Description,
		Type:        request.Type,
		Status:      request.Status,
		SortOrder:   int(request.SortOrder),
		WeekId:   request.WeekId,
	}

	// Parse dates
	if request.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", request.StartDate); err == nil {
			holidayWeek.StartDate = startDate
		}
	}

	if request.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", request.EndDate); err == nil {
			holidayWeek.EndDate = endDate
		}
	}

	return holidayWeek
}

