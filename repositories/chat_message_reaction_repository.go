package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ChatMessageReactionRepository interface {
	// Basic CRUD operations
	CreateReaction(ctx context.Context, reaction *models.ChatMessageReaction) error
	DeleteReaction(ctx context.Context, messageID, userID uint64, emoji string) error
	DeleteAllReactionsByMessageID(ctx context.Context, messageID uint64) error
	DeleteAllReactionsByUserAndMessageID(ctx context.Context, messageID, userID uint64) error
	GetReactionsByMessageID(ctx context.Context, messageID uint64) ([]models.ChatMessageReaction, error)
	ExistReaction(ctx context.Context, messageID, userID uint64, emoji string) (bool, error)

	// Aggregation methods
	GetReactionsSummaryByMessageID(ctx context.Context, messageID uint64) ([]ReactionSummary, error)
}

type ReactionSummary struct {
	Emoji   string   `json:"emoji"`
	Count   int64    `json:"count"`
	Users   []string `json:"users"`
	UserIDs []uint64 `json:"user_ids"`
}

type chatMessageReactionRepository struct{}

func NewChatMessageReactionRepository() ChatMessageReactionRepository {
	return &chatMessageReactionRepository{}
}

// CreateReaction tạo reaction mới hoặc tăng quantity nếu đã tồn tại
func (r *chatMessageReactionRepository) CreateReaction(ctx context.Context, reaction *models.ChatMessageReaction) error {
	// Check if reaction already exists
	var existingReaction models.ChatMessageReaction
	err := db.MasterDB.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", reaction.MessageID, reaction.UserID, reaction.Emoji).
		First(&existingReaction).Error

	if err == nil {
		// Reaction already exists, increment quantity
		return db.MasterDB.WithContext(ctx).
			Model(&existingReaction).
			Update("quantity", existingReaction.Quantity+1).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new reaction with quantity = 1
	if reaction.Quantity == 0 {
		reaction.Quantity = 1
	}
	return db.MasterDB.WithContext(ctx).Create(reaction).Error
}

// DeleteReaction xóa hoàn toàn reaction của user (không quan tâm quantity)
func (r *chatMessageReactionRepository) DeleteReaction(ctx context.Context, messageID, userID uint64, emoji string) error {
	// Delete the reaction record completely, regardless of quantity
	return db.MasterDB.WithContext(ctx).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&models.ChatMessageReaction{}).Error
}

// GetReactionsByMessageID lấy tất cả reactions của một tin nhắn
func (r *chatMessageReactionRepository) GetReactionsByMessageID(ctx context.Context, messageID uint64) ([]models.ChatMessageReaction, error) {
	var reactions []models.ChatMessageReaction
	err := db.ReplicaDB.WithContext(ctx).
		Preload("User").
		Where("message_id = ?", messageID).
		Find(&reactions).Error

	return reactions, err
}

// ExistReaction kiểm tra xem reaction có tồn tại không
func (r *chatMessageReactionRepository) ExistReaction(ctx context.Context, messageID, userID uint64, emoji string) (bool, error) {
	var count int64
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessageReaction{}).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Count(&count).Error

	return count > 0, err
}

// GetReactionsSummaryByMessageID lấy tổng hợp reactions theo emoji
func (r *chatMessageReactionRepository) GetReactionsSummaryByMessageID(ctx context.Context, messageID uint64) ([]ReactionSummary, error) {
	var results []ReactionSummary

	// Query that sums quantity instead of counting records
	type ReactionCount struct {
		Emoji string
		Count int64
	}

	var reactionCounts []ReactionCount
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessageReaction{}).
		Select("emoji, SUM(quantity) as count").
		Where("message_id = ?", messageID).
		Group("emoji").
		Order("count DESC, emoji ASC").
		Find(&reactionCounts).Error

	if err != nil {
		return nil, err
	}

	// Get user details for each emoji
	for _, rc := range reactionCounts {
		var reactions []models.ChatMessageReaction
		err := db.ReplicaDB.WithContext(ctx).
			Preload("User").
			Where("message_id = ? AND emoji = ?", messageID, rc.Emoji).
			Find(&reactions).Error

		if err != nil {
			return nil, err
		}

		var users []string
		var userIDs []uint64
		for _, reaction := range reactions {
			if reaction.User.Name != "" {
				// Add each user only once, regardless of quantity
				users = append(users, reaction.User.Name)
			}
			userIDs = append(userIDs, reaction.UserID)
		}

		results = append(results, ReactionSummary{
			Emoji:   rc.Emoji,
			Count:   rc.Count,
			Users:   users,
			UserIDs: userIDs,
		})
	}

	return results, nil
}

// DeleteAllReactionsByMessageID xóa tất cả reactions của một message
func (r *chatMessageReactionRepository) DeleteAllReactionsByMessageID(ctx context.Context, messageID uint64) error {
	return db.MasterDB.WithContext(ctx).
		Where("message_id = ?", messageID).
		Delete(&models.ChatMessageReaction{}).Error
}

// DeleteAllReactionsByUserAndMessageID xóa tất cả reactions của một user trên một message cụ thể
func (r *chatMessageReactionRepository) DeleteAllReactionsByUserAndMessageID(ctx context.Context, messageID, userID uint64) error {
	return db.MasterDB.WithContext(ctx).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		Delete(&models.ChatMessageReaction{}).Error
}

// Helper functions (keeping them for compatibility but simplifying)
// func parseStringArray(str string) []string {
// 	if str == "" {
// 		return []string{}
// 	}
// 	return []string{str}
// }

// func parseUint64Array(str string) []uint64 {
// 	if str == "" {
// 		return []uint64{}
// 	}
// 	return []uint64{}
// }
