package dto

type CopyScheduleResponse struct {
	CopyScheduleSuccess  bool   `gorm:"-" json:"copy_schedule_success"`
	CopyScheduleMessage  string `gorm:"-" json:"copy_schedule_message"`
}
