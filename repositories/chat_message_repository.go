package repositories

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type ChatMessageRepository interface {
	Create(ctx context.Context, message *models.ChatMessage) error
	CreateWithMedias(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error
	GetByCourseID(ctx context.Context, courseId uint64, page, limit int) ([]models.ChatMessage, int64, error)
	GetByID(ctx context.Context, id uint64) (*models.ChatMessage, error)
	Delete(ctx context.Context, id uint64) error
	TogglePin(ctx context.Context, id uint64) error
	GetPinnedByCourseID(ctx context.Context, courseId uint64) ([]models.ChatMessage, error)
	CountByCourseID(ctx context.Context, courseId uint64) (int64, error)
	GetRecentByCourseID(ctx context.Context, courseId uint64, limit int) ([]models.ChatMessage, error)

	CreateReply(ctx context.Context, message *models.ChatMessage) error
	GetRepliesByID(ctx context.Context, messageID uint64, page, limit int) ([]models.ChatMessage, int64, error)
	GetWithReplies(ctx context.Context, messageID uint64) (*models.ChatMessage, error)
	GetReplyCount(ctx context.Context, messageID uint64) (int64, error)
	DeleteRepliesByID(ctx context.Context, messageID uint64) error

	// Private message methods
	SendToRecipient(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error
	GetMessagesWithRecipient(ctx context.Context, courseId, senderId, recipientId uint64, page, limit int) ([]models.ChatMessage, int64, error)
	GetRecentSenders(ctx context.Context, courseId, currentUserId uint64, limit int) ([]models.ChatMessage, error)
}

type chatMessageRepository struct{}

func NewChatMessageRepository() ChatMessageRepository {
	return &chatMessageRepository{}
}

func (r *chatMessageRepository) Create(ctx context.Context, message *models.ChatMessage) error {
	return db.MasterDB.WithContext(ctx).Create(message).Error
}

func (r *chatMessageRepository) CreateWithMedias(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error {
	tx := db.MasterDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		config.Log.Errorf("Failed to begin transaction in CreateWithMedias: %v", tx.Error)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in CreateWithMedias: %v", r)
			tx.Rollback()
		}
	}()

	if err := tx.Create(message).Error; err != nil {
		config.Log.Errorf("Failed to create message in CreateWithMedias: %v", err)
		tx.Rollback()
		return fmt.Errorf("failed to create message: %w", err)
	}

	if len(mediaIDs) > 0 {
		// Validate media IDs before creating relationships
		for _, mediaID := range mediaIDs {
			if mediaID <= 0 {
				tx.Rollback()
				return fmt.Errorf("invalid media ID: %d", mediaID)
			}
		}

		messageMedias := make([]models.MessageMedia, len(mediaIDs))
		for i, mediaID := range mediaIDs {
			messageMedias[i] = models.MessageMedia{
				MessageID: message.ID,
				MediaID:   mediaID,
				SortOrder: i,
			}
		}

		if err := tx.CreateInBatches(messageMedias, 100).Error; err != nil {
			config.Log.Errorf("Failed to create message medias in CreateWithMedias: messageID=%d, mediaIDs=%v, error=%v", message.ID, mediaIDs, err)
			tx.Rollback()
			return fmt.Errorf("failed to create message medias: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		config.Log.Errorf("Failed to commit transaction in CreateWithMedias: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *chatMessageRepository) GetByCourseID(ctx context.Context, courseId uint64, page, limit int) ([]models.ChatMessage, int64, error) {
	var messages []models.ChatMessage
	var total int64

	dbQuery := db.ReplicaDB.WithContext(ctx).Model(&models.ChatMessage{})

	if err := dbQuery.Where("course_id = ?", courseId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := dbQuery.Where("course_id = ?", courseId).
		Preload("User").
		Preload("ReplyToMessage").
		Preload("ReplyToMessage.User").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

func (r *chatMessageRepository) GetByID(ctx context.Context, id uint64) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := db.ReplicaDB.WithContext(ctx).
		Preload("User").
		Preload("ReplyToMessage").
		Preload("ReplyToMessage.User").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		First(&message, id).Error

	return &message, err
}

func (r *chatMessageRepository) Delete(ctx context.Context, id uint64) error {
	return db.MasterDB.WithContext(ctx).Delete(&models.ChatMessage{}, id).Error
}

func (r *chatMessageRepository) TogglePin(ctx context.Context, id uint64) error {
	return db.MasterDB.WithContext(ctx).Model(&models.ChatMessage{}).
		Where("id = ?", id).
		Update("is_pinned", gorm.Expr("NOT is_pinned")).Error
}

func (r *chatMessageRepository) GetPinnedByCourseID(ctx context.Context, courseId uint64) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := db.ReplicaDB.WithContext(ctx).
		Where("course_id = ? AND is_pinned = ?", courseId, true).
		Preload("User").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Order("created_at DESC").
		Find(&messages).Error

	return messages, err
}

func (r *chatMessageRepository) CountByCourseID(ctx context.Context, courseId uint64) (int64, error) {
	var count int64
	err := db.ReplicaDB.WithContext(ctx).Model(&models.ChatMessage{}).
		Where("course_id = ?", courseId).
		Count(&count).Error

	return count, err
}

func (r *chatMessageRepository) GetRecentByCourseID(ctx context.Context, courseId uint64, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := db.ReplicaDB.WithContext(ctx).
		Where("course_id = ?", courseId).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

func (r *chatMessageRepository) CreateReply(ctx context.Context, message *models.ChatMessage) error {
	return db.MasterDB.WithContext(ctx).Create(message).Error
}

func (r *chatMessageRepository) GetRepliesByID(ctx context.Context, messageID uint64, page, limit int) ([]models.ChatMessage, int64, error) {
	var replies []models.ChatMessage
	var total int64

	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ? AND deleted_at IS NULL", messageID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = db.ReplicaDB.WithContext(ctx).
		Preload("User").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Where("reply_to_message_id = ? AND deleted_at IS NULL", messageID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&replies).Error

	return replies, total, err
}

func (r *chatMessageRepository) GetWithReplies(ctx context.Context, messageID uint64) (*models.ChatMessage, error) {
	var message models.ChatMessage

	err := db.ReplicaDB.WithContext(ctx).
		Preload("User").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Preload("ReplyToMessage").
		Preload("ReplyToMessage.User").
		Where("id = ? AND deleted_at IS NULL", messageID).
		First(&message).Error

	return &message, err
}

func (r *chatMessageRepository) GetReplyCount(ctx context.Context, messageID uint64) (int64, error) {
	var count int64
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ? AND deleted_at IS NULL", messageID).
		Count(&count).Error

	return count, err
}

func (r *chatMessageRepository) DeleteRepliesByID(ctx context.Context, messageID uint64) error {
	return db.MasterDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ?", messageID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *chatMessageRepository) SendToRecipient(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error {
	tx := db.MasterDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(mediaIDs) > 0 {
		messageMedias := make([]models.MessageMedia, len(mediaIDs))
		for i, mediaID := range mediaIDs {
			messageMedias[i] = models.MessageMedia{
				MessageID: message.ID,
				MediaID:   mediaID,
				SortOrder: i,
			}
		}

		if err := tx.CreateInBatches(messageMedias, 100).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *chatMessageRepository) GetMessagesWithRecipient(ctx context.Context, courseId, senderId, recipientId uint64, page, limit int) ([]models.ChatMessage, int64, error) {
	var messages []models.ChatMessage
	var total int64

	query := db.ReplicaDB.WithContext(ctx).Model(&models.ChatMessage{}).
		Where("course_id = ? AND recipient_id IS NOT NULL", courseId).
		Where("((user_id = ? AND recipient_id = ?) OR (user_id = ? AND recipient_id = ?))", senderId, recipientId, recipientId, senderId)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Recipient").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

func (r *chatMessageRepository) GetRecentSenders(ctx context.Context, courseId, currentUserId uint64, limit int) ([]models.ChatMessage, error) {
	var allMessages []models.ChatMessage
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("course_id = ? AND recipient_id IS NOT NULL", courseId).
		Where("(user_id = ? OR recipient_id = ?)", currentUserId, currentUserId).
		Preload("User").
		Preload("Recipient").
		Preload("MessageMedias", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("MessageMedias.Media").
		Order("created_at DESC").
		Find(&allMessages).Error

	if err != nil {
		return nil, err
	}

	senderMap := make(map[uint64]*models.ChatMessage)
	for i := range allMessages {
		msg := &allMessages[i]
		var senderId uint64
		if msg.UserID == currentUserId {
			if msg.RecipientID != nil {
				senderId = *msg.RecipientID
			} else {
				continue
			}
		} else {
			senderId = msg.UserID
		}

		if existing, ok := senderMap[senderId]; !ok || msg.CreatedAt.After(existing.CreatedAt) {
			senderMap[senderId] = msg
		}
	}

	messages := make([]models.ChatMessage, 0, len(senderMap))
	for _, msg := range senderMap {
		messages = append(messages, *msg)
	}

	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].CreatedAt.Before(messages[j].CreatedAt) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}

	if len(messages) > limit {
		messages = messages[:limit]
	}

	return messages, nil
}

