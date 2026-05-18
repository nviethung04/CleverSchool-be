package services

import (
	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *facultyService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	rows, err := f.GetRows("Faculties")
	if err != nil {
		return fmt.Errorf("cannot read 'Faculties' sheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		fID, _ := strconv.ParseInt(row[0], 10, 64)
		schoolID, _ := strconv.ParseInt(row[2], 10, 64)

		name := ""
		if len(row) > 2 && !isEmpty(row[2]) {
			name = row[2]
		}

		code := ""
		if len(row) > 3 && !isEmpty(row[3]) {
			code = row[3]
		}

		nameHead := ""
		if len(row) > 4 && !isEmpty(row[4]) {
			nameHead = row[4]
		}

		description := ""
		if len(row) > 5 && !isEmpty(row[5]) {
			description = row[5]
		}

		var fileInfo models.MediaInfo
		if len(row) > 6 && !isEmpty(row[6]) {
			fileUrl := utils.StripDomain(row[6], models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo = mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		}

		faculty := models.Faculty{
			ID:          fID,
			SchoolId:    schoolID,
			Name:        name,
			Code:        code,
			NameHead:    nameHead,
			Description: description,
			AvatarInfo:  fileInfo,
			Status:      true,
		}

		_, err := s.repo.CreateOrUpdateFaculty(&faculty)

		if err != nil {
			config.Log.Error("Import faculty", faculty)
		}
	}

	return nil
}
