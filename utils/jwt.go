package utils

import (
	"be-lms/redis"
	"be-lms/repositories"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Parse JWT và lấy userID, bỏ qua kiểm tra hạn (expiry)
func GetUserIDIgnoreExpiry(tokenStr string) (int64, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, &CustomClaims{})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	return int64(claims.UserID), nil
}

var jwtSecret []byte

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is not set in environment variables")
	}
	jwtSecret = []byte(secret)
}

type CustomClaims struct {
	UserID  int    `json:"user_id"`
	RoleID  int    `json:"role_id"`
	SchoolID int    `json:"school_id"`
	Role    string `json:"role"`
	RoleIDs  []int    `json:"role_ids"`
	Roles    []string `json:"roles"`
	RoleType string `json:"role_type"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, roleID, schoolID int, roleName string, roleIDs []int, roles []string, roleType string) (string, error) {
	claims := CustomClaims{
		UserID:  userID,
		RoleID:  roleID,
		SchoolID: schoolID,
		Role:    roleName,
		RoleIDs:  roleIDs,
		Roles:    roles,
		RoleType: roleType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // Token hết hạn sau 7 ngày
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func CheckToken(c *gin.Context, authRepo repositories.AuthRepository, tokenStr string) (*CustomClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	isValid, _ := redis.IsTokenValid(claims.UserID, tokenStr)

	if !isValid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func CheckTokenWithoutContext(tokenStr string) (*CustomClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	isValid, _ := redis.IsTokenValid(claims.UserID, tokenStr)

	if !isValid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GetClaims(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	isValid, _ := redis.IsTokenValid(claims.UserID, tokenStr)

	if !isValid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GetUserID(tokenStr string) (int64, error) {
	claims, err := GetClaims(tokenStr)
	if err != nil {
		return 0, err
	}
	return int64(claims.UserID), nil
}
