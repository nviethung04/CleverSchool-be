package services

import (
	"be-lms/models"
	"be-lms/repositories"
	"encoding/json"
	_ "fmt"
	"strconv"
)

type ExerciseUserService interface {
	SaveExerciseUser(user *models.ExerciseUser) error
	CalculateExerciseScoreService(exerciseID, userID int64) (float64, error)
	SaveExerciseScoreService(exerciseID, userID int64, score float64) error
	CalculateExerciseRatioService(exerciseID, userID int64) (float64, error)
	SaveExerciseRatioService(exerciseID, userID int64, ratio float64) error
}

type exerciseUserService struct {
	repo                  repositories.ExerciseUserRepository
	clonedQuestionService ClonedQuestionService
}

func NewExerciseUserService(repo repositories.ExerciseUserRepository, clonedQuestionService ClonedQuestionService) ExerciseUserService {
	return &exerciseUserService{repo: repo, clonedQuestionService: clonedQuestionService}
}

func (s *exerciseUserService) SaveExerciseUser(user *models.ExerciseUser) error {
	return s.repo.SaveExerciseUser(user)
}

func (s *exerciseUserService) CalculateExerciseScoreService(exerciseID, userID int64) (float64, error) {
	total := 0.0
	if v, err := s.repo.SumScoreFillInBlank(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreGroup(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreLabeling(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreManual(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreMatching(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScorePosition(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	if v, err := s.repo.SumScoreUser(exerciseID, userID); err == nil {
		total += v
	} else {
		return 0, err
	}
	return total, nil
}

func (s *exerciseUserService) SaveExerciseScoreService(exerciseID, userID int64, score float64) error {
	return s.repo.UpdateExerciseUserScore(exerciseID, userID, score)
}

func (s *exerciseUserService) CalculateExerciseRatioService(exerciseID, userID int64) (float64, error) {
	questionsMap, err := s.clonedQuestionService.GetQuestionsMap(exerciseID, "exercise")
	if err != nil {
		return 0, err
	}
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
			var meta struct {
				Attributes json.RawMessage `json:"attributes"`
			}
			_ = json.Unmarshal(q.Metadata, &meta)
			if len(meta.Attributes) > 0 {
				_ = json.Unmarshal(meta.Attributes, &attributes)
			}
		}
		if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
			var opts struct {
				Attributes json.RawMessage `json:"attributes"`
			}
			_ = json.Unmarshal(opts.Attributes, &attributes)
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
	// numerator
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
		correct, err := s.repo.GetCorrectCountByTable(table, exerciseID, userID)
		if err != nil {
			return 0, err
		}
		answers, err := s.repo.GetAnswerCountByTable(table, exerciseID, userID)
		if err != nil {
			return 0, err
		}
		for qid, total := range answers {
			c := correct[qid]
			ratio := 0.0
			if total > 0 {
				ratio = float64(c) / float64(total)
			}
			weight := 1.0
			if q, ok := questionsMap[fmtInt64(qid)]; ok {
				var attributes []struct {
					Value struct {
						Weight int `json:"weight"`
					} `json:"value"`
				}
				if q.Metadata != nil && len(q.Metadata) > 0 {
					var meta struct {
						Attributes json.RawMessage `json:"attributes"`
					}
					_ = json.Unmarshal(q.Metadata, &meta)
					if len(meta.Attributes) > 0 {
						_ = json.Unmarshal(meta.Attributes, &attributes)
					}
				}
				if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
					var opts struct {
						Attributes json.RawMessage `json:"attributes"`
					}
					_ = json.Unmarshal(opts.Attributes, &attributes)
				}
				w := 0
				for _, a := range attributes {
					w += a.Value.Weight
				}
				if w > 0 {
					weight = float64(w)
				}
			}
			numerator += ratio * weight
		}
	}
	manualScores, err := s.repo.GetManualScoringByExercise(exerciseID, userID)
	if err != nil {
		return 0, err
	}
	for qid, score := range manualScores {
		weight := 1.0
		if q, ok := questionsMap[fmtInt64(qid)]; ok {
			var attributes []struct {
				Value struct {
					Weight int `json:"weight"`
				} `json:"value"`
			}
			if q.Metadata != nil && len(q.Metadata) > 0 {
				var meta struct {
					Attributes json.RawMessage `json:"attributes"`
				}
				_ = json.Unmarshal(q.Metadata, &meta)
				if len(meta.Attributes) > 0 {
					_ = json.Unmarshal(meta.Attributes, &attributes)
				}
			}
			if q.Options != nil && len(q.Options) > 0 && len(attributes) == 0 {
				var opts struct {
					Attributes json.RawMessage `json:"attributes"`
				}
				_ = json.Unmarshal(opts.Attributes, &attributes)
			}
			w := 0
			for _, a := range attributes {
				w += a.Value.Weight
			}
			if w > 0 {
				weight = float64(w)
			}
		}
		numerator += (score / 10.0) * weight
	}
	return (numerator / denominator) * 100, nil
}

func (s *exerciseUserService) SaveExerciseRatioService(exerciseID, userID int64, ratio float64) error {
	return s.repo.UpdateExerciseUserRatio(exerciseID, userID, ratio)
}

func fmtInt64(v int64) string { return strconv.FormatInt(v, 10) }
