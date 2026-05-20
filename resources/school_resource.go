package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
)

type SchoolResource interface {
	FormatSchool(school *models.School) *prot.School
	FormatSchools(schools []*models.School) []*prot.School
	FormatModelSchool(school *prot.SchoolRequest) *models.School
}

type SchoolResourceImpl struct{}

func NewSchoolResource() SchoolResource {
	return &SchoolResourceImpl{}
}

func (r *SchoolResourceImpl) FormatSchool(school *models.School) *prot.School {
	if school == nil {
		return nil
	}

	var parentId int64
	if school.ParentSchoolID != nil {
		parentId = *school.ParentSchoolID
	}

	wardName := school.WardName
	provinceCode := school.ProvinceCode
	provinceName := school.ProvinceName

	if wardName == "" {
		wardName = school.Ward.FullName
	}

	if provinceCode == "" && school.Ward.Province.Code != "" {
		provinceCode = school.Ward.Province.Code
		provinceName = school.Ward.Province.FullName
	}

	return &prot.School{
		Id:             school.ID,
		Name:           school.Name,
		ShortName:      school.ShortName,
		Type:           school.Type,
		WardCode:       school.WardCode,
		WardName:       wardName,
		ProvinceCode:   provinceCode,
		ProvinceName:   provinceName,
		AddressVn:      school.AddressVN,
		AddressEn:      school.AddressEN,
		ParentSchoolId: parentId,
		ContactName:    school.ContactName,
		ContactPhone:   school.ContactPhone,
		ContactMail:    school.ContactMail,
		Status:         school.Status,
		Logo:           utils.StaticURL(school.LogoInfo.Path, models.Storage),
		StudentCount:   school.StudentCount,
		ClassCount:     school.ClassCount,
		CreatedAt:      school.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      school.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *SchoolResourceImpl) FormatSchools(schools []*models.School) []*prot.School {
	result := make([]*prot.School, 0, len(schools))
	for _, u := range schools {
		if u == nil {
			continue
		}
		if formatted := r.FormatSchool(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *SchoolResourceImpl) FormatModelSchool(school *prot.SchoolRequest) *models.School {
	if school == nil {
		return nil
	}

	var parentID *int64
	if school.ParentSchoolId != 0 {
		parentID = &school.ParentSchoolId
	}

	logoUrl := utils.StripDomain(school.Logo, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	logoInfo := mediaRepo.GetMediaInfo(logoUrl, models.Storage)

	return &models.School{
		ID:             school.Id,
		Name:           school.Name,
		ShortName:      school.ShortName,
		Type:           school.Type,
		WardCode:       school.WardCode,
		AddressVN:      school.AddressVn,
		AddressEN:      school.AddressEn,
		ParentSchoolID: parentID,
		ContactName:    school.ContactName,
		ContactPhone:   school.ContactPhone,
		ContactMail:    school.ContactMail,
		Status:         school.Status,
		LogoInfo:       logoInfo,
	}
}

