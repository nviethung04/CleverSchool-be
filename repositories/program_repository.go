package repositories

import (
	"database/sql"

	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"errors"

	"gorm.io/gorm"
)

type ProgramRepository interface {
	base.BaseRepositoryInterface[models.Program]
	DeleteProgram(id int) error
	FindCourseByProgramID(id int) models.Course
	GetProgramExportData(id int64) (*models.Program, []models.Chapter, error)
	GetProgramLessonPlans(id int64) ([]LessonPlanExportRow, error)
	CreateByCourse(courseId int64) (int64, error)
	ReSyncCourse(programId, courseId int64, syncIds []int64) error
	ReSyncLessonScheduleByCourse(newLessonId, oldLessonId, courseId int64) error
	ClearLessonSchedule(lessonIds []int64, courseId int64) error
	ReSyncHomeworkByCourse(homeworkId, syncHomeworkId, lessonId, syncLessonId, courseId int64) error
	ClearHomework(homeworkIds []int64, courseId int64) error
	UpdateCompletionLesson(newLessonId, oldLessonId int64) error
	ChapterUpdateSortPosition(id int64) error

	GetIdsBySchoolId(schoolId int64) []int64
	GetIdsBySubjectAndFaculty(subjectId, facultyId int64) []int64

	DeleteProgramRefSubjectsByProgramID(programID int64) error
	GetProgramRefSubjectsByProgramID(programID int64) ([]models.ProgramRefSubject, error)
	DeleteProgramRefSubjectsByProgramAndSubjectID(programID int64, subjectID int64) error
	UpdateOrCreateProgramRefSubject(programSubject models.ProgramRefSubject) error

	GetFailedUserIdsById(programId int64) ([]int64, error)
}

type programRepository struct {
	*base.BaseRepository[models.Program]
}

func NewProgramRepository() ProgramRepository {

	return &programRepository{
		BaseRepository: base.NewBaseRepository[models.Program](),
	}
}

func (r *programRepository) FindCourseByProgramID(id int) models.Course {
	var course models.Course
	db.MasterDB.Where("program_id = ?", id).First(&course)
	return course
}

func (r *programRepository) GetProgramExportData(id int64) (*models.Program, []models.Chapter, error) {
	var program models.Program
	if err := db.ReplicaDB.
		Preload("Subjects").
		Preload("Courses").
		First(&program, id).Error; err != nil {
		return nil, nil, err
	}

	var chapters []models.Chapter
	if err := db.ReplicaDB.
		Where("program_id = ?", id).
		Order("sort_position ASC").
		Order("id ASC").
		Preload("Lessons", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_position ASC").Order("id ASC")
		}).
		Preload("Headings", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_position ASC").Order("id ASC")
		}).
		Preload("Headings.Lessons", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("sort_position ASC").Order("id ASC")
		}).
		Find(&chapters).Error; err != nil {
		return nil, nil, err
	}

	return &program, chapters, nil
}

type LessonPlanExportRow struct {
	LessonID        int64
	LessonTitle     string
	LessonPlanID    int64
	LessonPlanName  string
	LessonPlanDesc  string
	LessonPlanParts []models.LessonPlanPart
}

func (r *programRepository) GetProgramLessonPlans(id int64) ([]LessonPlanExportRow, error) {
	var rows []LessonPlanExportRow

	query := `
		SELECT
			l.id AS lesson_id,
			l.title AS lesson_title,
			lp.id AS lesson_plan_id,
			lp.name AS lesson_plan_name,
			lp.description AS lesson_plan_description
		FROM lessons l
		JOIN chapters c ON c.id = l.chapter_id
		JOIN programs p ON p.id = c.program_id
		LEFT JOIN LATERAL (
			SELECT lprl.lesson_plan_id
			FROM lesson_plan_ref_lessons lprl
			WHERE lprl.lesson_id = l.id
				AND (lprl.course_id IS NULL OR lprl.course_id = 0)
			ORDER BY lprl.lesson_plan_id ASC
			LIMIT 1
		) AS primary_lp ON true
		LEFT JOIN lesson_plans lp ON lp.id = primary_lp.lesson_plan_id
		WHERE p.id = ?
		ORDER BY l.sort_position ASC, l.id ASC, lp.id ASC
	`

	type scanRow struct {
		LessonID       sql.NullInt64
		LessonTitle    sql.NullString
		LessonPlanID   sql.NullInt64
		LessonPlanName sql.NullString
		LessonPlanDesc sql.NullString
	}

	var temp []scanRow
	if err := db.ReplicaDB.Raw(query, id).Scan(&temp).Error; err != nil {
		return nil, err
	}

	partRepo := NewLessonPlanPartRepository()

	for _, r := range temp {
		lpParts := []models.LessonPlanPart{}
		if r.LessonPlanID.Valid && r.LessonPlanID.Int64 > 0 {
			lpParts, _ = partRepo.GetByLessonPlanID(r.LessonPlanID.Int64)
		}

		rows = append(rows, LessonPlanExportRow{
			LessonID:        r.LessonID.Int64,
			LessonTitle:     r.LessonTitle.String,
			LessonPlanID:    r.LessonPlanID.Int64,
			LessonPlanName:  r.LessonPlanName.String,
			LessonPlanDesc:  r.LessonPlanDesc.String,
			LessonPlanParts: lpParts,
		})
	}

	return rows, nil
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
	if err := tx.Table("lessons").Where("program_id = ?", programId).
		Pluck("id", &lessonIDs).Error; err != nil {
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
		`DELETE FROM lessons WHERE program_id = ?`,
		`DELETE FROM chapters WHERE program_id = ?`,
		`DELETE FROM courses WHERE program_id = ?`,
	}

	for _, q := range queries {
		if err := tx.Exec(q, programId).Error; err != nil {
			tx.Rollback()
			return err
		}
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
		// SubjectId:   course.SubjectId,
		ImageInfo: course.ImageInfo,
		Status:    course.Status,
		Target:    course.Target,
		Duration:  course.Duration,
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

func (r *programRepository) GetIdsBySchoolId(schoolId int64) []int64 {
	var programIDs []int64

	query := `
		SELECT DISTINCT c.program_id
		FROM courses c
		INNER JOIN course_schools cs ON cs.course_id = c.id
		WHERE cs.school_id = ?
			AND c.program_id IS NOT NULL
	`

	if err := db.ReplicaDB.Raw(query, schoolId).Scan(&programIDs).Error; err != nil {
		return programIDs
	}

	return programIDs
}

func (r *programRepository) GetIdsBySubjectAndFaculty(subjectId, facultyId int64) []int64 {
	var programIDs []int64

	if subjectId > 0 {
		query := `
			SELECT DISTINCT program_id
			FROM program_ref_subjects
			WHERE subject_id = ?
		`

		if err := db.ReplicaDB.Raw(query, subjectId).Scan(&programIDs).Error; err != nil {
			return programIDs
		}

		return programIDs
	}

	if facultyId > 0 {
		query := `
			SELECT DISTINCT prs.program_id
			FROM program_ref_subjects prs
			INNER JOIN subjects s ON s.id = prs.subject_id
			WHERE s.faculty_id = ?
		`

		if err := db.ReplicaDB.Raw(query, facultyId).Scan(&programIDs).Error; err != nil {
			return programIDs
		}

		return programIDs
	}

	return programIDs
}

func (r *programRepository) DeleteProgramRefSubjectsByProgramID(programID int64) error {
	if err := db.MasterDB.Exec("DELETE FROM program_ref_subjects WHERE program_id = ?", programID).Error; err != nil {
		return err
	}
	return nil
}

func (r *programRepository) GetProgramRefSubjectsByProgramID(programID int64) ([]models.ProgramRefSubject, error) {
	var programSubjects []models.ProgramRefSubject
	err := db.MasterDB.Where("program_id = ?", programID).Find(&programSubjects).Error
	return programSubjects, err
}

func (r *programRepository) DeleteProgramRefSubjectsByProgramAndSubjectID(programID int64, subjectID int64) error {
	var programSubjects []models.ProgramRefSubject
	if err := db.MasterDB.Where("program_id = ? AND subject_id = ?", programID, subjectID).
		Delete(&programSubjects).Error; err != nil {
		return err
	}
	return nil
}

func (r *programRepository) UpdateOrCreateProgramRefSubject(programSubject models.ProgramRefSubject) error {
	var existing models.ProgramRefSubject

	err := db.MasterDB.
		Where("program_id = ? AND subject_id = ?", programSubject.ProgramId, programSubject.SubjectId).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(&programSubject).Error
		}
		return err
	}

	return nil
}

func (r *programRepository) GetFailedUserIdsById(programId int64) ([]int64, error) {
	var failedUserIds []int64
	
	err := db.ReplicaDB.Table("user_courses").
		Joins("JOIN courses ON courses.id = user_courses.course_id").
		Where("courses.program_id = ? AND user_courses.is_failed = ?", programId, true).
		Distinct("user_courses.user_id").
		Pluck("user_courses.user_id", &failedUserIds).Error
	
	if err != nil {
		return nil, err
	}
	
	return failedUserIds, nil
}
