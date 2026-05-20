package jobs

import (
	"be-cleverschool/config"

	"github.com/robfig/cron/v3"
)

func StartDailyCourseStatisticsCronJob() {
	c := cron.New(cron.WithLocation(Location()))
	
	// Chạy lúc 1h sáng hàng ngày (0 1 * * *)
	_, err := c.AddFunc("0 1 * * *", func() {
		config.Log.Info("🚀 Bắt đầu Daily Course Statistics Cron Job")
		
		job := NewDailyCourseStatisticsCronJob()
		if err := job.Run(); err != nil {
			config.Log.Errorf("❌ Daily Course Statistics Job failed: %v", err)
		} else {
			config.Log.Info("✅ Daily Course Statistics Job completed successfully")
		}
	})
	
	if err != nil {
		config.Log.Errorf("❌ Failed to schedule Daily Course Statistics Cron Job: %v", err)
		return
	}
	
	c.Start()
	config.Log.Info("⏰ Daily Course Statistics Cron Job scheduled - runs daily at 1:00 AM")
}

