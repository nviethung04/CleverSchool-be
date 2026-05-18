package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"time"
)


type SemesterResource interface {
	FormatSemester(semester *models.Semester) *prot.Semester
	FormatSemesters(semesters []*models.Semester) []*prot.Semester
	FormatModelSemester(semester *prot.SemesterRequest) *models.Semester
}

type SemesterResourceImpl struct{}

func NewSemesterResource() SemesterResource {
	return &SemesterResourceImpl{}
}

func (r *SemesterResourceImpl) FormatSemester(semester *models.Semester) *prot.Semester {
	holidayResource := NewHolidayResource()

	holidays := make([]*models.Holiday, 0, len(semester.Holidays))
	for i := range semester.Holidays {
		holidays = append(holidays, &semester.Holidays[i])
	}
	return &prot.Semester{
		Id:          semester.ID,
		Name:        semester.Name,
		Description: semester.Description,
		StartDate:   semester.StartDate.Format("2006-01-02"),
		EndDate:     semester.EndDate.Format("2006-01-02"),
		StartWeekId:   semester.StartWeekId,
		EndWeekId:   semester.EndWeekId,
		Status:      semester.Status,
		SortOrder:   int32(semester.SortOrder),
		PreviousSemesterId:   semester.PreviousSemesterId,
		Holidays:    holidayResource.FormatHolidays(holidays),
		CreatedAt:   semester.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   semester.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SemesterResourceImpl) FormatSemesters(semesters []*models.Semester) []*prot.Semester {
	var result []*prot.Semester
	for _, semester := range semesters {
		result = append(result, r.FormatSemester(semester))
	}
	return result
}

func (r *SemesterResourceImpl) FormatModelSemester(request *prot.SemesterRequest) *models.Semester {
	semester := &models.Semester{
		ID:          request.Id,
		Name:        request.Name,
		Description: request.Description,
		Status:      request.Status,
		SortOrder:   int(request.SortOrder),
		PreviousSemesterId:   request.PreviousSemesterId,
	}

	if request.StartWeekId != 0 {
		semester.StartWeekId = request.StartWeekId
	}

	if request.EndWeekId != 0 {
		semester.EndWeekId = request.EndWeekId
	}

	// Parse dates
	if request.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", request.StartDate); err == nil {
			semester.StartDate = startDate
		}
	}

	if request.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", request.EndDate); err == nil {
			semester.EndDate = endDate
		}
	}

	if request.BeginDate != "" {
		if beginDate, err := time.Parse("2006-01-02", request.BeginDate); err == nil {
			semester.BeginDate = beginDate
		}
	} else {
		semester.BeginDate = semester.StartDate
	}

	return semester
}
