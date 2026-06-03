package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"

	"gorm.io/gorm"
)

type DashboardTeacherHomeworkOverviewRepository interface {
	GetOverview(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherHomeworkOverviewStatsRequest) (*dto.DashboardTeacherHomeworkOverviewStats, error)
}

type dashboardTeacherHomeworkOverviewRepository struct{}

func NewDashboardTeacherHomeworkOverviewRepository() DashboardTeacherHomeworkOverviewRepository {
	return &dashboardTeacherHomeworkOverviewRepository{}
}

func (r *dashboardTeacherHomeworkOverviewRepository) GetOverview(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherHomeworkOverviewStatsRequest) (*dto.DashboardTeacherHomeworkOverviewStats, error) {
	var overview dto.DashboardTeacherHomeworkOverviewStats

	// Helper function to create base query with direct homework_users filtering
	createBaseQuery := func() *gorm.DB {
		query := db.ReplicaDB.Table("homework_users hu")

		// Apply filters directly on homework_users table
		if req.HomeworkID > 0 {
			query = query.Where("hu.homework_id = ?", req.HomeworkID)
		}
		if req.UserID > 0 {
			query = query.Where("hu.user_id = ?", req.UserID)
		}
		if req.StartDate > 0 && req.EndDate > 0 {
			// Lọc homework trong khoảng thời gian từ start_date đến end_date
			query = query.Where("DATE(hu.created_at) >= DATE(to_timestamp(?)) AND DATE(hu.created_at) <= DATE(to_timestamp(?))", req.StartDate, req.EndDate)
		} else if req.StartDate > 0 {
			query = query.Where("DATE(hu.created_at) >= DATE(to_timestamp(?))", req.StartDate)
		} else if req.EndDate > 0 {
			query = query.Where("DATE(hu.created_at) <= DATE(to_timestamp(?))", req.EndDate)
		}

		// Join với users và homeworks để filter deleted records
		query = query.Joins("JOIN users u ON hu.user_id = u.id").
			Joins("LEFT JOIN homeworks h ON hu.homework_id = h.id").
			Where("u.deleted_at IS NULL").
			Where("h.deleted_at IS NULL")

		// Join với homework_ref_lessons để có thể filter theo lesson_id và course_id
		query = query.Joins("LEFT JOIN homework_ref_lessons hrl ON hrl.homework_id = hu.homework_id")
		
		// Filter theo lesson_id nếu có
		if req.LessonID > 0 {
			query = query.Where("hrl.lesson_id = ?", req.LessonID)
		}

		// Filter theo course_id nếu có - chỉ lấy học sinh thuộc course đó
		if req.CourseID > 0 {
			query = query.Joins("LEFT JOIN user_courses uc ON uc.user_id = hu.user_id AND uc.course_id = ?", req.CourseID).
				Where("uc.course_id IS NOT NULL")
		}

		// Nếu chỉ lấy homework cùng khóa (role_id = 2 hoặc 3)
		if onlyUserCourses && userID > 0 {
			query = query.Joins("LEFT JOIN user_courses uc2 ON hrl.course_id = uc2.course_id").
				Where("uc2.user_id = ?", userID)
		}

		return query
	}

	// Count total submitted (distinct theo homework_id, user_id, lesson_id)
	var totalSubmitted int64
	baseQuery := createBaseQuery()
	if err := baseQuery.Select("COUNT(DISTINCT (hu.homework_id, hu.user_id, hu.lesson_id))").Scan(&totalSubmitted).Error; err != nil {
		return nil, err
	}
	overview.TotalSubmitted = totalSubmitted

	// Count total scored (status_scoring = 2) - using status_scoring column
	scoredQuery := createBaseQuery()
	if err := scoredQuery.Where("hu.status_scoring = ?", 2).
		Select("COUNT(DISTINCT (hu.homework_id, hu.user_id, hu.lesson_id))").Scan(&overview.TotalScored).Error; err != nil {
		return nil, err
	}

	// Count total unscored (status_scoring = 1) - using status_scoring column
	unscoredQuery := createBaseQuery()
	if err := unscoredQuery.Where("hu.status_scoring = ?", 1).
		Select("COUNT(DISTINCT (hu.homework_id, hu.user_id, hu.lesson_id))").Scan(&overview.TotalUnscored).Error; err != nil {
		return nil, err
	}

	// Count total no manual scoring (status_scoring = 0) - using status_scoring column
	noManualScoringQuery := createBaseQuery()
	if err := noManualScoringQuery.Where("hu.status_scoring = ?", 0).
		Select("COUNT(DISTINCT (hu.homework_id, hu.user_id, hu.lesson_id))").Scan(&overview.TotalNoManualScoring).Error; err != nil {
		return nil, err
	}

	// Calculate completion rate: total_scored / (total_scored + total_unscored)
	totalScoredAndUnscored := overview.TotalScored + overview.TotalUnscored
	if totalScoredAndUnscored > 0 {
		overview.CompletionRate = float64(overview.TotalScored) / float64(totalScoredAndUnscored) * 100
	} else {
		overview.CompletionRate = 0
	}

	return &overview, nil
}