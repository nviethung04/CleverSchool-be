package requests

type GradeRequest struct {
	GradeID      int64 `json:"grade_id" binding:"required"`
	AssessmentID int64 `json:"assessment_id" binding:"required"`
}

type SubjectRequest struct {
	SubjectID int64          `json:"subject_id" binding:"required"`
	Grades    []GradeRequest `json:"grades" binding:"required"`
}

type DashboardAssessmentReportExcelRequest struct {
	SchoolIDs string           `json:"school_ids" binding:"required"`
	Subjects  []SubjectRequest `json:"subjects" binding:"required"`
}
