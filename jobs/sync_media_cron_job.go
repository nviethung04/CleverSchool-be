package jobs

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/repositories/base"
	"time"

	"github.com/robfig/cron/v3"
)

type SyncMediaInfoRepo interface {
	SyncMediaInfo() error
	SyncMediaInfoForRecentlyUpdatedMedia(within time.Duration) error
}

func StartSyncMediaCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// // run at 2AM every days
	c.AddFunc("0 2 * * *", func() {
		SyncAllQuestionsMediaInfoJob()
	})

	// run every minute
	// c.AddFunc("* * * * *", func() {
	// 	SyncMediaInfoForRecentlyUpdatedMediaJob(30 * time.Minute)
	// })

	// run every 30 minutes - sync media info for records with recently updated media
	// c.AddFunc("0,30 * * * *", func() {
	// 	SyncMediaInfoForRecentlyUpdatedMediaJob(30 * time.Minute)
	// })

	c.Start()
}

func SyncAllQuestionsMediaInfoJob() {
	config.Log.Info("Starting sync MediaInfo for all questions and answers")

	repos := []SyncMediaInfoRepo{
		base.NewBaseRepository[*models.Question](),
		base.NewBaseRepository[*models.Answer](),
		base.NewBaseRepository[*models.AnswerCoordinates](),
		base.NewBaseRepository[*models.AnswerGroup](),
		base.NewBaseRepository[*models.AnswerMatching](),
		base.NewBaseRepository[*models.AnswerPosition](),
		base.NewBaseRepository[*models.GroupAnswer](),
		base.NewBaseRepository[*models.User](),
		base.NewBaseRepository[*models.Certificate](),
		base.NewBaseRepository[*models.Course](),
		base.NewBaseRepository[*models.Degree](),
		base.NewBaseRepository[*models.Exam](),
		base.NewBaseRepository[*models.Homework](),
		base.NewBaseRepository[*models.LessonPlanPart](),
		base.NewBaseRepository[*models.LessonPlan](),
		base.NewBaseRepository[*models.School](),
		base.NewBaseRepository[*models.Skill](),
		base.NewBaseRepository[*models.Tag](),
		base.NewBaseRepository[*models.Topic](),
		base.NewBaseRepository[*models.HomeworkQuestionUserManualScoring](),
		base.NewBaseRepository[*models.ExamQuestionUserManualScoring](),
	}

	for _, repo := range repos {
		err := repo.SyncMediaInfo()
		if err != nil {
			config.Log.Error("Failed to sync MediaInfo:", err)
		}
	}
	config.Log.Info("Finished syncing MediaInfo")
}

// SyncMediaInfoForRecentlyUpdatedMediaJob syncs media info only for records
// that have *_info.id matching medias updated within the specified duration
func SyncMediaInfoForRecentlyUpdatedMediaJob(within time.Duration) {
	config.Log.Infof("Starting sync MediaInfo for records with media updated within %v", within)

	repos := []SyncMediaInfoRepo{
		base.NewBaseRepository[*models.Question](),
		base.NewBaseRepository[*models.Answer](),
		base.NewBaseRepository[*models.AnswerCoordinates](),
		base.NewBaseRepository[*models.AnswerGroup](),
		base.NewBaseRepository[*models.AnswerMatching](),
		base.NewBaseRepository[*models.AnswerPosition](),
		base.NewBaseRepository[*models.GroupAnswer](),
		base.NewBaseRepository[*models.User](),
		base.NewBaseRepository[*models.Certificate](),
		base.NewBaseRepository[*models.Course](),
		base.NewBaseRepository[*models.Degree](),
		base.NewBaseRepository[*models.Exam](),
		base.NewBaseRepository[*models.Homework](),
		base.NewBaseRepository[*models.LessonPlanPart](),
		base.NewBaseRepository[*models.LessonPlan](),
		base.NewBaseRepository[*models.School](),
		base.NewBaseRepository[*models.Skill](),
		base.NewBaseRepository[*models.Tag](),
		base.NewBaseRepository[*models.Topic](),
		base.NewBaseRepository[*models.HomeworkQuestionUserManualScoring](),
		base.NewBaseRepository[*models.ExamQuestionUserManualScoring](),
	}

	for _, repo := range repos {
		err := repo.SyncMediaInfoForRecentlyUpdatedMedia(within)
		if err != nil {
			config.Log.Error("Failed to sync MediaInfo for recently updated media:", err)
		}
	}

	config.Log.Info("Finished syncing MediaInfo for recently updated media")
}

