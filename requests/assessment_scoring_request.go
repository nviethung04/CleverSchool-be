package requests

type AssessmentSaveScoreSubcriteriaRequest struct {
	SubcriteriaID int64   `json:"subcriteria_id"`
	Score         float64 `json:"score"`
}

type AssessmentSaveScoreCriteriaRequest struct {
	CriteriaID   int64                                  `json:"criteria_id"`
	Score        float64                                `json:"score"`
	Subcriteria  []AssessmentSaveScoreSubcriteriaRequest `json:"subcriteria"`
}

type AssessmentSaveScoreStudentRequest struct {
	UserID    int64                           `json:"user_id"`
	FileInfos []map[string]interface{}        `json:"file_infos"`
	Criteria  []AssessmentSaveScoreCriteriaRequest `json:"criteria"`
}

type AssessmentSaveScoreBulkRequest struct {
	AssessmentID int64                             `json:"assessment_id" binding:"required"`
	CourseID     int64                             `json:"course_id" binding:"required"`
	Students     []AssessmentSaveScoreStudentRequest `json:"students" binding:"required"`
}

type AssessmentPublishRequest struct {
	CourseID     int64 `json:"course_id" binding:"required"`
	AssessmentID int64 `json:"assessment_id" binding:"required"`
	Publish      bool  `json:"publish"`
}

type AssessmentStudentsRequest struct {
	CourseID     int64 `form:"course_id" binding:"required"`
	AssessmentID int64 `form:"assessment_id" binding:"required"`
	StudentID    *int64 `form:"student_id"`
	Page         int   `form:"page"`
	Limit        int   `form:"limit"`
	GetScore     bool  `form:"get_score"`
}

type StudyReportListRequest struct {
	SubjectID    *int64 `form:"subject_id"`
	CourseID     *int64 `form:"course_id"`
	AssessmentID *int64 `form:"assessment_id"`
	Page         int    `form:"page"`
	PerPage      int    `form:"per_page"`
}

type StudyReportEvaluatesRequest struct {
	AssessmentID *int64 `form:"assessment_id"`
	IsCompleted  *bool  `form:"is_completed"`
}

type StudyReportPublishRequest struct {
	CourseID     int64 `json:"course_id" binding:"required"`
	AssessmentID int64 `json:"assessment_id" binding:"required"`
	Publish      bool  `json:"publish"`
}

type StudyReportCreateRequest struct {
	StudyReportCriteriaID int64                   `json:"study_report_criteria_id" binding:"required"`
	SubjectID             int64                   `json:"subject_id"`
	AssessmentID          int64                   `json:"assessment_id" binding:"required"`
	StudentID             int64                   `json:"student_id" binding:"required"`
	TeacherID             int64                   `json:"teacher_id"`
	CourseID              int64                   `json:"course_id" binding:"required"`
	GeneralComment        string                  `json:"general_comment"`
	IsCompleted           bool                    `json:"is_completed"`
	StarSkills            []map[string]interface{} `json:"star_skills"`
	CheckSkills           []map[string]interface{} `json:"check_skills"`
}

type AssessmentSubmitRequest struct {
	AssessmentID int64                    `json:"assessment_id" binding:"required"`
	CourseID     int64                    `json:"course_id" binding:"required"`
	FileInfos    []map[string]interface{} `json:"file_infos"`
}

type CreateGroupWithCriteriaRequest struct {
	Group struct {
		Name                              string `json:"name"`
		SubjectID                         int64  `json:"subject_id"`
		HasFile                           bool   `json:"has_file"`
		CopyFromAssessmentCriteriaGroupID int64  `json:"copy_from_assessment_criteria_group_id"`
	} `json:"group"`
	Criteria []struct {
		ID           interface{} `json:"id"`
		Name         string      `json:"name"`
		Description  string      `json:"description"`
		MaxScore     float64     `json:"max_score"`
		Subcriteria  []struct {
			ID          interface{} `json:"id"`
			Name        string      `json:"name"`
			Description string      `json:"description"`
			MaxScore    float64     `json:"max_score"`
		} `json:"subcriteria"`
	} `json:"criteria"`
}

type StudyReportCriteriaCreateRequest struct {
	SubjectID   int64                    `json:"subject_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	MaxStar     int                      `json:"max_star"`
	Notes       []map[string]interface{} `json:"notes"`
	StarSkills  []map[string]interface{} `json:"star_skills"`
	CheckSkills []map[string]interface{} `json:"check_skills"`
}
