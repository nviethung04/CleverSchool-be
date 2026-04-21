package models

type AppConfig struct {
	ForceUpdate      bool   `json:"force_update"`
	Host             string `json:"host"`
	Mode             string `json:"mode"`
	Version          string `json:"version"`
	Platform         string `json:"platform"`
	UrlVideoTutorial string `json:"url_video_tutorial"`
}

// Config modes
const (
	ModeRelease = "release"
	ModeDebug   = "debug"
)

// Platforms
const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)
