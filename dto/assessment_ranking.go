package dto

type AssessmentRanking struct {
	StudentID   int64   `json:"id"`
	StudentName string  `json:"name"`
	Avatar      string  `json:"avatar"`
	Class       string  `json:"class"` // Tên lớp, nhiều lớp cách nhau bởi dấu phẩy
	Score       float64 `json:"score"`
	Star        int     `json:"star"`
	Rank        int     `json:"rank"`
	IsMe        bool    `json:"is_me"`
}
