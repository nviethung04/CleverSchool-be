package services

import (
	"be-cleverschool/config"
	"be-cleverschool/models"
	"be-cleverschool/prot"

	"github.com/gin-gonic/gin"
)

type AppConfigService interface {
	GetConfig(c *gin.Context) (*prot.AppConfig, error)
}

type appConfigService struct {
	host string
}

func NewAppConfigService() AppConfigService {
	// TODO: Load host from environment variable or config file
	host := config.LoadConfig().AppConfigUrl

	return &appConfigService{
		host: host,
	}
}

func (s *appConfigService) GetConfig(c *gin.Context) (*prot.AppConfig, error) {
	// Get platform and version from headers
	platform := c.GetHeader("platform")
	version := c.GetHeader("version")

	// Default values if not provided
	if platform == "" {
		platform = models.PlatformWeb
	}

	if version == "" {
		version = "1.0.0"
	}

	// Determine if force update is needed based on version
	forceUpdate := s.shouldForceUpdate(platform, version)

	config := &prot.AppConfig{
		ForceUpdate:      forceUpdate,
		Host:             s.host,
		Mode:             models.ModeDebug,
		Version:          s.getLatestVersion(platform),
		Platform:         platform,
		UrlVideoTutorial: "https://Clever School-7.gitbook.io/product-docs/",
	}

	return config, nil
}

// shouldForceUpdate determines if the client should force update based on version
func (s *appConfigService) shouldForceUpdate(platform, version string) bool {
	// TODO: Implement version comparison logic
	// For now, return false (no forced update)
	// You can implement semantic version comparison here
	// Example: if version < minRequiredVersion then return true

	return false
}

// getLatestVersion returns the latest available version for the platform
func (s *appConfigService) getLatestVersion(platform string) string {
	// TODO: Get latest version from database or config
	// For now, return hardcoded versions

	switch platform {
	case models.PlatformIOS:
		return "1.0"
	case models.PlatformAndroid:
		return "1.0"
	case models.PlatformWeb:
		return "1.0"
	default:
		return "1.0"
	}
}

