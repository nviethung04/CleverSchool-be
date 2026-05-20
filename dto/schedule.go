package dto

import (
	"be-cleverschool/models"
	"time"
)

type ScheduleGroup struct {
	WeekId     int
	WeekNumber int
	Year       int
	StartDate  time.Time
	EndDate    time.Time
	Days       map[string][]models.LessonSchedule
}

type LessonSchedule struct {
	WeekId           int
	WeekNumberByYear int
	WeekNumber       int
	Year             int
	StartDate        time.Time
	EndDate          time.Time
	Lessons          []models.LessonSchedule
}

type LessonScheduleDetail struct {
	Groups           []LessonSchedule
	Semesters       []models.Semester
}

