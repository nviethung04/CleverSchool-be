package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"strings"

	"gorm.io/gorm"
)

type DashboardSchoolsRepository interface {
	GetDashboardSchools(selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error)
	GetDashboardSchoolsWithPagination(search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error)
	GetDashboardSchoolsCount(search string) (int64, error)
}

type dashboardSchoolsRepository struct {
	db *gorm.DB
}

func NewDashboardSchoolsRepository() DashboardSchoolsRepository {
	return &dashboardSchoolsRepository{
		db: db.ReplicaDB,
	}
}

// GetDashboardSchools lấy danh sách dashboard schools với điều kiện cố định
func (r *dashboardSchoolsRepository) GetDashboardSchools(selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error) {
	var schools []dto.DashboardSchoolsResponse

	// Tính các khoảng thời gian
	sep15 := "2025-09-15"

	// Lấy tất cả school_id từ bảng schools
	var schoolIDs []int64
	err := r.db.Table("schools").Select("id").Find(&schoolIDs).Error
	if err != nil {
		return schools, err
	}

	// Query 1: Lấy dữ liệu từ 8/9 (active_teachers, total_teachers)
	var sep8DataList []struct {
		SchoolID       int64 `json:"school_id"`
		TotalTeachers  int64 `json:"total_teachers"`
		ActiveTeachers int64 `json:"active_teachers"`
	}
	teacherQuery := r.db.Table("dashboard_report_schools").
		Select("school_id, total_teachers, active_teachers").
		Where("start_date = ?", sep15)
	if selectedEndDate != "" {
		teacherQuery = teacherQuery.Where("end_date = ?", selectedEndDate)
	}
	teacherQuery.Find(&sep8DataList)

	// Query 2: Lấy dữ liệu từ 15/9 (active_students, total_students, students_completed_homework)
	var sep15DataList []struct {
		SchoolID                  int64 `json:"school_id"`
		TotalStudents             int64 `json:"total_students"`
		ActiveStudents            int64 `json:"active_students"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
	}
	studentQuery := r.db.Table("dashboard_report_schools").
		Select("school_id, total_students, active_students, students_completed_homework").
		Where("start_date = ?", sep15)
	if selectedEndDate != "" {
		studentQuery = studentQuery.Where("end_date = ?", selectedEndDate)
	}
	studentQuery.Find(&sep15DataList)

	// Query 3: Lấy dữ liệu từ selected week (dựa trên start_date và end_date từ param)
	var selectedWeekDataList []struct {
		SchoolID                  int64 `json:"school_id"`
		StudentsCompletedHomework int64 `json:"students_completed_homework"`
		ActiveStudents            int64 `json:"active_students"`
		ActiveTeachers            int64 `json:"active_teachers"`
	}
	r.db.Table("dashboard_report_schools").
		Select("school_id, students_completed_homework, active_students, active_teachers").
		Where("start_date = ? AND end_date = ?", selectedStartDate, selectedEndDate).
		Find(&selectedWeekDataList)

	// Query 4: Lấy tên schools
	var schoolInfo []struct {
		SchoolID   int64  `json:"school_id"`
		SchoolName string `json:"school_name"`
	}
	r.db.Table("schools").
		Select("id as school_id, name as school_name").
		Find(&schoolInfo)

	// Tạo map để lookup nhanh
	sep8Map := make(map[int64]struct {
		TotalTeachers  int64
		ActiveTeachers int64
	})
	for _, data := range sep8DataList {
		sep8Map[data.SchoolID] = struct {
			TotalTeachers  int64
			ActiveTeachers int64
		}{
			TotalTeachers:  data.TotalTeachers,
			ActiveTeachers: data.ActiveTeachers,
		}
	}

	sep15Map := make(map[int64]struct {
		TotalStudents             int64
		ActiveStudents            int64
		StudentsCompletedHomework int64
	})
	for _, data := range sep15DataList {
		sep15Map[data.SchoolID] = struct {
			TotalStudents             int64
			ActiveStudents            int64
			StudentsCompletedHomework int64
		}{
			TotalStudents:             data.TotalStudents,
			ActiveStudents:            data.ActiveStudents,
			StudentsCompletedHomework: data.StudentsCompletedHomework,
		}
	}

	selectedWeekMap := make(map[int64]struct {
		StudentsCompletedHomework int64
		ActiveStudents            int64
		ActiveTeachers            int64
	})
	for _, data := range selectedWeekDataList {
		selectedWeekMap[data.SchoolID] = struct {
			StudentsCompletedHomework int64
			ActiveStudents            int64
			ActiveTeachers            int64
		}{
			StudentsCompletedHomework: data.StudentsCompletedHomework,
			ActiveStudents:            data.ActiveStudents,
			ActiveTeachers:            data.ActiveTeachers,
		}
	}

	schoolInfoMap := make(map[int64]string)
	for _, info := range schoolInfo {
		schoolInfoMap[info.SchoolID] = info.SchoolName
	}

	// Gộp dữ liệu theo school_id
	for _, schoolID := range schoolIDs {
		school := dto.DashboardSchoolsResponse{
			SchoolID: schoolID,
		}

		// Lấy tên school
		if schoolName, exists := schoolInfoMap[schoolID]; exists {
			school.SchoolName = schoolName
		}

		// Dữ liệu từ 8/9
		if sep8Data, exists := sep8Map[schoolID]; exists {
			school.TotalTeachers = sep8Data.TotalTeachers
			school.ActiveTeachersFromSep8 = sep8Data.ActiveTeachers
		}

		// Dữ liệu từ 15/9
		if sep15Data, exists := sep15Map[schoolID]; exists {
			school.TotalStudents = sep15Data.TotalStudents
			school.ActiveStudentsFromSep15 = sep15Data.ActiveStudents
			school.StudentsCompletedHomeworkFromSep15 = sep15Data.StudentsCompletedHomework
		}

		// Dữ liệu từ selected week
		if selectedWeekData, exists := selectedWeekMap[schoolID]; exists {
			school.StudentsCompletedHomeworkSelectedWeek = selectedWeekData.StudentsCompletedHomework
			school.ActiveStudentsSelectedWeek = selectedWeekData.ActiveStudents
			school.ActiveTeachersSelectedWeek = selectedWeekData.ActiveTeachers
		} else {
			// Nếu không có dữ liệu selected week, set -1
			school.StudentsCompletedHomeworkSelectedWeek = -1
			school.ActiveStudentsSelectedWeek = -1
			school.ActiveTeachersSelectedWeek = -1
		}

		schools = append(schools, school)
	}

	return schools, nil
}

// GetDashboardSchoolsWithPagination lấy danh sách dashboard schools với phân trang và tìm kiếm
func (r *dashboardSchoolsRepository) GetDashboardSchoolsWithPagination(search string, page, limit int, selectedStartDate, selectedEndDate string) ([]dto.DashboardSchoolsResponse, error) {
	// Lấy tất cả schools trước
	allSchools, err := r.GetDashboardSchools(selectedStartDate, selectedEndDate)
	if err != nil {
		return nil, err
	}

	// Filter by search if provided
	var filteredSchools []dto.DashboardSchoolsResponse
	if search != "" {
		for _, school := range allSchools {
			if strings.Contains(strings.ToLower(school.SchoolName), strings.ToLower(search)) {
				filteredSchools = append(filteredSchools, school)
			}
		}
	} else {
		filteredSchools = allSchools
	}

	// Apply pagination
	offset := (page - 1) * limit
	end := offset + limit

	if offset >= len(filteredSchools) {
		return []dto.DashboardSchoolsResponse{}, nil
	}

	if end > len(filteredSchools) {
		end = len(filteredSchools)
	}

	return filteredSchools[offset:end], nil
}

// GetDashboardSchoolsCount đếm tổng số dashboard schools với tìm kiếm
func (r *dashboardSchoolsRepository) GetDashboardSchoolsCount(search string) (int64, error) {
	var count int64
	
	baseQuery := r.db.Table("schools")
	
	// Filter by search if provided
	if search != "" {
		baseQuery = baseQuery.Where("name ILIKE ?", "%"+search+"%")
	}
	
	err := baseQuery.Count(&count).Error
	return count, err
}
