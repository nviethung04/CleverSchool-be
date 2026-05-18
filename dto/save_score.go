package dto

import "be-Clever School/models"

type AnswerWithScoreMultipleChoice struct {
	models.Answer
	Score float64
}

type AnswerWithScoreFillInBlank struct {
	models.Answer
	Score float64
}
