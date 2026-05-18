package dto

import "be-Clever School/models"

type QuestionWithScore struct {
	ID       int64   `gorm:"column:id" json:"id"`
	Content  string  `gorm:"column:content" json:"content"`
	FileURL  string  `gorm:"column:file_url" json:"file_url"`
	Type     string  `gorm:"column:type" json:"type"`
	Kind     string  `gorm:"column:kind" json:"kind"`
	MaxScore float64 `gorm:"column:max_score" json:"max_score"`
}

// ExamAnswerMultipleChoice Multiple Choice: Lấy câu trả lời của người dùng và câu trả lời đúng (is_correct = true)
type ExamAnswerMultipleChoice struct {
	ID                 int64    `json:"id"`
	ExamID             int64    `json:"exam_id"`
	QuestionID         int64    `json:"question_id"`
	UserID             int64    `json:"user_id"`
	AnswerIDs          []int64  `json:"answer_ids"`
	AnswerContents     []string `json:"answer_contents"`
	AnswerFileURLs     []string `json:"answer_file_urls"`
	AnswerKinds        []string `json:"answer_kinds"`
	TrueAnswerIDs      []int64  `json:"true_answer_ids"`
	TrueAnswers        []string `json:"true_answers"`
	TrueAnswerFileURLs []string `json:"true_answer_file_urls"`
	TrueAnswerKinds    []string `json:"true_answer_kinds"`
	IsCorrect          bool     `json:"is_correct"`
	Score              float64  `json:"score"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	AllAnswerIDs       []int64  `json:"all_answer_ids"`
	AllAnswerFileURLs  []string `json:"all_answer_file_urls"`
	AllAnswerKinds     []string `json:"all_answer_kinds"`
	AllAnswerContents  []string `json:"all_answer_contents"`
}

type ExamAnswerFillInBlank struct {
	SortPosition int     `gorm:"column:sort_position" json:"sort_position"`
	Answer       string  `gorm:"column:answer" json:"answer"`
	TrueAnswer   string  `gorm:"column:true_answer" json:"true_answer"`
	IsCorrect    bool    `gorm:"column:is_correct" json:"is_correct"`
	Score        float64 `gorm:"column:score" json:"score"`
}

type ExamAnswerPosition struct {
	SortPosition        int32   `json:"sort_position"`
	AnswerGroupPosition int64   `json:"answer_group_position"`
	AnswerContent       string  `json:"answer_content"`
	AnswerFileURL       string  `json:"answer_file_url"`
	AnswerFileKind      string  `json:"answer_file_kind"`
	TrueAnswerID        int64   `json:"true_answer_id"`
	TrueAnswerContent   string  `json:"true_answer_content"`
	TrueAnswerFileURL   string  `json:"true_answer_file_url"`
	TrueAnswerFileKind  string  `json:"true_answer_file_kind"`
	IsCorrect           bool    `json:"is_correct"`
	Score               float64 `json:"score"`
}

type ExamAnswerMatching struct {
	ID                int64   `gorm:"column:id" json:"id"`
	ExamID            int64   `gorm:"column:exam_id" json:"exam_id"`
	QuestionID        int64   `gorm:"column:question_id" json:"question_id"`
	UserID            int64   `gorm:"column:user_id" json:"user_id"`
	FirstItemID       int64   `gorm:"column:first_item_id" json:"first_item_id"`
	SecondItemID      int64   `gorm:"column:second_item_id" json:"second_item_id"`
	FirstItemContent  string  `gorm:"column:first_item_content" json:"first_item_content"`
	SecondItemContent string  `gorm:"column:second_item_content" json:"second_item_content"`
	FirstItemFileURL  string  `gorm:"column:first_item_file_url" json:"first_item_file_url"`
	FirstItemKind     string  `gorm:"column:first_item_kind" json:"first_item_kind"`
	SecondItemFileURL string  `gorm:"column:second_item_file_url" json:"second_item_file_url"`
	SecondItemKind    string  `gorm:"column:second_item_kind" json:"second_item_kind"`
	IsCorrect         bool    `gorm:"column:is_correct" json:"is_correct"`
	Score             float64 `gorm:"column:score" json:"score"`
	CreatedAt         string  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         string  `gorm:"column:updated_at" json:"updated_at"`
}

type ExamAnswerLabeling struct {
	BlankID             int64   `json:"blank_id"`
	TrueBlankContent    string  `json:"true_blank_content"`
	TrueBlankFileURL    string  `json:"true_blank_file_url"`
	TrueBlankKind       string  `json:"true_blank_kind"`
	BlankPositionX      float64 `json:"blank_position_x"`
	BlankPositionY      float64 `json:"blank_position_y"`
	BlankPositionWidth  float64 `json:"blank_position_width"`
	BlankPositionHeight float64 `json:"blank_position_height"`
	AnswerID            int64   `json:"answer_id"`
	AnswerContent       string  `json:"answer_content"`
	AnswerFileURL       string  `json:"answer_file_url"`
	AnswerKind          string  `json:"answer_kind"`
	IsCorrect           bool    `json:"is_correct"`
	Score               float64 `json:"score"`
}

type ExamAnswerGroup struct {
	AnswerID       int64   `json:"answer_id"`
	GroupID        int64   `json:"group_id"`
	AnswerContent  string  `json:"answer_content"`
	GroupContent   string  `json:"group_content"`
	AnswerFileURL  string  `json:"answer_file_url"`
	AnswerFileKind string  `json:"answer_file_kind"`
	GroupFileURL   string  `json:"group_file_url"`
	GroupFileKind  string  `json:"group_file_kind"`
	IsCorrect      bool    `json:"is_correct"`
	Score          float64 `json:"score"`
}

type ExamAnswerManual struct {
	Answer         string           `json:"answer"`
	AnswerFileInfo models.MediaInfo `json:"answer_file_info"`
	Score          float64          `json:"score"`
	IsScored       bool             `json:"is_scored"`
}
