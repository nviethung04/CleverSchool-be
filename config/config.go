package config

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBMasterURL  string
	DBReplicaURL string
	DBName       string
	JWTSecret    string

	RedisEnabled  bool
	RedisURL      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// MeiliSearch
	MeiliEnabled bool
	MeiliHost    string
	MeiliAPIKey  string

	AppDebug          bool
	AllowOrigins      []string
	EnableSwagger     bool
	BasicAuthUsername string
	BasicAuthPassword string
	AppUrl            string
	H5PUrl            string

	// API Configuration
	APIDomain string

	MasterPassword string

	DiscordHookUrl string

	// Internal API Configuration
	InternalAPIKey string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: No .env file found")
	}

	dbMasterURL := firstNonEmpty(os.Getenv("DATABASE_URL"), buildPostgresURL(
		os.Getenv("DB_MASTER_USER"),
		os.Getenv("DB_MASTER_PASSWORD"),
		os.Getenv("DB_MASTER_HOST"),
		os.Getenv("DB_MASTER_PORT"),
		os.Getenv("DB_MASTER_NAME"),
	))
	dbReplicaURL := firstNonEmpty(os.Getenv("DATABASE_REPLICA_URL"), "")
	if dbReplicaURL == "" && os.Getenv("DB_REPLICA_HOST") != "" {
		dbReplicaURL = buildPostgresURL(
			os.Getenv("DB_REPLICA_USER"),
			os.Getenv("DB_REPLICA_PASSWORD"),
			os.Getenv("DB_REPLICA_HOST"),
			os.Getenv("DB_REPLICA_PORT"),
			os.Getenv("DB_REPLICA_NAME"),
		)
	}

	dbName := firstNonEmpty(os.Getenv("DB_MASTER_NAME"), os.Getenv("PGDATABASE"), dbNameFromURL(dbMasterURL))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisURL := os.Getenv("REDIS_URL")
	redisEnabled := os.Getenv("REDIS_ENABLED") == "true" || redisURL != ""
	if os.Getenv("REDIS_ENABLED") == "false" {
		redisEnabled = false
	}

	allowOrigins := []string{}
	if allowOriginStr := os.Getenv("ALLOW_ORIGINS"); allowOriginStr != "" {
		for _, origin := range strings.Split(allowOriginStr, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowOrigins = append(allowOrigins, origin)
			}
		}
	}
	if len(allowOrigins) == 0 {
		allowOrigins = []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
	}

	basicAuthUsername := os.Getenv("BASIC_AUTH_USERNAME")
	if basicAuthUsername == "" {
		basicAuthUsername = "supperadmin"
	}

	basicAuthPassword := os.Getenv("BASIC_AUTH_PASSWORD")
	if basicAuthPassword == "" {
		basicAuthPassword = hashPassword("supperadmin")
	}

	h5pUrl := os.Getenv("H5P_URL")
	if h5pUrl == "" {
		h5pUrl = "http://10.10.2.135:8080"
	}

	apiDomain := os.Getenv("API_DOMAIN")
	if apiDomain == "" {
		if railwayDomain := os.Getenv("RAILWAY_PUBLIC_DOMAIN"); railwayDomain != "" {
			apiDomain = "https://" + railwayDomain
		} else {
			apiDomain = "http://localhost:8080"
		}
	}

	masterPassword := os.Getenv("MASTER_PASSWORD")
	if masterPassword == "" {
		masterPassword = "super@enspire@80a"
	}

	discordWebhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")

	if discordWebhookUrl == "" {
		discordWebhookUrl = "https://discord.com/api/webhooks/1422069034106753065/suR91UiySSw6Whtit4fFsAnXqJeWiZWkRJw7iFQ73hhrBNYvSJukGWxw8Vrr6Iq1x7_S"
	}

	internalAPIKey := os.Getenv("INTERNAL_API_KEY")
	if internalAPIKey == "" {
		internalAPIKey = "internal-api-key-2025"
	}

	appURL := os.Getenv("APP_FULL_URL")
	if appURL == "" {
		appURL = apiDomain
	}

	return Config{
		Port:         port,
		DBMasterURL:  dbMasterURL,
		DBReplicaURL: dbReplicaURL,
		JWTSecret:    os.Getenv("JWT_SECRET"),

		RedisEnabled:  redisEnabled,
		RedisURL:      redisURL,
		RedisHost:     firstNonEmpty(os.Getenv("REDIS_HOST"), os.Getenv("REDISHOST")),
		RedisPort:     firstNonEmpty(os.Getenv("REDIS_PORT"), os.Getenv("REDISPORT")),
		RedisPassword: firstNonEmpty(os.Getenv("REDIS_PASSWORD"), os.Getenv("REDISPASSWORD")),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		MeiliEnabled: os.Getenv("MEILI_ENABLED") == "true",
		MeiliHost:    os.Getenv("MEILI_HOST"),
		MeiliAPIKey:  os.Getenv("MEILI_API_KEY"),

		AppDebug:          os.Getenv("APP_DEBUG") == "true",
		EnableSwagger:     os.Getenv("ENABLE_SWAGGER") == "true",
		AllowOrigins:      allowOrigins,
		BasicAuthUsername: basicAuthUsername,
		BasicAuthPassword: basicAuthPassword,
		AppUrl:            appURL,

		DBName:    dbName,
		APIDomain: apiDomain,
		H5PUrl:    h5pUrl,

		MasterPassword: masterPassword,
		DiscordHookUrl: discordWebhookUrl,
		InternalAPIKey: internalAPIKey,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func buildPostgresURL(user, pass, host, port, name string) string {
	if host == "" || name == "" {
		return ""
	}
	if port == "" {
		port = "5432"
	}
	return "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + name + "?sslmode=disable"
}

func dbNameFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.Trim(parsed.Path, "/")
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error parsing %s, using default value: %d", key, defaultValue)
		return defaultValue
	}
	return intValue
}

func hashPassword(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}
