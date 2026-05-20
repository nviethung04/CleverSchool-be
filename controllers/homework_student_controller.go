package controllers

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HomeworkStudentController struct {
	service services.HomeworkStudentService
}

func NewHomeworkStudentController(service services.HomeworkStudentService) *HomeworkStudentController {
	return &HomeworkStudentController{service: service}
}

func (ctl *HomeworkStudentController) GetHomeworkStudents(c *gin.Context) {
	homeworkIDStr := c.Query("homework_id")
	courseIDStr := c.Query("course_id")
	if homeworkIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "homework_id is required"})
		return
	}
	homeworkID, err := strconv.ParseInt(homeworkIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid homework_id"})
		return
	}
	var courseID int64
	if courseIDStr != "" {
		courseID, err = strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course_id"})
			return
		}
	}
	info, students, err := ctl.service.GetHomeworkStudents(homeworkID, courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var questionFiles []*prot.HomeworkFile

	for _, qf := range info.FileInfos {
		questionFiles = append(questionFiles, &prot.HomeworkFile{
			Type: qf.Type,
			Url:  utils.StaticURL(qf.Disk, models.Storage),
		})
	}

	resp := &prot.HomeworkStudentsResponse{
		HomeworkInfo: &prot.HomeworkStudentsResponse_HomeworkInfo{
			Name:           info.Name,
			Description:    info.Description,
			Status:         info.Status,
			CoverImage:     utils.StaticURL(info.CoverImage, models.Storage),
			IsAssigned:     info.IsAssigned,
			TotalQuestions: info.TotalQuestions,
			QuestionForm: info.QuestionForm,
			QuestionFiles: questionFiles,
			CreatedAt:      info.CreatedAt.Unix(),
		},
		Students: []*prot.HomeworkStudentInfo{},
	}
	for _, s := range students {
		resp.Students = append(resp.Students, &prot.HomeworkStudentInfo{
			UserId:                  s.UserID,
			Name:                    s.Name,
			Username:                s.Username,
			Avatar:                  utils.StaticURL(s.Avatar, models.Storage),
			QuestionsCompleted:      s.QuestionsCompleted,
			LastQuestionIdCompleted: s.LastQuestionIDCompleted,
			UpdatedAt:               s.UpdatedAt.Unix(),
			ManualQuestionsCount:    s.ManualQuestionsCount,
			CourseId:                s.CourseID,
			IsSubmitted:             s.IsSubmitted,
		})
	}
	utils.Respond(c, resp, err, "")
}
