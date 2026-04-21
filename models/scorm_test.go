package models

import (
	"testing"
	"time"
)

func TestScormActivity(t *testing.T) {
	// Test creating a new ScormActivity
	activity := &ScormActivity{
		Title:       "Test Course",
		Description: "Test Description",
		Version:     "1.2",
		LaunchURL:   "/scorm/test/index.html",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate required fields
	if activity.Title == "" {
		t.Error("Title should not be empty")
	}
	if activity.Version == "" {
		t.Error("Version should not be empty")
	}
	if activity.LaunchURL == "" {
		t.Error("LaunchURL should not be empty")
	}

	// Test table name
	if activity.TableName() != "scorm_activities" {
		t.Errorf("Expected table name 'scorm_activities', got '%s'", activity.TableName())
	}

	t.Logf("ScormActivity created successfully: %+v", activity)
}

func TestScormAttempt(t *testing.T) {
	// Test creating a new ScormAttempt
	now := time.Now()
	attempt := &ScormAttempt{
		ID:         "test_attempt_001",
		ActivityID: 1,
		UserID:     1,
		Version:    "1.2",
		Status:     "incomplete",
		StartTime:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Validate required fields
	if attempt.ID == "" {
		t.Error("ID should not be empty")
	}
	if attempt.ActivityID == 0 {
		t.Error("ActivityID should not be zero")
	}
	if attempt.UserID == 0 {
		t.Error("UserID should not be zero")
	}
	if attempt.Version == "" {
		t.Error("Version should not be empty")
	}
	if attempt.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}

	// Test table name
	if attempt.TableName() != "scorm_attempts" {
		t.Errorf("Expected table name 'scorm_attempts', got '%s'", attempt.TableName())
	}

	t.Logf("ScormAttempt created successfully: %+v", attempt)
}

func TestScormCMI(t *testing.T) {
	// Test creating a new ScormCMI
	now := time.Now()
	cmi := &ScormCMI{
		AttemptID: "test_attempt_001",
		Element:   "cmi.core.lesson_status",
		Value:     "completed",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate required fields
	if cmi.AttemptID == "" {
		t.Error("AttemptID should not be empty")
	}
	if cmi.Element == "" {
		t.Error("Element should not be empty")
	}

	// Test table name
	if cmi.TableName() != "scorm_cmi" {
		t.Errorf("Expected table name 'scorm_cmi', got '%s'", cmi.TableName())
	}

	t.Logf("ScormCMI created successfully: %+v", cmi)
}

func TestScormSession(t *testing.T) {
	// Test creating a new ScormSession
	now := time.Now()
	session := &ScormSession{
		AttemptID: "test_attempt_001",
		SessionID: "test_session_001",
		UserAgent: "Mozilla/5.0 (Test Browser)",
		IPAddress: "127.0.0.1",
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate required fields
	if session.AttemptID == "" {
		t.Error("AttemptID should not be empty")
	}
	if session.SessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if session.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}

	// Test table name
	if session.TableName() != "scorm_sessions" {
		t.Errorf("Expected table name 'scorm_sessions', got '%s'", session.TableName())
	}

	t.Logf("ScormSession created successfully: %+v", session)
}

func TestScormModelsIntegration(t *testing.T) {
	// Test creating related models
	activity := &ScormActivity{
		Title:       "Integration Test Course",
		Description: "Testing model relationships",
		Version:     "2004",
		LaunchURL:   "/scorm/integration/index.html",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	attempt := &ScormAttempt{
		ID:         "integration_attempt_001",
		ActivityID: 1, // This would reference the activity
		UserID:     1,
		Version:    activity.Version,
		Status:     "incomplete",
		StartTime:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	cmi := &ScormCMI{
		AttemptID: attempt.ID,
		Element:   "cmi.completion_status",
		Value:     "incomplete",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session := &ScormSession{
		AttemptID: attempt.ID,
		SessionID: "integration_session_001",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate relationships
	if cmi.AttemptID != attempt.ID {
		t.Error("CMI AttemptID should match Attempt ID")
	}
	if session.AttemptID != attempt.ID {
		t.Error("Session AttemptID should match Attempt ID")
	}
	if attempt.Version != activity.Version {
		t.Error("Attempt Version should match Activity Version")
	}

	t.Log("All SCORM models integrated successfully")
	t.Logf("Activity: %+v", activity)
	t.Logf("Attempt: %+v", attempt)
	t.Logf("CMI: %+v", cmi)
	t.Logf("Session: %+v", session)
}
