package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func DashboardTeacherHomeworkUnscoredCollection(homeworks []dto.DashboardTeacherHomeworkUnscored) []*prot.DashboardTeacherHomeworkUnscored {
	var result []*prot.DashboardTeacherHomeworkUnscored
	for _, h := range homeworks {
		result = append(result, &prot.DashboardTeacherHomeworkUnscored{
			UserId:           h.UserID,
			HomeworkId:       h.HomeworkID,
			CourseId:        h.CourseID,
			LessonId:         h.LessonID,
			StudentName:      h.StudentName,
			CourseName:       h.CourseName,
			ObjectTitle:      h.ObjectTitle,
			SubjectName:      h.SubjectName,
			HomeworkName:     h.HomeworkName,
			LessonTitle:      h.LessonTitle,
			SubmittedAt:      h.SubmittedAt.Unix(),
			Ratio:            h.Ratio,
			TotalQuestions:   int32(h.TotalQuestions),
			UnscoredQuestions: int32(h.UnscoredQuestions),
			ManualQuestions:  int32(h.ManualQuestions),
			HasComment:       h.HasComment,
			CommentContent:   h.CommentContent,
		})
	}
	return result
}
