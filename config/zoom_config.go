package config

import (
	"os"
)

type ZoomConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	BaseURL      string
}

func GetZoomConfig() *ZoomConfig {
	return &ZoomConfig{
		ClientID:     os.Getenv("ZOOM_CLIENT_ID"),
		ClientSecret: os.Getenv("ZOOM_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("ZOOM_REDIRECT_URL"),
		BaseURL:      "https://api.zoom.us/v2",
	}
}

// Zoom OAuth URLs
func (z *ZoomConfig) GetAuthURL(state string) string {
	return "https://zoom.us/oauth/authorize?response_type=code&client_id=" + z.ClientID +
		"&redirect_uri=" + z.RedirectURL + "&state=" + state
}

func (z *ZoomConfig) GetTokenURL() string {
	return "https://zoom.us/oauth/token"
}
