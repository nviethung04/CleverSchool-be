package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
)

type WeekResource interface {
	FormatWeek(week *models.Week) *prot.Week
	FormatWeeks(weeks []*models.Week) []*prot.Week
}

type WeekResourceImpl struct{}

func NewWeekResource() WeekResource {
	return &WeekResourceImpl{}
}

func (r *WeekResourceImpl) FormatWeek(week *models.Week) *prot.Week {
	if week == nil {
		return nil
	}

	return &prot.Week{
		Id:         int64(week.ID),
		Year:       int32(week.Year),
		WeekNumber: int32(week.WeekNumber),
		StartDate:  week.StartDate.Format("2006-01-02"),
		EndDate:    week.EndDate.Format("2006-01-02"),
		CreatedAt:  week.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  week.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *WeekResourceImpl) FormatWeeks(weeks []*models.Week) []*prot.Week {
	result := make([]*prot.Week, 0, len(weeks))
	for index, u := range weeks {
		if formatted := r.FormatWeek(u); formatted != nil {
			formatted.Number = int32(index + 1)
			result = append(result, formatted)
		}
	}
	return result
}
