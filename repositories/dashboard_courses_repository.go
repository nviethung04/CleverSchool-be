package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"time"

	"gorm.io/gorm"
)

type DashboardCoursesRepository interface {
	GetDashboardCourses(schoolID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error)
	GetDashboardCoursesWithPagination(schoolID int64, search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error)
	GetDashboardCoursesCount(schoolID int64, search string) (int64, error)
}

type dashboardCoursesRepository struct {
	db *gorm.DB
}

func NewDashboardCoursesRepository() DashboardCoursesRepository {
	return &dashboardCoursesRepository{
		db: db.ReplicaDB,
	}
}

// GetDashboardCourses lấy danh sách dashboard courses với điều kiện cố định
func (r *dashboardCoursesRepository) GetDashboardCourses(schoolID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error) {
	var courses []dto.DashboardCoursesResponse

	// Tính các khoảng thời gian
	sep8 := "2025-09-08"
	sep15 := "2025-09-15"
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Lấy tất cả course_id từ bảng courses
	var courseIDs []int64
	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	err := baseQuery.Select("c.id").Find(&courseIDs).Error
	if err != nil {
		return courses, err
	}

	// Query 1: Lấy dữ liệu từ 8/9 (active_teachers, total_teachers, teacher_ids, teacher_infos)
	var sep8DataList []struct {
		CourseID       int64                        `json:"course_id"`
		TotalTeachers  int64                        `json:"total_teachers"`
		ActiveTeachers int64                        `json:"active_teachers"`
		TeacherIDs     string                       `json:"teacher_ids"`
		TeacherInfos   models.DashboardTeacherInfos `json:"teacher_infos"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_teachers, active_teachers, teacher_ids, teacher_infos").
		Where("start_date = ?", sep8).
		Find(&sep8DataList)

	// Query 2: Lấy dữ liệu từ 15/9 (active_students, total_students, students_completed_homework)
	var sep15DataList []struct {
		CourseID                  int64 `json:"course_id"`
		TotalStudents             int64 `json:"total_students"`
		ActiveStudents            int64 `json:"active_students"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_students, active_students, students_completed_homework").
		Where("start_date = ?", sep15).
		Find(&sep15DataList)

	// Query 3: Lấy dữ liệu từ selected week (dựa trên start_date và end_date từ param)
	var selectedWeekDataList []struct {
		CourseID                  int64 `json:"course_id"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
		ActiveStudents            int64 `json:"active_students"`
		ActiveTeachers            int64 `json:"active_teachers"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, students_completed_homework, active_students, active_teachers").
		Where("start_date = ? AND end_date = ?", selectedStartDate, selectedEndDate).
		Find(&selectedWeekDataList)

	// Query 4: Lấy tên courses và thông tin school
	var courseSchoolInfo []struct {
		CourseID   int64  `json:"course_id"`
		CourseName string `json:"course_name"`
		ObjectTitle string `json:"object_title"`  // Thêm object_title
		SchoolID   int64  `json:"school_id"`
		SchoolName string `json:"school_name"`
	}
	r.db.Table("courses c").
		Select("c.id as course_id, c.name as course_name, c.object_title, s.id as school_id, s.name as school_name").
		Joins("LEFT JOIN course_schools cs ON c.id = cs.course_id").
		Joins("LEFT JOIN schools s ON cs.school_id = s.id").
		Find(&courseSchoolInfo)

	// Query 5: Lấy homework statistics từ bảng dashboard_report_courses (từ 15/9/2025 đến hôm qua)
	var homeworkStatsData []struct {
		CourseID           int64 `json:"course_id"`
		TotalHomeworks     int64 `json:"total_homeworks"`
		AssignedHomeworks  int64 `json:"assigned_homeworks"`
		CompletedHomeworks int64 `json:"completed_homeworks"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_homeworks, assigned_homeworks, completed_homeworks").
		Where("start_date = ? AND end_date = ?", sep15, yesterday).
		Find(&homeworkStatsData)

	// Tạo map để lookup nhanh
	sep8Map := make(map[int64]struct {
		TotalTeachers  int64
		ActiveTeachers int64
		TeacherIDs     string
		TeacherInfos   models.DashboardTeacherInfos
	})
	for _, data := range sep8DataList {
		sep8Map[data.CourseID] = struct {
			TotalTeachers  int64
			ActiveTeachers int64
			TeacherIDs     string
			TeacherInfos   models.DashboardTeacherInfos
		}{
			TotalTeachers:  data.TotalTeachers,
			ActiveTeachers: data.ActiveTeachers,
			TeacherIDs:     data.TeacherIDs,
			TeacherInfos:   data.TeacherInfos,
		}
	}

	sep15Map := make(map[int64]struct {
		TotalStudents             int64
		ActiveStudents            int64
		StudentsCompletedHomework int64
	})
	for _, data := range sep15DataList {
		sep15Map[data.CourseID] = struct {
			TotalStudents             int64
			ActiveStudents            int64
			StudentsCompletedHomework int64
		}{
			TotalStudents:             data.TotalStudents,
			ActiveStudents:            data.ActiveStudents,
			StudentsCompletedHomework: data.StudentsCompletedHomework,
		}
	}

	selectedWeekMap := make(map[int64]struct {
		StudentsCompletedHomework int64
		ActiveStudents            int64
		ActiveTeachers            int64
	})
	for _, data := range selectedWeekDataList {
		selectedWeekMap[data.CourseID] = struct {
			StudentsCompletedHomework int64
			ActiveStudents            int64
			ActiveTeachers            int64
		}{
			StudentsCompletedHomework: data.StudentsCompletedHomework,
			ActiveStudents:            data.ActiveStudents,
			ActiveTeachers:            data.ActiveTeachers,
		}
	}

	courseSchoolMap := make(map[int64]struct {
		CourseName  string
		ObjectTitle string
		SchoolID    int64
		SchoolName  string
	})
	for _, info := range courseSchoolInfo {
		courseSchoolMap[info.CourseID] = struct {
			CourseName  string
			ObjectTitle string
			SchoolID    int64
			SchoolName  string
		}{
			CourseName:  info.CourseName,
			ObjectTitle: info.ObjectTitle,
			SchoolID:    info.SchoolID,
			SchoolName:  info.SchoolName,
		}
	}

	// Tạo map cho homework statistics
	homeworkStatsMap := make(map[int64]struct {
		TotalHomeworks     int64
		AssignedHomeworks  int64
		CompletedHomeworks int64
	})
	for _, data := range homeworkStatsData {
		homeworkStatsMap[data.CourseID] = struct {
			TotalHomeworks     int64
			AssignedHomeworks  int64
			CompletedHomeworks int64
		}{
			TotalHomeworks:     data.TotalHomeworks,
			AssignedHomeworks:  data.AssignedHomeworks,
			CompletedHomeworks: data.CompletedHomeworks,
		}
	}

	// Gộp dữ liệu theo course_id
	for _, courseID := range courseIDs {
		course := dto.DashboardCoursesResponse{
			CourseID: courseID,
		}

		// Lấy thông tin course và school
		if courseSchoolInfo, exists := courseSchoolMap[courseID]; exists {
			course.CourseName = courseSchoolInfo.CourseName
			course.ObjectTitle = courseSchoolInfo.ObjectTitle
			course.SchoolID = courseSchoolInfo.SchoolID
			course.SchoolName = courseSchoolInfo.SchoolName
		}

		// Dữ liệu từ 8/9
		if sep8Data, exists := sep8Map[courseID]; exists {
			course.TotalTeachers = sep8Data.TotalTeachers
			course.ActiveTeachersFromSep8 = sep8Data.ActiveTeachers
			course.TeacherIDs = sep8Data.TeacherIDs
			course.TeacherInfos = sep8Data.TeacherInfos
		}

		// Dữ liệu từ 15/9
		if sep15Data, exists := sep15Map[courseID]; exists {
			course.TotalStudents = sep15Data.TotalStudents
			course.ActiveStudentsFromSep15 = sep15Data.ActiveStudents
			course.StudentsCompletedHomeworkFromSep15 = sep15Data.StudentsCompletedHomework
		}

		// Dữ liệu từ selected week
		if selectedWeekData, exists := selectedWeekMap[courseID]; exists {
			course.StudentsCompletedHomeworkSelectedWeek = selectedWeekData.StudentsCompletedHomework
			course.ActiveStudentsSelectedWeek = selectedWeekData.ActiveStudents
			course.ActiveTeachersSelectedWeek = selectedWeekData.ActiveTeachers
		} else {
			// Nếu không có dữ liệu selected week, set -1
			course.StudentsCompletedHomeworkSelectedWeek = -1
			course.ActiveStudentsSelectedWeek = -1
			course.ActiveTeachersSelectedWeek = -1
		}

		// Homework statistics từ bảng dashboard_report_courses (15/9/2025 đến hôm qua)
		if homeworkStats, exists := homeworkStatsMap[courseID]; exists {
			course.TotalHomeworks = homeworkStats.TotalHomeworks
			course.AssignedHomeworks = homeworkStats.AssignedHomeworks
			course.CompletedHomeworks = homeworkStats.CompletedHomeworks
		}

		courses = append(courses, course)
	}

	return courses, nil
}

// GetDashboardCoursesWithPagination lấy danh sách dashboard courses với phân trang và tìm kiếm
func (r *dashboardCoursesRepository) GetDashboardCoursesWithPagination(schoolID int64, search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error) {
	var courses []dto.DashboardCoursesResponse

	// Tính các khoảng thời gian
	sep8 := "2025-09-08"
	sep15 := "2025-09-15"
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Lấy tất cả course_id từ bảng courses với filter
	var courseIDs []int64
	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	// Search filter
	if search != "" {
		baseQuery = baseQuery.Joins("LEFT JOIN course_schools cs_search ON c.id = cs_search.course_id").
			Joins("LEFT JOIN schools s_search ON cs_search.school_id = s_search.id").
			Where("(c.name ILIKE ? OR s_search.name ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	err := baseQuery.Select("c.id").Find(&courseIDs).Error
	if err != nil {
		return courses, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if len(courseIDs) > offset {
		if offset+limit > len(courseIDs) {
			courseIDs = courseIDs[offset:]
		} else {
			courseIDs = courseIDs[offset : offset+limit]
		}
	} else {
		return courses, nil
	}

	// Query 1: Lấy dữ liệu từ 8/9 (active_teachers, total_teachers, teacher_ids)
	var sep8DataList []struct {
		CourseID       int64                        `json:"course_id"`
		TotalTeachers  int64                        `json:"total_teachers"`
		ActiveTeachers int64                        `json:"active_teachers"`
		TeacherIDs     string                       `json:"teacher_ids"`
		TeacherInfos   models.DashboardTeacherInfos `json:"teacher_infos"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_teachers, active_teachers, teacher_ids, teacher_infos").
		Where("start_date = ? AND course_id IN ?", sep8, courseIDs).
		Find(&sep8DataList)

	// Query 2: Lấy dữ liệu từ 15/9 (active_students, total_students, students_completed_homework)
	var sep15DataList []struct {
		CourseID                  int64 `json:"course_id"`
		TotalStudents             int64 `json:"total_students"`
		ActiveStudents            int64 `json:"active_students"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_students, active_students, students_completed_homework").
		Where("start_date = ? AND course_id IN ?", sep15, courseIDs).
		Find(&sep15DataList)

	// Query 3: Lấy dữ liệu từ selected week (dựa trên start_date và end_date từ param)
	var selectedWeekDataList []struct {
		CourseID                  int64 `json:"course_id"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
		ActiveStudents            int64 `json:"active_students"`
		ActiveTeachers            int64 `json:"active_teachers"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, students_completed_homework, active_students, active_teachers").
		Where("start_date = ? AND end_date = ? AND course_id IN ?", selectedStartDate, selectedEndDate, courseIDs).
		Find(&selectedWeekDataList)

	// Query 4: Lấy tên courses và thông tin school
	var courseSchoolInfo []struct {
		CourseID   int64  `json:"course_id"`
		CourseName string `json:"course_name"`
		ObjectTitle string `json:"object_title"`  // Thêm object_title
		SchoolID   int64  `json:"school_id"`
		SchoolName string `json:"school_name"`
	}
	r.db.Table("courses c").
		Select("c.id as course_id, c.name as course_name, c.object_title, s.id as school_id, s.name as school_name").
		Joins("LEFT JOIN course_schools cs ON c.id = cs.course_id").
		Joins("LEFT JOIN schools s ON cs.school_id = s.id").
		Where("c.id IN ?", courseIDs).
		Find(&courseSchoolInfo)

	// Query 5: Lấy homework statistics từ bảng dashboard_report_courses (từ 15/9/2025 đến hôm qua)
	var homeworkStatsData []struct {
		CourseID           int64 `json:"course_id"`
		TotalHomeworks     int64 `json:"total_homeworks"`
		AssignedHomeworks  int64 `json:"assigned_homeworks"`
		CompletedHomeworks int64 `json:"completed_homeworks"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_homeworks, assigned_homeworks, completed_homeworks").
		Where("start_date = ? AND end_date = ? AND course_id IN ?", sep15, yesterday, courseIDs).
		Find(&homeworkStatsData)

	// Tạo map để lookup nhanh
	sep8Map := make(map[int64]struct {
		TotalTeachers  int64
		ActiveTeachers int64
		TeacherIDs     string
		TeacherInfos   models.DashboardTeacherInfos
	})
	for _, data := range sep8DataList {
		sep8Map[data.CourseID] = struct {
			TotalTeachers  int64
			ActiveTeachers int64
			TeacherIDs     string
			TeacherInfos   models.DashboardTeacherInfos
		}{
			TotalTeachers:  data.TotalTeachers,
			ActiveTeachers: data.ActiveTeachers,
			TeacherIDs:     data.TeacherIDs,
			TeacherInfos:   data.TeacherInfos,
		}
	}

	sep15Map := make(map[int64]struct {
		TotalStudents             int64
		ActiveStudents            int64
		StudentsCompletedHomework int64
	})
	for _, data := range sep15DataList {
		sep15Map[data.CourseID] = struct {
			TotalStudents             int64
			ActiveStudents            int64
			StudentsCompletedHomework int64
		}{
			TotalStudents:             data.TotalStudents,
			ActiveStudents:            data.ActiveStudents,
			StudentsCompletedHomework: data.StudentsCompletedHomework,
		}
	}

	selectedWeekMap := make(map[int64]struct {
		StudentsCompletedHomework int64
		ActiveStudents            int64
		ActiveTeachers            int64
	})
	for _, data := range selectedWeekDataList {
		selectedWeekMap[data.CourseID] = struct {
			StudentsCompletedHomework int64
			ActiveStudents            int64
			ActiveTeachers            int64
		}{
			StudentsCompletedHomework: data.StudentsCompletedHomework,
			ActiveStudents:            data.ActiveStudents,
			ActiveTeachers:            data.ActiveTeachers,
		}
	}

	courseSchoolMap := make(map[int64]struct {
		CourseName  string
		ObjectTitle string
		SchoolID    int64
		SchoolName  string
	})
	for _, info := range courseSchoolInfo {
		courseSchoolMap[info.CourseID] = struct {
			CourseName  string
			ObjectTitle string
			SchoolID    int64
			SchoolName  string
		}{
			CourseName:  info.CourseName,
			ObjectTitle: info.ObjectTitle,
			SchoolID:    info.SchoolID,
			SchoolName:  info.SchoolName,
		}
	}

	// Tạo map cho homework statistics
	homeworkStatsMap := make(map[int64]struct {
		TotalHomeworks     int64
		AssignedHomeworks  int64
		CompletedHomeworks int64
	})
	for _, data := range homeworkStatsData {
		homeworkStatsMap[data.CourseID] = struct {
			TotalHomeworks     int64
			AssignedHomeworks  int64
			CompletedHomeworks int64
		}{
			TotalHomeworks:     data.TotalHomeworks,
			AssignedHomeworks:  data.AssignedHomeworks,
			CompletedHomeworks: data.CompletedHomeworks,
		}
	}

	// Gộp dữ liệu theo course_id
	for _, courseID := range courseIDs {
		course := dto.DashboardCoursesResponse{
			CourseID: courseID,
		}

		// Lấy thông tin course và school
		if courseSchoolInfo, exists := courseSchoolMap[courseID]; exists {
			course.CourseName = courseSchoolInfo.CourseName
			course.ObjectTitle = courseSchoolInfo.ObjectTitle
			course.SchoolID = courseSchoolInfo.SchoolID
			course.SchoolName = courseSchoolInfo.SchoolName
		}

		// Dữ liệu từ 8/9
		if sep8Data, exists := sep8Map[courseID]; exists {
			course.TotalTeachers = sep8Data.TotalTeachers
			course.ActiveTeachersFromSep8 = sep8Data.ActiveTeachers
			course.TeacherIDs = sep8Data.TeacherIDs
			course.TeacherInfos = sep8Data.TeacherInfos
		}

		// Dữ liệu từ 15/9
		if sep15Data, exists := sep15Map[courseID]; exists {
			course.TotalStudents = sep15Data.TotalStudents
			course.ActiveStudentsFromSep15 = sep15Data.ActiveStudents
			course.StudentsCompletedHomeworkFromSep15 = sep15Data.StudentsCompletedHomework
		}

		// Dữ liệu từ selected week
		if selectedWeekData, exists := selectedWeekMap[courseID]; exists {
			course.StudentsCompletedHomeworkSelectedWeek = selectedWeekData.StudentsCompletedHomework
			course.ActiveStudentsSelectedWeek = selectedWeekData.ActiveStudents
			course.ActiveTeachersSelectedWeek = selectedWeekData.ActiveTeachers
		} else {
			// Nếu không có dữ liệu selected week, set -1
			course.StudentsCompletedHomeworkSelectedWeek = -1
			course.ActiveStudentsSelectedWeek = -1
			course.ActiveTeachersSelectedWeek = -1
		}

		// Homework statistics từ bảng dashboard_report_courses (15/9/2025 đến hôm qua)
		if homeworkStats, exists := homeworkStatsMap[courseID]; exists {
			course.TotalHomeworks = homeworkStats.TotalHomeworks
			course.AssignedHomeworks = homeworkStats.AssignedHomeworks
			course.CompletedHomeworks = homeworkStats.CompletedHomeworks
		}

		courses = append(courses, course)
	}

	return courses, nil
}

// GetDashboardCoursesCount đếm tổng số courses
func (r *dashboardCoursesRepository) GetDashboardCoursesCount(schoolID int64, search string) (int64, error) {
	var count int64

	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	// Search filter
	if search != "" {
		baseQuery = baseQuery.Joins("LEFT JOIN course_schools cs_search ON c.id = cs_search.course_id").
			Joins("LEFT JOIN schools s_search ON cs_search.school_id = s_search.id").
			Where("(c.name ILIKE ? OR s_search.name ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	err := baseQuery.Count(&count).Error
	return count, err
}
