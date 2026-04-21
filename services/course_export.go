package services

import (
	"be-lms/config"
	"be-lms/repositories"
	"be-lms/utils"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *courseService) Export(c *gin.Context) (string, error) {
	programRepo := repositories.NewProgramRepository()
	isVtg := config.LoadConfig().IsVtg

	allowedFilters := []string{"status", "program_id"}

	filter, _, _, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)

	programRepo.SetContext(c)
	programs, err := programRepo.GetAll()
	if err != nil {
		return "", err
	}

	courses, err := s.repo.GetAll()
	if err != nil {
		return "", err
	}

	f := excelize.NewFile()

	// === Style header ===
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
		},
	})

	// === Style row xen kẽ ===
	oddRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	if err != nil {
		config.Log.Error("Failed to create oddRowStyle:", err)
		return "", err
	}

	evenRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		config.Log.Error("Failed to create evenRowStyle:", err)
		return "", err
	}

	// ===== Sheet: Courses =====
	classSheet := "Courses"
	f.SetSheetName("Sheet1", classSheet)

	courseHeaders := []string{}

	if isVtg {
		courseHeaders = []string{
			"ID", "Subject Id", "Name", "Description", "Type",
			"Time", "Target", "Duration", "Start Date", "End Date", "Image",
		}
	} else {
		courseHeaders = []string{
			"ID", "Program Id", "Name", "Description", "Type",
			"Time", "Target", "Duration", "Start Date", "End Date", "Image",
		}
	}

	for i, h := range courseHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(classSheet, cell, h)
		f.SetCellStyle(classSheet, cell, cell, headerStyle)
	}
	for idx, c := range courses {
		row := idx + 2
		f.SetCellValue(classSheet, fmt.Sprintf("A%d", row), c.ID)
		f.SetCellValue(classSheet, fmt.Sprintf("B%d", row), c.ProgramId)
		f.SetCellValue(classSheet, fmt.Sprintf("c%d", row), c.Name)
		f.SetCellValue(classSheet, fmt.Sprintf("D%d", row), c.Description)
		f.SetCellValue(classSheet, fmt.Sprintf("E%d", row), c.Type)
		f.SetCellValue(classSheet, fmt.Sprintf("F%d", row), c.Time)
		f.SetCellValue(classSheet, fmt.Sprintf("G%d", row), c.Target)
		f.SetCellValue(classSheet, fmt.Sprintf("H%d", row), c.Duration)
		f.SetCellValue(classSheet, fmt.Sprintf("I%d", row), c.StartDate.Format("2006-01-02"))
		f.SetCellValue(classSheet, fmt.Sprintf("J%d", row), c.EndDate.Format("2006-01-02"))
		f.SetCellValue(classSheet, fmt.Sprintf("K%d", row), c.ImageInfo.Path)
	}

	for idx := range courses {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(courseHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(classSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(classSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Programs =====
	programSheet := "Programs"
	if isVtg {
		programSheet = "Subjects"
	}
	f.NewSheet(programSheet)

	gradeHeaders := []string{"ID", "Name"}
	for i, h := range gradeHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(programSheet, cell, h)
		f.SetCellStyle(programSheet, cell, cell, headerStyle)
	}
	for idx, program := range programs {
		row := idx + 2
		f.SetCellValue(programSheet, fmt.Sprintf("A%d", row), program.ID)
		f.SetCellValue(programSheet, fmt.Sprintf("B%d", row), program.Name)
	}

	for idx := range programs {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(gradeHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(programSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(programSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("courses_%d.xlsx", time.Now().Unix())
	if isVtg {
		filename = fmt.Sprintf("credit_class_%d.xlsx", time.Now().Unix())
	}

	// Save Excel file to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", fmt.Errorf("failed to write Excel to buffer: %v", err)
	}

	// Upload to S3
	fileURL, err := UploadToS3(buf.Bytes(), filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	config.Log.Info("Export course Excel to S3:", fileURL)

	return fileURL, nil
}

// ExportUsers xuất danh sách user theo khóa học
func (s *courseService) ExportUsers(c *gin.Context, courseId int64) (string, error) {
	// Lấy thông tin khóa học
	course, err := s.repo.FindByID(int(courseId))
	if err != nil {
		return "", fmt.Errorf("course not found: %w", err)
	}

	// Lấy danh sách user trong khóa học (tất cả roles)
	users, _, err := s.repo.GetUsers(courseId, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get users: %w", err)
	}

	roleRepo := repositories.NewRoleRepository()
	roleRepo.SetContext(c)
	roles, err := roleRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get roles: %w", err)
	}

	// Lấy danh sách user classes và classes
	userRepo := repositories.NewUserRepository()
	userRepo.SetContext(c)
	userClasses, err := userRepo.GetAllUserClasses()
	if err != nil {
		return "", fmt.Errorf("failed to get user classes: %w", err)
	}

	classRepo := repositories.NewClassRepository()
	classRepo.SetContext(c)
	classes, err := classRepo.GetAll()
	if err != nil {
		return "", fmt.Errorf("failed to get classes: %w", err)
	}

	// Tạo map: userID -> []ClassID / []ClassName
	userClassMap := make(map[int64][]int64)
	userClassNameMap := make(map[int64][]string)
	for _, uc := range userClasses {
		userClassMap[uc.UserId] = append(userClassMap[uc.UserId], uc.ClassId)
		for _, c := range classes {
			if c.ID == uc.ClassId {
				userClassNameMap[uc.UserId] = append(userClassNameMap[uc.UserId], c.Name)
			}
		}
	}

	f := excelize.NewFile()

	// === Style header ===
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
		},
	})
	if err != nil {
		return "", err
	}

	// === Style row xen kẽ ===
	oddRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		config.Log.Error("Failed to create oddRowStyle:", err)
		return "", err
	}

	evenRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		config.Log.Error("Failed to create evenRowStyle:", err)
		return "", err
	}

	// ===== Sheet: Users =====
	userSheet := "Users"
	f.SetSheetName("Sheet1", userSheet)

	isVtg := config.LoadConfig().IsVtg

	userHeaders := []string{
		"ID", "Identifier", "Username",
		"Code", "Name", "Email",
		"Phone", "Role ID", "Password",
		"Avatar", "Status", "Class IDs",
		"Class Names",
	}

	if isVtg {
		userHeaders = append(userHeaders, "Is Independent Student", "Is Failed Subject")
	}

	for i, h := range userHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(userSheet, cell, h)
		f.SetCellStyle(userSheet, cell, cell, headerStyle)
	}

	for idx, u := range users {
		row := idx + 2

		var roleId int
		for _, r := range u.Roles {
			roleId = int(r.ID)
			break
		}

		f.SetCellValue(userSheet, fmt.Sprintf("A%d", row), u.ID)
		f.SetCellValue(userSheet, fmt.Sprintf("B%d", row), u.Identifier)
		f.SetCellValue(userSheet, fmt.Sprintf("C%d", row), u.Username)
		f.SetCellValue(userSheet, fmt.Sprintf("D%d", row), u.Code)
		f.SetCellValue(userSheet, fmt.Sprintf("E%d", row), u.Name)
		f.SetCellValue(userSheet, fmt.Sprintf("F%d", row), u.Email)
		f.SetCellValue(userSheet, fmt.Sprintf("G%d", row), u.PhoneNumber)
		f.SetCellValue(userSheet, fmt.Sprintf("H%d", row), roleId)
		f.SetCellValue(userSheet, fmt.Sprintf("I%d", row), "")
		f.SetCellValue(userSheet, fmt.Sprintf("J%d", row), u.AvatarInfo.Path)
		f.SetCellValue(userSheet, fmt.Sprintf("K%d", row), u.Status)

		// Class IDs và Class Names
		classIDs := userClassMap[int64(u.ID)]
		classNames := userClassNameMap[int64(u.ID)]

		var classStrs []string
		for _, cid := range classIDs {
			classStrs = append(classStrs, strconv.FormatInt(cid, 10))
		}

		classNameStrs := append([]string{}, classNames...)

		f.SetCellValue(userSheet, fmt.Sprintf("L%d", row), strings.Join(classStrs, ","))
		f.SetCellValue(userSheet, fmt.Sprintf("M%d", row), strings.Join(classNameStrs, ","))

		if isVtg {
			f.SetCellValue(userSheet, fmt.Sprintf("N%d", row), u.IsIndependentStudent)
			f.SetCellValue(userSheet, fmt.Sprintf("O%d", row), u.IsFailedSubject)
		}
	}

	// Áp dụng style xen kẽ
	for idx := range users {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(userHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(userSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(userSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Roles =====
	roleSheet := "Roles"
	f.NewSheet(roleSheet)

	roleHeaders := []string{"ID", "Name"}

	for i, h := range roleHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(roleSheet, cell, h)
		f.SetCellStyle(roleSheet, cell, cell, headerStyle)
	}
	for idx, role := range roles {
		row := idx + 2
		f.SetCellValue(roleSheet, fmt.Sprintf("A%d", row), role.ID)
		f.SetCellValue(roleSheet, fmt.Sprintf("B%d", row), role.Name)
	}

	for idx := range roles {
		row := idx + 2
		for col := 'A'; col <= 'B'; col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(roleSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(roleSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Course Info =====
	courseSheet := "Course Info"
	f.NewSheet(courseSheet)

	courseHeaders := []string{"ID", "Name", "Description"}
	for i, h := range courseHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(courseSheet, cell, h)
		f.SetCellStyle(courseSheet, cell, cell, headerStyle)
	}

	f.SetCellValue(courseSheet, "A2", course.ID)
	f.SetCellValue(courseSheet, "B2", course.Name)
	f.SetCellValue(courseSheet, "C2", course.Description)

	for col := 'A'; col <= 'C'; col++ {
		cell := fmt.Sprintf("%c2", col)
		f.SetCellStyle(courseSheet, cell, cell, evenRowStyle)
	}

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("course_%d_users_%d.xlsx", courseId, time.Now().Unix())
	if isVtg {
		filename = fmt.Sprintf("credit_class_%d_users_%d.xlsx", courseId, time.Now().Unix())
	}

	// Save Excel file to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", fmt.Errorf("failed to write Excel to buffer: %v", err)
	}

	// Upload to S3
	fileURL, err := UploadToS3(buf.Bytes(), filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	config.Log.Info("Export course users Excel to S3:", fileURL)

	return fileURL, nil
}
