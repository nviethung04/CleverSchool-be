package models

type ExamQuestion struct {
	ID               int64 `gorm:"primaryKey"`
	ExamID           int64
	QuestionID       int64
	Score            float64
	IsSourceQuestion bool `gorm:"default:false"`
}

type HomeworkQuestion struct {
	ID               int64 `gorm:"primaryKey"`
	HomeworkID       int64
	QuestionID       int64
	Score            float64
	IsSourceQuestion bool `gorm:"default:false"`
}

type LessonPlanPartQuestion struct {
	ID               int64 `gorm:"primaryKey"`
	LessonPlanPartID int64
	QuestionID       int64
	Score            float64
	IsSourceQuestion bool `gorm:"default:false"`
}

type LevelTestQuestion struct {
	ID               int64 `gorm:"primaryKey"`
	LevelTestID      int64
	QuestionID       int64
	Score            float64
	IsSourceQuestion bool `gorm:"default:false"`
}

type ExerciseQuestion struct {
	ID               int64 `gorm:"primaryKey"`
	ExerciseID       int64
	QuestionID       int64
	Score            float64
	IsSourceQuestion bool `gorm:"default:false"`
}