package resources

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"sort"
)

type ProgramResource interface {
	FormatProgram(program *models.Program) *prot.Program
	FormatPrograms(programs []*models.Program) []*prot.Program
	FormatModelProgram(program *prot.ProgramRequest) *models.Program
}

type ProgramResourceImpl struct {
	CompleteLessonIds []int64
	LessonSchedules   []models.LessonSchedule
}

func NewProgramResource() ProgramResource {
	return &ProgramResourceImpl{
		CompleteLessonIds: []int64{},
		LessonSchedules: []models.LessonSchedule{},
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

	return &prot.Program{
		Id:              int64(program.ID),
		Name:            program.Name,
		Description:     program.Description,
		Courses:         courseResource.FormatCourses(courses),
		Image:           utils.StaticURL(program.ImageInfo.Path, models.Storage),
		Status:          program.Status,
		Target:          program.Target,
		Duration:		 int32(program.Duration),
		CurrentStudents: totalStudents,
		Chapters:        chapterResource.FormatDetailChapters(chapters),
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

	return &models.Program{
		ID:          int64(program.Id),
		SubjectId:   int64(program.SubjectId),
		Name:        program.Name,
		Description: program.Description,
		ImageInfo:   imageInfo,
		Status:      program.Status,
		Target:      program.Target,
		Duration:	 int(program.Duration),
	}
}
