package services

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GoogleMeetingService interface {
	CreateMeeting(userID uint, request *CreateGoogleMeetingRequest) (*models.GoogleMeeting, error)
	ListMeetingNotifications(userID uint, limit, offset int) ([]models.MeetingNotification, error)
	MarkMeetingNotificationRead(id uint, userID uint) error
	GetMeeting(id uint) (*models.GoogleMeeting, error)
	UpdateMeeting(id uint, request *UpdateGoogleMeetingRequest) (*models.GoogleMeeting, error)
	DeleteMeeting(id uint) error
	ForceDeleteMeeting(id uint, deletedBy uint) error
	GetUserMeetings(userID uint, page, limit int) ([]*models.GoogleMeeting, error)
	GetCourseMeetings(courseID uint) ([]*models.GoogleMeeting, error)
	GetLessonMeetings(lessonID uint) ([]*models.GoogleMeeting, error)
	GetUpcomingMeetings(userID uint, limit int) ([]*models.GoogleMeeting, error)
	UpdateRecording(id uint, recordingURL, status string, durationMinutes int) error
}

func (s *googleMeetingService) ListMeetingNotifications(userID uint, limit, offset int) ([]models.MeetingNotification, error) {
	return s.notifRepo.ListByUserAndProvider(userID, "google", limit, offset)
}

func (s *googleMeetingService) MarkMeetingNotificationRead(id uint, userID uint) error {
	return s.notifRepo.MarkRead(id, userID)
}

type googleMeetingService struct {
	repo      repositories.GoogleMeetingRepository
	authRepo  repositories.GoogleAuthRepository
	notifRepo repositories.MeetingNotificationRepository
	pubsub    ChatPubSubService
	config    *config.GoogleConfig
}

func (s *googleMeetingService) populateComputed(m *models.GoogleMeeting) {
	if m == nil {
		return
	}
	m.JoinURL = m.MeetURL
	if m.ShortCode != "" {
		m.PublicJoinURL = fmt.Sprintf("/api/meet/%s", m.ShortCode)
	}
}

func (s *googleMeetingService) populateComputedList(items []*models.GoogleMeeting) {
	for _, m := range items {
		s.populateComputed(m)
	}
}

// Request/Response structs
type CreateGoogleMeetingRequest struct {
	Title       string                         `json:"title" binding:"required"`
	Description string                         `json:"description"`
	StartTime   time.Time                      `json:"start_time" binding:"required"`
	EndTime     time.Time                      `json:"end_time" binding:"required"`
	TimeZone    string                         `json:"time_zone"`
	CourseID    *uint                          `json:"course_id"`
	LessonID    *uint                          `json:"lesson_id"`
	Attendees   []models.GoogleMeetingAttendee `json:"attendees"`
}

type UpdateGoogleMeetingRequest struct {
	Title       string                         `json:"title"`
	Description string                         `json:"description"`
	StartTime   time.Time                      `json:"start_time"`
	EndTime     time.Time                      `json:"end_time"`
	Attendees   []models.GoogleMeetingAttendee `json:"attendees"`
}

// Google Calendar API structs
type CalendarEvent struct {
	Summary        string          `json:"summary"`
	Description    string          `json:"description,omitempty"`
	Start          *EventTime      `json:"start"`
	End            *EventTime      `json:"end"`
	Attendees      []*Attendee     `json:"attendees,omitempty"`
	ConferenceData *ConferenceData `json:"conferenceData,omitempty"`
}

type EventTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type Attendee struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
}

type ConferenceData struct {
	CreateRequest *CreateRequest `json:"createRequest"`
}

type CreateRequest struct {
	RequestId             string                 `json:"requestId"`
	ConferenceSolutionKey *ConferenceSolutionKey `json:"conferenceSolutionKey"`
}

type ConferenceSolutionKey struct {
	Type string `json:"type"`
}

type CalendarEventResponse struct {
	ID             string                  `json:"id"`
	Summary        string                  `json:"summary"`
	Description    string                  `json:"description"`
	HtmlLink       string                  `json:"htmlLink"`
	Start          *EventTime              `json:"start"`
	End            *EventTime              `json:"end"`
	ConferenceData *ConferenceDataResponse `json:"conferenceData,omitempty"`
}

type ConferenceDataResponse struct {
	EntryPoints []*EntryPoint `json:"entryPoints,omitempty"`
}

type EntryPoint struct {
	EntryPointType string `json:"entryPointType"`
	Uri            string `json:"uri"`
	Label          string `json:"label,omitempty"`
}

func NewGoogleMeetingService() GoogleMeetingService {
	return &googleMeetingService{
		repo:      repositories.NewGoogleMeetingRepository(),
		authRepo:  repositories.NewGoogleAuthRepository(),
		notifRepo: repositories.NewMeetingNotificationRepository(),
		pubsub:    NewChatPubSubService(),
		config:    config.GetGoogleConfig(),
	}
}

func (s *googleMeetingService) generateShortCode() string {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateMeeting creates a new Google Meet meeting via Calendar API
func (s *googleMeetingService) CreateMeeting(userID uint, request *CreateGoogleMeetingRequest) (*models.GoogleMeeting, error) {
	config.Log.Infof("🎥 Creating Google Meet for user %d: %s", userID, request.Title)

	// Default timezone
	if request.TimeZone == "" {
		request.TimeZone = "UTC"
	}

	// Get user's Google account
	account, err := s.authRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not connected to Google: %v", err)
	}

	// Check if token needs refresh
	if time.Now().After(account.ExpiresAt) {
		authService := NewGoogleAuthService()
		if err := authService.RefreshToken(userID); err != nil {
			return nil, fmt.Errorf("failed to refresh Google token: %v", err)
		}
		// Reload account with fresh token
		account, _ = s.authRepo.GetByUserID(userID)
	}

	// Create calendar event with Meet link
	eventResp, err := s.createCalendarEvent(account.AccessToken, request)
	if err != nil {
		return nil, err
	}

	// Extract Meet URL from response
	meetURL := ""
	if eventResp.ConferenceData != nil && len(eventResp.ConferenceData.EntryPoints) > 0 {
		for _, entryPoint := range eventResp.ConferenceData.EntryPoints {
			if entryPoint.EntryPointType == "video" {
				meetURL = entryPoint.Uri
				break
			}
		}
	}
	if meetURL == "" {
		return nil, fmt.Errorf("failed to generate Google Meet URL")
	}

	// Convert attendees to JSON
	attendeesJSON := ""
	if len(request.Attendees) > 0 {
		attendeesBytes, _ := json.Marshal(request.Attendees)
		attendeesJSON = string(attendeesBytes)
	}

	// Save to database
	shortCode := s.generateShortCode()
	meeting := &models.GoogleMeeting{
		CalendarEventID: eventResp.ID,
		UserID:          userID,
		CourseID:        request.CourseID,
		LessonID:        request.LessonID,
		Title:           request.Title,
		Description:     request.Description,
		StartTime:       request.StartTime,
		EndTime:         request.EndTime,
		MeetURL:         meetURL,
		CalendarURL:     eventResp.HtmlLink,
		TimeZone:        request.TimeZone,
		Status:          models.GoogleMeetingStatusScheduled,
		Attendees:       attendeesJSON,
		ShortCode:       shortCode,
	}

	if err := s.repo.Create(meeting); err != nil {
		return nil, fmt.Errorf("failed to save meeting to database: %v", err)
	}

	// Re-fetch from database to ensure all fields are properly populated (same as Microsoft)
	savedMeeting, err := s.repo.GetByID(meeting.ID)
	if err != nil {
		config.Log.Warnf("🔍 CreateMeeting - Failed to fetch saved meeting: %v", err)
		// Continue with in-memory meeting if fetch fails
		savedMeeting = meeting
	} else {
		config.Log.Infof("🔍 CreateMeeting - Fetched from DB: ID=%d, LessonID=%v, CourseID=%v", savedMeeting.ID, savedMeeting.LessonID, savedMeeting.CourseID)
	}

	s.populateComputed(savedMeeting)
	go s.createMeetingNotifications(savedMeeting)
	config.Log.Infof("✅ Created Google Meet successfully: %s", meetURL)
	return savedMeeting, nil
}

func (s *googleMeetingService) createMeetingNotifications(meeting *models.GoogleMeeting) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Warnf("meeting notification panic: %v", r)
		}
	}()

	if meeting == nil {
		return
	}

	config.Log.Infof("meeting notif: start meeting_id=%d course_id=%v lesson_id=%v", meeting.ID, meeting.CourseID, meeting.LessonID)

	courseID := meeting.CourseID
	if courseID == nil && meeting.LessonID != nil {
		var cid uint
		if err := db.MasterDB.Table("lessons").Select("course_id").Where("id = ?", *meeting.LessonID).Scan(&cid).Error; err == nil && cid > 0 {
			courseID = &cid
		}
	}

	if courseID == nil {
		return
	}

	var userIDs []uint
	if err := db.MasterDB.Table("user_courses").Where("course_id = ?", *courseID).Pluck("user_id", &userIDs).Error; err != nil {
		config.Log.Warnf("meeting notif: failed to load users for course %d: %v", *courseID, err)
		return
	}
	if len(userIDs) == 0 {
		config.Log.Warnf("meeting notif: no users found for course %d", *courseID)
		return
	}

	config.Log.Infof("meeting notif: course %d -> %d users", *courseID, len(userIDs))

	notifs := make([]models.MeetingNotification, 0, len(userIDs))
	for _, uid := range userIDs {
		notifs = append(notifs, models.MeetingNotification{
			MeetingID:   meeting.ID,
			UserID:      uid,
			CourseID:    courseID,
			LessonID:    meeting.LessonID,
			Provider:    "google",
			Title:       meeting.Title,
			Description: meeting.Description,
		})
	}

	if err := s.notifRepo.BulkCreate(notifs); err != nil {
		config.Log.Warnf("meeting notif: bulk insert failed: %v", err)
	} else {
		config.Log.Infof("meeting notif: inserted %d notifications for meeting %d", len(notifs), meeting.ID)
	}

	for _, n := range notifs {
		payload := map[string]interface{}{
			"type":         "meeting_notification",
			"notification": n,
		}
		if s.pubsub != nil {
			if err := s.pubsub.NotifyUser(uint64(n.UserID), "meeting_notification", payload); err != nil {
				config.Log.Warnf("meeting notif: publish user %d failed: %v", n.UserID, err)
			} else {
				config.Log.Infof("meeting notif: published to user %d for meeting %d", n.UserID, meeting.ID)
			}
		}
	}
}

// createCalendarEvent creates event in Google Calendar with Meet
func (s *googleMeetingService) createCalendarEvent(accessToken string, request *CreateGoogleMeetingRequest) (*CalendarEventResponse, error) {
	// Prepare attendees
	var attendees []*Attendee
	for _, att := range request.Attendees {
		attendees = append(attendees, &Attendee{
			Email:       att.Email,
			DisplayName: att.DisplayName,
		})
	}

	// Create calendar event
	event := &CalendarEvent{
		Summary:     request.Title,
		Description: request.Description,
		Start: &EventTime{
			DateTime: request.StartTime.Format(time.RFC3339),
			TimeZone: request.TimeZone,
		},
		End: &EventTime{
			DateTime: request.EndTime.Format(time.RFC3339),
			TimeZone: request.TimeZone,
		},
		Attendees: attendees,
		ConferenceData: &ConferenceData{
			CreateRequest: &CreateRequest{
				RequestId: fmt.Sprintf("meet-%d-%d", request.StartTime.Unix(), time.Now().UnixNano()),
				ConferenceSolutionKey: &ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		},
	}

	// Convert to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %v", err)
	}

	config.Log.Infof("📅 Creating Calendar event: %s", string(eventJSON))

	// Make API request with notification settings
	apiURL := s.config.GetCalendarAPI() + "/calendars/primary/events?conferenceDataVersion=1&sendNotifications=true&sendUpdates=all"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(eventJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	config.Log.Infof("📡 Google Calendar API response status: %d", resp.StatusCode)
	config.Log.Infof("📡 Google Calendar API response body: %s", bodyString)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google Calendar API error: status %d, body: %s", resp.StatusCode, bodyString)
	}

	var eventResp CalendarEventResponse
	if err := json.Unmarshal(bodyBytes, &eventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	config.Log.Infof("✅ Calendar event created successfully - ID: %s, Meet URL: %s", eventResp.ID, eventResp.ConferenceData)
	return &eventResp, nil
}

// GetMeeting retrieves meeting by ID
func (s *googleMeetingService) GetMeeting(id uint) (*models.GoogleMeeting, error) {
	m, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	s.populateComputed(m)
	return m, nil
}

// UpdateMeeting updates an existing meeting
func (s *googleMeetingService) UpdateMeeting(id uint, request *UpdateGoogleMeetingRequest) (*models.GoogleMeeting, error) {
	meeting, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %v", err)
	}

	// Update fields
	if request.Title != "" {
		meeting.Title = request.Title
	}
	if request.Description != "" {
		meeting.Description = request.Description
	}
	if !request.StartTime.IsZero() {
		meeting.StartTime = request.StartTime
	}
	if !request.EndTime.IsZero() {
		meeting.EndTime = request.EndTime
	}
	if len(request.Attendees) > 0 {
		attendeesBytes, _ := json.Marshal(request.Attendees)
		meeting.Attendees = string(attendeesBytes)
	}

	if err := s.repo.Update(meeting); err != nil {
		return nil, fmt.Errorf("failed to update meeting: %v", err)
	}
	return meeting, nil
}

// DeleteMeeting deletes a meeting
func (s *googleMeetingService) DeleteMeeting(id uint) error {
	return s.repo.Delete(id)
}

// ForceDeleteMeeting permanently removes a meeting (bypasses soft delete) with audit
func (s *googleMeetingService) ForceDeleteMeeting(id uint, deletedBy uint) error {
	_ = s.repo.LogForceDelete(id, deletedBy) // best-effort audit
	return s.repo.ForceDelete(id)
}

// GetUserMeetings retrieves meetings for a user with pagination
func (s *googleMeetingService) GetUserMeetings(userID uint, page, limit int) ([]*models.GoogleMeeting, error) {
	offset := (page - 1) * limit
	items, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if limit > 0 && page > 0 {
		start := offset
		if start > len(items) {
			start = len(items)
		}
		end := start + limit
		if end > len(items) {
			end = len(items)
		}
		items = items[start:end]
	}

	s.populateComputedList(items)
	return items, nil
}

// GetCourseMeetings retrieves meetings for a course
func (s *googleMeetingService) GetCourseMeetings(courseID uint) ([]*models.GoogleMeeting, error) {
	items, err := s.repo.GetByCourseID(courseID)
	if err == nil {
		s.populateComputedList(items)
	}
	return items, err
}

// GetLessonMeetings retrieves meetings for a lesson
func (s *googleMeetingService) GetLessonMeetings(lessonID uint) ([]*models.GoogleMeeting, error) {
	items, err := s.repo.GetByLessonID(lessonID)
	if err == nil {
		s.populateComputedList(items)
	}
	return items, err
}

// GetUpcomingMeetings retrieves upcoming meetings for a user
func (s *googleMeetingService) GetUpcomingMeetings(userID uint, limit int) ([]*models.GoogleMeeting, error) {
	return s.repo.GetUpcoming(userID, limit)
}

func (s *googleMeetingService) UpdateRecording(id uint, recordingURL, status string, durationMinutes int) error {
	return s.repo.UpdateRecording(id, recordingURL, status, durationMinutes)
}
