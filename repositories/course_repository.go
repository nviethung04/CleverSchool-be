package repositories

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CourseRepository interface {
	base.BaseRepositoryInterface[models.Course]
	base.BeforeQueryHook
	GetUsers(courseId int64, roleId int64) ([]models.User, error)
	ReplaceUserCourse(courseId int64, userIds []int64, roleId int64, mainTeacherId int64) error
	AddUserCourse(courseId int64, userIds []int64, roleId int64) error
	DeleteOldSchool(courseId int64, schoolId int64) error
	StoreSchool(courseSchool models.CourseSchool) error
	DeleteNotInStudyShifts(courseID int64, validKeys []string) error
	StoreStudyShifts(courseRefStudyShifts []models.CourseRefStudyShift) error
	DeleteOldStudyShifts(courseId int64, StudyShiftIds []int64) error
	UpdateCurrentStudentToCourse(courseId int64) (err error)

	GetExams(courseId int64) ([]models.Exam, error)
	GetExamUsers(userId int64, courseId int64) ([]models.ExamUser, error)
	StoreSemester(courseId int64, semesterIds []int64) error
	GetCloneIds(courseId int64) []int64
}

type courseRepository struct {
	*base.BaseRepository[models.Course]
}

func NewCourseRepository() CourseRepository {
	repo := &courseRepository{
		BaseRepository: base.NewBaseRepository[models.Course](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func DeleteOldChapters(courseId int64, chapterIds []int64) error {
	if len(chapterIds) > 0 {
		if err := db.MasterDB.Where("id NOT IN (?) AND course_id = ?", chapterIds, courseId).Delete(&models.Chapter{}).Error; err != nil {
			return err
		}
	} else {
		if err := db.MasterDB.Where("course_id = ?", courseId).Delete(&models.Chapter{}).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (r *courseRepository) DeleteOldSchool(courseId int64, schoolId int64) error {
	if schoolId != 0 {
		if err := db.MasterDB.Where("school_id != ? AND course_id = ?", schoolId, courseId).Delete(&models.CourseSchool{}).Error; err != nil {
			return err
		}
	} else {
		if err := db.MasterDB.Where("course_id = ?", courseId).Delete(&models.CourseSchool{}).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (r *courseRepository) StoreSchool(courseSchool models.CourseSchool) error {
	return db.MasterDB.
		Where("course_id = ? AND school_id = ?", courseSchool.CourseId, courseSchool.SchoolId).
		FirstOrCreate(&courseSchool).Error
}

func (r *courseRepository) DeleteOldStudyShifts(courseId int64, StudyShiftIds []int64) error {
	if len(StudyShiftIds) > 0 {
		if err := db.MasterDB.Where("shift_id NOT IN (?) AND course_id = ?", StudyShiftIds, courseId).Delete(&models.CourseRefStudyShift{}).Error; err != nil {
			return err
		}
	} else {
		if err := db.MasterDB.Where("shift_id = ?", courseId).Delete(&models.CourseRefStudyShift{}).Error; err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (r *courseRepository) StoreStudyShifts(shifts []models.CourseRefStudyShift) error {
	return db.MasterDB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "course_id"}, {Name: "shift_id"}, {Name: "day_of_week"}},
			DoNothing: true,
		}).
		Create(&shifts).Error
}

func (r *courseRepository) DeleteNotInStudyShifts(courseID int64, validKeys []string) error {
	if len(validKeys) == 0 {
		return db.MasterDB.Where("course_id = ?", courseID).Delete(&models.CourseRefStudyShift{}).Error
	}

	query := db.MasterDB.Where("course_id = ?", courseID)

	for _, key := range validKeys {
		var shiftID, dayOfWeek int
		fmt.Sscanf(key, "%d_%d", &shiftID, &dayOfWeek)
		query = query.Not("shift_id = ? AND day_of_week = ?", shiftID, dayOfWeek)
	}

	return query.Delete(&models.CourseRefStudyShift{}).Error
}

func (r *courseRepository) GetUsers(courseId int64, roleId int64) ([]models.User, error) {
	var users []models.User

	query := db.MasterDB.
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Where("user_courses.course_id = ? AND users.deleted_at IS NULL", courseId).
		Preload("UserCourses", "course_id = ?", courseId)

	if roleId > 0 {
		query = query.Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("urr.role_id = ?", roleId).
			Order("user_courses.main_teacher DESC")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *courseRepository) ReplaceUserCourse(courseId int64, userIds []int64, roleId int64, mainTeacherId int64) error {
	if len(userIds) == 0 {
		userIds = []int64{0}
	}

	tx := db.MasterDB
	var userIdsToDelete []int64

	if roleId != 0 {
		err := tx.Table("user_courses AS uc").
			Joins("JOIN users AS u ON u.id = uc.user_id").
			Joins("JOIN user_ref_roles AS urr ON urr.user_id = u.id").
			Where("uc.course_id = ?", courseId).
			Where("urr.role_id = ?", roleId).
			Where("uc.user_id NOT IN ?", userIds).
			Pluck("uc.user_id", &userIdsToDelete).Error
		if err != nil {
			return err
		}
	} else {
		err := tx.Model(&models.UserCourse{}).
			Where("course_id = ? AND user_id NOT IN ?", courseId, userIds).
			Pluck("user_id", &userIdsToDelete).Error
		if err != nil {
			return err
		}
	}

	// Delete old user
	if len(userIdsToDelete) > 0 {
		if err := tx.
			Where("course_id = ? AND user_id IN ?", courseId, userIdsToDelete).
			Delete(&models.UserCourse{}).Error; err != nil {
			return err
		}
	}

	// Add or update users
	for _, uid := range userIds {
		var count int64
		query := tx.Table("user_courses AS uc").
			Where("uc.course_id = ? AND uc.user_id = ?", courseId, uid)

		if roleId != 0 {
			query = query.
				Joins("JOIN users AS u ON u.id = uc.user_id").
				Joins("JOIN user_ref_roles AS urr ON urr.user_id = u.id").
				Where("urr.role_id = ?", roleId)
		}

		if err := query.Count(&count).Error; err != nil {
			return err
		}

		if count != 1 {
			if roleId != 0 {
				var userRoleCount int64
				err := tx.Table("users AS u").
					Joins("JOIN user_ref_roles AS urr ON urr.user_id = u.id").
					Where(`
						u.id = ?
						AND urr.role_id = ?
						AND u.deleted_at IS NULL
					`, uid, roleId).
					Count(&userRoleCount).Error
				if err != nil {
					return err
				}
				if userRoleCount == 0 {
					continue
				}
			}

			newUC := models.UserCourse{
				UserId:      uid,
				CourseId:    courseId,
				MainTeacher: uid == mainTeacherId && roleId == models.TeacherRoleId,
			}
			if err := tx.Create(&newUC).Error; err != nil {
				return err
			}
		} else if roleId == models.TeacherRoleId {
			tx.Model(&models.UserCourse{}).
				Where("course_id = ? AND user_id = ?", courseId, uid).
				Update("main_teacher", uid == mainTeacherId)
		}
	}

	r.UpdateCurrentStudentToCourse(courseId)
	return nil
}

func (r *courseRepository) AddUserCourse(courseId int64, userIds []int64, roleId int64) error {
	tx := db.MasterDB

	for _, uid := range userIds {
		q := tx.Model(&models.UserCourse{}).
			Where("course_id = ? AND user_id = ?", courseId, uid)

		if roleId != 0 {
			q = q.Joins("JOIN users ON user_courses.user_id = users.id").
				Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
				Where("urr.role_id = ?", roleId)
		}

		var count int64
		if err := q.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			uc := models.UserCourse{UserId: uid, CourseId: courseId}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		}
	}

	r.UpdateCurrentStudentToCourse(courseId)

	return nil
}

func (r *courseRepository) UpdateCurrentStudentToCourse(courseId int64) (err error) {
	var count int64
	query := db.MasterDB.Model(&models.UserCourse{}).
		Joins("JOIN users ON users.id = user_courses.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("user_courses.course_id = ? AND urr.role_id = ? AND users.deleted_at IS NULL", courseId, models.StudentRoleId)

	err = query.Count(&count).Error
	if err != nil {
		return err
	}

	err = db.MasterDB.Model(&models.Course{}).
		Where("id = ?", courseId).
		Update("current_students", count).Error
	if err != nil {
		config.Log.Error("UpdateCurrentStudentToCourse", err)
		return err
	}
	return nil
}

func (r *courseRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	var model models.Course

	var deletedUserIDs []int64

	if err := db.MasterDB.
		Model(&models.Course{}).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Pluck("id", &deletedUserIDs).Error; err != nil {
		return err
	}

	if len(deletedUserIDs) > 0 {
		if err := db.MasterDB.
			Where("course_id IN (?)", deletedUserIDs).
			Delete(&models.UserCourse{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("course_id IN (?)", deletedUserIDs).
			Delete(&models.CourseSchool{}).Error; err != nil {
			return err
		}
	}

	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

func (r *courseRepository) GetExams(courseId int64) ([]models.Exam, error) {
	var exams []models.Exam

	err := db.ReplicaDB.Model(&models.Exam{}).
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id").
		Joins("JOIN courses ON courses.program_id = chapters.program_id").
		Where("courses.id = ? AND exams.deleted_at IS NULL", courseId).
		Preload("Lessons"). // Ensure "Lessons" relation is set in Exam model
		Find(&exams).Error

	if err != nil {
		return nil, err
	}

	return exams, nil
}

func (r *courseRepository) GetExamUsers(userId int64, courseId int64) ([]models.ExamUser, error) {
	var examUsers []models.ExamUser

	err := db.ReplicaDB.Model(&models.ExamUser{}).
		Joins("JOIN exams ON exams.id = exam_users.exam_id").
		Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Joins("JOIN lessons ON lessons.id = erl.lesson_id").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id").
		Joins("JOIN courses ON courses.program_id = chapters.program_id").
		Where("exam_users.user_id = ? AND courses.id = ?", userId, courseId).
		Find(&examUsers).Error

	if err != nil {
		return nil, err
	}

	return examUsers, nil
}

func (r *courseRepository) StoreSemester(courseId int64, semesterIds []int64) error {
	var existing []int64

	err := db.MasterDB.Model(&models.CourseRefSemester{}).
		Where("course_id = ?", courseId).
		Pluck("semester_id", &existing).Error
	if err != nil {
		return err
	}

	existingMap := make(map[int64]bool)
	for _, id := range existing {
		existingMap[id] = true
	}

	newMap := make(map[int64]bool)
	for _, id := range semesterIds {
		newMap[id] = true
	}

	var toAdd []models.CourseRefSemester
	for _, id := range semesterIds {
		if !existingMap[id] {
			toAdd = append(toAdd, models.CourseRefSemester{
				CourseId:   courseId,
				SemesterId: id,
			})
		}
	}

	var toDelete []int64
	for _, id := range existing {
		if !newMap[id] {
			toDelete = append(toDelete, id)
		}
	}

	if len(toAdd) > 0 {
		err = db.MasterDB.Create(&toAdd).Error
		if err != nil {
			return err
		}
	}

	if len(toDelete) > 0 {
		err = db.MasterDB.Where("course_id = ? AND semester_id IN ?", courseId, toDelete).
			Delete(&models.CourseRefSemester{}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *courseRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	return query
}

func (r *courseRepository) GetCloneIds(courseId int64) []int64 {
	var ids []int64
	err := db.MasterDB.
		Model(&models.Course{}).
		Where("clone_info->>'clone_id' = ?", fmt.Sprintf("%d", courseId)).
		Pluck("id", &ids).Error
	if err != nil {
		return nil
	}
	return ids
}
