package models

// DashboardOverviewResult represents the result of get_dashboard_overview function
type DashboardOverviewResult struct {
	SchoolTotalCount    int64 `json:"school_total_count" gorm:"column:school_total_count"`
	SchoolCurrentCount  int64 `json:"school_current_count" gorm:"column:school_current_count"`
	SchoolPreviousCount int64 `json:"school_previous_count" gorm:"column:school_previous_count"`

	UserTotalCount    int64 `json:"user_total_count" gorm:"column:user_total_count"`
	UserCurrentCount  int64 `json:"user_current_count" gorm:"column:user_current_count"`
	UserPreviousCount int64 `json:"user_previous_count" gorm:"column:user_previous_count"`

	StudentTotalCount    int64 `json:"student_total_count" gorm:"column:student_total_count"`
	StudentCurrentCount  int64 `json:"student_current_count" gorm:"column:student_current_count"`
	StudentPreviousCount int64 `json:"student_previous_count" gorm:"column:student_previous_count"`

	TeacherTotalCount    int64 `json:"teacher_total_count" gorm:"column:teacher_total_count"`
	TeacherCurrentCount  int64 `json:"teacher_current_count" gorm:"column:teacher_current_count"`
	TeacherPreviousCount int64 `json:"teacher_previous_count" gorm:"column:teacher_previous_count"`

	ActivityTotalCount    int64 `json:"activity_total_count" gorm:"column:activity_total_count"`
	ActivityCurrentCount  int64 `json:"activity_current_count" gorm:"column:activity_current_count"`
	ActivityPreviousCount int64 `json:"activity_previous_count" gorm:"column:activity_previous_count"`
	ActivityNewCount      int64 `json:"activity_new_count" gorm:"column:activity_new_count"`
}
