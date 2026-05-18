package services

import (
	"be-Clever School/config"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"errors"

	"github.com/gin-gonic/gin"
)

type ChapterService interface {
	GetAll(c *gin.Context) ([]models.Chapter, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Chapter, error)
	Create(c *gin.Context, req *prot.ChapterRequest) (*models.Chapter, error)
	Update(c *gin.Context, req *prot.ChapterRequest) (*models.Chapter, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Chapter, error)
	SortLessons(c *gin.Context, id int, req *prot.ChapterLessonSort) error
}

type chapterService struct {
	repo repositories.ChapterRepository
}

func NewChapterService(repo repositories.ChapterRepository) ChapterService {
	return &chapterService{repo: repo}
}

func (s *chapterService) GetAll(c *gin.Context) ([]models.Chapter, int64, error) {
	allowedFilters := []string{"status", "program_id"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"title", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	chapters, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return chapters, rows, nil
}

func (s *chapterService) GetByID(c *gin.Context, id int) (*prot.Chapter, error) {
	s.repo.SetPreload([]string{
		"Lessons",
	})

	s.repo.SetContext(c)
	chapter, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	return s.RespondDetail(c, chapter)
}

func (s *chapterService) Create(c *gin.Context, req *prot.ChapterRequest) (*models.Chapter, error) {
	chapterResource := resources.NewChapterResource()
	chapter := chapterResource.FormatModelChapter(req)

	s.repo.SetContext(c)

	var sortPosition int
	if chapter.ProgramId != 0 {
		sortPosition, _ = s.repo.GetPositionByIdAndProgram(chapter.ProgramId, 0)
	}
	chapter.SortPosition = sortPosition

	err := s.Validate(chapter)
	if err != nil {
		return nil, err
	}

	if err = s.repo.Create(chapter); err != nil {
		return nil, err
	}

	id := int(chapter.ID)

	err = s.StoreLessons(c, int64(id), req, chapter)
	if err != nil {
		config.Log.Error(err)
	}

	s.repo.SetPreload([]string{
		"Lessons",
	})

	newChapter, _ := s.repo.FindNewByID(id)

	return newChapter, nil
}

func (s *chapterService) Update(c *gin.Context, req *prot.ChapterRequest) (*models.Chapter, error) {
	chapterResource := resources.NewChapterResource()
	chapter := chapterResource.FormatModelChapter(req)

	s.repo.SetContext(c)
	s.repo.SetOmit([]string{
		"clone_info",
	})

	var sortPosition int
	if chapter.ProgramId != 0 {
		sortPosition, _ = s.repo.GetPositionByIdAndProgram(chapter.ProgramId, req.Id)
	}
	chapter.SortPosition = sortPosition

	err := s.Validate(chapter)
	if err != nil {
		return nil, err
	}

	if err = s.repo.Update(chapter); err != nil {
		return nil, err
	}

	id := int(chapter.ID)

	err = s.StoreLessons(c, int64(id), req, chapter)
	if err != nil {
		config.Log.Error(err)
	}

	s.repo.SetPreload([]string{
		"Lessons",
	})

	updatedChapter, _ := s.repo.FindNewByID(id)

	return updatedChapter, nil
}

func (s *chapterService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	return s.repo.Delete(id)
}

func (s *chapterService) Restore(c *gin.Context, id int) (*models.Chapter, error) {
	s.repo.SetContext(c)
	chapter, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return chapter, nil
}

func (s *chapterService) SortLessons(c *gin.Context, id int, req *prot.ChapterLessonSort) error {
	repo := repositories.NewLessonRepository()
	for index, lesson := range req.Lessons {
		repo.UpdatePositionById(id, int(lesson.Id), index)
	}
	return nil
}

func (s *chapterService) StoreLessons(c *gin.Context, id int64, req *prot.ChapterRequest, chapter *models.Chapter) error {
	var lessonIds []int64
	var isVtg = config.LoadConfig().IsVtg

	if isVtg && chapter.Time == 5 {
		return s.repo.UpdateLessonVtg(chapter)
	}

	for index, lesson := range req.Lessons {
		lessonRepo := repositories.NewLessonRepository()

		if lesson.Id != 0 {
			s.repo.UpdateLessonChapterId(id, lesson.Id)
			lessonIds = append(lessonIds, int64(lesson.Id))
		} else {
			lessonModel := models.Lesson{
				Title:        lesson.Title,
				ChapterID:    id,
				SortPosition: index,
			}

			lessonRepo.SetContext(c)
			if err := lessonRepo.Create(&lessonModel); err != nil {
				return err
			}

			lessonIds = append(lessonIds, lessonModel.ID)
		}
	}

	s.repo.ClearOldLessonByChapterId(id, lessonIds)

	return nil
}

func (cs *chapterService) RespondDetail(c *gin.Context, chapter *models.Chapter) (*prot.Chapter, error) {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := NewLessonService(lessonRepo)
	hideLessonIds := lessonService.HideLessonIds(c)
	studyingLessonIds := lessonService.StudyingLessonIds(c, 0, chapter.ID)
	completeLessonIds := lessonService.CompletionLessonIds(c, 0, chapter.ID)

	chapterResource := resources.NewChapterResource()
	if impl, ok := chapterResource.(*resources.ChapterResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
	}
	formattedChapter := chapterResource.FormatChapterDetail(chapter)

	return formattedChapter, nil
}

func (cs *chapterService) Validate(chapter *models.Chapter) error {
	var isVtg = config.LoadConfig().IsVtg

	if !isVtg {
		return nil
	}

	if chapter.Time < 0 || chapter.Time%5 != 0 {
		return errors.New(i18n.Localize("messages.chapter_time_invalid"))
	}

	isValid := cs.repo.ValidTime(chapter.ProgramId, chapter.ID, chapter.Time)

	if !isValid {
		return errors.New(i18n.Localize("messages.chapter_time_overlap"))
	}
	return nil
}
