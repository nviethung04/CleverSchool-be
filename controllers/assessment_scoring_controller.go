package controllers

import (
	"be-lms/requests"
	"be-lms/services"
	"be-lms/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AssessmentScoringController struct {
	svc services.AssessmentScoringService
}

func NewAssessmentScoringController(svc services.AssessmentScoringService) *AssessmentScoringController {
	return &AssessmentScoringController{svc: svc}
}

func (ctl *AssessmentScoringController) SaveScoresBulk(c *gin.Context) {
	var req requests.AssessmentSaveScoreBulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	data, err := ctl.svc.SaveScoresBulk(&req, actorID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetTeacherStudents(c *gin.Context) {
	var req requests.AssessmentStudentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := ctl.svc.GetTeacherStudents(&req)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetPublish(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Query("course_id"), 10, 64)
	assessmentID, _ := strconv.ParseInt(c.Query("assessment_id"), 10, 64)
	data, err := ctl.svc.GetPublish(courseID, assessmentID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) SetPublish(c *gin.Context) {
	var req requests.AssessmentPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := ctl.svc.SetPublish(&req)
	utils.Respond(c, map[string]bool{"success": err == nil}, err, "")
}

func (ctl *AssessmentScoringController) StudentSubmit(c *gin.Context) {
	var req requests.AssessmentSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := utils.GetUserID(c.GetHeader("Token"))
	if err != nil {
		utils.Respond(c, nil, err, "")
		return
	}
	err = ctl.svc.StudentSubmit(userID, &req)
	utils.Respond(c, map[string]string{"message": "submitted"}, err, "")
}

func (ctl *AssessmentScoringController) CreateCriteriaGroup(c *gin.Context) {
	var req requests.CreateGroupWithCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	data, err := ctl.svc.CreateCriteriaGroup(&req, actorID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) UpdateCriteriaGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req requests.CreateGroupWithCriteriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	data, err := ctl.svc.UpdateCriteriaGroup(id, &req, actorID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) ListCriteriaGroups(c *gin.Context) {
	var subjectID *int64
	if v := c.Query("subject_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		subjectID = &id
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	data, err := ctl.svc.ListCriteriaGroups(subjectID, c.Query("keyword"), page, limit)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetCriteriaGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	data, err := ctl.svc.GetCriteriaGroup(id)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) DeleteCriteriaGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	err := ctl.svc.DeleteCriteriaGroup(id, actorID)
	utils.Respond(c, nil, err, "Delete successful")
}

func (ctl *AssessmentScoringController) ListStudyReportCriterias(c *gin.Context) {
	var subjectID *int64
	if v := c.Query("subject_id"); v != "" {
		id, _ := strconv.ParseInt(v, 10, 64)
		subjectID = &id
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	data, err := ctl.svc.ListStudyReportCriterias(subjectID, c.Query("keyword"), page, limit)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetStudyReportCriteria(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	data, err := ctl.svc.GetStudyReportCriteria(id)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) CreateStudyReportCriteria(c *gin.Context) {
	var req requests.StudyReportCriteriaCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := ctl.svc.CreateStudyReportCriteria(&req)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) UpdateStudyReportCriteria(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req requests.StudyReportCriteriaCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := ctl.svc.UpdateStudyReportCriteria(id, &req)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) DeleteStudyReportCriteria(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := ctl.svc.DeleteStudyReportCriteria(id)
	utils.Respond(c, nil, err, "Delete successful")
}

func (ctl *AssessmentScoringController) ListStudyReports(c *gin.Context) {
	var req requests.StudyReportListRequest
	_ = c.ShouldBindQuery(&req)
	userID, _ := utils.GetUserID(c.GetHeader("Token"))
	isStudent := isStudentOnly(c)
	data, err := ctl.svc.ListStudyReports(&req, userID, isStudent)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetStudyReportDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := utils.GetUserID(c.GetHeader("Token"))
	isStudent := isStudentOnly(c)
	data, err := ctl.svc.GetStudyReportDetail(id, userID, isStudent)
	utils.Respond(c, data, err, "")
}

func isStudentOnly(c *gin.Context) bool {
	if roleIDsVal, ok := c.Get("roleIDs"); ok {
		if ids, ok := roleIDsVal.([]int); ok {
			if len(ids) == 1 && ids[0] == 3 {
				return true
			}
			for _, id := range ids {
				if id != 3 {
					return false
				}
			}
			return len(ids) > 0
		}
	}
	if roleType, ok := c.Get("roleType"); ok {
		if s, ok := roleType.(string); ok && s == "student" {
			return true
		}
	}
	return false
}

func (ctl *AssessmentScoringController) CreateStudyReport(c *gin.Context) {
	var req requests.StudyReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	data, err := ctl.svc.CreateStudyReport(&req, actorID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) UpdateStudyReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req requests.StudyReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorID, _ := utils.GetUserID(c.GetHeader("Token"))
	data, err := ctl.svc.UpdateStudyReport(id, &req, actorID)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) GetStudyReportEvaluates(c *gin.Context) {
	courseID, _ := strconv.ParseInt(c.Param("course_id"), 10, 64)
	var req requests.StudyReportEvaluatesRequest
	_ = c.ShouldBindQuery(&req)
	data, err := ctl.svc.GetStudyReportEvaluates(courseID, &req)
	utils.Respond(c, data, err, "")
}

func (ctl *AssessmentScoringController) SetStudyReportPublish(c *gin.Context) {
	var req requests.StudyReportPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := ctl.svc.SetStudyReportPublish(&req)
	utils.Respond(c, map[string]bool{"success": err == nil}, err, "")
}
