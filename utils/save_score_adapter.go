package utils

import (
	"encoding/json"
	"fmt"
)

// LegacyAnswerRequest represents the old frontend format
type LegacyAnswerRequest struct {
	QuestionID   int64                  `json:"question_id"`
	QuestionType string                 `json:"question_type"`
	AnswerData   map[string]interface{} `json:"answer_data"`
	IsCorrect    bool                   `json:"is_correct"`
	Score        float64                `json:"score"`
}

// LegacySaveScoreBulkRequest represents the old frontend format
type LegacySaveScoreBulkRequest struct {
	ContestRoundID   int64                 `json:"contest_round_id"`
	UserID           int64                 `json:"user_id"`
	Score            float64               `json:"score"`
	Ratio            float64               `json:"ratio"`
	Time             int64                 `json:"time"`
	HasManualScoring bool                  `json:"has_manual_scoring"`
	Answers          []LegacyAnswerRequest `json:"answers"`
}

// ConvertLegacySaveScoreRequest converts old frontend format to new protobuf format
func ConvertLegacySaveScoreRequest(legacyJSON []byte) ([]byte, error) {
	var legacy LegacySaveScoreBulkRequest
	if err := json.Unmarshal(legacyJSON, &legacy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legacy format: %v", err)
	}

	// Convert to new format
	newRequest := map[string]interface{}{
		"contest_round_id": legacy.ContestRoundID,
		"time":             legacy.Time,
		"list_answers":     make([]map[string]interface{}, 0, len(legacy.Answers)),
	}

	// Convert answers
	for _, answer := range legacy.Answers {
		newAnswer := map[string]interface{}{
			"type_question": answer.QuestionType,
			"question_id":   answer.QuestionID,
			"data":          answer.AnswerData,
		}
		newRequest["list_answers"] = append(newRequest["list_answers"].([]map[string]interface{}), newAnswer)
	}

	return json.Marshal(newRequest)
}
