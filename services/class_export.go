package services

import (
	"be-lms/config"
	"be-lms/repositories"
	"be-lms/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *classService) Export(c *gin.Context) (string, error) {
	schoolRepo := repositories.NewSchoolRepository()
	gradeRepo := repositories.NewGradeRepository()

	allowedFilters := []string{"status", "school_id"}

	filter, _, _, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)

	schools, err := schoolRepo.GetAll()
	if err != nil {
		return "", err
	}

	grades, err := gradeRepo.GetAll()
	if err != nil {
		return "", err
	}

	classes, err := s.repo.GetAll()
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

	// ===== Sheet: Classes =====
	classSheet := "Classes"
	f.SetSheetName("Sheet1", classSheet)

	classHeaders := []string{
		"ID", "School Id", "Grade Id",
		"Name", "Max student", "Teacher - Name",
		"Teacher - Phone", "Teacher - Email", "Status",
	}
	for i, h := range classHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(classSheet, cell, h)
		f.SetCellStyle(classSheet, cell, cell, headerStyle)
	}
	for idx, c := range classes {
		row := idx + 2
		f.SetCellValue(classSheet, fmt.Sprintf("A%d", row), c.ID)
		f.SetCellValue(classSheet, fmt.Sprintf("B%d", row), c.SchoolId)
		f.SetCellValue(classSheet, fmt.Sprintf("C%d", row), c.GradeId)
		f.SetCellValue(classSheet, fmt.Sprintf("D%d", row), c.Name)
		f.SetCellValue(classSheet, fmt.Sprintf("E%d", row), c.MaxStudents)
		f.SetCellValue(classSheet, fmt.Sprintf("F%d", row), c.TeacherInfo.Name)
		f.SetCellValue(classSheet, fmt.Sprintf("G%d", row), c.TeacherInfo.Phone)
		f.SetCellValue(classSheet, fmt.Sprintf("H%d", row), c.TeacherInfo.Email)
		f.SetCellValue(classSheet, fmt.Sprintf("I%d", row), c.Status)
	}

	for idx := range classes {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(classHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(classSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(classSheet, cell, cell, evenRowStyle)
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

	// ===== Sheet: Grades =====
	gradeSheet := "Grades"
	f.NewSheet(gradeSheet)

	gradeHeaders := []string{"ID", "Name"}
	for i, h := range gradeHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(gradeSheet, cell, h)
		f.SetCellStyle(gradeSheet, cell, cell, headerStyle)
	}
	for idx, grade := range grades {
		row := idx + 2
		f.SetCellValue(gradeSheet, fmt.Sprintf("A%d", row), grade.ID)
		f.SetCellValue(gradeSheet, fmt.Sprintf("B%d", row), grade.NameVN)
	}

	for idx := range grades {
		row := idx + 2
		for col := 'A'; col <= rune('A'+len(gradeHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(gradeSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(gradeSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("classes_%d.xlsx", time.Now().Unix())

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

	config.Log.Info("Export class Excel to S3:", fileURL)

	return fileURL, nil
}
