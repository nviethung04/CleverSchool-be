package dto

import (
	"be-cleverschool/models"
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
	TypeId     int64
	Name       string
	TypeUserId int64
	TypeName   string
	Score      float32
	Type       string           `json:"type"` // "exam", "homework", "exercise"
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
	Week      int32
	Duration  float32
	IsCurrent bool
}

type DeviceUsage struct {
	Name    string
	Count   int32
	Percent float32
}

type SystemUsageOverview struct {
	WeeklyUsages  []WeeklyUsage
	DeviceUsages  []DeviceUsage
	AverageUsed   models.AverageUsed
	CompletedRate float32
	CompleteCount int32
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

type NotAssignAssignment struct {
	TeacherId            int64
	AssignAssignmentId   int64
	TeacherName          string
	AssignAssignmentName string
	QuestionIds          string
	QuestionCount        int32
}

type NotGraded struct {
	TeacherId     int64
	NotGradedId   int64
	StudentId     int64
	CourseId      int64
	TeacherName   string
	StudentName   string
	CourseName    string
	LessonName    string
	NotGradedName string
	QuestionIds   string
	QuestionCount int32
	SubmittedAt   string
}

type WeeklyAssignAssignmentRate struct {
	Id                           int64
	WeekNumber                   string
	Total                        int32
	AssignAssignment             int32
	NotAssignAssignmentExams     []NotAssignAssignment
	NotAssignAssignmentHomeworks []NotAssignAssignment
	NotAssignAssignmentExercises []NotAssignAssignment
}

type WeeklySubmitRate struct {
	Id                 int64
	WeekNumber         string
	Total              int32
	NotGraded          int32
	NotGradedExams     []NotGraded
	NotGradedHomeworks []NotGraded
	NotGradedExercises []NotGraded
}

type NotSubmittedExam struct {
	SubmittedId int64   // exam_id
	Name        string  // exam_name
	UserIds     []int64 // Danh sách user_id chưa làm bài này
}

type NotSubmittedGrouped struct {
	SubmittedId   int64   // homework_id hoặc exercise_id
	Name          string  // homework_name hoặc exercise_name
	SubmittedType string  // "homework" hoặc "exercise"
	StudentIds    []int64 // Danh sách student_id chưa làm bài này
}

type WeeklyPerformanceOverviewRate struct {
	Id                      int64
	WeekNumber              string
	AssignAssignmentPercent float32
	SubmitPercent           float32
	NotGradedPercent        float32
	TotalRequired           int32
	TotalSubmitted          int32
	NotSubmittedExams       []NotSubmittedExam
	NotSubmittedHomeworks   []NotSubmittedGrouped
	NotSubmittedExercises   []NotSubmittedGrouped
}

type TeacherPerformanceOverview struct {
	GradingSummary                 GradingSummary
	WeeklyMarkingRates             []WeeklyMarkingRate
	WeeklyAssignAssignmentRates    []WeeklyAssignAssignmentRate
	WeeklySubmitRates              []WeeklySubmitRate
	WeeklyPerformanceOverviewRates []WeeklyPerformanceOverviewRate
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
	ObjectType        string
}

