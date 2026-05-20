package resources

import (
	"be-cleverschool/dto"
	"be-cleverschool/prot"
)

func DashboardTeacherHomeworkListResource(homework dto.DashboardTeacherHomeworkListItem) *prot.DashboardTeacherHomeworkListItem {
	return &prot.DashboardTeacherHomeworkListItem{
		HomeworkId:         homework.HomeworkID,
		HomeworkName:       homework.HomeworkName,
		TotalQuestions:     homework.TotalQuestions,
		QuestionsCompleted: homework.QuestionsCompleted,
		Status:             homework.Status,
		SubmittedAt:        homework.SubmittedAt,
		AssignedAt:         homework.AssignedAt,
		LessonId:           homework.LessonID,
		LessonName:         homework.LessonName,
		AssignLate:         homework.AssignLate,
	}
}

func DashboardTeacherHomeworkListCollection(homeworks []dto.DashboardTeacherHomeworkListItem) []*prot.DashboardTeacherHomeworkListItem {
	var result []*prot.DashboardTeacherHomeworkListItem
	for _, homework := range homeworks {
		result = append(result, DashboardTeacherHomeworkListResource(homework))
	}
	return result
}


