package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type DashboardExamRankingRepository interface {
	GetExamRanking(req *requests.DashboardExamRankingRequest, userID int64) ([]dto.DashboardExamRanking, int64, int64, error)
	GetExamScoreChart(req *requests.DashboardExamRankingRequest) ([]dto.ExamScoreChart, error)
	GetExamRankingWithoutPagination(req *requests.DashboardExamRankingRequest) ([]dto.DashboardExamRanking, int64, error)
}

type dashboardExamRankingRepository struct{}

func NewDashboardExamRankingRepository() DashboardExamRankingRepository {
	return &dashboardExamRankingRepository{}
}

func (r *dashboardExamRankingRepository) applyAssignedExamFilters(q *gorm.DB, req *requests.DashboardExamRankingRequest, courseColumn string) *gorm.DB {
	q = q.Joins("JOIN exams e ON e.id = eu.exam_id AND e.deleted_at IS NULL").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = e.id").
		Joins("JOIN lessons l ON l.id = erl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN courses c ON c.program_id = ch.program_id AND c.deleted_at IS NULL").
		Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0").
		Where("(erl.course_id IS NULL OR erl.course_id = 0 OR erl.course_id = " + courseColumn + ")")

	if req.CourseID > 0 {
		q = q.Where("c.id = ?", req.CourseID)
	}
	if req.ExamID > 0 {
		q = q.Where("eu.exam_id = ?", req.ExamID)
	}
	if req.LessonID > 0 {
		q = q.Where("l.id = ?", req.LessonID)
	}
	if req.StartDate != nil {
		q = q.Where("DATE(eu.created_at) >= DATE(to_timestamp(?))", *req.StartDate)
	}
	if req.EndDate != nil {
		q = q.Where("DATE(eu.created_at) <= DATE(to_timestamp(?))", *req.EndDate)
	}
	return q
}

func (r *dashboardExamRankingRepository) buildExamStatsSubquery(req *requests.DashboardExamRankingRequest) *gorm.DB {
	q := db.ReplicaDB.Table("exam_users eu").
		Select(`
			eu.user_id,
			AVG(eu.ratio) as average_ratio,
			AVG(eu.time) as average_time,
			COUNT(DISTINCT eu.exam_id) as number_exams
		`).
		Where("eu.ratio IS NOT NULL")
	return r.applyAssignedExamFilters(q, req, "c.id").Group("eu.user_id")
}

func (r *dashboardExamRankingRepository) buildRankingQuery(req *requests.DashboardExamRankingRequest) *gorm.DB {
	statsSub := r.buildExamStatsSubquery(req)

	return db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			COALESCE(exam_stats.average_ratio, 0) as average_ratio,
			COALESCE(exam_stats.average_time, 0) as average_time,
			COALESCE(exam_stats.number_exams, 0) as number_exams
		`).
		Joins("JOIN user_courses ON users.id = user_courses.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("LEFT JOIN (?) AS exam_stats ON exam_stats.user_id = users.id", statsSub).
		Where("urr.role_id = 3 AND users.deleted_at IS NULL AND user_courses.course_id = ?", req.CourseID)
}

func (r *dashboardExamRankingRepository) GetExamRanking(req *requests.DashboardExamRankingRequest, userID int64) ([]dto.DashboardExamRanking, int64, int64, error) {
	var rankings []dto.DashboardExamRanking
	var totalCount int64

	query := r.buildRankingQuery(req)

	countQuery := db.ReplicaDB.Table("(?) AS ranked_students", query.Session(&gorm.Session{}))
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, 0, err
	}

	if err := query.Find(&rankings).Error; err != nil {
		return nil, 0, 0, err
	}

	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].AverageRatio == rankings[j].AverageRatio {
			return rankings[i].StudentID < rankings[j].StudentID
		}
		return rankings[i].AverageRatio > rankings[j].AverageRatio
	})

	var currentRanking int64 = 0
	if userID > 0 {
		for i, ranking := range rankings {
			if ranking.StudentID == userID {
				currentRanking = int64(i + 1)
				break
			}
		}
	}

	if strings.ToLower(req.OrderBy) == "asc" {
		for i, j := 0, len(rankings)-1; i < j; i, j = i+1, j-1 {
			rankings[i], rankings[j] = rankings[j], rankings[i]
		}
	}

	var paginatedRankings []dto.DashboardExamRanking
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		end := offset + req.Limit

		if offset < len(rankings) {
			if end > len(rankings) {
				end = len(rankings)
			}
			paginatedRankings = rankings[offset:end]
		}
	} else {
		paginatedRankings = rankings
	}

	return paginatedRankings, totalCount, currentRanking, nil
}

func (r *dashboardExamRankingRepository) GetExamRankingWithoutPagination(req *requests.DashboardExamRankingRequest) ([]dto.DashboardExamRanking, int64, error) {
	var rankings []dto.DashboardExamRanking
	var totalCount int64

	query := r.buildRankingQuery(req)

	countQuery := db.ReplicaDB.Table("(?) AS ranked_students", query.Session(&gorm.Session{}))
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.Find(&rankings).Error
	return rankings, totalCount, err
}

func scoreRangeKey(score float64) string {
	switch {
	case score < 10:
		return "0-10"
	case score < 20:
		return "10-20"
	case score < 30:
		return "20-30"
	case score < 40:
		return "30-40"
	case score < 50:
		return "40-50"
	case score < 60:
		return "50-60"
	case score < 70:
		return "60-70"
	case score < 80:
		return "70-80"
	case score < 90:
		return "80-90"
	default:
		return "90-100"
	}
}

func (r *dashboardExamRankingRepository) GetExamScoreChart(req *requests.DashboardExamRankingRequest) ([]dto.ExamScoreChart, error) {
	scoreRanges := map[string]int64{
		"0-10":   0,
		"10-20":  0,
		"20-30":  0,
		"30-40":  0,
		"40-50":  0,
		"50-60":  0,
		"60-70":  0,
		"70-80":  0,
		"80-90":  0,
		"90-100": 0,
	}

	var ratios []float64
	q := db.ReplicaDB.Table("exam_users eu").
		Select("eu.ratio").
		Where("eu.ratio IS NOT NULL")
	q = r.applyAssignedExamFilters(q, req, "c.id")

	if err := q.Pluck("ratio", &ratios).Error; err != nil {
		return nil, err
	}

	for _, ratio := range ratios {
		scoreRanges[scoreRangeKey(ratio)]++
	}

	order := []string{"0-10", "10-20", "20-30", "30-40", "40-50", "50-60", "60-70", "70-80", "80-90", "90-100"}
	sortedDistributions := make([]dto.ExamScoreChart, 0, len(order))
	for _, rangeKey := range order {
		sortedDistributions = append(sortedDistributions, dto.ExamScoreChart{
			ScoreRange:   rangeKey,
			StudentCount: scoreRanges[rangeKey],
		})
	}

	return sortedDistributions, nil
}
