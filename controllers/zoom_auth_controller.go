package controllers

import (
	"be-Clever School/repositories"
	"be-Clever School/services"
	"be-Clever School/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ZoomAuthController struct {
	zoomAuthService services.ZoomAuthService
}

func NewZoomAuthController() *ZoomAuthController {
	zoomAuthRepo := repositories.NewZoomAuthRepository()
	zoomAuthService := services.NewZoomAuthService(zoomAuthRepo)

	return &ZoomAuthController{
		zoomAuthService: zoomAuthService,
	}
}

// POST /api/zoom/auth/connect
func (zac *ZoomAuthController) Connect(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authURL, err := zac.zoomAuthService.GetAuthURL(int64(userID.(int)))
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"auth_url": authURL,
	}, nil, "")
}

// GET /api/zoom/auth/callback
func (zac *ZoomAuthController) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		utils.Respond(c, nil, nil, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	zoomAccount, err := zac.zoomAuthService.HandleCallback(code, state)
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"message":      "Zoom account connected successfully",
		"zoom_account": zoomAccount,
	}, nil, "")
}

// POST /api/zoom/auth/disconnect
func (zac *ZoomAuthController) Disconnect(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := zac.zoomAuthService.DisconnectAccount(int64(userID.(int)))
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"message": "Zoom account disconnected successfully",
	}, nil, "")
}

// GET /api/zoom/auth/status
func (zac *ZoomAuthController) Status(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isConnected, zoomAccount, err := zac.zoomAuthService.GetConnectionStatus(int64(userID.(int)))
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	response := gin.H{
		"is_connected": isConnected,
	}

	if isConnected && zoomAccount != nil {
		response["zoom_account"] = gin.H{
			"zoom_user_id":  zoomAccount.ZoomUserID,
			"account_email": zoomAccount.AccountEmail,
			"connected_at":  zoomAccount.CreatedAt,
		}
	}

	utils.Respond(c, response, nil, "")
}

// POST /api/zoom/auth/refresh
func (zac *ZoomAuthController) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Respond(c, nil, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := zac.zoomAuthService.RefreshToken(int64(userID.(int)))
	if err != nil {
		utils.Respond(c, nil, err, "", http.StatusInternalServerError)
		return
	}

	utils.Respond(c, gin.H{
		"message": "Token refreshed successfully",
	}, nil, "")
}
