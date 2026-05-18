package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
)

type ExamCourseExam struct {
	ID                int64
	Name              string
	LessonTitle       string
	Description       string
	CoverImage        string
	Deadline          int64
	Status            int32
	IsAssigned        bool
	StudentsDone      int32
	StudentsNotScored int32
}

type ExamCourseChapter struct {
	ChapterInfo struct {
		Title       string
		Description string
		Status      bool
	}
	Exams []ExamCourseExam
}

type ExamCourseDetail struct {
	CourseInfo struct {
		Name        string
		Description string
		Status      bool
	}
	Chapters []ExamCourseChapter
}

type ExamCourseRepository interface {
	GetExamCourseDetail(courseID int64, isAssigned *bool) (ExamCourseDetail, error)
}

type examCourseRepository struct{}

func NewExamCourseRepository() ExamCourseRepository {
	return &examCourseRepository{}
}

func (r *examCourseRepository) GetExamCourseDetail(courseID int64, isAssigned *bool) (ExamCourseDetail, error) {
	var result ExamCourseDetail
	// Lấy thông tin course
	var course models.Course
	err := db.MasterDB.Select("name, description, status").Where("id = ?", courseID).First(&course).Error
	if err != nil {
		return result, err
	}
	result.CourseInfo.Name = course.Name
	result.CourseInfo.Description = course.Description
	result.CourseInfo.Status = course.Status

	// Lấy danh sách chapter
	var chapters []models.Chapter
	err = db.MasterDB.Select("id, title, description, status").Where("course_id = ?", courseID).Order("sort_position").Find(&chapters).Error
	if err != nil {
		return result, err
	}

	for _, ch := range chapters {
		chapter := ExamCourseChapter{}
		chapter.ChapterInfo.Title = ch.Title
		chapter.ChapterInfo.Description = ch.Description
		chapter.ChapterInfo.Status = ch.Status

		// Lấy danh sách lesson thuộc chapter
		var lessons []models.Lesson
		db.MasterDB.Select("id, title").Where("chapter_id = ?", ch.ID).Find(&lessons)
		lessonMap := map[int64]string{}
		lessonIDs := []int64{}
		for _, l := range lessons {
			lessonMap[l.ID] = l.Title
			lessonIDs = append(lessonIDs, l.ID)
		}

		// Lấy danh sách exam thuộc các lessons này, lọc theo is_assigned nếu có
		var exams []models.Exam
		if len(lessonIDs) > 0 {
			dbQuery := db.MasterDB.
				Select("exams.id, exams.name, exams.description, exams.cover_image_info, exams.deadline, exams.status").
				Joins("JOIN exam_ref_lessons erl ON erl.exam_id = exams.id").
				Where("erl.lesson_id IN ?", lessonIDs)

			if isAssigned != nil {
				if *isAssigned {
					dbQuery = dbQuery.Where("erl.assigned_by IS NOT NULL AND erl.assigned_by > 0")
				} else {
					dbQuery = dbQuery.Where("(erl.assigned_by IS NULL OR erl.assigned_by = 0)")
				}
			}
			dbQuery.Preload("Lessons", "ExamRefLessons").Find(&exams)
		}

		for _, ex := range exams {
			var isAssigned bool
			var lesson models.Lesson

			if len(ex.Lessons) > 0 {
				for _, l := range lessons {
					if l.ChapterID == ch.ID {
						lesson = l
						for _, ref := range ex.ExamRefLessons {
							if ref.LessonId == l.ID {
								isAssigned = ref.AssignedBy != nil && *ref.AssignedBy > 0
							}
						}
					}
				}
			}
			exam := ExamCourseExam{
				ID:          ex.ID,
				Name:        ex.Name,
				LessonTitle: lesson.Title,
				Description: ex.Description,
				CoverImage:  ex.CoverImageInfo.Path,
				Deadline:    ex.Deadline.Unix(),
				Status:      int32(ex.Status),
				IsAssigned:  isAssigned,
			}

			var studentsDone int64
			var studentsNotScored int64
			db.MasterDB.Model(&models.ExamUser{}).Where("exam_id = ?", ex.ID).Distinct("user_id").Count(&studentsDone)
			db.MasterDB.Model(&models.ExamQuestionUserManualScoring{}).Where("exam_id = ? AND is_scored = false", ex.ID).Distinct("user_id").Count(&studentsNotScored)
			exam.StudentsDone = int32(studentsDone)
			exam.StudentsNotScored = int32(studentsNotScored)
			chapter.Exams = append(chapter.Exams, exam)
		}
		result.Chapters = append(result.Chapters, chapter)
	}
	return result, nil
}
