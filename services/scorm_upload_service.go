package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"be-cleverschool/models"
)

type ScormUploadService struct {
	uploadDir string
}

func NewScormUploadService() *ScormUploadService {
	return &ScormUploadService{
		uploadDir: "./scorm-packages",
	}
}

// UploadScormPackage uploads and extracts a SCORM ZIP package
func (s *ScormUploadService) UploadScormPackage(filePath, originalName string) (*models.ScormActivity, error) {
	// Create upload directory if not exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique package name
	packageName := s.generatePackageName(originalName)
	packageDir := filepath.Join(s.uploadDir, packageName)

	// Extract ZIP file
	if err := s.extractZip(filePath, packageDir); err != nil {
		return nil, fmt.Errorf("failed to extract ZIP: %w", err)
	}

	// Validate SCORM package
	scormInfo, err := s.validateScormPackage(packageDir)
	if err != nil {
		// Clean up invalid package
		os.RemoveAll(packageDir)
		return nil, fmt.Errorf("invalid SCORM package: %w", err)
	}

	// Create SCORM activity record
	activity := &models.ScormActivity{
		Title:       scormInfo.Title,
		Description: scormInfo.Description,
		Version:     scormInfo.Version,
		LaunchURL:   scormInfo.LaunchURL,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return activity, nil
}

// extractZip extracts a ZIP file to the specified directory
func (s *ScormUploadService) extractZip(zipPath, extractDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer reader.Close()

	// Create extraction directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// Extract files
	for _, file := range reader.File {
		filePath := filepath.Join(extractDir, file.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(filePath, filepath.Clean(extractDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, file.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create file
		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Open source file
		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open source file: %w", err)
		}

		// Copy content
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	return nil
}

// validateScormPackage validates the extracted SCORM package
func (s *ScormUploadService) validateScormPackage(packageDir string) (*ScormPackageInfo, error) {
	info := &ScormPackageInfo{}

	// Check for imsmanifest.xml
	manifestPath := filepath.Join(packageDir, "imsmanifest.xml")
	if _, err := os.Stat(manifestPath); err == nil {
		// Parse manifest to get package info
		if err := s.parseManifest(manifestPath, info); err != nil {
			return nil, fmt.Errorf("failed to parse manifest: %w", err)
		}
	} else {
		// Fallback: look for common launch files
		if err := s.findLaunchFile(packageDir, info); err != nil {
			return nil, fmt.Errorf("failed to find launch file: %w", err)
		}
	}

	// Validate required fields
	if info.Title == "" {
		info.Title = filepath.Base(packageDir)
	}
	if info.LaunchURL == "" {
		return nil, fmt.Errorf("no launch file found")
	}
	if info.Version == "" {
		info.Version = "1.2" // Default to SCORM 1.2
	}

	return info, nil
}

// parseManifest parses imsmanifest.xml to extract package information
func (s *ScormUploadService) parseManifest(manifestPath string, info *ScormPackageInfo) error {
	// This is a simplified parser - in production you might want to use a proper XML parser
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Extract title
	if strings.Contains(contentStr, "<title>") {
		start := strings.Index(contentStr, "<title>") + 7
		end := strings.Index(contentStr, "</title>")
		if end > start {
			info.Title = strings.TrimSpace(contentStr[start:end])
		}
	}

	// Extract description
	if strings.Contains(contentStr, "<description>") {
		start := strings.Index(contentStr, "<description>") + 13
		end := strings.Index(contentStr, "</description>")
		if end > start {
			info.Description = strings.TrimSpace(contentStr[start:end])
		}
	}

	// Extract launch file from resources
	if strings.Contains(contentStr, "href=") {
		start := strings.Index(contentStr, "href=\"") + 6
		end := strings.Index(contentStr[start:], "\"")
		if end > 0 {
			info.LaunchURL = "/scorm/" + filepath.Base(filepath.Dir(manifestPath)) + "/" + contentStr[start:start+end]
		}
	}

	// Try to detect SCORM version
	if strings.Contains(contentStr, "http://www.imsproject.org/xsd/imscp_rootv1p1p2") {
		info.Version = "2004"
	} else {
		info.Version = "1.2"
	}

	return nil
}

// findLaunchFile looks for common launch files in the package
func (s *ScormUploadService) findLaunchFile(packageDir string, info *ScormPackageInfo) error {
	commonLaunchFiles := []string{
		"index_Clever School.html",
		"index_scorm.html",
		"index.html",
		"story_html5.html",
		"story.html",
		"launch.html",
	}

	for _, fileName := range commonLaunchFiles {
		filePath := filepath.Join(packageDir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			info.LaunchURL = "/scorm/" + filepath.Base(packageDir) + "/" + fileName
			return nil
		}
	}

	return fmt.Errorf("no launch file found")
}

// generatePackageName generates a unique package name
func (s *ScormUploadService) generatePackageName(originalName string) string {
	// Remove .zip extension
	name := strings.TrimSuffix(originalName, ".zip")

	// Add timestamp to make it unique
	timestamp := time.Now().Format("20060102_150405")

	return fmt.Sprintf("%s_%s", name, timestamp)
}

// ScormPackageInfo contains information extracted from a SCORM package
type ScormPackageInfo struct {
	Title       string
	Description string
	Version     string
	LaunchURL   string
}

// ListUploadedPackages lists all uploaded SCORM packages
func (s *ScormUploadService) ListUploadedPackages() ([]string, error) {
	if _, err := os.Stat(s.uploadDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(s.uploadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload directory: %w", err)
	}

	var packages []string
	for _, entry := range entries {
		if entry.IsDir() {
			packages = append(packages, entry.Name())
		}
	}

	return packages, nil
}

// DeletePackage deletes an uploaded SCORM package
func (s *ScormUploadService) DeletePackage(packageName string) error {
	packagePath := filepath.Join(s.uploadDir, packageName)

	if err := os.RemoveAll(packagePath); err != nil {
		return fmt.Errorf("failed to delete package: %w", err)
	}

	return nil
}

// GetPackageInfo gets information about a specific package
func (s *ScormUploadService) GetPackageInfo(packageName string) (*ScormPackageInfo, error) {
	packagePath := filepath.Join(s.uploadDir, packageName)

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package not found: %s", packageName)
	}

	return s.validateScormPackage(packagePath)
}

