package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExamCourseController struct {
	service services.ExamCourseService
}

func NewExamCourseController(service services.ExamCourseService) *ExamCourseController {
	return &ExamCourseController{service: service}
}

func (ctl *ExamCourseController) GetExamCourseDetail(c *gin.Context) {
    req, err, message := utils.GetBody[*prot.ExamCourseRequest](c, func() *prot.ExamCourseRequest {
		return &prot.ExamCourseRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	isAssigned := req.IsAssigned
	data, err := ctl.service.GetExamCourseDetail(req.CourseId, &isAssigned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := &prot.ExamCourseResponse{
		CourseInfo: &prot.ExamCourseInfo{
			Name:        data.CourseInfo.Name,
			Description: data.CourseInfo.Description,
			Status:      data.CourseInfo.Status,
		},
		Chapters: []*prot.ExamCourseChapter{},
	}
	for _, ch := range data.Chapters {
		chapter := &prot.ExamCourseChapter{
			ChapterInfo: &prot.ExamCourseChapterInfo{
				Title:       ch.ChapterInfo.Title,
				Description: ch.ChapterInfo.Description,
				Status:      ch.ChapterInfo.Status,
			},
			Exams: []*prot.ExamCourseExamInfo{},
		}
		for _, ex := range ch.Exams {
			exam := &prot.ExamCourseExamInfo{
				Id:                ex.ID,
				Name:              ex.Name,
				LessonTitle:       ex.LessonTitle,
				Description:       ex.Description,
				CoverImage:        ex.CoverImage,
				Deadline:          ex.Deadline,
				Status:            ex.Status,
				IsAssigned:        ex.IsAssigned,
				StudentsDone:      ex.StudentsDone,
				StudentsNotScored: ex.StudentsNotScored,
			}
			chapter.Exams = append(chapter.Exams, exam)
		}
		resp.Chapters = append(resp.Chapters, chapter)
	}
	utils.Respond(c, resp, err, "")
}
