package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"

	"gorm.io/gorm/clause"
)

type H5pRepository interface {
	base.BaseRepositoryInterface[models.H5pContent]
	CreateOrUpdate(content *models.H5pContent) error
	StoreScore(models.H5pContentScore) error
	StoreContentUserData(record models.H5pContentUserData) error
	GetContentUserData(contentId string, userId string) (RawResult, error)
	GetContentUserDataByContentIdAndUser(contentId string, userId string) (RawResults, error)
	FindByContentID(contentId string) (*models.H5pContent, error)
	UpdateByContentId(content *models.H5pContent) error
	DeleteByContentId(contentId string) error
}
type RawResult struct {
	UserState  string `gorm:"column:user_state"`
	Invalidate bool   `gorm:"column:invalidate"`
	Preload    bool   `gorm:"column:preload"`
}

type RawResults []struct {
	UserState string `gorm:"column:user_state"`
}

type h5pContentRepository struct {
	*base.BaseRepository[models.H5pContent]
}

func NewH5pContentRepository() H5pRepository {
	return &h5pContentRepository{
		BaseRepository: base.NewBaseRepository[models.H5pContent](),
	}
}

func (h *h5pContentRepository) FindByContentID(contentId string) (*models.H5pContent, error) {
	var content models.H5pContent
	err := db.MasterDB.Where("content_id = ?", contentId).First(&content).Error
	return &content, err
}

func (h *h5pContentRepository) CreateOrUpdate(content *models.H5pContent) error {
	return db.MasterDB.Table("h5p_contents").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "content_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "parameters", "metadata",
			"updated_at", "updated_by",
		}),
	}).Create(&content).Error
}

func (h *h5pContentRepository) UpdateByContentId(content *models.H5pContent) error {
	err := db.MasterDB.Table("h5p_contents").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "content_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "parameters", "metadata",
			"updated_at", "updated_by",
		}),
	}).Save(&content).Error
	return err
}

func (h *h5pContentRepository) DeleteByContentId(contentId string) error {
	return db.MasterDB.Where("content_id = ?", contentId).Delete(&models.H5pContent{}).Error
}

func (h *h5pContentRepository) StoreContentUserData(record models.H5pContentUserData) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "content_id"}, {Name: "context_id"}, {Name: "sub_content_id"}, {Name: "user_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (h *h5pContentRepository) StoreScore(score models.H5pContentScore) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "content_id"}, {Name: "user_id"}},
		UpdateAll: true,
	}).Create(&score).Error
}

func (h *h5pContentRepository) GetContentUserData(contentId string, userId string) (RawResult, error) {
	rawResult := RawResult{}

	err := db.MasterDB.
		Table("h5p_content_user_data").
		Select("user_state", "invalidate", "preload").
		Where("content_id = ? AND user_id = ?", contentId, userId).
		Take(&rawResult).Error

	return rawResult, err
}

func (h *h5pContentRepository) GetContentUserDataByContentIdAndUser(contentId string, userId string) (RawResults, error) {
	rawResults := RawResults{}

	err := db.MasterDB.
		Table("h5p_content_user_data").
		Select("user_state").
		Where("content_id = ? AND user_id = ?", contentId, userId).
		Scan(&rawResults).Error

	return rawResults, err
}
