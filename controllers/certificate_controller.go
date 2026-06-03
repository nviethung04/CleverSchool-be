package controllers

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
)

type CertificateController struct {
	*GenericController[models.Certificate, prot.Certificate, *prot.Certificate]
}

func NewCertificateController(service services.CertificateService) *CertificateController {
	certificateResource := resources.NewCertificateResource()
	certificateResourceAdapter := NewCertificateResourceAdapter(certificateResource)

	genericController := NewGenericController(
		service,
		certificateResourceAdapter,
		func() *prot.Certificate {
			return &prot.Certificate{}
		},
		func(certificates []*prot.Certificate, totalCount uint64) interface{} {
			return &prot.CertificateListResponse{
				Certificates:       certificates,
				TotalCount: int64(totalCount),
			}
		},
	)

	return &CertificateController{
		GenericController: genericController,
	}
}
