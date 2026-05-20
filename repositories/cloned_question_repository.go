package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
	"encoding/json"
	"fmt"
	"time"
)

type ClonedQuestion struct {
	ID             json.Number     `json:"id"`
	Type           string          `json:"type"`
	Content        json.RawMessage `json:"content"`
	Options        json.RawMessage `json:"options"`
	Metadata       json.RawMessage `json:"metadata"`
	CorrectAnswers json.RawMessage `json:"correct_answers"`
	// ... các trường khác nếu cần
}

type ClonedQuestionRepository interface {
	base.BaseRepositoryInterface[models.ClonedQuestion]
	FindByAssignment(assignmentId int64, assignmentType string) (*models.ClonedQuestion, error)
	PermanentlyDeleteOldRecords(before time.Time) error
	GetClonedQuestions(assignmentID int64, assignmentType string) ([]ClonedQuestion, error)
	DeleteByAssignment(assignmentId int64, assignmentType string) error
}

type clonedQuestionRepository struct {
	*base.BaseRepository[models.ClonedQuestion]
}

func NewClonedQuestionRepository() ClonedQuestionRepository {
	return &clonedQuestionRepository{
		BaseRepository: base.NewBaseRepository[models.ClonedQuestion](),
	}
}

func (r *clonedQuestionRepository) FindByAssignment(assignmentID int64, assignmentType string) (*models.ClonedQuestion, error) {
	var cloned models.ClonedQuestion
	err := db.MasterDB.
		Where("assignment_id = ? AND assignment_type = ?", assignmentID, assignmentType).
		First(&cloned).Error
	if err != nil {
		return nil, err
	}
	return &cloned, nil
}

func (r *clonedQuestionRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	tx := db.MasterDB

	deleteQuery := `
		DELETE FROM cloned_questions
		WHERE assignment_type = ?
			AND updated_at < ?
			AND NOT EXISTS (
				SELECT 1 FROM %s WHERE %s.id = cloned_questions.assignment_id
			)
	`

	type assignment struct {
		Type      string
		TableName string
	}

	assignments := []assignment{
		{models.ClonedQuestionTypeHomework, "homeworks"},
		{models.ClonedQuestionTypeExam, "exams"},
		{models.ClonedQuestionTypeExercise, "exercises"},
		{models.ClonedQuestionTypeLessonPlanPart, "lesson_plan_parts"},
		{models.ClonedQuestionTypeLevelTest, "level_tests"},
		{models.ClonedQuestionTypeContestRound, "contest_rounds"},
	}

	for _, a := range assignments {
		q := fmt.Sprintf(deleteQuery, a.TableName, a.TableName)
		if err := tx.Exec(q, a.Type, before).Error; err != nil {
			return err
		}
	}

	return nil
}

// Hàm mới: lấy danh sách câu hỏi từ trường questions dạng JSON
func (r *clonedQuestionRepository) GetClonedQuestions(assignmentID int64, assignmentType string) ([]ClonedQuestion, error) {
	var raw struct {
		Questions json.RawMessage `gorm:"column:questions"`
	}
	err := db.MasterDB.Table("cloned_questions").
		Select("questions").
		Where("assignment_id = ? AND assignment_type = ?", assignmentID, assignmentType).
		First(&raw).Error
	if err != nil {
		return nil, err
	}
	var questions []ClonedQuestion
	if err := json.Unmarshal(raw.Questions, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *clonedQuestionRepository) DeleteByAssignment(assignmentId int64, assignmentType string) error {
	return db.MasterDB.Where("assignment_id = ? AND assignment_type = ?", assignmentId, assignmentType).Delete(&models.ClonedQuestion{}).Error
}
