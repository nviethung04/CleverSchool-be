package dto

type ChatMessageReactionRequest struct {
	Emoji string `json:"emoji" binding:"required" validate:"required"`
}

type ChatMessageReactionResponse struct {
	Success   bool                    `json:"success"`
	Reactions []ReactionSummaryDTO    `json:"reactions"`
	Message   string                  `json:"message,omitempty"`
}

type ReactionSummaryDTO struct {
	Emoji      string            `json:"emoji"`
	Count      int64             `json:"count"`
	Users      []string          `json:"users"`
	UserCounts map[string]int    `json:"userCounts"`
}
