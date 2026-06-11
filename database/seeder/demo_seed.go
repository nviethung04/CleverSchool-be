package main

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/resources"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

const (
	demoSchoolMarker = "CS-DEMO"
	demoPassword     = "demo123"
)

// SeedDemoMVP tạo trường, khóa học, GV/HS, câu hỏi, exam/homework đã giao — idempotent.
func (s *Seeder) SeedDemoMVP() {
	var existing models.School
	if err := db.MasterDB.Where("short_name = ?", demoSchoolMarker).First(&existing).Error; err == nil {
		fmt.Printf("ℹ️  Demo đã có (school_id=%d). Bỏ qua seed.\n", existing.ID)
		printDemoCredentials()
		return
	}

	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now.AddDate(0, 6, 0)

	school := models.School{
		Name:      "CleverSchool Demo",
		ShortName: demoSchoolMarker,
		Type:      "school",
		Status:    true,
		AddressVN: "123 Đường Demo, Hà Nội",
	}
	if err := db.MasterDB.Create(&school).Error; err != nil {
		fmt.Println("❌ Tạo school:", err)
		return
	}

	subject := models.Subject{Name: "Tiếng Anh Demo", Description: "Môn demo MVP", Status: true}
	if err := db.MasterDB.Create(&subject).Error; err != nil {
		fmt.Println("❌ Tạo subject:", err)
		return
	}

	program := models.Program{
		Name:        "Chương trình Demo A1",
		Description: "Program demo cho test MVP",
		Status:      true,
		Duration:    120,
	}
	// Baseline migration 0002: chỉ name, description, duration, status (không có image_info).
	if err := db.MasterDB.Select("Name", "Description", "Status", "Duration").Create(&program).Error; err != nil {
		fmt.Println("❌ Tạo program:", err)
		return
	}

	course := models.Course{
		SubjectId:   subject.ID,
		ProgramId:   program.ID,
		Name:        "Khóa Demo - Học kỳ 1",
		ObjectTitle: "khoa-demo-hk1",
		Description: "Khóa học mẫu để test login, bài học, exam, homework",
		Type:        "online",
		Level:       "A1",
		Status:      true,
		State:       "active",
		StartDate:   start,
		EndDate:     end,
	}
	if err := db.MasterDB.Create(&course).Error; err != nil {
		fmt.Println("❌ Tạo course:", err)
		return
	}

	if err := db.MasterDB.Create(&models.CourseSchool{CourseId: course.ID, SchoolId: school.ID}).Error; err != nil {
		fmt.Println("❌ course_schools:", err)
		return
	}

	var gradeID int64
	if err := insertDemoGrade(&gradeID); err != nil {
		fmt.Println("❌ Tạo grade:", err)
		return
	}

	class := models.Class{
		SchoolId: school.ID,
		GradeId:  gradeID,
		Name:     "10A1 Demo",
		Status:   true,
	}
	if err := db.MasterDB.Create(&class).Error; err != nil {
		fmt.Println("❌ Tạo class:", err)
		return
	}

	var chapterID int64
	if err := db.MasterDB.Raw(`
		INSERT INTO chapters (course_id, program_id, title, object_title, description, sort_position, status)
		VALUES (?, ?, ?, ?, ?, ?, TRUE) RETURNING id`,
		course.ID, program.ID, "Chương 1 - Chào hỏi", "chuong-1-chao-hoi", "Nội dung demo chương 1", 1,
	).Scan(&chapterID).Error; err != nil {
		fmt.Println("❌ Tạo chapter:", err)
		return
	}

	lesson := models.Lesson{
		ChapterID:    chapterID,
		Title:        "Bài 1 - Hello",
		ObjectTitle:  "bai-1-hello",
		Description:  "Bài học demo",
		Status:       true,
		SortPosition: 1,
	}
	if err := db.MasterDB.Create(&lesson).Error; err != nil {
		fmt.Println("❌ Tạo lesson:", err)
		return
	}

	teacher, err := s.ensureDemoUser("teacher1", "Giáo viên Demo", int(school.ID), models.TeacherRoleId)
	if err != nil {
		fmt.Println("❌ teacher1:", err)
		return
	}
	student1, err := s.ensureDemoUser("student1", "Học sinh Demo 1", int(school.ID), models.StudentRoleId)
	if err != nil {
		fmt.Println("❌ student1:", err)
		return
	}
	student2, err := s.ensureDemoUser("student2", "Học sinh Demo 2", int(school.ID), models.StudentRoleId)
	if err != nil {
		fmt.Println("❌ student2:", err)
		return
	}

	for _, uc := range []models.UserCourse{
		{UserId: teacher.ID, CourseId: course.ID, IsCurrent: true, MainTeacher: true, StartTime: start, EndTime: end},
		{UserId: student1.ID, CourseId: course.ID, IsCurrent: true, StartTime: start, EndTime: end},
		{UserId: student2.ID, CourseId: course.ID, IsCurrent: true, StartTime: start, EndTime: end},
	} {
		if err := db.MasterDB.Where("user_id = ? AND course_id = ?", uc.UserId, uc.CourseId).
			FirstOrCreate(&uc).Error; err != nil {
			fmt.Println("❌ user_courses:", err)
			return
		}
	}

	for _, ucl := range []models.UserClass{
		{UserId: student1.ID, ClassId: class.ID, IsCurrent: true, StartTime: start, EndTime: end},
		{UserId: student2.ID, ClassId: class.ID, IsCurrent: true, StartTime: start, EndTime: end},
	} {
		if err := db.MasterDB.Where("user_id = ? AND class_id = ?", ucl.UserId, ucl.ClassId).
			FirstOrCreate(&ucl).Error; err != nil {
			fmt.Println("❌ user_classes:", err)
			return
		}
	}

	mcq, writing, err := s.ensureDemoQuestions(subject.ID)
	if err != nil {
		fmt.Println("❌ questions:", err)
		return
	}

	exam := models.Exam{
		ProgramId:      program.ID,
		Name:           "Kiểm tra Demo - Chào hỏi",
		ObjectTitle:    "kiem-tra-demo-chao-hoi",
		Status:         1,
		TimeLimit:      1800,
		MaxScore:       20,
		Description:    "Bài kiểm tra demo (MCQ + tự luận)",
		Deadline:       end,
		TotalQuestions: 2,
		QuestionForm:   models.QuestionFormQuestionType,
	}
	if err := db.MasterDB.Create(&exam).Error; err != nil {
		fmt.Println("❌ exam:", err)
		return
	}

	homework := models.Homework{
		ProgramId:      program.ID,
		Name:           "Bài tập về nhà Demo",
		ObjectTitle:    "bttn-demo",
		Status:         1,
		MaxScore:       10,
		Description:    "Homework demo",
		TotalQuestions: 1,
		QuestionForm:   models.QuestionFormQuestionType,
	}
	if err := db.MasterDB.Create(&homework).Error; err != nil {
		fmt.Println("❌ homework:", err)
		return
	}

	examQJSON, err := marshalQuestionsJSON([]*models.Question{mcq, writing})
	if err != nil {
		fmt.Println("❌ exam cloned JSON:", err)
		return
	}
	hwQJSON, err := marshalQuestionsJSON([]*models.Question{writing})
	if err != nil {
		fmt.Println("❌ homework cloned JSON:", err)
		return
	}

	if err := db.MasterDB.Create(&models.ClonedQuestion{
		ProgramId:      program.ID,
		AssignmentID:   exam.ID,
		AssignmentType: models.ClonedQuestionTypeExam,
		Questions:      examQJSON,
	}).Error; err != nil {
		fmt.Println("❌ cloned exam questions:", err)
		return
	}
	if err := db.MasterDB.Create(&models.ClonedQuestion{
		ProgramId:      program.ID,
		AssignmentID:   homework.ID,
		AssignmentType: models.ClonedQuestionTypeHomework,
		Questions:      hwQJSON,
	}).Error; err != nil {
		fmt.Println("❌ cloned homework questions:", err)
		return
	}

	assignAt := now
	teacherID := teacher.ID
	if err := db.MasterDB.Create(&models.ExamRefLesson{
		ExamId:     exam.ID,
		LessonId:   lesson.ID,
		CourseId:   course.ID,
		AssignedAt: &assignAt,
		AssignedBy: &teacherID,
	}).Error; err != nil {
		fmt.Println("❌ exam_ref_lessons:", err)
		return
	}
	if err := db.MasterDB.Model(&models.Exam{}).Where("id = ?", exam.ID).Updates(map[string]interface{}{
		"is_assigned": true,
		"assigned_at": assignAt,
		"assigned_by": teacherID,
	}).Error; err != nil {
		fmt.Println("❌ exam assigned flags:", err)
		return
	}

	if err := db.MasterDB.Create(&models.HomeworkRefLesson{
		HomeworkId: homework.ID,
		LessonId:   lesson.ID,
		CourseId:   course.ID,
		AssignedAt: &assignAt,
		AssignedBy: &teacherID,
	}).Error; err != nil {
		fmt.Println("❌ homework_ref_lessons:", err)
		return
	}
	if err := db.MasterDB.Model(&models.Homework{}).Where("id = ?", homework.ID).Updates(map[string]interface{}{
		"is_assigned": true,
		"assigned_at": assignAt,
		"assigned_by": teacherID,
	}).Error; err != nil {
		fmt.Println("❌ homework assigned flags:", err)
		return
	}

	fmt.Println("✅ Seed demo MVP xong.")
	fmt.Printf("   school_id=%d course_id=%d lesson_id=%d exam_id=%d homework_id=%d\n",
		school.ID, course.ID, lesson.ID, exam.ID, homework.ID)
	printDemoCredentials()
}

func (s *Seeder) ensureDemoUser(username, name string, schoolID int, roleID int64) (*models.User, error) {
	var user models.User
	err := db.MasterDB.Where("username = ?", username).First(&user).Error
	if err == nil {
		_ = s.ensureUserRole(user.ID, roleID)
		return &user, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user = models.User{
		Username: username,
		Name:     name,
		Password: string(hashed),
		Status:   true,
		SchoolID: schoolID,
		Email:    username + "@demo.local",
	}
	if err := db.MasterDB.Create(&user).Error; err != nil {
		return nil, err
	}
	if err := s.ensureUserRole(user.ID, roleID); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Seeder) ensureDemoQuestions(subjectID int64) (*models.Question, *models.Question, error) {
	mcqTitle := "[DEMO] MCQ - Thủ đô Việt Nam"
	var mcq models.Question
	err := db.MasterDB.Where("title = ?", mcqTitle).Preload("Answers").First(&mcq).Error
	if err == gorm.ErrRecordNotFound {
		mcq = models.Question{
			Kind:             "text",
			QuestionType:     models.QuestionTypeMultipleChoice,
			Title:            mcqTitle,
			Content:          "Thủ đô của Việt Nam là gì?",
			Point:            10,
			TimeLimitSeconds: 120,
			Status:           true,
			SubjectId:        subjectID,
			Answers: []models.Answer{
				{Content: "Hà Nội", IsCorrect: true, Kind: "text"},
				{Content: "TP. Hồ Chí Minh", IsCorrect: false, Kind: "text"},
				{Content: "Đà Nẵng", IsCorrect: false, Kind: "text"},
			},
		}
		if err := db.MasterDB.Create(&mcq).Error; err != nil {
			return nil, nil, err
		}
		_ = db.MasterDB.Preload("Answers").First(&mcq, mcq.ID)
	} else if err != nil {
		return nil, nil, err
	}

	writingTitle := "[DEMO] Writing - Giới thiệu bản thân"
	var writing models.Question
	err = db.MasterDB.Where("title = ?", writingTitle).First(&writing).Error
	if err == gorm.ErrRecordNotFound {
		writing = models.Question{
			Kind:             "text",
			QuestionType:     models.QuestionTypeWriting,
			Title:            writingTitle,
			Content:          "Viết 2-3 câu giới thiệu về bản thân.",
			Point:            10,
			TimeLimitSeconds: 300,
			Status:           true,
			SubjectId:        subjectID,
		}
		if err := db.MasterDB.Create(&writing).Error; err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	return &mcq, &writing, nil
}

func marshalQuestionsJSON(questions []*models.Question) ([]byte, error) {
	res := resources.NewQuestionResource()
	opts := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}
	var buf strings.Builder
	buf.WriteString("[")
	for i, q := range questions {
		formatted := res.FormatQuestion(q)
		if formatted == nil {
			continue
		}
		b, err := opts.Marshal(formatted)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
		if i < len(questions)-1 {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	return []byte(buf.String()), nil
}

// insertDemoGrade — model Grade dùng number/name_vn; baseline migration 0002 dùng name/status.
func insertDemoGrade(gradeID *int64) error {
	grade := models.Grade{Number: 10, NameVN: "Khối 10 Demo", NameEN: "Grade 10 Demo"}
	if err := db.MasterDB.Select("Number", "NameVN", "NameEN").Create(&grade).Error; err == nil {
		*gradeID = grade.ID
		return nil
	}
	return db.MasterDB.Raw(
		`INSERT INTO grades (name, status) VALUES (?, true) RETURNING id`,
		"Khối 10 Demo",
	).Scan(gradeID).Error
}

func printDemoCredentials() {
	fmt.Println("")
	fmt.Println("── Tài khoản demo (mật khẩu chung: demo123) ──")
	fmt.Println("   admin    — quản trị (seed Role)")
	fmt.Println("   teacher1 — giáo viên khóa demo")
	fmt.Println("   student1 — học sinh 1")
	fmt.Println("   student2 — học sinh 2")
	fmt.Println("── Luồng test ──")
	fmt.Println("   1. Login teacher1 → khóa 'Khóa Demo - Học kỳ 1'")
	fmt.Println("   2. Login student1 → làm exam/homework bài 1")
	fmt.Println("   3. Login teacher1 → chấm bài tự luận")
	fmt.Println("")
}
