package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageMediaRepository interface {
	base.BaseRepositoryInterface[models.MessageMedia]
	base.BeforeQueryHook

	// MessageMedia-specific methods
	CreateMessageMedia(ctx context.Context, messageMedia *models.MessageMedia) error
	CreateMessageMedias(ctx context.Context, messageMedias []models.MessageMedia) error
	DeleteMessageMediasByMessageID(ctx context.Context, messageID uint64) error
	GetMessageMediasByMessageID(ctx context.Context, messageID uint64) ([]models.MessageMedia, error)
	GetMediasByMessageID(ctx context.Context, messageID uint64) ([]models.Media, error)
	UpdateMessageMediaSortOrder(ctx context.Context, messageMediaID uint64, sortOrder int) error
	GetMessageMediaByID(ctx context.Context, id uint64) (*models.MessageMedia, error)
	DeleteMessageMedia(ctx context.Context, id uint64) error
}

type messageMediaRepository struct {
	*base.BaseRepository[models.MessageMedia]
}

func NewMessageMediaRepository() MessageMediaRepository {
	return &messageMediaRepository{
		BaseRepository: base.NewBaseRepository[models.MessageMedia](),
	}
}

// CreateMessageMedia tạo một liên kết message-media mới
func (r *messageMediaRepository) CreateMessageMedia(ctx context.Context, messageMedia *models.MessageMedia) error {
	return db.MasterDB.WithContext(ctx).Create(messageMedia).Error
}

// CreateMessageMedias tạo nhiều liên kết message-media cùng lúc
func (r *messageMediaRepository) CreateMessageMedias(ctx context.Context, messageMedias []models.MessageMedia) error {
	return db.MasterDB.WithContext(ctx).CreateInBatches(messageMedias, 100).Error
}

// DeleteMessageMediasByMessageID xóa tất cả media của một message
func (r *messageMediaRepository) DeleteMessageMediasByMessageID(ctx context.Context, messageID uint64) error {
	return db.MasterDB.WithContext(ctx).
		Where("message_id = ?", messageID).
		Delete(&models.MessageMedia{}).Error
}

// GetMessageMediasByMessageID lấy danh sách message-media theo messageID
func (r *messageMediaRepository) GetMessageMediasByMessageID(ctx context.Context, messageID uint64) ([]models.MessageMedia, error) {
	var messageMedias []models.MessageMedia
	err := db.ReplicaDB.WithContext(ctx).
		Where("message_id = ?", messageID).
		Order("sort_order ASC").
		Find(&messageMedias).Error

	return messageMedias, err
}

// GetMediasByMessageID lấy danh sách media theo messageID với đầy đủ thông tin media
func (r *messageMediaRepository) GetMediasByMessageID(ctx context.Context, messageID uint64) ([]models.Media, error) {
	var medias []models.Media
	err := db.ReplicaDB.WithContext(ctx).
		Table("medias").
		Joins("JOIN message_medias ON medias.id = message_medias.media_id").
		Where("message_medias.message_id = ?", messageID).
		Order("message_medias.sort_order ASC").
		Find(&medias).Error

	return medias, err
}

// UpdateMessageMediaSortOrder cập nhật thứ tự sắp xếp của message-media
func (r *messageMediaRepository) UpdateMessageMediaSortOrder(ctx context.Context, messageMediaID uint64, sortOrder int) error {
	return db.MasterDB.WithContext(ctx).
		Model(&models.MessageMedia{}).
		Where("id = ?", messageMediaID).
		Update("sort_order", sortOrder).Error
}

// GetMessageMediaByID lấy message-media theo ID
func (r *messageMediaRepository) GetMessageMediaByID(ctx context.Context, id uint64) (*models.MessageMedia, error) {
	var messageMedia models.MessageMedia
	err := db.ReplicaDB.WithContext(ctx).
		Preload("Media").
		Preload("Message").
		First(&messageMedia, id).Error

	return &messageMedia, err
}

// DeleteMessageMedia xóa một liên kết message-media
func (r *messageMediaRepository) DeleteMessageMedia(ctx context.Context, id uint64) error {
	return db.MasterDB.WithContext(ctx).Delete(&models.MessageMedia{}, id).Error
}

// BeforeQuery thực hiện các xử lý trước khi query
func (r *messageMediaRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	// Có thể thêm các điều kiện filter chung ở đây
	return query
}

