package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
)

type HeadingResource interface {
	FormatHeading(heading *models.Heading) *prot.Heading
	FormatHeadings(headings []*models.Heading) []*prot.Heading
	FormatModelHeading(heading *prot.HeadingRequest) *models.Heading
}

type HeadingResourceImpl struct{}

func NewHeadingResource() HeadingResource {
	return &HeadingResourceImpl{}
}

func (r *HeadingResourceImpl) FormatHeading(heading *models.Heading) *prot.Heading {
	if heading == nil {
		return nil
	}

	lessons := make([]*prot.LessonInfo, 0, len(heading.Lessons))
	for _, l := range heading.Lessons {
		lessons = append(lessons, &prot.LessonInfo{
			Id:           l.ID,
			Title:        l.Title,
			SortPosition: int32(l.SortPosition),
		})
	}

	return &prot.Heading{
		Id:           int64(heading.ID),
		ChapterId:    heading.ChapterId,
		Name:         heading.Name,
		Description:  heading.Description,
		Time:         heading.Time,
		Lessons:      lessons,
		SortPosition: int32(heading.SortPosition),
		Target:       heading.Target,
		CreatedAt:    heading.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    heading.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *HeadingResourceImpl) FormatHeadings(headings []*models.Heading) []*prot.Heading {
	result := make([]*prot.Heading, 0, len(headings))
	for _, t := range headings {
		if formatted := r.FormatHeading(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *HeadingResourceImpl) FormatModelHeading(heading *prot.HeadingRequest) *models.Heading {
	if heading == nil {
		return nil
	}

	return &models.Heading{
		ID:           int64(heading.Id),
		ChapterId:    heading.ChapterId,
		Name:         heading.Name,
		Description:  heading.Description,
		Time:         heading.Time,
		SortPosition: int(heading.SortPosition),
		Target:       heading.Target,
	}
}
