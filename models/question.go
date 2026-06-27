package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	QuestionTypeFillInBlanks   = "fill_in_blanks"
	QuestionTypeVocabulary     = "vocabulary"
	QuestionTypeWriting        = "writing"
	QuestionTypeSpeaking       = "speaking"
	QuestionTypeCategory       = "category"
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeOrdering       = "ordering"
	QuestionTypeMatching       = "matching"
	QuestionTypeLabeling       = "labeling"
	QuestionTypeDragDrop       = "drag_drop"
	DisplayWordCount           = "word_count"
	DisplayVertical            = "vertical"
	DisplayHorizontal          = "horizontal"
)

var QuestionTypes = []string{
	QuestionTypeFillInBlanks,
	QuestionTypeVocabulary,
	QuestionTypeWriting,
	QuestionTypeSpeaking,
	QuestionTypeCategory,
	QuestionTypeMultipleChoice,
	QuestionTypeOrdering,
	QuestionTypeMatching,
	QuestionTypeLabeling,
	QuestionTypeDragDrop,
}

type Question struct {
	ID           int64  `gorm:"primaryKey"`
	Kind         string `gorm:"type:KIND_ENUM;not null"`
	QuestionType string `gorm:"type:QUESTION_TYPE_ENUM;not null"`

	Title       string
	Description string
	Content     string
	Keywords    string
	FileInfo    MediaInfo `gorm:"type:jsonb" json:"file_info"`
	FileInfos   MediaInfos `gorm:"column:file_infos;type:jsonb" json:"file_infos"`

	TimeLimitSeconds int
	// Extra metadata for writing/speaking question types
	MaxCharacters    int  `gorm:"column:max_characters" json:"max_characters"`
	AllowImageUpload bool `gorm:"column:allow_image_upload" json:"allow_image_upload"`
	MaxRecordingTime int  `gorm:"column:max_recording_time" json:"max_recording_time"`
	SortPosition     int
	Point            float64
	IsRandom         int16 `gorm:"default:0"`
	ImageWidth       int
	ImageHeight      int

	Status bool
	Display string `gorm:"type:display_enum;default:'horizontal'" json:"display"`

	SourceQuestionId int64           `gorm:"null" json:"source_question_id"`
	Source           *SourceQuestion `gorm:"foreignKey:SourceQuestionId"`

	SubjectId int64           `gorm:"null" json:"subject_id"`
	Subject           *Subject `gorm:"foreignKey:SubjectId"`

	Answers           []Answer               `gorm:"foreignKey:QuestionID"`
	AnswerPositions   []AnswerPosition       `gorm:"foreignKey:QuestionID"`
	AnswerGroups      []AnswerGroup          `gorm:"foreignKey:QuestionID"`
	AnswerCoordinates []AnswerCoordinates    `gorm:"foreignKey:QuestionID"`
	AnswerMatchings   []AnswerMatching       `gorm:"foreignKey:QuestionID"`
	Attributes        []QuestionAttribute    `gorm:"many2many:question_ref_attributes"`
	RefAttributes     []QuestionRefAttribute `gorm:"foreignKey:QuestionID;references:ID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedBy int64          `json:"created_by"`
	UpdatedBy int64          `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
