package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"time"

	"github.com/gin-gonic/gin"
)

type LessonPlanPartRepository interface {
	GetAllWithPaging(req *requests.GetLessonPlanPartRequest, c *gin.Context) ([]models.LessonPlanPart, int64, error)
	GetByID(id int64, c *gin.Context) (*models.LessonPlanPart, error)
	Create(part *models.LessonPlanPart) error
	Update(part *models.LessonPlanPart) error
	Delete(id int64, userID int64) error
}

type lessonPlanPartRepository struct{}

func NewLessonPlanPartRepository() LessonPlanPartRepository {
	return &lessonPlanPartRepository{}
}

func (r *lessonPlanPartRepository) GetAllWithPaging(req *requests.GetLessonPlanPartRequest, c *gin.Context) ([]models.LessonPlanPart, int64, error) {
	var parts []models.LessonPlanPart
	var total int64

	query := db.ReplicaDB.Model(&models.LessonPlanPart{})

	if req.LessonPlanID > 0 {
		query = query.Where("lesson_plan_id = ?", req.LessonPlanID)
	}

	if req.CourseID > 0 {
		query = query.Where("(course_id = ? OR course_id IS NULL OR course_id = 0)", req.CourseID)
	} else {
		query = query.Where("(course_id IS NULL OR course_id = 0)")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	err := query.Find(&parts).Error
	if err != nil {
		return nil, 0, err
	}

	return parts, total, nil
}

func (r *lessonPlanPartRepository) GetByID(id int64, c *gin.Context) (*models.LessonPlanPart, error) {
	var part models.LessonPlanPart

	query := db.ReplicaDB.Model(&models.LessonPlanPart{})

	err := query.Where("id = ?", id).First(&part).Error
	if err != nil {
		return nil, err
	}
	return &part, nil
}

func (r *lessonPlanPartRepository) Create(part *models.LessonPlanPart) error {
	query := db.MasterDB
	omit := []string{}
	if part.ProgramId == 0 {
		omit = append(omit, "ProgramId")
	}
	if part.CourseID == 0 {
		omit = append(omit, "CourseID")
	}
	if len(omit) > 0 {
		query = query.Omit(omit...)
	}
	return query.Create(part).Error
}

func (r *lessonPlanPartRepository) Update(part *models.LessonPlanPart) error {
	updates := map[string]interface{}{}

	if part.LessonPlanID != 0 {
		updates["lesson_plan_id"] = part.LessonPlanID
	}

	// if part.CourseID != 0 {
	// 	updates["course_id"] = part.CourseID
	// }

	if part.Title != "" {
		updates["title"] = part.Title
	}
	if part.Tag != "" {
		updates["tag"] = part.Tag
	}
	if part.CoverImageInfo.Path != "" {
		updates["cover_image_info"] = part.CoverImageInfo
	}
	if part.SortPosition != 0 {
		updates["sort_position"] = part.SortPosition
	}
	if part.Time != 0 {
		updates["time"] = part.Time
	}
	updates["is_classwork"] = part.IsClasswork
	if part.FileType != "" {
		updates["file_type"] = part.FileType
	}
	if part.LinkInfo.Path != "" {
		updates["link_info"] = part.LinkInfo
	}
	if part.LinkType != "" {
		updates["link_type"] = part.LinkType
	}
	if part.GuideTeacher != "" {
		updates["guide_teacher"] = part.GuideTeacher
	}
	if part.GuideStudent != "" {
		updates["guide_student"] = part.GuideStudent
	}
	if part.File != "" {
		updates["file"] = part.File
	}

	updates["updated_by"] = 1
	updates["updated_at"] = time.Now()

	return db.MasterDB.Omit("clone_info").Model(&models.LessonPlanPart{}).Where("id = ?", part.ID).Updates(updates).Error
}

func (r *lessonPlanPartRepository) Delete(id int64, deletedBy int64) error {
	if err := db.MasterDB.Model(&models.LessonPlanPart{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&models.LessonPlanPart{}, id).Error
}
