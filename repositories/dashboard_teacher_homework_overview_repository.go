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

// ScoringData dùng cho query chi tiết scoring (debug/kiểm tra).
type ScoringData struct {
	HomeworkId    int64 `json:"homework_id"`
	UserId        int64 `json:"user_id"`
	LessonId      int64 `json:"lesson_id"`
	StatusScoring int32 `json:"status_scoring"`
}

type dashboardTeacherHomeworkOverviewRepository struct{}

func NewDashboardTeacherHomeworkOverviewRepository() DashboardTeacherHomeworkOverviewRepository {
	return &dashboardTeacherHomeworkOverviewRepository{}
}

func (r *dashboardTeacherHomeworkOverviewRepository) GetOverview(userID int64, onlyUserCourses bool, req *requests.DashboardTeacherHomeworkOverviewStatsRequest) (*dto.DashboardTeacherHomeworkOverviewStats, error) {
	var overview dto.DashboardTeacherHomeworkOverviewStats

	// Nếu không truyền course_id, lấy danh sách course_id của user từ bảng user_courses
	var userCourseIDs []int64
	if req.CourseID == 0 && userID > 0 {
		if err := db.ReplicaDB.Table("user_courses").
			Select("course_id").
			Where("user_id = ?", userID).
			Scan(&userCourseIDs).Error; err != nil {
			return nil, err
		}

		// Nếu user không có khóa nào thì trả về overview rỗng
		if len(userCourseIDs) == 0 {
			return &overview, nil
		}
	}

	// Helper function để tạo base query thống nhất giữa các API
	createBaseQuery := func() *gorm.DB {
		query := db.ReplicaDB.Table("homework_users hu").
			Joins("JOIN users u ON hu.user_id = u.id").
			Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
			Joins("JOIN homeworks h ON hu.homework_id = h.id").
			Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = hu.homework_id AND hrl.lesson_id = hu.lesson_id").
			Joins("JOIN lessons l ON l.id = hu.lesson_id").
			Joins("JOIN lesson_schedules ls ON ls.course_id = hrl.course_id AND ls.lesson_id = hrl.lesson_id").
			Joins("JOIN weeks w ON w.id = ls.week_id").
			Joins("JOIN user_courses student_courses ON student_courses.user_id = hu.user_id AND student_courses.course_id = hrl.course_id").
			Where("u.deleted_at IS NULL").
			Where("h.deleted_at IS NULL").
			Where("urr.role_id = ?", 3)

		// Filter theo course_id
		if req.CourseID > 0 {
			query = query.
				Where("student_courses.course_id = ?", req.CourseID).
				Where("hrl.course_id = ?", req.CourseID)
		} else if len(userCourseIDs) > 0 {
			query = query.
				Where("student_courses.course_id IN (?)", userCourseIDs).
				Where("hrl.course_id IN (?)", userCourseIDs)
		}

		if req.HomeworkID > 0 {
			query = query.Where("hu.homework_id = ?", req.HomeworkID)
		}
		if req.UserID > 0 {
			query = query.Where("hu.user_id = ?", req.UserID)
		}
		if req.LessonID > 0 {
			query = query.Where("hu.lesson_id = ?", req.LessonID)
		}
		if req.StartDate > 0 && req.EndDate > 0 {
			query = query.Where("hu.updated_at BETWEEN to_timestamp(?) AND to_timestamp(?)", req.StartDate, req.EndDate)
		} else if req.StartDate > 0 {
			query = query.Where("hu.updated_at >= to_timestamp(?)", req.StartDate)
		} else if req.EndDate > 0 {
			query = query.Where("hu.updated_at <= to_timestamp(?)", req.EndDate)
		}

		if onlyUserCourses && userID > 0 {
			query = query.Joins("JOIN user_courses teacher_courses ON teacher_courses.course_id = hrl.course_id").
				Where("teacher_courses.user_id = ?", userID)
		}

		return query
	}

	// Lấy toàn bộ bản ghi distinct (homework_id, user_id, lesson_id, status_scoring) rồi tính overview trong memory
	var scoringDataList []ScoringData
	baseQuery := createBaseQuery()
	if err := baseQuery.Select("DISTINCT hu.homework_id, hu.user_id, hu.lesson_id, hu.status_scoring").Scan(&scoringDataList).Error; err != nil {
		return nil, err
	}

	// Dedupe theo (homework_id, user_id, lesson_id) — join có thể sinh trùng dòng
	type scoringKey struct{ HomeworkId, UserId, LessonId int64 }
	dedupe := make(map[scoringKey]int32)
	for _, row := range scoringDataList {
		k := scoringKey{row.HomeworkId, row.UserId, row.LessonId}
		if _, exists := dedupe[k]; !exists {
			dedupe[k] = row.StatusScoring
		}
	}

	for _, status := range dedupe {
		overview.TotalSubmitted++
		switch status {
		case 0:
			overview.TotalNoManualScoring++
		case 1:
			overview.TotalUnscored++
		case 2:
			overview.TotalScored++
		}
	}

	// Completion rate = scored / (scored + unscored)
	totalScoredAndUnscored := overview.TotalScored + overview.TotalUnscored
	if totalScoredAndUnscored > 0 {
		overview.CompletionRate = float64(overview.TotalScored) / float64(totalScoredAndUnscored) * 100
	} else {
		overview.CompletionRate = 0
	}

	return &overview, nil
}
