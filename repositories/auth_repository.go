package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/i18n"
	"be-Clever School/models"

	"gorm.io/gorm"

	"errors"
	"time"
)

type AuthRepository interface {
	CreateUser(user models.User, roleId int64) (*models.User, error, string)
	GetByUsername(username string) (*models.User, error)
	GetByKey(username string) (*models.User, error)
	GetByID(userID int) (*models.User, error, string)
	SaveToken(userID int, token string, expiresTime time.Time) error
	UpdatePassword(userID int64, hashedPassword string) error
	CheckExistingUser(username, email string) error

	GetUsersByKey(value string) ([]models.User, error)
	GetRoleById(id int) (*models.Role, error)
}

type authRepository struct{}

func NewAuthRepository() AuthRepository {
	return &authRepository{}
}

func (r *authRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User

	err := db.ReplicaDB.
		Where("username = ?", username).
		Preload("Roles").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(i18n.Localize("messages.user_not_found"))
		}
		return nil, err
	}

	return &user, nil
}

func (r *authRepository) GetByID(userID int) (*models.User, error, string) {
	var user models.User
	err := db.ReplicaDB.Where("id = ?", userID).
		Preload("Roles").
		Preload("School").
		Preload("UserAddress").
		Preload("Certificates").
		Preload("Degrees").
		Preload("Departments").
		Preload("Positions").
		Preload("UserClasses.Class").
		Preload("UserClasses.Class.School").
		Preload("UserCourses.Course").
		Preload("Subjects").
		First(&user).Error
	if err != nil {
		// Nếu không tìm thấy user
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(i18n.Localize("messages.user_not_found")), "messages.user_not_found"
		}
		// Nếu có lỗi khác
		return nil, err, "messages.error_get_by_id"
	}
	return &user, nil, ""
}

func (r *authRepository) SaveToken(userID int, token string, expiresTime time.Time) error {
	// Cập nhật token và thời gian hết hạn vào MasterDB
	err := db.MasterDB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"token":         token,
		"expires_time":  expiresTime,
		"last_login_at": time.Now(),
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *authRepository) CreateUser(user models.User, roleId int64) (*models.User, error, string) {
	if err := db.MasterDB.Create(&user).Error; err != nil {
		return nil, err, "messages.error_create_user"
	}

	userRefRole := models.UserRefRole{
		UserId: user.ID,
		RoleId: roleId,
	}

	if err := db.MasterDB.Create(&userRefRole).Error; err != nil {
		return nil, err, "messages.error_create_user"
	}

	if user.MemberType == models.MemberTypeExternal {
		courseIds := config.LoadConfig().PublicCourseIds

		for _, courseId := range courseIds {
			userCourse := models.UserCourse{
				UserId:   user.ID,
				CourseId: int64(courseId),
			}
			db.MasterDB.Create(&userCourse)
		}
	}

	return &user, nil, ""
}

func (r *authRepository) UpdatePassword(userID int64, hashedPassword string) error {
	return db.MasterDB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func (r *authRepository) GetByKey(value string) (*models.User, error) {
	var user models.User

	err := db.ReplicaDB.
		Where("email = ? OR username = ? OR code = ? OR identifier = ?", value, value, value, value).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *authRepository) GetUsersByKey(value string) ([]models.User, error) {
	var users []models.User

	err := db.ReplicaDB.Preload("Roles").
		Where(`
			(users.email = ? OR users.username = ? OR users.code = ? OR users.identifier = ?)
		`, value, value, value, value).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *authRepository) CheckExistingUser(username, email string) error {
	var existingUser models.User

	condition := "username = ?"
	args := []interface{}{username}

	if email != "" {
		condition += " OR email = ?"
		args = append(args, email)
	}

	err := db.MasterDB.
		Unscoped().
		Where(condition, args...).
		First(&existingUser).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return errors.New(i18n.Localize("messages.account_already_exists"))
}

func (r *authRepository) GetRoleById(id int) (*models.Role, error) {
	var role models.Role

	err := db.ReplicaDB.
		Where("id = ?", id).
		First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(i18n.Localize("messages.role_not_found"))
		}
		return nil, err
	}

	return &role, nil
}
