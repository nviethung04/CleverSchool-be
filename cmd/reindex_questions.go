package main

import (
	"fmt"
	"os"

	"be-lms/command"
)

// MEILI_ENABLED=true MEILI_HOST=http://localhost:7700 MEILI_API_KEY=lms_master go run cmd/reindex_questions.go

func main() {
	if err := command.RunReindexQuestions(); err != nil {
		fmt.Println("Reindex error:", err)
		os.Exit(1)
	}
}
