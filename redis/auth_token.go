package redis

import (
	"be-Clever School/config"
	"be-Clever School/database/db"
	"be-Clever School/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type SessionInfo struct {
	Id        string    `json:"id"`
	Token     string    `json:"token"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Expires   time.Time `json:"expires"`
	LoginAt   time.Time `json:"login_at"`
	IsCurrent bool      `json:"is_current"`
}

func SaveUserToken(c *gin.Context, userID int, token, ip, userAgent string, expires time.Duration) error {

	ctx := context.Background()
	key := fmt.Sprintf("user_sessions:%d", userID)
	newSessionID := GenerateSessionId(ip, userAgent)
	repo := NewUserSessionRepository()

	// Chỉ xóa toàn bộ token cũ nếu là học sinh (roleID == 3)
	roleID := c.GetInt("roleID")
	if roleID == 3 {
		_ = db.RedisClient.Del(ctx, key).Err()
		_ = repo.DeleteByUserID(userID)
	}

	sessionInfo := SessionInfo{
		Id:        newSessionID,
		Token:     token,
		IP:        ip,
		UserAgent: userAgent,
		Expires:   time.Now().Add(expires),
		LoginAt:   time.Now(),
		IsCurrent: true,
	}

	// Async save session to DB
	go func(s SessionInfo, uID int) {
		if err := SaveSessionToDB(uID, s); err != nil {
			config.Log.Error("failed to save session to DB: %v\n", err)
		}
	}(sessionInfo, userID)

	jsonData, err := json.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	if err := db.RedisClient.HSet(ctx, key, token, jsonData).Err(); err != nil {
		return err
	}

	_ = db.RedisClient.Expire(ctx, key, expires).Err()

	return nil
}

func IsTokenValid(userID int, token string) (bool, error) {
	key := fmt.Sprintf("user_sessions:%d", userID)

	ok, err := db.RedisClient.HExists(context.Background(), key, token).Result()
	if err == nil && ok {
		return ok, nil
	}

	config.Log.Infof("Redis error: %s. Falling back to DB...\n", err)

	repo := NewUserSessionRepository()

	session, err := repo.GetByToken(token)
	if err != nil {
		return false, err
	}

	if session != nil && session.Token == token {
		return true, nil
	}

	return false, nil
}

func ListTokens(userID int, currentToken string) ([]SessionInfo, error) {
	key := fmt.Sprintf("user_sessions:%d", userID)

	rawData, err := db.RedisClient.HGetAll(context.Background(), key).Result()
	var sessions []SessionInfo

	if err == nil && 1 == 2 {
		for _, jsonStr := range rawData {
			var info SessionInfo
			if err := json.Unmarshal([]byte(jsonStr), &info); err == nil {
				info.IsCurrent = info.Token == currentToken
				sessions = append(sessions, info)
			}
		}
	} else {
		if err != redis.Nil {
			config.Log.Infof("Redis error: %s. Falling back to DB...\n", err)
		}

		repo := NewUserSessionRepository()

		dbSessions, dbErr := repo.ListByUserID(userID)
		if dbErr != nil {
			return nil, fmt.Errorf("failed to fetch sessions from DB: %w", dbErr)
		}

		for _, s := range dbSessions {
			sessions = append(sessions, SessionInfo{
				Id:        s.ID,
				Token:     s.Token,
				IP:        s.IP,
				UserAgent: s.UserAgent,
				Expires:   s.Expires,
				LoginAt:   s.LoginAt,
				IsCurrent: s.Token == currentToken,
			})
		}
	}

	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].IsCurrent != sessions[j].IsCurrent {
			return sessions[i].IsCurrent
		}
		return sessions[i].LoginAt.After(sessions[j].LoginAt)
	})

	return sessions, nil
}

func ClearUserToken(userID int, token string, ip string, userAgent string) error {
	key := fmt.Sprintf("user_sessions:%d", userID)
	if err := db.RedisClient.HDel(context.Background(), key, token).Err(); err != nil {
		config.Log.Info("Redis delete failed (continuing with DB): %v\n", err)
	}

	repo := NewUserSessionRepository()
	sessionID := GenerateSessionId(ip, userAgent)

	if err := repo.DeleteByID(sessionID); err != nil {
		return fmt.Errorf("failed to delete session in DB: %w", err)
	}

	return nil
}

func ClearUserTokenId(userID int, tokenID string) error {
	key := fmt.Sprintf("user_sessions:%d", userID)

	rawData, err := db.RedisClient.HGetAll(context.Background(), key).Result()
	if err != nil {
		return fmt.Errorf("failed to fetch sessions: %w", err)
	}

	for token, jsonStr := range rawData {
		var info SessionInfo
		if err := json.Unmarshal([]byte(jsonStr), &info); err != nil {
			continue
		}

		if info.Id == tokenID {
			return db.RedisClient.HDel(context.Background(), key, token).Err()
		}
	}

	repo := NewUserSessionRepository()
	repo.DeleteByToken(tokenID)

	return fmt.Errorf("token ID not found")
}

func ClearAllUserTokens(userID int) error {
	key := fmt.Sprintf("user_sessions:%d", userID)
	if err := db.RedisClient.Del(context.Background(), key).Err(); err != nil {
		config.Log.Info("Redis delete failed (continuing with DB): %v\n", err)
	}

	repo := NewUserSessionRepository()

	if err := repo.DeleteByUserID(userID); err != nil {
		return fmt.Errorf("failed to delete session in DB: %w", err)
	}

	return nil
}

func GetSessionID(c *gin.Context) string {
	cfg := config.LoadConfig()
	sessionKey := strings.Replace(cfg.DBName, "_db_", "_", 1) + "_session"

	cookie, err := c.Request.Cookie(sessionKey)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func GenerateSessionId(ip, userAgent string) string {
	raw := fmt.Sprintf("%s|%s", ip, userAgent)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func SaveSessionToDB(userID int, session SessionInfo) error {
	userSession := models.UserSession{
		ID:        session.Id,
		UserID:    userID,
		Token:     session.Token,
		IP:        session.IP,
		UserAgent: session.UserAgent,
		Expires:   session.Expires,
		LoginAt:   session.LoginAt,
		IsCurrent: session.IsCurrent,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo := NewUserSessionRepository()

	return repo.Save(userSession)
}

// Lấy token mới nhất của user (token đăng nhập cuối cùng)
func GetLatestUserToken(userID int) (string, error) {
	key := fmt.Sprintf("user_sessions:%d", userID)
	rawData, err := db.RedisClient.HGetAll(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	var latestToken string
	var latestLogin time.Time
	for token, jsonStr := range rawData {
		var info SessionInfo
		if err := json.Unmarshal([]byte(jsonStr), &info); err == nil {
			if info.LoginAt.After(latestLogin) {
				latestLogin = info.LoginAt
				latestToken = token
			}
		}
	}
	return latestToken, nil
}
