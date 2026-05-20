package controllers

import (
	"be-cleverschool/config"
	"be-cleverschool/i18n"
	"be-cleverschool/prot"
	"be-cleverschool/redis"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{service}
}

func (ac *AuthController) Login(c *gin.Context) {
	input, err, message := utils.GetBody[*prot.LoginRequest](c, func() *prot.LoginRequest { return &prot.LoginRequest{} })

	if err != nil {
		config.Log.Error("error: ", err)
		utils.Respond(c, nil, err, message)
		return
	}

	user, wrongPassword, err, message := ac.service.Login(c, input.Username, input.Password)

	if err == nil {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("h5p_token", user.Token, 7*24*60*60, "/", "", true, true)

		utils.Respond(c, user, nil, "")
		return
	}

	config.Log.Info("Login: ", input)

	if wrongPassword {
		utils.Respond(c, nil, err, message, 401)
		return
	}

	utils.Respond(c, nil, err, message)
}

func (ac *AuthController) Logout(c *gin.Context) {
	err, message := ac.service.Logout(c)

	utils.Respond(c, &prot.LogoutResponse{
		Success: true,
	}, err, message)
}

func (ac *AuthController) LogoutAll(c *gin.Context) {
	err, message := ac.service.LogoutAll(c)
	utils.Respond(c, nil, err, message)
}

func (ac *AuthController) ListSessions(c *gin.Context) {
	sessions, err, message := ac.service.ListSessions(c)

	if err != nil {
		utils.Respond(c, nil, err, message)
		return
	}

	var sessionPtrs []*redis.SessionInfo
	for i := range sessions {
		sessionPtrs = append(sessionPtrs, &sessions[i])
	}

	sessionResource := resources.NewSessionResource()
	sessionResponse := sessionResource.FormatSessions(sessionPtrs)

	list := &prot.ListSession{Sessions: sessionResponse}

	utils.Respond(c, list, err, message)
}

func (ac *AuthController) Profile(c *gin.Context) {
	user, err, message := ac.service.Profile(c)

	userResource := resources.NewUserResource()
	userResponse := userResource.FormatUser(user)

	utils.Respond(c, userResponse, err, message)
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	id := utils.GetCurrentUserId(c)

	req, err, message := utils.GetBody[*prot.ChangePasswordRequest](c, func() *prot.ChangePasswordRequest {
		return &prot.ChangePasswordRequest{}
	})

	if err != nil {
		config.Log.Error("error: ", err)
		utils.Respond(c, nil, err, message)
		return
	}

	err, message = ac.service.ChangePassword(c, id, req)
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	utils.Respond(c, req, nil, "")
}

func (ac *AuthController) ForgotPassword(c *gin.Context) {
	req, err, _ := utils.GetBody[*prot.ForgotPasswordRequest](c, func() *prot.ForgotPasswordRequest {
		return &prot.ForgotPasswordRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	response, err, message := ac.service.ForgotPassword(c, req)
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	utils.Respond(c, response, nil, "")
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	req, err, _ := utils.GetBody[*prot.ResetPasswordRequest](c, func() *prot.ResetPasswordRequest {
		return &prot.ResetPasswordRequest{}
	})
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	response, err, message := ac.service.ResetPassword(c, req)
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	utils.Respond(c, response, nil, "")
}

func (cc *AuthController) Register(c *gin.Context) {
	req, err, _ := utils.GetBody[*prot.RegisterRequest](c, func() *prot.RegisterRequest {
		return &prot.RegisterRequest{}
	})

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	if req.Email == "" && req.PhoneNumber == "" {
		utils.Respond(c, nil, errors.New(i18n.Localize("messages.email_or_phone_required")), "messages.email_or_phone_required")
		return
	}

	user, err, message := cc.service.Register(c, req)
	if err != nil {
		utils.Respond(c, nil, err, message, http.StatusBadRequest)
		return
	}

	userResource := resources.NewUserResource()
	userFormatted := userResource.FormatUser(user)

	utils.Respond(c, userFormatted, nil, "")
}

func (ac *AuthController) AcceptRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("role"))

	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.id_invalid")), "messages.id_invalid", http.StatusBadRequest)
		return
	}

	user, wrongPassword, err, message := ac.service.AcceptRole(c, id)

	if err == nil {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("h5p_token", user.Token, 7*24*60*60, "/", "", true, true)

		utils.Respond(c, user, nil, "")
		return
	}

	if wrongPassword {
		utils.Respond(c, nil, err, message, 401)
		return
	}

	utils.Respond(c, nil, err, message)
}

