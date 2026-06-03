package repositories

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/table_manager"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Helper function to convert string to SQL NULL or quoted string
func getSQLNullValue(value string) string {
	if value == "" {
		return "NULL"
	}
	return fmt.Sprintf("'%s'", value)
}

type DashboardRepository interface {
	GetUserRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error)
	GetTeacherRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error)
	GetStudentRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error)
	GetActivity(filter dto.FilterDashboardAdmin) (int32, int32, int32, int32, error)
	GetSchool(filter dto.FilterDashboardAdmin) (int32, int32, int32, error)
	GetUserOverview(filter dto.FilterDashboardAdmin) (*dto.UserOverview, error)
	GetCourseOverview(filter dto.FilterDashboardAdmin) (*dto.CourseOverviewDataItem, error)
	GetLearningOverview(filter dto.FilterDashboardAdmin) (*dto.LearningOverviewDataItem, error)
	GetScoreDistribution(filter dto.FilterDashboardAdmin) (*dto.ScoreDistributionOverview, error)
	GetSystemUsage(filter dto.FilterDashboardAdmin) (*dto.SystemUsageOverview, error)
	GetQuestionBank(filter dto.FilterDashboardAdmin) (*dto.QuestionBankOverview, error)
	GetRiskWarning(filter dto.FilterDashboardAdmin) (*dto.RiskAndWarning, error)
	GetTeacherPerformance(filter dto.FilterDashboardAdmin) (*dto.TeacherPerformanceOverview, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (r *dashboardRepository) GetSchool(filter dto.FilterDashboardAdmin) (int32, int32, int32, error) {
	var currentCount, previousCount, totalCount int64

	// === Current ===
	currentQuery := db.ReplicaDB.Model(&models.School{}).
		Where("schools.created_at >= ? AND schools.created_at <= ?", filter.StartTime, filter.EndTime)

	// === Previous ===
	previousQuery := db.ReplicaDB.Model(&models.School{}).
		Where("schools.created_at >= ? AND schools.created_at <= ?", filter.StartPreviousTime, filter.EndPreviousTime)

	// === Total ===
	totalQuery := db.ReplicaDB.Model(&models.School{}).
		Where("schools.created_at <= ?", filter.EndTime)

	// === Address filter ===
	if filter.Address.WardCode != "" {
		currentQuery = currentQuery.Where("schools.ward_code = ?", filter.Address.WardCode)
		previousQuery = previousQuery.Where("schools.ward_code = ?", filter.Address.WardCode)
		totalQuery = totalQuery.Where("schools.ward_code = ?", filter.Address.WardCode)
	} else if filter.Address.ProvinceCode != "" {
		currentQuery = currentQuery.Joins("JOIN wards ON wards.code = schools.ward_code").
			Where("wards.province_code = ?", filter.Address.ProvinceCode)
		previousQuery = previousQuery.Joins("JOIN wards ON wards.code = schools.ward_code").
			Where("wards.province_code = ?", filter.Address.ProvinceCode)
		totalQuery = totalQuery.Joins("JOIN wards ON wards.code = schools.ward_code").
			Where("wards.province_code = ?", filter.Address.ProvinceCode)
	}

	// === SchoolId filter ===
	if filter.SchoolId > 0 {
		currentQuery = currentQuery.Where("schools.id = ?", filter.SchoolId)
		previousQuery = previousQuery.Where("schools.id = ?", filter.SchoolId)
		totalQuery = totalQuery.Where("schools.id = ?", filter.SchoolId)
	}

	// === Count ===
	if err := currentQuery.Count(&currentCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := previousQuery.Count(&previousCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := totalQuery.Count(&totalCount).Error; err != nil {
		return 0, 0, 0, err
	}

	return int32(totalCount), int32(currentCount), int32(previousCount), nil
}

func (r *dashboardRepository) GetUserRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error) {
	var currentCount, previousCount, totalCount int64

	// currentQuery := db.ReplicaDB.Model(&models.User{}).
	// 	Where("users.created_at >= ? AND users.created_at <= ?", filter.StartTime, filter.EndTime)

	currentQuery := db.ReplicaDB.Model(&models.User{}).
		Where("users.created_at <= ?", filter.EndTime)

	previousQuery := db.ReplicaDB.Model(&models.User{}).
		Where("users.created_at >= ? AND users.created_at <= ?", filter.StartPreviousTime, filter.EndPreviousTime)

	totalQuery := db.ReplicaDB.Model(&models.User{}).
		Where("users.created_at <= ?", filter.EndTime)

	if filter.Address.ProvinceCode != "" || filter.Address.WardCode != "" {
		currentQuery = currentQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")

		if filter.Address.WardCode != "" {
			currentQuery = currentQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			previousQuery = previousQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			totalQuery = totalQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
		} else if filter.Address.ProvinceCode != "" {
			currentQuery = currentQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			previousQuery = previousQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			totalQuery = totalQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
		}
	}

	if filter.CourseId > 0 || filter.ProgramId > 0 {
		currentQuery = currentQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")

		if filter.CourseId > 0 {
			currentQuery = currentQuery.Where("uc.course_id = ?", filter.CourseId)
			previousQuery = previousQuery.Where("uc.course_id = ?", filter.CourseId)
			totalQuery = totalQuery.Where("uc.course_id = ?", filter.CourseId)
		}

		if filter.ProgramId > 0 {
			currentQuery = currentQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			previousQuery = previousQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			totalQuery = totalQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
		}
	}

	if filter.SchoolId > 0 {
		currentQuery = currentQuery.Joins(`
			LEFT JOIN user_classes uc_school ON uc_school.user_id = users.id
			LEFT JOIN classes cls ON cls.id = uc_school.class_id
		`).Where(`
			users.school_id = ? OR cls.school_id = ?
		`, filter.SchoolId, filter.SchoolId)

		previousQuery = previousQuery.Joins(`
			LEFT JOIN user_classes uc_school ON uc_school.user_id = users.id
			LEFT JOIN classes cls ON cls.id = uc_school.class_id
		`).Where(`
			users.school_id = ? OR cls.school_id = ?
		`, filter.SchoolId, filter.SchoolId)

		totalQuery = totalQuery.Joins(`
			LEFT JOIN user_classes uc_school ON uc_school.user_id = users.id
			LEFT JOIN classes cls ON cls.id = uc_school.class_id
		`).Where(`
			users.school_id = ? OR cls.school_id = ?
		`, filter.SchoolId, filter.SchoolId)
	}

	if err := currentQuery.Distinct("users.id").Count(&currentCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := previousQuery.Distinct("users.id").Count(&previousCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := totalQuery.Distinct("users.id").Count(&totalCount).Error; err != nil {
		return 0, 0, 0, err
	}

	return int32(totalCount), int32(currentCount), int32(previousCount), nil
}

func (r *dashboardRepository) GetStudentRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error) {
	var currentCount, totalCount, previousCount int64

	currentQuery := db.ReplicaDB.Model(&models.User{}).
		// Where("users.created_at >= ? AND users.created_at <= ?", filter.StartTime, filter.EndTime).
		Where("users.created_at <= ?", filter.EndTime).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.StudentRoleId)

	previousQuery := db.ReplicaDB.Model(&models.User{}).
		// Where("users.created_at >= ? AND users.created_at <= ?", filter.StartPreviousTime, filter.EndPreviousTime).
		Where("users.created_at <= ?", filter.EndPreviousTime).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.StudentRoleId)

	totalQuery := db.ReplicaDB.Model(&models.User{}).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.StudentRoleId)
		// Where("users.created_at <= ?", filter.EndTime)

	if filter.Address.ProvinceCode != "" || filter.Address.WardCode != "" {
		currentQuery = currentQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")

		if filter.Address.WardCode != "" {
			currentQuery = currentQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			previousQuery = previousQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			totalQuery = totalQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
		} else if filter.Address.ProvinceCode != "" {
			currentQuery = currentQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			previousQuery = previousQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			totalQuery = totalQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
		}
	}

	if filter.CourseId > 0 || filter.ProgramId > 0 {
		currentQuery = currentQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")

		if filter.CourseId > 0 {
			currentQuery = currentQuery.Where("uc.course_id = ?", filter.CourseId)
			previousQuery = previousQuery.Where("uc.course_id = ?", filter.CourseId)
			totalQuery = totalQuery.Where("uc.course_id = ?", filter.CourseId)
		}

		if filter.ProgramId > 0 {
			currentQuery = currentQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			previousQuery = previousQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			totalQuery = totalQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
		}
	}

	if filter.SchoolId > 0 {
		currentQuery = currentQuery.Where("users.school_id = ?", filter.SchoolId)
		previousQuery = previousQuery.Where("users.school_id = ?", filter.SchoolId)
		totalQuery = totalQuery.Where("users.school_id = ?", filter.SchoolId)
	}

	if err := currentQuery.Distinct("users.id").Count(&currentCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := previousQuery.Distinct("users.id").Count(&previousCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := totalQuery.Distinct("users.id").Count(&totalCount).Error; err != nil {
		return 0, 0, 0, err
	}

	return int32(totalCount), int32(currentCount), int32(previousCount), nil
}

func (r *dashboardRepository) GetTeacherRegister(filter dto.FilterDashboardAdmin) (int32, int32, int32, error) {
	var currentCount, previousCount, totalCount int64

	// === Current range ===
	currentQuery := db.ReplicaDB.Model(&models.User{}).
		Where("users.created_at <= ?", filter.EndTime).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.TeacherRoleId)

	// === Previous range ===
	previousQuery := db.ReplicaDB.Model(&models.User{}).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("users.created_at <= ?", filter.EndPreviousTime).
		Where("urr.role_id = ?", models.TeacherRoleId)

	// === Total range ===
	totalQuery := db.ReplicaDB.Model(&models.User{}).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.TeacherRoleId)

	// === Address filter ===
	if filter.Address.ProvinceCode != "" || filter.Address.WardCode != "" {
		currentQuery = currentQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_addresses ua ON ua.user_id = users.id")

		if filter.Address.WardCode != "" {
			currentQuery = currentQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			previousQuery = previousQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
			totalQuery = totalQuery.Where("ua.ward_code = ?", filter.Address.WardCode)
		} else if filter.Address.ProvinceCode != "" {
			currentQuery = currentQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			previousQuery = previousQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
			totalQuery = totalQuery.Where("ua.province_code = ?", filter.Address.ProvinceCode)
		}
	}

	// === Course and Program filter ===
	if filter.CourseId > 0 || filter.ProgramId > 0 {
		currentQuery = currentQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		previousQuery = previousQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")
		totalQuery = totalQuery.Joins("JOIN user_courses uc ON uc.user_id = users.id")

		if filter.CourseId > 0 {
			currentQuery = currentQuery.Where("uc.course_id = ?", filter.CourseId)
			previousQuery = previousQuery.Where("uc.course_id = ?", filter.CourseId)
			totalQuery = totalQuery.Where("uc.course_id = ?", filter.CourseId)
		}

		if filter.ProgramId > 0 {
			currentQuery = currentQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			previousQuery = previousQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
			totalQuery = totalQuery.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", filter.ProgramId)
		}
	}

	// === School filter ===
	if filter.SchoolId > 0 {
		currentQuery = currentQuery.
			Joins("JOIN user_classes urc ON urc.user_id = users.id").
			Joins("JOIN classes c ON c.id = urc.class_id").
			Where("c.school_id = ?", filter.SchoolId)

		previousQuery = previousQuery.
			Joins("JOIN user_classes urc ON urc.user_id = users.id").
			Joins("JOIN classes c ON c.id = urc.class_id").
			Where("c.school_id = ?", filter.SchoolId)

		totalQuery = totalQuery.
			Joins("JOIN user_classes urc ON urc.user_id = users.id").
			Joins("JOIN classes c ON c.id = urc.class_id").
			Where("c.school_id = ?", filter.SchoolId)
	}

	// === Thực thi ===
	if err := currentQuery.Distinct("users.id").Count(&currentCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := previousQuery.Distinct("users.id").Count(&previousCount).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := totalQuery.Distinct("users.id").Count(&totalCount).Error; err != nil {
		return 0, 0, 0, err
	}

	return int32(totalCount), int32(currentCount), int32(previousCount), nil
}

func (r *dashboardRepository) GetActivityUserIds(startTime, endTime time.Time, courseId, schoolId, programId int64) (models.ActivityIds, error) {
	var allUserIDs []int64
	historyUseRepo := NewHistoryUseRepository()

	current := startTime
	for current.Before(endTime) || current.Equal(endTime) {
		monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		if monthEnd.After(endTime) {
			monthEnd = endTime
		}

		var userIDs []int64
		var activityIds models.ActivityIds

		now := time.Now()
		isCurrentMonth := now.Year() == current.Year() && now.Month() == current.Month()

		if isCurrentMonth {
			idsRes, err := historyUseRepo.ActivityIds(monthStart, monthEnd)
			if err != nil {
				return models.ActivityIds{}, err
			}
			userIDs = idsRes.Ids

			var historyUses []models.HistoryUse
			err = db.ReplicaDB.
				Where("type = ?", "month").
				Find(&historyUses).Error
			if err != nil {
				config.Log.Error(err)
				return models.ActivityIds{}, err
			}

			for _, historyUse := range historyUses {
				var activityIds models.ActivityIds
				if err := json.Unmarshal(historyUse.ActivityIds, &activityIds); err != nil {
					config.Log.Error(err)
					return models.ActivityIds{}, err
				}

				for _, id := range activityIds.Ids {
					userIDs = append(userIDs, id)
				}
			}
		} else {
			var historyUse models.HistoryUse
			err := db.ReplicaDB.
				Where("date = ? AND type = ?", monthStart.Format("2006-01-02"), "month").
				First(&historyUse).Error
			if err != nil {
				return models.ActivityIds{}, err
			}
			if err := json.Unmarshal(historyUse.ActivityIds, &activityIds); err != nil {
				return models.ActivityIds{}, err
			}

			for _, id := range activityIds.Ids {
				userIDs = append(userIDs, id)
			}
		}

		allUserIDs = append(allUserIDs, userIDs...)

		current = monthStart.AddDate(0, 1, 0)
	}

	unique := make(map[int64]struct{})
	var ids []int64
	for _, id := range allUserIDs {
		if _, ok := unique[id]; !ok {
			unique[id] = struct{}{}
			ids = append(ids, id)
		}
	}

	// filter by course, program, school
	q := db.ReplicaDB.Model(&models.User{})
	if courseId != 0 || programId != 0 {
		q = q.Joins("JOIN user_courses uc ON uc.user_id = users.id")

		if courseId != 0 {
			q = q.Where("uc.course_id = ?", courseId)
		}
		if programId != 0 {
			q = q.Joins("JOIN courses co ON co.id = uc.course_id").
				Where("co.program_id = ?", programId)
		}
	}
	if schoolId != 0 {
		q = q.Where("users.school_id = ?", schoolId)
	}

	// lấy danh sách user id thỏa điều kiện course/program/school
	var filteredIDs []int64
	err := q.Where("users.id IN ?", ids).
		Pluck("users.id", &filteredIDs).Error
	if err != nil {
		return models.ActivityIds{}, err
	}

	return models.ActivityIds{Ids: filteredIDs}, nil
}

func (r *dashboardRepository) getUserOverviewFallback(filter dto.FilterDashboardAdmin) (*dto.UserOverview, error) {
	// Fallback method using individual calls
	config.Log.Info("Using fallback method - calling individual methods...")

	userTotal, userCurrent, userPrevious, err := r.GetUserRegister(filter)
	if err != nil {
		config.Log.Error("Error in GetUserRegister:", err)
		return nil, err
	}

	studentTotal, studentCurrent, studentPrevious, err := r.GetStudentRegister(filter)
	if err != nil {
		config.Log.Error("Error in GetStudentRegister:", err)
		return nil, err
	}

	teacherTotal, teacherCurrent, teacherPrevious, err := r.GetTeacherRegister(filter)
	if err != nil {
		config.Log.Error("Error in GetTeacherRegister:", err)
		return nil, err
	}

	activityTotal, activityCurrent, _, activityNew, err := r.GetActivity(filter)
	if err != nil {
		config.Log.Error("Error in GetActivity:", err)
		return nil, err
	}

	schoolTotal, schoolCurrent, schoolPrevious, err := r.GetSchool(filter)
	if err != nil {
		config.Log.Error("Error in GetSchool:", err)
		return nil, err
	}

	config.Log.Info("Fallback method results:",
		"UserTotal", userTotal, "UserCurrent", userCurrent,
		"StudentTotal", studentTotal, "StudentCurrent", studentCurrent,
		"TeacherTotal", teacherTotal, "TeacherCurrent", teacherCurrent,
		"ActivityTotal", activityTotal, "ActivityCurrent", activityCurrent,
		"SchoolTotal", schoolTotal, "SchoolCurrent", schoolCurrent,
	)

	return &dto.UserOverview{
		User: &dto.DashboardAdminItem{
			Total:       userTotal,
			TotalNumber: userCurrent,
			ChangeValue: float32(userCurrent - userPrevious),
		},
		Student: &dto.DashboardAdminItem{
			Total:       studentTotal,
			TotalNumber: studentCurrent,
			ChangeValue: float32(studentCurrent - studentPrevious),
		},
		Teacher: &dto.DashboardAdminItem{
			Total:       teacherTotal,
			TotalNumber: teacherCurrent,
			ChangeValue: float32(teacherCurrent - teacherPrevious),
		},
		Activity: &dto.DashboardAdminItem{
			Total:       activityTotal,
			TotalNumber: activityCurrent,
			ChangeValue: float32(activityNew),
		},
		School: &dto.DashboardAdminItem{
			Total:       schoolTotal,
			TotalNumber: schoolCurrent,
			ChangeValue: float32(schoolCurrent - schoolPrevious),
		},
	}, nil
}

func (r *dashboardRepository) GetActivity(filter dto.FilterDashboardAdmin) (int32, int32, int32, int32, error) {
	currentUserIds, err := r.GetActivityUserIds(filter.StartTime, filter.EndTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	previousIds, err := r.GetActivityUserIds(filter.StartPreviousTime, filter.EndPreviousTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	totalIds, err := r.GetActivityUserIds(filter.StartTime, filter.EndPreviousTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	totalCount := int64(len(totalIds.Ids))
	currentCount := int64(len(currentUserIds.Ids))
	previousCount := int64(len(previousIds.Ids))
	newCount := currentCount - previousCount

	return int32(totalCount), int32(currentCount), int32(previousCount), int32(newCount), nil
}

func (r *dashboardRepository) GetCourseOverview(filter dto.FilterDashboardAdmin) (*dto.CourseOverviewDataItem, error) {
	var (
		programCount,
		lessonCount,
		lessonPlanCount,
		questionCount,
		questionUsedCount,
		homeworkCount,
		examCount,
		userExamCount,
		completedLessonPlanCount,
		submittedHomeworkCount,
		submittedExamCount,
		unsubmittedUserExamCount int64
	)

	type QuestionID struct {
		ID int
	}

	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	db := db.ReplicaDB

	// === Program ===
	var programId int64
	programQuery := db.Model(&models.Program{}).
		Joins("LEFT JOIN courses co ON co.program_id = programs.id")

	if filter.CourseId > 0 {
		programQuery = programQuery.Where("co.id = ?", filter.CourseId)

		db.Model(&models.Course{}).
			Select("program_id").
			Where("id = ?", filter.CourseId).
			Scan(&programId)
	} else if filter.ProgramId > 0 {
		programId = filter.ProgramId
		programQuery = programQuery.Where("programs.id = ?", filter.ProgramId)
	}
	if filter.SchoolId > 0 {
		programQuery = programQuery.Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}
	if filterByDate {
		programQuery = programQuery.Where("programs.created_at <= ?", filter.EndTime)
		// programQuery = programQuery.Where("courses.created_at >= ? AND courses.created_at <= ?", filter.StartTime, filter.EndTime)
	}
	programQuery.Distinct("programs.id").Count(&programCount)

	// === Lesson ===
	lessonQuery := db.Model(&models.Lesson{}).
		Joins("JOIN chapters AS c ON lessons.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id")

	if programId > 0 {
		lessonQuery = lessonQuery.Where("p.id = ?", programId)
	}
	if filter.SchoolId > 0 {
		lessonQuery = lessonQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}
	if filterByDate {
		lessonQuery = lessonQuery.Where("lessons.created_at <= ?", filter.EndTime)
		// lessonQuery = lessonQuery.Where("lessons.created_at >= ? AND lessons.created_at <= ?", filter.StartTime, filter.EndTime)
	}
	lessonQuery.Count(&lessonCount)

	// === Lesson Plan ===
	lessonPlanQuery := db.Table("lesson_plans AS lp")

	if programId > 0 || filter.SchoolId > 0 {
		lessonPlanQuery = lessonPlanQuery.
			Joins("JOIN lesson_plan_ref_lessons AS lprl ON lp.id = lprl.lesson_plan_id").
			Joins("JOIN lessons AS l ON lprl.lesson_id = l.id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
	}

	if programId > 0 {
		lessonPlanQuery = lessonPlanQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		lessonPlanQuery = lessonPlanQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		lessonPlanQuery = lessonPlanQuery.Where("lp.created_at <= ?", filter.EndTime)
		// lessonPlanQuery = lessonPlanQuery.Where("lp.created_at >= ? AND lp.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	lessonPlanQuery.Where("lp.deleted_at IS NULL").
		Distinct("lp.id").
		Count(&lessonPlanCount)

	// === Question ===
	questionQuery := db.Model(&models.Question{})
	if filterByDate {
		questionQuery = questionQuery.Where("created_at <= ?", filter.EndTime)
		// questionQuery = questionQuery.Where("created_at >= ? AND created_at <= ?", filter.StartTime, filter.EndTime)
	}
	questionQuery.Count(&questionCount)

	// === Homework ===
	homeworkQuery := db.Table("homeworks AS h").
		Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons AS l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters AS c ON l.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id")

	if programId > 0 {
		homeworkQuery = homeworkQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		homeworkQuery = homeworkQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		homeworkQuery = homeworkQuery.Where("h.created_at <= ?", filter.EndTime)
		// homeworkQuery = homeworkQuery.Where("h.created_at >= ? AND h.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	homeworkQuery.Where("h.deleted_at IS NULL").
		Select("COUNT(DISTINCT h.id)").
		Count(&homeworkCount)

	// === Exam ===
	examQuery := db.Table("exams AS e").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
		Joins("JOIN chapters AS c ON l.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id")

	if programId > 0 {
		examQuery = examQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		examQuery = examQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		examQuery = examQuery.Where("e.created_at <= ?", filter.EndTime)
		// examQuery = examQuery.Where("e.created_at >= ? AND e.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	examQuery.Where("e.deleted_at IS NULL").
		Select("COUNT(DISTINCT e.id)").
		Count(&examCount)

	// === User Exam ===
	userExamQuery := db.Table("exam_users AS eu").
		Joins("JOIN exams AS e ON eu.exam_id = e.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
		Joins("JOIN chapters AS c ON l.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id").
		Where("e.deleted_at IS NULL")

	if programId > 0 {
		userExamQuery = userExamQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		userExamQuery = userExamQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		userExamQuery = userExamQuery.Where("eu.created_at <= ?", filter.EndTime)
		// userExamQuery = userExamQuery.Where("eu.created_at >= ? AND eu.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	subQuery := userExamQuery.Select("DISTINCT eu.user_id, eu.exam_id")

	db.Table("(?) as sub", subQuery).Count(&userExamCount)

	// === Submitted Homework ===
	submittedHomeworkQuery := db.Model(&models.HomeworkUser{}).
		Joins("JOIN homeworks AS h ON h.id = homework_users.homework_id").
		Joins("JOIN homework_ref_lessons AS hrl ON hrl.homework_id = h.id").
		Joins("JOIN lessons AS l ON l.id = hrl.lesson_id").
		Joins("JOIN chapters AS c ON l.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id")

	if programId > 0 {
		submittedHomeworkQuery = submittedHomeworkQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		submittedHomeworkQuery = submittedHomeworkQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		submittedHomeworkQuery = submittedHomeworkQuery.Where("homework_users.created_at <= ?", filter.EndTime)
		// submittedHomeworkQuery = submittedHomeworkQuery.Where("homework_users.created_at >= ? AND homework_users.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	submittedHomeworkQuery.
		Select("COUNT(DISTINCT homework_users.homework_id)").
		Scan(&submittedHomeworkCount)

	// === Submitted Exam ===
	submittedExamQuery := db.Model(&models.ExamUser{}).
		Joins("JOIN exams AS e ON e.id = exam_users.exam_id").
		Joins("JOIN exam_ref_lessons AS erl ON erl.exam_id = e.id").
		Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
		Joins("JOIN chapters AS c ON l.chapter_id = c.id").
		Joins("JOIN programs p ON c.program_id = p.id")

	if programId > 0 {
		submittedExamQuery = submittedExamQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		submittedExamQuery = submittedExamQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		submittedExamQuery = submittedExamQuery.
			Where("exam_users.created_at <= ?", filter.EndTime)
		// submittedExamQuery = submittedExamQuery.
		//   Where("exam_users.created_at >= ? AND exam_users.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	err := submittedExamQuery.
		Select("COALESCE(COUNT(DISTINCT exam_users.exam_id), 0)").
		Scan(&submittedExamCount).Error

	if err != nil {
		config.Log.Error("Error executing submitted exam query: ", err)
	}

	// === Cloned QuestionIDs ===
	var questionIDs []QuestionID

	queryClonedQuestion := `
		SELECT DISTINCT (jsonb_array_elements(cq.questions)->>'id')::int AS id
		FROM cloned_questions cq
		LEFT JOIN homeworks hw
			ON cq.assignment_type = 'homework' AND cq.assignment_id = hw.id
		LEFT JOIN homework_ref_lessons hrl
			ON hw.id = hrl.homework_id
		LEFT JOIN exams ex
			ON cq.assignment_type = 'exam' AND cq.assignment_id = ex.id
		LEFT JOIN exam_ref_lessons erl
			ON ex.id = erl.exam_id
		LEFT JOIN lessons l
			ON l.id = COALESCE(hrl.lesson_id, erl.lesson_id)
		LEFT JOIN chapters c
			ON c.id = l.chapter_id
		LEFT JOIN programs p
			ON p.id = c.program_id
	`

	var conditions []string
	var args []interface{}

	if filterByDate {
		conditions = append(conditions, "cq.created_at <= ?")
		args = append(args, filter.EndTime)
	}

	if programId > 0 {
		conditions = append(conditions, "p.id = ?")
		args = append(args, programId)
	}

	if filter.CourseId > 0 {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM courses co
				WHERE co.program_id = p.id AND co.id = ?
			)
		`)
		args = append(args, filter.CourseId)
	}

	if filter.SchoolId > 0 {
		conditions = append(conditions, `
			EXISTS (
				SELECT 1 FROM course_schools cs
				JOIN courses co ON co.id = cs.course_id
				WHERE co.program_id = p.id AND cs.school_id = ?
			)
		`)
		args = append(args, filter.SchoolId)
	}

	if len(conditions) > 0 {
		queryClonedQuestion += " WHERE " + strings.Join(conditions, " AND ")
	}

	db.Raw(queryClonedQuestion, args...).Scan(&questionIDs)

	uniqueQuestionIDs := make([]int, 0)
	seen := make(map[int]struct{})
	for _, q := range questionIDs {
		if _, ok := seen[q.ID]; !ok {
			seen[q.ID] = struct{}{}
			uniqueQuestionIDs = append(uniqueQuestionIDs, q.ID)
		}
	}

	// === Question used ===
	questionUsedQuery := db.Model(&models.Question{}).
		Where("id IN (?)", uniqueQuestionIDs)
	if filterByDate {
		questionUsedQuery = questionUsedQuery.Where("created_at <= ?", filter.EndTime)
		// questionUsedQuery = questionUsedQuery.Where("created_at >= ? AND created_at <= ?", filter.StartTime, filter.EndTime)
	}
	questionUsedQuery.Count(&questionUsedCount)

	// === Completed Lesson Plan ===
	completedLessonPlanQuery := db.Table("lesson_plans AS lp").
		Joins("JOIN lesson_plan_completes AS lpc ON lp.id = lpc.lesson_plan_id")

	if programId > 0 || filter.SchoolId > 0 {
		completedLessonPlanQuery = completedLessonPlanQuery.
			Joins("JOIN lesson_plan_ref_lessons AS lprl ON lp.id = lprl.lesson_plan_id").
			Joins("JOIN lessons AS l ON lprl.lesson_id = l.id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
	}

	if programId > 0 {
		completedLessonPlanQuery = completedLessonPlanQuery.Where("p.id = ?", programId)
	}

	if filter.SchoolId > 0 {
		completedLessonPlanQuery = completedLessonPlanQuery.
			Joins("JOIN courses co ON co.program_id = p.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		completedLessonPlanQuery = completedLessonPlanQuery.Where("lpc.completed_at <= ?", filter.EndTime)
		// completedLessonPlanQuery = completedLessonPlanQuery.
		//    Where("lpc.completed_at >= ? AND lpc.completed_at <= ?", filter.StartTime, filter.EndTime)
	}

	completedLessonPlanQuery.
		Distinct("lp.id").
		Count(&completedLessonPlanCount)

	// === Ungraded ===
	ungradedQuery := `
		SELECT COUNT(*) AS ungraded_exam_users
		FROM (
			SELECT DISTINCT eu.exam_id, eu.user_id
			FROM exam_users eu
			JOIN exams e ON eu.exam_id = e.id
			JOIN exam_ref_lessons erl ON erl.exam_id = e.id
			JOIN lessons l ON erl.lesson_id = l.id
			JOIN chapters c ON l.chapter_id = c.id
			JOIN courses co ON co.program_id = c.program_id
			WHERE EXISTS (
				SELECT 1
				FROM exam_question_user_manual_scoring eqms
				WHERE eqms.exam_id = eu.exam_id
					AND eqms.user_id = eu.user_id
					AND eqms.is_scored = FALSE
	`

	if programId > 0 {
		ungradedQuery += " AND c.program_id = ?"
	}
	if filter.SchoolId > 0 {
		ungradedQuery += " AND EXISTS (SELECT 1 FROM course_schools cs WHERE cs.course_id = co.id AND cs.school_id = ?)"
	}

	ungradedQuery += `
			)
		) AS ungraded_subquery
	`

	if programId > 0 && filter.SchoolId > 0 {
		db.Raw(ungradedQuery, programId, filter.SchoolId).Scan(&unsubmittedUserExamCount)
	} else if programId > 0 {
		db.Raw(ungradedQuery, programId).Scan(&unsubmittedUserExamCount)
	} else if filter.SchoolId > 0 {
		db.Raw(ungradedQuery, filter.SchoolId).Scan(&unsubmittedUserExamCount)
	} else {
		db.Raw(ungradedQuery).Scan(&unsubmittedUserExamCount)
	}

	return &dto.CourseOverviewDataItem{
		Programs:          int32(programCount),
		Lessons:           int32(lessonCount),
		LessonPlans:       int32(lessonPlanCount),
		Questions:         int32(questionCount),
		Homeworks:         int32(homeworkCount),
		Exams:             int32(examCount),
		UserExams:         int32(userExamCount),
		LessonPlanCurrent: int32(completedLessonPlanCount),
		HomeworkCurrent:   int32(submittedHomeworkCount),
		ExamCurrent:       int32(submittedExamCount),
		QuestionCurrent:   int32(questionUsedCount),
		UserExamCurrent:   int32(max(0, userExamCount-unsubmittedUserExamCount)),
	}, nil
}

func (r *dashboardRepository) GetLearningOverview(filter dto.FilterDashboardAdmin) (*dto.LearningOverviewDataItem, error) {
	db := db.ReplicaDB

	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	type ExamWithUser struct {
		ExamID     int64            `json:"exam_id"`
		ExamName   string           `json:"exam_name"`
		ExamUserID int64            `json:"exam_user_id"`
		UserID     int64            `json:"user_id"`
		UserName   string           `json:"user_name"`
		Ratio      float64          `json:"ratio"`
		AvatarInfo models.MediaInfo `json:"avatar_info"`
	}

	type WeekScore struct {
		WeekID     int64   `json:"week_id"`
		WeekNumber int     `json:"week_number"`
		AvgScore   float64 `json:"avg_score"`
	}

	// === 1. Lấy tuần ===
	var weeks []struct {
		ID         int64
		WeekNumber int
	}

	weeksQuery := db.Table("weeks").
		Select("id, week_number")

	if filterByDate {
		weeksQuery = weeksQuery.
			Where("end_date >= ?", filter.StartTime).
			Where("end_date <= ?", filter.EndTime).
			Order("id ASC")
	} else {
		weeksQuery = weeksQuery.
			Where("start_date <= NOW()").
			Order("id DESC").
			Limit(20)
	}

	if err := weeksQuery.Scan(&weeks).Error; err != nil {
		return nil, err
	}

	// === 2. Query avg score ===
	var weekScores []struct {
		WeekID   int64
		AvgScore float64
	}

	selectClause := `
		w.id AS week_id,
		COALESCE(SUM(eu.ratio), 0) / NULLIF(COUNT(eu.id), 0) AS avg_score
	`
	weekQuery := db.Table("weeks w").
		Select(selectClause).
		Joins(`LEFT JOIN exam_users eu ON eu.created_at BETWEEN w.start_date AND w.end_date`).
		Joins(`LEFT JOIN exams e ON eu.exam_id = e.id`).
		Joins(`LEFT JOIN exam_ref_lessons erl ON erl.exam_id = e.id`).
		Joins(`LEFT JOIN lessons l ON erl.lesson_id = l.id`).
		Joins(`LEFT JOIN chapters c ON l.chapter_id = c.id`).
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`)

	if filter.CourseId > 0 {
		weekQuery = weekQuery.Where("co.id = ?", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		weekQuery = weekQuery.Where("c.program_id = ?", filter.ProgramId)
	}

	if filter.SchoolId > 0 {
		weekQuery = weekQuery.Joins(
			`JOIN course_schools cs ON cs.course_id = co.id AND cs.school_id = ?`, filter.SchoolId,
		)
	}

	if filter.ClassId > 0 {
		weekQuery = weekQuery.
			Joins(`JOIN users u ON eu.user_id = u.id`).
			Joins(`JOIN user_classes uc ON uc.user_id = u.id AND uc.class_id = ?`, filter.ClassId)
	}

	if filter.TeacherId > 0 {
		weekQuery = weekQuery.
			Joins(`JOIN user_courses uc2 ON uc2.course_id = co.id AND uc2.user_id = ?`, filter.TeacherId)
	}

	if err := weekQuery.
		Where("w.start_date <= NOW()").
		Group("w.id").
		Order("w.id DESC").
		Limit(20).
		Scan(&weekScores).Error; err != nil {
		return nil, err
	}

	// === 3. Map đủ 20 tuần ===
	scoreMap := make(map[int64]float64)
	for _, ws := range weekScores {
		scoreMap[ws.WeekID] = ws.AvgScore
	}

	learningScoreWeeks := make([]dto.LearningScoreWeek, 0, len(weeks))
	gemsByWeeks := make([]dto.GemsByWeek, 0, len(weeks))
	for _, w := range weeks {
		score := float32(0)
		if v, ok := scoreMap[w.ID]; ok {
			score = float32(v)
		}
		learningScoreWeeks = append(learningScoreWeeks, dto.LearningScoreWeek{
			Id:         w.ID,
			WeekNumber: int32(w.WeekNumber),
			Score:      score,
		})
		gemsByWeeks = append(gemsByWeeks, dto.GemsByWeek{
			Id:         w.ID,
			WeekNumber: int32(w.WeekNumber),
			Count:      0,
		})
	}

	type TopStudent struct {
		UserId      int64
		UserName    string
		Avatar      models.MediaInfo
		AvgRatio    float64
		ExamId      int64
		ExamUserId  int64
		StudentName string
		ExamName    string
		SchoolName  string
		ClassName   string
	}

	// === 4. Top Highest ===
	var topHighestExam []TopStudent

	topHighestQuery := db.Model(&models.ExamUser{}).
		Select(`
			exam_users.user_id,
			users.name AS user_name,
			users.avatar_info,
			AVG(exam_users.ratio) AS avg_ratio,
			exam_users.exam_id as exam_id,
			exam_users.id as exam_user_id,
			e.name AS exam_name,
			users.name AS student_name,
			COALESCE(schools.name, '') AS school_name,
			COALESCE(MIN(classes.name), '') AS class_name
		`).
		Joins("JOIN users ON exam_users.user_id = users.id").
		Joins("JOIN exams e ON exam_users.exam_id = e.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons l ON erl.lesson_id = l.id").
		Joins("JOIN chapters c ON l.chapter_id = c.id").
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`).
		Joins(`LEFT JOIN user_classes uc ON uc.user_id = users.id`).
		Joins(`LEFT JOIN schools ON schools.id = users.school_id`).
		Joins(`LEFT JOIN classes ON classes.id = uc.class_id`)

	if filterByDate {
		topHighestQuery = topHighestQuery.Where("exam_users.created_at <= ?", filter.EndTime)
		// topHighestQuery = topHighestQuery.Where("exam_users.created_at >= ? AND exam_users.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	if filter.CourseId > 0 {
		topHighestQuery = topHighestQuery.Where("co.id = ?", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		topHighestQuery = topHighestQuery.Where("c.program_id = ?", filter.ProgramId)
	}

	if filter.SchoolId > 0 {
		topHighestQuery = topHighestQuery.Where("schools.id = ?", filter.SchoolId)
	}

	if filter.ClassId > 0 {
		topHighestQuery = topHighestQuery.Where("uc.class_id = ?", filter.ClassId)
	}

	if filter.TeacherId > 0 {
		topHighestQuery = topHighestQuery.Joins(`JOIN user_courses uc2 ON uc2.course_id = co.id AND uc2.user_id = ?`, filter.TeacherId)
	}

	err := topHighestQuery.Group("exam_users.user_id, users.name, users.avatar_info, exam_users.exam_id, exam_users.id, e.name, users.name, schools.name, classes.name").
		Having("AVG(exam_users.ratio) > ?", 75).
		Order("avg_ratio DESC").
		Limit(10).
		Scan(&topHighestExam).Error

	if err != nil {
		return nil, err
	}

	topHighest := make([]dto.StudentScore, 0, len(topHighestExam))
	for _, student := range topHighestExam {
		topHighest = append(topHighest, dto.StudentScore{
			Id:         student.UserId,
			ExamId:     student.ExamId,
			ExamUserId: student.ExamUserId,
			Name:       student.UserName,
			ExamName:   student.ExamName,
			Score:      float32(student.AvgRatio),
			AvatarInfo: student.Avatar,
			ClassName:  student.ClassName,
			SchoolName: student.SchoolName,
		})
	}

	// === 5. Top Lowest ===
	var topLowestExam []TopStudent

	topLowestQuery := db.Model(&models.ExamUser{}).
		Select(`
			exam_users.user_id,
			users.name AS user_name,
			users.avatar_info,
			AVG(exam_users.ratio) AS avg_ratio,
			exam_users.exam_id as exam_id,
			exam_users.id as exam_user_id,
			e.name AS exam_name,
			users.name AS student_name,
			COALESCE(schools.name, '') AS school_name,
			COALESCE(MIN(classes.name), '') AS class_name
		`).
		Joins("JOIN users ON exam_users.user_id = users.id").
		Joins("JOIN exams e ON exam_users.exam_id = e.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons l ON erl.lesson_id = l.id").
		Joins("JOIN chapters c ON l.chapter_id = c.id").
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`).
		Joins(`LEFT JOIN user_classes uc ON uc.user_id = users.id`).
		Joins(`LEFT JOIN schools ON schools.id = users.school_id`).
		Joins(`LEFT JOIN classes ON classes.id = uc.class_id`)

	if filterByDate {
		topLowestQuery = topLowestQuery.Where("exam_users.created_at <= ?", filter.EndTime)
		// topLowestQuery = topLowestQuery.Where("exam_users.created_at >= ? AND exam_users.created_at <= ?", filter.StartTime, filter.EndTime)
	}

	if filter.CourseId > 0 {
		topLowestQuery = topLowestQuery.Where("co.id = ?", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		topLowestQuery = topLowestQuery.Where("c.program_id = ?", filter.ProgramId)
	}

	if filter.SchoolId > 0 {
		topLowestQuery = topLowestQuery.Where("schools.id = ?", filter.SchoolId)
	}

	if filter.ClassId > 0 {
		topLowestQuery = topLowestQuery.Where("uc.class_id = ?", filter.ClassId)
	}

	if filter.TeacherId > 0 {
		topLowestQuery = topLowestQuery.Joins(`JOIN user_courses uc2 ON uc2.course_id = co.id AND uc2.user_id = ?`, filter.TeacherId)
	}

	err = topLowestQuery.Group("exam_users.user_id, users.name, users.avatar_info, exam_users.exam_id, exam_users.id, e.name, users.name, schools.name, classes.name").
		Having("AVG(exam_users.ratio) < ?", 25).
		Order("avg_ratio ASC").
		Limit(10).
		Scan(&topLowestExam).Error

	if err != nil {
		return nil, err
	}

	topLowest := make([]dto.StudentScore, 0, len(topLowestExam))

	for _, student := range topLowestExam {
		topLowest = append(topLowest, dto.StudentScore{
			Id:         student.UserId,
			ExamId:     student.ExamId,
			ExamUserId: student.ExamUserId,
			Name:       student.UserName,
			ExamName:   student.ExamName,
			Score:      float32(student.AvgRatio),
			AvatarInfo: student.Avatar,
			ClassName:  student.ClassName,
			SchoolName: student.SchoolName,
		})
	}

	var averageScore struct {
		TotalAvgRatio  float64
		TotalUserCount int64
	}

	avgRatioQuery := db.Table("exam_users eu").
		Select(`
            COALESCE(AVG(eu.ratio), 0) AS total_avg_ratio,
            COUNT(DISTINCT eu.user_id) AS total_user_count
        `).
		Joins(`LEFT JOIN users u ON eu.user_id = u.id`).
		Joins(`JOIN exams e ON eu.exam_id = e.id`).
		Joins(`JOIN exam_ref_lessons erl ON erl.exam_id = e.id`).
		Joins(`LEFT JOIN lessons l ON erl.lesson_id = l.id`).
		Joins(`LEFT JOIN chapters c ON l.chapter_id = c.id`).
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`).
		Joins(`LEFT JOIN user_classes uc ON uc.user_id = u.id`).
		Joins(`LEFT JOIN schools ON schools.id = u.school_id`)

	if filterByDate {
		avgRatioQuery = avgRatioQuery.Where("eu.created_at <= ?", filter.EndTime)
	}

	if filter.CourseId > 0 {
		avgRatioQuery = avgRatioQuery.Where("co.id = ?", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		avgRatioQuery = avgRatioQuery.Where("c.program_id = ?", filter.ProgramId)
	}

	if filter.SchoolId > 0 {
		avgRatioQuery = avgRatioQuery.Where("u.school_id = ?", filter.SchoolId)
	}

	if filter.ClassId > 0 {
		avgRatioQuery = avgRatioQuery.Where("uc.class_id = ?", filter.ClassId)
	}

	if filter.TeacherId > 0 {
		avgRatioQuery = avgRatioQuery.Joins(`JOIN user_courses uc2 ON uc2.course_id = co.id AND uc2.user_id = ?`, filter.TeacherId)
	}

	if err := avgRatioQuery.Scan(&averageScore).Error; err != nil {
		return nil, err
	}

	var scoreCounts struct {
		HighRatioCount int64
		LowRatioCount  int64
		TotalCount     int64
	}

	subQuery := db.Table("exam_users eu").
		Select("eu.id AS exam_user_id, eu.ratio AS avg_ratio").
		Joins(`JOIN exams e ON eu.exam_id = e.id`).
		Joins(`JOIN exam_ref_lessons erl ON erl.exam_id = e.id`).
		Joins(`LEFT JOIN users ON eu.user_id = users.id`).
		Joins(`LEFT JOIN lessons l ON erl.lesson_id = l.id`).
		Joins(`LEFT JOIN chapters c ON l.chapter_id = c.id`).
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`).
		Joins(`LEFT JOIN user_classes uc ON uc.user_id = users.id`).
		Joins(`LEFT JOIN schools ON schools.id = users.school_id`)

	if filterByDate {
		subQuery = subQuery.Where("eu.created_at <= ?", filter.EndTime)
	}

	if filter.CourseId > 0 {
		subQuery = subQuery.Where("co.id = ?", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		subQuery = subQuery.Where("c.program_id = ?", filter.ProgramId)
	}

	if filter.SchoolId > 0 {
		subQuery = subQuery.Where("schools.id = ?", filter.SchoolId)
	}

	if filter.ClassId > 0 {
		subQuery = subQuery.Where("uc.class_id = ?", filter.ClassId)
	}

	if filter.TeacherId > 0 {
		subQuery = subQuery.Joins(`JOIN user_courses uc2 ON uc2.course_id = co.id AND uc2.user_id = ?`, filter.TeacherId)
	}

	ratioCountQuery := db.Table("(?) AS sub", subQuery).
		Select(`
			SUM(CASE WHEN sub.avg_ratio >= 75 THEN 1 ELSE 0 END) AS high_ratio_count,
			SUM(CASE WHEN sub.avg_ratio < 25 THEN 1 ELSE 0 END) AS low_ratio_count,
			COUNT(*) AS total_count
		`)

	if err := ratioCountQuery.Scan(&scoreCounts).Error; err != nil {
		return nil, err
	}

	totalAverageScore := float32(0)
	if averageScore.TotalAvgRatio > 0 {
		totalAverageScore = float32(averageScore.TotalAvgRatio)
	}

	highestPercent := float32(0)
	if scoreCounts.TotalCount > 0 && scoreCounts.HighRatioCount > 0 {
		highestPercent = float32(scoreCounts.HighRatioCount) / float32(scoreCounts.TotalCount) * 100
	}

	lowestPercent := float32(0)
	if scoreCounts.TotalCount > 0 && scoreCounts.LowRatioCount > 0 {
		lowestPercent = float32(scoreCounts.LowRatioCount) / float32(scoreCounts.TotalCount) * 100
	}

	behaviorPercent := float32(0)
	if highestPercent+lowestPercent > 0 {
		behaviorPercent = 100 - (highestPercent + lowestPercent)
	}

	return &dto.LearningOverviewDataItem{
		LearningScoreWeeks: learningScoreWeeks,
		GemsByWeeks:        gemsByWeeks,
		TopHighest:         topHighest,
		TopLowest:          topLowest,
		TotalGem:           0,
		TotalAverageScore:  totalAverageScore,
		HighestPercent:     highestPercent,
		LowestPercent:      lowestPercent,
		BehaviorPercent:    behaviorPercent,
	}, nil
}

func (r *dashboardRepository) GetScoreDistribution(filter dto.FilterDashboardAdmin) (*dto.ScoreDistributionOverview, error) {
	db := db.ReplicaDB

	// === STEP 1: Lấy danh sách khối ===
	type Grade struct {
		ID     int64
		Number int64
	}

	var grades []Grade
	err := db.Model(&models.Grade{}).
		Select("id, number").
		Order("number").
		Find(&grades).Error
	if err != nil {
		return nil, err
	}

	// === STEP 2: Lấy phân phối điểm GROUP BY bin + grade_id ===
	type BinResult struct {
		Bin     int64
		Count   int64
		GradeID int64
	}

	var results []BinResult

	query := db.Table("exam_users AS eu").
		Select(`FLOOR(eu.ratio / 10) * 10 AS bin, cl.grade_id, COUNT(*) AS count`).
		Joins(`JOIN exams e ON eu.exam_id = e.id`).
		Joins(`JOIN exam_ref_lessons erl ON erl.exam_id = e.id`).
		Joins(`JOIN lessons l ON erl.lesson_id = l.id`).
		Joins(`JOIN chapters c ON l.chapter_id = c.id`).
		Joins(`JOIN users u ON eu.user_id = u.id`).
		Joins(`JOIN user_classes uc ON uc.user_id = u.id`).
		Joins(`JOIN classes cl ON uc.class_id = cl.id`).
		Joins(`LEFT JOIN courses co ON co.program_id = c.program_id`)

	// === Filters ===
	if filter.CourseId > 0 {
		query = query.Where(`co.id = ?`, filter.CourseId)
	}
	if filter.ProgramId > 0 {
		query = query.Where(`c.program_id = ?`, filter.ProgramId)
	}
	if filter.SchoolId > 0 {
		query = query.Where("u.school_id = ?", filter.SchoolId)
	}
	if filter.ClassId > 0 {
		query = query.Joins(`JOIN user_classes uc2 ON uc2.user_id = u.id AND uc2.class_id = ?`, filter.ClassId)
	}
	if filter.TeacherId > 0 {
		query = query.Joins(`JOIN user_courses uc3 ON uc3.course_id = co.id AND uc3.user_id = ?`, filter.TeacherId)
	}

	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0
	if filterByDate {
		query = query.Where(`eu.created_at <= ?`, filter.EndTime)
		// query = query.Where(`eu.created_at >= ? AND eu.created_at <= ?`, filter.StartTime, filter.EndTime)
	}

	err = query.
		Group(`bin, cl.grade_id`).
		Order(`cl.grade_id, bin`).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// === STEP 3: Map bin theo grade ===
	resultMap := make(map[int64]map[int64]int64)
	for _, r := range results {
		if _, ok := resultMap[r.GradeID]; !ok {
			resultMap[r.GradeID] = make(map[int64]int64)
		}
		resultMap[r.GradeID][r.Bin] = r.Count
	}

	// === STEP 4: Build items theo khối ===
	min := int64(0)
	max := int64(100)
	step := int64(10)

	var items []dto.ScoreDistributionItem
	itemNumber := 1

	for _, g := range grades {
		bins := resultMap[g.ID]

		scores := []float32{}
		counts := []float32{}

		for b := min; b <= max; b += step {
			scores = append(scores, float32(b))
			counts = append(counts, float32(bins[b]))
		}

		items = append(items, dto.ScoreDistributionItem{
			Number: int32(g.Number),
			Scores: scores,
			Counts: counts,
		})

		itemNumber++
	}

	// === Trả về ===
	return &dto.ScoreDistributionOverview{
		Items: items,
	}, nil
}

func (r *dashboardRepository) GetSystemUsage(filter dto.FilterDashboardAdmin) (*dto.SystemUsageOverview, error) {
	db := db.ReplicaDB

	historyRepo := NewHistoryUseRepository()

	type DeviceUsageResult struct {
		Device  string
		Count   int64
		Percent float64
	}

	type WeeklyUsageResult struct {
		WeekLabel string
		Duration  float64
	}

	var weeklyResults []WeeklyUsageResult

	// Lấy danh sách bảng cần query theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", filter.StartTime, filter.EndTime)

	var unionQueries []string
	var args []interface{}
	for _, tableName := range tableNames {
		// Kiểm tra bảng có tồn tại không
		var exists bool
		err := db.Raw(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = ?
				)
			`, tableName).Scan(&exists).Error

		if err != nil || !exists {
			continue
		}

		// Tạo subquery cho từng bảng tháng
		subQuery := fmt.Sprintf(`
				SELECT
					CONCAT(EXTRACT('week' FROM al.created_at)::INT, '/', EXTRACT('isoyear' FROM al.created_at)::INT) AS week_label,
					COUNT(DISTINCT FLOOR(EXTRACT(EPOCH FROM al.created_at) / 300)) * 5 AS duration
				FROM %s al
			`, tableName)

		// Thêm JOIN clauses
		joinClause := ""
		whereClause := "WHERE al.created_at >= ? AND al.created_at <= ?"
		var queryArgs []interface{}
		queryArgs = append(queryArgs, filter.StartTime, filter.EndTime)

		baseJoin := " JOIN users ON users.id = al.user_id"
		joinClause = baseJoin

		if filter.SchoolId > 0 {
			whereClause += " AND users.school_id = ?"
			queryArgs = append(queryArgs, filter.SchoolId)
		}

		if filter.CourseId > 0 || filter.ProgramId > 0 {
			joinClause += " JOIN user_courses ON user_courses.user_id = al.user_id"
			if filter.CourseId > 0 {
				whereClause += " AND user_courses.course_id = ?"
				queryArgs = append(queryArgs, filter.CourseId)
			}
			if filter.ProgramId > 0 {
				joinClause += " JOIN courses ON courses.id = user_courses.course_id"
				whereClause += " AND courses.program_id = ?"
				queryArgs = append(queryArgs, filter.ProgramId)
			}
		}

		if filter.RoleId > 0 {
			if !strings.Contains(joinClause, "JOIN user_ref_roles") {
				joinClause += " JOIN user_ref_roles urr ON urr.user_id = users.id"
			}
			whereClause += " AND urr.role_id = ?"
			queryArgs = append(queryArgs, filter.RoleId)
		}

		args = append(args, queryArgs...)

		subQuery += joinClause + " " + whereClause + " GROUP BY week_label"
		unionQueries = append(unionQueries, subQuery)
	}

	if len(unionQueries) > 0 {
		// Gộp tất cả subqueries bằng UNION ALL và tính tổng
		mainQuery := fmt.Sprintf(`
				SELECT week_label, SUM(duration) as duration
				FROM (
					%s
				) AS combined
				GROUP BY week_label
				ORDER BY week_label ASC
			`, strings.Join(unionQueries, " UNION ALL "))

		err := db.Raw(mainQuery, args...).Scan(&weeklyResults).Error
		if err != nil {
			return nil, err
		}
	}

	var weeks []models.Week

	err := db.Raw(`
			SELECT week_number, start_date, end_date, year
			FROM weeks
			WHERE
				start_date >= ?
				AND end_date <= ?
				AND start_date < NOW()
			ORDER BY year, week_number
		`,
		filter.StartTime.AddDate(0, 0, -4),
		filter.EndTime.AddDate(0, 0, 4),
	).Scan(&weeks).Error

	if err != nil {
		return nil, err
	}

	usageMap := make(map[string]float64)
	for _, r := range weeklyResults {
		usageMap[r.WeekLabel] = r.Duration
	}

	weeklyUsages := make([]dto.WeeklyUsage, 0, len(weeks))
	for _, w := range weeks {
		weekLabel := fmt.Sprintf("%d/%d", w.WeekNumber, w.Year)

		duration := float32(usageMap[weekLabel])

		weeklyUsages = append(weeklyUsages, dto.WeeklyUsage{
			Week:     weekLabel,
			Duration: duration,
		})
	}

	// === Thiết bị từ logs ===
	var deviceResults []DeviceUsageResult

	// Tạo UNION query cho device từ các bảng theo tháng
	var deviceUnionQueries []string
	var deviceArgs []interface{}

	for _, tableName := range tableNames {
		// Kiểm tra bảng có tồn tại không
		var exists bool
		err := db.Raw(`
				SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = ?
				)
			`, tableName).Scan(&exists).Error

		if err != nil || !exists {
			continue
		}

		// Tạo subquery để lấy device của user từ mỗi bảng tháng
		subQuery := fmt.Sprintf(`
				SELECT DISTINCT ON (al.user_id)
					al.user_id,
					al.device
				FROM %s al
			`, tableName)

		// Thêm JOIN clauses
		joinClause := ""
		whereClause := "WHERE al.device IS NOT NULL AND al.created_at >= ? AND al.created_at <= ?"

		baseJoin := " JOIN users u ON u.id = al.user_id"
		joinClause = baseJoin

		if filter.SchoolId > 0 {
			whereClause += fmt.Sprintf(" AND u.school_id = %d", filter.SchoolId)
		}

		if filter.CourseId > 0 || filter.ProgramId > 0 {
			joinClause += " JOIN user_courses uc ON uc.user_id = al.user_id"
			if filter.CourseId > 0 {
				whereClause += fmt.Sprintf(" AND uc.course_id = %d", filter.CourseId)
			}
			if filter.ProgramId > 0 {
				joinClause += " JOIN courses co ON co.id = uc.course_id"
				whereClause += fmt.Sprintf(" AND co.program_id = %d", filter.ProgramId)
			}
		}

		if filter.RoleId > 0 {
			if !strings.Contains(joinClause, "JOIN user_ref_roles") {
				joinClause += " JOIN user_ref_roles urr ON urr.user_id = u.id"
			}
			whereClause += fmt.Sprintf(" AND urr.role_id = %d", filter.RoleId)
		}

		subQuery += joinClause + " " + whereClause + " ORDER BY al.user_id, al.created_at DESC"
		deviceUnionQueries = append(deviceUnionQueries, subQuery)
		deviceArgs = append(deviceArgs, filter.StartTime, filter.EndTime)
	}

	if len(deviceUnionQueries) > 0 {
		// Gộp kết quả từ tất cả các bảng tháng và tính phần trăm
		deviceMainQuery := fmt.Sprintf(`
				SELECT
					t.device AS device,
					COUNT(*) AS count,
					COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () AS percent
				FROM (
					SELECT DISTINCT ON (user_id) user_id, device
					FROM (
						%s
					) AS all_devices
					ORDER BY user_id, device
				) AS t
				GROUP BY t.device
			`, strings.Join(deviceUnionQueries, " UNION ALL "))

		err = db.Raw(deviceMainQuery, deviceArgs...).Scan(&deviceResults).Error
		if err != nil {
			return nil, err
		}
	}

	deviceMap := make(map[string]dto.DeviceUsage)
	for _, d := range deviceResults {
		deviceMap[d.Device] = dto.DeviceUsage{
			Name:    d.Device,
			Count:   int32(d.Count),
			Percent: float32(d.Percent),
		}
	}

	devices := []string{"Desktop", "Mobile", "Tablet"}
	deviceUsages := make([]dto.DeviceUsage, 0, len(devices))
	for _, device := range devices {
		if usage, ok := deviceMap[device]; ok {
			deviceUsages = append(deviceUsages, usage)
		} else {
			deviceUsages = append(deviceUsages, dto.DeviceUsage{
				Name:    device,
				Count:   0,
				Percent: 0,
			})
		}
	}

	averageUsed, _ := historyRepo.AverageUsed(filter.StartTime, filter.EndTime, filter.SchoolId, filter.CourseId, filter.ProgramId, filter.RoleId)

	return &dto.SystemUsageOverview{
		WeeklyUsages: weeklyUsages,
		DeviceUsages: deviceUsages,
		AverageUsed:  averageUsed,
	}, nil
}

func (r *dashboardRepository) GetQuestionBank(filter dto.FilterDashboardAdmin) (*dto.QuestionBankOverview, error) {
	db := db.ReplicaDB
	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	var totalQuestions int64
	if err := db.Model(&models.Question{}).Count(&totalQuestions).Error; err != nil {
		return nil, err
	}

	// Query tổng theo loại
	var typeResults []struct {
		QuestionType string
		Total        int32
	}

	questionQuery := db.
		Model(&models.Question{}).
		Select("question_type, COUNT(*) as total")

	if filterByDate {
		questionQuery = questionQuery.Where(`created_at <= ?`, filter.EndTime)
		// questionQuery = questionQuery.Where(`created_at >= ? AND created_at <= ?`, filter.StartTime, filter.EndTime)
	}

	if err := questionQuery.Group("question_type").Scan(&typeResults).Error; err != nil {
		return nil, err
	}

	var types []dto.QuestionTypeItem
	for _, t := range typeResults {
		percent := float32(t.Total) * 100 / float32(totalQuestions)
		types = append(types, dto.QuestionTypeItem{
			Type:    t.QuestionType,
			Total:   t.Total,
			Percent: percent,
		})
	}

	// Lấy toàn bộ question_attributes
	var allAttributes []struct {
		ID       int64
		ParentID *int64
		Name     string
	}

	if err := db.
		Table("question_attributes").
		Select("id, parent_id, name").
		Find(&allAttributes).Error; err != nil {
		return nil, err
	}

	// Đếm số question gắn attribute
	var countResults []struct {
		AttributeID int64
		Count       int32
	}

	countQuery := `
		SELECT qa.id AS attribute_id, COUNT(q.id) AS count
		FROM question_attributes qa
		INNER JOIN question_ref_attributes qra ON qra.question_attribute_id = qa.id
		INNER JOIN questions q ON q.id = qra.question_id AND q.deleted_at IS NULL
		WHERE qa.parent_id IS NOT NULL
	`

	var args []interface{}
	if filterByDate {
		countQuery += ` AND q.created_at <= ?`
		args = append(args, filter.EndTime)
		// countQuery += ` AND q.created_at >= ? AND q.created_at <= ?`
		// args = append(args, filter.StartTime, filter.EndTime)
	}

	countQuery += ` GROUP BY qa.id`

	if err := db.Raw(countQuery, args...).Scan(&countResults).Error; err != nil {
		return nil, err
	}

	// Map count vào
	countMap := make(map[int64]int32)
	for _, row := range countResults {
		countMap[row.AttributeID] = row.Count
	}

	// Tạo cây kết quả
	rootMap := map[int64]dto.QuestionAttributeItem{}
	childMap := map[int64][]dto.QuestionDistributionItem{}

	for _, attr := range allAttributes {
		count := countMap[attr.ID]
		percent := float32(count) * 100 / float32(totalQuestions)

		if attr.ParentID == nil {
			rootMap[attr.ID] = dto.QuestionAttributeItem{
				Name: attr.Name,
			}
		} else {
			childMap[*attr.ParentID] = append(childMap[*attr.ParentID], dto.QuestionDistributionItem{
				Name:    attr.Name,
				Count:   count,
				Percent: percent,
			})
		}
	}

	var attributes []dto.QuestionAttributeItem
	for id, root := range rootMap {
		root.Items = childMap[id]
		attributes = append(attributes, root)
	}

	// Media usage
	var withAudio, withImage int64

	db.Model(&models.Question{}).
		Where(`file_info->>'path' IS NOT NULL AND kind = ?`, "audio").
		Count(&withAudio)

	db.Model(&models.Question{}).
		Where(`file_info->>'path' IS NOT NULL AND kind = ?`, "image").
		Count(&withImage)

	mediaUsage := dto.MediaUsage{
		TotalQuestions: int32(totalQuestions),
		WithAudio:      int32(withAudio),
		WithImage:      int32(withImage),
	}

	if totalQuestions > 0 {
		mediaUsage.AudioPercent = float32(withAudio) * 100 / float32(totalQuestions)
		mediaUsage.ImagePercent = float32(withImage) * 100 / float32(totalQuestions)
	}

	// Final result
	return &dto.QuestionBankOverview{
		Types:      types,
		Attributes: attributes,
		MediaUsage: mediaUsage,
	}, nil
}

func (r *dashboardRepository) GetRiskWarning(filter dto.FilterDashboardAdmin) (*dto.RiskAndWarning, error) {
	db := db.ReplicaDB

	var (
		inactiveStudents    []dto.InactiveStudent
		slowGradingTeachers []dto.SlowGradingTeacher
		decliningStudents   []dto.DecliningStudent
	)

	// ====== Date condition ======
	var dateCondition string
	var dateArgs []interface{}
	if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
		dateCondition = " AND eu.created_at >= ? AND eu.created_at <= ? "
		dateArgs = append(dateArgs, filter.StartTime, filter.EndTime)
	}

	// Inactive Students
	whereInactive := "urr.role_id = ? AND u.deleted_at IS NULL"
	joinCourse := ""

	if filter.SchoolId > 0 {
		whereInactive += fmt.Sprintf(" AND u.school_id = %d", filter.SchoolId)
	}
	if filter.CourseId > 0 {
		joinCourse = "LEFT JOIN user_courses uco ON uco.user_id = u.id"
		whereInactive += fmt.Sprintf(" AND uco.course_id = %d", filter.CourseId)
	} else if filter.ProgramId > 0 {
		joinCourse = "LEFT JOIN user_courses uco ON uco.user_id = u.id LEFT JOIN courses co ON co.id = uco.course_id"
		whereInactive += fmt.Sprintf(" AND co.program_id = %d", filter.ProgramId)
	}

	inactiveSQL := fmt.Sprintf(`
		SELECT
			u.id,
			u.name,
			u.avatar_info,
			c.name AS class,
			u.last_login_at AS last_login,
			DATE_PART('day', NOW() - COALESCE(u.last_login_at, u.created_at)) AS absent_days
		FROM users u
		LEFT JOIN user_classes uc ON uc.user_id = u.id
		LEFT JOIN classes c ON uc.class_id = c.id
		JOIN user_ref_roles urr ON urr.user_id = u.id
		%s
		WHERE %s
		GROUP BY u.id, u.name, u.avatar_info, c.name, u.created_at, u.last_login_at
		HAVING DATE_PART('day', NOW() - COALESCE(u.last_login_at, u.created_at)) >= 7
		ORDER BY absent_days DESC
		LIMIT 20
	`, joinCourse, whereInactive)

	if err := db.Raw(inactiveSQL, models.StudentRoleId).Scan(&inactiveStudents).Error; err != nil {
		return nil, err
	}

	// Declining Students
	whereDeclining := "1=1"
	joinCourseDeclining := ""
	if filter.SchoolId > 0 {
		whereDeclining += fmt.Sprintf(" AND u.school_id = %d", filter.SchoolId)
	}
	if filter.CourseId > 0 {
		whereDeclining += fmt.Sprintf(" AND uco.course_id = %d", filter.CourseId)
	} else if filter.ProgramId > 0 {
		joinCourseDeclining = "LEFT JOIN courses co2 ON co2.id = uco.course_id"
		whereDeclining += fmt.Sprintf(" AND co2.program_id = %d", filter.ProgramId)
	}

	decliningSQL := fmt.Sprintf(`
		WITH scored AS (
			SELECT
				eu.user_id,
				eu.id,
				eu.ratio,
				LAG(eu.ratio) OVER (PARTITION BY eu.user_id ORDER BY eu.created_at) AS prev_ratio
			FROM exam_users eu
			JOIN users u ON u.id = eu.user_id
			LEFT JOIN user_courses uco ON uco.user_id = u.id
			%s
			WHERE %s %s
		),
		filtered AS (
			SELECT user_id
			FROM scored
			GROUP BY user_id
			HAVING COUNT(*) >= 3 AND BOOL_AND(prev_ratio IS NULL OR ratio <= prev_ratio)
		)
		SELECT
			u.id,
			u.name,
			u.avatar_info,
			c.name AS class,
			json_agg(
				json_build_object('id', eu.id, 'score', eu.ratio)
				ORDER BY eu.created_at
			) AS score
		FROM users u
		JOIN exam_users eu ON eu.user_id = u.id
		JOIN user_ref_roles urr ON urr.user_id = u.id
		LEFT JOIN user_classes uc2 ON uc2.user_id = u.id
		LEFT JOIN classes c ON c.id = uc2.class_id
		WHERE u.id IN (SELECT user_id FROM filtered)
			AND urr.role_id = ?
			AND u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.avatar_info, c.name
		ORDER BY u.id
		LIMIT 20
	`, joinCourseDeclining, whereDeclining, dateCondition)

	decliningArgs := append(dateArgs, models.StudentRoleId)
	if err := db.Raw(decliningSQL, decliningArgs...).Scan(&decliningStudents).Error; err != nil {
		return nil, err
	}

	// Slow Grading Teachers
	whereSlow := "r.id = ? AND ut.deleted_at IS NULL"
	if filter.SchoolId > 0 {
		whereSlow += fmt.Sprintf(" AND ut.school_id = %d", filter.SchoolId)
	}
	if filter.CourseId > 0 {
		whereSlow += fmt.Sprintf(" AND co.id = %d", filter.CourseId)
	}
	if filter.ProgramId > 0 {
		whereSlow += fmt.Sprintf(" AND ch.program_id = %d", filter.ProgramId)
	}
	if filter.TeacherId > 0 {
		whereSlow += fmt.Sprintf(" AND ut.id = %d", filter.TeacherId)
	}

	slowGradingSQL := fmt.Sprintf(`
		WITH submissions AS (
			SELECT
				eu.exam_id,
				eu.user_id AS student_id,
				uct.user_id AS teacher_id
			FROM exam_users eu
			JOIN exams e ON e.id = eu.exam_id
			JOIN exam_ref_lessons erl ON erl.exam_id = e.id
			JOIN lessons l ON l.id = erl.lesson_id
			JOIN chapters ch ON ch.id = l.chapter_id
			JOIN courses co ON co.program_id = ch.program_id
			JOIN user_courses uct ON uct.course_id = co.id AND uct.main_teacher = TRUE
			JOIN users ut ON ut.id = uct.user_id
			JOIN user_ref_roles urr ON urr.user_id = ut.id
			JOIN roles r ON r.id = urr.role_id
			WHERE %s %s
		),
		grading AS (
			SELECT
				eq.exam_id,
				eq.user_id AS student_id,
				MIN(eq.created_at) AS submitted_at,
				MIN(eq.scoring_at) AS first_scored_at
			FROM exam_question_user_manual_scoring eq
			GROUP BY eq.exam_id, eq.user_id
		)
		SELECT
			s.teacher_id,
			ut.name AS teacher_name,
			ut.avatar_info,
			COUNT(DISTINCT s.exam_id || '-' || s.student_id) AS total_submissions,
			COUNT(DISTINCT CASE WHEN g.first_scored_at IS NOT NULL THEN s.exam_id || '-' || s.student_id END) AS graded_submissions,
			ROUND(EXTRACT(EPOCH FROM AVG(COALESCE(g.first_scored_at, NOW()) - g.submitted_at)) / 3600, 2) AS avg_wait_hours
		FROM submissions s
		JOIN grading g ON s.exam_id = g.exam_id AND s.student_id = g.student_id
		JOIN users ut ON ut.id = s.teacher_id
		GROUP BY s.teacher_id, ut.name, ut.avatar_info
		ORDER BY avg_wait_hours DESC
		LIMIT 20
	`, whereSlow, dateCondition)

	slowGradingArgs := append([]interface{}{models.TeacherRoleId}, dateArgs...)
	if err := db.Raw(slowGradingSQL, slowGradingArgs...).Scan(&slowGradingTeachers).Error; err != nil {
		return nil, err
	}

	return &dto.RiskAndWarning{
		InactiveStudents:    inactiveStudents,
		DecliningStudents:   decliningStudents,
		SlowGradingTeachers: slowGradingTeachers,
	}, nil
}

func (r *dashboardRepository) GetTeacherPerformance(filter dto.FilterDashboardAdmin) (*dto.TeacherPerformanceOverview, error) {
	db := db.ReplicaDB
	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	// Determine which types to include
	includeExam := filter.GradingType == "" || filter.GradingType == "all" || filter.GradingType == "exam"
	includeHomework := filter.GradingType == "" || filter.GradingType == "all" || filter.GradingType == "homework"
	includeExercise := filter.GradingType == "" || filter.GradingType == "all" || filter.GradingType == "exercise"

	var (
		weeklyMarkingRates       []dto.WeeklyMarkingRate
		weeklyMarkings           []dto.WeeklyMarking
		ungradedSubmissions      []dto.UngradedSubmission
		totalSubmissions         int32
		totalUngradedSubmissions int32
	)

	// --- Weekly marking ---
	var weekArgs []interface{}
	var unionParts []string

	whereClause := ""
	if filterByDate {
		whereClause = " AND w.end_date >= ? AND w.end_date <= ?"
	} else {
		whereClause = " AND w.end_date <= NOW()"
	}

	// Exam weekly marking
	if includeExam {
		examWeekSQL := `
			SELECT
				w.id,
				w.week_number,
				w.start_date,
				w.end_date,
				COUNT(DISTINCT eu.id) AS assigned_count,
				COUNT(DISTINCT CASE
					WHEN equms.scoring_by IS NOT NULL THEN equms.exam_id || '-' || equms.user_id
				END) AS marked_count
			FROM weeks w
			LEFT JOIN exam_users eu
				ON eu.created_at >= w.start_date AND eu.created_at < w.end_date + INTERVAL '1 day'
			LEFT JOIN exam_question_user_manual_scoring equms
				ON equms.created_at >= w.start_date AND equms.created_at < w.end_date + INTERVAL '1 day'
				AND equms.exam_id = eu.exam_id AND equms.user_id = eu.user_id
			WHERE 1=1` + whereClause + `
			GROUP BY w.id, w.week_number, w.start_date, w.end_date`
		unionParts = append(unionParts, examWeekSQL)
		if filterByDate {
			weekArgs = append(weekArgs, filter.StartTime, filter.EndTime)
		}
	}

	// Homework weekly marking
	if includeHomework {
		homeworkWeekSQL := `
			SELECT
				w.id,
				w.week_number,
				w.start_date,
				w.end_date,
				COUNT(DISTINCT hu.id) AS assigned_count,
				COUNT(DISTINCT CASE
					WHEN hqums.scoring_by IS NOT NULL THEN hqums.homework_id || '-' || hqums.user_id
				END) AS marked_count
			FROM weeks w
			LEFT JOIN homework_users hu
				ON hu.created_at >= w.start_date AND hu.created_at < w.end_date + INTERVAL '1 day'
			LEFT JOIN homework_question_user_manual_scoring hqums
				ON hqums.created_at >= w.start_date AND hqums.created_at < w.end_date + INTERVAL '1 day'
				AND hqums.homework_id = hu.homework_id AND hqums.user_id = hu.user_id
			WHERE 1=1` + whereClause + `
			GROUP BY w.id, w.week_number, w.start_date, w.end_date`
		unionParts = append(unionParts, homeworkWeekSQL)
		if filterByDate {
			weekArgs = append(weekArgs, filter.StartTime, filter.EndTime)
		}
	}

	// Exercise weekly marking
	if includeExercise {
		exerciseWeekSQL := `
			SELECT
				w.id,
				w.week_number,
				w.start_date,
				w.end_date,
				COUNT(DISTINCT eu.id) AS assigned_count,
				COUNT(DISTINCT CASE
					WHEN equms.scoring_by IS NOT NULL THEN equms.exercise_id || '-' || equms.user_id
				END) AS marked_count
			FROM weeks w
			LEFT JOIN exercise_users eu
				ON eu.created_at >= w.start_date AND eu.created_at < w.end_date + INTERVAL '1 day'
			LEFT JOIN exercise_question_user_manual_scoring equms
				ON equms.created_at >= w.start_date AND equms.created_at < w.end_date + INTERVAL '1 day'
				AND equms.exercise_id = eu.exercise_id AND equms.user_id = eu.user_id
			WHERE 1=1` + whereClause + `
			GROUP BY w.id, w.week_number, w.start_date, w.end_date`
		unionParts = append(unionParts, exerciseWeekSQL)
		if filterByDate {
			weekArgs = append(weekArgs, filter.StartTime, filter.EndTime)
		}
	}

	if len(unionParts) == 0 {
		// No types selected
		weeklyMarkings = []dto.WeeklyMarking{}
	} else {
		// Combine all parts with UNION ALL and aggregate
		weekSQL := `
			SELECT
				id,
				week_number,
				start_date,
				end_date,
				SUM(assigned_count)::bigint AS assigned_count,
				SUM(marked_count)::bigint AS marked_count
			FROM (
		` + strings.Join(unionParts, " UNION ALL ") + `
			) AS combined
			GROUP BY id, week_number, start_date, end_date
			ORDER BY start_date DESC`

		if err := db.Raw(weekSQL, weekArgs...).Scan(&weeklyMarkings).Error; err != nil {
			return nil, err
		}
	}

	// --- Ungraded submissions ---
	// Build ungraded submissions query with UNION ALL for all types
	var ungradedUnionParts []string
	var ungradedArgs []interface{}

	if includeExam {
		examDateFilter := ""
		if filterByDate {
			examDateFilter = " AND eq.created_at >= ? AND eq.created_at <= ?"
			ungradedArgs = append(ungradedArgs, filter.StartTime, filter.EndTime)
		}
		examSQL := `WITH ungraded_exam AS (
				SELECT eq.exam_id AS id_value, 'exam' AS type_value, eq.user_id AS student_id,
					eq.question_id, eq.created_at
				FROM exam_question_user_manual_scoring eq
				WHERE eq.scoring_by IS NULL
					AND NOT EXISTS (SELECT 1 FROM exam_question_user_manual_scoring eq2
						WHERE eq2.exam_id = eq.exam_id AND eq2.question_id = eq.question_id
						AND eq2.user_id = eq.user_id AND eq2.scoring_by IS NOT NULL)` + examDateFilter + `
			), agg_exam AS (
				SELECT id_value, type_value, student_id, ARRAY_AGG(question_id) AS question_ids,
					COUNT(*) AS question_count, MIN(created_at) AS submitted_at
				FROM ungraded_exam GROUP BY id_value, type_value, student_id
			), exam_course AS (
				SELECT e.id AS exam_id, e.name AS exam_name, l.id AS lesson_id,
					c.id AS chapter_id, c.program_id
				FROM exams e
				JOIN exam_ref_lessons erl ON erl.exam_id = e.id
				JOIN lessons l ON l.id = erl.lesson_id
				JOIN chapters c ON c.id = l.chapter_id
			)
			SELECT 'exam' AS type, u.id AS teacher_id, u.name AS teacher_name,
				s.id AS student_id, s.name AS student_name, a.id_value AS id,
				ec.exam_name AS name, a.question_ids, a.question_count, a.submitted_at
			FROM agg_exam a
			JOIN exam_course ec ON a.id_value = ec.exam_id
			JOIN courses co ON co.program_id = ec.program_id
			JOIN user_courses uc ON uc.course_id = co.id AND uc.main_teacher = TRUE
			JOIN users u ON u.id = uc.user_id
			JOIN user_ref_roles urr ON urr.user_id = u.id
			JOIN roles r ON r.id = urr.role_id
			JOIN users s ON s.id = a.student_id
			WHERE r.id = ?`
		ungradedUnionParts = append(ungradedUnionParts, examSQL)
		ungradedArgs = append(ungradedArgs, models.TeacherRoleId)
	}

	if includeHomework {
		homeworkDateFilter := ""
		if filterByDate {
			homeworkDateFilter = " AND hq.created_at >= ? AND hq.created_at <= ?"
			ungradedArgs = append(ungradedArgs, filter.StartTime, filter.EndTime)
		}
		homeworkSQL := `WITH ungraded_homework AS (
				SELECT hq.homework_id AS id_value, 'homework' AS type_value, hq.user_id AS student_id,
					hq.question_id, hq.created_at
				FROM homework_question_user_manual_scoring hq
				WHERE hq.scoring_by IS NULL
					AND NOT EXISTS (SELECT 1 FROM homework_question_user_manual_scoring hq2
						WHERE hq2.homework_id = hq.homework_id AND hq2.question_id = hq.question_id
						AND hq2.user_id = hq.user_id AND hq2.scoring_by IS NOT NULL)` + homeworkDateFilter + `
			), agg_homework AS (
				SELECT id_value, type_value, student_id, ARRAY_AGG(question_id) AS question_ids,
					COUNT(*) AS question_count, MIN(created_at) AS submitted_at
				FROM ungraded_homework GROUP BY id_value, type_value, student_id
			), homework_course AS (
				SELECT h.id AS homework_id, h.name AS homework_name, l.id AS lesson_id,
					c.id AS chapter_id, c.program_id
				FROM homeworks h
				JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id
				JOIN lessons l ON l.id = hrl.lesson_id
				JOIN chapters c ON c.id = l.chapter_id
			)
			SELECT 'homework' AS type, u.id AS teacher_id, u.name AS teacher_name,
				s.id AS student_id, s.name AS student_name, a.id_value AS id,
				hc.homework_name AS name, a.question_ids, a.question_count, a.submitted_at
			FROM agg_homework a
			JOIN homework_course hc ON a.id_value = hc.homework_id
			JOIN courses co ON co.program_id = hc.program_id
			JOIN user_courses uc ON uc.course_id = co.id AND uc.main_teacher = TRUE
			JOIN users u ON u.id = uc.user_id
			JOIN user_ref_roles urr ON urr.user_id = u.id
			JOIN roles r ON r.id = urr.role_id
			JOIN users s ON s.id = a.student_id
			WHERE r.id = ?`
		ungradedUnionParts = append(ungradedUnionParts, homeworkSQL)
		ungradedArgs = append(ungradedArgs, models.TeacherRoleId)
	}

	if includeExercise {
		exerciseDateFilter := ""
		if filterByDate {
			exerciseDateFilter = " AND eq.created_at >= ? AND eq.created_at <= ?"
			ungradedArgs = append(ungradedArgs, filter.StartTime, filter.EndTime)
		}
		exerciseSQL := `WITH ungraded_exercise AS (
				SELECT eq.exercise_id AS id_value, 'exercise' AS type_value, eq.user_id AS student_id,
					eq.question_id, eq.created_at
				FROM exercise_question_user_manual_scoring eq
				WHERE eq.scoring_by IS NULL
					AND NOT EXISTS (SELECT 1 FROM exercise_question_user_manual_scoring eq2
						WHERE eq2.exercise_id = eq.exercise_id AND eq2.question_id = eq.question_id
						AND eq2.user_id = eq.user_id AND eq2.scoring_by IS NOT NULL)` + exerciseDateFilter + `
			), agg_exercise AS (
				SELECT id_value, type_value, student_id, ARRAY_AGG(question_id) AS question_ids,
					COUNT(*) AS question_count, MIN(created_at) AS submitted_at
				FROM ungraded_exercise GROUP BY id_value, type_value, student_id
			), exercise_course AS (
				SELECT e.id AS exercise_id, e.name AS exercise_name, l.id AS lesson_id,
					c.id AS chapter_id, c.program_id
				FROM exercises e
				JOIN exercise_ref_lessons erl ON erl.exercise_id = e.id
				JOIN lessons l ON l.id = erl.lesson_id
				JOIN chapters c ON c.id = l.chapter_id
			)
			SELECT 'exercise' AS type, u.id AS teacher_id, u.name AS teacher_name,
				s.id AS student_id, s.name AS student_name, a.id_value AS id,
				ec.exercise_name AS name, a.question_ids, a.question_count, a.submitted_at
			FROM agg_exercise a
			JOIN exercise_course ec ON a.id_value = ec.exercise_id
			JOIN courses co ON co.program_id = ec.program_id
			JOIN user_courses uc ON uc.course_id = co.id AND uc.main_teacher = TRUE
			JOIN users u ON u.id = uc.user_id
			JOIN user_ref_roles urr ON urr.user_id = u.id
			JOIN roles r ON r.id = urr.role_id
			JOIN users s ON s.id = a.student_id
			WHERE r.id = ?`
		ungradedUnionParts = append(ungradedUnionParts, exerciseSQL)
		ungradedArgs = append(ungradedArgs, models.TeacherRoleId)
	}

	if len(ungradedUnionParts) == 0 {
		ungradedSubmissions = []dto.UngradedSubmission{}
	} else {
		// Wrap each CTE query in a subquery for UNION ALL compatibility
		wrappedParts := make([]string, len(ungradedUnionParts))
		for i, part := range ungradedUnionParts {
			wrappedParts[i] = "SELECT * FROM (" + part + ") AS subquery_" + strconv.Itoa(i)
		}
		ungradedSQL := strings.Join(wrappedParts, " UNION ALL ") + " ORDER BY teacher_id, type, id, student_id"
		if err := db.Raw(ungradedSQL, ungradedArgs...).Scan(&ungradedSubmissions).Error; err != nil {
			return nil, err
		}
	}

	// --- Merge weekly ---
	beginAt, _ := time.Parse("2006-01-02", "2025-09-08")
	for _, week := range weeklyMarkings {
		ungradedSubmissionByWeeks := []dto.UngradedSubmission{}
		startDate, _ := time.Parse(time.RFC3339, week.StartDate)
		endDate, _ := time.Parse(time.RFC3339, week.EndDate)

		if startDate.Before(beginAt) {
			continue
		}

		markingPercent := float32(100)
		if week.AssignedCount > 0 {
			markingPercent = float32(week.AssignedCount - week.MarkedCount) * 100 / float32(week.AssignedCount)
		}

		// Use map to track unique submissions by Type and Id
		seenMap := make(map[string]dto.UngradedSubmission)
		for _, u := range ungradedSubmissions {
			submittedAt, _ := time.Parse(time.RFC3339, u.SubmittedAt)
			if submittedAt.After(startDate) && submittedAt.Before(endDate.AddDate(0, 0, 1)) {
				// Create unique key from Type and Id
				uniqueKey := fmt.Sprintf("%s_%d", u.Type, u.Id)
				// Only add if not already seen
				if _, exists := seenMap[uniqueKey]; !exists {
					seenMap[uniqueKey] = u
				}
			}
		}
		// Convert map values to slice
		for _, u := range seenMap {
			ungradedSubmissionByWeeks = append(ungradedSubmissionByWeeks, u)
		}

		diff := startDate.Sub(beginAt)
		weekNumber := int(diff.Hours() / (24 * 7))
		if weekNumber < 0 {
			weekNumber = 0
		}
		weekNumber++

		totalSubmissions += week.AssignedCount
		totalUngradedSubmissions += int32(len(ungradedSubmissionByWeeks))

		weeklyMarkingRates = append(weeklyMarkingRates, dto.WeeklyMarkingRate{
			Id:                  week.Id,
			WeekNumber:          strconv.Itoa(weekNumber),
			MarkedCount:         week.AssignedCount - week.MarkedCount,
			AssignedCount:       week.AssignedCount,
			MarkingPercent:      markingPercent,
			UngradedSubmissions: ungradedSubmissionByWeeks,
		})
	}

	gradedSubmissions := totalSubmissions - totalUngradedSubmissions

	if gradedSubmissions < 0 {
		gradedSubmissions = 0
	}

	ungradedPercent := float32(0)
	if totalSubmissions > 0 {
		ungradedPercent = float32(totalUngradedSubmissions) * 100 / float32(totalSubmissions)
	}

	gradingSummary := dto.GradingSummary{
		TotalSubmissions:    totalSubmissions,
		GradedSubmissions:   gradedSubmissions,
		UngradedSubmissions: totalUngradedSubmissions,
		UngradedPercent:     ungradedPercent,
	}

	return &dto.TeacherPerformanceOverview{
		GradingSummary:     gradingSummary,
		WeeklyMarkingRates: weeklyMarkingRates,
	}, nil
}

func (r *dashboardRepository) GetUserOverview(filter dto.FilterDashboardAdmin) (*dto.UserOverview, error) {
	// Use single SQL function to get most data at once
	var result models.DashboardOverviewResult

	// Convert to UTC for database
	endTimeUTC := filter.EndTime.UTC()
	startPreviousTimeUTC := filter.StartPreviousTime.UTC()
	endPreviousTimeUTC := filter.EndPreviousTime.UTC()
	startTimeUTC := filter.StartTime.UTC()

	// Use raw SQL with string interpolation instead of prepared statement
	query := fmt.Sprintf(`
		SELECT * FROM get_dashboard_overview(
			'%s'::timestamp,
			'%s'::timestamp,
			'%s'::timestamp,
			'%s'::timestamp,
			'%s'::timestamp,
			%s,
			%s,
			%d,
			%d
		)`,
		endTimeUTC.Format("2006-01-02 15:04:05"),
		startPreviousTimeUTC.Format("2006-01-02 15:04:05"),
		endPreviousTimeUTC.Format("2006-01-02 15:04:05"),
		endPreviousTimeUTC.Format("2006-01-02 15:04:05"),
		startTimeUTC.Format("2006-01-02 15:04:05"),
		getSQLNullValue(filter.Address.ProvinceCode),
		getSQLNullValue(filter.Address.WardCode),
		filter.CourseId,
		filter.SchoolId,
	)

	err := db.MasterDB.Raw(query).Scan(&result).Error

	if err != nil {
		config.Log.Error("Error calling get_dashboard_overview:", err)
		// Fallback to individual methods if SQL function fails
		config.Log.Info("Falling back to individual methods...")
		return r.getUserOverviewFallback(filter)
	}

	// If all values are 0, try fallback method
	if result.SchoolTotalCount == 0 && result.UserTotalCount == 0 && result.StudentTotalCount == 0 && result.TeacherTotalCount == 0 {
		config.Log.Info("SQL function returned all zeros, trying fallback method...")
		return r.getUserOverviewFallback(filter)
	}

	// Get activity data separately using the existing method
	activityTotal, activityCurrent, _, activityNew, err := r.GetActivity(filter)
	if err != nil {
		return nil, err
	}

	return &dto.UserOverview{
		User: &dto.DashboardAdminItem{
			Total:       int32(result.UserTotalCount),
			TotalNumber: int32(result.UserCurrentCount),
			ChangeValue: float32(result.UserCurrentCount - result.UserPreviousCount),
		},
		Student: &dto.DashboardAdminItem{
			Total:       int32(result.StudentCurrentCount),
			TotalNumber: int32(result.StudentPreviousCount),
			ChangeValue: float32(result.StudentCurrentCount - result.StudentPreviousCount),
		},
		Teacher: &dto.DashboardAdminItem{
			Total:       int32(result.TeacherCurrentCount),
			TotalNumber: int32(result.TeacherPreviousCount),
			ChangeValue: float32(result.TeacherCurrentCount - result.TeacherPreviousCount),
		},
		Activity: &dto.DashboardAdminItem{
			Total:       activityTotal,
			TotalNumber: activityCurrent,
			ChangeValue: float32(activityNew),
		},
		School: &dto.DashboardAdminItem{
			Total:       int32(result.SchoolTotalCount),
			TotalNumber: int32(result.SchoolPreviousCount),
			ChangeValue: float32(result.SchoolCurrentCount),
		},
	}, nil
}

func (r *dashboardRepository) Percent(count, all int64) float32 {
	if all == 0 {
		return 0
	}
	return float32(count) * 100 / float32(all)
}
