package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/resources"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
)

type CertificateService interface {
	GetAll(c *gin.Context) ([]models.Certificate, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Certificate, error)
	Create(c *gin.Context, req *prot.Certificate) (*models.Certificate, error)
	Update(c *gin.Context, req *prot.Certificate) (*models.Certificate, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Certificate, error)
}

type certificateService struct {
	repo repositories.CertificateRepository
}

func NewCertificateService(repo repositories.CertificateRepository) CertificateService {
	return &certificateService{repo}
}

func (s *certificateService) GetAll(c *gin.Context) ([]models.Certificate, int64, error) {
	allowedFilters := []string{"user_id", "course_id", "status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	certificatees, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return certificatees, rows, nil
}

func (s *certificateService) GetByID(c *gin.Context, id int) (*prot.Certificate, error) {
	certificate, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	certificateResource := resources.NewCertificateResource()
	formattedCertificate := certificateResource.FormatCertificate(certificate)

	return formattedCertificate, nil
}

func (s *certificateService) Create(c *gin.Context, req *prot.Certificate) (*models.Certificate, error) {
	certificateResource := resources.NewCertificateResource()
	certificate := certificateResource.FormatModelCertificate(req)

	s.repo.SetContext(c)

	err := s.repo.Create(certificate)
	if err != nil {
		return nil, err
	}

	id := int(certificate.ID)
	newCertificate, _ := s.repo.FindNewByID(id)

	return newCertificate, nil
}

func (s *certificateService) Update(c *gin.Context, req *prot.Certificate) (*models.Certificate, error) {
	certificateResource := resources.NewCertificateResource()
	certificate := certificateResource.FormatModelCertificate(req)

	s.repo.SetContext(c)

	err := s.repo.Update(certificate)
	if err != nil {
		return nil, err
	}

	id := int(certificate.ID)
	updateCertificate, _ := s.repo.FindNewByID(id)

	return updateCertificate, nil
}

func (s *certificateService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *certificateService) Restore(c *gin.Context, id int) (*models.Certificate, error) {
	s.repo.SetContext(c)
	certificate, err := s.repo.Restore(id)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}
