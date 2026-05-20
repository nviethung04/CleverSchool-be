package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type QuestionAttributeRepository interface {
	base.BaseRepositoryInterface[models.QuestionAttribute]
	GetParents() ([]models.QuestionAttribute, error)
}

type questionAttributeRepository struct {
	*base.BaseRepository[models.QuestionAttribute]
}

func NewQuestionAttributeRepository() QuestionAttributeRepository {
	return &questionAttributeRepository{
		BaseRepository: base.NewBaseRepository[models.QuestionAttribute](),
	}
}

func (r *questionAttributeRepository) GetParents() ([]models.QuestionAttribute, error) {
	var entities []models.QuestionAttribute

	query := db.ReplicaDB.Model(models.QuestionAttribute{})
	query = r.ApplySearch(query)

	err := query.Where("parent_id = ? OR parent_id IS NULL", 0).Preload("Nodes").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return entities, err
}
