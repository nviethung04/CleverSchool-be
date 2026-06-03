package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ScormService struct {
	db *gorm.DB
}

func NewScormService() *ScormService {
	return &ScormService{
		db: db.MasterDB,
	}
}

// CreateScormActivity creates a new SCORM activity
func (s *ScormService) CreateScormActivity(title, description, version, launchURL string) (*models.ScormActivity, error) {
	activity := &models.ScormActivity{
		Title:       title,
		Description: description,
		Version:     version,
		LaunchURL:   launchURL,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(activity).Error; err != nil {
		return nil, fmt.Errorf("failed to create SCORM activity: %w", err)
	}

	return activity, nil
}

// GetScormActivity retrieves a SCORM activity by ID
func (s *ScormService) GetScormActivity(id uint) (*models.ScormActivity, error) {
	var activity models.ScormActivity
	if err := s.db.First(&activity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("SCORM activity not found")
		}
		return nil, fmt.Errorf("failed to get SCORM activity: %w", err)
	}

	return &activity, nil
}

// CreateScormAttempt creates a new SCORM attempt
func (s *ScormService) CreateScormAttempt(activityID, userID uint) (*models.ScormAttempt, error) {
	// Get activity to get version
	activity, err := s.GetScormActivity(activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity for version: %w", err)
	}

	attempt := &models.ScormAttempt{
		ActivityID: activityID,
		UserID:     userID,
		Version:    activity.Version,
		Status:     "incomplete",
		StartTime:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.db.Create(attempt).Error; err != nil {
		return nil, fmt.Errorf("failed to create SCORM attempt: %w", err)
	}

	return attempt, nil
}

// GetScormAttempt retrieves a SCORM attempt by ID
func (s *ScormService) GetScormAttempt(id string) (*models.ScormAttempt, error) {
	var attempt models.ScormAttempt
	if err := s.db.First(&attempt, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("SCORM attempt not found")
		}
		return nil, fmt.Errorf("failed to get SCORM attempt: %w", err)
	}

	return &attempt, nil
}

// SetScormCMI sets a CMI data element for a SCORM attempt
func (s *ScormService) SetScormCMI(attemptID, element, value string) error {
	// Upsert CMI data
	var cmi models.ScormCMI
	err := s.db.Where("attempt_id = ? AND element = ?", attemptID, element).First(&cmi).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new CMI record
			cmi = models.ScormCMI{
				AttemptID: attemptID,
				Element:   element,
				Value:     value,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := s.db.Create(&cmi).Error; err != nil {
				return fmt.Errorf("failed to create CMI data: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check CMI data: %w", err)
		}
	} else {
		// Update existing CMI record
		cmi.Value = value
		cmi.UpdatedAt = time.Now()
		if err := s.db.Save(&cmi).Error; err != nil {
			return fmt.Errorf("failed to update CMI data: %w", err)
		}
	}

	return nil
}

// UpdateScormAttemptScore updates the score fields of a SCORM attempt
func (s *ScormService) UpdateScormAttemptScore(id string, scoreRaw, scoreMin, scoreMax *float64) error {
	attempt, err := s.GetScormAttempt(id)
	if err != nil {
		return err
	}

	attempt.ScoreRaw = scoreRaw
	attempt.ScoreMin = scoreMin
	attempt.ScoreMax = scoreMax
	attempt.UpdatedAt = time.Now()

	if err := s.db.Save(attempt).Error; err != nil {
		return fmt.Errorf("failed to update SCORM attempt score: %w", err)
	}

	return nil
}

// UpdateScormAttemptStatus updates the status fields of a SCORM attempt
func (s *ScormService) UpdateScormAttemptStatus(id string, status, successStatus, completionStatus string) error {
	attempt, err := s.GetScormAttempt(id)
	if err != nil {
		return err
	}

	attempt.Status = status
	attempt.SuccessStatus = successStatus
	attempt.CompletionStatus = completionStatus
	attempt.UpdatedAt = time.Now()

	if err := s.db.Save(attempt).Error; err != nil {
		return fmt.Errorf("failed to update SCORM attempt status: %w", err)
	}

	return nil
}

// UpdateScormAttemptTime updates the time fields of a SCORM attempt
func (s *ScormService) UpdateScormAttemptTime(id string, totalTime string) error {
	attempt, err := s.GetScormAttempt(id)
	if err != nil {
		return err
	}

	attempt.TotalTime = totalTime
	attempt.UpdatedAt = time.Now()

	if err := s.db.Save(attempt).Error; err != nil {
		return fmt.Errorf("failed to update SCORM attempt time: %w", err)
	}

	return nil
}

// GetScormAttemptsByUser gets all SCORM attempts for a specific user
func (s *ScormService) GetScormAttemptsByUser(userID uint) ([]models.ScormAttempt, error) {
	var attempts []models.ScormAttempt
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("failed to get SCORM attempts for user: %w", err)
	}
	return attempts, nil
}

// GetScormAttemptsByActivity gets all SCORM attempts for a specific activity
func (s *ScormService) GetScormAttemptsByActivity(activityID uint) ([]models.ScormAttempt, error) {
	var attempts []models.ScormAttempt
	if err := s.db.Where("activity_id = ?", activityID).Order("created_at DESC").Find(&attempts).Error; err != nil {
		return nil, fmt.Errorf("failed to get SCORM attempts for activity: %w", err)
	}
	return attempts, nil
}

// GetScormCMI retrieves a CMI data element for a SCORM attempt
func (s *ScormService) GetScormCMI(attemptID, element string) (string, error) {
	var cmi models.ScormCMI
	if err := s.db.Where("attempt_id = ? AND element = ?", attemptID, element).First(&cmi).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // Return empty string if not found (SCORM standard)
		}
		return "", fmt.Errorf("failed to get CMI data: %w", err)
	}

	return cmi.Value, nil
}

// CompleteScormAttempt marks a SCORM attempt as completed
func (s *ScormService) CompleteScormAttempt(id string, status string) error {
	attempt, err := s.GetScormAttempt(id)
	if err != nil {
		return err
	}

	attempt.Status = status
	attempt.EndTime = &time.Time{}
	*attempt.EndTime = time.Now()
	attempt.UpdatedAt = time.Now()

	if err := s.db.Save(attempt).Error; err != nil {
		return fmt.Errorf("failed to complete SCORM attempt: %w", err)
	}

	return nil
}
