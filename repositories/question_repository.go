package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
)

type QuestionRepository interface {
	base.BaseRepositoryInterface[models.Question]
	UpdateOrCreate(question models.Question) (int64, error)
	GetQuestionIdsAndScoresByLessonPlanPartId(lessonPlanPartkId int64) ([]QuesstionScore, error)
	GetQuestionIdsAndScoresByLevelTestId(examId int64) ([]QuesstionScore, error)
	UpdateOrCreateAttribute(attribute models.QuestionRefAttribute) error
	DeleteOldAttribute(questionId int64, attributeIds []int64) error

	GetHomeworkIDsByQuestionID(questionId int64) ([]int64, error)
	GetExamIDsByQuestionID(questionId int64) ([]int64, error)
	GetLessonPlanPartIDsByQuestionID(questionId int64) ([]int64, error)
	GetLevelTestIDsByQuestionID(questionId int64) ([]int64, error)
	GetQuestionIdsByQuestionAttributeIds(questionAttributeIDs []int64) ([]int64, error)
}

type questionRepository struct {
	*base.BaseRepository[models.Question]
}

func NewQuestionRepository() QuestionRepository {
	return &questionRepository{
		BaseRepository: base.NewBaseRepository[models.Question](),
	}
}

func questionOmitZeroSourceID(sourceQuestionID int64, omit []string) []string {
	if sourceQuestionID == 0 {
		return append(omit, "SourceQuestionId")
	}
	return omit
}

// Create omits source_question_id when 0 — FK requires NULL or valid source_questions.id.
func (r *questionRepository) Create(entity *models.Question) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	if err := r.BeforeCreate(entity); err != nil {
		return err
	}
	omit := questionOmitZeroSourceID(entity.SourceQuestionId, []string{"author_id"})
	return db.MasterDB.Omit(omit...).Create(entity).Error
}

func (r *questionRepository) Update(entity *models.Question) error {
	if entity == nil {
		return fmt.Errorf("entity is nil")
	}
	if err := r.BeforeUpdate(entity); err != nil {
		return err
	}
	omit := []string{"created_at", "created_by", "author_id"}
	omit = questionOmitZeroSourceID(entity.SourceQuestionId, omit)
	return db.MasterDB.Omit(omit...).Save(entity).Error
}

func (r *questionRepository) Delete(id int) error {
	var model models.Question

	if err := db.MasterDB.First(&model, id).Error; err != nil {
		return err
	}

	if err := r.BeforeDelete(id, &model); err != nil {
		return err
	}

	omit := questionOmitZeroSourceID(model.SourceQuestionId, []string{"author_id"})
	if err := db.MasterDB.Omit(omit...).Save(&model).Error; err != nil {
		return err
	}

	model.ID = int64(id)
	return db.MasterDB.Delete(&model).Error
}

func (r *questionRepository) UpdateOrCreate(question models.Question) (int64, error) {
	if question.ID != 0 {
		query := db.MasterDB.Model(&models.Question{}).Where("id = ?", question.ID)
		if question.SourceQuestionId == 0 {
			query = query.Omit("SourceQuestionId")
		}
		result := query.Updates(question)
		return question.ID, result.Error
	}

	query := db.MasterDB.Omit("author_id")
	if question.SourceQuestionId == 0 {
		query = query.Omit("SourceQuestionId")
	}
	result := query.Create(&question)
	return question.ID, result.Error
}

type QuesstionScore struct {
	QuestionID int64   `gorm:"column:question_id"`
	Score      float64 `gorm:"column:score"`
}

func (r *questionRepository) GetQuestionIdsAndScoresByLevelTestId(levelTestId int64) ([]QuesstionScore, error) {
	var records []QuesstionScore
	err := db.MasterDB.
		Table("level_test_questions").
		Select("question_id, score").
		Where("level_test_id = ?", levelTestId).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *questionRepository) GetQuestionIdsAndScoresByLessonPlanPartId(lessonPlanPartkId int64) ([]QuesstionScore, error) {
	var records []QuesstionScore
	err := db.MasterDB.
		Table("lesson_plan_part_questions").
		Select("question_id, score").
		Where("lesson_plan_part_id = ?", lessonPlanPartkId).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}


func StoreAnswers[T models.AnswerGroup | models.AnswerCoordinates | models.AnswerPosition | models.Answer | models.AnswerMatching](answers []T) error {
	if len(answers) > 0 {
		if err := db.MasterDB.Create(&answers).Error; err != nil {
			return err
		}
	}
	return nil
}

func DeleteOldAnswers[T models.AnswerGroup | models.AnswerCoordinates | models.AnswerPosition | models.Answer | models.AnswerMatching](questionId int64, ids []int64) error {
	var model T
	if len(ids) > 0 {
		if err := db.MasterDB.Where("id NOT IN (?) AND question_id = ?", ids, questionId).Delete(&model).Error; err != nil {
			return err
		}
	} else {
		if err := db.MasterDB.Where("question_id = ?", questionId).Delete(&model).Error; err != nil {
			return err
		}
	}
	return nil
}

func UpdateOrCreateAnswers[T any](answers []T) ([]int64, error) {
	if len(answers) == 0 {
		return nil, nil
	}

	var (
		toCreate []T
		toUpdate []T
	)

	for _, a := range answers {
		v := reflect.ValueOf(a)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		idField := v.FieldByName("ID")

		if idField.IsValid() && idField.Kind() == reflect.Int64 && idField.Int() != 0 {
			toUpdate = append(toUpdate, a)
		} else {
			toCreate = append(toCreate, a)
		}
	}

	for i := range toUpdate {
		if err := db.MasterDB.Save(&toUpdate[i]).Error; err != nil {
			return nil, err
		}
	}

	if len(toCreate) > 0 {
		ptrToCreate := make([]*T, len(toCreate))
		for i := range toCreate {
			ptrToCreate[i] = &toCreate[i]
		}
		if err := db.MasterDB.Create(&ptrToCreate).Error; err != nil {
			return nil, err
		}
	}

	var ids []int64

	for _, a := range toUpdate {
		v := reflect.ValueOf(a)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		idField := v.FieldByName("ID")
		if idField.IsValid() && idField.Kind() == reflect.Int64 {
			ids = append(ids, idField.Int())
		}
	}

	for _, a := range toCreate {
		v := reflect.ValueOf(a)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		idField := v.FieldByName("ID")
		if idField.IsValid() && idField.Kind() == reflect.Int64 {
			ids = append(ids, idField.Int())
		}
	}

	return ids, nil
}

func UpdateOrCreateGroup(group models.GroupAnswer) (int64, error) {
	if group.ID != 0 {
		if err := db.MasterDB.Save(&group).Error; err != nil {
			return 0, err
		}
		return group.ID, nil
	} else {
		if err := db.MasterDB.Create(&group).Error; err != nil {
			return 0, err
		}
		return group.ID, nil
	}
}

func (r *questionRepository) UpdateOrCreateAttribute(attribute models.QuestionRefAttribute) error {
	var existing models.QuestionRefAttribute

	err := db.MasterDB.
		Where("question_id = ? AND question_attribute_id = ?", attribute.QuestionID, attribute.AttributeID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(&attribute).Error
		}
		return err
	}

	existing.ParentAttributeID = attribute.ParentAttributeID

	return db.MasterDB.Save(&existing).Error
}

func (r *questionRepository) DeleteOldAttribute(questionId int64, attributeIds []int64) error {
	return db.MasterDB.
		Where("question_id = ? AND question_attribute_id NOT IN ?", questionId, attributeIds).
		Delete(&models.QuestionRefAttribute{}).Error
}

func (r *questionRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	var model models.Question

	var deletedUserIDs []int64

	if err := db.MasterDB.
		Model(&models.Question{}).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Pluck("id", &deletedUserIDs).Error; err != nil {
		return err
	}

	if len(deletedUserIDs) > 0 {
		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.AnswerCoordinates{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.AnswerGroup{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.AnswerMatching{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.AnswerPosition{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.Answer{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.AnswerMatching{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("question_id IN (?)", deletedUserIDs).
			Delete(&models.QuestionRefAttribute{}).Error; err != nil {
			return err
		}

		var groupIds []int64

		if err := db.MasterDB.
			Model(&models.AnswerGroup{}).
			Pluck("group_id", &groupIds).Error; err != nil {
			return err
		}

		if len(groupIds) > 0 {
			if err := db.MasterDB.
				Where("id NOT IN ?", groupIds).
				Delete(&models.GroupAnswer{}).Error; err != nil {
				return err
			}
		}
	}

	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

func (r *questionRepository) GetHomeworkIDsByQuestionID(questionID int64) ([]int64, error) {
	var homeworkIDs []int64

	err := db.MasterDB.
		Model(&models.HomeworkQuestion{}).
		Where("question_id = ?", questionID).
		Pluck("homework_id", &homeworkIDs).Error

	if err != nil {
		return nil, err
	}
	return homeworkIDs, nil
}

func (r *questionRepository) GetExamIDsByQuestionID(questionID int64) ([]int64, error) {
	var examIDs []int64

	err := db.MasterDB.
		Model(&models.ExamQuestion{}).
		Where("question_id = ?", questionID).
		Pluck("exam_id", &examIDs).Error

	if err != nil {
		return nil, err
	}
	return examIDs, nil
}

func (r *questionRepository) GetLessonPlanPartIDsByQuestionID(questionID int64) ([]int64, error) {
	var lessonPlanPartIDs []int64

	err := db.MasterDB.
		Model(&models.LessonPlanPartQuestion{}).
		Where("question_id = ?", questionID).
		Pluck("lesson_plan_part_id", &lessonPlanPartIDs).Error

	if err != nil {
		return nil, err
	}
	return lessonPlanPartIDs, nil
}

func (r *questionRepository) GetLevelTestIDsByQuestionID(questionID int64) ([]int64, error) {
	var levelTestIDs []int64

	err := db.MasterDB.
		Model(&models.LevelTestQuestion{}).
		Where("question_id = ?", questionID).
		Pluck("level_test_id", &levelTestIDs).Error

	if err != nil {
		return nil, err
	}
	return levelTestIDs, nil
}

func (r *questionRepository) GetQuestionIdsByQuestionAttributeIds(questionAttributeIDs []int64) ([]int64, error) {
	var questionIDs []int64

	if len(questionAttributeIDs) == 0 {
		return questionIDs, nil
	}

	err := db.MasterDB.
		Model(&models.QuestionRefAttribute{}).
		Where("question_attribute_id IN ?", questionAttributeIDs).
		Pluck("DISTINCT question_id", &questionIDs).Error

	if err != nil {
		return nil, err
	}

	return questionIDs, nil
}
