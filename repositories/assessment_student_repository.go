package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
)

// AssessmentStudentRepository phục vụ API cũ /teacher/assessment/students
type AssessmentStudentRepository interface {
	GetStudentsByCourse(courseID int64, studentID *int64, limit, offset int) ([]dto.AssessmentStudentItem, int64, error)
	GetCourseInfo(courseID int64) (string, string, string, error) // name, object_title, teacher_names
}

type assessmentStudentRepository struct{}

func NewAssessmentStudentRepository() AssessmentStudentRepository {
	return &assessmentStudentRepository{}
}

func (r *assessmentStudentRepository) GetStudentsByCourse(courseID int64, studentID *int64, limit, offset int) ([]dto.AssessmentStudentItem, int64, error) {
	var students []dto.AssessmentStudentItem
	var total int64

	// Count total distinct students in course
	countQuery := db.ReplicaDB.Table("user_courses uc").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", courseID)

	if studentID != nil && *studentID > 0 {
		countQuery = countQuery.Where("uc.user_id = ?", *studentID)
	}

	if err := countQuery.Distinct("uc.user_id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch student list
	query := db.ReplicaDB.Table("user_courses uc").
		Select("uc.user_id AS student_id, u.name AS student_name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND uc.course_id = ?", courseID)

	if studentID != nil && *studentID > 0 {
		query = query.Where("uc.user_id = ?", *studentID)
	}

	query = query.Group("uc.user_id, u.name").
		Order("u.name ASC")

	if limit > 0 && offset >= 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&students).Error; err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *assessmentStudentRepository) GetCourseInfo(courseID int64) (string, string, string, error) {
	// Lấy name và object_title từ bảng courses
	var course struct {
		Name        string
		ObjectTitle string
	}
	err := db.ReplicaDB.Table("courses").
		Select("name, object_title").
		Where("id = ?", courseID).
		First(&course).Error
	if err != nil {
		return "", "", "", err
	}

	// Lấy danh sách tên giáo viên từ user_courses join users join user_ref_roles
	var teachers []struct {
		Name string
	}
	err = db.ReplicaDB.Table("user_courses uc").
		Select("u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("uc.course_id = ? AND urr.role_id = 2 AND u.deleted_at IS NULL", courseID).
		Find(&teachers).Error
	if err != nil {
		return "", "", "", err
	}

	// Ghép tên giáo viên thành chuỗi phân cách bởi dấu phẩy
	teacherNames := ""
	for i, teacher := range teachers {
		if i > 0 {
			teacherNames += ", "
		}
		teacherNames += teacher.Name
	}

	return course.Name, course.ObjectTitle, teacherNames, nil
}

func (r *assessmentStudentRepository) GetStudentsByClass(classID int64, studentID *int64, limit, offset int) ([]dto.AssessmentStudentItem, int64, error) {
	// 1. Lấy danh sách ID học sinh thuộc lớp
	var studentIDs []int64
	studentIDQuery := db.ReplicaDB.Table("user_classes").
		Select("user_id").
		Where("class_id = ?", classID)

	if studentID != nil && *studentID > 0 {
		studentIDQuery = studentIDQuery.Where("user_id = ?", *studentID)
	}

	if err := studentIDQuery.Pluck("user_id", &studentIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(studentIDs) == 0 {
		return []dto.AssessmentStudentItem{}, 0, nil
	}

	// 2. Lấy thông tin học sinh từ danh sách ID với WHERE IN
	var students []dto.AssessmentStudentItem
	var total int64

	// Count total
	countQuery := db.ReplicaDB.Table("users u").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND u.id IN (?)", studentIDs)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch student list
	query := db.ReplicaDB.Table("users u").
		Select("u.id AS student_id, u.name AS student_name").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND u.id IN (?)", studentIDs).
		Order("u.name ASC")

	if limit > 0 && offset >= 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&students).Error; err != nil {
		return nil, 0, err
	}

	return students, total, nil
}


