package services

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/repositories"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ZoomAuthService interface {
	GetAuthURL(userID int64) (string, error)
	HandleCallback(code, state string) (*models.ZoomAccount, error)
	RefreshToken(userID int64) error
	DisconnectAccount(userID int64) error
	GetConnectionStatus(userID int64) (bool, *models.ZoomAccount, error)
}

type zoomAuthService struct {
	zoomAuthRepo repositories.ZoomAuthRepository
	zoomConfig   *config.ZoomConfig
}

type ZoomTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type ZoomUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Type  int    `json:"type"`
}

func NewZoomAuthService(zoomAuthRepo repositories.ZoomAuthRepository) ZoomAuthService {
	return &zoomAuthService{
		zoomAuthRepo: zoomAuthRepo,
		zoomConfig:   config.GetZoomConfig(),
	}
}

func (s *zoomAuthService) GetAuthURL(userID int64) (string, error) {
	state := fmt.Sprintf("user_%d_%d", userID, time.Now().Unix())
	return s.zoomConfig.GetAuthURL(state), nil
}

func (s *zoomAuthService) HandleCallback(code, state string) (*models.ZoomAccount, error) {
	// Exchange code for tokens
	tokenResp, err := s.exchangeCodeForToken(code)
	if err != nil {
		return nil, err
	}

	// Get user info from Zoom
	userInfo, err := s.getZoomUserInfo(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// Extract userID from state
	var userID int64
	fmt.Sscanf(state, "user_%d_%d", &userID, new(int64))

	// Create or update zoom account
	zoomAccount := &models.ZoomAccount{
		UserID:       userID,
		ZoomUserID:   userInfo.ID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		AccountEmail: userInfo.Email,
		IsActive:     true,
	}

	// Delete existing account if any
	s.zoomAuthRepo.Delete(userID)

	// Create new account
	err = s.zoomAuthRepo.Create(zoomAccount)
	if err != nil {
		return nil, err
	}

	return zoomAccount, nil
}

func (s *zoomAuthService) RefreshToken(userID int64) error {
	account, err := s.zoomAuthRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	// Refresh token logic
	tokenResp, err := s.refreshAccessToken(account.RefreshToken)
	if err != nil {
		return err
	}

	// Update tokens
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()
	return s.zoomAuthRepo.UpdateTokens(userID, tokenResp.AccessToken, tokenResp.RefreshToken, expiresAt)
}

func (s *zoomAuthService) DisconnectAccount(userID int64) error {
	return s.zoomAuthRepo.Delete(userID)
}

func (s *zoomAuthService) GetConnectionStatus(userID int64) (bool, *models.ZoomAccount, error) {
	if !s.zoomAuthRepo.IsConnected(userID) {
		return false, nil, nil
	}

	account, err := s.zoomAuthRepo.GetByUserID(userID)
	if err != nil {
		return false, nil, err
	}

	return true, account, nil
}

// Private helper methods
func (s *zoomAuthService) exchangeCodeForToken(code string) (*ZoomTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.zoomConfig.RedirectURL)
	data.Set("client_id", s.zoomConfig.ClientID)
	data.Set("client_secret", s.zoomConfig.ClientSecret)

	fmt.Printf("🔄 Exchanging code for token - URL: %s\n", s.zoomConfig.GetTokenURL())
	fmt.Printf("🔄 Request data: %v\n", data)

	resp, err := http.PostForm(s.zoomConfig.GetTokenURL(), data)
	if err != nil {
		fmt.Printf("❌ HTTP POST error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Zoom response status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Read body error: %v\n", err)
		return nil, err
	}

	fmt.Printf("📡 Zoom response body: %s\n", string(body))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("zoom API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp ZoomTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		fmt.Printf("❌ JSON unmarshal error: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ Token response: AccessToken=%s, RefreshToken=%s, ExpiresIn=%d\n",
		tokenResp.AccessToken[:10]+"...", tokenResp.RefreshToken[:10]+"...", tokenResp.ExpiresIn)

	return &tokenResp, err
}

func (s *zoomAuthService) refreshAccessToken(refreshToken string) (*ZoomTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.zoomConfig.ClientID)
	data.Set("client_secret", s.zoomConfig.ClientSecret)

	resp, err := http.PostForm(s.zoomConfig.GetTokenURL(), data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp ZoomTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	return &tokenResp, err
}

func (s *zoomAuthService) getZoomUserInfo(accessToken string) (*ZoomUserResponse, error) {
	url := s.zoomConfig.BaseURL + "/users/me"
	fmt.Printf("🔄 Getting user info from: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Create request error: %v\n", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ HTTP GET error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("📡 User info response status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Read body error: %v\n", err)
		return nil, err
	}

	fmt.Printf("📡 User info response body: %s\n", string(body))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("zoom user info API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var userResp ZoomUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		fmt.Printf("❌ JSON unmarshal error: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ User info: ID=%s, Email=%s\n", userResp.ID, userResp.Email)

	return &userResp, err
}
