package repositories

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories/base"
	"database/sql"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LessonRepository interface {
	base.BaseRepositoryInterface[models.Lesson]
	base.BeforeQueryHook
	UpdateOrCreateDependency(dependency models.LessonDependency) error
	DeleteOldDependency(lessonId int64, dependencyIds []int64) error
	UpdateOrCreateTag(tag models.LessonRefTag) error
	DeleteOldTag(lessonId int64, tagIds []int64) error
	UpdateOrCreateTopic(topic models.LessonRefTopic) error
	DeleteOldTopic(lessonId int64, topicIds []int64) error
	UpdateOrCreateSkill(skill models.LessonRefSkill) error
	DeleteOldSkill(lessonId int64, skillIds []int64) error
	UpdateExamByLessonId(lessonId int64, examId int64) error
	UpdateHomeworkByLessonId(lessonId int64, examId int64) error
	UpdateOrCreateLessonPlanRefLesson(lessonId int64, lessonPlanId int64) error
	DeleteLessonPlanRefLesson(lessonId int64, lessonPlanIds []int64) error
	DropExamByLessonId(lessonId int64, examIds []int64) error
	DropHomeworkByLessonId(lessonId int64, homeworkIds []int64) error

	UpdateLessonSchedule(sch *models.LessonSchedule) error
	CreateLessonSchedules(schedules []models.LessonSchedule) error
	DeleteLessonSchedulesNotIn(lessonID int64, keepIDs []int64) error
	GetScheduleDateByWeek(weekID int64) (time.Time, error)
	GetWeekByDate(date time.Time) (*models.Week, error)
	DeleteLessonSchedulesByCourse(courseID int64, lessonIds []int64) error

	LessonIdsInSchedule(filters map[string]interface{}) ([]int64, error)
	GetCourseIdByLessonId(lessonId int64) (int64, error)
	Completion(lessonCompletion *prot.LessonCompletion) (*prot.LessonCompletion, error)
	CompletionLessonIds(studentId int64, courseId int64, chapterId int64) []int64
	CompletionLessonPlanIds(courseId int64, chapterId int64, lessonId int64) []int64
	UpdatePositionById(chapterId, id, position int) error

	UpdateHomeworkRefLesson(lessonId, courseId int64, homeworkIds []int64) error
	UpdateExamRefLesson(lessonId, courseId int64, examIds []int64) error
	UpdateExerciseRefLesson(lessonId, courseId int64, exerciseIds []int64) error
	CreateHomeworkRefLesson(lessonId int64, homeworkId int64) error
	CreateExamRefLesson(lessonId int64, examId int64) error
	CreateExerciseRefLesson(lessonId int64, exerciseId int64) error
	GetLessonSchedulesByCourse(courseID int64) ([]models.LessonSchedule, error)

	GetLessonPlanByCourse(lessonId, courseId int64) ([]*models.LessonPlan, []*models.LessonPlan, error)
	GetExamByCourse(lessonId, courseId int64) ([]*models.Exam, []*models.Exam, error)
	GetHomeworkByCourse(lessonId, courseId int64) ([]*models.Homework, []*models.Homework, error)
	GetExerciseByCourse(lessonId, courseId int64) ([]*models.Exercise, []*models.Exercise, error)

	StoreLessonPlanByCourse(lessonId, courseId int64, lessonPlanIds []int64) ([]*models.LessonPlan, []*models.LessonPlan, error)
	StoreExamByCourse(lessonId, courseId int64, examIds []int64) ([]*models.Exam, []*models.Exam, error)
	StoreHomeworkByCourse(lessonId, courseId int64, homeworkIds []int64) ([]*models.Homework, []*models.Homework, error)
	StoreExerciseByCourse(lessonId, courseId int64, exerciseIds []int64) ([]*models.Exercise, []*models.Exercise, error)

	UpdateLessonPlan(lessonId int64, lessonPlanId int64) error
	GetPositionByIdAndChapter(chapterId, id int64) (int, error)
}

type lessonRepository struct {
	*base.BaseRepository[models.Lesson]
}

func NewLessonRepository() LessonRepository {
	repo := &lessonRepository{
		BaseRepository: base.NewBaseRepository[models.Lesson](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *lessonRepository) UpdateOrCreateDependency(dependency models.LessonDependency) error {
	var existing models.LessonDependency
	err := db.MasterDB.
		Where("lesson_id = ? AND dependency_id = ?", dependency.LessonID, dependency.DependencyID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(&dependency).Error
	}

	return err
}

func (r *lessonRepository) DeleteOldDependency(lessonId int64, dependencyIds []int64) error {
	return db.MasterDB.
		Where("lesson_id = ? AND dependency_id NOT IN ?", lessonId, dependencyIds).
		Delete(&models.LessonDependency{}).Error
}

func (r *lessonRepository) UpdateOrCreateSkill(skill models.LessonRefSkill) error {
	var existing models.LessonRefSkill
	err := db.MasterDB.
		Where("lesson_id = ? AND skill_id = ?", skill.LessonID, skill.SkillID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(&skill).Error
	}
	return err
}

func (r *lessonRepository) DeleteOldSkill(lessonId int64, skillIds []int64) error {
	return db.MasterDB.
		Where("lesson_id = ? AND skill_id NOT IN ?", lessonId, skillIds).
		Delete(&models.LessonRefSkill{}).Error
}

func (r *lessonRepository) DropHomeworkByLessonId(lessonId int64, homeworkIds []int64) error {
    query := db.MasterDB.Table("homework_ref_lessons").Where("lesson_id = ?", lessonId)

    if len(homeworkIds) > 0 {
        query = query.Where("homework_id NOT IN ?", homeworkIds)
    }

    return query.Delete(nil).Error
}

func (r *lessonRepository) DropExamByLessonId(lessonId int64, examIds []int64) error {
    query := db.MasterDB.Table("exam_ref_lessons").Where("lesson_id = ?", lessonId)

    if len(examIds) > 0 {
        query = query.Where("exam_id NOT IN ?", examIds)
    }

    // Xoá các liên kết exam với lesson
    return query.Delete(nil).Error
}

func (r *lessonRepository) UpdateOrCreateTopic(topic models.LessonRefTopic) error {
	var existing models.LessonRefTopic
	err := db.MasterDB.
		Where("lesson_id = ? AND topic_id = ?", topic.LessonID, topic.TopicID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(&topic).Error
	}
	return err
}

func (r *lessonRepository) DeleteOldTopic(lessonId int64, topicIds []int64) error {
	return db.MasterDB.
		Where("lesson_id = ? AND topic_id NOT IN ?", lessonId, topicIds).
		Delete(&models.LessonRefTopic{}).Error
}

func (r *lessonRepository) UpdateOrCreateTag(tag models.LessonRefTag) error {
	var existing models.LessonRefTag
	err := db.MasterDB.
		Where("lesson_id = ? AND tag_id = ?", tag.LessonID, tag.TagID).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(&tag).Error
	}
	return err
}

func (r *lessonRepository) DeleteOldTag(lessonId int64, tagIds []int64) error {
	return db.MasterDB.
		Where("lesson_id = ? AND tag_id NOT IN ?", lessonId, tagIds).
		Delete(&models.LessonRefTag{}).Error
}

func (r *lessonRepository) UpdateExamByLessonId(lessonId int64, examId int64) error {
    if err := db.MasterDB.Table("exam_ref_lessons").
        Where("exam_id = ?", examId).
        Delete(nil).Error; err != nil {
        return err
	}

    return db.MasterDB.Table("exam_ref_lessons").
        Create(map[string]interface{}{
            "exam_id": examId,
            "lesson_id": lessonId,
        }).Error
}

func (r *lessonRepository) UpdateHomeworkByLessonId(lessonId int64, homeworkId int64) error {
    if err := db.MasterDB.Table("homework_ref_lessons").
        Where("homework_id = ?", homeworkId).
        Delete(nil).Error; err != nil {
        return err
	}

    return db.MasterDB.Table("homework_ref_lessons").
        Create(map[string]interface{}{
            "homework_id": homeworkId,
            "lesson_id": lessonId,
        }).Error
}

func (r *lessonRepository) UpdateOrCreateLessonPlanRefLesson(lessonId int64, lessonPlanId int64) error {
	var record models.LessonPlansLesson
	err := db.MasterDB.Where("lesson_plan_id = ? AND lesson_id = ?", lessonPlanId, lessonId).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newRecord := models.LessonPlansLesson{
				LessonPlanID: lessonPlanId,
				LessonID:     int(lessonId),
			}
			return db.MasterDB.Create(&newRecord).Error
		}
		return err
	}
	return nil
}

func (r *lessonRepository) DeleteLessonPlanRefLesson(lessonId int64, lessonPlanIds []int64) error {
	// 1. Xoá các dòng trùng nhau (giữ lại duy nhất một bản ghi cho mỗi lesson_id + lesson_plan_id)
	if err := db.MasterDB.Exec(`
		DELETE FROM lesson_plan_ref_lessons t
		USING (
			SELECT MIN(ctid) AS keep_ctid, lesson_id, lesson_plan_id
			FROM lesson_plan_ref_lessons
			WHERE lesson_id = ?
			GROUP BY lesson_id, lesson_plan_id
		) AS sub
		WHERE
			t.lesson_id = sub.lesson_id
			AND t.lesson_plan_id = sub.lesson_plan_id
			AND t.ctid <> sub.keep_ctid
			AND t.lesson_id = ?
	`, lessonId, lessonId).Error; err != nil {
		config.Log.Errorf("❌ Failed to delete duplicated rows: %v", err)
		return nil
	}

	// 2. Xoá những lesson_plan_id không có trong danh sách truyền vào
	if len(lessonPlanIds) > 0 {
		if err := db.MasterDB.
			Where("lesson_id = ? AND lesson_plan_id NOT IN ?", lessonId, lessonPlanIds).
			Delete(&models.LessonPlansLesson{}).Error; err != nil {
			config.Log.Error("failed to delete unmatched lesson plan ids: %w", err)
			return nil
		}
	} else {
		// Nếu danh sách lessonPlanIds rỗng thì xoá hết theo lessonId
		if err := db.MasterDB.
			Where("lesson_id = ?", lessonId).
			Delete(&models.LessonPlansLesson{}).Error; err != nil {
			config.Log.Error("failed to delete all by lessonId: %w", err)
			return nil
		}
	}

	return nil
}

func (r *lessonRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	var model models.Lesson

	var deletedLessonIDs []int64

	if err := db.MasterDB.
		Model(&models.Lesson{}).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Pluck("id", &deletedLessonIDs).Error; err != nil {
		return err
	}

	if len(deletedLessonIDs) > 0 {
		if err := db.MasterDB.
			Where("lesson_id IN (?)", deletedLessonIDs).
			Delete(&models.LessonRefSkill{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("lesson_id IN (?)", deletedLessonIDs).
			Delete(&models.LessonRefTag{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("lesson_id IN (?)", deletedLessonIDs).
			Delete(&models.LessonRefTopic{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("lesson_id IN (?)", deletedLessonIDs).
			Delete(&models.LessonSchedule{}).Error; err != nil {
			return err
		}
	}

	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

func (r *lessonRepository) UpdateLessonSchedule(sch *models.LessonSchedule) error {
	return db.MasterDB.Model(&models.LessonSchedule{}).Where("id = ?", sch.ID).Updates(sch).Error
}

func (r *lessonRepository) CreateLessonSchedules(schedules []models.LessonSchedule) error {
	return db.MasterDB.Create(&schedules).Error
}

func (r *lessonRepository) DeleteLessonSchedulesNotIn(lessonID int64, keepIDs []int64) error {
	query := db.MasterDB.Where("lesson_id = ?", lessonID)

	if len(keepIDs) > 0 {
		query = query.Where("id NOT IN ?", keepIDs)
	}

	return query.Delete(&models.LessonSchedule{}).Error
}

func (r *lessonRepository) GetScheduleDateByWeek(weekID int64) (time.Time, error) {
	var week models.Week
	if err := db.MasterDB.First(&week, weekID).Error; err != nil {
		return time.Time{}, err
	}

	return week.StartDate, nil
}

func (r *lessonRepository) GetWeekByDate(date time.Time) (*models.Week, error) {
	var week models.Week
	err := db.MasterDB.Where("start_date <= ? AND end_date >= ?", date, date).First(&week).Error
	return &week, err
}

func (r *lessonRepository) DeleteLessonSchedulesByCourse(courseID int64, lessonIds []int64) error {
	// Nếu có lessonIds, xóa tiếp theo lesson_id
	if len(lessonIds) > 0 {
		if err := db.MasterDB.
			Where("lesson_id IN ? AND course_id = ?", lessonIds, courseID).
			Delete(&models.LessonSchedule{}).Error; err != nil {
			return err
		}
	} else {
		if err := db.MasterDB.
			Where("course_id = ?", courseID).
			Delete(&models.LessonSchedule{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *lessonRepository) LessonIdsInSchedule(filters map[string]interface{}) ([]int64, error) {
	var lessonIDs []int64

	query := db.MasterDB.Model(&models.LessonSchedule{})

	if weekID, ok := filters["week_id"].(int64); ok && weekID > 0 {
		query = query.Where("week_id = ?", weekID)
	}
	if courseID, ok := filters["course_id"].(int64); ok && courseID > 0 {
		query = query.Where("lesson_schedules.course_id = ?", courseID)
	}

	if userID, ok := filters["user_id"].(int); ok && userID > 0 {
		query = query.Joins("JOIN user_courses ON user_courses.course_id = lesson_schedules.course_id").
			Where("user_courses.user_id = ?", userID)
	}

	err := query.
		Distinct("lesson_id").
		Pluck("lesson_id", &lessonIDs).Error

	return lessonIDs, err
}

func (r *lessonRepository) GetCourseIdByLessonId(lessonId int64) (int64, error) {
	var courseID int64

	err := db.ReplicaDB.
		Table("lessons l").
		Select("c.id").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id").
		Joins("JOIN courses c ON c.program_id = ch.program_id").
		Where("l.id = ? AND l.deleted_at IS NULL", lessonId).
		Scan(&courseID).Error

	if err != nil {
		return 0, err
	}

	return courseID, nil
}

func (r *lessonRepository) UpdatePositionById(chapterId, id, position int) error {
	err := db.MasterDB.
		Table("lessons").
		Where("id = ? AND deleted_at IS NULL AND chapter_id = ?", id, chapterId).
		Update("sort_position", position).Error

	return err
}

func (r *lessonRepository) Completion(lessonCompletion *prot.LessonCompletion) (*prot.LessonCompletion, error) {
	var existing models.LessonCompletion

	err := db.ReplicaDB.Where("lesson_id = ? AND student_id = ?", lessonCompletion.Id, lessonCompletion.StudentId).First(&existing).Error

	if lessonCompletion.IsComplete {
		if err == gorm.ErrRecordNotFound {
			newCompletion := models.LessonCompletion{
				LessonID:    lessonCompletion.Id,
				StudentID:   lessonCompletion.StudentId,
				CompletedAt: time.Now(),
			}
			if err := db.MasterDB.Create(&newCompletion).Error; err != nil {
				return nil, err
			}
			lessonCompletion.CompletionAt = newCompletion.CompletedAt.Format("2006-01-02")
			return lessonCompletion, nil
		} else if err != nil {
			return nil, err
		}

		lessonCompletion.CompletionAt = existing.CompletedAt.Format("2006-01-02")
		return lessonCompletion, nil
	} else {
		if err == nil {
			if err := db.MasterDB.Where("lesson_id = ? AND student_id = ?", lessonCompletion.Id, lessonCompletion.StudentId).Delete(&existing).Error; err != nil {
				return nil, err
			}
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		return lessonCompletion, nil
	}
}

func (r *lessonRepository) CompletionLessonIds(studentId int64, courseId int64, chapterId int64) []int64 {
	var lessonIds []int64

	if chapterId > 0 {
		var lessons []models.Lesson
		if err := db.ReplicaDB.
			Where("chapter_id = ?", chapterId).
			Find(&lessons).Error; err == nil {
			for _, l := range lessons {
				lessonIds = append(lessonIds, l.ID)
			}
		}
	} else if courseId > 0 {
		var chapters []models.Chapter
		if err := db.ReplicaDB.
			Where("course_id = ?", courseId).
			Find(&chapters).Error; err == nil {
			var chapterIDs []int64
			for _, ch := range chapters {
				chapterIDs = append(chapterIDs, ch.ID)
			}

			var lessons []models.Lesson
			if len(chapterIDs) > 0 {
				if err := db.ReplicaDB.
					Where("chapter_id IN ?", chapterIDs).
					Find(&lessons).Error; err == nil {
					for _, l := range lessons {
						lessonIds = append(lessonIds, l.ID)
					}
				}
			}
		}
	}

	if len(lessonIds) == 0 {
		var completions []models.LessonCompletion
		err := db.ReplicaDB.
			Where("student_id = ?", studentId).
			Find(&completions).Error

		if err != nil {
			return []int64{}
		}

		completedLessonIds := make([]int64, 0, len(completions))
		for _, c := range completions {
			completedLessonIds = append(completedLessonIds, c.LessonID)
		}

		return completedLessonIds
	}

	var completions []models.LessonCompletion
	err := db.ReplicaDB.
		Where("student_id = ? AND lesson_id IN ?", studentId, lessonIds).
		Find(&completions).Error

	if err != nil {
		return []int64{}
	}

	completedLessonIds := make([]int64, 0, len(completions))
	for _, c := range completions {
		completedLessonIds = append(completedLessonIds, c.LessonID)
	}

	return completedLessonIds
}

func (r *lessonRepository) CompletionLessonPlanIds(courseId, chapterId, lessonId int64) []int64 {
	var lessonIds []int64

	// Ưu tiên theo thứ tự: lessonId > chapterId > courseId
	if lessonId > 0 {
		lessonIds = append(lessonIds, lessonId)
	} else if chapterId > 0 {
		var lessons []models.Lesson
		if err := db.ReplicaDB.
			Where("chapter_id = ?", chapterId).
			Find(&lessons).Error; err == nil {
			for _, l := range lessons {
				lessonIds = append(lessonIds, l.ID)
			}
		}
	} else if courseId > 0 {
		var chapters []models.Chapter
		if err := db.ReplicaDB.
			Where("course_id = ?", courseId).
			Find(&chapters).Error; err == nil {
			var chapterIds []int64
			for _, ch := range chapters {
				chapterIds = append(chapterIds, ch.ID)
			}
			var lessons []models.Lesson
			if len(chapterIds) > 0 {
				if err := db.ReplicaDB.
					Where("chapter_id IN ?", chapterIds).
					Find(&lessons).Error; err == nil {
					for _, l := range lessons {
						lessonIds = append(lessonIds, l.ID)
					}
				}
			}
		}
	}

	if len(lessonIds) == 0 {
		return []int64{}
	}

	// Tìm lesson_plan_id từ bảng liên kết lesson_plan_ref_lessons
	var lessonPlanRef []struct {
		LessonPlanID int64
	}
	if err := db.ReplicaDB.
		Table("lesson_plan_ref_lessons").
		Select("lesson_plan_id").
		Where("lesson_id IN ?", lessonIds).
		Group("lesson_plan_id").
		Find(&lessonPlanRef).Error; err != nil {
		return []int64{}
	}

	if len(lessonPlanRef) == 0 {
		return []int64{}
	}

	var lessonPlanIDs []int64
	for _, ref := range lessonPlanRef {
		lessonPlanIDs = append(lessonPlanIDs, ref.LessonPlanID)
	}

	// Lọc những lesson_plan_id đã complete
	var completes []models.LessonPlanComplete
	if err := db.ReplicaDB.
		Where("lesson_plan_id IN ?", lessonPlanIDs).
		Find(&completes).Error; err != nil {
		return []int64{}
	}

	var completedPlanIDs []int64
	for _, c := range completes {
		completedPlanIDs = append(completedPlanIDs, c.LessonPlanID)
	}

	return completedPlanIDs
}

func (r *lessonRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	return query
}

func FilterProgram(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	programId := ctx.Query("program_id")

	if programId != "" {
		query = query.Where("lessons.program_id = ?", programId)
	} else {
		byProgram := ctx.Query("by_program")

		if byProgram == "true" || byProgram == "1" || byProgram == "yes" {
			query = query.Where("lessons.program_id >= 1")
		} else {
			query = query.Where("lessons.program_id IS NULL OR lessons.program_id = 0")
		}
	}

	return query
}

func (r *lessonRepository) UpdateExerciseRefLesson(lessonId, courseId int64, exerciseIds []int64) error {
	tx := db.MasterDB.Begin()

	var oldRefs []models.ExerciseRefLesson
	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		var programRefs []models.ExerciseRefLesson
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&programRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		idSet := make(map[int64]struct{}, len(exerciseIds))
		for _, id := range exerciseIds {
			idSet[id] = struct{}{}
		}
		for _, ref := range programRefs {
			if _, ok := idSet[ref.ExerciseId]; !ok {
				exerciseIds = append(exerciseIds, ref.ExerciseId)
				idSet[ref.ExerciseId] = struct{}{}
			}
		}
	}

	type oldAssign struct {
		AssignedBy *int64
		AssignedAt *time.Time
	}
	oldMap := make(map[int64]oldAssign)
	for _, ref := range oldRefs {
		oldMap[ref.ExerciseId] = oldAssign{
			AssignedBy: ref.AssignedBy,
			AssignedAt: ref.AssignedAt,
		}
	}

	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Delete(&models.ExerciseRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Delete(&models.ExerciseRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(exerciseIds) == 0 {
		return tx.Commit().Error
	}

	var refs []models.ExerciseRefLesson
	for _, exerciseId := range exerciseIds {
		ref := models.ExerciseRefLesson{
			LessonId:   lessonId,
			ExerciseId: exerciseId,
		}
		if courseId > 0 {
			ref.CourseId = courseId
		} else {
			ref.CourseId = 0
		}

		if old, ok := oldMap[exerciseId]; ok {
			ref.AssignedBy = old.AssignedBy
			ref.AssignedAt = old.AssignedAt
		}

		refs = append(refs, ref)
	}

	if err := tx.Create(&refs).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *lessonRepository) UpdateExamRefLesson(lessonId, courseId int64, examIds []int64) error {
	tx := db.MasterDB.Begin()

	var oldRefs []models.ExamRefLesson
	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		var programRefs []models.ExamRefLesson
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&programRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		idSet := make(map[int64]struct{}, len(examIds))
		for _, id := range examIds {
			idSet[id] = struct{}{}
		}
		for _, ref := range programRefs {
			if _, ok := idSet[ref.ExamId]; !ok {
				examIds = append(examIds, ref.ExamId)
				idSet[ref.ExamId] = struct{}{}
			}
		}
	}

	type oldAssign struct {
		AssignedBy *int64
		AssignedAt *time.Time
	}
	oldMap := make(map[int64]oldAssign)
	for _, ref := range oldRefs {
		oldMap[ref.ExamId] = oldAssign{
			AssignedBy: ref.AssignedBy,
			AssignedAt: ref.AssignedAt,
		}
	}

	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Delete(&models.ExamRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Delete(&models.ExamRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(examIds) == 0 {
		return tx.Commit().Error
	}

	var refs []models.ExamRefLesson
	for _, examId := range examIds {
		ref := models.ExamRefLesson{
			LessonId:   lessonId,
			ExamId: examId,
		}
		if courseId > 0 {
			ref.CourseId = courseId
		} else {
			ref.CourseId = 0
		}

		if old, ok := oldMap[examId]; ok {
			ref.AssignedBy = old.AssignedBy
			ref.AssignedAt = old.AssignedAt
		}

		refs = append(refs, ref)
	}

	if err := tx.Create(&refs).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *lessonRepository) UpdateHomeworkRefLesson(lessonId, courseId int64, homeworkIds []int64) error {
	tx := db.MasterDB.Begin()

	var oldRefs []models.HomeworkRefLesson
	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Find(&oldRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		var programRefs []models.HomeworkRefLesson
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Find(&programRefs).Error; err != nil {
			tx.Rollback()
			return err
		}

		idSet := make(map[int64]struct{}, len(homeworkIds))
		for _, id := range homeworkIds {
			idSet[id] = struct{}{}
		}
		for _, ref := range programRefs {
			if _, ok := idSet[ref.HomeworkId]; !ok {
				homeworkIds = append(homeworkIds, ref.HomeworkId)
				idSet[ref.HomeworkId] = struct{}{}
			}
		}
	}

	type oldAssign struct {
		AssignedBy *int64
		AssignedAt *time.Time
	}
	oldMap := make(map[int64]oldAssign)
	for _, ref := range oldRefs {
		oldMap[ref.HomeworkId] = oldAssign{
			AssignedBy: ref.AssignedBy,
			AssignedAt: ref.AssignedAt,
		}
	}

	if courseId == 0 {
		if err := tx.Where("lesson_id = ? AND (course_id IS NULL OR course_id = 0)", lessonId).
			Delete(&models.HomeworkRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
			Delete(&models.HomeworkRefLesson{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(homeworkIds) == 0 {
		return tx.Commit().Error
	}

	var refs []models.HomeworkRefLesson
	for _, homeworkId := range homeworkIds {
		ref := models.HomeworkRefLesson{
			LessonId:   lessonId,
			HomeworkId: homeworkId,
		}
		if courseId > 0 {
			ref.CourseId = courseId
		} else {
			ref.CourseId = 0
		}

		if old, ok := oldMap[homeworkId]; ok {
			ref.AssignedBy = old.AssignedBy
			ref.AssignedAt = old.AssignedAt
		}

		refs = append(refs, ref)
	}

	if err := tx.Create(&refs).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *lessonRepository) CreateHomeworkRefLesson(lessonId int64, homeworkId int64) error {
	ref := models.HomeworkRefLesson{
		LessonId:   lessonId,
		HomeworkId: homeworkId,
	}

	if err := db.MasterDB.
		Where("lesson_id = ? AND homework_id = ?", lessonId, homeworkId).
		FirstOrCreate(&ref).Error; err != nil {
		return err
	}

	return nil
}

func (r *lessonRepository) CreateExamRefLesson(lessonId int64, examId int64) error {
	ref := models.ExamRefLesson{
		LessonId:   lessonId,
		ExamId:     examId,
	}

	if err := db.MasterDB.
		Where("lesson_id = ? AND exam_id = ?", lessonId, examId).
		FirstOrCreate(&ref).Error; err != nil {
		return err
	}

	return nil
}

func (r *lessonRepository) CreateExerciseRefLesson(lessonId int64, exerciseId int64) error {
	ref := models.ExerciseRefLesson{
		LessonId:    lessonId,
		ExerciseId:  exerciseId,
	}

	if err := db.MasterDB.
		Where("lesson_id = ? AND exercise_id = ?", lessonId, exerciseId).
		FirstOrCreate(&ref).Error; err != nil {
		return err
	}

	return nil
}

func (r *lessonRepository) GetLessonSchedulesByCourse(courseID int64) ([]models.LessonSchedule, error) {
	var schedules []models.LessonSchedule
	if err := db.ReplicaDB.
		Where("course_id = ?", courseID).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *lessonRepository) GetLessonPlanByCourse(lessonId int64, courseId int64) ([]*models.LessonPlan, []*models.LessonPlan, error) {
	var programLessonPlans []*models.LessonPlan
    var courseLessonPlan  []*models.LessonPlan

    if err := db.ReplicaDB.
        Joins("JOIN lesson_plan_ref_lessons lrl ON lrl.lesson_plan_id = lesson_plans.id").
		Where("lrl.lesson_id = ? AND (lrl.course_id = 0 OR lrl.course_id IS NULL)", lessonId).
        Find(&programLessonPlans).Error; err != nil {
        return nil, nil, err
    }

    if err := db.ReplicaDB.
        Joins("JOIN lesson_plan_ref_lessons lrl ON lrl.lesson_plan_id = lesson_plans.id").
        Where("lrl.lesson_id = ? AND lrl.course_id = ?", lessonId, courseId).
        Find(&courseLessonPlan).Error; err != nil {
        return nil, nil, err
    }

    return programLessonPlans, courseLessonPlan, nil
}

func (r *lessonRepository) GetExamByCourse(lessonId int64, courseId int64) ([]*models.Exam, []*models.Exam,error) {
    var programExams []*models.Exam
    var courseExams  []*models.Exam

    if err := db.ReplicaDB.Preload("ExamRefLessons").
        Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
		Where("erl.lesson_id = ? AND (erl.course_id = 0 OR erl.course_id IS NULL)", lessonId).
        Find(&programExams).Error; err != nil {
        return nil, nil, err
    }

    if err := db.ReplicaDB.Preload("ExamRefLessons").
        Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
        Where("erl.lesson_id = ? AND erl.course_id = ?", lessonId, courseId).
        Find(&courseExams).Error; err != nil {
        return nil, nil, err
    }

    return programExams, courseExams, nil
}

func (r *lessonRepository) GetHomeworkByCourse(lessonId int64, courseId int64) ([]*models.Homework, []*models.Homework, error) {
    var programHomeworks []*models.Homework
    var courseHomeworks  []*models.Homework

    if err := db.ReplicaDB.Preload("HomeworkRefLessons").
        Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
        Where("hrl.lesson_id = ? AND (hrl.course_id = 0 OR hrl.course_id IS NULL)", lessonId).
        Find(&programHomeworks).Error; err != nil {
        return nil, nil, err
    }

    if err := db.ReplicaDB.Preload("HomeworkRefLessons").
        Joins("JOIN homework_ref_lessons hrl ON hrl.homework_id = homeworks.id").
        Where("hrl.lesson_id = ? AND hrl.course_id = ?", lessonId, courseId).
        Find(&courseHomeworks).Error; err != nil {
        return nil, nil, err
    }

    return programHomeworks, courseHomeworks, nil
}
func (r *lessonRepository)GetExerciseByCourse(lessonId int64, courseId int64) ([]*models.Exercise, []*models.Exercise, error) {
	var programExercises []*models.Exercise
    var courseExercises  []*models.Exercise

    if err := db.ReplicaDB.Preload("ExerciseRefLessons").
        Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id").
		Where("erl.lesson_id = ? AND (erl.course_id = 0 OR erl.course_id IS NULL)", lessonId).
        Find(&programExercises).Error; err != nil {
        return nil, nil, err
    }

    if err := db.ReplicaDB.Preload("ExerciseRefLessons").
        Joins("JOIN exercise_ref_lessons erl ON erl.exercise_id = exercises.id").
        Where("erl.lesson_id = ? AND erl.course_id = ?", lessonId, courseId).
        Find(&courseExercises).Error; err != nil {
        return nil, nil, err
    }

    return programExercises, courseExercises, nil
}

func (r *lessonRepository) StoreLessonPlanByCourse(lessonId, courseId int64, lessonPlanIds []int64) ([]*models.LessonPlan, []*models.LessonPlan, error) {
	config.Log.Infof("courseId: %d, lessonId: %d", courseId, lessonId)
	var programLessonPlanRefs []models.LessonPlanRefLesson
	if err := db.ReplicaDB.
		Where("lesson_id = ? AND (course_id = 0 OR course_id IS NULL)", lessonId).
		Find(&programLessonPlanRefs).Error; err != nil {
		return nil, nil, err
	}

	for _, ref := range programLessonPlanRefs {
		lessonPlanIds = append(lessonPlanIds, ref.LessonPlanId)
	}

	if err := db.MasterDB.
		Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
		Delete(&models.LessonPlanRefLesson{}).Error; err != nil {
		return nil, nil, err
	}

	unique := make(map[int64]struct{})
	for _, id := range lessonPlanIds {
		unique[id] = struct{}{}
	}

	for lpId := range unique {
		ref := &models.LessonPlanRefLesson{
			LessonId:     lessonId,
			LessonPlanId: lpId,
			CourseId:     courseId,
		}
		if err := db.MasterDB.Create(ref).Error; err != nil {
			return nil, nil, err
		}
	}

	return r.GetLessonPlanByCourse(lessonId, courseId)
}

func (r *lessonRepository) StoreExamByCourse(lessonId, courseId int64, examIds []int64) ([]*models.Exam, []*models.Exam, error) {
    var programExamRefs []models.ExamRefLesson
    if err := db.ReplicaDB.
        Where("lesson_id = ? AND (course_id = 0 OR course_id IS NULL)", lessonId).
        Find(&programExamRefs).Error; err != nil {
        return nil, nil, err
    }

    merged := make(map[int64]struct{})
    for _, id := range examIds {
        merged[id] = struct{}{}
    }
    for _, ref := range programExamRefs {
        merged[ref.ExamId] = struct{}{}
    }

    var oldRefs []models.ExamRefLesson
    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Find(&oldRefs).Error; err != nil {
        return nil, nil, err
    }
    assignedMap := make(map[int64]models.ExamRefLesson)
    for _, ref := range oldRefs {
        assignedMap[ref.ExamId] = ref
    }

    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Delete(&models.ExamRefLesson{}).Error; err != nil {
        return nil, nil, err
    }

    for examId := range merged {
        ref := &models.ExamRefLesson{
            ExamId:   examId,
            LessonId: lessonId,
            CourseId: courseId,
        }
        if old, ok := assignedMap[examId]; ok {
            ref.AssignedAt = old.AssignedAt
            ref.AssignedBy = old.AssignedBy
        }
        if err := db.MasterDB.Create(ref).Error; err != nil {
            return nil, nil, err
        }
    }

    return r.GetExamByCourse(lessonId, courseId)
}

func (r *lessonRepository) StoreHomeworkByCourse(lessonId, courseId int64, homeworkIds []int64) ([]*models.Homework, []*models.Homework, error) {
    var programHomeworkRefs []models.HomeworkRefLesson
    if err := db.ReplicaDB.
        Where("lesson_id = ? AND (course_id = 0 OR course_id IS NULL)", lessonId).
        Find(&programHomeworkRefs).Error; err != nil {
        return nil, nil, err
    }

    merged := make(map[int64]struct{})
    for _, id := range homeworkIds {
        merged[id] = struct{}{}
    }
    for _, ref := range programHomeworkRefs {
        merged[ref.HomeworkId] = struct{}{}
    }

    var oldRefs []models.HomeworkRefLesson
    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Find(&oldRefs).Error; err != nil {
        return nil, nil, err
    }
    assignedMap := make(map[int64]models.HomeworkRefLesson)
    for _, ref := range oldRefs {
        assignedMap[ref.HomeworkId] = ref
    }

    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Delete(&models.HomeworkRefLesson{}).Error; err != nil {
        return nil, nil, err
    }

    for homeworkId := range merged {
        ref := &models.HomeworkRefLesson{
            HomeworkId: homeworkId,
            LessonId:   lessonId,
            CourseId:   courseId,
        }
        if old, ok := assignedMap[homeworkId]; ok {
            ref.AssignedAt = old.AssignedAt
            ref.AssignedBy = old.AssignedBy
        }
        if err := db.MasterDB.Create(ref).Error; err != nil {
            return nil, nil, err
        }
    }

    return r.GetHomeworkByCourse(lessonId, courseId)
}

func (r *lessonRepository) StoreExerciseByCourse(lessonId, courseId int64, exerciseIds []int64) ([]*models.Exercise, []*models.Exercise, error) {
    var programExerciseRefs []models.ExerciseRefLesson
    if err := db.ReplicaDB.
        Where("lesson_id = ? AND (course_id = 0 OR course_id IS NULL)", lessonId).
        Find(&programExerciseRefs).Error; err != nil {
        return nil, nil, err
    }

    merged := make(map[int64]struct{})
    for _, id := range exerciseIds {
        merged[id] = struct{}{}
    }
    for _, ref := range programExerciseRefs {
        merged[ref.ExerciseId] = struct{}{}
    }

    var oldRefs []models.ExerciseRefLesson
    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Find(&oldRefs).Error; err != nil {
        return nil, nil, err
    }
    assignedMap := make(map[int64]models.ExerciseRefLesson)
    for _, ref := range oldRefs {
        assignedMap[ref.ExerciseId] = ref
    }

    if err := db.MasterDB.
        Where("lesson_id = ? AND course_id = ?", lessonId, courseId).
        Delete(&models.ExerciseRefLesson{}).Error; err != nil {
        return nil, nil, err
    }

    for exerciseId := range merged {
        ref := &models.ExerciseRefLesson{
            ExerciseId: exerciseId,
            LessonId:   lessonId,
            CourseId:   courseId,
        }
        if old, ok := assignedMap[exerciseId]; ok {
            ref.AssignedAt = old.AssignedAt
            ref.AssignedBy = old.AssignedBy
        }
        if err := db.MasterDB.Create(ref).Error; err != nil {
            return nil, nil, err
        }
    }

    return r.GetExerciseByCourse(lessonId, courseId)
}

func (r *lessonRepository) GetPositionByIdAndChapter(chapterId, id int64) (int, error) {
	var position int

	if id > 0 {
		err := db.MasterDB.Model(&models.Lesson{}).
			Where("id = ? AND chapter_id = ?", id, chapterId).
			Pluck("sort_position", &position).Error
		if err != nil {
			return position, err
		}
	}

	if position == 0 {
		var countZero int64
		if err := db.MasterDB.Model(&models.Lesson{}).
			Where("chapter_id = ? AND sort_position = 0 AND id != ?", chapterId, id).
			Count(&countZero).Error; err != nil {
			return position, err
		}

		if countZero > 0 {
			var maxPos sql.NullInt64
			if err := db.MasterDB.Model(&models.Lesson{}).
				Select("MAX(sort_position)").
				Where("chapter_id = ?", chapterId).
				Scan(&maxPos).Error; err != nil {
				return position, err
			}

			if maxPos.Valid {
				position = int(maxPos.Int64) + 1
			} else {
				position = 1
			}

			return position, nil
		} else {
			return position, errors.New("no lesson with sort_position=0 found")
		}
	} else {
		return position, nil
	}
}

func (r *lessonRepository) UpdateLessonPlan(lessonId int64, lessonPlanId int64) error {
	if err := db.MasterDB.
		Where("lesson_plan_id != ? AND lesson_id = ?", lessonPlanId, lessonId).
		Delete(&models.LessonPlanRefLesson{}).Error; err != nil {
		return err
	}

	if lessonPlanId == 0 {
		return nil
	}

	ref := models.LessonPlanRefLesson{
		LessonPlanId: lessonPlanId,
		LessonId:     lessonId,
	}
	if err := db.MasterDB.Create(&ref).Error; err != nil {
		return err
	}

	return nil
}
