package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"time"

	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(passwordReset *models.PasswordReset) error
	FindByToken(token string) (*models.PasswordReset, error)
	FindByEmail(email string) ([]models.PasswordReset, error)
	Update(passwordReset *models.PasswordReset) error
	DeleteExpired() error
	SetContext(ctx interface{})
}

type passwordResetRepository struct {
	db  *gorm.DB
	ctx interface{}
}

func NewPasswordResetRepository() PasswordResetRepository {
	return &passwordResetRepository{
		db: db.MasterDB,
	}
}

func (r *passwordResetRepository) SetContext(ctx interface{}) {
	r.ctx = ctx
}

func (r *passwordResetRepository) Create(passwordReset *models.PasswordReset) error {
	return r.db.Create(passwordReset).Error
}

func (r *passwordResetRepository) FindByToken(token string) (*models.PasswordReset, error) {
	var passwordReset models.PasswordReset
	err := r.db.Where("token = ? AND status = ? AND expires_at > ?", token, true, time.Now()).
		Preload("User").
		First(&passwordReset).Error
	if err != nil {
		return nil, err
	}
	return &passwordReset, nil
}

func (r *passwordResetRepository) FindByEmail(email string) ([]models.PasswordReset, error) {
	var passwordResets []models.PasswordReset
	err := r.db.Where("email = ? AND status = ? AND expires_at > ?", email, true, time.Now()).
		Find(&passwordResets).Error
	if err != nil {
		return nil, err
	}
	return passwordResets, nil
}

func (r *passwordResetRepository) Update(passwordReset *models.PasswordReset) error {
	return r.db.Save(passwordReset).Error
}

func (r *passwordResetRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.PasswordReset{}).Error
}

