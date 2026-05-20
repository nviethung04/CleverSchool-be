package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"fmt"
)

type FirebaseRepository interface {
	ValidateTarget(targetType string, targetValue int64) (bool, error)
	GetUserIDsByClass(classID int64) ([]int64, error)
	GetUserIDsByCourse(courseID int64) ([]int64, error)
	GetUserIDsBySchool(schoolID int64) ([]int64, error)
	GetSchoolIDsByUserID(userID int64) ([]int64, error)
}

type firebaseRepository struct{}

func NewFirebaseRepository() FirebaseRepository {
	return &firebaseRepository{}
}

func (r *firebaseRepository) ValidateTarget(targetType string, targetValue int64) (bool, error) {
	var count int64
	var err error

	switch targetType {
	case "class":
		err = db.ReplicaDB.
			Model(&models.Class{}).
			Where("id = ?", targetValue).
			Count(&count).Error

	case "course":
		err = db.ReplicaDB.
			Model(&models.Course{}).
			Where("id = ?", targetValue).
			Count(&count).Error

	case "user":
		err = db.ReplicaDB.
			Model(&models.User{}).
			Where("id = ?", targetValue).
			Count(&count).Error

	case "school":
		err = db.ReplicaDB.
			Model(&models.School{}).
			Where("id = ?", targetValue).
			Count(&count).Error

	default:
		return false, fmt.Errorf("invalid target type: %s", targetType)
	}

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *firebaseRepository) GetUserIDsByClass(classID int64) ([]int64, error) {
	var userIDs []int64

	err := db.ReplicaDB.
		Model(&models.UserClass{}).
		Where("class_id = ?", classID).
		Pluck("user_id", &userIDs).Error

	return userIDs, err
}

func (r *firebaseRepository) GetUserIDsByCourse(courseID int64) ([]int64, error) {
	var userIDs []int64
	err := db.ReplicaDB.
		Model(&models.UserCourse{}).
		Where("course_id = ?", courseID).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

func (r *firebaseRepository) GetUserIDsBySchool(schoolID int64) ([]int64, error) {
	var userIDs []int64

	err := db.ReplicaDB.
		Model(&models.UserClass{}).
		Joins("JOIN classes ON classes.id = user_classes.class_id").
		Joins("JOIN users ON users.id = user_classes.user_id").
		Where("classes.school_id = ?", schoolID).
		Where("classes.deleted_at IS NULL").
		Where("users.deleted_at IS NULL").
		Pluck("user_classes.user_id", &userIDs).Error

	return userIDs, err
}

func (r *firebaseRepository) GetSchoolIDsByUserID(userID int64) ([]int64, error) {
	var schoolIDs []int64

	err := db.ReplicaDB.
		Model(&models.UserClass{}).
		Joins("JOIN classes ON classes.id = user_classes.class_id").
		Where("user_classes.user_id = ?", userID).
		Where("classes.deleted_at IS NULL").
		Distinct().
		Pluck("classes.school_id", &schoolIDs).Error

	return schoolIDs, err
}
