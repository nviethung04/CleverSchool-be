package command

import (
	"be-Clever School/database/db"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type UpdateHomeworkLessonIdsCommand struct{}

func NewUpdateHomeworkLessonIdsCommand() *UpdateHomeworkLessonIdsCommand {
	return &UpdateHomeworkLessonIdsCommand{}
}

func (cmd *UpdateHomeworkLessonIdsCommand) Execute(c *gin.Context) {
	// Danh sách các bảng cần cập nhật lesson_id
	tables := []string{
		"homework_comments",
		"homework_question_user_fill_in_blanks",
		"homework_question_user_groups",
		"homework_question_user_labelings",
		"homework_question_user_manual_scoring",
		"homework_question_user_matchings",
		"homework_question_user_positions",
		"homework_user_skip_questions",
		"homework_users",
	}

	results := make(map[string]interface{})
	totalUpdated := 0
	chunkSize := 100

	for _, table := range tables {
		// Kiểm tra xem bảng có tồn tại không
		if !db.MasterDB.Migrator().HasTable(table) {
			results[table] = map[string]interface{}{
				"status":  "skipped",
				"message": "Table does not exist",
				"updated": 0,
			}
			continue
		}

		// Kiểm tra xem bảng có cột lesson_id không
		if !db.MasterDB.Migrator().HasColumn(table, "lesson_id") {
			results[table] = map[string]interface{}{
				"status":  "skipped",
				"message": "Column lesson_id does not exist",
				"updated": 0,
			}
			continue
		}

		// Xử lý từng bảng với chunk processing
		tableUpdated, chunksProcessed, err := cmd.processTableInChunks(table, chunkSize)
		if err != nil {
			results[table] = map[string]interface{}{
				"status":           "error",
				"message":          fmt.Sprintf("Failed to update table: %v", err),
				"updated":          0,
				"chunks_processed": 0,
			}
			continue
		}

		totalUpdated += tableUpdated
		results[table] = map[string]interface{}{
			"status":           "success",
			"message":          "Updated successfully",
			"updated":          tableUpdated,
			"chunks_processed": chunksProcessed,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Update completed successfully",
		"total_updated": totalUpdated,
		"results":       results,
	})
}

// processTableInChunks xử lý cập nhật từng bảng theo chunk
func (cmd *UpdateHomeworkLessonIdsCommand) processTableInChunks(table string, chunkSize int) (int, int, error) {
	totalUpdated := 0
	chunksProcessed := 0
	lastProcessedID := int64(0)

	for {
		// Lấy danh sách homework_id cần cập nhật trong chunk này
		// Sử dụng ID-based pagination thay vì OFFSET để tránh bỏ sót records
		var homeworkIDs []int64
		query := fmt.Sprintf(`
			SELECT DISTINCT h.homework_id
			FROM %s h
			WHERE h.homework_id IS NOT NULL
			AND h.homework_id > %d
			AND EXISTS (
				SELECT 1 FROM homework_ref_lessons hrl 
				WHERE hrl.homework_id = h.homework_id
			)
			ORDER BY h.homework_id ASC
			LIMIT %d
		`, table, lastProcessedID, chunkSize)

		if err := db.MasterDB.Raw(query).Scan(&homeworkIDs).Error; err != nil {
			return totalUpdated, chunksProcessed, fmt.Errorf("failed to get homework IDs: %v", err)
		}

		// Nếu không còn record nào, thoát khỏi loop
		if len(homeworkIDs) == 0 {
			break
		}

		// Cập nhật chunk này
		chunkUpdated, err := cmd.updateChunk(table, homeworkIDs)
		if err != nil {
			return totalUpdated, chunksProcessed, fmt.Errorf("failed to update chunk: %v", err)
		}

		totalUpdated += chunkUpdated
		chunksProcessed++
		
		// Cập nhật lastProcessedID để chunk tiếp theo
		lastProcessedID = homeworkIDs[len(homeworkIDs)-1]

		// Nếu chunk này ít hơn chunkSize, có nghĩa là đã hết records
		if len(homeworkIDs) < chunkSize {
			break
		}
	}

	return totalUpdated, chunksProcessed, nil
}

// updateChunk cập nhật một chunk homework_ids
func (cmd *UpdateHomeworkLessonIdsCommand) updateChunk(table string, homeworkIDs []int64) (int, error) {
	if len(homeworkIDs) == 0 {
		return 0, nil
	}

	// Tạo placeholders cho IN clause
	placeholders := make([]string, len(homeworkIDs))
	args := make([]interface{}, len(homeworkIDs))
	for i, id := range homeworkIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	// Cập nhật lesson_id cho chunk này
	// Logic: Lấy lesson_id từ homework_ref_lessons dựa vào homework_id
	// Nếu có nhiều lesson_id cho cùng homework_id, lấy cái có created_at lớn nhất
	updateQuery := fmt.Sprintf(`
		UPDATE %s 
		SET lesson_id = hrl.lesson_id
		FROM homework_ref_lessons hrl
		WHERE %s.homework_id = hrl.homework_id
		AND %s.homework_id IN (%s)
		AND hrl.created_at = (
			SELECT MAX(created_at) 
			FROM homework_ref_lessons hrl2 
			WHERE hrl2.homework_id = %s.homework_id
		)
	`, table, table, table, strings.Join(placeholders, ","), table)

	result := db.MasterDB.Exec(updateQuery, args...)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to execute update query: %v", result.Error)
	}

	return int(result.RowsAffected), nil
}
