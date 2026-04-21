package models

func (LessonPlansLesson) TableName() string {
	return "lesson_plan_ref_lessons"
}

type LessonPlansLesson struct {
	LessonPlanID int64 `json:"lesson_plan_id"`
	LessonID     int   `json:"lesson_id"`
}
