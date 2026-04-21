package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"be-lms/services"

	"github.com/gin-gonic/gin"
)

type ScormUploadController struct {
	uploadService *services.ScormUploadService
	scormService  *services.ScormService
}

func NewScormUploadController() *ScormUploadController {
	return &ScormUploadController{
		uploadService: services.NewScormUploadService(),
		scormService:  services.NewScormService(),
	}
}

// UploadScormPackage handles SCORM package upload
func (c *ScormUploadController) UploadScormPackage(ctx *gin.Context) {
	// Get uploaded file
	file, err := ctx.FormFile("scorm_package")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded or invalid file",
		})
		return
	}

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".zip") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Only ZIP files are allowed",
		})
		return
	}

	// Validate file size (max 100MB)
	if file.Size > 100*1024*1024 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "File size too large. Maximum size is 100MB",
		})
		return
	}

	// Create temporary file
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create temporary directory",
		})
		return
	}

	tempFile := filepath.Join(tempDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, tempFile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save uploaded file",
		})
		return
	}

	// Clean up temp file after processing
	defer os.Remove(tempFile)

	// Upload and extract SCORM package
	activity, err := c.uploadService.UploadScormPackage(tempFile, file.Filename)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to process SCORM package: %v", err),
		})
		return
	}

	// Save to database
	savedActivity, err := c.scormService.CreateScormActivity(
		activity.Title,
		activity.Description,
		activity.Version,
		activity.LaunchURL,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save SCORM activity to database",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "SCORM package uploaded successfully",
		"activity": gin.H{
			"id":          savedActivity.ID,
			"title":       savedActivity.Title,
			"description": savedActivity.Description,
			"version":     savedActivity.Version,
			"launchURL":   savedActivity.LaunchURL,
			"status":      savedActivity.Status,
		},
		"file": gin.H{
			"originalName": file.Filename,
			"size":         file.Size,
		},
	})
}

// ListUploadedPackages lists all uploaded SCORM packages
func (c *ScormUploadController) ListUploadedPackages(ctx *gin.Context) {
	packages, err := c.uploadService.ListUploadedPackages()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list packages",
		})
		return
	}

	var packageInfos []gin.H
	for _, pkg := range packages {
		info, err := c.uploadService.GetPackageInfo(pkg)
		if err != nil {
			// Skip packages with errors
			continue
		}

		packageInfos = append(packageInfos, gin.H{
			"name":        pkg,
			"title":       info.Title,
			"description": info.Description,
			"version":     info.Version,
			"launchURL":   info.LaunchURL,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":  true,
		"packages": packageInfos,
		"count":    len(packageInfos),
	})
}

// GetPackageInfo gets information about a specific package
func (c *ScormUploadController) GetPackageInfo(ctx *gin.Context) {
	packageName := ctx.Param("name")
	if packageName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Package name is required",
		})
		return
	}

	info, err := c.uploadService.GetPackageInfo(packageName)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Package not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"package": gin.H{
			"name":        packageName,
			"title":       info.Title,
			"description": info.Description,
			"version":     info.Version,
			"launchURL":   info.LaunchURL,
		},
	})
}

// DeletePackage deletes an uploaded SCORM package
func (c *ScormUploadController) DeletePackage(ctx *gin.Context) {
	packageName := ctx.Param("name")
	if packageName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Package name is required",
		})
		return
	}

	// Check if package exists
	_, err := c.uploadService.GetPackageInfo(packageName)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Package not found",
		})
		return
	}

	// Delete package
	if err := c.uploadService.DeletePackage(packageName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete package",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Package deleted successfully",
		"package": packageName,
	})
}

// UploadForm serves the upload form HTML

func (c *ScormUploadController) UploadForm(ctx *gin.Context) {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SCORM Package Upload</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            text-align: center;
        }
        .upload-form {
            margin: 20px 0;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input[type="file"] {
            width: 100%;
            padding: 10px;
            border: 2px dashed #3498db;
            border-radius: 4px;
            background-color: #f8f9fa;
        }
        button {
            background-color: #3498db;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover {
            background-color: #2980b9;
        }
        .progress {
            margin: 20px 0;
            display: none;
        }
        .progress-bar {
            width: 100%;
            height: 20px;
            background-color: #ecf0f1;
            border-radius: 10px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background-color: #3498db;
            width: 0%;
            transition: width 0.3s ease;
        }
        .result {
            margin-top: 20px;
            padding: 15px;
            border-radius: 4px;
            display: none;
        }
        .result.success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .result.error {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .package-list {
            margin-top: 30px;
        }
        .package-item {
            background-color: #f8f9fa;
            padding: 15px;
            margin: 10px 0;
            border-radius: 4px;
            border-left: 4px solid #3498db;
        }
        .package-title {
            font-weight: bold;
            color: #2c3e50;
        }
        .package-info {
            color: #7f8c8d;
            font-size: 14px;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 SCORM Package Upload</h1>

        <div class="upload-form">
            <h3>Upload SCORM Package</h3>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="form-group">
                    <label for="scorm_package">Select SCORM Package (.zip):</label>
                    <input type="file" id="scorm_package" name="scorm_package" accept=".zip" required>
                </div>
                <button type="submit">🚀 Upload Package</button>
            </form>
        </div>

        <div class="progress" id="progress">
            <div class="progress-bar">
                <div class="progress-fill" id="progressFill"></div>
            </div>
            <p id="progressText">Uploading...</p>
        </div>

        <div class="result" id="result"></div>

        <div class="package-list">
            <h3>📦 Uploaded Packages</h3>
            <div id="packageList">Loading...</div>
        </div>
    </div>

    <script>
        // Upload form handling
        document.getElementById('uploadForm').addEventListener('submit', async function(e) {
            e.preventDefault();

            const formData = new FormData();
            const fileInput = document.getElementById('scorm_package');
            const file = fileInput.files[0];

            if (!file) {
                showResult('error', 'Please select a file');
                return;
            }

            if (!file.name.toLowerCase().endsWith('.zip')) {
                showResult('error', 'Please select a ZIP file');
                return;
            }

            formData.append('scorm_package', file);

            // Show progress
            showProgress();

            try {
                const response = await fetch('/api/scorm/upload', {
                    method: 'POST',
                    body: formData
                });

                const result = await response.json();

                if (response.ok) {
                    showResult('success', 'Package uploaded successfully! ' + result.message);
                    loadPackages(); // Refresh package list
                } else {
                    showResult('error', 'Upload failed: ' + result.error);
                }
            } catch (error) {
                showResult('error', 'Upload failed: ' + error.message);
            } finally {
                hideProgress();
            }
        });

        function showProgress() {
            document.getElementById('progress').style.display = 'block';
            document.getElementById('result').style.display = 'none';
        }

        function hideProgress() {
            document.getElementById('progress').style.display = 'none';
        }

        function showResult(type, message) {
            const resultDiv = document.getElementById('result');
            resultDiv.className = 'result ' + type;
            resultDiv.textContent = message;
            resultDiv.style.display = 'block';
        }

        // Load uploaded packages
        async function loadPackages() {
            try {
                const response = await fetch('/api/scorm/packages');
                const result = await response.json();

                if (response.ok) {
                    displayPackages(result.packages);
                } else {
                    document.getElementById('packageList').innerHTML = 'Error loading packages: ' + result.error;
                }
            } catch (error) {
                document.getElementById('packageList').innerHTML = 'Error loading packages: ' + error.message;
            }
        }

        function displayPackages(packages) {
            const packageList = document.getElementById('packageList');

            if (packages.length === 0) {
                packageList.innerHTML = '<p>No packages uploaded yet.</p>';
                return;
            }

            let html = '';
            packages.forEach(pkg => {
                html += '<div class="package-item">';
                html += '<div class="package-title">' + (pkg.title || pkg.name) + '</div>';
                html += '<div class="package-info">';
                html += '<strong>Name:</strong> ' + pkg.name + '<br>';
                html += '<strong>Version:</strong> ' + pkg.version + '<br>';
                html += '<strong>Description:</strong> ' + (pkg.description || 'No description') + '<br>';
                html += '<strong>Launch URL:</strong> ' + pkg.launchURL + '<br>';
                html += '<button onclick="deletePackage(\'' + pkg.name + '\')" style="background-color: #e74c3c; margin-top: 10px;">🗑️ Delete</button>';
                html += '</div>';
                html += '</div>';
            });

            packageList.innerHTML = html;
        }

        async function deletePackage(packageName) {
            if (!confirm('Are you sure you want to delete this package?')) {
                return;
            }

            try {
                const response = await fetch('/api/scorm/packages/' + packageName, {
                    method: 'DELETE'
                });

                const result = await response.json();

                if (response.ok) {
                    showResult('success', 'Package deleted successfully!');
                    loadPackages(); // Refresh package list
                } else {
                    showResult('error', 'Delete failed: ' + result.error);
                }
            } catch (error) {
                showResult('error', 'Delete failed: ' + error.message);
            }
        }

        // Load packages on page load
        document.addEventListener('DOMContentLoaded', function() {
            loadPackages();
        });
    </script>
</body>
</html>`

	ctx.Header("Content-Type", "text/html")
	ctx.String(http.StatusOK, html)
}
