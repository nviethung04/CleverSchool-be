package services

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/repositories"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type MicrosoftAuthService interface {
	GetAuthURL(userID uint) string
	ExchangeCodeForToken(ctx *gin.Context, code, state string) (*models.MicrosoftAccount, error)
	RefreshToken(account *models.MicrosoftAccount) error
	GetUserProfile(accessToken string) (*MicrosoftUserProfile, error)
	IsConnected(userID uint) bool
	DisconnectAccount(userID uint) error
}

type microsoftAuthService struct {
	repo        repositories.MicrosoftAuthRepository
	oauthConfig *oauth2.Config
}

type MicrosoftUserProfile struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
}

type MicrosoftTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func NewMicrosoftAuthService(repo repositories.MicrosoftAuthRepository) MicrosoftAuthService {
	microsoftConfig := config.GetMicrosoftConfig()

	// Use v2.0 endpoints to receive JWT access tokens with delegated scopes
	authURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", microsoftConfig.TenantID)
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", microsoftConfig.TenantID)

	oauthConfig := &oauth2.Config{
		ClientID:     microsoftConfig.ClientID,
		ClientSecret: microsoftConfig.ClientSecret,
		RedirectURL:  microsoftConfig.RedirectURL,
		Scopes: []string{
			"https://graph.microsoft.com/User.Read",
			"https://graph.microsoft.com/Calendars.ReadWrite",
			"https://graph.microsoft.com/OnlineMeetings.ReadWrite",
			"offline_access",
			"openid",
			"profile",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	return &microsoftAuthService{
		repo:        repo,
		oauthConfig: oauthConfig,
	}
}

// GetAuthURL generates Microsoft OAuth URL
func (s *microsoftAuthService) GetAuthURL(userID uint) string {
	state := fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())

	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCodeForToken exchanges authorization code for access token
func (s *microsoftAuthService) ExchangeCodeForToken(ctx *gin.Context, code, state string) (*models.MicrosoftAccount, error) {
	config.Log.Infof("🔄 Processing Microsoft OAuth callback - code: %s, state: %s", code[:20]+"...", state)

	tenantID := config.GetMicrosoftConfig().TenantID

	// Extract user ID from state
	parts := strings.Split(state, "_")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid state format")
	}

	userID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in state: %v", err)
	}

	config.Log.Infof("🔄 Exchanging code for token")

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}

	config.Log.Infof("✅ Successfully exchanged code for token")
	config.Log.Infof("🔑 Token scope: %v", token.Extra("scope"))
	config.Log.Infof("🔑 Access token len: %d", len(token.AccessToken))
	if len(token.AccessToken) > 0 {
		head := token.AccessToken
		if len(head) > 20 {
			head = head[:20]
		}
		config.Log.Infof("🔑 Access token head: %s", head)
	}
	if rt := token.RefreshToken; rt != "" {
		config.Log.Infof("🔑 Refresh token len: %d", len(rt))
		head := rt
		if len(head) > 20 {
			head = head[:20]
		}
		config.Log.Infof("🔑 Refresh token head: %s", head)
	} else {
		config.Log.Warn("⚠️ Refresh token is empty")
	}

	// Get user profile
	config.Log.Infof("👤 Fetching Microsoft user info")
	profile, err := s.GetUserProfile(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}

	config.Log.Infof("✅ Retrieved user info: %s (%s)", profile.DisplayName, profile.UserPrincipalName)

	// Check if account already exists
	existingAccount, err := s.repo.GetByUserID(uint(userID))

	var account *models.MicrosoftAccount
	if err == nil && existingAccount != nil {
		// Update existing account
		existingAccount.AccessToken = token.AccessToken
		existingAccount.RefreshToken = token.RefreshToken
		existingAccount.ExpiresAt = token.Expiry
		existingAccount.MicrosoftUserID = profile.ID
		existingAccount.AccountEmail = getEmailFromProfile(profile)
		existingAccount.TenantID = tenantID
		existingAccount.IsActive = true

		if updater, ok := s.repo.(interface {
			UpdateTokensWithStatusAndTenantID(userID uint, accessToken, refreshToken string, expiresAt time.Time, tenantID string, isActive bool) error
		}); ok {
			err = updater.UpdateTokensWithStatusAndTenantID(uint(userID), existingAccount.AccessToken, existingAccount.RefreshToken, existingAccount.ExpiresAt, tenantID, true)
		} else if updater2, ok := s.repo.(interface {
			UpdateTokensWithStatus(userID uint, accessToken, refreshToken string, expiresAt time.Time, isActive bool) error
		}); ok {
			err = updater2.UpdateTokensWithStatus(uint(userID), existingAccount.AccessToken, existingAccount.RefreshToken, existingAccount.ExpiresAt, true)
		} else {
			err = s.repo.UpdateTokens(uint(userID), existingAccount.AccessToken, existingAccount.RefreshToken, existingAccount.ExpiresAt)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update account: %v", err)
		}
		account = existingAccount
		config.Log.Infof("✅ Updated existing Microsoft account for user %d", userID)
		if account.AccessToken != "" {
			head := account.AccessToken
			if len(head) > 20 {
				head = head[:20]
			}
			config.Log.Infof("💾 Persisted access token head: %s", head)
		}
		if account.RefreshToken != "" {
			head := account.RefreshToken
			if len(head) > 20 {
				head = head[:20]
			}
			config.Log.Infof("💾 Persisted refresh token head: %s", head)
		}
	} else {
		// Create new account
		account = &models.MicrosoftAccount{
			UserID:          uint(userID),
			MicrosoftUserID: profile.ID,
			AccountEmail:    getEmailFromProfile(profile),
			AccessToken:     token.AccessToken,
			RefreshToken:    token.RefreshToken,
			ExpiresAt:       token.Expiry,
			TenantID:        tenantID,
			IsActive:        true,
		}

		err = s.repo.Create(account)
		if err != nil {
			return nil, fmt.Errorf("failed to create account: %v", err)
		}
		config.Log.Infof("✅ Created new Microsoft account for user %d", userID)
		if account.AccessToken != "" {
			head := account.AccessToken
			if len(head) > 20 {
				head = head[:20]
			}
			config.Log.Infof("💾 Persisted access token head: %s", head)
		}
		if account.RefreshToken != "" {
			head := account.RefreshToken
			if len(head) > 20 {
				head = head[:20]
			}
			config.Log.Infof("💾 Persisted refresh token head: %s", head)
		}
	}

	return account, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *microsoftAuthService) RefreshToken(account *models.MicrosoftAccount) error {
	if account == nil {
		return fmt.Errorf("account is nil")
	}
	if account.RefreshToken == "" {
		return fmt.Errorf("refresh token is empty")
	}

	data := url.Values{}
	data.Set("client_id", s.oauthConfig.ClientID)
	data.Set("client_secret", s.oauthConfig.ClientSecret)
	data.Set("refresh_token", account.RefreshToken)
	data.Set("grant_type", "refresh_token")

	tenant := account.TenantID
	if tenant == "" {
		tenant = config.GetMicrosoftConfig().TenantID
	}
	if tenant == "" {
		tenant = "common"
	}

	tokenURL := "https://login.microsoftonline.com/" + tenant + "/oauth2/v2.0/token"
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp MicrosoftTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	// Update account with new tokens
	account.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		account.RefreshToken = tokenResp.RefreshToken
	}
	account.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// UpdateTokens expects a user_id (not the microsoft_accounts primary key)
	return s.repo.UpdateTokens(account.UserID, account.AccessToken, account.RefreshToken, account.ExpiresAt)
}

// GetUserProfile gets user profile from Microsoft Graph API
func (s *microsoftAuthService) GetUserProfile(accessToken string) (*MicrosoftUserProfile, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user profile, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var profile MicrosoftUserProfile
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user profile: %v", err)
	}

	return &profile, nil
}

// IsConnected checks if user has connected Microsoft account
func (s *microsoftAuthService) IsConnected(userID uint) bool {
	return s.repo.IsConnected(userID)
}

// DisconnectAccount disconnects Microsoft account
func (s *microsoftAuthService) DisconnectAccount(userID uint) error {
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("account not found: %v", err)
	}

	account.IsActive = false
	account.AccessToken = ""
	account.RefreshToken = ""

	// Persist tokens cleared + is_active=false
	if updater, ok := s.repo.(interface {
		UpdateTokensWithStatus(userID uint, accessToken, refreshToken string, expiresAt time.Time, isActive bool) error
	}); ok {
		return updater.UpdateTokensWithStatus(userID, account.AccessToken, account.RefreshToken, account.ExpiresAt, false)
	}
	return s.repo.UpdateTokens(userID, account.AccessToken, account.RefreshToken, account.ExpiresAt)
}

// Helper function to get email from profile
func getEmailFromProfile(profile *MicrosoftUserProfile) string {
	if profile.Mail != "" {
		return profile.Mail
	}
	return profile.UserPrincipalName
}

