package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"time"
)

type CourseRosterUserRow struct {
	UserID    uint
	Email     string
	Username  string
	FullName  string
	School    string
	ClassName string
	Course    string
}

type MeetingAttendanceRepository interface {
	Create(attendance *models.MeetingAttendance) error
	GetByID(id uint) (*models.MeetingAttendance, error)
	GetByMeetingID(meetingID uint) ([]*models.MeetingAttendance, error)
	GetByMeetingIDWithDetails(meetingID uint) ([]*models.MeetingAttendance, error) // Preload User details and filter out teachers
	GetByUserID(userID uint) ([]*models.MeetingAttendance, error)
	GetByMeetingAndUser(meetingID, userID uint) (*models.MeetingAttendance, error)
	Update(attendance *models.MeetingAttendance) error
	UpdateLeftAt(id uint, leftAt time.Time) error
	GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error)
	GetUserAttendanceHistory(userID uint, limit, offset int) ([]*models.MeetingAttendance, error)
	CheckUserAttended(meetingID, userID uint) (bool, error)
	GetRosterUserRowsByCourseID(courseID uint) ([]CourseRosterUserRow, error)
}

type meetingAttendanceRepository struct{}

func NewMeetingAttendanceRepository() MeetingAttendanceRepository {
	return &meetingAttendanceRepository{}
}

// Create creates a new attendance record
func (r *meetingAttendanceRepository) Create(attendance *models.MeetingAttendance) error {
	return db.MasterDB.Create(attendance).Error
}

func (r *meetingAttendanceRepository) GetByID(id uint) (*models.MeetingAttendance, error) {
	var attendance models.MeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").Preload("User").First(&attendance, id).Error
	return &attendance, err
}

func (r *meetingAttendanceRepository) GetByMeetingID(meetingID uint) ([]*models.MeetingAttendance, error) {
	var attendances []*models.MeetingAttendance
	err := db.ReplicaDB.Preload("User").Preload("Meeting.Course").Preload("Meeting.Lesson").
		Where("meeting_id = ?", meetingID).
		Order("joined_at ASC").
		Find(&attendances).Error
	return attendances, err
}

// GetByMeetingIDWithDetails returns attendances with full user details (School, Classes, Courses) excluding teachers
func (r *meetingAttendanceRepository) GetByMeetingIDWithDetails(meetingID uint) ([]*models.MeetingAttendance, error) {
	var attendances []*models.MeetingAttendance
	err := db.ReplicaDB.
		Preload("User").
		Preload("User.School").
		Preload("User.UserClasses").Preload("User.UserClasses.Class").
		Preload("User.UserCourses").Preload("User.UserCourses.Course").
		Preload("User.Roles").
		Preload("Meeting.Course").
		Preload("Meeting.Lesson").
		Where("meeting_id = ?", meetingID).
		Order("joined_at ASC").
		Find(&attendances).Error

	if err != nil {
		return attendances, err
	}

	// Filter out teachers (TeacherRoleId = 2)
	filtered := make([]*models.MeetingAttendance, 0, len(attendances))
	for _, att := range attendances {
		isTeacher := false
		for _, role := range att.User.Roles {
			if role.ID == models.TeacherRoleId {
				isTeacher = true
				break
			}
		}
		if !isTeacher {
			filtered = append(filtered, att)
		}
	}

	return filtered, nil
}

func (r *meetingAttendanceRepository) GetByUserID(userID uint) ([]*models.MeetingAttendance, error) {
	var attendances []*models.MeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Find(&attendances).Error
	return attendances, err
}

func (r *meetingAttendanceRepository) GetByMeetingAndUser(meetingID, userID uint) (*models.MeetingAttendance, error) {
	var attendance models.MeetingAttendance
	err := db.ReplicaDB.
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Order("joined_at DESC").
		First(&attendance).Error
	return &attendance, err
}

func (r *meetingAttendanceRepository) Update(attendance *models.MeetingAttendance) error {
	return db.MasterDB.Save(attendance).Error
}

func (r *meetingAttendanceRepository) UpdateLeftAt(id uint, leftAt time.Time) error {
	return db.MasterDB.Model(&models.MeetingAttendance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"left_at":          leftAt,
			"duration_minutes": db.MasterDB.Raw("EXTRACT(EPOCH FROM (? - joined_at)) / 60", leftAt),
		}).Error
}

func (r *meetingAttendanceRepository) GetAttendanceSummary(meetingID uint) (*models.AttendanceSummary, error) {
	var summary models.AttendanceSummary
	summary.MeetingID = meetingID

	var totalJoined int64
	db.ReplicaDB.Model(&models.MeetingAttendance{}).
		Where("meeting_id = ?", meetingID).
		Count(&totalJoined)
	summary.TotalJoined = int(totalJoined)

	var totalPresent int64
	db.ReplicaDB.Model(&models.MeetingAttendance{}).
		Where("meeting_id = ? AND is_present = true", meetingID).
		Count(&totalPresent)
	summary.TotalPresent = int(totalPresent)

	var avgDuration float64
	db.ReplicaDB.Model(&models.MeetingAttendance{}).
		Where("meeting_id = ? AND duration_minutes > 0", meetingID).
		Select("COALESCE(AVG(duration_minutes), 0)").
		Scan(&avgDuration)
	summary.AverageDuration = int(avgDuration)

	return &summary, nil
}

func (r *meetingAttendanceRepository) GetUserAttendanceHistory(userID uint, limit, offset int) ([]*models.MeetingAttendance, error) {
	var attendances []*models.MeetingAttendance
	err := db.ReplicaDB.Preload("Meeting").Preload("Meeting.Course").Preload("Meeting.Lesson").
		Where("user_id = ?", userID).
		Order("joined_at DESC").
		Limit(limit).Offset(offset).
		Find(&attendances).Error
	return attendances, err
}

func (r *meetingAttendanceRepository) CheckUserAttended(meetingID, userID uint) (bool, error) {
	var count int64
	err := db.ReplicaDB.Model(&models.MeetingAttendance{}).
		Where("meeting_id = ? AND user_id = ?", meetingID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *meetingAttendanceRepository) GetRosterUserRowsByCourseID(courseID uint) ([]CourseRosterUserRow, error) {
	rows := make([]CourseRosterUserRow, 0)
	err := db.ReplicaDB.
		Table("user_courses uc").
		Select("u.id as user_id, u.email, u.username, u.name as full_name, s.name as school, c.name as course, cls.name as class_name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("LEFT JOIN schools s ON s.id = u.school_id").
		Joins("LEFT JOIN user_classes uc2 ON uc2.user_id = u.id").
		Joins("LEFT JOIN classes cls ON cls.id = uc2.class_id").
		Joins("LEFT JOIN courses c ON c.id = uc.course_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Joins("JOIN roles r ON r.id = urr.role_id").
		Where("uc.course_id = ? AND r.name = ?", courseID, "student").
		Group("u.id, u.email, u.username, u.name, s.name, c.name, cls.name").
		Find(&rows).Error
	return rows, err
}
