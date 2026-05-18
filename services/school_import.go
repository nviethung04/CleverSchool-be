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

func (s *schoolService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	rows, err := f.GetRows("Schools")
	if err != nil {
		return fmt.Errorf("cannot read 'Schools' sheet: %w", err)
	}

	if len(rows) < 2 {
		return errors.New("no data to import")
	}

	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}

		sID, _ := strconv.ParseInt(row[0], 10, 64)

		var fileUrl string
		if len(row) >= 10 && row[9] != "" {
			fileUrl = utils.StripDomain(row[9], models.Storage)
		}

		mediaRepo := repositories.NewMediaRepository()
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		school := models.School{
			ID:           sID,
			Name:         row[1],
			ShortName:    row[2],
			Type:         row[3],
			AddressVN:    row[4],
			AddressEN:    row[5],
			ContactName:  row[6],
			ContactPhone: row[7],
			WardCode:     row[8],
			LogoInfo:     fileInfo,
		}

		_, err := s.repo.CreateOrUpdateSchool(&school)

		if err != nil {
			config.Log.Error("Import school", school)
		}
	}

	return nil
}
