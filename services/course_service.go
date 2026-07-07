package services

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CourseService interface {
	GetAll(c *gin.Context) ([]models.Course, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Course, error)
	Create(c *gin.Context, req *prot.CourseRequest) (*models.Course, error)
	Update(c *gin.Context, req *prot.CourseRequest) (*models.Course, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Course, error)
	GetUsers(c *gin.Context, id int64) ([]models.User, error)
	StoreUsers(c *gin.Context, id int64) ([]models.User, error)
	AddUsers(c *gin.Context, id int64) ([]models.User, error)
	GetScore(c *gin.Context, id int64) (*prot.ScoreByCourse, error)
	CopySchedule(id, targetCourseId, programId int64) dto.CopyScheduleResponse
	GetCourseFamily(c *gin.Context, courseID int64) (*prot.CourseFamilyResponse, error)
	SyncAllFamilyCourses(c *gin.Context, courseID int64) error
}

type courseService struct {
	repo                      repositories.CourseRepository
	lessonScheduleCopyService LessonScheduleCopySharedService
}

func NewCourseService(repo repositories.CourseRepository) CourseService {
	return &courseService{
		repo:                      repo,
		lessonScheduleCopyService: NewLessonScheduleCopySharedService(),
	}
}

func (s *courseService) GetAll(c *gin.Context) ([]models.Course, int64, error) {
	allowedFilters := []string{"status", "program_id", "description", "state"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Schools",
		"Program",
		"Program.Chapters",
		"Program.Chapters.Lessons",
	})

	courses, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return courses, rows, nil
}

func (s *courseService) GetByID(c *gin.Context, id int) (*prot.Course, error) {
	if id == 0 {
		return nil, errors.New("id is required")
	}

	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Schools",
		"Subject",
		"CourseRefStudyShifts",
		"CourseRefStudyShifts.StudyShift",
		"Users",
		"Users.Roles",
		"Semesters",
		"Program",
		"Program.Chapters",
		"Program.Chapters.Lessons",
	})

	course, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.RespondDetail(c, course)
}

func (s *courseService) Create(c *gin.Context, req *prot.CourseRequest) (*models.Course, error) {
	courseResource := resources.NewCourseResource()
	course := courseResource.FormatModelCourse(req)

	s.repo.SetContext(c)

	err := s.repo.Create(course)
	if err != nil {
		return nil, err
	}

	s.StoreSchool(int64(course.ID), req)
	s.StoreStudyShifts(int64(course.ID), req)
	s.StoreSemester(int64(course.ID), req)

	id := int(course.ID)

	s.repo.SetPreload([]string{
		"Schools",
		"Subject",
		"CourseRefStudyShifts",
		"CourseRefStudyShifts.StudyShift",
		"Users",
		"Users.Roles",
		"Semesters",
		"Program",
		"Program.Chapters",
		"Program.Chapters.Lessons",
	})

	newCourse, _ := s.repo.FindNewByID(id)

	return newCourse, nil
}

func (s *courseService) Update(c *gin.Context, req *prot.CourseRequest) (*models.Course, error) {
	courseResource := resources.NewCourseResource()
	course := courseResource.FormatModelCourse(req)

	s.repo.SetContext(c)
	s.repo.SetOmit([]string{
		"clone_info",
	})

	err := s.repo.Update(course)
	if err != nil {
		return nil, err
	}

	s.StoreSchool(int64(course.ID), req)
	s.StoreStudyShifts(int64(course.ID), req)
	s.StoreSemester(int64(course.ID), req)

	id := int(course.ID)

	s.repo.SetPreload([]string{
		"Schools",
		"Subject",
		"CourseRefStudyShifts",
		"CourseRefStudyShifts.StudyShift",
		"Users",
		"Users.Roles",
		"Semesters",
		"Program",
		"Program.Chapters",
		"Program.Chapters.Lessons",
	})

	updateCourse, _ := s.repo.FindNewByID(id)

	return updateCourse, nil
}

func (s *courseService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	course, err := s.repo.FindByID(id)

	if err != nil {
		config.Log.Error(err)
		return err
	}

	if course == nil {
		return errors.New("course not found")
	}

	err = s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *courseService) StoreSchool(id int64, request *prot.CourseRequest) error {
	if request.School == nil || request.School.Id == 0 {
		if err := s.repo.DeleteOldSchool(id, 0); err != nil {
			return err
		}

		return nil
	}

	school := models.CourseSchool{
		CourseId: id,
		SchoolId: request.School.Id,
	}

	if err := s.repo.DeleteOldSchool(id, school.SchoolId); err != nil {
		return err
	}

	return s.repo.StoreSchool(school)
}

func (s *courseService) Restore(c *gin.Context, id int) (*models.Course, error) {
	s.repo.SetContext(c)
	course, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (s *courseService) GetUsers(c *gin.Context, id int64) ([]models.User, error) {
	role := c.Query("role")
	roleIdStr := c.Query("role_id")

	roleRepo := repositories.NewRoleRepository()

	var roleModel *models.Role
	var err error

	if role != "" {
		roleModel, err = roleRepo.FindByName(role)
	} else if roleIdStr != "" {
		roleId, err := strconv.Atoi(roleIdStr)
		if err != nil {
			return nil, err
		}
		roleModel, err = roleRepo.FindByID(roleId)
	}

	if err != nil || roleModel == nil {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	students, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return students, nil
}

func (s *courseService) StoreUsers(c *gin.Context, id int64) ([]models.User, error) {
	var req prot.UserCourseRequest
	var mainTeacherId int64
	var defaultMainTeacherId int64

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))

	if err != nil {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	userIds := make([]int64, 0, len(req.Users))
	for index, u := range req.Users {
		if req.RoleId == models.TeacherRoleId && index == 0 {
			defaultMainTeacherId = u.Id
		}

		if mainTeacherId == 0 && req.RoleId == models.TeacherRoleId && u.MainTeacher {
			mainTeacherId = u.Id
		}
		userIds = append(userIds, u.Id)
	}

	if req.RoleId == models.TeacherRoleId && mainTeacherId == 0 {
		mainTeacherId = defaultMainTeacherId
	}

	if err := s.repo.ReplaceUserCourse(id, userIds, roleModel.ID, mainTeacherId); err != nil {
		return nil, err
	}

	newStudents, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return newStudents, nil
}

func (s *courseService) AddUsers(c *gin.Context, id int64) ([]models.User, error) {
	var req prot.UserCourseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	userIds := make([]int64, 0, len(req.Users))
	for _, u := range req.Users {
		userIds = append(userIds, u.Id)
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))
	if err != nil {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	if err := s.repo.AddUserCourse(id, userIds, roleModel.ID); err != nil {
		return nil, err
	}

	newStudents, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, err
	}

	return newStudents, nil
}

func (s *courseService) CopySchedule(id, targetCourseId, programId int64) dto.CopyScheduleResponse {
	var copyScheduleResponse dto.CopyScheduleResponse
	if targetCourseId > 0 {
		var sourceProgramID int64
		err := db.ReplicaDB.Table("courses").
			Select("program_id").
			Where("id = ?", targetCourseId).
			Scan(&sourceProgramID).Error
		if err != nil {
			config.Log.Error(fmt.Sprintf("Lỗi khi kiểm tra program_id của course nguồn: %v", err))
			copyScheduleResponse.CopyScheduleSuccess = false
			copyScheduleResponse.CopyScheduleMessage = "Lỗi khi kiểm tra course nguồn"
		} else if (sourceProgramID != programId && programId > 0) {
			copyScheduleResponse.CopyScheduleSuccess = false
			copyScheduleResponse.CopyScheduleMessage = "Hai course không cùng program, không thể copy lịch học"
		} else {
			copyResult, err := s.lessonScheduleCopyService.CopyLessonSchedulesBetweenCourses(targetCourseId, id)
			if err != nil {
				config.Log.Error(fmt.Sprintf("Lỗi khi copy lesson schedules: %v", err))
				copyScheduleResponse.CopyScheduleSuccess = false
				copyScheduleResponse.CopyScheduleMessage = fmt.Sprintf("Lỗi hệ thống: %v", err)
			} else if !copyResult.Success {
				copyScheduleResponse.CopyScheduleSuccess = false
				copyScheduleResponse.CopyScheduleMessage = copyResult.Message
			} else {
				config.Log.Info(fmt.Sprintf("Copy lesson schedules thành công: %s", copyResult.Message))
				copyScheduleResponse.CopyScheduleSuccess = true
				copyScheduleResponse.CopyScheduleMessage = copyResult.Message
			}
		}
	} else {
		copyScheduleResponse.CopyScheduleSuccess = false
		copyScheduleResponse.CopyScheduleMessage = "Không copy schedule"
	}

	return copyScheduleResponse
}

func (s *courseService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	if schoolIDStr := c.Query("school_id"); schoolIDStr != "" {
		filter["Schools.id"] = schoolIDStr + ":course_schools:courses:schools:course_id:school_id:id:id"
	}

	if byUserParam := c.Query("by_user"); byUserParam != "" {
		if byUserParam == "true" || byUserParam == "1" || byUserParam == "yes" {
			userId := utils.GetCurrentUserId(c)
			userIdStr := strconv.Itoa(userId)

			filter["UserCourses.user_id"] = userIdStr + ":courses:user_courses:id:course_id:user_id"
		}
	}

	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if parentID, err := strconv.ParseInt(parentIDStr, 10, 64); err == nil {
			ids := s.repo.GetCloneIds(parentID)

			if len(ids) > 0 {
				idStrs := make([]string, len(ids))
				for i, id := range ids {
					idStrs[i] = strconv.FormatInt(id, 10)
				}
				filter["id"] = "in:" + strings.Join(idStrs, ",")
			} else {
				filter["id"] = "in:-1"
			}
		}
	}

	return filter, nil
}

func (s *courseService) StoreStudyShifts(courseID int64, request *prot.CourseRequest) error {
	studyShifts := make([]models.CourseRefStudyShift, 0, len(request.StudyShifts))
	validKeys := make([]string, 0, len(request.StudyShifts))

	for _, shift := range request.StudyShifts {
		studyShifts = append(studyShifts, models.CourseRefStudyShift{
			CourseID:  courseID,
			ShiftID:   shift.Id,
			DayOfWeek: shift.DayOfWeek,
		})

		validKeys = append(validKeys, fmt.Sprintf("%d_%d", shift.Id, shift.DayOfWeek))
	}

	if len(studyShifts) > 0 {
		if err := s.repo.StoreStudyShifts(studyShifts); err != nil {
			return err
		}
	}

	if err := s.repo.DeleteNotInStudyShifts(courseID, validKeys); err != nil {
		return err
	}

	return nil
}

func (s *courseService) GetScore(c *gin.Context, id int64) (*prot.ScoreByCourse, error) {
	roleId := utils.GetCurrentRoleId(c)
	userId := utils.GetCurrentUserId(c)

	if roleId != models.StudentRoleId {
		return nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	exams, err := s.repo.GetExams(id)
	examUsers, err := s.repo.GetExamUsers(int64(userId), id)

	if err != nil {
		return nil, err
	}

	examResults := []*prot.ExamResult{}
	var maxScore float32
	var totalScore float32
	var doneCount int

	for _, exam := range exams {
		examUser := models.ExamUser{}
		status := "unsuccess"
		var dateStr string
		var timeStr string
		var score float32

		for _, e := range examUsers {
			if e.ExamID == exam.ID {
				examUser = e
				status = "success"
				break
			}
		}

		if examUser.Time != nil {
			minutes := int(*examUser.Time) / 60
			if minutes == 0 {
				minutes = 1
			}
			timeStr = fmt.Sprintf("%d", minutes)
		}

		if examUser.Ratio != nil {
			score = float32(*examUser.Ratio) / 10
		}

		if score > maxScore {
			maxScore = score
		}

		if status == "success" {
			totalScore += score
			doneCount++
			dateStr = examUser.CreatedAt.Format("02/01/2006")
		}

		score = float32(math.Round(float64(score)*10) / 10)

		examMinutes := int(*&exam.TimeLimit) / 60

		if examMinutes == 0 {
			examMinutes = 1
		}

		examTimeStr := fmt.Sprintf("%d", examMinutes)

		var lessonTitle string
		if len(exam.Lessons) > 0 {
			lessonTitle = exam.Lessons[0].Title
		}

		examResults = append(examResults, &prot.ExamResult{
			Id:       exam.ID,
			Lesson:   lessonTitle,
			Exam:     exam.Name,
			Date:     dateStr,
			Time:     timeStr,
			ExamTime: examTimeStr,
			Score:    score,
			MaxScore: 10,
			Status:   status,
		})
	}

	sort.Slice(examResults, func(i, j int) bool {
		if examResults[i].Status == "success" && examResults[j].Status != "success" {
			return true
		}
		if examResults[i].Status != "success" && examResults[j].Status == "success" {
			return false
		}
		return false
	})

	var averageScore float32
	if doneCount > 0 {
		averageScore = totalScore / float32(doneCount)
	}

	maxScore = float32(math.Round(float64(maxScore)*10) / 10)
	averageScore = float32(math.Round(float64(averageScore)*10) / 10)

	scoreByCourse := prot.ScoreByCourse{
		ExamCount:    int32(len(exams)),
		FinishExam:   int32(doneCount),
		ExamResults:  examResults,
		MaxScore:     maxScore,
		AverageScore: averageScore,
	}

	return &scoreByCourse, nil
}

func (s *courseService) StoreSemester(id int64, request *prot.CourseRequest) error {
	semesterIds := make([]int64, 0, len(request.Semesters))
	for _, s := range request.Semesters {
		semesterIds = append(semesterIds, s.Id)
	}

	config.Log.Info(semesterIds)

	return s.repo.StoreSemester(id, semesterIds)
}

func (cs *courseService) RespondDetail(c *gin.Context, course *models.Course) (*prot.Course, error) {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := NewLessonService(lessonRepo)
	completeLessonIds := lessonService.CompletionLessonIds(c, course.ID, 0)

	lessonSchedules, _ := lessonRepo.GetLessonSchedulesByCourse(course.ID)

	courseResource := resources.NewCourseResource()
	if impl, ok := courseResource.(*resources.CourseResourceImpl); ok {
		impl.CompleteLessonIds = completeLessonIds
		impl.LessonSchedules = lessonSchedules
	}
	formattedCourse := courseResource.FormatCourseDetail(course)

	return formattedCourse, nil
}
