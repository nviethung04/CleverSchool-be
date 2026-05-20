package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories/base"
)

type HeadingRepository interface {
	base.BaseRepositoryInterface[models.Heading]
	UpdateOrCreateLesson(headingId int64, req *prot.HeadingRequest) error
	ValidTime(heading *models.Heading) bool
}

type headingRepository struct {
	*base.BaseRepository[models.Heading]
}

func NewHeadingRepository() HeadingRepository {
	return &headingRepository{
		BaseRepository: base.NewBaseRepository[models.Heading](),
	}
}

func (r *headingRepository) UpdateOrCreateLesson(headingId int64, req *prot.HeadingRequest) error {
	var count int64
	if err := db.ReplicaDB.Model(&models.Lesson{}).
		Where("heading_id = ?", headingId).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	lesson := models.Lesson{
		Title:     req.Name,
		HeadingID: headingId,
		ChapterID: req.ChapterId,
	}
	if err := db.MasterDB.Omit("author_id").Create(&lesson).Error; err != nil {
		return err
	}

	return nil
}

func (r *headingRepository) ValidTime(heading *models.Heading) bool {
	if heading.ChapterId <= 0 {
		return true
	}

	var chapterTime int32
	if err := db.ReplicaDB.
		Model(&models.Chapter{}).
		Where("id = ?", heading.ChapterId).
		Select("time").
		Scan(&chapterTime).Error; err != nil {
		return false
	}

	var totalHeadingTime int32
	query := db.ReplicaDB.
		Model(&models.Heading{}).
		Where("chapter_id = ?", heading.ChapterId)

	if heading.ID > 0 {
		query = query.Where("id <> ?", heading.ID)
	}

	if err := query.
		Select("COALESCE(SUM(time), 0)").
		Scan(&totalHeadingTime).Error; err != nil {
		return false
	}

	return totalHeadingTime+heading.Time <= chapterTime
}
