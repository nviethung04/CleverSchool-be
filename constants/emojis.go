package constants

// ValidEmojis contains the list of allowed emojis for chat reactions
var ValidEmojis = []string{
	// Basic emotions
	"👍", "👎", "❤️", "😂", "😮", "😢", "😡",
	
	// Reactions
	"👏", "🔥", "💯", "🎉", "💪", "🙏", "👀",
	
	// Faces
	"😊", "😍", "🤔", "😭", "🥰", "💡", "🤩",
	
	// Objects
	"✨", "🚀", "💝", "🌟", "📚", "🎯", "💎",
	
	// Additional popular reactions
	"🤝", "👌", "💖", "🙌", "🤞", "💪", "🎊",
	"😄", "😎", "🤗", "👨‍🎓", "🙄",  "🙋‍♂️", "🤨",
	"🎈", "🏆", "⭐", "💫", "🔥", "❌", "✅",
}

// EmojiCategories groups emojis by category for better organization
var EmojiCategories = map[string][]string{
	"positive": {"👍", "❤️", "😂", "👏", "🔥", "💯", "🎉", "😊", "😍", "🥰", "✨", "🌟"},
	"negative": {"👎", "😢", "😡", "😭", "💡"},
	"neutral":  {"😮", "🤔", "👀", "👨‍🎓", "🙄",  "🙋‍♂️", "🤨"},
	"objects":  {"🚀", "📝", "📚", "🎯", "💎", "🏆", "⭐","❌","✅"},
}

// IsValidEmoji checks if an emoji is in the allowed list
func IsValidEmoji(emoji string) bool {
	for _, validEmoji := range ValidEmojis {
		if emoji == validEmoji {
			return true
		}
	}
	return false
}

// GetEmojisByCategory returns emojis for a specific category
func GetEmojisByCategory(category string) []string {
	if emojis, exists := EmojiCategories[category]; exists {
		return emojis
	}
	return []string{}
}

// GetAllCategories returns all available emoji categories
func GetAllCategories() []string {
	categories := make([]string, 0, len(EmojiCategories))
	for category := range EmojiCategories {
		categories = append(categories, category)
	}
	return categories
}
