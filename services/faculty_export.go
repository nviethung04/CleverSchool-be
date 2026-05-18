package services

import (
	"be-Clever School/config"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *facultyService) Export(c *gin.Context) (string, error) {
	schoolRepo := repositories.NewSchoolRepository()

	allowedFilters := []string{"status", "school_id"}

	filter, _, _, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	s.repo.SetContext(c)
	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)

	schoolRepo.SetContext(c)
	schools, err := schoolRepo.GetAll()
	if err != nil {
		return "", err
	}

	faculties, err := s.repo.GetAll()
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

	// ===== Sheet: Faculties =====
	facultySheet := "Faculties"
	f.SetSheetName("Sheet1", facultySheet)

	facultyHeaders := []string{
		"ID", "School Id", "Name",
		"Code", "Name head", "Description", "Avatar",
	}

	for i, h := range facultyHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(facultySheet, cell, h)
		f.SetCellStyle(facultySheet, cell, cell, headerStyle)
	}
	for idx, fa := range faculties {
		row := idx + 2
		f.SetCellValue(facultySheet, fmt.Sprintf("A%d", row), fa.ID)
		f.SetCellValue(facultySheet, fmt.Sprintf("B%d", row), fa.SchoolId)
		f.SetCellValue(facultySheet, fmt.Sprintf("C%d", row), fa.Name)
		f.SetCellValue(facultySheet, fmt.Sprintf("D%d", row), fa.Code)
		f.SetCellValue(facultySheet, fmt.Sprintf("E%d", row), fa.NameHead)
		f.SetCellValue(facultySheet, fmt.Sprintf("F%d", row), fa.Description)
		f.SetCellValue(facultySheet, fmt.Sprintf("G%d", row), fa.AvatarInfo.Path)
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

	// ===== Sheet: Schools =====
	schoolSheet := "Schools"
	f.NewSheet(schoolSheet)

	schoolHeaders := []string{"ID", "Name"}
	for i, h := range schoolHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(schoolSheet, cell, h)
		f.SetCellStyle(schoolSheet, cell, cell, headerStyle)
	}
	for idx, school := range schools {
		row := idx + 2
		f.SetCellValue(schoolSheet, fmt.Sprintf("A%d", row), school.ID)
		f.SetCellValue(schoolSheet, fmt.Sprintf("B%d", row), school.Name)
	}

	for idx := range schools {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(schoolHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(schoolSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(schoolSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("faculties_%d.xlsx", time.Now().Unix())

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

	config.Log.Info("Export faculty Excel to S3:", fileURL)

	return fileURL, nil
}
