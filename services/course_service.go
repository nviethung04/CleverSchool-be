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
	"mime/multipart"
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
	GetUsers(c *gin.Context, id int64) ([]models.User, []int64, error)
	StoreUsers(c *gin.Context, id int64) ([]models.User, []int64, error)
	AddUsers(c *gin.Context, id int64) ([]models.User, []int64, error)
	GetScore(c *gin.Context, id int64) (*prot.ScoreByCourse, error)
	CopySchedule(id, targetCourseId, programId int64) dto.CopyScheduleResponse
	Export(c *gin.Context) (string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
	ExportUsers(c *gin.Context, courseId int64) (string, error)
	ImportUsers(c *gin.Context, courseId int64, fileHeader *multipart.FileHeader) error
	GetCoursesFamilyByProgramId(programId int64) (*prot.ListCoursesFamilyResponse, error)
	NotEligibleForFinalExamIds(c *gin.Context, programId, userId int64) []int64
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
		"Program.Chapters.Headings",
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
		"Program.Chapters.Headings",
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

	// Nếu có parent_course_id, set parent và copy lịch học
	if req.ParentCourseId > 0 {
		// Validate parent course
		parentInfo, err := s.getParentCourseInfo(req.ParentCourseId)
		if err != nil {
			config.Log.Error(fmt.Sprintf("Lỗi khi kiểm tra khóa cha: %v", err))
		} else if parentInfo != nil {
			// Kiểm tra parent phải là root (parent_course_id = 0) và cùng program_id
			if parentInfo.ParentCourseID == 0 && parentInfo.ProgramID == course.ProgramId {
				// Set parent_course_id cho khóa mới
				if err := db.MasterDB.Model(&models.Course{}).
					Where("id = ?", course.ID).
					Update("parent_course_id", req.ParentCourseId).Error; err != nil {
					config.Log.Error(fmt.Sprintf("Lỗi khi set parent_course_id: %v", err))
				} else {
					// Copy lịch học từ parent sang khóa mới
					copyResult, err := s.lessonScheduleCopyService.CopyLessonSchedulesBetweenCourses(req.ParentCourseId, int64(course.ID))
					if err != nil {
						config.Log.Error(fmt.Sprintf("Lỗi khi copy lesson schedules: %v", err))
					} else if !copyResult.Success {
						config.Log.Error(fmt.Sprintf("Copy lesson schedules thất bại: %s", copyResult.Message))
					} else {
						config.Log.Info(fmt.Sprintf("Copy lesson schedules thành công: %s", copyResult.Message))
					}
				}
			}
		}
	}

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
		"Program.Chapters.Headings",
	})

	newCourse, _ := s.repo.FindNewByID(id)

	return newCourse, nil
}

func (s *courseService) Update(c *gin.Context, req *prot.CourseRequest) (*models.Course, error) {
	courseResource := resources.NewCourseResource()
	course := courseResource.FormatModelCourse(req)

	s.repo.SetContext(c)

	// Nếu không truyền parent_course_id (hoặc = 0), thì không cập nhật cột này
	omitFields := []string{
		"clone_info",
	}
	if req.ParentCourseId == 0 {
		omitFields = append(omitFields, "parent_course_id")
	} else {
		// Nếu có parent_course_id, set vào model
		course.ParentCourseId = req.ParentCourseId
	}

	s.repo.SetOmit(omitFields)

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
		"Program.Chapters.Headings",
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

func (s *courseService) GetUsers(c *gin.Context, id int64) ([]models.User, []int64, error) {
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
			return nil, nil, err
		}
		roleModel, err = roleRepo.FindByID(roleId)
	}

	if err != nil || roleModel == nil {
		return nil, nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	students, failedUserIds, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, nil, err
	}

	return students, failedUserIds, nil
}

func (s *courseService) StoreUsers(c *gin.Context, id int64) ([]models.User, []int64, error) {
	var mainTeacherId int64
	var defaultMainTeacherId int64

	req, err, _ := utils.GetBody[*prot.UserCourseRequest](c, func() *prot.UserCourseRequest {
		return &prot.UserCourseRequest{}
	})

	if err != nil {
		return nil, nil, err
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))

	if err != nil {
		return nil, nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	userIds := make([]int64, 0, len(req.Users))
	failedUserIds := make([]int64, 0, len(req.Users))
	for index, u := range req.Users {
		if req.RoleId == models.TeacherRoleId && index == 0 {
			defaultMainTeacherId = u.Id
		}

		if mainTeacherId == 0 && req.RoleId == models.TeacherRoleId && u.MainTeacher {
			mainTeacherId = u.Id
		}
		userIds = append(userIds, u.Id)

		if u.IsFailed {
			failedUserIds = append(failedUserIds, u.Id)
		}
	}

	if req.RoleId == models.TeacherRoleId && mainTeacherId == 0 {
		mainTeacherId = defaultMainTeacherId
	}

	if err := s.repo.ReplaceUserCourse(id, userIds, roleModel.ID, mainTeacherId, failedUserIds); err != nil {
		return nil, nil, err
	}

	newStudents, failedUserIds, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, nil, err
	}

	return newStudents, failedUserIds, nil
}

func (s *courseService) AddUsers(c *gin.Context, id int64) ([]models.User, []int64, error) {
	req, err, _ := utils.GetBody[*prot.UserCourseRequest](c, func() *prot.UserCourseRequest {
		return &prot.UserCourseRequest{}
	})

	if err != nil {
		return nil, nil, err
	}

	userIds := make([]int64, 0, len(req.Users))
	failedUserIds := make([]int64, 0, len(req.Users))
	for _, u := range req.Users {
		userIds = append(userIds, u.Id)

		if u.IsFailed {
			failedUserIds = append(failedUserIds, u.Id)
		}
	}

	roleRepo := repositories.NewRoleRepository()
	roleModel, err := roleRepo.FindByID(int(req.RoleId))
	if err != nil {
		return nil, nil, errors.New(i18n.Localize("messages.role_invalid"))
	}

	if err := s.repo.AddUserCourse(id, userIds, roleModel.ID, failedUserIds); err != nil {
		return nil, nil, err
	}

	newStudents, failedUserIds, err := s.repo.GetUsers(id, roleModel.ID)
	if err != nil {
		return nil, nil, err
	}

	return newStudents, failedUserIds, nil
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
		} else if sourceProgramID != programId && programId > 0 {
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

			// courseIds := s.repo.GetIdsByUserId(int64(userId))

			// if len(courseIds) == 0 && utils.GetCurrentMemberType(c) == models.MemberTypeExternal {
			// 	courseIds = config.LoadConfig().PublicCourseIds
			// }

			// idStrs := make([]string, len(courseIds))
			// for i, id := range courseIds {
			// 	idStrs[i] = strconv.FormatInt(id, 10)
			// }

			// filter["courses.id"] = "in:" + strings.Join(idStrs, ",")

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
				filter["courses.id"] = "in:" + strings.Join(idStrs, ",")
			} else {
				filter["courses.id"] = "in:-1"
			}
		}
	}

	// Loại bỏ subject_id khỏi filter vì đã xử lý trong BeforeQuery hook
	// subject_id được lấy từ programs.subject_id, không phải courses.subject_id
	delete(filter, "subject_id")

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
		FinishExam:   int32(len(examUsers)),
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

	return s.repo.StoreSemester(id, semesterIds)
}

func (cs *courseService) RespondDetail(c *gin.Context, course *models.Course) (*prot.Course, error) {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := NewLessonService(lessonRepo)
	hideLessonIds := lessonService.HideLessonIds(c)
	studyingLessonIds := lessonService.StudyingLessonIds(c, course.ID, 0)
	completeLessonIds := lessonService.CompletionLessonIds(c, course.ID, 0)

	lessonSchedules, _ := lessonRepo.GetLessonSchedulesByCourse(course.ID)
	userId := utils.GetCurrentUserId(c)

	courseResource := resources.NewCourseResource()
	if impl, ok := courseResource.(*resources.CourseResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
		impl.LessonSchedules = lessonSchedules
		impl.NotEligibleForFinalExamIds = cs.NotEligibleForFinalExamIds(c, course.ID, int64(userId))
	}
	formattedCourse := courseResource.FormatCourseDetail(course)

	return formattedCourse, nil
}

type parentCourseInfo struct {
	ID             int64
	ProgramID      int64
	ParentCourseID int64
}

func (s *courseService) getParentCourseInfo(courseID int64) (*parentCourseInfo, error) {
	var info parentCourseInfo
	err := db.ReplicaDB.Table("courses").
		Select("id, program_id, parent_course_id").
		Where("id = ? AND deleted_at IS NULL", courseID).
		Scan(&info).Error

	if err != nil {
		return nil, err
	}

	if info.ID == 0 {
		return nil, fmt.Errorf("không tìm thấy khóa học %d", courseID)
	}

	return &info, nil
}

func (s *courseService) GetCoursesFamilyByProgramId(programId int64) (*prot.ListCoursesFamilyResponse, error) {
	parentCourses, allCourses, err := s.repo.GetCoursesFamilyByProgramId(programId)
	if err != nil {
		return nil, err
	}

	// Tạo map để tìm children nhanh hơn
	childrenMap := make(map[int64][]*prot.CourseFamilyItem)
	for _, course := range allCourses {
		if course.ParentCourseId > 0 {
			childrenMap[course.ParentCourseId] = append(childrenMap[course.ParentCourseId], &prot.CourseFamilyItem{
				Id:          course.ID,
				Name:        course.Name,
				ObjectTitle: course.ObjectTitle,
			})
		}
	}

	// Tạo response với parent courses và children của chúng
	var parentCoursesResponse []*prot.CourseFamilyParent
	for _, parent := range parentCourses {
		children := childrenMap[parent.ID]
		if children == nil {
			children = []*prot.CourseFamilyItem{}
		}

		parentCoursesResponse = append(parentCoursesResponse, &prot.CourseFamilyParent{
			Id:              parent.ID,
			Name:            parent.Name,
			ObjectTitle:     parent.ObjectTitle,
			ChildrenCourses: children,
		})
	}

	return &prot.ListCoursesFamilyResponse{
		ParentCourses: parentCoursesResponse,
	}, nil
}

func (s *courseService) NotEligibleForFinalExamIds(c *gin.Context, courseId, userId int64) []int64 {
	isVtg := config.LoadConfig().IsVtg
	roleId := utils.GetCurrentRoleId(c)

	if !isVtg || roleId != models.StudentRoleId {
		return []int64{}
	}

	notEligibleForFinalExamIds, err := s.repo.NotEligibleForFinalExamIds(courseId, userId)
	if err != nil {
		config.Log.Error(fmt.Sprintf("error getting not eligible for final exam ids: %v", err))
		return []int64{}
	}

	config.Log.Info(fmt.Sprintf("notEligibleForFinalExamIds: %v", notEligibleForFinalExamIds))

	return notEligibleForFinalExamIds
}
