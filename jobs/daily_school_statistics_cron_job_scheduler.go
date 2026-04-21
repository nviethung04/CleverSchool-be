package jobs

import (
	"be-lms/config"

	"github.com/robfig/cron/v3"
)

func StartDailySchoolStatisticsCronJob() {
	c := cron.New(cron.WithLocation(Location()))
	
	// Chạy lúc 1h sáng hàng ngày (0 1 * * *)
	_, err := c.AddFunc("0 1 * * *", func() {
		config.Log.Info("🚀 Bắt đầu Daily School Statistics Cron Job")
		
		job := NewDailySchoolStatisticsCronJob()
		if err := job.Run(); err != nil {
			config.Log.Errorf("❌ Daily School Statistics Job failed: %v", err)
		} else {
			config.Log.Info("✅ Daily School Statistics Job completed successfully")
		}
	})
	
	if err != nil {
		config.Log.Errorf("❌ Failed to schedule Daily School Statistics Cron Job: %v", err)
		return
	}
	
	c.Start()
	config.Log.Info("⏰ Daily School Statistics Cron Job scheduled - runs daily at 1:00 AM")
}
