package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/table_manager"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DashboardCoursesRepository interface {
	GetDashboardCourses(schoolID, teacherID, courseID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error)
	GetDashboardCoursesWithPagination(schoolID, teacherID, courseID int64, search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error)
	GetDashboardCoursesCount(schoolID, teacherID, courseID int64, search string) (int64, error)
	GetCourseStudents(courseID int64, startDate, endDate string) (*dto.DashboardCourseStudentsResponse, error)
	GetCourseHomeworks(courseID int64, startDate, endDate string) (*dto.DashboardCourseHomeworksResponse, error)
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
func (r *dashboardCoursesRepository) GetDashboardCourses(schoolID, teacherID, courseID int64, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error) {
	var courses []dto.DashboardCoursesResponse

	// Tính các khoảng thời gian
	sep8 := "2025-09-08"
	// Lấy tất cả course_id từ bảng courses
	var courseIDs []int64
	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	// Filter by course_id if provided
	if courseID > 0 {
		baseQuery = baseQuery.Where("c.id = ?", courseID)
	}

	err := baseQuery.Select("c.id").Find(&courseIDs).Error
	if err != nil {
		return courses, err
	}

	// Filter by teacher_id if provided (check in teacher_ids column from dashboard_report_courses)
	if teacherID > 0 {
		var filteredCourseIDs []int64
		teacherIDStr := fmt.Sprintf("%d", teacherID)
		r.db.Table("dashboard_report_courses").
			Select("DISTINCT course_id").
			Where("start_date = ?", sep8).
			Where("(teacher_ids = ? OR teacher_ids LIKE ? OR teacher_ids LIKE ? OR teacher_ids LIKE ?)",
				teacherIDStr,
				teacherIDStr+",%",
				"%,"+teacherIDStr+",%",
				"%,"+teacherIDStr).
			Find(&filteredCourseIDs)

		// Intersect với courseIDs đã filter trước đó
		filteredMap := make(map[int64]bool)
		for _, id := range filteredCourseIDs {
			filteredMap[id] = true
		}

		var newCourseIDs []int64
		for _, id := range courseIDs {
			if filteredMap[id] {
				newCourseIDs = append(newCourseIDs, id)
			}
		}
		courseIDs = newCourseIDs
	}

	// Query 1: Lấy dữ liệu từ selected week (dựa trên start_date và end_date từ param)
	var selectedWeekDataList []struct {
		CourseID                      int64                        `json:"course_id"`
		TotalStudents                 int64                        `json:"total_students"`
		TotalTeachers                 int64                        `json:"total_teachers"`
		StudentsCompletedHomework     int64                        `json:"students_completed_homework"`
		ActiveStudents                int64                        `json:"active_students"`
		ActiveTeachers                int64                        `json:"active_teachers"`
		TeacherIDs                    string                       `json:"teacher_ids"`
		TeacherInfos                  models.DashboardTeacherInfos `json:"teacher_infos"`
		TotalHomeworks                int64                        `json:"total_homeworks"`
		AssignedHomeworks             int64                        `json:"assigned_homeworks"`
		StudentsCompletedAllHomeworks int64                        `json:"students_completed_all_homeworks"`
		StudentsDoingHomeworks        int64                        `json:"students_doing_homeworks"`
		StudentsNotStartedAnyHomework int64                        `json:"students_not_started_any_homework"`
		StudentActiveNotStartedAnyHomework int64                   `json:"student_active_not_started_any_homework"`
		HomeworkOver50PercentStudentComplete int64                 `json:"homework_over_50_percent_student_complete"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_students, total_teachers, students_completed_homework, active_students, active_teachers, teacher_ids, teacher_infos, total_homeworks, assigned_homeworks, students_completed_all_homeworks, students_doing_homeworks, students_not_started_any_homework, student_active_not_started_any_homework, homework_over_50_percent_student_complete").
		Where("start_date = ? AND end_date = ?", selectedStartDate, selectedEndDate).
		Find(&selectedWeekDataList)

	// Query 2: Lấy tên courses và thông tin school
	var courseSchoolInfo []struct {
		CourseID    int64  `json:"course_id"`
		CourseName  string `json:"course_name"`
		ObjectTitle string `json:"object_title"` // Thêm object_title
		SchoolID    int64  `json:"school_id"`
		SchoolName  string `json:"school_name"`
	}
	r.db.Table("courses c").
		Select("c.id as course_id, c.name as course_name, c.object_title, s.id as school_id, s.name as school_name").
		Joins("LEFT JOIN course_schools cs ON c.id = cs.course_id").
		Joins("LEFT JOIN schools s ON cs.school_id = s.id").
		Find(&courseSchoolInfo)

	// Tạo map để lookup nhanh
	selectedWeekMap := make(map[int64]struct {
		TotalStudents                 int64
		TotalTeachers                 int64
		StudentsCompletedHomework     int64
		ActiveStudents                int64
		ActiveTeachers                int64
		TeacherIDs                    string
		TeacherInfos                  models.DashboardTeacherInfos
		TotalHomeworks                int64
		AssignedHomeworks             int64
		StudentsCompletedAllHomeworks int64
		StudentsDoingHomeworks        int64
		StudentsNotStartedAnyHomework int64
		StudentActiveNotStartedAnyHomework int64
		HomeworkOver50PercentStudentComplete int64
	})
	for _, data := range selectedWeekDataList {
		selectedWeekMap[data.CourseID] = struct {
			TotalStudents                 int64
			TotalTeachers                 int64
			StudentsCompletedHomework     int64
			ActiveStudents                int64
			ActiveTeachers                int64
			TeacherIDs                    string
			TeacherInfos                  models.DashboardTeacherInfos
			TotalHomeworks                int64
			AssignedHomeworks             int64
			StudentsCompletedAllHomeworks int64
			StudentsDoingHomeworks        int64
			StudentsNotStartedAnyHomework int64
			StudentActiveNotStartedAnyHomework int64
			HomeworkOver50PercentStudentComplete int64
		}{
			TotalStudents:                 data.TotalStudents,
			TotalTeachers:                 data.TotalTeachers,
			StudentsCompletedHomework:     data.StudentsCompletedHomework,
			ActiveStudents:                data.ActiveStudents,
			ActiveTeachers:                data.ActiveTeachers,
			TeacherIDs:                    data.TeacherIDs,
			TeacherInfos:                  data.TeacherInfos,
			TotalHomeworks:                data.TotalHomeworks,
			AssignedHomeworks:             data.AssignedHomeworks,
			StudentsCompletedAllHomeworks: data.StudentsCompletedAllHomeworks,
			StudentsDoingHomeworks:        data.StudentsDoingHomeworks,
			StudentsNotStartedAnyHomework: data.StudentsNotStartedAnyHomework,
			StudentActiveNotStartedAnyHomework: data.StudentActiveNotStartedAnyHomework,
			HomeworkOver50PercentStudentComplete: data.HomeworkOver50PercentStudentComplete,
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

		// Dữ liệu từ selected week (tất cả các trường đều lấy theo start_date và end_date)
		if selectedWeekData, exists := selectedWeekMap[courseID]; exists {
			course.TotalStudents = selectedWeekData.TotalStudents
			course.TotalTeachers = selectedWeekData.TotalTeachers
			course.TeacherIDs = selectedWeekData.TeacherIDs
			course.TeacherInfos = selectedWeekData.TeacherInfos
			course.StudentsCompletedHomeworkSelectedWeek = selectedWeekData.StudentsCompletedHomework
			course.ActiveStudentsSelectedWeek = selectedWeekData.ActiveStudents
			course.ActiveTeachersSelectedWeek = selectedWeekData.ActiveTeachers
			course.TotalHomeworks = selectedWeekData.TotalHomeworks
			course.AssignedHomeworks = selectedWeekData.AssignedHomeworks
			course.StudentsCompletedAllHomeworks = selectedWeekData.StudentsCompletedAllHomeworks
			course.StudentsDoingHomeworks = selectedWeekData.StudentsDoingHomeworks
			course.StudentsNotStartedAnyHomework = selectedWeekData.StudentsNotStartedAnyHomework
			course.StudentActiveNotStartedAnyHomework = selectedWeekData.StudentActiveNotStartedAnyHomework
			course.HomeworkOver50PercentStudentComplete = selectedWeekData.HomeworkOver50PercentStudentComplete
		} else {
			// Nếu không có dữ liệu selected week, set -1
			course.StudentsCompletedHomeworkSelectedWeek = -1
			course.ActiveStudentsSelectedWeek = -1
			course.ActiveTeachersSelectedWeek = -1
		}

		courses = append(courses, course)
	}

	return courses, nil
}

// GetDashboardCoursesWithPagination lấy danh sách dashboard courses với phân trang và tìm kiếm
func (r *dashboardCoursesRepository) GetDashboardCoursesWithPagination(schoolID, teacherID, courseID int64, search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardCoursesResponse, error) {
	var courses []dto.DashboardCoursesResponse

	// Tính các khoảng thời gian
	sep8 := "2025-09-08"
	// Lấy tất cả course_id từ bảng courses với filter
	var courseIDs []int64
	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	// Filter by course_id if provided
	if courseID > 0 {
		baseQuery = baseQuery.Where("c.id = ?", courseID)
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

	// Filter by teacher_id if provided (check in teacher_ids column from dashboard_report_courses)
	if teacherID > 0 {
		var filteredCourseIDs []int64
		teacherIDStr := fmt.Sprintf("%d", teacherID)
		r.db.Table("dashboard_report_courses").
			Select("DISTINCT course_id").
			Where("start_date = ?", sep8).
			Where("(teacher_ids = ? OR teacher_ids LIKE ? OR teacher_ids LIKE ? OR teacher_ids LIKE ?)",
				teacherIDStr,
				teacherIDStr+",%",
				"%,"+teacherIDStr+",%",
				"%,"+teacherIDStr).
			Find(&filteredCourseIDs)

		// Intersect với courseIDs đã filter trước đó
		filteredMap := make(map[int64]bool)
		for _, id := range filteredCourseIDs {
			filteredMap[id] = true
		}

		var newCourseIDs []int64
		for _, id := range courseIDs {
			if filteredMap[id] {
				newCourseIDs = append(newCourseIDs, id)
			}
		}
		courseIDs = newCourseIDs
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

	// Query 1: Lấy dữ liệu từ selected week (dựa trên start_date và end_date từ param)
	var selectedWeekDataList []struct {
		CourseID                      int64                        `json:"course_id"`
		TotalStudents                 int64                        `json:"total_students"`
		TotalTeachers                 int64                        `json:"total_teachers"`
		StudentsCompletedHomework     int64                        `json:"students_completed_homework"`
		ActiveStudents                int64                        `json:"active_students"`
		ActiveTeachers                int64                        `json:"active_teachers"`
		TeacherIDs                    string                       `json:"teacher_ids"`
		TeacherInfos                  models.DashboardTeacherInfos `json:"teacher_infos"`
		TotalHomeworks                int64                        `json:"total_homeworks"`
		AssignedHomeworks             int64                        `json:"assigned_homeworks"`
		StudentsCompletedAllHomeworks int64                        `json:"students_completed_all_homeworks"`
		StudentsDoingHomeworks        int64                        `json:"students_doing_homeworks"`
		StudentsNotStartedAnyHomework int64                        `json:"students_not_started_any_homework"`
		StudentActiveNotStartedAnyHomework int64                   `json:"student_active_not_started_any_homework"`
		HomeworkOver50PercentStudentComplete int64                 `json:"homework_over_50_percent_student_complete"`
	}
	r.db.Table("dashboard_report_courses").
		Select("course_id, total_students, total_teachers, students_completed_homework, active_students, active_teachers, teacher_ids, teacher_infos, total_homeworks, assigned_homeworks, students_completed_all_homeworks, students_doing_homeworks, students_not_started_any_homework, student_active_not_started_any_homework, homework_over_50_percent_student_complete").
		Where("start_date = ? AND end_date = ? AND course_id IN ?", selectedStartDate, selectedEndDate, courseIDs).
		Find(&selectedWeekDataList)

	// Query 2: Lấy tên courses và thông tin school
	var courseSchoolInfo []struct {
		CourseID    int64  `json:"course_id"`
		CourseName  string `json:"course_name"`
		ObjectTitle string `json:"object_title"` // Thêm object_title
		SchoolID    int64  `json:"school_id"`
		SchoolName  string `json:"school_name"`
	}
	r.db.Table("courses c").
		Select("c.id as course_id, c.name as course_name, c.object_title, s.id as school_id, s.name as school_name").
		Joins("LEFT JOIN course_schools cs ON c.id = cs.course_id").
		Joins("LEFT JOIN schools s ON cs.school_id = s.id").
		Where("c.id IN ?", courseIDs).
		Find(&courseSchoolInfo)

	// Query 5: Bỏ query này vì đã lấy trong selectedWeekDataList

	// Tạo map để lookup nhanh
	selectedWeekMap := make(map[int64]struct {
		TotalStudents                 int64
		TotalTeachers                 int64
		StudentsCompletedHomework     int64
		ActiveStudents                int64
		ActiveTeachers                int64
		TeacherIDs                    string
		TeacherInfos                  models.DashboardTeacherInfos
		TotalHomeworks                int64
		AssignedHomeworks             int64
		StudentsCompletedAllHomeworks int64
		StudentsDoingHomeworks        int64
		StudentsNotStartedAnyHomework int64
		StudentActiveNotStartedAnyHomework int64
		HomeworkOver50PercentStudentComplete int64
	})
	for _, data := range selectedWeekDataList {
		selectedWeekMap[data.CourseID] = struct {
			TotalStudents                 int64
			TotalTeachers                 int64
			StudentsCompletedHomework     int64
			ActiveStudents                int64
			ActiveTeachers                int64
			TeacherIDs                    string
			TeacherInfos                  models.DashboardTeacherInfos
			TotalHomeworks                int64
			AssignedHomeworks             int64
			StudentsCompletedAllHomeworks int64
			StudentsDoingHomeworks        int64
			StudentsNotStartedAnyHomework int64
			StudentActiveNotStartedAnyHomework int64
			HomeworkOver50PercentStudentComplete int64
		}{
			TotalStudents:                 data.TotalStudents,
			TotalTeachers:                 data.TotalTeachers,
			StudentsCompletedHomework:     data.StudentsCompletedHomework,
			ActiveStudents:                data.ActiveStudents,
			ActiveTeachers:                data.ActiveTeachers,
			TeacherIDs:                    data.TeacherIDs,
			TeacherInfos:                  data.TeacherInfos,
			TotalHomeworks:                data.TotalHomeworks,
			AssignedHomeworks:             data.AssignedHomeworks,
			StudentsCompletedAllHomeworks: data.StudentsCompletedAllHomeworks,
			StudentsDoingHomeworks:        data.StudentsDoingHomeworks,
			StudentsNotStartedAnyHomework: data.StudentsNotStartedAnyHomework,
			StudentActiveNotStartedAnyHomework: data.StudentActiveNotStartedAnyHomework,
			HomeworkOver50PercentStudentComplete: data.HomeworkOver50PercentStudentComplete,
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

		// Dữ liệu từ selected week (tất cả các trường đều lấy theo start_date và end_date)
		if selectedWeekData, exists := selectedWeekMap[courseID]; exists {
			course.TotalStudents = selectedWeekData.TotalStudents
			course.TotalTeachers = selectedWeekData.TotalTeachers
			course.TeacherIDs = selectedWeekData.TeacherIDs
			course.TeacherInfos = selectedWeekData.TeacherInfos
			course.StudentsCompletedHomeworkSelectedWeek = selectedWeekData.StudentsCompletedHomework
			course.ActiveStudentsSelectedWeek = selectedWeekData.ActiveStudents
			course.ActiveTeachersSelectedWeek = selectedWeekData.ActiveTeachers
			course.TotalHomeworks = selectedWeekData.TotalHomeworks
			course.AssignedHomeworks = selectedWeekData.AssignedHomeworks
			course.StudentsCompletedAllHomeworks = selectedWeekData.StudentsCompletedAllHomeworks
			course.StudentsDoingHomeworks = selectedWeekData.StudentsDoingHomeworks
			course.StudentsNotStartedAnyHomework = selectedWeekData.StudentsNotStartedAnyHomework
			course.StudentActiveNotStartedAnyHomework = selectedWeekData.StudentActiveNotStartedAnyHomework
			course.HomeworkOver50PercentStudentComplete = selectedWeekData.HomeworkOver50PercentStudentComplete
		} else {
			// Nếu không có dữ liệu selected week, set -1
			course.StudentsCompletedHomeworkSelectedWeek = -1
			course.ActiveStudentsSelectedWeek = -1
			course.ActiveTeachersSelectedWeek = -1
		}

		courses = append(courses, course)
	}

	return courses, nil
}

// GetDashboardCoursesCount đếm tổng số courses
func (r *dashboardCoursesRepository) GetDashboardCoursesCount(schoolID, teacherID, courseID int64, search string) (int64, error) {
	sep8 := "2025-09-08"

	// Lấy tất cả course_id từ bảng courses với filter
	var courseIDs []int64
	baseQuery := r.db.Table("courses c")

	// Filter by school_id if provided
	if schoolID > 0 {
		baseQuery = baseQuery.Joins("INNER JOIN course_schools cs ON c.id = cs.course_id").
			Where("cs.school_id = ?", schoolID)
	}

	// Filter by course_id if provided
	if courseID > 0 {
		baseQuery = baseQuery.Where("c.id = ?", courseID)
	}

	// Search filter
	if search != "" {
		baseQuery = baseQuery.Joins("LEFT JOIN course_schools cs_search ON c.id = cs_search.course_id").
			Joins("LEFT JOIN schools s_search ON cs_search.school_id = s_search.id").
			Where("(c.name ILIKE ? OR s_search.name ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	err := baseQuery.Select("c.id").Find(&courseIDs).Error
	if err != nil {
		return 0, err
	}

	// Filter by teacher_id if provided (check in teacher_ids column from dashboard_report_courses)
	if teacherID > 0 {
		var filteredCourseIDs []int64
		teacherIDStr := fmt.Sprintf("%d", teacherID)
		r.db.Table("dashboard_report_courses").
			Select("DISTINCT course_id").
			Where("start_date = ?", sep8).
			Where("(teacher_ids = ? OR teacher_ids LIKE ? OR teacher_ids LIKE ? OR teacher_ids LIKE ?)",
				teacherIDStr,
				teacherIDStr+",%",
				"%,"+teacherIDStr+",%",
				"%,"+teacherIDStr).
			Find(&filteredCourseIDs)

		// Intersect với courseIDs đã filter trước đó
		filteredMap := make(map[int64]bool)
		for _, id := range filteredCourseIDs {
			filteredMap[id] = true
		}

		var newCourseIDs []int64
		for _, id := range courseIDs {
			if filteredMap[id] {
				newCourseIDs = append(newCourseIDs, id)
			}
		}
		courseIDs = newCourseIDs
	}

	return int64(len(courseIDs)), nil
}

// createActivityLogsUnionQuery tạo UNION query cho các bảng activity_logs theo tháng
func (r *dashboardCoursesRepository) createActivityLogsUnionQuery(startTime, endTime time.Time) (string, error) {
	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", startTime, endTime)

	// Tạo UNION ALL cho các bảng tháng
	var unionParts []string
	for _, tableName := range tableNames {
		// Kiểm tra bảng có tồn tại không
		var exists bool
		err := db.ReplicaDB.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = ?
			)
		`, tableName).Scan(&exists).Error

		if err != nil || !exists {
			continue
		}

		unionParts = append(unionParts, fmt.Sprintf("SELECT * FROM %s", tableName))
	}

	// Nếu không có bảng nào tồn tại, trả về query rỗng
	if len(unionParts) == 0 {
		return "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs", nil
	}

	// Tạo virtual table từ UNION các bảng tháng
	return fmt.Sprintf("(%s) AS all_activity_logs", strings.Join(unionParts, " UNION ALL ")), nil
}

// createAllActivityLogsUnionQuery tạo UNION query cho TẤT CẢ các bảng activity_logs (không giới hạn theo thời gian)
func (r *dashboardCoursesRepository) createAllActivityLogsUnionQuery() (string, error) {
	// Lấy tất cả các bảng activity_logs từ database
	var tableNames []string
	result := db.ReplicaDB.Raw(`
		SELECT table_name
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE 'activity_logs_%'
		ORDER BY table_name
	`).Pluck("table_name", &tableNames)

	if result.Error != nil {
		return "", result.Error
	}

	// Tạo UNION ALL cho các bảng
	var unionParts []string
	for _, tableName := range tableNames {
		unionParts = append(unionParts, fmt.Sprintf("SELECT * FROM %s", tableName))
	}

	// Nếu không có bảng nào tồn tại, trả về query rỗng
	if len(unionParts) == 0 {
		return "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs", nil
	}

	// Tạo virtual table từ UNION các bảng
	return fmt.Sprintf("(%s) AS all_activity_logs", strings.Join(unionParts, " UNION ALL ")), nil
}

// GetCourseStudents lấy danh sách học sinh trong course kèm thông tin homework
func (r *dashboardCoursesRepository) GetCourseStudents(courseID int64, startDate, endDate string) (*dto.DashboardCourseStudentsResponse, error) {
	// Parse dates
	var startTime, endTime time.Time
	var err error
	
	// Thử parse Unix timestamp
	if startTimestamp, err := strconv.ParseInt(startDate, 10, 64); err == nil {
		startTime = time.Unix(startTimestamp, 0)
	} else {
		// Parse YYYY-MM-DD format
		startTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %v", err)
		}
	}
	
	if endTimestamp, err := strconv.ParseInt(endDate, 10, 64); err == nil {
		endTime = time.Unix(endTimestamp, 0)
	} else {
		// Parse YYYY-MM-DD format
		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %v", err)
		}
	}
	
	// Tạo endTime đến cuối ngày (23:59:59)
	endOfDay := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())
	
	// Tạo UNION query cho TẤT CẢ activity_logs (không giới hạn theo thời gian)
	allActivityLogs, err := r.createAllActivityLogsUnionQuery()
	if err != nil {
		return nil, err
	}
	
	// Bước 1: Lấy danh sách học sinh trong course (role_id = 3)
	var students []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	err = r.db.Table("users u").
		Select("u.id, u.name").
		Joins("INNER JOIN user_courses uc ON u.id = uc.user_id").
		Joins("INNER JOIN user_ref_roles urr ON u.id = urr.user_id").
		Where("uc.course_id = ?", courseID).
		Where("u.deleted_at IS NULL").
		Where("urr.role_id = 3").
		Find(&students).Error
	if err != nil {
		return nil, err
	}
	
	// Bước 2: Lấy danh sách học sinh active (có trong activity_logs) - không giới hạn theo thời gian
	var activeStudentIDs []int64
	if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
		activeQuery := fmt.Sprintf(`
			SELECT DISTINCT u.id
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			INNER JOIN %s ON u.id = all_activity_logs.user_id
			WHERE uc.course_id = ?
			AND u.deleted_at IS NULL
			AND urr.role_id = 3
			AND all_activity_logs.role_id = 3
		`, allActivityLogs)
		r.db.Raw(activeQuery, courseID).Pluck("id", &activeStudentIDs)
	}
	
	// Tạo map để check active nhanh
	activeMap := make(map[int64]bool)
	for _, id := range activeStudentIDs {
		activeMap[id] = true
	}
	
	// Bước 3: Lấy danh sách homework trong khoảng thời gian
	var homeworks []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	err = r.db.Table("homeworks h").
		Select("DISTINCT h.id, h.name").
		Joins("INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id").
		Joins("INNER JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("INNER JOIN weeks w ON ls.week_id = w.id").
		Where("hrl.course_id = ?", courseID).
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_at IS NOT NULL").
		Where("? <= w.start_date AND ? >= w.end_date", startTime, endTime).
		Find(&homeworks).Error
	if err != nil {
		return nil, err
	}
	
	// Bước 4: Lấy trạng thái hoàn thành của từng homework cho từng học sinh
	homeworkCompletionMap := make(map[int64]map[int64]bool) // map[userID][homeworkID] = isCompleted
	homeworkLateMap := make(map[int64]map[int64]bool)       // map[userID][homeworkID] = isCompletedLate
	
	for _, student := range students {
		homeworkCompletionMap[student.ID] = make(map[int64]bool)
		homeworkLateMap[student.ID] = make(map[int64]bool)
		for _, homework := range homeworks {
			homeworkCompletionMap[student.ID][homework.ID] = false
			homeworkLateMap[student.ID][homework.ID] = false
		}
	}
	
	// Query để lấy các homework đã hoàn thành (bỏ điều kiện thời gian để lấy tất cả)
	var completedHomeworks []struct {
		UserID     int64     `json:"user_id"`
		HomeworkID int64     `json:"homework_id"`
		UpdatedAt  time.Time `json:"updated_at"`
	}
	err = r.db.Table("homework_users hu").
		Select("hu.user_id, h.id as homework_id, hu.updated_at").
		Joins("INNER JOIN homeworks h ON hu.homework_id = h.id").
		Joins("INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id AND hu.lesson_id = hrl.lesson_id").
		Where("hrl.course_id = ?", courseID).
		Where("hu.questions_completed >= h.total_questions").
		Find(&completedHomeworks).Error
	if err != nil {
		return nil, err
	}
	
	// Cập nhật trạng thái hoàn thành và kiểm tra muộn
	for _, completed := range completedHomeworks {
		if studentMap, exists := homeworkCompletionMap[completed.UserID]; exists {
			if _, homeworkExists := studentMap[completed.HomeworkID]; homeworkExists {
				studentMap[completed.HomeworkID] = true
				// Kiểm tra nếu hoàn thành sau endOfDay thì đánh dấu là muộn (tính theo updated_at)
				if completed.UpdatedAt.After(endOfDay) {
					if lateMap, lateExists := homeworkLateMap[completed.UserID]; lateExists {
						if _, lateHomeworkExists := lateMap[completed.HomeworkID]; lateHomeworkExists {
							lateMap[completed.HomeworkID] = true
						}
					}
				}
			}
		}
	}
	
	// Bước 5: Gộp dữ liệu thành response
	var result dto.DashboardCourseStudentsResponse
	for _, student := range students {
		studentDTO := dto.DashboardCourseStudent{
			ID:       student.ID,
			Name:     student.Name,
			IsActive: activeMap[student.ID],
			Homeworks: []dto.DashboardCourseStudentHomework{},
		}
		
		// Thêm danh sách homework
		for _, homework := range homeworks {
			isCompleted := false
			isCompletedLate := false
			if studentMap, exists := homeworkCompletionMap[student.ID]; exists {
				if completed, exists := studentMap[homework.ID]; exists {
					isCompleted = completed
				}
			}
			if lateMap, exists := homeworkLateMap[student.ID]; exists {
				if late, exists := lateMap[homework.ID]; exists {
					isCompletedLate = late
				}
			}
			
			studentDTO.Homeworks = append(studentDTO.Homeworks, dto.DashboardCourseStudentHomework{
				ID:              homework.ID,
				Name:            homework.Name,
				IsCompleted:     isCompleted,
				IsCompletedLate: isCompletedLate,
			})
		}
		
		result.Students = append(result.Students, studentDTO)
	}
	
	return &result, nil
}

func (r *dashboardCoursesRepository) GetCourseHomeworks(courseID int64, startDate, endDate string) (*dto.DashboardCourseHomeworksResponse, error) {
	if endDate == "" {
		return nil, fmt.Errorf("end_date is required")
	}

	var startTime *time.Time
	if startDate != "" {
		if startTimestamp, err := strconv.ParseInt(startDate, 10, 64); err == nil {
			t := time.Unix(startTimestamp, 0)
			startTime = &t
		} else if t, err := time.Parse("2006-01-02", startDate); err == nil {
			// Parse với UTC để giữ nguyên thời gian chuẩn
			normalized := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
			startTime = &normalized
		} else {
			return nil, fmt.Errorf("invalid start_date format: %v", err)
		}
	}

	var endTime time.Time
	if endTimestamp, err := strconv.ParseInt(endDate, 10, 64); err == nil {
		endTime = time.Unix(endTimestamp, 0)
	} else {
		var err error
		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %v", err)
		}
		// Parse với UTC để giữ nguyên thời gian chuẩn
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), endTime.Hour(), endTime.Minute(), endTime.Second(), endTime.Nanosecond(), time.UTC)
	}

	// Tính endOfDay với cùng timezone của endTime (không thêm bớt timezone)
	endOfDay := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, endTime.Location())

	var rows []struct {
		ID         int64      `json:"id"`
		Name       string     `json:"name"`
		AssignedAt *time.Time `json:"assigned_at"`
	}

	query := r.db.Table("homeworks h").
		Select("DISTINCT h.id, h.name, hrl.assigned_at").
		Joins("INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id").
		Joins("INNER JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
		Joins("INNER JOIN weeks w ON ls.week_id = w.id").
		Where("hrl.course_id = ?", courseID).
		Where("h.deleted_at IS NULL").
		Where("hrl.assigned_at IS NOT NULL")

	if startTime != nil {
		query = query.Where("? <= w.start_date AND ? >= w.end_date", *startTime, endTime)
	}

	if err := query.Order("hrl.assigned_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	var response dto.DashboardCourseHomeworksResponse
	for _, row := range rows {
		// assigned_at từ DB là timestamp without timezone, lấy nguyên không thêm bớt gì
		assignedAt := row.AssignedAt
		isLate := false
		if assignedAt != nil {
			// So sánh trực tiếp với endOfDay (cùng timezone)
			if assignedAt.After(endOfDay) {
				isLate = true
			}
		}
		response.Homeworks = append(response.Homeworks, dto.DashboardCourseHomework{
			ID:             row.ID,
			Name:           row.Name,
			AssignedAt:     assignedAt,
			IsAssignedLate: isLate,
		})
	}

	return &response, nil
}
