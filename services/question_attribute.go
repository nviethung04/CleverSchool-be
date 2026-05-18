package services

import (
	"be-Clever School/models"
	"be-Clever School/prot"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *questionService) StoreAttribute(c *gin.Context, id int64, req *prot.Question) error {
	attributeIds := make([]int64, 0, len(req.Attributes))

	for _, value := range req.Attributes {
		attribute := models.QuestionRefAttribute{
			QuestionID:        id,
			AttributeID:       value.Value.Id,
			ParentAttributeID: &value.Id,
		}

		s.repo.UpdateOrCreateAttribute(attribute)

		attributeIds = append(attributeIds, value.Value.Id)
	}

	s.repo.DeleteOldAttribute(id, attributeIds)

	return nil
}

func (s *questionService) SaveAttributes(questionID int64, row []string, attributes []models.QuestionAttribute, col int) error {
	for i, attribute := range attributes {
		currentCol := i + col + 1
		if currentCol >= len(row) {
			continue
		}
		strValue := strings.TrimSpace(row[currentCol])
		if strValue == "" {
			continue
		}

		attrID, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid attribute id at column %d: %v", i+11, err)
		}

		if attrID != 0 {
			s.repo.UpdateOrCreateAttribute(models.QuestionRefAttribute{
				ParentAttributeID: &attribute.ID,
				AttributeID:       attrID,
				QuestionID:        questionID,
			})
		}
	}

	return nil
}
