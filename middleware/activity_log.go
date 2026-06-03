package middleware

import (
	"be-lms/config"
	"be-lms/models"
	"be-lms/repositories"
	"be-lms/services"
	"be-lms/utils"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)                  // lưu nội dung trả về
	return w.ResponseWriter.Write(b) // ghi ra client
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token    string `json:"token"`
		RoleID   int    `json:"role_id"`
		RoleName string `json:"role_name"`
		Type     string `json:"type"`
	} `json:"data"`
	Error string `json:"error"`
}

func ActivityLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body
		var requestBodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBodyBytes = bodyBytes
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // reset body
			}
		}

		// Save attribute if PUT (before c.Next())
		var attributeData []byte
		method := c.Request.Method
		pathUrl := c.Request.URL.Path
		queryParams := c.Request.URL.RawQuery

		type modelHandler struct {
			Prefix string
			GetFn  func(c *gin.Context, id int) (interface{}, error)
		}

		modelHandlers := []modelHandler{
			{
				Prefix: "/api/manage/questions/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewQuestionRepository()
					service := services.NewQuestionService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/question-attributes/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewQuestionAttributeRepository()
					service := services.NewQuestionAttributeService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/users/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewUserRepository()
					service := services.NewUserService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/lessons/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewLessonRepository()
					service := services.NewLessonService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/certificates/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewCertificateRepository()
					service := services.NewCertificateService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/certificates/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewCertificateRepository()
					service := services.NewCertificateService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/chapters/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewChapterRepository()
					service := services.NewChapterService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/classes/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewClassRepository()
					service := services.NewClassService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/courses/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewCourseRepository()
					service := services.NewCourseService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/degrees/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewDegreeRepository()
					service := services.NewDegreeService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/departments/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewDepartmentRepository()
					service := services.NewDepartmentService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/employee-positions/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewEmployeePositionRepository()
					service := services.NewEmployeePositionService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/exams/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewExamRepository()
					service := services.NewExamService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/grades/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewGradeRepository()
					service := services.NewGradeService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/homeworks/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewHomeworkRepository()
					service := services.NewHomeworkService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/lesson-plans/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewLessonPlanRepository()
					service := services.NewLessonPlanService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/schools/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewSchoolRepository()
					service := services.NewSchoolService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/skills/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewSkillRepository()
					service := services.NewSkillService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/subjects/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewSubjectRepository()
					service := services.NewSubjectService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/tags/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewTagRepository()
					service := services.NewTagService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/topics/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewTopicRepository()
					service := services.NewTopicService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/users/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewUserRepository()
					service := services.NewUserService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/roles/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewRoleRepository()
					service := services.NewRoleService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/feedbacks/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewFeedbackRepository()
					service := services.NewFeedbackService(repo)
					return service.GetByID(c, id)
				},
			},
			{
				Prefix: "/api/manage/study-shifts/",
				GetFn: func(c *gin.Context, id int) (interface{}, error) {
					repo := repositories.NewStudyShiftRepository()
					service := services.NewStudyShiftService(repo)
					return service.GetByID(c, id)
				},
			},
		}

		skipPaths := []string{"completion", "change-password", "profile", "sync", "sort-lessons", "complete"}
		skipGetAttributeData := false

		for _, p := range skipPaths {
			if strings.Contains(pathUrl, p) {
				skipGetAttributeData = true
				break
			}
		}

		if method == "PUT" && !skipGetAttributeData {
			for _, model := range modelHandlers {
				idStr := path.Base(pathUrl) // lấy phần cuối "4054"
				id, err := strconv.Atoi(idStr)
				if err != nil {
					config.Log.Warn("Invalid ID in path: %v, url: %s", err, pathUrl)
					break
				}

				// Only update cloned questions
				if strings.Contains(model.Prefix, "api/manage/questions/") && (strings.Contains(queryParams, "exam_id=") || strings.Contains(queryParams, "homework_id=")) {
					repo := repositories.NewQuestionRepository()
					service := services.NewQuestionService(repo)

					var assignmentID int
					var assignmentType string

					if strings.Contains(queryParams, "exam_id=") {
						assignmentType = "exam"
						if examIDStr := c.Query("exam_id"); examIDStr != "" {
							if parsedID, err := strconv.Atoi(examIDStr); err == nil {
								assignmentID = parsedID
							} else {
								config.Log.Warn("Invalid exam_id:", err)
								break
							}
						}
					} else if strings.Contains(queryParams, "homework_id=") {
						assignmentType = "homework"
						if homeworkIDStr := c.Query("homework_id"); homeworkIDStr != "" {
							if parsedID, err := strconv.Atoi(homeworkIDStr); err == nil {
								assignmentID = parsedID
							} else {
								config.Log.Warn("Invalid homework_id:", err)
								break
							}
						}
					}

					attribute, err := service.GetClonedByID(c, id, assignmentID, assignmentType)

					if attribute != nil {
						attributeData, err = json.Marshal(attribute)
						if err != nil {
							config.Log.Warn("Marshal error: %v", err)
						}
					}

					break
				} else if strings.HasPrefix(pathUrl, model.Prefix) {
					attribute, err := model.GetFn(c, id)
					if err != nil {
						config.Log.Warn("GetByID error: %v", err)
						break
					}

					if attribute != nil {
						attributeData, err = json.Marshal(attribute)
						if err != nil {
							config.Log.Warn("Marshal error: %v", err)
						}
					}
					break
				}
			}
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Continue processing request
		c.Next()

		// After processing
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		agent := c.Request.UserAgent()

		// Skip by header
		skipLogHeader := c.GetHeader("X-Skip-Log")
		if strings.ToLower(skipLogHeader) == "true" ||
			!strings.HasPrefix(pathUrl, "/api/") ||
			strings.Contains(pathUrl, "/import") ||
			strings.Contains(pathUrl, "/scorm") ||
			strings.Contains(pathUrl, "/api/tus-uploads") ||
			(strings.Contains(pathUrl, "/api/medias") && method == "GET") ||
			pathUrl == "/api/h5p/upload" ||
			pathUrl == "/api/upload-file" {
			return
		}

		// Get userID and roleID
		var userID, roleID *int
		if val, exists := c.Get("userID"); exists {
			id := val.(int)
			userID = &id
		}
		if val, exists := c.Get("roleID"); exists {
			rid := val.(int)
			roleID = &rid
		}

		// chỉ lưu body khi code != 200
		var respBody datatypes.JSON

		if statusCode != http.StatusOK || (statusCode == http.StatusOK && pathUrl == "/api/login") {
			respBody = datatypes.JSON([]byte(blw.body.String()))
		} else {
			respBody = datatypes.JSON([]byte{}) // hoặc nil nếu field cho phép NULL
		}

		// Create log entry
		go func() {
			sessionID := GenerateSessionId(clientIP, agent)
			logEntry := models.ActivityLog{
				UserID:       userID,
				RoleID:       roleID,
				Method:       method,
				Path:         pathUrl,
				StatusCode:   statusCode,
				ResponseBody: respBody,
				ClientIP:     clientIP,
				LatencyMs:    int(latency.Milliseconds()),
				CreatedAt:    time.Now(),
				SessionID:    &sessionID,
				Agent:        &agent,
			}

			// Ignore body/query log for sensitive paths
			if (pathUrl != "/api/login" && pathUrl != "/api/change-password") || statusCode != http.StatusOK {
				if len(requestBodyBytes) > 0 {
					logEntry.RequestBody = datatypes.JSON(requestBodyBytes)
				}
				if queryParams != "" {
					logEntry.QueryParams = datatypes.JSON([]byte(`"` + queryParams + `"`))
				}
			} else if statusCode == http.StatusOK && pathUrl == "/api/login" {
				var resp LoginResponse
				if err := json.Unmarshal(respBody, &resp); err == nil {
					tokenStr := resp.Data.Token
					authRepo := repositories.NewAuthRepository()
					claims, err := utils.CheckToken(c, authRepo, tokenStr)

					if err == nil {
						currentUserID := claims.UserID
						currentRoleID := claims.RoleID

						logEntry.UserID = &currentUserID
						logEntry.RoleID = &currentRoleID
						logEntry.ResponseBody = datatypes.JSON([]byte{})
					}
				}
			}

			// Write attributes if any
			if method == "PUT" && len(attributeData) > 0 {
				logEntry.Attribute = datatypes.JSON(attributeData)
			}

			// Save to DB using monthly table strategy
			if err := repositories.NewMonthlyActivityLogRepository().Create(logEntry); err != nil {
				config.Log.Error("ActivityLoggerMiddleware:")
				config.Log.Error(err)
			}
		}()
	}
}

func GenerateSessionId(ip, userAgent string) string {
	raw := fmt.Sprintf("%s|%s", ip, userAgent)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}
