package resources

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"sort"
)

type ProgramResource interface {
	FormatProgram(program *models.Program) *prot.Program
	FormatPrograms(programs []*models.Program) []*prot.Program
	FormatModelProgram(program *prot.ProgramRequest) *models.Program
}

type ProgramResourceImpl struct {
	HideLessonIds     []int64
	StudyingLessonIds []int64
	CompleteLessonIds []int64
	LessonSchedules   []models.LessonSchedule
}

func NewProgramResource() ProgramResource {
	return &ProgramResourceImpl{
		HideLessonIds:     []int64{},
		StudyingLessonIds: []int64{},
		CompleteLessonIds: []int64{},
		LessonSchedules:   []models.LessonSchedule{},
	}
}

func (r *ProgramResourceImpl) FormatPrograms(programs []*models.Program) []*prot.Program {
	result := make([]*prot.Program, 0, len(programs))
	for _, u := range programs {
		if formatted := r.FormatProgram(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ProgramResourceImpl) FormatProgram(program *models.Program) *prot.Program {
	if program == nil {
		return nil
	}

	chapterResource := NewChapterResource()

	if impl, ok := chapterResource.(*ChapterResourceImpl); ok {
		impl.HideLessonIds = r.HideLessonIds
		impl.StudyingLessonIds = r.StudyingLessonIds
		impl.CompleteLessonIds = r.CompleteLessonIds
		impl.LessonSchedules = r.LessonSchedules
	}

	chapters := make([]*models.Chapter, 0, len(program.Chapters))
	for i := range program.Chapters {
		chapters = append(chapters, &program.Chapters[i])
	}

	sort.Slice(chapters, func(i, j int) bool {
		if chapters[i].SortPosition == chapters[j].SortPosition {
			return chapters[i].ID < chapters[j].ID
		}
		return chapters[i].SortPosition < chapters[j].SortPosition
	})

	var totalStudents int32
	courseResource := NewCourseResource()
	courses := make([]*models.Course, 0, len(program.Courses))
	for i, course := range program.Courses {
		totalStudents += course.CurrentStudents
		courses = append(courses, &program.Courses[i])
	}

	programVTGInfo := &prot.ProgramVTGInfo{
		Code:            "",
		ManagementCode:  "",
		NumberOfCredits: 0,
		Time:            "",
		TheoreticalTime: "",
		DiscussionTime:  "",
		PracticeTime:    "",
		TestTime:        "",
		StudyModule:     "",
		NumberOfFrequent: 0,
		NumberOfEvaluate: 0,
	}

	if program.Detail.VTGData != (models.ProgramVTGData{}) {
		programVTGInfo.Code = program.Detail.VTGData.Code
		programVTGInfo.ManagementCode = program.Detail.VTGData.ManagementCode
		programVTGInfo.NumberOfCredits = int32(program.Detail.VTGData.NumberOfCredits)
		programVTGInfo.Time = program.Detail.VTGData.Time
		programVTGInfo.TheoreticalTime = program.Detail.VTGData.TheoreticalTime
		programVTGInfo.DiscussionTime = program.Detail.VTGData.DiscussionTime
		programVTGInfo.PracticeTime = program.Detail.VTGData.PracticeTime
		programVTGInfo.TestTime = program.Detail.VTGData.TestTime
		programVTGInfo.StudyModule = program.Detail.VTGData.StudyModule
		programVTGInfo.NumberOfFrequent = int32(program.Detail.VTGData.NumberOfFrequent)
		programVTGInfo.NumberOfEvaluate = int32(program.Detail.VTGData.NumberOfEvaluate)
	}

	var subjectId int64

	if len(program.Subjects) > 0 {
		subjectId = program.Subjects[0].ID
	}

	subjectResource := NewSubjectResource()

	subjects := make([]*models.Subject, 0, len(program.Subjects))
	for i := range program.Subjects {
		subjects = append(subjects, &program.Subjects[i])
	}

	return &prot.Program{
		Id:              int64(program.ID),
		SubjectId:       subjectId,
		Name:            program.Name,
		Description:     program.Description,
		Courses:         courseResource.FormatCourses(courses),
		Image:           utils.StaticURL(program.ImageInfo.Path, models.Storage),
		Status:          program.Status,
		Target:          program.Target,
		Duration:        int32(program.Duration),
		CurrentStudents: totalStudents,
		Chapters:        chapterResource.FormatDetailChapters(chapters),
		Subjects:        subjectResource.FormatSubjects(subjects),
		Type:            program.Type,
		VtgInfo:         programVTGInfo,
		BookUrl:         program.BookUrl,
		CreatedAt:       program.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       program.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *ProgramResourceImpl) FormatModelProgram(program *prot.ProgramRequest) *models.Program {
	if program == nil {
		return nil
	}

	imageUrl := utils.StripDomain(program.Image, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	imageInfo := mediaRepo.GetMediaInfo(imageUrl, models.Storage)

	programType := program.Type

	if programType != models.TypePractice && programType != models.TypeModule && programType != models.TypeIntegration {
		programType = models.TypeTheory
	}

	var programVTGInfo models.ProgramVTGData

	if program.VtgInfo != nil {
		programVTGInfo = models.ProgramVTGData{
			Code:            program.VtgInfo.Code,
			ManagementCode:  program.VtgInfo.ManagementCode,
			NumberOfCredits: int(program.VtgInfo.NumberOfCredits),
			Time:            program.VtgInfo.Time,
			TheoreticalTime: program.VtgInfo.TheoreticalTime,
			DiscussionTime:  program.VtgInfo.DiscussionTime,
			PracticeTime:    program.VtgInfo.PracticeTime,
			TestTime:        program.VtgInfo.TestTime,
			StudyModule:     program.VtgInfo.StudyModule,
			NumberOfFrequent: int(program.VtgInfo.NumberOfFrequent),
			NumberOfEvaluate: int(program.VtgInfo.NumberOfEvaluate),
		}
	}

	programDetail := models.ProgramDetail{
		VTGData: programVTGInfo,
	}

	return &models.Program{
		ID: int64(program.Id),
		// SubjectId:   int64(program.SubjectId),
		Name:        program.Name,
		Description: program.Description,
		ImageInfo:   imageInfo,
		Status:      program.Status,
		Target:      program.Target,
		Duration:    int(program.Duration),
		Type:        programType,
		Detail:      programDetail,
		BookUrl:     program.BookUrl,
	}
}
