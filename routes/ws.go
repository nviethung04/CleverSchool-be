package routes

import (
	"be-Clever School/i18n"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"be-Clever School/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func RegisterWebsocketRoutes(router *gin.Engine) {
	api := router.Group("/api")
	authRepo := repositories.NewAuthRepository()

	ws := api.Group("/ws")

	pubsubService := services.NewChatPubSubService()

	ws.GET("/notifications", func(c *gin.Context) {
		fmt.Printf("[WS] /notifications incoming from %s\n", c.ClientIP())
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("[WS] upgrade error: %v\n", err)
			return
		}

		tokenStr := extractToken(c)
		if tokenStr == "" {
			if _, message, err := conn.ReadMessage(); err == nil {
				var payload map[string]interface{}
				if jsonErr := json.Unmarshal(message, &payload); jsonErr == nil {
					if t, ok := payload["token"].(string); ok {
						tokenStr = t
					}
				}
			}
		}

		claims, err := authenticateToken(c, authRepo, tokenStr)
		if err != nil {
			fmt.Printf("[WS] auth failed: %v\n", err)
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, err.Error()))
			conn.Close()
			return
		}

		fmt.Printf("[WS] authenticated userID=%d\n", claims.UserID)

		msgs, cancel, err := pubsubService.SubscribeToUser(uint64(claims.UserID))
		if err != nil {
			fmt.Printf("[WS] subscribe error: %v\n", err)
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
			conn.Close()
			return
		}

		go func() {
			defer func() {
				cancel()
				conn.Close()
			}()
			for msg := range msgs {
				if msg == nil {
					continue
				}
				if err := conn.WriteJSON(msg); err != nil {
					fmt.Printf("[WS] write error userID=%d: %v\n", claims.UserID, err)
					return
				}
			}
		}()

		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	})

	ws.GET("/notifications/debug/:userID", func(c *gin.Context) {
		idStr := c.Param("userID")
		id, _ := strconv.ParseUint(idStr, 10, 64)
		if id == 0 {
			c.Status(http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		msgs, cancel, err := pubsubService.SubscribeToUser(id)
		if err != nil {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, err.Error()))
			conn.Close()
			return
		}

		go func() {
			defer func() {
				cancel()
				conn.Close()
			}()
			for msg := range msgs {
				if msg == nil {
					continue
				}
				conn.WriteJSON(msg)
			}
		}()

		for {
			if _, _, err := conn.NextReader(); err != nil {
				break
			}
		}
	})
}

func extractToken(c *gin.Context) string {
	tokenStr := c.GetHeader("Token")

	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = strings.TrimSpace(authHeader[7:])
		}
	}

	if tokenStr == "" {
		if cookieToken, err := c.Cookie("token"); err == nil && cookieToken != "" {
			tokenStr = cookieToken
		}
	}

	if tokenStr == "" {
		if q := c.Query("token"); q != "" {
			tokenStr = q
		}
	}

	return tokenStr
}

func authenticateToken(c *gin.Context, authRepo repositories.AuthRepository, tokenStr string) (*utils.CustomClaims, error) {
	if tokenStr == "" {
		return nil, errors.New(i18n.Localize("messages.missing_token"))
	}

	claims, err := utils.CheckToken(c, authRepo, tokenStr)
	if err != nil {
		return nil, errors.New(i18n.Localize("messages.invalid_token"))
	}

	return claims, nil
}
