package dto

import (
	"be-lms/models"
	"time"
)

type DashboardAdminItem struct {
	Total       int32
	TotalNumber int32
	ChangeValue float32
}

type UserOverview struct {
	User     *DashboardAdminItem
	Student  *DashboardAdminItem
	Teacher  *DashboardAdminItem
	Activity *DashboardAdminItem
	School   *DashboardAdminItem
}

type CourseOverviewDataItem struct {
	Programs          int32
	Lessons           int32
	LessonPlans       int32
	Questions         int32
	Homeworks         int32
	Exams             int32
	UserExams         int32
	LessonPlanCurrent int32
	HomeworkCurrent   int32
	ExamCurrent       int32
	UserExamCurrent   int32
	QuestionCurrent   int32
}

type LearningScoreWeek struct {
	Id         int64
	WeekNumber int32
	Score      float32
}

type GemsByWeek struct {
	Id         int64
	WeekNumber int32
	Count      int32
}

type StudentScore struct {
	Id         int64
	ExamId     int64
	Name       string
	ExamUserId int64
	ExamName   string
	Score      float32
	AvatarInfo models.MediaInfo `json:"avatar_info"`
	ClassName  string
	SchoolName string
}

type LearningOverviewDataItem struct {
	LearningScoreWeeks []LearningScoreWeek
	GemsByWeeks        []GemsByWeek
	TopHighest         []StudentScore
	TopLowest          []StudentScore
	TotalGem           int32
	TotalAverageScore  float32
	HighestPercent     float32
	LowestPercent      float32
	BehaviorPercent    float32
}

type ScoreDistributionItem struct {
	Number int32
	Scores []float32
	Counts []float32
}

type ScoreDistribution struct {
	ScoreRange   string `json:"score_range"`
	StudentCount int64  `json:"student_count"`
}

type ExamScoreChart struct {
	ScoreRange   string `json:"score_range"`
	StudentCount int64  `json:"student_count"`
}

type ScoreDistributionOverview struct {
	Items []ScoreDistributionItem
}

type WeeklyUsage struct {
	Week     string
	Duration float32
}

type DeviceUsage struct {
	Name    string
	Count   int32
	Percent float32
}

type SystemUsageOverview struct {
	WeeklyUsages []WeeklyUsage
	DeviceUsages []DeviceUsage
	AverageUsed  models.AverageUsed
}

type QuestionDistributionItem struct {
	Name    string
	Count   int32
	Percent float32
}

type QuestionAttributeItem struct {
	Name  string
	Items []QuestionDistributionItem
}

type MediaUsage struct {
	TotalQuestions int32
	WithAudio      int32
	WithImage      int32
	AudioPercent   float32
	ImagePercent   float32
}

type QuestionTypeItem struct {
	Type    string
	Total   int32
	Percent float32
}

type QuestionBankOverview struct {
	Attributes []QuestionAttributeItem
	MediaUsage MediaUsage
	Types      []QuestionTypeItem
}

type InactiveStudent struct {
	Id         int64            `json:"id"`
	Name       string           `json:"name"`
	Class      string           `json:"class"`
	LastLogin  string           `json:"last_login"`
	AbsentDays int32            `json:"absent_days"`
	AvatarInfo models.MediaInfo `json:"avatar_info"`
}

type ScoreItem struct {
	Id    int64   `json:"id"`
	Score float64 `json:"score"`
}

type DecliningStudent struct {
	Id         int64
	Name       string
	Class      string           `json:"class"`
	Score      string           `gorm:"column:score"`
	AvatarInfo models.MediaInfo `json:"avatar_info"`
}

type SlowGradingTeacher struct {
	TeacherId         int64
	TeacherName       string
	GradedSubmissions int32
	TotalSubmissions  int32
	AvgWaitHours      float32
	AvatarInfo        models.MediaInfo `json:"avatar_info"`
}

type RiskAndWarning struct {
	InactiveStudents    []InactiveStudent
	DecliningStudents   []DecliningStudent
	SlowGradingTeachers []SlowGradingTeacher
}

type GradingSummary struct {
	TotalSubmissions    int32
	GradedSubmissions   int32
	UngradedSubmissions int32
	UngradedPercent     float32
}

type WeeklyMarking struct {
	Id             int64
	WeekNumber     string
	StartDate      string
	EndDate        string
	MarkingPercent float32
	AssignedCount  int32
	MarkedCount    int32
}

type WeeklyMarkingRate struct {
	Id                  int64
	WeekNumber          string
	MarkingPercent      float32
	AssignedCount       int32
	MarkedCount         int32
	UngradedSubmissions []UngradedSubmission
}

type UngradedSubmission struct {
	Type          string // "exam", "homework", "exercise"
	TeacherId     int64
	Id            int64 // exam_id, homework_id, or exercise_id
	StudentId     int64
	TeacherName   string
	StudentName   string
	Name          string // exam_name, homework_name, or exercise_name
	QuestionIds   string
	QuestionCount int32
	SubmittedAt   string
}

// UngradedSubmissionExam is kept for backward compatibility
type UngradedSubmissionExam struct {
	TeacherId     int64
	ExamId        int64
	StudentId     int64
	TeacherName   string
	StudentName   string
	ExamName      string
	QuestionIds   string
	QuestionCount int32
	SubmittedAt   string
}

type TeacherPerformanceOverview struct {
	GradingSummary     GradingSummary
	WeeklyMarkingRates []WeeklyMarkingRate
}

type FilterAddress struct {
	ProvinceCode string
	DistrictCode string
	WardCode     string
}

type FilterDashboardAdmin struct {
	TimeType          string
	StartTime         time.Time
	EndTime           time.Time
	StartPreviousTime time.Time
	EndPreviousTime   time.Time
	Address           FilterAddress
	ProgramId         int64
	CourseId          int64
	TeacherId         int64
	SchoolId          int64
	ClassId           int64
	RoleId            int64
	Year              int32
	Month             int32
	Quarter           int32
	LastYear          int32
	LastMonth         int32
	LastQuarter       int32
	GradingType              string
}
