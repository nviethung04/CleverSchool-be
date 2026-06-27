package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"strconv"
)

func (s *questionService) StoreAnswers(questionID int64, request *prot.Question, questionType string) error {
	switch questionType {
	case "multiple_choice":
		modelsAnswers := make([]models.Answer, 0, len(request.Options.Answers))
		answerIds := make([]int64, 0, len(request.Options.Answers))

		for index, ans := range request.Options.Answers {
			isCorrect := false

			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			for _, id := range request.CorrectAnswers.Ids {
				if id == indexGroup {
					isCorrect = true
					break
				}
			}

			fileUrl := utils.StripDomain(ans.Media.Url, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.Answer{
				ID:         int64(ans.Id),
				QuestionID: &questionID,
				Content:    ans.Text,
				FileInfo:   mediaInfo,
				Kind:       ans.Media.Type,
				Point:      float64(ans.Point),
				IsCorrect:  isCorrect || ans.IsCorrect,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.Answer](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "fill_in_blanks":
		modelsAnswers := make([]models.AnswerPosition, 0, len(request.Options.Answers))
		answerIds := make([]int64, 0, len(request.Options.Answers))

		for index, ans := range request.Options.Answers {
			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			indexGroupKey := strconv.FormatInt(indexGroup, 10)

			var correctPosition int64

			correctPositionStr, ok := request.CorrectAnswers.List[indexGroupKey]

			if !ok {
				correctPosition = int64(ans.CorrectPosition)
			} else {
				var err error
				correctPosition, err = strconv.ParseInt(correctPositionStr, 10, 64)
				if err != nil {
					correctPosition = int64(ans.CorrectPosition)
				}
			}

			groupPosition := int(ans.GroupPosition)

			if groupPosition == 0 {
				groupPosition = int(correctPosition)
			}

			fileUrl := utils.StripDomain(ans.Media.Url, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.AnswerPosition{
				ID:              int64(ans.Id),
				QuestionID:      &questionID,
				Content:         ans.Text,
				FileInfo:        mediaInfo,
				Kind:            ans.Media.Type,
				Point:           float64(ans.Point),
				CorrectPosition: int(correctPosition),
				GroupPosition:   groupPosition,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.AnswerPosition](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "ordering":
		modelsAnswers := make([]models.AnswerPosition, 0, len(request.Options.Answers))
		answerIds := make([]int64, 0, len(request.Options.Answers))

		for index, ans := range request.Options.Answers {
			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			correctPosition := 0
			groupPosition := int(ans.GroupPosition)

			for index, id := range request.CorrectAnswers.Ids {
				if id == indexGroup {
					correctPosition = index + 1
					break
				}
			}

			if correctPosition == 0 && ans.CorrectPosition != 0 {
				correctPosition = int(ans.CorrectPosition)
			}

			if groupPosition == 0 {
				groupPosition = correctPosition
			}

			fileUrl := utils.StripDomain(ans.Media.Url, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.AnswerPosition{
				ID:              int64(ans.Id),
				QuestionID:      &questionID,
				Content:         ans.Text,
				FileInfo:        mediaInfo,
				Kind:            ans.Media.Type,
				Point:           float64(ans.Point),
				CorrectPosition: int(correctPosition),
				GroupPosition:   groupPosition,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.AnswerPosition](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "matching":
		modelsAnswers := make([]models.AnswerMatching, 0, len(request.Options.Targets))
		answerIds := make([]int64, 0, len(request.Options.Targets))

		for index, ans := range request.Options.Targets {
			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			var correctPosition int64

			indexGroupKey := strconv.FormatInt(indexGroup, 10)

			correctPositionStr, ok := request.CorrectAnswers.List[indexGroupKey]

			if !ok {
				correctPosition = 0
			} else {
				var err error
				correctPosition, err = strconv.ParseInt(correctPositionStr, 10, 64)
				if err != nil {
					correctPosition = 0
				}
			}

			if correctPosition == 0 && ans.Content.CorrectPosition != 0 {
				correctPosition = int64(ans.Content.CorrectPosition)
			}

			source := prot.SourceContent{}

			groupPosition := 0
			matchingGroupPosition := int(ans.Content.GroupPosition)

			if request.Options != nil && request.Options.Sources != nil {
				for sourceIndex, sou := range request.Options.Sources {
					if int64(sourceIndex+1) == correctPosition {
						source.Media = sou.Content.Media
						source.Text = sou.Content.Text
						groupPosition = int(sou.GroupPosition)
						break
					}
				}
			}

			var fileUrl string
			var kind string
			if source.Media != nil {
				fileUrl = source.Media.Url
				kind = source.Media.Type
			}

			var matchingFileUrl string
			var matchingKind string
			if ans.Content != nil && ans.Content.Media != nil {
				matchingFileUrl = ans.Content.Media.Url
				matchingKind = ans.Content.Media.Type
			}

			fileUrlStrip := utils.StripDomain(fileUrl, models.Storage)
			matchingFileUrlStrip := utils.StripDomain(matchingFileUrl, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrlStrip)
			matchingMediaInfo := s.GetMediaInFo(matchingFileUrlStrip)

			answer := models.AnswerMatching{
				ID:                    int64(ans.Id),
				QuestionID:            &questionID,
				Content:               source.Text,
				MatchingContent:       ans.Content.Text,
				FileInfo:              mediaInfo,
				MatchingFileInfo:      matchingMediaInfo,
				Kind:                  kind,
				MatchingKind:          matchingKind,
				Point:                 float64(ans.Content.Point),
				CorrectPosition:       int(correctPosition),
				GroupPosition:         groupPosition,
				MatchingGroupPosition: matchingGroupPosition,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.AnswerMatching](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "drag_drop":
		modelsAnswers := make([]models.AnswerPosition, 0, len(request.Options.Answers))
		answerIds := make([]int64, 0, len(request.Options.Answers))

		for index, ans := range request.Options.Answers {
			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			indexGroupKey := strconv.FormatInt(indexGroup, 10)

			var correctPosition int64

			correctPositionStr, ok := request.CorrectAnswers.List[indexGroupKey]

			if !ok {
				correctPosition = 0
			} else {
				var err error
				correctPosition, err = strconv.ParseInt(correctPositionStr, 10, 64)
				if err != nil {
					correctPosition = 0
				}
			}

			if correctPosition == 0 && ans.CorrectPosition != 0 {
				correctPosition = int64(ans.CorrectPosition)
			}

			groupPosition := int(ans.GroupPosition)

			if groupPosition == 0 {
				groupPosition = int(correctPosition)
			}

			fileUrl := utils.StripDomain(ans.Media.Url, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.AnswerPosition{
				ID:              int64(ans.Id),
				QuestionID:      &questionID,
				Content:         ans.Text,
				FileInfo:        mediaInfo,
				Kind:            ans.Media.Type,
				Point:           float64(ans.Point),
				CorrectPosition: int(correctPosition),
				GroupPosition:   groupPosition,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.AnswerPosition](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "labeling":
		modelsAnswers := make([]models.AnswerCoordinates, 0, len(request.Options.Labels))
		answerIds := make([]int64, 0, len(request.Options.Labels))

		for index, ans := range request.Options.Labels {
			positionX := 0
			positionY := 0
			positionWidth := 0
			positionHeight := 0

			indexGroup := string(index + 1)
			if ans.Id != 0 {
				indexGroup = string(ans.Id)
			}

			var correctId int64

			correctIdStr, ok := request.CorrectAnswers.List[indexGroup]
			if !ok {
				correctId = 0
			} else {
				var err error
				correctId, err = strconv.ParseInt(correctIdStr, 10, 64)
				if err != nil {
					correctId = 0
				}
			}

			if ok {
				for _, blank := range request.Options.Blanks {
					if blank.Id == correctId {
						positionX = int(blank.Coordinates.X)
						positionY = int(blank.Coordinates.Y)
						positionWidth = int(blank.Coordinates.W)
						positionHeight = int(blank.Coordinates.H)
					}
				}
			} else {
				for indexBlank, blank := range request.Options.Blanks {
					if index == indexBlank {
						positionX = int(blank.Coordinates.X)
						positionY = int(blank.Coordinates.Y)
						positionWidth = int(blank.Coordinates.W)
						positionHeight = int(blank.Coordinates.H)
					}
				}
			}

			groupPosition := int(ans.GroupPosition)

			fileUrl := utils.StripDomain(ans.Media.Url, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.AnswerCoordinates{
				ID:             int64(ans.Id),
				QuestionID:     &questionID,
				Content:        ans.Text,
				FileInfo:       mediaInfo,
				Kind:           ans.Media.Type,
				Point:          float64(ans.Point),
				PositionX:      int(positionX),
				PositionY:      int(positionY),
				PositionWidth:  int(positionWidth),
				PositionHeight: int(positionHeight),
				GroupPosition:  groupPosition,
			}

			if answer.ID != 0 {
				if err := db.MasterDB.Save(&answer).Error; err != nil {
					return err
				} else {
					answerIds = append(answerIds, answer.ID)
				}
			} else {
				modelsAnswers = append(modelsAnswers, answer)
			}
		}

		if err := repositories.DeleteOldAnswers[models.AnswerCoordinates](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "category":
		if request.Options == nil {
			return nil
		}

		modelsAnswers := make([]models.AnswerGroup, 0, len(request.Options.Items))
		answerIds := make([]int64, 0, len(request.Options.Items))
		indexGroups := make([]*int64, 0, len(request.Options.Categories))

		for _, gro := range request.Options.Categories {
			mediaKind := "text"
			mediaURL := ""
			if gro.Media != nil {
				if gro.Media.Type != "" {
					mediaKind = gro.Media.Type
				}
				mediaURL = gro.Media.Url
			}

			fileUrl := utils.StripDomain(mediaURL, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			group := models.GroupAnswer{
				Content:  gro.Name,
				Kind:     mediaKind,
				FileInfo: mediaInfo,
			}

			if gro.Id != 0 {
				var existing models.GroupAnswer
				if err := db.MasterDB.First(&existing, int64(gro.Id)).Error; err == nil {
					group.ID = existing.ID
					group.SortPosition = existing.SortPosition
				}
			}

			if group.ID != 0 {
				if err := db.MasterDB.Save(&group).Error; err != nil {
					return err
				}
			} else {
				if err := db.MasterDB.Create(&group).Error; err != nil {
					return err
				}
			}

			indexGroups = append(indexGroups, &group.ID)
		}

		for index, ans := range request.Options.Items {
			var groupId int64

			indexGroup := int64(index + 1)
			if ans.Id != 0 {
				indexGroup = int64(ans.Id)
			}

			indexGroupKey := strconv.FormatInt(indexGroup, 10)

			if request.CorrectAnswers != nil {
				if groupIdStr, ok := request.CorrectAnswers.List[indexGroupKey]; ok {
					if parsed, err := strconv.ParseInt(groupIdStr, 10, 64); err == nil {
						groupId = parsed
					}
				}
			}

			if groupId == 0 && ans.GroupPosition > 0 && int(ans.GroupPosition) <= len(indexGroups) {
				if groupIdPtr := indexGroups[ans.GroupPosition-1]; groupIdPtr != nil {
					groupId = *groupIdPtr
				}
			}

			if groupId == 0 {
				continue
			}

			itemMediaKind := "text"
			itemMediaURL := ""
			if ans.Media != nil {
				if ans.Media.Type != "" {
					itemMediaKind = ans.Media.Type
				}
				itemMediaURL = ans.Media.Url
			}

			fileUrl := utils.StripDomain(itemMediaURL, models.Storage)
			mediaInfo := s.GetMediaInFo(fileUrl)

			answer := models.AnswerGroup{
				ID:         int64(ans.Id),
				QuestionID: &questionID,
				Content:    ans.Text,
				FileInfo:   mediaInfo,
				Kind:       itemMediaKind,
				Point:      float64(ans.Point),
				GroupID:    &groupId,
			}

			if answer.ID != 0 {
				var existing models.AnswerGroup
				if err := db.MasterDB.First(&existing, answer.ID).Error; err == nil && existing.QuestionID != nil && *existing.QuestionID == questionID {
					if err := db.MasterDB.Save(&answer).Error; err != nil {
						return err
					}
					answerIds = append(answerIds, answer.ID)
					continue
				}
				answer.ID = 0
			}

			modelsAnswers = append(modelsAnswers, answer)
		}

		if err := repositories.DeleteOldAnswers[models.AnswerGroup](questionID, answerIds); err != nil {
			return err
		}

		if err := repositories.StoreAnswers(modelsAnswers); err != nil {
			return err
		}

		return nil
	case "writing":
		return nil
	case "speaking":
		return nil
	}

	return nil
}

func (s *questionService) GetMediaInFo(url string) models.MediaInfo {
	mediaInfo := models.MediaInfo{
		Path: url,
		Disk: models.Storage,
	}

	mediaRepo := repositories.NewMediaRepository()
	media, err := mediaRepo.FindByStaticUrl(url, models.Storage)

	if err == nil {
		mediaInfo.Id = media.ID
		mediaInfo.Path = *media.StaticURL
	}

	return mediaInfo
}
