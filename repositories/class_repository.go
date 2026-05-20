package repositories

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ClassRepository interface {
	base.BaseRepositoryInterface[models.Class]
	GetUsers(classId int64, roleId int64) ([]models.User, []int64, error)
	ReplaceUserClass(classId int64, userIds []int64, roleId int64) error
	AddUserClass(classId int64, userIds []int64, roleId int64) error

	GetUserClass(userId int64) error
	UpdateCurrentStudentToClass(classId int64) (err error)
	CreateOrUpdateClass(user *models.Class) (int64, error)
	UpdateSchoolStats(schoolId int64) error
}

type classRepository struct {
	*base.BaseRepository[models.Class]
}

func NewClassRepository() ClassRepository {
	repo := &classRepository{
		BaseRepository: base.NewBaseRepository[models.Class](),
	}

	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *classRepository) GetUsers(classId int64, roleId int64) ([]models.User, []int64, error) {
	var users []models.User

	query := db.MasterDB.
		Joins("JOIN user_classes ON user_classes.user_id = users.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("user_classes.class_id = ?", classId).
		Preload("UserClasses", "class_id = ?", classId)

	if roleId > 0 {
		query = query.Where("urr.role_id = ?", roleId)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	// Classes don't have is_failed field, return empty slice
	return users, []int64{}, nil
}

func (r *classRepository) ReplaceUserClass(classId int64, userIds []int64, roleId int64) error {
	if len(userIds) == 0 {
		userIds = []int64{0}
	}

	tx := db.MasterDB

	var userIdsToDelete []int64

	if roleId != 0 {
		err := tx.Table("user_classes").
			Joins("JOIN users ON users.id = user_classes.user_id").
			Where("user_classes.class_id = ?", classId).
			Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("urr.role_id = ?", roleId).
			Where("user_classes.user_id NOT IN ?", userIds).
			Pluck("user_classes.user_id", &userIdsToDelete).Error
		if err != nil {
			return err
		}
	} else {
		err := tx.Model(&models.UserClass{}).
			Where("class_id = ? AND user_id NOT IN ?", classId, userIds).
			Pluck("user_id", &userIdsToDelete).Error
		if err != nil {
			return err
		}
	}

	// Delete old user
	if len(userIdsToDelete) > 0 {
		if err := tx.
			Where("class_id = ? AND user_id IN ?", classId, userIdsToDelete).
			Delete(&models.UserClass{}).Error; err != nil {
			return err
		}
	}

	// Add user
	for _, uid := range userIds {
		var count int64
		query := tx.Model(&models.UserClass{}).
			Where("class_id = ? AND user_id = ?", classId, uid)

		if roleId != 0 {
			query = query.Joins("JOIN users ON user_classes.user_id = users.id").
				Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
				Where("urr.role_id = ?", roleId)
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}

		if count != 1 {
			if roleId != 0 {
				var userRoleCount int64
				err := tx.Model(&models.User{}).
					Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
					Where("id = ? AND urr.role_id = ?", uid, roleId).
					Count(&userRoleCount).Error
				if err != nil {
					return err
				}
				if userRoleCount == 0 {
					continue
				}
			}

			newUC := models.UserClass{UserId: uid, ClassId: classId}
			if err := tx.Create(&newUC).Error; err != nil {
				return err
			}
		}
	}

	r.UpdateCurrentStudentToClass(classId)

	return nil
}

func (r *classRepository) AddUserClass(classId int64, userIds []int64, roleId int64) error {
	tx := db.MasterDB

	for _, uid := range userIds {
		q := tx.Model(&models.UserClass{}).
			Where("class_id = ? AND user_id = ?", classId, uid)

		if roleId != 0 {
			q = q.Joins("JOIN users ON user_classes.user_id = users.id").
				Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
				Where("urr.role_id = ?", roleId)
		}

		var count int64
		if err := q.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			uc := models.UserClass{UserId: uid, ClassId: classId, IsCurrent: true}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		}
	}

	r.UpdateCurrentStudentToClass(classId)

	return nil
}

func (r *classRepository) UpdateCurrentStudentToClass(classId int64) (err error) {
	// Get school_id by class
	var class models.Class
	err = db.MasterDB.Unscoped().Where("id = ?", classId).First(&class).Error
	if err != nil {
		config.Log.Error("UpdateCurrentStudentToClass - get class ", err)
		return err
	}

	var count int64
	query := db.MasterDB.Model(&models.UserClass{}).
		Joins("JOIN users ON users.id = user_classes.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where(`
			user_classes.class_id = ?
			AND urr.role_id = ?
			AND users.deleted_at IS NULL
		`, classId, models.StudentRoleId)

	err = query.Count(&count).Error
	if err != nil {
		return err
	}

	err = db.MasterDB.Model(&models.Class{}).
		Where("id = ?", classId).
		Update("current_students", count).Error
	if err != nil {
		config.Log.Error("UpdateCurrentStudentToClass", err)
		return err
	}

	// Update student_count, class_count to school
	err = r.UpdateSchoolStats(class.SchoolId)
	if err != nil {
		config.Log.Error("UpdateCurrentStudentToClass - update school stats", err)
		return err
	}

	return nil
}

func (r *classRepository) UpdateSchoolStats(schoolId int64) error {
	var studentCount int64
	var classCount int64

	// Count class of school
	err := db.MasterDB.Model(&models.Class{}).
		Where("school_id = ?", schoolId).
		Count(&classCount).Error
	if err != nil {
		return err
	}

	// Count student of school
	err = db.MasterDB.
		Table("user_classes").
		Joins("JOIN users ON users.id = user_classes.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("JOIN classes ON classes.id = user_classes.class_id").
		Where(`
			classes.school_id = ? AND
			classes.deleted_at IS NULL AND
			urr.role_id = ? AND
			users.deleted_at IS NULL
		`, schoolId, models.StudentRoleId).
		Count(&studentCount).Error

	if err != nil {
		return err
	}

	// Update schools
	err = db.MasterDB.Model(&models.School{}).
		Where("id = ?", schoolId).
		Updates(map[string]interface{}{
			"student_count": studentCount,
			"class_count":   classCount,
		}).Error
	if err != nil {
		config.Log.Error("UpdateSchoolStats", err)
		return err
	}

	return nil
}

func (r *classRepository) GetUserClass(userId int64) error {
	return nil
}

func (r *classRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	var model models.Class

	var deletedUserIDs []int64

	if err := db.MasterDB.
		Model(&models.Class{}).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Pluck("id", &deletedUserIDs).Error; err != nil {
		return err
	}

	if len(deletedUserIDs) > 0 {
		if err := db.MasterDB.
			Where("class_id IN (?)", deletedUserIDs).
			Delete(&models.UserClass{}).Error; err != nil {
			return err
		}
	}

	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

func (r *classRepository) CreateOrUpdateClass(class *models.Class) (int64, error) {
	var existing models.Class
	var err error

	if class.ID > 0 {
		err = db.MasterDB.First(&existing, class.ID).Error
	}

	if err == nil {
		existing.Name = class.Name
		existing.FacultyId = class.FacultyId
		existing.SchoolId = class.SchoolId
		existing.GradeId = class.GradeId
		existing.MaxStudents = class.MaxStudents
		existing.TeacherInfo = class.TeacherInfo
		existing.Status = class.Status
		if saveErr := db.MasterDB.Save(&existing).Error; saveErr != nil {
			return 0, saveErr
		}
		return int64(existing.ID), nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.MasterDB.Create(class).Error; createErr != nil {
			return 0, createErr
		}
		return int64(class.ID), nil
	}

	return 0, err
}

func (r *classRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	schoolId := r.GetAdminSchoolId(ctx)

	if schoolId > 0 {
		query = query.Where("school_id = ?", schoolId)
	}

	return query
}

