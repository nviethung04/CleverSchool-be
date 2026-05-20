package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/requests"
)

type WardRepository interface {
	GetAll(req *requests.GetWardsRequest) ([]models.Ward, error)
	GetByCode(code string) (*models.Ward, error)
}

type wardRepository struct{}

func NewWardRepository() WardRepository {
	return &wardRepository{}
}

func (r *wardRepository) GetAll(req *requests.GetWardsRequest) ([]models.Ward, error) {
	var wards []models.Ward

	query := db.ReplicaDB.Model(&models.Ward{})

	if req.ProvinceCode != "" {
		query = query.Where("province_code = ?", req.ProvinceCode)
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err := query.Find(&wards).Error
	if err != nil {
		return nil, err
	}
	return wards, nil
}

func (r *wardRepository) GetByCode(code string) (*models.Ward, error) {
	var ward models.Ward
	err := db.ReplicaDB.Where("code = ?", code).First(&ward).Error
	if err != nil {
		return nil, err
	}
	return &ward, nil
}

