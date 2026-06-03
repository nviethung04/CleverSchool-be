package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"encoding/json"
	"time"
)

type HomeworkInfo struct {
	Name           string
	Description    string
	Status         int32
	CoverImage     string
	IsAssigned     bool
	TotalQuestions int32
	CreatedAt      time.Time
}

type HomeworkStudentInfo struct {
	UserID                  int64
	Name                    string
	Username                string
	Avatar                  string
	QuestionsCompleted      int64
	LastQuestionIDCompleted int64
	UpdatedAt               time.Time
	ManualQuestionsCount    int32
	CourseID                int64
}

type HomeworkStudentRepository interface {
	GetHomeworkStudents(homeworkID int64, courseID int64) (HomeworkInfo, []HomeworkStudentInfo, error)
}

type homeworkStudentRepository struct{}

func NewHomeworkStudentRepository() HomeworkStudentRepository {
	return &homeworkStudentRepository{}
}

func (r *homeworkStudentRepository) GetHomeworkStudents(homeworkID int64, courseID int64) (HomeworkInfo, []HomeworkStudentInfo, error) {
	var info HomeworkInfo
	var students []HomeworkStudentInfo
	var isAssigned bool

	// Lấy thông tin homework
	var hw models.Homework
	err := db.ReplicaDB.Model(&models.Homework{}).
		Where("id = ?", homeworkID).
		Preload("HomeworkRefLessons").
		First(&hw).Error
	if err != nil {
		return info, students, err
	}

	for _, ref := range hw.HomeworkRefLessons {
		if ref.HomeworkId == homeworkID {
			isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
		}
	}

	info.Name = hw.Name
	info.Description = hw.Description
	info.Status = int32(hw.Status)
	info.CoverImage = hw.CoverImageInfo.Path
	info.CreatedAt = hw.CreatedAt
	info.IsAssigned = isAssigned

	// Lấy total_questions từ cloned_questions (đếm số lượng phần tử trong mảng questions)
	type questionsRaw struct {
		Questions json.RawMessage `gorm:"column:questions"`
	}
	var raw questionsRaw
	err = db.ReplicaDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", homeworkID, "homework").
		First(&raw).Error
	if err == nil && len(raw.Questions) > 0 {
		var arr []interface{}
		if err := json.Unmarshal(raw.Questions, &arr); err == nil {
			info.TotalQuestions = int32(len(arr))
		}
	}

	// Lấy danh sách học sinh với một query duy nhất
	type StudentResult struct {
		UserID                  int64            `gorm:"column:user_id"`
		Name                    string           `gorm:"column:name"`
		Username                string           `gorm:"column:username"`
		AvatarInfo              models.MediaInfo `gorm:"column:avatar_info"`
		QuestionsCompleted      int64            `gorm:"column:questions_completed"`
		LastQuestionIDCompleted int64            `gorm:"column:last_question_id_completed"`
		UpdatedAt               time.Time        `gorm:"column:updated_at"`
		ManualQuestionsCount    int32            `gorm:"column:manual_questions_count"`
		CourseID                int64            `gorm:"column:course_id"`
	}

	var studentResults []StudentResult
	query := db.ReplicaDB.Table("users").
		Select(`users.id as user_id, users.name, users.username, users.avatar_info,
			COALESCE(hu.questions_completed, 0) as questions_completed,
			COALESCE(hu.last_question_id_completed, 0) as last_question_id_completed,
			COALESCE(hu.updated_at, users.created_at) as updated_at,
			COALESCE(manual.count, 0) as manual_questions_count,
			hrl.course_id as course_id`).
		Joins("JOIN user_courses ON user_courses.user_id = users.id").
		Joins("JOIN courses ON courses.id = user_courses.course_id AND courses.deleted_at IS NULL").
		Joins("JOIN chapters ON chapters.program_id = courses.program_id AND chapters.deleted_at IS NULL").
		Joins("JOIN lessons ON lessons.chapter_id = chapters.id AND lessons.deleted_at IS NULL").
		Joins("JOIN homework_ref_lessons hrl ON hrl.lesson_id = lessons.id").
		Joins("JOIN homeworks ON homeworks.id = hrl.homework_id AND homeworks.deleted_at IS NULL").
		Joins(`LEFT JOIN homework_users hu ON hu.homework_id = homeworks.id AND hu.user_id = users.id`).
		Joins(`LEFT JOIN (
			SELECT user_id, COUNT(DISTINCT question_id) as count
			FROM homework_question_user_manual_scoring
			WHERE homework_id = ?
			GROUP BY user_id
		) as manual ON manual.user_id = users.id`, homeworkID).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("homeworks.id = ? AND urr.role_id = 3 AND users.deleted_at IS NULL", homeworkID)

	// Thêm điều kiện lọc theo course_id nếu có
	if courseID > 0 {
		query = query.Where("hrl.course_id = ?", courseID)
	}

	err = query.Scan(&studentResults).Error

	if err != nil {
		return info, students, err
	}

	for _, sr := range studentResults {
		students = append(students, HomeworkStudentInfo{
			UserID:                  sr.UserID,
			Name:                    sr.Name,
			Username:                sr.Username,
			Avatar:                  sr.AvatarInfo.Path,
			QuestionsCompleted:      sr.QuestionsCompleted,
			LastQuestionIDCompleted: sr.LastQuestionIDCompleted,
			UpdatedAt:               sr.UpdatedAt,
			ManualQuestionsCount:    sr.ManualQuestionsCount,
			CourseID:                sr.CourseID,
		})
	}
	return info, students, nil
}
