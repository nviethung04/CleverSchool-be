package config

import (
	"os"
)

// MicrosoftConfig holds Microsoft OAuth configuration
type MicrosoftConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TenantID     string
}

// GetMicrosoftConfig returns Microsoft OAuth configuration from environment variables
func GetMicrosoftConfig() *MicrosoftConfig {
	return &MicrosoftConfig{
		ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("MICROSOFT_REDIRECT_URL"),
		TenantID:     os.Getenv("MICROSOFT_TENANT_ID"),
	}
}

// GetAuthURL returns Microsoft OAuth authorization URL
func (mc *MicrosoftConfig) GetAuthURL(state string) string {
	scopes := "https://graph.microsoft.com/Calendars.ReadWrite https://graph.microsoft.com/OnlineMeetings.ReadWrite https://graph.microsoft.com/User.Read offline_access"

	return "https://login.microsoftonline.com/" + mc.TenantID + "/oauth2/v2.0/authorize?" +
		"client_id=" + mc.ClientID +
		"&response_type=code" +
		"&redirect_uri=" + mc.RedirectURL +
		"&response_mode=query" +
		"&scope=" + scopes +
		"&state=" + state +
		"&prompt=consent"
}

// GetTokenURL returns Microsoft OAuth token exchange URL
func (mc *MicrosoftConfig) GetTokenURL() string {
	return "https://login.microsoftonline.com/" + mc.TenantID + "/oauth2/v2.0/token"
}

// GetUserInfoURL returns Microsoft Graph user info API URL
func (mc *MicrosoftConfig) GetUserInfoURL() string {
	return "https://graph.microsoft.com/v1.0/me"
}

// GetGraphAPI returns Microsoft Graph API base URL
func (mc *MicrosoftConfig) GetGraphAPI() string {
	return "https://graph.microsoft.com/v1.0"
}
