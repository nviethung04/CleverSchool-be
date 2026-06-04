package config

import (
	"os"
	"strings"
)

// EnvBool reads a boolean env var (true, 1, yes).
func EnvBool(key string, defaultValue bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// S3Configured is true when AWS credentials and bucket are set (upload, log export, cleanup).
func S3Configured() bool {
	return os.Getenv("AWS_BUCKET") != "" &&
		os.Getenv("AWS_ACCESS_KEY_ID") != "" &&
		os.Getenv("AWS_SECRET_ACCESS_KEY") != ""
}

// DashboardJobsEnabled turns on background dashboard pre-cache and daily statistics jobs.
// Off by default — not required for login, courses, lessons, exams, or homework (MVP).
func DashboardJobsEnabled() bool {
	return EnvBool("ENABLE_DASHBOARD_JOBS", false)
}
