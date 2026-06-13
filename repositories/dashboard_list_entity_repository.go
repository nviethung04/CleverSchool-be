package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/repositories/base"
	"be-lms/requests"

	"github.com/gin-gonic/gin"
)

type DashboardListEntityRepository interface {
	GetSchools(c *gin.Context,req *requests.DashboardSchoolListRequest) ([]dto.DashboardSchool, int64, error)
	GetPrograms(c *gin.Context, schoolID int64, req *requests.DashboardProgramListRequest) ([]dto.DashboardProgram, int64, error)
	GetCourses(c *gin.Context, userID int64, schoolID int64, onlyUserCourses bool, req *requests.DashboardCourseListRequest) ([]dto.DashboardCourse, int64, error)
	GetTeachers(c *gin.Context, req *requests.DashboardTeacherListRequest) ([]dto.DashboardTeacher, int64, error)
	GetSubjects(req *requests.DashboardSubjectListRequest) ([]dto.DashboardSubject, int64, error)
	GetExams(userID int64, onlyUserExams bool, req *requests.DashboardExamListRequest) ([]dto.DashboardExam, int64, error)
	GetHomeworks(userID int64, onlyUserHomeworks bool, req *requests.DashboardHomeworkListRequest) ([]dto.DashboardHomework, int64, error)
	GetLessons(userID int64, onlyUserLessons bool, req *requests.DashboardLessonListRequest) ([]dto.DashboardLesson, int64, error)
}

type dashboardListEntityRepository struct{}

func NewDashboardListEntityRepository() DashboardListEntityRepository {
	return &dashboardListEntityRepository{}
}

func (r *dashboardListEntityRepository) GetSchools(c *gin.Context, req *requests.DashboardSchoolListRequest) ([]dto.DashboardSchool, int64, error) {
	var schools []dto.DashboardSchool
	var totalCount int64

	query := db.ReplicaDB.Table("schools").
		Select("DISTINCT schools.id, schools.name as school_name, schools.short_name as school_short_name, wards.name as ward_name, provinces.name as province_name").
		Joins("LEFT JOIN wards ON schools.ward_code = wards.code").
		Joins("LEFT JOIN provinces ON wards.province_code = provinces.code").
		Where("schools.deleted_at IS NULL")

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
		query = query.Order("schools.id ASC")
	}

	err := query.Find(&schools).Error
	return schools, totalCount, err
}

func (r *dashboardListEntityRepository) GetPrograms(c *gin.Context, schoolID int64, req *requests.DashboardProgramListRequest) ([]dto.DashboardProgram, int64, error) {
	var programs []dto.DashboardProgram
	var totalCount int64

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)
	if schoolIdByRole > 0 {
		schoolID = int64(schoolIdByRole)
	}

	query := db.ReplicaDB.Table("programs").
		Select("DISTINCT programs.id, programs.name").
		Where("programs.deleted_at IS NULL")

	if schoolID > 0 {
		query = query.
			Joins("JOIN courses ON courses.program_id = programs.id AND courses.deleted_at IS NULL").
			Joins("JOIN course_schools ON course_schools.course_id = courses.id").
			Where("course_schools.school_id = ?", schoolID)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
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
		query = query.Order("programs.id ASC")
	}

	err := query.Find(&programs).Error
	return programs, totalCount, err
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
		Select("DISTINCT courses.id, courses.name, courses.object_title").
		Where("courses.deleted_at IS NULL")

	if onlyUserCourses && userID > 0 {
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
		req.SchoolID  = int64(schoolIdByRole)
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

	// Tạo subquery để count total chính xác - phải có logic giống query chính
	countQuery := db.ReplicaDB.Table("homeworks").
		Select("COUNT(DISTINCT homeworks.id)").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
		Joins("LEFT JOIN lessons ON hrl.lesson_id = lessons.id").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN subjects ON courses.subject_id = subjects.id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("homeworks.deleted_at IS NULL")

	if req.CourseID > 0 {
		countQuery = countQuery.Where("courses.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		countQuery = countQuery.Where("subjects.id = ?", req.SubjectID)
	}
	if req.LessonID > 0 {
		countQuery = countQuery.Where("lessons.id = ?", req.LessonID)
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
		countQuery = countQuery.Where("user_courses.user_id = ?", userID)
	}

	if err := countQuery.Scan(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Query chính để lấy data - sử dụng DISTINCT ON để tránh duplicate
	query := db.ReplicaDB.Table("homeworks").
		Select("DISTINCT ON (homeworks.id) homeworks.id, homeworks.name, courses.id as course_id, subjects.id as subject_id, courses.name as course_name, subjects.name as subject_name, hrl.lesson_id, lessons.title as lesson_title, hrl.assigned_at").
		Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
		Joins("LEFT JOIN lessons ON hrl.lesson_id = lessons.id").
		Joins("LEFT JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("LEFT JOIN courses ON courses.program_id = chapters.program_id").
		Joins("LEFT JOIN subjects ON courses.subject_id = subjects.id").
		Joins("LEFT JOIN user_courses ON courses.id = user_courses.course_id").
		Where("homeworks.deleted_at IS NULL")

	if req.CourseID > 0 {
		query = query.Where("courses.id = ?", req.CourseID)
	}
	if req.SubjectID > 0 {
		query = query.Where("subjects.id = ?", req.SubjectID)
	}
	if req.LessonID > 0 {
		query = query.Where("lessons.id = ?", req.LessonID)
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
		query = query.Where("user_courses.user_id = ?", userID)
	}

	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		query = query.Offset(offset).Limit(req.Limit)
	}
	if len(req.Sort) > 0 {
		// Với DISTINCT ON, ORDER BY phải bắt đầu với homeworks.id
		query = query.Order("homeworks.id ASC")
		for field, order := range req.Sort {
			query = query.Order(field + " " + order)
		}
	} else {
		query = query.Order("homeworks.id ASC")
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
		countQuery = countQuery.Where("chapters.id = ?", req.ChapterID)
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
		query = query.Where("chapters.id = ?", req.ChapterID)
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
