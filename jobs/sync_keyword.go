package jobs

import (
	"be-Clever School/repositories"
	"be-Clever School/services"

	"github.com/robfig/cron/v3"
)

func StartSyncKeywordCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// run every minute
	c.AddFunc("* * * * *", func() {
	})

	// run every 30 minutes
	c.AddFunc("0,30 * * * *", func() {
		questionService := services.NewQuestionService(
			repositories.NewQuestionRepository(),
		)
		questionService.SyncKeywords(nil, 0)
	})

	c.Start()
}
