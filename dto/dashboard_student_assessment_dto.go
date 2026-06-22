package dto

type DashboardStudentAssessmentItemDTO struct {
	Assessment DashboardStudentAssessmentDetailDTO `json:"assessment"`
	IsScored   bool                                `json:"is_scored"`
	TotalScore float64                             `json:"total_score"`
	Files      []string                            `json:"files"`
}

type DashboardStudentAssessmentDetailDTO struct {
	ID                           string                                    `json:"id"`
	Name                         string                                    `json:"name"`
	Description                  string                                    `json:"description"`
	Type                         string                                    `json:"type"`
	ProgramID                    string                                    `json:"program_id"`
	CreatedAt                    string                                    `json:"created_at"`
	CreatedBy                    string                                    `json:"created_by"`
	UpdatedAt                    string                                    `json:"updated_at"`
	UpdatedBy                    string                                    `json:"updated_by"`
	FileInfos                    []DashboardStudentAssessmentFileInfoDTO   `json:"file_infos"`
	Files                        []string                                  `json:"files"`
	StudyReportCriteriaID        string                                    `json:"study_report_criteria_id"`
	SubjectID                    string                                    `json:"subject_id"`
	AssessmentCriteriaGroupID    string                                    `json:"assessment_criteria_group_id"`
	AssessmentCriteriaGroupName  string                                    `json:"assessment_criteria_group_name"`
	StudyReportCriteriaName      string                                    `json:"study_report_criteria_name"`
	SubjectName                  string                                    `json:"subject_name"`
	Criteria                     []DashboardStudentAssessmentCriterionDTO `json:"criteria"`
	LessonID                     string                                    `json:"lesson_id,omitempty"`
	LessonTitle                  string                                    `json:"lesson_title,omitempty"`
	IsAssigned                   bool                                      `json:"is_assigned,omitempty"`
	HasFile                      bool                                      `json:"has_file"`
}

type DashboardStudentAssessmentFileInfoDTO struct {
	ID   int64  `json:"id"`
	Disk string `json:"disk"`
	Path string `json:"path"`
}

type DashboardStudentAssessmentCriterionDTO struct {
	ID            string                                       `json:"id"`
	SubjectID     string                                       `json:"subject_id"`
	Name          string                                       `json:"name"`
	Description   string                                       `json:"description"`
	HasSubcriteria bool                                        `json:"has_subcriteria"`
	MaxScore      float64                                      `json:"max_score"`
	CreatedAt     string                                       `json:"created_at"`
	CreatedBy     string                                       `json:"created_by"`
	UpdatedAt     string                                       `json:"updated_at"`
	UpdatedBy     string                                       `json:"updated_by"`
	Subcriteria   []DashboardStudentAssessmentSubcriterionDTO `json:"subcriteria"`
	Score         float64                                      `json:"score"`
}

type DashboardStudentAssessmentSubcriterionDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MaxScore    float64 `json:"max_score"`
	Score       float64 `json:"score"`
}

type DashboardStudentAssessmentListResponseDTO struct {
	Type        string                              `json:"@type"`
	Assessments []DashboardStudentAssessmentItemDTO `json:"assessments"`
	Total       string                              `json:"total"`
}
