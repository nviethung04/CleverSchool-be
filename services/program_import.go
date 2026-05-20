package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"

	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *programService) Import(c *gin.Context, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		return fmt.Errorf("cannot read excel file: %w", err)
	}

	var programSheetRows [][]string
	programSheetRows, err = f.GetRows("Program")
	if err != nil {
		programSheetRows, err = f.GetRows("Subject")
		if err != nil {
			return fmt.Errorf("cannot read 'Program' or 'Subject' sheet: %w", err)
		}
	}

	programData, err := parseProgramSheet(programSheetRows)
	if err != nil {
		return err
	}

	s.repo.SetContext(c)
	programModel, err := s.upsertProgram(programData)
	if err != nil {
		return err
	}

	if len(programData.SubjectIDs) > 0 {
		if err := s.syncProgramSubjects(programModel.ID, programData.SubjectIDs); err != nil {
			return err
		}
	}

	isVtg := config.LoadConfig().IsVtg
	isTheoryType := programModel.Type == models.TypeTheory

	chapterRepo := repositories.NewChapterRepository()
	chapterRepo.SetContext(c)
	lessonRepo := repositories.NewLessonRepository()
	lessonRepo.SetContext(c)
	headingRepo := repositories.NewHeadingRepository()
	headingRepo.SetContext(c)

	var currentChapter *models.Chapter
	lessonMap := make(map[int64]*models.Lesson)

	if !isTheoryType {
		chapterRepo.SetFilter(map[string]interface{}{"program_id": programModel.ID})
		chapterRepo.SetSort(map[string]string{"sort_position": "ASC", "id": "ASC"})
		existingChapters, _, err := chapterRepo.FindAll()
		if err == nil && len(existingChapters) > 0 {
			currentChapter = &existingChapters[0]
		} else {
			chapter, err := s.upsertChapter(chapterRepo, programModel.ID, 0, "", "Chapter 1", "")
			if err != nil {
				return err
			}
			currentChapter = chapter
		}

		lessonRows, err := f.GetRows("Lessons")
		if err != nil {
			return fmt.Errorf("cannot read 'Lessons' sheet: %w", err)
		}

		lessonSort := 0
		for idx, row := range lessonRows {
			if idx == 0 {
				continue
			}

			lessonIDStr := cellValue(row, 0)
			lessonTitle := cellValue(row, 1)
			lessonDescription := cellValue(row, 2)

			if lessonIDStr == "" && lessonTitle == "" && lessonDescription == "" {
				continue
			}

			lesson, err := s.upsertLessonWithHeading(lessonRepo, currentChapter.ID, 0, lessonSort, lessonIDStr, lessonTitle, lessonDescription)
			if err != nil {
				return err
			}
			lessonMap[lesson.ID] = lesson
		}
	} else {
		var chapterRows [][]string
		chapterRows, err = f.GetRows("Chapters")
		if err != nil {
			chapterRows, err = f.GetRows("Lessons")
			if err != nil {
				return fmt.Errorf("cannot read 'Chapters' or 'Lessons' sheet: %w", err)
			}
		}

		chapterSort := 0
		lessonSort := make(map[int64]int)
		var currentHeading *models.Heading
		headingSort := make(map[int64]int)
		chaptersWithHeadings := make(map[int64]bool)

		for idx, row := range chapterRows {
			if idx == 0 {
				continue
			}

			chapterIDStr := cellValue(row, 0)
			chapterTitle := cellValue(row, 1)
			chapterDescription := cellValue(row, 2)

			if chapterIDStr != "" || chapterTitle != "" || chapterDescription != "" {
				chapter, err := s.upsertChapter(chapterRepo, programModel.ID, chapterSort, chapterIDStr, chapterTitle, chapterDescription)
				if err != nil {
					return err
				}
				currentChapter = chapter
				lessonSort[currentChapter.ID] = 0
				headingSort[currentChapter.ID] = 0
				currentHeading = nil
				chapterSort++
			}

			if currentChapter == nil {
				continue
			}

			var headingID int64 = 0
			var headingTime int32 = 0
			var skipLessonFromSheet bool = false

			if isVtg {
				headingIDStr := cellValue(row, 3)
				headingName := cellValue(row, 4)
				headingDescription := cellValue(row, 5)
				headingTimeStr := cellValue(row, 6)

				if headingIDStr != "" || headingName != "" || headingDescription != "" {
					headingTime = int32(0)
					if headingTimeStr != "" {
						if t, err := strconv.Atoi(headingTimeStr); err == nil {
							headingTime = int32(t)
						}
					}
					heading, err := s.upsertHeading(headingRepo, currentChapter.ID, headingSort[currentChapter.ID], headingIDStr, headingName, headingDescription, headingTime)
					if err != nil {
						return err
					}
					currentHeading = heading
					headingID = heading.ID
					headingSort[currentChapter.ID]++
					chaptersWithHeadings[currentChapter.ID] = true

					if heading.Time == models.HeadingMiniTime {
						skipLessonFromSheet = true
						lesson, err := s.upsertLessonForHeading(lessonRepo, currentChapter.ID, headingID, headingName, headingDescription)
						if err != nil {
							return err
						}
						lessonMap[lesson.ID] = lesson
					} else {
						lessonSort[currentChapter.ID] = 0
					}
				} else if currentHeading != nil {
					headingID = currentHeading.ID
					chaptersWithHeadings[currentChapter.ID] = true
					if currentHeading.Time == models.HeadingMiniTime {
						skipLessonFromSheet = true
					}
				}
			}

			if skipLessonFromSheet {
				continue
			}

			lessonColOffset := 3
			if isVtg {
				lessonColOffset = 7
			}

			lessonIDStr := cellValue(row, lessonColOffset)
			lessonTitle := cellValue(row, lessonColOffset+1)
			lessonDescription := cellValue(row, lessonColOffset+2)

			if lessonIDStr == "" && lessonTitle == "" && lessonDescription == "" {
				continue
			}

			if isVtg && headingID == 0 {
				continue
			}

			lesson, err := s.upsertLessonWithHeading(lessonRepo, currentChapter.ID, headingID, lessonSort[currentChapter.ID], lessonIDStr, lessonTitle, lessonDescription)
			if err != nil {
				return err
			}
			lessonMap[lesson.ID] = lesson
			lessonSort[currentChapter.ID]++
		}

		if isVtg {
			allChapters := make(map[int64]bool)
			for chapterID := range lessonSort {
				allChapters[chapterID] = true
			}
			for chapterID := range headingSort {
				allChapters[chapterID] = true
			}

			for chapterID := range allChapters {
				if !chaptersWithHeadings[chapterID] {
					if err := lessonRepo.DeleteLessonsByChapterID(chapterID); err != nil {
						config.Log.Errorf("Failed to delete lessons in chapter %d without headings: %v", chapterID, err)
					}
				}
			}

			if err := lessonRepo.DeleteLessonsWithoutHeadingIDByProgramID(programModel.ID); err != nil {
				config.Log.Errorf("Failed to delete lessons without heading_id in VTG program: %v", err)
			}
		}
	}

	lessonPlanRows, err := f.GetRows("Lesson Plans")
	if err == nil && len(lessonPlanRows) > 1 {

		lessonPlanRepo := repositories.NewLessonPlanRepository()
		lessonPlanPartRepo := repositories.NewLessonPlanPartRepository()

		var currentLessonID int64
		var currentLessonPlan *models.LessonPlan
		partSort := make(map[int64]int16)

		for idx, row := range lessonPlanRows {
			if idx == 0 {
				continue
			}

			lessonIDStr := cellValue(row, 0)
			lessonPlanIDStr := cellValue(row, 2)
			lessonPlanName := cellValue(row, 3)
			lessonPlanDesc := cellValue(row, 4)

			if lessonIDStr != "" {
				lessonID, err := strconv.ParseInt(lessonIDStr, 10, 64)
				if err != nil || lessonID <= 0 {
					continue
				}

				_, exists := lessonMap[lessonID]
				if !exists {
					continue
				}

				currentLessonID = lessonID
				partSort[lessonID] = 0

				lessonPlan, err := s.upsertLessonPlan(lessonPlanRepo, programModel.ID, lessonPlanIDStr, lessonPlanName, lessonPlanDesc, lessonID)
				if err != nil {
					return err
				}
				currentLessonPlan = lessonPlan
			}

			if currentLessonPlan == nil || currentLessonID == 0 {
				continue
			}

			partIDStr := cellValue(row, 5)
			partTitle := cellValue(row, 6)
			partObjectTitle := cellValue(row, 7)
			partTag := cellValue(row, 8)
			partGuideTeacher := cellValue(row, 9)
			partGuideStudent := cellValue(row, 10)
			partFileType := cellValue(row, 11)
			partLinkType := cellValue(row, 12)
			partFile := cellValue(row, 13)
			partCoverImage := cellValue(row, 14)
			partLink := cellValue(row, 15)

			if partIDStr == "" && partTitle == "" {
				continue
			}

			partSort[currentLessonID]++
			if _, err := s.upsertLessonPlanPart(lessonPlanPartRepo, currentLessonPlan.ID, programModel.ID, partSort[currentLessonID], partIDStr, partTitle, partObjectTitle, partTag, partGuideTeacher, partGuideStudent, partFileType, partLinkType, partFile, partCoverImage, partLink); err != nil {
				return err
			}
		}
	}

	return nil
}

type importedProgram struct {
	ID          int64
	Name        string
	Description string
	Target      string
	Type        string
	Duration    int
	Status      bool
	Image       string
	VTGData     models.ProgramVTGData
	SubjectIDs  []int64
}

func parseProgramSheet(rows [][]string) (*importedProgram, error) {
	data := &importedProgram{
		Type:     models.TypeTheory,
		Status:   true,
		Duration: 0,
		VTGData:  models.ProgramVTGData{},
	}

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(row[0]))
		value := strings.TrimSpace(row[1])

		switch key {
		case "id":
			if id, err := strconv.ParseInt(value, 10, 64); err == nil {
				data.ID = id
			}
		case "name":
			data.Name = value
		case "description":
			data.Description = value
		case "target":
			data.Target = value
		case "type":
			if value != "" {
				data.Type = value
			}
		case "duration":
			if duration, err := strconv.Atoi(value); err == nil {
				data.Duration = duration
			}
		case "status":
			data.Status = true
		case "image":
			data.Image = value
		case "code":
			data.VTGData.Code = value
		case "management code":
			data.VTGData.ManagementCode = value
		case "module":
			data.VTGData.StudyModule = value
		case "credits":
			if credits, err := strconv.Atoi(value); err == nil {
				data.VTGData.NumberOfCredits = credits
			}
		case "time":
			data.VTGData.Time = value
		case "theoretical time":
			data.VTGData.TheoreticalTime = value
		case "practice time":
			data.VTGData.PracticeTime = value
		case "discussion time":
			data.VTGData.DiscussionTime = value
		case "test time":
			data.VTGData.TestTime = value
		case "subject ids":
			data.SubjectIDs = utils.ParseIDs(value)
		}
	}

	if strings.TrimSpace(data.Image) == "-" {
		data.Image = ""
	}

	if data.Name == "" {
		return nil, errors.New("program name is required")
	}
	if data.Description == "" {
		data.Description = data.Name
	}
	if data.Target == "" {
		data.Target = data.Name
	}
	if data.Type == "" {
		data.Type = models.TypeTheory
	}

	return data, nil
}

func (s *programService) upsertProgram(data *importedProgram) (*models.Program, error) {
	mediaRepo := repositories.NewMediaRepository()
	imagePath := utils.StripDomain(data.Image, models.Storage)
	imageInfo := mediaRepo.GetMediaInfo(imagePath, models.Storage)

	if data.ID > 0 {
		existing, err := s.repo.FindNewByID(int(data.ID))
		if err != nil {
			return nil, fmt.Errorf("program id %d not found: %w", data.ID, err)
		}

		existing.Name = data.Name
		existing.Description = data.Description
		existing.Target = data.Target
		existing.Duration = data.Duration
		existing.Type = data.Type
		existing.Status = data.Status
		if imageInfo.Path != "" {
			existing.ImageInfo = imageInfo
		}
		existing.Detail.VTGData = data.VTGData

		if err := s.repo.Update(existing); err != nil {
			return nil, err
		}

		return existing, nil
	}

	program := &models.Program{
		Name:        data.Name,
		Description: data.Description,
		Target:      data.Target,
		Duration:    data.Duration,
		Type:        data.Type,
		Status:      data.Status,
		ImageInfo:   imageInfo,
		Detail: models.ProgramDetail{
			VTGData: data.VTGData,
		},
	}

	if err := s.repo.Create(program); err != nil {
		return nil, err
	}

	return program, nil
}

func (s *programService) upsertChapter(
	repo repositories.ChapterRepository,
	programID int64,
	sortPosition int,
	idStr, title, description string,
) (*models.Chapter, error) {
	chapterTitle := fallbackText(title, fmt.Sprintf("Chapter %d", sortPosition+1))
	chapterDescription := fallbackText(description, chapterTitle)

	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
		existing, err := repo.FindByID(int(id))
		if err != nil {
			return nil, fmt.Errorf("chapter id %d not found: %w", id, err)
		}

		existing.ProgramId = programID
		existing.Title = chapterTitle
		existing.ObjectTitle = chapterTitle
		existing.Description = chapterDescription
		existing.Target = chapterDescription
		existing.SortPosition = sortPosition
		existing.Status = true

		if err := repo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	chapter := &models.Chapter{
		ProgramId:    programID,
		Title:        chapterTitle,
		ObjectTitle:  chapterTitle,
		Description:  chapterDescription,
		Target:       chapterDescription,
		Status:       true,
		SortPosition: sortPosition,
	}

	if err := repo.Create(chapter); err != nil {
		return nil, err
	}

	return chapter, nil
}

func (s *programService) upsertLesson(
	repo repositories.LessonRepository,
	chapterID int64,
	sortPosition int,
	idStr, title, description string,
) (*models.Lesson, error) {
	return s.upsertLessonWithHeading(repo, chapterID, 0, sortPosition, idStr, title, description)
}

func (s *programService) upsertLessonForHeading(
	repo repositories.LessonRepository,
	chapterID, headingID int64,
	headingName, headingDescription string,
) (*models.Lesson, error) {
	repo.SetFilter(map[string]interface{}{"heading_id": headingID})
	existingLessons, err := repo.GetAll()
	if err != nil {
		return nil, err
	}

	var lesson *models.Lesson
	if len(existingLessons) > 0 {
		lesson = &existingLessons[0]
		lesson.ChapterID = chapterID
		lesson.HeadingID = headingID
		lesson.Title = headingName
		lesson.ObjectTitle = headingName
		lesson.Description = headingDescription
		lesson.Status = true

		repo.SetOmit([]string{"author_id"})
		if err := repo.Update(lesson); err != nil {
			return nil, err
		}

		if len(existingLessons) > 1 {
			for i := 1; i < len(existingLessons); i++ {
				_ = repo.Delete(int(existingLessons[i].ID))
			}
		}
	} else {
		lesson = &models.Lesson{
			ChapterID:    chapterID,
			HeadingID:    headingID,
			Title:        headingName,
			ObjectTitle:  headingName,
			Description:  headingDescription,
			Status:       true,
			SortPosition: 0,
		}

		repo.SetOmit([]string{"author_id"})
		if err := repo.Create(lesson); err != nil {
			return nil, err
		}
	}

	repo.SetFilter(map[string]interface{}{"chapter_id": chapterID})
	allChapterLessons, err := repo.GetAll()
	if err == nil {
		for _, otherLesson := range allChapterLessons {
			if otherLesson.HeadingID != headingID {
				_ = repo.Delete(int(otherLesson.ID))
			}
		}
	}

	return lesson, nil
}

func (s *programService) upsertLessonWithHeading(
	repo repositories.LessonRepository,
	chapterID, headingID int64,
	sortPosition int,
	idStr, title, description string,
) (*models.Lesson, error) {
	lessonTitle := fallbackText(title, fmt.Sprintf("Lesson %d", sortPosition))
	lessonDescription := fallbackText(description, lessonTitle)

	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
		existing, err := repo.FindByID(int(id))
		if err != nil {
			return nil, fmt.Errorf("lesson id %d not found: %w", id, err)
		}

		existing.ChapterID = chapterID
		existing.HeadingID = headingID
		existing.Title = lessonTitle
		existing.ObjectTitle = lessonTitle
		existing.Description = lessonDescription
		existing.SortPosition = sortPosition
		existing.Status = true

		repo.SetOmit([]string{"author_id"})
		if err := repo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	lesson := &models.Lesson{
		ChapterID:    chapterID,
		HeadingID:    headingID,
		Title:        lessonTitle,
		ObjectTitle:  lessonTitle,
		Description:  lessonDescription,
		Status:       true,
		SortPosition: sortPosition,
	}

	repo.SetOmit([]string{"author_id"})
	if err := repo.Create(lesson); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (s *programService) upsertHeading(
	repo repositories.HeadingRepository,
	chapterID int64,
	sortPosition int,
	idStr, name, description string,
	time int32,
) (*models.Heading, error) {
	headingName := fallbackText(name, fmt.Sprintf("Heading %d", sortPosition+1))
	headingDescription := fallbackText(description, headingName)

	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
		existing, err := repo.FindByID(int(id))
		if err != nil {
			return nil, fmt.Errorf("heading id %d not found: %w", id, err)
		}

		existing.ChapterId = chapterID
		existing.Name = headingName
		existing.Description = headingDescription
		existing.Time = time
		existing.SortPosition = sortPosition

		if err := repo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	heading := &models.Heading{
		ChapterId:    chapterID,
		Name:         headingName,
		Description:  headingDescription,
		Time:         time,
		SortPosition: sortPosition,
	}

	if err := repo.Create(heading); err != nil {
		return nil, err
	}

	return heading, nil
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func cellValue(row []string, index int) string {
	if index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func (s *programService) upsertLessonPlan(
	repo repositories.LessonPlanRepository,
	programID int64,
	lessonPlanIDStr, name, description string,
	lessonID int64,
) (*models.LessonPlan, error) {
	lessonPlanName := fallbackText(name, "Lesson Plan")
	lessonPlanDesc := fallbackText(description, lessonPlanName)

	var lessonPlan *models.LessonPlan

	if lessonPlanIDStr != "" {
		if id, err := strconv.ParseInt(lessonPlanIDStr, 10, 64); err == nil && id > 0 {
			existing, err := repo.GetByID(int(id), nil)
			if err != nil {
				return nil, fmt.Errorf("lesson plan id %d not found: %w", id, err)
			}

			existing.ProgramId = programID
			existing.Name = lessonPlanName
			existing.ObjectTitle = lessonPlanName
			existing.Description = lessonPlanDesc
			existing.Status = 1

			if err := repo.Update(existing); err != nil {
				return nil, err
			}

			lessonPlan = existing

			lessonRepo := repositories.NewLessonRepository()
			if err := lessonRepo.DeleteProgramLevelLessonPlanRefLesson(lessonID); err != nil {
				return nil, err
			}

			if err := lessonRepo.CreateLessonPlanRefLesson(lessonID, lessonPlan.ID); err != nil {
				return nil, err
			}
		}
	}

	if lessonPlan == nil {
		lessonPlan = &models.LessonPlan{
			ProgramId:    programID,
			Name:         lessonPlanName,
			ObjectTitle:  lessonPlanName,
			Description:  lessonPlanDesc,
			Status:       1,
			SortPosition: 0,
		}

		if err := repo.Create(lessonPlan, int(lessonID)); err != nil {
			return nil, err
		}
	}

	return lessonPlan, nil
}

func (s *programService) upsertLessonPlanPart(
	repo repositories.LessonPlanPartRepository,
	lessonPlanID, programID int64,
	sortPosition int16,
	idStr, title, objectTitle, tag, guideTeacher, guideStudent, fileType, linkType, file, coverImage, link string,
) (*models.LessonPlanPart, error) {
	mediaRepo := repositories.NewMediaRepository()

	partTitle := fallbackText(title, fmt.Sprintf("Part %d", sortPosition))
	partObjectTitle := fallbackText(objectTitle, partTitle)

	coverImagePath := utils.StripDomain(coverImage, models.Storage)
	coverImageInfo := mediaRepo.GetMediaInfo(coverImagePath, models.Storage)

	linkPath := utils.StripDomain(link, models.Storage)
	linkInfo := mediaRepo.GetMediaInfo(linkPath, models.Storage)

	if idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
			existing, err := repo.GetByID(id, nil)
			if err != nil {
				return nil, fmt.Errorf("lesson plan part id %d not found: %w", id, err)
			}

			existing.LessonPlanID = lessonPlanID
			existing.ProgramId = programID
			existing.Title = partTitle
			existing.ObjectTitle = partObjectTitle
			existing.Tag = tag
			existing.GuideTeacher = guideTeacher
			existing.GuideStudent = guideStudent
			existing.FileType = fileType
			existing.LinkType = linkType
			existing.File = file
			if coverImageInfo.Path != "" {
				existing.CoverImageInfo = coverImageInfo
			}
			if linkInfo.Path != "" {
				existing.LinkInfo = linkInfo
			}
			existing.SortPosition = sortPosition

			if err := repo.Update(existing); err != nil {
				return nil, err
			}
			return existing, nil
		}
	}

	part := &models.LessonPlanPart{
		LessonPlanID:   lessonPlanID,
		ProgramId:      programID,
		Title:          partTitle,
		ObjectTitle:    partObjectTitle,
		Tag:            tag,
		GuideTeacher:   guideTeacher,
		GuideStudent:   guideStudent,
		FileType:       fileType,
		LinkType:       linkType,
		File:           file,
		CoverImageInfo: coverImageInfo,
		LinkInfo:       linkInfo,
		SortPosition:   sortPosition,
	}

	if err := repo.Create(part); err != nil {
		return nil, err
	}

	return part, nil
}

func (s *programService) syncProgramSubjects(programID int64, subjectIDs []int64) error {
	unique := make(map[int64]struct{})
	normalized := make([]int64, 0, len(subjectIDs))
	for _, id := range subjectIDs {
		if id <= 0 {
			continue
		}
		if _, exists := unique[id]; exists {
			continue
		}
		unique[id] = struct{}{}
		normalized = append(normalized, id)
	}

	if len(normalized) == 0 {
		return nil
	}

	existingRefs, err := s.repo.GetProgramRefSubjectsByProgramID(programID)
	if err != nil {
		return err
	}

	desired := make(map[int64]struct{}, len(normalized))
	for _, id := range normalized {
		desired[id] = struct{}{}
	}

	for _, ref := range existingRefs {
		if _, ok := desired[ref.SubjectId]; !ok {
			if err := s.repo.DeleteProgramRefSubjectsByProgramAndSubjectID(programID, ref.SubjectId); err != nil {
				return err
			}
		}
	}

	for _, id := range normalized {
		subjectRef := models.ProgramRefSubject{
			ProgramId: programID,
			SubjectId: id,
		}
		if err := s.repo.UpdateOrCreateProgramRefSubject(subjectRef); err != nil {
			return err
		}
	}

	return nil
}
