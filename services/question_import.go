package services

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"be-cleverschool/utils"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *questionService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	attibuteRepo := repositories.NewQuestionAttributeRepository()
	attibutes, err := attibuteRepo.GetParents()
	if err != nil {
		return err
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

	fillInBlanksRows, err := f.GetRows("Question Fill in blanks")
	config.Log.Info("Import question, Question Fill in blanks...")
	if err == nil {
		s.UpdateOrCreateAnswerPositionQuestions(fillInBlanksRows, models.QuestionTypeFillInBlanks, attibutes)
	}

	orderingRows, err := f.GetRows("Question Ordering")
	config.Log.Info("Import question, Question Ordering...")
	if err == nil {
		s.UpdateOrCreateAnswerPositionQuestions(orderingRows, models.QuestionTypeOrdering, attibutes)
	}

	dragDropRows, err := f.GetRows("Question Drag drop")
	config.Log.Info("Import question, Question Drag drop...")
	if err == nil {
		s.UpdateOrCreateAnswerPositionQuestions(dragDropRows, models.QuestionTypeDragDrop, attibutes)
	}

	multipleChoiceRows, err := f.GetRows("Question Multiple choice")
	config.Log.Info("Import question, Question Multiple choice...")
	if err == nil {
		s.UpdateOrCreateAnswerQuestions(multipleChoiceRows, attibutes)
	}

	matchingRows, err := f.GetRows("Question Matching")
	config.Log.Info("Import question, Question Matching...")
	if err == nil {
		s.UpdateOrCreateAnswerMatchingQuestions(matchingRows, attibutes)
	}

	labelingRows, err := f.GetRows("Question Labeling")
	config.Log.Info("Import question, Question Labeling...")
	if err == nil {
		s.UpdateOrCreateAnswerLabelingQuestions(labelingRows, attibutes)
	}

	categoryRows, err := f.GetRows("Question Category")
	config.Log.Info("Import question, Question Category...")
	if err == nil {
		s.UpdateOrCreateAnswerCategoryQuestions(categoryRows, attibutes)
	}

	writingRows, err := f.GetRows("Question Writing")
	config.Log.Info("Import question, Question Writing...")
	if err == nil {
		s.UpdateOrCreateWritingAndSpeakingQuestions(writingRows, models.QuestionTypeWriting, attibutes)
	}

	speakingRows, err := f.GetRows("Question Speaking")
	config.Log.Info("Import question, Question Speaking...")
	if err == nil {
		s.UpdateOrCreateWritingAndSpeakingQuestions(speakingRows, models.QuestionTypeSpeaking, attibutes)
	}

	s.SyncKeywords(c, 0)

	return nil
}

func (s *questionService) UpdateOrCreateAnswerPositionQuestions(rows [][]string, questionType string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	var (
		currentQuestionId int64
		currentAnswers    []models.AnswerPosition
		currentFileInfos  models.MediaInfos
	)

	importCount := 0

	for i, row := range rows {
		if i == 0 || len(row) < 7 {
			config.Log.Warn("Import question, UpdateOrCreateAnswerPositionQuestions:")
			config.Log.Warn(row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip answer parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 10) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			// persist immediately so we don't lose data if the file has no more question rows
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		// New question
		if row[2] != "" {
			// Save first answer
			if currentQuestionId != 0 && len(currentAnswers) > 0 {
				ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerPosition](currentAnswers)
				repositories.DeleteOldAnswers[models.AnswerPosition](currentQuestionId, ids)
			}

			// Save file infos for previous question
			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			qID, _ := strconv.ParseInt(row[0], 10, 64)
			sourceID, _ := strconv.ParseInt(row[1], 10, 64)
			point, _ := strconv.ParseFloat(row[7], 64)
			timeLimit, _ := strconv.Atoi(row[8])
			isRandom, _ := strconv.Atoi(row[9])

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     questionType,
				TimeLimitSeconds: timeLimit,
				IsRandom:         int16(isRandom),
			}

			id, err := s.repo.UpdateOrCreate(question)
			if err != nil {
				config.Log.Warn(err)
				currentQuestionId = 0
				currentAnswers = nil
				continue
			}

			s.SaveAttributes(id, row, attibutes, 16)

			currentQuestionId = id
			currentAnswers = nil
			// ensure file infos are initialized for this question id (in case UpdateOrCreate skips zero-values)
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			importCount++
		}

		// append answer
		if currentQuestionId != 0 {
			if len(row) < 12 {
				config.Log.Warn("Import question, currentQuestionId:")
				config.Log.Warn(row)
				continue
			}

			var answerID int64
			if len(row) > 10 && row[10] != "" {
				answerID, _ = strconv.ParseInt(row[10], 10, 64)
			}

			var fileUrl string
			if len(row) > 12 && row[12] != "" {
				fileUrl = row[12]
			}

			var kind string
			if len(row) > 13 && row[13] != "" {
				kind = row[13]
			} else {
				kind = "text"
			}

			fileUrl = utils.StripDomain(fileUrl, models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			answer := models.AnswerPosition{
				ID:         answerID,
				QuestionID: &currentQuestionId,
				Content:    row[11],
				FileInfo:   fileInfo,
				Kind:       kind,
			}

			if len(row) > 15 && row[15] != "" {
				var groupPosition string
				if len(row) > 16 && row[16] != "" {
					groupPosition = row[16]
				} else if len(row) > 15 {
					groupPosition = row[15]
				}

				answer.Point, _ = strconv.ParseFloat(row[14], 64)
				answer.CorrectPosition, _ = strconv.Atoi(row[15])
				answer.GroupPosition, _ = strconv.Atoi(groupPosition)
			}

			currentAnswers = append(currentAnswers, answer)
		}
	}

	// save last current
	if currentQuestionId != 0 && len(currentAnswers) > 0 {
		ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerPosition](currentAnswers)
		repositories.DeleteOldAnswers[models.AnswerPosition](currentQuestionId, ids)
	}
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateAnswerPositionQuestions done")

	return nil
}

// isMediaOnlyRow detects the exported "extra media rows" which only contain Kind + File url.
// We also require answerIdCol to be empty to avoid misclassifying normal answer rows.
func isMediaOnlyRow(row []string, kindCol, urlCol, answerIdCol int) bool {
	if len(row) <= urlCol {
		return false
	}
	kind := ""
	if len(row) > kindCol {
		kind = row[kindCol]
	}
	url := row[urlCol]
	answerID := ""
	if len(row) > answerIdCol {
		answerID = row[answerIdCol]
	}
	return (kind != "" || url != "") && answerID == ""
}

func (s *questionService) parseMediaDetail(row []string, kindCol, urlCol int) models.MediaDetail {
	kind := "text"
	if len(row) > kindCol && row[kindCol] != "" {
		kind = row[kindCol]
	}
	fileUrl := ""
	if len(row) > urlCol && row[urlCol] != "" {
		fileUrl = utils.StripDomain(row[urlCol], models.Storage)
	}
	mediaRepo := repositories.NewMediaRepository()
	fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
	return models.MediaDetail{
		Id:   fileInfo.Id,
		Disk: fileInfo.Disk,
		Path: fileInfo.Path,
		Type: kind,
	}
}

func (s *questionService) UpdateOrCreateAnswerQuestions(rows [][]string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	var (
		currentQuestionId int64
		currentAnswers    []models.Answer
		currentFileInfos  models.MediaInfos
	)

	importCount := 0

	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			config.Log.Warn("Import question, UpdateOrCreateAnswerQuestions:")
			config.Log.Warn(row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip answer parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 10) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		// New question
		if row[2] != "" {
			// Save first answer
			if currentQuestionId != 0 && len(currentAnswers) > 0 {
				ids, _ := repositories.UpdateOrCreateAnswers[models.Answer](currentAnswers)
				repositories.DeleteOldAnswers[models.Answer](currentQuestionId, ids)
			}

			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			qID, _ := strconv.ParseInt(row[0], 10, 64)
			sourceID, _ := strconv.ParseInt(row[1], 10, 64)
			point, _ := strconv.ParseFloat(row[7], 64)
			timeLimit, _ := strconv.Atoi(row[8])
			isRandom, _ := strconv.Atoi(row[9])

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     models.QuestionTypeMultipleChoice,
				TimeLimitSeconds: timeLimit,
				IsRandom:         int16(isRandom),
			}

			id, err := s.repo.UpdateOrCreate(question)
			if err != nil {
				config.Log.Warn(err)
				currentQuestionId = 0
				currentAnswers = nil
				continue
			}

			s.SaveAttributes(id, row, attibutes, 15)

			currentQuestionId = id
			currentAnswers = nil
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			importCount++
		}

		// append answer
		if currentQuestionId != 0 {
			answerID, _ := strconv.ParseInt(row[10], 10, 64)

			fileUrl := utils.StripDomain(row[12], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kindAnswer string
			if len(row) > 13 && row[13] != "" {
				kindAnswer = row[13]
			} else {
				kindAnswer = "text"
			}

			answer := models.Answer{
				ID:         answerID,
				QuestionID: &currentQuestionId,
				Content:    row[11],
				FileInfo:   fileInfo,
				Kind:       kindAnswer,
			}

			var isCorrectStr string

			if len(row) > 14 && row[14] != "" {
				isCorrectStr = row[14]
			} else {
				isCorrectStr = "false"
			}

			isCorrect := isCorrectStr == "true" || isCorrectStr == "1" || isCorrectStr == "TRUE"
			answer.IsCorrect = isCorrect

			var point float64
			if len(row) > 15 && row[15] != "" {
				point, _ = strconv.ParseFloat(row[15], 64)
			} else {
				point = 0
			}
			answer.Point = point

			currentAnswers = append(currentAnswers, answer)
		}
	}

	// save last current
	if currentQuestionId != 0 && len(currentAnswers) > 0 {
		ids, _ := repositories.UpdateOrCreateAnswers[models.Answer](currentAnswers)
		repositories.DeleteOldAnswers[models.Answer](currentQuestionId, ids)
	}
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateAnswerQuestions done")

	return nil
}

func (s *questionService) UpdateOrCreateAnswerMatchingQuestions(rows [][]string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	var (
		currentQuestionId int64
		currentAnswers    []models.AnswerMatching
		currentFileInfos  models.MediaInfos
	)

	importCount := 0

	for i, row := range rows {
		if i == 0 || len(row) < 19 {
			config.Log.Warnf("Import question, UpdateOrCreateAnswerMatchingQuestions: Row %d has insufficient columns. Expected: 19, Got: %d", i, len(row))
			config.Log.Warnf("Row data: %v", row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip answer parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 10) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		// New question
		if row[2] != "" {
			// Save first answer
			if currentQuestionId != 0 && len(currentAnswers) > 0 {
				ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerMatching](currentAnswers)
				repositories.DeleteOldAnswers[models.AnswerMatching](currentQuestionId, ids)
			}

			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			qID, _ := strconv.ParseInt(row[0], 10, 64)
			sourceID, _ := strconv.ParseInt(row[1], 10, 64)
			point, _ := strconv.ParseFloat(row[7], 64)
			timeLimit, _ := strconv.Atoi(row[8])
			isRandom, _ := strconv.Atoi(row[9])

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     models.QuestionTypeMatching,
				TimeLimitSeconds: timeLimit,
				IsRandom:         int16(isRandom),
			}

			id, err := s.repo.UpdateOrCreate(question)
			if err != nil {
				config.Log.Warn(err)
				currentQuestionId = 0
				currentAnswers = nil
				continue
			}

			s.SaveAttributes(id, row, attibutes, 19)

			currentQuestionId = id
			currentAnswers = nil
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}
			importCount++
		}

		// append answer
		if currentQuestionId != 0 {
			// Validate row has enough columns for answer data

			answerID, _ := strconv.ParseInt(row[10], 10, 64)

			fileUrl := utils.StripDomain(row[12], models.Storage)
			matchingFileUrl := utils.StripDomain(row[15], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
			matchingFileInfo := mediaRepo.GetMediaInfo(matchingFileUrl, models.Storage)

			var kindAnswer string
			if len(row) > 13 && row[13] != "" {
				kindAnswer = row[13]
			} else {
				kindAnswer = "text"
			}

			var kindAnswerMatching string
			if len(row) > 16 && row[16] != "" {
				kindAnswerMatching = row[16]
			} else {
				kindAnswerMatching = "text"
			}

			answer := models.AnswerMatching{
				ID:               answerID,
				QuestionID:       &currentQuestionId,
				Content:          row[11],
				FileInfo:         fileInfo,
				Kind:             kindAnswer,
				MatchingContent:  row[14],
				MatchingFileInfo: matchingFileInfo,
				MatchingKind:     kindAnswerMatching,
			}

			// Safe access to row elements with bounds checking
			if len(row) > 17 {
				isCorrect := row[17] == "true" || row[17] == "1" || row[17] == "TRUE"
				answer.IsCorrect = isCorrect
			}

			if len(row) > 18 {
				answer.Point, _ = strconv.ParseFloat(row[18], 64)
			}

			if len(row) > 19 {
				answer.CorrectPosition, _ = strconv.Atoi(row[19])
			}

			// var groupPositionStr string
			// if len(row) > 20 && row[20] != "" {
			// 	groupPositionStr = row[20]
			// } else if len(row) > 19 {
			// 	groupPositionStr = row[19]
			// } else {
			// 	groupPositionStr = "0"
			// }

			// var matchingGroupPositionStr string
			// if len(row) > 21 && row[21] != "" {
			// 	matchingGroupPositionStr = row[21]
			// } else {
			// 	matchingGroupPositionStr = groupPositionStr
			// }

			// answer.GroupPosition, _ = strconv.Atoi(groupPositionStr)
			// answer.MatchingGroupPosition, _ = strconv.Atoi(matchingGroupPositionStr)

			currentAnswers = append(currentAnswers, answer)
		}
	}

	// save last current
	if currentQuestionId != 0 && len(currentAnswers) > 0 {
		ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerMatching](currentAnswers)
		repositories.DeleteOldAnswers[models.AnswerMatching](currentQuestionId, ids)
	}
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateAnswerMatchingQuestions done")

	return nil
}

func (s *questionService) UpdateOrCreateAnswerLabelingQuestions(rows [][]string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	var (
		currentQuestionId int64
		currentAnswers    []models.AnswerCoordinates
		currentFileInfos  models.MediaInfos
	)

	importCount := 0

	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			config.Log.Warn("Import question, UpdateOrCreateAnswerLabelingQuestions:")
			config.Log.Warn(row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip answer parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 10) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		// New question
		if row[2] != "" {
			// Save first answer
			if currentQuestionId != 0 && len(currentAnswers) > 0 {
				ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerCoordinates](currentAnswers)
				repositories.DeleteOldAnswers[models.AnswerCoordinates](currentQuestionId, ids)
			}

			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			qID, _ := strconv.ParseInt(row[0], 10, 64)
			sourceID, _ := strconv.ParseInt(row[1], 10, 64)
			point, _ := strconv.ParseFloat(row[7], 64)
			timeLimit, _ := strconv.Atoi(row[8])
			isRandom, _ := strconv.Atoi(row[9])

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     models.QuestionTypeLabeling,
				TimeLimitSeconds: timeLimit,
				IsRandom:         int16(isRandom),
			}

			id, err := s.repo.UpdateOrCreate(question)
			if err != nil {
				config.Log.Warn(err)
				currentQuestionId = 0
				currentAnswers = nil
				continue
			}

			s.SaveAttributes(id, row, attibutes, 19)

			currentQuestionId = id
			currentAnswers = nil
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}
			importCount++
		}

		// append answer
		if currentQuestionId != 0 {
			answerID, _ := strconv.ParseInt(row[10], 10, 64)

			fileUrl := utils.StripDomain(row[12], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kindAnswer string
			if len(row) > 13 && row[13] != "" {
				kindAnswer = row[13]
			} else {
				kindAnswer = "text"
			}

			answer := models.AnswerCoordinates{
				ID:         answerID,
				QuestionID: &currentQuestionId,
				Content:    row[11],
				FileInfo:   fileInfo,
				Kind:       kindAnswer,
			}

			answer.PositionX, _ = strconv.Atoi(row[14])
			answer.PositionY, _ = strconv.Atoi(row[15])
			answer.PositionWidth, _ = strconv.Atoi(row[16])
			answer.PositionHeight, _ = strconv.Atoi(row[17])
			answer.Point, _ = strconv.ParseFloat(row[18], 64)
			answer.GroupPosition, _ = strconv.Atoi(row[19])

			currentAnswers = append(currentAnswers, answer)
		}
	}

	// save last current
	if currentQuestionId != 0 && len(currentAnswers) > 0 {
		ids, _ := repositories.UpdateOrCreateAnswers[models.AnswerCoordinates](currentAnswers)
		repositories.DeleteOldAnswers[models.AnswerCoordinates](currentQuestionId, ids)
	}
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateAnswerLabelingQuestions done")

	return nil
}

func (s *questionService) UpdateOrCreateAnswerCategoryQuestions(rows [][]string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	var (
		currentQuestionId int64
		currentAnswers    []models.AnswerGroup
		mappingGroup      = make(map[string]int64)
		currentFileInfos  models.MediaInfos
	)

	importCount := 0

	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			config.Log.Warn("Import question, UpdateOrCreateAnswerCategoryQuestions:")
			config.Log.Warn(row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 10) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		if row[2] != "" {
			if currentQuestionId != 0 && len(currentAnswers) > 0 {
				ids, err := repositories.UpdateOrCreateAnswers[models.AnswerGroup](currentAnswers)
				if err == nil {
					repositories.DeleteOldAnswers[models.AnswerGroup](currentQuestionId, ids)
				} else {
					return err
				}
			}

			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			mappingGroup = make(map[string]int64)

			qID, err1 := strconv.ParseInt(row[0], 10, 64)
			if err1 != nil {
				qID = 0
			}

			sourceID, err2 := strconv.ParseInt(row[1], 10, 64)
			if err2 != nil {
				sourceID = 0
			}

			point, err3 := strconv.ParseFloat(row[7], 64)
			if err3 != nil {
				point = 0
			}

			timeLimit, err4 := strconv.Atoi(row[8])
			if err4 != nil {
				timeLimit = 0
			}

			isRandom, err5 := strconv.Atoi(row[9])
			if err5 != nil {
				isRandom = 0
			}

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     models.QuestionTypeCategory,
				TimeLimitSeconds: timeLimit,
				IsRandom:         int16(isRandom),
			}

			id, err := s.repo.UpdateOrCreate(question)
			if err != nil {
				config.Log.Warn(err)
				currentQuestionId = 0
				currentAnswers = nil
				continue
			}

			s.SaveAttributes(id, row, attibutes, 18)

			currentQuestionId = id
			currentAnswers = nil
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}
			importCount++
		}

		groupID, err := strconv.ParseInt(row[14], 10, 64)
		if err != nil {
			groupID = 0
		}

		if groupID == 0 {
			if id, exists := mappingGroup[row[15]]; exists {
				groupID = id
			} else {

				fileUrl := utils.StripDomain(row[16], models.Storage)
				mediaRepo := repositories.NewMediaRepository()
				fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

				var kindAnswer string
				if len(row) > 17 && row[17] != "" {
					kindAnswer = row[17]
				} else {
					kindAnswer = "text"
				}

				group := models.GroupAnswer{
					Content:  row[15],
					FileInfo: fileInfo,
					Kind:     kindAnswer,
				}

				newID, err := repositories.UpdateOrCreateGroup(group)
				if err != nil {
					groupID = 0
				} else {
					groupID = newID
					mappingGroup[group.Content] = groupID
				}
			}
		}

		currentGroupId := groupID

		if currentQuestionId != 0 {
			answerID, _ := strconv.ParseInt(row[10], 10, 64)
			var pointStr string
			if len(row) > 18 && row[18] != "" {
				pointStr = row[18]
			} else {
				pointStr = "0"
			}

			point, _ := strconv.ParseFloat(pointStr, 64)

			fileUrl := utils.StripDomain(row[12], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kindAnswer string
			if len(row) > 17 && row[17] != "" {
				kindAnswer = row[17]
			} else {
				kindAnswer = "text"
			}

			answer := models.AnswerGroup{
				ID:         answerID,
				QuestionID: &currentQuestionId,
				GroupID:    &currentGroupId,
				Content:    row[11],
				FileInfo:   fileInfo,
				Kind:       kindAnswer,
				Point:      point,
			}

			currentAnswers = append(currentAnswers, answer)
		}
	}

	if currentQuestionId != 0 && len(currentAnswers) > 0 {
		ids, err := repositories.UpdateOrCreateAnswers[models.AnswerGroup](currentAnswers)
		if err == nil {
			repositories.DeleteOldAnswers[models.AnswerGroup](currentQuestionId, ids)
		} else {
			return err
		}
	}
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateAnswerCategoryQuestions done")

	return nil
}

func (s *questionService) UpdateOrCreateWritingAndSpeakingQuestions(rows [][]string, questionType string, attibutes []models.QuestionAttribute) error {
	if len(rows) <= 1 {
		return nil
	}

	importCount := 0
	var (
		currentQuestionId int64
		currentFileInfos  models.MediaInfos
	)

	for i, row := range rows {
		if i == 0 || len(row) < 8 {
			config.Log.Warn("Import question, UpdateOrCreateWritingAndSpeakingQuestions:")
			config.Log.Warn(row)
			continue
		}

		// Media-only row (Kind + File url only): append to currentFileInfos and skip parsing
		if row[2] == "" && isMediaOnlyRow(row, 5, 6, 0) && currentQuestionId != 0 {
			currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			continue
		}

		if row[2] != "" {
			// Save file infos for previous question
			if currentQuestionId != 0 && len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			qID, err1 := strconv.ParseInt(row[0], 10, 64)
			if err1 != nil {
				qID = 0
			}

			sourceID, err2 := strconv.ParseInt(row[1], 10, 64)
			if err2 != nil {
				sourceID = 0
			}

			point, err3 := strconv.ParseFloat(row[7], 64)
			if err3 != nil {
				point = 0
			}

			timeLimit, err4 := strconv.Atoi(row[8])
			if err4 != nil {
				timeLimit = 0
			}

			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			var kind string
			if len(row) > 5 && row[5] != "" {
				kind = row[5]
			} else {
				kind = "text"
			}

			// Init file infos for this question (include first media)
			currentFileInfos = models.MediaInfos{}
			if row[5] != "" || row[6] != "" {
				currentFileInfos = append(currentFileInfos, s.parseMediaDetail(row, 5, 6))
			}

			question := models.Question{
				ID:               qID,
				SourceQuestionId: sourceID,
				Title:            row[2],
				Description:      row[3],
				Content:          row[4],
				Kind:             kind,
				FileInfo:         fileInfo,
				FileInfos:        currentFileInfos,
				Point:            point,
				QuestionType:     questionType,
				TimeLimitSeconds: timeLimit,
			}

			id, err := s.repo.UpdateOrCreate(question)

			if err != nil {
				config.Log.Warn(err)
				continue
			}

			s.SaveAttributes(id, row, attibutes, 8)

			currentQuestionId = id
			if len(currentFileInfos) > 0 {
				_ = s.repo.UpdateFileInfos(currentQuestionId, currentFileInfos)
			}

			importCount++
		}
	}

	// save last current
	if currentQuestionId != 0 && len(currentFileInfos) > 0 {
		_ = s.repo.Update(&models.Question{ID: currentQuestionId, FileInfos: currentFileInfos})
	}

	config.Log.Info("Count import questions:", importCount)
	config.Log.Info("Import question, UpdateOrCreateWritingAndSpeakingQuestions done")

	return nil
}

