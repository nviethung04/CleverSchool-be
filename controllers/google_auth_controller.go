package controllers

import (
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type GoogleAuthController struct {
	service services.GoogleAuthService
}

func NewGoogleAuthController() *GoogleAuthController {
	return &GoogleAuthController{
		service: services.NewGoogleAuthService(),
	}
}

// Connect initiates Google OAuth flow
func (ctrl *GoogleAuthController) Connect(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)
	host := c.Request.Host

	authURL, err := ctrl.service.GetAuthURL(uint(userID), host)
	if err != nil {
		utils.Respond(c, nil, err, "Failed to generate authorization URL")
		return
	}

	utils.Respond(c, gin.H{
		"auth_url": authURL,
	}, nil, "")
}

// Callback handles OAuth callback
func (ctrl *GoogleAuthController) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing code or state parameter",
		})
		return
	}

	_, err := ctrl.service.HandleCallback(code, state)
	if err != nil {
		html := "<html><head><meta charset=\"utf-8\"/></head><body style=\"font-family:Arial; padding:24px\"><h3>Không thể đăng nhập Google</h3><p>Vui lòng thử lại.</p><pre style=\"white-space:pre-wrap;\">" + strings.ReplaceAll(err.Error(), "<", "&lt;") + "</pre></body></html>"
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(html))
		return
	}

	html := "<html><head><meta charset=\"utf-8\"/></head><body style=\"font-family:Arial; padding:24px\"><h3>Đã đăng nhập Google thành công</h3><p>Bạn có thể đóng cửa sổ này và quay lại hệ thống.</p><button onclick=\"window.close()\">Đóng</button><script>setTimeout(function(){try{window.close()}catch(e){}}, 1500);</script></body></html>"
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// Status checks Google connection status
func (ctrl *GoogleAuthController) Status(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	isConnected := ctrl.service.IsConnected(uint(userID))

	utils.Respond(c, gin.H{
		"is_connected": isConnected,
		"user_id":      userID,
	}, nil, "")
}

// Disconnect removes Google connection
func (ctrl *GoogleAuthController) Disconnect(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	err := ctrl.service.Disconnect(uint(userID))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to disconnect Google account")
		return
	}

	utils.Respond(c, gin.H{
		"message": "Google account disconnected successfully",
	}, nil, "")
}

// RefreshToken refreshes the access token
func (ctrl *GoogleAuthController) RefreshToken(c *gin.Context) {
	userID := utils.GetCurrentUserId(c)

	err := ctrl.service.RefreshToken(uint(userID))
	if err != nil {
		utils.Respond(c, nil, err, "Failed to refresh token")
		return
	}

	utils.Respond(c, gin.H{
		"message": "Token refreshed successfully",
	}, nil, "")
}

