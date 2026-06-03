package services

import (
	"be-lms/models"
	"be-lms/repositories"
	"encoding/json"
	"strconv"
)

type HomeworkUserService interface {
	SaveHomeworkUser(homeworkUser *models.HomeworkUser) error
	CalculateHomeworkScoreService(homeworkID, userID int64) (float64, error)
	SaveHomeworkScoreService(homeworkID, userID int64, score float64) error
	UpdateHomeworkStatusScoringService(homeworkID, userID int64) error
	CalculateHomeworkRatioService(homeworkID, userID int64) (float64, error)
	SaveHomeworkRatioService(homeworkID, userID int64, score float64) error
	CheckHomeworkHasQuestionsNeedManualGrading(homeworkID, userID int64) (bool, error)
}

type homeworkUserService struct {
	repo                  repositories.HomeworkUserRepository
	clonedQuestionService ClonedQuestionService
	manualScoringRepo     repositories.SaveScoreManualScoringRepository
}

func NewHomeworkUserService(repo repositories.HomeworkUserRepository, clonedQuestionService ClonedQuestionService, manualScoringRepo repositories.SaveScoreManualScoringRepository) HomeworkUserService {
	return &homeworkUserService{
		repo:                  repo,
		clonedQuestionService: clonedQuestionService,
		manualScoringRepo:     manualScoringRepo,
	}
}

func (s *homeworkUserService) SaveHomeworkUser(homeworkUser *models.HomeworkUser) error {
	return s.repo.SaveHomeworkUser(homeworkUser)
}

func (s *homeworkUserService) CalculateHomeworkScoreService(homeworkID, userID int64) (float64, error) {
	total := 0.0
	if v, err := s.repo.SumScoreFillInBlank(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreGroup(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreLabeling(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreManual(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreMatching(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScorePosition(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreUser(homeworkID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	return total, nil
}

func (s *homeworkUserService) SaveHomeworkScoreService(homeworkID, userID int64, score float64) error {
	return s.repo.UpdateHomeworkUserScore(homeworkID, userID, score)
}

func (s *homeworkUserService) UpdateHomeworkStatusScoringService(homeworkID, userID int64) error {
	return s.repo.UpdateHomeworkUserStatusScoring(homeworkID, userID)
}

func (s *homeworkUserService) CalculateHomeworkRatioService(homeworkID, userID int64) (float64, error) {
	// Lấy danh sách câu hỏi từ cloned_question
	questionsMap, err := s.clonedQuestionService.GetQuestionsMap(homeworkID, "homework")
	if err != nil {
		return 0, err
	}
	// Tính denominator (tổng weight)
	denominator := 0.0
	for _, q := range questionsMap {
		var attributes []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Value struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Weight int    `json:"weight"`
			} `json:"value"`
		}
		weight := 0
		if q.Metadata != nil && len(q.Metadata) > 0 {
			// Nếu attributes nằm trong metadata
			var meta struct {
				Attributes json.RawMessage `json:"attributes"`
			}
			if err := json.Unmarshal(q.Metadata, &meta); err == nil && len(meta.Attributes) > 0 {
				_ = json.Unmarshal(meta.Attributes, &attributes)
			}
		}
		if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
			// Nếu attributes nằm trong options (phòng trường hợp cũ)
			var opts struct {
				Attributes json.RawMessage `json:"attributes"`
			}
			if err := json.Unmarshal(q.Options, &opts); err == nil && len(opts.Attributes) > 0 {
				_ = json.Unmarshal(opts.Attributes, &attributes)
			}
		}
		for _, attr := range attributes {
			weight += attr.Value.Weight
		}
		if weight == 0 {
			weight = 1
		}
		denominator += float64(weight)
	}
	if denominator == 0 {
		return 0, nil
	}
	// Tính numerator như cũ
	numerator := 0.0
	tables := []string{
		"homework_question_user_fill_in_blanks",
		"homework_question_user_groups",
		"homework_question_user_labelings",
		"homework_question_user_matchings",
		"homework_question_user_positions",
		"homework_question_users",
	}
	for _, table := range tables {
		correct, err := s.repo.GetCorrectCountByTable(table, homeworkID, userID)
		if err != nil {
			return 0, err
		}
		answerCount, err := s.repo.GetAnswerCountByTable(table, homeworkID, userID)
		if err != nil {
			return 0, err
		}
		for qid, cnt := range correct {
			q, ok := questionsMap[strconv.FormatInt(qid, 10)]
			weight := 0
			if ok {
				var attributes []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Value struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Weight int    `json:"weight"`
					} `json:"value"`
				}
				if q.Metadata != nil && len(q.Metadata) > 0 {
					var meta struct {
						Attributes json.RawMessage `json:"attributes"`
					}
					if err := json.Unmarshal(q.Metadata, &meta); err == nil && len(meta.Attributes) > 0 {
						_ = json.Unmarshal(meta.Attributes, &attributes)
					}
				}
				if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
					var opts struct {
						Attributes json.RawMessage `json:"attributes"`
					}
					if err := json.Unmarshal(q.Options, &opts); err == nil && len(opts.Attributes) > 0 {
						_ = json.Unmarshal(opts.Attributes, &attributes)
					}
				}
				for _, attr := range attributes {
					weight += attr.Value.Weight
				}
			}
			if weight == 0 {
				weight = 1
			}
			denom := answerCount[qid]
			if denom == 0 {
				continue
			}
			numerator += float64(cnt) * float64(weight) / float64(denom)
		}
	}
	manualScores, err := s.repo.GetManualScoringByHomework(homeworkID, userID)
	if err != nil {
		return 0, err
	}
	for qid, score := range manualScores {
		q, ok := questionsMap[strconv.FormatInt(qid, 10)]
		weight := 0
		if ok {
			var attributes []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Value struct {
					ID     string `json:"id"`
					Name   string `json:"name"`
					Weight int    `json:"weight"`
				} `json:"value"`
			}
			if q.Metadata != nil && len(q.Metadata) > 0 {
				var meta struct {
					Attributes json.RawMessage `json:"attributes"`
				}
				if err := json.Unmarshal(q.Metadata, &meta); err == nil && len(meta.Attributes) > 0 {
					_ = json.Unmarshal(meta.Attributes, &attributes)
				}
			}
			if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
				var opts struct {
					Attributes json.RawMessage `json:"attributes"`
				}
				if err := json.Unmarshal(q.Options, &opts); err == nil && len(opts.Attributes) > 0 {
					_ = json.Unmarshal(opts.Attributes, &attributes)
				}
			}
			for _, attr := range attributes {
				weight += attr.Value.Weight
			}
		}
		if weight == 0 {
			weight = 1
		}
		numerator += (score / 10.0) * float64(weight)
	}
	ratio := (numerator / denominator) * 100
	return ratio, nil
}

func (s *homeworkUserService) SaveHomeworkRatioService(homeworkID, userID int64, score float64) error {
	return s.repo.UpdateHomeworkUserRatio(homeworkID, userID, score)
}

func (s *homeworkUserService) CheckHomeworkHasQuestionsNeedManualGrading(homeworkID, userID int64) (bool, error) {
	return s.manualScoringRepo.CheckUnscoredQuestionsHomework(homeworkID, userID)
}
