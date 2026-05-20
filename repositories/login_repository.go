package repositories

//
//import (
//	"be-Clever School/database/db"
//	"be-Clever School/models"
//	"errors"
//	"gorm.io/gorm"
//	"time"
//)
//
//type LoginRepository interface {
//	GetByUsername(username string) (*models.User, error)
//	GetByID(userID int) (*models.User, error)
//	SaveToken(userID int, token string, expiresTime time.Time) error
//}
//
//type loginRepository struct{}
//
//func NewLoginRepository() LoginRepository {
//	return &loginRepository{}
//}
//
//func (r *loginRepository) GetByUsername(username string) (*models.User, error) {
//	var user models.User
//	err := db.ReplicaDB.Where("username = ?", username).First(&user).Error
//	if err != nil {
//		// Nếu không tìm thấy user
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return nil, errors.New("user not found")
//		}
//		// Nếu có lỗi khác
//		return nil, err
//	}
//	return &user, nil
//}
//
//func (r *loginRepository) GetByID(userID int) (*models.User, error) {
//	var user models.User
//	err := db.ReplicaDB.Where("id = ?", userID).First(&user).Error
//	if err != nil {
//		// Nếu không tìm thấy user
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return nil, errors.New("user not found")
//		}
//		// Nếu có lỗi khác
//		return nil, err
//	}
//	return &user, nil
//}
//
//func (r *loginRepository) SaveToken(userID int, token string, expiresTime time.Time) error {
//	// Cập nhật token và thời gian hết hạn vào MasterDB
//	err := db.MasterDB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
//		"token":        token,
//		"expires_time": expiresTime,
//	}).Error
//
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
