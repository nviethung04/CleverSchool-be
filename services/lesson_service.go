package services

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"be-Clever School/database/db"

	"github.com/gin-gonic/gin"
)

type LessonService interface {
	GetAll(c *gin.Context) ([]models.Lesson, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Lesson, error)
	Create(c *gin.Context, req *prot.LessonRequest) (*models.Lesson, error)
	Update(c *gin.Context, req *prot.LessonRequest) (*models.Lesson, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Lesson, error)
	StoreLessonSchedules(c *gin.Context, list *prot.LessonSchedulesInfoRequest) (int64, error)
	LessonSchedules(c *gin.Context, filters map[string]interface{}) (dto.LessonScheduleDetail, *time.Time, []models.Week, bool, error)
	Completion(c *gin.Context, lessonCompletion *prot.LessonCompletion) (*prot.LessonCompletion, error)
	CompletionLessonIds(c *gin.Context, courseId int64, chapterId int64) []int64
	StudyingLessonIds(c *gin.Context, courseId int64, chapterId int64) []int64
	Studying(c *gin.Context, lessonCompletion *prot.LessonStudying) (*prot.LessonStudying, error)
	HideLessonIds(c *gin.Context) []int64

	GetLessonPlanByCourse(c *gin.Context, lessonId, courseId int64) (*prot.LessonPlanByCourse, error)
	StoreLessonPlanByCourse(c *gin.Context, lessonId, courseId int64, req *prot.LessonPlanByCourse) (*prot.LessonPlanByCourse, error)
	GetExamByCourse(c *gin.Context, lessonId, courseId int64) (*prot.ExamByCourse, error)
	StoreExamByCourse(c *gin.Context, lessonId, courseId int64, req *prot.ExamByCourse) (*prot.ExamByCourse, error)
	GetHomeworkByCourse(c *gin.Context, lessonId, courseId int64) (*prot.HomeworkByCourse, error)
	StoreHomeworkByCourse(c *gin.Context, lessonId, courseId int64, req *prot.HomeworkByCourse) (*prot.HomeworkByCourse, error)
	GetExerciseByCourse(c *gin.Context, lessonId, courseId int64) (*prot.ExerciseByCourse, error)
	StoreExerciseByCourse(c *gin.Context, lessonId, courseId int64, req *prot.ExerciseByCourse) (*prot.ExerciseByCourse, error)
}

type lessonService struct {
	repo              repositories.LessonRepository
	assessmentService AssessmentCourseService
}

func NewLessonService(repo repositories.LessonRepository) LessonService {
	return &lessonService{
		repo:              repo,
		assessmentService: NewAssessmentCourseService(),
	}
}

func (s *lessonService) GetAll(c *gin.Context) ([]models.Lesson, int64, error) {
	allowedFilters := []string{"chapter_id", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"title", "description", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Dependencies",
		"Skills",
		"Tags",
		"Topics",
		"Heading",
		"Chapter",
		"Chapter.Program",
		"Author",
		"Exams",
		"Homeworks:deleted_at IS NULL",
		"Exercises",
		"LessonPlans",
		"Exams.ExamRefLessons",
		"Homeworks.HomeworkRefLessons",
		"Exercises.ExerciseRefLessons",
		"Assessments",
		"Assessments.AssessmentRefLessons",
		"LessonPlans.Author",
		"TeachingPlans",
	})

	s.repo.SetAlias(map[string]string{
		"*":                  "lessons.*",
		"lessons.created_by": "author_id",
	})

	lessons, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return lessons, rows, nil
}

func (s *lessonService) GetByID(c *gin.Context, id int) (*prot.Lesson, error) {
	s.repo.SetPreload([]string{
		"Dependencies",
		"Skills",
		"Tags",
		"Topics",
		"Heading",
		"Chapter",
		"Chapter.Program",
		"Author",
		"Exams",
		"Homeworks:deleted_at IS NULL",
		"Exercises",
		"LessonPlans",
		"Schedules",
		"Schedules.Week",
		"Schedules.Course",
		"Exams.ExamRefLessons",
		"Homeworks.HomeworkRefLessons",
		"Exercises.ExerciseRefLessons",
		"Assessments",
		"Assessments.AssessmentRefLessons",
		"LessonPlans.Author",
		"TeachingPlans",
	})

	s.repo.SetAlias(map[string]string{
		"*":                  "*",
		"lessons.created_by": "author_id",
	})

	s.repo.SetContext(c)
	lesson, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	userID := utils.GetCurrentUserId(c)

	// Batch query trạng thái hoàn thành exam
	examIDs := make([]int64, 0)
	for _, exam := range lesson.Exams {
		examIDs = append(examIDs, exam.ID)
	}
	examCompletionMap := map[int64]bool{}
	if len(examIDs) > 0 {
		var rows []struct{ ExamID int64 }
		db.ReplicaDB.Table("exam_users").
			Select("exam_id").
			Where("exam_id IN ? AND user_id = ?", examIDs, userID).
			Scan(&rows)
		for _, row := range rows {
			examCompletionMap[row.ExamID] = true
		}
	}

	// Batch query số câu đã làm homework
	homeworkIDs := make([]int64, 0)
	for _, hw := range lesson.Homeworks {
		homeworkIDs = append(homeworkIDs, hw.ID)
	}
	homeworkCompletedMap := map[int64]int32{}
	homeworkSubmittedMap := map[int64]bool{}
	if len(homeworkIDs) > 0 {
		var rows []struct {
			HomeworkID         int64
			QuestionsCompleted int64
		}
		db.ReplicaDB.Table("homework_users").
			Select("homework_id, questions_completed").
			Where("homework_id IN ? AND user_id = ?", homeworkIDs, userID).
			Scan(&rows)
		for _, row := range rows {
			homeworkCompletedMap[row.HomeworkID] = int32(row.QuestionsCompleted)
			homeworkSubmittedMap[row.HomeworkID] = true
		}
	}

	// Lấy tổng số câu hỏi cho từng exam/homework/exercise
	examTotalQuestions := map[int64]int32{}
	for _, exam := range lesson.Exams {
		total, _ := CountExamQuestionsFromCloned(exam.ID)
		examTotalQuestions[exam.ID] = total
	}
	homeworkTotalQuestions := map[int64]int32{}
	for _, hw := range lesson.Homeworks {
		total, _ := CountHomeworkQuestionsFromCloned(hw.ID)
		homeworkTotalQuestions[hw.ID] = total
	}

	// Batch query trạng thái hoàn thành exercise
	exerciseIDs := make([]int64, 0)
	for _, exercise := range lesson.Exercises {
		exerciseIDs = append(exerciseIDs, exercise.ID)
	}
	exerciseCompletionMap := map[int64]bool{}
	if len(exerciseIDs) > 0 {
		var rows []struct{ ExerciseID int64 }
		db.ReplicaDB.Table("exercise_users").
			Select("exercise_id").
			Where("exercise_id IN ? AND user_id = ?", exerciseIDs, userID).
			Scan(&rows)
		for _, row := range rows {
			exerciseCompletionMap[row.ExerciseID] = true
		}
	}

	exerciseTotalQuestions := map[int64]int32{}
	for _, exercise := range lesson.Exercises {
		total, _ := CountExerciseQuestionsFromCloned(exercise.ID)
		exerciseTotalQuestions[exercise.ID] = total
	}

	lessonResource := resources.NewLessonResourceWithCompletionAndQuestions(
		examCompletionMap, examTotalQuestions, homeworkCompletedMap, homeworkTotalQuestions,
		exerciseCompletionMap, exerciseTotalQuestions, homeworkSubmittedMap,
	)

	hideLessonIds := s.HideLessonIds(c)
	studyingLessonIds := s.StudyingLessonIds(c, 0, 0)
	completeLessonIds := s.CompletionLessonIds(c, 0, 0)
	completeLessonPlanIds := s.CompletionLessonPlanIds(c, 0, 0, lesson.ID)
	courseId, _ := strconv.Atoi(c.Query("course_id"))
	var publishAssessmentIds []int64

	if courseId > 0 {
		assessmentRepo := repositories.NewAssessmentRepository()
		publishAssessmentIds = assessmentRepo.GetPublishAssessmentIds(int64(courseId))
	}

	if impl, ok := lessonResource.(*resources.LessonResourceImpl); ok {
		impl.HideLessonIds = hideLessonIds
		impl.StudyingLessonIds = studyingLessonIds
		impl.CompleteLessonIds = completeLessonIds
		impl.CompleteLessonPlanIds = completeLessonPlanIds
		impl.CourseId = int64(courseId)
		impl.PublishAssessmentIds = publishAssessmentIds
	}

	formattedLesson := lessonResource.FormatLesson(lesson)

	return formattedLesson, nil
}

func (s *lessonService) Create(c *gin.Context, req *prot.LessonRequest) (*models.Lesson, error) {
	lessonResource := resources.NewLessonResource()
	lesson := lessonResource.FormatModelLesson(req)

	s.repo.SetContext(c)

	var sortPosition int
	if lesson.ChapterID > 0 {
		sortPosition, _ = s.repo.GetPositionByIdAndChapter(req.Chapter.Id, 0)
	}
	lesson.SortPosition = sortPosition

	if lesson.ChapterID == 0 && req.ProgramId > 0 {
		chapterId, err := s.repo.CreateChapter(req.ProgramId)

		if err != nil {
			return nil, err
		}

		lesson.ChapterID = chapterId
	}

	if lesson.HeadingID > 0 {
		headingRepo := repositories.NewHeadingRepository()
		heading, _ := headingRepo.FindByID(int(lesson.HeadingID))

		if heading.ID > 0 && (heading.ChapterId == lesson.ChapterID || lesson.ChapterID == 0) {
			lesson.ChapterID = heading.ChapterId
		} else {
			lesson.HeadingID = 0
			config.Log.Info("Heading not found id: ", lesson.HeadingID)
		}
	}

	err := s.repo.Create(lesson)
	if err != nil {
		return nil, err
	}

	id := int(lesson.ID)

	s.StoreDependencies(c, int64(id), req)
	s.StoreTags(c, int64(id), req)
	s.StoreTopics(c, int64(id), req)
	s.StoreSkills(c, int64(id), req)
	s.StoreExams(c, int64(id), 0, req)
	s.StoreHomeworks(c, int64(id), 0, req)
	s.StoreExercises(c, int64(id), 0, req)
	s.StoreAssessments(c, int64(id), 0, req)
	s.StoreLessonPlans(c, int64(id), 0, req)
	s.StoreTeachingPlans(c, int64(id), req)
	// s.StoreSchedules(c, int64(id), req)

	s.repo.SetPreload([]string{
		"Dependencies",
		"Skills",
		"Tags",
		"Topics",
		"Heading",
		"Chapter",
		"Chapter.Program",
		"Author",
		"Exams",
		"Homeworks:deleted_at IS NULL",
		"Exercises",
		"LessonPlans",
		"Schedules",
		"Schedules.Week",
		"Schedules.Course",
		"Exams.ExamRefLessons",
		"Homeworks.HomeworkRefLessons",
		"Exercises.ExerciseRefLessons",
		"Assessments",
		"Assessments.AssessmentRefLessons",
		"LessonPlans.Author",
		"TeachingPlans",
	})

	s.repo.SetAlias(map[string]string{
		"*":                  "*",
		"lessons.created_by": "author_id",
	})

	newLesson, _ := s.repo.FindNewByID(id)

	return newLesson, nil
}

func (s *lessonService) Update(c *gin.Context, req *prot.LessonRequest) (*models.Lesson, error) {
	courseId, _ := strconv.Atoi(c.Query("course_id"))

	oldLesson, _ := s.repo.FindNewByID(int(req.Id))

	if req.Chapter == nil {
		req.Chapter = &prot.ChapterInfo{}
	}

	if oldLesson.ChapterID != req.Chapter.Id {
		scheduleRepo := repositories.NewLessonScheduleRepository()
		scheduleRepo.UpdateCourseIdByLesson(req.Id, req.Chapter.Id, oldLesson.ChapterID)
	}

	lessonResource := resources.NewLessonResource()
	lesson := lessonResource.FormatModelLesson(req)

	s.repo.SetContext(c)
	s.repo.SetOmit([]string{
		"clone_info",
	})

	id := int(lesson.ID)

	if courseId == 0 {
		var sortPosition int
		if lesson.ChapterID > 0 {
			sortPosition, _ = s.repo.GetPositionByIdAndChapter(req.Chapter.Id, req.Id)
		}
		lesson.SortPosition = sortPosition

		if lesson.ChapterID == 0 && req.ProgramId > 0 {
			chapterId, err := s.repo.CreateChapter(req.ProgramId)

			if err != nil {
				return nil, err
			}

			lesson.ChapterID = chapterId
		}

		if lesson.HeadingID > 0 {
			headingRepo := repositories.NewHeadingRepository()
			heading, _ := headingRepo.FindByID(int(lesson.HeadingID))

			config.Log.Info("Heading: ", heading)

			if heading.ID > 0 && (heading.ChapterId == lesson.ChapterID || lesson.ChapterID == 0) {
				lesson.ChapterID = heading.ChapterId
			} else {
				lesson.HeadingID = 0
				config.Log.Info("Heading not found id: ", lesson.HeadingID)
			}
		}

		err := s.repo.Update(lesson)
		if err != nil {
			return nil, err
		}

		s.StoreDependencies(c, int64(id), req)
		s.StoreTags(c, int64(id), req)
		s.StoreTopics(c, int64(id), req)
		s.StoreSkills(c, int64(id), req)
		s.StoreLessonPlans(c, int64(id), int64(courseId), req)
		s.StoreTeachingPlans(c, int64(id), req)
	}

	s.StoreExams(c, int64(id), int64(courseId), req)
	s.StoreHomeworks(c, int64(id), int64(courseId), req)
	s.StoreExercises(c, int64(id), int64(courseId), req)
	s.StoreAssessments(c, int64(id), int64(courseId), req)
	// s.StoreSchedules(c, int64(id), req)

	s.repo.SetPreload([]string{
		"Dependencies",
		"Skills",
		"Tags",
		"Topics",
		"Heading",
		"Chapter",
		"Chapter.Program",
		"Author",
		"Exams",
		"Homeworks:deleted_at IS NULL",
		"Exercises",
		"LessonPlans",
		"Schedules",
		"Schedules.Week",
		"Schedules.Course",
		"Exams.ExamRefLessons",
		"Homeworks.HomeworkRefLessons",
		"Exercises.ExerciseRefLessons",
		"Assessments",
		"Assessments.AssessmentRefLessons",
		"LessonPlans.Author",
		"TeachingPlans",
	})

	s.repo.SetAlias(map[string]string{
		"*":                  "*",
		"lessons.created_by": "author_id",
	})

	updateLesson, _ := s.repo.FindNewByID(id)

	return updateLesson, nil
}

func (s *lessonService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *lessonService) Restore(c *gin.Context, id int) (*models.Lesson, error) {
	s.repo.SetContext(c)
	lesson, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *lessonService) StoreDependencies(c *gin.Context, id int64, req *prot.LessonRequest) error {
	dependencyIds := make([]int64, 0, len(req.Dependencies))

	for _, value := range req.Dependencies {
		dependency := models.LessonDependency{
			LessonID:     id,
			DependencyID: value.Id,
		}
		s.repo.UpdateOrCreateDependency(dependency)

		dependencyIds = append(dependencyIds, value.Id)
	}

	s.repo.DeleteOldDependency(id, dependencyIds)

	return nil
}

func (s *lessonService) StoreTags(c *gin.Context, id int64, req *prot.LessonRequest) error {
	tagIds := make([]int64, 0, len(req.Tags))

	for _, value := range req.Tags {
		tag := models.LessonRefTag{
			LessonID: id,
			TagID:    value.Id,
		}

		s.repo.UpdateOrCreateTag(tag)

		tagIds = append(tagIds, value.Id)
	}

	s.repo.DeleteOldTag(id, tagIds)

	return nil
}

func (s *lessonService) StoreTopics(c *gin.Context, id int64, req *prot.LessonRequest) error {
	topicIds := make([]int64, 0, len(req.Topics))

	for _, value := range req.Topics {
		topic := models.LessonRefTopic{
			LessonID: id,
			TopicID:  value.Id,
		}

		s.repo.UpdateOrCreateTopic(topic)

		topicIds = append(topicIds, value.Id)
	}

	s.repo.DeleteOldTopic(id, topicIds)

	return nil
}

func (s *lessonService) StoreSkills(c *gin.Context, id int64, req *prot.LessonRequest) error {
	skillIds := make([]int64, 0, len(req.Skills))

	for _, value := range req.Skills {
		skill := models.LessonRefSkill{
			LessonID: id,
			SkillID:  value.Id,
		}

		s.repo.UpdateOrCreateSkill(skill)

		skillIds = append(skillIds, value.Id)
	}

	s.repo.DeleteOldSkill(id, skillIds)

	return nil
}

func (s *lessonService) StoreExams(c *gin.Context, id, courseId int64, req *prot.LessonRequest) error {
	examIdMap := make(map[int64]struct{})
	examIds := []int64{}

	for _, exam := range req.Exams {
		if exam.Id != 0 {
			if _, exists := examIdMap[exam.Id]; !exists {
				examIdMap[exam.Id] = struct{}{}
				examIds = append(examIds, exam.Id)
			}
		}
	}

	err := s.repo.UpdateExamRefLesson(id, courseId, examIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *lessonService) StoreExercises(c *gin.Context, id, courseId int64, req *prot.LessonRequest) error {
	exerciseIdMap := make(map[int64]struct{})
	exerciseIds := []int64{}

	for _, exercise := range req.Exercises {
		if exercise.Id != 0 {
			if _, exists := exerciseIdMap[exercise.Id]; !exists {
				exerciseIdMap[exercise.Id] = struct{}{}
				exerciseIds = append(exerciseIds, exercise.Id)
			}
		}
	}

	err := s.repo.UpdateExerciseRefLesson(id, courseId, exerciseIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *lessonService) StoreHomeworks(c *gin.Context, id, courseId int64, req *prot.LessonRequest) error {
	homeworkIdMap := make(map[int64]struct{})
	homeworkIds := []int64{}

	for _, homework := range req.Homeworks {
		if homework.Id != 0 {
			if _, exists := homeworkIdMap[homework.Id]; !exists {
				homeworkIdMap[homework.Id] = struct{}{}
				homeworkIds = append(homeworkIds, homework.Id)
			}
		}
	}

	err := s.repo.UpdateHomeworkRefLesson(id, courseId, homeworkIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *lessonService) StoreAssessments(c *gin.Context, id, courseId int64, req *prot.LessonRequest) error {
	assessmentIdMap := make(map[int64]struct{})
	assessmentIds := []int64{}

	for _, assessment := range req.Assessment {
		if assessment.Id != 0 {
			if _, exists := assessmentIdMap[assessment.Id]; !exists {
				assessmentIdMap[assessment.Id] = struct{}{}
				assessmentIds = append(assessmentIds, assessment.Id)
			}
		}
	}

	// Nếu courseId = 0 (không truyền course_id hoặc create/update lesson ở program level)
	if courseId == 0 {
		// Thêm/update bản ghi assessment_ref_lessons với course_id = 0
		err := s.repo.UpdateAssessmentRefLesson(id, 0, assessmentIds)
		if err != nil {
			return err
		}

		// Gọi service AssignCoursesToAssessment để gán cho tất cả courses thuộc program
		for _, assessmentID := range assessmentIds {
			if err := s.assessmentService.AssignCoursesToAssessment(c, assessmentID, id); err != nil {
				// Log lỗi nhưng không dừng toàn bộ quá trình
				config.Log.Error(fmt.Sprintf("Error assigning courses to assessment %d for lesson %d: %v", assessmentID, id, err))
			}
		}
		return nil
	}

	// Nếu courseId > 0 (có truyền course_id - chỉ gán cho course cụ thể)
	// Chỉ thêm bản ghi với course_id đó, không thêm với course_id = 0
	// Không gọi AssignCoursesToAssessment
	err := s.repo.UpdateAssessmentRefLesson(id, courseId, assessmentIds)
	if err != nil {
		return err
	}

	return nil
}

func (s *lessonService) StoreLessonPlans(c *gin.Context, id, courseId int64, req *prot.LessonRequest) error {
	isVtg := config.LoadConfig().IsVtg

	if isVtg {
		var lessonPlanIds []int64

		for _, lessonPlan := range req.LessonPlans {
			lessonPlanIds = append(lessonPlanIds, lessonPlan.Id)
		}

		config.Log.Info(lessonPlanIds)

		err := s.repo.UpdateLessonPlans(id, lessonPlanIds)
		if err != nil {
			return err
		}
		return nil
	} else {
		var lessonPlanId int64
		for _, lessonPlan := range req.LessonPlans {
			lessonPlanId = lessonPlan.Id
			break
		}

		err := s.repo.UpdateLessonPlan(id, lessonPlanId)
		if err != nil {
			return err
		}

		return nil
	}
}

func (s *lessonService) StoreSchedules(c *gin.Context, lessonID int64, req *prot.LessonRequest) error {
	if req == nil {
		return nil
	}

	scheduleIDs := make([]int64, 0, len(req.Schedules))
	newSchedules := make([]models.LessonSchedule, 0, len(req.Schedules))

	for _, sch := range req.Schedules {
		var parsedDate time.Time
		var courseId int64
		var err error

		if sch.ScheduledDate != "" {
			parsedDate, err = time.Parse("2006-01-02", sch.ScheduledDate)
		} else {
			continue
		}

		if parsedDate.IsZero() {
			continue
		}

		week, err := s.repo.GetWeekByDate(parsedDate)
		if err != nil {
			continue
		}
		if sch.CourseId != 0 {
			courseId = sch.CourseId
		}

		if parsedDate.IsZero() || courseId == 0 {
			continue
		}

		schedule := models.LessonSchedule{
			ID:            sch.Id,
			LessonID:      lessonID,
			CourseID:      courseId,
			ScheduledDate: parsedDate,
			WeekID:        week.ID,
		}

		if schedule.ID != 0 {
			if err := s.repo.UpdateLessonSchedule(&schedule); err != nil {
				return err
			}
		} else {
			newSchedules = append(newSchedules, schedule)
		}

		scheduleIDs = append(scheduleIDs, schedule.ID)
	}

	if err := s.repo.DeleteLessonSchedulesNotIn(lessonID, scheduleIDs); err != nil {
		return err
	}

	if len(newSchedules) > 0 {
		if err := s.repo.CreateLessonSchedules(newSchedules); err != nil {
			return err
		}
	}

	return nil
}

func (s *lessonService) StoreLessonSchedules(c *gin.Context, req *prot.LessonSchedulesInfoRequest) (int64, error) {
	if req == nil || len(req.Weeks) == 0 {
		return 0, nil
	}

	courseID := req.CourseId

	if courseID == 0 {
		if firstLessonID := req.Weeks[0].Schedules[0].Lesson.Id; firstLessonID > 0 {
			cid, err := s.repo.GetCourseIdByLessonId(firstLessonID)
			if err != nil {
				return 0, err
			}
			courseID = cid
		}
		if courseID == 0 {
			return 0, nil
		}
	}

	// ✅ Tính danh sách lessonIds (duy nhất)
	lessonIdSet := make(map[int64]bool)
	for _, week := range req.Weeks {
		for _, sched := range week.Schedules {
			if sched.Lesson != nil && sched.Lesson.Id > 0 {
				lessonIdSet[sched.Lesson.Id] = true
			}
		}
	}
	var lessonIds []int64
	for id := range lessonIdSet {
		lessonIds = append(lessonIds, id)
	}

	// Xoá toàn bộ lịch cũ theo course
	if err := s.repo.DeleteLessonSchedulesByCourse(courseID, lessonIds); err != nil {
		return courseID, err
	}

	// Check trùng lesson trong cùng 1 tuần
	weekLessonSet := make(map[int64]map[int64]bool) // map[weekID][lessonID] = true

	// Set preload 1 lần
	s.repo.SetPreload([]string{"Chapter", "LessonPlans"})

	var newSchedules []models.LessonSchedule

	for _, week := range req.Weeks {
		weekID := int64(week.WeekId)
		if weekID == 0 {
			continue
		}

		startDate, err := s.repo.GetScheduleDateByWeek(weekID)
		if err != nil || startDate.IsZero() {
			continue
		}

		// Khởi tạo map con nếu chưa có
		if _, exists := weekLessonSet[weekID]; !exists {
			weekLessonSet[weekID] = make(map[int64]bool)
		}

		for index, sched := range week.Schedules {
			lessonID := sched.Lesson.Id
			if lessonID == 0 {
				continue
			}

			// Bỏ qua nếu lesson đã có trong tuần này
			if weekLessonSet[weekID][lessonID] {
				continue
			}

			lesson, err := s.repo.FindByID(int(lessonID))
			if err != nil || lesson.ID == 0 {
				continue
			}

			var lessonPlanId int64

			if len(lesson.LessonPlans) > 0 {
				lessonPlanId = lesson.LessonPlans[0].ID
			}

			for _, lessonPlan := range sched.Lesson.LessonPlans {
				if lessonPlan.IsActive {
					lessonPlanId = lessonPlan.Id
					break
				}
			}

			newSchedules = append(newSchedules, models.LessonSchedule{
				LessonID:      lessonID,
				CourseID:      courseID,
				ScheduledDate: startDate,
				WeekID:        weekID,
				SortPosition:  index,
				LessonPlanID:  lessonPlanId,
			})
			weekLessonSet[weekID][lessonID] = true
		}
	}

	if len(newSchedules) > 0 {
		if err := s.repo.CreateLessonSchedules(newSchedules); err != nil {
			return courseID, err
		}
	}

	return courseID, nil
}

func (s *lessonService) LessonSchedules(c *gin.Context, filters map[string]interface{}) (dto.LessonScheduleDetail, *time.Time, []models.Week, bool, error) {
	lessonIDs, err := s.repo.LessonIdsInSchedule(c, filters)
	if err != nil {
		return dto.LessonScheduleDetail{}, nil, []models.Week{}, false, err
	}

	filter := map[string]interface{}{}

	if len(lessonIDs) == 0 {
		filter["lesson_ids"] = "0"
	} else {
		idStrings := make([]string, len(lessonIDs))
		for i, lid := range lessonIDs {
			idStrings[i] = strconv.FormatInt(lid, 10)
		}

		filter["lesson_ids"] = strings.Join(idStrings, ",")
	}

	courseID, ok := filters["course_id"].(int64)
	if ok && courseID > 0 {
		filter["course_id"] = courseID
	}

	repoSchedule := repositories.NewLessonScheduleRepository()
	lessonSchedules, err := repoSchedule.GetSchedules(filter)

	if err != nil {
		return dto.LessonScheduleDetail{}, nil, []models.Week{}, false, err
	}

	return s.LessonSchedulesByWeek(c, lessonSchedules, filters)
}

func (s *lessonService) LessonSchedulesByWeek(c *gin.Context, schedules []models.LessonSchedule, filters map[string]interface{}) (dto.LessonScheduleDetail, *time.Time, []models.Week, bool, error) {
	var result []dto.LessonSchedule

	var beginDate *time.Time
	var holidayWeeks []models.Week
	var semesters []models.Semester
	var autoIndex bool

	if semesterID, ok := filters["semester_id"].(int64); ok && semesterID != 0 {
		semesterRepo := repositories.NewSemesterRepository()
		semester, err := semesterRepo.FindByID(int(semesterID))

		semesters = append(semesters, *semester)

		if semester == nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		holidayWeeks, err = semesterRepo.GetHolidayWeeksBySemesterID(semesterID)

		if err != nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		beginDate = &semester.BeginDate

		weekRepo := repositories.NewWeekRepository()

		beginWeek, err := weekRepo.GetByDate(*beginDate)

		beginDate = &beginWeek.StartDate

		if err != nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		start := semester.StartDate
		end := semester.EndDate

		weeks, err := weekRepo.FindWeeksByDateRange(start, end)
		if err != nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		scheduleMap := make(map[int64][]models.LessonSchedule)
		seen := make(map[int64]bool)

		for _, sched := range schedules {
			if seen[sched.ID] {
				continue
			}
			seen[sched.ID] = true

			scheduleMap[sched.Week.ID] = append(scheduleMap[sched.Week.ID], sched)
		}

		for _, w := range weeks {
			diffDays := int(w.StartDate.Sub(start).Hours() / 24)
			weekNumber := (diffDays / 7) + 1
			if diffDays < 0 {
				weekNumber = 0
			}

			group := dto.LessonSchedule{
				WeekId:           int(w.ID),
				WeekNumberByYear: w.WeekNumber,
				WeekNumber:       weekNumber,
				Year:             w.Year,
				StartDate:        w.StartDate,
				EndDate:          w.EndDate,
				Lessons:          scheduleMap[w.ID],
			}
			result = append(result, group)
		}
	} else if _, ok := filters["week_by_course"].(bool); ok {
		courseID, ok := filters["course_id"].(int64)

		if courseID == 0 || !ok {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, nil
		}

		courseRepo := repositories.NewCourseRepository()
		weakRepo := repositories.NewWeekRepository()

		courseRepo.SetContext(c)
		courseRepo.SetPreload([]string{"Semesters"})
		course, err := courseRepo.FindByID(int(courseID))

		if err != nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		start := course.StartDate
		duration := course.Duration
		end := start.Add(time.Duration(duration*7) * time.Hour * 24)

		if len(course.Semesters) > 0 {
			semesters = append(semesters, course.Semesters...)
			semesterRepo := repositories.NewSemesterRepository()

			for _, s := range course.Semesters {
				if s.EndDate.After(end) {
					endWWeek, err := weakRepo.GetByDate(s.EndDate)

					if err == nil {
						end = endWWeek.EndDate
					} else {
						end = s.EndDate
					}
				}

				if s.BeginDate.Before(start) {
					startWWeek, err := weakRepo.GetByDate(s.BeginDate)

					if err == nil {
						start = startWWeek.StartDate
					} else {
						start = s.BeginDate
					}
				}

				holidays, _ := semesterRepo.GetHolidayWeeksBySemesterID(s.ID)

				if len(holidays) > 0 {
					holidayWeeks = append(holidayWeeks, holidays...)
				}
			}
		}

		beginDate = &start

		weekRepo := repositories.NewWeekRepository()

		weeks, err := weekRepo.FindWeeksByDateRange(start, end)
		if err != nil {
			return dto.LessonScheduleDetail{}, beginDate, holidayWeeks, autoIndex, err
		}

		scheduleMap := make(map[int64][]models.LessonSchedule)
		seen := make(map[int64]bool)

		for _, sched := range schedules {
			if seen[sched.ID] {
				continue
			}
			seen[sched.ID] = true

			scheduleMap[sched.Week.ID] = append(scheduleMap[sched.Week.ID], sched)
		}

		for _, w := range weeks {
			diffDays := int(w.StartDate.Sub(start).Hours() / 24)
			weekNumber := (diffDays / 7) + 1
			if diffDays < 0 {
				weekNumber = 0
			}

			group := dto.LessonSchedule{
				WeekId:           int(w.ID),
				WeekNumberByYear: w.WeekNumber,
				WeekNumber:       weekNumber,
				Year:             w.Year,
				StartDate:        w.StartDate,
				EndDate:          w.EndDate,
				Lessons:          scheduleMap[w.ID],
			}
			result = append(result, group)
		}
	} else {
		groupMap := make(map[string]*dto.LessonSchedule)

		for _, sched := range schedules {
			year := sched.Week.Year
			weekId := sched.Week.ID
			weekNumber := sched.Week.WeekNumber
			startDate := sched.Week.StartDate
			endDate := sched.Week.EndDate

			if beginDate == nil || startDate.Before(*beginDate) {
				beginDate = &startDate
			}

			key := fmt.Sprintf("%d-%d", year, weekNumber)

			if _, exists := groupMap[key]; !exists {
				groupMap[key] = &dto.LessonSchedule{
					WeekId:           int(weekId),
					WeekNumber:       weekNumber,
					WeekNumberByYear: weekNumber,
					Year:             year,
					StartDate:        startDate,
					EndDate:          endDate,
					Lessons:          []models.LessonSchedule{},
				}
			}

			lessonExists := false
			for _, existing := range groupMap[key].Lessons {
				if existing.LessonID == sched.LessonID {
					lessonExists = true
					break
				}
			}
			if !lessonExists {
				groupMap[key].Lessons = append(groupMap[key].Lessons, sched)
			}
		}

		for _, g := range groupMap {
			result = append(result, *g)
		}

		autoIndex = true
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Year == result[j].Year {
			return result[i].WeekNumber < result[j].WeekNumber
		}
		return result[i].Year < result[j].Year
	})

	return dto.LessonScheduleDetail{
		Semesters: semesters,
		Groups:    result,
	}, beginDate, holidayWeeks, autoIndex, nil
}

func (s *lessonService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	filters := make(map[string]interface{})

	if weekIDStr := c.Query("schedule_week_id"); weekIDStr != "" {
		if weekID, err := strconv.ParseInt(weekIDStr, 10, 64); err == nil {
			filters["schedule_week_id"] = weekID
		}
	}

	if scheduleCourseIDStr := c.Query("schedule_course_id"); scheduleCourseIDStr != "" {
		if courseID, err := strconv.ParseInt(scheduleCourseIDStr, 10, 64); err == nil {
			filters["schedule_course_id"] = courseID
		}
	}

	if byUserParam := c.Query("schedule_by_user"); byUserParam == "true" || byUserParam == "1" || byUserParam == "yes" {
		userID := utils.GetCurrentUserId(c)
		filters["user_id"] = userID
	}

	if len(filters) > 0 {
		lessonIDs, err := s.repo.LessonIdsInSchedule(c, filters)
		if err != nil {
			return nil, err
		}

		if len(lessonIDs) > 0 {
			idStrings := make([]string, len(lessonIDs))
			for i, id := range lessonIDs {
				idStrings[i] = strconv.FormatInt(id, 10)
			}
			filter["id"] = "in:" + strings.Join(idStrings, ",")
		} else {
			filter["id"] = "in:0"
		}
	}

	programIDStr := c.Query("program_id")

	if programIDStr != "" {
		filter["Chapter.program_id"] = programIDStr + ":lessons:chapters:chapter_id:id:program_id"
	}

	return filter, nil
}

func (s *lessonService) Completion(c *gin.Context, lessonCompletion *prot.LessonCompletion) (*prot.LessonCompletion, error) {
	return s.repo.Completion(lessonCompletion)
}

func (s *lessonService) CompletionLessonIds(c *gin.Context, courseId int64, chapterId int64) []int64 {
	studentId := utils.GetCurrentUserId(c)
	if utils.GetCurrentRoleId(c) != models.StudentRoleId {
		return []int64{}
	}
	return s.repo.CompletionLessonIds(int64(studentId), courseId, chapterId)
}

func (s *lessonService) StudyingLessonIds(c *gin.Context, courseId int64, chapterId int64) []int64 {
	studentId := utils.GetCurrentUserId(c)
	if utils.GetCurrentRoleId(c) != models.StudentRoleId {
		return []int64{}
	}
	return s.repo.StudyingLessonIds(int64(studentId), courseId, chapterId)
}

func (s *lessonService) HideLessonIds(c *gin.Context) []int64 {
	userId := utils.GetCurrentUserId(c)
	roleId := utils.GetCurrentRoleId(c)

	if roleId != models.StudentRoleId && roleId != models.TeacherRoleId {
		return []int64{}
	}

	ids := s.repo.HideLessonIds(int64(userId))

	return ids
}

func (s *lessonService) Studying(c *gin.Context, lessonStudying *prot.LessonStudying) (*prot.LessonStudying, error) {
	return s.repo.Studying(lessonStudying)
}

func (s *lessonService) CompletionLessonPlanIds(c *gin.Context, courseId int64, chapterId int64, lessonId int64) []int64 {
	return s.repo.CompletionLessonPlanIds(courseId, chapterId, lessonId)
}

func (s *lessonService) GetLessonPlanByCourse(c *gin.Context, lessonId, courseId int64) (*prot.LessonPlanByCourse, error) {
	lessonPlans, lessonPlanByCourses, err := s.repo.GetLessonPlanByCourse(lessonId, courseId)
	return s.ReturnLessonPlanByCourse(lessonPlans, lessonPlanByCourses, err)
}

func (s *lessonService) GetExamByCourse(c *gin.Context, lessonId, courseId int64) (*prot.ExamByCourse, error) {
	exams, examByCourses, err := s.repo.GetExamByCourse(lessonId, courseId)
	return s.ReturnExamByCourse(exams, examByCourses, courseId, err)
}

func (s *lessonService) GetHomeworkByCourse(c *gin.Context, lessonId, courseId int64) (*prot.HomeworkByCourse, error) {
	homeworks, homeworkByCourses, err := s.repo.GetHomeworkByCourse(lessonId, courseId)
	return s.ReturnHomeworkByCourse(homeworks, homeworkByCourses, courseId, err)
}

func (s *lessonService) GetExerciseByCourse(c *gin.Context, lessonId, courseId int64) (*prot.ExerciseByCourse, error) {
	exercises, exerciseByCourses, err := s.repo.GetExerciseByCourse(lessonId, courseId)
	return s.ReturnExerciseByCourse(exercises, exerciseByCourses, courseId, err)
}

func (s *lessonService) StoreLessonPlanByCourse(c *gin.Context, lessonId, courseId int64, req *prot.LessonPlanByCourse) (*prot.LessonPlanByCourse, error) {
	lessonPlanReq := req.LessonPlans
	lessonPlanIds := []int64{}

	for _, lessonPlan := range lessonPlanReq {
		lessonPlanIds = append(lessonPlanIds, lessonPlan.Id)
	}

	lessonPlans, lessonPlanByCourses, err := s.repo.StoreLessonPlanByCourse(lessonId, courseId, lessonPlanIds)
	return s.ReturnLessonPlanByCourse(lessonPlans, lessonPlanByCourses, err)
}

func (s *lessonService) StoreExamByCourse(c *gin.Context, lessonId, courseId int64, req *prot.ExamByCourse) (*prot.ExamByCourse, error) {
	examReq := req.Exams
	examIds := []int64{}

	for _, exam := range examReq {
		examIds = append(examIds, exam.Id)
	}

	exams, examByCourses, err := s.repo.StoreExamByCourse(lessonId, courseId, examIds)
	return s.ReturnExamByCourse(exams, examByCourses, courseId, err)
}

func (s *lessonService) StoreHomeworkByCourse(c *gin.Context, lessonId, courseId int64, req *prot.HomeworkByCourse) (*prot.HomeworkByCourse, error) {
	homeworkReq := req.Homeworks
	homeworkIds := []int64{}

	for _, homework := range homeworkReq {
		homeworkIds = append(homeworkIds, homework.Id)
	}

	homeworks, homeworkByCourses, err := s.repo.StoreHomeworkByCourse(lessonId, courseId, homeworkIds)
	return s.ReturnHomeworkByCourse(homeworks, homeworkByCourses, courseId, err)
}

func (s *lessonService) StoreExerciseByCourse(c *gin.Context, lessonId, courseId int64, req *prot.ExerciseByCourse) (*prot.ExerciseByCourse, error) {
	exerciseReq := req.Exercises
	exerciseIds := []int64{}

	for _, exercise := range exerciseReq {
		exerciseIds = append(exerciseIds, exercise.Id)
	}

	exercises, exerciseByCourses, err := s.repo.StoreExerciseByCourse(lessonId, courseId, exerciseIds)
	return s.ReturnExerciseByCourse(exercises, exerciseByCourses, courseId, err)
}

func (s *lessonService) ReturnLessonPlanByCourse(lessonPlans, lessonPlanByCourses []*models.LessonPlan, err error) (*prot.LessonPlanByCourse, error) {
	if err != nil {
		return nil, err
	}

	lessonPlanResource := resources.NewLessonPlanResource()
	lessonPlanFormat := lessonPlanResource.FormatLessonPlans(lessonPlans)
	lessonPlanByCourseFormat := lessonPlanResource.FormatLessonPlans(lessonPlanByCourses)

	if len(lessonPlanByCourseFormat) == 0 {
		lessonPlanByCourseFormat = lessonPlanFormat
	}

	return &prot.LessonPlanByCourse{
		LessonPlanByPrograms: lessonPlanFormat,
		LessonPlans:          lessonPlanByCourseFormat,
	}, nil
}

func (s *lessonService) ReturnExamByCourse(exams, examByCourses []*models.Exam, courseId int64, err error) (*prot.ExamByCourse, error) {
	if err != nil {
		return nil, err
	}

	examResource := resources.NewExamResource()
	if impl, ok := examResource.(*resources.ExamResourceImpl); ok {
		impl.CourseId = courseId
	}

	examFormat := examResource.FormatExams(exams)
	examByCourseFormat := examResource.FormatExams(examByCourses)

	if len(examByCourseFormat) == 0 {
		examByCourseFormat = examFormat
	}

	return &prot.ExamByCourse{
		ExamByPrograms: examFormat,
		Exams:          examByCourseFormat,
	}, nil
}

func (s *lessonService) ReturnHomeworkByCourse(homeworks, homeworkByCourses []*models.Homework, courseId int64, err error) (*prot.HomeworkByCourse, error) {
	if err != nil {
		return nil, err
	}

	homeworkResource := resources.NewHomeworkResource()
	if impl, ok := homeworkResource.(*resources.HomeworkResourceImpl); ok {
		impl.CourseId = courseId
	}

	homeworkFormat := homeworkResource.FormatHomeworks(homeworks)
	homeworkByCourseFormat := homeworkResource.FormatHomeworks(homeworkByCourses)

	if len(homeworkByCourseFormat) == 0 {
		homeworkByCourseFormat = homeworkFormat
	}

	return &prot.HomeworkByCourse{
		Homeworks:          homeworkByCourseFormat,
		HomeworkByPrograms: homeworkFormat,
	}, nil
}

func (s *lessonService) ReturnExerciseByCourse(exercises, exerciseByCourses []*models.Exercise, courseId int64, err error) (*prot.ExerciseByCourse, error) {
	if err != nil {
		return nil, err
	}

	exerciseResource := resources.NewExerciseResource()
	if impl, ok := exerciseResource.(*resources.ExerciseResourceImpl); ok {
		impl.CourseId = courseId
	}

	exerciseFormat := exerciseResource.FormatExercises(exercises)
	exerciseByCourseFormat := exerciseResource.FormatExercises(exerciseByCourses)

	if len(exerciseByCourseFormat) == 0 {
		exerciseByCourseFormat = exerciseFormat
	}

	return &prot.ExerciseByCourse{
		Exercises:          exerciseByCourseFormat,
		ExerciseByPrograms: exerciseFormat,
	}, nil
}

func (s *lessonService) StoreTeachingPlans(c *gin.Context, id int64, req *prot.LessonRequest) error {
	var teachingPlanIds []int64

	if req.TeachingPlan != nil && req.TeachingPlan.Id > 0 {
		teachingPlanIds = append(teachingPlanIds, req.TeachingPlan.Id)
	}

	s.repo.UpdateTeachingPlans(id, teachingPlanIds)

	return nil
}
