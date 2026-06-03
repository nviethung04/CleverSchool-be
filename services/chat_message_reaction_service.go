package services

import (
	"be-lms/constants"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/resources"
	"context"
	"errors"
	"fmt"
	"time"
)

type ChatMessageReactionService interface {
	AddReaction(ctx context.Context, courseID, messageID, userID uint64, emoji string) (*ReactionResponse, error)
	RemoveReaction(ctx context.Context, courseID, messageID, userID uint64, emoji string) (*ReactionResponse, error)
	GetMessageReactions(ctx context.Context, courseID, messageID uint64) (*ReactionResponse, error)
	DeleteAllReactions(ctx context.Context, courseID, messageID, userID uint64) (*ReactionResponse, error)
}

type ReactionResponse struct {
	Success   bool              `json:"success"`
	Reactions []ReactionSummary `json:"reactions"`
	Message   string            `json:"message,omitempty"`
}

type ReactionSummary struct {
	Emoji      string         `json:"emoji"`
	Count      int64          `json:"count"`
	Users      []string       `json:"users"`
	UserCounts map[string]int `json:"userCounts"`
}

type chatMessageReactionService struct {
	reactionRepo     repositories.ChatMessageReactionRepository
	chatMessageRepo  repositories.ChatMessageRepository
	userRepo         repositories.UserRepository
	courseRepo       repositories.CourseRepository
	reactionResource *resources.ChatMessageReactionResource
}

func NewChatMessageReactionService(
	reactionRepo repositories.ChatMessageReactionRepository,
	chatMessageRepo repositories.ChatMessageRepository,
	userRepo repositories.UserRepository,
	courseRepo repositories.CourseRepository,
) ChatMessageReactionService {
	return &chatMessageReactionService{
		reactionRepo:     reactionRepo,
		chatMessageRepo:  chatMessageRepo,
		userRepo:         userRepo,
		courseRepo:       courseRepo,
		reactionResource: resources.NewChatMessageReactionResource(),
	}
}

// AddReaction thêm reaction vào tin nhắn
func (s *chatMessageReactionService) AddReaction(ctx context.Context, courseID, messageID, userID uint64, emoji string) (*ReactionResponse, error) {
	// Validate emoji
	if !s.isValidEmoji(emoji) {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.invalid_emoji"),
		}, errors.New("invalid emoji")
	}

	// Check if message exists and belongs to course
	message, err := s.chatMessageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.message_not_found"),
		}, err
	}

	if message.CourseID != courseID {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("message does not belong to course")
	}

	// Check if user has access to course
	if !s.userHasAccessToCourse(ctx, userID, courseID) {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("user does not have access to course")
	}

	// Create reaction
	reaction := &models.ChatMessageReaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now(),
	}

	err = s.reactionRepo.CreateReaction(ctx, reaction)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.create_failed"),
		}, err
	}

	// Get updated reactions
	reactions, err := s.getFormattedReactions(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.fetch_failed"),
		}, err
	}

	return &ReactionResponse{
		Success:   true,
		Reactions: reactions,
		Message:   i18n.Localize("messages.reaction_added"),
	}, nil
}

// RemoveReaction xóa reaction khỏi tin nhắn
func (s *chatMessageReactionService) RemoveReaction(ctx context.Context, courseID, messageID, userID uint64, emoji string) (*ReactionResponse, error) {
	// Validate emoji
	if !s.isValidEmoji(emoji) {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.invalid_emoji"),
		}, errors.New("invalid emoji")
	}

	// Check if message exists and belongs to course
	message, err := s.chatMessageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.message_not_found"),
		}, err
	}

	if message.CourseID != courseID {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("message does not belong to course")
	}

	// Check if user has access to course
	if !s.userHasAccessToCourse(ctx, userID, courseID) {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("user does not have access to course")
	}

	// Remove reaction
	err = s.reactionRepo.DeleteReaction(ctx, messageID, userID, emoji)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.delete_failed"),
		}, err
	}

	// Get updated reactions
	reactions, err := s.getFormattedReactions(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.fetch_failed"),
		}, err
	}

	return &ReactionResponse{
		Success:   true,
		Reactions: reactions,
		Message:   i18n.Localize("messages.reaction_removed"),
	}, nil
}

// GetMessageReactions lấy danh sách reactions của tin nhắn
func (s *chatMessageReactionService) GetMessageReactions(ctx context.Context, courseID, messageID uint64) (*ReactionResponse, error) {
	// Check if message exists and belongs to course
	message, err := s.chatMessageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.message_not_found"),
		}, err
	}

	if message.CourseID != courseID {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("message does not belong to course")
	}

	// Get reactions
	reactions, err := s.getFormattedReactions(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.fetch_failed"),
		}, err
	}

	return &ReactionResponse{
		Success:   true,
		Reactions: reactions,
	}, nil
}

// Helper functions

func (s *chatMessageReactionService) isValidEmoji(emoji string) bool {
	return constants.IsValidEmoji(emoji)
}

func (s *chatMessageReactionService) userHasAccessToCourse(ctx context.Context, userID, courseID uint64) bool {
	// Check if user is enrolled in course
	_, err := s.courseRepo.FindByID(int(courseID))
	if err != nil {
		fmt.Printf("DEBUG: Course not found with ID %d: %v\n", courseID, err)
		return false
	}

	// Get course users to check if user has access
	users, err := s.courseRepo.GetUsers(int64(courseID), 0) // 0 means all roles
	if err != nil {
		fmt.Printf("DEBUG: Error getting users for course %d: %v\n", courseID, err)
		return false
	}

	fmt.Printf("DEBUG: Found %d users for course %d, looking for userID %d\n", len(users), courseID, userID)
	for _, user := range users {
		fmt.Printf("DEBUG: Checking user ID %d against %d\n", user.ID, userID)
		if uint64(user.ID) == userID {
			fmt.Printf("DEBUG: User %d has access to course %d\n", userID, courseID)
			return true
		}
	}

	fmt.Printf("DEBUG: User %d does NOT have access to course %d\n", userID, courseID)
	return false
}

func (s *chatMessageReactionService) getFormattedReactions(ctx context.Context, messageID uint64) ([]ReactionSummary, error) {
	repoSummary, err := s.reactionRepo.GetReactionsSummaryByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	var reactions []ReactionSummary
	for _, summary := range repoSummary {
		userCounts := make(map[string]int)

		// Count occurrences of each user name
		for _, userName := range summary.Users {
			userCounts[userName]++
		}

		reactions = append(reactions, ReactionSummary{
			Emoji:      summary.Emoji,
			Count:      summary.Count,
			Users:      summary.Users,
			UserCounts: userCounts,
		})
	}

	return reactions, nil
}

// DeleteAllReactions xóa tất cả reactions của user hiện tại cho một message cụ thể
func (s *chatMessageReactionService) DeleteAllReactions(ctx context.Context, courseID, messageID, userID uint64) (*ReactionResponse, error) {
	// Check if message exists and belongs to course
	message, err := s.chatMessageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.message_not_found"),
		}, err
	}

	if message.CourseID != courseID {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("message does not belong to course")
	}

	// Check if user has access to course
	if !s.userHasAccessToCourse(ctx, userID, courseID) {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.unauthorized"),
		}, errors.New("user does not have access to course")
	}

	// Delete all reactions of this user for the message
	err = s.reactionRepo.DeleteAllReactionsByUserAndMessageID(ctx, messageID, userID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.error_occurred"),
		}, err
	}

	// Get updated reactions after deletion
	reactions, err := s.getFormattedReactions(ctx, messageID)
	if err != nil {
		return &ReactionResponse{
			Success: false,
			Message: i18n.Localize("messages.fetch_failed"),
		}, err
	}

	return &ReactionResponse{
		Success:   true,
		Reactions: reactions,
		Message:   "All your reactions deleted successfully",
	}, nil
}
