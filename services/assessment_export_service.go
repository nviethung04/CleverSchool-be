package services

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/repositories"
	"be-cleverschool/requests"
	"be-cleverschool/utils"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// buildSkillTypeTreeFromFlat xây dựng cây từ flat list (giống như trong resources)
func buildSkillTypeTreeFromFlat(flatTypes []models.StudyReportSkillType) []*prot.StudyReportSkillType {
	nodes := make(map[int64]*prot.StudyReportSkillType, len(flatTypes))
	roots := make([]*prot.StudyReportSkillType, 0)

	// Tạo nodes
	for _, t := range flatTypes {
		nodes[t.ID] = &prot.StudyReportSkillType{
			Id:        t.ID,
			SkillId:   t.SkillId,
			NameVn:    t.NameVn,
			NameEn:    t.NameEn,
			SortOrder: int32(t.SortOrder),
			Level:     int32(t.Level),
			ParentId: func() int64 {
				if t.ParentId != nil {
					return *t.ParentId
				}
				return 0
			}(),
			NodeTypes: make([]*prot.StudyReportSkillType, 0),
		}
	}

	// Xây dựng cây
	for _, t := range flatTypes {
		node := nodes[t.ID]
		if t.ParentId != nil && nodes[*t.ParentId] != nil {
			parent := nodes[*t.ParentId]
			parent.NodeTypes = append(parent.NodeTypes, node)
		} else {
			roots = append(roots, node)
		}
	}

	// Sắp xếp
	var sortTree func(list []*prot.StudyReportSkillType)
	sortTree = func(list []*prot.StudyReportSkillType) {
		for i := 0; i < len(list)-1; i++ {
			for j := i + 1; j < len(list); j++ {
				if list[i].SortOrder > list[j].SortOrder ||
					(list[i].SortOrder == list[j].SortOrder && list[i].Id > list[j].Id) {
					list[i], list[j] = list[j], list[i]
				}
			}
		}
		for _, n := range list {
			sortTree(n.NodeTypes)
		}
	}
	sortTree(roots)

	return roots
}

// flattenSkillTypes lấy tất cả các leaf nodes (con bé nhất) từ cây types
func flattenSkillTypes(types []*prot.StudyReportSkillType) []struct {
	ID   int64
	Name string
} {
	var result []struct {
		ID   int64
		Name string
	}

	var flattenRecursive func(nodes []*prot.StudyReportSkillType)
	flattenRecursive = func(nodes []*prot.StudyReportSkillType) {
		for _, node := range nodes {
			if len(node.NodeTypes) == 0 {
				// Leaf node - con bé nhất
				result = append(result, struct {
					ID   int64
					Name string
				}{
					ID:   node.Id,
					Name: node.NameVn,
				})
			} else {
				// Có con, tiếp tục đệ quy
				flattenRecursive(node.NodeTypes)
			}
		}
	}

	flattenRecursive(types)
	return result
}

// getMaxDepth tính độ sâu tối đa của cây types
func getMaxDepth(types []*prot.StudyReportSkillType) int {
	maxDepth := 0
	var depthRecursive func(nodes []*prot.StudyReportSkillType, currentDepth int)
	depthRecursive = func(nodes []*prot.StudyReportSkillType, currentDepth int) {
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
		for _, node := range nodes {
			if len(node.NodeTypes) > 0 {
				depthRecursive(node.NodeTypes, currentDepth+1)
			}
		}
	}
	depthRecursive(types, 1)
	return maxDepth
}

// EvaluationColumnInfo cấu trúc thông tin cột cho tiêu chí đánh giá
type EvaluationColumnInfo struct {
	Type      string // "star" hoặc "check"
	SkillID   int64  // 0 nếu là parent column (Tiêu chí tặng sao/Tiêu chí check)
	SkillName string
	MaxStar   *int
	Types     []struct {
		ID   int64
		Name string
	}
	TypeTree []*prot.StudyReportSkillType // Cây types để render phân cấp
	StartCol int                          // Cột bắt đầu (0-based)
	EndCol   int                          // Cột kết thúc (0-based, exclusive)
	IsParent bool                         // true nếu là cột cha (Tiêu chí tặng sao/Tiêu chí check)
}

// renderLevelTypes render các types ở một level cụ thể
func renderLevelTypes(f *excelize.File, sheetName string, evalCol EvaluationColumnInfo, targetLevel int, row int, styleID int, isName bool) {
	var renderRecursive func(nodes []*prot.StudyReportSkillType, currentLevel int, colOffset *int)
	renderRecursive = func(nodes []*prot.StudyReportSkillType, currentLevel int, colOffset *int) {
		for _, node := range nodes {
			if currentLevel == targetLevel {
				// Tìm số lượng leaf nodes con của node này
				leafCount := countLeafNodes(node.NodeTypes)
				if leafCount == 0 {
					leafCount = 1 // Nếu không có con, vẫn có 1 cột (chính nó là leaf)
				}

				startCol := evalCol.StartCol + *colOffset + 1
				endCol := evalCol.StartCol + *colOffset + leafCount

				if isName {
					startCell := getCellName(row, startCol)
					endCell := getCellName(row, endCol)
					f.SetCellValue(sheetName, startCell, node.NameVn)
					for c := startCol; c <= endCol; c++ {
						cell := getCellName(row, c)
						f.SetCellStyle(sheetName, cell, cell, styleID)
					}
					if endCol > startCol {
						f.MergeCell(sheetName, startCell, endCell)
					}
				} else {
					startCell := getCellName(row, startCol)
					endCell := getCellName(row, endCol)
					f.SetCellValue(sheetName, startCell, node.Id)
					for c := startCol; c <= endCol; c++ {
						cell := getCellName(row, c)
						f.SetCellStyle(sheetName, cell, cell, styleID)
					}
					if endCol > startCol {
						f.MergeCell(sheetName, startCell, endCell)
					}
				}

				*colOffset += leafCount
			} else if currentLevel < targetLevel {
				// Chưa đến level cần render, tiếp tục đệ quy
				renderRecursive(node.NodeTypes, currentLevel+1, colOffset)
			}
			// Nếu currentLevel > targetLevel, không làm gì (đã vượt quá)
		}
	}

	colOffset := 0
	renderRecursive(evalCol.TypeTree, 1, &colOffset)
}

// countLeafNodes đếm số lượng leaf nodes trong cây
func countLeafNodes(nodes []*prot.StudyReportSkillType) int {
	if len(nodes) == 0 {
		return 0
	}
	count := 0
	for _, node := range nodes {
		if len(node.NodeTypes) == 0 {
			count++
		} else {
			count += countLeafNodes(node.NodeTypes)
		}
	}
	return count
}

type AssessmentExportService interface {
	ExportExcel(c *gin.Context, req *requests.AssessmentExportExcelRequest) error
	ImportExcel(c *gin.Context, fileReader io.Reader) (interface{}, error)
}

type assessmentExportService struct {
	repo                repositories.AssessmentExportRepository
	scoreRepo           repositories.AssessmentScoreRepository
	studyReportRepo     repositories.StudyReportRepository
	assessmentRepo      repositories.AssessmentRepository
}

func NewAssessmentExportService() AssessmentExportService {
	return &assessmentExportService{
		repo:            repositories.NewAssessmentExportRepository(),
		scoreRepo:       repositories.NewAssessmentScoreRepository(),
		studyReportRepo: repositories.NewStudyReportRepository(),
		assessmentRepo:  repositories.NewAssessmentRepository(),
	}
}

func (s *assessmentExportService) ExportExcel(c *gin.Context, req *requests.AssessmentExportExcelRequest) error {
	// Parse course_ids từ string "1,2,3,4"
	courseIDs, err := parseCourseIDs(req.CourseIDs)
	if err != nil {
		return fmt.Errorf("invalid course_ids format: %w", err)
	}

	if len(courseIDs) == 0 {
		return fmt.Errorf("course_ids cannot be empty")
	}

	// Lấy assessment với criteria và subcriteria
	assessment, err := s.repo.GetAssessmentWithCriteria(req.AssessmentID)
	if err != nil {
		return fmt.Errorf("failed to get assessment: %w", err)
	}

	// Lấy StudyReportCriteria nếu có
	var studyReportCriteria *models.StudyReportCriteria
	if assessment.StudyReportCriteriaID > 0 {
		studyReportCriteria, err = s.repo.GetStudyReportCriteria(assessment.StudyReportCriteriaID)
		if err != nil {
			// Không bắt buộc phải có, nên chỉ log warning
			fmt.Printf("Warning: failed to get study report criteria: %v\n", err)
		}
	}

	// Tạo file Excel với nhiều sheet (mỗi course một sheet)
	return s.createExcelFileWithMultipleSheets(c, assessment, courseIDs, studyReportCriteria)
}

// parseCourseIDs parse string "1,2,3,4" thành []int64
func parseCourseIDs(courseIDsStr string) ([]int64, error) {
	if courseIDsStr == "" {
		return nil, fmt.Errorf("course_ids is empty")
	}

	parts := strings.Split(courseIDsStr, ",")
	courseIDs := make([]int64, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		courseID, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid course_id: %s", part)
		}

		courseIDs = append(courseIDs, courseID)
	}

	return courseIDs, nil
}

// createExcelFileWithMultipleSheets tạo file Excel với nhiều sheet (mỗi course một sheet)
func (s *assessmentExportService) createExcelFileWithMultipleSheets(c *gin.Context, assessment *models.Assessment, courseIDs []int64, studyReportCriteria *models.StudyReportCriteria) error {
	f := excelize.NewFile()
	defer f.Close()

	// Tạo sheet cho mỗi course
	firstSheetIndex := 0
	for i, courseID := range courseIDs {
		// Lấy thông tin course
		course, err := s.repo.GetCourseByID(courseID)
		if err != nil {
			return fmt.Errorf("failed to get course %d: %w", courseID, err)
		}

		// Lấy danh sách học sinh theo course
		students, err := s.repo.GetStudentsByCourseID(courseID)
		if err != nil {
			return fmt.Errorf("failed to get students for course %d: %w", courseID, err)
		}

		// Tên sheet: hiển thị course.ObjectTitle, nếu không có thì dùng course.Name kèm số thứ tự
		var sheetName string
		if course.ObjectTitle != "" {
			sheetName = course.ObjectTitle
		} else {
			// Nếu không có object_title, dùng tên khóa kèm số thứ tự
			sheetName = fmt.Sprintf("%s %d", course.Name, i+1)
		}
		// Loại bỏ các ký tự không hợp lệ trong tên sheet (Excel chỉ cho phép tối đa 31 ký tự)
		sheetName = sanitizeSheetName(sheetName)

		var index int
		if i == 0 {
			// Sheet đầu tiên: rename Sheet1 thay vì tạo mới
			f.SetSheetName("Sheet1", sheetName)
			index = 0
		} else {
			// Các sheet tiếp theo: tạo mới
			index, err = f.NewSheet(sheetName)
			if err != nil {
				return fmt.Errorf("failed to create sheet for course %d: %w", courseID, err)
			}
		}

		// Tạo nội dung cho sheet này
		err = s.createSheetForCourse(f, sheetName, assessment, students, course, studyReportCriteria)
		if err != nil {
			return fmt.Errorf("failed to create sheet content for course %d: %w", courseID, err)
		}

		// Set sheet đầu tiên làm active
		if i == 0 {
			firstSheetIndex = index
		}
	}

	// Set sheet đầu tiên làm active
	f.SetActiveSheet(firstSheetIndex)

	// Set response headers cho file Excel .xlsx
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// Tên file: assessment_report_ngày xuất (format: YYYYMMDD)
	now := time.Now()
	fileName := fmt.Sprintf("assessment_report_%s.xlsx", now.Format("20060102"))
	// Loại bỏ các ký tự không hợp lệ trong tên file
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	fileName = strings.ReplaceAll(fileName, "*", "_")
	fileName = strings.ReplaceAll(fileName, "?", "_")
	fileName = strings.ReplaceAll(fileName, "\"", "_")
	fileName = strings.ReplaceAll(fileName, "<", "_")
	fileName = strings.ReplaceAll(fileName, ">", "_")
	fileName = strings.ReplaceAll(fileName, "|", "_")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Transfer-Encoding", "binary")

	return f.Write(c.Writer)
}

// sanitizeSheetName loại bỏ các ký tự không hợp lệ và giới hạn độ dài tên sheet
func sanitizeSheetName(name string) string {
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

// createSheetForCourse tạo nội dung cho một sheet của course
func (s *assessmentExportService) createSheetForCourse(f *excelize.File, sheetName string, assessment *models.Assessment, students []repositories.StudentWithCourse, course *models.Course, studyReportCriteria *models.StudyReportCriteria) error {

	// Style cho các tiêu đề cũ (tên khóa, tên học sinh, mã định danh, nhận xét): màu xanh nhạt
	basicHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#000000",
			Size:  12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#B4C6E7"}, // Màu xanh nhạt hơn #4F81BD
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Protection: &excelize.Protection{
			Locked: true, // Lock để bảo vệ
		},
	})

	// Style cho "Tiêu chí điểm" (hàng 1): màu đậm hơn #d9ead3 (màu các ô dưới)
	scoreHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#000000",
			Size:  12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#93c47d"}, // Màu đậm hơn #d9ead3
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Protection: &excelize.Protection{
			Locked: true,
		},
	})

	// Style cho "Tiêu chí đánh giá" (hàng 1): màu đậm hơn các ô dưới
	evalHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#000000",
			Size:  12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffd966"}, // Màu đậm hơn #fce5cd, #fff2cc, #d0e0e3
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Protection: &excelize.Protection{
			Locked: true,
		},
	})

	// Các style được tạo động trong code khi cần thiết

	// Tạo style cho ID rows (có màu xanh giống headerStyle)
	idStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  10,
			Color: "#000000",
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
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Tạo style cho student rows (locked cho tên học sinh)
	studentStyleLocked, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Protection: &excelize.Protection{
			Locked: true, // Lock để bảo vệ
		},
	})

	// Đã bỏ style cho cột lớp vì không còn cột lớp

	// Tạo style cho tên học sinh với màu xám nhạt (cho hàng chẵn)
	studentStyleLockedGray, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"}, // Màu xám nhạt hơn
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Protection: &excelize.Protection{
			Locked: true, // Lock để bảo vệ
		},
	})

	// Tạo style cho cột "Tên khóa" - căn giữa, in đậm, màu trắng
	courseNameStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
			Bold: true,
		},
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
		Protection: &excelize.Protection{
			Locked: true, // Lock để bảo vệ
		},
	})

	// Tạo style cho cột "Tên khóa" - căn giữa, in đậm, màu xám nhạt (cho hàng chẵn)
	courseNameStyleGray, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"}, // Màu xám nhạt hơn
			Pattern: 1,
		},
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
		Protection: &excelize.Protection{
			Locked: true, // Lock để bảo vệ
		},
	})

	// Tính toán cấu trúc cột
	type ColumnInfo struct {
		CriterionID       int64
		CriterionName     string
		CriterionMaxScore float32
		Subcriteria       []struct {
			ID       int64
			Name     string
			MaxScore float32
		}
		HasSubcriteria bool // Đánh dấu có subcriteria thực sự hay không
		StartCol       int  // Cột bắt đầu (0-based)
		EndCol         int  // Cột kết thúc (0-based, exclusive)
	}

	var columns []ColumnInfo
	// Cột A, B: Metadata (sẽ bị ẩn)
	// Cột C: Tên khóa
	// Cột D: Tên học sinh
	// Cột E: Mã định danh
	// Cột F: Student ID (sẽ bị ẩn)
	// Từ cột G trở đi: Criteria và Subcriteria
	currentCol := 6 // Bắt đầu từ cột G (0=A, 1=B, 2=C, 3=D, 4=E, 5=F, 6=G)

	for _, ref := range assessment.AssessmentRefCriteria {
		if ref.AssessmentCriterion.ID == 0 {
			continue
		}

		criterion := ref.AssessmentCriterion
		colInfo := ColumnInfo{
			CriterionID:       criterion.ID,
			CriterionName:     criterion.Name,
			CriterionMaxScore: criterion.MaxScore,
			StartCol:          currentCol,
			HasSubcriteria:    false,
		}

		// Lấy subcriteria
		for _, sub := range criterion.AssessmentSubcriteria {
			if sub.ID > 0 {
				colInfo.HasSubcriteria = true
				colInfo.Subcriteria = append(colInfo.Subcriteria, struct {
					ID       int64
					Name     string
					MaxScore float32
				}{
					ID:       sub.ID,
					Name:     sub.Name,
					MaxScore: sub.MaxScore,
				})
			}
		}

		// Nếu không có subcriteria, tạo 1 cột cho criterion (nhưng không đánh dấu HasSubcriteria = true)
		if len(colInfo.Subcriteria) == 0 {
			colInfo.Subcriteria = append(colInfo.Subcriteria, struct {
				ID       int64
				Name     string
				MaxScore float32
			}{
				ID:       criterion.ID,
				Name:     criterion.Name,
				MaxScore: criterion.MaxScore,
			})
		}

		colInfo.EndCol = currentCol + len(colInfo.Subcriteria)
		currentCol = colInfo.EndCol
		columns = append(columns, colInfo)
	}

	// Tính điểm cuối cùng của các cột "Tiêu chí điểm"
	scoreEndCol := currentCol
	if len(columns) > 0 {
		lastCol := columns[len(columns)-1]
		scoreEndCol = lastCol.EndCol
	}

	// Tính toán cấu trúc cột cho "Tiêu chí đánh giá" (nếu có StudyReportCriteria)
	var evaluationColumns []EvaluationColumnInfo
	evaluationStartCol := scoreEndCol // Bắt đầu sau các cột "Tiêu chí điểm"

	if studyReportCriteria != nil {
		// Tách skills thành star và check
		var starSkills []models.StudyReportSkill
		var checkSkills []models.StudyReportSkill

		for _, skill := range studyReportCriteria.Skills {
			if skill.Type == models.StudyReportTypeStar {
				starSkills = append(starSkills, skill)
			} else if skill.Type == models.StudyReportTypeCheck {
				checkSkills = append(checkSkills, skill)
			}
		}

		// Xử lý star skills
		if len(starSkills) > 0 {
			// Cột lớn nhất: "Tiêu chí tặng sao"
			maxStarText := "Tiêu chí tặng sao"
			if studyReportCriteria.MaxStar != nil && *studyReportCriteria.MaxStar > 0 {
				maxStarText = fmt.Sprintf("Tiêu chí tặng sao (max: %d)", *studyReportCriteria.MaxStar)
			}

			starStartCol := evaluationStartCol
			starEndCol := starStartCol

			// Tính tổng số cột cho star skills (chỉ tính leaf nodes)
			for _, skill := range starSkills {
				typeTree := buildSkillTypeTreeFromFlat(skill.Types)
				flattenedTypes := flattenSkillTypes(typeTree)
				typeCount := len(flattenedTypes)
				if typeCount == 0 {
					typeCount = 1 // Nếu không có types, vẫn có 1 cột cho skill
				}
				starEndCol += typeCount
			}

			if starEndCol > starStartCol {
				evalCol := EvaluationColumnInfo{
					Type:      "star",
					SkillID:   0,
					SkillName: maxStarText,
					MaxStar:   studyReportCriteria.MaxStar,
					StartCol:  starStartCol,
					EndCol:    starEndCol,
					IsParent:  true,
				}
				evaluationColumns = append(evaluationColumns, evalCol)
				evaluationStartCol = starEndCol
			}

			// Thêm các cột cho từng star skill (skill cha merge các leaf nodes)
			currentStarCol := starStartCol
			for _, skill := range starSkills {
				skillStartCol := currentStarCol

				// Build tree từ flat list
				typeTree := buildSkillTypeTreeFromFlat(skill.Types)

				// Flatten để lấy leaf nodes
				flattenedTypes := flattenSkillTypes(typeTree)
				typeCount := len(flattenedTypes)

				if typeCount == 0 {
					typeCount = 1
					// Nếu không có types, tạo 1 type giả từ skill
					evalCol := EvaluationColumnInfo{
						Type:      "star",
						SkillID:   skill.ID,
						SkillName: skill.NameVn,
						StartCol:  skillStartCol,
						EndCol:    skillStartCol + 1,
						IsParent:  false,
						Types: []struct {
							ID   int64
							Name string
						}{
							{ID: skill.ID, Name: skill.NameVn},
						},
					}
					evaluationColumns = append(evaluationColumns, evalCol)
					currentStarCol = skillStartCol + 1
				} else {
					// Có types, tạo 1 cột cha cho skill merge các cột leaf nodes
					evalCol := EvaluationColumnInfo{
						Type:      "star",
						SkillID:   skill.ID,
						SkillName: skill.NameVn,
						StartCol:  skillStartCol,
						EndCol:    skillStartCol + typeCount,
						IsParent:  false,
						Types:     flattenedTypes,
						TypeTree:  typeTree, // Lưu tree để render phân cấp
					}
					evaluationColumns = append(evaluationColumns, evalCol)
					currentStarCol = skillStartCol + typeCount
				}
			}
		}

		// Xử lý check skills
		if len(checkSkills) > 0 {
			// Cột lớn nhất: "Tiêu chí check"
			checkStartCol := evaluationStartCol
			checkEndCol := checkStartCol

			// Tính tổng số cột cho check skills (chỉ tính leaf nodes)
			for _, skill := range checkSkills {
				typeTree := buildSkillTypeTreeFromFlat(skill.Types)
				flattenedTypes := flattenSkillTypes(typeTree)
				typeCount := len(flattenedTypes)
				if typeCount == 0 {
					typeCount = 1
				}
				checkEndCol += typeCount
			}

			if checkEndCol > checkStartCol {
				evalCol := EvaluationColumnInfo{
					Type:      "check",
					SkillID:   0,
					SkillName: "Tiêu chí check",
					StartCol:  checkStartCol,
					EndCol:    checkEndCol,
					IsParent:  true,
				}
				evaluationColumns = append(evaluationColumns, evalCol)
				evaluationStartCol = checkEndCol
			}

			// Thêm các cột cho từng check skill (skill cha merge các leaf nodes)
			currentCheckCol := checkStartCol
			for _, skill := range checkSkills {
				skillStartCol := currentCheckCol

				// Build tree từ flat list
				typeTree := buildSkillTypeTreeFromFlat(skill.Types)

				// Flatten để lấy leaf nodes
				flattenedTypes := flattenSkillTypes(typeTree)
				typeCount := len(flattenedTypes)

				if typeCount == 0 {
					typeCount = 1
					// Nếu không có types, tạo 1 type giả từ skill
					evalCol := EvaluationColumnInfo{
						Type:      "check",
						SkillID:   skill.ID,
						SkillName: skill.NameVn,
						StartCol:  skillStartCol,
						EndCol:    skillStartCol + 1,
						IsParent:  false,
						Types: []struct {
							ID   int64
							Name string
						}{
							{ID: skill.ID, Name: skill.NameVn},
						},
					}
					evaluationColumns = append(evaluationColumns, evalCol)
					currentCheckCol = skillStartCol + 1
				} else {
					// Có types, tạo 1 cột cha cho skill merge các cột leaf nodes
					evalCol := EvaluationColumnInfo{
						Type:      "check",
						SkillID:   skill.ID,
						SkillName: skill.NameVn,
						StartCol:  skillStartCol,
						EndCol:    skillStartCol + typeCount,
						IsParent:  false,
						Types:     flattenedTypes,
						TypeTree:  typeTree, // Lưu tree để render phân cấp
					}
					evaluationColumns = append(evaluationColumns, evalCol)
					currentCheckCol = skillStartCol + typeCount
				}
			}
		}
	}

	// Cập nhật currentCol để bao gồm cả các cột "Tiêu chí đánh giá"
	evaluationEndCol := scoreEndCol
	if len(evaluationColumns) > 0 {
		lastEvalCol := evaluationColumns[len(evaluationColumns)-1]
		evaluationEndCol = lastEvalCol.EndCol
	}
	currentCol = evaluationEndCol

	// Thêm cột "Nhận xét" ở cuối cùng
	commentCol := currentCol + 1 // Cột sau các cột tiêu chí

	// Hàng 1: "Tiêu chí điểm" - merge từ cột G (7) đến cột cuối cùng của tiêu chí điểm
	if scoreEndCol > 6 {
		startCell := "G1" // Cột G (7)
		endCell := getCellName(1, scoreEndCol)
		f.SetCellValue(sheetName, startCell, "Tiêu chí điểm")
		// Apply style cho tất cả các ô trong range (màu #b6d7a8)
		for c := 7; c <= scoreEndCol; c++ {
			cell := getCellName(1, c)
			f.SetCellStyle(sheetName, cell, cell, scoreHeaderStyle)
		}
		f.MergeCell(sheetName, startCell, endCell)
	}

	// Hàng 1: "Tiêu chí đánh giá" - merge từ cột sau tiêu chí điểm đến cột cuối cùng
	if len(evaluationColumns) > 0 && evaluationEndCol > scoreEndCol {
		evalStartCol := scoreEndCol + 1
		startCell := getCellName(1, evalStartCol)
		endCell := getCellName(1, evaluationEndCol)
		f.SetCellValue(sheetName, startCell, "Tiêu chí đánh giá")
		// Apply style cho tất cả các ô trong range (màu #ffe599)
		for c := evalStartCol; c <= evaluationEndCol; c++ {
			cell := getCellName(1, c)
			f.SetCellStyle(sheetName, cell, cell, evalHeaderStyle)
		}
		f.MergeCell(sheetName, startCell, endCell)
	}

	// Header cho các cột ở hàng 1 (bắt đầu từ cột C)
	// Cột C: Tên khóa
	f.SetCellValue(sheetName, "C1", "Tên khóa")
	// Cột D: Tên học sinh
	f.SetCellValue(sheetName, "D1", "Tên học sinh")
	// Cột E: Mã định danh
	f.SetCellValue(sheetName, "E1", "Mã định danh")
	// Cột F: Student ID (sẽ bị ẩn)
	f.SetCellValue(sheetName, "F1", "Student ID")

	// (Sẽ được merge xuống lastHeaderRow sau khi tính toán được lastHeaderRow)

	// Cột "Nhận xét" - sẽ được merge xuống lastHeaderRow sau

	// Hàng 2: Tên các tiêu chí cha (kèm max_score) - text màu đỏ
	for _, col := range columns {
		if col.StartCol < col.EndCol {
			// StartCol là 0-based, Excel là 1-based, nên +1
			startCell := getCellName(2, col.StartCol+1)
			// EndCol là exclusive (0-based), nên cell cuối cùng trong range là EndCol (1-based)
			// Ví dụ: StartCol=2, EndCol=5 (exclusive) -> range là [2,3,4] -> endCell là col 5 (1-based) = E2
			endCell := getCellName(2, col.EndCol)
			// Tạo tên tiêu chí kèm max_score
			criterionNameWithScore := col.CriterionName
			if col.CriterionMaxScore > 0 {
				criterionNameWithScore = fmt.Sprintf("%s (%.2f)", col.CriterionName, col.CriterionMaxScore)
			}
			f.SetCellValue(sheetName, startCell, criterionNameWithScore)
			// Apply style với text màu đỏ và nền #d9ead3
			scoreHeaderRow2Style, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold:  true,
					Color: "#ff0000",
					Size:  12,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#d9ead3"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Protection: &excelize.Protection{
					Locked: true,
				},
			})
			for c := col.StartCol; c < col.EndCol; c++ {
				cell := getCellName(2, c+1)
				f.SetCellStyle(sheetName, cell, cell, scoreHeaderRow2Style)
			}
			// Merge cells nếu có nhiều hơn 1 cột
			if col.EndCol-col.StartCol > 1 {
				f.MergeCell(sheetName, startCell, endCell)
			}
		}
	}

	// Hàng 3: ID các tiêu chí cha (sẽ bị ẩn)
	for _, col := range columns {
		if col.StartCol < col.EndCol {
			startCell := getCellName(3, col.StartCol+1)
			endCell := getCellName(3, col.EndCol)
			f.SetCellValue(sheetName, startCell, col.CriterionID)
			// Apply style to all cells in range trước
			for c := col.StartCol; c < col.EndCol; c++ {
				cell := getCellName(3, c+1)
				f.SetCellStyle(sheetName, cell, cell, idStyle)
			}
			// Merge cells nếu có nhiều hơn 1 cột
			if col.EndCol-col.StartCol > 1 {
				f.MergeCell(sheetName, startCell, endCell)
			}
		}
	}

	// Hàng 4: Tên các tiêu chí con (kèm max_score) - chỉ render khi có subcriteria thực sự
	// Text màu đỏ, nền #d9ead3
	for _, col := range columns {
		// Chỉ render hàng 4 nếu có subcriteria thực sự
		if col.HasSubcriteria {
			for i, sub := range col.Subcriteria {
				cell := getCellName(4, col.StartCol+i+1)
				// Tạo tên tiêu chí con kèm max_score
				subNameWithScore := sub.Name
				if sub.MaxScore > 0 {
					subNameWithScore = fmt.Sprintf("%s (%.2f)", sub.Name, sub.MaxScore)
				}
				f.SetCellValue(sheetName, cell, subNameWithScore)
				// Style với text màu đỏ, nền #d9ead3
				subHeaderRow4Style, _ := f.NewStyle(&excelize.Style{
					Font: &excelize.Font{
						Bold:  true,
						Color: "#ff0000",
						Size:  12,
					},
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{"#d9ead3"},
						Pattern: 1,
					},
					Alignment: &excelize.Alignment{
						Horizontal: "center",
						Vertical:   "center",
						WrapText:   true,
					},
					Border: []excelize.Border{
						{Type: "left", Color: "000000", Style: 1},
						{Type: "top", Color: "000000", Style: 1},
						{Type: "bottom", Color: "000000", Style: 1},
						{Type: "right", Color: "000000", Style: 1},
					},
					Protection: &excelize.Protection{
						Locked: true,
					},
				})
				f.SetCellStyle(sheetName, cell, cell, subHeaderRow4Style)
			}
		}
	}

	// Hàng 5: ID các tiêu chí con (sẽ bị ẩn) - chỉ render khi có subcriteria thực sự
	for _, col := range columns {
		// Chỉ render hàng 5 nếu có subcriteria thực sự
		if col.HasSubcriteria {
			for i, sub := range col.Subcriteria {
				cell := getCellName(5, col.StartCol+i+1)
				f.SetCellValue(sheetName, cell, sub.ID)
				f.SetCellStyle(sheetName, cell, cell, idStyle)
			}
		}
	}

	// Render các hàng cho "Tiêu chí đánh giá"
	// Tìm các cột lớn nhất (parent columns) cho star và check
	var _ *EvaluationColumnInfo
	var _ *EvaluationColumnInfo
	for i := range evaluationColumns {
		if evaluationColumns[i].IsParent && evaluationColumns[i].Type == "star" {
			_ = &evaluationColumns[i]
		}
		if evaluationColumns[i].IsParent && evaluationColumns[i].Type == "check" {
			_ = &evaluationColumns[i]
		}
	}

	// Hàng 2: Tên các skill cha (merge các types con) - text màu đỏ
	// KHÔNG render starParentCol và checkParentCol ở hàng 2 vì chúng sẽ merge và che mất các skill cha
	for _, evalCol := range evaluationColumns {
		if !evalCol.IsParent && evalCol.SkillID > 0 && (len(evalCol.Types) > 0 || len(evalCol.TypeTree) > 0) {
			startCell := getCellName(2, evalCol.StartCol+1)
			endCell := getCellName(2, evalCol.EndCol)
			f.SetCellValue(sheetName, startCell, evalCol.SkillName)
			// Tạo style với text màu đỏ
			var bgColor string
			if evalCol.Type == "star" {
				// Star skill: màu xen kẽ, bắt đầu với màu 1
				bgColor = "#fce5cd"
			} else if evalCol.Type == "check" {
				// Check skill: màu #d0e0e3
				bgColor = "#d0e0e3"
			} else {
				bgColor = "#4F81BD"
			}
			evalHeaderRow2Style, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold:  true,
					Color: "#ff0000",
					Size:  12,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{bgColor},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Protection: &excelize.Protection{
					Locked: true,
				},
			})
			for c := evalCol.StartCol; c < evalCol.EndCol; c++ {
				cell := getCellName(2, c+1)
				f.SetCellStyle(sheetName, cell, cell, evalHeaderRow2Style)
			}
			if evalCol.EndCol-evalCol.StartCol > 1 {
				f.MergeCell(sheetName, startCell, endCell)
			}
		}
	}

	// Hàng 3: ID các skill cha (sẽ bị ẩn)
	// KHÔNG render starParentCol và checkParentCol ở hàng 3 vì chúng sẽ merge và che mất các skill cha
	for _, evalCol := range evaluationColumns {
		if !evalCol.IsParent && evalCol.SkillID > 0 && (len(evalCol.Types) > 0 || len(evalCol.TypeTree) > 0) {
			startCell := getCellName(3, evalCol.StartCol+1)
			endCell := getCellName(3, evalCol.EndCol)
			f.SetCellValue(sheetName, startCell, evalCol.SkillID)
			for c := evalCol.StartCol; c < evalCol.EndCol; c++ {
				cell := getCellName(3, c+1)
				f.SetCellStyle(sheetName, cell, cell, idStyle)
			}
			if evalCol.EndCol-evalCol.StartCol > 1 {
				f.MergeCell(sheetName, startCell, endCell)
			}
		}
	}

	// Tính max depth và render phân cấp theo hàng
	maxDepth := 0
	for _, evalCol := range evaluationColumns {
		if !evalCol.IsParent && len(evalCol.TypeTree) > 0 {
			depth := getMaxDepth(evalCol.TypeTree)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	// Tạo map để lưu màu cho mỗi skill (để đảm bảo màu nhất quán giữa các hàng)
	skillColorMap := make(map[int64]string)
	starSkillIndexForLevels := 0
	for _, evalCol := range evaluationColumns {
		if !evalCol.IsParent && evalCol.SkillID > 0 {
			if evalCol.Type == "star" {
				// Star skill: xen kẽ màu #fce5cd và #fff2cc
				if starSkillIndexForLevels%2 == 0 {
					skillColorMap[evalCol.SkillID] = "#fce5cd"
				} else {
					skillColorMap[evalCol.SkillID] = "#fff2cc"
				}
				starSkillIndexForLevels++
			} else if evalCol.Type == "check" {
				// Check skill: màu #d0e0e3
				skillColorMap[evalCol.SkillID] = "#d0e0e3"
			}
		}
	}

	// Render các hàng phân cấp (mỗi level một hàng)
	currentRow := 4
	for _, evalCol := range evaluationColumns {
		if !evalCol.IsParent && evalCol.SkillID > 0 {
			if evalCol.Type == "star" {
				// Star skill: xen kẽ màu #fce5cd và #fff2cc
				if starSkillIndexForLevels%2 == 0 {
					skillColorMap[evalCol.SkillID] = "#fce5cd"
				} else {
					skillColorMap[evalCol.SkillID] = "#fff2cc"
				}
				starSkillIndexForLevels++
			} else if evalCol.Type == "check" {
				// Check skill: màu #d0e0e3
				skillColorMap[evalCol.SkillID] = "#d0e0e3"
			}
		}
	}

	for level := 1; level <= maxDepth; level++ {
		// Hàng tên (chẵn): Tên các types ở level này
		for _, evalCol := range evaluationColumns {
			if !evalCol.IsParent && len(evalCol.TypeTree) > 0 {
				// Lấy màu từ map (đã tạo ở trên) hoặc mặc định
				bgColor := skillColorMap[evalCol.SkillID]
				if bgColor == "" {
					bgColor = "#4F81BD"
				}

				// Tạo style với text màu đỏ cho hàng tên
				levelHeaderStyle, _ := f.NewStyle(&excelize.Style{
					Font: &excelize.Font{
						Bold:  true,
						Color: "#ff0000",
						Size:  12,
					},
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{bgColor},
						Pattern: 1,
					},
					Alignment: &excelize.Alignment{
						Horizontal: "center",
						Vertical:   "center",
						WrapText:   true,
					},
					Border: []excelize.Border{
						{Type: "left", Color: "000000", Style: 1},
						{Type: "top", Color: "000000", Style: 1},
						{Type: "bottom", Color: "000000", Style: 1},
						{Type: "right", Color: "000000", Style: 1},
					},
					Protection: &excelize.Protection{
						Locked: true,
					},
				})
				renderLevelTypes(f, sheetName, evalCol, level, currentRow, levelHeaderStyle, true)
			}
		}
		currentRow++

		// Hàng ID (lẻ): ID các types ở level này (sẽ bị ẩn)
		for _, evalCol := range evaluationColumns {
			if !evalCol.IsParent && len(evalCol.TypeTree) > 0 {
				// Lấy màu từ map hoặc mặc định
				bgColor := skillColorMap[evalCol.SkillID]
				if bgColor == "" {
					bgColor = "#4F81BD"
				}

				// Tạo style cho hàng ID
				levelIdStyle, _ := f.NewStyle(&excelize.Style{
					Font: &excelize.Font{
						Size:  10,
						Color: "#000000",
					},
					Fill: excelize.Fill{
						Type:    "pattern",
						Color:   []string{bgColor},
						Pattern: 1,
					},
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
				renderLevelTypes(f, sheetName, evalCol, level, currentRow, levelIdStyle, false)
			}
		}
		currentRow++
	}

	// Từ hàng sau các hàng phân cấp: Danh sách học sinh
	startRow := currentRow
	lastHeaderRow := startRow - 1 // Hàng cuối cùng của header (trước phần nhập điểm)

	// Thêm metadata vào cột A và B (ẩn)
	f.SetCellValue(sheetName, "A1", "course_id")
	f.SetCellValue(sheetName, "B1", course.ID)
	f.SetCellValue(sheetName, "A2", "assessment_id")
	f.SetCellValue(sheetName, "B2", assessment.ID)
	f.SetCellValue(sheetName, "A3", "start_row")
	f.SetCellValue(sheetName, "B3", startRow)

	// Ẩn cột A và B
	f.SetColVisible(sheetName, "A", false)
	f.SetColVisible(sheetName, "B", false)

	// Merge các cột C, D, E, F trong phần header (từ hàng 1 đến lastHeaderRow)
	if lastHeaderRow >= 1 {
		// Cột C: Tên khóa - merge trong header (màu xanh nhạt)
		cellCStart := "C1"
		cellCEnd := getCellName(lastHeaderRow, 3)
		for r := 1; r <= lastHeaderRow; r++ {
			cell := getCellName(r, 3)
			f.SetCellStyle(sheetName, cell, cell, basicHeaderStyle)
		}
		if lastHeaderRow > 1 {
			f.MergeCell(sheetName, cellCStart, cellCEnd)
		}

		// Cột D: Tên học sinh - merge trong header (màu xanh nhạt)
		cellDStart := "D1"
		cellDEnd := getCellName(lastHeaderRow, 4)
		for r := 1; r <= lastHeaderRow; r++ {
			cell := getCellName(r, 4)
			f.SetCellStyle(sheetName, cell, cell, basicHeaderStyle)
		}
		if lastHeaderRow > 1 {
			f.MergeCell(sheetName, cellDStart, cellDEnd)
		}

		// Cột E: Mã định danh - merge trong header (màu xanh nhạt)
		cellEStart := "E1"
		cellEEnd := getCellName(lastHeaderRow, 5)
		for r := 1; r <= lastHeaderRow; r++ {
			cell := getCellName(r, 5)
			f.SetCellStyle(sheetName, cell, cell, basicHeaderStyle)
		}
		if lastHeaderRow > 1 {
			f.MergeCell(sheetName, cellEStart, cellEEnd)
		}

		// Cột F: Student ID (sẽ bị ẩn) - merge trong header (màu xanh nhạt)
		cellFStart := "F1"
		cellFEnd := getCellName(lastHeaderRow, 6)
		for r := 1; r <= lastHeaderRow; r++ {
			cell := getCellName(r, 6)
			f.SetCellStyle(sheetName, cell, cell, basicHeaderStyle)
		}
		if lastHeaderRow > 1 {
			f.MergeCell(sheetName, cellFStart, cellFEnd)
		}
	}

	// Áp dụng màu cho các cột header (trừ các ô đã có style riêng)
	for r := 1; r <= lastHeaderRow; r++ {
		// Các cột dưới "Tiêu chí điểm": màu #d9ead3
		for c := 7; c <= scoreEndCol; c++ {
			cell := getCellName(r, c)
			// Kiểm tra xem cell đã có style chưa (hàng 2 có style riêng)
			if r == 2 {
				// Hàng 2 đã có style riêng với text màu đỏ
				continue
			}
			// Tạo style với nền #d9ead3
			scoreHeaderCellStyle, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold:  true,
					Color: "#000000",
					Size:  12,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#d9ead3"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Protection: &excelize.Protection{
					Locked: true,
				},
			})
			f.SetCellStyle(sheetName, cell, cell, scoreHeaderCellStyle)
		}

		// Các cột dưới "Tiêu chí đánh giá": áp dụng màu theo loại skill
		if len(evaluationColumns) > 0 && evaluationEndCol > scoreEndCol {
			starSkillIndex := 0
			for _, evalCol := range evaluationColumns {
				if !evalCol.IsParent && evalCol.SkillID > 0 {
					// Xác định màu dựa trên loại skill
					var bgColor string
					if evalCol.Type == "star" {
						// Star skill: xen kẽ màu #fce5cd và #fff2cc
						if starSkillIndex%2 == 0 {
							bgColor = "#fce5cd"
						} else {
							bgColor = "#fff2cc"
						}
						starSkillIndex++
					} else if evalCol.Type == "check" {
						// Check skill: màu #d0e0e3
						bgColor = "#d0e0e3"
					} else {
						bgColor = "#4F81BD"
					}

					// Áp dụng cho các cột của skill này
					for c := evalCol.StartCol; c < evalCol.EndCol; c++ {
						cell := getCellName(r, c+1)
						if r == 2 {
							// Hàng 2 đã có style riêng với text màu đỏ
							continue
						}
						evalHeaderCellStyle, _ := f.NewStyle(&excelize.Style{
							Font: &excelize.Font{
								Bold:  true,
								Color: "#000000",
								Size:  12,
							},
							Fill: excelize.Fill{
								Type:    "pattern",
								Color:   []string{bgColor},
								Pattern: 1,
							},
							Alignment: &excelize.Alignment{
								Horizontal: "center",
								Vertical:   "center",
								WrapText:   true,
							},
							Border: []excelize.Border{
								{Type: "left", Color: "000000", Style: 1},
								{Type: "top", Color: "000000", Style: 1},
								{Type: "bottom", Color: "000000", Style: 1},
								{Type: "right", Color: "000000", Style: 1},
							},
							Protection: &excelize.Protection{
								Locked: true,
							},
						})
						f.SetCellStyle(sheetName, cell, cell, evalHeaderCellStyle)
					}
				}
			}
		}

		// Cột "Nhận xét": màu xanh nhạt (giống các tiêu đề cũ)
		commentCell := getCellName(r, commentCol)
		if r == 2 {
			// Hàng 2: text màu đỏ, nền xanh nhạt
			commentHeaderRow2Style, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold:  true,
					Color: "#ff0000",
					Size:  12,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#B4C6E7"}, // Màu xanh nhạt
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Protection: &excelize.Protection{
					Locked: true,
				},
			})
			f.SetCellStyle(sheetName, commentCell, commentCell, commentHeaderRow2Style)
		} else {
			// Các hàng khác: màu xanh nhạt
			f.SetCellStyle(sheetName, commentCell, commentCell, basicHeaderStyle)
		}
	}

	// Merge cột "Nhận xét" từ hàng 1 đến lastHeaderRow
	if lastHeaderRow >= 1 {
		commentCellStart := getCellName(1, commentCol)
		commentCellEnd := getCellName(lastHeaderRow, commentCol)
		f.SetCellValue(sheetName, commentCellStart, "Nhận xét")
		// Style đã được áp dụng trong vòng lặp ở trên
		if lastHeaderRow > 1 {
			f.MergeCell(sheetName, commentCellStart, commentCellEnd)
		}
	}

	// Không cần merge cột lớp nữa vì đã bỏ 2 cột lớp

	for i, student := range students {
		row := startRow + i

		// Xác định style dựa trên số thứ tự hàng (hàng chẵn = xám, hàng lẻ = trắng)
		// i bắt đầu từ 0, nên hàng đầu tiên (i=0) là lẻ, hàng thứ 2 (i=1) là chẵn
		isEvenRow := i%2 == 1
		lockedStudentStyle := studentStyleLocked
		if isEvenRow {
			lockedStudentStyle = studentStyleLockedGray
		}

		// Cột C: Tên khóa - sẽ merge sau
		// Cột D: Tên học sinh - tô màu xám cách dòng
		cellD := getCellName(row, 4)
		f.SetCellValue(sheetName, cellD, student.Name)
		f.SetCellStyle(sheetName, cellD, cellD, lockedStudentStyle)

		// Cột E: Mã định danh - tô màu xám cách dòng
		cellE := getCellName(row, 5)
		f.SetCellValue(sheetName, cellE, student.IdentifierCode)
		f.SetCellStyle(sheetName, cellE, cellE, lockedStudentStyle)

		// Cột F: Student ID (sẽ bị ẩn)
		cellF := getCellName(row, 6)
		f.SetCellValue(sheetName, cellF, student.ID)
		f.SetCellStyle(sheetName, cellF, cellF, lockedStudentStyle)

		// Các cột tiêu chí điểm để trống (để nhập điểm) - màu #d9ead3
		for col := 7; col <= scoreEndCol; col++ {
			cell := getCellName(row, col)
			// Tạo style với nền #d9ead3
			scoreDataStyle, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Size: 11,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#d9ead3"},
					Pattern: 1,
				},
				Alignment: &excelize.Alignment{
					Vertical: "center",
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 1},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
				Protection: &excelize.Protection{
					Locked: false, // Unlock để nhập điểm
				},
			})
			f.SetCellStyle(sheetName, cell, cell, scoreDataStyle)
		}

		// Các cột tiêu chí đánh giá
		for _, evalCol := range evaluationColumns {
			if !evalCol.IsParent && evalCol.SkillID > 0 && len(evalCol.Types) > 0 {
				// Lấy màu từ map (đã tạo ở trên) hoặc mặc định
				bgColor := skillColorMap[evalCol.SkillID]
				if bgColor == "" {
					bgColor = "#000000"
				}

				// Chỉ render các cột có SkillID (không phải parent lớn nhất)
				for i := range evalCol.Types {
					cell := getCellName(row, evalCol.StartCol+i+1)
					// Tạo style với màu nền tương ứng
					evalDataStyle, _ := f.NewStyle(&excelize.Style{
						Font: &excelize.Font{
							Size: 11,
						},
						Fill: excelize.Fill{
							Type:    "pattern",
							Color:   []string{bgColor},
							Pattern: 1,
						},
						Alignment: &excelize.Alignment{
							Vertical: "center",
						},
						Border: []excelize.Border{
							{Type: "left", Color: "000000", Style: 1},
							{Type: "top", Color: "000000", Style: 1},
							{Type: "bottom", Color: "000000", Style: 1},
							{Type: "right", Color: "000000", Style: 1},
						},
						Protection: &excelize.Protection{
							Locked: false,
						},
					})

					if evalCol.Type == "check" {
						// Thêm checkbox cho các ô check - sử dụng data validation với list
						dv := excelize.NewDataValidation(true)
						dv.Sqref = cell
						dv.SetDropList([]string{"☐", "☑"})
						dv.ShowInputMessage = false
						dv.ShowErrorMessage = false
						if err := f.AddDataValidation(sheetName, dv); err != nil {
							// Nếu không thể thêm data validation, set giá trị mặc định
							f.SetCellValue(sheetName, cell, "☐")
						} else {
							f.SetCellValue(sheetName, cell, "☐")
						}
					}
					f.SetCellStyle(sheetName, cell, cell, evalDataStyle)
				}
			}
		}

		// Cột "Nhận xét" - màu #ead1dc
		commentCell := getCellName(row, commentCol)
		commentDataStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Size: 11,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#ead1dc"},
				Pattern: 1,
			},
			Alignment: &excelize.Alignment{
				Vertical: "center",
			},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
			Protection: &excelize.Protection{
				Locked: false,
			},
		})
		f.SetCellStyle(sheetName, commentCell, commentCell, commentDataStyle)
	}

	// Merge cột "Tên khóa" (C) từ hàng đầu tiên học sinh đến hàng cuối cùng
	if len(students) > 0 {
		firstStudentRow := startRow
		lastStudentRow := startRow + len(students) - 1
		cellCStart := getCellName(firstStudentRow, 3)
		cellCEnd := getCellName(lastStudentRow, 3)
		f.SetCellValue(sheetName, cellCStart, course.Name)
		// Áp dụng style cho tất cả các ô trong range (căn giữa, in đậm)
		for r := firstStudentRow; r <= lastStudentRow; r++ {
			cell := getCellName(r, 3)
			// Xác định style dựa trên số thứ tự hàng
			isEvenRow := (r-firstStudentRow)%2 == 1
			styleToUse := courseNameStyle
			if isEvenRow {
				styleToUse = courseNameStyleGray
			}
			f.SetCellStyle(sheetName, cell, cell, styleToUse)
		}
		if lastStudentRow > firstStudentRow {
			f.MergeCell(sheetName, cellCStart, cellCEnd)
		}
	}

	// Cột "Nhận xét" không cần merge, mỗi học sinh một ô riêng (giống cột "Tên học sinh")
	// Style đã được áp dụng trong vòng lặp render học sinh ở trên

	// Set column widths
	f.SetColWidth(sheetName, "C", "C", 25) // Tên khóa
	f.SetColWidth(sheetName, "D", "D", 25) // Tên học sinh
	f.SetColWidth(sheetName, "E", "E", 20) // Mã định danh
	f.SetColWidth(sheetName, "F", "F", 12) // Student ID (sẽ bị ẩn)
	// Set width cho các cột criteria (mỗi cột 15)
	for col := 7; col < currentCol; col++ {
		colName := getColumnName(col)
		f.SetColWidth(sheetName, colName, colName, 15)
	}
	// Set width cho cột "Nhận xét" (rộng hơn)
	commentColName := getColumnName(commentCol)
	f.SetColWidth(sheetName, commentColName, commentColName, 40)

	// Ẩn các hàng ID (hàng lẻ sau hàng 1)
	// Hàng 3: Parent criteria IDs
	f.SetRowVisible(sheetName, 3, false)
	// Hàng 5: Sub-criteria IDs
	f.SetRowVisible(sheetName, 5, false)
	// Ẩn các hàng ID của tiêu chí đánh giá (hàng lẻ sau hàng 4)
	// Chỉ ẩn các hàng ID trong phần header (trước startRow)
	if maxDepth > 0 {
		hideRow := 5 // Bắt đầu từ hàng 5 (sau hàng 4)
		for level := 1; level <= maxDepth; level++ {
			hideRow += 1 // Hàng tên
			hideRow += 1 // Hàng ID - ẩn hàng này
			// Chỉ ẩn nếu hàng này vẫn trong phần header (nhỏ hơn startRow)
			if hideRow < startRow {
				f.SetRowVisible(sheetName, hideRow, false)
			}
		}
	}

	// Ẩn cột F (Student ID)
	f.SetColVisible(sheetName, "F", false)

	// Freeze panes: cố định các cột C-F và các hàng tiêu chí
	// Freeze tại cell G{startRow} để:
	// - Các cột C-F (Tên khóa, Tên học sinh, Mã định danh, Student ID) được cố định
	// - Các hàng tiêu chí (từ hàng 1 đến startRow-1) được cố định
	// - Chỉ các cột từ G trở đi và các hàng từ startRow trở đi mới scroll
	freezeCell := getCellName(startRow, 7) // Cột G (7), hàng startRow
	// Freeze panes: XSplit = 6 (cố định 6 cột đầu: A, B, C, D, E, F - nhưng A, B, F bị ẩn)
	// YSplit = startRow - 1 (cố định các hàng từ 1 đến startRow-1)
	// TopLeftCell = G{startRow} (cell bắt đầu scroll)
	if err := f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      startRow - 1,  // Số hàng được freeze (từ hàng 1 đến startRow-1)
		XSplit:      6,             // Số cột được freeze (6 cột đầu: A, B, C, D, E, F)
		TopLeftCell: freezeCell,    // Cell bắt đầu scroll (G{startRow})
		ActivePane:  "bottomRight", // Pane hoạt động là bottomRight (phần scroll)
	}); err != nil {
		return fmt.Errorf("failed to set freeze panes: %w", err)
	}

	// Bảo vệ sheet với mật khẩu "vandz"
	// Mặc định tất cả các ô đã được lock, chỉ các ô có Protection.Locked = false mới có thể chỉnh sửa
	protectOpts := excelize.SheetProtectionOptions{
		AlgorithmName:       "SHA-512",
		Password:            "vandz",
		SelectLockedCells:   true,
		SelectUnlockedCells: true,
		FormatCells:         false,
		FormatColumns:       false,
		FormatRows:          false,
		InsertColumns:       false,
		InsertRows:          false,
		InsertHyperlinks:    false,
		DeleteColumns:       false,
		DeleteRows:          false,
		Sort:                false,
		AutoFilter:          false,
		PivotTables:         false,
		EditObjects:         false,
		EditScenarios:       false,
	}
	if err := f.ProtectSheet(sheetName, &protectOpts); err != nil {
		return fmt.Errorf("failed to protect sheet: %w", err)
	}

	return nil
}

// Helper functions
func getCellName(row, col int) string {
	colName := getColumnName(col)
	return fmt.Sprintf("%s%d", colName, row)
}

func getColumnName(col int) string {
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

// ImportExcel nhập file Excel đã điền dữ liệu và tạo assessment_scores + study_reports
func (s *assessmentExportService) ImportExcel(c *gin.Context, fileReader io.Reader) (interface{}, error) {
	// Mở file Excel
	f, err := excelize.OpenReader(fileReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read excel file: %w", err)
	}
	defer f.Close()

	// Lấy danh sách sheet names
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("no sheets found in excel file")
	}

	userID := int64(utils.GetCurrentUserId(c))
	now := time.Now().UTC()

	type importResult struct {
		SheetName       string `json:"sheet_name"`
		Processed       int    `json:"processed"`
		Success         int    `json:"success"`
		Failed          int    `json:"failed"`
		Errors          []string `json:"errors,omitempty"`
	}

	results := make([]importResult, 0)

	// Xử lý từng sheet
	for _, sheetName := range sheetList {
		result := importResult{
			SheetName: sheetName,
			Errors:    make([]string, 0),
		}

		// 1. Đọc metadata từ cột A, B (hàng 1-3)
		// A1="course_id", B1=course_id
		// A2="assessment_id", B2=assessment_id
		// A3="start_row", B3=start_row
		courseIDStr, err := f.GetCellValue(sheetName, "B1")
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot read course_id: %v", err))
			results = append(results, result)
			continue
		}
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil || courseID <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid course_id: %s", courseIDStr))
			results = append(results, result)
			continue
		}

		assessmentIDStr, err := f.GetCellValue(sheetName, "B2")
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot read assessment_id: %v", err))
			results = append(results, result)
			continue
		}
		assessmentID, err := strconv.ParseInt(assessmentIDStr, 10, 64)
		if err != nil || assessmentID <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid assessment_id: %s", assessmentIDStr))
			results = append(results, result)
			continue
		}

		startRowStr, err := f.GetCellValue(sheetName, "B3")
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot read start_row: %v", err))
			results = append(results, result)
			continue
		}
		startRow, err := strconv.Atoi(startRowStr)
		if err != nil || startRow <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("invalid start_row: %s", startRowStr))
			results = append(results, result)
			continue
		}

		// 2. Lấy assessment với criteria
		assessment, err := s.repo.GetAssessmentWithCriteria(assessmentID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot get assessment: %v", err))
			results = append(results, result)
			continue
		}

		// 3. Lấy lesson_id từ assessment_ref_lessons
		lessonID, err := s.repo.GetLessonIDByAssessmentAndCourse(assessmentID, courseID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot get lesson_id: %v", err))
			results = append(results, result)
			continue
		}
		if lessonID <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("lesson_id not found for assessment_id=%d, course_id=%d", assessmentID, courseID))
			results = append(results, result)
			continue
		}

		// 4. Lấy subject_id từ assessment
		subjectID := assessment.SubjectID
		if subjectID <= 0 {
			// Nếu không có subject_id trong assessment, lấy từ lesson (lesson -> chapter -> program -> program_ref_subjects)
			var result struct {
				SubjectID int64
			}
			err := db.ReplicaDB.Table("lessons l").
				Joins("JOIN chapters c ON c.id = l.chapter_id").
				Joins("JOIN programs p ON p.id = c.program_id").
				Joins("JOIN program_ref_subjects prs ON prs.program_id = p.id").
				Where("l.id = ? AND l.deleted_at IS NULL", lessonID).
				Select("prs.subject_id AS subject_id").
				Limit(1).
				Scan(&result).Error
			if err == nil && result.SubjectID > 0 {
				subjectID = result.SubjectID
			}
		}

		// 5. Lấy StudyReportCriteria nếu có
		var studyReportCriteria *models.StudyReportCriteria
		if assessment.StudyReportCriteriaID > 0 {
			studyReportCriteria, err = s.repo.GetStudyReportCriteria(assessment.StudyReportCriteriaID)
			if err != nil {
				// Không bắt buộc phải có
				studyReportCriteria = nil
			}
		}

		// 6. Map các cột với criteria/subcriteria IDs (từ hàng 2-5)
		// Hàng 2: Tên criteria (merge cells) -> hàng 3: criterion_id
		// Hàng 4: Tên subcriteria -> hàng 5: subcriterion_id
		columnToSubcriterionMap := make(map[int]struct {
			CriterionID    int64
			SubcriterionID *int64
		})

		// Đọc từ hàng 3 để lấy criterion_id và từ hàng 5 để lấy subcriterion_id
		rows, err := f.GetRows(sheetName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("cannot read rows: %v", err))
			results = append(results, result)
			continue
		}

		// Map cột với criterion/subcriterion (cột F là 6, bắt đầu từ cột 7)
		// Hàng 3 (index 2): Parent criteria IDs
		// Hàng 5 (index 4): Sub-criteria IDs
		if len(rows) > 4 {
			// Hàng 3: Parent criteria IDs
			row3 := rows[2] // Index 2 = hàng 3
			// Hàng 5: Sub-criteria IDs (chỉ có khi có subcriteria)
			row5 := rows[4] // Index 4 = hàng 5

			// Duyệt từ cột 7 trở đi (index 6)
			// Lưu ý: Khi có subcriteria, một criterion có thể merge nhiều cột (hàng 3),
			// nhưng mỗi cột có subcriterion_id riêng (hàng 5)
			// Cần map tất cả các cột, kể cả các cột merge
			var currentCriterionID int64 = 0
			for colIdx := 6; colIdx < len(row3); colIdx++ {
				// Lấy criterion_id từ hàng 3 (có thể là merge cell nên có thể trống)
				criterionIDStr := row3[colIdx]
				if criterionIDStr != "" {
					criterionID, err := strconv.ParseInt(criterionIDStr, 10, 64)
					if err == nil && criterionID > 0 {
						// Kiểm tra criterion này có trong assessment không
						validCriterion := false
						for _, ref := range assessment.AssessmentRefCriteria {
							if ref.AssessmentCriterion.ID == criterionID {
								validCriterion = true
								break
							}
						}
						if validCriterion {
							currentCriterionID = criterionID
						} else {
							currentCriterionID = 0
						}
					} else {
						currentCriterionID = 0
					}
				}

				// Nếu không có criterion_id hợp lệ, bỏ qua cột này
				if currentCriterionID == 0 {
					continue
				}

				// Kiểm tra criterion này có subcriteria không
				hasSubcriteria := false
				for _, ref := range assessment.AssessmentRefCriteria {
					if ref.AssessmentCriterion.ID == currentCriterionID {
						hasSubcriteria = ref.AssessmentCriterion.HasSubcriteria && len(ref.AssessmentCriterion.AssessmentSubcriteria) > 0
						break
					}
				}

				// Nếu không có subcriteria: subcriterion_id = nil, điểm ghi vào criterion_id
				// Nếu có subcriteria: điểm phải ghi vào subcriterion_id (từ hàng 5)
				var subcriterionID *int64
				if hasSubcriteria {
					// Có subcriteria: lấy subcriterion_id từ hàng 5
					if len(row5) > colIdx {
						subcriterionIDStr := row5[colIdx]
						if subcriterionIDStr != "" {
							subID, err := strconv.ParseInt(subcriterionIDStr, 10, 64)
							if err == nil && subID > 0 {
								subcriterionID = &subID
							}
						}
					}
					// Nếu không tìm thấy subcriterion_id từ hàng 5, bỏ qua cột này
					if subcriterionID == nil {
						continue
					}
				}
				// Nếu không có subcriteria: subcriterionID = nil (điểm ghi vào criterion_id)

				columnToSubcriterionMap[colIdx+1] = struct {
					CriterionID    int64
					SubcriterionID *int64
				}{
					CriterionID:    currentCriterionID,
					SubcriterionID: subcriterionID,
				}
			}
		}

		// 7. Tìm cột "Nhận xét" (cột cuối cùng có header "Nhận xét")
		commentCol := 0
		for rowIdx := 0; rowIdx < len(rows) && rowIdx < startRow-1; rowIdx++ {
			row := rows[rowIdx]
			for colIdx := len(row) - 1; colIdx >= 0; colIdx-- {
				if strings.Contains(strings.ToLower(row[colIdx]), "nhận xét") || strings.Contains(strings.ToLower(row[colIdx]), "comment") {
					commentCol = colIdx + 1
					break
				}
			}
			if commentCol > 0 {
				break
			}
		}

		// 8. Map các cột với skill_type IDs cho tiêu chí đánh giá (nếu có)
		// Chỉ map cho các cột KHÔNG có trong columnToSubcriterionMap (tức là không phải tiêu chí điểm)
		columnToSkillTypeMap := make(map[int]int64) // col -> skill_type_id
		columnToSkillMap := make(map[int]int64)      // col -> skill_id

		if studyReportCriteria != nil && len(rows) > 5 {
			// Tìm cột cuối cùng của tiêu chí điểm (cột lớn nhất trong columnToSubcriterionMap)
			maxScoreCol := 0
			for col := range columnToSubcriterionMap {
				if col > maxScoreCol {
					maxScoreCol = col
				}
			}

			// Xác định cột bắt đầu của tiêu chí đánh giá (cột đầu tiên không có trong columnToSubcriterionMap)
			evalStartCol := maxScoreCol + 1
			if evalStartCol < 7 {
				evalStartCol = 7 // Ít nhất từ cột G
			}

			// Xác định cột kết thúc (trước cột nhận xét)
			maxCol := len(rows[0])
			if commentCol > 0 {
				maxCol = commentCol - 1
			}

			// Duyệt từng cột từ evalStartCol đến maxCol
			for colIdx := evalStartCol - 1; colIdx < maxCol; colIdx++ {
				// Bỏ qua nếu cột này đã có trong columnToSubcriterionMap (là tiêu chí điểm)
				if _, isScoreCol := columnToSubcriterionMap[colIdx+1]; isScoreCol {
					continue
				}

				// Duyệt từ dưới lên (từ hàng gần startRow nhất lên hàng 5) để tìm skill_type_id
				var skillTypeID int64
				var skillID int64
				found := false
				for rowIdx := startRow - 2; rowIdx >= 4; rowIdx-- {
					if rowIdx < len(rows) {
						row := rows[rowIdx]
						if colIdx < len(row) {
							idStr := row[colIdx]
							if idStr != "" {
								id, err := strconv.ParseInt(idStr, 10, 64)
								if err == nil && id > 0 {
									// Kiểm tra id này có phải là skill_type_id trong study_report_criteria không
									for _, skill := range studyReportCriteria.Skills {
										// Tìm trong tất cả types (bao gồm nested)
										var findType func(types []models.StudyReportSkillType) bool
										findType = func(types []models.StudyReportSkillType) bool {
											for _, skillType := range types {
												if skillType.ID == id {
													// Tìm thấy skill_type_id
													skillTypeID = id
													skillID = skill.ID
													return true
												}
												// Đệ quy tìm trong các con
												if findType(skillType.NodeTypes) {
													return true
												}
											}
											return false
										}
										if findType(skill.Types) {
											found = true
											break
										}
									}
									if found {
										break
									}
								}
							}
						}
					}
				}
				if found && skillTypeID > 0 && skillID > 0 {
					columnToSkillTypeMap[colIdx+1] = skillTypeID
					columnToSkillMap[colIdx+1] = skillID
				}
			}
		}

		// 9. Xử lý từng hàng học sinh (từ startRow trở đi)
		if len(rows) < startRow {
			result.Errors = append(result.Errors, "no student rows found")
			results = append(results, result)
			continue
		}

		for rowIdx := startRow - 1; rowIdx < len(rows); rowIdx++ {
			row := rows[rowIdx]
			result.Processed++

			// Lấy student_id từ cột F (index 5)
			if len(row) < 6 {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: insufficient columns", rowIdx+1))
				continue
			}

			studentIDStr := row[5] // Cột F (index 5)
			studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
			if err != nil || studentID <= 0 {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: invalid student_id", rowIdx+1))
				continue
			}

			// Lấy comment từ cột cuối cùng
			comment := ""
			if commentCol > 0 && len(row) >= commentCol {
				comment = row[commentCol-1]
			}

			// Xử lý tiêu chí điểm (assessment_scores + assessment_score_details)
			var totalScore float64
			var scoreDetails []*models.AssessmentScoreDetail

			// Xác định cột kết thúc (trước cột nhận xét)
			maxCol := len(row)
			if commentCol > 0 {
				maxCol = commentCol - 1
			}

			for colIdx := 7; colIdx <= maxCol; colIdx++ { // Bắt đầu từ cột G (7)
				if mapping, ok := columnToSubcriterionMap[colIdx]; ok {
					scoreStr := ""
					if colIdx-1 < len(row) {
						scoreStr = row[colIdx-1] // colIdx là 1-based trong map, nhưng row là 0-based
					}
					if scoreStr != "" {
						score, err := strconv.ParseFloat(scoreStr, 64)
						if err == nil && score >= 0 {
							totalScore += score
							scoreDetails = append(scoreDetails, &models.AssessmentScoreDetail{
								CriterionID:    mapping.CriterionID,
								SubcriterionID: mapping.SubcriterionID, // nil nếu không có subcriteria
								Score:          score,
								CreatedAt:      now,
								CreatedBy:      userID,
								UpdatedAt:      now,
								UpdatedBy:      userID,
							})
						}
					}
				}
				// Tiếp tục duyệt các cột khác (có thể là tiêu chí đánh giá)
			}

			// Tạo AssessmentScore
			assessmentScore := &models.AssessmentScore{
				AssessmentID: assessmentID,
				StudentID:    studentID,
				LessonID:     lessonID,
				CourseID:     courseID,
				TotalScore:   totalScore,
				StatusScored: 1, // Đã chấm điểm
				IsLate:       false,
				CreatedAt:    now,
				CreatedBy:    userID,
				UpdatedAt:    now,
				UpdatedBy:    userID,
			}

			// Xóa các scores cũ trước
			err = s.scoreRepo.DeleteOldScores(assessmentID, studentID, courseID, userID)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: error deleting old scores: %v", rowIdx+1, err))
				continue
			}

			// Tạo AssessmentScore mới
			err = s.scoreRepo.Create(assessmentScore)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: error creating assessment score: %v", rowIdx+1, err))
				continue
			}

			// Cập nhật AssessmentScoreID cho các detail
			for _, detail := range scoreDetails {
				detail.AssessmentScoreID = assessmentScore.ID
			}

			// Tạo score details
			if len(scoreDetails) > 0 {
				err = s.scoreRepo.CreateDetails(scoreDetails)
				if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: error creating score details: %v", rowIdx+1, err))
					continue
				}
			}

			// Xử lý tiêu chí đánh giá (study_reports + study_report_ref_skill_types)
			if studyReportCriteria != nil && subjectID > 0 {
				// Tìm hoặc tạo study_report
				existingReport, err := s.studyReportRepo.FindByStudentCourseSubjectAssessment(
					studentID, courseID, subjectID, assessmentID)

				studyReport := &models.StudyReport{
					StudentId:             studentID,
					CourseId:              courseID,
					SubjectId:             subjectID,
					AssessmentId:          assessmentID,
					TeacherId:             userID,
					StudyReportCriteriaId: assessment.StudyReportCriteriaID,
					GeneralComment:        comment,
					IsCompleted:           true,
					AssessmentScoreId:     &assessmentScore.ID,
				}

				// Set name và description từ study_report_criteria
				if studyReportCriteria.Name != "" {
					studyReport.Name = studyReportCriteria.Name
				}
				if studyReportCriteria.Description != "" {
					studyReport.Description = studyReportCriteria.Description
				}

				if err == nil && existingReport != nil {
					// Update
					studyReport.ID = existingReport.ID
					studyReport.CreatedAt = existingReport.CreatedAt
					studyReport.CreatedBy = existingReport.CreatedBy
					s.studyReportRepo.SetContext(c)
					err = s.studyReportRepo.Update(studyReport)
				} else {
					// Create
					studyReport.CreatedAt = now
					studyReport.CreatedBy = userID
					s.studyReportRepo.SetContext(c)
					err = s.studyReportRepo.Create(studyReport)
				}

				if err != nil {
					result.Failed++
					result.Errors = append(result.Errors, fmt.Sprintf("row %d: error creating/updating study report: %v", rowIdx+1, err))
					continue
				}

				// Thu thập skill values từ Excel
				// Sử dụng map để đảm bảo mỗi skill_type_id chỉ xuất hiện một lần (do unique constraint)
				skillValueMap := make(map[int64]repositories.StudyReportSkillValue)
				
				// Xác định cột kết thúc (trước cột nhận xét)
				maxCol := len(row)
				if commentCol > 0 {
					maxCol = commentCol - 1
				}
				
				for colIdx := 7; colIdx <= maxCol; colIdx++ {
					if skillTypeID, ok := columnToSkillTypeMap[colIdx]; ok {
						skillID := columnToSkillMap[colIdx]
						valueStr := ""
						if colIdx-1 < len(row) {
							valueStr = row[colIdx-1] // colIdx là 1-based trong map, nhưng row là 0-based
						}

						// Xác định loại skill (star hoặc check)
						isStarSkill := false
						for _, skill := range studyReportCriteria.Skills {
							if skill.ID == skillID {
								isStarSkill = skill.Type == models.StudyReportTypeStar
								break
							}
						}

						skillValue := repositories.StudyReportSkillValue{
							SkillID:     skillID,
							SkillTypeID: skillTypeID,
						}

						shouldSave := false
						if isStarSkill {
							// Star skill: chỉ lưu khi có giá trị
							if valueStr != "" {
								star, err := strconv.ParseInt(valueStr, 10, 32)
								if err == nil && star >= 0 {
									starInt32 := int32(star)
									skillValue.StarValue = &starInt32
									shouldSave = true
								}
							}
						} else {
							// Check skill: luôn lưu, nếu không có giá trị thì is_check = false
							isCheck := false
							if valueStr == "☑" || valueStr == "✓" || valueStr == "true" || valueStr == "1" {
								isCheck = true
							}
							skillValue.CheckValue = &isCheck
							shouldSave = true
						}

						// Lưu vào map
						if shouldSave {
							skillValueMap[skillTypeID] = skillValue
						}
					}
				}

				// Chuyển map thành slice
				skillValues := make([]repositories.StudyReportSkillValue, 0, len(skillValueMap))
				for _, skillValue := range skillValueMap {
					skillValues = append(skillValues, skillValue)
				}

				// Replace skill values
				if len(skillValues) > 0 {
					err = s.studyReportRepo.ReplaceSkillValues(studyReport.ID, assessment.StudyReportCriteriaID, skillValues)
					if err != nil {
						result.Failed++
						result.Errors = append(result.Errors, fmt.Sprintf("row %d: error replacing skill values: %v", rowIdx+1, err))
						continue
					}
				}
			}

			result.Success++
		}

		results = append(results, result)
	}

	return gin.H{
		"message": "Import completed",
		"results": results,
	}, nil
}
