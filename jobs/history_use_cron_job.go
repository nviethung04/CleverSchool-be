package jobs

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/repositories"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func StartHistoryUseCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Chạy lúc 3h sáng Thứ 2 hàng tuần → lưu tuần trước
	c.AddFunc("0 3 * * 1", func() {
		err := SyncWeeklyHistoryUseJob()
		if err != nil {
			config.Log.Error("Failed to sync weekly history use:", err)
		}
	})

	// Chạy 3h sáng ngày 1 hàng tháng → lưu tháng trước
	c.AddFunc("0 3 1 * *", func() {
		err := SyncMonthlyHistoryUseJob()
		if err != nil {
			config.Log.Error("Failed to sync monthly history use:", err)
		}
	})

	// chạy bù
	// c.AddFunc("* * * * *", func() {
	c.AddFunc("0 17 * * *", func() {
		// err := SyncWeeklyHistoryUseJob(time.Date(2025, 8, 4, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 8, 11, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 8, 18, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 8, 25, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 9, 1, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 9, 8, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncWeeklyHistoryUseJob(time.Date(2025, 9, 15, 0, 0, 0, 0, time.Local))
		// if err != nil {
		// 	config.Log.Error("Failed to sync weekly history use:", err)
		// }

		// err = SyncMonthlyHistoryUseJob(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
		// if err != nil {
		// 	config.Log.Error("Failed to sync monthly history use:", err)
		// }
	})

	// Test chạy bù full tuần/tháng
	// c.AddFunc("*/15 * * * *", func() {
	// 	SyncAllWeekHistoryUseJob()
	// 	SyncAllMonthlyHistoryUseJob()
	// })

	c.Start()
}

// WEEKLY: lấy ngày chủ nhật tuần trước làm mốc
func SyncWeeklyHistoryUseJob(dateOpt ...time.Time) error {
    // Xác định ngày chủ nhật
    var lastSunday time.Time
    if len(dateOpt) > 0 && !dateOpt[0].IsZero() {
        // Nếu truyền vào ngày → dùng ngày đó
        lastSunday = time.Date(
            dateOpt[0].Year(),
            dateOpt[0].Month(),
            dateOpt[0].Day(),
            0, 0, 0, 0,
            dateOpt[0].Location(),
        )
    } else {
        // Nếu không truyền gì → mặc định chủ nhật tuần trước
        now := time.Now()
        lastSunday = now.AddDate(0, 0, -1) // hôm qua
        lastSunday = time.Date(lastSunday.Year(), lastSunday.Month(), lastSunday.Day(), 0, 0, 0, 0, lastSunday.Location())
    }

    historyUseRepo := repositories.NewHistoryUseRepository()
    if err := historyUseRepo.SaveWeeklyUsage(lastSunday); err != nil {
        config.Log.Errorf("Failed to save weekly usage for %s: %v", lastSunday.Format("2006-01-02"), err)
        return fmt.Errorf("failed to save weekly usage: %w", err)
    }

    config.Log.Infof("Finished SyncWeeklyHistoryUseJob for %s", lastSunday.Format("2006-01-02"))
    return nil
}

// MONTHLY: Lấy tháng trước
// MONTHLY: Có thể truyền tháng tùy chọn, mặc định tháng trước
func SyncMonthlyHistoryUseJob(dateOpt ...time.Time) error {
	var targetMonth time.Time

	if len(dateOpt) > 0 && !dateOpt[0].IsZero() {
		// Nếu truyền thời gian vào → lấy thời gian đó
		targetMonth = time.Date(
			dateOpt[0].Year(),
			dateOpt[0].Month(),
			1, 0, 0, 0, 0,
			dateOpt[0].Location(),
		)
	} else {
		// Không truyền gì → mặc định lấy tháng trước
		now := time.Now()
		previousMonth := now.AddDate(0, -1, 0)
		targetMonth = time.Date(
			previousMonth.Year(),
			previousMonth.Month(),
			1, 0, 0, 0, 0,
			previousMonth.Location(),
		)
	}

	historyUseRepo := repositories.NewHistoryUseRepository()
	if err := historyUseRepo.SaveMonthlyUsage(targetMonth); err != nil {
		config.Log.Errorf("Failed to save monthly usage for %s: %v", targetMonth.Format("2006-01"), err)
		return fmt.Errorf("failed to save monthly usage: %w", err)
	}

	config.Log.Infof("Finished SyncMonthlyHistoryUseJob for %s", targetMonth.Format("2006-01"))
	return nil
}

// Chạy bù toàn bộ tuần
func SyncAllWeekHistoryUseJob() error {
	dbConn := db.ReplicaDB

	// Lấy tất cả tuần đã kết thúc
	var weeks []struct {
		ID        int64
		StartDate time.Time
		EndDate   time.Time
	}

	err := dbConn.Raw(`
		SELECT id, start_date, end_date
		FROM weeks
		WHERE end_date < CURRENT_DATE
		ORDER BY start_date ASC
	`).Scan(&weeks).Error
	if err != nil {
		return fmt.Errorf("failed to get all weeks: %w", err)
	}

	historyUseRepo := repositories.NewHistoryUseRepository()

	for _, w := range weeks {
		// Mốc: lấy ngày Chủ nhật của tuần (hoặc end_date)
		targetDate := w.EndDate
		if err := historyUseRepo.SaveWeeklyUsage(targetDate); err != nil {
			config.Log.Errorf("Failed to save weekly usage for week [%d] (%s - %s): %v",
				w.ID,
				w.StartDate.Format("2006-01-02"),
				w.EndDate.Format("2006-01-02"),
				err,
			)
		}
	}

	config.Log.Info("Finished SyncAllWeekHistoryUseJob")
	return nil
}

// Chạy bù toàn bộ tháng
func SyncAllMonthlyHistoryUseJob() error {
	location := time.Now().Location()

	// Mốc bắt đầu cứng: 01/01/2025
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, location)

	// Mốc kết thúc: Tháng hiện tại (ngày 1)
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)

	historyUseRepo := repositories.NewHistoryUseRepository()

	current := start
	for !current.After(end) {
		if err := historyUseRepo.SaveMonthlyUsage(current); err != nil {
			config.Log.Errorf("Failed to save monthly usage for %s: %v", current.Format("2006-01"), err)
		}
		current = current.AddDate(0, 1, 0)
	}

	config.Log.Info("Finished SyncAllMonthlyHistoryUseJob")
	return nil
}

