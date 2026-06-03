package models

import (
	"time"

	"gorm.io/gorm"
)

type Contest struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:text;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      int16          `json:"status"`
	StartTime   *time.Time     `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	ImageInfo   MediaInfo      `gorm:"type:jsonb" json:"image_info"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   int64          `json:"created_by"`
	UpdatedAt   time.Time      `json:"updated_at"`
	UpdatedBy   int64          `json:"updated_by"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy   int64          `gorm:"column:deleted_by" json:"deleted_by"`

	// Relationships
	ContestRounds []ContestRound `gorm:"foreignKey:ContestId;references:ID" json:"contest_rounds"`
	Creator       *User          `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
}

type ContestRound struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestId    int64          `gorm:"not null" json:"contest_id"`
	Name         string         `gorm:"type:text" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	SortPosition int64          `json:"sort_position"`
	JoinLevel    string         `gorm:"type:join_level_enum" json:"join_level"`
	StartTime    *time.Time     `json:"start_time"`
	EndTime      *time.Time     `json:"end_time"`
	CreatedAt    time.Time      `json:"created_at"`
	CreatedBy    int64          `json:"created_by"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UpdatedBy    int64          `json:"updated_by"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy    int64          `gorm:"column:deleted_by" json:"deleted_by"`

	QuestionForm string `gorm:"type:question_form_enum;not null"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`

	// Relationships
	Contest Contest `gorm:"foreignKey:ContestId" json:"contest"`
}

type ContestRoundUser struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId   int64     `json:"contest_round_id"`
	UserId           int64     `json:"user_id"`
	Score            float64   `gorm:"type:numeric(5,2)" json:"score"`
	Ratio            float64   `gorm:"type:numeric(5,2)" json:"ratio"`
	Time             int64     `json:"time"`
	HasManualScoring bool      `json:"has_manual_scoring"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
}

type ContestRoundQuestionUserPosition struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId      int64     `json:"contest_round_id"`
	UserId              int64     `json:"user_id"`
	QuestionId          int64     `json:"question_id"`
	AnswerGroupPosition int64     `json:"answer_group_position"`
	SortPosition        int32     `json:"sort_position"`
	IsCorrect           bool      `json:"is_correct"`
	Score               float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt           time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserMultipleChoice struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64     `json:"contest_round_id"`
	UserId         int64     `json:"user_id"`
	QuestionId     int64     `json:"question_id"`
	AnswerId       int64     `json:"answer_id"`
	TrueAnswerId   int64     `json:"true_answer_id"`
	IsCorrect      bool      `json:"is_correct"`
	Score          float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserMatching struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64     `json:"contest_round_id"`
	UserId         int64     `json:"user_id"`
	QuestionId     int64     `json:"question_id"`
	FirstItemId    int64     `json:"first_item_id"`
	SecondItemId   int64     `json:"second_item_id"`
	IsCorrect      bool      `json:"is_correct"`
	Score          float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserManualScoring struct {
	ID             int64                  `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64                  `json:"contest_round_id"`
	UserId         int64                  `json:"user_id"`
	QuestionId     int64                  `json:"question_id"`
	Answer         string                 `gorm:"type:text" json:"answer"`
	Score          float64                `gorm:"type:numeric(5,2)" json:"score"`
	IsScored       bool                   `json:"is_scored"`
	CreatedAt      time.Time              `json:"created_at"`
	ScoringBy      int64                  `json:"scoring_by"`
	ScoringAt      *time.Time             `json:"scoring_at"`
	FileInfo       map[string]interface{} `gorm:"type:jsonb" json:"file_info"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserLabeling struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64     `json:"contest_round_id"`
	QuestionId     int64     `json:"question_id"`
	UserId         int64     `json:"user_id"`
	BlankId        int64     `json:"blank_id"`
	AnswerId       int64     `json:"answer_id"`
	IsCorrect      bool      `json:"is_correct"`
	Score          float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserGroup struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64     `json:"contest_round_id"`
	UserId         int64     `json:"user_id"`
	QuestionId     int64     `json:"question_id"`
	AnswerId       int64     `json:"answer_id"`
	GroupId        int64     `json:"group_id"`
	IsCorrect      bool      `json:"is_correct"`
	Score          float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundQuestionUserFillInBlank struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64     `json:"contest_round_id"`
	UserId         int64     `json:"user_id"`
	QuestionId     int64     `json:"question_id"`
	Answer         string    `gorm:"type:text" json:"answer"`
	SortPosition   int32     `json:"sort_position"`
	IsCorrect      bool      `json:"is_correct"`
	Score          float64   `gorm:"type:numeric(5,2)" json:"score"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
	Question     Question     `gorm:"foreignKey:QuestionId" json:"question"`
}

type ContestRoundJoinerSchool struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64          `json:"contest_round_id"`
	SchoolId       int64          `json:"school_id"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      int64          `json:"created_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy      int64          `gorm:"column:deleted_by" json:"deleted_by"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId;references:ID" json:"contest_round"`
	School       School       `gorm:"foreignKey:SchoolId;references:ID" json:"school"`
}

type ContestRoundJoinerProvince struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64          `json:"contest_round_id"`
	ProvinceId     int64          `json:"province_id"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      int64          `json:"created_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy      int64          `gorm:"column:deleted_by" json:"deleted_by"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	Province     Province     `gorm:"foreignKey:ProvinceId;references:ID" json:"province"`
}

type ContestRoundJoinerPerson struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64          `json:"contest_round_id"`
	UserId         int64          `json:"user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      int64          `json:"created_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy      int64          `gorm:"column:deleted_by" json:"deleted_by"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	User         User         `gorm:"foreignKey:UserId" json:"user"`
}

func (ContestRoundJoinerPerson) TableName() string {
	return "contest_round_joiner_persons"
}

type ContestRoundJoinerClass struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ContestRoundId int64          `json:"contest_round_id"`
	ClassId        int64          `json:"class_id"`
	CreatedAt      time.Time      `json:"created_at"`
	CreatedBy      int64          `json:"created_by"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
	DeletedBy      int64          `gorm:"column:deleted_by" json:"deleted_by"`

	// Relationships
	ContestRound ContestRound `gorm:"foreignKey:ContestRoundId" json:"contest_round"`
	Class        Class        `gorm:"foreignKey:ClassId" json:"class"`
}
