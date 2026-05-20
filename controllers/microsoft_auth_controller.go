package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type MicrosoftAuthController struct {
	service services.MicrosoftAuthService
}

func NewMicrosoftAuthController() *MicrosoftAuthController {
	return &MicrosoftAuthController{
		service: services.NewMicrosoftAuthService(repositories.NewMicrosoftAuthRepository()),
	}
}

// @Summary Connect Microsoft Account
// @Description Generate Microsoft OAuth URL to connect account
// @Tags Microsoft Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "auth_url"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /microsoft/auth/connect [get]
func (c *MicrosoftAuthController) ConnectAccount(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)

	authURL := c.service.GetAuthURL(uint(userID))

	config.Log.Infof("🔗 Generated Microsoft auth URL for user %d", userID)

	c.ResponseSuccess(ctx, gin.H{
		"auth_url": authURL,
	})
} // @Summary Microsoft OAuth Callback
// @Description Handle Microsoft OAuth callback and save account
// @Tags Microsoft Auth
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} map[string]interface{} "account info"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Router /microsoft/auth/callback [get]
func (c *MicrosoftAuthController) HandleCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")

	if code == "" {
		config.Log.Error("❌ Microsoft OAuth callback: missing code parameter")
		c.ResponseError(ctx, http.StatusBadRequest, "Missing authorization code", nil)
		return
	}

	if state == "" {
		config.Log.Error("❌ Microsoft OAuth callback: missing state parameter")
		c.ResponseError(ctx, http.StatusBadRequest, "Missing state parameter", nil)
		return
	}

	account, err := c.service.ExchangeCodeForToken(ctx, code, state)
	if err != nil {
		config.Log.Errorf("❌ Failed to exchange Microsoft code for token: %v", err)
		c.ResponseError(ctx, http.StatusBadRequest, "Failed to connect Microsoft account", err.Error())
		return
	}

	config.Log.Infof("✅ Microsoft account connected successfully for user %d: %s", account.UserID, account.AccountEmail)
	accept := ctx.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		html := `<!doctype html>
<html>
	<head>
		<meta charset="utf-8" />
		<title>Microsoft Connected</title>
		<meta name="viewport" content="width=device-width,initial-scale=1" />
		<style>body{font-family:Arial,Helvetica,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0} .card{padding:20px;border-radius:8px;border:1px solid #e5e7eb;box-shadow:0 2px 6px rgba(0,0,0,.06);text-align:center}</style>
	</head>
	<body>
		<div class="card">
			<h3>Microsoft account connected</h3>
			<p>Account: ` + account.AccountEmail + `</p>
			<p>Trang này sẽ đóng sau <span id="count">5</span> giây...</p>
			<p>Nếu cửa sổ không đóng, hãy đóng thủ công.</p>
		</div>
		<script>
			(function(){
				var count = 5;
				var el = document.getElementById('count');
				var t = setInterval(function(){
					count -= 1; el.textContent = count;
					if(count <= 0){
						clearInterval(t);
						try{
							if(window.opener){
								// Notify parent window that Microsoft account is connected
								window.opener.postMessage({type: 'MICROSOFT_AUTH_SUCCESS'}, '*');
							}
							window.close();
						} catch(e){
							try{ window.close(); }catch(e){}
						}
					}
				}, 1000);
			})();
		</script>
	</body>
</html>`

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Microsoft account connected successfully",
		"account": gin.H{
			"id":           account.ID,
			"email":        account.AccountEmail,
			"tenant_id":    account.TenantID,
			"connected_at": account.CreatedAt,
		},
	})
}

// @Summary Check Microsoft Connection
// @Description Check if user has connected Microsoft account
// @Tags Microsoft Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "connection status"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /microsoft/auth/status [get]
func (c *MicrosoftAuthController) CheckConnection(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)

	isConnected := c.service.IsConnected(uint(userID))

	c.ResponseSuccess(ctx, gin.H{
		"connected": isConnected,
	})
}

// @Summary Disconnect Microsoft Account
// @Description Disconnect and remove Microsoft account from user
// @Tags Microsoft Auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "success message"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /microsoft/auth/disconnect [delete]
func (c *MicrosoftAuthController) DisconnectAccount(ctx *gin.Context) {
	userID := utils.GetCurrentUserId(ctx)

	if err := c.service.DisconnectAccount(uint(userID)); err != nil {
		config.Log.Errorf("❌ Failed to disconnect Microsoft account for user %d: %v", userID, err)
		c.ResponseError(ctx, http.StatusInternalServerError, "Failed to disconnect Microsoft account", err.Error())
		return
	}

	config.Log.Infof("🔌 Microsoft account disconnected for user %d", userID)

	c.ResponseSuccess(ctx, gin.H{
		"message": "Microsoft account disconnected successfully",
	})
}

// ResponseSuccess sends successful response
func (c *MicrosoftAuthController) ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// ResponseError sends error response
func (c *MicrosoftAuthController) ResponseError(ctx *gin.Context, statusCode int, message string, details interface{}) {
	response := gin.H{
		"success": false,
		"message": message,
	}

	if details != nil {
		response["details"] = details
	}

	ctx.JSON(statusCode, response)
}

