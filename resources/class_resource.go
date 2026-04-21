package resources

import (
	"be-lms/dto"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"time"
)

type ClassResource interface {
	FormatClass(class *models.Class) *prot.ClassResponse
	FormatClasses(classs []*models.Class) []*prot.ClassResponse
	FormatModelClass(class *prot.ClassRequest) *models.Class
	FormatClassByUserReponse(class *dto.Class) *prot.ClassByUserResponse
	FormatClassByUser(class *models.Class) *prot.ClassByUser
	MapUsersToDTOStudents(users []models.User, activityLogs map[int64]models.ActivityLog, currentUserId int64) []dto.Student
	MapCoursesToDTOCourses(users []models.Course) []dto.Course
}

type ClassResourceImpl struct{}

func NewClassResource() ClassResource {
	return &ClassResourceImpl{}
}

func (r *ClassResourceImpl) FormatClass(class *models.Class) *prot.ClassResponse {
	if class == nil {
		return nil
	}

	var maxStudents int32
	if class.MaxStudents != 0 {
		maxStudents = class.MaxStudents
	}

	var currentStudents int32
	if class.CurrentStudents != 0 {
		currentStudents = class.CurrentStudents
	}

	var grade *prot.Grade
	gradeResource := NewGradeResource()

	if class.Grade != nil {
		grade = gradeResource.FormatGrade(class.Grade)
	}

	var faculty *prot.Faculty
	facultyResource := NewFacultyResource()
	if class.Faculty != nil {
		faculty = facultyResource.FormatFaculty(class.Faculty)
	}

	var classMainId int64
	var classMainName string
	if class.ClassMain != nil {
		classMainId = class.ClassMain.ID
		classMainName = class.ClassMain.Name
	} else if class.ClassMainId > 0 {
		classMainId = class.ClassMainId
	}

	return &prot.ClassResponse{
		Id:              class.ID,
		SchoolId:        class.SchoolId,
		GradeId:         class.GradeId,
		Name:            class.Name,
		Status:          class.Status,
		CurrentStudents: currentStudents,
		MaxStudents:     maxStudents,
		TeacherInfo: &prot.TeacherInfo{
			Name:  class.TeacherInfo.Name,
			Phone: class.TeacherInfo.Phone,
			Email: class.TeacherInfo.Email,
		},
		FacultyId:      class.FacultyId,
		Grade:          grade,
		Faculty:        faculty,
		ClassMainId:    classMainId,
		ClassMainName:  classMainName,
		CreatedAt:      class.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      class.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *ClassResourceImpl) FormatClassByUser(class *models.Class) *prot.ClassByUser {
	if class == nil {
		return nil
	}

	return &prot.ClassByUser{
		Id:       class.ID,
		SchoolId: class.SchoolId,
		Name:     class.Name,
		TeacherInfo: &prot.TeacherInfo{
			Name:  class.TeacherInfo.Name,
			Phone: class.TeacherInfo.Phone,
			Email: class.TeacherInfo.Email,
		},
	}
}

func (r *ClassResourceImpl) FormatClasses(classs []*models.Class) []*prot.ClassResponse {
	result := make([]*prot.ClassResponse, 0, len(classs))
	for _, u := range classs {
		if formatted := r.FormatClass(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *ClassResourceImpl) FormatModelClass(class *prot.ClassRequest) *models.Class {
	if class == nil {
		return nil
	}

	var teacherInfo models.TeacherInfo
	if class.TeacherInfo != nil {
		teacherInfo = models.TeacherInfo{
			Name:  class.TeacherInfo.Name,
			Phone: class.TeacherInfo.Phone,
			Email: class.TeacherInfo.Email,
		}
	}

	return &models.Class{
		ID:          class.Id,
		SchoolId:    class.SchoolId,
		GradeId:     class.GradeId,
		FacultyId:   class.FacultyId,
		Name:        class.Name,
		Status:      class.Status,
		MaxStudents: class.MaxStudents,
		TeacherInfo: teacherInfo,
	}
}

func (r *ClassResourceImpl) FormatClassByUserReponse(class *dto.Class) *prot.ClassByUserResponse {
	if class == nil {
		return nil
	}

	var total int32 = int32(len(class.Students))
	var active int32

	students := make([]*prot.StudentInClass, 0, total)
	for _, s := range class.Students {
		if s.IsActivity {
			active++
		}
		students = append(students, &prot.StudentInClass{
			Id:           s.ID,
			Name:         s.Name,
			Username:     s.Username,
			Email:        s.Email,
			Avatar:       s.Avatar,
			IsActivity:   s.IsActivity,
			ActivityDate: s.ActivityDate,
		})
	}

	classResponse := r.FormatClassByUser(&class.Class)

	courses := make([]*prot.CourseByUser, 0, int32(len(class.Courses)))

	for _, c := range class.Courses {
		courses = append(courses, &prot.CourseByUser{
			Id:   c.ID,
			Name: c.Name,
		})
	}

	return &prot.ClassByUserResponse{
		TotalStudent:   int64(total),
		ActiveStudents: int64(active),
		Class:          classResponse,
		Students:       students,
		Courses:        courses,
	}
}

func (r *ClassResourceImpl) MapUsersToDTOStudents(users []models.User, activityLogs map[int64]models.ActivityLog, currentUserId int64) []dto.Student {
	var (
		now              = time.Now()
		currentUserFirst []dto.Student
		others           []dto.Student
	)

	for _, u := range users {
		student := dto.Student{
			ID:           u.ID,
			Name:         u.Name,
			Username:     u.Username,
			Email:        u.Email,
			Avatar:       u.AvatarInfo.Path,
			IsActivity:   false,
			ActivityDate: i18n.Localize("user_class.not_active"),
		}

		if log, ok := activityLogs[u.ID]; ok {
			duration := now.Sub(log.CreatedAt)
			days := int(duration.Hours() / 24)

			switch {
			case days == 0:
				student.ActivityDate = i18n.Localize("user_class.today")
			case days == 1:
				student.ActivityDate = i18n.Localize("user_class.yesterday")
			case days < 30:
				student.ActivityDate = i18n.Localize("user_class.days_ago", map[string]interface{}{"Days": days})
			default:
				months := days / 30
				student.ActivityDate = i18n.Localize("user_class.months_ago", map[string]interface{}{"Months": months})
			}

			if duration <= 5*time.Minute {
				student.IsActivity = true
			}
		}

		if u.ID == currentUserId {
			currentUserFirst = append(currentUserFirst, student)
		} else {
			others = append(others, student)
		}
	}

	return append(currentUserFirst, others...)
}

func (r *ClassResourceImpl) MapCoursesToDTOCourses(cousers []models.Course) []dto.Course {
	var currentCousers []dto.Course

	for _, c := range cousers {
		couser := dto.Course{
			ID:   c.ID,
			Name: c.Name,
		}

		currentCousers = append(currentCousers, couser)
	}

	return currentCousers
}
