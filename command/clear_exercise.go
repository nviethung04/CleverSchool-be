package command

import (
	"be-Clever School/database/db"

	"github.com/gin-gonic/gin"
)

type ClearExerciseCommand struct{}

func NewClearExerciseCommand() *ClearExerciseCommand {
	return &ClearExerciseCommand{}
}

func (c *ClearExerciseCommand) Execute(ctx *gin.Context) {
	// Xóa theo thứ tự an toàn: bảng con trước, bảng cha sau
	statements := []struct {
		name string
		sql  string
	}{
		{"exercise_question_user_positions", "DELETE FROM exercise_question_user_positions"},
		{"exercise_question_user_matchings", "DELETE FROM exercise_question_user_matchings"},
		{"exercise_question_user_manual_scoring", "DELETE FROM exercise_question_user_manual_scoring"},
		{"exercise_question_user_labelings", "DELETE FROM exercise_question_user_labelings"},
		{"exercise_question_user_groups", "DELETE FROM exercise_question_user_groups"},
		{"exercise_question_user_fill_in_blanks", "DELETE FROM exercise_question_user_fill_in_blanks"},
		{"exercise_question_users", "DELETE FROM exercise_question_users"},
		{"exercise_comments", "DELETE FROM exercise_comments"},
		{"exercise_ref_lessons", "DELETE FROM exercise_ref_lessons"},
		{"exercise_questions", "DELETE FROM exercise_questions"},
		{"exercise_users", "DELETE FROM exercise_users"},
		{"cloned_questions(exercise)", "DELETE FROM cloned_questions WHERE assignment_type = 'exercise'"},
		{"exercises", "DELETE FROM exercises"},
	}

	results := make(map[string]int64)
	var errors []string

	tx := db.MasterDB.Begin()
	for _, st := range statements {
		res := tx.Exec(st.sql)
		if res.Error != nil {
			errors = append(errors, st.name+": "+res.Error.Error())
			continue
		}
		results[st.name] = res.RowsAffected
	}

	if len(errors) > 0 {
		tx.Rollback()
		ctx.JSON(500, gin.H{
			"ok":      false,
			"errors":  errors,
			"deleted": results,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		ctx.JSON(500, gin.H{"ok": false, "error": err.Error()})
		return
	}

	// Tổng số hàng đã xóa
	var total int64
	for _, v := range results {
		total += v
	}

	ctx.JSON(200, gin.H{
		"ok":            true,
		"deleted":       results,
		"deleted_total": total,
		"tables": []string{
			"exercises", "exercise_users", "exercise_ref_lessons", "exercise_questions",
			"exercise_question_users", "exercise_question_user_positions", "exercise_question_user_matchings",
			"exercise_question_user_manual_scoring", "exercise_question_user_labelings",
			"exercise_question_user_groups", "exercise_question_user_fill_in_blanks", "exercise_comments",
			"cloned_questions(exercise)",
		},
		"note": "Hard delete executed for all exercise-related tables and cloned_questions with assignment_type=exercise",
	})
}
