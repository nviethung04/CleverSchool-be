package models

import (
	"time"
)

type ExamQuestionUser struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	ExamID       int64     `gorm:"column:exam_id" json:"exam_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	AnswerID     int64     `gorm:"column:answer_id" json:"answer_id"`
	TrueAnswerID int64     `gorm:"column:true_answer_id" json:"true_answer_id"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExamQuestionUser) TableName() string {
	return "exam_question_users"
}

type HomeworkQuestionUser struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID   int64     `gorm:"column:homework_id" json:"homework_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	AnswerID     int64     `gorm:"column:answer_id" json:"answer_id"`
	TrueAnswerID int64     `gorm:"column:true_answer_id" json:"true_answer_id"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (HomeworkQuestionUser) TableName() string {
	return "homework_question_users"
}

type HomeworkUserQuestion struct {
	ID             int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID     int64     `gorm:"column:homework_id" json:"homework_id"`
	UserID         int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID     int64     `gorm:"column:question_id" json:"question_id"`
	RatioScore     float64   `gorm:"column:ratio_score" json:"ratio_score"`
	IsAllCorrect   bool      `gorm:"column:is_all_correct" json:"is_all_correct"`
	Star           int       `gorm:"column:star" json:"star"`
	NumberOptions  int       `gorm:"column:number_options" json:"number_options"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	LessonID       int64     `gorm:"column:lesson_id" json:"lesson_id"`
	NumberTimeSent int       `gorm:"column:number_time_sent" json:"number_time_sent"`
	Weight         float64   `gorm:"column:weight" json:"weight"`
}

func (HomeworkUserQuestion) TableName() string {
	return "homework_user_questions"
}

type ExerciseQuestionUser struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	ExerciseID   int64     `gorm:"column:exercise_id" json:"exercise_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	AnswerID     int64     `gorm:"column:answer_id" json:"answer_id"`
	TrueAnswerID int64     `gorm:"column:true_answer_id" json:"true_answer_id"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExerciseQuestionUser) TableName() string {
	return "exercise_question_users"
}

type LevelTestQuestionUser struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	LevelTestID  int64     `gorm:"column:level_test_id" json:"level_test_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	AnswerID     int64     `gorm:"column:answer_id" json:"answer_id"`
	TrueAnswerID int64     `gorm:"column:true_answer_id" json:"true_answer_id"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (LevelTestQuestionUser) TableName() string {
	return "level_test_question_users"
}

type ExamQuestionUserFillInBlank struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	ExamID       int64     `gorm:"column:exam_id" json:"exam_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	Answer       string    `gorm:"column:answer" json:"answer"`
	SortPosition int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExamQuestionUserFillInBlank) TableName() string {
	return "exam_question_user_fill_in_blanks"
}

type HomeworkQuestionUserFillInBlank struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID   int64     `gorm:"column:homework_id" json:"exam_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	Answer       string    `gorm:"column:answer" json:"answer"`
	SortPosition int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (HomeworkQuestionUserFillInBlank) TableName() string {
	return "homework_question_user_fill_in_blanks"
}

type ExerciseQuestionUserFillInBlank struct {
	ID           int64     `gorm:"primaryKey;column:id" json:"id"`
	ExerciseID   int64     `gorm:"column:exercise_id" json:"exercise_id"`
	LessonID     int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID   int64     `gorm:"column:question_id" json:"question_id"`
	Answer       string    `gorm:"column:answer" json:"answer"`
	SortPosition int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect    bool      `gorm:"column:is_correct" json:"is_correct"`
	Score        float64   `gorm:"column:score" json:"score"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExerciseQuestionUserFillInBlank) TableName() string {
	return "exercise_question_user_fill_in_blanks"
}

type ExamQuestionUserPosition struct {
	ID                  int64     `gorm:"primaryKey;column:id" json:"id"`
	ExamID              int64     `gorm:"column:exam_id" json:"exam_id"`
	LessonID            int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID              int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID          int64     `gorm:"column:question_id" json:"question_id"`
	AnswerGroupPosition int64     `gorm:"column:answer_group_position" json:"answer_group_position"`
	SortPosition        int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect           bool      `gorm:"column:is_correct" json:"is_correct"`
	Score               float64   `gorm:"column:score" json:"score"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExamQuestionUserPosition) TableName() string {
	return "exam_question_user_positions"
}

type HomeworkQuestionUserPosition struct {
	ID                  int64     `gorm:"primaryKey;column:id" json:"id"`
	HomeworkID          int64     `gorm:"column:homework_id" json:"homework_id"`
	LessonID            int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID              int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID          int64     `gorm:"column:question_id" json:"question_id"`
	AnswerGroupPosition int64     `gorm:"column:answer_group_position" json:"answer_group_position"`
	SortPosition        int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect           bool      `gorm:"column:is_correct" json:"is_correct"`
	Score               float64   `gorm:"column:score" json:"score"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (HomeworkQuestionUserPosition) TableName() string {
	return "homework_question_user_positions"
}

type ExerciseQuestionUserPosition struct {
	ID                  int64     `gorm:"primaryKey;column:id" json:"id"`
	ExerciseID          int64     `gorm:"column:exercise_id" json:"exercise_id"`
	LessonID            int64     `gorm:"column:lesson_id" json:"lesson_id"`
	UserID              int64     `gorm:"column:user_id" json:"user_id"`
	QuestionID          int64     `gorm:"column:question_id" json:"question_id"`
	AnswerGroupPosition int64     `gorm:"column:answer_group_position" json:"answer_group_position"`
	SortPosition        int       `gorm:"column:sort_position" json:"sort_position"`
	IsCorrect           bool      `gorm:"column:is_correct" json:"is_correct"`
	Score               float64   `gorm:"column:score" json:"score"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ExerciseQuestionUserPosition) TableName() string {
	return "exercise_question_user_positions"
}

type ExamQuestionUserMatching struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	ExamID       int64     `gorm:"column:exam_id"`
	LessonID     int64     `gorm:"column:lesson_id"`
	UserID       int64     `gorm:"column:user_id"`
	QuestionID   int64     `gorm:"column:question_id"`
	FirstItemID  int64     `gorm:"column:first_item_id"`
	SecondItemID int64     `gorm:"column:second_item_id"`
	IsCorrect    bool      `gorm:"column:is_correct"`
	Score        float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ExamQuestionUserMatching) TableName() string {
	return "exam_question_user_matchings"
}

type HomeworkQuestionUserMatching struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	HomeworkID   int64     `gorm:"column:homework_id"`
	LessonID     int64     `gorm:"column:lesson_id"`
	UserID       int64     `gorm:"column:user_id"`
	QuestionID   int64     `gorm:"column:question_id"`
	FirstItemID  int64     `gorm:"column:first_item_id"`
	SecondItemID int64     `gorm:"column:second_item_id"`
	IsCorrect    bool      `gorm:"column:is_correct"`
	Score        float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (HomeworkQuestionUserMatching) TableName() string {
	return "homework_question_user_matchings"
}

type ExerciseQuestionUserMatching struct {
	ID           int64     `gorm:"primaryKey;column:id"`
	ExerciseID   int64     `gorm:"column:exercise_id"`
	LessonID     int64     `gorm:"column:lesson_id"`
	UserID       int64     `gorm:"column:user_id"`
	QuestionID   int64     `gorm:"column:question_id"`
	FirstItemID  int64     `gorm:"column:first_item_id"`
	SecondItemID int64     `gorm:"column:second_item_id"`
	IsCorrect    bool      `gorm:"column:is_correct"`
	Score        float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (ExerciseQuestionUserMatching) TableName() string {
	return "exercise_question_user_matchings"
}

type ExamQuestionUserLabeling struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	ExamID     int64     `gorm:"column:exam_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	QuestionID int64     `gorm:"column:question_id"`
	UserID     int64     `gorm:"column:user_id"`
	BlankID    int64     `gorm:"column:blank_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (ExamQuestionUserLabeling) TableName() string {
	return "exam_question_user_labelings"
}

type HomeworkQuestionUserLabeling struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	HomeworkID int64     `gorm:"column:homework_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	QuestionID int64     `gorm:"column:question_id"`
	UserID     int64     `gorm:"column:user_id"`
	BlankID    int64     `gorm:"column:blank_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (HomeworkQuestionUserLabeling) TableName() string {
	return "homework_question_user_labelings"
}

type ExerciseQuestionUserLabeling struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	ExerciseID int64     `gorm:"column:exercise_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	QuestionID int64     `gorm:"column:question_id"`
	UserID     int64     `gorm:"column:user_id"`
	BlankID    int64     `gorm:"column:blank_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (ExerciseQuestionUserLabeling) TableName() string {
	return "exercise_question_user_labelings"
}

type ExamQuestionUserGroup struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	ExamID     int64     `gorm:"column:exam_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	UserID     int64     `gorm:"column:user_id"`
	QuestionID int64     `gorm:"column:question_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	GroupID    int64     `gorm:"column:group_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (ExamQuestionUserGroup) TableName() string {
	return "exam_question_user_groups"
}

type HomeworkQuestionUserGroup struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	HomeworkID int64     `gorm:"column:homework_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	UserID     int64     `gorm:"column:user_id"`
	QuestionID int64     `gorm:"column:question_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	GroupID    int64     `gorm:"column:group_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (HomeworkQuestionUserGroup) TableName() string {
	return "homework_question_user_groups"
}

type ExerciseQuestionUserGroup struct {
	ID         int64     `gorm:"primaryKey;column:id"`
	ExerciseID int64     `gorm:"column:exercise_id"`
	LessonID   int64     `gorm:"column:lesson_id"`
	UserID     int64     `gorm:"column:user_id"`
	QuestionID int64     `gorm:"column:question_id"`
	AnswerID   int64     `gorm:"column:answer_id"`
	GroupID    int64     `gorm:"column:group_id"`
	IsCorrect  bool      `gorm:"column:is_correct"`
	Score      float64   `gorm:"column:score"` // numeric(5,2)
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (ExerciseQuestionUserGroup) TableName() string {
	return "exercise_question_user_groups"
}

type ExamQuestionUserManualScoring struct {
	ID         int64      `gorm:"primaryKey;column:id"`
	ExamID     int64      `gorm:"column:exam_id"`
	LessonID   int64      `gorm:"column:lesson_id"`
	UserID     int64      `gorm:"column:user_id"`
	QuestionID int64      `gorm:"column:question_id"`
	Answer     *string    `gorm:"column:answer"` // text
	FileInfo   MediaInfo  `gorm:"type:jsonb" json:"file_info"`
	Score      *float64   `gorm:"column:score"`      // numeric(5,2)
	IsScored   bool       `gorm:"column:is_scored"`  // boolean
	CreatedAt  *time.Time `gorm:"column:created_at"` // timestamp
	ScoringBy  *int64     `gorm:"column:scoring_by"` // bigint
	ScoringAt  *time.Time `gorm:"column:scoring_at"` // timestamp
}

// TableName chỉ định tên bảng tương ứng trong DB
func (ExamQuestionUserManualScoring) TableName() string {
	return "exam_question_user_manual_scoring"
}

type HomeworkQuestionUserManualScoring struct {
	ID         int64      `gorm:"primaryKey;column:id"`
	HomeworkID int64      `gorm:"column:homework_id"`
	LessonID   int64      `gorm:"column:lesson_id"`
	UserID     int64      `gorm:"column:user_id"`
	QuestionID int64      `gorm:"column:question_id"`
	Answer     *string    `gorm:"column:answer"` // text
	FileInfo   MediaInfo  `gorm:"type:jsonb" json:"file_info"`
	Score      *float64   `gorm:"column:score"`      // numeric(5,2)
	IsScored   bool       `gorm:"column:is_scored"`  // boolean
	CreatedAt  *time.Time `gorm:"column:created_at"` // timestamp
	ScoringBy  *int64     `gorm:"column:scoring_by"` // bigint
	ScoringAt  *time.Time `gorm:"column:scoring_at"` // timestamp
}

// TableName sets the insert table name for this struct type
func (HomeworkQuestionUserManualScoring) TableName() string {
	return "homework_question_user_manual_scoring"
}

type ExerciseQuestionUserManualScoring struct {
	ID         int64      `gorm:"primaryKey;column:id"`
	ExerciseID int64      `gorm:"column:exercise_id"`
	LessonID   int64      `gorm:"column:lesson_id"`
	UserID     int64      `gorm:"column:user_id"`
	QuestionID int64      `gorm:"column:question_id"`
	Answer     *string    `gorm:"column:answer"` // text
	FileInfo   MediaInfo  `gorm:"type:jsonb" json:"file_info"`
	Score      *float64   `gorm:"column:score"`      // numeric(5,2)
	IsScored   bool       `gorm:"column:is_scored"`  // boolean
	CreatedAt  *time.Time `gorm:"column:created_at"` // timestamp
	ScoringBy  *int64     `gorm:"column:scoring_by"` // bigint
	ScoringAt  *time.Time `gorm:"column:scoring_at"` // timestamp
}

func (ExerciseQuestionUserManualScoring) TableName() string {
	return "exercise_question_user_manual_scoring"
}
