package resources

import (
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"time"
)

type CertificateResource interface {
	FormatCertificate(certificate *models.Certificate) *prot.Certificate
	FormatCertificates(certificates []*models.Certificate) []*prot.Certificate
	FormatModelCertificate(certificate *prot.Certificate) *models.Certificate
}

type CertificateResourceImpl struct{}

func NewCertificateResource() CertificateResource {
	return &CertificateResourceImpl{}
}

func (r *CertificateResourceImpl) FormatCertificate(certificate *models.Certificate) *prot.Certificate {
	if certificate == nil {
		return nil
	}

	return &prot.Certificate{
		Id:              certificate.ID,
		UserId:          certificate.UserId,
		CourseId:        certificate.CourseId,
		Name:            certificate.Name,
		IssuedBy:        certificate.IssuedBy,
		IssuedDate:      certificate.IssuedDate.Format("2006-01-02"),
		ExpiryDate:      certificate.ExpiryDate.Format("2006-01-02"),
		CertificateCode: certificate.CertificateCode,
		Description:     certificate.Description,
		Status:          certificate.Status,
		Grade:           certificate.Grade,
		Rating:          certificate.Rating,
		FileUrl:         utils.StaticURL(certificate.FileInfo.Path, models.Storage),
		CreatedAt:       certificate.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       certificate.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *CertificateResourceImpl) FormatCertificates(certificates []*models.Certificate) []*prot.Certificate {
	result := make([]*prot.Certificate, 0, len(certificates))
	for _, u := range certificates {
		if formatted := r.FormatCertificate(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *CertificateResourceImpl) FormatModelCertificate(certificate *prot.Certificate) *models.Certificate {
	if certificate == nil {
		return nil
	}

	var issuedDate, expiryDate time.Time
	var err error

	if certificate.IssuedDate != "" {
		issuedDate, err = time.Parse("2006-01-02", certificate.IssuedDate)
		if err != nil {
			issuedDate = time.Time{}
		}
	}

	if certificate.ExpiryDate != "" {
		expiryDate, err = time.Parse("2006-01-02", certificate.ExpiryDate)
		if err != nil {
			expiryDate = time.Time{}
		}
	}

	fileUrl := utils.StripDomain(certificate.FileUrl, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

	return &models.Certificate{
		ID:              certificate.Id,
		UserId:          certificate.UserId,
		CourseId:        certificate.CourseId,
		Name:            certificate.Name,
		IssuedBy:        certificate.IssuedBy,
		IssuedDate:      issuedDate,
		ExpiryDate:      expiryDate,
		CertificateCode: certificate.CertificateCode,
		Description:     certificate.Description,
		Status:          certificate.Status,
		FileInfo:        fileInfo,
		Grade:           certificate.Grade,
		Rating:          certificate.Rating,
	}
}

