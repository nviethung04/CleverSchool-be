package services

import (
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	_ "time"
)

type ProvinceService interface {
	GetAll() ([]*prot.Province, error)
	GetByCode(code string) (*prot.Province, error)
	Wards(code string) ([]*prot.Ward, error)
}

type provinceService struct {
	repo           repositories.ProvinceRepository
	localeProvider i18n.LocaleProvider
}

func NewProvinceService(repo repositories.ProvinceRepository) ProvinceService {
	return &provinceService{repo: repo, localeProvider: i18n.GlobalLocaleProvider{}}
}

func NewProvinceServiceWithLocale(repo repositories.ProvinceRepository, provider i18n.LocaleProvider) ProvinceService {
	if provider == nil {
		provider = i18n.GlobalLocaleProvider{}
	}
	return &provinceService{repo: repo, localeProvider: provider}
}

func (s *provinceService) GetAll() ([]*prot.Province, error) {
	// Không dùng paging ở đây, nếu cần thì thêm param requests.GetProvincesRequest vào
	parts, err := s.repo.GetAll(&requests.GetProvincesRequest{})
	if err != nil {
		return nil, err
	}

	var res []*prot.Province
	for _, p := range parts {
		res = append(res, s.modelToProtoProvince(&p))
	}
	return res, nil
}

func (s *provinceService) GetByCode(code string) (*prot.Province, error) {
	p, err := s.repo.GetByCode(code)
	if err != nil {
		return nil, err
	}
	return s.modelToProtoProvince(p), nil
}

func (s *provinceService) Wards(code string) ([]*prot.Ward, error) {
	wards, err := s.repo.GetWards(code)
	if err != nil {
		return nil, err
	}

	var protoWards []*prot.Ward
	for _, w := range wards {
		protoWards = append(protoWards, s.modelToProtoWard(&w))
	}

	return protoWards, nil
}

func (s *provinceService) modelToProtoProvince(p *models.Province) *prot.Province {
	var tr models.ProvinceTranslation
	_ = i18n.FillTranslationWithProvider(p, &tr, s.localeProvider)
	return &prot.Province{
		Code:     p.Code,
		Name:     tr.Name,
		FullName: tr.FullName,
	}
}

func (s *provinceService) modelToProtoWard(w *models.Ward) *prot.Ward {
	var tr models.WardTranslation
	_ = i18n.FillTranslationWithProvider(w, &tr, s.localeProvider)
	return &prot.Ward{
		Code:         w.Code,
		ProvinceCode: w.ProvinceCode,
		Name:         tr.Name,
		FullName:     tr.FullName,
	}
}
