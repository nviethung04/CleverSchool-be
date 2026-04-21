package services

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

func (s *userService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	rows, err := f.GetRows("Users")
	if err != nil {
		return fmt.Errorf("cannot read 'Users' sheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	isVtg := config.LoadConfig().IsVtg

	courseUserMap := make(map[int64][]int64)
	classUserMap := make(map[int64][]int64)
	roleUserMap := make(map[int64][]int64)

	classRepo := repositories.NewClassRepository()

	userIds := make([]int64, 0, len(rows)-1)

	for _, row := range rows[1:] {
		if len(row) < 9 {
			continue
		}

		fileUrl := utils.StripDomain(row[9], models.Storage)
		mediaRepo := repositories.NewMediaRepository()
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		status := false
		if row[10] == "1" || row[10] == "true" || row[10] == "TRUE" {
			status = true
		}

		schoolId := 0
		schoolCode := ""

		if len(row) > 11 && strings.TrimSpace(row[11]) != "" {
			classID, _ := strconv.ParseInt(row[11], 10, 64)

			if classID > 0 {
				classRepo.SetPreload([]string{
					"School",
				})
				classStudent, _ := classRepo.FindByID(int(classID))

				if classStudent != nil && classStudent.SchoolId > 0 && classStudent.School != nil {
					schoolId = int(classStudent.SchoolId)
					schoolCode = classStudent.School.ShortName
				}
			}
		}

		rawRoleId := row[7]
		parsedRoleId, err := strconv.ParseInt(rawRoleId, 10, 64)
		if err != nil {
			continue
		}
		roleId := int(parsedRoleId)

		id, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			id = 0
		}

		username := row[2]
		name := row[4]
		isNewUser := false

		if username == "" {
			if name == "" {
				continue
			}

			username = s.GenerateUsername(name, schoolCode, roleId)
			isNewUser = true
		}

		identifier := row[1]
		code := row[3]

		user := models.User{
			ID:          id,
			Username:    username,
			Identifier:  identifier,
			Code:        code,
			Name:        name,
			Email:       row[5],
			PhoneNumber: row[6],
			AvatarInfo:  fileInfo,
			Status:      status,
			SchoolID:    schoolId,
		}

		// Parse Is Independent Student và Is Failed Subject nếu isVtg = true
		if isVtg {
			if len(row) > 15 && strings.TrimSpace(row[15]) != "" {
				isIndependentStudent := false
				if row[15] == "1" || row[15] == "true" || row[15] == "TRUE" {
					isIndependentStudent = true
				}
				user.IsIndependentStudent = isIndependentStudent
			}

			if len(row) > 16 && strings.TrimSpace(row[16]) != "" {
				isFailedSubject := false
				if row[16] == "1" || row[16] == "true" || row[16] == "TRUE" {
					isFailedSubject = true
				}
				user.IsFailedSubject = isFailedSubject
			}
		}

		if row[8] != "" || isNewUser {
			userPassword := row[8]
			if userPassword == "" {
				if identifier != "" {
					userPassword = identifier
				} else if code != "" {
					userPassword = code
				} else {
					userPassword = username
				}
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)

			if err != nil {
				user.Password = ""
			}
			user.Password = string(hashedPassword)
		}

		userId, err := s.repo.CreateOrUpdateUser(&user)

		if err != nil {
			config.Log.Error("Import user", user)
		} else {
			userIds = append(userIds, userId)
			if len(row) > 11 && strings.TrimSpace(row[11]) != "" {
				classIds := utils.ParseIDs(row[11])

				for _, classID := range classIds {
					classUserMap[classID] = append(classUserMap[classID], userId)
				}
			}
			if len(row) > 12 && strings.TrimSpace(row[12]) != "" {
				courseIds := utils.ParseIDs(row[12])

				for _, courseID := range courseIds {
					courseUserMap[courseID] = append(courseUserMap[courseID], userId)
				}
			}

			if roleId > 0 {
				roleUserMap[int64(roleId)] = append(roleUserMap[int64(roleId)], userId)
			}
		}
	}

	courseRepo := repositories.NewCourseRepository()
	roleRepo := repositories.NewRoleRepository()

	if len(userIds) > 0 {
		s.repo.DeleteOldClassCourseRole(userIds)
	}

	for classId, userIds := range classUserMap {
		classRepo.AddUserClass(classId, userIds, 0)
	}

	for courseId, userIds := range courseUserMap {
		courseRepo.AddUserCourse(courseId, userIds, 0, []int64{})
	}

	for roleId, userIds := range roleUserMap {
		roleRepo.AddUserRole(roleId, userIds)
	}

	return nil
}

func (s *userService) GenerateUsername(name string, schoolName string, roleId int) string {
	// Xử lý tên người: bỏ dấu, chuyển thành chữ thường, loại bỏ khoảng trắng thừa
	cleanName := utils.RemoveAccents(name)
	cleanName = strings.ToLower(cleanName)
	cleanName = strings.Join(strings.Fields(cleanName), " ")

	// Xử lý schoolCode: nếu nhiều từ lấy chữ cái đầu mỗi từ, nếu 1 từ lấy nguyên từ
	schoolNameClean := utils.RemoveAccents(schoolName)
	words := strings.Fields(schoolNameClean)

	var schoolCode string
	if len(words) == 1 {
		schoolCode = strings.ToLower(words[0])
	} else {
		for _, w := range words {
			runes := []rune(w)
			if len(runes) > 0 {
				schoolCode += string(runes[0])
			}
		}
		schoolCode = strings.ToLower(schoolCode)
	}

	// Tách tên người thành từng phần
	parts := strings.Fields(cleanName)
	if len(parts) == 0 {
		return ""
	}

	// Lấy họ/chữ cuối làm mainName
	mainName := parts[len(parts)-1]

	// Lấy chữ cái đầu của các phần còn lại làm initials
	var initials string
	if len(parts) > 1 {
		for _, p := range parts[:len(parts)-1] {
			runes := []rune(p)
			if len(runes) > 0 {
				initials += string(runes[0])
			}
		}
	}

	// Tạo baseUsername: nếu roleId == 2 thì không nối schoolCode
	var baseUsername string
	if roleId == 2 {
		baseUsername = fmt.Sprintf("%s%s%s", mainName, initials, "gv")
	} else {
		baseUsername = fmt.Sprintf("%s%s%s", mainName, initials, schoolCode)
	}

	index := 1
	username := ""

	// Kiểm tra username tồn tại hay chưa, nếu có thì thêm số đếm
	for {
		candidate := baseUsername
		if index > 1 {
			candidate = fmt.Sprintf("%s%d", baseUsername, index)
		}

		exists, _ := s.repo.UsernameExits(candidate, 0)
		if !exists {
			username = candidate
			break
		}

		index++
	}

	return username
}
