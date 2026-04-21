package dto

type HomeworkRanking struct {
	StudentID   int64   `json:"id"`
	StudentName string  `json:"name"`
	Avatar      string  `json:"avatar"`
	Class       string  `json:"class"` // Tên lớp, nhiều lớp cách nhau bởi dấu phẩy
	Star        int64   `json:"star"`
	Exp         float64 `json:"exp"`
	Rank        int     `json:"rank"` // 1, 2, 3, ...
	IsMe        bool    `json:"is_me"`
}
