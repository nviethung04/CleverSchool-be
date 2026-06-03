package services

import (
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *questionService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, []repositories.QuesstionScore, error) {
	filter = utils.ApplyFilterDate(c, filter)

	homeworkIDStr := c.Query("homework_id")
	lessonPlanPartIDStr := c.Query("lesson_plan_part_id")
	examIDStr := c.Query("exam_id")
	levelTestIDStr := c.Query("level_test_id")
	questionAttributeIDsStr := c.Query("attribute_ids")

	if v, ok := filter["homework_id"]; ok {
		homeworkIDStr = fmt.Sprintf("%v", v)
	}

	var allQuestionIds []int64
	var questionScores []repositories.QuesstionScore

	appendQuestionIDs := func(ids []int64) {
		allQuestionIds = append(allQuestionIds, ids...)
	}

	if homeworkIDStr != "" {
		_, err := strconv.ParseInt(homeworkIDStr, 10, 64)
		if err != nil {
			return filter, questionScores, err
		}

		// Không còn bảng homework_questions, trả về slice rỗng
		records := []repositories.QuesstionScore{}
		questionScores = records
	}

	if lessonPlanPartIDStr != "" {
		lessonPlanPartID, err := strconv.ParseInt(lessonPlanPartIDStr, 10, 64)
		if err != nil {
			return filter, questionScores, err
		}

		records, err := s.repo.GetQuestionIdsAndScoresByLessonPlanPartId(lessonPlanPartID)
		if err != nil {
			return filter, questionScores, err
		}

		for _, record := range records {
			appendQuestionIDs([]int64{record.QuestionID})
		}

		questionScores = records
	}

	if examIDStr != "" {
		_, err := strconv.ParseInt(examIDStr, 10, 64)
		if err != nil {
			return filter, questionScores, err
		}

		// Không còn bảng exam_questions, trả về slice rỗng
		records := []repositories.QuesstionScore{}
		questionScores = records
	}

	if levelTestIDStr != "" {
		levelTestID, err := strconv.ParseInt(levelTestIDStr, 10, 64)
		if err != nil {
			return filter, questionScores, err
		}

		records, err := s.repo.GetQuestionIdsAndScoresByLevelTestId(levelTestID)
		if err != nil {
			return filter, questionScores, err
		}

		for _, record := range records {
			appendQuestionIDs([]int64{record.QuestionID})
		}

		questionScores = records
	}

	if questionAttributeIDsStr != "" {
		ids := utils.ParseIDs(questionAttributeIDsStr)
		var idStrs []string
		for _, id := range ids {
			idStrs = append(idStrs, fmt.Sprintf("%d", id))
		}
		filter["question_attribute_id"] = fmt.Sprintf("in:%s", strings.Join(idStrs, ","))
	}

	if len(allQuestionIds) > 0 {
		filter["id"] = fmt.Sprintf("in:%s", strings.Trim(strings.Replace(fmt.Sprint(allQuestionIds), " ", ",", -1), "[]"))
	}

	return filter, questionScores, nil
}

func (s *questionService) GetKey(assignmentType string) (string, error) {
	var err error
	var filterKey string

	switch assignmentType {
	case models.ClonedQuestionTypeHomework:
		filterKey = "homework_id"
	case models.ClonedQuestionTypeExam:
		filterKey = "exam_id"
	case models.ClonedQuestionTypeLessonPlanPart:
		filterKey = "lesson_plan_part_id"
	case models.ClonedQuestionTypeLevelTest:
		filterKey = "lesson_plan_part_id"
	case models.ClonedQuestionTypeExercise:
		filterKey = "exercise_id"
	case models.ClonedQuestionTypeContestRound:
		filterKey = "contest_round_id"
	default:
		return filterKey, err
	}

	return filterKey, err
}
