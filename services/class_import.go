package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *classService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	rows, err := f.GetRows("Classes")
	if err != nil {
		return fmt.Errorf("cannot read 'Classes' sheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	facultyRepo := repositories.NewFacultyRepository()
	isVtg := config.LoadConfig().IsVtg
	facultyRepo.SetContext(c)
	faculties, err := facultyRepo.GetAll()

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		cID, _ := strconv.ParseInt(row[0], 10, 64)

		var schoolID, facultyID int64

		if isVtg {
			facultyID, _ = strconv.ParseInt(row[1], 10, 64)

			for _, faculty := range faculties {
				if faculty.ID == facultyID {
					schoolID = faculty.SchoolId
					break
				}
			}
		} else {
			schoolID, _ = strconv.ParseInt(row[1], 10, 64)
		}

		gradeID, _ := strconv.ParseInt(row[2], 10, 64)
		maxStudents, _ := strconv.ParseInt(row[4], 10, 32)
		status := false
		if row[8] == "1" || row[8] == "true" || row[8] == "TRUE" {
			status = true
		}

		teacherInfo := models.TeacherInfo{
			Name:  row[5],
			Phone: row[6],
			Email: row[7],
		}

		class := models.Class{
			ID:          cID,
			SchoolId:    schoolID,
			FacultyId:   facultyID,
			GradeId:     gradeID,
			Name:        row[3],
			MaxStudents: int32(maxStudents),
			TeacherInfo: teacherInfo,
			Status:      status,
		}

		_, err := s.repo.CreateOrUpdateClass(&class)

		if err != nil {
			config.Log.Error("Import class", class)
		}
	}

	return nil
}
