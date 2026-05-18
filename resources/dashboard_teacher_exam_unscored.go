package resources

import (
	"be-Clever School/dto"
	"be-Clever School/prot"
)

func DashboardTeacherExamUnscoredResource(exam dto.DashboardTeacherExamUnscored) *prot.DashboardTeacherExamUnscored {
	return &prot.DashboardTeacherExamUnscored{
		UserId:            exam.UserID,
		ExamId:            exam.ExamID,
		CourseId:          exam.CourseID,
		LessonId:          exam.LessonID,
		StudentName:       exam.StudentName,
		CourseName:        exam.CourseName,
		SubjectName:       exam.SubjectName,
		ExamName:          exam.ExamName,
		LessonTitle:       exam.LessonTitle,
		SubmittedAt:       exam.SubmittedAt.Unix(),
		Deadline:          exam.Deadline.Unix(),
		SubmitOnTime:      exam.SubmitOnTime,
		Ratio:             exam.Ratio,
		TotalQuestions:    int32(exam.TotalQuestions),
		UnscoredQuestions: int32(exam.UnscoredQuestions),
	}
}

func DashboardTeacherExamUnscoredCollection(exams []dto.DashboardTeacherExamUnscored) []*prot.DashboardTeacherExamUnscored {
	var result []*prot.DashboardTeacherExamUnscored
	for _, exam := range exams {
		result = append(result, DashboardTeacherExamUnscoredResource(exam))
	}
	return result
} 