package resources

import (
	"be-lms/dto"
	"be-lms/prot"
)

func DashboardTeacherExamScoredResource(exam dto.DashboardTeacherExamScored) *prot.DashboardTeacherExamScored {
	return &prot.DashboardTeacherExamScored{
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

func DashboardTeacherExamScoredCollection(exams []dto.DashboardTeacherExamScored) []*prot.DashboardTeacherExamScored {
	var result []*prot.DashboardTeacherExamScored
	for _, exam := range exams {
		result = append(result, DashboardTeacherExamScoredResource(exam))
	}
	return result
}
