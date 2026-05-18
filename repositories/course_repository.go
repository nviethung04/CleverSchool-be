package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CourseRepository interface {
	base.BaseRepositoryInterface[models.Course]
	base.BeforeQueryHook
	GetUsers(courseId int64, roleId int64) ([]models.User, []int64, error)
	ReplaceUserCourse(courseId int64, userIds []int64, roleId int64, mainTeacherId int64, failedUserIds []int64) error
	AddUserCourse(courseId int64, userIds []int64, roleId int64, failedUserIds []int64) error
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

	GetIdsByUserId(userId int64) []int64
	CreateOrUpdateCourse(course *models.Course) (int64, error)
	GetCoursesFamilyByProgramId(programId int64) ([]models.Course, []models.Course, error)

	NotEligibleForFinalExamIds(courseId, userId int64) ([]int64, error)

	GetCourseIdsBySubjectId(subjectId int64) ([]int64, error)
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

func (r *courseRepository) GetUsers(courseId int64, roleId int64) ([]models.User, []int64, error) {
	var users []models.User

	query := db.MasterDB.
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Where("user_courses.course_id = ? AND users.deleted_at IS NULL", courseId).
		Preload("UserCourses", "course_id = ?", courseId).
		Preload("UserCourses.Course").
		Preload("UserClasses").
		Preload("UserClasses.Class").
		Preload("UserClasses.Class.School").
		Preload("UserAddress").
		Preload("School").
		Preload("Subjects").
		Preload("Certificates").
		Preload("Degrees").
		Preload("Departments").
		Preload("Positions").
		Preload("Roles")

	if roleId > 0 {
		query = query.Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("urr.role_id = ?", roleId).
			Order("user_courses.main_teacher DESC")
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	// Get list of failed user IDs
	var failedUserIds []int64
	failedQuery := db.MasterDB.Table("user_courses").
		Where("course_id = ? AND is_failed = ?", courseId, true)

	if roleId > 0 {
		failedQuery = failedQuery.
			Joins("JOIN users ON users.id = user_courses.user_id").
			Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
			Where("urr.role_id = ? AND users.deleted_at IS NULL", roleId)
	}

	if err := failedQuery.Pluck("user_courses.user_id", &failedUserIds).Error; err != nil {
		return nil, nil, err
	}

	return users, failedUserIds, nil
}

func (r *courseRepository) ReplaceUserCourse(courseId int64, userIds []int64, roleId int64, mainTeacherId int64, failedUserIds []int64) error {
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
		if err := tx.Table("user_courses").
			Where("course_id = ? AND user_id IN ?", courseId, userIdsToDelete).
			Delete(&models.UserCourse{}).Error; err != nil {
			return err
		}
	}

	// Helper function to check if user is in failed list
	isUserFailed := func(userId int64) bool {
		for _, fid := range failedUserIds {
			if fid == userId {
				return true
			}
		}
		return false
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
				IsFailed:    isUserFailed(uid),
			}
			if err := tx.Create(&newUC).Error; err != nil {
				return err
			}
		} else {
			// Update existing user course
			updates := map[string]interface{}{
				"is_failed": isUserFailed(uid),
			}

			if roleId == models.TeacherRoleId {
				updates["main_teacher"] = uid == mainTeacherId
			}

			tx.Table("user_courses").
				Where("course_id = ? AND user_id = ?", courseId, uid).
				Updates(updates)
		}
	}

	r.UpdateCurrentStudentToCourse(courseId)
	return nil
}

func (r *courseRepository) AddUserCourse(courseId int64, userIds []int64, roleId int64, failedUserIds []int64) error {
	tx := db.MasterDB

	// Helper function to check if user is in failed list
	isUserFailed := func(userId int64) bool {
		for _, fid := range failedUserIds {
			if fid == userId {
				return true
			}
		}
		return false
	}

	for _, uid := range userIds {
		q := tx.Table("user_courses AS uc").
			Where("uc.course_id = ? AND uc.user_id = ?", courseId, uid)

		if roleId != 0 {
			q = q.Joins("JOIN users AS u ON u.id = uc.user_id").
				Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
				Where("urr.role_id = ?", roleId)
		}

		var count int64
		if err := q.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			uc := models.UserCourse{
				UserId:   uid,
				CourseId: courseId,
				IsFailed: isUserFailed(uid),
			}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		} else {
			// Update existing user course with is_failed
			tx.Table("user_courses").
				Where("course_id = ? AND user_id = ?", courseId, uid).
				Update("is_failed", isUserFailed(uid))
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
	schoolId := r.GetAdminSchoolId(ctx)

	if schoolId > 0 {
		query = query.Joins("JOIN course_schools cs ON cs.course_id = courses.id AND cs.school_id = ?", schoolId)
	}

	// Filter theo subject_id từ programs
	// Logic: courses JOIN programs theo courses.program_id = programs.id, lấy subject_id từ programs.subject_id
	if subjectIdStr := ctx.Query("subject_id"); subjectIdStr != "" {
		if subjectId, err := strconv.ParseInt(subjectIdStr, 10, 64); err == nil {
			if subjectId > 0 {

				// JOIN program_ref_subjects thay cho programs.subject_id
				query = query.
					Joins("JOIN program_ref_subjects prs ON prs.program_id = courses.program_id").
					Where("prs.subject_id = ?", subjectId)

			} else {
				// subject_id = 0 → lấy courses không có program hoặc không có mapping subject
				query = query.
					Joins("LEFT JOIN program_ref_subjects prs ON prs.program_id = courses.program_id").
					Where("courses.program_id IS NULL OR prs.subject_id IS NULL")
			}
		}
	}

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

func (r *courseRepository) GetIdsByUserId(userId int64) []int64 {
	var ids []int64
	err := db.ReplicaDB.
		Model(&models.UserCourse{}).
		Where("user_id = ?", userId).
		Pluck("course_id", &ids).Error
	if err != nil {
		return nil
	}
	return ids
}

func (r *courseRepository) CreateOrUpdateCourse(course *models.Course) (int64, error) {
	var existing models.Course
	var err error

	if course.ID > 0 {
		err = db.MasterDB.First(&existing, course.ID).Error
	}

	if err == nil {
		existing.Name = course.Name
		existing.ProgramId = course.ProgramId
		existing.Description = course.Description
		existing.Type = course.Type
		existing.Time = course.Time
		existing.Target = course.Target
		existing.StartDate = course.StartDate
		existing.EndDate = course.EndDate
		existing.ImageInfo = course.ImageInfo
		existing.Status = course.Status
		existing.Duration = course.Duration
		if saveErr := db.MasterDB.Save(&existing).Error; saveErr != nil {
			return 0, saveErr
		}
		return int64(existing.ID), nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.MasterDB.Create(course).Error; createErr != nil {
			return 0, createErr
		}
		return int64(course.ID), nil
	}

	return 0, err
}

func (r *courseRepository) GetCoursesFamilyByProgramId(programId int64) ([]models.Course, []models.Course, error) {
	var parentCourses []models.Course
	var allCourses []models.Course

	// Lấy tất cả courses thuộc program_id (bao gồm cả parent và children)
	err := db.ReplicaDB.Model(&models.Course{}).
		Where("program_id = ? AND deleted_at IS NULL", programId).
		Select("id, name, object_title, parent_course_id, program_id").
		Find(&allCourses).Error

	if err != nil {
		return nil, nil, err
	}

	// Lọc ra các parent courses (parent_course_id = 0)
	for _, course := range allCourses {
		if course.ParentCourseId == 0 {
			parentCourses = append(parentCourses, course)
		}
	}

	return parentCourses, allCourses, nil
}

func (r *courseRepository) NotEligibleForFinalExamIds(courseId, userId int64) ([]int64, error) {

	var courseIds []int64

	sql := `
		WITH ranked_scores AS (
		SELECT
			e.program_id,
			eu.score,
			ROW_NUMBER() OVER (
			PARTITION BY e.program_id
			ORDER BY eu.score DESC NULLS LAST
			) AS rn
		FROM exams e
		JOIN exam_users eu ON eu.exam_id = e.id
		WHERE e.type = ?
			AND eu.user_id = ?
			AND e.deleted_at IS NULL
			AND eu.score IS NOT NULL
		),
		program_requirements AS (
		SELECT
			p.id AS program_id,
			COALESCE((p.detail->'vtg_data'->>'number_of_frequent')::int, 0) AS required_count
		FROM programs p
		WHERE p.deleted_at IS NULL
		),
		user_program_stats AS (
		SELECT
			pr.program_id,
			pr.required_count,
			COUNT(rs.program_id) AS exam_count,
			MIN(CASE WHEN rs.rn <= pr.required_count THEN rs.score ELSE NULL END) AS min_top_score
		FROM program_requirements pr
		LEFT JOIN ranked_scores rs ON rs.program_id = pr.program_id
		GROUP BY pr.program_id, pr.required_count
		),
		not_eligible_programs AS (
		SELECT program_id
		FROM user_program_stats
		WHERE required_count > 0
			AND (
				exam_count < required_count
				OR min_top_score IS NULL
				OR min_top_score <= ?
			)
		),
		course_completion_stats AS (
		SELECT
			c.id AS course_id,
			c.program_id,
			COUNT(DISTINCT l.id) AS total_lessons,
			COUNT(DISTINCT lc.lesson_id) AS completed_lessons,
			CASE 
				WHEN COUNT(DISTINCT l.id) > 0 
				THEN (COUNT(DISTINCT lc.lesson_id) * 100.0 / COUNT(DISTINCT l.id))
				ELSE 0
			END AS completion_rate
		FROM courses c
		JOIN chapters ch ON ch.program_id = c.program_id AND ch.deleted_at IS NULL
		JOIN lessons l ON l.chapter_id = ch.id AND l.deleted_at IS NULL
		LEFT JOIN lesson_completions lc ON lc.lesson_id = l.id AND lc.student_id = ?
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.program_id
		)
		SELECT DISTINCT c.id
		FROM courses c
		JOIN user_courses uc ON uc.course_id = c.id
		LEFT JOIN not_eligible_programs nep ON nep.program_id = c.program_id
		LEFT JOIN course_completion_stats ccs ON ccs.course_id = c.id
		WHERE uc.user_id = ?
			AND c.deleted_at IS NULL
			AND (
				nep.program_id IS NOT NULL
				OR COALESCE(ccs.completion_rate, 0) < ?
			)
	`

	args := []interface{}{models.ExamTypeFrequent, userId, models.MinTopScore, userId, userId, models.LessonCompletionCompleted}
	if courseId > 0 {
		sql += ` AND c.id = ?`
		args = append(args, courseId)
	}

	err := db.ReplicaDB.Raw(sql, args...).Scan(&courseIds).Error
	return courseIds, err
}

func (r *courseRepository) GetCourseIdsBySubjectId(subjectId int64) ([]int64, error) {
	var courseIds []int64
	query := db.ReplicaDB.Model(&models.Course{})

	if subjectId > 0 {
		query = query.
			Joins("JOIN program_ref_subjects prs ON prs.program_id = courses.program_id").
			Where("prs.subject_id = ?", subjectId)
	} else {
		query = query.
			Joins("LEFT JOIN program_ref_subjects prs ON prs.program_id = courses.program_id").
			Where("courses.program_id IS NULL OR prs.subject_id IS NULL")
	}

	err := query.Pluck("courses.id", &courseIds).Error
	return courseIds, err
}
