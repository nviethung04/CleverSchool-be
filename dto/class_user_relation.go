package dto

type ClassUserRelationRequest struct {
	ClassID    int64   `json:"class_id" binding:"required"`
	StudentIDs []int64 `json:"student_ids" binding:"required"`
}

type ClassUserRelationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
} 