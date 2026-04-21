package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories/base"
	"be-lms/requests"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardListEntityRepository interface {
	GetSchools(c *gin.Context, userID int64, onlyUserSchools bool, req *requests.DashboardSchoolListRequest) ([]dto.DashboardSchool, int64, error)
	GetCourses(c *gin.Context, userID int64, schoolID int64, onlyUserCourses bool, req *requests.DashboardCourseListRequest) ([]dto.DashboardCourse, int64, error)
	GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) ([]dto.DashboardTeacher, int64, error)
	GetSubjects(req *requests.DashboardSubjectListRequest) ([]dto.DashboardSubject, int64, error)
	GetExams(userID int64, onlyUserExams bool, req *requests.DashboardExamListRequest) ([]dto.DashboardExam, int64, error)
	GetHomeworks(userID int64, onlyUserHomeworks bool, req *requests.DashboardHomeworkListRequest) ([]dto.DashboardHomework, int64, error)
	GetLessons(userID int64, onlyUserLessons bool, req *requests.DashboardLessonListRequest) ([]dto.DashboardLesson, int64, error)
	GetChapters(req *requests.DashboardChapterListRequest) ([]dto.DashboardChapter, int64, error)
	GetClasses(c *gin.Context, req *requests.DashboardClassListRequest) ([]dto.DashboardClass, int64, error)
	GetClassMains(c *gin.Context, req *requests.DashboardClassMainListRequest) ([]dto.DashboardClassMain, int64, error)
}

type dashboardListEntityRepository struct{}

func NewDashboardListEntityRepository() DashboardListEntityRepository {
	return &dashboardListEntityRepository{}
}

func (r *dashboardListEntityRepository) GetSchools(c *gin.Context, userID int64, onlyUserSchools bool, req *requests.DashboardSchoolListRequest) ([]dto.DashboardSchool, int64, error) {
	var schools []dto.DashboardSchool
	var totalCount int64

	query := db.ReplicaDB.Table("schools").
		Select("DISTINCT schools.id, schools.name as school_name, schools.short_name as school_short_name, wards.name as ward_name, provinces.name as province_name").
		Joins("LEFT JOIN wards ON schools.ward_code = wards.code").
		Joins("LEFT JOIN provinces ON wards.province_code = provinces.code").
		Where("schools.deleted_at IS NULL")

	// Nếu role = 2 (giáo viên), chỉ lấy các trường thuộc giáo viên đó
	// Join: users -> user_courses -> courses -> course_schools -> schools
	if onlyUserSchools && userID > 0 {
		query = query.
			Joins("JOIN course_schools cs ON cs.school_id = schools.id").
			Joins("JOIN courses c ON c.id = cs.course_id").
			Joins("JOIN user_courses uc ON uc.course_id = c.id").
			Joins("JOIN users u ON u.id = uc.user_id").
			Where("u.id = ?", userID).
			Where("u.deleted_at IS NULL").
			Where("c.deleted_at IS NULL")
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	baseRepo := base.NewBaseRepository[models.School]()
	schoolId := baseRepo.GetAdminSchoolId(c)

	if schoolId > 0 {
		query = query.Where("schools.id = ?", schoolId)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("schools.name ASC")
	}

	err := query.Find(&schools).Error
	return schools, totalCount, err
}

func (r *dashboardListEntityRepository) GetCourses(c *gin.Context, userID int64, schoolID int64, onlyUserCourses bool, req *requests.DashboardCourseListRequest) ([]dto.DashboardCourse, int64, error) {
	var courses []dto.DashboardCourse
	var totalCount int64

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)

	if schoolIdByRole > 0 {
		schoolID = int64(schoolIdByRole)
	}

	query := db.ReplicaDB.Table("courses").
		Select("DISTINCT courses.id, courses.name, courses.object_title, courses.parent_course_id, courses.program_id, programs.name AS program_name").
		Joins("LEFT JOIN programs ON courses.program_id = programs.id").
		Where("courses.deleted_at IS NULL")

	if onlyUserCourses && userID > 0 {
		// var courseIds []int64
		// db.ReplicaDB.Model(&models.UserCourse{}).
		// 	Where("user_id = ?", userID).
		// 	Pluck("course_id", &courseIds)

		// memberType, _ := c.Get("memberType")

		// if len(courseIds) == 0 && memberType.(string) == models.MemberTypeExternal {
		// 	courseIds = config.LoadConfig().PublicCourseIds
		// }

		// query = query.Where("courses.id IN ?", courseIds)

		query = query.Joins("JOIN user_courses ON courses.id = user_courses.course_id").
			Where("user_courses.user_id = ?", userID)
		if schoolID > 0 {
			query = query.Joins("LEFT JOIN course_schools ON courses.id = course_schools.course_id").
				Where("course_schools.school_id = ?", schoolID)
		}
	} else {
		query = query.Joins("LEFT JOIN course_schools ON courses.id = course_schools.course_id")
		if schoolID > 0 {
			query = query.Where("course_schools.school_id = ?", schoolID)
		}
	}

	if req.ProgramID > 0 {
		query = query.Where("courses.program_id = ?", req.ProgramID)
	}

	if req.SubjectID > 0 {
		query = query.Where("programs.subject_id = ?", req.SubjectID)
	}

	if req.ParentCourseID != nil {
		query = query.Where("courses.parent_course_id = ?", *req.ParentCourseID)
	}

	// Filter theo assessment_id: lấy distinct course_id từ assessment_ref_lessons
	if req.AssessmentID != nil && *req.AssessmentID > 0 {
		query = query.
			Joins("INNER JOIN assessment_ref_lessons ON assessment_ref_lessons.course_id = courses.id").
			Where("assessment_ref_lessons.assessment_id = ?", *req.AssessmentID)
	}

	// Count total - dùng DISTINCT để đếm đúng khi có JOIN
	countQuery := db.ReplicaDB.Table("courses").
		Select("DISTINCT courses.id").
		Joins("LEFT JOIN programs ON courses.program_id = programs.id").
		Where("courses.deleted_at IS NULL")

	if onlyUserCourses && userID > 0 {
		countQuery = countQuery.Joins("JOIN user_courses ON courses.id = user_courses.course_id").
			Where("user_courses.user_id = ?", userID)
		if schoolID > 0 {
			countQuery = countQuery.Joins("LEFT JOIN course_schools ON courses.id = course_schools.course_id").
				Where("course_schools.school_id = ?", schoolID)
		}
	} else {
		countQuery = countQuery.Joins("LEFT JOIN course_schools ON courses.id = course_schools.course_id")
		if schoolID > 0 {
			countQuery = countQuery.Where("course_schools.school_id = ?", schoolID)
		}
	}

	if req.ProgramID > 0 {
		countQuery = countQuery.Where("courses.program_id = ?", req.ProgramID)
	}

	if req.SubjectID > 0 {
		countQuery = countQuery.Where("programs.subject_id = ?", req.SubjectID)
	}

	if req.ParentCourseID != nil {
		countQuery = countQuery.Where("courses.parent_course_id = ?", *req.ParentCourseID)
	}

	if req.AssessmentID != nil && *req.AssessmentID > 0 {
		countQuery = countQuery.
			Joins("INNER JOIN assessment_ref_lessons ON assessment_ref_lessons.course_id = courses.id").
			Where("assessment_ref_lessons.assessment_id = ?", *req.AssessmentID)
	}

	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("courses.id ASC")
	}

	err := query.Find(&courses).Error
	return courses, totalCount, err
}

func (r *dashboardListEntityRepository) GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) ([]dto.DashboardTeacher, int64, error) {
	var teachers []dto.DashboardTeacher
	var totalCount int64

	query := db.ReplicaDB.Table("users").
		Select("DISTINCT users.id, users.name").
		Joins("LEFT JOIN user_courses ON users.id = user_courses.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("users.deleted_at IS NULL AND urr.role_id = 2")

	// Filter by course_id if provided
	if req.CourseID > 0 {
		query = query.Where("user_courses.course_id = ?", req.CourseID)
	}

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)

	if schoolIdByRole > 0 {
		req.SchoolID = int64(schoolIdByRole)
	}

	if req.ProgramID > 0 {
		query = query.Joins("LEFT JOIN courses ON user_courses.course_id = courses.id").
			Where("courses.program_id = ?", req.ProgramID)
	}

	// Filter by school_id if provided
	if req.SchoolID > 0 {
		query = query.Joins("JOIN user_classes urc ON urc.user_id = users.id").
			Joins("JOIN classes c ON c.id = urc.class_id").
			Where("c.school_id = ?", req.SchoolID)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("users.id ASC")
	}

	err := query.Find(&teachers).Error
	return teachers, totalCount, err
}

func (r *dashboardListEntityRepository) GetSubjects(req *requests.DashboardSubjectListRequest) ([]dto.DashboardSubject, int64, error) {
	var subjects []dto.DashboardSubject
	var totalCount int64

	query := db.ReplicaDB.Table("subjects").
		Select("DISTINCT subjects.id, subjects.name").
		Where("subjects.deleted_at IS NULL")

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("subjects.id ASC")
	}

	err := query.Find(&subjects).Error
	return subjects, totalCount, err
}

func (r *dashboardListEntityRepository) GetExams(userID int64, onlyUserExams bool, req *requests.DashboardExamListRequest) ([]dto.DashboardExam, int64, error) {
	var exams []dto.DashboardExam
	var totalCount int64

	// Tạo subquery để count total chính xác
	countQuery := db.ReplicaDB.Table("exams").
		Select("COUNT(DISTINCT exams.id)").
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.exam_id = exams.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Joins("LEFT JOIN lessons ON erl.lesson_id = lessons.id").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN subjects ON courses.subject_id = subjects.id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("exams.deleted_at IS NULL")

	if req.CourseID > 0 {
		countQuery = countQuery.Where("courses.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		countQuery = countQuery.Where("subjects.id = ?", req.SubjectID)
	}
	if req.LessonID > 0 {
		countQuery = countQuery.Where("lessons.id = ?", req.LessonID)
	}
	if onlyUserExams && userID > 0 {
		countQuery = countQuery.Where("user_courses.user_id = ?", userID)
	}

	if err := countQuery.Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Query chính để lấy data
	query := db.ReplicaDB.Table("exams").
		Select("DISTINCT exams.id, exams.name, courses.id as course_id, subjects.id as subject_id").
		Joins("LEFT JOIN exam_ref_lessons erl ON erl.exam_id = exams.id AND erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Joins("LEFT JOIN lessons ON erl.lesson_id = lessons.id").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN subjects ON courses.subject_id = subjects.id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("exams.deleted_at IS NULL")

	if req.CourseID > 0 {
		query = query.Where("courses.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		query = query.Where("subjects.id = ?", req.SubjectID)
	}
	if req.LessonID > 0 {
		query = query.Where("lessons.id = ?", req.LessonID)
	}
	if onlyUserExams && userID > 0 {
		query = query.Where("user_courses.user_id = ?", userID)
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("exams.id ASC")
	}

	err := query.Find(&exams).Error
	return exams, totalCount, err
}

func (r *dashboardListEntityRepository) GetHomeworks(userID int64, onlyUserHomeworks bool, req *requests.DashboardHomeworkListRequest) ([]dto.DashboardHomework, int64, error) {
	var homeworks []dto.DashboardHomework
	var totalCount int64

	// Parse start_date và end_date (có thể là Unix timestamp hoặc YYYY-MM-DD)
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if timestamp, err := strconv.ParseInt(req.StartDate, 10, 64); err == nil {
			// Unix timestamp
			t := time.Unix(timestamp, 0)
			startDate = &t
		} else if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			// YYYY-MM-DD format
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if timestamp, err := strconv.ParseInt(req.EndDate, 10, 64); err == nil {
			// Unix timestamp
			t := time.Unix(timestamp, 0)
			endDate = &t
		} else if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			// YYYY-MM-DD format
			endDate = &t
		}
	}

	// Parse homework_ids từ string (cách nhau bởi dấu phẩy) thành array
	var homeworkIDs []int64
	if req.HomeworkIDs != "" {
		idsStr := strings.Split(req.HomeworkIDs, ",")
		for _, idStr := range idsStr {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				homeworkIDs = append(homeworkIDs, id)
			}
		}
	}

	// Parse lesson_ids từ string (cách nhau bởi dấu phẩy) thành array
	var lessonIDs []int64
	if req.LessonIDs != "" {
		idsStr := strings.Split(req.LessonIDs, ",")
		for _, idStr := range idsStr {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				lessonIDs = append(lessonIDs, id)
			}
		}
	}
	// Nếu có LessonID (single) và chưa có LessonIDs, thêm vào
	if req.LessonID > 0 && len(lessonIDs) == 0 {
		lessonIDs = append(lessonIDs, req.LessonID)
	}

	// Tạo subquery để count total chính xác - phải có logic giống query chính
	// Join giống như teacher/homework: INNER JOIN với lesson_schedules và weeks (BẮT BUỘC)
	countQuery := db.ReplicaDB.Table("homeworks h").
		Select("COUNT(DISTINCT h.id)").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON hrl.lesson_id = l.id").
		Joins("JOIN chapters ch ON l.chapter_id = ch.id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("LEFT JOIN subjects s ON c.subject_id = s.id").
		Joins("JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("LEFT JOIN user_courses uc ON c.id = uc.course_id").
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_at IS NOT NULL")

	if req.CourseID > 0 {
		countQuery = countQuery.Where("c.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		countQuery = countQuery.Where("s.id = ?", req.SubjectID)
	}
	if req.ChapterID > 0 {
		countQuery = countQuery.Where("ch.id = ?", req.ChapterID)
	}
	if len(lessonIDs) > 0 {
		countQuery = countQuery.Where("l.id IN (?)", lessonIDs)
	} else if req.LessonID > 0 {
		countQuery = countQuery.Where("l.id = ?", req.LessonID)
	}
	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		countQuery = countQuery.Where("h.id IN (?)", homeworkIDs)
	}
	// Filter theo start_date và end_date qua weeks (giống teacher/homework)
	if startDate != nil && endDate != nil {
		countQuery = countQuery.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}
	if req.IsAssigned != nil {
		if *req.IsAssigned {
			// Chỉ lấy homework đã được giao (assigned_at IS NOT NULL)
			countQuery = countQuery.Where("hrl.assigned_at IS NOT NULL")
		} else {
			// Lấy homework chưa được giao (assigned_at IS NULL)
			countQuery = countQuery.Where("hrl.assigned_at IS NULL")
		}
	}
	if onlyUserHomeworks && userID > 0 {
		countQuery = countQuery.Where("uc.user_id = ?", userID)
	}

	if err := countQuery.Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Query chính để lấy data - sử dụng DISTINCT ON để tránh duplicate
	// Join giống như teacher/homework: INNER JOIN với lesson_schedules và weeks (BẮT BUỘC)
	query := db.ReplicaDB.Table("homeworks h").
		Select("DISTINCT ON (h.id) h.id, h.name, c.id as course_id, s.id as subject_id, c.name as course_name, s.name as subject_name, hrl.lesson_id, l.title as lesson_title, hrl.assigned_at").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons l ON hrl.lesson_id = l.id").
		Joins("JOIN chapters ch ON l.chapter_id = ch.id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Joins("LEFT JOIN subjects s ON c.subject_id = s.id").
		Joins("JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("JOIN weeks w ON ls.week_id = w.id").
		Joins("LEFT JOIN user_courses uc ON c.id = uc.course_id").
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_at IS NOT NULL")

	if req.CourseID > 0 {
		query = query.Where("c.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		query = query.Where("s.id = ?", req.SubjectID)
	}
	if req.ChapterID > 0 {
		query = query.Where("ch.id = ?", req.ChapterID)
	}
	if len(lessonIDs) > 0 {
		query = query.Where("l.id IN (?)", lessonIDs)
	} else if req.LessonID > 0 {
		query = query.Where("l.id = ?", req.LessonID)
	}
	// Filter theo homework_ids nếu có
	if len(homeworkIDs) > 0 {
		query = query.Where("h.id IN (?)", homeworkIDs)
	}
	// Filter theo start_date và end_date qua weeks (giống teacher/homework)
	if startDate != nil && endDate != nil {
		query = query.Where("? <= w.start_date AND ? >= w.end_date", *startDate, *endDate)
	}
	if req.IsAssigned != nil {
		if *req.IsAssigned {
			// Chỉ lấy homework đã được giao (assigned_at IS NOT NULL)
			query = query.Where("hrl.assigned_at IS NOT NULL")
		} else {
			// Lấy homework chưa được giao (assigned_at IS NULL)
			query = query.Where("hrl.assigned_at IS NULL")
		}
	}
	if onlyUserHomeworks && userID > 0 {
		query = query.Where("uc.user_id = ?", userID)
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}
	if len(req.Sort) > 0 {
		// Với DISTINCT ON, ORDER BY phải bắt đầu với h.id
		query = query.Order("h.id ASC")
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("h.id ASC")
	}

	err := query.Find(&homeworks).Error
	return homeworks, totalCount, err
}

func (r *dashboardListEntityRepository) GetLessons(userID int64, onlyUserLessons bool, req *requests.DashboardLessonListRequest) ([]dto.DashboardLesson, int64, error) {
	var lessons []dto.DashboardLesson
	var totalCount int64

	// Tạo subquery để count total chính xác
	countQuery := db.ReplicaDB.Table("lessons").
		Select("COUNT(DISTINCT lessons.id)").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("lessons.deleted_at IS NULL")

	if req.CourseID > 0 {
		countQuery = countQuery.Where("courses.id = ?", req.CourseID)
	}
	if req.ChapterID > 0 {
		// Filter trực tiếp theo lessons.chapter_id (foreign key trực tiếp)
		countQuery = countQuery.Where("lessons.chapter_id = ?", req.ChapterID)
	}
	if onlyUserLessons && userID > 0 {
		countQuery = countQuery.Where("user_courses.user_id = ?", userID)
	}

	if err := countQuery.Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Query chính để lấy data
	query := db.ReplicaDB.Table("lessons").
		Select("DISTINCT lessons.id, lessons.title, chapters.id as chapter_id, chapters.title as chapter_name, courses.id as course_id, courses.name as course_name").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("lessons.deleted_at IS NULL")

	if req.CourseID > 0 {
		query = query.Where("courses.id = ?", req.CourseID)
	}
	if req.ChapterID > 0 {
		// Filter trực tiếp theo lessons.chapter_id (foreign key trực tiếp)
		query = query.Where("lessons.chapter_id = ?", req.ChapterID)
	}
	if onlyUserLessons && userID > 0 {
		query = query.Where("user_courses.user_id = ?", userID)
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("lessons.id ASC")
	}

	err := query.Find(&lessons).Error
	return lessons, totalCount, err
}

func (r *dashboardListEntityRepository) GetChapters(req *requests.DashboardChapterListRequest) ([]dto.DashboardChapter, int64, error) {
	var chapters []dto.DashboardChapter
	var totalCount int64

	query := db.ReplicaDB.Table("chapters").
		Select("DISTINCT chapters.id, chapters.title, courses.id as course_id, courses.name as course_name").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Where("chapters.deleted_at IS NULL")

	if req.CourseID > 0 {
		query = query.Where("courses.id = ?", req.CourseID)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("chapters.id ASC")
	}

	err := query.Find(&chapters).Error
	return chapters, totalCount, err
}

func (r *dashboardListEntityRepository) GetClasses(c *gin.Context, req *requests.DashboardClassListRequest) ([]dto.DashboardClass, int64, error) {
	var classes []dto.DashboardClass
	var totalCount int64

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)

	schoolID := req.SchoolID
	if schoolIdByRole > 0 {
		schoolID = int64(schoolIdByRole)
	}

	query := db.ReplicaDB.Table("classes c").
		Select("DISTINCT c.id, c.name AS class_name, c.class_main_id, COALESCE(cm.name, '') AS class_main_name").
		Joins("LEFT JOIN classes_main cm ON c.class_main_id = cm.id AND cm.deleted_at IS NULL").
		Where("c.deleted_at IS NULL")

	if schoolID > 0 {
		query = query.Where("c.school_id = ?", schoolID)
	}

	if req.ClassMainID > 0 {
		query = query.Where("c.class_main_id = ?", req.ClassMainID)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("c.id ASC")
	}

	err := query.Find(&classes).Error
	return classes, totalCount, err
}

func (r *dashboardListEntityRepository) GetClassMains(c *gin.Context, req *requests.DashboardClassMainListRequest) ([]dto.DashboardClassMain, int64, error) {
	var classMains []dto.DashboardClassMain
	var totalCount int64

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)

	schoolID := req.SchoolID
	if schoolIdByRole > 0 {
		schoolID = int64(schoolIdByRole)
	}

	query := db.ReplicaDB.Table("classes_main cm").
		Select("cm.id, cm.name, cm.school_id").
		Where("cm.deleted_at IS NULL")

	if schoolID > 0 {
		query = query.Where("cm.school_id = ?", schoolID)
	}

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}

	// Apply sorting
	if len(req.Sort) > 0 {
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("cm.id ASC")
	}

	err := query.Find(&classMains).Error
	return classMains, totalCount, err
}
