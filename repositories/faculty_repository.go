package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"errors"

	"gorm.io/gorm"
)

type FacultyRepository interface {
	base.BaseRepositoryInterface[models.Faculty]
	CreateOrUpdateFaculty(faculty *models.Faculty) (int64, error)
}

type facultyRepository struct {
	*base.BaseRepository[models.Faculty]
}

func NewFacultyRepository() FacultyRepository {
	return &facultyRepository{
		BaseRepository: base.NewBaseRepository[models.Faculty](),
	}
}

func (r *facultyRepository) CreateOrUpdateFaculty(faculty *models.Faculty) (int64, error) {
	var existing models.Faculty
	var err error

	if faculty.ID > 0 {
		err = db.MasterDB.First(&existing, faculty.ID).Error
	}

	if err == nil {
		existing.Name = faculty.Name
		existing.SchoolId = faculty.SchoolId
		existing.Code = faculty.Code
		existing.Description = faculty.Description
		existing.NameHead = faculty.NameHead
		existing.AvatarInfo = faculty.AvatarInfo
		existing.Status = faculty.Status
		if saveErr := db.MasterDB.Save(&existing).Error; saveErr != nil {
			return 0, saveErr
		}
		return int64(existing.ID), nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.MasterDB.Create(faculty).Error; createErr != nil {
			return 0, createErr
		}
		return int64(faculty.ID), nil
	}

	return 0, err
}
