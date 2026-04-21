package models

import (
	"time"
)

type Week struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	Year       int       `json:"year"`
	WeekNumber int       `json:"week_number"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
