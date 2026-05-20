package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
	"database/sql"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChapterRepository interface {
	base.BaseRepositoryInterface[models.Chapter]
	base.BeforeQueryHook
	UpdateLessonChapterId(chapterId int64, lessonId int64) error
	GetPositionByIdAndProgram(programId, id int64) (int, error)
	ClearOldLessonByChapterId(chapterId int64, lessonIds []int64) error
	ValidTime(programId, chapterId int64, time int32) bool
	UpdateLessonVtg(chapter *models.Chapter) error
	UpdatePositionById(programId, id, position int) error
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
	query := db.MasterDB.Model(&models.Lesson{}).
		Where("chapter_id = ?", chapterId)

	if len(lessonIds) > 0 {
		query = query.Where("id NOT IN ?", lessonIds)
	}

	return query.Update("chapter_id", 0).Error
}

func (r *chapterRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	return query
}

func (r *chapterRepository) ValidTime(programId, chapterId int64, time int32) bool {
	if programId == 0 {
		return true
	}

	var program models.Program
	if err := db.ReplicaDB.
		Select("detail").
		First(&program, programId).Error; err != nil {
		return false
	}

	programTime, err := strconv.Atoi(program.Detail.VTGData.Time)
	if err != nil {
		return false
	}

	if programTime == 0 {
		return true
	}

	var totalChapterTime int32

	query := db.ReplicaDB.
		Model(&models.Chapter{}).
		Select("COALESCE(SUM(time), 0)").
		Where("program_id = ?", programId)

	if chapterId > 0 {
		query = query.Where("id <> ?", chapterId)
	}

	if err := query.Scan(&totalChapterTime).Error; err != nil {
		return false
	}

	return totalChapterTime+time <= int32(programTime)
}

func (r *chapterRepository) UpdateLessonVtg(chapter *models.Chapter) error {
	var lessons []models.Lesson

	if err := db.ReplicaDB.
		Where("chapter_id = ?", chapter.ID).
		Order("sort_position ASC, id ASC").
		Find(&lessons).Error; err != nil {
		return err
	}

	if len(lessons) == 0 {
		lesson := models.Lesson{
			ChapterID:    chapter.ID,
			Title:        chapter.Title,
			SortPosition: 0,
		}
		return db.MasterDB.Omit("author_id").Create(&lesson).Error
	}

	if len(lessons) > 1 {
		ids := make([]int64, 0, len(lessons)-1)
		for _, l := range lessons[1:] {
			ids = append(ids, l.ID)
		}

		if err := db.MasterDB.
			Where("id IN ?", ids).
			Delete(&models.Lesson{}).Error; err != nil {
			return err
		}
	}

	return db.MasterDB.
		Model(&models.Lesson{}).
		Where("id = ?", lessons[0].ID).
		Omit("author_id").
		Updates(map[string]interface{}{
			"title":         chapter.Title,
			"sort_position": 0,
		}).Error
}

func (r *chapterRepository) UpdatePositionById(programId, id, position int) error {
	err := db.MasterDB.
		Table("chapters").
		Where("id = ? AND deleted_at IS NULL AND program_id = ?", id, programId).
		Update("sort_position", position).Error

	return err
}
