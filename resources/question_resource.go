package resources

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"sort"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
)

type PositionKey struct {
	Id             int64
	PositionX      int32
	PositionY      int32
	PositionWidth  int32
	PositionHeight int32
}

type MatchingKey struct {
	Id      int64
	FileUrl string
	Kind    string
}

type GroupKey struct {
	Id      int64
	FileUrl string
	Kind    string
	Content string
}

type QuestionResource interface {
	FormatQuestion(question *models.Question) *prot.Question
	FormatQuestions(questions []*models.Question) []*prot.Question
	FormatModelQuestion(question *prot.Question) *models.Question
	ParseProtQuestionFromJSON(data []byte) (*prot.Question, error)
	FormatStripDomain(question *prot.Question) *prot.Question
	FormatStaticURL(question *prot.Question) *prot.Question
	FormatByRole(question *prot.Question, roleId int64) *prot.Question
	FormatCorrectAnswers(question *prot.Question) *prot.QuestionCorrectAnswers
	FormatGroups(question *prot.Question) []*prot.GroupPosition
	FormatMedias(question *prot.Question) *prot.Question
	FormatMediaUrls(question *prot.Question) *prot.Question
}

type QuestionResourceImpl struct{}

func NewQuestionResource() QuestionResource {
	return &QuestionResourceImpl{}
}

func (resource *QuestionResourceImpl) FormatQuestion(question *models.Question) *prot.Question {
	if question == nil {
		return nil
	}
	var answerPointers []*models.Answer
	for i := range question.Answers {
		answerPointers = append(answerPointers, &question.Answers[i])
	}

	var answerPositions []*models.AnswerPosition
	for i := range question.AnswerPositions {
		answerPositions = append(answerPositions, &question.AnswerPositions[i])
	}

	options := resource.FormatAnswersByType(question, question.QuestionType)
	correctAnswers := resource.FormatCorrectAnswersByType(question, question.QuestionType)

	metaData := prot.QuestionMetaData{
		Instructions: question.Title,
		Description:  question.Description,
		Points:       question.Point,
		Time:         float32(question.TimeLimitSeconds),
		IsRandom:     question.IsRandom != 0,
		Display: question.Display,
		MaxCharacters:    int32(question.MaxCharacters),
		AllowImageUpload: question.AllowImageUpload,
		MaxRecordingTime: int32(question.MaxRecordingTime),
	}

	mediaQuestions := make([]*prot.MediaQuestion, 0, len(question.FileInfos))

	for _, info := range question.FileInfos {
		if info.Path != "" {
			url := utils.StaticURL(info.Path, models.Storage)

			mediaType := info.Type
			if mediaType == "" {
				mediaType = question.Kind
			}

			mediaQuestions = append(mediaQuestions, &prot.MediaQuestion{
				Type: mediaType,
				Url:  url,
			})
		}
	}

	var mediaQuestion *prot.MediaQuestion
	if len(mediaQuestions) > 0 {
		mediaQuestion = mediaQuestions[0]
	}

	questionContent := prot.QuestionContent{
		Text:  question.Content,
		Media: mediaQuestion,
		Medias: mediaQuestions,
	}

	attributes := make([]*prot.QuestionAttributeInfo, 0, len(question.RefAttributes))

	for _, attr := range question.RefAttributes {
		var weight int32
		if attr.Weight > 0 {
			weight = int32(attr.Weight)
		} else {
			weight = int32(attr.ParentAttribute.Weight)
		}

		attributes = append(attributes, &prot.QuestionAttributeInfo{
			Id:   attr.ParentAttribute.ID,
			Name: attr.ParentAttribute.Name,
			Value: &prot.QuestionAttributeValue{
				Id:     attr.AttributeID,
				Name:   attr.Attribute.Name,
				Weight: weight,
			},
		})
	}

	subjectId := int64(0)
	if question.SubjectId != 0 {
		subjectId = question.SubjectId
	}

	var subject *prot.QuestionSubjectInfo
	if question.Subject != nil {
		subject = &prot.QuestionSubjectInfo{
			Id:   question.Subject.ID,
			Name: question.Subject.Name,
			Description: question.Subject.Description,
		}
	}

	questionFormatted := &prot.Question{
		Id:             question.ID,
		Type:           question.QuestionType,
		Status:         question.Status,
		SortPosition:   int32(question.SortPosition),
		Options:        options,
		Metadata:       &metaData,
		Content:        &questionContent,
		CorrectAnswers: correctAnswers,
		Attributes:     attributes,
		SubjectId:      subjectId,
		Subject:        subject,
		CreatedAt:      question.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      question.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if utils.InArray(question.QuestionType, []string{models.QuestionTypeOrdering, models.QuestionTypeFillInBlanks, models.QuestionTypeDragDrop}) {
		questionFormatted = resource.GetQuestionWithIndexPosition(question, answerPositions, &metaData, &questionContent)
	}
	return questionFormatted
}

func (r *QuestionResourceImpl) FormatQuestions(questions []*models.Question) []*prot.Question {
	result := make([]*prot.Question, 0, len(questions))
	for _, q := range questions {
		if formatted := r.FormatQuestion(q); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *QuestionResourceImpl) FormatModelQuestion(question *prot.Question) *models.Question {
	if question == nil {
		return nil
	}

	var fileInfos []models.MediaDetail
	mediaRepo := repositories.NewMediaRepository()

	for _, media := range question.Content.Medias {
		fileUrl := utils.StripDomain(media.Url, models.Storage)
		info := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		fileInfos = append(fileInfos, models.MediaDetail{
			Path: fileUrl,
			Id:   info.Id,
			Type: media.Type,
			Disk: info.Disk,
		})
	}

	if len(fileInfos) == 0 && question.Content.Media.Url != "" {
		fileUrl := utils.StripDomain(question.Content.Media.Url, models.Storage)
		info := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		fileInfos = append(fileInfos, models.MediaDetail{
			Path: fileUrl,
			Id:   info.Id,
			Type: question.Content.Media.Type,
			Disk: info.Disk,
		})
	}

	subjectId := int64(0)
	if question.SubjectId != 0 {
		subjectId = question.SubjectId
	}

	display := models.DisplayWordCount
	if question.Metadata.Display != "" {
		display = question.Metadata.Display
	}

	mediaKind := "text"
	if question.Content != nil && question.Content.Media != nil && question.Content.Media.Type != "" {
		mediaKind = question.Content.Media.Type
	}

	return &models.Question{
		ID:               question.Id,
		Kind:             mediaKind,
		QuestionType:     question.Type,
		Title:            question.Metadata.Instructions,
		Description:      question.Metadata.Description,
		FileInfos:        fileInfos,
		Content:          question.Content.Text,
		TimeLimitSeconds: int(question.Metadata.Time),
		Point:            float64(question.Metadata.Points),
		IsRandom:         utils.BoolToInt16(question.Metadata.IsRandom),
		Status:           question.Status,
		SourceQuestionId: int64(question.SourceQuestionId),
		SubjectId:        subjectId,
		Display: display,
		MaxCharacters:    int(question.Metadata.MaxCharacters),
		AllowImageUpload: question.Metadata.AllowImageUpload,
		MaxRecordingTime: int(question.Metadata.MaxRecordingTime),
	}
}

func (r *QuestionResourceImpl) ParseProtQuestionFromJSON(data []byte) (*prot.Question, error) {
	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	var protoQuestion prot.Question
	if err := unmarshalOpts.Unmarshal(data, &protoQuestion); err != nil {
		return nil, err
	}

	if protoQuestion.Options != nil {
		protoQuestion.Options.Groups = r.FormatGroups(&protoQuestion)
	}

	if protoQuestion.Metadata != nil && protoQuestion.Metadata.Display == "" {
		protoQuestion.Metadata.Display = models.DisplayWordCount
	}

	return &protoQuestion, nil
}

func (r *QuestionResourceImpl) FormatStripDomain(question *prot.Question) *prot.Question {
	if question == nil || question.Content == nil || question.Content.Media == nil {
		return question
	}

	question.Content.Media.Url = utils.StripDomain(question.Content.Media.Url, models.Storage)

	for index, media := range question.Content.Medias {
		question.Content.Medias[index].Url = utils.StripDomain(media.Url, models.Storage)
	}

	if question.Options != nil {
		for indexAnswer, answer := range question.Options.Answers {
			if answer.Media != nil {
				question.Options.Answers[indexAnswer].Media.Url = utils.StripDomain(answer.Media.Url, models.Storage)
			}
		}

		for indexSource, source := range question.Options.Sources {
			if source.Content != nil && source.Content.Media != nil {
				question.Options.Sources[indexSource].Content.Media.Url = utils.StripDomain(source.Content.Media.Url, models.Storage)
			}
		}

		for indexTarget, target := range question.Options.Targets {
			if target.Content != nil && target.Content.Media != nil {
				question.Options.Targets[indexTarget].Content.Media.Url = utils.StripDomain(target.Content.Media.Url, models.Storage)
			}
		}

		for indexLabel, label := range question.Options.Labels {
			if label.Media != nil {
				question.Options.Labels[indexLabel].Media.Url = utils.StripDomain(label.Media.Url, models.Storage)
			}
		}

		for indexItem, item := range question.Options.Items {
			if item.Media != nil {
				question.Options.Items[indexItem].Media.Url = utils.StripDomain(item.Media.Url, models.Storage)
			}
		}

		for indexQuestion, questionItem := range question.Options.Questions {
			if questionItem.Media != nil {
				question.Options.Questions[indexQuestion].Media.Url = utils.StripDomain(questionItem.Media.Url, models.Storage)
			}
		}

		for indexCategory, category := range question.Options.Categories {
			if category.Media != nil {
				question.Options.Categories[indexCategory].Media.Url = utils.StripDomain(category.Media.Url, models.Storage)
			}
		}

		for indexGroup, group := range question.Options.Groups {
			if group.Media != nil {
				question.Options.Groups[indexGroup].Media.Url = utils.StripDomain(group.Media.Url, models.Storage)
			}
		}
	}

	return question
}

func (r *QuestionResourceImpl) FormatByRole(question *prot.Question, roleId int64) *prot.Question {
	if roleId == models.StudentRoleId {
		question.CorrectAnswers = nil
		return question
	}

	return question
}

func (r *QuestionResourceImpl) FormatMedias(question *prot.Question) *prot.Question {
	if question.Content.Media.Url != "" && len(question.Content.Medias) == 0 {
		question.Content.Medias = append(question.Content.Medias, &prot.MediaQuestion{
			Url: question.Content.Media.Url,
			Type: question.Content.Media.Type,
			File: question.Content.Media.File,
		})
	}

	question.Content.Media = nil

	if len(question.Content.Medias) > 0 {
		question.Content.Media = question.Content.Medias[0]
	}

	return question
}

func (r *QuestionResourceImpl) FormatMediaUrls(question *prot.Question) *prot.Question {
	if question == nil || question.Content == nil {
		return question
	}

	if len(question.Content.Medias) == 0 && question.Content.Media != nil && question.Content.Media.Url != "" {
		question.Content.Medias = append(question.Content.Medias, &prot.MediaQuestion{
			Url:  utils.StaticURL(question.Content.Media.Url, models.Storage),
			Type: question.Content.Media.Type,
			File: question.Content.Media.File,
		})
		return question
	}

	for i, media := range question.Content.Medias {
		if media != nil && media.Url != "" {
			question.Content.Medias[i].Url = utils.StaticURL(media.Url, models.Storage)
		}
	}

	return question
}

func (r *QuestionResourceImpl) FormatStaticURL(question *prot.Question) *prot.Question {
	if question.Content != nil && question.Content.Media != nil {
		question.Content.Media.Url = utils.StaticURL(question.Content.Media.Url, models.Storage)

		for i, media := range question.Content.Medias {
			question.Content.Medias[i].Url = utils.StaticURL(media.Url, models.Storage)
		}
	}

	if question.Options != nil {
		for i, answer := range question.Options.Answers {
			if answer.Media != nil {
				question.Options.Answers[i].Media.Url = utils.StaticURL(answer.Media.Url, models.Storage)
			}
		}

		for i, source := range question.Options.Sources {
			if source.Content != nil && source.Content.Media != nil {
				question.Options.Sources[i].Content.Media.Url = utils.StaticURL(source.Content.Media.Url, models.Storage)
			}
		}

		for i, target := range question.Options.Targets {
			if target.Content != nil && target.Content.Media != nil {
				question.Options.Targets[i].Content.Media.Url = utils.StaticURL(target.Content.Media.Url, models.Storage)
			}
		}

		for i, label := range question.Options.Labels {
			if label.Media != nil {
				question.Options.Labels[i].Media.Url = utils.StaticURL(label.Media.Url, models.Storage)
			}
		}

		for i, item := range question.Options.Items {
			if item.Media != nil {
				question.Options.Items[i].Media.Url = utils.StaticURL(item.Media.Url, models.Storage)
			}
		}

		for i, questionItem := range question.Options.Questions {
			if questionItem.Media != nil {
				question.Options.Questions[i].Media.Url = utils.StaticURL(questionItem.Media.Url, models.Storage)
			}
		}

		for i, category := range question.Options.Categories {
			if category.Media != nil {
				question.Options.Categories[i].Media.Url = utils.StaticURL(category.Media.Url, models.Storage)
			}
		}

		for i, group := range question.Options.Groups {
			if group.Media != nil {
				question.Options.Groups[i].Media.Url = utils.StaticURL(group.Media.Url, models.Storage)
			}
		}
	}

	return question
}

func (resource *QuestionResourceImpl) FormatCorrectAnswersByType(question *models.Question, questionType string) *prot.QuestionCorrectAnswers {
	switch questionType {
	case models.QuestionTypeCategory:
		answers := question.AnswerGroups

		correctAnswers := make(map[string]string)

		for _, a := range answers {
			key := strconv.FormatInt(a.ID, 10)
			value := strconv.FormatInt(int64(a.Group.ID), 10)
			if a.GroupID != nil && *a.GroupID > 0 {
				value = strconv.FormatInt(*a.GroupID, 10)
			} else if value == "0" && a.Group.ID > 0 {
				value = strconv.FormatInt(int64(a.Group.ID), 10)
			}
			correctAnswers[key] = value
		}

		return &prot.QuestionCorrectAnswers{
			List: correctAnswers,
		}
	case models.QuestionTypeLabeling:
		answers := question.AnswerCoordinates

		correctAnswers := make(map[string]string)

		for _, a := range answers {
			key := strconv.FormatInt(a.ID, 10)
			correctAnswers[key] = key
		}

		return &prot.QuestionCorrectAnswers{
			List: correctAnswers,
		}
	case models.QuestionTypeDragDrop:
		answers := question.AnswerPositions

		correctAnswers := make(map[string]string)

		for _, a := range answers {
			if a.CorrectPosition != 0 {
				key := strconv.FormatInt(int64(a.CorrectPosition), 10)
				value := strconv.FormatInt(int64(a.GroupPosition), 10)
				correctAnswers[key] = value
			}
		}

		return &prot.QuestionCorrectAnswers{
			List: correctAnswers,
		}
	case models.QuestionTypeMatching:
		answers := question.AnswerMatchings

		correctAnswers := make(map[string]string)

		for _, a := range answers {
			key := strconv.FormatInt(a.ID, 10)
			correctAnswers[key] = key
		}

		return &prot.QuestionCorrectAnswers{
			List: correctAnswers,
		}
	case models.QuestionTypeOrdering:
		answers := question.AnswerPositions

		sort.Slice(answers, func(i, j int) bool {
			return answers[i].CorrectPosition < answers[j].CorrectPosition
		})

		indexIds := make([]int64, len(answers))
		for i, a := range answers {
			indexIds[i] = int64(a.ID)
		}

		correctAnswers := make(map[string]string)

		for _, a := range answers {
			key := strconv.FormatInt(a.ID, 10)
			value := strconv.FormatInt(int64(a.GroupPosition), 10)
			correctAnswers[key] = value
		}

		return &prot.QuestionCorrectAnswers{
			Ids:  indexIds,
			List: correctAnswers,
		}
	case models.QuestionTypeFillInBlanks:
		answers := question.AnswerPositions
		correctAnswers := make(map[string]string)

		for _, a := range answers {
			key := strconv.FormatInt(int64(a.CorrectPosition), 10)
			correctAnswers[key] = a.Content
		}

		return &prot.QuestionCorrectAnswers{
			List: correctAnswers,
		}
	case models.QuestionTypeMultipleChoice:
		answers := question.Answers
		var indexIds []int64

		for _, a := range answers {
			if a.IsCorrect {
				indexIds = append(indexIds, int64(a.ID))
			}
		}

		return &prot.QuestionCorrectAnswers{
			Ids: indexIds,
		}
	}

	return nil
}

func (resource *QuestionResourceImpl) FormatCorrectAnswers(question *prot.Question) *prot.QuestionCorrectAnswers {
    if question == nil {
		config.Log.Info("question is nil")
        return nil
    }

    opts := question.Options
    if opts == nil {
        return nil
    }

    switch question.Type {
    case models.QuestionTypeCategory:
		correct := make(map[string]string)

		for _, item := range opts.Items {
			key := strconv.FormatInt(int64(item.Id), 10)
			val := ""

			for index, category := range opts.Categories {
				if (index + 1) == int(item.GroupPosition) {
					val = strconv.FormatInt(int64(category.Id), 10)
				}
			}
			correct[key] = val
		}
		return &prot.QuestionCorrectAnswers{List: correct}
    case models.QuestionTypeLabeling:
        correct := make(map[string]string)
        for _, lbl := range opts.Labels {
            key := strconv.FormatInt(int64(lbl.GetId()), 10)
            correct[key] = key
        }
        return &prot.QuestionCorrectAnswers{List: correct}
    case models.QuestionTypeDragDrop:
        correct := make(map[string]string)
        for _, ans := range opts.Answers {
            if ans.GetCorrectPosition() != 0 {
                key := strconv.FormatInt(int64(ans.GetCorrectPosition()), 10)
                val := strconv.FormatInt(int64(ans.GetGroupPosition()), 10)
                correct[key] = val
            }
        }
        return &prot.QuestionCorrectAnswers{List: correct}
    case models.QuestionTypeMatching:
        correct := make(map[string]string)
        for _, tgt := range opts.Targets {
            key := strconv.FormatInt(tgt.GetId(), 10)
            correct[key] = key
        }
        return &prot.QuestionCorrectAnswers{List: correct}
    case models.QuestionTypeOrdering:
        answers := opts.Answers
        sort.Slice(answers, func(i, j int) bool {
            return answers[i].GetCorrectPosition() < answers[j].GetCorrectPosition()
        })

        ids := make([]int64, 0, len(answers))
        for _, a := range answers {
            ids = append(ids, int64(a.GetId()))
        }

        correct := make(map[string]string)
        for _, a := range answers {
            key := strconv.FormatInt(int64(a.GetId()), 10)
            val := strconv.FormatInt(int64(a.GetGroupPosition()), 10)
            correct[key] = val
        }
        return &prot.QuestionCorrectAnswers{Ids: ids, List: correct}
    case models.QuestionTypeFillInBlanks:
        correct := make(map[string]string)
        for _, a := range opts.Answers {
            key := strconv.FormatInt(int64(a.GetCorrectPosition()), 10)
            correct[key] = a.GetText()
        }
        return &prot.QuestionCorrectAnswers{List: correct}
    case models.QuestionTypeMultipleChoice:
        var ids []int64
        for _, a := range opts.Answers {
            if a.GetIsCorrect() {
                ids = append(ids, int64(a.GetId()))
            }
        }
        return &prot.QuestionCorrectAnswers{Ids: ids}
    }

    return nil
}

func (resource *QuestionResourceImpl) FormatAnswersByType(question *models.Question, questionType string) *prot.QuestionOption {
	switch questionType {
	case models.QuestionTypeCategory:
		answers := question.AnswerGroups
		items := make([]*prot.AnswerContent, 0, len(answers))
		categories := make([]*prot.GroupAnswer, 0, len(answers))
		addedGroups := make(map[uint64]bool)
		groupIDToPosition := make(map[uint64]int32)

		for _, a := range answers {
			groupID := uint64(0)
			if a.GroupID != nil {
				groupID = uint64(*a.GroupID)
			} else if a.Group.ID != 0 {
				groupID = uint64(a.Group.ID)
			}
			if groupID == 0 {
				continue
			}

			if !addedGroups[groupID] {
				cat := &prot.GroupAnswer{
					Id:   groupID,
					Name: a.Group.Content,
					Media: &prot.MediaAnswer{
						Type: a.Group.Kind,
						Url:  utils.StaticURL(a.Group.FileInfo.Path, models.Storage),
					},
				}
				categories = append(categories, cat)
				addedGroups[groupID] = true
				groupIDToPosition[groupID] = int32(len(categories))
			}
		}

		for _, a := range answers {
			groupID := uint64(0)
			if a.GroupID != nil {
				groupID = uint64(*a.GroupID)
			} else if a.Group.ID != 0 {
				groupID = uint64(a.Group.ID)
			}

			items = append(items, &prot.AnswerContent{
				Id:            uint64(a.ID),
				Text:          a.Content,
				Point:         a.Point,
				GroupPosition: groupIDToPosition[groupID],
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
			})
		}

		return &prot.QuestionOption{
			Items:      items,
			Categories: categories,
		}
	case models.QuestionTypeLabeling:
		answers := question.AnswerCoordinates
		labels := make([]*prot.AnswerContent, 0, len(answers))
		blanks := make([]*prot.AnswerBlank, 0, len(answers))
		groupMap := make(map[int32][]models.AnswerCoordinates)
		groups := make([]*prot.GroupPosition, 0, len(answers))

		for _, a := range answers {
			id := uint64(a.ID)
			labels = append(labels, &prot.AnswerContent{
				Id:    id,
				Text:  a.Content,
				Point: a.Point,
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
				GroupPosition: int32(a.GroupPosition),
			})
			blanks = append(blanks, &prot.AnswerBlank{
				Id: int64(id),
				Coordinates: &prot.AnswerCoordinates{
					X: int32(a.PositionX),
					Y: int32(a.PositionY),
					W: int32(a.PositionWidth),
					H: int32(a.PositionHeight),
				},
			})

			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		groups = append(groups,
			BuildGroupPositions[models.AnswerCoordinates](
				groupMap,
				func(a models.AnswerCoordinates) string { return a.Content },
				func(a models.AnswerCoordinates) string { return a.Kind },
				func(a models.AnswerCoordinates) models.MediaInfo { return a.FileInfo },
			)...)

		return &prot.QuestionOption{
			Labels: labels,
			Blanks: blanks,
			Groups: groups,
		}
	case models.QuestionTypeMatching:
		answers := question.AnswerMatchings
		sources := make([]*prot.AnswerSource, 0, len(answers))
		targets := make([]*prot.AnswerTarget, 0, len(answers))
		groups := make([]*prot.GroupPosition, 0, len(answers))
		groupMap := make(map[int32][]models.AnswerMatching)
		matchingGroupMap := make(map[int32][]models.AnswerMatching)

		for _, a := range answers {
			id := int64(a.ID)
			sources = append(sources, &prot.AnswerSource{
				Id: id,
				Content: &prot.SourceContent{
					Text: a.Content,
					Media: &prot.MediaAnswer{
						Type: a.Kind,
						Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
					},
				},
				GroupPosition: int32(a.GroupPosition),
			})

			targets = append(targets, &prot.AnswerTarget{
				Id: id,
				Content: &prot.AnswerContent{
					CorrectPosition: int32(a.CorrectPosition),
					Text:            a.MatchingContent,
					Point:           a.Point,
					Media: &prot.MediaAnswer{
						Type: a.MatchingKind,
						Url:  utils.StaticURL(a.MatchingFileInfo.Path, models.Storage),
					},
					GroupPosition: int32(a.MatchingGroupPosition),
				},
			})

			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}

			if a.MatchingGroupPosition != 0 {
				gp := int32(a.MatchingGroupPosition)
				matchingGroupMap[gp] = append(matchingGroupMap[gp], a)
			}
		}

		groups = append(groups,
			BuildGroupPositions[models.AnswerMatching](
				groupMap,
				func(a models.AnswerMatching) string { return a.Content },
				func(a models.AnswerMatching) string { return a.Kind },
				func(a models.AnswerMatching) models.MediaInfo { return a.FileInfo },
			)...)

		groups = append(groups,
			BuildGroupPositions[models.AnswerMatching](
				matchingGroupMap,
				func(a models.AnswerMatching) string { return a.MatchingContent },
				func(a models.AnswerMatching) string { return a.MatchingKind },
				func(a models.AnswerMatching) models.MediaInfo { return a.MatchingFileInfo },
			)...)

		uniqueGroups := make(map[int32]*prot.GroupPosition)
		for _, g := range groups {
			if _, exists := uniqueGroups[g.Position]; !exists {
				uniqueGroups[g.Position] = g
			}
		}

		groups = make([]*prot.GroupPosition, 0, len(uniqueGroups))
		for _, g := range uniqueGroups {
			groups = append(groups, g)
		}

		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Position < groups[j].Position
		})

		return &prot.QuestionOption{
			Sources: sources,
			Targets: targets,
			Groups:  groups,
		}
	case models.QuestionTypeOrdering:
		answers := question.AnswerPositions
		new_answers := make([]*prot.AnswerContent, 0, len(answers))
		groupMap := make(map[int32][]models.AnswerPosition)
		groups := make([]*prot.GroupPosition, 0, len(answers))

		for _, a := range answers {
			new_answers = append(new_answers, &prot.AnswerContent{
				Id:              uint64(a.ID),
				Text:            a.Content,
				Point:           a.Point,
				CorrectPosition: int32(a.CorrectPosition),
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
				GroupPosition: int32(a.GroupPosition),
			})

			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		groups = append(groups,
			BuildGroupPositions[models.AnswerPosition](
				groupMap,
				func(a models.AnswerPosition) string { return a.Content },
				func(a models.AnswerPosition) string { return a.Kind },
				func(a models.AnswerPosition) models.MediaInfo { return a.FileInfo },
			)...)

		return &prot.QuestionOption{
			Answers: new_answers,
			Groups:  groups,
		}

	case models.QuestionTypeFillInBlanks:
		answers := question.AnswerPositions
		new_answers := make([]*prot.AnswerContent, 0, len(answers))
		groupMap := make(map[int32][]models.AnswerPosition)
		groups := make([]*prot.GroupPosition, 0, len(answers))

		for _, a := range answers {
			new_answers = append(new_answers, &prot.AnswerContent{
				Id:              uint64(a.ID),
				Text:            a.Content,
				Point:           a.Point,
				CorrectPosition: int32(a.CorrectPosition),
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
				GroupPosition: int32(a.GroupPosition),
			})

			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		groups = append(groups,
			BuildGroupPositions[models.AnswerPosition](
				groupMap,
				func(a models.AnswerPosition) string { return a.Content },
				func(a models.AnswerPosition) string { return a.Kind },
				func(a models.AnswerPosition) models.MediaInfo { return a.FileInfo },
			)...)

		return &prot.QuestionOption{
			Answers: new_answers,
			Groups:  groups,
		}
	case models.QuestionTypeMultipleChoice:
		answers := question.Answers

		new_answers := make([]*prot.AnswerContent, 0, len(answers))

		for _, a := range answers {
			new_answers = append(new_answers, &prot.AnswerContent{
				Id:        uint64(a.ID),
				Text:      a.Content,
				Point:     a.Point,
				IsCorrect: a.IsCorrect,
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
			})
		}

		return &prot.QuestionOption{
			Answers: new_answers,
		}
	case models.QuestionTypeDragDrop:
		answers := question.AnswerPositions
		new_answers := make([]*prot.AnswerContent, 0, len(answers))
		groupMap := make(map[int32][]models.AnswerPosition)
		groups := make([]*prot.GroupPosition, 0, len(answers))

		for _, a := range answers {
			new_answers = append(new_answers, &prot.AnswerContent{
				Id:              uint64(a.ID),
				Text:            a.Content,
				Point:           a.Point,
				CorrectPosition: int32(a.CorrectPosition),
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
				GroupPosition: int32(a.GroupPosition),
			})

			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		groups = append(groups,
			BuildGroupPositions[models.AnswerPosition](
				groupMap,
				func(a models.AnswerPosition) string { return a.Content },
				func(a models.AnswerPosition) string { return a.Kind },
				func(a models.AnswerPosition) models.MediaInfo { return a.FileInfo },
			)...)

		return &prot.QuestionOption{
			Answers: new_answers,
			Groups:  groups,
		}
	case models.QuestionTypeWriting:
		return nil
	case models.QuestionTypeSpeaking:
		return nil
	default:
		answers := question.Answers
		new_answers := make([]*prot.AnswerContent, 0, len(answers))

		for _, a := range answers {
			new_answers = append(new_answers, &prot.AnswerContent{
				Id:    uint64(a.ID),
				Text:  a.Content,
				Point: a.Point,
				Media: &prot.MediaAnswer{
					Type: a.Kind,
					Url:  utils.StaticURL(a.FileInfo.Path, models.Storage),
				},
			})
		}

		return &prot.QuestionOption{
			Answers: new_answers,
		}
	}
}

func (resource *QuestionResourceImpl) GetIndexPositionAnswers(answers []*models.AnswerPosition) map[int32][]*models.AnswerPosition {
	positionAnswers := make(map[int32][]*models.AnswerPosition)

	for _, answer := range answers {
		if answer.CorrectPosition > 0 {
			pos := int32(answer.CorrectPosition)
			positionAnswers[pos] = append(positionAnswers[pos], answer)
		}
	}

	return positionAnswers
}

func (resource *QuestionResourceImpl) GetGroupAnswers(answers []*models.AnswerGroup) map[GroupKey][]*models.AnswerGroup {
	groups := make(map[GroupKey][]*models.AnswerGroup)

	for _, answer := range answers {
		key := GroupKey{
			Id:      answer.Group.ID,
			FileUrl: utils.StripDomain(answer.Group.FileInfo.Path, models.Storage),
			Kind:    answer.Group.Kind,
			Content: answer.Group.Content,
		}

		groups[key] = append(groups[key], answer)
	}

	return groups
}

func (resource *QuestionResourceImpl) GetPositionAnswers(answers []*models.AnswerCoordinates) map[PositionKey][]*models.AnswerCoordinates {
	positions := make(map[PositionKey][]*models.AnswerCoordinates)

	for _, answer := range answers {
		key := PositionKey{
			Id:             answer.ID,
			PositionX:      int32(answer.PositionX),
			PositionY:      int32(answer.PositionY),
			PositionWidth:  int32(answer.PositionWidth),
			PositionHeight: int32(answer.PositionHeight),
		}

		positions[key] = append(positions[key], answer)
	}

	return positions
}

func (resource *QuestionResourceImpl) GetMatchingAnswers(answers []*models.Answer) map[MatchingKey][]*models.Answer {
	matchings := make(map[MatchingKey][]*models.Answer)

	for _, answer := range answers {
		key := MatchingKey{
			Id:      answer.ID,
			FileUrl: utils.StripDomain(answer.FileInfo.Path, models.Storage),
			Kind:    answer.Kind,
		}

		matchings[key] = append(matchings[key], answer)
	}

	return matchings
}

func (resource *QuestionResourceImpl) GetQuestionWithIndexPosition(question *models.Question, answers []*models.AnswerPosition, meta *prot.QuestionMetaData, content *prot.QuestionContent) *prot.Question {
	positions := []int32{}
	positionAnswers := resource.GetIndexPositionAnswers(answers)

	correctAnswers := resource.FormatCorrectAnswersByType(question, question.QuestionType)

	for position := range positionAnswers {
		positions = append(positions, position)
	}

	sort.Slice(positions, func(i, j int) bool {
		return positions[i] < positions[j]
	})

	formattedAnswers := resource.FormatAnswersByType(question, question.QuestionType)

	attributes := make([]*prot.QuestionAttributeInfo, 0, len(question.RefAttributes))

	for _, attr := range question.RefAttributes {
		var weight int32
		if attr.Weight > 0 {
			weight = int32(attr.Weight)
		} else {
			weight = int32(attr.ParentAttribute.Weight)
		}

		attributes = append(attributes, &prot.QuestionAttributeInfo{
			Id:   attr.ParentAttribute.ID,
			Name: attr.ParentAttribute.Name,
			Value: &prot.QuestionAttributeValue{
				Id:     attr.AttributeID,
				Name:   attr.Attribute.Name,
				Weight: weight,
			},
		})
	}

	return &prot.Question{
		Id:             int64(question.ID),
		Type:           question.QuestionType,
		SortPosition:   int32(question.SortPosition),
		Options:        formattedAnswers,
		Status:         question.Status,
		Metadata:       meta,
		Content:        content,
		CorrectAnswers: correctAnswers,
		Attributes:     attributes,
		CreatedAt:      question.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      question.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func BuildGroupPositions[T any](
	m map[int32][]T,
	getContent func(T) string,
	getKind func(T) string,
	getMedia func(T) models.MediaInfo,
) []*prot.GroupPosition {
	var positions []int32
	for pos := range m {
		positions = append(positions, pos)
	}

	sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })

	var result []*prot.GroupPosition
	for _, pos := range positions {
		list := m[pos]
		if len(list) == 0 {
			continue
		}
		first := list[0]
		result = append(result, &prot.GroupPosition{
			Position: pos,
			Count:    int32(len(list)),
			Text:     getContent(first),
			Media: &prot.MediaAnswer{
				Type: getKind(first),
				Url:  utils.StaticURL(getMedia(first).Path, models.Storage),
			},
		})
	}
	return result
}

func (resource *QuestionResourceImpl) FormatGroups(question *prot.Question) []*prot.GroupPosition {
	groups := []*prot.GroupPosition{}

	if question == nil || question.Options == nil {
		return groups
	}

	switch question.Type {
	case models.QuestionTypeLabeling:
		groupMap := make(map[int32][]*prot.AnswerContent)
		for _, a := range question.Options.Labels {
			if a == nil {
				continue
			}
			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		for id := range groupMap {
			if len(groupMap[id]) == 0 || groupMap[id][0] == nil {
				continue
			}

			first := groupMap[id][0]
			gp := &prot.GroupPosition{
				Position: id,
				Count:    int32(len(groupMap[id])),
				Text:     first.Text,
			}

			if first.Media != nil {
				gp.Media = &prot.MediaAnswer{
					Type: first.Media.Type,
					Url:  utils.StaticURL(first.Media.Url, models.Storage),
				}
			}

			groups = append(groups, gp)
		}

		question.Options.Groups = groups

	case models.QuestionTypeFillInBlanks,
		models.QuestionTypeOrdering,
		models.QuestionTypeDragDrop:

		groupMap := make(map[int32][]*prot.AnswerContent)
		for _, a := range question.Options.Answers {
			if a == nil {
				continue
			}
			if a.GroupPosition != 0 {
				gp := int32(a.GroupPosition)
				groupMap[gp] = append(groupMap[gp], a)
			}
		}

		for id := range groupMap {
			if len(groupMap[id]) == 0 || groupMap[id][0] == nil {
				continue
			}

			first := groupMap[id][0]
			gp := &prot.GroupPosition{
				Position: id,
				Count:    int32(len(groupMap[id])),
				Text:     first.Text,
			}

			if first.Media != nil {
				gp.Media = &prot.MediaAnswer{
					Type: first.Media.Type,
					Url:  utils.StaticURL(first.Media.Url, models.Storage),
				}
			}

			groups = append(groups, gp)
		}
	}

	return groups
}
