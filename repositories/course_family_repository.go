package repositories

import (
	"be-lms/database/db"
	"be-lms/dto"
	"errors"

	"gorm.io/gorm"
)

type CourseFamilyRepository interface {
	GetCourseBasic(courseID int64) (*CourseFamilyBasic, error)
	GetCourseMember(courseID int64) (*dto.CourseFamilyMember, error)
	GetChildren(parentCourseID int64) ([]dto.CourseFamilyMember, error)
}

type CourseFamilyBasic struct {
	ID             int64
	ParentCourseID int64
}

type courseFamilyRepository struct{}

func NewCourseFamilyRepository() CourseFamilyRepository {
	return &courseFamilyRepository{}
}

func (r *courseFamilyRepository) GetCourseBasic(courseID int64) (*CourseFamilyBasic, error) {
	var basic CourseFamilyBasic
	err := db.ReplicaDB.Table("courses").
		Select("id, parent_course_id").
		Where("id = ? AND deleted_at IS NULL", courseID).
		Scan(&basic).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	if basic.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &basic, nil
}

func (r *courseFamilyRepository) GetCourseMember(courseID int64) (*dto.CourseFamilyMember, error) {
	var member dto.CourseFamilyMember
	err := r.baseMemberQuery().
		Where("c.id = ?", courseID).
		Scan(&member).Error

	if err != nil {
		return nil, err
	}

	if member.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &member, nil
}

func (r *courseFamilyRepository) GetChildren(parentCourseID int64) ([]dto.CourseFamilyMember, error) {
	var children []dto.CourseFamilyMember

	err := r.baseMemberQuery().
		Where("c.parent_course_id = ?", parentCourseID).
		Scan(&children).Error

	return children, err
}

func (r *courseFamilyRepository) baseMemberQuery() *gorm.DB {
	return db.ReplicaDB.Table("courses AS c").
		Select(`
			c.id,
			c.name,
			c.object_title,
			COALESCE(c.program_id, 0) AS program_id,
			COALESCE(p.name, '') AS program_name,
			COALESCE(string_agg(DISTINCT s.name, ', '), '') AS school_name
		`).
		Joins("LEFT JOIN programs p ON p.id = c.program_id").
		Joins("LEFT JOIN course_schools cs ON cs.course_id = c.id").
		Joins("LEFT JOIN schools s ON s.id = cs.school_id").
		Where("c.deleted_at IS NULL").
		Group("c.id, p.name")
}

