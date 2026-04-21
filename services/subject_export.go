package services

import (
	"be-lms/config"
	"be-lms/repositories"
	"be-lms/utils"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *subjectService) Export(c *gin.Context) (string, error) {
	facultyRepo := repositories.NewFacultyRepository()
	trainingLevelRepo := repositories.NewTrainingLevelRepository()

	allowedFilters := []string{"status"}

	filter, _, _, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)
	s.repo.SetPreload([]string{
		"Faculty",
		"TrainingLevels:deleted_at IS NULL",
	})

	subjects, _, err := s.repo.FindAll()
	if err != nil {
		return "", err
	}

	facultyRepo.SetContext(c)
	faculties, err := facultyRepo.GetAll()
	if err != nil {
		return "", err
	}

	trainingLevelRepo.SetContext(c)
	trainingLevels, err := trainingLevelRepo.GetAll()
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

	// ===== Sheet: Subjects =====
	isVtg := config.LoadConfig().IsVtg
	subjectSheet := "Subjects"
	if isVtg {
		subjectSheet = "Programs"
	}
	f.SetSheetName("Sheet1", subjectSheet)

	subjectHeaders := []string{
		"ID", "Faculty Id", "Training Level Ids",
		"Name", "Description",
	}
	for i, h := range subjectHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(subjectSheet, cell, h)
		f.SetCellStyle(subjectSheet, cell, cell, headerStyle)
	}
	for idx, s := range subjects {
		row := idx + 2

		var trainingLevelStrs []string

		if len(s.TrainingLevels) > 0 {
			for _, tl := range s.TrainingLevels {
				if tl.ID > 0 {
					trainingLevelStrs = append(trainingLevelStrs, fmt.Sprintf("%d", tl.ID))
				}
			}
		}

		trainingLevelIdsStr := strings.Join(trainingLevelStrs, ",")

		f.SetCellValue(subjectSheet, fmt.Sprintf("A%d", row), s.ID)
		f.SetCellValue(subjectSheet, fmt.Sprintf("B%d", row), s.FacultyId)
		f.SetCellValue(subjectSheet, fmt.Sprintf("C%d", row), trainingLevelIdsStr)
		f.SetCellValue(subjectSheet, fmt.Sprintf("D%d", row), s.Name)
		f.SetCellValue(subjectSheet, fmt.Sprintf("E%d", row), s.Description)
	}

	for idx := range subjects {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(subjectHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(subjectSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(subjectSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Faculties (chỉ export) =====
	facultySheet := "Faculties"
	f.NewSheet(facultySheet)

	facultyHeaders := []string{"ID", "Name", "Code"}
	for i, h := range facultyHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(facultySheet, cell, h)
		f.SetCellStyle(facultySheet, cell, cell, headerStyle)
	}
	for idx, faculty := range faculties {
		row := idx + 2
		f.SetCellValue(facultySheet, fmt.Sprintf("A%d", row), faculty.ID)
		f.SetCellValue(facultySheet, fmt.Sprintf("B%d", row), faculty.Name)
		f.SetCellValue(facultySheet, fmt.Sprintf("C%d", row), faculty.Code)
	}

	for idx := range faculties {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(facultyHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(facultySheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(facultySheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Training Levels (chỉ export) =====
	trainingLevelSheet := "TrainingLevels"
	f.NewSheet(trainingLevelSheet)

	trainingLevelHeaders := []string{"ID", "Name", "Description"}
	for i, h := range trainingLevelHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(trainingLevelSheet, cell, h)
		f.SetCellStyle(trainingLevelSheet, cell, cell, headerStyle)
	}
	for idx, trainingLevel := range trainingLevels {
		row := idx + 2
		f.SetCellValue(trainingLevelSheet, fmt.Sprintf("A%d", row), trainingLevel.ID)
		f.SetCellValue(trainingLevelSheet, fmt.Sprintf("B%d", row), trainingLevel.Name)
		f.SetCellValue(trainingLevelSheet, fmt.Sprintf("C%d", row), trainingLevel.Description)
	}

	for idx := range faculties {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(facultyHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(facultySheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(facultySheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Save file to S3 =====
	var filename string
	if isVtg {
		filename = fmt.Sprintf("programs_%d.xlsx", time.Now().Unix())
	} else {
		filename = fmt.Sprintf("subjects_%d.xlsx", time.Now().Unix())
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

	config.Log.Info("Export subject Excel to S3:", fileURL)

	return fileURL, nil
}
