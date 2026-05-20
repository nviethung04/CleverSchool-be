package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SchoolRepository interface {
	base.BaseRepositoryInterface[models.School]
	base.BeforeQueryHook
	GetAllSchool(req *requests.GetSchoolRequest, ctx *gin.Context) ([]models.School, int64, error)
	CreateOrUpdateSchool(user *models.School) (int64, error)
	GetAllIds() ([]int64, error)
	GetSchoolStudents(schoolId int64) ([]models.User, error)
}

type schoolRepository struct {
	*base.BaseRepository[models.School]
}

func NewSchoolRepository() SchoolRepository {
	repo := &schoolRepository{
		BaseRepository: base.NewBaseRepository[models.School](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

type SchoolWithLocation struct {
	models.School

	WardName     string `json:"ward_name"`
	ProvinceName string `json:"province_name"`
	ProvinceCode string `json:"province_code"`
}

func (r *schoolRepository) GetAllSchool(req *requests.GetSchoolRequest, ctx *gin.Context) ([]models.School, int64, error) {
	var results []SchoolWithLocation
	var total int64

	// Join Ward table để lấy tên ward và province
	query := db.ReplicaDB.Model(&models.School{}).
		Select("schools.*, wards.full_name as ward_name, provinces.full_name as province_name, provinces.code as province_code").
		Joins("left join wards on wards.code = schools.ward_code").
		Joins("left join provinces on provinces.code = wards.province_code").
		Where("schools.deleted_at IS NULL")

	query = r.BeforeQuery(query, ctx)

	if req.Name != "" {
		query = query.Where("unaccent(schools.name) ILIKE unaccent(?)", "%"+req.Name+"%")
	}

	if req.Type != "" {
		query = query.Where("unaccent(schools.type) ILIKE unaccent(?)", "%"+req.Type+"%")
	}

	if req.WardCode != "" {
		query = query.Where("schools.ward_code = ?", req.WardCode)
	}

	if req.ProvinceCode != "" {
		query = query.Where("wards.province_code = ?", req.ProvinceCode)
	}

	if req.Status != "" {
		query = query.Where("schools.status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err := query.Find(&results).Error

	if err != nil {
		return nil, 0, err
	}

	schools := make([]models.School, len(results))
	for i, r := range results {
		schools[i] = r.School
		schools[i].WardName = r.WardName
		schools[i].ProvinceName = r.ProvinceName
		schools[i].ProvinceCode = r.ProvinceCode
	}

	return schools, total, nil
}

func (r *schoolRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	schoolId := r.GetAdminSchoolId(ctx)

	if schoolId > 0 {
		query = query.Where("id = ?", schoolId)
	}

	return query
}

func (r *schoolRepository) CreateOrUpdateSchool(school *models.School) (int64, error) {
	var existing models.School
	var err error

	if school.ID > 0 {
		err = db.MasterDB.First(&existing, school.ID).Error
	}

	if err == nil {
		existing.Name = school.Name
		existing.ShortName = school.ShortName
		existing.Type = school.Type
		existing.AddressVN = school.AddressVN
		existing.AddressEN = school.AddressEN
		existing.ContactName = school.ContactName
		existing.ContactPhone = school.ContactPhone
		existing.WardCode = school.WardCode
		existing.LogoInfo = school.LogoInfo
		if saveErr := db.MasterDB.Save(&existing).Error; saveErr != nil {
			return 0, saveErr
		}
		return int64(existing.ID), nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.MasterDB.Create(school).Error; createErr != nil {
			return 0, createErr
		}
		return int64(school.ID), nil
	}

	return 0, err
}

func (r *schoolRepository) GetAllIds() ([]int64, error) {
	var ids []int64
	if err := db.MasterDB.Model(&models.School{}).Select("id").Find(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// GetSchoolStudents gets all students from a specific school
func (r *schoolRepository) GetSchoolStudents(schoolId int64) ([]models.User, error) {
	var students []models.User

	// Get students who belong to the school and have student role
	err := db.ReplicaDB.
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("users.school_id = ? AND roles.name = 'student' AND users.deleted_at IS NULL", schoolId).
		Preload("School").
		Find(&students).Error

	return students, err
}

