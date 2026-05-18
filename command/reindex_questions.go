package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
)

func RunReindexQuestions() error {
	cfg := config.LoadConfig()
	if !cfg.MeiliEnabled {
		fmt.Println("MEILI_ENABLED is false; skip reindex.")
		return nil
	}

	if err := db.ConnectPostgres(cfg); err != nil {
		return fmt.Errorf("connect postgres error: %w", err)
	}

	const batchSize = 1000
	var lastID int64 = 0
	total := 0

	httpClient := &http.Client{Timeout: 60 * time.Second}
	host := strings.TrimSpace(cfg.MeiliHost)
	if host == "" {
		return fmt.Errorf("MEILI_HOST is empty")
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	indexURL := strings.TrimRight(host, "/") + "/indexes/questions/documents"

	for {
		var questions []models.Question
		tx := db.MasterDB.Model(&models.Question{}).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize)
		if err := tx.Find(&questions).Error; err != nil {
			return err
		}
		if len(questions) == 0 {
			break
		}

		docs := make([]map[string]interface{}, 0, len(questions))
		for _, q := range questions {
			lastID = q.ID
			docs = append(docs, map[string]interface{}{
				"id":            q.ID,
				"status":        q.Status,
				"question_type": q.QuestionType,
				"title":         q.Title,
				"description":   q.Description,
				"content":       q.Content,
				"keywords":      q.Keywords,
				"point":         q.Point,
				"created_at":    q.CreatedAt.Format(time.RFC3339),
			})
		}

		body, _ := json.Marshal(docs)
		req, _ := http.NewRequest("POST", indexURL, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if cfg.MeiliAPIKey != "" {
			req.Header.Set("X-Meili-API-Key", cfg.MeiliAPIKey)
			req.Header.Set("Authorization", "Bearer "+cfg.MeiliAPIKey)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("meili index http status %d", resp.StatusCode)
		}
		total += len(docs)
		fmt.Printf("Indexed %d questions (total %d)\n", len(docs), total)
	}

	fmt.Printf("Reindex finished, total %d questions indexed.\n", total)
	return nil
}

func main() {
	// Allow running as: go run command/reindex_questions.go
	_ = os.Setenv("APP_DEBUG", "false")
	if err := RunReindexQuestions(); err != nil {
		fmt.Println("Reindex error:", err)
		os.Exit(1)
	}
}
