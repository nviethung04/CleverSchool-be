package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type GoogleAuthService interface {
	GetAuthURL(userID uint, host string) (string, error)
	HandleCallback(code, state string) (*models.GoogleAccount, error)
	RefreshToken(userID uint) error
	IsConnected(userID uint) bool
	Disconnect(userID uint) error
	GetUserInfo(accessToken string) (*GoogleUserInfo, error)
}

type googleAuthService struct {
	repo   repositories.GoogleAuthRepository
	config *config.GoogleConfig
}

func (s *googleAuthService) encodeState(userID uint, host string) string {
	hostEscaped := url.QueryEscape(host)
	return fmt.Sprintf("user_%d_%d_%s", userID, time.Now().Unix(), hostEscaped)
}

func (s *googleAuthService) parseState(state string) (userID uint, host string, err error) {
	// Backward compatible:
	// - old: user_{id}_{ts}
	// - new: user_{id}_{ts}_{host}
	parts := strings.SplitN(state, "_", 4)
	if len(parts) < 3 {
		return 0, "", fmt.Errorf("invalid state parameter")
	}

	userID64, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, "", fmt.Errorf("invalid user ID in state: %v", err)
	}

	if len(parts) == 4 {
		host, _ = url.QueryUnescape(parts[3])
	}

	return uint(userID64), host, nil
}

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

func NewGoogleAuthService() GoogleAuthService {
	return &googleAuthService{
		repo:   repositories.NewGoogleAuthRepository(),
		config: config.GetGoogleConfig(),
	}
}

func (s *googleAuthService) buildRedirectURI(host string) string {
	// If already absolute, return as is
	if strings.HasPrefix(s.config.RedirectURL, "http://") || strings.HasPrefix(s.config.RedirectURL, "https://") {
		return s.config.RedirectURL
	}
	scheme := "http"
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		scheme = "http"
	} else if host != "" {
		scheme = "https"
	}
	if host == "" {
		return s.config.RedirectURL
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, s.config.RedirectURL)
}

// GetAuthURL generates Google OAuth authorization URL
func (s *googleAuthService) GetAuthURL(userID uint, host string) (string, error) {
	state := s.encodeState(userID, host)
	redirectURI := s.buildRedirectURI(host)
	authURL := s.config.GetAuthURLWithRedirect(state, redirectURI)

	config.Log.Infof("🔗 Generated Google OAuth URL for user %d: %s", userID, authURL)
	return authURL, nil
}

// HandleCallback handles OAuth callback and exchanges code for token
func (s *googleAuthService) HandleCallback(code, state string) (*models.GoogleAccount, error) {
	config.Log.Infof("🔄 Processing Google OAuth callback - code: %s, state: %s",
		code[:10]+"...", state)

	userID, stateHost, err := s.parseState(state)
	if err != nil {
		return nil, err
	}

	redirectURI := s.buildRedirectURI(stateHost)
	tokenResp, err := s.exchangeCodeForToken(code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %v", err)
	}

	// Get user info from Google
	userInfo, err := s.GetUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	// Calculate token expiration
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Check if account already exists
	existingAccount, err := s.repo.GetByUserID(userID)
	if err == nil && existingAccount != nil {
		// Update existing account
		err = s.repo.UpdateTokens(userID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to update tokens: %v", err)
		}

		// Ensure no duplicate active rows remain
		_ = s.repo.DeleteOtherActive(userID, existingAccount.ID)

		config.Log.Infof("✅ Updated existing Google account for user %d", userID)
		return s.repo.GetByUserID(userID)
	} // Create new account
	account := &models.GoogleAccount{
		UserID:       userID,
		GoogleUserID: userInfo.ID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		AccountEmail: userInfo.Email,
		IsActive:     true,
	}

	err = s.repo.Create(account)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google account: %v", err)
	}

	// Ensure only one active row per user
	_ = s.repo.DeleteOtherActive(userID, account.ID)

	config.Log.Infof("✅ Created new Google account for user %d: %s", userID, userInfo.Email)
	return account, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (s *googleAuthService) exchangeCodeForToken(code string, redirectURI string) (*GoogleTokenResponse, error) {
	config.Log.Infof("🔄 Exchanging code for token")

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.ClientID)
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	config.Log.Infof("📡 Request to: %s", s.config.GetTokenURL())
	config.Log.Infof("📡 Request data: %s", data.Encode())

	resp, err := http.PostForm(s.config.GetTokenURL(), data)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	config.Log.Infof("📡 Google response status: %d", resp.StatusCode)

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	config.Log.Infof("📡 Google response body: %s", string(bodyBytes))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	config.Log.Infof("✅ Successfully exchanged code for token")
	return &tokenResp, nil
}

// GetUserInfo retrieves user information from Google API
func (s *googleAuthService) GetUserInfo(accessToken string) (*GoogleUserInfo, error) {
	config.Log.Infof("👤 Fetching Google user info")

	req, err := http.NewRequest("GET", s.config.GetUserInfoURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google API error: status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	config.Log.Infof("✅ Retrieved user info: %s (%s)", userInfo.Name, userInfo.Email)
	return &userInfo, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *googleAuthService) RefreshToken(userID uint) error {
	account, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get Google account: %v", err)
	}

	data := url.Values{}
	data.Set("refresh_token", account.RefreshToken)
	data.Set("client_id", s.config.ClientID)
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm(s.config.GetTokenURL(), data)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %v", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Google API error: status %d", resp.StatusCode)
	}

	// Update tokens in database
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	refreshToken := tokenResp.RefreshToken
	if refreshToken == "" {
		refreshToken = account.RefreshToken // Keep old refresh token if not provided
	}

	return s.repo.UpdateTokens(userID, tokenResp.AccessToken, refreshToken, expiresAt)
}

// IsConnected checks if user has connected Google account
func (s *googleAuthService) IsConnected(userID uint) bool {
	return s.repo.IsConnected(userID)
}

// Disconnect removes Google account connection
func (s *googleAuthService) Disconnect(userID uint) error {
	// Keep row for audit but clear tokens and deactivate
	return s.repo.DeactivateAndClear(userID)
}
