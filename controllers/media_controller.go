package controllers

import (
	"be-Clever School/config"
	"be-Clever School/i18n"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/services"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type MediaController struct {
	service services.MediaService
}

func NewMediaController(service services.MediaService) *MediaController {
	return &MediaController{service}
}

func (s *MediaController) UploadFile(c *gin.Context) {
	path := c.PostForm("path")

	config.Log.Info("path: ", path)

	if strings.Contains(path, "power_point") {
		powerPointService := services.NewPowerPointUploadService()

		file, err := c.FormFile("file")
		if err != nil {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.invalid_file")), "messages.invalid_file")
			return
		}

		// Validate file type
		if !powerPointService.IsValidArchiveFile(file.Filename) {
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.invalid_file_type")), "messages.invalid_file_type")
			return
		}

		// Get folder name from form or generate from filename
		folderName := c.PostForm("folder_name")
		if folderName == "" {
			subFolder := ""
			if strings.Contains(path, "power_point/") {
				subFolder = strings.TrimPrefix(path, "power_point/")
				subFolder = strings.Trim(subFolder, "/")
			}

			if subFolder != "" {
				folderName = subFolder + "/" + powerPointService.GenerateFolderName(file.Filename)
			} else {
				folderName = powerPointService.GenerateFolderName(file.Filename)
			}
		}

		// Validate folder name
		// if !powerPointService.IsValidFolderName(folderName) {
		// 	utils.Respond(c, nil, errors.New(i18n.Localize("messages.invalid_folder_name")))
		// 	return
		// }

		// Check if folder already exists
		if powerPointService.FolderExists(folderName) {
			config.Log.Error("FolderExists: ")
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.folder_already_exists")), "messages.folder_already_exists")
			return
		}

		result, err := powerPointService.UploadAndExtract(file, folderName)
		if err != nil {
			config.Log.Error("UploadAndExtract: ", err)
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.failed_to_process_file")), "messages.failed_to_process_file")
			return
		}

		// After extract, scan for one or many PowerPoint packages
		packages, scanErr := powerPointService.FindPackages(result.FolderName)
		if scanErr != nil {
			config.Log.Error("FindPackages: ", scanErr)
			utils.Respond(c, nil, errors.New(i18n.Localize("messages.failed_to_process_file")), "messages.failed_to_process_file")
			return
		}

		if len(packages) == 0 {
			// Fallback to original single behavior
			data, saveErr := powerPointService.SaveMedia(*result)
			utils.Respond(c, data, saveErr, "")
			return
		}

		if len(packages) == 1 {
			data, saveErr := powerPointService.SaveMedia(packages[0])
			utils.Respond(c, data, saveErr, "")
			return
		}

		// Multiple packages: save all and return a list
		var files []*prot.File
		for _, pkg := range packages {
			f, saveErr := powerPointService.SaveMediaWithoutMove(pkg)
			if saveErr != nil {
				config.Log.Error("SaveMediaWithoutMove: ", saveErr)
				// continue saving others; collect what we can
				continue
			}
			files = append(files, f)
		}

		// After saving all packages, move the entire folder to S3
		// This will handle the case where multiple packages are in the same folder
		if len(files) > 0 {
			go func() {
				localFolderPath := filepath.Join("power_point", result.FolderName)
				s3Prefix := fmt.Sprintf("power_point/%s", result.FolderName)

				// Check if local folder exists before trying to move
				if _, err := os.Stat(localFolderPath); err == nil {
					moveErr := powerPointService.MoveFolderToS3(localFolderPath, s3Prefix)
					if moveErr != nil {
						config.Log.Error(fmt.Sprintf("Failed to move folder %s to S3: %v", localFolderPath, moveErr))
					}
				}
			}()
		}

		resp := &prot.FileResponse{Files: files, TotalCount: uint64(len(files))}
		utils.Respond(c, resp, nil, "")
		return
	} else if path == "h5p" {
		h5pRepo := repositories.NewH5pContentRepository()
		h5pService := services.NewH5pService(h5pRepo)
		data, err := h5pService.Upload(c)

		if err != nil {
			utils.Respond(c, nil, err, "")
			return
		}
		utils.Respond(c, data, nil, "")
		return
	}

	data, err := s.service.UploadFile(c)
	utils.Respond(c, data, err, "")
}

func (s *MediaController) UploadFolder(c *gin.Context) {
	var req prot.Media

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid")
		return
	}

	err := s.service.UploadFolder(c, req)

	utils.Respond(c, nil, err, "")
}

func (s *MediaController) DeleteFolder(c *gin.Context) {
	folderIdStr := c.Param("folder_id")

	var folderId int64 = 0
	var err error

	if folderIdStr != "" {
		folderId, err = strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid folder id"), "invalid folder id")
			return
		}
	}

	err = s.service.DeleteFolder(c, folderId)

	utils.Respond(c, &prot.DeleteResponse{
		Id: folderId,
	}, err, "")
}

func (s *MediaController) DeleteFolderAndFiles(c *gin.Context) {
	folderIdStr := c.Param("folder_id")

	var folderId int64 = 0
	var err error

	if folderIdStr != "" {
		folderId, err = strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid folder id"), "invalid folder id")
			return
		}
	}

	err = s.service.DeleteFolderAndFiles(c, folderId)

	utils.Respond(c, &prot.DeleteResponse{
		Id: folderId,
	}, err, "")
}

func (s *MediaController) DeleteFile(c *gin.Context) {
	fileIdStr := c.Param("file_id")

	var fileId int64 = 0
	var err error

	if fileIdStr != "" {
		fileId, err = strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid file id"), "invalid file id")
			return
		}
	}

	err = s.service.DeleteFile(c, fileId)

	utils.Respond(c, &prot.DeleteResponse{
		Id: fileId,
	}, err, "")
}

func (s *MediaController) Folders(c *gin.Context) {
	folderIdStr := c.Query("folder_id")

	var folderId int64 = 0
	var err error

	if folderIdStr != "" {
		folderId, err = strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid folder id"), "invalid folder id")
			return
		}
	}

	data, err := s.service.Folders(c, folderId)
	utils.Respond(c, data, err, "")
}

func (s *MediaController) Files(c *gin.Context) {
	folderIdStr := c.Param("folder_id")

	var folderId int64 = 0
	var err error

	if folderIdStr != "" {
		folderId, err = strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			utils.Respond(c, nil, fmt.Errorf("invalid folder id"), "invalid folder id")
			return
		}
	}

	data, err := s.service.Files(c, folderId)
	utils.Respond(c, data, err, "")
}
