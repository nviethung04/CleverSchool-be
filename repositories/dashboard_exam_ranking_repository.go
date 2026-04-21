package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/requests"
	"sort"
	"strings"
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

func (r *dashboardExamRankingRepository) GetExamRanking(req *requests.DashboardExamRankingRequest, userID int64) ([]dto.DashboardExamRanking, int64, int64, error) {
	var rankings []dto.DashboardExamRanking
	var totalCount int64

	// Query 1: Lấy danh sách exam_ids thỏa mãn điều kiện
	var examIDs []int64
	examQuery := db.ReplicaDB.Table("exams").
		Select("DISTINCT exams.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON erl.lesson_id = lessons.id").
		Joins("JOIN lesson_schedules ON lesson_schedules.lesson_id = lessons.id").
		Joins("JOIN weeks ON lesson_schedules.week_id = weeks.id").
		Joins("JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("JOIN courses ON courses.program_id = chapters.program_id").
		Where("courses.id = ? AND exams.deleted_at IS NULL", req.CourseID)

	// Thực hiện query 1 để lấy exam_ids
	if err := examQuery.Pluck("exams.id", &examIDs).Error; err != nil {
		return nil, 0, 0, err
	}

	// Query 2: Lấy ranking của TẤT CẢ học sinh trong course (không có pagination, không có sorting)
	// Sử dụng LEFT JOIN để hiển thị cả học sinh không có bài kiểm tra
	query := db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			COALESCE(AVG(CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) THEN exam_users.ratio ELSE NULL END), 0) as average_ratio,
			COALESCE(AVG(CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) AND exam_users.time IS NOT NULL THEN exam_users.time ELSE NULL END), 0) as average_time,
			COALESCE(COUNT(DISTINCT CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) THEN exam_users.exam_id END), 0) as number_exams
		`, examIDs, examIDs, examIDs).
		Joins("JOIN user_courses ON users.id = user_courses.user_id").
		Joins("LEFT JOIN exam_users ON users.id = exam_users.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = 3 AND users.deleted_at IS NULL AND user_courses.course_id = ?", req.CourseID).
		Group("users.id, users.name, users.avatar_info")

	// Thêm điều kiện date filter nếu có
	if req.StartDate != nil {
		query = query.Where("DATE(exam_users.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(exam_users.created_at) <= DATE(to_timestamp(?))", req.EndDate)
	}

	// Count total - đếm tất cả học sinh trong course
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, 0, err
	}

	// Lấy TẤT CẢ dữ liệu (không có pagination, không có sorting)
	if err := query.Find(&rankings).Error; err != nil {
		return nil, 0, 0, err
	}

	// Sắp xếp giảm dần theo điểm trung bình (nếu điểm bằng nhau thì theo ID)
	// Sử dụng sort.Slice để sắp xếp trong memory
	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].AverageRatio == rankings[j].AverageRatio {
			// Nếu điểm bằng nhau, sắp xếp theo ID tăng dần
			return rankings[i].StudentID < rankings[j].StudentID
		}
		// Sắp xếp theo điểm giảm dần (cao nhất lên đầu)
		return rankings[i].AverageRatio > rankings[j].AverageRatio
	})

	// Tính current_ranking cho user hiện tại (nếu có)
	var currentRanking int64 = 0
	if userID > 0 {
		for i, ranking := range rankings {
			if ranking.StudentID == userID {
				currentRanking = int64(i + 1) // +1 vì index bắt đầu từ 0
				break
			}
		}
	}

	// Nếu order_by = "asc", đảo ngược list lại
	if strings.ToLower(req.OrderBy) == "asc" {
		// Đảo ngược list để sắp xếp tăng dần
		for i, j := 0, len(rankings)-1; i < j; i, j = i+1, j-1 {
			rankings[i], rankings[j] = rankings[j], rankings[i]
		}
	}

	// Áp dụng pagination từ list đã sắp xếp
	var paginatedRankings []dto.DashboardExamRanking
	if req.Limit > 0 && req.Page > 0 {
		offset := (req.Page - 1) * req.Limit
		end := offset + req.Limit

		// Đảm bảo không vượt quá độ dài của list
		if offset < int(len(rankings)) {
			if end > int(len(rankings)) {
				end = len(rankings)
			}
			paginatedRankings = rankings[offset:end]
		}
	} else {
		// Nếu không có pagination, trả về tất cả
		paginatedRankings = rankings
	}

	return paginatedRankings, totalCount, currentRanking, nil
}

func (r *dashboardExamRankingRepository) GetExamRankingWithoutPagination(req *requests.DashboardExamRankingRequest) ([]dto.DashboardExamRanking, int64, error) {
	var rankings []dto.DashboardExamRanking
	var totalCount int64

	// Query 1: Lấy danh sách exam_ids thỏa mãn điều kiện
	var examIDs []int64
	examQuery := db.ReplicaDB.Table("exams").
		Select("DISTINCT exams.id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON erl.lesson_id = lessons.id").
		Joins("JOIN lesson_schedules ON lesson_schedules.lesson_id = lessons.id").
		Joins("JOIN weeks ON lesson_schedules.week_id = weeks.id").
		Joins("JOIN chapters ON lessons.chapter_id = chapters.id").
		Joins("JOIN courses ON courses.program_id = chapters.program_id").
		Where("courses.id = ? AND exams.deleted_at IS NULL", req.CourseID)

	// Thực hiện query 1 để lấy exam_ids
	if err := examQuery.Pluck("exams.id", &examIDs).Error; err != nil {
		return nil, 0, err
	}

	// Query 2: Lấy ranking của tất cả học sinh trong course (không có pagination)
	// Sử dụng LEFT JOIN để hiển thị cả học sinh không có bài kiểm tra
	query := db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			COALESCE(AVG(CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) THEN exam_users.ratio ELSE NULL END), 0) as average_ratio,
			COALESCE(AVG(CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) AND exam_users.time IS NOT NULL THEN exam_users.time ELSE NULL END), 0) as average_time,
			COALESCE(COUNT(DISTINCT CASE WHEN exam_users.has_manual_scoring = false AND exam_users.exam_id IN (?) THEN exam_users.exam_id END), 0) as number_exams
		`, examIDs, examIDs, examIDs).
		Joins("JOIN user_courses ON users.id = user_courses.user_id").
		Joins("LEFT JOIN exam_users ON users.id = exam_users.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = 3 AND users.deleted_at IS NULL AND user_courses.course_id = ?", req.CourseID).
		Group("users.id, users.name, users.avatar_info")

	// Thêm điều kiện date filter nếu có
	if req.StartDate != nil {
		query = query.Where("DATE(exam_users.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(exam_users.created_at) <= DATE(to_timestamp(?))", req.EndDate)
	}

	// Count total - đếm tất cả học sinh trong course
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// KHÔNG áp dụng pagination - lấy tất cả dữ liệu
	// KHÔNG áp dụng sorting - để tránh ảnh hưởng đến score chart

	err := query.Find(&rankings).Error
	return rankings, totalCount, err
}

func (r *dashboardExamRankingRepository) GetExamScoreChart(req *requests.DashboardExamRankingRequest) ([]dto.ExamScoreChart, error) {
	// Lấy dữ liệu ranking đầy đủ (không bị pagination) để tính score chart
	rankings, _, err := r.GetExamRankingWithoutPagination(req)
	if err != nil {
		return nil, err
	}

	// Khởi tạo map để đếm số lượng học sinh theo từng mức điểm
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

	// Phân loại học sinh theo điểm trung bình
	for _, ranking := range rankings {
		score := ranking.AverageRatio
		var rangeKey string

		switch {
		case score == 0:
			rangeKey = "0-10"
		case score < 10:
			rangeKey = "0-10"
		case score < 20:
			rangeKey = "10-20"
		case score < 30:
			rangeKey = "20-30"
		case score < 40:
			rangeKey = "30-40"
		case score < 50:
			rangeKey = "40-50"
		case score < 60:
			rangeKey = "50-60"
		case score < 70:
			rangeKey = "60-70"
		case score < 80:
			rangeKey = "70-80"
		case score < 90:
			rangeKey = "80-90"
		default:
			rangeKey = "90-100"
		}

		scoreRanges[rangeKey]++
	}

	// Chuyển đổi map thành slice kết quả
	var distributions []dto.ExamScoreChart
	for rangeKey, count := range scoreRanges {
		distributions = append(distributions, dto.ExamScoreChart{
			ScoreRange:   rangeKey,
			StudentCount: count,
		})
	}

	// Sắp xếp theo thứ tự mức điểm
	// Sắp xếp thủ công để đảm bảo thứ tự đúng
	sortedDistributions := make([]dto.ExamScoreChart, 0, len(distributions))
	order := []string{"0-10", "10-20", "20-30", "30-40", "40-50", "50-60", "60-70", "70-80", "80-90", "90-100"}

	for _, rangeKey := range order {
		for _, dist := range distributions {
			if dist.ScoreRange == rangeKey {
				sortedDistributions = append(sortedDistributions, dist)
				break
			}
		}
	}

	return sortedDistributions, nil
}
