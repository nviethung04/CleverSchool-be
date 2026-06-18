package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
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
		utils.Respond(c, nil, err,  "messages.id_invalid", 400)
		return
	}

	users, err := cc.service.GetUsers(c, int64(id))
	cc.respondWithUsers(c, users, int64(id), err)
}

func (cc *CourseController) StoreUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err,  "messages.id_invalid", 400)
		return
	}

	users, err := cc.service.StoreUsers(c, int64(id))
	cc.respondWithUsers(c, users, int64(id), err)
}

func (cc *CourseController) AddUsers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err,  "messages.id_invalid", 400)
		return
	}

	students, err := cc.service.AddUsers(c, int64(id))
	cc.respondWithUsers(c, students, int64(id), err)
}

func (cc *CourseController) respondWithUsers(c *gin.Context, users []models.User, courseId int64, err error) {
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
		utils.Respond(c, nil, err,  "messages.id_invalid", 400)
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

func (cc *CourseController) GetCourseFamily(c *gin.Context) {
	courseID, err := strconv.ParseInt(c.Query("course_id"), 10, 64)
	if err != nil || courseID <= 0 {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	family, err := cc.service.GetCourseFamily(c, courseID)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data", 404)
		return
	}

	utils.Respond(c, family, nil, "")
}

func (cc *CourseController) SyncAllFamilyCourses(c *gin.Context) {
	req, err, message := utils.GetBody[*prot.SyncAllFamilyRequest](c, func() *prot.SyncAllFamilyRequest {
		return &prot.SyncAllFamilyRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}
	if req.CourseId <= 0 {
		utils.Respond(c, nil, err, "messages.id_invalid", 400)
		return
	}

	if err := cc.service.SyncAllFamilyCourses(c, req.CourseId); err != nil {
		utils.Respond(c, nil, err, "messages.create_data")
	}
}

func (cc *CourseController) RespondList(c *gin.Context, items []models.Course, totalCount int64, err error) {
	var coursePtrs []*models.Course
	for i := range items {
		coursePtrs = append(coursePtrs, &items[i])
	}

	courseResource := resources.NewCourseResource()
	var coursesResponse []*prot.Course

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
	completeLessonIds := lessonService.CompletionLessonIds(c, course.ID, 0)

	lessonSchedules, _ := lessonRepo.GetLessonSchedulesByCourse(course.ID)

	req, _, _ := utils.GetBody[*prot.CourseRequest](c, func() *prot.CourseRequest {
		return &prot.CourseRequest{}
	})

	copyScheduleResponse := cc.service.CopySchedule(course.ID, req.TargetCourseId, course.ProgramId)

	courseResource := resources.NewCourseResource()
	if impl, ok := courseResource.(*resources.CourseResourceImpl); ok {
		impl.CompleteLessonIds = completeLessonIds
		impl.LessonSchedules = lessonSchedules
		impl.CopyScheduleResponse = copyScheduleResponse
	}
	formattedCourse := courseResource.FormatCourseDetail(course)

	utils.Respond(c, formattedCourse, nil, "")
}
