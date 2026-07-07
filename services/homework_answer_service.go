package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/json"
	"strings"
	"google.golang.org/protobuf/types/known/structpb"
)

type HomeworkAnswerService interface {
	GetHomeworkAnswer(homeworkID, userID int64) (repositories.HomeworkAnswerOverview, []map[string]interface{}, []repositories.ManualAnswer, error)
	GetHomeworkAnswerProto(homeworkID, userID int64) (*prot.HomeworkAnswerResponse, error)
}

type homeworkAnswerService struct {
    repo repositories.HomeworkAnswerRepository
    skipRepo repositories.HomeworkUserSkipQuestionRepository
}

func NewHomeworkAnswerService(repo repositories.HomeworkAnswerRepository) HomeworkAnswerService {
    return &homeworkAnswerService{repo: repo, skipRepo: repositories.NewHomeworkUserSkipQuestionRepository()}
}

func (s *homeworkAnswerService) GetHomeworkAnswer(homeworkID, userID int64) (repositories.HomeworkAnswerOverview, []map[string]interface{}, []repositories.ManualAnswer, error) {
	overview, err := s.repo.GetHomeworkAnswerOverview(homeworkID, userID)
	if err != nil {
		return overview, nil, nil, err
	}
	questions, err := s.repo.GetClonedQuestions(homeworkID)
	if err != nil {
		return overview, nil, nil, err
	}
	questions = EnrichClonedQuestionMaps(questions)
	manualAnswers, err := s.repo.GetManualAnswers(homeworkID, userID)
	if err != nil {
		return overview, questions, nil, err
	}
	return overview, questions, manualAnswers, nil
}

func (s *homeworkAnswerService) GetHomeworkAnswerProto(homeworkID, userID int64) (*prot.HomeworkAnswerResponse, error) {
	overview, questions, manualAnswers, err := s.GetHomeworkAnswer(homeworkID, userID)
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
		HomeworkQuestionForm:    overview.HomeworkQuestionForm,
		QuestionFiles:           s.convertMediaInfosToProto(overview.QuestionFiles),
	}
	var protoQuestions []*structpb.Struct
	for _, q := range questions {
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

    return &prot.HomeworkAnswerResponse{
        Overview:      protoOverview,
        Questions:     protoQuestions,
        ManualAnswers: protoManualAnswers,
        SkippedQuestionIds: skippedIDs,
        IsAllScored: isAllScored,
        SubmitFiles:   s.convertMediaInfosToProto(overview.SubmitFiles),
    }, nil
}

func (s *homeworkAnswerService) convertMediaInfosToProto(medias models.MediaInfos) []*prot.HomeworkFile {
	var files []*prot.HomeworkFile
	for _, m := range medias {
		files = append(files, &prot.HomeworkFile{
			Type: m.Type,
			Url:  utils.StaticURL(m.Path, models.Storage),
		})
	}
	return files
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
