package controllers

import (
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LessonController struct {
	service services.LessonService
	*GenericController[models.Lesson, prot.Lesson, *prot.LessonRequest]
}

func NewLessonController(service services.LessonService) *LessonController {
	lessonResource := resources.NewLessonResource()
	lessonResourceAdapter := NewLessonResourceAdapter(lessonResource)

	genericController := NewGenericController(
		service,
		lessonResourceAdapter,
		func() *prot.LessonRequest {
			return &prot.LessonRequest{}
		},
		func(lessons []*prot.Lesson, totalCount uint64) interface{} {
			return &prot.LessonsResponse{
				Lessons:    lessons,
				TotalCount: totalCount,
			}
		},
	)

	ctl := &LessonController{
		GenericController: genericController,
		service:           service,
	}

	ctl.GenericController.WithRespondListHook(func(c *gin.Context, items []models.Lesson, totalCount int64, err error) {
		ctl.RespondList(c, items, totalCount, err)
	})

	ctl.GenericController.WithRespondDetailHook(func(c *gin.Context, item *models.Lesson) {
		ctl.RespondDetail(c, item)
	})

	return ctl
}

func (lc *LessonController) Completion(c *gin.Context) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId != models.StudentRoleId {
		utils.Respond(c, nil, fmt.Errorf("%s", i18n.Localize("messages.role_invalid")), "messages.role_invalid")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	lessonCompletion, err, _ := utils.GetBody[*prot.LessonCompletion](c, func() *prot.LessonCompletion {
		return &prot.LessonCompletion{}
	})

	lessonCompletion.Id = int64(id)
	lessonCompletion.StudentId = int64(utils.GetCurrentUserId(c))

	completion, err := lc.service.Completion(c, lessonCompletion)

	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, completion, nil, "")
}

func (lc *LessonController) Studying(c *gin.Context) {
	roleId := utils.GetCurrentRoleId(c)

	if roleId != models.StudentRoleId {
		utils.Respond(c, nil, fmt.Errorf("%s", i18n.Localize("messages.role_invalid")), "messages.role_invalid")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	req, err, _ := utils.GetBody[*prot.LessonStudying](c, func() *prot.LessonStudying {
		return &prot.LessonStudying{}
	})

	req.Id = int64(id)
	req.StudentId = int64(utils.GetCurrentUserId(c))

	studying, err := lc.service.Studying(c, req)

	if err != nil {
		utils.Respond(c, nil, err, err.Error())
		return
	}

	utils.Respond(c, studying, nil, "")
}

func (lc *LessonController) RespondList(c *gin.Context, lessons []models.Lesson, totalCount int64, err error) {
	var lessonPtrs []*models.Lesson
	for i := range lessons {
		lessonPtrs = append(lessonPtrs, &lessons[i])
	}

	hideLessonIds := lc.service.HideLessonIds(c)
	studyingLessonIds := lc.service.StudyingLessonIds(c, 0, 0)
	completeLessonIds := lc.service.CompletionLessonIds(c, 0, 0)
	lessonResource := resources.NewLessonResource()
	if impl, ok := lessonResource.(*resources.LessonResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
	}
	lessonsResponse := lessonResource.FormatLessons(lessonPtrs)

	list := &prot.LessonsResponse{
		Lessons:    lessonsResponse,
		TotalCount: uint64(totalCount),
	}

	utils.Respond(c, list, err, "")
}

func (lc *LessonController) RespondDetail(c *gin.Context, lesson *models.Lesson) {
	hideLessonIds := lc.service.HideLessonIds(c)
	studyingLessonIds := lc.service.StudyingLessonIds(c, 0, 0)
	completeLessonIds := lc.service.CompletionLessonIds(c, 0, 0)
	lessonResource := resources.NewLessonResource()
	courseId, _ := strconv.Atoi(c.Query("course_id"))

	// Get vocabularies for this lesson
	flashcardService := services.NewFlashcardService()
	lessonVocabularies, _ := flashcardService.GetLessonVocabularies(lesson.ID, nil)

	if impl, ok := lessonResource.(*resources.LessonResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
		impl.Flashcard = lessonVocabularies
		impl.CourseId = int64(courseId)
	}

	formattedLesson := lessonResource.FormatLesson(lesson)

	utils.Respond(c, formattedLesson, nil, "")
}

func (lc *LessonController) StoreLessonSchedules(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.LessonSchedulesInfoRequest](c, func() *prot.LessonSchedulesInfoRequest {
		return &prot.LessonSchedulesInfoRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	courseID, err := lc.service.StoreLessonSchedules(c, req)

	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	if courseID > 0 {
		filters := map[string]interface{}{
			"course_id":      courseID,
			"week_by_course": true,
		}

		data, _, _, _, err := lc.service.LessonSchedules(c, filters)
		if err == nil {
			lessonScheduleResource := resources.NewLessonScheduleResource()
			groupedResponse := lessonScheduleResource.FormatLessonSchedulesInfo(data)
			utils.Respond(c, groupedResponse, nil, "")
			return
		}
	}

	utils.Respond(c, nil, nil, "")
}

func (lc *LessonController) LessonSchedules(c *gin.Context) {
	filters := map[string]interface{}{}

	if dateStr := c.Query("date"); dateStr != "" {
		weekRepo := repositories.NewWeekRepository()
		parsedDate, _ := utils.ParseDate(dateStr)
		week, _ := weekRepo.GetByDate(parsedDate)

		filters["week_id"] = week.ID

	} else if weekIDStr := c.Query("week_id"); weekIDStr != "" {
		if weekID, err := strconv.ParseInt(weekIDStr, 10, 64); err == nil {
			filters["week_id"] = weekID
		}
	}

	if courseIDStr := c.Query("course_id"); courseIDStr != "" {
		if courseID, err := strconv.ParseInt(courseIDStr, 10, 64); err == nil {
			filters["course_id"] = courseID

			if byWeekParam := c.Query("week_by_course"); byWeekParam == "true" || byWeekParam == "1" || byWeekParam == "yes" {
				filters["week_by_course"] = true
			}
		}
	}

	if semesterIdStr := c.Query("semester_id"); semesterIdStr != "" {
		if semesterId, err := strconv.ParseInt(semesterIdStr, 10, 64); err == nil {
			filters["semester_id"] = semesterId
		}
	}

	if byUserParam := c.Query("by_user"); byUserParam == "true" || byUserParam == "1" || byUserParam == "yes" {
		userID := utils.GetCurrentUserId(c)
		filters["user_id"] = userID
	}

	data, beginDate, holidayWeeks, autoIndex, err := lc.service.LessonSchedules(c, filters)

	if err == nil {
		lessonScheduleResource := resources.NewLessonScheduleResource()

		if weekIDVal, ok := filters["week_id"]; !ok || weekIDVal == nil || weekIDVal == int64(0) {
			if impl, ok := lessonScheduleResource.(*resources.LessonScheduleResourceImpl); ok && beginDate != nil {
				impl.BeginTime = beginDate
				impl.HolidayWeeks = holidayWeeks
				impl.AutoIndex = autoIndex
			}

			groupedResponse := lessonScheduleResource.FormatLessonSchedulesInfo(data)
			utils.Respond(c, groupedResponse, err, "")
			return
		} else {
			var selected dto.LessonSchedule

			for _, item := range data.Groups {
				if int64(item.WeekId) == filters["week_id"] {
					selected = item
					break
				}
			}

			groupedResponse := lessonScheduleResource.FormatLessonSchedulesByWeek(selected)
			utils.Respond(c, groupedResponse, err, "")
			return
		}
	}

	utils.Respond(c, nil, err, "")
}

func (lc *LessonController) GetLessonPlanByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	lessonPlans, err := lc.service.GetLessonPlanByCourse(c, int64(id), int64(courseId))

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) GetExamByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	lessonPlans, err := lc.service.GetExamByCourse(c, int64(id), int64(courseId))

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) GetHomeworkByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	lessonPlans, err := lc.service.GetHomeworkByCourse(c, int64(id), int64(courseId))

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) GetExerciseByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	lessonPlans, err := lc.service.GetExerciseByCourse(c, int64(id), int64(courseId))

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) StoreLessonPlanByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	req, err, _ := utils.GetBody[*prot.LessonPlanByCourse](c, func() *prot.LessonPlanByCourse {
		return &prot.LessonPlanByCourse{}
	})

	lessonPlans, err := lc.service.StoreLessonPlanByCourse(c, int64(id), int64(courseId), req)

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) StoreExamByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	req, err, _ := utils.GetBody[*prot.ExamByCourse](c, func() *prot.ExamByCourse {
		return &prot.ExamByCourse{}
	})

	lessonPlans, err := lc.service.StoreExamByCourse(c, int64(id), int64(courseId), req)

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) StoreHomeworkByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	req, err, _ := utils.GetBody[*prot.HomeworkByCourse](c, func() *prot.HomeworkByCourse {
		return &prot.HomeworkByCourse{}
	})

	lessonPlans, err := lc.service.StoreHomeworkByCourse(c, int64(id), int64(courseId), req)

	utils.Respond(c, lessonPlans, err, "")
}

func (lc *LessonController) StoreExerciseByCourse(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("lesson_id"))
	courseId, _ := strconv.Atoi(c.Param("course_id"))

	req, err, _ := utils.GetBody[*prot.ExerciseByCourse](c, func() *prot.ExerciseByCourse {
		return &prot.ExerciseByCourse{}
	})

	lessonPlans, err := lc.service.StoreExerciseByCourse(c, int64(id), int64(courseId), req)

	utils.Respond(c, lessonPlans, err, "")
}
