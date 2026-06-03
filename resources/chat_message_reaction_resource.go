package resources

import (
	"be-lms/models"
)

type ChatMessageReactionResource struct{}

func NewChatMessageReactionResource() *ChatMessageReactionResource {
	return &ChatMessageReactionResource{}
}

type ReactionSummaryResource struct {
	Emoji      string            `json:"emoji"`
	Count      int64             `json:"count"`
	Users      []string          `json:"users"`
	UserCounts map[string]int    `json:"userCounts"`
}

type ChatMessageReactionResponseResource struct {
	Success   bool                      `json:"success"`
	Reactions []ReactionSummaryResource `json:"reactions"`
	Message   string                    `json:"message,omitempty"`
}

// FormatReactions formats reaction models to response resource
func (r *ChatMessageReactionResource) FormatReactions(reactions []models.ChatMessageReaction) []ReactionSummaryResource {
	// Group reactions by emoji
	emojiMap := make(map[string][]models.ChatMessageReaction)
	
	for _, reaction := range reactions {
		emojiMap[reaction.Emoji] = append(emojiMap[reaction.Emoji], reaction)
	}
	
	var result []ReactionSummaryResource
	
	for emoji, emojiReactions := range emojiMap {
		var users []string
		userCounts := make(map[string]int)
		
		for _, reaction := range emojiReactions {
			if reaction.User.Name != "" {
				users = append(users, reaction.User.Name)
				userCounts[reaction.User.Name]++
			}
		}
		
		summary := ReactionSummaryResource{
			Emoji:      emoji,
			Count:      int64(len(emojiReactions)),
			Users:      users,
			UserCounts: userCounts,
		}
		
		result = append(result, summary)
	}
	
	return result
}

// FormatReactionResponse formats complete reaction response
func (r *ChatMessageReactionResource) FormatReactionResponse(success bool, reactions []models.ChatMessageReaction, message string) *ChatMessageReactionResponseResource {
	return &ChatMessageReactionResponseResource{
		Success:   success,
		Reactions: r.FormatReactions(reactions),
		Message:   message,
	}
}
