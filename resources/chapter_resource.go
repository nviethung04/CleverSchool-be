package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"sort"
)

type ChapterResource interface {
	FormatChapter(chapter *models.Chapter) *prot.Chapter
	FormatChapterDetail(chapter *models.Chapter) *prot.Chapter
	FormatChapters(chapters []*models.Chapter) []*prot.Chapter
	FormatDetailChapters(chapters []*models.Chapter) []*prot.Chapter
	FormatModelChapter(chapter *prot.ChapterRequest) *models.Chapter
}

type ChapterResourceImpl struct {
	CompleteLessonIds []int64
	LessonSchedules   []models.LessonSchedule
}

func NewChapterResource() ChapterResource {
	return &ChapterResourceImpl{
		CompleteLessonIds: []int64{},
		LessonSchedules:   []models.LessonSchedule{},
	}
}

func (r *ChapterResourceImpl) FormatChapter(chapter *models.Chapter) *prot.Chapter {
	if chapter == nil {
		return nil
	}

	return &prot.Chapter{
		Id:          int64(chapter.ID),
		Title:       chapter.Title,
		Description: chapter.Description,
		ObjectTitle: chapter.ObjectTitle,
		Status:      chapter.Status,
		CreatedAt:   chapter.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   chapter.UpdatedAt.Format("2006-01-02 15:04:05"),
		ProgramId:   int64(chapter.ProgramId),
	}
}

func (r *ChapterResourceImpl) FormatChapterDetail(chapter *models.Chapter) *prot.Chapter {
	if chapter == nil {
		return nil
	}

	completedMap := make(map[int64]bool, len(r.CompleteLessonIds))
	for _, id := range r.CompleteLessonIds {
		completedMap[id] = true
	}

	// Tạo map order từ LessonSchedules
	lessonOrder := make(map[int64]int64)
	for _, ls := range r.LessonSchedules {
		order := int64(int(ls.WeekID)*1000 + ls.SortPosition)
		lessonOrder[ls.LessonID] = order
	}

	lessons := make([]*prot.LessonInfo, 0, len(chapter.Lessons))
	for _, lesson := range chapter.Lessons {
		lessons = append(lessons, &prot.LessonInfo{
			Id:           lesson.ID,
			Title:        lesson.Title,
			SortPosition: int32(lesson.SortPosition),
			IsComplete:   completedMap[lesson.ID],
		})
	}

	// Sort theo order
	// sort.Slice(lessons, func(i, j int) bool {
	// 	orderI, okI := lessonOrder[lessons[i].Id]
	// 	orderJ, okJ := lessonOrder[lessons[j].Id]

	// 	switch {
	// 	case okI && okJ: // cả 2 đều có trong LessonSchedules
	// 		if orderI == orderJ {
	// 			return lessons[i].Id < lessons[j].Id
	// 		}
	// 		return orderI < orderJ
	// 	case okI: // chỉ i có trong LessonSchedules
	// 		return true
	// 	case okJ: // chỉ j có trong LessonSchedules
	// 		return false
	// 	default: // cả 2 đều không có trong LessonSchedules
	// 		return lessons[i].Id < lessons[j].Id
	// 	}
	// })

	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].SortPosition == lessons[j].SortPosition {
			return lessons[i].Id < lessons[j].Id
		}
		return lessons[i].SortPosition < lessons[j].SortPosition
	})

	return &prot.Chapter{
		Id:          int64(chapter.ID),
		Title:       chapter.Title,
		Description: chapter.Description,
		ObjectTitle: chapter.ObjectTitle,
		Lessons:     lessons,
		Status:      chapter.Status,
		CreatedAt:   chapter.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   chapter.UpdatedAt.Format("2006-01-02 15:04:05"),
		ProgramId:   int64(chapter.ProgramId),
	}
}

func (r *ChapterResourceImpl) FormatChapters(chapters []*models.Chapter) []*prot.Chapter {
	result := make([]*prot.Chapter, 0, len(chapters))
	for _, u := range chapters {
		if formatted := r.FormatChapter(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ChapterResourceImpl) FormatDetailChapters(chapters []*models.Chapter) []*prot.Chapter {
	result := make([]*prot.Chapter, 0, len(chapters))
	for _, u := range chapters {
		if formatted := r.FormatChapterDetail(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ChapterResourceImpl) FormatModelChapter(chapter *prot.ChapterRequest) *models.Chapter {
	if chapter == nil {
		return nil
	}

	return &models.Chapter{
		ID:          int64(chapter.Id),
		Title:       chapter.Title,
		Description: chapter.Description,
		ObjectTitle: chapter.ObjectTitle,
		Status:      chapter.Status,
		ProgramId:   int64(chapter.ProgramId),
	}
}
