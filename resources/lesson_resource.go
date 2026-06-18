package resources

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/utils"
	"sort"
	"strings"
)

type LessonResource interface {
	FormatLesson(lesson *models.Lesson) *prot.Lesson
	FormatLessons(lessons []*models.Lesson) []*prot.Lesson
	FormatModelLesson(lesson *prot.LessonRequest) *models.Lesson
}

type LessonResourceImpl struct {
	CompleteLessonIds      []int64
	CompleteLessonPlanIds  []int64
	CourseId			   int64
	Flashcard              *dto.FlashcardLessonDTO
	examCompletionMap      map[int64]bool
	examTotalQuestions     map[int64]int32
	homeworkCompletedMap   map[int64]int32
	homeworkTotalQuestions map[int64]int32
	exerciseCompletionMap  map[int64]bool
	exerciseTotalQuestions map[int64]int32
}

func NewLessonResourceWithCompletionAndQuestions(
	examCompletionMap map[int64]bool,
	examTotalQuestions map[int64]int32,
	homeworkCompletedMap map[int64]int32,
	homeworkTotalQuestions map[int64]int32,
	exerciseCompletionMap map[int64]bool,
	exerciseTotalQuestions map[int64]int32,
) LessonResource {
	return &LessonResourceImpl{
		CompleteLessonIds:      []int64{},
		CompleteLessonPlanIds:  []int64{},
		CourseId:				0,
		Flashcard:              nil,
		examCompletionMap:      examCompletionMap,
		examTotalQuestions:     examTotalQuestions,
		homeworkCompletedMap:   homeworkCompletedMap,
		homeworkTotalQuestions: homeworkTotalQuestions,
		exerciseCompletionMap:  exerciseCompletionMap,
		exerciseTotalQuestions: exerciseTotalQuestions,
	}
}

func NewLessonResource() LessonResource {
	return &LessonResourceImpl{
		CompleteLessonIds: []int64{},
		CompleteLessonPlanIds: []int64{},
	}
}

func (r *LessonResourceImpl) FormatLesson(lesson *models.Lesson) *prot.Lesson {
	if lesson == nil {
		return nil
	}

	chapter := prot.ChapterInfo{
		Id:    lesson.Chapter.ID,
		Title: lesson.Chapter.Title,
		ObjectTitle:	 lesson.Chapter.ObjectTitle,
	}

	tags := make([]*prot.TagInfo, 0, len(lesson.Tags))
	for _, attr := range lesson.Tags {
		tags = append(tags, &prot.TagInfo{
			Id:   attr.ID,
			Name: attr.Name,
		})
	}

	topics := make([]*prot.TopicInfo, 0, len(lesson.Topics))
	for _, attr := range lesson.Topics {
		topics = append(topics, &prot.TopicInfo{
			Id:   attr.ID,
			Name: attr.Name,
		})
	}

	skills := make([]*prot.SkillInfo, 0, len(lesson.Skills))
	for _, attr := range lesson.Skills {
		skills = append(skills, &prot.SkillInfo{
			Id:   attr.ID,
			Name: attr.Name,
		})
	}

	dependencies := make([]*prot.LessonInfo, 0, len(lesson.Dependencies))
	for _, attr := range lesson.Dependencies {
		dependencies = append(dependencies, &prot.LessonInfo{
			Id:    attr.ID,
			Title: attr.Title,
		})
	}

	exams := r.GetExams(lesson)
	homeworks := r.GetHomeworks(lesson)
	exercises := r.GetExercises(lesson)
	assessments := r.GetAssessments(lesson)

	lessonPlans := make([]*prot.LessonPlanInfo, 0, len(lesson.LessonPlans))
	seen := make(map[int64]bool)

	completeMap := make(map[int64]bool, len(r.CompleteLessonPlanIds))
	for _, id := range r.CompleteLessonPlanIds {
		completeMap[id] = true
	}

	for _, attr := range lesson.LessonPlans {
		if seen[attr.ID] {
			continue
		}
		seen[attr.ID] = true

		if len(lessonPlans) > 0 {
			continue
		}

		lessonPlans = append(lessonPlans, &prot.LessonPlanInfo{
			Id:           attr.ID,
			Name:         attr.Name,
			Description:  attr.Description,
			ObjectTitle:	 attr.ObjectTitle,
			CoverImage:   utils.StaticURL(attr.CoverImageInfo.Path, models.Storage),
			Status:       int32(attr.Status),
			SortPosition: int32(attr.SortPosition),
			TotalTime:    int64(attr.TotalTime),
			IsComplete:   completeMap[attr.ID],
		})
	}

	sort.Slice(lessonPlans, func(i, j int) bool {
		if lessonPlans[i].SortPosition == lessonPlans[j].SortPosition {
			return lessonPlans[i].Id < lessonPlans[j].Id
		}
		return lessonPlans[i].SortPosition < lessonPlans[j].SortPosition
	})

	var author prot.AuthorInfo
	if lesson.Author != nil {
		author = prot.AuthorInfo{
			Id:          lesson.Author.ID,
			Name:        lesson.Author.Name,
			Username:    lesson.Author.Username,
			Email:       lesson.Author.Email,
			PhoneNumber: lesson.Author.PhoneNumber,
		}
	}

	var program prot.LessonProgramInfo
	if chapter.Id != 0 && lesson.Chapter.Program.ID != 0 {
		program = prot.LessonProgramInfo{
			Id:   lesson.Chapter.Program.ID,
			Name: lesson.Chapter.Program.Name,
			Description:	 lesson.Chapter.Program.Description,
		}
	}

	sort.Slice(lesson.Schedules, func(i, j int) bool {
		return lesson.Schedules[i].WeekID < lesson.Schedules[j].WeekID
	})

	schedules := make([]*prot.LessonScheduleInfo, 0, len(lesson.Schedules))
	for _, attr := range lesson.Schedules {
		var week *prot.WeekInfo
		if attr.Week != nil {
			week = &prot.WeekInfo{
				Id:         attr.Week.ID,
				Year:       int32(attr.Week.Year),
				WeekNumber: int32(attr.Week.WeekNumber),
				StartDate:  attr.Week.StartDate.Format("2006-01-02"),
				EndDate:    attr.Week.EndDate.Format("2006-01-02"),
			}
		}

		var weekNumber int32 = 1
		if attr.Course != nil && !attr.ScheduledDate.IsZero() && !attr.Course.StartDate.IsZero() {
			diff := attr.ScheduledDate.Sub(attr.Course.StartDate)
			weekNumber = int32(diff.Hours()/24/7) + 1
			if weekNumber < 1 {
				weekNumber = 1
			}
		}

		schedules = append(schedules, &prot.LessonScheduleInfo{
			Id:            attr.ID,
			CourseId:      attr.CourseID,
			ScheduledDate: attr.ScheduledDate.Format("2006-01-02"),
			WeekId:        attr.WeekID,
			WeekNumber:    weekNumber,
			Week:          week,
		})
	}

	isComplete := false
	for _, id := range r.CompleteLessonIds {
		if id == lesson.ID {
			isComplete = true
			break
		}
	}

	var vocabularies []*prot.Vocabulary

	if r.Flashcard != nil {
		for _, lv := range r.Flashcard.Vocabularies {
			if lv.Vocabulary != nil {
				var wordAudio prot.MediaFlashcard
				var exampleAudio prot.MediaFlashcard
				var image prot.MediaFlashcard

				if lv.Vocabulary.WordAudio != nil {
					wordAudio = prot.MediaFlashcard{
						Id:   lv.Vocabulary.WordAudio.ID,
						Type: lv.Vocabulary.WordAudio.FileType,
						Url:  utils.StaticURL(lv.Vocabulary.WordAudio.FilePath, models.Storage),
					}
				}

				if lv.Vocabulary.ExampleAudio != nil {
					wordAudio = prot.MediaFlashcard{
						Id:   lv.Vocabulary.ExampleAudio.ID,
						Type: lv.Vocabulary.ExampleAudio.FileType,
						Url:  utils.StaticURL(lv.Vocabulary.ExampleAudio.FilePath, models.Storage),
					}
				}

				if lv.Vocabulary.Image != nil {
					image = prot.MediaFlashcard{
						Id:   lv.Vocabulary.Image.ID,
						Type: lv.Vocabulary.Image.FileType,
						Url:  utils.StaticURL(lv.Vocabulary.Image.FilePath, models.Storage),
					}
				}
				vocabularies = append(vocabularies, &prot.Vocabulary{
					Id:           lv.Vocabulary.ID,
					Word:         lv.Vocabulary.Word,
					WordAudio:    &wordAudio,
					Image:        &image,
					ExampleAudio: &exampleAudio,
				})
			}
		}
	}

	return &prot.Lesson{
		Id:           int64(lesson.ID),
		Title:        lesson.Title,
		Description:  lesson.Description,
		ObjectTitle:	 lesson.ObjectTitle,
		Status:       lesson.Status,
		Views:        int32(lesson.Views),
		Chapter:      &chapter,
		Program:       &program,
		Tags:         tags,
		Topics:       topics,
		Skills:       skills,
		Dependencies: dependencies,
		Exams:        exams,
		Homeworks:    homeworks,
		Exercises:    exercises,
		Assessments:  assessments,
		LessonPlans:  lessonPlans,
		Schedules:    schedules,
		Author:       &author,
		IsComplete:   isComplete,
		Vocabularies: vocabularies,
		VocabularyCount: int32(len(vocabularies)),
		CreatedAt:    lesson.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    lesson.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *LessonResourceImpl) FormatLessons(lessons []*models.Lesson) []*prot.Lesson {
	result := make([]*prot.Lesson, 0, len(lessons))
	for _, u := range lessons {
		if formatted := r.FormatLesson(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *LessonResourceImpl) FormatModelLesson(lesson *prot.LessonRequest) *models.Lesson {
	if lesson == nil {
		return nil
	}

	var chapterID int64
	if lesson.Chapter != nil {
		chapterID = int64(lesson.Chapter.Id)
	}

	return &models.Lesson{
		ID:          int64(lesson.Id),
		ChapterID:   chapterID,
		Title:       strings.TrimSpace(lesson.Title),
		Description: strings.TrimSpace(lesson.Description),
		Status:      lesson.Status,
		Views:       int(lesson.Views),
		ObjectTitle: lesson.ObjectTitle,
	}
}

func (r *LessonResourceImpl) GetHomeworks(lesson *models.Lesson) []*prot.HomeworkInfo {
	homeworksMap := make(map[int64]*prot.HomeworkInfo)

	for _, attr := range lesson.Homeworks {
		for _, ref := range attr.HomeworkRefLessons {
			if ref.LessonId == lesson.ID && ref.HomeworkId == attr.ID && ref.CourseId == 0 {
				homeworksMap[attr.ID] = &prot.HomeworkInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					QuestionsCompleted: r.homeworkCompletedMap[attr.ID],
					TotalQuestions:     r.homeworkTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         false,
					IsProgram:          true,
				}
				break
			}
		}
	}

	if r.CourseId != 0 {
		for _, attr := range lesson.Homeworks {
			var (
				hasCourseRef bool
				isAssigned   bool
			)

			for _, ref := range attr.HomeworkRefLessons {
				if ref.LessonId == lesson.ID && ref.HomeworkId == attr.ID && ref.CourseId == r.CourseId {
					hasCourseRef = true
					if ref.AssignedBy != nil && *ref.AssignedBy > 0 {
						isAssigned = true
						break
					}
				}
			}

			if !hasCourseRef {
				continue
			}

			if existing, ok := homeworksMap[attr.ID]; ok {
				existing.IsAssigned = isAssigned
			} else {
				homeworksMap[attr.ID] = &prot.HomeworkInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					QuestionsCompleted: r.homeworkCompletedMap[attr.ID],
					TotalQuestions:     r.homeworkTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         isAssigned,
					IsProgram:          false,
				}
			}
		}
	}

	ids := make([]int64, 0, len(homeworksMap))
	for id := range homeworksMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	homeworks := make([]*prot.HomeworkInfo, 0, len(ids))
	for _, id := range ids {
		homeworks = append(homeworks, homeworksMap[id])
	}

	return homeworks
}

func (r *LessonResourceImpl) GetAssessments(lesson *models.Lesson) []*prot.AssessmentInfo {
	assessmentsMap := make(map[int64]*prot.AssessmentInfo)

	for _, attr := range lesson.Assessments {
		for _, ref := range attr.AssessmentRefLessons {
			if ref.LessonId == lesson.ID && ref.AssessmentId == attr.ID && ref.CourseId == 0 {
				assessmentsMap[attr.ID] = &prot.AssessmentInfo{
					Id:          attr.ID,
					Name:        attr.Name,
					Description: attr.Description,
					Type:        attr.Type,
					IsProgram:   true,
				}
				break
			}
		}
	}

	if r.CourseId != 0 {
		for _, attr := range lesson.Assessments {
			var hasCourseRef bool

			for _, ref := range attr.AssessmentRefLessons {
				if ref.LessonId == lesson.ID && ref.AssessmentId == attr.ID && ref.CourseId == r.CourseId {
					hasCourseRef = true
					break
				}
			}

			if !hasCourseRef {
				continue
			}

			if _, ok := assessmentsMap[attr.ID]; !ok {
				assessmentsMap[attr.ID] = &prot.AssessmentInfo{
					Id:          attr.ID,
					Name:        attr.Name,
					Description: attr.Description,
					Type:        attr.Type,
					IsProgram:   false,
				}
			}
		}
	}

	ids := make([]int64, 0, len(assessmentsMap))
	for id := range assessmentsMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	assessments := make([]*prot.AssessmentInfo, 0, len(ids))
	for _, id := range ids {
		assessments = append(assessments, assessmentsMap[id])
	}

	return assessments
}

func (r *LessonResourceImpl) GetExercises(lesson *models.Lesson) []*prot.ExerciseInfo {
	exercisesMap := make(map[int64]*prot.ExerciseInfo)

	for _, attr := range lesson.Exercises {
		for _, ref := range attr.ExerciseRefLessons {
			if ref.LessonId == lesson.ID && ref.ExerciseId == attr.ID && ref.CourseId == 0 {
				exercisesMap[attr.ID] = &prot.ExerciseInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					IsCompleted:    r.exerciseCompletionMap[attr.ID],
					TotalQuestions: r.exerciseTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         false,
					IsProgram:          true,
					TimeLimit:      int32(attr.TimeLimit),
				}
				break
			}
		}
	}

	if r.CourseId != 0 {
		for _, attr := range lesson.Exercises {
			var (
				hasCourseRef bool
				isAssigned   bool
			)

			for _, ref := range attr.ExerciseRefLessons {
				if ref.LessonId == lesson.ID && ref.ExerciseId == attr.ID && ref.CourseId == r.CourseId {
					hasCourseRef = true
					if ref.AssignedBy != nil && *ref.AssignedBy > 0 {
						isAssigned = true
						break
					}
				}
			}

			if !hasCourseRef {
				continue
			}

			if existing, ok := exercisesMap[attr.ID]; ok {
				existing.IsAssigned = isAssigned
			} else {
				exercisesMap[attr.ID] = &prot.ExerciseInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					IsCompleted:    r.exerciseCompletionMap[attr.ID],
					TotalQuestions: r.exerciseTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         isAssigned,
					IsProgram:          false,
					TimeLimit:      int32(attr.TimeLimit),
				}
			}
		}
	}

	ids := make([]int64, 0, len(exercisesMap))
	for id := range exercisesMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	exercises := make([]*prot.ExerciseInfo, 0, len(ids))
	for _, id := range ids {
		exercises = append(exercises, exercisesMap[id])
	}

	return exercises
}

func (r *LessonResourceImpl) GetExams(lesson *models.Lesson) []*prot.ExamInfo {
	examsMap := make(map[int64]*prot.ExamInfo)

	for _, attr := range lesson.Exams {
		for _, ref := range attr.ExamRefLessons {
			if ref.LessonId == lesson.ID && ref.ExamId == attr.ID && ref.CourseId == 0 {
				examsMap[attr.ID] = &prot.ExamInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					IsCompleted:    r.exerciseCompletionMap[attr.ID],
					TotalQuestions: r.exerciseTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         false,
					IsProgram:          true,
					TimeLimit:      int32(attr.TimeLimit),
				}
				break
			}
		}
	}

	if r.CourseId != 0 {
		for _, attr := range lesson.Exams {
			var (
				hasCourseRef bool
				isAssigned   bool
			)

			for _, ref := range attr.ExamRefLessons {
				if ref.LessonId == lesson.ID && ref.ExamId == attr.ID && ref.CourseId == r.CourseId {
					hasCourseRef = true
					if ref.AssignedBy != nil && *ref.AssignedBy > 0 {
						isAssigned = true
						break
					}
				}
			}

			if !hasCourseRef {
				continue
			}

			if existing, ok := examsMap[attr.ID]; ok {
				existing.IsAssigned = isAssigned
			} else {
				examsMap[attr.ID] = &prot.ExamInfo{
					Id:                 attr.ID,
					Name:               attr.Name,
					Description:        attr.Description,
					ObjectTitle:        attr.ObjectTitle,
					IsCompleted:    r.exerciseCompletionMap[attr.ID],
					TotalQuestions: r.exerciseTotalQuestions[attr.ID],
					CreatedAt:          attr.CreatedAt.Format("2006-01-02"),
					IsAssigned:         isAssigned,
					IsProgram:          false,
					TimeLimit:      int32(attr.TimeLimit),
				}
			}
		}
	}

	ids := make([]int64, 0, len(examsMap))
	for id := range examsMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	exams := make([]*prot.ExamInfo, 0, len(ids))
	for _, id := range ids {
		exams = append(exams, examsMap[id])
	}

	return exams
}
