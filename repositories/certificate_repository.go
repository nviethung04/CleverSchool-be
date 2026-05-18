package repositories

import (
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type CertificateRepository interface {
	base.BaseRepositoryInterface[models.Certificate]
}

type certificateRepository struct {
	*base.BaseRepository[models.Certificate]
}

func NewCertificateRepository() CertificateRepository {
	return &certificateRepository{
		BaseRepository: base.NewBaseRepository[models.Certificate](),
	}
}
