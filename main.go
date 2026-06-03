// main.go
package main

import (
	"be-lms/app"
	"be-lms/command"
	"be-lms/config"
	"be-lms/database/db"
	"embed"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/lib/pq"
)

//go:embed database/migrations/*.sql
var migrationFiles embed.FS

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "migrate:version":
			showMigrationVersion()
			return
		case "migrate:force":
			if len(os.Args) < 3 {
				log.Fatal("❌ Thiếu version cần force. Dùng: migrate:force <version>")
			}
			version, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("❌ Version không hợp lệ: %v", err)
			}
			forceMigration(version)
			return
		case "generate-school-statistics":
			command.GenerateSchoolStatisticsCommand()
			return
		case "generate-course-statistics":
			command.GenerateCourseStatisticsCommand()
			return
		case "daily-school-statistics":
			command.DailySchoolStatisticsCommand()
			return
		case "daily-course-statistics":
			command.DailyCourseStatisticsCommand()
			return
		}
	}

	// Nếu không có args hoặc args khác thì chạy app như bình thường
	app.RunAppServer()
}

func runMigrate() {
	fmt.Println("📦 Running DB migration...")

	cfg := config.LoadConfig()
	dbURL := cfg.DBMasterURL
	if dbURL == "" {
		log.Fatal("❌ DBMasterURL is not set")
	}

	dbInstance, err := postgres.WithInstance(db.OpenDB(dbURL), &postgres.Config{})
	if err != nil {
		log.Fatal("❌ DB instance error:", err)
	}

	sourceDriver, err := iofs.New(migrationFiles, "database/migrations")
	if err != nil {
		log.Fatal("❌ Migration source error:", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbInstance)
	if err != nil {
		log.Fatal("❌ Migrate init error:", err)
	}

	// 🔍 Lấy version trước khi migrate
	currentVer, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatal("❌ Get current version failed:", err)
	}
	if err == migrate.ErrNilVersion {
		currentVer = 0
	}

	fmt.Printf("📄 Current migration version: %d (dirty=%v)\n", currentVer, dirty)

	// ⚙️ Thực thi migrate
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("❌ Migration failed:", err)
	}

	// 🔍 Lấy version sau khi migrate
	newVer, dirty, err := m.Version()
	if err != nil {
		log.Fatal("❌ Get new version failed:", err)
	}

	fmt.Printf("✅ Migration successful. New version: %d (dirty=%v)\n", newVer, dirty)

	if newVer == currentVer {
		fmt.Println("ℹ️ No new migrations were applied.")
	} else {
		fmt.Printf("📌 Migrated from version %d → %d\n", currentVer, newVer)
	}
}

func showMigrationVersion() {
	cfg := config.LoadConfig()
	dbURL := cfg.DBMasterURL
	if dbURL == "" {
		log.Fatal("❌ DBMasterURL is not set")
	}

	dbInstance, err := postgres.WithInstance(db.OpenDB(dbURL), &postgres.Config{})
	if err != nil {
		log.Fatal("❌ DB instance error:", err)
	}

	sourceDriver, err := iofs.New(migrationFiles, "database/migrations")
	if err != nil {
		log.Fatal("❌ Migration source error:", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbInstance)
	if err != nil {
		log.Fatal("❌ Migrate init error:", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("❌ Get version error: %v", err)
	}
	fmt.Printf("📄 Current migration version: %d (dirty=%v)\n", version, dirty)
}

func forceMigration(version int) {
	cfg := config.LoadConfig()
	dbURL := cfg.DBMasterURL
	if dbURL == "" {
		log.Fatal("❌ DBMasterURL is not set")
	}

	dbInstance, err := postgres.WithInstance(db.OpenDB(dbURL), &postgres.Config{})
	if err != nil {
		log.Fatal("❌ DB instance error:", err)
	}

	sourceDriver, err := iofs.New(migrationFiles, "database/migrations")
	if err != nil {
		log.Fatal("❌ Migration source error:", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbInstance)
	if err != nil {
		log.Fatal("❌ Migrate init error:", err)
	}

	if err := m.Force(version); err != nil {
		log.Fatalf("❌ Force version failed: %v", err)
	}

	fmt.Printf("✅ Forced version to %d.\n", version)
}
