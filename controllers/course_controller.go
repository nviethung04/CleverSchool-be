package controllers

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseController struct {
	*GenericController[models.Course, prot.Course, *prot.CourseRequest]
	service services.CourseService
}

func NewCourseController(service services.CourseService) *CourseController {
	courseResource := resources.NewCourseResource()
	courseResourceAdapter := NewCourseResourceAdapter(courseResource)

	genericController := NewGenericController(
		service,
		courseResourceAdapter,
		func() *prot.CourseRequest {
			return &prot.CourseRequest{}
		},
		func(courses []*prot.Course, totalCount uint64) interface{} {
			return &prot.CoursesResponse{
				Courses:    courses,
				TotalCount: int64(totalCount),
			}
		},
	)

	ctl := &CourseController{
		GenericController: genericController,
		service:           service,
	}

	ctl.GenericController.WithRespondDetailHook(func(c *gin.Context, item *models.Course) {
		ctl.RespondDetail(c, item)
	})

	ctl.GenericController.WithRespondListHook(func(c *gin.Context, items []models.Course, totalCount int64, err error) {
		ctl.RespondList(c, items, totalCount, err)
	})

	return ctl
}

func (cc *CourseController) GetUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	users, failedUserIds, err := cc.service.GetUsers(c, int64(id))
	cc.respondWithUsers(c, users, int64(id), failedUserIds, err)
}

func (cc *CourseController) StoreUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	users, failedUserIds, err := cc.service.StoreUsers(c, int64(id))
	cc.respondWithUsers(c, users, int64(id), failedUserIds, err)
}

func (cc *CourseController) AddUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	students, failedUserIds, err := cc.service.AddUsers(c, int64(id))
	cc.respondWithUsers(c, students, int64(id), failedUserIds, err)
}

func (cc *CourseController) respondWithUsers(c *gin.Context, users []models.User, courseId int64, failedUserIds []int64, err error) {
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}

	ptrs := make([]*models.User, len(users))
	for i := range users {
		ptrs[i] = &users[i]
	}

	userResource := resources.NewUserResource()
	if impl, ok := userResource.(*resources.UserResourceImpl); ok {
		impl.CourseId = courseId
		impl.FailedUserIds = failedUserIds
	}
	formattedUsers := userResource.FormatUsers(ptrs)

	list := &prot.UsersResponse{
		Users:      formattedUsers,
		TotalCount: int64(len(formattedUsers)),
	}

	utils.Respond(c, list, nil, "")
}

func (cc *CourseController) GetScore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	score, err := cc.service.GetScore(c, int64(id))

	utils.Respond(c, score, err, "")
}

func (cc *CourseController) ResyncSchedule(c *gin.Context) {
	req, _, _ := utils.GetBody[*prot.ResyncScheduleRequest](c, func() *prot.ResyncScheduleRequest {
		return &prot.ResyncScheduleRequest{}
	})

	for _, id := range req.CourseIds {
		copyScheduleResponse := cc.service.CopySchedule(id, req.CourseId, 0)

		if !copyScheduleResponse.CopyScheduleSuccess {
			utils.Respond(c, &prot.ResyncScheduleResponse{
				Message: "Copy schedule failed, course id: " + strconv.Itoa(int(id)) + ". " + copyScheduleResponse.CopyScheduleMessage,
			}, nil, "")
			return
		}
	}

	utils.Respond(c, &prot.ResyncScheduleResponse{
		Message: "Ok",
	}, nil, "")
}

func (cc *CourseController) RespondList(c *gin.Context, items []models.Course, totalCount int64, err error) {
	var coursePtrs []*models.Course
	for i := range items {
		coursePtrs = append(coursePtrs, &items[i])
	}

	courseResource := resources.NewCourseResource()
	var coursesResponse []*prot.Course

	userId := utils.GetCurrentUserId(c)
	if impl, ok := courseResource.(*resources.CourseResourceImpl); ok {
		impl.NotEligibleForFinalExamIds = cc.service.NotEligibleForFinalExamIds(c, 0, int64(userId))
	}

	if byUserParam := c.Query("by_user"); byUserParam != "" {
		coursesResponse = courseResource.FormatDetailCourses(coursePtrs)
	} else {
		coursesResponse = courseResource.FormatCourses(coursePtrs)
	}

	list := &prot.CoursesResponse{
		Courses:    coursesResponse,
		TotalCount: int64(totalCount),
	}

	utils.Respond(c, list, err, "")
}

func (cc *CourseController) RespondDetail(c *gin.Context, course *models.Course) {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := services.NewLessonService(lessonRepo)
	hideLessonIds := lessonService.HideLessonIds(c)
	studyingLessonIds := lessonService.StudyingLessonIds(c, course.ID, 0)
	completeLessonIds := lessonService.CompletionLessonIds(c, course.ID, 0)

	lessonSchedules, _ := lessonRepo.GetLessonSchedulesByCourse(course.ID)

	req, _, _ := utils.GetBody[*prot.CourseRequest](c, func() *prot.CourseRequest {
		return &prot.CourseRequest{}
	})

	copyScheduleResponse := cc.service.CopySchedule(course.ID, req.ParentCourseId, course.ProgramId)
	userId := utils.GetCurrentUserId(c)

	courseResource := resources.NewCourseResource()
	if impl, ok := courseResource.(*resources.CourseResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
		impl.LessonSchedules = lessonSchedules
		impl.CopyScheduleResponse = copyScheduleResponse
		impl.NotEligibleForFinalExamIds = cc.service.NotEligibleForFinalExamIds(c, course.ID, int64(userId))
	}
	formattedCourse := courseResource.FormatCourseDetail(course)

	utils.Respond(c, formattedCourse, nil, "")
}

func (fc *CourseController) Export(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)

	go func() {
		url, err := fc.service.Export(cCp)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (cc *CourseController) ListCoursesFamily(c *gin.Context) {
	programIdStr := c.Query("program_id")
	if programIdStr == "" {
		utils.Respond(c, nil, errors.New("program_id is required"), "messages.program_id_required", 400)
		return
	}

	programId, err := strconv.ParseInt(programIdStr, 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.program_id_invalid", 400)
		return
	}

	result, err := cc.service.GetCoursesFamilyByProgramId(programId)
	if err != nil {
		utils.Respond(c, nil, err, "", 500)
		return
	}

	utils.Respond(c, result, nil, "")
}

func (fc *CourseController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := fc.service.Import(cCp, file)
		if err != nil {
			config.Log.Error("Course import failed", "error", err)
		} else {
			config.Log.Info("Course import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}

func (fc *CourseController) ExportUsers(c *gin.Context) {
	cCp := c.Copy()
	resultChan := make(chan *dto.MyExportResult)
	courseId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	go func() {
		url, err := fc.service.ExportUsers(cCp, courseId)
		resultChan <- &dto.MyExportResult{Url: url, Err: err}
	}()

	res := <-resultChan
	utils.Respond(c, &prot.Export{Url: res.Url}, res.Err, "")
}

func (fc *CourseController) ImportUsers(c *gin.Context) {
	courseId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf("file is required"), "file is required", 400)
		return
	}

	cCp := c.Copy()
	go func() {
		err := fc.service.ImportUsers(cCp, courseId, file)
		if err != nil {
			config.Log.Error("Course users import failed", "error", err)
		} else {
			config.Log.Info("Course users import finished successfully")
		}
	}()

	utils.Respond(c, &prot.Import{Message: i18n.Localize("messages.import_complete")}, nil, "")
}
