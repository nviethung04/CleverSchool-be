package services

import (
	"be-cleverschool/config"
	"be-cleverschool/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *schoolService) Export(c *gin.Context) (string, error) {
	allowedFilters := []string{}

	filter, _, _, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	s.repo.SetSearch(keyword, []string{})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)

	schools, err := s.repo.GetAll()
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

	// ===== Sheet: Schools =====
	schoolSheet := "Schools"
	f.SetSheetName("Sheet1", schoolSheet)

	schoolHeaders := []string{
		"ID", "Name", "Short Name",
		"Type", "Address VN", "Address EN",
		"Contact Name", "Contact Phone", "ward Code",
		"Logo",
	}
	for i, h := range schoolHeaders {
		col := string('A' + i)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(schoolSheet, cell, h)
		f.SetCellStyle(schoolSheet, cell, cell, headerStyle)
	}

	for idx, s := range schools {
		row := idx + 2
		f.SetCellValue(schoolSheet, fmt.Sprintf("A%d", row), s.ID)
		f.SetCellValue(schoolSheet, fmt.Sprintf("B%d", row), s.Name)
		f.SetCellValue(schoolSheet, fmt.Sprintf("C%d", row), s.ShortName)
		f.SetCellValue(schoolSheet, fmt.Sprintf("D%d", row), s.Type)
		f.SetCellValue(schoolSheet, fmt.Sprintf("E%d", row), s.AddressVN)
		f.SetCellValue(schoolSheet, fmt.Sprintf("F%d", row), s.AddressEN)
		f.SetCellValue(schoolSheet, fmt.Sprintf("G%d", row), s.ContactName)
		f.SetCellValue(schoolSheet, fmt.Sprintf("H%d", row), s.ContactPhone)
		f.SetCellValue(schoolSheet, fmt.Sprintf("I%d", row), s.WardCode)
		f.SetCellValue(schoolSheet, fmt.Sprintf("J%d", row), s.LogoInfo.Path)
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
	filename := fmt.Sprintf("schools_%d.xlsx", time.Now().Unix())

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

	config.Log.Info("Export school Excel to S3:", fileURL)

	return fileURL, nil
}

