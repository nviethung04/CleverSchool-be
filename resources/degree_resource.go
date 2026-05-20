package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"time"
)

type DegreeResource interface {
	FormatDegree(degree *models.Degree) *prot.Degree
	FormatDegrees(degrees []*models.Degree) []*prot.Degree
	FormatModelDegree(degree *prot.Degree) *models.Degree
}

type DegreeResourceImpl struct{}

func NewDegreeResource() DegreeResource {
	return &DegreeResourceImpl{}
}

func (r *DegreeResourceImpl) FormatDegree(degree *models.Degree) *prot.Degree {
	if degree == nil {
		return nil
	}

	return &prot.Degree{
		Id:             degree.ID,
		UserId:         degree.UserId,
		Name:           degree.Name,
		Major:          degree.Major,
		Institution:    degree.Institution,
		GraduationYear: int32(degree.GraduationYear),
		DegreeLevel:    degree.DegreeLevel,
		DegreeCode:     degree.DegreeCode,
		ReceivedDate:   degree.ReceivedDate.Format("2006-01-02"),
		Note:           degree.Note,
		FileUrl:        utils.StaticURL(degree.FileInfo.Path, models.Storage),
		Status:         degree.Status,
		CreatedAt:      degree.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      degree.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *DegreeResourceImpl) FormatDegrees(degrees []*models.Degree) []*prot.Degree {
	result := make([]*prot.Degree, 0, len(degrees))
	for _, u := range degrees {
		if formatted := r.FormatDegree(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *DegreeResourceImpl) FormatModelDegree(degree *prot.Degree) *models.Degree {
	if degree == nil {
		return nil
	}

	var receivedDate time.Time
	var err error

	if degree.ReceivedDate != "" {
		receivedDate, err = time.Parse("2006-01-02", degree.ReceivedDate)
		if err != nil {
			receivedDate = time.Time{}
		}
	}

	fileUrl := utils.StripDomain(degree.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	return &models.Degree{
		ID:             degree.Id,
		UserId:         degree.UserId,
		Name:           degree.Name,
		Major:          degree.Major,
		Institution:    degree.Institution,
		GraduationYear: int(degree.GraduationYear),
		DegreeLevel:    degree.DegreeLevel,
		DegreeCode:     degree.DegreeCode,
		ReceivedDate:   receivedDate,
		Note:           degree.Note,
		Status:         degree.Status,
		FileInfo:       fileInfo,
	}
}

