package config

import (
	"time"
)

// TUSConfig chứa cấu hình cho TUS upload service
type TUSConfig struct {
	// Chunk size cho upload (8MB)
	ChunkSize int64
	// Max upload size (5GB)
	MaxUploadSize int64
	// Timeout cho mỗi chunk
	ChunkTimeout time.Duration
	// Số lần retry khi upload thất bại
	MaxRetries int
	// Delay giữa các lần retry
	RetryDelays []time.Duration
	// Timeout cho toàn bộ upload
	UploadTimeout time.Duration
}

// GetTUSConfig trả về cấu hình mặc định cho TUS
func GetTUSConfig() *TUSConfig {
	return &TUSConfig{
		ChunkSize:     8 * 1024 * 1024,        // 8MB
		MaxUploadSize: 5 * 1024 * 1024 * 1024, // 5GB
		ChunkTimeout:  5 * time.Minute,
		MaxRetries:    5,
		RetryDelays: []time.Duration{
			0, 1 * time.Second, 3 * time.Second, 5 * time.Second, 10 * time.Second, 30 * time.Second,
		},
		UploadTimeout: 30 * time.Minute,
	}
}
