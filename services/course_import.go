package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

func (s *courseService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	rows, err := f.GetRows("Courses")
	if err != nil {
		return fmt.Errorf("cannot read 'Courses' sheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		cID, _ := strconv.ParseInt(row[0], 10, 64)
		programID, _ := strconv.ParseInt(row[1], 10, 64)

		name := ""
		if len(row) > 2 && !isEmpty(row[2]) {
			name = row[2]
		}

		description := ""
		if len(row) > 3 && !isEmpty(row[3]) {
			description = row[3]
		}

		courseType := ""
		if len(row) > 4 && !isEmpty(row[4]) {
			courseType = row[4]
		}

		timeValue := ""
		if len(row) > 5 && !isEmpty(row[5]) {
			timeValue = row[5]
		}

		target := ""
		if len(row) > 6 && !isEmpty(row[6]) {
			target = row[6]
		}

		var duration int
		if len(row) > 7 && !isEmpty(row[7]) {
			duration, _ = strconv.Atoi(strings.TrimSpace(row[7]))
		}

		var startDate, endDate time.Time
		if len(row) > 8 && !isEmpty(row[8]) {
			startDate, _ = parseExcelDate(row[8])
		}
		if len(row) > 9 && !isEmpty(row[9]) {
			endDate, _ = parseExcelDate(row[9])
		}

		var fileInfo models.MediaInfo
		if len(row) > 10 && !isEmpty(row[10]) {
			fileUrl := utils.StripDomain(row[10], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo = mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		}

		course := models.Course{
			ID:          cID,
			ProgramId:   programID,
			Name:        name,
			Description: description,
			Type:        courseType,
			Time:        timeValue,
			Target:      target,
			StartDate:   startDate,
			EndDate:     endDate,
			ImageInfo:   fileInfo,
			Duration:    duration,
			Status:      true,
		}

		_, err := s.repo.CreateOrUpdateCourse(&course)

		if err != nil {
			config.Log.Error("Import course", course)
		}
	}

	return nil
}

func parseExcelDate(v string) (time.Time, error) {
	if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
		t, err := excelize.ExcelDateToTime(floatVal, false)
		if err == nil {
			return t, nil
		}
	}

	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
		"02-01-2006",
		"01-02-2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse date: %s", v)
}

func isEmpty(v string) bool {
	return strings.TrimSpace(v) == "" ||
		v == "null" ||
		v == "NULL" ||
		v == "-" ||
		v == "--"
}

func (s *courseService) ImportUsers(c *gin.Context, courseId int64, fileHeader *multipart.FileHeader) error {
	course, err := s.repo.FindByID(int(courseId))
	if err != nil {
		return fmt.Errorf("course not found: %w", err)
	}

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

	userRepo := repositories.NewUserRepository()
	userRepo.SetContext(c)
	classRepo := repositories.NewClassRepository()
	classRepo.SetContext(c)
	roleRepo := repositories.NewRoleRepository()
	roleRepo.SetContext(c)

	userIds := make([]int64, 0)
	createdCount := 0
	updatedCount := 0

	for _, row := range rows[1:] {
		if len(row) < 5 {
			continue
		}

		id, _ := strconv.ParseInt(strings.TrimSpace(row[0]), 10, 64)
		identifier := strings.TrimSpace(row[1])
		username := strings.TrimSpace(row[2])
		code := strings.TrimSpace(row[3])
		name := strings.TrimSpace(row[4])
		email := ""
		phone := ""
		roleIdStr := ""
		password := ""
		avatar := ""
		status := false
		classIdsStr := ""

		if len(row) > 5 {
			email = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			phone = strings.TrimSpace(row[6])
		}
		if len(row) > 7 {
			roleIdStr = strings.TrimSpace(row[7])
		}
		if len(row) > 8 {
			password = strings.TrimSpace(row[8])
		}
		if len(row) > 9 {
			avatar = strings.TrimSpace(row[9])
		}
		if len(row) > 10 {
			statusStr := strings.TrimSpace(row[10])
			status = statusStr == "1" || statusStr == "true" || statusStr == "TRUE"
		}
		if len(row) > 11 {
			classIdsStr = strings.TrimSpace(row[11])
		}

		if name == "" && username == "" {
			continue
		}

		var userId int64
		var isNewUser bool

		if id > 0 {
			user, err := userRepo.FindByID(int(id))
			if err == nil && user != nil {
				userId = int64(user.ID)
				isNewUser = false

				if username != "" {
					user.Username = username
				}
				if identifier != "" {
					user.Identifier = identifier
				}
				if code != "" {
					user.Code = code
				}
				if name != "" {
					user.Name = name
				}
				if email != "" {
					user.Email = email
				}
				if phone != "" {
					user.PhoneNumber = phone
				}
				if avatar != "" {
					fileUrl := utils.StripDomain(avatar, models.Storage)
					mediaRepo := repositories.NewMediaRepository()
					user.AvatarInfo = mediaRepo.GetMediaInfo(fileUrl, models.Storage)
				}
				user.Status = status

				if password != "" {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
					if err == nil {
						user.Password = string(hashedPassword)
					}
				}

				// Parse Is Independent Student và Is Failed Subject nếu isVtg = true
				if isVtg {
					if len(row) > 12 && strings.TrimSpace(row[12]) != "" {
						isIndependentStudent := false
						if row[12] == "1" || row[12] == "true" || row[12] == "TRUE" {
							isIndependentStudent = true
						}
						user.IsIndependentStudent = isIndependentStudent
					}

					if len(row) > 13 && strings.TrimSpace(row[13]) != "" {
						isFailedSubject := false
						if row[13] == "1" || row[13] == "true" || row[13] == "TRUE" {
							isFailedSubject = true
						}
						user.IsFailedSubject = isFailedSubject
					}
				}

				_, err = userRepo.CreateOrUpdateUser(user)
				if err != nil {
					config.Log.Errorf("Failed to update user %d: %v", id, err)
					continue
				}
				updatedCount++
			} else {
				isNewUser = true
			}
		} else {
			isNewUser = true
		}

		if isNewUser {
			if username == "" {
				if name == "" {
					continue
				}
				schoolCode := ""
				if classIdsStr != "" {
					classIds := utils.ParseIDs(classIdsStr)
					if len(classIds) > 0 {
						classID := classIds[0]
						classRepo.SetPreload([]string{"School"})
						classStudent, _ := classRepo.FindByID(int(classID))
						if classStudent != nil && classStudent.School != nil {
							schoolCode = classStudent.School.ShortName
						}
					}
				}

				roleId := 0
				if roleIdStr != "" {
					roleId, _ = strconv.Atoi(roleIdStr)
				}

				username = generateUsernameForImport(name, schoolCode, roleId, userRepo)
			}

			if password == "" {
				if identifier != "" {
					password = identifier
				} else if code != "" {
					password = code
				} else {
					password = username
				}
			}

			var fileInfo models.MediaInfo
			if avatar != "" {
				fileUrl := utils.StripDomain(avatar, models.Storage)
				mediaRepo := repositories.NewMediaRepository()
				fileInfo = mediaRepo.GetMediaInfo(fileUrl, models.Storage)
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				config.Log.Warnf("Failed to hash password for user %s, skipping", username)
				continue
			}

			schoolId := 0
			if classIdsStr != "" {
				classIds := utils.ParseIDs(classIdsStr)
				if len(classIds) > 0 {
					classID := classIds[0]
					classRepo.SetPreload([]string{"School"})
					classStudent, _ := classRepo.FindByID(int(classID))
					if classStudent != nil && classStudent.SchoolId > 0 {
						schoolId = int(classStudent.SchoolId)
					}
				}
			}

			user := models.User{
				ID:          0,
				Username:    username,
				Identifier:  identifier,
				Code:        code,
				Name:        name,
				Email:       email,
				PhoneNumber: phone,
				AvatarInfo:  fileInfo,
				Status:      status,
				SchoolID:    schoolId,
				Password:    string(hashedPassword),
			}

			// Parse Is Independent Student và Is Failed Subject nếu isVtg = true
			if isVtg {
				if len(row) > 12 && strings.TrimSpace(row[12]) != "" {
					isIndependentStudent := false
					if row[12] == "1" || row[12] == "true" || row[12] == "TRUE" {
						isIndependentStudent = true
					}
					user.IsIndependentStudent = isIndependentStudent
				}

				if len(row) > 13 && strings.TrimSpace(row[13]) != "" {
					isFailedSubject := false
					if row[13] == "1" || row[13] == "true" || row[13] == "TRUE" {
						isFailedSubject = true
					}
					user.IsFailedSubject = isFailedSubject
				}
			}

			userId, err = userRepo.CreateOrUpdateUser(&user)
			if err != nil {
				config.Log.Errorf("Failed to create user %s: %v", username, err)
				continue
			}
			createdCount++

			if roleIdStr != "" {
				roleId, err := strconv.ParseInt(roleIdStr, 10, 64)
				if err == nil && roleId > 0 {
					roleRepo.AddUserRole(roleId, []int64{userId})
				}
			}
		}

		userIds = append(userIds, userId)
	}

	if len(userIds) == 0 {
		return errors.New("no valid users to import")
	}

	if err := s.repo.ReplaceUserCourse(courseId, userIds, 0, 0, []int64{}); err != nil {
		return fmt.Errorf("failed to replace users in course: %w", err)
	}

	config.Log.Infof("Successfully imported %d users to course %d (ID: %d). Created: %d, Updated: %d",
		len(userIds), courseId, course.ID, createdCount, updatedCount)

	return nil
}

func generateUsernameForImport(name string, schoolName string, roleId int, userRepo repositories.UserRepository) string {
	cleanName := utils.RemoveAccents(name)
	cleanName = strings.ToLower(cleanName)
	cleanName = strings.Join(strings.Fields(cleanName), " ")

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

	parts := strings.Fields(cleanName)
	if len(parts) == 0 {
		return ""
	}

	mainName := parts[len(parts)-1]

	var initials string
	if len(parts) > 1 {
		for _, p := range parts[:len(parts)-1] {
			runes := []rune(p)
			if len(runes) > 0 {
				initials += string(runes[0])
			}
		}
	}

	var baseUsername string
	if roleId == 2 {
		baseUsername = fmt.Sprintf("%s%s%s", mainName, initials, "gv")
	} else {
		baseUsername = fmt.Sprintf("%s%s%s", mainName, initials, schoolCode)
	}

	index := 1
	username := ""

	for {
		candidate := baseUsername
		if index > 1 {
			candidate = fmt.Sprintf("%s%d", baseUsername, index)
		}

		exists, _ := userRepo.UsernameExits(candidate, 0)
		if !exists {
			username = candidate
			break
		}

		index++
	}

	return username
}
