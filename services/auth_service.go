package services

import (
	"be-lms/config"
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/redis"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/hex"
	"errors"
	"math/rand"
	"sort"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(c *gin.Context, username, password string) (*prot.UserLoginResponse, bool, error, string)
	Logout(c *gin.Context) (error, string)
	LogoutAll(c *gin.Context) (error, string)
	ListSessions(c *gin.Context) ([]redis.SessionInfo, error, string)
	Register(c *gin.Context, req *prot.RegisterRequest) (*models.User, error, string)
	Profile(c *gin.Context) (*models.User, error, string)
	ChangePassword(c *gin.Context, id int, req *prot.ChangePasswordRequest) (error, string)
	ForgotPassword(c *gin.Context, req *prot.ForgotPasswordRequest) (*prot.ForgotPasswordResponse, error, string)
	ResetPassword(c *gin.Context, req *prot.ResetPasswordRequest) (*prot.ResetPasswordResponse, error, string)
	AcceptRole(c *gin.Context, id int) (*prot.UserLoginResponse, bool, error, string)
}

type authService struct {
	repo              repositories.AuthRepository
	passwordResetRepo repositories.PasswordResetRepository
	emailService      EmailService
}

func NewAuthService(repo repositories.AuthRepository) AuthService {
	return &authService{
		repo:              repo,
		passwordResetRepo: repositories.NewPasswordResetRepository(),
		emailService:      NewEmailService(),
	}
}

func (s *authService) Login(c *gin.Context, username, password string) (*prot.UserLoginResponse, bool, error, string) {
	// Cho phép đăng nhập bằng email hoặc username, trả về kiểu *models.User
	var user *models.User
	var err error
	var isTester bool

	if username == "" || password == "" {
		return nil, false, errors.New(i18n.Localize("messages.username_password_required")), "messages.username_password_required"
	}

	isTester = password == config.LoadConfig().MasterPassword

	users, err := s.repo.GetUsersByKey(username)
	if err != nil {
		return nil, false, err, "messages.user_not_found"
	}

	if len(users) == 0 {
		return nil, false, errors.New(i18n.Localize("messages.user_not_found")), "messages.user_not_found"
	}

	var roleId int

	for _, u := range users {
		hasUser := false

		if len(u.Roles) == 0 {
			continue
		}

		sort.Slice(u.Roles, func(i, j int) bool {
			return u.Roles[i].ID < u.Roles[j].ID
		})

		roleId = int(u.Roles[0].ID)

		if (!isTester && (roleId == models.StudentRoleId || roleId == models.TeacherRoleId) || roleId == models.AdminRoleId) {
			err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if err == nil {
				hasUser = true
				user = &u
				break
			}
		} else {
			hasUser = true
			user = &u
			break
		}

		if !hasUser {
			return nil, true, errors.New(i18n.Localize("messages.invalid_password")), "messages.invalid_password"
		}
	}

	var roleIds []int
	var roleNames []string
	var roleIdsInt32 []int32

	tempRoleId := roleId
	roleName := models.PageStudent
	roleType := models.PageStudent

	role, err := s.repo.GetRoleById(roleId)
	if err == nil {
		roleName = role.DefaultPageView
		roleType = role.DefaultPageView
		tempRoleId = int(role.DefaultPageID)
	}

	sort.Slice(user.Roles, func(i, j int) bool {
		return user.Roles[i].ID < user.Roles[j].ID
	})

	for _, role := range user.Roles {
		roleIds = append(roleIds, int(role.ID))
		roleIdsInt32 = append(roleIdsInt32, int32(role.ID))
		roleNames = append(roleNames, role.Name)
	}

	// Tạo JWT token
	token, err := utils.GenerateToken(int(user.ID), tempRoleId, user.SchoolID, roleName, roleIds, roleNames, roleType)
	if err != nil {
		return nil, false, errors.New(i18n.Localize("messages.failed_to_generate_token")), "messages.failed_to_generate_token"
	}

	// Chỉ xóa toàn bộ token cũ nếu là học sinh (roleID == 3)
	if tempRoleId == models.StudentRoleId && !isTester {
		_ = redis.ClearAllUserTokens(int(user.ID))
	}

	// Lưu token vào DB
	expires := 7 * 24 * time.Hour // 7 ngày hết hạn
	expiresTime := time.Now().Add(expires)
	if err := s.repo.SaveToken(int(user.ID), token, expiresTime); err != nil {
		return nil, false, errors.New(i18n.Localize("messages.auth_failed")), "messages.auth_failed"
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	_ = redis.SaveUserToken(c, int(user.ID), token, ip, userAgent, expires)

	resp := &prot.UserLoginResponse{
		Token:    token,
		RoleId:   int32(tempRoleId),
		RoleName: roleName,
		Type:     roleType,
		RoleIds:  roleIdsInt32,
		RoleNames: roleNames,
	}

	session := sessions.Default(c)
	session.Set("token", token)
	session.Save()

	return resp, false, nil, ""
}

func (s *authService) Logout(c *gin.Context) (error, string) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return errors.New(i18n.Localize("messages.user_not_authenticated")), "messages.user_not_authenticated"
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return errors.New(i18n.Localize("messages.invalid_user_id")), "messages.invalid_user_id"
	}

	tokenId := c.DefaultQuery("token_id", "")

	if tokenId != "" {
		if err := redis.ClearUserTokenId(userID, tokenId); err != nil {
			return err, "messages.error_clear_user_token_id"
		}
		return nil, ""
	}

	token := c.GetHeader("Token")
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := redis.ClearUserToken(userID, token, ip, userAgent); err != nil {
		return err, "messages.error_clear_user_token"
	}

	session := sessions.Default(c)
	session.Delete("token")
	session.Save()

	return nil, ""
}

func (s *authService) LogoutAll(c *gin.Context) (error, string) {
	token := c.GetHeader("Token")
	if token == "" {
		return errors.New(i18n.Localize("messages.missing_token")), "messages.missing_token"
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		return errors.New(i18n.Localize("messages.user_not_authenticated")), "messages.user_not_authenticated"
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return errors.New(i18n.Localize("messages.invalid_user_id")), "messages.invalid_user_id"
	}

	if err := redis.ClearAllUserTokens(userID); err != nil {
		return err, "messages.error_clear_all_user_tokens"
	}

	session := sessions.Default(c)
	session.Delete("token")
	session.Save()

	return nil, ""
}

func (s *authService) ListSessions(c *gin.Context) ([]redis.SessionInfo, error, string) {

	userIDVal, exists := c.Get("userID")
	if !exists {
		return nil, errors.New(i18n.Localize("messages.user_not_authenticated")), "messages.user_not_authenticated"
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return nil, errors.New(i18n.Localize("messages.invalid_user_id")), "messages.invalid_user_id"
	}

	token := c.GetHeader("Token")
	sessions, err := redis.ListTokens(userID, token)
	if err != nil {
		return nil, errors.New(i18n.Localize("messages.failed_to_list_sessions")), "messages.failed_to_list_sessions"
	}

	return sessions, nil, ""
}

func (s *authService) Register(c *gin.Context, req *prot.RegisterRequest) (*models.User, error, string) {
	var userName string

	if req.Email != "" {
		userName = req.Email
	} else if req.PhoneNumber != "" {
		userName = req.PhoneNumber
	}

	if err := s.repo.CheckExistingUser(userName, req.Email); err != nil {
		return nil, err, "messages.account_already_exists"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userName), bcrypt.DefaultCost)
	if err != nil {
		return nil, err, "messages.error_generate_from_password"
	}

	roleID := req.RoleId
	if roleID == 0 {
		roleID = models.StudentRoleId
	}

	user := models.User{
		Name:        req.FullName,
		Username:    userName,
		Email:       req.Email,
		Password:    string(hashedPassword),
		PhoneNumber: req.PhoneNumber,
		Status:      false,
	}

	return s.repo.CreateUser(user, int64(roleID))
}

func (s *authService) Profile(c *gin.Context) (*models.User, error, string) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return nil, errors.New(i18n.Localize("messages.user_not_authenticated")), "messages.user_not_authenticated"
	}

	userID, ok := userIDVal.(int)
	if !ok {
		return nil, errors.New(i18n.Localize("messages.invalid_user_id")), "messages.invalid_user_id"
	}

	return s.repo.GetByID(userID)
}

func (s *authService) ChangePassword(c *gin.Context, id int, req *prot.ChangePasswordRequest) (error, string) {
	user, err, _ := s.repo.GetByID(id)
	if err != nil {
		config.Log.Error(err)
		return errors.New(i18n.Localize("messages.user_not_found")), "messages.user_not_found"
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		config.Log.Error(err)
		return errors.New(i18n.Localize("messages.invalid_old_password")), "messages.invalid_old_password"
	}

	if req.NewPassword != req.ConfirmPassword {
		return errors.New(i18n.Localize("messages.confirm_password_mismatch")), "messages.confirm_password_mismatch"
	}

	if req.OldPassword == req.NewPassword {
		return errors.New(i18n.Localize("messages.new_password_must_differ")), "messages.new_password_must_differ"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		config.Log.Error(err)
		return errors.New(i18n.Localize("messages.password_hash_failed")), "messages.password_hash_failed"
	}

	err = s.repo.UpdatePassword(int64(id), string(hashedPassword))
	if err != nil {
		config.Log.Error(err)
		return errors.New(i18n.Localize("messages.no_password_update")), "messages.no_password_update"
	}

	return nil, ""
}

func (s *authService) ForgotPassword(c *gin.Context, req *prot.ForgotPasswordRequest) (*prot.ForgotPasswordResponse, error, string) {
	// check email
	user, err := s.repo.GetByKey(req.Username)
	if err != nil || user == nil || user.Email == "" {
		config.Log.Error("User not found for email/username:", req.Username, err)
		return &prot.ForgotPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.email_not_found"),
		}, nil, "messages.email_not_found"
	}

	// Check if there has been a recent password reset request (within 10 minutes)
	existingResets, err := s.passwordResetRepo.FindByEmail(user.Email)
	if err == nil && len(existingResets) > 0 {
		// Check if there are any requests within 10 minutes
		for _, reset := range existingResets {
			if time.Since(reset.CreatedAt) < 10*time.Minute {
				return &prot.ForgotPasswordResponse{
					Success: false,
					Message: i18n.Localize("messages.reset_password_too_frequent"),
				}, nil, "messages.reset_password_too_frequent"
			}
		}
	}

	token := s.generateResetToken()
	expiresAt := time.Now().Add(2 * time.Hour) // Token expires after 2 hours

	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		Status:    true,
		ExpiresAt: expiresAt,
	}

	err = s.passwordResetRepo.Create(passwordReset)
	if err != nil {
		config.Log.Error("Failed to create password reset:", err)
		return &prot.ForgotPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.reset_password_failed"),
		}, nil, "messages.reset_password_failed"
	}

	// send email
	err = s.emailService.SendPasswordResetEmail(user.Email, token, user.Username)
	if err != nil {
		config.Log.Error("Failed to send reset email:", err)
		return &prot.ForgotPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.email_send_failed"),
		}, nil, "messages.email_send_failed"
	}

	return &prot.ForgotPasswordResponse{
		Success: true,
		Message: i18n.Localize("messages.reset_password_email_sent"),
	}, nil, "messages.reset_password_email_sent"
}

func (s *authService) ResetPassword(c *gin.Context, req *prot.ResetPasswordRequest) (*prot.ResetPasswordResponse, error, string) {
	// Find the password reset record
	passwordReset, err := s.passwordResetRepo.FindByToken(req.Token)
	if err != nil {
		config.Log.Error("Invalid or expired token:", err)
		return &prot.ResetPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.invalid_or_expired_token"),
		}, nil, "messages.invalid_or_expired_token"
	}

	// Check new password
	if req.NewPassword != req.ConfirmPassword {
		return &prot.ResetPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.confirm_password_mismatch"),
		}, nil, "messages.confirm_password_mismatch"
	}

	if len(req.NewPassword) < 6 {
		return &prot.ResetPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.password_too_short"),
		}, nil, "messages.password_too_short"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		config.Log.Error("Failed to hash password:", err)
		return &prot.ResetPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.password_hash_failed"),
		}, nil, "messages.password_hash_failed"
	}

	err = s.repo.UpdatePassword(passwordReset.UserID, string(hashedPassword))
	if err != nil {
		config.Log.Error("Failed to update password:", err)
		return &prot.ResetPasswordResponse{
			Success: false,
			Message: i18n.Localize("messages.password_update_failed"),
		}, nil, "messages.password_update_failed"
	}

	// Update token
	now := time.Now()
	passwordReset.Status = false
	passwordReset.UsedAt = &now
	err = s.passwordResetRepo.Update(passwordReset)
	if err != nil {
		config.Log.Error("Failed to update password reset status:", err)
	}

	return &prot.ResetPasswordResponse{
		Success: true,
		Message: i18n.Localize("messages.password_reset_success"),
	}, nil, "messages.password_reset_success"
}

func (s *authService) AcceptRole(c *gin.Context, id int) (*prot.UserLoginResponse, bool, error, string) {
	userIDVal, _ := c.Get("userID")
	userID, _ := userIDVal.(int)

	schoolIDVal, _ := c.Get("schoolID")
	schoolID, _ := schoolIDVal.(int)

	roleIDsVal, _ := c.Get("roleIDs")
	roleNamesVal, _ := c.Get("roles")

	roleIDs, _ := roleIDsVal.([]int)
	roleNames, _ := roleNamesVal.([]string)

	valid := false
	for _, rid := range roleIDs {
		if rid == id {
			valid = true
			break
		}
	}
	if !valid {
		return nil, false, errors.New(i18n.Localize("messages.invalid_role")), "messages.invalid_role"
	}

	roleType := models.PageStudent
	roleName := models.PageStudent

	role, err := s.repo.GetRoleById(id)
	if err == nil {
		roleName = role.DefaultPageView
		roleType = role.DefaultPageView
	}

	oldToken := c.GetHeader("Token")

	token, err := utils.GenerateToken(userID, id, schoolID, roleName, roleIDs, roleNames, roleType)
	if err != nil {
		return nil, false, errors.New(i18n.Localize("messages.failed_to_generate_token")), "messages.failed_to_generate_token"
	}

	roleIdsInt32 := make([]int32, len(roleIDs))
	for i, id := range roleIDs {
		roleIdsInt32[i] = int32(id)
	}

	// Lưu token vào DB
	expires := 7 * 24 * time.Hour // 7 ngày hết hạn
	expiresTime := time.Now().Add(expires)
	if err := s.repo.SaveToken(int(userID), token, expiresTime); err != nil {
		return nil, false, errors.New(i18n.Localize("messages.auth_failed")), "messages.auth_failed"
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	_ = redis.SaveUserToken(c, int(userID), token, ip, userAgent, expires)

	// Trả về response
	response := &prot.UserLoginResponse{
		Token:    token,
		RoleId:   int32(id),
		RoleName: roleName,
		Type:     roleType,
		RoleIds:  roleIdsInt32,
		RoleNames: roleNames,
	}

	redis.ClearUserToken(userID, oldToken, ip, userAgent)

	session := sessions.Default(c)
	session.Set("token", token)
	session.Save()

	return response, true, nil, ""
}


func (s *authService) generateResetToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
