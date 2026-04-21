package jobs

import (
	"be-lms/config"
	"be-lms/redis"
	"be-lms/repositories"
	"time"

	"github.com/robfig/cron/v3"
)

type PermanentDeleter interface {
	PermanentlyDeleteOldRecords(before time.Time) error
}

func StartCleanupCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// run at 2AM every days
	c.AddFunc("0 2 * * *", func() {
		sevenDaysAgo := time.Now().AddDate(0, -1, 0)

		repos := []PermanentDeleter{
			repositories.NewClassRepository(),
			repositories.NewChapterRepository(),
			repositories.NewCourseRepository(),
			repositories.NewLessonRepository(),
			repositories.NewQuestionRepository(),
			//repositories.NewSchoolRepository(),
			repositories.NewSourceQuestionRepository(),
			repositories.NewSubjectRepository(),
			repositories.NewUserRepository(),
			repositories.NewSkillRepository(),
			repositories.NewTopicRepository(),
			repositories.NewTagRepository(),
			repositories.NewCertificateRepository(),
			repositories.NewDegreeRepository(),
			repositories.NewDepartmentRepository(),
			repositories.NewEmployeePositionRepository(),
			repositories.NewClonedQuestionRepository(),
		}

		for _, repo := range repos {
			err := repo.PermanentlyDeleteOldRecords(sevenDaysAgo)
			if err != nil {
				config.Log.Error("Xoá dữ liệu cũ thất bại:", err)
			}
		}
	})

	// Run ActivityLogRepository cleanup: 2AM, only on day 1 every month
	// c.AddFunc("0 2 1 * *", func() {
	// 	repo := repositories.NewActivityLogRepository()
	// 	err := repo.PermanentlyDeleteOldRecords(time.Now())
	// 	if err != nil {
	// 		config.Log.Error("Xoá ActivityLog cũ thất bại:", err)
	// 	}
	// })

	// run every minute
	c.AddFunc("* * * * *", func() {
		repo := redis.NewUserSessionRepository()
		err := repo.PermanentlyDeleteOldRecords()

		if err != nil {
			config.Log.Error("Xoá dữ liệu cũ thất bại:", err)
		}
	})

	// Delete reset password every hour
	c.AddFunc("0 * * * *", func() {
		passwordResetRepo := repositories.NewPasswordResetRepository()
		err := passwordResetRepo.DeleteExpired()
		if err != nil {
			config.Log.Error("Failed to delete expired password resets:", err)
		}
	})

	c.Start()
}
