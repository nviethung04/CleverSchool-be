package services

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *questionService) Export(c *gin.Context) (string, error) {
	attibuteRepo := repositories.NewQuestionAttributeRepository()
	attibutes, err := attibuteRepo.GetParents()

	allowedFilters := []string{"status", "question_type", "subject_id"}
	filter, _, _, _, _, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	filter, _, _ = s.ApplyFilter(c, filter)

	s.repo.SetPreload([]string{
		"Answers", "AnswerPositions", "AnswerGroups",
		"AnswerCoordinates", "AnswerMatchings", "AnswerGroups.Group",
		"RefAttributes",
	})

	s.repo.SetFilter(filter)
	s.repo.SetSort(map[string]string{
		"id": "asc",
	})
	questions, err := s.repo.GetAll()
	if err != nil {
		return "", err
	}

	var fillInBlanksQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeFillInBlanks {
			fillInBlanksQuestions = append(fillInBlanksQuestions, q)
		}
	}

	var orderingQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeOrdering {
			orderingQuestions = append(orderingQuestions, q)
		}
	}

	var drapDropQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeDragDrop {
			drapDropQuestions = append(drapDropQuestions, q)
		}
	}

	var multipleChoiceQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeMultipleChoice {
			multipleChoiceQuestions = append(multipleChoiceQuestions, q)
		}
	}

	var matchingQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeMatching {
			matchingQuestions = append(matchingQuestions, q)
		}
	}

	var labelingQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeLabeling {
			labelingQuestions = append(labelingQuestions, q)
		}
	}

	var categoryQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeCategory {
			categoryQuestions = append(categoryQuestions, q)
		}
	}

	var writingQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeWriting {
			writingQuestions = append(writingQuestions, q)
		}
	}

	var speakingQuestions []models.Question
	for _, q := range questions {
		if q.QuestionType == models.QuestionTypeSpeaking {
			speakingQuestions = append(speakingQuestions, q)
		}
	}

	f := excelize.NewFile()

	qfibSheet := "Question Fill in blanks"
	f.SetSheetName("Sheet1", qfibSheet)
	err = s.SetSheetAnswerPosition(f, qfibSheet, fillInBlanksQuestions, models.QuestionTypeFillInBlanks, attibutes)
	if err != nil {
		return "", err
	}

	qoSheet := "Question Ordering"
	f.NewSheet(qoSheet)
	err = s.SetSheetAnswerPosition(f, qoSheet, orderingQuestions, models.QuestionTypeOrdering, attibutes)
	if err != nil {
		return "", err
	}

	qddSheet := "Question Drag drop"
	f.NewSheet(qddSheet)
	err = s.SetSheetAnswerPosition(f, qddSheet, drapDropQuestions, models.QuestionTypeDragDrop, attibutes)
	if err != nil {
		return "", err
	}

	qmcSheet := "Question Multiple choice"
	f.NewSheet(qmcSheet)
	err = s.SetSheetAnswer(f, qmcSheet, multipleChoiceQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qmSheet := "Question Matching"
	f.NewSheet(qmSheet)
	err = s.SetSheetAnswerMatching(f, qmSheet, matchingQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qlSheet := "Question Labeling"
	f.NewSheet(qlSheet)
	err = s.SetSheetAnswerLabeling(f, qlSheet, labelingQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qcSheet := "Question Category"
	f.NewSheet(qcSheet)
	err = s.SetSheetAnswerCategory(f, qcSheet, categoryQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qwSheet := "Question Writing"
	f.NewSheet(qwSheet)
	err = s.SetSheetQuestion(f, qwSheet, writingQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qsSheet := "Question Speaking"
	f.NewSheet(qsSheet)
	err = s.SetSheetQuestion(f, qsSheet, speakingQuestions, attibutes)
	if err != nil {
		return "", err
	}

	qaSheet := "Question Attibute"
	f.NewSheet(qaSheet)
	err = s.SetSheetQuestionAttribute(f, qaSheet, attibutes)
	if err != nil {
		return "", err
	}

	// Áp dụng style cho tất cả các sheet
	s.applyStylesToAllSheets(f)

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("questions_%d.xlsx", time.Now().Unix())

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

	config.Log.Info("Export question Excel to S3:", fileURL)

	return fileURL, nil
}

func (s *questionService) SetSheetQuestionAttribute(f *excelize.File, sheetName string, attibutes []models.QuestionAttribute) error {
	headers := []string{
		"ID", "Parent ID", "Name",
		"Level", "Weight", "Description",
	}

	// Set header row
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, a := range attibutes {
		col := 1
		_ = s.SetCell(f, sheetName, col, row, a.ID)
		col++
		_ = s.SetCell(f, sheetName, col, row, nil)
		col++
		_ = s.SetCell(f, sheetName, col, row, a.Name)
		col++
		_ = s.SetCell(f, sheetName, col, row, a.Level)
		col++
		_ = s.SetCell(f, sheetName, col, row, a.Weight)
		col++
		_ = s.SetCell(f, sheetName, col, row, a.Description)

		row++

		for _, n := range a.Nodes {
			col := 1
			_ = s.SetCell(f, sheetName, col, row, n.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, a.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, n.Name)
			col++
			_ = s.SetCell(f, sheetName, col, row, n.Level)
			col++
			_ = s.SetCell(f, sheetName, col, row, n.Weight)
			col++
			_ = s.SetCell(f, sheetName, col, row, n.Description)

			row++
		}
	}

	return nil
}

func (s *questionService) SetSheetQuestion(f *excelize.File, sheetName string, questions []models.Question, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		col := 1
		_ = s.SetCell(f, sheetName, col, row, q.ID)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.Title)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.Description)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.Content)
		col++
		s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
		_ = s.SetCell(f, sheetName, col, row, q.Point)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
		col++
		_ = s.SetCell(f, sheetName, col, row, q.IsRandom)

		// Set attribute
		for _, attr := range attibutes {
			value := s.GetAttributeValue(q, attr)
			_ = s.SetCell(f, sheetName, col, row, value)
			col++
		}

		row++

		s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
	}

	return nil
}

func (s *questionService) SetSheetAnswerPosition(f *excelize.File, sheetName string, questions []models.Question, questionType string, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds", "Is random",
		"Answer - ID", "Answer - Content", "Answer - File url",
		"Answer - Kind", "Answer - Point", "Answer - Correct position", "Answer - Group position",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		for i, ans := range q.AnswerPositions {
			col := 1
			if i == 0 {
				_ = s.SetCell(f, sheetName, col, row, q.ID)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Title)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Description)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Content)
				col++
				s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
				_ = s.SetCell(f, sheetName, col, row, q.Point)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.IsRandom)
			} else {
				col += 9
			}

			// Answer
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Point)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.CorrectPosition)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.GroupPosition)

			if i == 0 {
				// Set attribute
				for _, attr := range attibutes {
					col++
					value := s.GetAttributeValue(q, attr)
					_ = s.SetCell(f, sheetName, col, row, value)
				}
			}

			row++

			if i == 0 {
				s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
			}
		}
	}

	return nil
}

func (s *questionService) SetSheetAnswerMatching(f *excelize.File, sheetName string, questions []models.Question, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds", "Is random",
		"Answer - ID", "Answer - Content", "Answer - File url",
		"Answer - Kind", "Answer - Maching content", "Answer - Maching file url",
		"Answer - Maching kind", "Answer - Is correct", "Answer - Point",
		"Answer - Correct position",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		for i, ans := range q.AnswerMatchings {
			col := 1
			if i == 0 {
				_ = s.SetCell(f, sheetName, col, row, q.ID)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Title)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Description)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Content)
				col++
				s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
				_ = s.SetCell(f, sheetName, col, row, q.Point)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.IsRandom)
			} else {
				col += 9
			}

			// Answer
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.MatchingContent)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.MatchingFileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.MatchingKind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.IsCorrect)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Point)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.CorrectPosition)

			if i == 0 {
				// Set attribute
				for _, attr := range attibutes {
					col++
					value := s.GetAttributeValue(q, attr)
					_ = s.SetCell(f, sheetName, col, row, value)
				}
			}

			row++

			if i == 0 {
				s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
			}
		}
	}

	return nil
}

func (s *questionService) SetSheetAnswer(f *excelize.File, sheetName string, questions []models.Question, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds", "Is random",
		"Answer - ID", "Answer - Content", "Answer - File url",
		"Answer - Kind", "Answer - Is correct", "Answer - Point",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		for i, ans := range q.Answers {
			col := 1
			if i == 0 {
				_ = s.SetCell(f, sheetName, col, row, q.ID)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Title)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Description)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Content)
				col++
				s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
				_ = s.SetCell(f, sheetName, col, row, q.Point)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.IsRandom)
			} else {
				col += 9
			}

			// Answer
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.IsCorrect)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Point)

			if i == 0 {
				// Set attribute
				for _, attr := range attibutes {
					col++
					value := s.GetAttributeValue(q, attr)
					_ = s.SetCell(f, sheetName, col, row, value)
				}
			}

			row++

			if i == 0 {
				s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
			}
		}
	}

	return nil
}

func (s *questionService) SetSheetAnswerLabeling(f *excelize.File, sheetName string, questions []models.Question, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds", "Is random",
		"Answer - ID", "Answer - Content", "Answer - File url",
		"Answer - Kind", "Answer - Position x", "Answer - Position y",
		"Answer - Position width", "Answer - Position height", "Answer - Point", "Answer - Group position",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		for i, ans := range q.AnswerCoordinates {
			col := 1
			if i == 0 {
				_ = s.SetCell(f, sheetName, col, row, q.ID)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Title)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Description)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Content)
				col++
				s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
				_ = s.SetCell(f, sheetName, col, row, q.Point)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.IsRandom)
			} else {
				col += 9
			}

			// Answer
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.PositionX)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.PositionY)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.PositionWidth)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.PositionHeight)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Point)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.GroupPosition)

			if i == 0 {
				// Set attribute
				for _, attr := range attibutes {
					col++
					value := s.GetAttributeValue(q, attr)
					_ = s.SetCell(f, sheetName, col, row, value)
				}
			}

			row++

			if i == 0 {
				s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
			}
		}
	}

	return nil
}

func (s *questionService) SetSheetAnswerCategory(f *excelize.File, sheetName string, questions []models.Question, attibutes []models.QuestionAttribute) error {
	answerHeaders := []string{
		"Question ID", "Source ID", "Title", "Description",
		"Content", "Kind", "File url",
		"Point", "Time limit seconds", "Is random",
		"Answer - ID", "Answer - Content", "Answer - File url",
		"Answer - Kind", "Answer - Group ID", "Answer - Group content",
		"Answer - Group file url", "Answer - Group kind", "Answer - Point",
	}

	for _, a := range attibutes {
		answerHeaders = append(answerHeaders, fmt.Sprintf("Attribute - %v - %s", a.ID, a.Name))
	}

	// Set header row
	for i, h := range answerHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	// Fill data
	row := 2
	for _, q := range questions {
		fileInfosToUse := s.getFileInfo(q)

		for i, ans := range q.AnswerGroups {
			col := 1
			if i == 0 {
				_ = s.SetCell(f, sheetName, col, row, q.ID)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.SourceQuestionId)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Title)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Description)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.Content)
				col++
				s.setKindAndPath(f, sheetName, &col, row, fileInfosToUse, q)
				_ = s.SetCell(f, sheetName, col, row, q.Point)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.TimeLimitSeconds)
				col++
				_ = s.SetCell(f, sheetName, col, row, q.IsRandom)
			} else {
				col += 9
			}

			// Answer
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Group.ID)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Group.Content)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Group.FileInfo.Path)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Group.Kind)
			col++
			_ = s.SetCell(f, sheetName, col, row, ans.Point)

			if i == 0 {
				// Set attribute
				for _, attr := range attibutes {
					col++
					value := s.GetAttributeValue(q, attr)
					_ = s.SetCell(f, sheetName, col, row, value)
				}
			}

			row++

			if i == 0 {
				s.appendKindPathOnlyRows(f, sheetName, &row, fileInfosToUse)
			}
		}
	}

	return nil
}

func (s *questionService) GetAttributeValue(question models.Question, attr models.QuestionAttribute) string {
	for _, ref := range question.RefAttributes {
		if ref.ParentAttributeID != nil && *ref.ParentAttributeID == attr.ID {
			return strconv.FormatInt(ref.AttributeID, 10)
		}
	}
	return ""
}

func (s *questionService) SetCell(f *excelize.File, sheet string, col, row int, value interface{}) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return err
	}
	return f.SetCellValue(sheet, cell, value)
}

// Helper functions để lấy style
func (s *questionService) getHeaderStyle(f *excelize.File) int {
	style, err := f.NewStyle(&excelize.Style{
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
		return 0
	}
	return style
}

func (s *questionService) getOddRowStyle(f *excelize.File) int {
	style, err := f.NewStyle(&excelize.Style{
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
		return 0
	}
	return style
}

func (s *questionService) getEvenRowStyle(f *excelize.File) int {
	style, err := f.NewStyle(&excelize.Style{
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
		return 0
	}
	return style
}

// Helper function để áp dụng style cho tất cả các sheet
func (s *questionService) applyStylesToAllSheets(f *excelize.File) {
	// Áp dụng style cho tất cả các sheet
	sheets := []string{
		"Question Fill in blanks",
		"Question Ordering",
		"Question Drag drop",
		"Question Multiple choice",
		"Question Matching",
		"Question Labeling",
		"Question Category",
		"Question Writing",
		"Question Speaking",
		"Question Attibute",
	}

	for _, sheetName := range sheets {
		// Lấy tất cả các cell có data và áp dụng style
		rows, err := f.GetRows(sheetName)
		if err != nil || len(rows) <= 1 {
			continue
		}

		// Áp dụng style cho header (row 1)
		for c := 1; c <= len(rows[0]); c++ {
			cell, _ := excelize.CoordinatesToCellName(c, 1)
			f.SetCellStyle(sheetName, cell, cell, s.getHeaderStyle(f))
		}

		// Áp dụng style cho data rows
		for r := 2; r <= len(rows); r++ {
			for c := 1; c <= len(rows[0]); c++ {
				cell, _ := excelize.CoordinatesToCellName(c, r)
				if (r-2)%2 == 0 {
					f.SetCellStyle(sheetName, cell, cell, s.getOddRowStyle(f))
				} else {
					f.SetCellStyle(sheetName, cell, cell, s.getEvenRowStyle(f))
				}
			}
		}
	}
}

func (s *questionService) getFileInfo(q models.Question) models.MediaInfos {
	fileInfosToUse := q.FileInfos
	if len(fileInfosToUse) == 0 {
		if q.FileInfo.Path != "" {
			fileInfosToUse = models.MediaInfos{models.MediaDetail{
				Id:   q.FileInfo.Id,
				Disk: q.FileInfo.Disk,
				Path: q.FileInfo.Path,
				Type: q.Kind,
			}}
		}
	} else {
		for i, fileInfo := range fileInfosToUse {
			if fileInfo.Type == "" {
				fileInfosToUse[i].Type = q.Kind
			}
		}
	}
	return fileInfosToUse
}

// setKindAndPath sets Kind and Path columns for a row
func (s *questionService) setKindAndPath(f *excelize.File, sheetName string, col *int, row int, fileInfosToUse models.MediaInfos, q models.Question) {
	if len(fileInfosToUse) > 0 {
		_ = s.SetCell(f, sheetName, *col, row, fileInfosToUse[0].Type)
		*col = *col + 1
		_ = s.SetCell(f, sheetName, *col, row, fileInfosToUse[0].Path)
	} else {
		_ = s.SetCell(f, sheetName, *col, row, q.Kind)
		*col = *col + 1
		_ = s.SetCell(f, sheetName, *col, row, q.FileInfo.Path)
	}
	*col = *col + 1
}

// appendKindPathOnlyRows appends extra rows for file infos (starting at index 1),
// filling ONLY the Kind and File url columns. It increments *row for each appended row.
//
// The sheet layout used in exports is:
// 1: Question ID, 2: Source ID, 3: Title, 4: Description, 5: Content, 6: Kind, 7: File url, ...
func (s *questionService) appendKindPathOnlyRows(f *excelize.File, sheetName string, row *int, fileInfosToUse models.MediaInfos) {
	if len(fileInfosToUse) <= 1 {
		return
	}

	for fileIdx := 1; fileIdx < len(fileInfosToUse); fileIdx++ {
		fileInfo := fileInfosToUse[fileIdx]
		col := 6 // Kind column
		_ = s.SetCell(f, sheetName, col, *row, fileInfo.Type)
		col++ // File url column
		_ = s.SetCell(f, sheetName, col, *row, fileInfo.Path)
		*row = *row + 1
	}
}

