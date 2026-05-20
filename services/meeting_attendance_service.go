package services

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/dto"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// formatDurationMMSS converts total seconds to MM:SS string.
func formatDurationMMSS(totalSeconds int) string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

type MeetingAttendanceService interface {
	JoinMeeting(meetingID, userID uint, joinMethod, deviceInfo, ipAddress string) (*models.MeetingAttendance, error)
	LeaveMeeting(meetingID, userID uint) error
	GetMeetingAttendances(meetingID uint) ([]*models.MeetingAttendance, error)
	GetMeetingAttendancesWithDetails(meetingID uint) ([]*dto.MeetingAttendanceResponse, error)
	ExportAttendancesToExcel(meetingID uint, writer io.Writer) error
	GetUserAttendanceHistory(userID uint, page, limit int) ([]*models.MeetingAttendance, error)
	GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error)
	JoinByShortCode(shortCode string, userID uint, roleId int, deviceInfo, ipAddress string) (*models.MicrosoftMeeting, *models.MeetingAttendance, error)
	GenerateShortCode() string
}

type meetingAttendanceService struct {
	attendanceRepo repositories.MeetingAttendanceRepository
	meetingRepo    repositories.MicrosoftMeetingRepository
}

func NewMeetingAttendanceService() MeetingAttendanceService {
	return &meetingAttendanceService{
		attendanceRepo: repositories.NewMeetingAttendanceRepository(),
		meetingRepo:    repositories.NewMicrosoftMeetingRepository(),
	}
}

func (s *meetingAttendanceService) isUserInCourse(userID uint, courseID uint) (bool, error) {
	var count int64
	err := db.ReplicaDB.Table("user_courses").
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *meetingAttendanceService) JoinMeeting(meetingID, userID uint, joinMethod, deviceInfo, ipAddress string) (*models.MeetingAttendance, error) {
	// Check if meeting exists
	meeting, err := s.meetingRepo.GetByID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("meeting not found: %v", err)
	}

	if meeting.CourseID != nil {
		ok, err := s.isUserInCourse(userID, *meeting.CourseID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate course membership: %v", err)
		}
		if !ok {
			return nil, fmt.Errorf("forbidden: user is not enrolled in this course")
		}
	}

	existing, err := s.attendanceRepo.GetByMeetingAndUser(meetingID, userID)
	if err == nil && existing.LeftAt == nil {
		config.Log.Infof("User %d already in meeting %d", userID, meetingID)
		return existing, nil
	}
	if err != nil {
		// If not found, we will create a new attendance record; other errors should stop.
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing attendance: %v", err)
		}
	}

	attendance := &models.MeetingAttendance{
		MeetingID:  meetingID,
		UserID:     userID,
		JoinedAt:   time.Now(),
		JoinMethod: joinMethod,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IsPresent:  true,
	}

	if err := s.attendanceRepo.Create(attendance); err != nil {
		return nil, fmt.Errorf("failed to create attendance record: %v", err)
	}

	config.Log.Infof("✅ User %d joined meeting %d (%s)", userID, meetingID, meeting.Title)
	return attendance, nil
}

func (s *meetingAttendanceService) LeaveMeeting(meetingID, userID uint) error {
	attendance, err := s.attendanceRepo.GetByMeetingAndUser(meetingID, userID)
	if err != nil {
		return fmt.Errorf("attendance record not found: %v", err)
	}

	leftAt := time.Now()
	duration := int(leftAt.Sub(attendance.JoinedAt).Minutes())

	attendance.LeftAt = &leftAt
	attendance.DurationMinutes = duration

	if err := s.attendanceRepo.Update(attendance); err != nil {
		return fmt.Errorf("failed to update attendance: %v", err)
	}

	config.Log.Infof("✅ User %d left meeting %d after %d minutes", userID, meetingID, duration)
	return nil
}

func (s *meetingAttendanceService) GetMeetingAttendances(meetingID uint) ([]*models.MeetingAttendance, error) {
	return s.attendanceRepo.GetByMeetingID(meetingID)
}

func (s *meetingAttendanceService) GetMeetingAttendancesWithDetails(meetingID uint) ([]*dto.MeetingAttendanceResponse, error) {
	attendances, err := s.attendanceRepo.GetByMeetingIDWithDetails(meetingID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.MeetingAttendanceResponse, 0, len(attendances))
	for _, att := range attendances {
		resp := &dto.MeetingAttendanceResponse{
			ID:              att.ID,
			MeetingID:       att.MeetingID,
			UserID:          att.UserID,
			Email:           att.User.Email,
			Username:        att.User.Username,
			FullName:        att.User.Name,
			SchoolID:        att.User.SchoolID,
			JoinedAt:        att.JoinedAt,
			LeftAt:          att.LeftAt,
			DurationMinutes: att.DurationMinutes,
			JoinMethod:      att.JoinMethod,
			DeviceInfo:      att.DeviceInfo,
			IPAddress:       att.IPAddress,
			IsPresent:       att.IsPresent,
			CreatedAt:       att.CreatedAt,
			UpdatedAt:       att.UpdatedAt,
		}

		if att.User.School.ID != 0 {
			resp.SchoolName = att.User.School.Name
		}

		resp.ClassNames = make([]string, 0)
		for _, uc := range att.User.UserClasses {
			if uc.Class != nil && uc.Class.Name != "" {
				resp.ClassNames = append(resp.ClassNames, uc.Class.Name)
			}
		}

		resp.CourseNames = make([]string, 0)
		for _, uc := range att.User.UserCourses {
			if uc.Course != nil && uc.Course.Name != "" {
				resp.CourseNames = append(resp.CourseNames, uc.Course.Name)
			}
		}

		if att.Meeting.ID != 0 && att.Meeting.Course != nil {
			resp.CourseName = att.Meeting.Course.Name
		}

		if att.Meeting.ID != 0 && att.Meeting.Lesson != nil {
			resp.LessonName = att.Meeting.Lesson.Title
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *meetingAttendanceService) GetUserAttendanceHistory(userID uint, page, limit int) ([]*models.MeetingAttendance, error) {
	offset := (page - 1) * limit
	return s.attendanceRepo.GetUserAttendanceHistory(userID, limit, offset)
}

func (s *meetingAttendanceService) GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error) {
	return s.attendanceRepo.GetAttendanceSummary(meetingID)
}

// ExportAttendancesToExcel exports attendance data to Excel file
func (s *meetingAttendanceService) ExportAttendancesToExcel(meetingID uint, writer io.Writer) error {
	// Get attendance data with details
	attendances, err := s.GetMeetingAttendancesWithDetails(meetingID)
	if err != nil {
		return fmt.Errorf("failed to get attendances: %v", err)
	}

	// Get meeting details
	meeting, err := s.meetingRepo.GetByID(meetingID)
	if err != nil {
		return fmt.Errorf("failed to get meeting: %v", err)
	}

	rosterUsers := make([]*dto.MeetingAttendanceResponse, 0)
	rosterSet := make(map[uint]bool)
	if meeting.CourseID != nil {
		rows, err := s.attendanceRepo.GetRosterUserRowsByCourseID(*meeting.CourseID)
		if err != nil {
			return fmt.Errorf("failed to load roster users: %v", err)
		}

		// group classes per user
		rosterMap := make(map[uint]*dto.MeetingAttendanceResponse)
		for _, r := range rows {
			entry, ok := rosterMap[r.UserID]
			if !ok {
				entry = &dto.MeetingAttendanceResponse{
					UserID:     r.UserID,
					Email:      r.Email,
					Username:   r.Username,
					FullName:   r.FullName,
					SchoolName: r.School,
					CourseName: r.Course,
				}
				rosterMap[r.UserID] = entry
			}
			if r.ClassName != "" {
				entry.ClassNames = append(entry.ClassNames, r.ClassName)
			}
		}
		for _, v := range rosterMap {
			rosterUsers = append(rosterUsers, v)
			rosterSet[v.UserID] = true
		}
	}

	attByUser := make(map[uint]*dto.MeetingAttendanceResponse)
	for _, att := range attendances {
		attByUser[att.UserID] = att
	}

	// Sort roster by full name for consistent ordering
	sort.Slice(rosterUsers, func(i, j int) bool {
		return strings.ToLower(rosterUsers[i].FullName) < strings.ToLower(rosterUsers[j].FullName)
	})

	unknownAttendances := make([]*dto.MeetingAttendanceResponse, 0)

	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Set sheet name
	sheetName := "Attendance"
	f.SetSheetName("Sheet1", sheetName)

	lessonTitle := ""
	if meeting.Lesson != nil {
		if meeting.Lesson.Title != "" {
			lessonTitle = meeting.Lesson.Title
		} else if meeting.Lesson.ObjectTitle != "" {
			lessonTitle = meeting.Lesson.ObjectTitle
		}
	}
	courseTitle := ""
	if meeting.Course != nil {
		if meeting.Course.Name != "" {
			courseTitle = meeting.Course.Name
		}
	}
	parts := []string{}
	if lessonTitle != "" {
		parts = append(parts, lessonTitle)
	}
	if courseTitle != "" {
		parts = append(parts, courseTitle)
	}
	label := strings.Join(parts, " / ")
	if label == "" {
		label = "(chưa đặt tên)"
	}
	title := fmt.Sprintf("Danh sách tham gia meeting %s", label)
	f.SetCellValue(sheetName, "A1", title)

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
	})
	f.SetCellStyle(sheetName, "A1", "A1", style)
	f.MergeCell(sheetName, "A1", "M1")
	f.SetRowHeight(sheetName, 1, 25)

	// Set column widths
	colWidths := map[string]float64{
		"A": 8,  // No.
		"B": 28, // Email
		"C": 18, // Username
		"D": 22, // Full Name
		"E": 18, // School
		"F": 25, // Classes
		"G": 25, // Courses
		"H": 18, // Lesson
		"I": 16, // Vắng mặt
		"J": 16, // Có mặt
		"K": 22, // Join Time
		"L": 22, // Leave Time
		"M": 18, // Duration
		"N": 18, // Join Method
		"O": 28, // Ghi chú
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// Add headers
	headers := []string{
		"No.",
		"Email",
		"Username",
		"Full Name",
		"School",
		"Classes",
		"Courses",
		"Lesson",
		"Absent",
		"Present",
		"Join Time",
		"Leave Time",
		"Duration (Min)",
		"Join Method",
		"Notes",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"366092"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%c3", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 3, 20)

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "left",
		},
	})
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			Horizontal: "center",
		},
	})

	rowNum := 4

	for _, roster := range rosterUsers {
		att, ok := attByUser[roster.UserID]
		present := ok
		var leaveDisplay *time.Time
		if present {
			leaveDisplay = att.LeftAt
			if leaveDisplay == nil && !meeting.EndTime.IsZero() {
				leaveDisplay = &meeting.EndTime
			}
		}
		durSeconds := 0
		if present {
			if leaveDisplay != nil {
				durSeconds = int(leaveDisplay.Sub(att.JoinedAt).Seconds())
				if durSeconds < 0 {
					durSeconds = 0
				}
			} else if att.DurationMinutes > 0 {
				durSeconds = att.DurationMinutes * 60
			}
		}
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), rowNum-3)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), roster.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), roster.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), roster.FullName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), roster.SchoolName)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), strings.Join(roster.ClassNames, ", "))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), roster.CourseName)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), lessonTitle)
		if present {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), "✓")
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowNum), att.JoinedAt.Format("2006-01-02 15:04:05"))
			if leaveDisplay != nil {
				f.SetCellValue(sheetName, fmt.Sprintf("L%d", rowNum), leaveDisplay.Format("2006-01-02 15:04:05"))
			}
			f.SetCellValue(sheetName, fmt.Sprintf("M%d", rowNum), formatDurationMMSS(durSeconds))
			f.SetCellValue(sheetName, fmt.Sprintf("N%d", rowNum), att.JoinMethod)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), "X")
		}
		for i := 0; i < len(headers); i++ {
			cell := fmt.Sprintf("%c%d", 'A'+i, rowNum)
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", rowNum), fmt.Sprintf("I%d", rowNum), centerStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", rowNum), fmt.Sprintf("J%d", rowNum), centerStyle)
		rowNum++
	}

	for _, att := range attendances {
		if rosterSet[att.UserID] {
			// already listed as roster student
			continue
		}
		unknownAttendances = append(unknownAttendances, att)
	}

	sort.Slice(unknownAttendances, func(i, j int) bool {
		return strings.ToLower(unknownAttendances[i].FullName) < strings.ToLower(unknownAttendances[j].FullName)
	})

	unknownSheet := "Unknown"
	_, _ = f.NewSheet(unknownSheet)

	unknownHeaders := []string{
		"No.", "Email", "Username", "Full Name", "School", "Classes", "Courses", "Lesson", "Present", "Join Time", "Leave Time", "Duration (Min)",
		"Join Method", "Notes",
	}
	for i, header := range unknownHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(unknownSheet, cell, header)
		f.SetCellStyle(unknownSheet, cell, cell, headerStyle)
	}
	f.SetRowHeight(unknownSheet, 1, 20)
	rowNumUnknown := 2
	for _, att := range unknownAttendances {
		leaveDisplay := att.LeftAt
		durSeconds := 0
		if leaveDisplay == nil && !meeting.EndTime.IsZero() {
			leaveDisplay = &meeting.EndTime
		}
		if leaveDisplay != nil {
			durSeconds = int(leaveDisplay.Sub(att.JoinedAt).Seconds())
			if durSeconds < 0 {
				durSeconds = 0
			}
		} else if att.DurationMinutes > 0 {
			durSeconds = att.DurationMinutes * 60
		}
		f.SetCellValue(unknownSheet, fmt.Sprintf("A%d", rowNumUnknown), rowNumUnknown-1)
		f.SetCellValue(unknownSheet, fmt.Sprintf("B%d", rowNumUnknown), att.Email)
		f.SetCellValue(unknownSheet, fmt.Sprintf("C%d", rowNumUnknown), att.Username)
		f.SetCellValue(unknownSheet, fmt.Sprintf("D%d", rowNumUnknown), att.FullName)
		f.SetCellValue(unknownSheet, fmt.Sprintf("E%d", rowNumUnknown), att.SchoolName)
		f.SetCellValue(unknownSheet, fmt.Sprintf("F%d", rowNumUnknown), strings.Join(att.ClassNames, ", "))
		f.SetCellValue(unknownSheet, fmt.Sprintf("G%d", rowNumUnknown), strings.Join(att.CourseNames, ", "))
		f.SetCellValue(unknownSheet, fmt.Sprintf("H%d", rowNumUnknown), att.LessonName)
		f.SetCellValue(unknownSheet, fmt.Sprintf("I%d", rowNumUnknown), "✓")
		f.SetCellValue(unknownSheet, fmt.Sprintf("J%d", rowNumUnknown), att.JoinedAt.Format("2006-01-02 15:04:05"))
		if leaveDisplay != nil {
			f.SetCellValue(unknownSheet, fmt.Sprintf("K%d", rowNumUnknown), leaveDisplay.Format("2006-01-02 15:04:05"))
		}
		f.SetCellValue(unknownSheet, fmt.Sprintf("L%d", rowNumUnknown), formatDurationMMSS(durSeconds))
		f.SetCellValue(unknownSheet, fmt.Sprintf("M%d", rowNumUnknown), att.JoinMethod)
		note := "Not in roster"
		if lessonTitle != "" {
			note = fmt.Sprintf("Not enrolled in %s", lessonTitle)
		}
		f.SetCellValue(unknownSheet, fmt.Sprintf("N%d", rowNumUnknown), note)
		for i := 0; i < len(unknownHeaders); i++ {
			cell := fmt.Sprintf("%c%d", 'A'+i, rowNumUnknown)
			f.SetCellStyle(unknownSheet, cell, cell, cellStyle)
		}
		f.SetCellStyle(unknownSheet, fmt.Sprintf("I%d", rowNumUnknown), fmt.Sprintf("I%d", rowNumUnknown), centerStyle)
		rowNumUnknown++
	}

	f.SetActiveSheet(0)

	// Write to writer
	return f.Write(writer)
}

func (s *meetingAttendanceService) JoinByShortCode(shortCode string, userID uint, roleId int, deviceInfo, ipAddress string) (*models.MicrosoftMeeting, *models.MeetingAttendance, error) {
	// Find meeting by short code
	meeting, err := s.meetingRepo.GetByShortCode(shortCode)
	if err != nil {
		return nil, nil, fmt.Errorf("meeting not found with code: %s", shortCode)
	}

	// Enforce course membership (if meeting belongs to a course)
	if meeting.CourseID != nil {
		if roleId != models.AdminRoleId && roleId != models.SchoolRoleId {
			ok, err := s.isUserInCourse(userID, *meeting.CourseID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to validate course membership: %v", err)
			}
			if !ok {
				return nil, nil, fmt.Errorf("forbidden: user is not enrolled in this course")
			}
		}
	}

	nowUTC := time.Now().UTC()
	endUTC := meeting.EndTime.UTC()

	if nowUTC.After(endUTC.Add(30 * time.Minute)) {
		return nil, nil, fmt.Errorf("meeting has ended")
	}

	attendance, err := s.JoinMeeting(meeting.ID, userID, "short_link", deviceInfo, ipAddress)
	if err != nil {
		return meeting, nil, err
	}

	return meeting, attendance, nil
}

func (s *meetingAttendanceService) GenerateShortCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
