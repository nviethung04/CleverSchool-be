package services

import (
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"encoding/json"
	"strconv"
)

type ExamUserService interface {
	SaveExamUser(examUser *models.ExamUser) error
	CalculateExamScoreService(examID, userID int64) (float64, error)
	SaveExamScoreService(examID, userID int64, score float64) error
	CalculateExamRatioService(examID, userID int64) (float64, error)
	SaveExamRatioService(examID, userID int64, score float64) error
	SaveExerciseUser(exerciseUser *models.ExerciseUser) error
	CalculateExerciseScoreService(exerciseID, userID int64) (float64, error)
	SaveExerciseScoreService(exerciseID, userID int64, score float64) error
	CalculateExerciseRatioService(exerciseID, userID int64) (float64, error)
	SaveExerciseRatioService(exerciseID, userID int64, score float64) error
}

type examUserService struct {
	repo                  repositories.ExamUserRepository
	clonedQuestionService ClonedQuestionService
}

func NewExamUserService(repo repositories.ExamUserRepository, clonedQuestionService ClonedQuestionService) ExamUserService {
	return &examUserService{
		repo:                  repo,
		clonedQuestionService: clonedQuestionService,
	}
}

func (s *examUserService) SaveExamUser(examUser *models.ExamUser) error {
	return s.repo.SaveExamUser(examUser)
}

func (s *examUserService) SaveExerciseUser(exerciseUser *models.ExerciseUser) error {
	return s.repo.SaveExerciseUser(exerciseUser)
}

func (s *examUserService) CalculateExamScoreService(examID, userID int64) (float64, error) {
	total := 0.0
	if v, err := s.repo.SumScoreFillInBlank(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreGroup(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreLabeling(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreManual(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreMatching(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScorePosition(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreUser(examID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	return total, nil
}

func (s *examUserService) SaveExamScoreService(examID, userID int64, score float64) error {
	return s.repo.UpdateExamUserScore(examID, userID, score)
}

func (s *examUserService) CalculateExamRatioService(examID, userID int64) (float64, error) {
	// Lấy danh sách câu hỏi từ cloned_question
	questionsMap, err := s.clonedQuestionService.GetQuestionsMap(examID, "exam")
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
		"exam_question_user_fill_in_blanks",
		"exam_question_user_groups",
		"exam_question_user_labelings",
		"exam_question_user_matchings",
		"exam_question_user_positions",
		"exam_question_users",
	}
	for _, table := range tables {
		correct, err := s.repo.GetCorrectCountByTable(table, examID, userID)
		if err != nil {
			return 0, err
		}
		answerCount, err := s.repo.GetAnswerCountByTable(table, examID, userID)
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
	manualScores, err := s.repo.GetManualScoringByExam(examID, userID)
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

func (s *examUserService) SaveExamRatioService(examID, userID int64, score float64) error {
	return s.repo.UpdateExamUserRatio(examID, userID, score)
}

func (s *examUserService) CalculateExerciseScoreService(exerciseID, userID int64) (float64, error) {
	total := 0.0
	if v, err := s.repo.SumScoreFillInBlankExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreGroupExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreLabelingExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreManualExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreMatchingExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScorePositionExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreUserExercise(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	return total, nil
}

func (s *examUserService) SaveExerciseScoreService(exerciseID, userID int64, score float64) error {
	return s.repo.UpdateExerciseUserScore(exerciseID, userID, score)
}

func (s *examUserService) CalculateExerciseRatioService(exerciseID, userID int64) (float64, error) {
	// Lấy danh sách câu hỏi từ cloned_question
	questionsMap, err := s.clonedQuestionService.GetQuestionsMap(exerciseID, "exercise")
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
		"exercise_question_user_fill_in_blanks",
		"exercise_question_user_groups",
		"exercise_question_user_labelings",
		"exercise_question_user_matchings",
		"exercise_question_user_positions",
		"exercise_question_users",
	}
	for _, table := range tables {
		correct, err := s.repo.GetCorrectCountByTableExercise(table, exerciseID, userID)
		if err != nil {
			return 0, err
		}
		answerCount, err := s.repo.GetAnswerCountByTableExercise(table, exerciseID, userID)
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
	manualScores, err := s.repo.GetManualScoringByExercise(exerciseID, userID)
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

func (s *examUserService) SaveExerciseRatioService(exerciseID, userID int64, score float64) error {
	return s.repo.UpdateExerciseUserRatio(exerciseID, userID, score)
}

