package jobs

import (
	"be-lms/config"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

func StartClearExportFilesCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Chạy lúc 3AM mỗi ngày
	c.AddFunc("0 3 * * *", func() {
		config.Log.Info("Start delete old file export...")

		exportDir := "public/exports"
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)

		deletedCount, err := clearOldExportFiles(exportDir, sevenDaysAgo)
		if err != nil {
			config.Log.Errorf("Error when deleting old export file: %v", err)
			return
		}

		config.Log.Infof("Deleted old export file %d (from %s and earlier)",
			deletedCount, sevenDaysAgo.Format("2006-01-02"))
	})

	c.Start()
}

func clearOldExportFiles(exportDir string, beforeTime time.Time) (int, error) {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		config.Log.Infof("Directory %s does not exist, skipping", exportDir)
		return 0, nil
	}

	deletedCount := 0

	err := filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			config.Log.Warnf("Error accessing file %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Chỉ xử lý file Excel (.xlsx)
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".xlsx") {
			return nil
		}

		// Kiểm tra thời gian tạo file
		if info.ModTime().Before(beforeTime) {
			// Xóa file
			if err := os.Remove(path); err != nil {
				config.Log.Errorf("Cannot delete file %s: %v", path, err)
				return nil
			}

			deletedCount++
			config.Log.Infof("Deleted old export file: %s (created at: %s)",
				info.Name(), info.ModTime().Format("2006-01-02 15:04:05"))
		}

		return nil
	})

	if err != nil {
		return deletedCount, fmt.Errorf("error while browsing directory %s: %w", exportDir, err)
	}

	return deletedCount, nil
}
