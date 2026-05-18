package main

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/jobs"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables - tìm file .env ở thư mục cha
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env") // fallback nếu chạy từ thư mục gốc

	// Load config
	cfg := config.LoadConfig()
	config.InitLogger()

	log.Println("🔄 Starting manual sync homework status scoring...")

	// Connect to database
	if err := db.ConnectPostgres(cfg); err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}

	// Connect to Redis (optional)
	if err := db.ConnectRedis(cfg); err != nil {
		log.Printf("⚠️ Redis connection failed (continuing without Redis): %v", err)
	}

	// Chạy job sync homework status scoring
	jobs.SyncHomeworkStatusScoringJob()

	log.Println("✅ Manual sync homework status scoring completed!")
}
