package resources

import (
	"be-lms/dto"
	"be-lms/prot"
)

func DashboardTeacherHomeworkScoredResource(homework dto.DashboardTeacherHomeworkScored) *prot.DashboardTeacherHomeworkScored {
	return &prot.DashboardTeacherHomeworkScored{
		UserId:            homework.UserID,
		HomeworkId:        homework.HomeworkID,
		CourseId:          homework.CourseID,
		LessonId:          homework.LessonID,
		StudentName:       homework.StudentName,
		CourseName:        homework.CourseName,
		SubjectName:       homework.SubjectName,
		HomeworkName:      homework.HomeworkName,
		LessonTitle:       homework.LessonTitle,
		SubmittedAt:       homework.SubmittedAt.Unix(),
		Ratio:             homework.Ratio,
		TotalQuestions:    int32(homework.TotalQuestions),
		UnscoredQuestions: int32(homework.UnscoredQuestions),
	}
}

func DashboardTeacherHomeworkScoredCollection(homeworks []dto.DashboardTeacherHomeworkScored) []*prot.DashboardTeacherHomeworkScored {
	var result []*prot.DashboardTeacherHomeworkScored
	for _, homework := range homeworks {
		result = append(result, DashboardTeacherHomeworkScoredResource(homework))
	}
	return result
}
