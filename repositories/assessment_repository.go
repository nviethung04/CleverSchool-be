package repositories

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/requests"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssessmentRepository interface {
	GetAllWithPaging(req *requests.GetAssessmentRequest, c *gin.Context) ([]models.Assessment, int64, error)
	GetByID(id int64) (*models.Assessment, error)
	Create(entity *models.Assessment) error
	Update(entity *models.Assessment) error
	Delete(id int64, deletedBy int64) error
	AssignRefLesson(ref models.AssessmentRefLesson) error

	GetPublishCourseIds(assessmentId int64) []int64
	GetPublishAssessmentIds(courseId int64) []int64
	GetPublish(courseId, assessmentId int64) bool
	UpdatePublish(assessmentId, courseId int64, courseIds []int64, publish bool) bool
	FindPublishAssessments(assessmentIds []int64) ([]models.PublishAssessment, error)
}

type assessmentRepository struct{}

func NewAssessmentRepository() AssessmentRepository {
	return &assessmentRepository{}
}

func (r *assessmentRepository) GetAllWithPaging(req *requests.GetAssessmentRequest, c *gin.Context) ([]models.Assessment, int64, error) {
	var items []models.Assessment
	var total int64

	// Xác định có cần join với assessment_ref_lessons không
	needJoinRefLesson := req.CourseID != nil || req.LessonID != nil || req.ProgramID != nil

	// Nếu có course_id mà không có program_id, lấy program_id từ courses
	if req.CourseID != nil && req.ProgramID == nil {
		var programID int64
		err := db.ReplicaDB.Table("courses").
			Select("program_id").
			Where("id = ? AND deleted_at IS NULL", *req.CourseID).
			Scan(&programID).Error
		if err == nil && programID > 0 {
			req.ProgramID = &programID
		}
	}

	// Xác định có dùng LATERAL JOIN không (khi chỉ có course_id mà không có lesson_id và không có program_id)
	// Tính lại sau khi có thể đã lấy program_id từ courses
	useLateralJoin := req.CourseID != nil && req.LessonID == nil && req.ProgramID == nil && needJoinRefLesson

	// Build SELECT fields
	selectFields := `
		assessments.*,
		acg.id AS assessment_criteria_group_id,
		acg.name AS assessment_criteria_group_name,
		acg.has_file AS assessment_criteria_group_has_file,
		s.name AS subject_name,
		src.name AS study_report_criteria_name`

	if needJoinRefLesson {
		if useLateralJoin {
			// Với LATERAL JOIN, các field từ subquery đã có sẵn
			selectFields += `,
			arl.lesson_id AS lesson_id,
			arl.title AS lesson_title,
			CASE WHEN arl.assigned_at IS NOT NULL THEN true ELSE false END AS is_assigned`
		} else {
			selectFields += `,
			arl.lesson_id AS lesson_id,
			l.title AS lesson_title,
			CASE WHEN arl.assigned_at IS NOT NULL THEN true ELSE false END AS is_assigned`
		}
	}

	query := db.ReplicaDB.Table("assessments").
		Select(selectFields).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (arc.assessment_id) 
				arc.assessment_id,
				acgr.assessment_criteria_group_id AS id,
				acg.name,
				acg.has_file
			FROM assessment_ref_criteria arc
			JOIN assessment_criteria ac ON ac.id = arc.assessment_criterion_id AND ac.deleted_at IS NULL
			JOIN assessment_criteria_group_ref_criteria acgr ON acgr.assessment_criteria_id = ac.id AND acgr.deleted_at IS NULL
			JOIN assessment_criteria_groups acg ON acg.id = acgr.assessment_criteria_group_id AND acg.deleted_at IS NULL
			WHERE arc.deleted_at IS NULL
			ORDER BY arc.assessment_id, acgr.assessment_criteria_group_id
		) acg ON acg.assessment_id = assessments.id`).
		Joins("LEFT JOIN subjects s ON s.id = assessments.subject_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN study_report_criterias src ON src.id = assessments.study_report_criteria_id AND src.deleted_at IS NULL")

	// JOIN với assessment_ref_lessons và lessons nếu cần
	if needJoinRefLesson {
		joinCondition := "arl.assessment_id = assessments.id"
		if req.CourseID != nil {
			joinCondition += fmt.Sprintf(" AND arl.course_id = %d", *req.CourseID)
		}
		if req.LessonID != nil {
			joinCondition += fmt.Sprintf(" AND arl.lesson_id = %d", *req.LessonID)
		}

		// Nếu chỉ có course_id mà không có lesson_id và không có program_id, dùng LATERAL JOIN để lấy một assessment_ref_lesson đầu tiên
		if useLateralJoin {
			// Dùng LATERAL JOIN để lấy assessment_ref_lesson đầu tiên cho mỗi assessment
			query = query.Joins(`LEFT JOIN LATERAL (
				SELECT arl2.lesson_id, arl2.assigned_at, l2.title
				FROM assessment_ref_lessons arl2
				LEFT JOIN lessons l2 ON l2.id = arl2.lesson_id AND l2.deleted_at IS NULL
				WHERE arl2.assessment_id = assessments.id AND arl2.course_id = ?
				ORDER BY arl2.assigned_at DESC NULLS LAST
				LIMIT 1
			) arl ON true`, *req.CourseID)
			// JOIN với chapters từ arl.lesson_id (vì không có alias l khi dùng LATERAL JOIN)
			query = query.Joins("LEFT JOIN lessons l ON l.id = arl.lesson_id AND l.deleted_at IS NULL")
		} else {
			query = query.
				Joins(fmt.Sprintf("LEFT JOIN assessment_ref_lessons arl ON %s", joinCondition)).
				Joins("LEFT JOIN lessons l ON l.id = arl.lesson_id AND l.deleted_at IS NULL")
		}

		// Luôn JOIN với chapters để có thể filter theo program_id từ chapters
		query = query.Joins("LEFT JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL")
	}

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("unaccent(assessments.name) ILIKE unaccent(?) OR unaccent(assessments.description) ILIKE unaccent(?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	if req.Type != "" {
		query = query.Where("assessments.type = ?", req.Type)
	}

	// Lấy assessment theo program_id từ chapters.program_id (không phải assessments.program_id)
	if req.ProgramID != nil {
		// Nếu chưa có join với assessment_ref_lessons, cần join để lấy program_id từ chapters
		if !needJoinRefLesson {
			query = query.
				Joins("LEFT JOIN assessment_ref_lessons arl_prog ON arl_prog.assessment_id = assessments.id").
				Joins("LEFT JOIN lessons l_prog ON l_prog.id = arl_prog.lesson_id AND l_prog.deleted_at IS NULL").
				Joins("LEFT JOIN chapters ch_prog ON ch_prog.id = l_prog.chapter_id AND ch_prog.deleted_at IS NULL")
			query = query.Where("ch_prog.program_id = ?", *req.ProgramID)
		} else {
			// Nếu đã có join với assessment_ref_lessons và chapters (ch), filter theo ch.program_id
			query = query.Where("ch.program_id = ?", *req.ProgramID)
		}
	}

	// Filter theo subject_id (từ bảng assessments trực tiếp)
	if req.SubjectID != nil {
		query = query.Where("assessments.subject_id = ?", *req.SubjectID)
	}

	// Filter theo class_main_id
	// Logic: classes (class_main_id) -> user_classes -> users -> user_courses -> courses -> assessment_ref_lessons -> assessments
	// Chỉ lấy users có role_id = 3 (học sinh)
	// Khi filter theo class_main_id, dùng query riêng để lấy DISTINCT assessment_id, sau đó WHERE IN
	var assessmentIDsFromClassMain []int64
	if req.ClassMainID != nil {
		// Lấy danh sách course_id từ các users thuộc classes có class_main_id này và có role_id = 3
		var courseIDs []int64
		err := db.ReplicaDB.Table("user_courses uc").
			Select("DISTINCT uc.course_id").
			Joins("JOIN user_classes ucl ON ucl.user_id = uc.user_id").
			Joins("JOIN classes c ON c.id = ucl.class_id").
			Joins("JOIN user_ref_roles urr ON urr.user_id = uc.user_id").
			Where("c.class_main_id = ? AND c.deleted_at IS NULL AND urr.role_id = 3", *req.ClassMainID).
			Pluck("uc.course_id", &courseIDs).Error

		if err != nil || len(courseIDs) == 0 {
			// Nếu có lỗi hoặc không có course nào, trả về rỗng
			query = query.Where("1 = 0")
		} else {
			// Query riêng để lấy DISTINCT assessment_id từ các course_id này
			err = db.ReplicaDB.Table("assessment_ref_lessons").
				Select("DISTINCT assessment_id").
				Where("course_id IN (?)", courseIDs).
				Pluck("assessment_id", &assessmentIDsFromClassMain).Error

			if err != nil || len(assessmentIDsFromClassMain) == 0 {
				// Nếu có lỗi hoặc không có assessment nào, trả về rỗng
				query = query.Where("1 = 0")
			} else {
				// Dùng WHERE IN với danh sách assessment_id
				query = query.Where("assessments.id IN (?)", assessmentIDsFromClassMain)
			}
		}
	}

	// Filter theo grade_id
	// Logic: classes (grade_id) -> user_classes -> users -> user_courses -> courses -> assessment_ref_lessons -> assessments
	// Chỉ lấy users có role_id = 3 (học sinh)
	// Khi filter theo grade_id, dùng query riêng để lấy DISTINCT assessment_id, sau đó WHERE IN
	var assessmentIDsFromGrade []int64
	if req.GradeID != nil {
		// Lấy danh sách course_id từ các users thuộc classes có grade_id này và có role_id = 3
		var courseIDs []int64
		err := db.ReplicaDB.Table("user_courses uc").
			Select("DISTINCT uc.course_id").
			Joins("JOIN user_classes ucl ON ucl.user_id = uc.user_id").
			Joins("JOIN classes c ON c.id = ucl.class_id").
			Joins("JOIN user_ref_roles urr ON urr.user_id = uc.user_id").
			Where("c.grade_id = ? AND c.deleted_at IS NULL AND urr.role_id = 3", *req.GradeID).
			Pluck("uc.course_id", &courseIDs).Error

		if err != nil || len(courseIDs) == 0 {
			// Nếu có lỗi hoặc không có course nào, trả về rỗng
			query = query.Where("1 = 0")
		} else {
			// Query riêng để lấy DISTINCT assessment_id từ các course_id này
			err = db.ReplicaDB.Table("assessment_ref_lessons").
				Select("DISTINCT assessment_id").
				Where("course_id IN (?)", courseIDs).
				Pluck("assessment_id", &assessmentIDsFromGrade).Error

			if err != nil || len(assessmentIDsFromGrade) == 0 {
				// Nếu có lỗi hoặc không có assessment nào, trả về rỗng
				query = query.Where("1 = 0")
			} else {
				// Dùng WHERE IN với danh sách assessment_id (không cần DISTINCT nữa vì đã filter)
				query = query.Where("assessments.id IN (?)", assessmentIDsFromGrade)
			}
		}
	}

	query = query.Where("assessments.deleted_at IS NULL")

	// Count total
	// Nếu đã filter theo grade_id hoặc class_main_id, total = số lượng assessment_id trong danh sách
	if req.GradeID != nil && len(assessmentIDsFromGrade) > 0 {
		total = int64(len(assessmentIDsFromGrade))
		// Không cần DISTINCT nữa vì đã filter bằng WHERE IN
	} else if req.ClassMainID != nil && len(assessmentIDsFromClassMain) > 0 {
		total = int64(len(assessmentIDsFromClassMain))
		// Không cần DISTINCT nữa vì đã filter bằng WHERE IN
	} else {
		// Count total - không cần DISTINCT vì không có JOIN gây duplicate
		// Chỉ khi có JOIN với assessment_ref_lessons mà không filter cụ thể mới cần DISTINCT
		// Nhưng trong trường hợp này, nếu có needJoinRefLesson thì đã có filter cụ thể (course_id, lesson_id, program_id)
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
	}

	if req.Limit > 0 && req.Page > 0 {
		query = query.Limit(req.Limit).Offset((req.Page - 1) * req.Limit)
	}

	if err := query.Order("assessments.created_at DESC").Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *assessmentRepository) GetByID(id int64) (*models.Assessment, error) {
	var entity models.Assessment

	// Lấy các trường từ join trước
	var groupID int64
	var groupName, subjectName, studyReportName string
	query := db.ReplicaDB.Table("assessments").
		Select(`
			assessments.*,
			acg.id AS assessment_criteria_group_id,
			acg.name AS assessment_criteria_group_name,
			acg.has_file AS assessment_criteria_group_has_file,
			s.name AS subject_name,
			src.name AS study_report_criteria_name
		`).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (arc.assessment_id) 
				arc.assessment_id,
				acgr.assessment_criteria_group_id AS id,
				acg.name,
				acg.has_file
			FROM assessment_ref_criteria arc
			JOIN assessment_criteria ac ON ac.id = arc.assessment_criterion_id AND ac.deleted_at IS NULL
			JOIN assessment_criteria_group_ref_criteria acgr ON acgr.assessment_criteria_id = ac.id AND acgr.deleted_at IS NULL
			JOIN assessment_criteria_groups acg ON acg.id = acgr.assessment_criteria_group_id AND acg.deleted_at IS NULL
			WHERE arc.deleted_at IS NULL
			ORDER BY arc.assessment_id, acgr.assessment_criteria_group_id
		) acg ON acg.assessment_id = assessments.id`).
		Joins("LEFT JOIN subjects s ON s.id = assessments.subject_id AND s.deleted_at IS NULL").
		Joins("LEFT JOIN study_report_criterias src ON src.id = assessments.study_report_criteria_id AND src.deleted_at IS NULL").
		Where("assessments.id = ? AND assessments.deleted_at IS NULL", id)

	if err := query.Scan(&entity).Error; err != nil {
		return nil, err
	}

	// Lưu các trường từ join trước khi preload
	groupID = entity.AssessmentCriteriaGroupID
	groupName = entity.AssessmentCriteriaGroupName
	groupHasFile := entity.AssessmentCriteriaGroupHasFile
	subjectName = entity.SubjectName
	studyReportName = entity.StudyReportCriteriaName

	// Preload criteria và subcriteria
	if err := db.ReplicaDB.
		Preload("AssessmentRefCriteria", "deleted_at IS NULL").
		Preload("AssessmentRefCriteria.AssessmentCriterion.AssessmentSubcriteria").
		Where("id = ?", entity.ID).
		First(&entity).Error; err != nil {
		return nil, err
	}

	// Gán lại các trường từ join sau khi preload
	entity.AssessmentCriteriaGroupID = groupID
	entity.AssessmentCriteriaGroupName = groupName
	entity.AssessmentCriteriaGroupHasFile = groupHasFile
	entity.SubjectName = subjectName
	entity.StudyReportCriteriaName = studyReportName

	return &entity, nil
}

func (r *assessmentRepository) Create(entity *models.Assessment) error {
	return db.MasterDB.Create(entity).Error
}

func (r *assessmentRepository) Update(entity *models.Assessment) error {
	updates := map[string]interface{}{}

	if entity.Name != "" {
		updates["name"] = entity.Name
	}
	if entity.Description != "" {
		updates["description"] = entity.Description
	}
	if entity.Type != "" {
		updates["type"] = entity.Type
	}
	if entity.ProgramID != 0 {
		updates["program_id"] = entity.ProgramID
	}
	// Cập nhật has_file nếu được set
	updates["has_file"] = entity.HasFile
	updates["updated_by"] = entity.UpdatedBy
	updates["update_at"] = entity.UpdatedAt

	return db.MasterDB.Model(&models.Assessment{}).Where("id = ?", entity.ID).Updates(updates).Error
}

func (r *assessmentRepository) Delete(id int64, deletedBy int64) error {
	entity := models.Assessment{}
	if err := db.MasterDB.First(&entity, id).Error; err != nil {
		return err
	}

	entity.DeletedBy = deletedBy

	if err := db.MasterDB.Save(&entity).Error; err != nil {
		return err
	}

	return db.MasterDB.Delete(&entity).Error
}

func (r *assessmentRepository) AssignRefLesson(ref models.AssessmentRefLesson) error {
	return db.MasterDB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "assessment_id"},
			{Name: "lesson_id"},
			{Name: "course_id"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"assigned_by": ref.AssignedBy,
			"assigned_at": ref.AssignedAt,
		}),
	}).Create(&ref).Error
}

func (r *assessmentRepository) FindPublishAssessments(assessmentIds []int64) ([]models.PublishAssessment, error) {
	publishes := make([]models.PublishAssessment, 0)

	if len(assessmentIds) == 0 {
		return publishes, nil
	}

	if err := db.ReplicaDB.
		Where("assessment_id IN ?", assessmentIds).
		Find(&publishes).Error; err != nil {
		return nil, err
	}

	return publishes, nil
}

func (r *assessmentRepository) GetPublishCourseIds(assessmentId int64) []int64 {
	var courseIds []int64

	err := db.ReplicaDB.
		Model(&models.PublishAssessment{}).
		Where("assessment_id = ?", assessmentId).
		Pluck("course_id", &courseIds).Error

	if err != nil {
		config.Log.Error(err)
	}

	return courseIds
}

func (r *assessmentRepository) GetPublishAssessmentIds(courseId int64) []int64 {
	var assessmentIds []int64

	err := db.ReplicaDB.
		Model(&models.PublishAssessment{}).
		Where("course_id = ?", courseId).
		Pluck("assessment_id", &assessmentIds).Error

	if err != nil {
		config.Log.Error(err)
		return assessmentIds
	}

	return assessmentIds
}

func (r *assessmentRepository) GetPublish(courseId, assessmentId int64) bool {
	var publish models.PublishAssessment

	err := db.ReplicaDB.
		Where("course_id = ? AND assessment_id = ?", courseId, assessmentId).
		First(&publish).Error

	if err != nil || err == gorm.ErrRecordNotFound {
		return false
	}

	return true
}

func (r *assessmentRepository) UpdatePublish(assessmentId, courseId int64, courseIds []int64, publish bool) bool {
	if courseId > 0 {
		if publish {
			var count int64
			if err := db.MasterDB.
				Model(&models.PublishAssessment{}).
				Where("assessment_id = ? AND course_id = ?", assessmentId, courseId).
				Count(&count).Error; err != nil {
				return false
			}

			if count == 0 {
				if err := db.MasterDB.Create(&models.PublishAssessment{
					AssessmentId: assessmentId,
					CourseId:     courseId,
				}).Error; err != nil {
					return false
				}
			}

		} else {
			if err := db.MasterDB.
				Where("assessment_id = ? AND course_id = ?", assessmentId, courseId).
				Delete(&models.PublishAssessment{}).Error; err != nil {
				return false
			}
		}

		return true
	}

	if err := db.MasterDB.
		Where("assessment_id = ?", assessmentId).
		Delete(&models.PublishAssessment{}).Error; err != nil {
		return false
	}

	if len(courseIds) == 0 {
		return true
	}

	publishes := make([]models.PublishAssessment, 0, len(courseIds))
	for _, cid := range courseIds {
		publishes = append(publishes, models.PublishAssessment{
			AssessmentId: assessmentId,
			CourseId:     cid,
		})
	}

	if err := db.MasterDB.Create(&publishes).Error; err != nil {
		return false
	}

	return true
}
