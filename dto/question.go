package dto

type MeiliRequest struct {
	Q                    string   `json:"q"`
	Limit                int      `json:"limit"`
	Offset               int      `json:"offset"`
	Filter               []string `json:"filter,omitempty"`
	AttributesToRetrieve []string `json:"attributesToRetrieve,omitempty"`
}

type MeiliResponse struct {
	Hits               []map[string]interface{} `json:"hits"`
	EstimatedTotalHits *int64                   `json:"estimatedTotalHits,omitempty"`
	TotalHits          *int64                   `json:"totalHits,omitempty"`
}

type MeiliQuestionDoc struct {
	ID           int64   `json:"id"`
	Status       bool    `json:"status"`
	QuestionType string  `json:"question_type"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Content      string  `json:"content"`
	Keywords     string  `json:"keywords"`
	Point        float64 `json:"point"`
	CreatedAt    string  `json:"created_at"`
}
