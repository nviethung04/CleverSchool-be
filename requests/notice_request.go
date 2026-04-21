package requests

type NoticeRequest struct {
	Cover       *string `json:"cover"`
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	Status      string  `json:"status" binding:"required"`
	Type        string  `json:"type" binding:"required"`
}
