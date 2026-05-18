package config

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
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
	AppConfigUrl      string
	H5PUrl            string

	// API Configuration
	APIDomain string

	MasterPassword string

	DiscordHookUrl string

	// Internal API Configuration
	InternalAPIKey string

	PublicCourseIds []int64

	IsVtg bool
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Warning: No .env file found")
	}

	dbName := os.Getenv("DB_MASTER_NAME")

	dbMasterURL := "postgres://" + os.Getenv("DB_MASTER_USER") + ":" + os.Getenv("DB_MASTER_PASSWORD") + "@" + os.Getenv("DB_MASTER_HOST") + ":" + os.Getenv("DB_MASTER_PORT") + "/" + os.Getenv("DB_MASTER_NAME") + "?sslmode=disable"
	dbReplicaURL := ""
	if os.Getenv("DB_REPLICA_HOST") != "" {
		dbReplicaURL = "postgres://" + os.Getenv("DB_REPLICA_USER") + ":" + os.Getenv("DB_REPLICA_PASSWORD") + "@" + os.Getenv("DB_REPLICA_HOST") + ":" + os.Getenv("DB_REPLICA_PORT") + "/" + os.Getenv("DB_REPLICA_NAME") + "?sslmode=disable"
	}

	allowOrigins := []string{}
	if allowOriginStr := os.Getenv("ALLOW_ORIGINS"); allowOriginStr != "" {
		for _, origin := range strings.Split(allowOriginStr, ",") {
			allowOrigins = append(allowOrigins, strings.TrimSpace(origin))
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
		apiDomain = "http://localhost:8080"
	}

	masterPassword := os.Getenv("MASTER_PASSWORD")
	if masterPassword == "" {
		masterPassword = "super@Clever School@80a"
	}

	discordWebhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")

	if discordWebhookUrl == "" {
		discordWebhookUrl = "https://discord.com/api/webhooks/1422069034106753065/suR91UiySSw6Whtit4fFsAnXqJeWiZWkRJw7iFQ73hhrBNYvSJukGWxw8Vrr6Iq1x7_S"
	}

	internalAPIKey := os.Getenv("INTERNAL_API_KEY")
	if internalAPIKey == "" {
		internalAPIKey = "internal-api-key-2025"
	}

	publicCourseIdsStr := os.Getenv("PUBLIC_COURSE_IDS")
	var publicCourseIds []int64

	if publicCourseIdsStr != "" {
		ids := strings.Split(publicCourseIdsStr, ",")

		for _, id := range ids {
			i, err := strconv.ParseInt(id, 10, 64)
			if err == nil {
				publicCourseIds = append(publicCourseIds, i)
			}
		}
	}

	return Config{
		Port:         os.Getenv("PORT"),
		DBMasterURL:  dbMasterURL,
		DBReplicaURL: dbReplicaURL,
		JWTSecret:    os.Getenv("JWT_SECRET"),

		RedisEnabled:  os.Getenv("REDIS_ENABLED") == "true",
		RedisHost:     os.Getenv("REDIS_HOST"),
		RedisPort:     os.Getenv("REDIS_PORT"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		MeiliEnabled: os.Getenv("MEILI_ENABLED") == "true",
		MeiliHost:    os.Getenv("MEILI_HOST"),
		MeiliAPIKey:  os.Getenv("MEILI_API_KEY"),

		AppDebug:          os.Getenv("APP_DEBUG") == "true",
		EnableSwagger:     os.Getenv("ENABLE_SWAGGER") == "true",
		AllowOrigins:      allowOrigins,
		BasicAuthUsername: basicAuthUsername,
		BasicAuthPassword: basicAuthPassword,
		AppUrl:            os.Getenv("APP_FULL_URL"),
		AppConfigUrl:      os.Getenv("APP_CONFIG_URL"),

		DBName:    dbName,
		APIDomain: apiDomain,
		H5PUrl:    h5pUrl,

		MasterPassword: masterPassword,
		DiscordHookUrl: discordWebhookUrl,
		InternalAPIKey: internalAPIKey,

		PublicCourseIds: publicCourseIds,
		IsVtg:           os.Getenv("IS_VTG") == "true",
	}
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
