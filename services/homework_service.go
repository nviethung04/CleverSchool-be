package services

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/repositories/base"
	"be-lms/requests"
	"be-lms/resources"
	"be-lms/utils"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type HomeworkService interface {
	GetAll(c *gin.Context) ([]models.Homework, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Homework, error)
	Create(c *gin.Context, req *prot.HomeworkRequest) (*models.Homework, error)
	Update(c *gin.Context, req *prot.HomeworkRequest) (*models.Homework, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Homework, error)
	Cloned(c *gin.Context, id int) (*prot.Homework, error)
	Assigned(c *gin.Context, id int) (*prot.AssignedHomework, error)
	AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error)
}

type homeworkService struct {
	repo repositories.HomeworkRepository
}

func NewHomeworkService(r repositories.HomeworkRepository) HomeworkService {
	return &homeworkService{repo: r}
}

func (s *homeworkService) GetAll(c *gin.Context) ([]models.Homework, int64, error) {
	var req requests.GetHomeworkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	homeworks, total, err := s.repo.GetAllWithPaging(&req, c)
	if err != nil {
		return nil, 0, err
	}

	return homeworks, total, nil
}

func (s *homeworkService) GetByID(c *gin.Context, id int) (*prot.Homework, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	// Lấy role_id từ token
	roleID := utils.GetCurrentRoleId(c)

	// Nếu là học sinh (role_id = 3), kiểm tra quyền truy cập
	if roleID == 3 {
		hasAccess, err := s.checkStudentHomeworkAccess(userID, int64(id))
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			return nil, fmt.Errorf("don't have permission to access this homework")
		}
	}

	hw, err := s.repo.GetByID(int64(id), userID, c)
	if err != nil {
		return nil, err
	}
	return dtoToProtoHomework(hw), nil
}

// checkStudentHomeworkAccess kiểm tra quyền truy cập homework của học sinh
func (s *homeworkService) checkStudentHomeworkAccess(userID, homeworkID int64) (bool, error) {
	// Sử dụng SQL JOIN để kiểm tra quyền truy cập trong 1 query
	var count int64
	err := db.ReplicaDB.Table("user_courses uc").
		Joins("JOIN homework_ref_lessons hrl ON uc.course_id = hrl.course_id").
		Where("uc.user_id = ? AND hrl.homework_id = ? AND hrl.course_id IS NOT NULL", userID, homeworkID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	// Nếu có ít nhất 1 course_id chung, học sinh được phép truy cập
	return count > 0, nil
}

func (s *homeworkService) Create(c *gin.Context, req *prot.HomeworkRequest) (*models.Homework, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	homeworkResource := resources.NewHomeworkResource()
	homework := homeworkResource.FormatModelHomework(req)
	homework.CreatedBy = userID

	err = s.repo.Create(homework)
	if err != nil {
		return nil, err
	}

	return homework, nil
}

func (s *homeworkService) Update(c *gin.Context, req *prot.HomeworkRequest) (*models.Homework, error) {
	tokenStr := c.GetHeader("Token")
	userID, err := utils.GetUserID(tokenStr)
	if err != nil {
		utils.Respond(c, nil, err, "")
		return nil, nil
	}

	homeworkResource := resources.NewHomeworkResource()
	homework := homeworkResource.FormatModelHomework(req)
	homework.UpdatedBy = userID

	err = s.repo.Update(homework)
	if err != nil {
		return nil, err
	}

	return homework, nil
}

func (s *homeworkService) Delete(c *gin.Context, id int) error {
	deletedBy := utils.GetCurrentUserId(c)
	return s.repo.Delete(int64(id), int64(deletedBy))
}

func (s *homeworkService) Restore(c *gin.Context, id int) (*models.Homework, error) {
	baseRepo := base.NewBaseRepository[*models.Homework]()
	baseRepo.SetContext(c)
	homework, err := baseRepo.Restore(id)
	if err != nil {
		return nil, err
	}

	return *homework, nil
}

func (s *homeworkService) Cloned(c *gin.Context, id int) (*prot.Homework, error) {
	data, err := s.repo.GetByID(int64(id), 0, c)

	if err != nil {
		return nil, err
	}

	var req = &prot.ClonedHomeworkRequest{}

	req, err, _ = utils.GetBody[*prot.ClonedHomeworkRequest](c, func() *prot.ClonedHomeworkRequest {
		return &prot.ClonedHomeworkRequest{}
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

	name := data.Name
	if req.Name != "" {
		name = req.Name
	}

	description := data.Description
	if req.Description != "" {
		description = req.Description
	}

	homework := &models.Homework{
		Name:           name,
		Status:         data.Status,
		Description:    description,
		MaxScore:       data.MaxScore,
		CoverImageInfo: data.CoverImageInfo,
		TotalQuestions: data.TotalQuestion,
		CreatedAt:      now,
		CreatedBy:      int64(createById),
		ObjectTitle:    data.ObjectTitle,
		ProgramId:      data.ProgramID,
		CloneInfo: &models.CloneInfo{
			CloneId:   int64(id),
			CourseId:  cloneCourseId,
			ClonedAt:  &now,
			ProgramId: data.ProgramID,
		},
		QuestionForm: data.QuestionForm,
		FileInfos: data.FileInfos,
	}

	err = s.repo.Create(homework)

	if err != nil {
		return nil, err
	}

	clonedQuestionRepo := repositories.NewClonedQuestionRepository()
	clonedQuestionRepo.SetContext(c)
	clonedQuestion, err := clonedQuestionRepo.FindByAssignment(int64(id), models.ClonedQuestionTypeHomework)

	if err != nil {
		return nil, err
	}

	if clonedQuestion != nil {
		newClonedQuestion := clonedQuestion
		newClonedQuestion.ID = int64(0)
		newClonedQuestion.AssignmentID = homework.ID
		newClonedQuestion.CreatedBy = int64(createById)
		newClonedQuestion.CreatedAt = now
		newClonedQuestion.CloneInfo = &models.CloneInfo{
			CloneId:   int64(clonedQuestion.ID),
			CourseId:  cloneCourseId,
			ClonedAt:  &now,
			ProgramId: data.ProgramID,
		}

		err := clonedQuestionRepo.Create(newClonedQuestion)
		if err != nil {
			return nil, err
		}
	}

	UpdateHomeworkTotalQuestions(homework.ID)

	newHomework, err := s.repo.GetByID(homework.ID, 0, c)

	if err != nil {
		return nil, err
	}

	if req.LessonId != 0 {
		lessonRepo := repositories.NewLessonRepository()
		lessonRepo.SetContext(c)
		err := lessonRepo.CreateHomeworkRefLesson(req.LessonId, newHomework.ID)
		if err != nil {
			return nil, err
		}
	}

	newModelHomework := dtoToProtoHomework(newHomework)
	return newModelHomework, nil
}

func (s *homeworkService) Assigned(c *gin.Context, id int) (*prot.AssignedHomework, error) {
	var req = &prot.AssignedHomework{}

	req, err, _ := utils.GetBody[*prot.AssignedHomework](c, func() *prot.AssignedHomework {
		return &prot.AssignedHomework{}
	})

	if err != nil {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	if req.CourseId == 0 || req.LessonId == 0 {
		return nil, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	ref := models.HomeworkRefLesson{
		HomeworkId: int64(id),
		LessonId:   req.LessonId,
		CourseId:   req.CourseId,
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

func (s *homeworkService) AssignedLesson(c *gin.Context, id int) (*prot.AssignedLessons, error) {
	homework, err := s.repo.AssignedLesson(int64(id))
	if err != nil {
		return nil, err
	}

	var lessons []*prot.AssignedLesson
	for _, lesson := range homework.Lessons {
		var isAssigned bool

		for _, ref := range homework.HomeworkRefLessons {
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

func CountHomeworkQuestionsFromCloned(homeworkID int64) (int32, error) {
	cloneRepo := repositories.NewClonedQuestionRepository()

	cloneData, err := cloneRepo.FindByAssignment(homeworkID, models.ClonedQuestionTypeHomework)

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

func UpdateHomeworkTotalQuestions(homeworkID int64) error {
	totalQuestions, err := CountHomeworkQuestionsFromCloned(homeworkID)
	if err != nil {
		return err
	}
	return db.MasterDB.Model(&models.Homework{}).Where("id = ?", homeworkID).Update("total_questions", int32(totalQuestions)).Error
}

func dtoToProtoHomework(hw *dto.HomeworkDTO) *prot.Homework {
	// Skip check assigned by homework
	// isAssigned := utils.BoolOrFalse(hw.IsAssigned)
	isAssigned := false

	questionFiles := make([]*prot.HomeworkFile, 0, len(hw.FileInfos))

	for _, fileInfo := range hw.FileInfos {
		questionFiles = append(questionFiles, &prot.HomeworkFile{
			Type: fileInfo.Type,
			Url:  utils.StaticURL(fileInfo.Path, models.Storage),
		})
	}

	return &prot.Homework{
		Id:                      hw.ID,
		LessonId:                hw.LessonID,
		Name:                    hw.Name,
		Status:                  int32(hw.Status),
		Description:             hw.Description,
		MaxScore:                float32(hw.MaxScore),
		CoverImage:              utils.StaticURL(hw.CoverImageInfo.Disk, models.Storage),
		CreatedAt:               hw.CreatedAt.Unix(),
		CreatedBy:               hw.CreatedBy,
		UpdatedAt:               hw.UpdatedAt.Unix(),
		UpdatedBy:               hw.UpdatedBy,
		IsAssigned:              isAssigned,
		TotalQuestions:          hw.TotalQuestion,
		QuestionCompleted:       hw.QuestionCompleted,
		LastQuestionIdCompleted: hw.LastQuestionIDCompleted,
		ObjectTitle:             hw.ObjectTitle,
		IsRandomQuestion:        hw.IsRandomQuestion,
		QuestionForm: hw.QuestionForm,
		QuestionFiles: questionFiles,
	}
}
