package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ProgramRepository interface {
	base.BaseRepositoryInterface[models.Program]
	DeleteProgram(id int) error
	FindCourseByProgramID(id int) models.Course
	CreateByCourse(courseId int64) (int64, error)
	ReSyncCourse(programId, courseId int64, syncIds []int64) error
	ReSyncLessonScheduleByCourse(newLessonId, oldLessonId, courseId int64) error
	ClearLessonSchedule(lessonIds []int64, courseId int64) error
	ReSyncHomeworkByCourse(homeworkId, syncHomeworkId, lessonId, syncLessonId, courseId int64) error
	ClearHomework(homeworkIds []int64, courseId int64) error
	UpdateCompletionLesson(newLessonId, oldLessonId int64) error
	ChapterUpdateSortPosition(id int64) error
}

type programRepository struct{
	*base.BaseRepository[models.Program]
}

func NewProgramRepository() ProgramRepository {

	return &programRepository{
		BaseRepository: base.NewBaseRepository[models.Program](),
	}
}

func (r *programRepository) Update(entity *models.Program) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	if err := r.BeforeUpdate(entity); err != nil {
		return err
	}

	omit := []string{"created_at", "created_by", "author_id"}
	if entity.SubjectId == 0 {
		omit = append(omit, "SubjectId")
	}

	return db.MasterDB.Omit(omit...).Save(entity).Error
}

func (r *programRepository) FindCourseByProgramID(id int) models.Course {
	var course models.Course
	db.MasterDB.Where("program_id = ?", id).First(&course)
	return course
}

func (r *programRepository) DeleteProgram(programId int) error {
	tx := db.MasterDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Lấy danh sách lesson_plan_id theo program
	var lessonPlanIDs []int64
	if err := tx.Table("lesson_plans").Where("program_id = ?", programId).
		Pluck("id", &lessonPlanIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Lấy danh sách lesson_id theo program
	var lessonIDs []int64
	// NOTE: lessons không có cột program_id. program_id nằm ở chapters, lessons trỏ tới chapters bằng chapter_id.
	if err := tx.Table("lessons").
		Joins("JOIN chapters ON chapters.id = lessons.chapter_id").
		Where("chapters.program_id = ?", programId).
		Pluck("lessons.id", &lessonIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Xóa trong lesson_plan_ref_lessons
	if len(lessonPlanIDs) > 0 || len(lessonIDs) > 0 {
		if err := tx.Exec(`
			DELETE FROM lesson_plan_ref_lessons
			WHERE lesson_plan_id IN (?) OR lesson_id IN (?)`,
			lessonPlanIDs, lessonIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Xóa các bảng còn lại
	queries := []string{
		`DELETE FROM homeworks WHERE program_id = ?`,
		`DELETE FROM exams WHERE program_id = ?`,
		`DELETE FROM lesson_plan_parts WHERE program_id = ?`,
		`DELETE FROM lesson_plans WHERE program_id = ?`,
		`DELETE FROM chapters WHERE program_id = ?`,
		`DELETE FROM courses WHERE program_id = ?`,
	}

	for _, q := range queries {
		if err := tx.Exec(q, programId).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// lessons: delete theo chapters.program_id (vì lessons không có program_id)
	if err := tx.Exec(`
		DELETE FROM lessons
		USING chapters
		WHERE lessons.chapter_id = chapters.id
		  AND chapters.program_id = ?`, programId).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *programRepository) ReSyncCourse(programId, courseId int64, syncIds []int64) error {
    if err := db.MasterDB.Model(&models.Chapter{}).
        Where("course_id = ?", courseId).
        Update("program_id", programId).Error; err != nil {
        return err
    }

	if len(syncIds) > 0 {
        if err := db.MasterDB.Model(&models.Course{}).
            Where("id IN ?", syncIds).
            Update("program_id", programId).Error; err != nil {
            return err
        }
    }

	return nil
}

func (r *programRepository) ReSyncLessonScheduleByCourse(newLessonID, oldLessonID, courseID int64) error {
	if err := db.MasterDB.Model(&models.LessonSchedule{}).
        Where("lesson_id = ?", oldLessonID).
        Update("lesson_id", newLessonID).Error; err != nil {
        return err
	}
	return nil
}

func (r *programRepository) ClearLessonSchedule(lessonIds []int64, courseId int64) error {
	if err := db.MasterDB.
		Where("lesson_id NOT IN ? AND course_id = ?", lessonIds, courseId).
		Delete(&models.LessonSchedule{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *programRepository) ReSyncHomeworkByCourse(
    newHomeworkID, oldHomeworkID int64,
    newLessonID, oldLessonID int64,
    courseID int64,
) error {
	if err := db.MasterDB.Table("homework_users").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_users").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_positions").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_matchings").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_manual_scoring").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_labelings").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_groups").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

	if err := db.MasterDB.Table("homework_question_user_fill_in_blanks").
		Where("homework_id = ?", oldHomeworkID).
		Update("homework_id", newHomeworkID).Error; err != nil {
		return err
	}

    var ref models.HomeworkRefLesson

	err := db.MasterDB.
		Where("homework_id = ? AND lesson_id = ? AND course_id = ?", newHomeworkID, newLessonID, courseID).
		First(&ref).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
    	var oldRef models.HomeworkRefLesson
		db.MasterDB.
			Where("homework_id = ? AND lesson_id = ?", oldHomeworkID, oldLessonID).
			First(&oldRef)

		newRef := models.HomeworkRefLesson{
			HomeworkId: newHomeworkID,
			LessonId:   newLessonID,
			CourseId:   courseID,
			AssignedAt: oldRef.AssignedAt,
			AssignedBy: oldRef.AssignedBy,
		}
		return db.MasterDB.Create(&newRef).Error
	}

	return nil
}

func (r *programRepository) ClearHomework(homeworkIds []int64, courseId int64) error {
	if err := db.MasterDB.
		Where("homework_id NOT IN ? AND course_id = ?", homeworkIds, courseId).
		Delete(&models.HomeworkRefLesson{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *programRepository) CreateByCourse(courseId int64) (int64, error) {
	var course models.Course

	if err := db.MasterDB.First(&course, courseId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("course not found")
		}
		return 0, err
	}

	newProgram := models.Program{
		Name:        course.Name,
		Description: course.Description,
		SubjectId:   course.SubjectId,
		ImageInfo:   course.ImageInfo,
		Status:      course.Status,
		Target:      course.Target,
		Duration:    course.Duration,
	}

	if err := db.MasterDB.Create(&newProgram).Error; err != nil {
		return 0, err
	}

	return newProgram.ID, nil
}

func (r *programRepository) UpdateCompletionLesson(newLessonId, oldLessonId int64) error {
	if err := db.MasterDB.Table("lesson_completions").
		Where("lesson_id = ?", newLessonId).
		Update("lesson_id", oldLessonId).Error; err != nil {
		return err
	}

	return nil
}

func (r *programRepository) ChapterUpdateSortPosition(programID int64) error {
	var chapters []models.Chapter

	if err := db.MasterDB.
		Where("program_id = ?", programID).
		Order("sort_position ASC").
		Order("id ASC").
		Find(&chapters).Error; err != nil {
		return err
	}

	for idx, chapter := range chapters {
		if chapter.SortPosition != int(idx) {
			if err := db.MasterDB.Model(&models.Chapter{}).
				Where("id = ?", chapter.ID).
				Update("sort_position", idx).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
