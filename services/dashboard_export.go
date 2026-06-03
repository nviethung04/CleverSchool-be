package services

import (
	"be-lms/config"
	"be-lms/i18n"
	"be-lms/prot"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func (s *dashboardService) Export(c *gin.Context) (string, error) {
	userRegister := s.cache.GetCachedUserRegister(c)
	studentRegister := s.cache.GetCachedStudentRegister(c)
	teacherRegister := s.cache.GetCachedTeacherRegister(c)
	activity := s.cache.GetCachedActivity(c)
	school := s.cache.GetCachedSchool(c)

	userOverview := &prot.UserOverview{
		User:     userRegister,
		Student:  studentRegister,
		Teacher:  teacherRegister,
		Activity: activity,
		School:   school,
	}

	courseOverview := s.cache.GetCachedCourseOverview(c)

	learning := s.cache.GetCachedLearningOverview(c)
	riskAndWarning := s.cache.GetCachedRiskWarning(c)
	questionBank := s.cache.GetCachedQuestionBank(c)
	systemUsageOverview := s.cache.GetCachedSystemUsage(c)

	f := excelize.NewFile()

	headerStyle, _ := f.NewStyle(&excelize.Style{
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
		},
	})

	// ===== Sheet: User overview =====
	userOverviewSheet := "User overview"
	f.SetSheetName("Sheet1", userOverviewSheet)
	s.SetSheetUserOverview(f, userOverviewSheet, userOverview, headerStyle)

	// ===== Sheet: Course Overview =====
	courseOverviewSheet := "Course Overview"
	f.NewSheet(courseOverviewSheet)
	s.SetSheetCourseOverview(f, courseOverviewSheet, courseOverview, headerStyle)

	// ===== Sheet: Overview learning =====
	overviewLearningSheet := "Overview learning"
	f.NewSheet(overviewLearningSheet)
	s.SetSheetOverviewLearning(f, overviewLearningSheet, learning, headerStyle)

	// ===== Sheet: Risk & Warning =====
	riskAndWarningSheet := "Risk & Warning"
	f.NewSheet(riskAndWarningSheet)
	s.SetSheetRiskAndWarning(f, riskAndWarningSheet, riskAndWarning, headerStyle)

	// ===== Sheet: System usage =====
	systemUsageSheet := "System usage"
	f.NewSheet(systemUsageSheet)
	s.SetSheetSystemUsageOverview(f, systemUsageSheet, systemUsageOverview, headerStyle)

	// ===== Sheet: Question Bank =====
	questionBankSheet := "Question Bank"
	f.NewSheet(questionBankSheet)
	s.SetSheetQuestionBank(f, questionBankSheet, questionBank, headerStyle)

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("dashboard_%d.xlsx", time.Now().Unix())

	// Save Excel file to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return "", fmt.Errorf("failed to write Excel to buffer: %v", err)
	}

	// Upload to S3
	fileURL, err := UploadToS3(buf.Bytes(), filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	config.Log.Info("Export user Excel to S3:", fileURL)

	return fileURL, nil
}

func (s *dashboardService) SetSheetUserOverview(f *excelize.File, sheetName string, userOverview *prot.UserOverview, headerStyle int) error {
	userOverviewHeaders := []string{
		"Title", "Total", "Change Value", "Description", "Status",
	}
	for i, h := range userOverviewHeaders {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	row := 2

	// User
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userOverview.User.Title)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), userOverview.User.TotalNumber)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", userOverview.User.ChangeValue))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), userOverview.User.Message)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), userOverview.User.Status)
	row++

	// School
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userOverview.School.Title)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), userOverview.School.TotalNumber)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", userOverview.School.ChangeValue))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), userOverview.School.Message)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), userOverview.School.Status)
	row++

	// Teacher
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userOverview.Teacher.Title)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), userOverview.Teacher.TotalNumber)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", userOverview.Teacher.ChangeValue))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), userOverview.Teacher.Message)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), userOverview.Teacher.Status)
	row++

	// Student
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userOverview.Student.Title)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), userOverview.Student.TotalNumber)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", userOverview.Student.ChangeValue))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), userOverview.Student.Message)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), userOverview.Student.Status)
	row++

	// Activity
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), userOverview.Activity.Title)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), userOverview.Activity.TotalNumber)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", userOverview.Activity.ChangeValue))
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), userOverview.Activity.Message)
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), userOverview.Activity.Status)

	return nil
}

func (s *dashboardService) SetSheetCourseOverview(f *excelize.File, sheetName string, courseOverview *prot.CourseOverview, headerStyle int) error {
	// Headers overviews
	overviewHeaders := []string{"Title", "Count"}
	for i, h := range overviewHeaders {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	row := 2
	for _, o := range courseOverview.Overviews {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), o.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), o.Count)
		row++
	}

	// Headers rates
	ratesHeaders := []string{"Title", "Total", "Complete Total", "Percent"}
	for i, h := range ratesHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row+1)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	row += 2 // sang dòng mới dưới headers
	for _, r := range courseOverview.Rates {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), r.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), r.Total)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), r.CompleteTotal)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%.2f%%", r.Percent))
		row++
	}

	return nil
}

func (s *dashboardService) SetSheetOverviewLearning(f *excelize.File, sheetName string, learning *prot.OverviewLearning, headerStyle int) error {
	row := 1

	// Section title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Top highest")
	row++

	topHighestHeaders := []string{"ID", "Exam Id", "Name", "Exam", "Score"}
	for i, h := range topHighestHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row++

	// Data
	for _, u := range learning.TopHighest {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), u.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), u.ExamId)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), u.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), u.ExamName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), u.Score)
		row++
	}

	row++ // 1 dòng trống

	// Section title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Top lowest")
	row++

	// Headers
	topLowestHeaders := []string{"ID", "Exam Id", "Name", "Exam", "Score"}
	for i, h := range topLowestHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row++

	// Data
	for _, u := range learning.TopLowest {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), u.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), u.ExamId)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), u.StudentName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), u.ExamName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), u.Score)
		row++
	}

	return nil
}

func (s *dashboardService) SetSheetRiskAndWarning(f *excelize.File, sheetName string, riskAndWarning *prot.RiskAndWarning, headerStyle int) error {
	row := 1

	// Section title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Inactive students")
	row++

	// Headers
	inactiveStudentHeaders := []string{"ID", "Name", "Class", "Last login", "Absent days"}
	for i, h := range inactiveStudentHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row++

	// Data
	for _, s := range riskAndWarning.InactiveStudents {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), s.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.ClassName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), s.LastLogin)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), s.AbsentDays)
		row++
	}

	row++ // 1 dòng trống

	// Section title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Declining students")
	row++

	// Headers
	decliningStudentHeaders := []string{"ID", "Name", "Class", "Scores"}
	for i, h := range decliningStudentHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row++

	for _, s := range riskAndWarning.DecliningStudents {
		var scoreStrings []string
		for _, score := range s.Scores {
			scoreStrings = append(scoreStrings, fmt.Sprintf("%v", score.Score))
		}
		scores := strings.Join(scoreStrings, ",")
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), s.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), s.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), s.ClassName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), scores)
		row++
	}

	row++ // 1 dòng trống

	// Section title
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Slow grading teachers")
	row++

	// Headers
	slowGradingTeacherHeaders := []string{"ID", "Name", "Graded", "Submitted", "Percent", "Average waiting time"}
	for i, h := range slowGradingTeacherHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row++

	for _, t := range riskAndWarning.SlowGradingTeachers {
		avgWaitHours := i18n.Localize("time.hours", map[string]interface{}{"Number": t.AvgWaitHours})
		percentStr := fmt.Sprintf("%.2f%%", t.Percent)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), t.Id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), t.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), t.Graded)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), t.Submitted)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), percentStr)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), avgWaitHours)
		row++
	}

	return nil
}

func (s *dashboardService) SetSheetQuestionBank(f *excelize.File, sheetName string, questionBank *prot.QuestionBankOverview, headerStyle int) error {
	// Attributes
	attrHeaders := []string{"Attribute", "Name", "Count", "Percent"}
	for i, h := range attrHeaders {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	row := 2
	for _, attr := range questionBank.Attributes {
		for _, item := range attr.Items {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), attr.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), item.Count)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("%.2f%%", item.Percent*100))
			row++
		}
	}

	// Types
	typeHeaders := []string{"Type", "Total", "Percent"}
	for i, h := range typeHeaders {
		cell := fmt.Sprintf("%s1", string('F'+i))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	rowType := 2
	for _, typ := range questionBank.Types {
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowType), typ.Type)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowType), typ.Total)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowType), fmt.Sprintf("%.2f%%", typ.Percent))
		rowType++
	}

	// Media usage
	mediaStartRow := max(row, rowType) + 2
	cellA := fmt.Sprintf("A%d", mediaStartRow)
	f.SetCellValue(sheetName, cellA, "Total questions")
	f.SetCellStyle(sheetName, cellA, cellA, headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", mediaStartRow), questionBank.MediaUsage.TotalQuestions)

	cellA = fmt.Sprintf("A%d", mediaStartRow+1)
	f.SetCellValue(sheetName, cellA, "With audio")
	f.SetCellStyle(sheetName, cellA, cellA, headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", mediaStartRow+1), questionBank.MediaUsage.WithAudio)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", mediaStartRow+1), fmt.Sprintf("%.2f%%", questionBank.MediaUsage.AudioPercent))

	cellA = fmt.Sprintf("A%d", mediaStartRow+2)
	f.SetCellValue(sheetName, cellA, "With image")
	f.SetCellStyle(sheetName, cellA, cellA, headerStyle)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", mediaStartRow+2), questionBank.MediaUsage.WithImage)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", mediaStartRow+2), fmt.Sprintf("%.2f%%", questionBank.MediaUsage.ImagePercent))

	return nil
}

func (s *dashboardService) SetSheetSystemUsageOverview(f *excelize.File, sheetName string, sys *prot.SystemUsageOverview, headerStyle int) error {
	// ===== Weekly Usage =====
	weeklyHeaders := []string{"Week", "Average Duration", "Average Duration (minutes)"}
	for i, h := range weeklyHeaders {
		cell := fmt.Sprintf("%s1", string('A'+i))
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	row := 2
	for _, w := range sys.WeeklyUsage {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), w.WeekLabel)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), w.AverageDuration)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), w.AverageDurationMinute)
		row++
	}

	// ===== Device Usage =====
	deviceStartRow := row + 2

	deviceHeaders := []string{"Device Name", "User Count", "Percentage"}
	for i, h := range deviceHeaders {
		cell := fmt.Sprintf("%s%d", string('A'+i), deviceStartRow)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	row = deviceStartRow + 1
	for _, d := range sys.DeviceUsages {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), d.DeviceName)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), d.UserCount)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.2f%%", d.Percentage))
		row++
	}

	return nil
}

// max helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
