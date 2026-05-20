package services

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/repositories/base"
	"be-cleverschool/requests"
	"be-cleverschool/resources"
	"be-cleverschool/utils"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExerciseService interface {
    GetAll(c *gin.Context) ([]models.Exercise, int64, error)
    GetByID(c *gin.Context, id int) (*prot.Exercise, error)
    Create(c *gin.Context, req *prot.ExerciseRequest) (*models.Exercise, error)
    Update(c *gin.Context, req *prot.ExerciseRequest) (*models.Exercise, error)
    Delete(c *gin.Context, id int) error
    Restore(c *gin.Context, id int) (*models.Exercise, error)
    Cloned(c *gin.Context, id int) (*prot.Exercise, error)
    Assigned(c *gin.Context, id int) (*prot.AssignedExercise, error)
    AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error)
}

type exerciseService struct {
    repo repositories.ExerciseRepository
}

func NewExerciseService(r repositories.ExerciseRepository) ExerciseService {
    return &exerciseService{repo: r}
}

func (s *exerciseService) GetAll(c *gin.Context) ([]models.Exercise, int64, error) {
    var req requests.GetExerciseRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        return nil, 0, err
    }

    exercises, total, err := s.repo.GetAllWithPaging(&req, c)
    if err != nil {
        return nil, 0, err
    }

    return exercises, total, err
}

func (s *exerciseService) GetByID(c *gin.Context, id int) (*prot.Exercise, error) {
    item, err := s.repo.GetByID(int64(id), c)
    if err != nil {
        return nil, err
    }
    res := resources.NewExerciseResource()
    return res.FormatExercise(item), nil
}

func (s *exerciseService) Create(c *gin.Context, req *prot.ExerciseRequest) (*models.Exercise, error) {
    res := resources.NewExerciseResource()
    model := res.FormatModelExercise(req)
    model.CreatedBy = int64(utils.GetCurrentUserId(c))
    err := s.repo.Create(model)
    if err != nil { return nil, err }

    // Ghi liên kết exercise_ref_lessons nếu có lesson_id trong query hoặc header (tuỳ đặc tả)
    // Hiện ExerciseRequest không có LessonId, nên tạm bỏ qua.
    return model, nil
}

func (s *exerciseService) Update(c *gin.Context, req *prot.ExerciseRequest) (*models.Exercise, error) {
    res := resources.NewExerciseResource()
    model := res.FormatModelExercise(req)

    model.UpdatedAt = time.Now().UTC()
    model.UpdatedBy = int64(utils.GetCurrentUserId(c))

    err := s.repo.Update(model)
    if err != nil { return nil, err }

    // Nếu có lesson_id (ví dụ từ query "lesson_id"), tạo liên kết vào exercise_ref_lessons
    if lessonIDStr := c.Query("lesson_id"); lessonIDStr != "" {
        if lessonID, convErr := strconv.ParseInt(lessonIDStr, 10, 64); convErr == nil {
            lessonRepo := repositories.NewLessonRepository()
            _ = lessonRepo.CreateExerciseRefLesson(lessonID, model.ID)
        }
    }

    return model, nil
}

func (s *exerciseService) Delete(c *gin.Context, id int) error {
    deletedBy := utils.GetCurrentUserId(c)
    return s.repo.Delete(int64(id), int64(deletedBy))
}

func (s *exerciseService) Restore(c *gin.Context, id int) (*models.Exercise, error) {
    baseRepo := base.NewBaseRepository[*models.Exercise]()
    baseRepo.SetContext(c)
    item, err := baseRepo.Restore(id)
    if err != nil { return nil, err }
    return *item, nil
}

func (s *exerciseService) Cloned(c *gin.Context, id int) (*prot.Exercise, error) {
    data, err := s.repo.GetByID(int64(id), c)
    if err != nil { return nil, err }

    var req = &prot.ClonedExamRequest{}
    req, err, _ = utils.GetBody[*prot.ClonedExamRequest](c, func() *prot.ClonedExamRequest { return &prot.ClonedExamRequest{} })
    // Cho phép body rỗng: nếu không parse được thì dùng request mặc định
    if err != nil {
        req = &prot.ClonedExamRequest{}
    }

    now := time.Now()
    var cloneCourseId int64
    if data.CloneInfo != nil { cloneCourseId = data.CloneInfo.CourseId }
    createById := utils.GetCurrentUserId(c)

    newItem := data
    newItem.ID = 0
    newItem.CreatedAt = now
    newItem.CreatedBy = int64(createById)
    newItem.CloneInfo = &models.CloneInfo{ CloneId: int64(data.ID), CourseId: cloneCourseId, ClonedAt: &now, ProgramId: data.ProgramId }
    if req.Description != "" { newItem.Description = req.Description }
    if req.Name != "" { newItem.Name = req.Name }

    err = s.repo.Create(newItem)
    if err != nil { return nil, err }

    clonedQuestionRepo := repositories.NewClonedQuestionRepository()
    clonedQuestionRepo.SetContext(c)
    clonedQuestion, err := clonedQuestionRepo.FindByAssignment(int64(id), models.ClonedQuestionTypeExercise)
    if err != nil { return nil, err }
    if clonedQuestion != nil {
        newClonedQuestion := clonedQuestion
        newClonedQuestion.ID = int64(0)
        newClonedQuestion.AssignmentID = newItem.ID
        newClonedQuestion.CreatedBy = int64(createById)
        newClonedQuestion.CreatedAt = now
        newClonedQuestion.CloneInfo = &models.CloneInfo{ CloneId: int64(clonedQuestion.ID), CourseId: cloneCourseId, ClonedAt: &now, ProgramId: data.ProgramId }
        if err := clonedQuestionRepo.Create(newClonedQuestion); err != nil { return nil, err }
    }

    UpdateExerciseTotalQuestions(newItem.ID)

    if err != nil { return nil, err }

    if req.LessonId != 0 {
        lessonRepo := repositories.NewLessonRepository()
        lessonRepo.SetContext(c)
        err := lessonRepo.CreateExerciseRefLesson(req.LessonId, newItem.ID)
        if err != nil { return nil, err }
    }

    resrc := resources.NewExerciseResource()
    protItem := resrc.FormatExercise(newItem)
    return protItem, nil
}

func (s *exerciseService) Assigned(c *gin.Context, id int) (*prot.AssignedExercise, error) {
	var req = &prot.AssignedExercise{}

	req, err, _ := utils.GetBody[*prot.AssignedExercise](c, func() *prot.AssignedExercise {
		return &prot.AssignedExercise{}
	})

	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	if req.CourseId == 0 || req.LessonId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	ref := models.ExerciseRefLesson{
		ExerciseId:   int64(id),
		LessonId: req.LessonId,
		CourseId: req.CourseId,
	}

	if req.IsAssigned {
		now := time.Now()
		ref.AssignedAt = &now
		ref.AssignedBy = utils.PtrInt64(int64(utils.GetCurrentUserId(c)))
	} else {
		ref.AssignedAt = nil
		ref.AssignedBy = nil
	}

	err = s.repo.Assigned(ref)

	if err != nil {
		return req, fmt.Errorf(i18n.Localize("messages.no_record_update"))
	}

	return req, nil
}

func (s *exerciseService) AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error) {
    exercise, err := s.repo.AssignedLesson(int64(id))
	if err != nil {
		return nil, err
	}

	var lessons []*prot.AssignedLesson
	for _, lesson := range exercise.Lessons {
		var isAssigned bool

		for _, ref := range exercise.ExerciseRefLessons {
			if ref.LessonId == lesson.ID {
				isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
			}
			if isAssigned {
				break
			}
		}
		lessons = append(lessons, &prot.AssignedLesson{
			Id: lesson.ID,
			Title: lesson.Title,
			Description: lesson.Description,
			ObjectTitle: lesson.ObjectTitle,
			IsAssigned: isAssigned,
		})
	}

	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].IsAssigned != lessons[j].IsAssigned {
			return lessons[i].IsAssigned && !lessons[j].IsAssigned
		}
		return lessons[i].Id < lessons[j].Id
	})
    return &prot.AssignedLessons{Lessons: lessons}, nil
}

func CountExerciseQuestionsFromCloned(exerciseID int64) (int32, error) {
    cloneRepo := repositories.NewClonedQuestionRepository()
    cloneData, err := cloneRepo.FindByAssignment(exerciseID, models.ClonedQuestionTypeExercise)
    if err != nil { return 0, err }
    var rawQuestions []json.RawMessage
    if err := json.Unmarshal(cloneData.Questions, &rawQuestions); err != nil {
        config.Log.Error("unmarshal rawQuestions error:", err.Error())
        return 0, err
    }
    return int32(len(rawQuestions)), nil
}

func UpdateExerciseTotalQuestions(exerciseID int64) error {
    totalQuestions, err := CountExerciseQuestionsFromCloned(exerciseID)
    if err != nil { return err }
    return db.MasterDB.Model(&models.Exercise{}).Where("id = ?", exerciseID).Update("total_questions", int32(totalQuestions)).Error
}



