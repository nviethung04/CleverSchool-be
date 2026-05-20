package command

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"time"

	"github.com/gin-gonic/gin"
)

type CopyHomeworkToExerciseCommand struct{}

func NewCopyHomeworkToExerciseCommand() *CopyHomeworkToExerciseCommand {
	return &CopyHomeworkToExerciseCommand{}
}

// Body: {"from_date":"YYYY-MM-DD"}
type copyRequest struct {
	FromDate string `json:"from_date"`
}

func (c *CopyHomeworkToExerciseCommand) Execute(ctx *gin.Context) {
	// Parse input JSON
	var req copyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.FromDate == "" {
		ctx.JSON(400, gin.H{"error": "invalid_request"})
		return
	}

	fromTime, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "invalid_from_date_format (YYYY-MM-DD)"})
		return
	}

	type result struct {
		Total        int     `json:"total"`
		Success      int     `json:"success"`
		Failed       int     `json:"failed"`
		FailedIDs    []int64 `json:"failed_ids"`
		Skipped      int     `json:"skipped"`
		ClonedCopied int     `json:"cloned_copied"`
	}
	res := result{}

	// Process in chunks of 100
	offset := 0
	const chunkSize = 100
	for {
		var homeworks []models.Homework
		if err := db.ReplicaDB.Where("created_at >= ?", fromTime).Order("id").Limit(chunkSize).Offset(offset).Find(&homeworks).Error; err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if len(homeworks) == 0 {
			break
		}
		res.Total += len(homeworks)

		for _, hw := range homeworks {
			// Copy one homework -> exercise
			ex := models.Exercise{
				Name:             hw.Name,
				Status:           hw.Status,
				MaxScore:         hw.MaxScore,
				Description:      hw.Description,
				CoverImageInfo:   hw.CoverImageInfo,
				CreatedAt:        hw.CreatedAt,
				UpdatedAt:        hw.UpdatedAt,
				CreatedBy:        hw.CreatedBy,
				UpdatedBy:        hw.UpdatedBy,
				DeletedBy:        hw.DeletedBy,
				TotalQuestions:   hw.TotalQuestions,
				CloneInfo:        hw.CloneInfo,
				ObjectTitle:      hw.ObjectTitle,
				ProgramId:        hw.ProgramId,
				IsRandomQuestion: hw.IsRandomQuestion,
			}

			// Nếu exercise với ID này đã tồn tại thì bỏ qua
			var exists int64
			if err := db.ReplicaDB.Model(&models.Exercise{}).Where("id = ?", hw.ID).Count(&exists).Error; err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			if exists > 0 {
				// Đã tồn tại exercise với ID này -> bỏ qua
				res.Skipped++
				continue
			}

			// Copy giữ nguyên ID của homework sang exercise
			ex.ID = hw.ID

			if err := db.MasterDB.Create(&ex).Error; err != nil {
				res.Failed++
				res.FailedIDs = append(res.FailedIDs, hw.ID)
				continue
			}

			// Copy cloned questions mapping homework -> exercise (optional)
			var cloned models.ClonedQuestion
			if err := db.ReplicaDB.Where("assignment_type = ? AND assignment_id = ?", models.ClonedQuestionTypeHomework, hw.ID).First(&cloned).Error; err == nil && cloned.ID != 0 {
				newCloned := models.ClonedQuestion{
					AssignmentID:   ex.ID,
					AssignmentType: models.ClonedQuestionTypeExercise,
					Questions:      cloned.Questions,
					CreatedAt:      cloned.CreatedAt,
					UpdatedAt:      cloned.UpdatedAt,
					CreatedBy:      cloned.CreatedBy,
					UpdatedBy:      cloned.UpdatedBy,
					DeletedBy:      cloned.DeletedBy,
					CloneInfo:      cloned.CloneInfo,
					ProgramId:      cloned.ProgramId,
				}
				if err := db.MasterDB.Create(&newCloned).Error; err != nil {
					// Treat as failure for this homework
					res.Failed++
					res.FailedIDs = append(res.FailedIDs, hw.ID)
					continue
				}
				res.ClonedCopied++
			}

			res.Success++
		}

		offset += chunkSize
	}

	ctx.JSON(200, gin.H{
		"total":               res.Total,
		"success":             res.Success,
		"failed":              res.Failed,
		"skipped":             res.Skipped,
		"cloned_copied":       res.ClonedCopied,
		"failed_homework_ids": res.FailedIDs,
	})
}

