package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/utils"
	"encoding/json"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

type HomeworkAnswerService interface {
	GetHomeworkAnswer(homeworkID, userID, lessonID int64) (repositories.HomeworkAnswerOverview, []map[string]interface{}, []repositories.ManualAnswer, error)
	GetHomeworkAnswerProto(homeworkID, userID, lessonID int64) (*prot.HomeworkAnswerResponse, error)
}

type homeworkAnswerService struct {
	repo     repositories.HomeworkAnswerRepository
	skipRepo repositories.HomeworkUserSkipQuestionRepository
}

func NewHomeworkAnswerService(repo repositories.HomeworkAnswerRepository) HomeworkAnswerService {
	return &homeworkAnswerService{repo: repo, skipRepo: repositories.NewHomeworkUserSkipQuestionRepository()}
}

func (s *homeworkAnswerService) GetHomeworkAnswer(homeworkID, userID, lessonID int64) (repositories.HomeworkAnswerOverview, []map[string]interface{}, []repositories.ManualAnswer, error) {
	overview, err := s.repo.GetHomeworkAnswerOverview(homeworkID, userID)
	if err != nil {
		return overview, nil, nil, err
	}
	questions, err := s.repo.GetClonedQuestions(homeworkID)
	if err != nil && overview.HomeworkQuestionForm == models.QuestionFormQuestionType {
		return overview, nil, nil, err
	}
	manualAnswers, err := s.repo.GetManualAnswers(homeworkID, userID)
	if err != nil && overview.HomeworkQuestionForm == models.QuestionFormQuestionType {
		return overview, questions, nil, err
	}
	return overview, questions, manualAnswers, nil
}

func (s *homeworkAnswerService) GetHomeworkAnswerProto(homeworkID, userID, lessonID int64) (*prot.HomeworkAnswerResponse, error) {
	overview, questions, manualAnswers, err := s.GetHomeworkAnswer(homeworkID, userID, lessonID)
	if err != nil {
		return nil, err
	}

	protoOverview := &prot.HomeworkAnswerOverview{
		StudentName:             overview.StudentName,
		HomeworkName:            overview.HomeworkName,
		HomeworkDescription:     overview.HomeworkDescription,
		HomeworkStatus:          overview.HomeworkStatus,
		HomeworkCoverImage:      utils.StaticURL(overview.HomeworkCoverImage, "storage"),
		TotalQuestions:          overview.TotalQuestions,
		QuestionsCompleted:      overview.QuestionsCompleted,
		LastQuestionIdCompleted: overview.LastQuestionIDCompleted,
		ManualQuestionsCount:    overview.ManualQuestionsCount,
	}
	var protoQuestions []*structpb.Struct
	hwqRepo := repositories.NewHomeworkUserQuestionRepository()
	// Lấy danh sách question IDs đã trả lời đúng
	completedQuestionIDs, err := s.getCompletedQuestionIDs(homeworkID, userID)

	if err != nil {
		// Nếu có lỗi, sử dụng danh sách rỗng
		completedQuestionIDs = []string{}
	}

	for _, q := range questions {
		// Chuyển tất cả id về dạng string và thêm trường is_completed
		s.processQuestionIDs(q, completedQuestionIDs)

		// Lấy thông tin star, ratio_score, number_time_sent từ homework_user_questions
		if idVal, ok := q["id"]; ok {
			if qidStr := s.convertIDToString(idVal); qidStr != "" {
				if qid, err := strconv.ParseInt(qidStr, 10, 64); err == nil {
					record, err := hwqRepo.GetLatestByHomeworkUserQuestion(homeworkID, userID, lessonID, qid)
					if err == nil && record != nil {
						q["star"] = record.Star
						q["ratio_score"] = record.RatioScore
						q["number_time_sent"] = record.NumberTimeSent
					} else {
						// Không có bản ghi -> trả -1 theo yêu cầu
						q["star"] = -1
						q["ratio_score"] = -1
						q["number_time_sent"] = -1
					}
				}
			}
		}

		// Xử lý media URL trong questions để trả về URL tuyệt đối
		s.processMediaURLsInContent(q)
		if s, err := structpb.NewStruct(q); err == nil {
			protoQuestions = append(protoQuestions, s)
		}
	}
	var protoManualAnswers []*prot.ManualAnswer
	for _, m := range manualAnswers {
		var contentStruct *structpb.Struct
		if len(m.Content) > 0 {
			var contentMap map[string]interface{}
			if err := json.Unmarshal(m.Content, &contentMap); err == nil {
				// Bổ sung hai trường yêu cầu vào content_struct: is_scored, manual_score
				contentMap["is_scored"] = m.Status
				if m.Score != nil {
					contentMap["manual_score"] = *m.Score
				}
				// Xử lý media URL trong content_struct để trả về URL tuyệt đối
				s.processMediaURLsInContent(contentMap)
				contentStruct, _ = structpb.NewStruct(contentMap)
			}
		}
		var manualScore float64
		if m.Score != nil {
			manualScore = *m.Score
		}
		protoManualAnswers = append(protoManualAnswers, &prot.ManualAnswer{
			Id:            m.ID,
			Type:          m.Type,
			Source:        "",
			Status:        m.Status,
			ContentStruct: contentStruct,
			Answer:        stringPtrToString(m.Answer),
			FileUrl:       utils.StaticURL(stringPtrToString(m.FileURL), models.Storage),
			IsScored:      m.Status,
			ManualScore:   manualScore,
		})
	}
	// Lấy ratio và comment
	ratio, _ := s.repo.GetHomeworkRatio(homeworkID, userID)
	comment, _ := s.repo.GetLatestHomeworkComment(homeworkID, userID)
	protoOverview.Ratio = ratio
	protoOverview.LatestComment = comment
	protoOverview.HomeworkQuestionForm = overview.HomeworkQuestionForm

	var questionFiles []*prot.HomeworkAnswerFile
	for _, qf := range overview.HomeworkFiles {
		questionFiles = append(questionFiles, &prot.HomeworkAnswerFile{
			Type: qf.Type,
			Url:  utils.StaticURL(qf.Path, models.Storage),
		})
	}
	protoOverview.QuestionFiles = questionFiles

	// Trả về response, có thể mở rộng proto nếu cần field riêng cho ratio/comment
	// Lấy danh sách câu skip
	skippedIDs, _ := s.skipRepo.ListQuestionIDs(homeworkID, userID)

	// Tính toán is_all_scored: true nếu tất cả câu manual đều đã được chấm
	isAllScored := true
	if len(manualAnswers) > 0 {
		for _, m := range manualAnswers {
			if !m.Status { // Status = false nghĩa là chưa được chấm
				isAllScored = false
				break
			}
		}
	}

	var protoSubmitFiles []*prot.HomeworkAnswerFile

	if overview.HomeworkQuestionForm == models.QuestionFormWriteType {
		var submitFiles models.MediaInfos
		submitFiles, _ = s.repo.GetSubmitFiles(homeworkID, userID)
		for _, qf := range submitFiles {
			protoSubmitFiles = append(protoSubmitFiles, &prot.HomeworkAnswerFile{
				Type: qf.Type,
				Url:  utils.StaticURL(qf.Path, models.Storage),
			})
		}
	}

	return &prot.HomeworkAnswerResponse{
		Overview:           protoOverview,
		Questions:          protoQuestions,
		ManualAnswers:      protoManualAnswers,
		SkippedQuestionIds: skippedIDs,
		IsAllScored:        isAllScored,
		SubmitFiles:        protoSubmitFiles,
		IsSubmitted:        overview.IsSubmitted,
		IsScored:           overview.IsScored,
		Rate:               overview.Rate,
	}, nil
}

func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// processMediaURLsInContent xử lý đệ quy tất cả các URL media trong content_struct để trả về URL tuyệt đối
func (s *homeworkAnswerService) processMediaURLsInContent(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			// Xử lý các trường media phổ biến
			if key == "media" || key == "medias" || key == "content" || key == "options" {
				s.processMediaURLsInContent(value)
			} else if key == "url" && s.isMediaURL(value) {
				// Nếu đây là URL media, chuyển đổi thành URL tuyệt đối
				if urlStr, ok := value.(string); ok && urlStr != "" {
					v[key] = utils.StaticURL(urlStr, models.Storage)
				}
			} else if key == "answers" || key == "sources" || key == "targets" || key == "labels" || key == "items" || key == "questions" || key == "categories" || key == "groups" {
				// Xử lý các mảng chứa media
				s.processMediaURLsInContent(value)
			} else {
				// Tiếp tục đệ quy cho các trường khác
				s.processMediaURLsInContent(value)
			}
		}
	case []interface{}:
		// Xử lý mảng
		for _, item := range v {
			s.processMediaURLsInContent(item)
		}
	}
}

// isMediaURL kiểm tra xem một giá trị có phải là URL media không
func (s *homeworkAnswerService) isMediaURL(value interface{}) bool {
	if urlStr, ok := value.(string); ok {
		// Kiểm tra nếu URL không chứa http/https và không rỗng
		return urlStr != "" && !strings.Contains(urlStr, "http")
	}
	return false
}

// processQuestionIDs chuyển tất cả id về dạng string và thêm trường is_completed
func (s *homeworkAnswerService) processQuestionIDs(data interface{}, completedQuestionIDs []string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Xử lý trường id ở level hiện tại
		if idValue, exists := v["id"]; exists {
			// Chuyển id sang string nếu chưa phải string
			if idStr := s.convertIDToString(idValue); idStr != "" {
				v["id"] = idStr

				// Thêm trường is_completed dựa trên question id
				// Chỉ thêm ở level question (không thêm ở các level nested khác)
				if _, isQuestion := v["type"]; isQuestion {
					v["is_completed"] = s.containsString(completedQuestionIDs, idStr)
				}
			}
		}

		// Đệ quy xử lý tất cả các trường khác
		for key, value := range v {
			// Bỏ qua các trường đặc biệt không cần xử lý id
			if key == "media" || key == "medias" || key == "content" || key == "options" ||
				key == "answers" || key == "sources" || key == "targets" || key == "labels" ||
				key == "items" || key == "questions" || key == "categories" || key == "groups" {
				s.processQuestionIDs(value, completedQuestionIDs)
			} else if key != "id" && key != "is_completed" {
				// Xử lý các trường khác (trừ id và is_completed đã xử lý ở trên)
				s.processQuestionIDs(value, completedQuestionIDs)
			}
		}
	case []interface{}:
		// Xử lý mảng
		for _, item := range v {
			s.processQuestionIDs(item, completedQuestionIDs)
		}
	}
}

// containsString kiểm tra xem một string có trong slice không
func (s *homeworkAnswerService) containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// convertIDToString chuyển id từ các kiểu dữ liệu khác nhau sang string
func (s *homeworkAnswerService) convertIDToString(idValue interface{}) string {
	switch v := idValue.(type) {
	case string:
		return v
	case int:
		return s.intToString(int64(v))
	case int32:
		return s.intToString(int64(v))
	case int64:
		return s.intToString(v)
	case float64:
		// JSON numbers thường được parse thành float64
		return s.intToString(int64(v))
	case float32:
		return s.intToString(int64(v))
	default:
		return ""
	}
}

// intToString chuyển int64 sang string
func (s *homeworkAnswerService) intToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

// getCompletedQuestionIDs lấy danh sách question_id đã trả lời đúng (dạng string)
// Một câu hỏi được tính là đúng khi TẤT CẢ các câu trả lời trong câu hỏi đó đều đúng
// Logic tương tự như updateQuestionsCompletedWithCorrectAnswers nhưng trả về danh sách thay vì cập nhật database
func (s *homeworkAnswerService) getCompletedQuestionIDs(homeworkID, userID int64) ([]string, error) {
	// Lấy danh sách tất cả question_id mà user đã trả lời
	var allAnsweredQuestions []int64

	// Lấy từ tất cả các bảng homework_question_*
	tables := []string{
		"homework_question_users",
		"homework_question_user_positions",
		"homework_question_user_matchings",
		"homework_question_user_manual_scoring",
		"homework_question_user_labelings",
		"homework_question_user_groups",
		"homework_question_user_fill_in_blanks",
	}

	for _, table := range tables {
		var questionIDs []int64
		err := db.MasterDB.Table(table).
			Select("DISTINCT question_id").
			Where("homework_id = ? AND user_id = ?", homeworkID, userID).
			Pluck("question_id", &questionIDs).Error
		if err != nil {
			return nil, err
		}
		allAnsweredQuestions = append(allAnsweredQuestions, questionIDs...)
	}

	// Loại bỏ trùng lặp
	uniqueQuestions := make(map[int64]bool)
	for _, qID := range allAnsweredQuestions {
		uniqueQuestions[qID] = true
	}

	var completedQuestionIDs []string

	// Kiểm tra từng câu hỏi xem có tất cả câu trả lời đều đúng không
	for questionID := range uniqueQuestions {
		isFullyCorrect := true

		// Kiểm tra homework_question_users
		var wrongCount int64
		err := db.MasterDB.Table("homework_question_users").
			Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
				homeworkID, userID, questionID).
			Count(&wrongCount).Error
		if err != nil {
			return nil, err
		}
		if wrongCount > 0 {
			isFullyCorrect = false
		}

		// Kiểm tra homework_question_user_positions
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_positions").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return nil, err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_matchings
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_matchings").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return nil, err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_labelings
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_labelings").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return nil, err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_groups
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_groups").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return nil, err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Kiểm tra homework_question_user_fill_in_blanks
		if isFullyCorrect {
			err = db.MasterDB.Table("homework_question_user_fill_in_blanks").
				Where("homework_id = ? AND user_id = ? AND question_id = ? AND is_correct = false",
					homeworkID, userID, questionID).
				Count(&wrongCount).Error
			if err != nil {
				return nil, err
			}
			if wrongCount > 0 {
				isFullyCorrect = false
			}
		}

		// Nếu tất cả câu trả lời đều đúng, thêm vào danh sách
		if isFullyCorrect {
			completedQuestionIDs = append(completedQuestionIDs, s.intToString(questionID))
		}
	}

	return completedQuestionIDs, nil
}
