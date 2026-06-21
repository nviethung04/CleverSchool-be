package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"fmt"
)

type ExerciseStudentService interface {
	GetExerciseStudentsByExerciseIDService(exerciseID int64, courseID int64, limit, page int) (interface{}, error)
	ResetExerciseAttemptService(exerciseID, userID, lessonID int64) error
}

type exerciseStudentService struct {
	repo repositories.ExerciseStudentRepository
}

func NewExerciseStudentService(repo repositories.ExerciseStudentRepository) ExerciseStudentService {
	return &exerciseStudentService{repo: repo}
}

func (s *exerciseStudentService) GetExerciseStudentsByExerciseIDService(exerciseID int64, courseID int64, limit, page int) (interface{}, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}

	students, total, err := s.repo.GetExerciseStudentsByExerciseID(exerciseID, courseID, limit, offset)
	if err != nil {
		return nil, err
	}

	var result []*prot.ExamStudentItem
	for _, st := range students {
		comment, _ := s.repo.GetExerciseComment(exerciseID, st.UserID)
		result = append(result, &prot.ExamStudentItem{
			UserId:        st.UserID,
			Name:          st.Name,
			Username:      st.Username,
			Avatar:        utils.StaticURL(st.Avatar, models.Storage),
			IsSubmitted:   st.IsSubmitted,
			SubmittedAt:   st.SubmittedAt,
			Duration:      st.Duration,
			UnscoredCount: st.UnscoredCount,
			Score:         st.Score,
			Ratio:         st.Ratio,
			Comment:       comment,
			CourseId:      st.CourseID,
		})
	}

	// Lấy exercise_info chi tiết
	info, err := s.repo.GetExerciseInfoByExerciseID(exerciseID)
	if err != nil {
		return nil, err
	}
	totalQuestions, _ := CountExerciseQuestionsFromCloned(exerciseID)

	// Map sang struct giống exam
	protoInfo := &prot.ExamStudentExamInfo{
		Name:              info.Name,
		Description:       info.Description,
		IsAssigned:        info.IsAssigned,
		Status:            info.Status,
		MaxScore:          info.MaxScore,
		TimeLimit:         info.TimeLimit,
		CreatedAt:         info.CreatedAt,
		Deadline:          info.Deadline,
		TotalQuestions:    int64(totalQuestions),
		TotalStudents:     int64(info.TotalStudents),
		SubmittedStudents: 0,
		GradedStudents:    0,
		AvgScore:          0,
	}
	return &prot.ExamStudentResponseWithCount{
		Students:   result,
		TotalCount: total,
		ExamInfo:   protoInfo,
	}, nil
}

func (s *exerciseStudentService) ResetExerciseAttemptService(exerciseID, userID, lessonID int64) error {
	if exerciseID == 0 || userID == 0 {
		return fmt.Errorf("exercise_id and user_id are required")
	}
	return repositories.ResetExerciseAttempt(exerciseID, userID, lessonID)
}
