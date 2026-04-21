package models

type SemesterRefHoliday struct {
	ID         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	SemesterId int64 `gorm:"not null" json:"semester_id"`
	HolidayId   int64 `gorm:"not null" json:"holiday_id"`
}
