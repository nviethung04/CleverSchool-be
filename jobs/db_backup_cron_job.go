package jobs

import (
	"be-lms/config"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

// StartDatabaseBackupCronJob schedules Postgres SQL dumps at 00:00 and 12:00 daily
func StartDatabaseBackupCronJob() {
	location := Location()
	c := cron.New(cron.WithLocation(location))

	// Run at 00:00 and 12:00 every day
	c.AddFunc("0 0,12 * * *", func() {
		backupDatabase()
	})

	// Debug: run every minute (bỏ comment nếu cần test)
	// c.AddFunc("*/5 * * * *", func() {
	// 	backupDatabase()
	// })

	c.Start()
}

func backupDatabase() {
	cfg := config.LoadConfig()

	// Resolve DB credentials
	pgHost := getFirstNonEmpty(os.Getenv("DB_MASTER_HOST"), "localhost")
	pgPort := getFirstNonEmpty(os.Getenv("DB_MASTER_PORT"), "5432")
	pgUser := getFirstNonEmpty(os.Getenv("DB_MASTER_USER"), "postgres")
	pgPass := os.Getenv("DB_MASTER_PASSWORD")
	pgDBName := getFirstNonEmpty(os.Getenv("DB_MASTER_NAME"), cfg.DBName)
	domain := getFirstNonEmpty(os.Getenv("API_DOMAIN"), "localhost")
	domain = extractDomainName(domain)

	// Prepare output directory and filename
	outputDir := filepath.Join("public", "exports", "db_backups")
	_ = os.MkdirAll(outputDir, 0o755)

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("%s-%s-%s.sql", sanitizeFileName(pgDBName), domain, timestamp)
	filePath := filepath.Join(outputDir, fileName)

	// Check if pg_dump is available
	pgDumpAvailable := false
	if _, err := exec.LookPath("pg_dump"); err == nil {
		pgDumpAvailable = true
	}

	// Check if Docker is available
	dockerAvailable := false
	if _, err := exec.LookPath("docker"); err == nil {
		dockerAvailable = true
	}

	// If neither pg_dump nor Docker available, try direct DB export
	if !pgDumpAvailable && !dockerAvailable {
		config.Log.Info("Neither pg_dump nor Docker available. Trying direct database export...")
		if err := exportDatabaseDirectly(pgHost, pgPort, pgUser, pgPass, pgDBName, filePath); err != nil {
			config.Log.Errorf("Direct database export failed: %v", err)
			_ = os.Remove(filePath)
			return
		}
		uploadAndClean(filePath, fileName)
		return
	}

	// Try to install pg_dump if not available but Docker is
	if !pgDumpAvailable && dockerAvailable {
		config.Log.Info("pg_dump not found. Trying to install...")

		installCmds := []string{
			"apt-get update && apt-get install -y postgresql-client",
			"apt-get update && apt-get install -y postgresql-client-15",
			"apt-get update && apt-get install -y postgresql-client-14",
		}

		installed := false
		for _, cmdStr := range installCmds {
			config.Log.Infof("Trying install with: %s", cmdStr)
			installCmd := exec.Command("bash", "-c", cmdStr)
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if ierr := installCmd.Run(); ierr == nil {
				installed = true
				break
			}
		}

		if !installed {
			config.Log.Info("Failed to install pg_dump. Falling back to Docker pg_dump...")
			if derr := dumpWithDockerPgDump("15", pgHost, pgPort, pgUser, pgPass, pgDBName, filePath); derr != nil {
				config.Log.Errorf("Docker pg_dump failed: %v", derr)
				_ = os.Remove(filePath)
				return
			}
			uploadAndClean(filePath, fileName)
			return
		}
	}

	// Try local pg_dump
	cmd := exec.Command("pg_dump",
		"--format=plain",
		"--no-owner",
		"--no-privileges",
		"-h", pgHost,
		"-p", pgPort,
		"-U", pgUser,
		pgDBName,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+pgPass)

	f, err := os.Create(filePath)
	if err != nil {
		config.Log.Errorf("Cannot create backup file %s: %v", filePath, err)
		return
	}
	defer f.Close()
	cmd.Stdout = f

	var stderr strings.Builder
	cmd.Stderr = &stderr

	config.Log.Infof("Starting database backup to %s", filePath)
	if err := cmd.Run(); err != nil {
		// If version mismatch, retry with Dockerized pg_dump
		if strings.Contains(stderr.String(), "server version:") &&
			strings.Contains(stderr.String(), "pg_dump version:") {
			major := extractServerMajorVersion(stderr.String())
			if major == "" {
				major = "15"
			}
			config.Log.Infof("Local pg_dump failed due to version mismatch. Retrying with Docker postgres:%s-alpine...", major)

			if derr := dumpWithDockerPgDump(major, pgHost, pgPort, pgUser, pgPass, pgDBName, filePath); derr != nil {
				config.Log.Errorf("Docker pg_dump also failed: %v", derr)
				_ = os.Remove(filePath)
				return
			}
		} else {
			config.Log.Errorf("Database backup failed: %v | %s", err, stderr.String())
			_ = os.Remove(filePath)
			return
		}
	}

	uploadAndClean(filePath, fileName)
}

// upload to S3 and clean up
func uploadAndClean(filePath, fileName string) {
	if err := uploadBackupToS3(filePath, fileName); err != nil {
		config.Log.Errorf("Failed to upload backup to S3: %v", err)
		return
	}
	if err := os.Remove(filePath); err != nil {
		config.Log.Warnf("Uploaded to S3 but failed to remove local backup %s: %v", filePath, err)
	}
	config.Log.Infof("Database backup completed and uploaded: %s", fileName)
}

func getFirstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "_")
	return replacer.Replace(name)
}

func extractServerMajorVersion(errMsg string) string {
	parts := strings.Split(errMsg, "server version:")
	if len(parts) < 2 {
		return ""
	}
	ver := strings.TrimSpace(parts[1])
	for i := 0; i < len(ver); i++ {
		if ver[i] == ' ' || ver[i] == '\n' || ver[i] == '\r' || ver[i] == ';' {
			ver = ver[:i]
			break
		}
	}
	if idx := strings.Index(ver, "."); idx > 0 {
		return ver[:idx]
	}
	return ver
}

func dumpWithDockerPgDump(major, host, port, user, password, dbname string, filePath string) error {
	image := fmt.Sprintf("postgres:%s-alpine", major)

	dockerHost := host
	if host == "localhost" || host == "127.0.0.1" {
		dockerHost = "host.docker.internal"
	}

	args := []string{
		"run", "--rm",
		"-e", "PGPASSWORD=" + password,
		image,
		"pg_dump",
		"--format=plain",
		"--no-owner",
		"--no-privileges",
		"-h", dockerHost,
		"-p", port,
		"-U", user,
		dbname,
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create backup file: %v", err)
	}
	defer f.Close()

	cmd := exec.Command("docker", args...)
	cmd.Stdout = f
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker pg_dump error: %v | %s", err, stderr.String())
	}
	return nil
}

// exportDatabaseDirectly exports database using direct SQL queries when pg_dump is not available
// exportDatabaseDirectly exports only data (INSERTs) without schema.
// Schema must be created separately using pg_dump --schema-only
func exportDatabaseDirectly(host, port, user, password, dbname, filePath string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer file.Close()

	file.WriteString("-- Data-only backup created at " + time.Now().Format("2006-01-02 15:04:05") + "\n")
	file.WriteString("-- Schema must be restored separately\n\n")

	// Get all table names
	rows, err := dbConn.Query(`
		SELECT schemaname, tablename
		FROM pg_tables
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY schemaname, tablename
	`)
	if err != nil {
		return fmt.Errorf("failed to get table list: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}
		tables = append(tables, fmt.Sprintf("%s.%s", schema, table))
	}

	// Export each table's data
	for _, table := range tables {
		config.Log.Infof("Exporting table data: %s", table)
		file.WriteString(fmt.Sprintf("\n-- Data for table %s\n", table))

		dataRows, err := dbConn.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			config.Log.Warnf("Failed to get data for %s: %v", table, err)
			file.WriteString(fmt.Sprintf("-- Failed to get data for %s\n", table))
			continue
		}

		columns, err := dataRows.Columns()
		if err != nil {
			config.Log.Warnf("Failed to get columns for %s: %v", table, err)
			dataRows.Close()
			continue
		}

		for dataRows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := dataRows.Scan(valuePtrs...); err != nil {
				config.Log.Warnf("Failed to scan row for %s: %v", table, err)
				continue
			}

			insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (",
				table, strings.Join(columns, ", "))

			valStrs := make([]string, len(values))
			for i, val := range values {
				if val == nil {
					valStrs[i] = "NULL"
				} else {
					switch v := val.(type) {
					case []byte:
						str := string(v) // convert []byte → string
						str = strings.ReplaceAll(str, "'", "''")
						// Fix timestamp lỗi +0000 +0000 hoặc +0000 UTC
						str = strings.ReplaceAll(str, "+0000 +0000", "+00")
						str = strings.ReplaceAll(str, "+0000 UTC", "+00")
						valStrs[i] = "'" + str + "'"
					default:
						str := fmt.Sprintf("%v", v)
						str = strings.ReplaceAll(str, "'", "''")
						str = strings.ReplaceAll(str, "+0000 +0000", "+00")
						str = strings.ReplaceAll(str, "+0000 UTC", "+00")
						valStrs[i] = "'" + str + "'"
					}
				}
			}
			insertSQL += strings.Join(valStrs, ", ") + ");\n"
			file.WriteString(insertSQL)
		}
		dataRows.Close()
	}

	config.Log.Info("Direct data-only export completed")
	return nil
}

func uploadBackupToS3(localPath string, fileName string) error {
	s3Disk := config.Disks[config.S3]

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
		if s3Disk.Endpoint != "" {
			ep := aws.Endpoint{
				URL:               s3Disk.Endpoint,
				SigningRegion:     s3Disk.Region,
				HostnameImmutable: true,
			}
			return ep, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(s3Disk.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Disk.Key, s3Disk.Secret, "")),
		awsconfig.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = s3Disk.PathStyle
	})

	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("cannot read backup file: %v", err)
	}

	key := fmt.Sprintf("public/exports/db_backups/%s", fileName)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Disk.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/sql"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %v", err)
	}

	config.Log.Infof("Uploaded backup to S3: %s/%s", strings.TrimRight(s3Disk.URL, "/"), key)
	return nil
}

func extractDomainName(rawURL string) string {
	u, err := url.Parse(rawURL)
	var host string
	if err != nil || u.Hostname() == "" {
		host = strings.TrimPrefix(rawURL, "http://")
		host = strings.TrimPrefix(host, "https://")
		host = strings.Split(host, "/")[0]
		host = strings.Split(host, ":")[0]
	} else {
		host = u.Hostname()
	}

	// Nếu là localhost hoặc 127.0.0.1 thì giữ nguyên
	if host == "localhost" || host == "127.0.0.1" {
		return host
	}

	if idx := strings.Index(host, "-api"); idx != -1 {
		host = host[:idx]
	} else {
		if dot := strings.Index(host, "."); dot != -1 {
			host = host[:dot]
		}
	}

	return strings.ReplaceAll(host, "-", "_")
}
