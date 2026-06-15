package services

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
)

type SaveScoreBulkService interface {
	SaveScoreBulkExam(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error)
	SaveScoreBulkExercise(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error)
	SaveScoreBulkHomework(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error)
	SaveScoreBulkContestRound(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error)
}

type saveScoreBulkService struct {
	serviceMC       SaveScoreMultipleChoiceService
	serviceFB       SaveScoreFillInBlankService
	serviceP        SaveScorePositionService
	serviceM        SaveScoreMatchingService
	serviceL        SaveScoreLabelingService
	serviceG        SaveScoreGroupService
	serviceMS       ManualScoringService
	examUserService ExamUserService
}

func NewSaveScoreBulkService(
	serviceMC SaveScoreMultipleChoiceService,
	serviceFB SaveScoreFillInBlankService,
	serviceP SaveScorePositionService,
	serviceM SaveScoreMatchingService,
	serviceL SaveScoreLabelingService,
	serviceG SaveScoreGroupService,
	serviceMS ManualScoringService,
	examUserService ExamUserService,
) SaveScoreBulkService {
	return &saveScoreBulkService{
		serviceMC:       serviceMC,
		serviceFB:       serviceFB,
		serviceP:        serviceP,
		serviceM:        serviceM,
		serviceL:        serviceL,
		serviceG:        serviceG,
		serviceMS:       serviceMS,
		examUserService: examUserService,
	}
}

func (s *saveScoreBulkService) SaveScoreBulkExam(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error) {
	if req.ExamId == 0 {
		return nil, errors.New("exam_id is required")
	}

	// Lấy thông tin exam
	var exam models.Exam
	if err := db.ReplicaDB.First(&exam, req.ExamId).Error; err != nil {
		return nil, fmt.Errorf("failed to get exam info: %v", err)
	}

	if exam.QuestionForm == models.QuestionFormWriteType {
		return s.SaveBulkWrite(req, userID, models.ClonedQuestionTypeExam)
	}

	if len(req.ListAnswers) == 0 {
		return nil, errors.New("answers is required")
	}

	// Lấy danh sách câu hỏi từ cloned_question 1 lần duy nhất
	clonedQuestionsMap := map[string]repositories.ClonedQuestion{}
	if s.serviceMC != nil {
		if mc, ok := s.serviceMC.(interface{ GetClonedQuestionService() ClonedQuestionService }); ok {
			clonedQuestionsMap, _ = mc.GetClonedQuestionService().GetQuestionsMap(req.ExamId, "exam")
		}
	}
	if len(clonedQuestionsMap) == 0 {
		// fallback nếu serviceMC không có method trên
		clonedQuestionService := NewClonedQuestionService(repositories.NewClonedQuestionRepository())
		clonedQuestionsMap, _ = clonedQuestionService.GetQuestionsMap(req.ExamId, "exam")
	}

	// Xóa hết đáp án cũ của user cho exam_id
	tx := db.MasterDB.Begin()
	if err := repositories.DeleteAllExamAnswersByExamIDAndUserID(req.ExamId, userID, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	response := &prot.SaveScoreBulkResponse{
		MultipleChoice: make([]*prot.SaveScoreResponse, 0),
		FillInBlank:    make([]*prot.SaveScoreResponseFillInBlank, 0),
		Ordering:       make([]*prot.SaveScoreResponsePosition, 0),
		Dragdrop:       make([]*prot.SaveScoreResponsePosition, 0),
		Matching:       make([]*prot.SaveScoreResponseMatching, 0),
		Labeling:       make([]*prot.SaveScoreResponseLabeling, 0),
		Category:       make([]*prot.SaveScoreResponseGroup, 0),
		Time:           req.Time,
		ExamName:       exam.Name,
		Manual:         make([]*prot.SaveScoreResponseManual, 0),
	}

	// Khởi tạo biến hasManualScoring = true
	hasManualScoring := false

	for _, answer := range req.ListAnswers {
		switch answer.TypeQuestion {
		case "multiple_choice":
			var answerIDs []int64
			if raw, ok := answer.Data["answer_ids"]; ok {
				b, err := protojson.Marshal(raw)
				if err != nil {
					return nil, fmt.Errorf("cannot marshal answer_ids for question %d: %v", answer.QuestionId, err)
				}

				if err := json.Unmarshal(b, &answerIDs); err != nil {
					return nil, fmt.Errorf("invalid answer_ids format for question %d: %v", answer.QuestionId, err)
				}
			}

			ids := make([]int64, len(answerIDs))
			copy(ids, answerIDs)

			reqMC := &prot.SaveScoreMultipleChoiceRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				AnswerIds:  ids,
			}
			resp, err := s.serviceMC.SaveScoreMultipleChoiceExam(reqMC, userID)
			if err != nil {
				return nil, err
			}
			response.MultipleChoice = append(response.MultipleChoice, resp)
			response.TotalScore += resp.Score

		case "fill_in_blank":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			answersInterface := answersVal.AsInterface()

			answers, ok := answersInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]string)
			for k, v := range answers {
				answerMap[k] = fmt.Sprintf("%v", v)
			}
			reqFB := &prot.SaveScoreFillInBlankRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceFB.SaveScoreFillInBlankExam(reqFB, userID)
			if err != nil {
				return nil, err
			}
			response.FillInBlank = append(response.FillInBlank, resp)
			response.TotalScore += resp.TotalScore

		case "ordering":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			answersInterface := answersVal.AsInterface()

			rawMap, ok := answersInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int32)
			for k, v := range rawMap {
				if num, ok := v.(float64); ok {
					answerMap[k] = int32(num)
				} else {
					return nil, fmt.Errorf("invalid value type for key %s in question %d", k, answer.QuestionId)
				}
			}

			reqO := &prot.SaveScorePositionRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceP.SaveScorePositionExam(reqO, userID)
			if err != nil {
				return nil, err
			}
			response.Ordering = append(response.Ordering, resp)
			response.TotalScore += resp.TotalScore

		case "dragdrop":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			rawInterface := answersVal.AsInterface()

			rawMap, ok := rawInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int32, len(rawMap))
			for k, v := range rawMap {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid value type for key %s in question %d", k, answer.QuestionId)
				}
				answerMap[k] = int32(num)
			}

			reqD := &prot.SaveScorePositionRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceP.SaveScorePositionExam(reqD, userID)
			if err != nil {
				return nil, err
			}
			response.Dragdrop = append(response.Dragdrop, resp)
			response.TotalScore += resp.TotalScore

		case "matching":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			rawInterface := answersVal.AsInterface()

			rawMap, ok := rawInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int64, len(rawMap))
			for k, v := range rawMap {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid value type for key %s in question %d", k, answer.QuestionId)
				}
				answerMap[k] = int64(num)
			}

			reqM := &prot.SaveScoreMatchingRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceM.SaveScoreMatchingExam(reqM, userID)
			if err != nil {
				return nil, err
			}
			response.Matching = append(response.Matching, resp)
			response.TotalScore += resp.TotalScore

		case "labeling":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			rawInterface := answersVal.AsInterface()
			rawMap, ok := rawInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int64, len(rawMap))
			for k, v := range rawMap {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid value type for key %s in question %d", k, answer.QuestionId)
				}
				answerMap[k] = int64(num)
			}

			reqL := &prot.SaveScoreLabelingRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceL.SaveScoreLabelingExam(reqL, userID)
			if err != nil {
				return nil, err
			}
			response.Labeling = append(response.Labeling, resp)
			response.TotalScore += resp.TotalScore

		case "category":
			answersVal, ok := answer.Data["answers"]
			if !ok {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			rawInterface := answersVal.AsInterface()
			rawMap, ok := rawInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int64, len(rawMap))
			for k, v := range rawMap {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid value type for key %s in question %d", k, answer.QuestionId)
				}
				answerMap[k] = int64(num)
			}

			reqG := &prot.SaveScoreGroupRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceG.SaveScoreGroupExam(reqG, userID)
			if err != nil {
				return nil, err
			}
			response.Category = append(response.Category, resp)
			response.TotalScore += resp.TotalScore

		case "manual":
			var answerContent, fileURL string
			if val, ok := answer.Data["answer"]; ok && val != nil {
				if str, ok := val.AsInterface().(string); ok {
					answerContent = str
				}
			}

			if val, ok := answer.Data["file_url"]; ok && val != nil {
				if str, ok := val.AsInterface().(string); ok {
					fileURL = str
				}
			}

			fileUrl := utils.StripDomain(fileURL, models.Storage)

			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			// Lưu câu trả lời cần chấm điểm thủ công
			reqManual := &prot.SaveAnswerManualRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				Answer:     answerContent,
				FileUrl:    fileUrl,
				FileInfo: &prot.FileInfo{
					Id:   fileInfo.Id,
					Path: fileInfo.Path,
					Disk: fileInfo.Disk,
				},
			}
			if err := s.serviceMS.SaveExamManualScoringService(*reqManual, userID); err != nil {
				return nil, err
			}

			// Thêm response cho câu hỏi manual
			manualResp := &prot.SaveScoreResponseManual{
				QuestionId: answer.QuestionId,
				Answer:     answerContent,
				FileUrl:    fileURL,
				IsSaved:    true,
			}
			response.Manual = append(response.Manual, manualResp)

			// Cập nhật hasManualScoring = true khi có câu hỏi manual
			hasManualScoring = true

		default:
			return nil, errors.New("invalid type_question")
		}
	}

	// Lưu kết quả vào bảng exam_users với trạng thái hasManualScoring
	examUser := &models.ExamUser{
		ExamID:           req.ExamId,
		UserID:           userID,
		Score:            &response.TotalScore,
		Time:             &req.Time,
		HasManualScoring: hasManualScoring,
	}
	if err := s.examUserService.SaveExamUser(examUser); err != nil {
		return nil, err
	}

	ratio, err := s.examUserService.CalculateExamRatioService(req.ExamId, userID)
	if err != nil {
		return nil, err
	}
	if err := s.examUserService.SaveExamRatioService(req.ExamId, userID, ratio); err != nil {
		return nil, err
	}
	response.Ratio = ratio
	return response, nil
}

func (s *saveScoreBulkService) SaveScoreBulkHomework(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error) {
	if req.HomeworkId == 0 {
		return nil, errors.New("homework_id is required")
	}

	// Lấy thông tin homework
	var homework models.Homework
	if err := db.ReplicaDB.First(&homework, req.HomeworkId).Error; err != nil {
		return nil, fmt.Errorf("failed to get homework info: %v", err)
	}

	if homework.QuestionForm == models.QuestionFormWriteType {
		return s.SaveBulkWrite(req, userID, models.ClonedQuestionTypeHomework)
	}

	if len(req.ListAnswers) == 0 {
		return nil, errors.New("answers is required")
	}

	response := &prot.SaveScoreBulkResponse{
		MultipleChoice: make([]*prot.SaveScoreResponse, 0),
		FillInBlank:    make([]*prot.SaveScoreResponseFillInBlank, 0),
		Ordering:       make([]*prot.SaveScoreResponsePosition, 0),
		Dragdrop:       make([]*prot.SaveScoreResponsePosition, 0),
		Matching:       make([]*prot.SaveScoreResponseMatching, 0),
		Labeling:       make([]*prot.SaveScoreResponseLabeling, 0),
		Category:       make([]*prot.SaveScoreResponseGroup, 0),
		Manual:         make([]*prot.SaveScoreResponseManual, 0),
	}

	//hasManualScoring := false

	for _, answer := range req.ListAnswers {
		switch answer.TypeQuestion {
		case "multiple_choice":
			answerIDs := []int64{}

			if raw, ok := answer.Data["answer_ids"]; ok && raw != nil {
				list := raw.GetListValue()
				if list == nil {
					return nil, fmt.Errorf("answer_ids is not a list for question %d", answer.QuestionId)
				}

				for _, v := range list.Values {
					f := v.GetNumberValue()
					answerIDs = append(answerIDs, int64(f))
				}
			}

			// copy vào slice mới nếu cần
			ids := make([]int64, len(answerIDs))
			copy(ids, answerIDs)

			reqMC := &prot.SaveScoreMultipleChoiceRequest{
				ExamId:     req.ExamId,
				QuestionId: answer.QuestionId,
				AnswerIds:  ids,
			}
			resp, err := s.serviceMC.SaveScoreMultipleChoiceHomework(reqMC, userID)
			if err != nil {
				return nil, err
			}
			response.MultipleChoice = append(response.MultipleChoice, resp)
			response.TotalScore += resp.Score

		case "fill_in_blank":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// answersVal là *structpb.Value
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]string)
			for k, v := range structAnswers.Fields {
				// mỗi v là *structpb.Value, chuyển thành string
				answerMap[k] = v.GetStringValue()
			}
			reqFB := &prot.SaveScoreFillInBlankRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceFB.SaveScoreFillInBlankHomework(reqFB, userID)
			if err != nil {
				return nil, err
			}
			response.FillInBlank = append(response.FillInBlank, resp)
			response.TotalScore += resp.TotalScore

		case "ordering":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// Lấy struct từ *structpb.Value
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			answerMap := make(map[string]int32)
			for k, v := range structAnswers.Fields {
				answerMap[k] = int32(v.GetNumberValue())
			}

			reqO := &prot.SaveScorePositionRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceP.SaveScorePositionHomework(reqO, userID)
			if err != nil {
				return nil, err
			}
			response.Ordering = append(response.Ordering, resp)
			response.TotalScore += resp.TotalScore

		case "dragdrop":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// answersVal là *structpb.Value, phải chuyển sang Struct
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			// Lấy map[string]int32
			answerMap := make(map[string]int32, len(structAnswers.Fields))
			for k, v := range structAnswers.Fields {
				answerMap[k] = int32(v.GetNumberValue())
			}
			reqD := &prot.SaveScorePositionRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceP.SaveScorePositionHomework(reqD, userID)
			if err != nil {
				return nil, err
			}
			response.Dragdrop = append(response.Dragdrop, resp)
			response.TotalScore += resp.TotalScore

		case "matching":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// Chuyển sang Struct
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			// Tạo map[string]int64
			answerMap := make(map[string]int64, len(structAnswers.Fields))
			for k, v := range structAnswers.Fields {
				answerMap[k] = int64(v.GetNumberValue())
			}

			reqM := &prot.SaveScoreMatchingRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceM.SaveScoreMatchingHomework(reqM, userID)
			if err != nil {
				return nil, err
			}
			response.Matching = append(response.Matching, resp)
			response.TotalScore += resp.TotalScore

		case "labeling":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// Chuyển sang Struct
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			// Tạo map[string]int64
			answerMap := make(map[string]int64, len(structAnswers.Fields))
			for k, v := range structAnswers.Fields {
				answerMap[k] = int64(v.GetNumberValue())
			}

			reqL := &prot.SaveScoreLabelingRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceL.SaveScoreLabelingHomework(reqL, userID)
			if err != nil {
				return nil, err
			}
			response.Labeling = append(response.Labeling, resp)
			response.TotalScore += resp.TotalScore

		case "category":
			answersVal, ok := answer.Data["answers"]
			if !ok || answersVal == nil {
				return nil, fmt.Errorf("missing answers for question %d", answer.QuestionId)
			}

			// Chuyển sang Struct
			structAnswers := answersVal.GetStructValue()
			if structAnswers == nil {
				return nil, fmt.Errorf("answers is not a struct for question %d", answer.QuestionId)
			}

			// Tạo map[string]int64
			answerMap := make(map[string]int64, len(structAnswers.Fields))
			for k, v := range structAnswers.Fields {
				answerMap[k] = int64(v.GetNumberValue())
			}

			reqG := &prot.SaveScoreGroupRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answers:    answerMap,
			}
			resp, err := s.serviceG.SaveScoreGroupHomework(reqG, userID)
			if err != nil {
				return nil, err
			}
			response.Category = append(response.Category, resp)
			response.TotalScore += resp.TotalScore

		case "manual":
			// Lấy dữ liệu từ data, cho phép trống
			var answerContent, fileURL string

			if val, ok := answer.Data["answer"]; ok && val != nil {
				if s := val.GetStringValue(); s != "" {
					answerContent = s
				}
			}

			if val, ok := answer.Data["file_url"]; ok && val != nil {
				if s := val.GetStringValue(); s != "" {
					fileURL = s
				}
			}

			fileUrl := utils.StripDomain(fileURL, models.Storage)

			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

			// Lưu câu trả lời cần chấm điểm thủ công
			reqManual := &prot.SaveAnswerManualRequest{
				HomeworkId: req.HomeworkId,
				LessonId:   req.LessonId,
				QuestionId: answer.QuestionId,
				Answer:     answerContent,
				FileUrl:    fileUrl,
				FileInfo: &prot.FileInfo{
					Id:   fileInfo.Id,
					Path: fileInfo.Path,
					Disk: fileInfo.Disk,
				},
			}
			if err := s.serviceMS.SaveHomeworkManualScoringService(*reqManual, userID); err != nil {
				return nil, err
			}

			// Thêm response cho câu hỏi manual
			manualResp := &prot.SaveScoreResponseManual{
				QuestionId: answer.QuestionId,
				Answer:     answerContent,
				FileUrl:    fileURL,
				IsSaved:    true,
			}
			response.Manual = append(response.Manual, manualResp)

			// Cập nhật hasManualScoring = true khi có câu hỏi manual
			//hasManualScoring = true

		default:
			return nil, errors.New("invalid type_question")
		}
	}

	// TẠM THỜI: stub cho exercise sẽ reuse logic homework
	return response, nil
}

func (s *saveScoreBulkService) SaveScoreBulkExercise(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error) {
	if req.ExerciseId == 0 {
		return nil, errors.New("exercise_id is required")
	}

	// Lấy thông tin exercise để trả tên
	var exercise models.Exercise
	if err := db.ReplicaDB.First(&exercise, req.ExerciseId).Error; err != nil {
		return nil, fmt.Errorf("failed to get exercise info: %v", err)
	}

	if exercise.QuestionForm == models.QuestionFormWriteType {
		return s.SaveBulkWrite(req, userID, models.ClonedQuestionTypeExercise)
	}

	if len(req.ListAnswers) == 0 {
		return nil, errors.New("answers is required")
	}

	clonedQuestionsMap := map[string]repositories.ClonedQuestion{}
	if s.serviceMC != nil {
		if mc, ok := s.serviceMC.(interface{ GetClonedQuestionService() ClonedQuestionService }); ok {
			clonedQuestionsMap, _ = mc.GetClonedQuestionService().GetQuestionsMap(req.ExerciseId, "exercise")
		}
	}
	if len(clonedQuestionsMap) == 0 {
		clonedQuestionService := NewClonedQuestionService(repositories.NewClonedQuestionRepository())
		clonedQuestionsMap, _ = clonedQuestionService.GetQuestionsMap(req.ExerciseId, "exercise")
	}

	tx := db.MasterDB.Begin()
	if err := repositories.DeleteAllExerciseAnswersByExerciseIDAndUserID(req.ExerciseId, userID, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	response := &prot.SaveScoreBulkResponse{
		MultipleChoice: make([]*prot.SaveScoreResponse, 0),
		FillInBlank:    make([]*prot.SaveScoreResponseFillInBlank, 0),
		Ordering:       make([]*prot.SaveScoreResponsePosition, 0),
		Dragdrop:       make([]*prot.SaveScoreResponsePosition, 0),
		Matching:       make([]*prot.SaveScoreResponseMatching, 0),
		Labeling:       make([]*prot.SaveScoreResponseLabeling, 0),
		Category:       make([]*prot.SaveScoreResponseGroup, 0),
		Manual:         make([]*prot.SaveScoreResponseManual, 0),
		Time:           req.Time,
		ExamName:       "",
		ExerciseName:   exercise.Name,
	}

	hasManualScoring := false

	for _, answer := range req.ListAnswers {
		switch answer.TypeQuestion {
		case "multiple_choice":
			var ids []int64
			if raw, ok := answer.Data["answer_ids"]; ok && raw != nil {
				arr, ok := raw.AsInterface().([]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid answer_ids format for question %d", answer.QuestionId)
				}
				ids = make([]int64, len(arr))
				for i, v := range arr {
					num, ok := v.(float64)
					if !ok {
						return nil, fmt.Errorf("invalid value type in answer_ids for question %d", answer.QuestionId)
					}
					ids[i] = int64(num)
				}
			}

			reqMC := &prot.SaveScoreMultipleChoiceRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				AnswerIds:  ids,
			}
			resp, err := s.serviceMC.SaveScoreMultipleChoiceExercise(reqMC, userID)
			if err != nil {
				return nil, err
			}
			response.MultipleChoice = append(response.MultipleChoice, resp)
			response.TotalScore += resp.Score

		case "fill_in_blank":
			var m map[string]string

			if raw, ok := answer.Data["answers"]; ok && raw != nil {
				tempInterface := raw.AsInterface()
				temp, ok := tempInterface.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
				}

				m = make(map[string]string, len(temp))
				for k, v := range temp {
					m[k] = fmt.Sprintf("%v", v)
				}
			}

			reqFB := &prot.SaveScoreFillInBlankRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceFB.SaveScoreFillInBlankExercise(reqFB, userID)
			if err != nil {
				return nil, err
			}
			response.FillInBlank = append(response.FillInBlank, resp)
			response.TotalScore += resp.TotalScore

		case "ordering":
			raw, ok := answer.Data["answers"]
			if !ok || raw == nil {
				return nil, fmt.Errorf("answers not found for question %d", answer.QuestionId)
			}

			tempInterface := raw.AsInterface()
			temp, ok := tempInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			m := make(map[string]int32, len(temp))
			for k, v := range temp {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid answer value type for key %s in question %d", k, answer.QuestionId)
				}
				m[k] = int32(num)
			}

			reqP := &prot.SaveScorePositionRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceP.SaveScorePositionExercise(reqP, userID)
			if err != nil {
				return nil, err
			}
			response.Ordering = append(response.Ordering, resp)
			response.TotalScore += resp.TotalScore

		case "dragdrop":
			raw, ok := answer.Data["answers"]
			if !ok || raw == nil {
				return nil, fmt.Errorf("answers not found for question %d", answer.QuestionId)
			}

			tempInterface := raw.AsInterface()
			temp, ok := tempInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			m := make(map[string]int32, len(temp))
			for k, v := range temp {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid answer value type for key %s in question %d", k, answer.QuestionId)
				}
				m[k] = int32(num)
			}

			reqD := &prot.SaveScorePositionRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceP.SaveScorePositionExercise(reqD, userID)
			if err != nil {
				return nil, err
			}
			response.Dragdrop = append(response.Dragdrop, resp)
			response.TotalScore += resp.TotalScore

		case "matching":
			raw, ok := answer.Data["answers"]
			if !ok || raw == nil {
				return nil, fmt.Errorf("answers not found for question %d", answer.QuestionId)
			}

			tempInterface := raw.AsInterface()
			temp, ok := tempInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			m := make(map[string]int64, len(temp))
			for k, v := range temp {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid answer value type for key %s in question %d", k, answer.QuestionId)
				}
				m[k] = int64(num)
			}

			reqM := &prot.SaveScoreMatchingRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceM.SaveScoreMatchingExercise(reqM, userID)
			if err != nil {
				return nil, err
			}
			response.Matching = append(response.Matching, resp)
			response.TotalScore += resp.TotalScore

		case "labeling":
			raw, ok := answer.Data["answers"]
			if !ok || raw == nil {
				return nil, fmt.Errorf("answers not found for question %d", answer.QuestionId)
			}

			tempInterface := raw.AsInterface()
			temp, ok := tempInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			m := make(map[string]int64, len(temp))
			for k, v := range temp {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid answer value type for key %s in question %d", k, answer.QuestionId)
				}
				m[k] = int64(num)
			}

			reqL := &prot.SaveScoreLabelingRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceL.SaveScoreLabelingExercise(reqL, userID)
			if err != nil {
				return nil, err
			}
			response.Labeling = append(response.Labeling, resp)
			response.TotalScore += resp.TotalScore

		case "category":
			raw, ok := answer.Data["answers"]
			if !ok || raw == nil {
				return nil, fmt.Errorf("answers not found for question %d", answer.QuestionId)
			}

			tempInterface := raw.AsInterface()
			temp, ok := tempInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answers format for question %d", answer.QuestionId)
			}

			m := make(map[string]int64, len(temp))
			for k, v := range temp {
				num, ok := v.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid answer value type for key %s in question %d", k, answer.QuestionId)
				}
				m[k] = int64(num)
			}

			reqG := &prot.SaveScoreGroupRequest{
				ExerciseId: req.ExerciseId,
				QuestionId: answer.QuestionId,
				Answers:    m,
			}
			resp, err := s.serviceG.SaveScoreGroupExercise(reqG, userID)
			if err != nil {
				return nil, err
			}
			response.Category = append(response.Category, resp)
			response.TotalScore += resp.TotalScore

		case "manual":
			var answerContent, fileURL string

			if val, ok := answer.Data["answer"]; ok && val != nil {
				if str, ok := val.AsInterface().(string); ok {
					answerContent = str
				}
			}

			if val, ok := answer.Data["file_url"]; ok && val != nil {
				if str, ok := val.AsInterface().(string); ok {
					fileURL = str
				}
			}

			fileUrl := utils.StripDomain(fileURL, models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
			if err := db.MasterDB.Table("exercise_question_user_manual_scoring").
				Create(map[string]interface{}{"exercise_id": req.ExerciseId, "user_id": userID, "question_id": answer.QuestionId, "answer": answerContent, "file_info": fileInfo, "is_scored": false}).Error; err != nil {
				return nil, err
			}
			response.Manual = append(response.Manual, &prot.SaveScoreResponseManual{QuestionId: answer.QuestionId, Answer: answerContent, FileUrl: fileURL, IsSaved: true})
			hasManualScoring = true

		default:
			return nil, errors.New("invalid type_question")
		}
	}

	// Save exercise_users + ratio
	exerciseUser := &models.ExerciseUser{ExerciseID: req.ExerciseId, UserID: userID, Score: &response.TotalScore, Time: &req.Time, HasManualScoring: hasManualScoring}
	if err := s.examUserService.SaveExerciseUser(exerciseUser); err != nil {
		return nil, err
	}
	ratio, err := s.examUserService.CalculateExerciseRatioService(req.ExerciseId, userID)
	if err != nil {
		return nil, err
	}
	if err := s.examUserService.SaveExerciseRatioService(req.ExerciseId, userID, ratio); err != nil {
		return nil, err
	}
	response.Ratio = ratio
	return response, nil
}

func (s *saveScoreBulkService) SaveScoreBulkContestRound(req *prot.SaveScoreBulkRequest, userID int64) (*prot.SaveScoreBulkResponse, error) {
	fmt.Printf("🔍 SaveScoreBulkContestRound: ContestRoundId=%d, ListAnswers=%d, UserID=%d\n", req.ContestRoundId, len(req.ListAnswers), userID)

	if req.ContestRoundId == 0 {
		return nil, errors.New("contest_round_id is required")
	}

	// Lấy thông tin contest round
	var contestRound models.ContestRound
	if err := db.ReplicaDB.First(&contestRound, req.ContestRoundId).Error; err != nil {
		return nil, fmt.Errorf("failed to get contest round info: %v", err)
	}

	if contestRound.QuestionForm == models.QuestionFormWriteType {
		return s.SaveBulkWrite(req, userID, models.ClonedQuestionTypeContestRound)
	}

	if len(req.ListAnswers) == 0 {
		return nil, errors.New("answers is required")
	}

	// Get contest round questions from cloned_questions table
	fmt.Printf("🔍 Getting contest round questions for ContestRoundId=%d\n", req.ContestRoundId)
	var clonedQuestion models.ClonedQuestion
	err := db.ReplicaDB.Where("assignment_id = ? AND assignment_type = ?", req.ContestRoundId, "contest_round").First(&clonedQuestion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get contest round questions: %v", err)
	}

	// Parse questions from JSONB
	var questions []map[string]interface{}
	if err := json.Unmarshal(clonedQuestion.Questions, &questions); err != nil {
		return nil, fmt.Errorf("failed to parse contest round questions: %v", err)
	}

	// Create a map for quick question lookup
	questionsMap := make(map[int64]map[string]interface{})
	for _, q := range questions {
		if id, ok := q["id"].(string); ok {
			if questionID, err := strconv.ParseInt(id, 10, 64); err == nil {
				questionsMap[questionID] = q
			}
		}
	}
	fmt.Printf("🔍 Loaded %d questions from contest round\n", len(questionsMap))

	response := &prot.SaveScoreBulkResponse{
		MultipleChoice:   make([]*prot.SaveScoreResponse, 0),
		FillInBlank:      make([]*prot.SaveScoreResponseFillInBlank, 0),
		Ordering:         make([]*prot.SaveScoreResponsePosition, 0),
		Dragdrop:         make([]*prot.SaveScoreResponsePosition, 0),
		Matching:         make([]*prot.SaveScoreResponseMatching, 0),
		Labeling:         make([]*prot.SaveScoreResponseLabeling, 0),
		Category:         make([]*prot.SaveScoreResponseGroup, 0),
		Manual:           make([]*prot.SaveScoreResponseManual, 0),
		ContestRoundId:   req.ContestRoundId,
		ContestRoundName: "", // Will be filled later if needed
	}

	totalScore := 0.0
	hasManualScoring := false

	// Process each answer - support all question types for contest rounds using correct tables
	for _, answer := range req.ListAnswers {
		fmt.Printf("🔍 Processing answer: TypeQuestion=%s, QuestionId=%d\n", answer.TypeQuestion, answer.QuestionId)
		switch answer.TypeQuestion {
		case "multiple_choice":
			// Get answer_ids from data
			var answerIDs []int64
			if raw, ok := answer.Data["answer_ids"]; ok && raw != nil {
				list := raw.GetListValue()
				if list != nil {
					for _, val := range list.Values {
						answerIDs = append(answerIDs, int64(val.GetNumberValue()))
					}
				}
			}

			// Get question info to calculate score
			questionData := questionsMap[answer.QuestionId]
			isCorrect := false
			score := 0.0

			if questionData != nil {
				// Get correct answers from question
				if options, ok := questionData["options"].(map[string]interface{}); ok {
					if answers, ok := options["answers"].([]interface{}); ok {
						correctAnswerIDs := make(map[int64]bool)
						var maxScore float64 = 1.0 // Default score

						// Get metadata for points
						if metadata, ok := questionData["metadata"].(map[string]interface{}); ok {
							if points, ok := metadata["points"].(float64); ok {
								maxScore = points
							}
						}

						// Find correct answers
						for _, ans := range answers {
							if ansMap, ok := ans.(map[string]interface{}); ok {
								if isCorrectVal, ok := ansMap["is_correct"].(bool); ok && isCorrectVal {
									if idStr, ok := ansMap["id"].(string); ok {
										if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
											correctAnswerIDs[id] = true
										}
									}
								}
							}
						}

						// Check if answer is correct (all selected answers must be correct)
						if len(answerIDs) > 0 && len(answerIDs) == len(correctAnswerIDs) {
							allCorrect := true
							for _, aid := range answerIDs {
								if !correctAnswerIDs[aid] {
									allCorrect = false
									break
								}
							}
							isCorrect = allCorrect
							if isCorrect {
								score = maxScore
							}
						}
					}
				}
			}

			// Save to contest_round_question_user_multiple_choices table
			for _, answerID := range answerIDs {
				if err := db.MasterDB.Table("contest_round_question_user_multiple_choices").
					Create(map[string]interface{}{
						"contest_round_id": req.ContestRoundId,
						"user_id":          userID,
						"question_id":      answer.QuestionId,
						"answer_id":        answerID,
						"is_correct":       isCorrect,
						"score":            score,
					}).Error; err != nil {
					return nil, fmt.Errorf("failed to save multiple choice answer: %v", err)
				}
			}

			response.MultipleChoice = append(response.MultipleChoice, &prot.SaveScoreResponse{
				QuestionId: answer.QuestionId,
				AnswerIds:  answerIDs, // Use all answer IDs
				IsCorrect:  isCorrect,
				Score:      score,
			})

		case "fill_in_blanks", "fill_in_blank":
			answers := make(map[string]string)

			if raw, ok := answer.Data["answers"]; ok && raw != nil {
				obj := raw.GetStructValue()
				if obj != nil {
					for key, value := range obj.Fields {
						answers[key] = value.GetStringValue()
					}
				}
			}

			// Save to contest_round_question_user_fill_in_blanks table
			for position, answerText := range answers {
				pos, _ := strconv.Atoi(position)
				if err := db.MasterDB.Table("contest_round_question_user_fill_in_blanks").
					Create(map[string]interface{}{
						"contest_round_id": req.ContestRoundId,
						"user_id":          userID,
						"question_id":      answer.QuestionId,
						"answer":           answerText,
						"sort_position":    pos,
						"is_correct":       false,
						"score":            0.0,
					}).Error; err != nil {
					return nil, err
				}
			}

			// Create a simple response
			response.FillInBlank = append(response.FillInBlank, &prot.SaveScoreResponseFillInBlank{
				QuestionId:   answer.QuestionId,
				TotalScore:   0.0,
				IsAllCorrect: false,
			})

		case "ordering", "drag_drop", "dragdrop":
			answers := make(map[string]int64)

			if raw, ok := answer.Data["answers"]; ok && raw != nil {
				obj := raw.GetStructValue()
				if obj != nil {
					for key, value := range obj.Fields {
						answers[key] = int64(value.GetNumberValue())
					}
				}
			}

			// Save to contest_round_question_user_positions table
			for answerID, position := range answers {
				answerIDInt, _ := strconv.ParseInt(answerID, 10, 64)
				if err := db.MasterDB.Table("contest_round_question_user_positions").
					Create(map[string]interface{}{
						"contest_round_id":      req.ContestRoundId,
						"user_id":               userID,
						"question_id":           answer.QuestionId,
						"answer_group_position": answerIDInt,
						"sort_position":         position,
						"is_correct":            false,
						"score":                 0.0,
					}).Error; err != nil {
					return nil, err
				}
			}

			// Create a simple response
			resp := &prot.SaveScoreResponsePosition{
				QuestionId:   answer.QuestionId,
				TotalScore:   0.0,
				IsAllCorrect: false,
			}

			if answer.TypeQuestion == "ordering" {
				response.Ordering = append(response.Ordering, resp)
			} else {
				response.Dragdrop = append(response.Dragdrop, resp)
			}

		case "labeling":
			fmt.Printf("🔍 Processing labeling question: %d\n", answer.QuestionId)
			answers := make(map[string]int64)

			if raw, ok := answer.Data["answers"]; ok && raw != nil {
				obj := raw.GetStructValue()
				if obj != nil {
					for key, value := range obj.Fields {
						answers[key] = int64(value.GetNumberValue())
					}
				}
			}

			// Get correct answers from question data
			var correctAnswers map[string]string
			var totalScore float64 = 0.0
			var isAllCorrect bool = true
			var correctCount int = 0
			var totalCount int = len(answers)

			if questionData, exists := questionsMap[answer.QuestionId]; exists {
				if correctAnswersData, ok := questionData["correct_answers"].(map[string]interface{}); ok {
					if listData, ok := correctAnswersData["list"].(map[string]interface{}); ok {
						correctAnswers = make(map[string]string)
						for k, v := range listData {
							if str, ok := v.(string); ok {
								correctAnswers[k] = str
							}
						}
					}
				}
			}

			fmt.Printf("🔍 User answers: %+v, Correct answers: %+v\n", answers, correctAnswers)

			// Save to contest_round_question_user_labelings table and calculate score
			for blankID, answerID := range answers {
				blankIDInt, _ := strconv.ParseInt(blankID, 10, 64)
				isCorrect := false
				score := 0.0

				// Check if answer is correct
				if correctAnswer, exists := correctAnswers[blankID]; exists {
					if strconv.FormatInt(answerID, 10) == correctAnswer {
						isCorrect = true
						score = 1.0 // Assuming 1 point per correct answer
						correctCount++
					}
				}

				if !isCorrect {
					isAllCorrect = false
				}

				totalScore += score

				if err := db.MasterDB.Table("contest_round_question_user_labelings").
					Create(map[string]interface{}{
						"contest_round_id": req.ContestRoundId,
						"question_id":      answer.QuestionId,
						"user_id":          userID,
						"blank_id":         blankIDInt,
						"answer_id":        answerID,
						"is_correct":       isCorrect,
						"score":            score,
					}).Error; err != nil {
					return nil, err
				}
			}

			fmt.Printf("🔍 Labeling question %d: Score=%f, Correct=%d/%d, IsAllCorrect=%t\n", answer.QuestionId, totalScore, correctCount, totalCount, isAllCorrect)

			response.Labeling = append(response.Labeling, &prot.SaveScoreResponseLabeling{
				QuestionId:   answer.QuestionId,
				TotalScore:   totalScore,
				IsAllCorrect: isAllCorrect,
			})

		case "writing", "speaking":
			// Manual scoring for contest round
			hasManualScoring = true // Set flag for manual scoring

			answerContent := ""
			if raw, ok := answer.Data["answer"]; ok && raw != nil {
				answerContent = raw.GetStringValue()
			}

			fileURL := ""
			if raw, ok := answer.Data["file_url"]; ok && raw != nil {
				if str := raw.GetStringValue(); str != "" {
					fileURL = str
				}
			}

			fileUrl := utils.StripDomain(fileURL, models.Storage)
			mediaRepo := repositories.NewMediaRepository()
			fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
			if err := db.MasterDB.Table("contest_round_question_user_manual_scoring").
				Create(map[string]interface{}{"contest_round_id": req.ContestRoundId, "user_id": userID, "question_id": answer.QuestionId, "answer": answerContent, "file_info": fileInfo, "is_scored": false}).Error; err != nil {
				return nil, err
			}
			response.Manual = append(response.Manual, &prot.SaveScoreResponseManual{QuestionId: answer.QuestionId, Answer: answerContent, FileUrl: fileURL, IsSaved: true})

		default:
			// For other question types, return error for now
			return nil, fmt.Errorf("question type %s not yet supported for contest rounds", answer.TypeQuestion)
		}
	}

	// Calculate total score and ratio
	fmt.Printf("🔍 Calculating total score from responses...\n")
	for _, mc := range response.MultipleChoice {
		totalScore += mc.Score
		fmt.Printf("🔍 MultipleChoice score: %f\n", mc.Score)
	}
	for _, fb := range response.FillInBlank {
		totalScore += fb.TotalScore
		fmt.Printf("🔍 FillInBlank score: %f\n", fb.TotalScore)
	}
	for _, ord := range response.Ordering {
		totalScore += ord.TotalScore
		fmt.Printf("🔍 Ordering score: %f\n", ord.TotalScore)
	}
	for _, dd := range response.Dragdrop {
		totalScore += dd.TotalScore
		fmt.Printf("🔍 Dragdrop score: %f\n", dd.TotalScore)
	}
	for _, match := range response.Matching {
		totalScore += match.TotalScore
		fmt.Printf("🔍 Matching score: %f\n", match.TotalScore)
	}
	for _, label := range response.Labeling {
		totalScore += label.TotalScore
		fmt.Printf("🔍 Labeling score: %f\n", label.TotalScore)
	}
	for _, cat := range response.Category {
		totalScore += cat.TotalScore
		fmt.Printf("🔍 Category score: %f\n", cat.TotalScore)
	}

	fmt.Printf("🔍 Total score calculated: %f\n", totalScore)
	response.TotalScore = totalScore

	// For now, use a simple ratio calculation
	// TODO: Get contest round info for proper ratio calculation
	ratio := 0.0
	if len(req.ListAnswers) > 0 {
		ratio = totalScore / float64(len(req.ListAnswers))
	}

	response.Ratio = ratio

	// Save to contest_round_users table
	fmt.Printf("🔍 Saving to contest_round_users: ContestRoundId=%d, UserID=%d, Score=%f, Ratio=%f\n", req.ContestRoundId, userID, totalScore, ratio)
	contestRoundUser := map[string]interface{}{
		"contest_round_id":   req.ContestRoundId,
		"user_id":            userID,
		"score":              totalScore,
		"ratio":              ratio,
		"time":               req.Time,
		"has_manual_scoring": hasManualScoring,
	}

	// Check if record exists, if yes update, if no create
	var existingRecord struct {
		ID int64 `gorm:"column:id"`
	}
	err = db.MasterDB.Table("contest_round_users").
		Where("contest_round_id = ? AND user_id = ?", req.ContestRoundId, userID).
		First(&existingRecord).Error

	if err != nil {
		// Record doesn't exist, create new one
		if err := db.MasterDB.Table("contest_round_users").Create(contestRoundUser).Error; err != nil {
			return nil, fmt.Errorf("failed to create contest round user record: %v", err)
		}
	} else {
		// Record exists, update it
		if err := db.MasterDB.Table("contest_round_users").
			Where("contest_round_id = ? AND user_id = ?", req.ContestRoundId, userID).
			Updates(contestRoundUser).Error; err != nil {
			return nil, fmt.Errorf("failed to update contest round user record: %v", err)
		}
	}

	fmt.Printf("🔍 SaveScoreBulkContestRound completed successfully. Response: ContestRoundId=%d, TotalScore=%f, Ratio=%f\n", response.ContestRoundId, response.TotalScore, response.Ratio)
	return response, nil
}

func (s *saveScoreBulkService) SaveBulkWrite(req *prot.SaveScoreBulkRequest, userID int64, typeQuestion string) (*prot.SaveScoreBulkResponse, error) {
	mediaRepo := repositories.NewMediaRepository()

	fileInfos := make([]models.MediaDetail, 0)

	for _, file := range req.Files{
		fileUrl := utils.StripDomain(file.Url, models.Storage)
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)
		fileInfos = append(fileInfos, models.MediaDetail{
			Path: utils.StripDomain(fileInfo.Path, models.Storage),
			Id:   fileInfo.Id,
			Type: file.Type,
			Disk: fileInfo.Disk,
		})
	}

	score := 0.0
	ratio := 0.0
	hasManualScoring := true

	var err error

	files := make([]*prot.BulkFile, 0)

	for _, file := range req.Files {
		files = append(files, &prot.BulkFile{
			Type: file.Type,
			Url: utils.StaticURL(file.Url, models.Storage),
		})
	}

	switch typeQuestion {
	case models.ClonedQuestionTypeExam:
		examUserRepo := repositories.NewExamUserRepository()
		err = examUserRepo.UpdateOrCreate(&models.ExamUser{
			ExamID: req.ExamId,
			UserID: userID,
			LessonID: req.LessonId,
			FileInfos: fileInfos,
			HasManualScoring: hasManualScoring,
			Score: &score,
			Ratio: &ratio,
		})
	case models.ClonedQuestionTypeHomework:
		homeworkUserRepo := repositories.NewHomeworkUserRepository()
		err = homeworkUserRepo.UpdateOrCreate(&models.HomeworkUser{
			HomeworkID: req.HomeworkId,
			UserID: userID,
			LessonID: req.LessonId,
			FileInfos: fileInfos,
			HasManualScoring: hasManualScoring,
			Score: score,
			Ratio: ratio,
		})
	case models.ClonedQuestionTypeExercise:
		homeworkUserRepo := repositories.NewHomeworkUserRepository()
		err = homeworkUserRepo.UpdateOrCreate(&models.HomeworkUser{
			HomeworkID: req.ExerciseId,
			UserID: userID,
			LessonID: req.LessonId,
			FileInfos: fileInfos,
			HasManualScoring: hasManualScoring,
			Score: score,
			Ratio: ratio,
		})
	case models.ClonedQuestionTypeContestRound:
		homeworkUserRepo := repositories.NewContestScoreRepository()
		err = homeworkUserRepo.UpdateOrCreateContestRoundUser(&models.ContestRoundUser{
			ContestRoundId: req.ContestRoundId,
			UserId: userID,
			FileInfos: fileInfos,
			HasManualScoring: hasManualScoring,
			Score: score,
			Ratio: ratio,
		})
	default:
		err = fmt.Errorf("unsupported question type: %v", typeQuestion)
	}

	if err == nil {
		response := &prot.SaveScoreBulkResponse{
			MultipleChoice:   make([]*prot.SaveScoreResponse, 0),
			FillInBlank:      make([]*prot.SaveScoreResponseFillInBlank, 0),
			Ordering:         make([]*prot.SaveScoreResponsePosition, 0),
			Dragdrop:         make([]*prot.SaveScoreResponsePosition, 0),
			Matching:         make([]*prot.SaveScoreResponseMatching, 0),
			Labeling:         make([]*prot.SaveScoreResponseLabeling, 0),
			Category:         make([]*prot.SaveScoreResponseGroup, 0),
			Manual:           make([]*prot.SaveScoreResponseManual, 0),
			ContestRoundId:   req.ContestRoundId,
			ContestRoundName: "", // Will be filled later if needed
			Files: files,
		}

		return response, nil
	}

	return nil, fmt.Errorf("Failed to update SaveBulkWrite")
}
