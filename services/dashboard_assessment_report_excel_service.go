package services

import (
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type DashboardAssessmentReportExcelService interface {
	ExportExcel(c *gin.Context, req *requests.DashboardAssessmentReportExcelRequest) error
}

type dashboardAssessmentReportExcelService struct {
	repo repositories.DashboardAssessmentReportExcelRepository
}

func NewDashboardAssessmentReportExcelService() DashboardAssessmentReportExcelService {
	return &dashboardAssessmentReportExcelService{
		repo: repositories.NewDashboardAssessmentReportExcelRepository(),
	}
}

func (s *dashboardAssessmentReportExcelService) ExportExcel(c *gin.Context, req *requests.DashboardAssessmentReportExcelRequest) error {
	// Parse school_ids
	schoolIDs, err := parseInt64Slice(req.SchoolIDs)
	if err != nil {
		return fmt.Errorf("invalid school_ids format: %w", err)
	}

	if len(schoolIDs) == 0 {
		return fmt.Errorf("school_ids cannot be empty")
	}

	if len(req.Subjects) == 0 {
		return fmt.Errorf("subjects cannot be empty")
	}

	// Lấy danh sách trường
	schools, err := s.repo.GetSchoolsByIDs(schoolIDs)
	if err != nil {
		return fmt.Errorf("failed to get schools: %w", err)
	}

	if len(schools) == 0 {
		return fmt.Errorf("no schools found")
	}

	// Thu thập tất cả assessment_ids từ subjects
	allAssessmentIDs := make([]int64, 0)
	assessmentInfoMap := make(map[int64]repositories.AssessmentInfo) // map[assessment_id]AssessmentInfo
	
	for _, subject := range req.Subjects {
		for _, grade := range subject.Grades {
			allAssessmentIDs = append(allAssessmentIDs, grade.AssessmentID)
		}
	}

	if len(allAssessmentIDs) == 0 {
		return fmt.Errorf("no assessments found in subjects")
	}

	// Lấy thông tin assessments
	assessments, err := s.repo.GetAssessmentsByIDs(allAssessmentIDs)
	if err != nil {
		return fmt.Errorf("failed to get assessments: %w", err)
	}

	if len(assessments) == 0 {
		return fmt.Errorf("no assessments found")
	}

	// Tạo map assessment_id -> AssessmentInfo
	for _, assessment := range assessments {
		assessmentInfoMap[assessment.ID] = assessment
	}

	// Lấy thông tin subjects và grades
	allSubjectIDs := make([]int64, 0)
	allGradeIDs := make([]int64, 0)
	for _, subject := range req.Subjects {
		allSubjectIDs = append(allSubjectIDs, subject.SubjectID)
		for _, grade := range subject.Grades {
			allGradeIDs = append(allGradeIDs, grade.GradeID)
		}
	}

	subjectNameMap, err := s.repo.GetSubjectsByIDs(allSubjectIDs)
	if err != nil {
		return fmt.Errorf("failed to get subjects: %w", err)
	}

	gradeNameMap, err := s.repo.GetGradesByIDs(allGradeIDs)
	if err != nil {
		return fmt.Errorf("failed to get grades: %w", err)
	}

	// Tạo file Excel
	f := excelize.NewFile()
	defer f.Close()

	// Lấy danh sách classes từ các schools
	classIDs, err := s.repo.GetClassesBySchoolIDs(schoolIDs)
	if err != nil {
		return fmt.Errorf("failed to get classes: %w", err)
	}

	if len(classIDs) == 0 {
		return fmt.Errorf("no classes found")
	}

	// Lấy danh sách học sinh theo class_ids
	allStudents, err := s.repo.GetStudentsByClassIDs(classIDs)
	if err != nil {
		return fmt.Errorf("failed to get students: %w", err)
	}

	if len(allStudents) == 0 {
		return fmt.Errorf("no students found")
	}

	// Lấy mapping class_id -> school_id
	classSchoolMap, err := s.repo.GetClassSchoolMap(classIDs)
	if err != nil {
		return fmt.Errorf("failed to get class-school mapping: %w", err)
	}

	// Lấy mapping student_id -> class_id từ user_classes
	studentClassMap, err := s.repo.GetStudentClassMap(classIDs)
	if err != nil {
		return fmt.Errorf("failed to get student-class mapping: %w", err)
	}

	// Group học sinh theo school_id
	studentsBySchool := make(map[int64][]repositories.StudentReportData)
	for _, student := range allStudents {
		if classID, exists := studentClassMap[student.StudentID]; exists {
			if schoolID, exists := classSchoolMap[classID]; exists {
				studentsBySchool[schoolID] = append(studentsBySchool[schoolID], student)
			}
		}
	}

	firstSheetIndex := 0
	for idx, school := range schools {
		students, exists := studentsBySchool[school.ID]
		if !exists || len(students) == 0 {
			// Nếu không có học sinh, bỏ qua trường này
			continue
		}

		// Lấy danh sách student_ids
		studentIDs := make([]int64, len(students))
		for i, student := range students {
			studentIDs[i] = student.StudentID
		}

		// Lấy assessment scores
		scoreMap, err := s.repo.GetAssessmentScores(studentIDs, allAssessmentIDs)
		if err != nil {
			return fmt.Errorf("failed to get assessment scores: %w", err)
		}

		// Tên sheet là tên trường (giới hạn 31 ký tự)
		sheetName := sanitizeSheetNameForReport(school.Name)

		var sheetIndex int
		if idx == 0 {
			// Sheet đầu tiên: rename Sheet1
			f.SetSheetName("Sheet1", sheetName)
			sheetIndex = 0
		} else {
			// Các sheet tiếp theo: tạo mới
			sheetIndex, err = f.NewSheet(sheetName)
			if err != nil {
				return fmt.Errorf("failed to create sheet for school %d: %w", school.ID, err)
			}
		}

		// Tạo nội dung cho sheet
		err = s.createSheetForSchool(f, sheetName, students, req.Subjects, scoreMap, assessmentInfoMap, subjectNameMap, gradeNameMap)
		if err != nil {
			return fmt.Errorf("failed to create sheet content for school %d: %w", school.ID, err)
		}

		if idx == 0 {
			firstSheetIndex = sheetIndex
		}
	}

	// Set sheet đầu tiên làm active
	f.SetActiveSheet(firstSheetIndex)

	// Set response headers
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	fileName := "assessment_report_excel.xlsx"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

	return f.Write(c.Writer)
}

// createSheetForSchool tạo nội dung cho một sheet của trường
func (s *dashboardAssessmentReportExcelService) createSheetForSchool(
	f *excelize.File,
	sheetName string,
	students []repositories.StudentReportData,
	subjects []requests.SubjectRequest,
	scoreMap repositories.AssessmentScoreMap,
	assessmentInfoMap map[int64]repositories.AssessmentInfo,
	subjectNameMap map[int64]string,
	gradeNameMap map[int64]string,
) error {
	// Tạo style cho header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
			Size:  12,
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
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
		},
	})

	// Tạo style cho dữ liệu
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Tạo style cho các cột điểm (căn giữa)
	scoreStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Tạo header (thứ tự: Khối -> Lớp gốc -> Lớp đã chia -> STT -> Tên học sinh)
	headers := []string{"Khối", "Lớp gốc", "Lớp đã chia", "STT", "Tên học sinh"}
	
	// Thêm các cột theo thứ tự subjects (số cột = số subjects)
	for _, subject := range subjects {
		subjectName := subjectNameMap[subject.SubjectID]
		if subjectName == "" {
			subjectName = fmt.Sprintf("Subject %d", subject.SubjectID)
		}
		headers = append(headers, subjectName)
	}

	// Ghi header
	for i, header := range headers {
		cell := getColumnNameForReport(i+1) + "1"
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Tính STT theo từng lớp đã chia (reset về 1 mỗi khi lớp đã chia thay đổi)
	sttCounter := 0
	currentClassKeyForSTT := ""
	
	// Ghi dữ liệu học sinh (thứ tự cột: Khối -> Lớp gốc -> Lớp đã chia -> STT -> Tên học sinh)
	for rowIdx, student := range students {
		row := rowIdx + 2 // Bắt đầu từ hàng 2
		classKey := student.ClassMainName + "|" + student.ClassName
		
		// Reset STT khi lớp đã chia thay đổi
		if classKey != currentClassKeyForSTT {
			sttCounter = 1
			currentClassKeyForSTT = classKey
		} else {
			sttCounter++
		}

		// Khối
		f.SetCellValue(sheetName, getColumnNameForReport(1)+fmt.Sprintf("%d", row), student.GradeNameVN)
		f.SetCellStyle(sheetName, getColumnNameForReport(1)+fmt.Sprintf("%d", row), getColumnNameForReport(1)+fmt.Sprintf("%d", row), dataStyle)

		// Lớp gốc
		f.SetCellValue(sheetName, getColumnNameForReport(2)+fmt.Sprintf("%d", row), student.ClassMainName)
		f.SetCellStyle(sheetName, getColumnNameForReport(2)+fmt.Sprintf("%d", row), getColumnNameForReport(2)+fmt.Sprintf("%d", row), dataStyle)

		// Lớp đã chia
		f.SetCellValue(sheetName, getColumnNameForReport(3)+fmt.Sprintf("%d", row), student.ClassName)
		f.SetCellStyle(sheetName, getColumnNameForReport(3)+fmt.Sprintf("%d", row), getColumnNameForReport(3)+fmt.Sprintf("%d", row), dataStyle)

		// STT (reset theo từng lớp đã chia)
		f.SetCellValue(sheetName, getColumnNameForReport(4)+fmt.Sprintf("%d", row), sttCounter)
		f.SetCellStyle(sheetName, getColumnNameForReport(4)+fmt.Sprintf("%d", row), getColumnNameForReport(4)+fmt.Sprintf("%d", row), dataStyle)

		// Tên học sinh
		f.SetCellValue(sheetName, getColumnNameForReport(5)+fmt.Sprintf("%d", row), student.StudentName)
		f.SetCellStyle(sheetName, getColumnNameForReport(5)+fmt.Sprintf("%d", row), getColumnNameForReport(5)+fmt.Sprintf("%d", row), dataStyle)

		// Các cột assessment (theo thứ tự subjects, mỗi subject 1 cột)
		colIdx := 6 // Bắt đầu từ cột F (sau Khối, Lớp gốc, Lớp đã chia, STT, Tên học sinh)
		for _, subject := range subjects {
			var scoreValue interface{}
			
			// Tìm grade phù hợp với grade_id của học sinh trong subject này
			var matchedAssessmentID int64
			found := false
			for _, grade := range subject.Grades {
				if student.GradeID == grade.GradeID {
					matchedAssessmentID = grade.AssessmentID
					found = true
					break
				}
			}
			
			if found {
				// Lấy assessment info để kiểm tra subject_id
				assessmentInfo, hasAssessment := assessmentInfoMap[matchedAssessmentID]
				if hasAssessment && assessmentInfo.SubjectID == subject.SubjectID {
					// Lấy điểm từ scoreMap
					if scoreMap[student.StudentID] != nil && scoreMap[student.StudentID][matchedAssessmentID] != nil {
						scoreValue = *scoreMap[student.StudentID][matchedAssessmentID]
					} else {
						// Nếu không có điểm, để trống
						scoreValue = ""
					}
				} else {
					// Không tìm thấy assessment hoặc subject không match, để trống
					scoreValue = ""
				}
			} else {
				// Không tìm thấy grade phù hợp, để trống
				scoreValue = ""
			}

			cell := getColumnNameForReport(colIdx) + fmt.Sprintf("%d", row)
			f.SetCellValue(sheetName, cell, scoreValue)
			f.SetCellStyle(sheetName, cell, cell, scoreStyle)
			colIdx++
		}
	}

	// Set column widths (thứ tự: Khối -> Lớp gốc -> Lớp đã chia -> STT -> Tên học sinh)
	f.SetColWidth(sheetName, "A", "A", 15) // Khối
	f.SetColWidth(sheetName, "B", "B", 20) // Lớp gốc
	f.SetColWidth(sheetName, "C", "C", 20) // Lớp đã chia
	f.SetColWidth(sheetName, "D", "D", 8)  // STT
	f.SetColWidth(sheetName, "E", "E", 30) // Tên học sinh
	
	// Set width cho các cột assessment (số cột = số subjects)
	for i := 0; i < len(subjects); i++ {
		colName := getColumnNameForReport(6 + i)
		f.SetColWidth(sheetName, colName, colName, 20)
	}

	return nil
}

// Helper functions
func parseInt64Slice(str string) ([]int64, error) {
	if str == "" {
		return nil, fmt.Errorf("empty string")
	}

	parts := strings.Split(str, ",")
	result := make([]int64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		val, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}

		result = append(result, val)
	}

	return result, nil
}

// sanitizeSheetNameForReport loại bỏ các ký tự không hợp lệ và giới hạn độ dài tên sheet
func sanitizeSheetNameForReport(name string) string {
	// Loại bỏ các ký tự không hợp lệ
	invalidChars := []string{"/", "\\", "?", "*", "[", "]"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Giới hạn độ dài tối đa 31 ký tự (Excel limit)
	if len(name) > 31 {
		name = name[:31]
	}

	return name
}

// getColumnNameForReport chuyển đổi số cột thành tên cột Excel (A, B, C, ..., AA, AB, ...)
func getColumnNameForReport(col int) string {
	col-- // Convert to 0-based (col 1 -> A, col 2 -> B, ...)
	if col < 26 {
		return string(rune('A' + col))
	}
	// Handle columns beyond Z (AA, AB, ...)
	result := ""
	for col >= 0 {
		result = string(rune('A'+(col%26))) + result
		col = col/26 - 1
	}
	return result
}


