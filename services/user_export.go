package services

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// removeVietnameseAccents loại bỏ dấu tiếng Việt
func removeVietnameseAccents(s string) string {
	accents := map[string]string{
		"à": "a", "á": "a", "ả": "a", "ã": "a", "ạ": "a", "â": "a", "ầ": "a", "ấ": "a", "ẩ": "a", "ẫ": "a", "ậ": "a", "ă": "a", "ằ": "a", "ắ": "a", "ẳ": "a", "ẵ": "a", "ặ": "a",
		"đ": "d",
		"è": "e", "é": "e", "ẻ": "e", "ẽ": "e", "ẹ": "e", "ê": "e", "ề": "e", "ế": "e", "ể": "e", "ễ": "e", "ệ": "e",
		"ì": "i", "í": "i", "ỉ": "i", "ĩ": "i", "ị": "i",
		"ò": "o", "ó": "o", "ỏ": "o", "õ": "o", "ọ": "o", "ô": "o", "ồ": "o", "ố": "o", "ổ": "o", "ỗ": "o", "ộ": "o", "ơ": "o", "ờ": "o", "ớ": "o", "ở": "o", "ỡ": "o", "ợ": "o",
		"ù": "u", "ú": "u", "ủ": "u", "ũ": "u", "ụ": "u", "ư": "u", "ừ": "u", "ứ": "u", "ử": "u", "ữ": "u", "ự": "u",
		"ỳ": "y", "ý": "y", "ỷ": "y", "ỹ": "y", "ỵ": "y",
		"À": "A", "Á": "A", "Ả": "A", "Ã": "A", "Ạ": "A", "Â": "A", "Ầ": "A", "Ấ": "A", "Ẩ": "A", "Ẫ": "A", "Ậ": "A", "Ă": "A", "Ằ": "A", "Ắ": "A", "Ẳ": "A", "Ẵ": "A", "Ặ": "A",
		"Đ": "D",
		"È": "E", "É": "E", "Ẻ": "E", "Ẽ": "E", "Ẹ": "E", "Ê": "E", "Ề": "E", "Ế": "E", "Ể": "E", "Ễ": "E", "Ệ": "E",
		"Ì": "I", "Í": "I", "Ỉ": "I", "Ĩ": "I", "Ị": "I",
		"Ò": "O", "Ó": "O", "Ỏ": "O", "Õ": "O", "Ọ": "O", "Ô": "O", "Ồ": "O", "Ố": "O", "Ổ": "O", "Ỗ": "O", "Ộ": "O", "Ơ": "O", "Ờ": "O", "Ớ": "O", "Ở": "O", "Ỡ": "O", "Ợ": "O",
		"Ù": "U", "Ú": "U", "Ủ": "U", "Ũ": "U", "Ụ": "U", "Ư": "U", "Ừ": "U", "Ứ": "U", "Ử": "U", "Ữ": "U", "Ự": "U",
		"Ỳ": "Y", "Ý": "Y", "Ỷ": "Y", "Ỹ": "Y", "Ỵ": "Y",
	}

	result := s
	for accented, plain := range accents {
		result = strings.ReplaceAll(result, accented, plain)
	}
	return result
}

func (s *userService) Export(c *gin.Context) (string, error) {
	classRepo := repositories.NewClassRepository()
	courseRepo := repositories.NewCourseRepository()
	roleRepo := repositories.NewRoleRepository()
	schoolRepo := repositories.NewSchoolRepository()

	classRepo.SetContext(c)
	courseRepo.SetContext(c)
	roleRepo.SetContext(c)
	schoolRepo.SetContext(c)

	allowedFilters := []string{
		"parent_id",
		"school_id",
	}

	filter, _, _, keyword, _, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return "", err
	}

	filter, _ = s.ApplyFilter(c, filter)

	s.repo.SetSearch(keyword, []string{
		"username",
		"code",
		"name",
		"email",
		"phone_number",
		"id",
	})

	sort := map[string]string{
		"name": "asc",
	}

	s.repo.SetPreload([]string{"Roles"})
	s.repo.SetFilter(filter)
	s.repo.SetSort(sort)

	roles, err := roleRepo.GetAll()
	if err != nil {
		return "", err
	}

	schoolIds, err := schoolRepo.GetAllIds()

	schoolFilter := map[string]interface{}{}

	if err == nil && len(schoolIds) > 0 {
		schoolIdStrs := make([]string, len(schoolIds))
		for i, id := range schoolIds {
			schoolIdStrs[i] = strconv.FormatInt(id, 10)
		}

		schoolFilter["school_id"] = "in:" + strings.Join(schoolIdStrs, ",")
	} else {
		schoolFilter["school_id"] = "in:-1"
	}

	classRepo.SetFilter(schoolFilter)
	classRepo.SetPreload([]string{"School"})
	classes, err := classRepo.GetAll()
	if err != nil {
		return "", err
	}

	courseRepo.SetContext(c)
	courses, err := courseRepo.GetAll()
	if err != nil {
		return "", err
	}

	users, err := s.repo.GetAll()
	if err != nil {
		return "", err
	}

	userClasses, err := s.repo.GetAllUserClasses()
	if err != nil {
		return "", err
	}

	userCourses, err := s.repo.GetAllUserCourses()
	if err != nil {
		return "", err
	}

	f := excelize.NewFile()

	// === Map: userID -> []CourseID / []ClassID ===
	userCourseMap := make(map[int64][]int64)
	userCourseNameMap := make(map[int64][]string)
	for _, uc := range userCourses {
		userCourseMap[uc.UserId] = append(userCourseMap[uc.UserId], uc.CourseId)
		for _, c := range courses {
			if c.ID == uc.CourseId {
				userCourseNameMap[uc.UserId] = append(userCourseNameMap[uc.UserId], c.Name)
			}
		}
	}

	userClassMap := make(map[int64][]int64)
	userClassNameMap := make(map[int64][]string)
	for _, uc := range userClasses {
		userClassMap[uc.UserId] = append(userClassMap[uc.UserId], uc.ClassId)
		for _, c := range classes {
			if c.ID == uc.ClassId {
				userClassNameMap[uc.UserId] = append(userClassNameMap[uc.UserId], c.Name)
			}
		}
	}

	// === Style header ===
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
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "top", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
		},
	})

	if err != nil {
		return "", err
	}

	// === Style row xen kẽ ===
	oddRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F5F5F5"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		config.Log.Error("Failed to create oddRowStyle:", err)
		return "", err
	}

	evenRowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		config.Log.Error("Failed to create evenRowStyle:", err)
		return "", err
	}

	// ===== Sheet: Users =====
	userSheet := "Users"
	f.SetSheetName("Sheet1", userSheet)

	userHeaders := []string{
		"ID", "Identifier", "Username",
		"Code", "Name", "Email",
		"Phone", "Role ID", "Password",
		"Avatar", "Status", "Class IDs",
		"Course IDs", "Class Names", "Course Names",
	}

	for i, h := range userHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(userSheet, cell, h)
		f.SetCellStyle(userSheet, cell, cell, headerStyle)
	}
	for idx, u := range users {
		row := idx + 2

		var roleId int
		for _, r := range u.Roles {
			roleId = int(r.ID)
			break
		}

		f.SetCellValue(userSheet, fmt.Sprintf("A%d", row), u.ID)
		f.SetCellValue(userSheet, fmt.Sprintf("B%d", row), u.Identifier)
		f.SetCellValue(userSheet, fmt.Sprintf("C%d", row), u.Username)
		f.SetCellValue(userSheet, fmt.Sprintf("D%d", row), u.Code)
		f.SetCellValue(userSheet, fmt.Sprintf("E%d", row), u.Name)
		f.SetCellValue(userSheet, fmt.Sprintf("F%d", row), u.Email)
		f.SetCellValue(userSheet, fmt.Sprintf("G%d", row), u.PhoneNumber)
		f.SetCellValue(userSheet, fmt.Sprintf("H%d", row), roleId)
		f.SetCellValue(userSheet, fmt.Sprintf("I%d", row), "")
		f.SetCellValue(userSheet, fmt.Sprintf("J%d", row), u.AvatarInfo.Path)
		f.SetCellValue(userSheet, fmt.Sprintf("K%d", row), u.Status)

		// IDs
		courseIDs := userCourseMap[int64(u.ID)]
		classIDs := userClassMap[int64(u.ID)]
		classNames := userClassNameMap[int64(u.ID)]
		courseNames := userCourseNameMap[int64(u.ID)]

		var classStrs []string
		for _, cid := range classIDs {
			classStrs = append(classStrs, fmt.Sprintf("%d", cid))
		}

		var courseStrs []string
		for _, cid := range courseIDs {
			courseStrs = append(courseStrs, fmt.Sprintf("%d", cid))
		}

		classNameStrs := append([]string{}, classNames...)
		courseNameStrs := append([]string{}, courseNames...)

		f.SetCellValue(userSheet, fmt.Sprintf("L%d", row), strings.Join(classStrs, ","))
		f.SetCellValue(userSheet, fmt.Sprintf("M%d", row), strings.Join(courseStrs, ","))

		f.SetCellValue(userSheet, fmt.Sprintf("N%d", row), strings.Join(classNameStrs, ","))
		f.SetCellValue(userSheet, fmt.Sprintf("O%d", row), strings.Join(courseNameStrs, ","))
	}

	// Áp dụng style xen kẽ sau khi đã set tất cả cell values
	for idx := range users {
		row := idx + 2
		// Áp dụng style cho từng cell trong hàng
		for col := 'A'; col <= rune('A'+len(userHeaders)-1); col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(userSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(userSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Roles =====
	roleSheet := "Roles"
	f.NewSheet(roleSheet)

	roleHeaders := []string{"ID", "Name"}

	for i, h := range roleHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(roleSheet, cell, h)
		f.SetCellStyle(roleSheet, cell, cell, headerStyle)
	}
	for idx, role := range roles {
		row := idx + 2
		f.SetCellValue(roleSheet, fmt.Sprintf("A%d", row), role.ID)
		f.SetCellValue(roleSheet, fmt.Sprintf("B%d", row), role.Name)
	}

	for idx := range roles {
		row := idx + 2
		for col := 'A'; col <= 'B'; col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(roleSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(roleSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Classes =====
	classSheet := "Classes"
	f.NewSheet(classSheet)

	classHeaders := []string{"ID", "Name", "School"}

	for i, h := range classHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(classSheet, cell, h)
		f.SetCellStyle(classSheet, cell, cell, headerStyle)
	}
	for idx, class := range classes {
		row := idx + 2
		f.SetCellValue(classSheet, fmt.Sprintf("A%d", row), class.ID)
		f.SetCellValue(classSheet, fmt.Sprintf("B%d", row), class.Name)
		f.SetCellValue(classSheet, fmt.Sprintf("C%d", row), class.School.Name)
	}

	for idx := range classes {
		row := idx + 2
		for col := 'A'; col <= 'C'; col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(classSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(classSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Sheet: Courses =====
	courseSheet := "Courses"
	f.NewSheet(courseSheet)

	courseHeaders := []string{"ID", "Name", "Description", "Object title"}

	for i, h := range courseHeaders {
		col := string(rune('A' + i))
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(courseSheet, cell, h)
		f.SetCellStyle(courseSheet, cell, cell, headerStyle)
	}
	for idx, course := range courses {
		row := idx + 2
		f.SetCellValue(courseSheet, fmt.Sprintf("A%d", row), course.ID)
		f.SetCellValue(courseSheet, fmt.Sprintf("B%d", row), course.Name)
		f.SetCellValue(courseSheet, fmt.Sprintf("C%d", row), course.Description)
		f.SetCellValue(courseSheet, fmt.Sprintf("D%d", row), course.ObjectTitle)
	}

	for idx := range courses {
		row := idx + 2
		for col := 'A'; col <= 'D'; col++ {
			cell := fmt.Sprintf("%c%d", col, row)
			if idx%2 == 0 {
				f.SetCellStyle(courseSheet, cell, cell, oddRowStyle)
			} else {
				f.SetCellStyle(courseSheet, cell, cell, evenRowStyle)
			}
		}
	}

	// ===== Save file to S3 =====
	filename := fmt.Sprintf("users_%d.xlsx", time.Now().Unix())

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

// ExportUsersPDF xuất file PDF danh sách tài khoản & mật khẩu theo trường/lớp
func (s *userService) ExportUsersPDF(schoolID, classID int) ([]byte, string, error) {
	// Lấy thông tin trường từ DB
	schoolRepo := repositories.NewSchoolRepository()
	schoolFilter := map[string]interface{}{"id": schoolID}
	schoolRepo.SetFilter(schoolFilter)
	schools, err := schoolRepo.GetAll()
	if err != nil || len(schools) == 0 {
		return nil, "", fmt.Errorf("không tìm thấy trường học với ID: %d", schoolID)
	}
	schoolName := schools[0].Name

	// Lấy thông tin lớp từ DB
	classRepo := repositories.NewClassRepository()
	classFilter := map[string]interface{}{"id": classID}
	classRepo.SetFilter(classFilter)
	classes, err := classRepo.GetAll()
	if err != nil || len(classes) == 0 {
		return nil, "", fmt.Errorf("không tìm thấy lớp học với ID: %d", classID)
	}
	className := classes[0].Name

	// Lấy danh sách user theo schoolID
	filter := map[string]interface{}{
		"school_id": schoolID,
	}
	s.repo.SetFilter(filter)
	s.repo.SetPreload([]string{"UserClasses"}) // Preload quan hệ UserClasses
	allUsers, err := s.repo.GetAll()
	if err != nil {
		return nil, "", err
	}

	// Lọc user theo classID thông qua UserClasses
	var users []models.User
	for _, user := range allUsers {
		for _, userClass := range user.UserClasses {
			if int64(classID) == userClass.ClassId {
				users = append(users, user)
				break
			}
		}
	}

	// Sử dụng logo từ domain public của Enspire
	logoURL := "https://enspire.vn/wp-content/uploads/2023/09/logo.png"

	config.Log.Info("Using Enspire public logo URL:", logoURL)

	// Sử dụng Google Fonts với Vietnamese subset
	fontCSS := `@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap&subset=vietnamese');`

	// Tạo rows (escape để tránh HTML injection)
	var rowsBuilder strings.Builder
	for i, user := range users {
		nameEsc := html.EscapeString(user.Name)
		userEsc := html.EscapeString(user.Username)
		fmt.Fprintf(&rowsBuilder, `
          <tr>
            <td class="stt">%d</td>
            <td class="name">%s</td>
            <td class="account">%s</td>
          </tr>`, i+1, nameEsc, userEsc)
	}
	rowsHTML := rowsBuilder.String()

	// Build HTML (chèn fontCSS ở đầu <style>)
	htmlContent := fmt.Sprintf(`
<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <style>
    %s
    :root{
      --accent:#1e88e5;
      --accent-dark:#1565c0;
      --yellow:#fbbc05;
      --muted:#6b7280;
      --bg:#f6f9fc;
      --card:#ffffff;
    }
    body {
      margin:0;
      padding:0;
      background:var(--bg);
      font-family:'Inter', 'Segoe UI', 'Arial', 'Helvetica Neue', sans-serif;
      color:#222;
    }
    .page{width:210mm;min-height:297mm;margin:12mm auto;padding:18mm;background:#fff;
      box-shadow:0 6px 18px rgba(16,24,40,0.08);border-radius:6px;}
    .header{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:24px;}
    .brand{display:flex;align-items:center;gap:12px;}
    .logo img{width:64px;height:64px;border-radius:8px;object-fit:contain;display:block;}
    .title h1{margin:0;font-size:18px;font-weight:700;color:var(--accent-dark);margin-bottom:6px;}
    .title p{margin:0;font-size:11px;color:var(--muted);}
    .contact{text-align:right;font-size:12px;line-height:1.6;}
    .hotline{display:inline-block;background:linear-gradient(90deg,var(--accent),var(--accent-dark));color:#fff;
      padding:6px 10px;margin-bottom:6px;border-radius:20px;font-weight:600;}
    .report-title{text-align:center;margin:8px 0 18px 0;}
    .report-title h2{
      margin:0;
      font-size:20px;
      font-weight:700;
      text-transform:uppercase;
      color:var(--yellow);
    }
    .report-sub{
      margin-top:8px;
      font-size:14px;
      font-weight:600;
      color:var(--accent-dark);
    }
    table{width:100%%;border-collapse:collapse;font-size:13px;color:var(--accent-dark);}
    thead th{
      padding:10px 8px;
      background:var(--yellow);
      color:#fff;
      font-weight:600;
      text-align:center;
      border:1px solid #ddd;
    }
    tbody td{
      padding:10px 8px;
      border:1px solid #ddd;
      text-align:center;
    }
    tbody tr:nth-child(odd){background:rgba(251,188,5,0.06);}
    .stt{width:10%%;}
    .name{width:40%%;}
    .account{width:50%%;}
    .note{margin-top:14px;font-size:11px;color:var(--muted);display:flex;justify-content:space-between;}
  </style>
</head>
<body>
  <div class="page">
    <div class="header">
      <div class="brand">
        <div class="logo">
          %s
        </div>
        <div class="title">
          <h1>Thông tin đăng nhập hệ thống LMS</h1>
          <p>Danh sách tài khoản</p>
        </div>
      </div>
      <div class="contact">
        <div class="hotline">0934609998</div>
        <div>HOTLINE HỖ TRỢ</div>
      </div>
    </div>

    <div class="report-title">
      <h2>TÀI KHOẢN &amp; MẬT KHẨU</h2>
      <div class="report-sub">%s — %s</div>
    </div>

    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th class="stt">STT</th>
            <th class="name">Họ và tên</th>
            <th class="account">Tài khoản / Mật khẩu</th>
          </tr>
        </thead>
        <tbody>
          %s
        </tbody>
      </table>
    </div>

    <div class="note">
      <div>%s — %s</div>
    </div>
  </div>
</body>
</html>`,
		fontCSS,
		fmt.Sprintf(`<img src="%s" alt="Enspire Logo"/>`, logoURL),
		schoolName, className,
		rowsHTML,
		schoolName, className)

	// Debug: Log để kiểm tra logo URL
	config.Log.Info("Generated HTML contains logo URL:", logoURL)

	// Convert HTML sang PDF bằng Fly.dev PDF Service
	pdfBuffer, err := callFlyPDFService(htmlContent)
	if err != nil {
		return nil, "", fmt.Errorf("PDF conversion failed: %v", err)
	}

	// Tạo filename cho download
	filename := fmt.Sprintf("tai-khoan-mat-khau-%s-%s.pdf",
		removeVietnameseAccents(schoolName),
		removeVietnameseAccents(className))

	// Trả PDF binary trực tiếp
	config.Log.Info("PDF generated successfully for download, size:", len(pdfBuffer), "bytes")
	return pdfBuffer, filename, nil
}

// callFlyPDFService gọi API PDF service trên Fly.dev
func callFlyPDFService(htmlContent string) ([]byte, error) {
	// Chuẩn bị request body
	requestBody := map[string]string{
		"content": htmlContent,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Gọi API
	apiURL := "https://pdf-service-green-sound-6593.fly.dev/pdf/generate-from-html"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Đọc response body (PDF binary)
	pdfBuffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Kiểm tra PDF format
	if len(pdfBuffer) < 4 || string(pdfBuffer[:4]) != "%PDF" {
		return nil, fmt.Errorf("response is not valid PDF format")
	}

	config.Log.Info("PDF generated successfully, size:", len(pdfBuffer), "bytes")
	return pdfBuffer, nil
}
