package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/requests"
)

type ProvinceRepository interface {
	GetAll(req *requests.GetProvincesRequest) ([]models.Province, error)
	GetByCode(code string) (*models.Province, error)
	GetWards(code string) ([]models.Ward, error)
}

type provinceRepository struct{}

func NewProvinceRepository() ProvinceRepository {
	return &provinceRepository{}
}

func (r *provinceRepository) GetAll(req *requests.GetProvincesRequest) ([]models.Province, error) {
	var provinces []models.Province

	query := db.ReplicaDB.Model(&models.Province{})

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err := query.Find(&provinces).Error
	if err != nil {
		return nil, err
	}
	return provinces, nil
}

func (r *provinceRepository) GetByCode(code string) (*models.Province, error) {
	var province models.Province
	err := db.ReplicaDB.Where("code = ?", code).First(&province).Error
	if err != nil {
		return nil, err
	}
	return &province, nil
}

func (r *provinceRepository) GetWards(code string) ([]models.Ward, error) {
	var wards []models.Ward
	err := db.ReplicaDB.Where("province_code = ?", code).Find(&wards).Error
	if err != nil {
		return nil, err
	}
	return wards, nil
}

