package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChatMessageRepository interface {
	base.BaseRepositoryInterface[models.ChatMessage]
	base.BeforeQueryHook

	// Chat-specific methods
	CreateMessage(ctx context.Context, message *models.ChatMessage) error
	CreateMessageWithMedias(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error
	GetMessagesByCourseID(ctx context.Context, courseId uint64, page, limit int) ([]models.ChatMessage, int64, error)
	GetMessageByID(ctx context.Context, id uint64) (*models.ChatMessage, error)
	UpdateMessage(ctx context.Context, message *models.ChatMessage) error
	DeleteMessage(ctx context.Context, id uint64) error
	TogglePinMessage(ctx context.Context, id uint64) error
	GetPinnedMessagesByCourseID(ctx context.Context, courseId uint64) ([]models.ChatMessage, error)
	CountMessagesByCourseID(ctx context.Context, courseId uint64) (int64, error)
	GetRecentMessagesByCourseID(ctx context.Context, courseId uint64, limit int) ([]models.ChatMessage, error)

	// Reply-specific methods
	CreateReply(ctx context.Context, message *models.ChatMessage) error
	GetRepliesByMessageID(ctx context.Context, messageID uint64, page, limit int) ([]models.ChatMessage, int64, error)
	GetMessageWithReplies(ctx context.Context, messageID uint64) (*models.ChatMessage, error)
	GetReplyCount(ctx context.Context, messageID uint64) (int64, error)
	DeleteRepliesByMessageID(ctx context.Context, messageID uint64) error
}

type chatMessageRepository struct {
	*base.BaseRepository[models.ChatMessage]
}

func NewChatMessageRepository() ChatMessageRepository {
	return &chatMessageRepository{
		BaseRepository: base.NewBaseRepository[models.ChatMessage](),
	}
}

// CreateMessage tạo một tin nhắn mới
func (r *chatMessageRepository) CreateMessage(ctx context.Context, message *models.ChatMessage) error {
	return db.MasterDB.WithContext(ctx).Create(message).Error
}

// CreateMessageWithMedias tạo tin nhắn mới kèm theo media files
func (r *chatMessageRepository) CreateMessageWithMedias(ctx context.Context, message *models.ChatMessage, mediaIDs []int64) error {
	// Start transaction
	tx := db.MasterDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create message
	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create message-media relationships if mediaIDs provided
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

// GetMessagesByCourseID lấy danh sách tin nhắn theo courseId với phân trang
func (r *chatMessageRepository) GetMessagesByCourseID(ctx context.Context, courseId uint64, page, limit int) ([]models.ChatMessage, int64, error) {
	var messages []models.ChatMessage
	var total int64

	dbQuery := db.ReplicaDB.WithContext(ctx).Model(&models.ChatMessage{})

	// Count total records
	if err := dbQuery.Where("course_id = ?", courseId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated messages with all relations
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

// GetMessageByID lấy tin nhắn theo ID với đầy đủ thông tin liên quan
func (r *chatMessageRepository) GetMessageByID(ctx context.Context, id uint64) (*models.ChatMessage, error) {
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

// UpdateMessage cập nhật tin nhắn
func (r *chatMessageRepository) UpdateMessage(ctx context.Context, message *models.ChatMessage) error {
	return db.MasterDB.WithContext(ctx).Save(message).Error
}

// DeleteMessage xóa tin nhắn (soft delete)
func (r *chatMessageRepository) DeleteMessage(ctx context.Context, id uint64) error {
	return db.MasterDB.WithContext(ctx).Delete(&models.ChatMessage{}, id).Error
}

// TogglePinMessage chuyển đổi trạng thái pin của tin nhắn
func (r *chatMessageRepository) TogglePinMessage(ctx context.Context, id uint64) error {
	return db.MasterDB.WithContext(ctx).Model(&models.ChatMessage{}).
		Where("id = ?", id).
		Update("is_pinned", gorm.Expr("NOT is_pinned")).Error
}

// GetPinnedMessagesByCourseID lấy danh sách tin nhắn đã pin trong khóa học
func (r *chatMessageRepository) GetPinnedMessagesByCourseID(ctx context.Context, courseId uint64) ([]models.ChatMessage, error) {
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

// CountMessagesByCourseID đếm tổng số tin nhắn trong khóa học
func (r *chatMessageRepository) CountMessagesByCourseID(ctx context.Context, courseId uint64) (int64, error) {
	var count int64
	err := db.ReplicaDB.WithContext(ctx).Model(&models.ChatMessage{}).
		Where("course_id = ?", courseId).
		Count(&count).Error

	return count, err
}

// GetRecentMessagesByCourseID lấy tin nhắn gần đây nhất trong khóa học
func (r *chatMessageRepository) GetRecentMessagesByCourseID(ctx context.Context, courseId uint64, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := db.ReplicaDB.WithContext(ctx).
		Where("course_id = ?", courseId).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

// BeforeQuery thực hiện các xử lý trước khi query (implement interface với đúng signature)
func (r *chatMessageRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	// Có thể thêm các điều kiện filter chung ở đây
	// Ví dụ: chỉ lấy tin nhắn chưa bị xóa
	return query.Where("deleted_at IS NULL")
}

// ========== REPLY METHODS ==========

// CreateReply tạo một reply message mới
func (r *chatMessageRepository) CreateReply(ctx context.Context, message *models.ChatMessage) error {
	return db.MasterDB.WithContext(ctx).Create(message).Error
}

// GetRepliesByMessageID lấy danh sách replies của một message với phân trang
func (r *chatMessageRepository) GetRepliesByMessageID(ctx context.Context, messageID uint64, page, limit int) ([]models.ChatMessage, int64, error) {
	var replies []models.ChatMessage
	var total int64

	// Count total replies
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ? AND deleted_at IS NULL", messageID).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Get paginated replies with user info
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

// GetMessageWithReplies lấy message kèm thông tin replies
func (r *chatMessageRepository) GetMessageWithReplies(ctx context.Context, messageID uint64) (*models.ChatMessage, error) {
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

// GetReplyCount đếm số lượng replies của một message
func (r *chatMessageRepository) GetReplyCount(ctx context.Context, messageID uint64) (int64, error) {
	var count int64
	err := db.ReplicaDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ? AND deleted_at IS NULL", messageID).
		Count(&count).Error

	return count, err
}

// DeleteRepliesByMessageID xóa tất cả replies của một message (soft delete)
func (r *chatMessageRepository) DeleteRepliesByMessageID(ctx context.Context, messageID uint64) error {
	// Soft delete all replies when parent message is deleted
	return db.MasterDB.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("reply_to_message_id = ?", messageID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
