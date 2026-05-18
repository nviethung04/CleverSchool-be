package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type ExtractProgress struct {
	JobID       string     `json:"job_id"`
	Status      string     `json:"status"`   // "pending", "downloading", "extracting", "uploading", "scanning", "completed", "failed"
	Progress    int        `json:"progress"` // 0-100
	Message     string     `json:"message"`
	Error       string     `json:"error,omitempty"`
	FinalURL    string     `json:"final_url,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

const (
	StatusPending     = "pending"
	StatusDownloading = "downloading"
	StatusExtracting  = "extracting"
	StatusUploading   = "uploading"
	StatusScanning    = "scanning"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
)

// GenerateJobID tạo job ID từ key
func GenerateJobID(key string) string {
	return fmt.Sprintf("extract:%s", key)
}

// SaveProgress lưu progress vào Redis
func SaveProgress(jobID string, progress *ExtractProgress) error {
	if db.RedisClient == nil {
		return nil // Nếu không có Redis, bỏ qua
	}

	progress.UpdatedAt = time.Now()
	data, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	// Lưu với TTL 1 giờ
	err = db.RedisClient.Set(context.Background(), jobID, data, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save progress: %w", err)
	}

	return nil
}

// GetProgress lấy progress từ Redis
func GetProgress(jobID string) (*ExtractProgress, error) {
	if db.RedisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	val, err := db.RedisClient.Get(context.Background(), jobID).Result()
	if err == redis.Nil {
		return nil, nil // Không tìm thấy
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	var progress ExtractProgress
	err = json.Unmarshal([]byte(val), &progress)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
	}

	progress.FinalURL = utils.StaticURL(progress.FinalURL, models.Storage)

	return &progress, nil
}

// UpdateProgress helper để update progress dễ dàng
func UpdateProgress(jobID, status string, progress int, message string) error {
	prog, err := GetProgress(jobID)
	if err != nil || prog == nil {
		// Tạo mới nếu chưa có
		prog = &ExtractProgress{
			JobID:     jobID,
			Status:    status,
			Progress:  progress,
			Message:   message,
			StartedAt: time.Now(),
		}
	} else {
		prog.Status = status
		prog.Progress = progress
		prog.Message = message
	}

	if status == StatusCompleted || status == StatusFailed {
		now := time.Now()
		prog.CompletedAt = &now
	}

	return SaveProgress(jobID, prog)
}
