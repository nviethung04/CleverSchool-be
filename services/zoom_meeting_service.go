package services

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/repositories"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ZoomMeetingService interface {
	CreateMeeting(userID int64, req *CreateMeetingRequest) (*models.ZoomMeeting, error)
	GetMeetingByID(id int64) (*models.ZoomMeeting, error)
	GetUserMeetings(userID int64, limit, offset int) ([]models.ZoomMeeting, error)
	GetCourseMeetings(courseID int64) ([]models.ZoomMeeting, error)
	GetLessonMeetings(lessonID int64) ([]models.ZoomMeeting, error)
	UpdateMeeting(id int64, req *UpdateMeetingRequest) (*models.ZoomMeeting, error)
	DeleteMeeting(id int64) error
	StartMeeting(id int64) (*StartMeetingResponse, error)
}

type zoomMeetingService struct {
	meetingRepo  repositories.ZoomMeetingRepository
	zoomAuthRepo repositories.ZoomAuthRepository
	zoomConfig   *config.ZoomConfig
}

type CreateMeetingRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	Duration    int       `json:"duration" binding:"required"` // minutes
	CourseID    *int64    `json:"course_id"`
	LessonID    *int64    `json:"lesson_id"`
	Password    string    `json:"password"`
	WaitingRoom bool      `json:"waiting_room"`
	IsRecording bool      `json:"is_recording"`
}

type UpdateMeetingRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	StartTime   *time.Time `json:"start_time"`
	Duration    *int       `json:"duration"`
	Password    *string    `json:"password"`
	WaitingRoom *bool      `json:"waiting_room"`
	IsRecording *bool      `json:"is_recording"`
}

type StartMeetingResponse struct {
	JoinURL string `json:"join_url"`
	HostURL string `json:"host_url"`
}

type ZoomCreateMeetingRequest struct {
	Topic     string              `json:"topic"`
	Type      int                 `json:"type"`
	StartTime string              `json:"start_time"`
	Duration  int                 `json:"duration"`
	Password  string              `json:"password,omitempty"`
	Settings  ZoomMeetingSettings `json:"settings"`
}

type ZoomMeetingSettings struct {
	WaitingRoom bool `json:"waiting_room"`
	Recording   bool `json:"auto_recording"`
}

type ZoomMeetingResponse struct {
	ID       int64  `json:"id"`
	Topic    string `json:"topic"`
	JoinURL  string `json:"join_url"`
	HostURL  string `json:"start_url"`
	Password string `json:"password"`
}

func NewZoomMeetingService(meetingRepo repositories.ZoomMeetingRepository, zoomAuthRepo repositories.ZoomAuthRepository) ZoomMeetingService {
	return &zoomMeetingService{
		meetingRepo:  meetingRepo,
		zoomAuthRepo: zoomAuthRepo,
		zoomConfig:   config.GetZoomConfig(),
	}
}

func (s *zoomMeetingService) CreateMeeting(userID int64, req *CreateMeetingRequest) (*models.ZoomMeeting, error) {
	// Get user's Zoom account
	zoomAccount, err := s.zoomAuthRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not connected to Zoom")
	}

	// Create meeting on Zoom
	zoomReq := &ZoomCreateMeetingRequest{
		Topic:     req.Title,
		Type:      models.MeetingTypeScheduled,
		StartTime: req.StartTime.Format("2006-01-02T15:04:05Z"),
		Duration:  req.Duration,
		Password:  req.Password,
		Settings: ZoomMeetingSettings{
			WaitingRoom: req.WaitingRoom,
			Recording:   req.IsRecording,
		},
	}

	zoomResp, err := s.createZoomMeeting(zoomAccount.AccessToken, zoomReq)
	if err != nil {
		return nil, err
	}

	// Save meeting to database
	meeting := &models.ZoomMeeting{
		ZoomMeetingID: fmt.Sprintf("%d", zoomResp.ID),
		UserID:        userID,
		CourseID:      req.CourseID,
		LessonID:      req.LessonID,
		Title:         req.Title,
		Description:   req.Description,
		StartTime:     req.StartTime,
		Duration:      req.Duration,
		JoinURL:       zoomResp.JoinURL,
		HostURL:       zoomResp.HostURL,
		Password:      zoomResp.Password,
		MeetingType:   models.MeetingTypeScheduled,
		Status:        models.MeetingStatusScheduled,
		IsRecording:   req.IsRecording,
		WaitingRoom:   req.WaitingRoom,
	}

	err = s.meetingRepo.Create(meeting)
	if err != nil {
		return nil, err
	}

	return meeting, nil
}

func (s *zoomMeetingService) GetMeetingByID(id int64) (*models.ZoomMeeting, error) {
	return s.meetingRepo.GetByID(id)
}

func (s *zoomMeetingService) GetUserMeetings(userID int64, limit, offset int) ([]models.ZoomMeeting, error) {
	return s.meetingRepo.GetByUserID(userID, limit, offset)
}

func (s *zoomMeetingService) GetCourseMeetings(courseID int64) ([]models.ZoomMeeting, error) {
	return s.meetingRepo.GetByCourseID(courseID)
}

func (s *zoomMeetingService) GetLessonMeetings(lessonID int64) ([]models.ZoomMeeting, error) {
	return s.meetingRepo.GetByLessonID(lessonID)
}

func (s *zoomMeetingService) UpdateMeeting(id int64, req *UpdateMeetingRequest) (*models.ZoomMeeting, error) {
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.WaitingRoom != nil {
		updates["waiting_room"] = *req.WaitingRoom
	}
	if req.IsRecording != nil {
		updates["is_recording"] = *req.IsRecording
	}

	err := s.meetingRepo.Update(id, updates)
	if err != nil {
		return nil, err
	}

	return s.meetingRepo.GetByID(id)
}

func (s *zoomMeetingService) DeleteMeeting(id int64) error {
	meeting, err := s.meetingRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete from Zoom (optional - could also just delete from DB)
	if zoomAccount, err := s.zoomAuthRepo.GetByUserID(meeting.UserID); err == nil {
		s.deleteZoomMeeting(zoomAccount.AccessToken, meeting.ZoomMeetingID)
	}

	return s.meetingRepo.Delete(id)
}

func (s *zoomMeetingService) StartMeeting(id int64) (*StartMeetingResponse, error) {
	meeting, err := s.meetingRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update status to started
	s.meetingRepo.UpdateStatus(meeting.ZoomMeetingID, models.MeetingStatusStarted)

	return &StartMeetingResponse{
		JoinURL: meeting.JoinURL,
		HostURL: meeting.HostURL,
	}, nil
}

// Private helper methods for Zoom API calls
func (s *zoomMeetingService) createZoomMeeting(accessToken string, req *ZoomCreateMeetingRequest) (*ZoomMeetingResponse, error) {
	jsonData, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", s.zoomConfig.BaseURL+"/users/me/meetings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var zoomResp ZoomMeetingResponse
	err = json.Unmarshal(body, &zoomResp)
	return &zoomResp, err
}

func (s *zoomMeetingService) deleteZoomMeeting(accessToken, meetingID string) error {
	req, err := http.NewRequest("DELETE", s.zoomConfig.BaseURL+"/meetings/"+meetingID, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
