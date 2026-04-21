package config

import (
	"os"
)

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GetGoogleConfig returns Google OAuth configuration from environment variables
func GetGoogleConfig() *GoogleConfig {
	return &GoogleConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	}
}

// GetAuthURL returns Google OAuth authorization URL
func (gc *GoogleConfig) GetAuthURL(state string) string {
	scopes := "https://www.googleapis.com/auth/calendar https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"

	return "https://accounts.google.com/o/oauth2/auth?" +
		"response_type=code" +
		"&client_id=" + gc.ClientID +
		"&redirect_uri=" + gc.RedirectURL +
		"&scope=" + scopes +
		"&state=" + state +
		"&access_type=offline" +
		"&prompt=consent"
}

func (gc *GoogleConfig) GetAuthURLWithRedirect(state, redirectURI string) string {
	scopes := "https://www.googleapis.com/auth/calendar https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"

	return "https://accounts.google.com/o/oauth2/auth?" +
		"response_type=code" +
		"&client_id=" + gc.ClientID +
		"&redirect_uri=" + redirectURI +
		"&scope=" + scopes +
		"&state=" + state +
		"&access_type=offline" +
		"&prompt=consent"
}

// GetTokenURL returns Google OAuth token exchange URL
func (gc *GoogleConfig) GetTokenURL() string {
	return "https://oauth2.googleapis.com/token"
}

// GetUserInfoURL returns Google user info API URL
func (gc *GoogleConfig) GetUserInfoURL() string {
	return "https://www.googleapis.com/oauth2/v2/userinfo"
}

// GetCalendarAPI returns Google Calendar API base URL
func (gc *GoogleConfig) GetCalendarAPI() string {
	return "https://www.googleapis.com/calendar/v3"
}
