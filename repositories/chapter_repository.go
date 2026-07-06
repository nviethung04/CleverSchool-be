package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChapterRepository interface {
	base.BaseRepositoryInterface[models.Chapter]
	base.BeforeQueryHook
	UpdateLessonChapterId(chapterId int64, lessonId int64) error
	GetPositionByIdAndProgram(programId, id int64) (int, error)
	ClearOldLessonByChapterId(chapterId int64, lessonIds []int64) error
}

type chapterRepository struct {
	*base.BaseRepository[models.Chapter]
}

func NewChapterRepository() ChapterRepository {
	repo := &chapterRepository{
		BaseRepository: base.NewBaseRepository[models.Chapter](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *chapterRepository) Create(entity *models.Chapter) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	if err := r.BeforeCreate(entity); err != nil {
		return err
	}

	query := db.MasterDB.Omit("author_id")
	if entity.CourseId == 0 {
		query = query.Omit("CourseId")
	}

	return query.Create(entity).Error
}

func (r *chapterRepository) Update(entity *models.Chapter) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	if err := r.BeforeUpdate(entity); err != nil {
		return err
	}

	omit := []string{"created_at", "created_by", "author_id", "clone_info"}
	if entity.CourseId == 0 {
		omit = append(omit, "CourseId")
	}

	return db.MasterDB.Omit(omit...).Save(entity).Error
}

// Override BaseRepository.Delete to avoid Save() overwriting nullable course_id with 0.
func (r *chapterRepository) Delete(id int) error {
	var model models.Chapter
	if err := r.BeforeDelete(id, &model); err != nil {
		return err
	}
	if err := db.MasterDB.Model(&models.Chapter{}).
		Where("id = ?", id).
		Update("deleted_by", model.DeletedBy).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&models.Chapter{}, id).Error
}

func (r *chapterRepository) UpdateLessonChapterId(chapterId int64, lessonId int64) error {
	return db.MasterDB.Model(&models.Lesson{}).
		Where("id = ?", lessonId).
		Update("chapter_id", chapterId).Error
}

func (r *chapterRepository) GetPositionByIdAndProgram(programId int64, id int64) (int, error) {
	var position int

	if id > 0 {
		err := db.MasterDB.Model(&models.Chapter{}).
			Select("sort_position").
			Where("id = ? AND program_id = ?", id, programId).
			Scan(&position).Error
		if err != nil {
			return position, err
		}
	}

	if position == 0 {
		var countZero int64
		if err := db.MasterDB.Model(&models.Chapter{}).
			Where("program_id = ? AND sort_position = 0 AND id != ?", programId, id).
			Count(&countZero).Error; err != nil {
			return position, err
		}

		if countZero > 0 {
			var maxPos sql.NullInt64
			if err := db.MasterDB.Model(&models.Chapter{}).
				Select("MAX(sort_position)").
				Where("program_id = ?", programId).
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
			return position, errors.New("no chapter with sort_position=0 found")
		}
	} else {
		return position, nil
	}
}

func (r *chapterRepository) ClearOldLessonByChapterId(chapterId int64, lessonIds []int64) error {
	return db.MasterDB.Model(&models.Lesson{}).
		Where("chapter_id = ? AND id NOT IN (?)", chapterId, lessonIds).
		Update("chapter_id", 0).Error
}

func (r *chapterRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	return query
}
