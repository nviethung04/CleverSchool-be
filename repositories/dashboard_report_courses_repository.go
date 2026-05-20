package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/table_manager"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DashboardReportCoursesRepository interface {
	CalculateCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error)
	SaveCourseStatistics(statistics []models.DashboardReportCourses) error
	GetExistingReport(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error)
	DeleteExistingReport(courseID int64, startDate, endDate time.Time) error
}

type dashboardReportCoursesRepository struct {
	dashboardCoursesRepo DashboardCoursesRepository
}

func NewDashboardReportCoursesRepository() DashboardReportCoursesRepository {
	return &dashboardReportCoursesRepository{
		dashboardCoursesRepo: NewDashboardCoursesRepository(),
	}
}

// createActivityLogsUnionQuery tạo UNION query cho các bảng activity_logs theo tháng
func (r *dashboardReportCoursesRepository) createActivityLogsUnionQuery(startTime, endTime time.Time) (string, error) {
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

func (r *dashboardReportCoursesRepository) CalculateCourseStatistics(startDate, endDate time.Time) ([]models.DashboardReportCourses, error) {
	// Lấy danh sách tất cả khóa học
	var courses []models.Course
	if err := db.ReplicaDB.Where("deleted_at IS NULL").Find(&courses).Error; err != nil {
		return nil, err
	}

	var results []models.DashboardReportCourses

	// Tạo endDate đến cuối ngày (23:59:59)
	endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// Tạo UNION query cho activity_logs
	allActivityLogs, err := r.createActivityLogsUnionQuery(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Batch processing với batch size 100
	batchSize := 10
	totalCourses := len(courses)

	fmt.Printf("🔄 Bắt đầu xử lý %d khóa học với batch size %d\n", totalCourses, batchSize)

	for i := 0; i < totalCourses; i += batchSize {
		end := i + batchSize
		if end > totalCourses {
			end = totalCourses
		}

		batch := courses[i:end]
		fmt.Printf("📦 Xử lý batch %d-%d/%d (%.1f%%)\n",
			i+1, end, totalCourses, float64(end)/float64(totalCourses)*100)

		// Xử lý batch
		batchResults, err := r.processCourseBatch(batch, startDate, endDate, endOfDay, allActivityLogs)
		if err != nil {
			fmt.Printf("❌ Lỗi batch %d-%d: %v\n", i+1, end, err)
			continue
		}

		results = append(results, batchResults...)
		fmt.Printf("✅ Hoàn thành batch %d-%d (%d records)\n", i+1, end, len(batchResults))
	}

	return results, nil
}

func (r *dashboardReportCoursesRepository) processCourseBatch(courses []models.Course, startDate, endDate, endOfDay time.Time, allActivityLogs string) ([]models.DashboardReportCourses, error) {
	var results []models.DashboardReportCourses

	for _, course := range courses {
		statistics := models.DashboardReportCourses{
			CourseID:  course.ID,
			StartDate: startDate,
			EndDate:   endDate,
		}

		// 1. Đếm tổng số học sinh của khóa học (qua user_courses)
		var totalStudents int64
		studentQuery := `
			SELECT COUNT(DISTINCT u.id)
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			WHERE uc.course_id = ? AND u.deleted_at IS NULL AND urr.role_id = 3 AND u.created_at <= ?
		`
		if err := db.ReplicaDB.Raw(studentQuery, course.ID, endOfDay).Scan(&totalStudents).Error; err != nil {
			return nil, err
		}
		statistics.TotalStudents = totalStudents

		// 2. Đếm tổng số giáo viên của khóa học và lấy danh sách teacher_ids + teacher_infos
		var totalTeachers int64
		var teacherIDs []string
		var teacherInfos []models.DashboardTeacherInfo
		teacherQuery := `
			SELECT u.id, u.username, u.name
			FROM users u
			INNER JOIN user_courses uc ON u.id = uc.user_id
			INNER JOIN user_ref_roles urr ON u.id = urr.user_id
			WHERE uc.course_id = ? AND u.deleted_at IS NULL AND urr.role_id = 2 AND u.created_at <= ?
		`

		var teachers []struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
		}
		if err := db.ReplicaDB.Raw(teacherQuery, course.ID, endOfDay).Scan(&teachers).Error; err != nil {
			return nil, err
		}

		totalTeachers = int64(len(teachers))
		for _, teacher := range teachers {
			teacherIDs = append(teacherIDs, strconv.FormatInt(teacher.ID, 10))
			teacherInfos = append(teacherInfos, models.DashboardTeacherInfo{
				ID:       teacher.ID,
				Username: teacher.Username,
				Name:     teacher.Name,
			})
		}
		statistics.TotalTeachers = totalTeachers
		statistics.TeacherIDs = strings.Join(teacherIDs, ",")
		statistics.TeacherInfos = teacherInfos

		// 3. Đếm số học sinh active trong thời gian (role_id = 3)
		var activeStudents int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
			studentQuery := fmt.Sprintf(`
				SELECT COUNT(DISTINCT all_activity_logs.user_id)
				FROM %s
				INNER JOIN users u ON all_activity_logs.user_id = u.id
				INNER JOIN user_courses uc ON u.id = uc.user_id
				WHERE uc.course_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(studentQuery, course.ID, 3, startDate, endDate).Scan(&activeStudents).Error; err != nil {
				// Nếu có lỗi, set về 0
				activeStudents = 0
			}
		}
		statistics.ActiveStudents = activeStudents

		// 4. Đếm số giáo viên active trong thời gian (role_id = 2)
		var activeTeachers int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" {
			teacherQuery := fmt.Sprintf(`
				SELECT COUNT(DISTINCT all_activity_logs.user_id)
				FROM %s
				INNER JOIN users u ON all_activity_logs.user_id = u.id
				INNER JOIN user_courses uc ON u.id = uc.user_id
				WHERE uc.course_id = ? AND u.deleted_at IS NULL AND all_activity_logs.role_id = ? AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)

			if err := db.ReplicaDB.Raw(teacherQuery, course.ID, 2, startDate, endDate).Scan(&activeTeachers).Error; err != nil {
				// Nếu có lỗi, set về 0
				activeTeachers = 0
			}
		}
		statistics.ActiveTeachers = activeTeachers

		// 6. Tính tổng số homework với join lesson_schedules và weeks theo date range
		// Logic: homework → lesson → chapter → program → course (không dựa vào hrl.course_id)
		// Thêm điều kiện: hrl.course_id = course_id HOẶC hrl.course_id = 0
		var totalHomeworks int64
		totalHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM courses c
			INNER JOIN chapters ch ON ch.program_id = c.program_id
			INNER JOIN lessons l ON l.chapter_id = ch.id
			INNER JOIN homework_ref_lessons hrl ON hrl.lesson_id = l.id
			INNER JOIN lesson_schedules ls ON ls.lesson_id = l.id AND ls.course_id = c.id
			INNER JOIN weeks w ON w.id = ls.week_id
			INNER JOIN homeworks h ON h.id = hrl.homework_id
			WHERE c.id = ? 
			AND h.deleted_at IS NULL
			AND hrl.created_at <= ?
			AND ? <= w.start_date AND ? >= w.end_date
			AND (hrl.course_id = ? OR hrl.course_id = 0)
		`
		if err := db.ReplicaDB.Raw(totalHomeworksQuery, course.ID, endOfDay, startDate, endDate, course.ID).Scan(&totalHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			totalHomeworks = 0
		}
		statistics.TotalHomeworks = totalHomeworks

		// 7. Tính số homework đã được giao (assigned_at IS NOT NULL) với join lesson_schedules và weeks theo date range
		var assignedHomeworks int64
		assignedHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM homeworks h
			INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id
			INNER JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id
			INNER JOIN weeks w ON ls.week_id = w.id
			WHERE hrl.course_id = ?
			AND hrl.assigned_at IS NOT NULL
			AND h.deleted_at IS NULL
			AND hrl.created_at <= ?
			AND ? <= w.start_date AND ? >= w.end_date
		`
		if err := db.ReplicaDB.Raw(assignedHomeworksQuery, course.ID, endOfDay, startDate, endDate).Scan(&assignedHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			assignedHomeworks = 0
		}
		statistics.AssignedHomeworks = assignedHomeworks

		// 8. Tính số homework đã hoàn thành (có ít nhất 1 user hoàn thành)
		var completedHomeworks int64
		completedHomeworksQuery := `
			SELECT COUNT(DISTINCT h.id)
			FROM homeworks h
			INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id
			INNER JOIN homework_users hu ON h.id = hu.homework_id AND hu.lesson_id = hrl.lesson_id
			WHERE hrl.course_id = ?
			AND h.deleted_at IS NULL
			AND hrl.assigned_at IS NOT NULL
			AND hrl.created_at <= ?
			AND hu.questions_completed >= h.total_questions
		`
		if err := db.ReplicaDB.Raw(completedHomeworksQuery, course.ID, endOfDay).Scan(&completedHomeworks).Error; err != nil {
			// Nếu có lỗi, set về 0
			completedHomeworks = 0
		}
		statistics.CompletedHomeworks = completedHomeworks

		// 9-12. Tính toán các số liệu về học sinh và homework dựa trên dữ liệu từ GetCourseStudents
		// Gọi GetCourseStudents để lấy dữ liệu học sinh và homework (giống như API dashboard/report/courses/students)
		startDateStr := startDate.Format("2006-01-02")
		endDateStr := endDate.Format("2006-01-02")
		studentsData, err := r.dashboardCoursesRepo.GetCourseStudents(course.ID, startDateStr, endDateStr)
		if err != nil {
			// Nếu có lỗi, set tất cả về 0
			statistics.StudentsCompletedAllHomeworks = 0
			statistics.StudentsCompletedHomework = 0
			statistics.StudentsNotStartedAnyHomework = 0
			statistics.StudentsDoingHomeworks = 0
			statistics.StudentActiveNotStartedAnyHomework = 0
		} else {
			// Đếm các học sinh theo các điều kiện
			var studentsCompletedAllHomeworksCount int64 = 0
			var studentsCompletedHomeworkCount int64 = 0
			var studentsNotStartedAnyHomeworkCount int64 = 0
			var studentsDoingHomeworksCount int64 = 0
			var studentActiveNotStartedAnyHomeworkCount int64 = 0

			for _, student := range studentsData.Students {
				// Kiểm tra nếu học sinh có homework nào không
				if len(student.Homeworks) == 0 {
					continue
				}

				// Đếm số homework đã hoàn thành và chưa hoàn thành
				allCompleted := true
				allCompletedNotLate := true
				allNotCompleted := true
				hasCompleted := false
				hasNotCompleted := false

				for _, homework := range student.Homeworks {
					if homework.IsCompleted {
						hasCompleted = true
						allNotCompleted = false
						if homework.IsCompletedLate {
							allCompletedNotLate = false
						}
					} else {
						hasNotCompleted = true
						allCompleted = false
						allCompletedNotLate = false
					}
				}

				// students_completed_all_homeworks: số học sinh có tất cả các homework có is_completed: true
				if allCompleted {
					studentsCompletedAllHomeworksCount++
				}

				// students_completed_homework: số học sinh có tất cả các homework có is_completed: true và homework.is_completed_late false
				if allCompleted && allCompletedNotLate {
					studentsCompletedHomeworkCount++
				}

				// students_not_started_any_homework: số học sinh có tất cả các homework có is_completed: false
				if allNotCompleted {
					studentsNotStartedAnyHomeworkCount++
				}

				// student_active_not_started_any_homework: số học sinh có tất cả các homework có is_completed: false và is_active: true
				if allNotCompleted && student.IsActive {
					studentActiveNotStartedAnyHomeworkCount++
				}

				// students_doing_homeworks: số học sinh có ít nhất 1 is_completed: false và 1 is_completed: true
				if hasCompleted && hasNotCompleted {
					studentsDoingHomeworksCount++
				}
			}

			statistics.StudentsCompletedAllHomeworks = studentsCompletedAllHomeworksCount
			statistics.StudentsCompletedHomework = studentsCompletedHomeworkCount
			statistics.StudentsNotStartedAnyHomework = studentsNotStartedAnyHomeworkCount
			statistics.StudentsDoingHomeworks = studentsDoingHomeworksCount
			statistics.StudentActiveNotStartedAnyHomework = studentActiveNotStartedAnyHomeworkCount
		}

		// 13. Tính số lượng homework có trên 50% học sinh active hoàn thành
		// Logic: Lấy danh sách học sinh active từ activity_logs (UNION ALL nhiều tháng),
		// sau đó với mỗi homework, đếm số học sinh active đã hoàn thành và so sánh tỷ lệ
		var homeworkOver50PercentStudentComplete int64
		if allActivityLogs != "(SELECT NULL as user_id, NULL as role_id, NULL as created_at, NULL as path, NULL as status_code WHERE 1=0) AS all_activity_logs" && activeStudents > 0 {
			// Bước 1: Lấy danh sách học sinh active (có trong activity_logs - UNION ALL nhiều tháng)
			// Phải dùng Raw query vì allActivityLogs là UNION ALL động
			var activeStudentIDs []int64
			activeStudentsQuery := fmt.Sprintf(`
				SELECT DISTINCT u.id
				FROM users u
				INNER JOIN user_courses uc ON u.id = uc.user_id
				INNER JOIN user_ref_roles urr ON u.id = urr.user_id
				INNER JOIN %s ON u.id = all_activity_logs.user_id
				WHERE uc.course_id = ?
				AND u.deleted_at IS NULL
				AND urr.role_id = 3
				AND all_activity_logs.role_id = 3
				AND all_activity_logs.created_at >= ? AND all_activity_logs.created_at <= ?
			`, allActivityLogs)
			if err := db.ReplicaDB.Raw(activeStudentsQuery, course.ID, startDate, endDate).Pluck("id", &activeStudentIDs).Error; err != nil {
				homeworkOver50PercentStudentComplete = 0
			} else if len(activeStudentIDs) > 0 {
				// Bước 2: Lấy danh sách homework trong khoảng thời gian (dùng GORM)
				var homeworks []models.Homework
				err := db.ReplicaDB.
					Table("homeworks h").
					Select("DISTINCT h.id").
					Joins("INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id").
					Joins("INNER JOIN lesson_schedules ls ON hrl.course_id = ls.course_id AND hrl.lesson_id = ls.lesson_id").
					Joins("INNER JOIN weeks w ON ls.week_id = w.id").
					Where("hrl.course_id = ?", course.ID).
					Where("h.deleted_at IS NULL").
					Where("hrl.assigned_at IS NOT NULL").
					Where("hrl.created_at <= ?", endOfDay).
					Where("? <= w.start_date AND ? >= w.end_date", startDate, endDate).
					Find(&homeworks).Error

				if err != nil {
					homeworkOver50PercentStudentComplete = 0
				} else if len(homeworks) > 0 {
					// Bước 3: Với mỗi homework, đếm số học sinh active đã hoàn thành (dùng GORM)
					homeworkOver50PercentCount := 0
					for _, homework := range homeworks {
						var completedActiveStudents int64

						// Đếm số học sinh active đã hoàn thành homework này (COUNT DISTINCT)
						err := db.ReplicaDB.
							Table("homework_users hu").
							Select("COUNT(DISTINCT hu.user_id)").
							Joins("INNER JOIN homeworks h ON hu.homework_id = h.id").
							Joins("INNER JOIN homework_ref_lessons hrl ON h.id = hrl.homework_id AND hu.lesson_id = hrl.lesson_id").
							Where("h.id = ?", homework.ID).
							Where("hu.questions_completed >= h.total_questions").
							Where("hrl.created_at <= ?", endOfDay).
							Where("hu.user_id IN ?", activeStudentIDs).
							Scan(&completedActiveStudents).Error

						if err != nil {
							continue
						}

						// Bước 4: So sánh tỷ lệ với tổng số học sinh active
						if float64(completedActiveStudents)/float64(len(activeStudentIDs)) >= 0.5 {
							homeworkOver50PercentCount++
						}
					}
					homeworkOver50PercentStudentComplete = int64(homeworkOver50PercentCount)
				}
			}
		}
		statistics.HomeworkOver50PercentStudentComplete = homeworkOver50PercentStudentComplete

		results = append(results, statistics)
	}

	return results, nil
}

func (r *dashboardReportCoursesRepository) SaveCourseStatistics(statistics []models.DashboardReportCourses) error {
	// Xóa các báo cáo cũ trong khoảng thời gian này trước khi lưu mới
	if len(statistics) > 0 {
		startDate := statistics[0].StartDate
		endDate := statistics[0].EndDate

		if err := db.MasterDB.Where("start_date = ? AND end_date = ?", startDate, endDate).
			Delete(&models.DashboardReportCourses{}).Error; err != nil {
			return err
		}
	}

	// Batch save với batch size 100
	batchSize := 100
	totalRecords := len(statistics)

	fmt.Printf("💾 Bắt đầu lưu %d records với batch size %d\n", totalRecords, batchSize)

	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		batch := statistics[i:end]
		fmt.Printf("💾 Lưu batch %d-%d/%d (%.1f%%)\n",
			i+1, end, totalRecords, float64(end)/float64(totalRecords)*100)

		if err := db.MasterDB.Create(&batch).Error; err != nil {
			fmt.Printf("❌ Lỗi lưu batch %d-%d: %v\n", i+1, end, err)
			return err
		}

		fmt.Printf("✅ Hoàn thành lưu batch %d-%d (%d records)\n", i+1, end, len(batch))
	}

	return nil
}

func (r *dashboardReportCoursesRepository) GetExistingReport(courseID int64, startDate, endDate time.Time) (*models.DashboardReportCourses, error) {
	var report models.DashboardReportCourses
	err := db.ReplicaDB.Where("course_id = ? AND start_date = ? AND end_date = ?", courseID, startDate, endDate).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *dashboardReportCoursesRepository) DeleteExistingReport(courseID int64, startDate, endDate time.Time) error {
	return db.MasterDB.Where("course_id = ? AND start_date = ? AND end_date = ?", courseID, startDate, endDate).
		Delete(&models.DashboardReportCourses{}).Error
}

