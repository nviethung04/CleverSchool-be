package requests

// FlashcardFilterRequest represents filters for flashcard queries
type FlashcardFilterRequest struct {
	Status     []string `form:"status"`     // new, learning, known, mastered
	Level      []string `form:"level"`      // beginner, intermediate, advanced
	Difficulty []int    `form:"difficulty"` // 1-5
	IsKnown    *bool    `form:"is_known"`
	Tags       []string `form:"tags"`

	// Pagination
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`

	// Sorting
	SortBy    string `form:"sort_by"`    // word, difficulty, created_at, last_studied_at
	SortOrder string `form:"sort_order"` // asc, desc
}
