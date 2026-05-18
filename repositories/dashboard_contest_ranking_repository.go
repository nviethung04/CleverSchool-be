package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/requests"
	"sort"
	"strings"
)

type DashboardContestRankingRepository interface {
	GetContestRanking(req *requests.DashboardContestRankingRequest, userID int64) ([]dto.DashboardContestRanking, int64, int64, error)
	GetContestScoreChart(req *requests.DashboardContestRankingRequest) ([]dto.ContestScoreChart, error)
	GetContestRankingWithoutPagination(req *requests.DashboardContestRankingRequest) ([]dto.DashboardContestRanking, int64, error)
}

type dashboardContestRankingRepository struct{}

func NewDashboardContestRankingRepository() DashboardContestRankingRepository {
	return &dashboardContestRankingRepository{}
}

func (r *dashboardContestRankingRepository) GetContestRanking(req *requests.DashboardContestRankingRequest, userID int64) ([]dto.DashboardContestRanking, int64, int64, error) {
	var rankings []dto.DashboardContestRanking
	var totalCount int64

	// Query 1: Lấy danh sách contest_round_ids thỏa mãn điều kiện
	var contestRoundIDs []int64
	contestRoundQuery := db.ReplicaDB.Table("contest_rounds").
		Select("DISTINCT contest_rounds.id").
		Joins("JOIN contests ON contest_rounds.contest_id = contests.id").
		Joins("JOIN contest_ref_courses crc ON crc.contest_id = contests.id").
		Where("crc.course_id = ? AND contest_rounds.deleted_at IS NULL", req.CourseID)

	// Thêm filter theo contest_id nếu có
	if req.ContestID > 0 {
		contestRoundQuery = contestRoundQuery.Where("contests.id = ?", req.ContestID)
	}

	// Thêm filter theo contest_round_id nếu có
	if req.ContestRoundID > 0 {
		contestRoundQuery = contestRoundQuery.Where("contest_rounds.id = ?", req.ContestRoundID)
	}

	// Thực hiện query 1 để lấy contest_round_ids
	if err := contestRoundQuery.Pluck("contest_rounds.id", &contestRoundIDs).Error; err != nil {
		return nil, 0, 0, err
	}

	// Nếu không có contest round nào, trả về empty
	if len(contestRoundIDs) == 0 {
		return []dto.DashboardContestRanking{}, 0, 0, nil
	}

	// Query 2: Lấy ranking của TẤT CẢ học sinh (không có pagination, không có sorting)
	// Sử dụng LEFT JOIN để hiển thị cả học sinh không có bài thi
	query := db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.ratio ELSE NULL END), 0) as average_ratio,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.score ELSE NULL END), 0) as average_score,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) AND contest_round_users.time IS NOT NULL THEN contest_round_users.time ELSE NULL END), 0) as average_time,
			COALESCE(COUNT(DISTINCT CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.contest_round_id END), 0) as number_rounds
		`, contestRoundIDs, contestRoundIDs, contestRoundIDs, contestRoundIDs).
		Joins("LEFT JOIN contest_round_users ON users.id = contest_round_users.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("LEFT JOIN user_classes uc ON uc.user_id = users.id").
		Joins("LEFT JOIN classes c ON c.id = uc.class_id").
		Joins("LEFT JOIN schools s ON s.id = users.school_id").
		Joins("LEFT JOIN provinces p ON p.code = s.ward_code").
		Where("urr.role_id = 3 AND users.deleted_at IS NULL").
		Group("users.id, users.name, users.avatar_info")

	// Thêm điều kiện course filter nếu có
	if req.CourseID > 0 {
		query = query.Joins("JOIN user_courses ON users.id = user_courses.user_id").
			Where("user_courses.course_id = ?", req.CourseID)
	}

	// Thêm điều kiện class filter nếu có
	if req.ClassID > 0 {
		query = query.Where("c.id = ?", req.ClassID)
	}

	// Thêm điều kiện school filter nếu có
	if req.SchoolID > 0 {
		query = query.Where("s.id = ?", req.SchoolID)
	}

	// Thêm điều kiện province filter nếu có
	if req.ProvinceCode != "" {
		query = query.Where("p.code = ?", req.ProvinceCode)
	}

	// Thêm điều kiện date filter nếu có
	if req.StartDate != nil {
		query = query.Where("DATE(contest_round_users.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(contest_round_users.created_at) <= DATE(to_timestamp(?))", req.EndDate)
	}

	// Count total - đếm tất cả học sinh
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
	var paginatedRankings []dto.DashboardContestRanking
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

func (r *dashboardContestRankingRepository) GetContestRankingWithoutPagination(req *requests.DashboardContestRankingRequest) ([]dto.DashboardContestRanking, int64, error) {
	var rankings []dto.DashboardContestRanking
	var totalCount int64

	// Query 1: Lấy danh sách contest_round_ids thỏa mãn điều kiện
	var contestRoundIDs []int64
	contestRoundQuery := db.ReplicaDB.Table("contest_rounds").
		Select("DISTINCT contest_rounds.id").
		Joins("JOIN contests ON contest_rounds.contest_id = contests.id").
		Joins("JOIN contest_ref_courses crc ON crc.contest_id = contests.id").
		Where("crc.course_id = ? AND contest_rounds.deleted_at IS NULL", req.CourseID)

	// Thêm filter theo contest_id nếu có
	if req.ContestID > 0 {
		contestRoundQuery = contestRoundQuery.Where("contests.id = ?", req.ContestID)
	}

	// Thêm filter theo contest_round_id nếu có
	if req.ContestRoundID > 0 {
		contestRoundQuery = contestRoundQuery.Where("contest_rounds.id = ?", req.ContestRoundID)
	}

	// Thực hiện query 1 để lấy contest_round_ids
	if err := contestRoundQuery.Pluck("contest_rounds.id", &contestRoundIDs).Error; err != nil {
		return nil, 0, err
	}

	// Nếu không có contest round nào, trả về empty
	if len(contestRoundIDs) == 0 {
		return []dto.DashboardContestRanking{}, 0, nil
	}

	// Query 2: Lấy ranking của tất cả học sinh (không có pagination)
	// Sử dụng LEFT JOIN để hiển thị cả học sinh không có bài thi
	query := db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.ratio ELSE NULL END), 0) as average_ratio,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.score ELSE NULL END), 0) as average_score,
			COALESCE(AVG(CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) AND contest_round_users.time IS NOT NULL THEN contest_round_users.time ELSE NULL END), 0) as average_time,
			COALESCE(COUNT(DISTINCT CASE WHEN contest_round_users.has_manual_scoring = false AND contest_round_users.contest_round_id IN (?) THEN contest_round_users.contest_round_id END), 0) as number_rounds
		`, contestRoundIDs, contestRoundIDs, contestRoundIDs, contestRoundIDs).
		Joins("LEFT JOIN contest_round_users ON users.id = contest_round_users.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("LEFT JOIN user_classes uc ON uc.user_id = users.id").
		Joins("LEFT JOIN classes c ON c.id = uc.class_id").
		Joins("LEFT JOIN schools s ON s.id = users.school_id").
		Joins("LEFT JOIN provinces p ON p.code = s.ward_code").
		Where("urr.role_id = 3 AND users.deleted_at IS NULL").
		Group("users.id, users.name, users.avatar_info")

	// Thêm điều kiện course filter nếu có
	if req.CourseID > 0 {
		query = query.Joins("JOIN user_courses ON users.id = user_courses.user_id").
			Where("user_courses.course_id = ?", req.CourseID)
	}

	// Thêm điều kiện class filter nếu có
	if req.ClassID > 0 {
		query = query.Where("c.id = ?", req.ClassID)
	}

	// Thêm điều kiện school filter nếu có
	if req.SchoolID > 0 {
		query = query.Where("s.id = ?", req.SchoolID)
	}

	// Thêm điều kiện province filter nếu có
	if req.ProvinceCode != "" {
		query = query.Where("p.code = ?", req.ProvinceCode)
	}

	// Thêm điều kiện date filter nếu có
	if req.StartDate != nil {
		query = query.Where("DATE(contest_round_users.created_at) >= DATE(to_timestamp(?))", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("DATE(contest_round_users.created_at) <= DATE(to_timestamp(?))", req.EndDate)
	}

	// Count total - đếm tất cả học sinh
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// KHÔNG áp dụng pagination - lấy tất cả dữ liệu
	// KHÔNG áp dụng sorting - để tránh ảnh hưởng đến score chart

	err := query.Find(&rankings).Error
	return rankings, totalCount, err
}

func (r *dashboardContestRankingRepository) GetContestScoreChart(req *requests.DashboardContestRankingRequest) ([]dto.ContestScoreChart, error) {
	// Lấy dữ liệu ranking đầy đủ (không bị pagination) để tính score chart
	rankings, _, err := r.GetContestRankingWithoutPagination(req)
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
	var distributions []dto.ContestScoreChart
	for rangeKey, count := range scoreRanges {
		distributions = append(distributions, dto.ContestScoreChart{
			ScoreRange:   rangeKey,
			StudentCount: count,
		})
	}

	// Sắp xếp theo thứ tự mức điểm
	// Sắp xếp thủ công để đảm bảo thứ tự đúng
	sortedDistributions := make([]dto.ContestScoreChart, 0, len(distributions))
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
