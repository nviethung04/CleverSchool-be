package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// StudentWithCourse chứa thông tin học sinh kèm course
type StudentWithCourse struct {
	models.User
	IdentifierCode string // Mã định danh được tính toán
}

type AssessmentExportRepository interface {
	GetAssessmentWithCriteria(assessmentID int64) (*models.Assessment, error)
	GetStudentsByCourseID(courseID int64) ([]StudentWithCourse, error)
	GetCourseByID(courseID int64) (*models.Course, error)
	GetStudyReportCriteria(criteriaID int64) (*models.StudyReportCriteria, error)
	GetLessonIDByAssessmentAndCourse(assessmentID, courseID int64) (int64, error)
}

type assessmentExportRepository struct{}

func NewAssessmentExportRepository() AssessmentExportRepository {
	return &assessmentExportRepository{}
}

func (r *assessmentExportRepository) GetAssessmentWithCriteria(assessmentID int64) (*models.Assessment, error) {
	var entity models.Assessment
	
	err := db.ReplicaDB.
		Preload("AssessmentRefCriteria", "deleted_at IS NULL").
		Preload("AssessmentRefCriteria.AssessmentCriterion.AssessmentSubcriteria", "deleted_at IS NULL").
		Where("id = ? AND deleted_at IS NULL", assessmentID).
		First(&entity).Error
	
	if err != nil {
		return nil, err
	}
	
	return &entity, nil
}

func (r *assessmentExportRepository) GetStudentsByCourseID(courseID int64) ([]StudentWithCourse, error) {
	var users []models.User
	
	// Lấy danh sách học sinh từ course với role_id = 3 (Student)
	err := db.ReplicaDB.
		Table("users").
		Select("users.*").
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Joins("JOIN user_ref_roles ON user_ref_roles.user_id = users.id").
		Where("user_courses.course_id = ?", courseID).
		Where("user_ref_roles.role_id = ?", 3).
		Where("users.deleted_at IS NULL").
		Order("users.name ASC").
		Find(&users).Error
	
	if err != nil {
		return nil, err
	}
	
	// Chuyển đổi sang StudentWithCourse và tính toán mã định danh
	students := make([]StudentWithCourse, len(users))
	for i, user := range users {
		students[i] = StudentWithCourse{
			User:           user,
			IdentifierCode: calculateIdentifierCode(user.Code, user.Identifier),
		}
	}
	
	// Sắp xếp theo tên (phần cuối cùng của name) tăng dần với hỗ trợ tiếng Việt
	c := collate.New(language.Vietnamese)
	sort.Slice(students, func(i, j int) bool {
		lastNameI := extractLastName(students[i].Name)
		lastNameJ := extractLastName(students[j].Name)
		return c.CompareString(lastNameI, lastNameJ) < 0
	})
	
	return students, nil
}

// extractLastName lấy tên từ name bằng cách duyệt ngược đến khi gặp ký tự cách
// Nếu không có ký tự cách thì lấy toàn bộ string
func extractLastName(name string) string {
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

// calculateIdentifierCode tính toán mã định danh theo logic:
// - Nếu cả code và identifier đều rỗng → trống
// - Nếu có cả 2 → hiển thị code
// - Nếu chỉ có 1 trong 2 → hiển thị cái không trống
func calculateIdentifierCode(code, identifier string) string {
	codeEmpty := code == "" || code == "null"
	identifierEmpty := identifier == "" || identifier == "null"
	
	if codeEmpty && identifierEmpty {
		return ""
	}
	
	if !codeEmpty && !identifierEmpty {
		return code
	}
	
	if !codeEmpty {
		return code
	}
	
	return identifier
}

func (r *assessmentExportRepository) GetCourseByID(courseID int64) (*models.Course, error) {
	var course models.Course
	
	err := db.ReplicaDB.
		Where("id = ? AND deleted_at IS NULL", courseID).
		First(&course).Error
	
	if err != nil {
		return nil, err
	}
	
	return &course, nil
}

func (r *assessmentExportRepository) GetStudyReportCriteria(criteriaID int64) (*models.StudyReportCriteria, error) {
	var criteria models.StudyReportCriteria
	
	err := db.ReplicaDB.
		Preload("Subject").
		Preload("Skills").
		Preload("Skills.Types").
		Where("id = ?", criteriaID).
		First(&criteria).Error
	
	if err != nil {
		return nil, err
	}
	
	return &criteria, nil
}

func (r *assessmentExportRepository) GetLessonIDByAssessmentAndCourse(assessmentID, courseID int64) (int64, error) {
	var lessonID int64
	
	err := db.ReplicaDB.Table("assessment_ref_lessons").
		Select("lesson_id").
		Where("assessment_id = ? AND course_id = ?", assessmentID, courseID).
		Order("created_at ASC").
		Limit(1).
		Scan(&lessonID).Error
	
	if err != nil {
		return 0, err
	}
	
	return lessonID, nil
}

