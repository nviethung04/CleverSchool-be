package repositories

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
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

