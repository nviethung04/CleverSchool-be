package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"be-lms/requests"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LessonPlanRepository interface {
	GetAllWithPaging(req *requests.GetLessonPlanRequest, sort map[string]string, c *gin.Context) ([]models.LessonPlan, int64, error)
	GetByID(id int, c *gin.Context) (*models.LessonPlan, error)
	Create(lp *models.LessonPlan, lessonID int) error
	Update(lp *models.LessonPlan) error
	SyncUpdate(lp *models.LessonPlan) error
	Delete(id int, userID int64) error
	DeleteLessonPlansLesson(lessonPlanID, lessonID int) error
	Complete(id int, userID int64, status bool) error
	CompleteByIds(ids []int) ([]models.LessonPlanComplete, error)
}

type lessonPlanRepository struct{}

func NewLessonPlanRepository() LessonPlanRepository {
	return &lessonPlanRepository{}
}

//func (r *lessonPlanRepository) GetAll(req *requests.GetLessonPlanRequest) ([]models.LessonPlan, error) {
//	var lessonPlans []models.LessonPlan
//
//	query := db.ReplicaDB.Model(&models.LessonPlan{}).Where("is_deleted = false")
//
//	if req.LessonID > 0 {
//		query = query.
//			Joins("JOIN lesson_plan_ref_lessons ON lesson_plan_ref_lessons.lesson_plan_id = lesson_plans.id").
//			Where("lesson_plan_ref_lessons.lesson_id = ?", req.LessonID)
//	}
//
//	if req.Limit > 0 && req.Page > 0 {
//		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
//	}
//
//	err := query.Order("sort_position ASC").Find(&lessonPlans).Error
//	return lessonPlans, err
//}

func (r *lessonPlanRepository) GetAllWithPaging(req *requests.GetLessonPlanRequest, sort map[string]string, c *gin.Context) ([]models.LessonPlan, int64, error) {
	var lessonPlans []models.LessonPlan
	var total int64

	query := db.ReplicaDB.Model(&models.LessonPlan{})

	if req.LessonID != nil {
		query = query.
			Joins("JOIN lesson_plan_ref_lessons ON lesson_plan_ref_lessons.lesson_plan_id = lesson_plans.id").
			Where("lesson_plan_ref_lessons.lesson_id = ?", req.LessonID).
			Joins("JOIN lessons ON lessons.id = lesson_plan_ref_lessons.lesson_id")
	}

	if req.Keyword != "" {
		baseRepository := base.NewBaseRepository[models.LessonPlan]()
		baseRepository.SetSearch(req.Keyword, []string{"name", "description", "id"})
		query = baseRepository.ApplySearch(query)
	}

	query = r.ApplySort(query, sort)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	query = query.Preload("Complete")

	err := query.Order("sort_position ASC").Find(&lessonPlans).Error
	if err != nil {
		return nil, 0, err
	}
	return lessonPlans, total, nil
}

func (r *lessonPlanRepository) GetByID(id int, c *gin.Context) (*models.LessonPlan, error) {
	var lessonPlan models.LessonPlan
	query := db.ReplicaDB.Model(&models.LessonPlan{})

	err := query.Preload("Complete").First(&lessonPlan, id).Error
	if err != nil {
		return nil, err
	}

	err = r.AfterFindById(id, &lessonPlan)
	if err != nil {
		return nil, err
	}

	return &lessonPlan, nil
}

func (r *lessonPlanRepository) Create(lp *models.LessonPlan, lessonID int) error {
	tx := db.MasterDB.Begin()

	query := tx.Omit("course_id")
	if lp.ProgramId == 0 {
		query = query.Omit("ProgramId")
	}

	if err := query.Create(lp).Error; err != nil {
		tx.Rollback()
		return err
	}

	if lessonID > 0 {
		link := models.LessonPlansLesson{
			LessonPlanID: lp.ID,
			LessonID:     lessonID,
		}
		if err := tx.Create(&link).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *lessonPlanRepository) Update(lp *models.LessonPlan) error {
	updates := map[string]interface{}{}

	if lp.Name != "" {
		updates["name"] = lp.Name
	}
	if lp.Description != "" {
		updates["description"] = lp.Description
	}
	if lp.CoverImageInfo.Path != "" {
		updates["cover_image_info"] = lp.CoverImageInfo
	}
	if lp.Status != 0 {
		updates["status"] = lp.Status
	}
	if lp.SortPosition != 0 {
		updates["sort_position"] = lp.SortPosition
	}
	if lp.TotalTime != 0 {
		updates["total_time"] = lp.TotalTime
	}
	if lp.UpdatedBy != 0 {
		updates["updated_by"] = lp.UpdatedBy
	}

	updates["updated_at"] = time.Now()

	return db.MasterDB.Omit("course_id", "clone_info").Model(&models.LessonPlan{}).Where("id = ?", lp.ID).Updates(updates).Error
}

func (r *lessonPlanRepository) SyncUpdate(lp *models.LessonPlan) error {
	updates := map[string]interface{}{}

	if lp.Name != "" {
		updates["name"] = lp.Name
	}
	if lp.Description != "" {
		updates["description"] = lp.Description
	}
	if lp.CoverImageInfo.Path != "" {
		updates["cover_image_info"] = lp.CoverImageInfo
	}
	if lp.Status != 0 {
		updates["status"] = lp.Status
	}
	if lp.SortPosition != 0 {
		updates["sort_position"] = lp.SortPosition
	}
	if lp.TotalTime != 0 {
		updates["total_time"] = lp.TotalTime
	}
	if lp.UpdatedBy != 0 {
		updates["updated_by"] = lp.UpdatedBy
	}

	updates["clone_info"] = lp.CloneInfo

	updates["updated_at"] = time.Now()

	return db.MasterDB.Omit("course_id").Model(&models.LessonPlan{}).Where("id = ?", lp.ID).Updates(updates).Error
}

func (r *lessonPlanRepository) Delete(id int, deletedBy int64) error {
	if err := db.MasterDB.Model(&models.LessonPlan{}).
		Where("id = ?", id).
		Update("deleted_by", deletedBy).Error; err != nil {
		return err
	}
	return db.MasterDB.Delete(&models.LessonPlan{}, id).Error
}

func (r *lessonPlanRepository) DeleteLessonPlansLesson(lessonPlanID, lessonID int) error {
	// Xoá bản ghi trong bảng trung gian
	return db.MasterDB.
		Where("lesson_plan_id = ? AND lesson_id = ?", lessonPlanID, lessonID).
		Delete(&models.LessonPlansLesson{}).Error
}

func (r *lessonPlanRepository) AfterFindById(id int, entity *models.LessonPlan) error {
	// ✅ Async add views
	go func(e *models.LessonPlan) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Println("panic in async view update:", rec)
			}
		}()

		val := reflect.ValueOf(e).Elem()
		field := val.FieldByName("Views")

		if field.IsValid() && field.Kind() == reflect.Int {
			newViews := field.Int() + 1

			if err := db.MasterDB.Model(e).Update("views", newViews).Error; err != nil {
				fmt.Println("async update views error:", err)
			}
		}
	}(entity)

	return nil
}

func (r *lessonPlanRepository) ApplySort(query *gorm.DB, sort map[string]string) *gorm.DB {
	for _, v := range sort {
		order := strings.ToLower(v)
		if order == "newest" {
			return query.Order("created_at DESC")
		}
		if order == "oldest" {
			return query.Order("created_at ASC")
		}
	}

	for field, order := range sort {
		lowerField := strings.ToLower(field)
		query = query.Order(fmt.Sprintf("%s %s", lowerField, order))
	}

	return query
}

func (r *lessonPlanRepository) Complete(id int, userID int64, status bool) error {
	if status {
		var complete models.LessonPlanComplete
		err := db.MasterDB.
			Where("lesson_plan_id = ? AND user_id = ?", id, userID).
			First(&complete).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				complete = models.LessonPlanComplete{
					UserID:       userID,
					LessonPlanID: int64(id),
					CompletedAt:  time.Now(),
				}
				if err := db.MasterDB.Create(&complete).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else {
		if err := db.MasterDB.
			Where("lesson_plan_id = ? AND user_id = ?", id, userID).
			Delete(&models.LessonPlanComplete{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *lessonPlanRepository) CompleteByIds(ids []int) ([]models.LessonPlanComplete, error) {
	var completes []models.LessonPlanComplete

	err := db.MasterDB.
		Where("lesson_plan_id IN ?", ids).
		Find(&completes).Error

	if err != nil {
		return nil, err
	}

	return completes, nil
}
