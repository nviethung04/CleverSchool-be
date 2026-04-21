package jobs

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/resources"
	"fmt"
	"time"
)

// RecalculateStudyReportsStarJob tính lại total_star và avg_star cho tất cả study_reports (chunk 100)
func RecalculateStudyReportsStarJob() (totalProcessed int64, totalUpdated int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("Panic in RecalculateStudyReportsStarJob: %v", r)
		}
	}()

	config.Log.Info("Starting RecalculateStudyReportsStarJob")

	if db.MasterDB == nil {
		config.Log.Error("MasterDB is nil - database not connected")
		return 0, 0, fmt.Errorf("database not connected")
	}

	startTime := time.Now()
	batchSize := 100
	var lastID int64 = 0
	var processed, updated int64

	studyReportRepo := repositories.NewStudyReportRepository()
	resource := resources.NewStudyReportResource()

	for {
		var studyReports []models.StudyReport

		err := db.MasterDB.Where("id > ?", lastID).
			Preload("Skills").
			Preload("Criteria").
			Preload("Criteria.Skills").
			Preload("Criteria.Skills.Types").
			Order("id ASC").
			Limit(batchSize).
			Find(&studyReports).Error

		if err != nil {
			return processed, updated, fmt.Errorf("failed to fetch study_reports: %w", err)
		}

		if len(studyReports) == 0 {
			break
		}

		for _, report := range studyReports {
			lastID = report.ID
			processed++

			_, _, _, totalStar, avgStar := resource.SplitReportSkills(&report)
			totalStarInt := int(totalStar)
			avgStarFloat := avgStar

			if err := studyReportRepo.UpdateTotalStarAndAvgStar(report.ID, totalStarInt, avgStarFloat); err != nil {
				config.Log.Errorf("StudyReport ID %d: Failed to update total_star and avg_star - %v", report.ID, err)
				continue
			}

			updated++
			config.Log.Infof("StudyReport ID %d: updated - total_star=%d, avg_star=%.2f",
				report.ID, totalStarInt, avgStarFloat)
		}

		config.Log.Infof("Progress: %d processed (updated: %d)", processed, updated)
	}

	duration := time.Since(startTime)
	config.Log.Infof("Finished RecalculateStudyReportsStarJob in %v, processed=%d, updated=%d", duration, processed, updated)

	return processed, updated, nil
}
