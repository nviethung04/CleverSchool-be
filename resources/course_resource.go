package resources

import (
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"fmt"
	"sort"
	"time"
)

type CourseResource interface {
	FormatCourse(course *models.Course) *prot.Course
	FormatCourseDetail(course *models.Course) *prot.Course
	FormatCourses(courses []*models.Course) []*prot.Course
	FormatDetailCourses(courses []*models.Course) []*prot.Course
	FormatModelCourse(course *prot.CourseRequest) *models.Course
}

type CourseResourceImpl struct {
	HideLessonIds              []int64
	StudyingLessonIds          []int64
	CompleteLessonIds          []int64
	LessonSchedules            []models.LessonSchedule
	CopyScheduleResponse       dto.CopyScheduleResponse
	NotEligibleForFinalExamIds []int64
}

func NewCourseResource() CourseResource {
	return &CourseResourceImpl{
		HideLessonIds:              []int64{},
		StudyingLessonIds:          []int64{},
		CompleteLessonIds:          []int64{},
		LessonSchedules:            []models.LessonSchedule{},
		CopyScheduleResponse:       dto.CopyScheduleResponse{},
		NotEligibleForFinalExamIds: []int64{},
	}
}

func (r *CourseResourceImpl) FormatCourse(course *models.Course) *prot.Course {
	if course == nil {
		return nil
	}

	chapterResource := NewChapterResource()

	if impl, ok := chapterResource.(*ChapterResourceImpl); ok {
		impl.HideLessonIds = r.HideLessonIds
		impl.StudyingLessonIds = r.StudyingLessonIds
		impl.CompleteLessonIds = r.CompleteLessonIds
		impl.LessonSchedules = r.LessonSchedules
	}

	chapters := make([]*models.Chapter, 0, len(course.Program.Chapters))
	for i := range course.Program.Chapters {
		chapters = append(chapters, &course.Program.Chapters[i])
	}

	sort.Slice(chapters, func(i, j int) bool {
		if chapters[i].SortPosition == chapters[j].SortPosition {
			return chapters[i].ID < chapters[j].ID
		}
		return chapters[i].SortPosition < chapters[j].SortPosition
	})

	var currentStudents int32
	if course.CurrentStudents != 0 {
		currentStudents = course.CurrentStudents
	}

	var schoolInfo *prot.SchooInfo = nil

	if len(course.Schools) > 0 {
		school := course.Schools[0]
		schoolInfo = &prot.SchooInfo{
			Id:   school.ID,
			Name: school.Name,
		}
	}

	var programInfo *prot.ProgramInfo = nil
	var subjectId int64 = 0

	if course.Program.ID != 0 {
		programInfo = &prot.ProgramInfo{
			Id:          course.Program.ID,
			Name:        course.Program.Name,
			Description: course.Program.Description,
			BookUrl:     course.Program.BookUrl,
		}
		// Lấy subject_id từ program, nếu không có thì trả về 0
		if len(course.Program.Subjects) > 0 {
			subjectId = int64(course.Program.Subjects[0].ID)
		}
	}

	notEligibleForFinalExam := false

	for _, id := range r.NotEligibleForFinalExamIds {
		if id == course.ID {
			notEligibleForFinalExam = true
			break
		}
	}

	return &prot.Course{
		Id:                   int64(course.ID),
		SubjectId:            subjectId,
		ProgramId:            int64(course.ProgramId),
		Name:                 course.Name,
		Description:          course.Description,
		Type:                 course.Type,
		Level:                course.Level,
		ObjectTitle:          course.ObjectTitle,
		Duration:             int32(course.Duration),
		Image:                utils.StaticURL(course.ImageInfo.Path, models.Storage),
		Status:               course.Status,
		CurrentStudents:      currentStudents,
		StartDate:            course.StartDate.Format("2006-01-02"),
		EndDate:              course.EndDate.Format("2006-01-02"),
		Time:                 course.Time,
		Target:               course.Target,
		Chapters:             chapterResource.FormatChapters(chapters),
		School:               schoolInfo,
		Program:              programInfo,
		Progress:             fmt.Sprintf("%d%%", GetProgress(*course)),
		State:                course.State,
		CreatedAt:            course.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            course.UpdatedAt.Format("2006-01-02 15:04:05"),
		CopyScheduleSuccess:  r.CopyScheduleResponse.CopyScheduleSuccess,
		CopyScheduleMessage:  r.CopyScheduleResponse.CopyScheduleMessage,
		EligibleForFinalExam: !notEligibleForFinalExam,
	}
}

func (r *CourseResourceImpl) FormatCourseDetail(course *models.Course) *prot.Course {
	if course == nil {
		return nil
	}

	chapterResource := NewChapterResource()

	if impl, ok := chapterResource.(*ChapterResourceImpl); ok {
		impl.HideLessonIds = r.HideLessonIds
		impl.StudyingLessonIds = r.StudyingLessonIds
		impl.CompleteLessonIds = r.CompleteLessonIds
		impl.LessonSchedules = r.LessonSchedules
	}

	chapters := make([]*models.Chapter, 0, len(course.Program.Chapters))
	for i := range course.Program.Chapters {
		chapters = append(chapters, &course.Program.Chapters[i])
	}

	sort.Slice(chapters, func(i, j int) bool {
		if chapters[i].SortPosition == chapters[j].SortPosition {
			return chapters[i].ID < chapters[j].ID
		}
		return chapters[i].SortPosition < chapters[j].SortPosition
	})

	var currentStudents int32
	if course.CurrentStudents != 0 {
		currentStudents = course.CurrentStudents
	}

	var schoolInfo *prot.SchooInfo = nil

	if len(course.Schools) > 0 {
		school := course.Schools[0]
		schoolInfo = &prot.SchooInfo{
			Id:   school.ID,
			Name: school.Name,
		}
	}
	sort.Slice(course.CourseRefStudyShifts, func(i, j int) bool {
		return course.CourseRefStudyShifts[i].DayOfWeek < course.CourseRefStudyShifts[j].DayOfWeek
	})

	studyShifts := make([]*prot.StudyShiftInfo, 0, len(course.CourseRefStudyShifts))

	for _, s := range course.CourseRefStudyShifts {
		studyShifts = append(studyShifts, &prot.StudyShiftInfo{
			Id:        s.StudyShift.ID,
			Name:      s.StudyShift.Name,
			StartTime: s.StudyShift.StartTime,
			EndTime:   s.StudyShift.EndTime,
			DayOfWeek: s.DayOfWeek,
		})
	}

	var teacherInfo *prot.CourseTeacherInfo

	found := false
	for _, u := range course.Users {
		for _, r := range u.Roles {
			if r.ID == models.TeacherRoleId {
				teacherInfo = &prot.CourseTeacherInfo{
					Id:          u.ID,
					Avatar:      utils.StaticURL(u.AvatarInfo.Path, models.Storage),
					Name:        u.Name,
					Description: u.Description,
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	var semesters []*prot.SemesterInfo
	for _, s := range course.Semesters {
		semesters = append(semesters, &prot.SemesterInfo{
			Id:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			StartDate:   s.StartDate.Format("2006-01-02"),
			EndDate:     s.EndDate.Format("2006-01-02"),
		})
	}

	sort.Slice(semesters, func(i, j int) bool {
		return semesters[i].StartDate < semesters[j].StartDate
	})

	var programInfo *prot.ProgramInfo = nil
	var subjectId int64 = 0

	if course.Program.ID != 0 {
		programInfo = &prot.ProgramInfo{
			Id:          course.Program.ID,
			Name:        course.Program.Name,
			Description: course.Program.Description,
			BookUrl:     course.Program.BookUrl,
		}
		// Lấy subject_id từ program, nếu không có thì trả về 0
		if len(course.Program.Subjects) > 0 {
			subjectId = int64(course.Program.Subjects[0].ID)
		}
	}

	notEligibleForFinalExam := false
	for _, id := range r.NotEligibleForFinalExamIds {
		if id == course.ID {
			notEligibleForFinalExam = true
			break
		}
	}

	return &prot.Course{
		Id:                   int64(course.ID),
		SubjectId:            subjectId,
		ProgramId:            int64(course.ProgramId),
		Name:                 course.Name,
		Description:          course.Description,
		Type:                 course.Type,
		ObjectTitle:          course.ObjectTitle,
		Level:                course.Level,
		Duration:             int32(course.Duration),
		Image:                utils.StaticURL(course.ImageInfo.Path, models.Storage),
		Status:               course.Status,
		CurrentStudents:      currentStudents,
		StartDate:            course.StartDate.Format("2006-01-02"),
		EndDate:              course.EndDate.Format("2006-01-02"),
		Time:                 course.Time,
		Target:               course.Target,
		Chapters:             chapterResource.FormatDetailChapters(chapters),
		School:               schoolInfo,
		Program:              programInfo,
		Teacher:              teacherInfo,
		StudyShifts:          studyShifts,
		Semesters:            semesters,
		State:                course.State,
		Progress:             fmt.Sprintf("%d%%", GetProgress(*course)),
		CreatedAt:            course.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:            course.UpdatedAt.Format("2006-01-02 15:04:05"),
		CopyScheduleSuccess:  r.CopyScheduleResponse.CopyScheduleSuccess,
		CopyScheduleMessage:  r.CopyScheduleResponse.CopyScheduleMessage,
		EligibleForFinalExam: !notEligibleForFinalExam,
	}
}

func (r *CourseResourceImpl) FormatCourses(courses []*models.Course) []*prot.Course {
	result := make([]*prot.Course, 0, len(courses))
	for _, u := range courses {
		if formatted := r.FormatCourse(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *CourseResourceImpl) FormatDetailCourses(courses []*models.Course) []*prot.Course {
	result := make([]*prot.Course, 0, len(courses))
	for _, u := range courses {
		if formatted := r.FormatCourseDetail(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *CourseResourceImpl) FormatModelCourse(course *prot.CourseRequest) *models.Course {
	if course == nil {
		return nil
	}

	startDate, err := time.Parse("2006-01-02", course.StartDate)
	if err != nil {
		startDate = time.Now()
	}

	endDate := startDate
	if course.Duration > 0 {
		endDate = startDate.Add(time.Duration(course.Duration*7) * time.Hour * 24)
	}

	imageUrl := utils.StripDomain(course.Image, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(imageUrl, models.Storage)

	state := course.State

	if state == "" {
		now := time.Now()

		if now.Before(startDate) {
			state = "coming"
		} else if now.After(endDate) {
			state = "finished"
		} else {
			state = "active"
		}
	}

	return &models.Course{
		ID:          int64(course.Id),
		SubjectId:   int64(course.SubjectId),
		ProgramId:   int64(course.ProgramId),
		Name:        course.Name,
		Description: course.Description,
		Type:        course.Type,
		ObjectTitle: course.ObjectTitle,
		Level:       course.Level,
		Duration:    int(course.Duration),
		ImageInfo:   imageInfo,
		Status:      course.Status,
		Target:      course.Target,
		StartDate:   startDate,
		EndDate:     endDate,
		Time:        course.Time,
		State:       state,
	}
}

func GetProgress(course models.Course) int32 {
	now := time.Now()

	start := time.Date(course.StartDate.Year(), course.StartDate.Month(), course.StartDate.Day(), 0, 0, 0, 0, course.StartDate.Location())
	end := time.Date(course.EndDate.Year(), course.EndDate.Month(), course.EndDate.Day(), 0, 0, 0, 0, course.EndDate.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var progress int32

	if today.Before(start) {
		progress = 0
	} else if today.After(end) {
		progress = 100
	} else {
		totalDays := int(end.Sub(start).Hours() / 24)
		passedDays := int(today.Sub(start).Hours() / 24)

		if totalDays > 0 {
			progress = int32((float64(passedDays) / float64(totalDays)) * 100)
		} else {
			progress = 0
		}
	}

	return progress
}
