package resources

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
)

func DashboardTeacherHomeworkScoredResource(homework dto.DashboardTeacherHomeworkScored) *prot.DashboardTeacherHomeworkScored {
	return &prot.DashboardTeacherHomeworkScored{
		UserId:            homework.UserID,
		HomeworkId:        homework.HomeworkID,
		CourseId:          homework.CourseID,
		LessonId:          homework.LessonID,
		StudentName:       homework.StudentName,
		CourseName:        homework.CourseName,
		ObjectTitle:       homework.ObjectTitle,
		SubjectName:       homework.SubjectName,
		HomeworkName:      homework.HomeworkName,
		LessonTitle:       homework.LessonTitle,
		SubmittedAt:       homework.SubmittedAt.Unix(),
		Ratio:             homework.Ratio,
		TotalQuestions:    int32(homework.TotalQuestions),
		UnscoredQuestions: int32(homework.UnscoredQuestions),
		HasComment:        homework.HasComment,
		CommentContent:    homework.CommentContent,
		Graders:           dashboardTeacherHomeworkGraderCollection(homework.Graders),
	}
}

func DashboardTeacherHomeworkScoredCollection(homeworks []dto.DashboardTeacherHomeworkScored) []*prot.DashboardTeacherHomeworkScored {
	var result []*prot.DashboardTeacherHomeworkScored
	for _, homework := range homeworks {
		result = append(result, DashboardTeacherHomeworkScoredResource(homework))
	}
	return result
}

func dashboardTeacherHomeworkGraderCollection(graders []dto.DashboardTeacherHomeworkGrader) []*prot.DashboardTeacherHomeworkGrader {
	if len(graders) == 0 {
		return nil
	}
	var result []*prot.DashboardTeacherHomeworkGrader
	for _, grader := range graders {
		result = append(result, &prot.DashboardTeacherHomeworkGrader{
			UserId:   grader.UserID,
			Username: grader.Username,
			Name:     grader.Name,
		})
	}
	return result
}

