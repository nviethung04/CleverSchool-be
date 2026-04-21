package repositories

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/models"
	"be-lms/table_manager"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
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
	GetSystemUsageType(filter dto.FilterDashboardAdmin) (*dto.SystemUsageOverview, error)
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

	// === Count (chạy song song) ===
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		setErr(currentQuery.Count(&currentCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(previousQuery.Count(&previousCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(totalQuery.Count(&totalCount).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return 0, 0, 0, firstErr
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
		Where("users.created_at <= ?", filter.EndPreviousTime)

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

	// Count chạy song song
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		setErr(currentQuery.Distinct("users.id").Count(&currentCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(previousQuery.Distinct("users.id").Count(&previousCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(totalQuery.Distinct("users.id").Count(&totalCount).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return 0, 0, 0, firstErr
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

	// Count chạy song song
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		setErr(currentQuery.Distinct("users.id").Count(&currentCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(previousQuery.Distinct("users.id").Count(&previousCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(totalQuery.Distinct("users.id").Count(&totalCount).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return 0, 0, 0, firstErr
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

	// Count chạy song song
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		setErr(currentQuery.Distinct("users.id").Count(&currentCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(previousQuery.Distinct("users.id").Count(&previousCount).Error)
	}()
	go func() {
		defer wg.Done()
		setErr(totalQuery.Distinct("users.id").Count(&totalCount).Error)
	}()
	wg.Wait()
	if firstErr != nil {
		return 0, 0, 0, firstErr
	}
	return int32(totalCount), int32(currentCount), int32(previousCount), nil
}

func (r *dashboardRepository) GetActivityUserIds(startTime, endTime time.Time, courseId, schoolId, programId int64) (models.ActivityIds, error) {
	type monthTask struct {
		monthStart     time.Time
		monthEnd       time.Time
		isCurrentMonth bool
	}
	var tasks []monthTask
	current := startTime
	now := time.Now()
	for current.Before(endTime) || current.Equal(endTime) {
		monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
		monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
		if monthEnd.After(endTime) {
			monthEnd = endTime
		}
		isCurrentMonth := now.Year() == current.Year() && now.Month() == current.Month()
		tasks = append(tasks, monthTask{monthStart, monthEnd, isCurrentMonth})
		current = monthStart.AddDate(0, 1, 0)
	}

	var allUserIDs []int64
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	historyUseRepo := NewHistoryUseRepository()
	wg := sync.WaitGroup{}
	for _, task := range tasks {
		wg.Add(1)
		go func(monthStart, monthEnd time.Time, isCurrentMonth bool) {
			defer wg.Done()
			var userIDs []int64
			if isCurrentMonth {
				idsRes, err := historyUseRepo.ActivityIds(monthStart, monthEnd)
				if err != nil {
					setErr(err)
					return
				}
				userIDs = idsRes.Ids
				var historyUses []models.HistoryUse
				if err := db.ReplicaDB.Where("type = ?", "month").Find(&historyUses).Error; err != nil {
					config.Log.Error(err)
					setErr(err)
					return
				}
				for _, historyUse := range historyUses {
					var activityIds models.ActivityIds
					if err := json.Unmarshal(historyUse.ActivityIds, &activityIds); err != nil {
						config.Log.Error(err)
						setErr(err)
						return
					}
					userIDs = append(userIDs, activityIds.Ids...)
				}
			} else {
				var historyUse models.HistoryUse
				if err := db.ReplicaDB.Where("date = ? AND type = ?", monthStart.Format("2006-01-02"), "month").First(&historyUse).Error; err != nil {
					setErr(err)
					return
				}
				var activityIds models.ActivityIds
				if err := json.Unmarshal(historyUse.ActivityIds, &activityIds); err != nil {
					setErr(err)
					return
				}
				userIDs = activityIds.Ids
			}
			mu.Lock()
			allUserIDs = append(allUserIDs, userIDs...)
			mu.Unlock()
		}(task.monthStart, task.monthEnd, task.isCurrentMonth)
	}
	wg.Wait()
	if firstErr != nil {
		return models.ActivityIds{}, firstErr
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
	var currentUserIds, previousIds, totalIds models.ActivityIds
	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		ids, err := r.GetActivityUserIds(filter.StartTime, filter.EndTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		currentUserIds = ids
		mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		ids, err := r.GetActivityUserIds(filter.StartPreviousTime, filter.EndPreviousTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		previousIds = ids
		mu.Unlock()
	}()
	go func() {
		defer wg.Done()
		ids, err := r.GetActivityUserIds(filter.StartTime, filter.EndPreviousTime, filter.CourseId, filter.SchoolId, filter.ProgramId)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		totalIds = ids
		mu.Unlock()
	}()
	wg.Wait()
	if firstErr != nil {
		return 0, 0, 0, 0, firstErr
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

	database := db.ReplicaDB

	// === Program (phải chạy trước để có programId) ===
	var programId int64
	programQuery := database.Model(&models.Program{})

	if filter.CourseId > 0 {
		programQuery = programQuery.
			Joins("JOIN courses co ON co.program_id = programs.id").
			Where("co.id = ?", filter.CourseId)

		database.Model(&models.Course{}).
			Select("program_id").
			Where("id = ?", filter.CourseId).
			Scan(&programId)
	} else if filter.ProgramId > 0 {
		programId = filter.ProgramId
		programQuery = programQuery.Where("programs.id = ?", filter.ProgramId)
	} else if filter.SchoolId > 0 {
		programQuery = programQuery.
			Joins("JOIN courses co ON co.program_id = programs.id").
			Joins("JOIN course_schools cs ON cs.course_id = co.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}

	if filterByDate {
		programQuery = programQuery.Where("programs.created_at <= ?", filter.EndTime)
	}
	programQuery.Distinct("programs.id").Count(&programCount)

	// === Chạy song song các query chỉ phụ thuộc programId / filter ===
	var questionIDs []QuestionID
	var firstErr error
	var mu sync.Mutex
	wg := &sync.WaitGroup{}

	// Lesson
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Model(&models.Lesson{}).
			Joins("JOIN chapters AS c ON lessons.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("lessons.created_at <= ?", filter.EndTime)
		}
		err := q.Distinct("lessons.id").Count(&lessonCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Lesson Plan
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Table("lesson_plans AS lp")
		if programId > 0 || filter.SchoolId > 0 {
			q = q.
				Joins("JOIN lesson_plan_ref_lessons AS lprl ON lp.id = lprl.lesson_plan_id").
				Joins("JOIN lessons AS l ON lprl.lesson_id = l.id").
				Joins("JOIN chapters AS c ON l.chapter_id = c.id").
				Joins("JOIN programs p ON c.program_id = p.id")
		}
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("lp.created_at <= ?", filter.EndTime)
		}
		err := q.Where("lp.deleted_at IS NULL").Distinct("lp.id").Count(&lessonPlanCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Question
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Model(&models.Question{})
		if filterByDate {
			q = q.Where("created_at <= ?", filter.EndTime)
		}
		err := q.Count(&questionCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Homework
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Table("homeworks AS h").
			Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = h.id").
			Joins("JOIN lessons AS l ON l.id = hrl.lesson_id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("h.created_at <= ?", filter.EndTime)
		}
		err := q.Where("h.deleted_at IS NULL").Select("COUNT(DISTINCT h.id)").Count(&homeworkCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Exam
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Table("exams AS e").
			Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
			Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("e.created_at <= ?", filter.EndTime)
		}
		err := q.Where("e.deleted_at IS NULL").Select("COUNT(DISTINCT e.id)").Count(&examCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// User Exam
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Table("exam_users AS eu").
			Joins("JOIN exams AS e ON eu.exam_id = e.id").
			Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
			Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id").
			Where("e.deleted_at IS NULL")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("eu.created_at <= ?", filter.EndTime)
		}
		subQuery := q.Select("DISTINCT eu.user_id, eu.exam_id")
		err := database.Table("(?) as sub", subQuery).Count(&userExamCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Submitted Homework
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Model(&models.HomeworkUser{}).
			Joins("JOIN homeworks AS h ON h.id = homework_users.homework_id").
			Joins("JOIN homework_ref_lessons AS hrl ON hrl.homework_id = h.id").
			Joins("JOIN lessons AS l ON l.id = hrl.lesson_id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("homework_users.created_at <= ?", filter.EndTime)
		}
		err := q.Select("COUNT(DISTINCT homework_users.homework_id)").Scan(&submittedHomeworkCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Submitted Exam
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Model(&models.ExamUser{}).
			Joins("JOIN exams AS e ON e.id = exam_users.exam_id").
			Joins("JOIN exam_ref_lessons AS erl ON erl.exam_id = e.id").
			Joins("JOIN lessons AS l ON l.id = erl.lesson_id").
			Joins("JOIN chapters AS c ON l.chapter_id = c.id").
			Joins("JOIN programs p ON c.program_id = p.id")
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("exam_users.created_at <= ?", filter.EndTime)
		}
		err := q.Select("COALESCE(COUNT(DISTINCT exam_users.exam_id), 0)").Scan(&submittedExamCount).Error
		if err != nil {
			config.Log.Error("Error executing submitted exam query: ", err)
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Cloned QuestionIDs (raw SQL)
	wg.Add(1)
	go func() {
		defer wg.Done()
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
		err := database.Raw(queryClonedQuestion, args...).Scan(&questionIDs).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Completed Lesson Plan
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Table("lesson_plans AS lp").
			Joins("JOIN lesson_plan_completes AS lpc ON lp.id = lpc.lesson_plan_id")
		if programId > 0 || filter.SchoolId > 0 {
			q = q.
				Joins("JOIN lesson_plan_ref_lessons AS lprl ON lp.id = lprl.lesson_plan_id").
				Joins("JOIN lessons AS l ON lprl.lesson_id = l.id").
				Joins("JOIN chapters AS c ON l.chapter_id = c.id").
				Joins("JOIN programs p ON c.program_id = p.id")
		}
		if programId > 0 {
			q = q.Where("p.id = ?", programId)
		}
		if filter.SchoolId > 0 {
			q = q.
				Joins("JOIN courses co ON co.program_id = p.id").
				Joins("JOIN course_schools cs ON cs.course_id = co.id").
				Where("cs.school_id = ?", filter.SchoolId)
		}
		if filterByDate {
			q = q.Where("lpc.completed_at <= ?", filter.EndTime)
		}
		err := q.Distinct("lp.id").Count(&completedLessonPlanCount).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	// Ungraded
	wg.Add(1)
	go func() {
		defer wg.Done()
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
		var err error
		if programId > 0 && filter.SchoolId > 0 {
			err = database.Raw(ungradedQuery, programId, filter.SchoolId).Scan(&unsubmittedUserExamCount).Error
		} else if programId > 0 {
			err = database.Raw(ungradedQuery, programId).Scan(&unsubmittedUserExamCount).Error
		} else if filter.SchoolId > 0 {
			err = database.Raw(ungradedQuery, filter.SchoolId).Scan(&unsubmittedUserExamCount).Error
		} else {
			err = database.Raw(ungradedQuery).Scan(&unsubmittedUserExamCount).Error
		}
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// Dedupe questionIDs → uniqueQuestionIDs, rồi query question used (phụ thuộc kết quả cloned)
	uniqueQuestionIDs := make([]int, 0)
	seen := make(map[int]struct{})
	for _, q := range questionIDs {
		if _, ok := seen[q.ID]; !ok {
			seen[q.ID] = struct{}{}
			uniqueQuestionIDs = append(uniqueQuestionIDs, q.ID)
		}
	}

	if len(uniqueQuestionIDs) > 0 {
		questionUsedQuery := database.Model(&models.Question{}).Where("id IN (?)", uniqueQuestionIDs)
		if filterByDate {
			questionUsedQuery = questionUsedQuery.Where("created_at <= ?", filter.EndTime)
		}
		questionUsedQuery.Count(&questionUsedCount)
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
	database := db.ReplicaDB

	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	includeExam := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exam"
	includeHomework := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "homework"
	includeExercise := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exercise"

	var userIds []int64
	hasFilter := filter.ProgramId > 0 || filter.CourseId > 0 || filter.SchoolId > 0 || filter.ClassId > 0 || filter.TeacherId > 0

	// === Phase 1: chạy song song userIds (nếu hasFilter) và weeks ===
	var weeks []struct {
		ID         int64
		WeekNumber int
	}
	var firstErr error
	var mu sync.Mutex
	wg1 := &sync.WaitGroup{}

	if hasFilter {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			userIdQuery := database.Table("users u")
			needUserCourses := filter.ProgramId > 0 || filter.CourseId > 0 || filter.TeacherId > 0
			if needUserCourses {
				userIdQuery = userIdQuery.Joins("JOIN user_courses uc ON uc.user_id = u.id").
					Joins("JOIN courses c ON c.id = uc.course_id")
				if filter.CourseId > 0 {
					userIdQuery = userIdQuery.Where("c.id = ?", filter.CourseId)
				}
				if filter.ProgramId > 0 {
					userIdQuery = userIdQuery.Where("c.program_id = ?", filter.ProgramId)
				}
				if filter.TeacherId > 0 {
					userIdQuery = userIdQuery.Where("uc.user_id = ?", filter.TeacherId)
				}
			}
			if filter.SchoolId > 0 {
				userIdQuery = userIdQuery.Where("u.school_id = ?", filter.SchoolId)
			}
			if filter.ClassId > 0 {
				userIdQuery = userIdQuery.Joins("JOIN user_classes ucl ON ucl.user_id = u.id").
					Where("ucl.class_id = ?", filter.ClassId)
			}
			err := userIdQuery.Distinct("u.id").Pluck("u.id", &userIds).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wg1.Add(1)
	go func() {
		defer wg1.Done()
		q := database.Table("weeks").Select("id, week_number")
		if filterByDate {
			q = q.Where("end_date >= ?", filter.StartTime).Where("end_date <= ?", filter.EndTime).Order("id ASC")
		} else {
			q = q.Where("start_date <= NOW()").Order("id DESC").Limit(20)
		}
		err := q.Scan(&weeks).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg1.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// === Phase 2: build union SQL/args (dùng userIds, hasFilter) ===
	buildWeekScoreQuery := func(userTable string) (string, []interface{}) {
		var args []interface{}
		whereClause := " WHERE w.start_date <= NOW()"
		if filterByDate {
			whereClause += " AND w.end_date >= ? AND w.end_date <= ?"
			args = append(args, filter.StartTime, filter.EndTime)
		}
		if len(userIds) > 0 {
			placeholders := strings.Repeat("?,", len(userIds))
			placeholders = placeholders[:len(placeholders)-1]
			whereClause += fmt.Sprintf(" AND %s.user_id IN (%s)", userTable, placeholders)
			for _, id := range userIds {
				args = append(args, id)
			}
		} else if hasFilter {
			whereClause += " AND 1=0"
		}
		query := fmt.Sprintf(`
			SELECT w.id AS week_id, COALESCE(AVG(%s.ratio), 0) AS avg_score
			FROM weeks w
			LEFT JOIN %s ON %s.created_at BETWEEN w.start_date AND w.end_date%s
			GROUP BY w.id`,
			userTable, userTable, userTable, whereClause)
		return query, args
	}

	var unionParts []string
	var unionArgs []interface{}
	if includeExam {
		examSQL, examArgs := buildWeekScoreQuery("exam_users")
		unionParts = append(unionParts, examSQL)
		unionArgs = append(unionArgs, examArgs...)
	}
	if includeHomework {
		homeworkSQL, homeworkArgs := buildWeekScoreQuery("homework_users")
		unionParts = append(unionParts, homeworkSQL)
		unionArgs = append(unionArgs, homeworkArgs...)
	}
	if includeExercise {
		exerciseSQL, exerciseArgs := buildWeekScoreQuery("exercise_users")
		unionParts = append(unionParts, exerciseSQL)
		unionArgs = append(unionArgs, exerciseArgs...)
	}

	buildTopQuery := func(tableName, assignmentTable, assignmentIdCol, assignmentNameCol, typeName string) (string, []interface{}) {
		var args []interface{}
		whereClause := " WHERE 1=1"
		if filterByDate {
			whereClause += fmt.Sprintf(" AND %s.created_at <= ?", tableName)
			args = append(args, filter.EndTime)
		}
		if len(userIds) > 0 {
			placeholders := strings.Repeat("?,", len(userIds))
			placeholders = placeholders[:len(placeholders)-1]
			whereClause += fmt.Sprintf(" AND %s.user_id IN (%s)", tableName, placeholders)
			for _, id := range userIds {
				args = append(args, id)
			}
		} else if hasFilter {
			whereClause += " AND 1=0"
		}
		query := fmt.Sprintf(`
			SELECT
				%s.user_id,
				users.name AS user_name,
				users.avatar_info,
				AVG(%s.ratio) AS avg_ratio,
				%s.%s as assignment_id,
				%s.id as user_record_id,
				a.%s AS assignment_name,
				'%s' AS type,
				users.name AS student_name,
				COALESCE(schools.name, '') AS school_name,
				COALESCE(MIN(classes.name), '') AS class_name
			FROM %s
			JOIN users ON %s.user_id = users.id
			LEFT JOIN %s a ON %s.%s = a.id
			LEFT JOIN user_classes uc ON uc.user_id = users.id
			LEFT JOIN schools ON schools.id = users.school_id
			LEFT JOIN classes ON classes.id = uc.class_id%s
			GROUP BY %s.user_id, users.name, users.avatar_info, %s.%s, %s.id, a.%s, users.name, schools.name, classes.name
			HAVING AVG(%s.ratio) > 75`,
			tableName, tableName, tableName, assignmentIdCol, tableName, assignmentNameCol, typeName,
			tableName, tableName, assignmentTable, tableName, assignmentIdCol, whereClause,
			tableName, tableName, assignmentIdCol, tableName, assignmentNameCol, tableName)
		return query, args
	}

	var topHighestUnionParts []string
	var topHighestArgs []interface{}
	if includeExam {
		q, a := buildTopQuery("exam_users", "exams", "exam_id", "name", "exam")
		topHighestUnionParts = append(topHighestUnionParts, q)
		topHighestArgs = append(topHighestArgs, a...)
	}
	if includeHomework {
		q, a := buildTopQuery("homework_users", "homeworks", "homework_id", "name", "homework")
		topHighestUnionParts = append(topHighestUnionParts, q)
		topHighestArgs = append(topHighestArgs, a...)
	}
	if includeExercise {
		q, a := buildTopQuery("exercise_users", "exercises", "exercise_id", "name", "exercise")
		topHighestUnionParts = append(topHighestUnionParts, q)
		topHighestArgs = append(topHighestArgs, a...)
	}

	buildTopLowestQuery := func(tableName, assignmentTable, assignmentIdCol, assignmentNameCol, typeName string) (string, []interface{}) {
		var args []interface{}
		whereClause := " WHERE 1=1"
		if filterByDate {
			whereClause += fmt.Sprintf(" AND %s.created_at <= ?", tableName)
			args = append(args, filter.EndTime)
		}
		if len(userIds) > 0 {
			placeholders := strings.Repeat("?,", len(userIds))
			placeholders = placeholders[:len(placeholders)-1]
			whereClause += fmt.Sprintf(" AND %s.user_id IN (%s)", tableName, placeholders)
			for _, id := range userIds {
				args = append(args, id)
			}
		} else if hasFilter {
			whereClause += " AND 1=0"
		}
		query := fmt.Sprintf(`
			SELECT
				%s.user_id,
				users.name AS user_name,
				users.avatar_info,
				AVG(%s.ratio) AS avg_ratio,
				%s.%s as assignment_id,
				%s.id as user_record_id,
				a.%s AS assignment_name,
				'%s' AS type,
				users.name AS student_name,
				COALESCE(schools.name, '') AS school_name,
				COALESCE(MIN(classes.name), '') AS class_name
			FROM %s
			JOIN users ON %s.user_id = users.id
			LEFT JOIN %s a ON %s.%s = a.id
			LEFT JOIN user_classes uc ON uc.user_id = users.id
			LEFT JOIN schools ON schools.id = users.school_id
			LEFT JOIN classes ON classes.id = uc.class_id%s
			GROUP BY %s.user_id, users.name, users.avatar_info, %s.%s, %s.id, a.%s, users.name, schools.name, classes.name
			HAVING AVG(%s.ratio) < 25`,
			tableName, tableName, tableName, assignmentIdCol, tableName, assignmentNameCol, typeName,
			tableName, tableName, assignmentTable, tableName, assignmentIdCol, whereClause,
			tableName, tableName, assignmentIdCol, tableName, assignmentNameCol, tableName)
		return query, args
	}

	var topLowestUnionParts []string
	var topLowestArgs []interface{}
	if includeExam {
		q, a := buildTopLowestQuery("exam_users", "exams", "exam_id", "name", "exam")
		topLowestUnionParts = append(topLowestUnionParts, q)
		topLowestArgs = append(topLowestArgs, a...)
	}
	if includeHomework {
		q, a := buildTopLowestQuery("homework_users", "homeworks", "homework_id", "name", "homework")
		topLowestUnionParts = append(topLowestUnionParts, q)
		topLowestArgs = append(topLowestArgs, a...)
	}
	if includeExercise {
		q, a := buildTopLowestQuery("exercise_users", "exercises", "exercise_id", "name", "exercise")
		topLowestUnionParts = append(topLowestUnionParts, q)
		topLowestArgs = append(topLowestArgs, a...)
	}

	buildAvgQuery := func(tableName string) (string, []interface{}) {
		var args []interface{}
		whereClause := " WHERE 1=1"
		if filterByDate {
			whereClause += fmt.Sprintf(" AND %s.created_at <= ?", tableName)
			args = append(args, filter.EndTime)
		}
		if len(userIds) > 0 {
			placeholders := strings.Repeat("?,", len(userIds))
			placeholders = placeholders[:len(placeholders)-1]
			whereClause += fmt.Sprintf(" AND %s.user_id IN (%s)", tableName, placeholders)
			for _, id := range userIds {
				args = append(args, id)
			}
		} else if hasFilter {
			whereClause += " AND 1=0"
		}
		query := fmt.Sprintf(`SELECT %s.user_id, %s.ratio FROM %s%s`, tableName, tableName, tableName, whereClause)
		return query, args
	}

	var avgUnionParts []string
	var avgArgs []interface{}
	if includeExam {
		q, a := buildAvgQuery("exam_users")
		avgUnionParts = append(avgUnionParts, q)
		avgArgs = append(avgArgs, a...)
	}
	if includeHomework {
		q, a := buildAvgQuery("homework_users")
		avgUnionParts = append(avgUnionParts, q)
		avgArgs = append(avgArgs, a...)
	}
	if includeExercise {
		q, a := buildAvgQuery("exercise_users")
		avgUnionParts = append(avgUnionParts, q)
		avgArgs = append(avgArgs, a...)
	}

	// === Phase 3: chạy song song 5 query (weekScores, topHighest, topLowest, averageScore, scoreCounts) ===
	type TopStudent struct {
		UserId         int64
		UserName       string
		Avatar         models.MediaInfo
		AvgRatio       float64
		AssignmentId   int64
		UserRecordId   int64
		StudentName    string
		AssignmentName string
		Type           string
		SchoolName     string
		ClassName      string
	}

	var weekScores []struct {
		WeekID   int64
		AvgScore float64
	}
	var topHighestStudents []TopStudent
	var topLowestStudents []TopStudent
	var averageScore struct {
		TotalAvgRatio  float64
		TotalUserCount int64
	}
	var scoreCounts struct {
		HighRatioCount int64
		LowRatioCount  int64
		TotalCount     int64
	}

	firstErr = nil
	wg3 := &sync.WaitGroup{}

	if len(unionParts) > 0 {
		wg3.Add(1)
		finalSQL := `SELECT week_id, AVG(avg_score) AS avg_score FROM (` + strings.Join(unionParts, " UNION ALL ") + `) AS combined GROUP BY week_id ORDER BY week_id DESC LIMIT 20`
		go func() {
			defer wg3.Done()
			err := database.Raw(finalSQL, unionArgs...).Scan(&weekScores).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	if len(topHighestUnionParts) > 0 {
		wg3.Add(1)
		finalTopHighestSQL := strings.Join(topHighestUnionParts, " UNION ALL ") + " ORDER BY avg_ratio DESC LIMIT 10"
		go func() {
			defer wg3.Done()
			err := database.Raw(finalTopHighestSQL, topHighestArgs...).Scan(&topHighestStudents).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	if len(topLowestUnionParts) > 0 {
		wg3.Add(1)
		finalTopLowestSQL := strings.Join(topLowestUnionParts, " UNION ALL ") + " ORDER BY avg_ratio ASC LIMIT 10"
		go func() {
			defer wg3.Done()
			err := database.Raw(finalTopLowestSQL, topLowestArgs...).Scan(&topLowestStudents).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	if len(avgUnionParts) > 0 {
		wg3.Add(1)
		finalAvgSQL := fmt.Sprintf(`
			SELECT COALESCE(AVG(ratio), 0) AS total_avg_ratio, COUNT(DISTINCT user_id) AS total_user_count
			FROM (%s) AS combined`, strings.Join(avgUnionParts, " UNION ALL "))
		go func() {
			defer wg3.Done()
			err := database.Raw(finalAvgSQL, avgArgs...).Scan(&averageScore).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()

		wg3.Add(1)
		finalCountSQL := fmt.Sprintf(`
			SELECT SUM(CASE WHEN ratio >= 75 THEN 1 ELSE 0 END) AS high_ratio_count,
				SUM(CASE WHEN ratio < 25 THEN 1 ELSE 0 END) AS low_ratio_count, COUNT(*) AS total_count
			FROM (%s) AS combined`, strings.Join(avgUnionParts, " UNION ALL "))
		go func() {
			defer wg3.Done()
			err := database.Raw(finalCountSQL, avgArgs...).Scan(&scoreCounts).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wg3.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// === Map tuần + build DTO ===
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

	topHighest := make([]dto.StudentScore, 0, len(topHighestStudents))
	for _, student := range topHighestStudents {
		topHighest = append(topHighest, dto.StudentScore{
			Id:         student.UserId,
			TypeId:     student.AssignmentId,
			TypeUserId: student.UserRecordId,
			Name:       student.UserName,
			TypeName:   student.AssignmentName,
			Score:      float32(student.AvgRatio),
			Type:       student.Type,
			AvatarInfo: student.Avatar,
			ClassName:  student.ClassName,
			SchoolName: student.SchoolName,
		})
	}

	topLowest := make([]dto.StudentScore, 0, len(topLowestStudents))
	for _, student := range topLowestStudents {
		topLowest = append(topLowest, dto.StudentScore{
			Id:         student.UserId,
			TypeId:     student.AssignmentId,
			TypeUserId: student.UserRecordId,
			Name:       student.UserName,
			TypeName:   student.AssignmentName,
			Score:      float32(student.AvgRatio),
			Type:       student.Type,
			AvatarInfo: student.Avatar,
			ClassName:  student.ClassName,
			SchoolName: student.SchoolName,
		})
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
	database := db.ReplicaDB

	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	includeExam := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exam"
	includeHomework := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "homework"
	includeExercise := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exercise"

	var userIds []int64
	hasFilter := filter.ProgramId > 0 || filter.CourseId > 0 || filter.SchoolId > 0 || filter.ClassId > 0 || filter.TeacherId > 0

	type Grade struct {
		ID     int64
		Number int64
	}
	var grades []Grade

	// === Phase 1: chạy song song userIds (nếu hasFilter) và grades ===
	var firstErr error
	var mu sync.Mutex
	wg := &sync.WaitGroup{}

	if hasFilter {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userIdQuery := database.Table("users u")
			needUserCourses := filter.ProgramId > 0 || filter.CourseId > 0 || filter.TeacherId > 0
			if needUserCourses {
				userIdQuery = userIdQuery.Joins("JOIN user_courses uc ON uc.user_id = u.id").
					Joins("JOIN courses c ON c.id = uc.course_id")
				if filter.CourseId > 0 {
					userIdQuery = userIdQuery.Where("c.id = ?", filter.CourseId)
				}
				if filter.ProgramId > 0 {
					userIdQuery = userIdQuery.Where("c.program_id = ?", filter.ProgramId)
				}
				if filter.TeacherId > 0 {
					userIdQuery = userIdQuery.Where("uc.user_id = ?", filter.TeacherId)
				}
			}
			if filter.SchoolId > 0 {
				userIdQuery = userIdQuery.Where("u.school_id = ?", filter.SchoolId)
			}
			if filter.ClassId > 0 {
				userIdQuery = userIdQuery.Joins("JOIN user_classes ucl ON ucl.user_id = u.id").
					Where("ucl.class_id = ?", filter.ClassId)
			}
			err := userIdQuery.Distinct("u.id").Pluck("u.id", &userIds).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Model(&models.Grade{}).
			Select("id, number").
			Order("number").
			Find(&grades).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// === Phase 2: build union và chạy query phân phối điểm ===
	type BinResult struct {
		Bin     int64
		Count   int64
		GradeID int64
	}

	var results []BinResult
	buildScoreDistQuery := func(userTable string) (string, []interface{}) {
		var args []interface{}
		whereClause := " WHERE 1=1"
		if filterByDate {
			whereClause += fmt.Sprintf(" AND %s.created_at <= ?", userTable)
			args = append(args, filter.EndTime)
		}
		if len(userIds) > 0 {
			placeholders := strings.Repeat("?,", len(userIds))
			placeholders = placeholders[:len(placeholders)-1]
			whereClause += fmt.Sprintf(" AND %s.user_id IN (%s)", userTable, placeholders)
			for _, id := range userIds {
				args = append(args, id)
			}
		} else if hasFilter {
			whereClause += " AND 1=0"
		}
		query := fmt.Sprintf(`
			SELECT 
				FLOOR(%s.ratio / 10) * 10 AS bin, 
				cl.grade_id,
				COUNT(DISTINCT %s.user_id) AS count
			FROM %s
			JOIN users u ON %s.user_id = u.id
			JOIN user_classes uc ON uc.user_id = u.id
			JOIN classes cl ON uc.class_id = cl.id
			%s
			GROUP BY bin, cl.grade_id`,
			userTable, userTable, userTable, userTable, whereClause)
		return query, args
	}

	var unionParts []string
	var unionArgs []interface{}
	if includeExam {
		q, a := buildScoreDistQuery("exam_users")
		unionParts = append(unionParts, q)
		unionArgs = append(unionArgs, a...)
	}
	if includeHomework {
		q, a := buildScoreDistQuery("homework_users")
		unionParts = append(unionParts, q)
		unionArgs = append(unionArgs, a...)
	}
	if includeExercise {
		q, a := buildScoreDistQuery("exercise_users")
		unionParts = append(unionParts, q)
		unionArgs = append(unionArgs, a...)
	}

	if len(unionParts) > 0 {
		finalSQL := fmt.Sprintf(`
			SELECT bin, grade_id, SUM(count) AS count
			FROM (%s) AS combined
			GROUP BY bin, grade_id
			ORDER BY grade_id, bin`, strings.Join(unionParts, " UNION ALL "))

		if err := database.Raw(finalSQL, unionArgs...).Scan(&results).Error; err != nil {
			return nil, err
		}
	}

	// === Map bin theo grade và build items ===
	resultMap := make(map[int64]map[int64]int64)
	for _, r := range results {
		if _, ok := resultMap[r.GradeID]; !ok {
			resultMap[r.GradeID] = make(map[int64]int64)
		}
		resultMap[r.GradeID][r.Bin] = r.Count
	}

	min := int64(0)
	max := int64(100)
	step := int64(10)
	var items []dto.ScoreDistributionItem

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
	}

	return &dto.ScoreDistributionOverview{
		Items: items,
	}, nil
}

func (r *dashboardRepository) GetSystemUsage(filter dto.FilterDashboardAdmin) (*dto.SystemUsageOverview, error) {
	database := db.ReplicaDB

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

	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", filter.StartTime, filter.EndTime)

	// Phase 1: một vòng lặp — kiểm tra exists một lần mỗi bảng, build cả weekly và device union
	var unionQueries []string
	var args []interface{}
	var deviceUnionQueries []string
	var deviceArgs []interface{}

	for _, tableName := range tableNames {
		var exists bool
		if err := database.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = ?
			)`, tableName).Scan(&exists).Error; err != nil || !exists {
			continue
		}

		// Weekly subquery
		subQuery := fmt.Sprintf(`
			SELECT
				CONCAT(EXTRACT('week' FROM al.created_at)::INT, '/', EXTRACT('isoyear' FROM al.created_at)::INT) AS week_label,
				COUNT(DISTINCT FLOOR(EXTRACT(EPOCH FROM al.created_at) / 300)) * 5 AS duration
			FROM %s al
		`, tableName)
		joinClause := " JOIN users ON users.id = al.user_id"
		whereClause := "WHERE al.created_at >= ? AND al.created_at <= ?"
		queryArgs := []interface{}{filter.StartTime, filter.EndTime}
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
		subQuery += joinClause + " " + whereClause + " GROUP BY week_label"
		unionQueries = append(unionQueries, subQuery)
		args = append(args, queryArgs...)

		// Device subquery
		deviceSubQuery := fmt.Sprintf(`
			SELECT DISTINCT ON (al.user_id) al.user_id, al.device
			FROM %s al
		`, tableName)
		deviceJoin := " JOIN users u ON u.id = al.user_id"
		deviceWhere := "WHERE al.device IS NOT NULL AND al.created_at >= ? AND al.created_at <= ?"
		if filter.SchoolId > 0 {
			deviceWhere += fmt.Sprintf(" AND u.school_id = %d", filter.SchoolId)
		}
		if filter.CourseId > 0 || filter.ProgramId > 0 {
			deviceJoin += " JOIN user_courses uc ON uc.user_id = al.user_id"
			if filter.CourseId > 0 {
				deviceWhere += fmt.Sprintf(" AND uc.course_id = %d", filter.CourseId)
			}
			if filter.ProgramId > 0 {
				deviceJoin += " JOIN courses co ON co.id = uc.course_id"
				deviceWhere += fmt.Sprintf(" AND co.program_id = %d", filter.ProgramId)
			}
		}
		if filter.RoleId > 0 {
			if !strings.Contains(deviceJoin, "JOIN user_ref_roles") {
				deviceJoin += " JOIN user_ref_roles urr ON urr.user_id = u.id"
			}
			deviceWhere += fmt.Sprintf(" AND urr.role_id = %d", filter.RoleId)
		}
		deviceSubQuery += deviceJoin + " " + deviceWhere + " ORDER BY al.user_id, al.created_at DESC"
		deviceUnionQueries = append(deviceUnionQueries, deviceSubQuery)
		deviceArgs = append(deviceArgs, filter.StartTime, filter.EndTime)
	}

	// Phase 2: chạy song song weeks, weeklyResults, deviceResults, averageUsed
	var weeklyResults []WeeklyUsageResult
	var weeks []models.Week
	var deviceResults []DeviceUsageResult
	var averageUsed models.AverageUsed

	var firstErr error
	var mu sync.Mutex
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Raw(`
			SELECT week_number, start_date, end_date, year
			FROM weeks
			WHERE start_date >= ? AND end_date <= ? AND start_date < NOW()
			ORDER BY year, week_number
		`, filter.StartTime.AddDate(0, 0, -4), filter.EndTime.AddDate(0, 0, 4)).Scan(&weeks).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	if len(unionQueries) > 0 {
		wg.Add(1)
		mainQuery := fmt.Sprintf(`
			SELECT week_label, SUM(duration) as duration
			FROM (%s) AS combined
			GROUP BY week_label
			ORDER BY week_label ASC
		`, strings.Join(unionQueries, " UNION ALL "))
		go func() {
			defer wg.Done()
			err := database.Raw(mainQuery, args...).Scan(&weeklyResults).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	if len(deviceUnionQueries) > 0 {
		wg.Add(1)
		deviceMainQuery := fmt.Sprintf(`
			SELECT t.device AS device, COUNT(*) AS count,
				COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () AS percent
			FROM (
				SELECT DISTINCT ON (user_id) user_id, device
				FROM (%s) AS all_devices
				ORDER BY user_id, device
			) AS t
			GROUP BY t.device
		`, strings.Join(deviceUnionQueries, " UNION ALL "))
		go func() {
			defer wg.Done()
			err := database.Raw(deviceMainQuery, deviceArgs...).Scan(&deviceResults).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		avg, err := historyRepo.AverageUsed(filter.StartTime, filter.EndTime, filter.SchoolId, filter.CourseId, filter.ProgramId, filter.RoleId)
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
			return
		}
		averageUsed = avg
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// Phase 3: build usageMap, weeklyUsages, deviceUsages
	usageMap := make(map[string]float64)
	for _, r := range weeklyResults {
		usageMap[r.WeekLabel] = r.Duration
	}

	beginAt, _ := time.Parse("2006-01-02", "2025-09-08")
	today := time.Now()
	weeklyUsages := make([]dto.WeeklyUsage, 0, len(weeks))
	for _, w := range weeks {
		startDate := w.StartDate
		endDate := w.EndDate
		if startDate.Before(beginAt) {
			continue
		}
		diff := startDate.Sub(beginAt)
		weekNumber := int(diff.Hours()/(24*7)) + 1
		weekLabel := fmt.Sprintf("%d/%d", w.WeekNumber, w.Year)
		duration := float32(usageMap[weekLabel])
		isCurrent := (today.Equal(startDate) || today.After(startDate)) &&
			(today.Equal(endDate) || today.Before(endDate.Add(24*time.Hour)))
		weeklyUsages = append(weeklyUsages, dto.WeeklyUsage{
			Week:      int32(weekNumber),
			Duration:  duration,
			IsCurrent: isCurrent,
		})
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
			deviceUsages = append(deviceUsages, dto.DeviceUsage{Name: device, Count: 0, Percent: 0})
		}
	}

	return &dto.SystemUsageOverview{
		WeeklyUsages: weeklyUsages,
		DeviceUsages: deviceUsages,
		AverageUsed:  averageUsed,
	}, nil
}

func (r *dashboardRepository) GetSystemUsageType(filter dto.FilterDashboardAdmin) (*dto.SystemUsageOverview, error) {
	database := db.ReplicaDB

	// Step 1: Find courses based on filter
	courseQuery := database.Table("courses c").
		Select("DISTINCT c.id").
		Where("c.deleted_at IS NULL")
	if filter.SchoolId > 0 {
		courseQuery = courseQuery.
			Joins("JOIN course_schools cs ON cs.course_id = c.id").
			Where("cs.school_id = ?", filter.SchoolId)
	}
	if filter.ProgramId > 0 {
		courseQuery = courseQuery.Where("c.program_id = ?", filter.ProgramId)
	}
	if filter.CourseId > 0 {
		courseQuery = courseQuery.Where("c.id = ?", filter.CourseId)
	}

	var courseIds []int64
	if err := courseQuery.Pluck("id", &courseIds).Error; err != nil {
		return nil, err
	}
	if len(courseIds) == 0 {
		return &dto.SystemUsageOverview{CompleteCount: 0, CompletedRate: 0}, nil
	}

	includeExam := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exam"
	includeHomework := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "homework"
	includeExercise := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exercise"

	type Assignment struct {
		AssignmentID int64
		CourseID     int64
		Type         string
	}

	// Step 2: Chạy song song 3 query assignments (exam, homework, exercise)
	var examAssignments, homeworkAssignments, exerciseAssignments []Assignment
	var firstErr error
	var mu sync.Mutex
	wgAssign := &sync.WaitGroup{}

	if includeExam {
		wgAssign.Add(1)
		go func() {
			defer wgAssign.Done()
			var out []Assignment
			q := database.Table("exam_ref_lessons erl").
				Select("DISTINCT erl.exam_id as assignment_id, erl.course_id, 'exam' as type").
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = erl.lesson_id AND ls.course_id = erl.course_id").
				Where("erl.course_id IN ?", courseIds).
				Where("erl.assigned_at IS NOT NULL")
			if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
				q = q.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			}
			err := q.Scan(&out).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			examAssignments = out
		}()
	}
	if includeHomework {
		wgAssign.Add(1)
		go func() {
			defer wgAssign.Done()
			var out []Assignment
			q := database.Table("homework_ref_lessons hrl").
				Select("DISTINCT hrl.homework_id as assignment_id, hrl.course_id, 'homework' as type").
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = hrl.lesson_id AND ls.course_id = hrl.course_id").
				Where("hrl.course_id IN ?", courseIds).
				Where("hrl.assigned_at IS NOT NULL")
			if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
				q = q.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			}
			err := q.Scan(&out).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			homeworkAssignments = out
		}()
	}
	if includeExercise {
		wgAssign.Add(1)
		go func() {
			defer wgAssign.Done()
			var out []Assignment
			q := database.Table("exercise_ref_lessons exrl").
				Select("DISTINCT exrl.exercise_id as assignment_id, exrl.course_id, 'exercise' as type").
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = exrl.lesson_id AND ls.course_id = exrl.course_id").
				Where("exrl.course_id IN ?", courseIds).
				Where("exrl.assigned_at IS NOT NULL")
			if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
				q = q.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			}
			err := q.Scan(&out).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			exerciseAssignments = out
		}()
	}

	wgAssign.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	allAssignments := make([]Assignment, 0, len(examAssignments)+len(homeworkAssignments)+len(exerciseAssignments))
	allAssignments = append(allAssignments, examAssignments...)
	allAssignments = append(allAssignments, homeworkAssignments...)
	allAssignments = append(allAssignments, exerciseAssignments...)

	if len(allAssignments) == 0 {
		return &dto.SystemUsageOverview{CompleteCount: 0, CompletedRate: 0}, nil
	}

	coursesWithAssignments := make(map[int64]bool)
	for _, a := range allAssignments {
		coursesWithAssignments[a.CourseID] = true
	}
	filteredCourseIds := make([]int64, 0, len(courseIds))
	for _, courseID := range courseIds {
		if coursesWithAssignments[courseID] {
			filteredCourseIds = append(filteredCourseIds, courseID)
		}
	}
	if len(filteredCourseIds) == 0 {
		return &dto.SystemUsageOverview{CompleteCount: 0, CompletedRate: 0}, nil
	}

	assignmentsByCourse := make(map[int64][]Assignment)
	var examIDs, homeworkIDs, exerciseIDs []int64
	for _, a := range allAssignments {
		assignmentsByCourse[a.CourseID] = append(assignmentsByCourse[a.CourseID], a)
		switch a.Type {
		case "exam":
			examIDs = append(examIDs, a.AssignmentID)
		case "homework":
			homeworkIDs = append(homeworkIDs, a.AssignmentID)
		case "exercise":
			exerciseIDs = append(exerciseIDs, a.AssignmentID)
		}
	}

	type StudentCourse struct {
		UserID   int64
		CourseID int64
	}
	type CompletionRecord struct {
		UserID       int64
		AssignmentID int64
		Type         string
	}

	// Step 3: Chạy song song studentCourses + 3 completion queries
	var studentCourses []StudentCourse
	var examCompletions []struct {
		UserID int64
		ExamID int64
	}
	var homeworkCompletions []struct {
		UserID     int64
		HomeworkID int64
	}
	var exerciseCompletions []struct {
		UserID     int64
		ExerciseID int64
	}

	firstErr = nil
	wgData := &sync.WaitGroup{}

	wgData.Add(1)
	go func() {
		defer wgData.Done()
		err := database.Table("user_courses uc").
			Select("DISTINCT uc.user_id, uc.course_id").
			Joins("JOIN users u ON u.id = uc.user_id").
			Joins("JOIN user_ref_roles ur ON ur.user_id = u.id").
			Where("ur.role_id = ?", models.StudentRoleId).
			Where("u.deleted_at IS NULL").
			Where("uc.course_id IN ?", filteredCourseIds).
			Scan(&studentCourses).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	if includeExam && len(examIDs) > 0 {
		wgData.Add(1)
		go func() {
			defer wgData.Done()
			err := database.Table("exam_users").
				Select("DISTINCT user_id, exam_id").
				Where("exam_id IN ?", examIDs).
				Scan(&examCompletions).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	if includeHomework && len(homeworkIDs) > 0 {
		wgData.Add(1)
		go func() {
			defer wgData.Done()
			err := database.Table("homework_users").
				Select("DISTINCT user_id, homework_id").
				Where("homework_id IN ?", homeworkIDs).
				Scan(&homeworkCompletions).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	if includeExercise && len(exerciseIDs) > 0 {
		wgData.Add(1)
		go func() {
			defer wgData.Done()
			err := database.Table("exercise_users").
				Select("DISTINCT user_id, exercise_id").
				Where("exercise_id IN ?", exerciseIDs).
				Scan(&exerciseCompletions).Error
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	wgData.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	if len(studentCourses) == 0 {
		return &dto.SystemUsageOverview{CompleteCount: 0, CompletedRate: 0}, nil
	}

	studentCoursesMap := make(map[int64][]int64)
	uniqueStudents := make(map[int64]bool)
	for _, sc := range studentCourses {
		studentCoursesMap[sc.UserID] = append(studentCoursesMap[sc.UserID], sc.CourseID)
		uniqueStudents[sc.UserID] = true
	}

	allCompletions := make([]CompletionRecord, 0,
		len(examCompletions)+len(homeworkCompletions)+len(exerciseCompletions))
	for _, ec := range examCompletions {
		allCompletions = append(allCompletions, CompletionRecord{UserID: ec.UserID, AssignmentID: ec.ExamID, Type: "exam"})
	}
	for _, hc := range homeworkCompletions {
		allCompletions = append(allCompletions, CompletionRecord{UserID: hc.UserID, AssignmentID: hc.HomeworkID, Type: "homework"})
	}
	for _, exc := range exerciseCompletions {
		allCompletions = append(allCompletions, CompletionRecord{UserID: exc.UserID, AssignmentID: exc.ExerciseID, Type: "exercise"})
	}

	completionMap := make(map[int64]map[string]bool)
	for _, comp := range allCompletions {
		if completionMap[comp.UserID] == nil {
			completionMap[comp.UserID] = make(map[string]bool)
		}
		completionMap[comp.UserID][comp.Type+":"+fmt.Sprintf("%d", comp.AssignmentID)] = true
	}

	var completeCount int64
	for userID := range uniqueStudents {
		requiredAssignments := make(map[string]bool)
		for _, courseID := range studentCoursesMap[userID] {
			for _, a := range assignmentsByCourse[courseID] {
				requiredAssignments[a.Type+":"+fmt.Sprintf("%d", a.AssignmentID)] = true
			}
		}
		if len(requiredAssignments) == 0 {
			continue
		}
		completed := completionMap[userID]
		for requiredKey := range requiredAssignments {
			if completed[requiredKey] {
				completeCount++
				break
			}
		}
	}

	totalUniqueStudents := int64(len(uniqueStudents))
	var completeRate float32
	if totalUniqueStudents > 0 {
		completeRate = float32(completeCount) / float32(totalUniqueStudents) * 100
	}

	return &dto.SystemUsageOverview{
		CompleteCount: int32(completeCount),
		CompletedRate: completeRate,
	}, nil
}

func (r *dashboardRepository) GetQuestionBank(filter dto.FilterDashboardAdmin) (*dto.QuestionBankOverview, error) {
	database := db.ReplicaDB
	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0

	var totalQuestions int64
	var typeResults []struct {
		QuestionType string
		Total        int32
	}
	var allAttributes []struct {
		ID       int64
		ParentID *int64
		Name     string
	}
	var countResults []struct {
		AttributeID int64
		Count       int32
	}
	var withAudio, withImage int64

	var firstErr error
	var mu sync.Mutex
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Model(&models.Question{}).Count(&totalQuestions).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		q := database.Model(&models.Question{}).Select("question_type, COUNT(*) as total")
		if filterByDate {
			q = q.Where(`created_at <= ?`, filter.EndTime)
		}
		err := q.Group("question_type").Scan(&typeResults).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Table("question_attributes").Select("id, parent_id, name").Find(&allAttributes).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
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
		}
		countQuery += ` GROUP BY qa.id`
		err := database.Raw(countQuery, args...).Scan(&countResults).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Model(&models.Question{}).
			Where(`file_info->>'path' IS NOT NULL AND kind = ?`, "audio").
			Count(&withAudio).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Model(&models.Question{}).
			Where(`file_info->>'path' IS NOT NULL AND kind = ?`, "image").
			Count(&withImage).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
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

	countMap := make(map[int64]int32)
	for _, row := range countResults {
		countMap[row.AttributeID] = row.Count
	}

	rootMap := map[int64]dto.QuestionAttributeItem{}
	childMap := map[int64][]dto.QuestionDistributionItem{}
	for _, attr := range allAttributes {
		count := countMap[attr.ID]
		percent := float32(count) * 100 / float32(totalQuestions)
		if attr.ParentID == nil {
			rootMap[attr.ID] = dto.QuestionAttributeItem{Name: attr.Name}
		} else {
			childMap[*attr.ParentID] = append(childMap[*attr.ParentID], dto.QuestionDistributionItem{
				Name: attr.Name, Count: count, Percent: percent,
			})
		}
	}

	var attributes []dto.QuestionAttributeItem
	for id, root := range rootMap {
		root.Items = childMap[id]
		attributes = append(attributes, root)
	}

	mediaUsage := dto.MediaUsage{
		TotalQuestions: int32(totalQuestions),
		WithAudio:      int32(withAudio),
		WithImage:      int32(withImage),
	}
	if totalQuestions > 0 {
		mediaUsage.AudioPercent = float32(withAudio) * 100 / float32(totalQuestions)
		mediaUsage.ImagePercent = float32(withImage) * 100 / float32(totalQuestions)
	}

	return &dto.QuestionBankOverview{
		Types:      types,
		Attributes: attributes,
		MediaUsage: mediaUsage,
	}, nil
}

func (r *dashboardRepository) GetRiskWarning(filter dto.FilterDashboardAdmin) (*dto.RiskAndWarning, error) {
	database := db.ReplicaDB

	var (
		inactiveStudents    []dto.InactiveStudent
		slowGradingTeachers []dto.SlowGradingTeacher
		decliningStudents   []dto.DecliningStudent
	)

	var dateCondition string
	var dateArgs []interface{}
	if !filter.StartTime.IsZero() && !filter.EndTime.IsZero() {
		dateCondition = " AND eu.created_at >= ? AND eu.created_at <= ? "
		dateArgs = append(dateArgs, filter.StartTime, filter.EndTime)
	}

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
		SELECT u.id, u.name, u.avatar_info, c.name AS class, u.last_login_at AS last_login,
			DATE_PART('day', NOW() - COALESCE(u.last_login_at, u.created_at)) AS absent_days
		FROM users u
		LEFT JOIN user_classes uc ON uc.user_id = u.id
		LEFT JOIN classes c ON uc.class_id = c.id
		JOIN user_ref_roles urr ON urr.user_id = u.id
		%s
		WHERE %s
		GROUP BY u.id, u.name, u.avatar_info, c.name, u.created_at, u.last_login_at
		HAVING DATE_PART('day', NOW() - COALESCE(u.last_login_at, u.created_at)) >= 7
		ORDER BY absent_days DESC LIMIT 20
	`, joinCourse, whereInactive)

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
			SELECT eu.user_id, eu.id, eu.ratio,
				LAG(eu.ratio) OVER (PARTITION BY eu.user_id ORDER BY eu.created_at) AS prev_ratio
			FROM exam_users eu
			JOIN users u ON u.id = eu.user_id
			LEFT JOIN user_courses uco ON uco.user_id = u.id
			%s
			WHERE %s %s
		),
		filtered AS (
			SELECT user_id FROM scored
			GROUP BY user_id
			HAVING COUNT(*) >= 3 AND BOOL_AND(prev_ratio IS NULL OR ratio <= prev_ratio)
		)
		SELECT u.id, u.name, u.avatar_info, c.name AS class,
			json_agg(json_build_object('id', eu.id, 'score', eu.ratio) ORDER BY eu.created_at) AS score
		FROM users u
		JOIN exam_users eu ON eu.user_id = u.id
		JOIN user_ref_roles urr ON urr.user_id = u.id
		LEFT JOIN user_classes uc2 ON uc2.user_id = u.id
		LEFT JOIN classes c ON c.id = uc2.class_id
		WHERE u.id IN (SELECT user_id FROM filtered) AND urr.role_id = ? AND u.deleted_at IS NULL
		GROUP BY u.id, u.name, u.avatar_info, c.name ORDER BY u.id LIMIT 20
	`, joinCourseDeclining, whereDeclining, dateCondition)

	decliningArgs := append(append([]interface{}{}, dateArgs...), models.StudentRoleId)

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
			SELECT eu.exam_id, eu.user_id AS student_id, uct.user_id AS teacher_id
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
			SELECT eq.exam_id, eq.user_id AS student_id,
				MIN(eq.created_at) AS submitted_at, MIN(eq.scoring_at) AS first_scored_at
			FROM exam_question_user_manual_scoring eq
			GROUP BY eq.exam_id, eq.user_id
		)
		SELECT s.teacher_id, ut.name AS teacher_name, ut.avatar_info,
			COUNT(DISTINCT s.exam_id || '-' || s.student_id) AS total_submissions,
			COUNT(DISTINCT CASE WHEN g.first_scored_at IS NOT NULL THEN s.exam_id || '-' || s.student_id END) AS graded_submissions,
			ROUND(EXTRACT(EPOCH FROM AVG(COALESCE(g.first_scored_at, NOW()) - g.submitted_at)) / 3600, 2) AS avg_wait_hours
		FROM submissions s
		JOIN grading g ON s.exam_id = g.exam_id AND s.student_id = g.student_id
		JOIN users ut ON ut.id = s.teacher_id
		GROUP BY s.teacher_id, ut.name, ut.avatar_info
		ORDER BY avg_wait_hours DESC LIMIT 20
	`, whereSlow, dateCondition)

	slowGradingArgs := append(append([]interface{}{}, models.TeacherRoleId), dateArgs...)

	var firstErr error
	var mu sync.Mutex
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Raw(inactiveSQL, models.StudentRoleId).Scan(&inactiveStudents).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Raw(decliningSQL, decliningArgs...).Scan(&decliningStudents).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := database.Raw(slowGradingSQL, slowGradingArgs...).Scan(&slowGradingTeachers).Error
		if err != nil {
			mu.Lock()
			if firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	return &dto.RiskAndWarning{
		InactiveStudents:    inactiveStudents,
		DecliningStudents:   decliningStudents,
		SlowGradingTeachers: slowGradingTeachers,
	}, nil
}

func (r *dashboardRepository) GetUserOverview(filter dto.FilterDashboardAdmin) (*dto.UserOverview, error) {
	var result models.DashboardOverviewResult

	endTimeUTC := filter.EndTime.UTC()
	startPreviousTimeUTC := filter.StartPreviousTime.UTC()
	endPreviousTimeUTC := filter.EndPreviousTime.UTC()
	startTimeUTC := filter.StartTime.UTC()

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

	var overviewErr error
	var activityTotal, activityCurrent, activityNew int32
	var activityErr error

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		overviewErr = db.MasterDB.Raw(query).Scan(&result).Error
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		activityTotal, activityCurrent, _, activityNew, activityErr = r.GetActivity(filter)
	}()

	wg.Wait()

	if overviewErr != nil {
		config.Log.Error("Error calling get_dashboard_overview:", overviewErr)
		config.Log.Info("Falling back to individual methods...")
		return r.getUserOverviewFallback(filter)
	}

	if result.SchoolTotalCount == 0 && result.UserTotalCount == 0 && result.StudentTotalCount == 0 && result.TeacherTotalCount == 0 {
		config.Log.Info("SQL function returned all zeros, trying fallback method...")
		return r.getUserOverviewFallback(filter)
	}

	if activityErr != nil {
		return nil, activityErr
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
