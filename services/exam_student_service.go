package services

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/requests"
	"be-lms/utils"
)

type ExamStudentService interface {
	GetExamStudentsByExamIDService(examID int64, courseID int64, limit, page int) (*prot.ExamStudentResponseWithCount, error)
	GetExamInfoWithStats(examID int64) (dto.ExamStudentInfo, int32, int32, int32, float64, error)
	GetExamByStudentService(req requests.GetExamByStudentRequest) (*prot.GetExamByStudentResponse, error)
}

type examStudentService struct {
	repo repositories.ExamStudentRepository
}

func NewExamStudentService(repo repositories.ExamStudentRepository) ExamStudentService {
	return &examStudentService{repo: repo}
}

func (s *examStudentService) GetExamStudentsByExamIDService(examID int64, courseID int64, limit, page int) (*prot.ExamStudentResponseWithCount, error) {
	offset := 0
	if page > 0 && limit > 0 {
		offset = (page - 1) * limit
	}
	students, total, err := s.repo.GetExamStudentsByExamID(examID, courseID, limit, offset)
	if err != nil {
		return nil, err
	}
	var result []*prot.ExamStudentItem
	for _, st := range students {
		comment, _ := s.repo.GetExamComment(examID, st.UserID)
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
	// Lấy exam info và thống kê
	info, _, submitted, graded, avgScore, err := s.GetExamInfoWithStats(examID)
	if err != nil {
		return nil, err
	}
	// Lấy số lượng câu hỏi từ cloned_questions
	totalQuestions, _ := CountExamQuestionsFromCloned(examID)
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
		SubmittedStudents: int64(submitted),
		GradedStudents:    int64(graded),
		AvgScore:          avgScore,
	}
	return &prot.ExamStudentResponseWithCount{
		Students:   result,
		TotalCount: total,
		ExamInfo:   protoInfo,
	}, nil
}

func (s *examStudentService) GetExamInfoWithStats(examID int64) (dto.ExamStudentInfo, int32, int32, int32, float64, error) {
	info, err := s.repo.GetExamInfoByExamID(examID)
	if err != nil {
		return info, 0, 0, 0, 0, err
	}
	// Lấy danh sách học sinh (không phân trang)
	students, _, err := s.repo.GetExamStudentsByExamID(examID, 0, 0, 0)
	if err != nil {
		return info, 0, 0, 0, 0, err
	}
	var submittedCount int32
	var gradedCount int32
	var sumScore float64
	var gradedScoreCount int32
	for _, st := range students {
		if st.IsSubmitted {
			submittedCount++
		}
		if st.UnscoredCount == 0 && st.IsSubmitted {
			gradedCount++
			sumScore += st.Score
			gradedScoreCount++
		}
	}
	avgScore := 0.0
	if gradedScoreCount > 0 {
		avgScore = sumScore / float64(gradedScoreCount)
	}
	return info, int32(len(students)), submittedCount, gradedCount, avgScore, nil
}

func (s *examStudentService) GetExamByStudentService(req requests.GetExamByStudentRequest) (*prot.GetExamByStudentResponse, error) {
	offset := 0
	if req.Page > 0 && req.Limit > 0 {
		offset = (req.Page - 1) * req.Limit
	}

	var protoCourses []*prot.GetExamByStudentCourse

	courses, total, err := s.repo.GetExamByStudentRepoWithDate(req.UserID, req.WeekID, req.CourseID, req.StartDate, req.EndDate, req.Limit, offset)
	if err != nil {
		return nil, err
	}

	for _, course := range courses {
		var protoLessons []*prot.GetExamByStudentLesson
		for _, lesson := range course.Lessons {
			var exams []dto.GetExamByStudentExamQuery
			err := s.repo.GetExamsByLesson(lesson.ID, req.UserID, req.CourseID, &exams)
			if err != nil {
				return nil, err
			}
			var protoExams []*prot.GetExamByStudentExam
			seenExam := make(map[int64]bool)
			for _, e := range exams {
				if seenExam[e.ID] {
					continue
				}
				isSubmitted, _ := s.repo.IsExamSubmitted(e.ID, req.UserID)
				unscoredCount, _ := s.repo.CountUnscoredManual(e.ID, req.UserID)
				protoExams = append(protoExams, &prot.GetExamByStudentExam{
					Id:            e.ID,
					Name:          e.Name,
					Status:        e.Status,
					Description:   e.Description,
					CoverImage:    utils.StaticURL(e.CoverImage, models.Storage),
					Deadline:      e.Deadline,
					IsSubmitted:   isSubmitted,
					UnscoredCount: int32(unscoredCount),
				})

				seenExam[e.ID] = true
			}

			// Lấy danh sách homeworks cho lesson này
			homeworks, err := s.repo.GetHomeworksByLesson(lesson.ID, req.UserID, req.CourseID)
			if err != nil {
				return nil, err
			}
			var protoHomeworks []*prot.GetExamByStudentHomework
			seenHomework := make(map[int64]bool)
			for _, hw := range homeworks {
				if seenHomework[hw.ID] {
					continue
				}
				totalQ, _ := CountHomeworkQuestionsFromCloned(hw.ID)
				protoHomeworks = append(protoHomeworks, &prot.GetExamByStudentHomework{
					Id:                      hw.ID,
					Name:                    hw.Name,
					Status:                  hw.Status,
					Description:             hw.Description,
					TotalQuestion:           totalQ,
					QuestionCompleted:       hw.QuestionCompleted,
					CoverImage:              utils.StaticURL(hw.CoverImage, models.Storage),
					LastQuestionIdCompleted: hw.LastQuestionIDCompleted,
				})

				seenHomework[hw.ID] = true
			}

			// Lấy danh sách exercises cho lesson này (logic tương tự exams)
			exercises, err := s.repo.GetExercisesByLesson(lesson.ID, req.UserID, req.CourseID)
			if err != nil { return nil, err }
			var protoExercises []*prot.GetExamByStudentExercise
			seenExercise := make(map[int64]bool)
			for _, ex := range exercises {
				if seenExercise[ex.ID] {
					continue
				}
				isSubmitted, _ := s.repo.IsExerciseSubmitted(ex.ID, req.UserID)
				unscoredCount, _ := s.repo.CountUnscoredManualExercise(ex.ID, req.UserID)
				totalQ, _ := CountExerciseQuestionsFromCloned(ex.ID)
				protoExercises = append(protoExercises, &prot.GetExamByStudentExercise{
					Id:            ex.ID,
					Name:          ex.Name,
					Status:        ex.Status,
					Description:   ex.Description,
					CoverImage:    utils.StaticURL(ex.CoverImage, models.Storage),
					Deadline:      ex.Deadline,
					IsSubmitted:   isSubmitted,
					UnscoredCount: int32(unscoredCount),
					TotalQuestion: totalQ,
				})

				seenExercise[ex.ID] = true
			}

			protoLessons = append(protoLessons, &prot.GetExamByStudentLesson{
				Id:          lesson.ID,
				Title:       lesson.Title,
				Description: lesson.Description,
				Status:      lesson.Status,
				Exams:       protoExams,
				Homeworks:   protoHomeworks,
				Exercises:   protoExercises,
			})
		}
		protoCourses = append(protoCourses, &prot.GetExamByStudentCourse{
			Id:          course.ID,
			Name:        course.Name,
			Description: course.Description,
			Status:      course.Status,
			SubjectName: course.SubjectName,
			Type:        course.Type,
			Image:       utils.StaticURL(course.Image, models.Storage),
			Level:       course.Level,
			Target:      course.Target,
			Lessons:     protoLessons,
		})
	}

	return &prot.GetExamByStudentResponse{
		Courses: protoCourses,
		Total:   total,
	}, nil
}
