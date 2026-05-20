package utils

import (
	"be-cleverschool/config"
	"be-cleverschool/i18n"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"be-cleverschool/prot"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func BoolToInt16(b bool) int16 {
	if b {
		return 1
	}
	return 0
}

func Contains(slice []uint64, value uint64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func GetCurrentMemberType(c *gin.Context) string {
	memberType, ok := c.Get("memberType")
	if !ok {
		return ""
	}
	return memberType.(string)
}

func GetCurrentUserId(c *gin.Context) int {
	idVal, ok := c.Get("userID")
	if !ok {
		return 0
	}

	var id int
	switch v := idVal.(type) {
	case int64:
		id = int(v)
	case int:
		id = v
	case float64:
		id = int(v)
	default:
		return 0
	}

	return id
}
func GetCurrentRoleId(c *gin.Context) int {
	idVal, ok := c.Get("roleID")
	if !ok {
		return 0
	}

	var id int
	switch v := idVal.(type) {
	case int64:
		id = int(v)
	case int:
		id = v
	case float64:
		id = int(v)
	default:
		return 0
	}

	return id
}
func GetCurrentSchoolId(c *gin.Context) int {
	idVal, ok := c.Get("schoolID")
	if !ok {
		return 0
	}

	var id int
	switch v := idVal.(type) {
	case int64:
		id = int(v)
	case int:
		id = v
	case float64:
		id = int(v)
	default:
		return 0
	}

	return id
}

func InArray(val string, array []string) bool {
	for _, item := range array {
		if item == val {
			return true
		}
	}
	return false
}

func ParsePaginationParams(c *gin.Context, allowedFilters []string) (map[string]interface{}, int, int, string, map[string]string, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "100"))
	if err != nil || perPage <= 0 {
		perPage = 100
	}

	filter := make(map[string]interface{})
	for _, filterKey := range allowedFilters {
		if val := c.Query(filterKey); val != "" {
			filter[filterKey] = val
		}
	}

	keyword := strings.TrimSpace(c.Query("keyword"))
	search := strings.TrimSpace(c.Query("search"))

	if search != "" {
		keyword = search
	}

	// ✅ Parse sort_*
	sort := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "sort_") && len(values) > 0 {
			field := strings.TrimPrefix(key, "sort_")
			order := strings.ToLower(strings.TrimSpace(values[0]))
			if order != "" {
				sort[field] = order
			}
		}
	}

	return filter, page, perPage, keyword, sort, nil
}

func StringPtr(s string) *string {
	return &s
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func DerefStr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func DerefInt64(i *int64) int64 {
	if i != nil {
		return *i
	}
	return 0
}

func GetString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func GetInt64(n *int64) int64 {
	if n != nil {
		return *n
	}
	return 0
}

func GetBody[T proto.Message](c *gin.Context, newFunc func() T) (T, error, string) {
	req := newFunc()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
	}

	// Reset body so it can be read again later if needed
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	accept := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept")))

	isProto := strings.Contains(contentType, "application/x-protobuf") ||
		strings.Contains(contentType, "application/protobuf") ||
		strings.Contains(accept, "application/x-protobuf") ||
		strings.Contains(accept, "application/protobuf")

	if isProto {
		if err := proto.Unmarshal(body, req); err != nil {
			return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
		}
	} else {
		// Đối với /save-score/bulk, không discard unknown fields để giữ question_id và data
		isSaveScoreBulk := strings.Contains(c.Request.RequestURI, "/save-score/bulk")
		unmarshalOpts := protojson.UnmarshalOptions{
			DiscardUnknown: !isSaveScoreBulk,
		}
		if err := unmarshalOpts.Unmarshal(body, req); err != nil {
			// Try to convert legacy SaveScoreBulk format
			if isSaveScoreBulk {
				convertedBody, convertErr := ConvertLegacySaveScoreRequest(body)
				if convertErr == nil {
					if err := unmarshalOpts.Unmarshal(convertedBody, req); err == nil {
						return req, nil, ""
					}
				}
			}
			return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
		}
	}

	return req, nil, ""
}

// GetSaveScoreBulkBody parses JSON request body for /save-score/bulk endpoint
// This function ensures question_id and data are preserved in each answer
func GetSaveScoreBulkBody(c *gin.Context) (*prot.SaveScoreBulkRequest, error, string) {
	req := &prot.SaveScoreBulkRequest{}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
	}

	// Reset body so it can be read again later if needed
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	accept := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept")))

	isProto := strings.Contains(contentType, "application/x-protobuf") ||
		strings.Contains(contentType, "application/protobuf") ||
		strings.Contains(accept, "application/x-protobuf") ||
		strings.Contains(accept, "application/protobuf")

	if isProto {
		// Proto format
		if err := proto.Unmarshal(body, req); err != nil {
			return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
		}
		return req, nil, ""
	}

	// JSON format - parse trực tiếp để giữ question_id và data
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(body, &jsonMap); err != nil {
		// Try to convert legacy SaveScoreBulk format
		convertedBody, convertErr := ConvertLegacySaveScoreRequest(body)
		if convertErr != nil {
			return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
		}
		if err := json.Unmarshal(convertedBody, &jsonMap); err != nil {
			return req, fmt.Errorf(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
		}
	}

	// Map các fields từ JSON vào proto struct
	if examId, ok := jsonMap["exam_id"].(float64); ok {
		req.ExamId = int64(examId)
	}
	if exerciseId, ok := jsonMap["exercise_id"].(float64); ok {
		req.ExerciseId = int64(exerciseId)
	}
	if levelTestId, ok := jsonMap["level_test_id"].(float64); ok {
		req.LevelTestId = int64(levelTestId)
	}
	if homeworkId, ok := jsonMap["homework_id"].(float64); ok {
		req.HomeworkId = int64(homeworkId)
	}
	if contestRoundId, ok := jsonMap["contest_round_id"].(float64); ok {
		req.ContestRoundId = int64(contestRoundId)
	}
	if lessonId, ok := jsonMap["lesson_id"].(float64); ok {
		req.LessonId = int64(lessonId)
	}
	if timeVal, ok := jsonMap["time"].(float64); ok {
		req.Time = int64(timeVal)
	}
	if userId, ok := jsonMap["user_id"].(float64); ok {
		req.UserId = int64(userId)
	}

	// Parse list_answers và giữ lại question_id và data
	listAnswersInterface, ok := jsonMap["list_answers"]
	if !ok {
		return req, fmt.Errorf("missing list_answers in request"), "messages.data_invalid"
	}

	listAnswers, ok := listAnswersInterface.([]interface{})
	if !ok {
		return req, fmt.Errorf("list_answers is not an array"), "messages.data_invalid"
	}

	// Lưu JSON maps gốc vào context để dùng khi convert
	answerJSONMaps := make([]map[string]interface{}, 0, len(listAnswers))

	req.ListAnswers = make([]*prot.AnswerRequest, 0, len(listAnswers))
	for i, answerInterface := range listAnswers {
		answerMap, ok := answerInterface.(map[string]interface{})
		if !ok {
			return req, fmt.Errorf("answer at index %d is not an object", i), "messages.data_invalid"
		}

		// Lưu JSON map gốc
		answerJSONMaps = append(answerJSONMaps, answerMap)

		answerReq := &prot.AnswerRequest{}

		// Set type_question
		if typeQuestion, ok := answerMap["type_question"].(string); ok {
			answerReq.TypeQuestion = typeQuestion
		} else {
			return req, fmt.Errorf("missing type_question in answer at index %d", i), "messages.data_invalid"
		}

		req.ListAnswers = append(req.ListAnswers, answerReq)
	}

	// Lưu JSON maps gốc vào context để dùng khi convert
	c.Set("save_score_bulk_answer_json_maps", answerJSONMaps)

	return req, nil, ""
}

func GetBodyFromBytes[T proto.Message](bodyBytes []byte, newFunc func() T) (T, error) {
	req := newFunc()

	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	if err := unmarshalOpts.Unmarshal(bodyBytes, req); err != nil {
		return req, fmt.Errorf(i18n.Localize("messages.data_invalid"))
	}

	return req, nil
}
func ParseIDs(s string) []int64 {
	var ids []int64

	normalized := strings.ReplaceAll(s, ".", ",")
	parts := strings.Split(normalized, ",")

	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

func Int64OrZero(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func Int64PtrOrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func RoundTo2Decimal(val float64) float64 {
	return math.Round(val*100) / 100
}

func ApplyFilterDate(c *gin.Context, filter map[string]interface{}) map[string]interface{} {
	beginAtStr := c.Query("begin_at")
	endAtStr := c.Query("end_at")

	if endAtStr != "" {
		t, err := time.Parse("2006-01-02", endAtStr)
		if err != nil {
			return filter
		}

		endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endAtStr = endOfDay.Format("2006-01-02 15:04:05")
	}

	if beginAtStr != "" && endAtStr != "" {
		filter["created_at"] = "between:" + beginAtStr + "," + endAtStr
	} else if beginAtStr != "" {
		filter["created_at"] = ">=:" + beginAtStr
	} else if endAtStr != "" {
		filter["created_at"] = "<=:" + endAtStr
	}

	return filter
}

func FormatNumberWithComma(n int) string {
	return fmt.Sprintf("%d", n)
}

func StaticURL(path string, storage string) string {
	// Returns integer if already an absolute URL
	if path == "" || strings.Contains(path, "http") {
		return path
	}

	storage = "s3"

	// if strings.Contains(path, "power-point") {
	// 	storage = config.Public
	// } else {
	// 	storage = "s3"
	// }

	disk, ok := config.Disks[storage]
	if !ok {
		return path
	}

	base := disk.URL
	if base == "" {
		base = "/" + storage
	}

	fullURL := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")

	fullURL = strings.ReplaceAll(fullURL, "+", "%2B")
	fullURL = strings.ReplaceAll(fullURL, "+", "%2B")
	fullURL = strings.ReplaceAll(fullURL, "power-point", "power_point")

	return fullURL
}

func StripDomain(fullURL string, storage string) string {
	appUrl := config.LoadConfig().AppUrl
	fullURL = strings.ReplaceAll(fullURL, "power_point", "power-point")

	if strings.HasPrefix(fullURL, appUrl) {
		trimmed := strings.TrimPrefix(fullURL, appUrl)
		return strings.TrimPrefix(trimmed, "/")
	}

	disk, ok := config.Disks[storage]
	if !ok {
		config.Log.Info("StripDomain disk error")
		return fullURL
	}

	base := disk.URL

	if fullURL == "" || base == "" {
		return fullURL
	}

	if strings.HasPrefix(fullURL, base) {
		trimmed := strings.TrimPrefix(fullURL, base)
		return strings.TrimPrefix(trimmed, "/")
	}

	return fullURL
}

func ToPtrSlice[T any](in []T) []*T {
	out := make([]*T, len(in))
	for i := range in {
		out[i] = &in[i]
	}
	return out
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func BoolOrFalse(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

func FormatPercent(value float64, decimalPlaces ...int) string {
	withSign := false
	places := 1

	if len(decimalPlaces) > 0 {
		last := decimalPlaces[len(decimalPlaces)-1]
		if last < 0 {
			withSign = true
			decimalPlaces = decimalPlaces[:len(decimalPlaces)-1]
		}
	}

	if len(decimalPlaces) > 0 {
		places = decimalPlaces[0]
	}

	sign := ""
	if withSign {
		if value > 0 {
			sign = "+"
		} else if value < 0 {
			sign = "-"
			value = -value
		}
	} else {
		value = math.Abs(value)
	}

	if value == float64(int(value)) {
		return fmt.Sprintf("%s%d%%", sign, int(value))
	}

	formatStr := fmt.Sprintf("%%s%%.%df%%%%", places)
	return fmt.Sprintf(formatStr, sign, value)
}

func FormatDurationMinutes(mins float32, decimalPlaces ...int) string {
	places := 1
	if len(decimalPlaces) > 0 {
		places = decimalPlaces[0]
	}

	if mins >= 60 {
		hours := mins / 60

		format := "%." + fmt.Sprintf("%df", places)
		hoursStr := fmt.Sprintf(format, hours)

		if hours == 1 {
			return i18n.Localize("time.hour", map[string]interface{}{"Number": hoursStr})
		} else {
			return i18n.Localize("time.hours", map[string]interface{}{"Number": hoursStr})
		}
	}

	if mins == 1 {
		return i18n.Localize("time.minute", map[string]interface{}{"Number": fmt.Sprintf("%d", int(mins))})
	} else {
		return i18n.Localize("time.minutes", map[string]interface{}{"Number": fmt.Sprintf("%d", int(mins))})
	}
}

func FormatFloatToString(value float64, decimalPlaces ...int) string {
	places := 1

	if len(decimalPlaces) > 0 {
		places = decimalPlaces[0]
	}

	if value == float64(int64(value)) {
		return fmt.Sprintf("%d", int64(value))
	}

	format := fmt.Sprintf("%%.%df", places)
	return fmt.Sprintf(format, value)
}

func FormatFloat(value float64, decimalPlaces ...int) float64 {
	places := 1
	if len(decimalPlaces) > 0 {
		places = decimalPlaces[0]
	}

	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

func SafePercent(numerator, denominator int32) float32 {
	if denominator == 0 {
		return 0
	}
	return float32(FormatFloat(float64(numerator)/float64(denominator)*100, 2))
}

var specialRuneMap = map[rune]string{
	'Đ': "D",
	'đ': "d",
	'Ł': "L",
	'ł': "l",
	'Ø': "O",
	'ø': "o",
	'Æ': "AE",
	'æ': "ae",
	'ß': "ss",
	'Þ': "Th",
	'þ': "th",
}

func RemoveAccents(s string) string {
	t := norm.NFD.String(s)
	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}

		if repl, ok := specialRuneMap[r]; ok {
			b.WriteString(repl)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func SanitizeClassName(className string) string {
	clean := RemoveAccents(className)
	clean = strings.ToLower(clean)
	clean = strings.Join(strings.Fields(clean), "")

	// Special: Nếu bắt đầu bằng "lop" thì bỏ nó đi
	if strings.HasPrefix(clean, "lop") {
		clean = strings.TrimPrefix(clean, "lop")
	}

	if clean == "" {
		return ""
	}

	// Nếu nhiều từ, lấy các chữ cái đầu, tối đa 3 ký tự
	if len(clean) > 3 {
		return clean[:3]
	}

	return clean
}

func NormalizeKeyword(s string) string {
	s = RemoveAccents(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return CopyFile(path, destPath)
	})
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func FormatMonthYear(month int32, year int32, locale string) string {
	if month < 1 || month > 12 {
		return ""
	}
	t := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	if locale == "vi" {
		return fmt.Sprintf("tháng %d/%d", month, year)
	}

	return t.Format("January 2006")
}

func FormatMonthOrQuarter(quarter int32, year int32, locale string) string {
	if quarter < 1 || quarter > 4 {
		return ""
	}

	if locale == "vi" {
		return fmt.Sprintf("quý %d/%d", quarter, year)
	}

	suffix := "th"
	switch quarter {
	case 1:
		suffix = "st"
	case 2:
		suffix = "nd"
	case 3:
		suffix = "rd"
	}

	return fmt.Sprintf("%d%s quarter %d", quarter, suffix, year)
}

func PtrInt64(i int64) *int64 {
	return &i
}

