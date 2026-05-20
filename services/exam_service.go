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
	"time"

	"github.com/gin-gonic/gin"
)

type ExamService interface {
	GetAll(c *gin.Context) ([]models.Exam, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Exam, error)
	Create(c *gin.Context, req *prot.ExamRequest) (*models.Exam, error)
	Update(c *gin.Context, req *prot.ExamRequest) (*models.Exam, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Exam, error)
	Cloned(c *gin.Context, id int) (*prot.Exam, error)
	Assigned(c *gin.Context, id int) (*prot.AssignedExam, error)
	AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error)
}

type examService struct {
	repo repositories.ExamRepository
}

func NewExamService(r repositories.ExamRepository) ExamService {
	return &examService{repo: r}
}

func (s *examService) GetAll(c *gin.Context) ([]models.Exam, int64, error) {
	var req requests.GetExamRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	exams, total, err := s.repo.GetAllWithPaging(&req, c)
	if err != nil {
		return nil, 0, err
	}

	return exams, total, err
}

func (s *examService) GetByID(c *gin.Context, id int) (*prot.Exam, error) {
	exam, err := s.repo.GetByID(int64(id), c)
	if err != nil {
		return nil, err
	}
	return modelToProtoExam(exam), nil
}

func (s *examService) Create(c *gin.Context, req *prot.ExamRequest) (*models.Exam, error) {
	examResource := resources.NewExamResource()
	exam := examResource.FormatModelExam(req)
	exam.CreatedBy = int64(utils.GetCurrentUserId(c))

	err := s.repo.Create(exam)
	if err != nil {
		return nil, err
	}
	return exam, nil
}

func (s *examService) Update(c *gin.Context, req *prot.ExamRequest) (*models.Exam, error) {
	examResource := resources.NewExamResource()
	exam := examResource.FormatModelExam(req)

	exam.UpdatedAt = time.Now().UTC()
	exam.UpdatedBy = int64(utils.GetCurrentUserId(c))

	err := s.repo.Update(exam)
	if err != nil {
		return nil, err
	}
	return exam, nil
}

func (s *examService) Delete(c *gin.Context, id int) error {
	deletedBy := utils.GetCurrentUserId(c)
	return s.repo.Delete(int64(id), int64(deletedBy))
}

func (s *examService) Restore(c *gin.Context, id int) (*models.Exam, error) {
	baseRepo := base.NewBaseRepository[*models.Exam]()
	baseRepo.SetContext(c)
	exam, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *exam, nil
}

func (s *examService) Cloned(c *gin.Context, id int) (*prot.Exam, error) {
	data, err := s.repo.GetByID(int64(id), c)

	if err != nil {
		return nil, err
	}

	var req = &prot.ClonedExamRequest{}

	req, err, _ = utils.GetBody[*prot.ClonedExamRequest](c, func() *prot.ClonedExamRequest {
		return &prot.ClonedExamRequest{}
	})

	if err != nil {
		return nil, err
	}

	now := time.Now()
	var cloneCourseId int64

	if data.CloneInfo != nil {
		cloneCourseId = data.CloneInfo.CourseId
	}

	createById := utils.GetCurrentUserId(c)

	newExam := data
	newExam.ID = 0
	newExam.CreatedAt = now
	newExam.CreatedBy = int64(createById)
	newExam.CloneInfo = &models.CloneInfo{
		CloneId:   int64(data.ID),
		CourseId:  cloneCourseId,
		ClonedAt:  &now,
		ProgramId: data.ProgramId,
	}

	if req.Description != "" {
		newExam.Description = req.Description
	}

	if req.Name != "" {
		newExam.Name = req.Name
	}

	err = s.repo.Create(newExam)

	if err != nil {
		return nil, err
	}

	clonedQuestionRepo := repositories.NewClonedQuestionRepository()
	clonedQuestionRepo.SetContext(c)
	clonedQuestion, err := clonedQuestionRepo.FindByAssignment(int64(id), models.ClonedQuestionTypeExam)

	if err != nil {
		return nil, err
	}

	if clonedQuestion != nil {
		newClonedQuestion := clonedQuestion
		newClonedQuestion.ID = int64(0)
		newClonedQuestion.AssignmentID = newExam.ID
		newClonedQuestion.CreatedBy = int64(createById)
		newClonedQuestion.CreatedAt = now
		newClonedQuestion.CloneInfo = &models.CloneInfo{
			CloneId:   int64(clonedQuestion.ID),
			CourseId:  cloneCourseId,
			ClonedAt:  &now,
			ProgramId: data.ProgramId,
		}

		err := clonedQuestionRepo.Create(newClonedQuestion)
		if err != nil {
			return nil, err
		}
	}

	UpdateExamTotalQuestions(newExam.ID)

	newUpdateExam, err := s.repo.GetByID(newExam.ID, c)

	if err != nil {
		return nil, err
	}

	if req.LessonId != 0 {
		lessonRepo := repositories.NewLessonRepository()
		lessonRepo.SetContext(c)
		err := lessonRepo.CreateExamRefLesson(req.LessonId, newUpdateExam.ID)
		if err != nil {
			return nil, err
		}
	}

	resources := resources.NewExamResource()
	protExam := resources.FormatExam(newUpdateExam)

	return protExam, nil
}

func (s *examService) Assigned(c *gin.Context, id int) (*prot.AssignedExam, error) {
	var req = &prot.AssignedExam{}

	req, err, _ := utils.GetBody[*prot.AssignedExam](c, func() *prot.AssignedExam {
		return &prot.AssignedExam{}
	})

	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	if req.CourseId == 0 || req.LessonId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	ref := models.ExamRefLesson{
		ExamId:   int64(id),
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
func (s *examService) AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error) {
	exam, err := s.repo.AssignedLesson(int64(id))
	if err != nil {
		return nil, err
	}

	var lessons []*prot.AssignedLesson
	for _, lesson := range exam.Lessons {
		var isAssigned bool

		for _, ref := range exam.ExamRefLessons {
			if ref.LessonId == lesson.ID {
				isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
			}
			if isAssigned {
				break
			}
		}
		lessons = append(lessons, &prot.AssignedLesson{
			Id:          lesson.ID,
			Title:       lesson.Title,
			Description: lesson.Description,
			ObjectTitle: lesson.ObjectTitle,
			IsAssigned:  isAssigned,
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

func modelToProtoExam(exam *models.Exam) *prot.Exam {
	var isAssigned bool

	for _, ref := range exam.ExamRefLessons {
		if ref.ExamId == exam.ID {
			isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
		}
	}

	questionFiles := make([]*prot.ExamFile, 0, len(exam.FileInfos))

	for _, fileInfo := range exam.FileInfos {
		questionFiles = append(questionFiles, &prot.ExamFile{
			Type: fileInfo.Type,
			Url:  utils.StaticURL(fileInfo.Path, models.Storage),
		})
	}

	return &prot.Exam{
		Id:               exam.ID,
		Name:             exam.Name,
		Status:           int32(exam.Status),
		TimeLimit:        exam.TimeLimit,
		MaxScore:         float32(exam.MaxScore),
		Description:      exam.Description,
		CoverImage:       utils.StaticURL(exam.CoverImageInfo.Path, models.Storage),
		CreatedAt:        exam.CreatedAt.Unix(),
		CreatedBy:        exam.CreatedBy,
		UpdatedAt:        exam.UpdatedAt.Unix(),
		UpdatedBy:        exam.UpdatedBy,
		Deadline:         exam.Deadline.Unix(),
		IsAssigned:       isAssigned,
		TotalQuestions:   int32(exam.TotalQuestions),
		IsRandomQuestion: exam.IsRandomQuestion,
		ObjectTitle:      exam.ObjectTitle,
		QuestionForm:     exam.QuestionForm,
		QuestionFiles:    questionFiles,
		Type:             exam.Type,
	}
}

// CountExamQuestionsFromCloned đếm số câu hỏi của exam từ bảng cloned_questions
func CountExamQuestionsFromCloned(examID int64) (int32, error) {
	cloneRepo := repositories.NewClonedQuestionRepository()
	cloneData, err := cloneRepo.FindByAssignment(examID, models.ClonedQuestionTypeExam)

	if err != nil {
		return 0, err
	}

	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(cloneData.Questions, &rawQuestions); err != nil {
		config.Log.Error("unmarshal rawQuestions error:", err.Error())
		return 0, err
	}

	return int32(len(rawQuestions)), nil
}

func UpdateExamTotalQuestions(examID int64) error {
	totalQuestions, err := CountExamQuestionsFromCloned(examID)
	if err != nil {
		return err
	}
	return db.MasterDB.Model(&models.Exam{}).Where("id = ?", examID).Update("total_questions", int32(totalQuestions)).Error
}

