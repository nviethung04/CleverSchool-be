package services

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"be-Clever School/config"
	"be-Clever School/models"
	"be-Clever School/repositories"
	"be-Clever School/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type programInfoRow struct {
	label string
	value string
}

func (s *programService) Export(c *gin.Context, id int) (string, error) {
	isVtg := config.LoadConfig().IsVtg

	programRepo := repositories.NewProgramRepository()
	programRepo.SetContext(c)
	program, chapters, err := programRepo.GetProgramExportData(int64(id))
	if err != nil {
		return "", err
	}
	lessonPlanRows, err := programRepo.GetProgramLessonPlans(int64(id))
	if err != nil {
		return "", err
	}

	programType := program.Type
	isTheoryType := programType == models.TypeTheory

	f := excelize.NewFile()

	programSheet := "Program"
	if isVtg {
		programSheet = "Subject"
	}

	if err := f.SetSheetName("Sheet1", programSheet); err != nil {
		return "", err
	}

	programHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	if err != nil {
		return "", err
	}

	oddProgramStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	evenProgramStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	infoRows := buildProgramInfoRows(program, isVtg)
	for idx, row := range infoRows {
		line := idx + 1
		if err := f.SetCellValue(programSheet, fmt.Sprintf("A%d", line), row.label); err != nil {
			return "", err
		}
		if err := f.SetCellValue(programSheet, fmt.Sprintf("B%d", line), row.value); err != nil {
			return "", err
		}

		// Apply header style to column A
		if err := f.SetCellStyle(programSheet, fmt.Sprintf("A%d", line), fmt.Sprintf("A%d", line), programHeaderStyle); err != nil {
			return "", err
		}

		// Apply odd/even style to column B
		styleID := oddProgramStyle
		if idx%2 == 1 {
			styleID = evenProgramStyle
		}
		if err := f.SetCellStyle(programSheet, fmt.Sprintf("B%d", line), fmt.Sprintf("B%d", line), styleID); err != nil {
			return "", err
		}
	}
	if err := f.SetColWidth(programSheet, "A", "A", 28); err != nil {
		return "", err
	}
	if err := f.SetColWidth(programSheet, "B", "B", 80); err != nil {
		return "", err
	}

	var sheetName string
	if !isTheoryType {
		sheetName = "Lessons"
	} else {
		sheetName = "Chapters"
	}
	if _, err := f.NewSheet(sheetName); err != nil {
		return "", err
	}

	var headers []string

	if !isTheoryType {
		headers = []string{"Lesson ID", "Lesson Title", "Lesson Description"}
	} else if isVtg {
		headers = []string{"Chapter ID", "Chapter Name", "Chapter Description", "Heading ID", "Heading Name", "Heading Description", "Heading Time", "Lesson ID", "Lesson Title", "Lesson Description"}
	} else {
		headers = []string{"Chapter ID", "Chapter Name", "Chapter Description", "Lesson ID", "Lesson Title", "Lesson Description"}
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	})
	if err != nil {
		return "", err
	}

	oddChapterStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	evenChapterStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	for i, h := range headers {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		if err := f.SetCellValue(sheetName, cell, h); err != nil {
			return "", err
		}
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return "", err
		}
	}

	if !isTheoryType {
		if err := f.SetColWidth(sheetName, "A", "C", 35); err != nil {
			return "", err
		}
	} else if isVtg {
		if err := f.SetColWidth(sheetName, "A", "C", 25); err != nil {
			return "", err
		}
		if err := f.SetColWidth(sheetName, "D", "G", 25); err != nil {
			return "", err
		}
		if err := f.SetColWidth(sheetName, "H", "J", 35); err != nil {
			return "", err
		}
	} else {
		if err := f.SetColWidth(sheetName, "A", "C", 25); err != nil {
			return "", err
		}
		if err := f.SetColWidth(sheetName, "D", "F", 35); err != nil {
			return "", err
		}
	}

	row := 2
	for i := range chapters {
		chapter := chapters[i]
		chapterStart := row

		if !isTheoryType {
			if i == 0 {
				for idx := range chapter.Lessons {
					lesson := chapter.Lessons[idx]
					writeLessonOnlyRow(f, sheetName, row, &lesson)
					row++
				}

				for headingIdx := range chapter.Headings {
					heading := chapter.Headings[headingIdx]
					for lessonIdx := range heading.Lessons {
						lesson := heading.Lessons[lessonIdx]
						writeLessonOnlyRow(f, sheetName, row, &lesson)
						row++
					}
				}

				chapterEnd := row - 1
				if chapterEnd >= chapterStart {
					styleID := oddChapterStyle
					startCell := fmt.Sprintf("A%d", chapterStart)
					endCell := fmt.Sprintf("C%d", chapterEnd)
					if err := f.SetCellStyle(sheetName, startCell, endCell, styleID); err != nil {
						return "", err
					}
				}
			}
			break
		} else if isVtg {
			firstRow := true

			if len(chapter.Headings) == 0 {
				if len(chapter.Lessons) == 0 {
					writeChapterRowWithHeading(f, sheetName, row, &chapter, nil, nil)
					row++
				} else {
					for idx := range chapter.Lessons {
						lesson := chapter.Lessons[idx]
						if firstRow {
							writeChapterRowWithHeading(f, sheetName, row, &chapter, nil, &lesson)
							firstRow = false
						} else {
							writeChapterRowWithHeading(f, sheetName, row, nil, nil, &lesson)
						}
						row++
					}
				}
			} else {
				for headingIdx := range chapter.Headings {
					heading := chapter.Headings[headingIdx]
					if len(heading.Lessons) == 0 {
						if firstRow {
							writeChapterRowWithHeading(f, sheetName, row, &chapter, &heading, nil)
							firstRow = false
						} else {
							writeChapterRowWithHeading(f, sheetName, row, nil, &heading, nil)
						}
						row++
					} else {
						for lessonIdx := range heading.Lessons {
							lesson := heading.Lessons[lessonIdx]
							if firstRow {
								writeChapterRowWithHeading(f, sheetName, row, &chapter, &heading, &lesson)
								firstRow = false
							} else if lessonIdx == 0 {
								writeChapterRowWithHeading(f, sheetName, row, nil, &heading, &lesson)
							} else {
								writeChapterRowWithHeading(f, sheetName, row, nil, nil, &lesson)
							}
							row++
						}
					}
				}

				lessonsWithoutHeading := make([]models.Lesson, 0)
				for idx := range chapter.Lessons {
					lesson := chapter.Lessons[idx]
					if lesson.HeadingID == 0 {
						lessonsWithoutHeading = append(lessonsWithoutHeading, lesson)
					}
				}

				if len(lessonsWithoutHeading) > 0 {
					for idx := range lessonsWithoutHeading {
						lesson := lessonsWithoutHeading[idx]
						if firstRow {
							writeChapterRowWithHeading(f, sheetName, row, &chapter, nil, &lesson)
							firstRow = false
						} else {
							writeChapterRowWithHeading(f, sheetName, row, nil, nil, &lesson)
						}
						row++
					}
				}
			}

			chapterEnd := row - 1
			if chapterEnd >= chapterStart {
				styleID := oddChapterStyle
				if i%2 == 1 {
					styleID = evenChapterStyle
				}
				startCell := fmt.Sprintf("A%d", chapterStart)
				endCell := fmt.Sprintf("J%d", chapterEnd)
				if err := f.SetCellStyle(sheetName, startCell, endCell, styleID); err != nil {
					return "", err
				}
			}
		} else {
			if len(chapter.Lessons) == 0 {
				writeChapterRow(f, sheetName, row, &chapter, nil)
				row++
			} else {
				for idx := range chapter.Lessons {
					lesson := chapter.Lessons[idx]
					if idx == 0 {
						writeChapterRow(f, sheetName, row, &chapter, &lesson)
					} else {
						writeChapterRow(f, sheetName, row, nil, &lesson)
					}
					row++
				}
			}

			chapterEnd := row - 1
			if chapterEnd >= chapterStart {
				styleID := oddChapterStyle
				if i%2 == 1 {
					styleID = evenChapterStyle
				}
				startCell := fmt.Sprintf("A%d", chapterStart)
				endCell := fmt.Sprintf("F%d", chapterEnd)
				if err := f.SetCellStyle(sheetName, startCell, endCell, styleID); err != nil {
					return "", err
				}
			}
		}
	}

	lessonPlanSheet := "Lesson Plans"
	if _, err := f.NewSheet(lessonPlanSheet); err != nil {
		return "", err
	}

	lpHeaders := []string{
		"Lesson ID",
		"Lesson Title",
		"Lesson Plan ID",
		"Lesson Plan Name",
		"Lesson Plan Description",
		"Part ID",
		"Part Title",
		"Part Object Title",
		"Part Tag",
		"Part Guide Teacher",
		"Part Guide Student",
		"Part File Type",
		"Part Link Type",
		"Part File",
		"Part Cover Image",
		"Part Link",
	}

	lpHeaderStyle := headerStyle

	oddLPSytle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	evenLPStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
	})
	if err != nil {
		return "", err
	}

	for i, h := range lpHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		if err := f.SetCellValue(lessonPlanSheet, cell, h); err != nil {
			return "", err
		}
		if err := f.SetCellStyle(lessonPlanSheet, cell, cell, lpHeaderStyle); err != nil {
			return "", err
		}
	}

	if err := f.SetColWidth(lessonPlanSheet, "A", "E", 25); err != nil {
		return "", err
	}
	if err := f.SetColWidth(lessonPlanSheet, "F", "P", 30); err != nil {
		return "", err
	}

	lessonOrder := make(map[int64]int)
	lessonSet := make(map[int64]bool)
	orderIndex := 0

	if !isTheoryType {
		if len(chapters) > 0 {
			firstChapter := chapters[0]
			for _, lesson := range firstChapter.Lessons {
				if !lessonSet[lesson.ID] {
					lessonOrder[lesson.ID] = orderIndex
					lessonSet[lesson.ID] = true
					orderIndex++
				}
			}
			for _, heading := range firstChapter.Headings {
				for _, lesson := range heading.Lessons {
					if !lessonSet[lesson.ID] {
						lessonOrder[lesson.ID] = orderIndex
						lessonSet[lesson.ID] = true
						orderIndex++
					}
				}
			}
		}
	} else {
		for _, chapter := range chapters {
			for _, lesson := range chapter.Lessons {
				if !lessonSet[lesson.ID] {
					lessonOrder[lesson.ID] = orderIndex
					lessonSet[lesson.ID] = true
					orderIndex++
				}
			}
			for _, heading := range chapter.Headings {
				for _, lesson := range heading.Lessons {
					if !lessonSet[lesson.ID] {
						lessonOrder[lesson.ID] = orderIndex
						lessonSet[lesson.ID] = true
						orderIndex++
					}
				}
			}
		}
	}

	filteredRows := make([]repositories.LessonPlanExportRow, 0)
	for _, row := range lessonPlanRows {
		if _, exists := lessonSet[row.LessonID]; exists {
			filteredRows = append(filteredRows, row)
		}
	}

	sort.Slice(filteredRows, func(i, j int) bool {
		orderI := lessonOrder[filteredRows[i].LessonID]
		orderJ := lessonOrder[filteredRows[j].LessonID]
		return orderI < orderJ
	})

	lpRowIndex := 2

	for i := range filteredRows {
		data := filteredRows[i]
		groupStart := lpRowIndex

		if len(data.LessonPlanParts) == 0 {
			writeLessonPlanRow(f, lessonPlanSheet, lpRowIndex, &data, nil, true)
			lpRowIndex++
		} else {
			for idx := range data.LessonPlanParts {
				part := data.LessonPlanParts[idx]
				writeLessonPlanRow(f, lessonPlanSheet, lpRowIndex, &data, &part, idx == 0)
				lpRowIndex++
			}
		}

		groupEnd := lpRowIndex - 1
		if groupEnd >= groupStart {
			styleID := oddLPSytle
			if i%2 == 1 {
				styleID = evenLPStyle
			}
			startCell := fmt.Sprintf("A%d", groupStart)
			endCell := fmt.Sprintf("P%d", groupEnd)
			if err := f.SetCellStyle(lessonPlanSheet, startCell, endCell, styleID); err != nil {
				return "", err
			}
		}
	}

	filename := fmt.Sprintf("program_%d_%d.xlsx", program.ID, time.Now().Unix())
	if isVtg {
		filename = fmt.Sprintf("subject_%d_%d.xlsx", program.ID, time.Now().Unix())
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", err
	}

	url, err := UploadToS3(buf.Bytes(), filename)
	if err != nil {
		return "", err
	}

	return url, nil
}

func buildProgramInfoRows(program *models.Program, isVtg bool) []programInfoRow {
	subjectNames := make([]string, 0, len(program.Subjects))
	subjectIDs := make([]string, 0, len(program.Subjects))
	for _, subject := range program.Subjects {
		subjectNames = append(subjectNames, subject.Name)
		subjectIDs = append(subjectIDs, fmt.Sprintf("%d", subject.ID))
	}

	courseNames := make([]string, 0, len(program.Courses))
	courseIDs := make([]string, 0, len(program.Courses))
	for _, course := range program.Courses {
		courseNames = append(courseNames, course.Name)
		courseIDs = append(courseIDs, fmt.Sprintf("%d", course.ID))
	}

	imageURL := utils.StaticURL(program.ImageInfo.Path, models.Storage)
	if strings.TrimSpace(imageURL) == "" {
		imageURL = "-"
	}

	rows := []programInfoRow{}

	if isVtg {
		rows = []programInfoRow{
			{"ID", fmt.Sprintf("%d", program.ID)},
			{"Name", program.Name},
			{"Description", program.Description},
			{"Image", imageURL},
			{"Type", program.Type},
			{"Target", program.Target},
			{"Duration", fmt.Sprintf("%d", program.Duration)},
			{"Program IDs", placeholderIfEmpty(subjectIDs)},
			{"Programs", placeholderIfEmpty(subjectNames)},
			{"Class IDs", placeholderIfEmpty(courseIDs)},
			{"Classes", placeholderIfEmpty(courseNames)},
		}
	} else {
		rows = []programInfoRow{
			{"ID", fmt.Sprintf("%d", program.ID)},
			{"Name", program.Name},
			{"Description", program.Description},
			{"Image", imageURL},
			{"Type", program.Type},
			{"Target", program.Target},
			{"Duration", fmt.Sprintf("%d", program.Duration)},
			{"Subject IDs", placeholderIfEmpty(subjectIDs)},
			{"Subjects", placeholderIfEmpty(subjectNames)},
			{"Course IDs", placeholderIfEmpty(courseIDs)},
			{"Courses", placeholderIfEmpty(courseNames)},
		}
	}

	// Only add VTG data if it exists
	if program.Detail.VTGData.Code != "" {
		rows = append(rows,
			programInfoRow{"Code", program.Detail.VTGData.Code},
			programInfoRow{"Management Code", program.Detail.VTGData.ManagementCode},
			programInfoRow{"Module", program.Detail.VTGData.StudyModule},
			programInfoRow{"Credits", fmt.Sprintf("%d", program.Detail.VTGData.NumberOfCredits)},
			programInfoRow{"Time", program.Detail.VTGData.Time},
			programInfoRow{"Theoretical Time", program.Detail.VTGData.TheoreticalTime},
			programInfoRow{"Practice Time", program.Detail.VTGData.PracticeTime},
			programInfoRow{"Discussion Time", program.Detail.VTGData.DiscussionTime},
			programInfoRow{"Test Time", program.Detail.VTGData.TestTime},
		)
	}

	return rows
}

func placeholderIfEmpty(values []string) string {
	clean := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return "-"
	}
	return strings.Join(clean, ", ")
}

func writeChapterRow(f *excelize.File, sheet string, row int, chapter *models.Chapter, lesson *models.Lesson) {
	if chapter != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), chapter.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), chapter.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), chapter.Description)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "")
	}

	if lesson != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), lesson.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), lesson.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), lesson.Description)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "")
	}
}

func writeChapterRowWithHeading(f *excelize.File, sheet string, row int, chapter *models.Chapter, heading *models.Heading, lesson *models.Lesson) {
	// Chapter columns (A, B, C)
	if chapter != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), chapter.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), chapter.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), chapter.Description)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "")
	}

	// Heading columns (D, E, F, G)
	if heading != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), heading.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), heading.Name)
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), heading.Description)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), heading.Time)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), "")
	}

	// Lesson columns (H, I, J)
	if lesson != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), lesson.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), lesson.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), lesson.Description)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), "")
	}
}

func writeLessonOnlyRow(f *excelize.File, sheet string, row int, lesson *models.Lesson) {
	// Lesson columns only (A, B, C)
	if lesson != nil {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), lesson.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), lesson.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), lesson.Description)
	} else {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "")
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "")
	}
}

func writeLessonPlanRow(
	f *excelize.File,
	sheet string,
	row int,
	data *repositories.LessonPlanExportRow,
	part *models.LessonPlanPart,
	showLesson bool,
) {
	if showLesson {
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), data.LessonID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), data.LessonTitle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), data.LessonPlanID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("D%d", row), data.LessonPlanName)
		_ = f.SetCellValue(sheet, fmt.Sprintf("E%d", row), data.LessonPlanDesc)
	} else {
		for _, col := range []string{"A", "B", "C", "D", "E"} {
			_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), "")
		}
	}

	if part != nil {
		link := part.LinkInfo.Path
		coverImage := part.CoverImageInfo.Path

		if link != "" {
			link = utils.StaticURL(link, models.Storage)
		}

		if coverImage != "" {
			coverImage = utils.StaticURL(coverImage, models.Storage)
		}

		_ = f.SetCellValue(sheet, fmt.Sprintf("F%d", row), part.ID)
		_ = f.SetCellValue(sheet, fmt.Sprintf("G%d", row), part.Title)
		_ = f.SetCellValue(sheet, fmt.Sprintf("H%d", row), part.ObjectTitle)
		_ = f.SetCellValue(sheet, fmt.Sprintf("I%d", row), part.Tag)
		_ = f.SetCellValue(sheet, fmt.Sprintf("J%d", row), part.GuideTeacher)
		_ = f.SetCellValue(sheet, fmt.Sprintf("K%d", row), part.GuideStudent)
		_ = f.SetCellValue(sheet, fmt.Sprintf("L%d", row), part.FileType)
		_ = f.SetCellValue(sheet, fmt.Sprintf("M%d", row), part.LinkType)
		_ = f.SetCellValue(sheet, fmt.Sprintf("N%d", row), part.File)
		_ = f.SetCellValue(sheet, fmt.Sprintf("O%d", row), coverImage)
		_ = f.SetCellValue(sheet, fmt.Sprintf("P%d", row), link)
	} else {
		for _, col := range []string{"F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"} {
			_ = f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), "")
		}
	}
}
