package services

import (
	"be-lms/config"
	_ "be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"context"
	"encoding/json"
	"strconv"
)

type GetExamAnswersService interface {
	GetExamAnswers(ctx context.Context, req *prot.GetExamAnswersRequest) (*prot.GetExamAnswersResponse, error)
}

type getExamAnswersService struct {
	repo repositories.GetExamAnswersRepository
}

func NewGetExamAnswersService(repo repositories.GetExamAnswersRepository) GetExamAnswersService {
	return &getExamAnswersService{
		repo: repo,
	}
}

func (s *getExamAnswersService) GetExamAnswers(ctx context.Context, req *prot.GetExamAnswersRequest) (*prot.GetExamAnswersResponse, error) {
	clonedRepo := repositories.NewClonedQuestionRepository()
	clonedQuestion, err := clonedRepo.FindByAssignment(req.ExamId, models.ClonedQuestionTypeExam)
	questionResource := resources.NewQuestionResource()

	if err != nil {
		return nil, err
	}

	var rawQuestions []json.RawMessage
	if err := json.Unmarshal(clonedQuestion.Questions, &rawQuestions); err != nil {
		return nil, err
	}

	questions := make([]*prot.Question, 0, len(rawQuestions))
	var totalScore float64
	for _, q := range rawQuestions {
		question, _ := questionResource.ParseProtQuestionFromJSON(q)
		questions = append(questions, question)
		// Lấy score từ metadata nếu có
		var qMap map[string]interface{}
		if err := json.Unmarshal(q, &qMap); err == nil {
			if meta, ok := qMap["metadata"].(map[string]interface{}); ok {
				if score, ok := meta["score"].(float64); ok {
					totalScore += score
				}
			}
		}
	}

	// Lấy thông tin tổng quan
	studentName, _ := s.repo.GetStudentName(req.UserId)
	duration, _, submittedAt, hasManualScoring, ratio, _ := s.repo.GetExamUserInfo(req.ExamId, req.UserId)
	// totalQuestions lấy từ cloned_question
	totalQuestions := len(rawQuestions)
	unscoredQuestions, _ := s.repo.GetUnscoredManualQuestions(req.ExamId, req.UserId)
	comment, _ := s.repo.GetExamComment(req.ExamId, req.UserId)

	overview := &prot.ExamAnswerOverview{
		StudentName:       studentName,
		Duration:          duration,
		TotalScore:        float32(totalScore),
		SubmittedAt:       submittedAt,
		HasManualScoring:  hasManualScoring,
		TotalQuestions:    int32(totalQuestions),
		UnscoredQuestions: int32(unscoredQuestions),
		Comment:           comment,
		Ratio:             float32(ratio),
	}

    var questionAnswers []*prot.QuestionAnswer
    for _, question := range questions {
        var (
            contentText string
            fileURL     string
            fileKind    string
        )

        if question.Content != nil {
            contentText = question.Content.Text
            if question.Content.Media != nil {
                fileURL = utils.StaticURL(question.Content.Media.Url, models.Storage)
                fileKind = question.Content.Media.Type
            } else if contentText != "" {
                fileKind = "text"
            }
        }

        questionAnswer := &prot.QuestionAnswer{
            Id:           question.Id,
            Content:      contentText,
            Type:         question.Type,
            FileUrl:      fileURL,
            QuestionKind: fileKind,
        }

		switch question.Type {
		case "multiple_choice":
			answers, err := s.repo.GetMultipleChoiceAnswer(req.ExamId, question.Id, req.UserId)

			if err != nil || len(answers) == 0 {
				allAnswerIds := []int64{}
				allAnswerContents := []string{}
				allAnswerFileUrls := []string{}
				allAnswerKinds := []string{}
				trueAnswerIds := []int64{}
				trueAnswers := []string{}
				trueAnswerFileUrls := []string{}
				trueAnswerKinds := []string{}

				for _, answer := range question.Options.Answers {
					allAnswerIds = append(allAnswerIds, int64(answer.Id))
					allAnswerContents = append(allAnswerContents, answer.Text)
					allAnswerFileUrls = append(allAnswerFileUrls, utils.StaticURL(answer.Media.Url, models.Storage))
					allAnswerKinds = append(allAnswerKinds, answer.Media.Type)

					if answer.IsCorrect {
						trueAnswerIds = append(trueAnswerIds, int64(answer.Id))
						trueAnswers = append(trueAnswers, answer.Text)
						trueAnswerFileUrls = append(trueAnswerFileUrls, utils.StaticURL(answer.Media.Url, models.Storage))
						trueAnswerKinds = append(trueAnswerKinds, answer.Media.Type)
					}
				}

				answerMultipleChoice := &prot.AnswerMultipleChoice{
					AnswerIds:          []int64{},
					TrueAnswerIds:      trueAnswerIds,
					Answers:            []string{},
					TrueAnswers:        trueAnswers,
					AnswerFileUrls:     []string{},
					AnswerKinds:        []string{},
					TrueAnswerFileUrls: trueAnswerFileUrls,
					TrueAnswerKinds:    trueAnswerKinds,
					IsAllCorrect:       false,
					Score:              0,
					AllAnswerIds:       allAnswerIds,
					AllAnswerFileUrls:  allAnswerFileUrls,
					AllAnswerKinds:     allAnswerKinds,
					AllAnswerContents:  allAnswerContents,
				}
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerMultipleChoice{
					AnswerMultipleChoice: answerMultipleChoice,
				}
			} else {
				allAnswerIds := []int64{}
				allAnswerContents := []string{}
				allAnswerFileUrls := []string{}
				allAnswerKinds := []string{}
				trueAnswerIds := []int64{}
				trueAnswers := []string{}
				trueAnswerFileUrls := []string{}
				trueAnswerKinds := []string{}
				answerIds := []int64{}
				answerContents := []string{}
				answerFileUrls := []string{}
				answerKinds := []string{}

				for _, answer := range question.Options.Answers {
					if answer.IsCorrect {
						trueAnswerIds = append(trueAnswerIds, int64(answer.Id))
						trueAnswers = append(trueAnswers, answer.Text)
						trueAnswerFileUrls = append(trueAnswerFileUrls, utils.StaticURL(answer.Media.Url, models.Storage))
						trueAnswerKinds = append(trueAnswerKinds, answer.Media.Type)
					}
				}

				for _, answer := range answers {
					answerIDsMap := make(map[int64]bool)
					for _, id := range answer.AnswerIDs {
						answerIDsMap[id] = true
					}

					for _, a := range question.Options.Answers {
						if _, found := answerIDsMap[int64(a.Id)]; found {
							answerIds = append(answerIds, int64(a.Id))
							answerContents = append(answerContents, a.Text)
							answerFileUrls = append(answerFileUrls, utils.StaticURL(a.Media.Url, models.Storage))

							answerKinds = append(answerKinds, a.Media.Type)
							break
						}
					}
				}

				for _, a := range question.Options.Answers {
					allAnswerIds = append(allAnswerIds, int64(a.Id))
					allAnswerContents = append(allAnswerContents, a.Text)
					allAnswerFileUrls = append(allAnswerFileUrls, utils.StaticURL(a.Media.Url, models.Storage))
					allAnswerKinds = append(allAnswerKinds, a.Media.Type)
				}

				answer := answers[0]

				answerMultipleChoice := &prot.AnswerMultipleChoice{
					AnswerIds:          answerIds,
					TrueAnswerIds:      trueAnswerIds,
					Answers:            answerContents,
					TrueAnswers:        trueAnswers,
					AnswerFileUrls:     answerFileUrls,
					AnswerKinds:        answerKinds,
					TrueAnswerFileUrls: trueAnswerFileUrls,
					TrueAnswerKinds:    trueAnswerKinds,
					IsAllCorrect:       answer.IsCorrect,
					Score:              float32(answer.Score),
					AllAnswerIds:       allAnswerIds,
					AllAnswerFileUrls:  allAnswerFileUrls,
					AllAnswerKinds:     allAnswerKinds,
					AllAnswerContents:  allAnswerContents,
				}
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerMultipleChoice{
					AnswerMultipleChoice: answerMultipleChoice,
				}
			}

		case "fill_in_blanks":
			var isAllCorrect bool
			answers, err := s.repo.GetFillInBlankAnswer(req.ExamId, question.Id, req.UserId)
			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerFillInBlank{
					AnswerFillInBlank: &prot.AnswerFillInBlank{
						Answers:      []*prot.FillInBlankAnswer{},
						IsAllCorrect: isAllCorrect,
					},
				}
				continue
			}

			var fillInBlankAnswers []*prot.FillInBlankAnswer
			if err != nil || len(answers) == 0 {
				for _, answer := range question.Options.Answers {
					fillInBlankAnswers = append(fillInBlankAnswers, &prot.FillInBlankAnswer{
						SortPosition: answer.CorrectPosition,
						Answer:       answer.Text,
						TrueAnswer:   answer.Text,
						IsCorrect:    false,
						Score:        float32(answer.Point),
					})
				}
			} else {
				isAllCorrect = true
				for _, answer := range answers {
					var (
						trueAnswerContent string
					)

					for _, a := range question.Options.Answers {
						if a.CorrectPosition == int32(answer.SortPosition) {
							trueAnswerContent = a.Text
							break
						}
					}

					fillInBlankAnswers = append(fillInBlankAnswers, &prot.FillInBlankAnswer{
						SortPosition: int32(answer.SortPosition),
						Answer:       answer.Answer,
						TrueAnswer:   trueAnswerContent,
						IsCorrect:    answer.IsCorrect,
						Score:        float32(answer.Score),
					})

					if !answer.IsCorrect && isAllCorrect {
						isAllCorrect = false
					}
				}
			}

			questionAnswer.Answer = &prot.QuestionAnswer_AnswerFillInBlank{
				AnswerFillInBlank: &prot.AnswerFillInBlank{
					Answers:      fillInBlankAnswers,
					IsAllCorrect: isAllCorrect,
				},
			}

		case "ordering", "drag_drop":
			var isAllCorrect bool
			answers, err := s.repo.GetPositionAnswer(req.ExamId, req.UserId, question.Id)
			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerPosition{
					AnswerPosition: &prot.AnswerPosition{
						Answers:      []*prot.PositionAnswer{},
						IsAllCorrect: isAllCorrect,
					},
				}
				config.Log.Error("GetPositionAnswer: ", err)
				break
			}

			var positionAnswers []*prot.PositionAnswer
			if err != nil || len(answers) == 0 {
				for _, answer := range question.Options.Answers {
					if answer.GroupPosition > 0 {
						positionAnswers = append(positionAnswers, &prot.PositionAnswer{
							SortPosition:        answer.CorrectPosition,
							AnswerGroupPosition: int64(answer.GroupPosition),
							AnswerContent:       answer.Text,
							AnswerFileUrl:       utils.StaticURL(answer.Media.Url, models.Storage),
							AnswerFileKind:      answer.Media.Type,
							TrueAnswerId:        int64(answer.Id),
							TrueAnswerContent:   answer.Text,
							TrueAnswerFileUrl:   utils.StaticURL(answer.Media.Url, models.Storage),
							TrueAnswerFileKind:  answer.Media.Type,
							IsCorrect:           false,
							Score:               answer.Point,
						})
					}
				}
			} else {
				isAllCorrect = true
				for _, answer := range answers {
					if answer.SortPosition > 0 {
						var (
							trueAnswerID       int64
							trueAnswerContent  string
							trueAnswerFileURL  string
							trueAnswerFileKind string
							answerContent      string
							answerFileURL      string
							answerFileKind     string
						)

						for _, a := range question.Options.Answers {
							if a.CorrectPosition == answer.SortPosition {
								trueAnswerID = int64(a.Id)
								trueAnswerContent = a.Text
								trueAnswerFileURL = a.Media.Url
								trueAnswerFileKind = a.Media.Type
								break
							}
						}

						for _, a := range question.Options.Answers {
							if a.GroupPosition == int32(answer.AnswerGroupPosition) {
								answerContent = a.Text
								answerFileURL = a.Media.Url
								answerFileKind = a.Media.Type
								break
							}
						}

						positionAnswers = append(positionAnswers, &prot.PositionAnswer{
							SortPosition:        answer.SortPosition,
							AnswerGroupPosition: answer.AnswerGroupPosition,
							AnswerContent:       answerContent,
							AnswerFileUrl:       utils.StaticURL(answerFileURL, models.Storage),
							AnswerFileKind:      answerFileKind,
							TrueAnswerId:        trueAnswerID,
							TrueAnswerContent:   trueAnswerContent,
							TrueAnswerFileUrl:   utils.StaticURL(trueAnswerFileURL, models.Storage),
							TrueAnswerFileKind:  trueAnswerFileKind,
							IsCorrect:           answer.IsCorrect,
							Score:               answer.Score,
						})

						if !answer.IsCorrect && isAllCorrect {
							isAllCorrect = false
						}
					}
				}
			}

			questionAnswer.Answer = &prot.QuestionAnswer_AnswerPosition{
				AnswerPosition: &prot.AnswerPosition{
					Answers:      positionAnswers,
					IsAllCorrect: isAllCorrect,
				},
			}

		case "matching":
			var isAllCorrect bool
			answers, err := s.repo.GetMatchingAnswer(req.ExamId, question.Id, req.UserId)

			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerMatching{
					AnswerMatching: &prot.AnswerMatching{
						Matches:      []*prot.MatchingPairAnswer{},
						IsAllCorrect: isAllCorrect,
					},
				}
				config.Log.Error("GetMatchingAnswer: ", err)
				break
			}

			var matches []*prot.MatchingPairAnswer
			if err != nil || len(answers) == 0 {
				for _, source := range question.Options.Sources {
					var (
						targetId       int64
						targetContent  string
						targetFileUrl  string
						targetFileKind string
						point          float32
					)

					for _, target := range question.Options.Targets {
						if target.Id == source.Id {
							targetId = target.Id
							targetContent = target.Content.Text
							targetFileUrl = target.Content.Media.Url
							targetFileKind = target.Content.Media.Type
							point = float32(target.Content.Point)
							break
						}
					}

					matches = append(matches, &prot.MatchingPairAnswer{
						FirstItemId:       source.Id,
						SecondItemId:      targetId,
						FirstItemContent:  source.Content.Text,
						SecondItemContent: targetContent,
						FirstItemFileUrl:  utils.StaticURL(source.Content.Media.Url, models.Storage),
						FirstItemKind:     source.Content.Media.Type,
						SecondItemFileUrl: utils.StaticURL(targetFileUrl, models.Storage),
						SecondItemKind:    targetFileKind,
						IsCorrect:         false,
						Score:             point,
					})
				}
			} else {
				isAllCorrect = true
				for _, a := range answers {
					var (
						targetContent  string
						targetFileUrl  string
						targetFileKind string
						sourceContent  string
						sourceFileUrl  string
						sourceFileKind string
					)

					for _, target := range question.Options.Targets {
						if target.Id == a.SecondItemID {
							targetContent = target.Content.Text
							targetFileUrl = target.Content.Media.Url
							targetFileKind = target.Content.Media.Type
							break
						}
					}

					for _, source := range question.Options.Sources {
						if source.Id == a.FirstItemID {
							sourceContent = source.Content.Text
							sourceFileUrl = source.Content.Media.Url
							sourceFileKind = source.Content.Media.Type
							break
						}
					}

					matches = append(matches, &prot.MatchingPairAnswer{
						FirstItemId:       a.FirstItemID,
						SecondItemId:      a.SecondItemID,
						FirstItemContent:  sourceContent,
						SecondItemContent: targetContent,
						FirstItemFileUrl:  utils.StaticURL(sourceFileUrl, models.Storage),
						FirstItemKind:     sourceFileKind,
						SecondItemFileUrl: utils.StaticURL(targetFileUrl, models.Storage),
						SecondItemKind:    targetFileKind,
						IsCorrect:         a.IsCorrect,
						Score:             float32(a.Score),
					})

					if !a.IsCorrect && isAllCorrect {
						isAllCorrect = false
					}
				}
			}
			questionAnswer.Answer = &prot.QuestionAnswer_AnswerMatching{
				AnswerMatching: &prot.AnswerMatching{
					Matches:      matches,
					IsAllCorrect: isAllCorrect,
				},
			}

		case "labeling":
			var isAllCorrect bool
			answers, err := s.repo.GetLabelingAnswer(req.ExamId, question.Id, req.UserId)
			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerLabeling{
					AnswerLabeling: &prot.AnswerLabeling{
						Answers:      []*prot.LabelingAnswer{},
						IsAllCorrect: isAllCorrect,
					},
				}
				config.Log.Error("GetLabelingAnswer: ", err)
				continue
			}

			var labelingAnswers []*prot.LabelingAnswer
			if err != nil || len(answers) == 0 {
				for _, label := range question.Options.Labels {
					var (
						blankId             int64
						blankPositionX      float32
						blankPositionY      float32
						blankPositionWidth  float32
						blankPositionHeight float32
					)

					for _, b := range question.Options.Blanks {
						if b.Id == int64(label.Id) {
							blankId = b.Id
							blankPositionX = float32(b.Coordinates.X)
							blankPositionY = float32(b.Coordinates.Y)
							blankPositionWidth = float32(b.Coordinates.W)
							blankPositionHeight = float32(b.Coordinates.H)
							break
						}
					}

					labelingAnswers = append(labelingAnswers, &prot.LabelingAnswer{
						BlankId:             blankId,
						TrueBlankContent:    label.Text,
						TrueBlankFileUrl:    utils.StaticURL(label.Media.Url, models.Storage),
						TrueBlankKind:       label.Media.Type,
						BlankPositionX:      blankPositionX,
						BlankPositionY:      blankPositionY,
						BlankPositionWidth:  blankPositionWidth,
						BlankPositionHeight: blankPositionHeight,
						AnswerId:            int64(label.Id),
						AnswerContent:       label.Text,
						AnswerFileUrl:       utils.StaticURL(label.Media.Url, models.Storage),
						AnswerKind:          label.Media.Type,
						IsCorrect:           false,
						Score:               float32(label.Point),
					})
				}
			} else {
				isAllCorrect = true
				for _, a := range answers {
					var (
						trueBlankContent    string
						trueBlankFileUrl    string
						trueBlankKind       string
						blankPositionX      float32
						blankPositionY      float32
						blankPositionWidth  float32
						blankPositionHeight float32
						answerFileURL       string
						answerKind          string
						answerContent       string
					)

					for _, label := range question.Options.Labels {
						if int64(label.Id) == a.AnswerID {
							answerContent = label.Text
							answerFileURL = utils.StaticURL(label.Media.Url, models.Storage)
							answerKind = label.Media.Type
							break
						}
					}

					for _, label := range question.Options.Labels {
						if int64(label.Id) == a.BlankID {
							trueBlankContent = label.Text
							trueBlankFileUrl = utils.StaticURL(label.Media.Url, models.Storage)
							trueBlankKind = label.Media.Type
							break
						}
					}

					for _, blank := range question.Options.Blanks {
						if int64(blank.Id) == a.BlankID {
							blankPositionX = float32(blank.Coordinates.X)
							blankPositionY = float32(blank.Coordinates.Y)
							blankPositionWidth = float32(blank.Coordinates.W)
							blankPositionHeight = float32(blank.Coordinates.H)
							break
						}
					}

					labelingAnswers = append(labelingAnswers, &prot.LabelingAnswer{
						BlankId:             a.BlankID,
						TrueBlankContent:    trueBlankContent,
						TrueBlankFileUrl:    utils.StaticURL(trueBlankFileUrl, models.Storage),
						TrueBlankKind:       trueBlankKind,
						BlankPositionX:      blankPositionX,
						BlankPositionY:      blankPositionY,
						BlankPositionWidth:  blankPositionWidth,
						BlankPositionHeight: blankPositionHeight,
						AnswerId:            a.AnswerID,
						AnswerContent:       answerContent,
						AnswerFileUrl:       utils.StaticURL(answerFileURL, models.Storage),
						AnswerKind:          answerKind,
						IsCorrect:           a.IsCorrect,
						Score:               float32(a.Score),
					})

					if !a.IsCorrect && isAllCorrect {
						isAllCorrect = false
					}
				}
			}

			questionAnswer.Answer = &prot.QuestionAnswer_AnswerLabeling{
				AnswerLabeling: &prot.AnswerLabeling{
					Answers:      labelingAnswers,
					IsAllCorrect: isAllCorrect,
				},
			}

		case "category":
			answers, err := s.repo.GetGroupAnswer(req.ExamId, question.Id, req.UserId)
			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerGroup{
					AnswerGroup: &prot.AnswerGroup{
						Groups:       []*prot.CategoryAnswer{},
						IsAllCorrect: false,
					},
				}
				config.Log.Error("GetGroupAnswer: ", err)
				continue
			}

			var categoryAnswers []*prot.CategoryAnswer
			isAllCorrect := true

			if err != nil || len(answers) == 0 {
				for _, item := range question.Options.Items {
					var (
						groupId       int64
						groupContent  string
						groupFileUrl  string
						groupFileKind string
					)

					catID, ok := question.CorrectAnswers.List[strconv.FormatUint(item.Id, 10)]
					if ok {
						for _, category := range question.Options.Categories {
							categoryId := strconv.FormatInt(int64(category.Id), 10)
							if categoryId == catID {
								groupId = int64(category.Id)
								groupContent = category.Name
								groupFileUrl = utils.StaticURL(category.Media.Url, models.Storage)
								groupFileKind = category.Media.Type
								break
							}
						}
					}

					categoryAnswers = append(categoryAnswers, &prot.CategoryAnswer{
						AnswerId:       int64(item.Id),
						GroupId:        groupId,
						AnswerContent:  item.Text,
						GroupContent:   groupContent,
						AnswerFileUrl:  utils.StaticURL(item.Media.Url, models.Storage),
						AnswerFileKind: item.Media.Type,
						GroupFileUrl:   groupFileUrl,
						GroupFileKind:  groupFileKind,
						IsCorrect:      false,
						Score:          float32(item.Point),
					})
				}

				isAllCorrect = false
			} else {
				for _, answer := range answers {
					var (
						answerContent  string
						answerFileUrl  string
						answerFileKind string
						groupContent   string
						groupFileUrl   string
						groupFileKind  string
					)

					for _, category := range question.Options.Categories {
						if int64(category.Id) == answer.GroupID {
							groupContent = category.Name
							groupFileUrl = utils.StaticURL(category.Media.Url, models.Storage)
							groupFileKind = category.Media.Type
							break
						}
					}

					for _, item := range question.Options.Items {
						if int64(item.Id) == answer.AnswerID {
							answerContent = item.Text
							answerFileUrl = utils.StaticURL(item.Media.Url, models.Storage)
							answerFileKind = item.Media.Type
							break
						}
					}

					categoryAnswers = append(categoryAnswers, &prot.CategoryAnswer{
						AnswerId:       answer.AnswerID,
						GroupId:        answer.GroupID,
						AnswerContent:  answerContent,
						GroupContent:   groupContent,
						AnswerFileUrl:  answerFileUrl,
						AnswerFileKind: answerFileKind,
						GroupFileUrl:   groupFileUrl,
						GroupFileKind:  groupFileKind,
						IsCorrect:      answer.IsCorrect,
						Score:          float32(answer.Score),
					})
					if !answer.IsCorrect {
						isAllCorrect = false
					}
				}
			}

			questionAnswer.Answer = &prot.QuestionAnswer_AnswerGroup{
				AnswerGroup: &prot.AnswerGroup{
					Groups:       categoryAnswers,
					IsAllCorrect: isAllCorrect,
				},
			}

		case "writing", "speaking":
			answer, err := s.repo.GetManualAnswer(req.ExamId, question.Id, req.UserId)
			if err != nil {
				questionAnswer.Answer = &prot.QuestionAnswer_AnswerManual{
					AnswerManual: &prot.AnswerManual{
						Answer:        "",
						AnswerFileUrl: "",
						Score:         0,
						IsScored:      false,
					},
				}
				config.Log.Error(err)
				continue
			}

			questionAnswer.Answer = &prot.QuestionAnswer_AnswerManual{
				AnswerManual: &prot.AnswerManual{
					Answer:        answer.Answer,
					AnswerFileUrl: utils.StaticURL(answer.AnswerFileInfo.Path, models.Storage),
					Score:         float32(answer.Score),
					IsScored:      answer.IsScored,
				},
			}
		}

		questionAnswers = append(questionAnswers, questionAnswer)
	}

	return &prot.GetExamAnswersResponse{
		Questions: questionAnswers,
		Overview:  overview,
	}, nil
}
