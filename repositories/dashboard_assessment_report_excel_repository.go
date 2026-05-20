package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// StudentReportData chứa dữ liệu học sinh cho report Excel
type StudentReportData struct {
	StudentID     int64  `json:"student_id"`
	StudentName   string `json:"student_name"`
	ClassName     string `json:"class_name"`
	ClassMainName string `json:"class_main_name"`
	GradeNameVN   string `json:"grade_name_vn"`
	GradeID       int64  `json:"grade_id"`
}

// AssessmentInfo chứa thông tin assessment
type AssessmentInfo struct {
	ID          int64  `json:"id"`
	SubjectID   int64  `json:"subject_id"`
	SubjectName string `json:"subject_name"`
}

// AssessmentScoreMap map[student_id][assessment_id]total_score
type AssessmentScoreMap map[int64]map[int64]*float64

type DashboardAssessmentReportExcelRepository interface {
	GetSchoolsByIDs(schoolIDs []int64) ([]models.School, error)
	GetClassesBySchoolIDs(schoolIDs []int64) ([]int64, error)
	GetStudentsByClassIDs(classIDs []int64) ([]StudentReportData, error)
	GetClassSchoolMap(classIDs []int64) (map[int64]int64, error)
	GetStudentClassMap(classIDs []int64) (map[int64]int64, error)
	GetAssessmentsByIDs(assessmentIDs []int64) ([]AssessmentInfo, error)
	GetAssessmentScores(studentIDs []int64, assessmentIDs []int64) (AssessmentScoreMap, error)
	GetSubjectsByIDs(subjectIDs []int64) (map[int64]string, error)
	GetGradesByIDs(gradeIDs []int64) (map[int64]string, error)
}

type dashboardAssessmentReportExcelRepository struct{}

func NewDashboardAssessmentReportExcelRepository() DashboardAssessmentReportExcelRepository {
	return &dashboardAssessmentReportExcelRepository{}
}

// GetSchoolsByIDs lấy danh sách trường theo IDs
func (r *dashboardAssessmentReportExcelRepository) GetSchoolsByIDs(schoolIDs []int64) ([]models.School, error) {
	var schools []models.School
	if len(schoolIDs) == 0 {
		return schools, nil
	}
	
	err := db.ReplicaDB.
		Where("id IN (?) AND deleted_at IS NULL", schoolIDs).
		Find(&schools).Error
	
	return schools, err
}

// GetClassesBySchoolIDs lấy danh sách class_ids từ school_ids
func (r *dashboardAssessmentReportExcelRepository) GetClassesBySchoolIDs(schoolIDs []int64) ([]int64, error) {
	var classIDs []int64
	
	if len(schoolIDs) == 0 {
		return classIDs, nil
	}
	
	err := db.ReplicaDB.Table("classes").
		Select("id").
		Where("school_id IN (?) AND deleted_at IS NULL", schoolIDs).
		Pluck("id", &classIDs).Error
	
	return classIDs, err
}

// GetStudentsByClassIDs lấy danh sách học sinh theo class_ids từ user_classes, kèm thông tin lớp, lớp gốc, khối
// Sắp xếp theo: lớp gốc, lớp đã chia, tên học sinh (duyệt ngược string, theo thứ tự tiếng Việt)
func (r *dashboardAssessmentReportExcelRepository) GetStudentsByClassIDs(classIDs []int64) ([]StudentReportData, error) {
	var students []StudentReportData
	
	if len(classIDs) == 0 {
		return students, nil
	}
	
	err := db.ReplicaDB.Table("users u").
		Select(`
			u.id AS student_id,
			u.name AS student_name,
			COALESCE(c.name, '') AS class_name,
			COALESCE(cm.name, '') AS class_main_name,
			COALESCE(g.name_vn, '') AS grade_name_vn,
			COALESCE(g.id, 0) AS grade_id
		`).
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3").
		Joins("JOIN user_classes uc ON uc.user_id = u.id AND uc.is_current = true").
		Joins("LEFT JOIN classes c ON c.id = uc.class_id AND c.deleted_at IS NULL").
		Joins("LEFT JOIN classes_main cm ON cm.id = c.class_main_id AND cm.deleted_at IS NULL").
		Joins("LEFT JOIN grades g ON g.id = c.grade_id AND g.deleted_at IS NULL").
		Where("uc.class_id IN (?) AND u.deleted_at IS NULL AND u.status = true", classIDs).
		Group("u.id, u.name, c.name, cm.name, g.name_vn, g.id").
		Order("cm.name ASC, c.name ASC, u.name ASC").
		Find(&students).Error
	
	if err != nil {
		return nil, err
	}
	
	// Sắp xếp lại theo tiếng Việt: lớp gốc, lớp đã chia, tên học sinh (duyệt ngược)
	c := collate.New(language.Vietnamese)
	sort.Slice(students, func(i, j int) bool {
		// So sánh lớp gốc
		if students[i].ClassMainName != students[j].ClassMainName {
			return c.CompareString(students[i].ClassMainName, students[j].ClassMainName) < 0
		}
		
		// So sánh lớp đã chia
		if students[i].ClassName != students[j].ClassName {
			return c.CompareString(students[i].ClassName, students[j].ClassName) < 0
		}
		
		// So sánh tên học sinh (duyệt ngược - lấy phần cuối của tên)
		lastNameI := extractLastNameForReport(students[i].StudentName)
		lastNameJ := extractLastNameForReport(students[j].StudentName)
		return c.CompareString(lastNameI, lastNameJ) < 0
	})
	
	return students, err
}

// extractLastNameForReport lấy tên từ name bằng cách duyệt ngược đến khi gặp ký tự cách
// Nếu không có ký tự cách thì lấy toàn bộ string
func extractLastNameForReport(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	
	// Duyệt ngược từ cuối string đến khi gặp ký tự cách
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == ' ' {
			// Tìm thấy ký tự cách, lấy phần sau ký tự cách
			return strings.TrimSpace(name[i+1:])
		}
	}
	
	// Không có ký tự cách, trả về toàn bộ string
	return name
}

// GetClassSchoolMap lấy mapping class_id -> school_id
func (r *dashboardAssessmentReportExcelRepository) GetClassSchoolMap(classIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64)
	
	if len(classIDs) == 0 {
		return result, nil
	}
	
	var data []struct {
		ID       int64 `gorm:"column:id"`
		SchoolID int64 `gorm:"column:school_id"`
	}
	
	err := db.ReplicaDB.Table("classes").
		Select("id, school_id").
		Where("id IN (?) AND deleted_at IS NULL", classIDs).
		Find(&data).Error
	
	if err != nil {
		return nil, err
	}
	
	for _, item := range data {
		result[item.ID] = item.SchoolID
	}
	
	return result, nil
}

// GetStudentClassMap lấy mapping student_id -> class_id từ user_classes
func (r *dashboardAssessmentReportExcelRepository) GetStudentClassMap(classIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64)
	
	if len(classIDs) == 0 {
		return result, nil
	}
	
	var data []struct {
		UserID  int64 `gorm:"column:user_id"`
		ClassID int64 `gorm:"column:class_id"`
	}
	
	err := db.ReplicaDB.Table("user_classes").
		Select("user_id, class_id").
		Where("class_id IN (?) AND is_current = true", classIDs).
		Find(&data).Error
	
	if err != nil {
		return nil, err
	}
	
	for _, item := range data {
		result[item.UserID] = item.ClassID
	}
	
	return result, nil
}

// GetAssessmentsByIDs lấy thông tin assessments theo IDs
func (r *dashboardAssessmentReportExcelRepository) GetAssessmentsByIDs(assessmentIDs []int64) ([]AssessmentInfo, error) {
	var assessments []AssessmentInfo
	
	if len(assessmentIDs) == 0 {
		return assessments, nil
	}
	
	err := db.ReplicaDB.Table("assessments a").
		Select(`
			a.id,
			a.subject_id,
			COALESCE(s.name, '') AS subject_name
		`).
		Joins("LEFT JOIN subjects s ON s.id = a.subject_id AND s.deleted_at IS NULL").
		Where("a.id IN (?) AND a.deleted_at IS NULL", assessmentIDs).
		Order("a.id ASC").
		Find(&assessments).Error
	
	return assessments, err
}

// GetAssessmentScores lấy điểm assessment theo student_ids và assessment_ids
func (r *dashboardAssessmentReportExcelRepository) GetAssessmentScores(studentIDs []int64, assessmentIDs []int64) (AssessmentScoreMap, error) {
	result := make(AssessmentScoreMap)
	
	if len(studentIDs) == 0 || len(assessmentIDs) == 0 {
		return result, nil
	}
	
	var scores []struct {
		StudentID    int64    `gorm:"column:student_id"`
		AssessmentID int64    `gorm:"column:assessment_id"`
		TotalScore   *float64 `gorm:"column:total_score"`
	}
	
	err := db.ReplicaDB.Table("assessment_scores").
		Select("student_id, assessment_id, total_score").
		Where("student_id IN (?) AND assessment_id IN (?) AND deleted_at IS NULL", studentIDs, assessmentIDs).
		Order("created_at DESC").
		Find(&scores).Error
	
	if err != nil {
		return nil, err
	}
	
	// Xử lý duplicate: chỉ lấy bản ghi mới nhất (đã ORDER BY created_at DESC)
	seen := make(map[int64]map[int64]bool)
	for _, score := range scores {
		if seen[score.StudentID] == nil {
			seen[score.StudentID] = make(map[int64]bool)
		}
		
		// Chỉ lấy bản ghi đầu tiên (mới nhất) cho mỗi cặp student_id + assessment_id
		if !seen[score.StudentID][score.AssessmentID] {
			seen[score.StudentID][score.AssessmentID] = true
			
			// Chỉ thêm vào map nếu total_score >= 0 (không âm)
			if score.TotalScore != nil && *score.TotalScore >= 0 {
				if result[score.StudentID] == nil {
					result[score.StudentID] = make(map[int64]*float64)
				}
				result[score.StudentID][score.AssessmentID] = score.TotalScore
			}
		}
	}
	
	return result, nil
}

// GetSubjectsByIDs lấy tên các subjects theo IDs
func (r *dashboardAssessmentReportExcelRepository) GetSubjectsByIDs(subjectIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	
	if len(subjectIDs) == 0 {
		return result, nil
	}
	
	var subjects []struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	
	err := db.ReplicaDB.Table("subjects").
		Select("id, name").
		Where("id IN (?) AND deleted_at IS NULL", subjectIDs).
		Find(&subjects).Error
	
	if err != nil {
		return nil, err
	}
	
	for _, subject := range subjects {
		result[subject.ID] = subject.Name
	}
	
	return result, nil
}

// GetGradesByIDs lấy tên các grades theo IDs
func (r *dashboardAssessmentReportExcelRepository) GetGradesByIDs(gradeIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	
	if len(gradeIDs) == 0 {
		return result, nil
	}
	
	var grades []struct {
		ID     int64  `gorm:"column:id"`
		NameVN string `gorm:"column:name_vn"`
	}
	
	err := db.ReplicaDB.Table("grades").
		Select("id, name_vn").
		Where("id IN (?) AND deleted_at IS NULL", gradeIDs).
		Find(&grades).Error
	
	if err != nil {
		return nil, err
	}
	
	for _, grade := range grades {
		result[grade.ID] = grade.NameVN
	}
	
	return result, nil
}

