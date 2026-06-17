package main

import (
	"be-lms/config"
	"be-lms/database/db"
	"be-lms/models"
	redisperm "be-lms/redis"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Seeder struct{}

func main() {
	// Load env
	_ = godotenv.Load()

	// Load config
	cfg := config.LoadConfig()

	// Connect Database
	if err := db.ConnectPostgres(cfg); err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}

	if cfg.RedisEnabled {
		if err := db.ConnectRedis(cfg); err != nil {
			log.Println("⚠️ Redis không kết nối được — seed vẫn chạy, nhớ xóa cache permission thủ công")
		}
	}

	fmt.Println("🎉 Create fake data")

	seed := &Seeder{}

	fmt.Println("🎉 Fake data has been saved to the database.!")

	model := flag.String("model", "", "Task to run (e.g., role, user)")
	// count := flag.Int("count", 10, "Number of records to seed")
	flag.Parse()

	switch *model {
	case "Question":
		seed.SeedQuestions()
	case "QuestionAttribute":
		seed.SeedQuestionAttributes()
	case "Role":
		seed.SeedRoles()
	case "UserClassAndCourse":
		seed.UserClassAndCourse()
	case "Week":
		currentYear := time.Now().Year()
		seed.SeedWeeksForYear(currentYear)
	case "Flashcard":
		seed.SeedFlashcards()
	case "Lesson4Flashcard":
		SeedLesson4Flashcards()
	default:
		seed.SeedQuestions()
		seed.SeedRoles()
		return
	}
}
func (s *Seeder) SeedWeeksForYear(year int) {
	var count int64
	if err := db.MasterDB.Model(&models.Week{}).
		Where("year = ?", year).
		Count(&count).Error; err != nil {
		log.Printf("Error checking existing weeks for year %d: %v", year, err)
		return
	}

	if count > 0 {
		log.Printf("Weeks for year %d already exist. Skipping seeding.", year)
		return
	}

	// Bắt đầu từ Thứ Hai của tuần chứa ngày 4/1 (ISO week 1)
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	start := jan4
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}

	// Lặp tạo tuần đến khi ISO year > năm cần seed
	for {
		isoYear, isoWeek := start.ISOWeek()
		if isoYear > year {
			break
		}

		if isoYear == year {
			end := start.AddDate(0, 0, 6)

			weekData := models.Week{
				Year:       year,
				WeekNumber: isoWeek,
				StartDate:  start,
				EndDate:    end,
			}

			if err := db.MasterDB.Create(&weekData).Error; err != nil {
				log.Printf("Error seeding week %d/%d: %v", isoWeek, year, err)
			}
		}

		// Chuyển sang tuần tiếp theo
		start = start.AddDate(0, 0, 7)
	}
}

func (s *Seeder) SeedQuestionAttributes() {
	var subjectId int64
	if err := db.MasterDB.Table("subjects").
		Where("deleted_at IS NULL").
		Order("id ASC").
		Limit(1).
		Pluck("id", &subjectId).Error; err != nil || subjectId == 0 {
		log.Println("⚠️ Skip QuestionAttribute seed: no subjects in database")
		return
	}

	var existing int64
	db.MasterDB.Model(&models.QuestionAttribute{}).
		Where("subject_id = ? AND name = ? AND COALESCE(parent_id, 0) = 0", subjectId, "Kỹ năng").
		Count(&existing)
	if existing > 0 {
		log.Println("ℹ️ Question attributes already seeded — skipping")
		return
	}

	attributes := []models.QuestionAttribute{
		{
			ParentID:  nil,
			SubjectId: &subjectId,
			Name:      "Kỹ năng",
			Level:     1,
			Weight:    1,
			Nodes: []models.QuestionAttribute{
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Vocabulary",
					Level:     2,
					Weight:    1,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Structure",
					Level:     2,
					Weight:    2,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Phonics",
					Level:     2,
					Weight:    3,
				},
			},
		},
		{
			ParentID:  nil,
			SubjectId: &subjectId,
			Name:      "Mức độ",
			Level:     1,
			Weight:    1,
			Nodes: []models.QuestionAttribute{
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Easy",
					Level:     2,
					Weight:    1,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Medium",
					Level:     2,
					Weight:    2,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Hard",
					Level:     2,
					Weight:    3,
				},
			},
		},
		{
			ParentID:  nil,
			SubjectId: &subjectId,
			Name:      "Mức độ nhận thức",
			Level:     1,
			Weight:    1,
			Nodes: []models.QuestionAttribute{
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Nhận diện",
					Level:     2,
					Weight:    1,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Ghi nhớ",
					Level:     2,
					Weight:    1,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Vận dụng",
					Level:     2,
					Weight:    2,
				},
				{
					ParentID:  nil,
					SubjectId: &subjectId,
					Name:      "Vận dụng nâng cao/sáng tạo",
					Level:     2,
					Weight:    1,
				},
			},
		},
	}

	for _, a := range attributes {
		if err := db.MasterDB.Create(&a).Error; err != nil {
			log.Printf("Error seeding attributes: %v", err)
		}
	}
}

func (s *Seeder) SeedQuestions() {
	questions := []models.Question{
		{
			Kind:             "image",
			QuestionType:     "multiple_choice",
			Title:            "Chọn đáp án đúng.",
			Content:          "Thủ đô của Việt Nam là gì?",
			Point:            10,
			TimeLimitSeconds: 60,
			Answers: []models.Answer{
				{Content: "Hà Nội", IsCorrect: true, Kind: "image"},
				{Content: "TP. Hồ Chí Minh", IsCorrect: false, Kind: "image"},
				{Content: "Đà Nẵng", IsCorrect: false, Kind: "image"},
				{Content: "Huế", IsCorrect: false, Kind: "image"},
			},
		},
		{
			Kind:             "audio",
			QuestionType:     "fill_in_blanks",
			Title:            "Điền từ đúng vào chỗ trống.",
			Content:          "[blank1] là quốc gia lớn nhất Đông Nam Á về diện tích. Thủ đô của nó là [blank2].",
			Point:            10,
			TimeLimitSeconds: 30,
			AnswerPositions: []models.AnswerPosition{
				{Content: "Một quốc gia Đông Nam Á", IsCorrect: true, Kind: "text", Point: 5, CorrectPosition: 1},
				{Content: "Thủ đô của quốc gia đó", IsCorrect: true, Kind: "text", Point: 5, CorrectPosition: 2},
			},
		},
		{
			Kind:             "video",
			QuestionType:     "ordering",
			Title:            "Kéo các hành tinh vào đúng thứ tự.",
			Content:          "Sắp xếp các hành tinh theo thứ tự từ gần Mặt Trời nhất.",
			Point:            40,
			TimeLimitSeconds: 30,
			AnswerPositions: []models.AnswerPosition{
				{Content: "Sao Thủy", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 1},
				{Content: "Sao Kim", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 2},
				{Content: "Trái Đất", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 3},
				{Content: "Sao Hỏa", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 4},
			},
		},
		{
			Kind:             "image",
			QuestionType:     "matching",
			Title:            "Kéo mỗi quốc gia đến lá cờ tương ứng",
			Content:          "Ghép quốc gia với lá cờ tương ứng.",
			Point:            10,
			TimeLimitSeconds: 30,
			AnswerPositions: []models.AnswerPosition{
				{Content: "Việt Nam", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 1},
				{Content: "Nhật Bản", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 1},
				{Content: "Hàn Quốc", IsCorrect: true, Kind: "image", Point: 10, CorrectPosition: 1},
				{Content: "Việt Nam", IsCorrect: false, Kind: "image", Point: 10, CorrectPosition: 1},
			},
		},
		{
			Kind:             "image",
			QuestionType:     "drag_drop",
			Title:            "Kéo các từ đúng vào chỗ trống. Lưu ý: Có một số đáp án không sử dụng.",
			Content:          "[blank1] là hành tinh lớn nhất. [blank2] là hành tinh gần Mặt Trời nhất. [blank3] có vành đai nổi bật.",
			Point:            8,
			TimeLimitSeconds: 30,
			AnswerPositions: []models.AnswerPosition{
				{Content: "Sao Mộc", IsCorrect: true, Kind: "text", Point: 1, CorrectPosition: 1},
				{Content: "Sao Thủy", IsCorrect: true, Kind: "text", Point: 1, CorrectPosition: 2},
				{Content: "Sao Thổ", IsCorrect: true, Kind: "text", Point: 1, CorrectPosition: 3},
				{Content: "Sao Kim", IsCorrect: false, Kind: "text", Point: 1, CorrectPosition: 0},
				{Content: "Sao Hỏa", IsCorrect: false, Kind: "text", Point: 1, CorrectPosition: 0},
			},
		},
		{
			Kind:             "image",
			QuestionType:     "labeling",
			Title:            "Kéo nhãn đến đúng vị trí trên bông hoa. Lưu ý: Có một số nhãn không sử dụng.",
			Content:          "Gắn nhãn các phần của bông hoa.",
			Point:            10,
			TimeLimitSeconds: 30,
			ImageWidth:       500,
			ImageHeight:      250,
			AnswerCoordinates: []models.AnswerCoordinates{
				{Content: "Cánh hoa", Kind: "text", Point: 5, PositionX: 100, PositionY: 200, PositionWidth: 20, PositionHeight: 5},
				{Content: "Thân", Kind: "text", Point: 5, PositionX: 150, PositionY: 300, PositionWidth: 20, PositionHeight: 5},
				{Content: "Lá", Kind: "text", Point: 5, PositionX: 120, PositionY: 250, PositionWidth: 20, PositionHeight: 5},
				{Content: "Rễ", Kind: "text", Point: 5, PositionX: 0, PositionY: 0, PositionWidth: 20, PositionHeight: 5},
			},
		},
		{
			Kind:             "video",
			QuestionType:     "category",
			Title:            "Kéo mỗi động vật vào môi trường sống tương ứng.",
			Content:          "Phân loại động vật theo môi trường sống.",
			Point:            10,
			TimeLimitSeconds: 30,
			AnswerGroups: []models.AnswerGroup{
				{Content: "Sư tử", Group: models.GroupAnswer{Content: "Đất liền", Kind: "text"}, Kind: "image", Point: 2},
				{Content: "Cá heo", Group: models.GroupAnswer{Content: "Biển", Kind: "text"}, Kind: "image", Point: 2},
				{Content: "Chim đại bàng", Group: models.GroupAnswer{Content: "Bầu trời", Kind: "text"}, Kind: "image", Point: 2},
				{Content: "Cá mập", Group: models.GroupAnswer{Content: "Biển", Kind: "text"}, Kind: "image", Point: 2},
				{Content: "Hươu", Group: models.GroupAnswer{Content: "Đất liền", Kind: "text"}, Kind: "image", Point: 2},
				{Content: "Chim cánh cụt", Group: models.GroupAnswer{Content: "Biển", Kind: "text"}, Kind: "image", Point: 2},
			},
		},
		{
			Kind:             "text",
			QuestionType:     "speaking",
			Title:            "Đọc và trả lời câu hỏi.",
			Content:          "Bạn tên gì?",
			Point:            10,
			TimeLimitSeconds: 30,
		},
		{
			Kind:             "text",
			QuestionType:     "writing",
			Title:            "Đọc và trả lời câu hỏi.",
			Content:          "Quả gì mà lăn lốc lốc?",
			Point:            10,
			TimeLimitSeconds: 30,
		},
	}

	if err := db.MasterDB.Exec("DELETE FROM answers").Error; err != nil {
		log.Fatalf("Error deleting old answers: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM answer_groups").Error; err != nil {
		log.Fatalf("Error deleting old answer_groups: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM group_answers").Error; err != nil {
		log.Fatalf("Error deleting old group_answers: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM answer_positions").Error; err != nil {
		log.Fatalf("Error deleting old answer_positions: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM answer_coordinates").Error; err != nil {
		log.Fatalf("Error deleting old answer_coordinates: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM answer_matchings").Error; err != nil {
		log.Fatalf("Error deleting old answer_matchings: %v", err)
	}

	if err := db.MasterDB.Exec("DELETE FROM questions").Error; err != nil {
		log.Fatalf("Error deleting old questions: %v", err)
	}

	for _, q := range questions {
		if err := db.MasterDB.Create(&q).Error; err != nil {
			log.Printf("Error seeding question: %v", err)
		}
	}
}
func (s *Seeder) SeedRoles() {
	if err := db.MasterDB.Exec("DELETE FROM role_permissions").Error; err != nil {
		fmt.Println("Lỗi khi xóa bảng role_permissions:", err)
		return
	}
	if err := db.MasterDB.Exec("DELETE FROM permissions").Error; err != nil {
		fmt.Println("Lỗi khi xóa bảng permissions:", err)
		return
	}

	adminRole, err := s.ensureRole(models.AdminRoleId, "Admin", models.PageAdmin)
	if err != nil {
		fmt.Println("Lỗi role Admin:", err)
		return
	}
	teacherRole, err := s.ensureRole(models.TeacherRoleId, "Teacher", models.PageTeacher)
	if err != nil {
		fmt.Println("Lỗi role Teacher:", err)
		return
	}
	studentRole, err := s.ensureRole(models.StudentRoleId, "Student", models.PageStudent)
	if err != nil {
		fmt.Println("Lỗi role Student:", err)
		return
	}

	adminUser, err := s.ensureAdminUser("admin", "admin123", "Quản trị hệ thống")
	if err != nil {
		fmt.Println("Lỗi tạo user admin:", err)
		return
	}
	if err := s.ensureUserRole(adminUser.ID, adminRole.ID); err != nil {
		fmt.Println("Lỗi gán role admin:", err)
		return
	}

	s.seedPermissions(config.GetPermissions(), adminRole)
	s.seedPermissions(config.GetTeacherPermissions(), teacherRole)
	s.seedPermissions(config.GetStudentPermissions(), studentRole)
	s.seedPermissionString("internal.command", "system", "Lệnh nội bộ", adminRole)

	s.clearRolePermissionCache()

	fmt.Printf("✅ Seed roles & permissions xong. Đăng nhập: username=admin password=admin123 (user_id=%d, role_id=%d)\n",
		adminUser.ID, adminRole.ID)
}

func (s *Seeder) clearRolePermissionCache() {
	if db.RedisClient == nil {
		fmt.Println("ℹ️  Redis chưa bật — sau seed chạy: redis-cli DEL permissions:role:1 permissions:role:2 permissions:role:3")
		return
	}
	for _, roleID := range []int{int(models.AdminRoleId), int(models.TeacherRoleId), int(models.StudentRoleId)} {
		if err := redisperm.NewRoleRedis(roleID).ClearRolePermissionsCache(); err != nil {
			fmt.Printf("⚠️ Không xóa được cache role %d: %v\n", roleID, err)
		}
	}
	fmt.Println("✅ Đã làm mới cache permission trên Redis")
}

func (s *Seeder) ensureRole(id int64, name, defaultPageView string) (*models.Role, error) {
	var role models.Role
	err := db.MasterDB.First(&role, id).Error
	if err == nil {
		return &role, nil
	}
	role = models.Role{
		ID:              id,
		Name:            name,
		DefaultPageView: defaultPageView,
		Status:          true,
	}
	if err := db.MasterDB.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *Seeder) ensureAdminUser(username, password, displayName string) (*models.User, error) {
	var user models.User
	if err := db.MasterDB.Where("username = ?", username).First(&user).Error; err == nil {
		return &user, nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user = models.User{
		Username: username,
		Name:     displayName,
		Password: string(hashedPassword),
		Status:   true,
	}
	// Không set school_id=0 (vi phạm FK); để NULL
	if err := db.MasterDB.Select("Username", "Name", "Password", "Status").Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Seeder) ensureUserRole(userID, roleID int64) error {
	var ref models.UserRefRole
	err := db.MasterDB.Where("user_id = ? AND role_id = ?", userID, roleID).First(&ref).Error
	if err == nil {
		return nil
	}
	return db.MasterDB.Create(&models.UserRefRole{UserId: userID, RoleId: roleID}).Error
}

func (s *Seeder) seedPermissionString(permission, group, name string, role *models.Role) {
	var perm models.Permission
	db.MasterDB.Where(models.Permission{Permission: permission}).
		Assign(models.Permission{Group: group, Name: name}).
		FirstOrCreate(&perm)

	var rolePerm models.RolePermission
	if err := db.MasterDB.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rolePerm).Error; err != nil {
		db.MasterDB.Create(&models.RolePermission{
			RoleID:       role.ID,
			PermissionID: int64(perm.ID),
		})
	}
}

func (s *Seeder) seedPermissions(perms map[string]config.PermissionGroup, role *models.Role) {
	for key, val := range perms {
		for idx, action := range val.Actions {
			permissionStr := fmt.Sprintf("%s.%s", key, action)

			name := ""
			if idx < len(val.Names) {
				name = val.Names[idx]
			}

			display := true
			var perm models.Permission
			db.MasterDB.Where(models.Permission{Permission: permissionStr}).
				Assign(models.Permission{
					Group:        val.Group,
					Name:         name,
					SortPosition: val.SortPosition,
					IsDisplay:    &display,
				}).
				FirstOrCreate(&perm)

			var rolePerm models.RolePermission
			if err := db.MasterDB.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).First(&rolePerm).Error; err != nil {
				db.MasterDB.Create(&models.RolePermission{
					RoleID:       role.ID,
					PermissionID: int64(perm.ID),
				})
			}
		}
	}
}

func (s *Seeder) UserClassAndCourse() {
	if err := db.MasterDB.Exec("INSERT INTO user_classes (class_id, user_id, is_current) VALUES (1, 1, TRUE)").Error; err != nil {
		fmt.Println("Lỗi khi thêm dữ liệu vào bảng user_classes:", err)
		return
	}

	if err := db.MasterDB.Exec("INSERT INTO user_courses (course_id, user_id, is_current) VALUES (1, 1, TRUE)").Error; err != nil {
		fmt.Println("Lỗi khi thêm dữ liệu vào bảng user_courses:", err)
		return
	}
}

//go run database/seeder/seeder.go --model=Role
