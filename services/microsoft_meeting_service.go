package services

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories"
)

type MeetingPermissionInfo struct {
	HasOnlineMeetingPermission bool     `json:"has_online_meeting_permission"`
	HasCalendarPermission      bool     `json:"has_calendar_permission"`
	UserInfo                   UserInfo `json:"user_info"`
	Groups                     []Group  `json:"groups"`
	Error                      string   `json:"error,omitempty"`
}

type UserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	JobTitle    string `json:"job_title"`
}

type Group struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type MicrosoftMeetingService interface {
	CheckMeetingPermissions(userID uint) (*MeetingPermissionInfo, error)
	CreateMeeting(userID uint, request *CreateMicrosoftMeetingRequest) (*models.MicrosoftMeeting, error)
	GetMeeting(id uint) (*models.MicrosoftMeeting, error)
	UpdateMeeting(id uint, request *UpdateMicrosoftMeetingRequest) (*models.MicrosoftMeeting, error)
	DeleteMeeting(id uint) error
	GetUserMeetings(userID uint, page, limit int) ([]*models.MicrosoftMeeting, error)
	GetCourseMeetings(courseID uint) ([]*models.MicrosoftMeeting, error)
	GetLessonMeetings(lessonID uint) ([]*models.MicrosoftMeeting, error)
	GetUpcomingMeetings(userID uint, limit int) ([]*models.MicrosoftMeeting, error)
	GetByShortCode(code string) (*models.MicrosoftMeeting, error)
	FetchRecording(meetingID uint) (string, error)
	SetRecordingShareURL(meetingID uint, shareURL string) error
}

type microsoftMeetingService struct {
	repo      repositories.MicrosoftMeetingRepository
	authRepo  repositories.MicrosoftAuthRepository
	notifRepo repositories.MeetingNotificationRepository
	pubsub    ChatPubSubService
	config    *config.MicrosoftConfig
}

// Request/Response structs
type CreateMicrosoftMeetingRequest struct {
	Title       string                            `json:"title" binding:"required"`
	Description string                            `json:"description"`
	StartTime   time.Time                         `json:"start_time" binding:"required"`
	EndTime     time.Time                         `json:"end_time" binding:"required"`
	TimeZone    string                            `json:"time_zone"`
	CourseID    *uint                             `json:"course_id"`
	LessonID    *uint                             `json:"lesson_id"`
	ClassID     *uint                             `json:"class_id"`
	Attendees   []models.MicrosoftMeetingAttendee `json:"attendees"`
}

type UpdateMicrosoftMeetingRequest struct {
	Title       string                            `json:"title"`
	Description string                            `json:"description"`
	StartTime   time.Time                         `json:"start_time"`
	EndTime     time.Time                         `json:"end_time"`
	Attendees   []models.MicrosoftMeetingAttendee `json:"attendees"`
}

// Microsoft Graph API structs
type GraphEvent struct {
	Subject               string             `json:"subject"`
	Body                  *GraphItemBody     `json:"body,omitempty"`
	Start                 *GraphDateTimeZone `json:"start"`
	End                   *GraphDateTimeZone `json:"end"`
	Attendees             []*GraphAttendee   `json:"attendees,omitempty"`
	IsOnlineMeeting       bool               `json:"isOnlineMeeting"`
	OnlineMeetingProvider string             `json:"onlineMeetingProvider"`
}

type GraphItemBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type GraphDateTimeZone struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type GraphAttendee struct {
	EmailAddress *GraphEmailAddress `json:"emailAddress"`
	Type         string             `json:"type"`
}

type GraphEmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type GraphEventResponse struct {
	ID            string              `json:"id"`
	Subject       string              `json:"subject"`
	Body          *GraphItemBody      `json:"body"`
	WebLink       string              `json:"webLink"`
	Start         *GraphDateTimeZone  `json:"start"`
	End           *GraphDateTimeZone  `json:"end"`
	OnlineMeeting *GraphOnlineMeeting `json:"onlineMeeting"`
}

type GraphOnlineMeeting struct {
	JoinURL         string   `json:"joinUrl"`
	ConferenceID    string   `json:"conferenceId"`
	TollNumber      string   `json:"tollNumber,omitempty"`
	TollFreeNumbers []string `json:"tollFreeNumbers,omitempty"`
}

func NewMicrosoftMeetingService() MicrosoftMeetingService {
	return &microsoftMeetingService{
		repo:      repositories.NewMicrosoftMeetingRepository(),
		authRepo:  repositories.NewMicrosoftAuthRepository(),
		notifRepo: repositories.NewMeetingNotificationRepository(),
		pubsub:    NewChatPubSubService(),
		config:    config.GetMicrosoftConfig(),
	}
}

// CreateMeeting creates a new Microsoft Teams meeting via Graph API
func (s *microsoftMeetingService) CreateMeeting(userID uint, request *CreateMicrosoftMeetingRequest) (*models.MicrosoftMeeting, error) {
	defer func() {
		if r := recover(); r != nil {
			config.Log.Errorf("💥 Panic in CreateMeeting: %v", r)
		}
	}()

	config.Log.Infof("🎥 Creating Microsoft Teams meeting for user %d: %s", userID, request.Title)

	// Get user's Microsoft account
	config.Log.Infof("🔍 Getting Microsoft account for user %d", userID)
	account, err := s.authRepo.GetByUserID(userID)
	if err != nil {
		config.Log.Errorf("❌ Failed to get Microsoft account: %v", err)
		return nil, fmt.Errorf("user not connected to Microsoft: %v", err)
	}

	config.Log.Infof("✅ Found Microsoft account: %s", account.AccountEmail)

	// Check if token needs refresh
	if time.Now().After(account.ExpiresAt) {
		authService := NewMicrosoftAuthService(s.authRepo)
		if err := authService.RefreshToken(account); err != nil {
			return nil, fmt.Errorf("failed to refresh Microsoft token: %v", err)
		}
		// Reload account with fresh token
		account, _ = s.authRepo.GetByUserID(userID)
	}

	// Try creating OnlineMeetings first to test permissions
	config.Log.Infof("🔧 Testing OnlineMeetings API permissions...")
	teamsURL, onlineMeetingID, conferenceCode, err := s.createOnlineMeeting(account.AccessToken, request)

	var (
		eventResp   *GraphEventResponse // Declare here for both paths
		onlineErr   error
		calendarErr error
	)

	if err != nil {
		onlineErr = err
		config.Log.Errorf("❌ OnlineMeetings API failed: %v", onlineErr)
		// Fallback to Calendar Events API
		eventResp, err = s.createGraphEvent(account.AccessToken, request)
		if err != nil {
			calendarErr = err
			return nil, fmt.Errorf("onlineMeetings failed: %v; calendarEvents failed: %v", onlineErr, calendarErr)
		}

		// Extract Teams URL from Calendar response (likely null)
		config.Log.Infof("🔍 DEBUG: Checking eventResp.OnlineMeeting...")
		if eventResp.OnlineMeeting != nil {
			teamsURL = eventResp.OnlineMeeting.JoinURL
			onlineMeetingID = eventResp.OnlineMeeting.ConferenceID // fallback if createOnlineMeeting failed
			conferenceCode = eventResp.OnlineMeeting.ConferenceID
			config.Log.Infof("🔍 DEBUG: Got Teams URL from Calendar: %s", teamsURL)
		} else {
			config.Log.Infof("🔍 DEBUG: eventResp.OnlineMeeting is nil")
		}

		// If OnlineMeetings failed but Calendar event exists, allow calendar-only fallback
		if teamsURL == "" && eventResp != nil {
			teamsURL = eventResp.WebLink   // use calendar web link as join URL
			onlineMeetingID = eventResp.ID // use event ID as unique identifier
			conferenceCode = eventResp.ID
			config.Log.Infof("📎 Fallback to calendar-only event: joinUrl=%s, eventId=%s", teamsURL, onlineMeetingID)
		}

		// Still no URL? return detailed error
		if teamsURL == "" {
			return nil, fmt.Errorf("Teams join URL missing. OnlineMeetings failed: %v; calendarEvent id=%s webLink=%s onlineMeeting=nil",
				onlineErr,
				s.safeString(eventResp.ID),
				s.safeString(eventResp.WebLink),
			)
		}
	} else {
		config.Log.Infof("✅ OnlineMeetings API success: %s", teamsURL)
		// Create calendar event for OnlineMeetings success case
		eventResp, err = s.createGraphEvent(account.AccessToken, request)
		if err != nil {
			config.Log.Warnf("⚠️ Failed to create calendar event but OnlineMeeting exists: %v", err)
			eventResp = nil // Ensure it's nil on failure
		}
	}

	// Convert attendees to JSON
	attendeesJSON := ""
	if len(request.Attendees) > 0 {
		attendeesBytes, _ := json.Marshal(request.Attendees)
		attendeesJSON = string(attendeesBytes)
	}

	// Final validation
	if teamsURL == "" {
		return nil, fmt.Errorf("failed to create Teams meeting URL - check OnlineMeetings.ReadWrite permissions (last online error: %v)", onlineErr)
	}

	if onlineMeetingID == "" {
		onlineMeetingID = conferenceCode
	}

	// Prepare calendar info (safe access)
	calendarEventID := ""
	if eventResp != nil {
		calendarEventID = eventResp.ID
	}

	guestJoinURL := teamsURL

	shortCode := s.generateShortCode()

	conferenceCode = s.normalizeConferenceID(conferenceCode, teamsURL, calendarEventID, eventResp)

	// Save to database
	// IMPORTANT: Store times as-is (don't convert to UTC).
	// Times from request are already in the meeting's local timezone
	// The database will store them as naive timestamps without timezone info
	meeting := &models.MicrosoftMeeting{
		OnlineMeetingID: onlineMeetingID,
		CalendarEventID: calendarEventID,
		UserID:          userID,
		CourseID:        request.CourseID,
		LessonID:        request.LessonID,
		ClassID:         request.ClassID,
		Title:           request.Title,
		Description:     request.Description,
		StartTime:       request.StartTime, // Keep in local timezone, don't convert to UTC
		EndTime:         request.EndTime,   // Keep in local timezone, don't convert to UTC
		JoinURL:         teamsURL,
		JoinWebURL:      guestJoinURL,
		ConferenceID:    conferenceCode,
		ShortCode:       shortCode,
		TimeZone:        request.TimeZone,
		Status:          models.MicrosoftMeetingStatusScheduled,
		Attendees:       attendeesJSON,
	}

	// Use atomic UPSERT to prevent race condition with duplicate online_meeting_id
	savedMeeting, err := s.repo.CreateOrGetByOnlineMeetingID(meeting)
	if err != nil {
		return nil, fmt.Errorf("failed to save meeting to database: %v", err)
	}

	// If this is an existing meeting (different ID), log it
	if savedMeeting.ID != 0 && savedMeeting.ID != meeting.ID {
		config.Log.Warnf("⚠️ Duplicate online_meeting_id detected, returning existing meeting id=%d", savedMeeting.ID)
	}

	// Populate local display times
	s.attachLocalTimes(savedMeeting)

	// Create notifications for course members (sync, not async to avoid race conditions)
	s.createMeetingNotifications(savedMeeting)

	config.Log.Infof("✅ Created Microsoft Teams meeting successfully: %s", teamsURL)
	return savedMeeting, nil
}

func (s *microsoftMeetingService) createMeetingNotifications(meeting *models.MicrosoftMeeting) {
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
		config.Log.Warnf("meeting notif: no course_id found for meeting %d", meeting.ID)
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
			Provider:    "microsoft",
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

// createOnlineMeeting creates Teams meeting via OnlineMeetings API
func (s *microsoftMeetingService) createOnlineMeeting(accessToken string, request *CreateMicrosoftMeetingRequest) (string, string, string, error) {
	onlineMeetingReq := map[string]interface{}{
		"subject":              request.Title,
		"startDateTime":        request.StartTime.Format(time.RFC3339),
		"endDateTime":          request.EndTime.Format(time.RFC3339),
		"allowedPresenters":    "everyone",
		"isAnonymousStartTime": true,
		"accessLevel":          "everyone",
	}

	reqJSON, _ := json.Marshal(onlineMeetingReq)
	config.Log.Infof("📞 Creating OnlineMeeting: %s", string(reqJSON))

	apiURL := s.config.GetGraphAPI() + "/me/onlineMeetings"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create OnlineMeeting request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to call OnlineMeetings API: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	config.Log.Infof("📞 OnlineMeetings API response status: %d", resp.StatusCode)
	config.Log.Infof("📞 OnlineMeetings API response: %s", bodyString)

	if resp.StatusCode != 201 {
		return "", "", "", fmt.Errorf("OnlineMeetings API error: status %d, body: %s", resp.StatusCode, bodyString)
	}

	var meetingResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &meetingResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse OnlineMeeting response: %v", err)
	}

	joinUrl := ""
	if joinWebUrl, ok := meetingResp["joinWebUrl"].(string); ok {
		joinUrl = joinWebUrl
	}

	meetingID := ""
	if id, ok := meetingResp["id"].(string); ok {
		meetingID = id
	}
	conferenceCode := ""
	if confId, ok := meetingResp["conferenceId"].(string); ok {
		conferenceCode = confId
	}
	if threadId := extractThreadID(joinUrl, meetingID); threadId != "" {
		conferenceCode = threadId
	}
	if meetingID == "" {
		meetingID = conferenceCode
	}

	return joinUrl, meetingID, conferenceCode, nil
}

// decodeMeetingID converts Graph onlineMeeting id (base64url) to threadId string if possible
func decodeMeetingID(encoded string) string {
	if encoded == "" {
		return ""
	}
	// base64url padding fix
	str := encoded
	if m := len(str) % 4; m != 0 {
		str += strings.Repeat("=", 4-m)
	}
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return ""
	}
	decoded := string(data)
	if thread := findThreadID(decoded); thread != "" {
		return thread
	}
	return ""
}

// extractThreadID tries to get threadId from joinUrl query param or decoded meetingID
func extractThreadID(joinUrl, meetingID string) string {
	if thread := findThreadID(joinUrl); thread != "" {
		return thread
	}
	if thread := decodeMeetingID(meetingID); thread != "" {
		return thread
	}
	return ""
}

// findThreadID finds 19:meeting_...@thread.v2 in a string
func findThreadID(s string) string {
	re := regexp.MustCompile(`19:meeting_[A-Za-z0-9_-]+@thread\.v2`)
	if match := re.FindString(s); match != "" {
		return match
	}
	// some joinUrl encode threadId in query param threadId=...
	if u, err := url.Parse(s); err == nil {
		if val := u.Query().Get("threadId"); val != "" {
			if decodedVal, err := url.QueryUnescape(val); err == nil {
				if match := re.FindString(decodedVal); match != "" {
					return match
				}
			}
		}
	}
	return ""
}

// createGraphEvent creates event in Microsoft Graph with Teams meeting
func (s *microsoftMeetingService) createGraphEvent(accessToken string, request *CreateMicrosoftMeetingRequest) (*GraphEventResponse, error) {
	// Prepare attendees
	var attendees []*GraphAttendee
	for _, att := range request.Attendees {
		attendees = append(attendees, &GraphAttendee{
			EmailAddress: &GraphEmailAddress{
				Address: att.Email,
				Name:    att.DisplayName,
			},
			Type: "required",
		})
	}

	// Set default timezone if not provided
	timezone := request.TimeZone
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		config.Log.Warnf("⚠️ Invalid timezone %s, fallback to UTC", timezone)
		loc = time.UTC
		timezone = "UTC"
	}

	startLocal := request.StartTime.In(loc)
	endLocal := request.EndTime.In(loc)

	// Create Graph event
	event := &GraphEvent{
		Subject: request.Title,
		Body: &GraphItemBody{
			ContentType: "HTML",
			Content:     request.Description,
		},
		Start: &GraphDateTimeZone{
			DateTime: startLocal.Format("2006-01-02T15:04:05.000"),
			TimeZone: timezone,
		},
		End: &GraphDateTimeZone{
			DateTime: endLocal.Format("2006-01-02T15:04:05.000"),
			TimeZone: timezone,
		},
		Attendees:             attendees,
		IsOnlineMeeting:       true,
		OnlineMeetingProvider: "teamsForBusiness",
	}

	// Convert to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %v", err)
	}

	config.Log.Infof("📅 Creating Graph event: %s", string(eventJSON))

	// Make API request
	apiURL := s.config.GetGraphAPI() + "/me/events"
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

	config.Log.Infof("📡 Microsoft Graph API response status: %d", resp.StatusCode)
	config.Log.Infof("📡 Microsoft Graph API response body: %s", bodyString)

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("microsoft Graph API error: status %d, body: %s", resp.StatusCode, bodyString)
	}

	var eventResp GraphEventResponse
	if err := json.Unmarshal(bodyBytes, &eventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Safe logging
	teamsURL := "null"
	if eventResp.OnlineMeeting != nil {
		teamsURL = eventResp.OnlineMeeting.JoinURL
	}
	config.Log.Infof("✅ Graph event created successfully - ID: %s, Teams URL: %s", eventResp.ID, teamsURL)
	return &eventResp, nil
}

// GetMeeting retrieves meeting by ID
func (s *microsoftMeetingService) GetMeeting(id uint) (*models.MicrosoftMeeting, error) {
	m, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	s.attachLocalTimes(m)
	return m, nil
}

// UpdateMeeting updates an existing meeting
func (s *microsoftMeetingService) UpdateMeeting(id uint, request *UpdateMicrosoftMeetingRequest) (*models.MicrosoftMeeting, error) {
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

	// Attach local display times
	s.attachLocalTimes(meeting)

	return meeting, nil
}

// DeleteMeeting deletes a meeting
func (s *microsoftMeetingService) DeleteMeeting(id uint) error {
	// Xóa meeting từ database
	return s.repo.Delete(id)
}

// CheckMeetingPermissions kiểm tra quyền tạo meeting của user
func (s *microsoftMeetingService) CheckMeetingPermissions(userID uint) (*MeetingPermissionInfo, error) {
	// Lấy Microsoft account
	account, err := s.authRepo.GetByUserID(userID)
	if err != nil {
		return &MeetingPermissionInfo{
			Error: "Microsoft account not found",
		}, err
	}

	permissionInfo := &MeetingPermissionInfo{}

	// Lấy thông tin user
	userInfo, err := s.getUserInfo(account.AccessToken)
	if err == nil {
		permissionInfo.UserInfo = *userInfo
	}

	// Test OnlineMeetings permission
	hasOnlinePermission := s.testOnlineMeetingPermission(account.AccessToken)
	permissionInfo.HasOnlineMeetingPermission = hasOnlinePermission

	// Test Calendar permission
	hasCalendarPermission := s.testCalendarPermission(account.AccessToken)
	permissionInfo.HasCalendarPermission = hasCalendarPermission

	// Lấy groups của user
	groups, err := s.getUserGroups(account.AccessToken)
	if err == nil {
		permissionInfo.Groups = groups
	}

	return permissionInfo, nil
}

func (s *microsoftMeetingService) getUserInfo(accessToken string) (*UserInfo, error) {
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API call failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	userInfo := &UserInfo{
		ID:          getStringValue(result, "id"),
		DisplayName: getStringValue(result, "displayName"),
		Email:       getStringValue(result, "mail"),
		JobTitle:    getStringValue(result, "jobTitle"),
	}

	return userInfo, nil
}

func (s *microsoftMeetingService) testOnlineMeetingPermission(accessToken string) bool {
	// Test bằng cách gọi API OnlineMeetings
	reqBody := map[string]interface{}{
		"subject": "Permission Test",
	}
	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://graph.microsoft.com/v1.0/me/onlineMeetings", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Nếu không có lỗi 403 thì có permission
	return resp.StatusCode != 403
}

func (s *microsoftMeetingService) testCalendarPermission(accessToken string) bool {
	// Test bằng cách lấy calendar events
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/events?$top=1", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (s *microsoftMeetingService) getUserGroups(accessToken string) ([]Group, error) {
	req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me/memberOf", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API call failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Parse groups
	var groups []Group
	if value, ok := result["value"].([]interface{}); ok {
		for _, v := range value {
			if m, ok := v.(map[string]interface{}); ok {
				groups = append(groups, Group{
					ID:          s.safeString(getStringValue(m, "id")),
					DisplayName: s.safeString(getStringValue(m, "displayName")),
				})
			}
		}
	}

	return groups, nil
}

func (s *microsoftMeetingService) safeString(val string) string {
	if val == "" {
		return "null"
	}
	return val
}

func getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func (s *microsoftMeetingService) attachLocalTimes(m *models.MicrosoftMeeting) {
	if m == nil {
		return
	}
	if m.RecordingURL != "" && m.RecordingStatus == "none" {
		m.RecordingStatus = "available"
	}

	// StartTime and EndTime are already stored in local timezone (not UTC)
	// Just format them with the proper timezone label
	start := m.StartTime
	end := m.EndTime

	// Format: "HH:mm DD/MM/YYYY (TimeZone)"
	// e.g., "15:04 02/01/2006 (Asia/Ho_Chi_Minh)"
	tz := m.TimeZone
	if tz == "" {
		tz = "UTC"
	}
	m.StartTimeLocal = start.Format("15:04 02/01/2006") + " (" + tz + ")"
	m.EndTimeLocal = end.Format("15:04 02/01/2006") + " (" + tz + ")"
}

// GetUserMeetings retrieves meetings for a user with pagination
func (s *microsoftMeetingService) GetUserMeetings(userID uint, page, limit int) ([]*models.MicrosoftMeeting, error) {
	offset := (page - 1) * limit
	meetings, err := s.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, m := range meetings {
		s.attachLocalTimes(m)
	}
	return meetings, nil
}

// GetCourseMeetings retrieves meetings for a course
func (s *microsoftMeetingService) GetCourseMeetings(courseID uint) ([]*models.MicrosoftMeeting, error) {
	meetings, err := s.repo.GetByCourseID(courseID)
	if err != nil {
		return nil, err
	}
	for _, m := range meetings {
		s.attachLocalTimes(m)
	}
	return meetings, nil
}

// GetLessonMeetings retrieves meetings for a lesson
func (s *microsoftMeetingService) GetLessonMeetings(lessonID uint) ([]*models.MicrosoftMeeting, error) {
	meetings, err := s.repo.GetByLessonID(lessonID)
	if err != nil {
		return nil, err
	}
	for _, m := range meetings {
		s.attachLocalTimes(m)
	}
	return meetings, nil
}

// GetUpcomingMeetings retrieves upcoming meetings for a user
func (s *microsoftMeetingService) GetUpcomingMeetings(userID uint, limit int) ([]*models.MicrosoftMeeting, error) {
	if limit == 0 {
		limit = 5
	}
	meetings, err := s.repo.GetUpcoming(userID, limit)
	if err != nil {
		return nil, err
	}
	for _, m := range meetings {
		s.attachLocalTimes(m)
	}
	return meetings, nil
}

func (s *microsoftMeetingService) GetByShortCode(code string) (*models.MicrosoftMeeting, error) {
	m, err := s.repo.GetByShortCode(code)
	if err != nil {
		return nil, err
	}
	s.attachLocalTimes(m)
	return m, nil
}

func (s *microsoftMeetingService) GetClassMeetings(classID uint) ([]*models.MicrosoftMeeting, error) {
	return s.repo.GetByClassID(classID)
}

func (s *microsoftMeetingService) GetMeetingByShortCode(shortCode string) (*models.MicrosoftMeeting, error) {
	m, err := s.repo.GetByShortCode(shortCode)
	if err != nil {
		return nil, err
	}
	s.attachLocalTimes(m)
	return m, nil
}

func (s *microsoftMeetingService) UpdateRecording(meetingID uint, recordingURL, status string) error {
	return s.repo.UpdateRecording(meetingID, recordingURL, status)
}

func (s *microsoftMeetingService) SetRecordingShareURL(meetingID uint, shareURL string) error {
	return s.repo.UpdateRecordingShareURL(meetingID, shareURL)
}

func (s *microsoftMeetingService) FetchRecording(meetingID uint) (string, error) {
	meeting, err := s.repo.GetByID(meetingID)
	if err != nil {
		return "", fmt.Errorf("meeting not found: %v", err)
	}

	if meeting.RecordingURL != "" {
		return meeting.RecordingURL, nil
	}

	if publicFolder := os.Getenv("NEXT_PUBLIC_PUBLIC_MS_RECORD_FOLDER"); publicFolder != "" {
		return publicFolder, nil
	}

	return "", fmt.Errorf("Recording not available")
}

func (s *microsoftMeetingService) searchAndLinkRecording(accessToken string, meeting *models.MicrosoftMeeting) (string, error) {
	defaultShareURL := ""
	queries := []string{"Meeting Recording"}
	if meeting.Title != "" {
		queries = append(queries, meeting.Title)
	}
	if meeting.ShortCode != "" {
		queries = append(queries, meeting.ShortCode)
	}
	if meeting.OnlineMeetingID != "" {
		queries = append(queries, meeting.OnlineMeetingID)
	}
	if meeting.ConferenceID != "" {
		queries = append(queries, meeting.ConferenceID)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	meetingEnd := meeting.EndTime
	windowStart := meetingEnd.Add(-14 * 24 * time.Hour)
	resultURL := ""

	for _, q := range queries {
		searchURL := fmt.Sprintf("%s/me/drive/root/search(q='%s')", s.config.GetGraphAPI(), url.QueryEscape(q))
		config.Log.Infof("🔍 Searching recording with query: %s", q)
		req, _ := http.NewRequest("GET", searchURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return
			}

			items, _ := result["value"].([]interface{})
			if len(items) == 0 {
				return
			}

			var (
				bestID       string
				bestWebURL   string
				bestDownload string
			)
			for _, it := range items {
				m, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				id, _ := m["id"].(string)
				name, _ := m["name"].(string)
				webUrl, _ := m["webUrl"].(string)
				downloadUrl, _ := m["@microsoft.graph.downloadUrl"].(string)
				if id == "" || name == "" {
					continue
				}
				lmStr, _ := m["lastModifiedDateTime"].(string)
				lm, _ := time.Parse(time.RFC3339, lmStr)
				if !lm.IsZero() && (lm.Before(windowStart) || lm.After(meetingEnd.Add(72*time.Hour))) {
					continue
				}
				bestID = id
				bestWebURL = webUrl
				bestDownload = downloadUrl
				break
			}

			if bestID == "" {
				return
			}
			linkURL := fmt.Sprintf("%s/me/drive/items/%s/createLink", s.config.GetGraphAPI(), bestID)
			body := map[string]string{"type": "view", "scope": "anonymous"}
			bodyBytes, _ := json.Marshal(body)
			linkReq, _ := http.NewRequest("POST", linkURL, bytes.NewBuffer(bodyBytes))
			linkReq.Header.Set("Authorization", "Bearer "+accessToken)
			linkReq.Header.Set("Content-Type", "application/json")
			linkResp, err := client.Do(linkReq)
			if err == nil {
				defer linkResp.Body.Close()
				if linkResp.StatusCode == 200 || linkResp.StatusCode == 201 {
					var linkResult map[string]interface{}
					if err := json.NewDecoder(linkResp.Body).Decode(&linkResult); err == nil {
						if link, ok := linkResult["link"].(map[string]interface{}); ok {
							if urlStr, ok := link["webUrl"].(string); ok && urlStr != "" {
								resultURL = urlStr
								return
							}
						}
						if urlStr, ok := linkResult["webUrl"].(string); ok && urlStr != "" {
							resultURL = urlStr
							return
						}
					}
				}
			}

			if bestWebURL != "" {
				resultURL = bestWebURL
				return
			}
			if bestDownload != "" {
				resultURL = bestDownload
				return
			}
		}()

		if resultURL != "" {
			return resultURL, nil
		}
	}

	if defaultShareURL != "" {
		return defaultShareURL, nil
	}

	return "", fmt.Errorf("no suitable recording item")
}

func (s *microsoftMeetingService) generateShortCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude confusing chars: I,O,0,1
	code := make([]byte, 8)
	for i := range code {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		code[i] = charset[int(randomByte[0])%len(charset)]
	}
	return string(code)
}

func (s *microsoftMeetingService) normalizeConferenceID(conferenceID, joinURL, calendarEventID string, eventResp *GraphEventResponse) string {
	const maxLen = 100

	if conferenceID == "" {
		conferenceID = calendarEventID
	}

	if conferenceID == "" && joinURL != "" {
		hash := sha256.Sum256([]byte(joinURL))
		conferenceID = fmt.Sprintf("cal-%x", hash[:12])
	}

	if len(conferenceID) > maxLen {
		conferenceID = conferenceID[:maxLen]
	}
	return conferenceID
}
