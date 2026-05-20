package services

import (
	"be-cleverschool/config"
	"be-cleverschool/dto"
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/redis"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/utils"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Cache keys prefix
	DashboardUserRegisterKey       = "dashboard:user_register"
	DashboardStudentRegisterKey    = "dashboard:student_register"
	DashboardTeacherRegisterKey    = "dashboard:teacher_register"
	DashboardActivityKey           = "dashboard:activity"
	DashboardUserOverviewKey       = "dashboard:user_overview"
	DashboardCourseOverviewKey     = "dashboard:course_overview"
	DashboardLearningKey           = "dashboard:learning"
	DashboardScoreDistributionKey  = "dashboard:score_distribution"
	DashboardSystemUsageKey        = "dashboard:system_usage"
	DashboardSystemUsageTypeKey    = "dashboard:system_usage_type"
	DashboardTeacherPerformanceKey = "dashboard:teacher_performance"
	DashboardQuestionBankKey       = "dashboard:question_bank"
	DashboardRiskWarningKey        = "dashboard:risk_warning"

	// Cache TTL
	DashboardCacheTTL = 30 * time.Minute // 30 minute
)

type DashboardCacheService struct {
	repo            repositories.DashboardRepository
	activityLogRepo repositories.MonthlyActivityLogRepository
}

func NewDashboardCacheService(repo repositories.DashboardRepository) *DashboardCacheService {
	return &DashboardCacheService{
		repo:            repo,
		activityLogRepo: repositories.NewMonthlyActivityLogRepository(),
	}
}

// Generate cache key with filter parameters
func (dcs *DashboardCacheService) generateCacheKey(baseKey string, filters map[string]interface{}) string {
	if len(filters) == 0 {
		return baseKey
	}

	filterStr, _ := json.Marshal(filters)
	return fmt.Sprintf("%s:%s", baseKey, string(filterStr))
}

// Cache for User Register
func (dcs *DashboardCacheService) GetCachedUserRegister(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current":  filter.StartTime.Unix(),
		"start_previous": filter.StartPreviousTime.Unix(),
		"end_current":    filter.EndTime.Unix(),
		"end_previous":   filter.EndPreviousTime.Unix(),
		"province_code":  filter.Address.ProvinceCode,
		"ward_code":      filter.Address.WardCode,
		"program_id":     filter.ProgramId,
		"course_id":      filter.CourseId,
		"school_id":      filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardUserRegisterKey, filters)

	cachedData, err := redis.RememberCache[map[string]int32](cacheKey, DashboardCacheTTL, func() (map[string]int32, error) {
		totalRegister, currentRegister, lastRegister, err := dcs.repo.GetUserRegister(filter)
		if err != nil {
			return map[string]int32{"total": 0, "current": 0, "last": 0}, nil
		}
		return map[string]int32{"total": totalRegister, "current": currentRegister, "last": lastRegister}, nil
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		totalRegister, currentRegister, lastRegister, _ := dcs.repo.GetUserRegister(filter)
		return dcs.formatUserRegisterData(totalRegister, currentRegister, lastRegister, filter, c)
	}

	return dcs.formatUserRegisterData(cachedData["total"], cachedData["current"], cachedData["last"], filter, c)
}

// Cache for Student Register
func (dcs *DashboardCacheService) GetCachedStudentRegister(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current":  filter.StartTime.Unix(),
		"start_previous": filter.StartPreviousTime.Unix(),
		"end_current":    filter.EndTime.Unix(),
		"end_previous":   filter.EndPreviousTime.Unix(),
		"province_code":  filter.Address.ProvinceCode,
		"ward_code":      filter.Address.WardCode,
		"program_id":     filter.ProgramId,
		"course_id":      filter.CourseId,
		"school_id":      filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardStudentRegisterKey, filters)

	cachedData, err := redis.RememberCache[map[string]int32](cacheKey, DashboardCacheTTL, func() (map[string]int32, error) {
		totalRegister, currentRegister, oldRegister, err := dcs.repo.GetStudentRegister(filter)
		if err != nil {
			return map[string]int32{"total": 0, "current": 0, "old": 0}, nil
		}
		return map[string]int32{"total": totalRegister, "current": currentRegister, "old": oldRegister}, nil
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		totalRegister, currentRegister, oldRegister, _ := dcs.repo.GetStudentRegister(filter)
		return dcs.formatStudentRegisterData(totalRegister, currentRegister, oldRegister, c)
	}

	return dcs.formatStudentRegisterData(cachedData["total"], cachedData["current"], cachedData["old"], c)
}

// Cache for Teacher Register
func (dcs *DashboardCacheService) GetCachedTeacherRegister(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current":  filter.StartTime.Unix(),
		"start_previous": filter.StartPreviousTime.Unix(),
		"end_current":    filter.EndTime.Unix(),
		"end_previous":   filter.EndPreviousTime.Unix(),
		"province_code":  filter.Address.ProvinceCode,
		"ward_code":      filter.Address.WardCode,
		"program_id":     filter.ProgramId,
		"course_id":      filter.CourseId,
		"school_id":      filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardTeacherRegisterKey, filters)

	cachedData, err := redis.RememberCache[map[string]int32](cacheKey, DashboardCacheTTL, func() (map[string]int32, error) {
		totalRegister, currentRegister, oldRegister, err := dcs.repo.GetTeacherRegister(filter)
		if err != nil {
			return map[string]int32{"total": 0, "current": 0, "old": 0}, nil
		}
		return map[string]int32{"total": totalRegister, "current": currentRegister, "old": oldRegister}, nil
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		totalRegister, currentRegister, oldRegister, _ := dcs.repo.GetTeacherRegister(filter)
		return dcs.formatTeacherRegisterData(totalRegister, currentRegister, oldRegister, c)
	}

	return dcs.formatTeacherRegisterData(cachedData["total"], cachedData["current"], cachedData["old"], c)
}

// Cache for Activity
func (dcs *DashboardCacheService) GetCachedActivity(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"province_code": filter.Address.ProvinceCode,
		"ward_code":     filter.Address.WardCode,
		"program_id":    filter.ProgramId,
		"course_id":     filter.CourseId,
		"school_id":     filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardActivityKey, filters)

	cachedData, err := redis.RememberCache[map[string]int32](cacheKey, DashboardCacheTTL, func() (map[string]int32, error) {
		totalActivity, currentActivity, oldActivity, newActivity, err := dcs.repo.GetActivity(filter)
		if err != nil {
			return map[string]int32{"total": 0, "current": 0, "old": 0, "new": 0}, nil
		}
		return map[string]int32{"total": totalActivity, "current": currentActivity, "old": oldActivity, "new": newActivity}, nil
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		totalActivity, currentActivity, oldActivity, newActivity, _ := dcs.repo.GetActivity(filter)
		return dcs.formatActivityData(totalActivity, currentActivity, oldActivity, newActivity, filter, c)
	}

	return dcs.formatActivityData(cachedData["total"], cachedData["current"], cachedData["old"], cachedData["new"], filter, c)
}

// Cache for User Online (active users in last 15 minutes)
func (dcs *DashboardCacheService) GetCachedUserOnline(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"minutes":    15,
		"program_id": filter.ProgramId,
		"course_id":  filter.CourseId,
		"school_id":  filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey("dashboard:user_online", filters)

	cachedData, err := redis.RememberCache[int](cacheKey, 1*time.Minute, func() (int, error) {
		count, err := dcs.activityLogRepo.GetActiveUsersCount(15, filter.SchoolId, filter.CourseId, filter.ProgramId)
		if err != nil {
			return 0, err
		}
		return int(count), nil
	})

	// Not use cache
	if err != nil {
		// Fallback to direct fetch if cache fails
		activeUsersCount, _ := dcs.activityLogRepo.GetActiveUsersCount(15, filter.SchoolId, filter.CourseId, filter.ProgramId)
		return dcs.formatUserOnlineData(int32(activeUsersCount), c)
	}

	return dcs.formatUserOnlineData(int32(cachedData), c)
}

// Cache for User Overview (unified method)
func (dcs *DashboardCacheService) GetCachedUserOverview(c *gin.Context) (*prot.UserOverview, error) {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current":  filter.StartTime.Unix(),
		"start_previous": filter.StartPreviousTime.Unix(),
		"end_current":    filter.EndTime.Unix(),
		"end_previous":   filter.EndPreviousTime.Unix(),
		"program_id":     filter.ProgramId,
		"course_id":      filter.CourseId,
		"school_id":      filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardUserOverviewKey, filters)

	cachedData, err := redis.RememberCache[*dto.UserOverview](cacheKey, DashboardCacheTTL, func() (*dto.UserOverview, error) {
		return dcs.repo.GetUserOverview(filter)
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		dtoData, fallbackErr := dcs.repo.GetUserOverview(filter)
		if fallbackErr != nil {
			return nil, fallbackErr
		}
		return dcs.convertUserOverviewToProt(c, dtoData), nil
	}

	return dcs.convertUserOverviewToProt(c, cachedData), nil
}

// Helper function to convert DTO to prot
func (dcs *DashboardCacheService) convertUserOverviewToProt(c *gin.Context, dto *dto.UserOverview) *prot.UserOverview {
	if dto == nil {
		return nil
	}

	filter := dcs.getFilter(c, Month)

	// Get active users count (15 minutes)
	activeUsersCount, err := dcs.activityLogRepo.GetActiveUsersCount(15, filter.SchoolId, filter.CourseId, filter.ProgramId)
	if err != nil {
		config.Log.Error("GetActiveUsersCount error: ", err)
	}
	config.Log.Info("Active users count (15 min): ", activeUsersCount)
	userOnline := dcs.formatUserOnlineData(int32(activeUsersCount), c)

	return &prot.UserOverview{
		User:       dcs.formatUserRegisterData(dto.User.TotalNumber, dto.User.TotalNumber, dto.User.TotalNumber-int32(dto.User.ChangeValue), filter, c),
		Student:    dcs.formatStudentRegisterData(dto.Student.Total, dto.Student.Total, dto.Student.TotalNumber, c),
		Teacher:    dcs.formatTeacherRegisterData(dto.Teacher.Total, dto.Teacher.Total, dto.Teacher.TotalNumber, c),
		Activity:   dcs.formatActivityData(dto.Activity.TotalNumber, dto.Activity.TotalNumber, dto.Activity.TotalNumber-int32(dto.Activity.ChangeValue), int32(dto.Activity.ChangeValue), filter, c),
		School:     dcs.formatSchoolData(dto.School.Total, int32(dto.School.ChangeValue), dto.School.TotalNumber, filter, c),
		UserOnline: userOnline,
	}
}

// Cache for School
func (dcs *DashboardCacheService) GetCachedSchool(c *gin.Context) *prot.DashboardAdminItem {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"province_code": filter.Address.ProvinceCode,
		"ward_code":     filter.Address.WardCode,
		"program_id":    filter.ProgramId,
		"school_id":     filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardActivityKey, filters)

	cachedData, err := redis.RememberCache[map[string]int32](cacheKey, DashboardCacheTTL, func() (map[string]int32, error) {
		totalActivity, currentActivity, oldActivity, err := dcs.repo.GetSchool(filter)
		if err != nil {
			return map[string]int32{"total": 0, "current": 0, "old": 0}, nil
		}
		return map[string]int32{"total": totalActivity, "current": currentActivity, "old": oldActivity}, nil
	})

	if err != nil {
		// Fallback to direct fetch if cache fails
		totalActivity, currentActivity, oldActivity, _ := dcs.repo.GetSchool(filter)
		return dcs.formatSchoolData(totalActivity, currentActivity, oldActivity, filter, c)
	}

	return dcs.formatSchoolData(cachedData["total"], cachedData["current"], cachedData["old"], filter, c)
}

// Cache cho Course Overview (static data)
func (dcs *DashboardCacheService) GetCachedCourseOverview(c *gin.Context) *prot.CourseOverview {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"course_id":     filter.CourseId,
		"school_id":     filter.SchoolId,
	}

	cacheKey := dcs.generateCacheKey(DashboardCourseOverviewKey, filters)

	data, err := redis.RememberCache[*dto.CourseOverviewDataItem](cacheKey, DashboardCacheTTL, func() (*dto.CourseOverviewDataItem, error) {
		return dcs.repo.GetCourseOverview(filter)
	})

	if err != nil || data == nil {
		data, err = dcs.repo.GetCourseOverview(filter)
		if err != nil || data == nil {
			config.Log.Error("GetCourseOverview: ", err)
			return &prot.CourseOverview{}
		}
	}

	return dcs.formatCourseOverviewData(*data, c)
}

// Cache cho Learning Overview
func (dcs *DashboardCacheService) GetCachedLearningOverview(c *gin.Context) *prot.OverviewLearning {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"course_id":     filter.CourseId,
		"school_id":     filter.SchoolId,
		"class_id":      filter.ClassId,
		"teacher_id":    filter.TeacherId,
		"object_type":   filter.ObjectType,
	}

	cacheKey := dcs.generateCacheKey(DashboardLearningKey, filters)

	data, err := redis.RememberCache[*dto.LearningOverviewDataItem](cacheKey, DashboardCacheTTL, func() (*dto.LearningOverviewDataItem, error) {
		return dcs.repo.GetLearningOverview(filter)
	})

	if err != nil || data == nil {
		data, err = dcs.repo.GetLearningOverview(filter)
		if err != nil || data == nil {
			config.Log.Error("GetLearningOverview: ", err)
			return &prot.OverviewLearning{}
		}
	}

	return dcs.formatLearnigOverviewData(*data, c)
}

// Cache cho Score Distribution
func (dcs *DashboardCacheService) GetCachedScoreDistribution(c *gin.Context) *prot.ScoreDistributionOverview {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"course_id":     filter.CourseId,
		"school_id":     filter.SchoolId,
		"class_id":      filter.ClassId,
		"teacher_id":    filter.TeacherId,
		"object_type":   filter.ObjectType,
	}

	cacheKey := dcs.generateCacheKey(DashboardScoreDistributionKey, filters)

	data, err := redis.RememberCache[*dto.ScoreDistributionOverview](cacheKey, DashboardCacheTTL, func() (*dto.ScoreDistributionOverview, error) {
		return dcs.repo.GetScoreDistribution(filter)
	})

	if err != nil {
		data, err = dcs.repo.GetScoreDistribution(filter)

		if err != nil || data == nil {
			config.Log.Error("GetScoreDistribution: ", err)
			return &prot.ScoreDistributionOverview{}
		}
	}

	return dcs.formatScoreDistributionData(*data, c)
}

// Cache cho System Usage
func (dcs *DashboardCacheService) GetCachedSystemUsage(c *gin.Context) *prot.SystemUsageOverview {
	filter := dcs.getFilter(c, Week)
	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"school_id":     filter.SchoolId,
		"course_id":     filter.CourseId,
		"role_id":       filter.RoleId,
	}

	filterTypes := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"school_id":     filter.SchoolId,
		"course_id":     filter.CourseId,
		"object_type":   filter.ObjectType,
	}

	cacheKey := dcs.generateCacheKey(DashboardSystemUsageKey, filters)
	cacheTypeKey := dcs.generateCacheKey(DashboardSystemUsageTypeKey, filterTypes)

	data, err := redis.RememberCache[*dto.SystemUsageOverview](cacheKey, DashboardCacheTTL, func() (*dto.SystemUsageOverview, error) {
		return dcs.repo.GetSystemUsage(filter)
	})

	if err != nil {
		data, err = dcs.repo.GetSystemUsage(filter)

		if err != nil || data == nil {
			config.Log.Error("GetSystemUsage: ", err)
			return &prot.SystemUsageOverview{}
		}
	}

	dataType, err := redis.RememberCache[*dto.SystemUsageOverview](cacheTypeKey, DashboardCacheTTL, func() (*dto.SystemUsageOverview, error) {
		return dcs.repo.GetSystemUsageType(filter)
	})

	if err != nil {
		dataType, err = dcs.repo.GetSystemUsageType(filter)

		if err != nil || data == nil {
			config.Log.Error("GetSystemUsageType: ", err)
			return &prot.SystemUsageOverview{}
		}
	}

	data.CompleteCount = dataType.CompleteCount
	data.CompletedRate = dataType.CompletedRate

	return dcs.formatSystemUsageData(*data, c)
}

// Cache cho Teacher Performance
func (dcs *DashboardCacheService) GetCachedTeacherPerformance(c *gin.Context) *prot.TeacherPerformanceOverview {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"school_id":     filter.SchoolId,
		"course_id":     filter.CourseId,
		"teacher_id":    filter.TeacherId,
		"object_type":   filter.ObjectType,
	}

	cacheKey := dcs.generateCacheKey(DashboardTeacherPerformanceKey, filters)

	data, err := redis.RememberCache[*dto.TeacherPerformanceOverview](cacheKey, DashboardCacheTTL, func() (*dto.TeacherPerformanceOverview, error) {
		return dcs.repo.GetTeacherPerformance(filter)
	})

	if err != nil {
		data, err = dcs.repo.GetTeacherPerformance(filter)

		if err != nil || data == nil {
			config.Log.Error("GetTeacherPerformance: ", err)
			return &prot.TeacherPerformanceOverview{}
		}
	}
	return dcs.formatTeacherPerformanceData(*data, c)
}

// Cache cho Question Bank
func (dcs *DashboardCacheService) GetCachedQuestionBank(c *gin.Context) *prot.QuestionBankOverview {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
	}

	cacheKey := dcs.generateCacheKey(DashboardQuestionBankKey, filters)

	data, err := redis.RememberCache[*dto.QuestionBankOverview](cacheKey, DashboardCacheTTL, func() (*dto.QuestionBankOverview, error) {
		return dcs.repo.GetQuestionBank(filter)
	})

	if err != nil {
		data, err = dcs.repo.GetQuestionBank(filter)

		if err != nil || data == nil {
			config.Log.Error("GetQuestionBank: ", err)
			return &prot.QuestionBankOverview{}
		}
	}

	return dcs.formatQuestionBankData(*data, c)
}

// Cache cho Risk and Warning
func (dcs *DashboardCacheService) GetCachedRiskWarning(c *gin.Context) *prot.RiskAndWarning {
	filter := dcs.getFilter(c, Month)

	filters := map[string]interface{}{
		"start_current": filter.StartTime.Unix(),
		"end_current":   filter.EndTime.Unix(),
		"program_id":    filter.ProgramId,
		"school_id":     filter.SchoolId,
		"course_id":     filter.CourseId,
		"teacher_id":    filter.TeacherId,
	}

	cacheKey := dcs.generateCacheKey(DashboardRiskWarningKey, filters)

	data, err := redis.RememberCache[*dto.RiskAndWarning](cacheKey, DashboardCacheTTL, func() (*dto.RiskAndWarning, error) {
		return dcs.repo.GetRiskWarning(filter)
	})

	if err != nil {
		data, err = dcs.repo.GetRiskWarning(filter)

		if err != nil || data == nil {
			config.Log.Error("GetRiskWarning: ", err)
			return &prot.RiskAndWarning{}
		}
	}

	return dcs.formatRiskWarningData(*data, c)
}

// Clear tất cả cache của dashboard
func (dcs *DashboardCacheService) ClearAllDashboardCache() error {
	keys := []string{
		DashboardUserRegisterKey,
		DashboardStudentRegisterKey,
		DashboardTeacherRegisterKey,
		DashboardActivityKey,
		DashboardCourseOverviewKey,
		DashboardLearningKey,
		DashboardScoreDistributionKey,
		DashboardSystemUsageKey,
		DashboardTeacherPerformanceKey,
		DashboardQuestionBankKey,
		DashboardRiskWarningKey,
		DashboardSystemUsageTypeKey,
	}

	for _, key := range keys {
		if err := redis.ClearCacheByPrefix(key); err != nil {
			config.Log.Errorf("Failed to clear cache for key '%s': %v", key, err)
			return err
		}
	}

	return nil
}

// Clear cache by type
func (dcs *DashboardCacheService) ClearCacheByType(cacheType string) error {
	switch cacheType {
	case "user_register":
		return redis.ClearCacheByPrefix(DashboardUserRegisterKey)
	case "student_register":
		return redis.ClearCacheByPrefix(DashboardStudentRegisterKey)
	case "teacher_register":
		return redis.ClearCacheByPrefix(DashboardTeacherRegisterKey)
	case "activity":
		return redis.ClearCacheByPrefix(DashboardActivityKey)
	case "course_overview":
		return redis.DeleteCache(DashboardCourseOverviewKey)
	case "learning":
		return redis.DeleteCache(DashboardLearningKey)
	case "score_distribution":
		return redis.DeleteCache(DashboardScoreDistributionKey)
	case "system_usage":
		return redis.ClearCacheByPrefix(DashboardSystemUsageKey)
	case "teacher_performance":
		return redis.DeleteCache(DashboardTeacherPerformanceKey)
	case "question_bank":
		return redis.DeleteCache(DashboardQuestionBankKey)
	case "risk_warning":
		return redis.DeleteCache(DashboardRiskWarningKey)
	default:
		return fmt.Errorf("unknown cache type: %s", cacheType)
	}
}

func (dcs *DashboardCacheService) getTime(timeType string) (time.Time, time.Time) {
	now := time.Now()
	var startCurrent, startPrevious time.Time
	switch timeType {
	case "day":
		startCurrent = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startPrevious = startCurrent.AddDate(0, 0, -1)
	case "week":
		offset := int(now.Weekday()) - 1
		if offset < 0 {
			offset = 6
		}
		startCurrent = time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
		startPrevious = startCurrent.AddDate(0, 0, -7)
	case "24h":
		startCurrent = now.Add(-24 * time.Hour)
		startPrevious = startCurrent.AddDate(0, 0, -1)
	case "month":
		fallthrough
	default:
		startCurrent = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		startPrevious = startCurrent.AddDate(0, -1, 0)
	}

	return startCurrent, startPrevious
}

func (dcs *DashboardCacheService) getFilter(c *gin.Context, userTimeDefault string) dto.FilterDashboardAdmin {
	// === Params ===
	provinceCode := c.DefaultQuery("province_code", "")
	wardCode := c.DefaultQuery("ward_code", "")
	programIdStr := c.DefaultQuery("program_id", "")
	courseIdStr := c.DefaultQuery("course_id", "")
	teacherIdStr := c.DefaultQuery("teacher_id", "")
	schoolIdStr := c.DefaultQuery("school_id", "")
	classIdStr := c.DefaultQuery("class_id", "")
	roleIdStr := c.DefaultQuery("role_id", "")
	timeType := c.DefaultQuery("time_type", userTimeDefault)
	yearStr := c.DefaultQuery("year", "")
	monthStr := c.DefaultQuery("month", "")
	quarterStr := c.DefaultQuery("quarter", "")
	objectTypeStr := c.DefaultQuery("object_type", "all")

	parseInt64 := func(s string) int64 {
		if s == "" {
			return 0
		}
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0
		}
		return val
	}

	programId := parseInt64(programIdStr)
	courseId := parseInt64(courseIdStr)
	teacherId := parseInt64(teacherIdStr)
	schoolId := parseInt64(schoolIdStr)
	classId := parseInt64(classIdStr)
	roleId := parseInt64(roleIdStr)

	roleIdByUser := utils.GetCurrentRoleId(c)
	if roleIdByUser == models.SchoolRoleId {
		schoolId = int64(utils.GetCurrentSchoolId(c))
	}

	baseRepo := base.NewBaseRepository[models.School]()
	schoolIdByRole := baseRepo.GetAdminSchoolId(c)

	if schoolIdByRole > 0 {
		schoolId = int64(schoolIdByRole)
	}

	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)
	quarter, _ := strconv.Atoi(quarterStr)

	now := time.Now()
	endHour, endMin, endSec := 23, 59, 59

	var startCurrent, endCurrent, startPrevious, endPrevious time.Time

	if month > 0 {
		if year == 0 {
			year = now.Year()
		}
		startCurrent = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, now.Location())
		endCurrent = time.Date(year, time.Month(month), daysInMonth(year, int(month)), endHour, endMin, endSec, 0, now.Location())
		startPrevious = startCurrent.AddDate(0, -1, 0)
		endPrevious = time.Date(startPrevious.Year(), startPrevious.Month(), daysInMonth(startPrevious.Year(), int(startPrevious.Month())), endHour, endMin, endSec, 0, now.Location())
		timeType = "by_month"

	} else if quarter > 0 {
		if year == 0 {
			year = now.Year()
		}
		startMonth := (quarter-1)*3 + 1
		startCurrent = time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, now.Location())
		endCurrent = time.Date(year, time.Month(startMonth+2), daysInMonth(year, startMonth+2), endHour, endMin, endSec, 0, now.Location())
		startPrevious = startCurrent.AddDate(0, -3, 0)
		endPrevious = endCurrent.AddDate(0, -3, 0)
		timeType = "by_quarter"

	} else if year > 0 {
		startCurrent = time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
		endCurrent = time.Date(year, 12, 31, endHour, endMin, endSec, 0, now.Location())
		startPrevious = startCurrent.AddDate(-1, 0, 0)
		endPrevious = endCurrent.AddDate(-1, 0, 0)
		timeType = "by_year"

	} else {
		switch timeType {
		case "day":
			startCurrent = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endCurrent = time.Date(now.Year(), now.Month(), now.Day(), endHour, endMin, endSec, 0, now.Location())
			startPrevious = startCurrent.AddDate(0, 0, -1)
			endPrevious = time.Date(startPrevious.Year(), startPrevious.Month(), startPrevious.Day(), endHour, endMin, endSec, 0, now.Location())

		case "week":
			offset := int(now.Weekday()) - 1
			if offset < 0 {
				offset = 6
			}
			startCurrent = time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, now.Location())
			endCurrent = startCurrent.AddDate(0, 0, 6).Add(time.Hour*23 + time.Minute*59 + time.Second*59)
			startPrevious = startCurrent.AddDate(0, 0, -7)
			endPrevious = endCurrent.AddDate(0, 0, -7)

		case "24h":
			startCurrent = now.Add(-24 * time.Hour)
			endCurrent = now
			startPrevious = startCurrent.Add(-24 * time.Hour)
			endPrevious = startCurrent

		case "month", "by_month", "":
			startCurrent = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endCurrent = time.Date(now.Year(), now.Month(), daysInMonth(now.Year(), int(now.Month())), endHour, endMin, endSec, 0, now.Location())
			startPrevious = startCurrent.AddDate(0, -1, 0)
			endPrevious = time.Date(startPrevious.Year(), startPrevious.Month(), daysInMonth(startPrevious.Year(), int(startPrevious.Month())), endHour, endMin, endSec, 0, now.Location())

		default:
			startCurrent = now
			endCurrent = now
			startPrevious = now
			endPrevious = now
		}
	}

	var lastYear, lastMonth, lastQuarter int32
	if month > 0 {
		if month > 1 {
			lastMonth = int32(month - 1)
			lastYear = int32(year)
		} else {
			lastMonth = 12
			lastYear = int32(year) - 1
		}
	} else if quarter > 0 {
		if quarter > 1 {
			lastQuarter = int32(quarter - 1)
			lastYear = int32(year)
		} else {
			lastQuarter = 4
			lastYear = int32(year) - 1
		}
	} else if year > 0 {
		lastYear = int32(year) - 1
	}

	return dto.FilterDashboardAdmin{
		TimeType:          timeType,
		StartTime:         startCurrent,
		EndTime:           endCurrent,
		StartPreviousTime: startPrevious,
		EndPreviousTime:   endPrevious,
		Address: dto.FilterAddress{
			ProvinceCode: provinceCode,
			WardCode:     wardCode,
		},
		ProgramId:   programId,
		CourseId:    courseId,
		TeacherId:   teacherId,
		SchoolId:    schoolId,
		ClassId:     classId,
		RoleId:      roleId,
		Year:        int32(year),
		Month:       int32(month),
		Quarter:     int32(quarter),
		LastYear:    lastYear,
		LastMonth:   lastMonth,
		LastQuarter: lastQuarter,
		ObjectType:  objectTypeStr,
	}
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (dcs *DashboardCacheService) formatUserRegisterData(totalActivity, currentRegister, lastRegister int32, filter dto.FilterDashboardAdmin, c *gin.Context) *prot.DashboardAdminItem {
	userTime := filter.TimeType
	var timeStr string

	switch userTime {
	case "day":
		timeStr = i18n.Localize("dashboard.last_day")
	case "week":
		timeStr = i18n.Localize("dashboard.last_week")
	case "by_quarter":
		quarterStr := utils.FormatMonthOrQuarter(filter.LastQuarter, filter.LastYear, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.by_date", map[string]interface{}{"Date": quarterStr})
	case "by_month":
		monthStr := utils.FormatMonthYear(filter.LastMonth, filter.LastYear, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.by_date", map[string]interface{}{"Date": monthStr})
	case "by_year":
		timeStr = i18n.Localize("dashboard.by_date", map[string]interface{}{"Date": filter.LastYear})
	default:
		timeStr = i18n.Localize("dashboard.last_month")
	}

	var status, message string
	var percent float32
	diff := currentRegister - lastRegister

	if lastRegister == 0 {
		if currentRegister == 0 {
			percent = 0
			status = "same"
		} else {
			percent = 100
			status = "up"
			message = i18n.Localize("dashboard.user_up", map[string]interface{}{"Count": diff, "Time": timeStr})
		}
	} else {
		percent = (float32(diff) / float32(lastRegister)) * 100

		if diff > 0 {
			status = "up"
			message = i18n.Localize("dashboard.user_up", map[string]interface{}{"Count": diff, "Time": timeStr})
		} else if diff < 0 {
			status = "down"
			message = i18n.Localize("dashboard.user_down", map[string]interface{}{"Count": -diff, "Time": timeStr})
		} else {
			status = "same"
		}
	}

	return &prot.DashboardAdminItem{
		Total:       currentRegister,
		TotalNumber: totalActivity,
		ChangeValue: float32(utils.FormatFloat(float64(percent))),
		Title:       i18n.Localize("dashboard.total_users"),
		Message:     message,
		Status:      status,
	}
}

func (dcs *DashboardCacheService) formatStudentRegisterData(totalActivity, currentRegister, oldRegister int32, c *gin.Context) *prot.DashboardAdminItem {
	totalRegister := totalActivity
	var status, message string
	var percent float32
	diff := currentRegister - oldRegister

	if oldRegister == 0 {
		if currentRegister == 0 {
			percent = 0
			status = "same"
		} else {
			percent = 100
			status = "up"
			formattedPercent := utils.FormatPercent(float64(percent), 1)
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		}
	} else {
		percent = (float32(diff) / float32(oldRegister)) * 100
		formattedPercent := utils.FormatPercent(float64(float32(currentRegister)/float32(totalRegister)*100), 1)

		if percent > 0 {
			status = "up"
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		} else if percent < 0 {
			status = "down"
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		} else {
			status = "same"
			formattedPercent := utils.FormatPercent(float64(100), 1)
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		}
	}

	return &prot.DashboardAdminItem{
		Total:       currentRegister,
		TotalNumber: currentRegister,
		ChangeValue: float32(utils.FormatFloat(float64(percent))),
		Title:       i18n.Localize("dashboard.student"),
		Message:     message,
		Status:      status,
	}
}

func (dcs *DashboardCacheService) formatTeacherRegisterData(totalActivity, currentRegister, oldRegister int32, c *gin.Context) *prot.DashboardAdminItem {
	totalRegister := totalActivity

	var status, message string
	var percent float32
	diff := currentRegister - oldRegister

	if oldRegister == 0 {
		if currentRegister == 0 {
			percent = 0
			status = "same"
		} else {
			percent = 100
			status = "up"
			formattedPercent := utils.FormatPercent(float64(percent), 1)
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		}
	} else {
		percent = (float32(diff) / float32(oldRegister)) * 100
		formattedPercent := utils.FormatPercent(float64((float32(currentRegister)/float32(totalRegister))*100), 1)

		if percent > 0 {
			status = "up"
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		} else if percent < 0 {
			status = "down"
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		} else {
			status = "same"
			formattedPercent := utils.FormatPercent(float64(100), 1)
			message = i18n.Localize("dashboard.percent_total", map[string]interface{}{"Percent": formattedPercent})
		}
	}

	return &prot.DashboardAdminItem{
		Total:       currentRegister,
		TotalNumber: currentRegister,
		ChangeValue: float32(utils.FormatFloat(float64(percent))),
		Title:       i18n.Localize("dashboard.teacher"),
		Message:     message,
		Status:      status,
	}
}

func (dcs *DashboardCacheService) formatUserOnlineData(activeUsersCount int32, c *gin.Context) *prot.DashboardAdminItem {
	return &prot.DashboardAdminItem{
		Total:       activeUsersCount,
		TotalNumber: activeUsersCount,
		ChangeValue: 0,
		Title:       i18n.Localize("dashboard.user_online"),
		Message:     i18n.Localize("dashboard.currently_online"),
		Status:      "online",
	}
}

func (dcs *DashboardCacheService) formatSchoolData(totalActivity, currentActivity, oldActivity int32, filter dto.FilterDashboardAdmin, c *gin.Context) *prot.DashboardAdminItem {
	userTime := filter.TimeType
	var timeStr string

	switch userTime {
	case "month":
		timeStr = i18n.Localize("dashboard.new_schools_this_month")
	case "by_quarter":
		quarterStr := utils.FormatMonthOrQuarter(filter.Quarter, filter.Year, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.new_schools_this_date", map[string]interface{}{"Date": quarterStr})
	case "by_month":
		monthStr := utils.FormatMonthYear(filter.Month, filter.Year, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.new_schools_this_date", map[string]interface{}{"Date": monthStr})
	case "by_year":
		timeStr = i18n.Localize("dashboard.new_schools_this_date", map[string]interface{}{"Date": filter.Year})
	default:
		timeStr = i18n.Localize("dashboard.new_schools_this_week")
	}

	var status, message string
	var percent float32

	if oldActivity == 0 {
		if currentActivity == 0 {
			percent = 0
			status = "same"
		} else {
			percent = 100
			status = "up"
			message = i18n.Localize("dashboard.count_schools", map[string]interface{}{"Count": currentActivity, "Time": timeStr})
		}
	} else {
		percent = (float32(currentActivity) / float32(oldActivity)) * 100

		if percent > 0 {
			status = "up"
			message = i18n.Localize("dashboard.count_schools", map[string]interface{}{"Count": currentActivity, "Time": timeStr})
		} else if percent < 0 {
			status = "down"
			message = i18n.Localize("dashboard.count_schools", map[string]interface{}{"Count": currentActivity, "Time": timeStr})
		} else {
			status = "same"
		}
	}

	return &prot.DashboardAdminItem{
		Total:       currentActivity,
		TotalNumber: totalActivity,
		ChangeValue: float32(utils.FormatFloat(float64(percent))),
		Title:       i18n.Localize("dashboard.school"),
		Message:     message,
		Status:      status,
	}
}

func (dcs *DashboardCacheService) formatActivityData(totalActivity, currentActivity, oldActivity, newActivity int32, filter dto.FilterDashboardAdmin, c *gin.Context) *prot.DashboardAdminItem {
	userTime := filter.TimeType
	var timeStr, title string

	switch userTime {
	case "month":
		timeStr = i18n.Localize("dashboard.new_users_this_month")
		title = i18n.Localize("dashboard.active_this_month")
	case "by_quarter":
		quarterStr := utils.FormatMonthOrQuarter(filter.Quarter, filter.Year, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.new_users_this_date", map[string]interface{}{"Date": quarterStr})
		title = i18n.Localize("dashboard.active_this_date", map[string]interface{}{"Date": quarterStr})
	case "by_month":
		monthStr := utils.FormatMonthYear(filter.Month, filter.Year, i18n.GetCurrentLocale())
		timeStr = i18n.Localize("dashboard.new_users_this_date", map[string]interface{}{"Date": monthStr})
		title = i18n.Localize("dashboard.active_this_date", map[string]interface{}{"Date": monthStr})
	case "by_year":
		timeStr = i18n.Localize("dashboard.new_users_this_date", map[string]interface{}{"Date": filter.Year})
		title = i18n.Localize("dashboard.active_this_date", map[string]interface{}{"Date": filter.Year})
	default:
		timeStr = i18n.Localize("dashboard.new_users_this_week")
	}

	var status, message string
	var percent float32

	if oldActivity == 0 {
		if currentActivity == 0 {
			percent = 0
			status = "same"
		} else {
			percent = 100
			status = "up"
			if newActivity > 0 {
				message = i18n.Localize("dashboard.count_users", map[string]interface{}{"Count": newActivity, "Time": timeStr})
			}
		}
	} else {
		percent = (float32(currentActivity-oldActivity) / float32(currentActivity)) * 100

		if percent > 0 {
			status = "up"
		} else if percent < 0 {
			status = "down"
		} else {
			status = "same"
		}

		if newActivity > 0 {
			message = i18n.Localize("dashboard.count_users", map[string]interface{}{"Count": newActivity, "Time": timeStr})
		}
	}

	return &prot.DashboardAdminItem{
		Total:       currentActivity,
		TotalNumber: currentActivity,
		ChangeValue: float32(utils.FormatFloat(float64(percent))),
		Title:       title,
		Message:     message,
		Status:      status,
	}
}

func (dcs *DashboardCacheService) formatCourseOverviewData(dataItem dto.CourseOverviewDataItem, c *gin.Context) *prot.CourseOverview {
	progranText := i18n.Localize("dashboard.program")

	if config.LoadConfig().IsVtg {
		progranText = i18n.Localize("dashboard.subject")
	}

	return &prot.CourseOverview{
		Overviews: []*prot.CourseOverviewItem{
			{Title: progranText, Count: dataItem.Programs},
			{Title: i18n.Localize("dashboard.lesson"), Count: dataItem.Lessons},
			{Title: i18n.Localize("dashboard.homework"), Count: dataItem.Homeworks},
			{Title: i18n.Localize("dashboard.lesson_plan"), Count: dataItem.LessonPlans},
			{Title: i18n.Localize("dashboard.exam"), Count: dataItem.Exams},
			{Title: i18n.Localize("dashboard.banking_question"), Count: dataItem.Questions},
		},
		Rates: []*prot.CourseOverviewRate{
			{
				Title:         i18n.Localize("dashboard.lesson_plan_completed"),
				Total:         dataItem.LessonPlans,
				CompleteTotal: dataItem.LessonPlanCurrent,
				Percent:       utils.SafePercent(dataItem.LessonPlanCurrent, dataItem.LessonPlans),
			},
			{
				Title:         i18n.Localize("dashboard.exercise_done"),
				Total:         dataItem.Homeworks,
				CompleteTotal: dataItem.HomeworkCurrent,
				Percent:       utils.SafePercent(dataItem.HomeworkCurrent, dataItem.Homeworks),
			},
			{
				Title:         i18n.Localize("dashboard.exam_done"),
				Total:         dataItem.Exams,
				CompleteTotal: dataItem.ExamCurrent,
				Percent:       utils.SafePercent(dataItem.ExamCurrent, dataItem.Exams),
			},
			{
				Title:         i18n.Localize("dashboard.the_exam_has_been_graded"),
				Total:         dataItem.UserExams,
				CompleteTotal: dataItem.UserExamCurrent,
				Percent:       utils.SafePercent(dataItem.UserExamCurrent, dataItem.UserExams),
			},
			{
				Title:         i18n.Localize("dashboard.questions_used"),
				Total:         dataItem.Questions,
				CompleteTotal: dataItem.QuestionCurrent,
				Percent:       utils.SafePercent(dataItem.QuestionCurrent, dataItem.Questions),
			},
		},
	}
}

func (dcs *DashboardCacheService) formatLearnigOverviewData(dataItem dto.LearningOverviewDataItem, c *gin.Context) *prot.OverviewLearning {
	scoreWeeks := []*prot.LearningScoreWeek{}

	for _, scoreWeek := range dataItem.LearningScoreWeeks {
		scoreWeeks = append(scoreWeeks, &prot.LearningScoreWeek{
			Id:        scoreWeek.Id,
			WeekLabel: i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": scoreWeek.WeekNumber}),
			AvgScore:  float32(math.Round(float64(scoreWeek.Score)*100) / 100),
		})
	}

	GemsByWeeks := []*prot.GemsByWeek{}
	for _, gemsByWeek := range dataItem.GemsByWeeks {
		GemsByWeeks = append(GemsByWeeks, &prot.GemsByWeek{
			Id:        gemsByWeek.Id,
			WeekLabel: i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": gemsByWeek.WeekNumber}),
			GemCount:  gemsByWeek.Count,
		})
	}

	topHighest := []*prot.StudentScore{}
	for _, highest := range dataItem.TopHighest {
		topHighest = append(topHighest, &prot.StudentScore{
			Id:          highest.Id,
			TypeId:      highest.TypeId,
			TypeUserId:  highest.TypeUserId,
			StudentName: highest.Name,
			TypeName:    highest.TypeName,
			Score:       highest.Score,
			Avatar:      utils.StaticURL(highest.AvatarInfo.Path, models.Storage),
			SchoolName:  highest.SchoolName,
			ClassName:   highest.ClassName,
			Type:        highest.Type,
		})
	}

	topLowest := []*prot.StudentScore{}
	for _, lowest := range dataItem.TopLowest {
		topLowest = append(topLowest, &prot.StudentScore{
			Id:          lowest.Id,
			TypeId:      lowest.TypeId,
			TypeUserId:  lowest.TypeUserId,
			StudentName: lowest.Name,
			TypeName:    lowest.TypeName,
			Score:       lowest.Score,
			Avatar:      utils.StaticURL(lowest.AvatarInfo.Path, models.Storage),
			SchoolName:  lowest.SchoolName,
			ClassName:   lowest.ClassName,
			Type:        lowest.Type,
		})
	}

	totalAverageScore := fmt.Sprintf("%.2f%%", dataItem.TotalAverageScore)
	highestPercent := fmt.Sprintf("%.2f%%", dataItem.HighestPercent)
	lowestPercent := fmt.Sprintf("%.2f%%", dataItem.LowestPercent)
	behaviorPercent := fmt.Sprintf("%.2f%%", dataItem.BehaviorPercent)

	return &prot.OverviewLearning{
		AvgScoresByWeek: scoreWeeks,
		GemsByWeek:      GemsByWeeks,
		TopHighest:      topHighest,
		TopLowest:       topLowest,
		TotalGem:        dataItem.TotalGem,
		AverageScore:    totalAverageScore,
		HighestPercent:  highestPercent,
		LowestPercent:   lowestPercent,
		BehaviorPercent: behaviorPercent,
	}
}

func (dcs *DashboardCacheService) formatScoreDistributionData(dataItem dto.ScoreDistributionOverview, c *gin.Context) *prot.ScoreDistributionOverview {
	distributions := []*prot.ScoreDistributionItem{}

	for _, item := range dataItem.Items {
		distributions = append(distributions, &prot.ScoreDistributionItem{
			Grade:  i18n.Localize("dashboard.grade_number", map[string]interface{}{"Number": item.Number}),
			Scores: item.Scores,
			Counts: item.Counts,
		})
	}
	return &prot.ScoreDistributionOverview{
		Distributions: distributions,
	}
}

func (dcs *DashboardCacheService) formatSystemUsageData(dataItem dto.SystemUsageOverview, c *gin.Context) *prot.SystemUsageOverview {
	weeklyUsage := []*prot.WeeklyUsage{}
	deviceUsages := []*prot.DeviceUsage{}
	averageUsed := []*prot.AverageUsedItem{}

	for _, usage := range dataItem.WeeklyUsages {
		weeklyUsage = append(weeklyUsage, &prot.WeeklyUsage{
			WeekLabel:             i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": usage.Week}),
			AverageDuration:       utils.FormatDurationMinutes(usage.Duration, 2),
			AverageDurationMinute: int32(usage.Duration),
			IsCurrent:             usage.IsCurrent,
		})
	}

	for _, device := range dataItem.DeviceUsages {
		deviceUsages = append(deviceUsages, &prot.DeviceUsage{
			DeviceName: device.Name,
			UserCount:  device.Count,
			Percentage: float32(utils.FormatFloat(float64(device.Percent), 2)),
		})
	}

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "daily_active_users",
		Title: i18n.Localize("dashboard.daily_active_users", map[string]interface{}{
			"Number": dataItem.AverageUsed.DailyActiveUsers,
		}),
		Value: strconv.FormatFloat(dataItem.AverageUsed.DailyActiveUsers, 'f', 0, 64),
	})

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "minutes_per_session",
		Title: i18n.Localize("dashboard.minutes_per_session", map[string]interface{}{
			"Number": dataItem.AverageUsed.MinutesPerSession,
		}),
		Value: strconv.FormatFloat(dataItem.AverageUsed.MinutesPerSession, 'f', 0, 64),
	})

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "returning_rate",
		Title: i18n.Localize("dashboard.returning_rate", map[string]interface{}{
			"Number": dataItem.AverageUsed.ReturningRate,
		}),
		Value: fmt.Sprintf("%.2f%%", dataItem.AverageUsed.ReturningRate*100),
	})

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "engagement_rate",
		Title: i18n.Localize("dashboard.engagement_rate", map[string]interface{}{
			"Number": dataItem.AverageUsed.EngagementRate,
		}),
		Value: fmt.Sprintf("%.2f%%", dataItem.AverageUsed.EngagementRate*100),
	})

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "complete_count",
		Title: i18n.Localize("dashboard.complete_count", map[string]interface{}{
			"Number": dataItem.CompleteCount,
		}),
		Value: fmt.Sprintf("%d", dataItem.CompleteCount),
	})

	averageUsed = append(averageUsed, &prot.AverageUsedItem{
		Key: "complete_rate",
		Title: i18n.Localize("dashboard.complete_rate", map[string]interface{}{
			"Number": dataItem.CompletedRate,
		}),
		Value: fmt.Sprintf("%.2f%%", dataItem.CompletedRate),
	})

	return &prot.SystemUsageOverview{
		WeeklyUsage:  weeklyUsage,
		DeviceUsages: deviceUsages,
		AverageUsed:  averageUsed,
	}
}

func (dcs *DashboardCacheService) formatTeacherPerformanceData(dataItem dto.TeacherPerformanceOverview, c *gin.Context) *prot.TeacherPerformanceOverview {
	weeklyMarkingRates := []*prot.WeeklyMarkingRate{}

	var totalExercise, totalAssigned, gradedSubmission, ungradedSubmission int32

	for _, week := range dataItem.WeeklyMarkingRates {
		ungradedSubmissionExams := []*prot.UngradedSubmission{}
		ungradedSubmissionHomeworks := []*prot.UngradedSubmission{}
		ungradedSubmissionExercises := []*prot.UngradedSubmission{}

		for _, submission := range week.UngradedSubmissions {
			questionIdsStr := submission.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")

			var questionIds []int64

			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			parsed, _ := time.Parse(time.RFC3339, submission.SubmittedAt)
			formatted := parsed.Format("2006-01-02 15:04")

			if submission.Type == "homework" {
				ungradedSubmissionHomeworks = append(ungradedSubmissionHomeworks, &prot.UngradedSubmission{
					TeacherId:     submission.TeacherId,
					ExamId:        submission.Id,
					StudentId:     submission.StudentId,
					TeacherName:   submission.TeacherName,
					StudentName:   submission.StudentName,
					ExamName:      submission.Name,
					QuestionCount: submission.QuestionCount,
					SubmittedAt:   formatted,
					QuestionIds:   questionIds,
				})
			}

			if submission.Type == "exercise" {
				ungradedSubmissionExercises = append(ungradedSubmissionExercises, &prot.UngradedSubmission{
					TeacherId:     submission.TeacherId,
					ExamId:        submission.Id,
					StudentId:     submission.StudentId,
					TeacherName:   submission.TeacherName,
					StudentName:   submission.StudentName,
					ExamName:      submission.Name,
					QuestionCount: submission.QuestionCount,
					SubmittedAt:   formatted,
					QuestionIds:   questionIds,
				})
			}

			if submission.Type == "exam" {
				ungradedSubmissionExams = append(ungradedSubmissionExams, &prot.UngradedSubmission{
					TeacherId:     submission.TeacherId,
					ExamId:        submission.Id,
					StudentId:     submission.StudentId,
					TeacherName:   submission.TeacherName,
					StudentName:   submission.StudentName,
					ExamName:      submission.Name,
					QuestionCount: submission.QuestionCount,
					SubmittedAt:   formatted,
					QuestionIds:   questionIds,
				})
			}
		}

		weeklyMarkingRates = append(weeklyMarkingRates, &prot.WeeklyMarkingRate{
			Id:                          week.Id,
			WeekLabel:                   i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": week.WeekNumber}),
			MarkingPercent:              float32(utils.FormatFloat(float64(week.MarkingPercent), 2)),
			AssignedCount:               week.AssignedCount,
			MarkedCount:                 week.MarkedCount,
			UngradedSubmissionHomeworks: ungradedSubmissionHomeworks,
			UngradedSubmissionExams:     ungradedSubmissionExams,
			UngradedSubmissionExercises: ungradedSubmissionExercises,
		})
	}

	// Format WeeklyAssignAssignmentRate
	weeklyAssignAssignmentRates := []*prot.WeeklyAssignAssignmentRate{}
	for _, week := range dataItem.WeeklyAssignAssignmentRates {
		notAssignAssignmentExams := []*prot.NotAssignAssignment{}
		notAssignAssignmentHomeworks := []*prot.NotAssignAssignment{}
		notAssignAssignmentExercises := []*prot.NotAssignAssignment{}

		for _, notAssign := range week.NotAssignAssignmentExams {
			questionIdsStr := notAssign.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			notAssignAssignmentExams = append(notAssignAssignmentExams, &prot.NotAssignAssignment{
				TeacherId:            notAssign.TeacherId,
				AssignAssignmentId:   notAssign.AssignAssignmentId,
				TeacherName:          notAssign.TeacherName,
				AssignAssignmentName: notAssign.AssignAssignmentName,
				QuestionIds:          questionIds,
				QuestionCount:        notAssign.QuestionCount,
			})
		}

		for _, notAssign := range week.NotAssignAssignmentHomeworks {
			questionIdsStr := notAssign.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			notAssignAssignmentHomeworks = append(notAssignAssignmentHomeworks, &prot.NotAssignAssignment{
				TeacherId:            notAssign.TeacherId,
				AssignAssignmentId:   notAssign.AssignAssignmentId,
				TeacherName:          notAssign.TeacherName,
				AssignAssignmentName: notAssign.AssignAssignmentName,
				QuestionIds:          questionIds,
				QuestionCount:        notAssign.QuestionCount,
			})
		}

		for _, notAssign := range week.NotAssignAssignmentExercises {
			questionIdsStr := notAssign.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			notAssignAssignmentExercises = append(notAssignAssignmentExercises, &prot.NotAssignAssignment{
				TeacherId:            notAssign.TeacherId,
				AssignAssignmentId:   notAssign.AssignAssignmentId,
				TeacherName:          notAssign.TeacherName,
				AssignAssignmentName: notAssign.AssignAssignmentName,
				QuestionIds:          questionIds,
				QuestionCount:        notAssign.QuestionCount,
			})
		}

		totalExercise += week.Total
		totalAssigned += week.AssignAssignment

		weeklyAssignAssignmentRates = append(weeklyAssignAssignmentRates, &prot.WeeklyAssignAssignmentRate{
			Id:                           week.Id,
			WeekLabel:                    i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": week.WeekNumber}),
			Total:                        week.Total,
			AssignAssignment:             week.AssignAssignment,
			NotAssignAssignmentExams:     notAssignAssignmentExams,
			NotAssignAssignmentHomeworks: notAssignAssignmentHomeworks,
			NotAssignAssignmentExercises: notAssignAssignmentExercises,
		})
	}

	// Format WeeklySubmitRate
	weeklySubmitRates := []*prot.WeeklySubmitRate{}
	for _, week := range dataItem.WeeklySubmitRates {
		notGradedExams := []*prot.NotGraded{}
		notGradedHomeworks := []*prot.NotGraded{}
		notGradedExercises := []*prot.NotGraded{}

		// Use maps to track unique items by (not_graded_id, student_id) to avoid duplicates
		examSeenMap := make(map[string]bool)
		homeworkSeenMap := make(map[string]bool)
		exerciseSeenMap := make(map[string]bool)

		for _, notGraded := range week.NotGradedExams {
			uniqueKey := fmt.Sprintf("%d_%d", notGraded.NotGradedId, notGraded.StudentId)
			if examSeenMap[uniqueKey] {
				continue
			}
			examSeenMap[uniqueKey] = true

			questionIdsStr := notGraded.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			parsed, _ := time.Parse(time.RFC3339, notGraded.SubmittedAt)
			formatted := parsed.Format("2006-01-02 15:04")

			notGradedExams = append(notGradedExams, &prot.NotGraded{
				TeacherId:     notGraded.TeacherId,
				NotGradedId:   notGraded.NotGradedId,
				StudentId:     notGraded.StudentId,
				TeacherName:   notGraded.TeacherName,
				StudentName:   notGraded.StudentName,
				NotGradedName: notGraded.NotGradedName,
				QuestionIds:   questionIds,
				QuestionCount: notGraded.QuestionCount,
				SubmittedAt:   formatted,
			})
		}

		for _, notGraded := range week.NotGradedHomeworks {
			uniqueKey := fmt.Sprintf("%d_%d", notGraded.NotGradedId, notGraded.StudentId)
			if homeworkSeenMap[uniqueKey] {
				continue
			}
			homeworkSeenMap[uniqueKey] = true

			questionIdsStr := notGraded.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			parsed, _ := time.Parse(time.RFC3339, notGraded.SubmittedAt)
			formatted := parsed.Format("2006-01-02 15:04")

			notGradedHomeworks = append(notGradedHomeworks, &prot.NotGraded{
				TeacherId:     notGraded.TeacherId,
				NotGradedId:   notGraded.NotGradedId,
				StudentId:     notGraded.StudentId,
				CourseId:      notGraded.CourseId,
				TeacherName:   notGraded.TeacherName,
				StudentName:   notGraded.StudentName,
				CourseName:    notGraded.CourseName,
				LessonName:    notGraded.LessonName,
				NotGradedName: notGraded.NotGradedName,
				QuestionIds:   questionIds,
				QuestionCount: notGraded.QuestionCount,
				SubmittedAt:   formatted,
			})
		}

		for _, notGraded := range week.NotGradedExercises {
			uniqueKey := fmt.Sprintf("%d_%d", notGraded.NotGradedId, notGraded.StudentId)
			if exerciseSeenMap[uniqueKey] {
				continue
			}
			exerciseSeenMap[uniqueKey] = true

			questionIdsStr := notGraded.QuestionIds
			questionIdsStr = strings.Trim(questionIdsStr, "{}")
			var questionIds []int64
			if questionIdsStr != "" {
				parts := strings.Split(questionIdsStr, ",")
				for _, p := range parts {
					id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
					if err == nil {
						questionIds = append(questionIds, id)
					}
				}
			}

			parsed, _ := time.Parse(time.RFC3339, notGraded.SubmittedAt)
			formatted := parsed.Format("2006-01-02 15:04")

			notGradedExercises = append(notGradedExercises, &prot.NotGraded{
				TeacherId:     notGraded.TeacherId,
				NotGradedId:   notGraded.NotGradedId,
				StudentId:     notGraded.StudentId,
				TeacherName:   notGraded.TeacherName,
				StudentName:   notGraded.StudentName,
				NotGradedName: notGraded.NotGradedName,
				QuestionIds:   questionIds,
				QuestionCount: notGraded.QuestionCount,
				SubmittedAt:   formatted,
			})
		}

		gradedSubmission += week.Total
		ungradedSubmission += week.NotGraded

		weeklySubmitRates = append(weeklySubmitRates, &prot.WeeklySubmitRate{
			Id:                 week.Id,
			WeekLabel:          i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": week.WeekNumber}),
			Total:              week.Total,
			NotGraded:          week.NotGraded,
			NotGradedExams:     notGradedExams,
			NotGradedHomeworks: notGradedHomeworks,
			NotGradedExercises: notGradedExercises,
		})
	}

	// Format WeeklyPerformanceOverviewRate
	weeklyPerformanceOverviewRates := []*prot.WeeklyPerformanceOverviewRate{}
	for _, week := range dataItem.WeeklyPerformanceOverviewRates {
		// Format exams (grouped by submitted_id)
		notSubmittedExamsList := []*prot.NotSubmittedExam{}
		for _, exam := range week.NotSubmittedExams {
			notSubmittedExamsList = append(notSubmittedExamsList, &prot.NotSubmittedExam{
				SubmittedId: exam.SubmittedId,
				Name:        exam.Name,
				UserIds:     exam.UserIds,
			})
		}

		// Format homeworks (grouped by submitted_id)
		notSubmittedHomeworksList := []*prot.NotSubmittedGrouped{}
		for _, homework := range week.NotSubmittedHomeworks {
			notSubmittedHomeworksList = append(notSubmittedHomeworksList, &prot.NotSubmittedGrouped{
				SubmittedId:   homework.SubmittedId,
				Name:          homework.Name,
				SubmittedType: homework.SubmittedType,
				StudentIds:    homework.StudentIds,
			})
		}

		// Format exercises (grouped by submitted_id)
		notSubmittedExercisesList := []*prot.NotSubmittedGrouped{}
		for _, exercise := range week.NotSubmittedExercises {
			notSubmittedExercisesList = append(notSubmittedExercisesList, &prot.NotSubmittedGrouped{
				SubmittedId:   exercise.SubmittedId,
				Name:          exercise.Name,
				SubmittedType: exercise.SubmittedType,
				StudentIds:    exercise.StudentIds,
			})
		}

		weeklyPerformanceOverviewRates = append(weeklyPerformanceOverviewRates, &prot.WeeklyPerformanceOverviewRate{
			Id:                      week.Id,
			WeekLabel:               i18n.Localize("dashboard.week_number", map[string]interface{}{"Number": week.WeekNumber}),
			AssignAssignmentPercent: float32(utils.FormatFloat(float64(week.AssignAssignmentPercent), 2)),
			SubmitPercent:           float32(utils.FormatFloat(float64(week.SubmitPercent), 2)),
			NotGradedPercent:        float32(utils.FormatFloat(float64(week.NotGradedPercent), 2)),
			TotalRequired:           week.TotalRequired,
			TotalSubmitted:          week.TotalSubmitted,
			NotSubmittedExams:       notSubmittedExamsList,
			NotSubmittedHomeworks:   notSubmittedHomeworksList,
			NotSubmittedExercises:   notSubmittedExercisesList,
		})
	}

	var ungradedPercent float32

	if ungradedSubmission > 0 {
		ungradedPercent = float32(ungradedSubmission) / float32(gradedSubmission) * 100
	} else {
		ungradedPercent = 0
	}

	return &prot.TeacherPerformanceOverview{
		GradingSummary: &prot.GradingSummary{
			TotalSubmissions:    dataItem.GradingSummary.TotalSubmissions,
			GradedSubmissions:   gradedSubmission,
			UngradedSubmissions: ungradedSubmission,
			UngradedPercent:     float32(utils.FormatFloat(float64(ungradedPercent), 2)),
			TotalExercise:       totalExercise,
			TotalAssigned:       totalAssigned,
		},
		WeeklyMarkingRates:             weeklyMarkingRates,
		WeeklyAssignAssignmentRates:    weeklyAssignAssignmentRates,
		WeeklySubmitRates:              weeklySubmitRates,
		WeeklyPerformanceOverviewRates: weeklyPerformanceOverviewRates,
	}
}

func (dcs *DashboardCacheService) formatQuestionBankData(dataItem dto.QuestionBankOverview, c *gin.Context) *prot.QuestionBankOverview {
	attributes := []*prot.QuestionAttributeItem{}
	types := []*prot.QuestionTypeItem{}

	for _, attribute := range dataItem.Attributes {
		items := []*prot.QuestionDistributionItem{}

		for _, item := range attribute.Items {
			items = append(items, &prot.QuestionDistributionItem{
				Name:    item.Name,
				Count:   item.Count,
				Percent: float32(utils.FormatFloat(float64(item.Percent), 2)),
			})
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].Percent > items[j].Percent
		})

		attributes = append(attributes, &prot.QuestionAttributeItem{
			Name:  attribute.Name,
			Items: items,
		})
	}

	for _, t := range dataItem.Types {
		types = append(types, &prot.QuestionTypeItem{
			Type:    i18n.Localize("question.type." + t.Type),
			Total:   t.Total,
			Percent: float32(utils.FormatFloat(float64(t.Percent), 2)),
		})
	}

	sort.Slice(types, func(i, j int) bool {
		return types[i].Percent > types[j].Percent
	})

	return &prot.QuestionBankOverview{
		Attributes: attributes,
		MediaUsage: &prot.MediaUsage{
			TotalQuestions: dataItem.MediaUsage.TotalQuestions,
			WithAudio:      dataItem.MediaUsage.WithAudio,
			WithImage:      dataItem.MediaUsage.WithImage,
			AudioPercent:   float32(utils.FormatFloat(float64(dataItem.MediaUsage.AudioPercent), 2)),
			ImagePercent:   float32(utils.FormatFloat(float64(dataItem.MediaUsage.ImagePercent), 2)),
		},
		Types: types,
	}
}

func (dcs *DashboardCacheService) formatRiskWarningData(dataItem dto.RiskAndWarning, c *gin.Context) *prot.RiskAndWarning {
	inactiveStudents := []*prot.InactiveStudent{}
	decliningStudents := []*prot.DecliningStudent{}
	slowGradingTeachers := []*prot.SlowGradingTeacher{}

	for _, student := range dataItem.InactiveStudents {
		var lastLoginStr string = student.LastLogin
		var lastLoginFormatted string

		if lastLoginStr != "" {
			if t, err := time.Parse(time.RFC3339, lastLoginStr); err == nil {
				lastLoginFormatted = t.Format("2006-01-02 15:04:05")
			}
		} else {
			lastLoginFormatted = ""
		}

		inactiveStudents = append(inactiveStudents, &prot.InactiveStudent{
			Id:         student.Id,
			Name:       student.Name,
			ClassName:  student.Class,
			LastLogin:  lastLoginFormatted,
			AbsentDays: student.AbsentDays,
			Avatar:     utils.StaticURL(student.AvatarInfo.Path, models.Storage),
		})
	}

	for _, student := range dataItem.DecliningStudents {
		var scores []*prot.ScoreItem

		if student.Score != "" {
			var parsed []dto.ScoreItem
			err := json.Unmarshal([]byte(student.Score), &parsed)
			if err != nil {
				config.Log.Warnf("Cannot parse Score JSON for student %d: %v", student.Id, err)
			} else {
				for _, s := range parsed {
					scores = append(scores, &prot.ScoreItem{
						Id:    s.Id,
						Score: float32(s.Score),
					})
				}
			}
		}

		decliningStudents = append(decliningStudents, &prot.DecliningStudent{
			Id:        student.Id,
			Name:      student.Name,
			ClassName: student.Class,
			Scores:    scores,
			Avatar:    utils.StaticURL(student.AvatarInfo.Path, models.Storage),
		})
	}

	for _, teacher := range dataItem.SlowGradingTeachers {
		var gradedPercent float64 = 100

		if teacher.TotalSubmissions > 0 {
			gradedPercent = float64(teacher.GradedSubmissions) * 100 / float64(teacher.TotalSubmissions)
		}

		slowGradingTeachers = append(slowGradingTeachers, &prot.SlowGradingTeacher{
			Id:           teacher.TeacherId,
			Name:         teacher.TeacherName,
			Graded:       teacher.GradedSubmissions,
			Submitted:    teacher.TotalSubmissions,
			Percent:      float32(utils.FormatFloat(float64(gradedPercent), 2)),
			AvgWaitHours: teacher.AvgWaitHours,
			Avatar:       utils.StaticURL(teacher.AvatarInfo.Path, models.Storage),
		})
	}

	return &prot.RiskAndWarning{
		InactiveStudents:    inactiveStudents,
		DecliningStudents:   decliningStudents,
		SlowGradingTeachers: slowGradingTeachers,
	}
}

