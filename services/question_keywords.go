package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"strings"
)

func appendKeyword(keywords []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return keywords
	}
	return append(keywords, value)
}

func appendAnswerContentKeywords(keywords []string, items []*prot.AnswerContent) []string {
	for _, item := range items {
		if item != nil {
			keywords = appendKeyword(keywords, item.Text)
		}
	}
	return keywords
}

func buildKeywordsFromRequest(req *prot.Question) string {
	if req == nil {
		return ""
	}

	var keywords []string

	if req.Metadata != nil {
		keywords = appendKeyword(keywords, req.Metadata.Instructions)
		keywords = appendKeyword(keywords, req.Metadata.Description)
	}
	if req.Content != nil {
		keywords = appendKeyword(keywords, req.Content.Text)
	}
	if req.Options != nil {
		keywords = appendAnswerContentKeywords(keywords, req.Options.Answers)
		keywords = appendAnswerContentKeywords(keywords, req.Options.Labels)
		keywords = appendAnswerContentKeywords(keywords, req.Options.Items)
		keywords = appendAnswerContentKeywords(keywords, req.Options.Questions)
		for _, group := range req.Options.Groups {
			if group != nil {
				keywords = appendKeyword(keywords, group.Text)
			}
		}
		for _, category := range req.Options.Categories {
			if category != nil {
				keywords = appendKeyword(keywords, category.Name)
			}
		}
		for _, source := range req.Options.Sources {
			if source != nil && source.Content != nil {
				keywords = appendKeyword(keywords, source.Content.Text)
			}
		}
		for _, target := range req.Options.Targets {
			if target != nil && target.Content != nil {
				keywords = appendKeyword(keywords, target.Content.Text)
			}
		}
	}

	return strings.Join(keywords, " | ")
}

func (s *questionService) syncKeywordsFromRequest(questionID int64, req *prot.Question) error {
	keywords := buildKeywordsFromRequest(req)
	return db.MasterDB.Model(&models.Question{}).
		Where("id = ?", questionID).
		Update("keywords", keywords).Error
}
