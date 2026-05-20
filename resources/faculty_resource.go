package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
)

type FacultyResource interface {
	FormatFaculty(faculty *models.Faculty) *prot.Faculty
	FormatFaculties(faculties []*models.Faculty) []*prot.Faculty
	FormatModelFaculty(faculty *prot.FacultyRequest) *models.Faculty
}

type FacultyResourceImpl struct{}

func NewFacultyResource() FacultyResource {
	return &FacultyResourceImpl{}
}

func (r *FacultyResourceImpl) FormatFaculty(faculty *models.Faculty) *prot.Faculty {
	if faculty == nil {
		return nil
	}

	var school *prot.School
	schoolResource := NewSchoolResource()
	if faculty.School != nil {
		school = schoolResource.FormatSchool(faculty.School)
	}

	return &prot.Faculty{
		Id:        int64(faculty.ID),
		SchoolId:  faculty.SchoolId,
		School:    school,
		Name:      faculty.Name,
		Code:      faculty.Code,
		NameHead:  faculty.NameHead,
		Avatar:    utils.StaticURL(faculty.AvatarInfo.Path, models.Storage),
		Status:    faculty.Status,
		CreatedAt: faculty.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: faculty.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *FacultyResourceImpl) FormatFaculties(faculties []*models.Faculty) []*prot.Faculty {
	result := make([]*prot.Faculty, 0, len(faculties))
	for _, t := range faculties {
		if formatted := r.FormatFaculty(t); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *FacultyResourceImpl) FormatModelFaculty(faculty *prot.FacultyRequest) *models.Faculty {
	if faculty == nil {
		return nil
	}

	avatar := utils.StripDomain(faculty.Avatar, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	avatarInfo := mediaRepo.GetMediaInfo(avatar, models.Storage)

	return &models.Faculty{
		ID:          int64(faculty.Id),
		SchoolId:    faculty.SchoolId,
		Name:        faculty.Name,
		Code:        faculty.Code,
		NameHead:    faculty.NameHead,
		AvatarInfo:  avatarInfo,
		Description: faculty.Description,
		Status:      faculty.Status,
	}
}

