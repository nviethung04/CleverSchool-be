package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ChapterController struct {
	service services.ChapterService
	*GenericController[models.Chapter, prot.Chapter, *prot.ChapterRequest]
}

func NewChapterController(service services.ChapterService) *ChapterController {
	chapterResource := resources.NewChapterResource()
	chapterResourceAdapter := NewChapterResourceAdapter(chapterResource)

	genericController := NewGenericController(
		service,
		chapterResourceAdapter,
		func() *prot.ChapterRequest {
			return &prot.ChapterRequest{}
		},
		func(chapters []*prot.Chapter, totalCount uint64) interface{} {
			return &prot.ChaptersResponse{
				Chapters:   chapters,
				TotalCount: totalCount,
			}
		},
	)

	ctl := &ChapterController{
		GenericController: genericController,
		service:           service,
	}

	ctl.GenericController.WithRespondDetailHook(func(c *gin.Context, item *models.Chapter) {
		ctl.RespondDetail(c, item)
	})

	ctl.GenericController.WithUsedError(func() bool {
		return true
	})

	return ctl
}

// Không cần override Create và Update nữa
// GenericController đã handle đủ sau khi thêm program_id vào protobuf

func (cc *ChapterController) SortLessons(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	req, err, message := utils.GetBody[*prot.ChapterLessonSort](c, func() *prot.ChapterLessonSort {
		return &prot.ChapterLessonSort{}
	})

	err = cc.service.SortLessons(c, id, req)

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	utils.Respond(c, req, err, "")
}

func (cc *ChapterController) RespondDetail(c *gin.Context, chapter *models.Chapter) {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := services.NewLessonService(lessonRepo)
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

	utils.Respond(c, formattedChapter, nil, "")
}
