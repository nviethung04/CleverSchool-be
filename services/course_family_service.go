package services

import (
	"be-cleverschool/dto"
	"be-cleverschool/repositories"
	"errors"

	"gorm.io/gorm"
)

type CourseFamilyService interface {
	GetCourseFamily(courseID int64) (*dto.CourseFamily, error)
}

type courseFamilyService struct {
	repo repositories.CourseFamilyRepository
}

func NewCourseFamilyService(repo repositories.CourseFamilyRepository) CourseFamilyService {
	return &courseFamilyService{repo: repo}
}

func (s *courseFamilyService) GetCourseFamily(courseID int64) (*dto.CourseFamily, error) {
	if courseID == 0 {
		return nil, errors.New("course_id is required")
	}

	course, err := s.repo.GetCourseBasic(courseID)
	if err != nil {
		return nil, err
	}

	parentID := course.ParentCourseID
	if parentID == 0 {
		parentID = course.ID
	}

	parent, err := s.repo.GetCourseMember(parentID)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(parent.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &dto.CourseFamily{
		Parent:   parent,
		Children: children,
	}, nil
}


