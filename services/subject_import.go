package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"be-Clever School/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *subjectService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}
	defer f.Close()

	// Get all sheet names
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return fmt.Errorf("excel file has no sheets")
	}

	// Try to read "Subjects" sheet first, if not found try "Programs" (for VTG)
	var rows [][]string
	var sheetName string

	rows, err = f.GetRows("Subjects")
	if err == nil {
		sheetName = "Subjects"
	} else {
		rows, err = f.GetRows("Programs")
		if err == nil {
			sheetName = "Programs"
		} else {
			// If both fail, try to use the first sheet
			if len(sheetList) > 0 {
				sheetName = sheetList[0]
				rows, err = f.GetRows(sheetName)
				if err != nil {
					return fmt.Errorf("cannot read sheet '%s' or any other sheet: available sheets are %v: %w", sheetName, sheetList, err)
				}
			} else {
				return fmt.Errorf("cannot read 'Subjects' or 'Programs' sheet: available sheets are %v: %w", sheetList, err)
			}
		}
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	s.repo.SetContext(c)

	for _, row := range rows[1:] {
		if len(row) < 3 {
			continue
		}

		id, _ := strconv.ParseInt(row[0], 10, 64)
		facultyID, _ := strconv.ParseInt(row[1], 10, 64)

		var trainingLevelIds []int64

		if len(row) > 2 && row[2] != "" {
			trainingLevelIdsStr := strings.TrimSpace(row[2])
			if strings.Contains(trainingLevelIdsStr, ",") || (len(trainingLevelIdsStr) > 0 && isNumeric(trainingLevelIdsStr)) {
				ids := strings.Split(trainingLevelIdsStr, ",")
				for _, idStr := range ids {
					idStr = strings.TrimSpace(idStr)
					if idStr != "" {
						if levelID, err := strconv.ParseInt(idStr, 10, 64); err == nil && levelID > 0 {
							trainingLevelIds = append(trainingLevelIds, levelID)
						}
					}
				}
			}
		}

		subject := models.Subject{
			ID:          id,
			FacultyId:   facultyID,
			Name:        row[3],
			Description: "",
			Status:      true,
		}

		if len(row) > 4 {
			subject.Description = row[4]
		}

		var subjectID int64
		if id > 0 {
			if err := s.repo.Update(&subject); err != nil {
				if err := s.repo.Create(&subject); err == nil {
					subjectID = subject.ID
				}
			} else {
				subjectID = id
			}
		} else {
			if err := s.repo.Create(&subject); err == nil {
				subjectID = subject.ID
			}
		}

		if subjectID > 0 && len(trainingLevelIds) > 0 {
			_ = s.repo.UpdateTrainingLevels(subjectID, trainingLevelIds)
		}
	}

	return nil
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
