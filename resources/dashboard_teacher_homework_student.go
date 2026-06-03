package resources

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/utils"
)

func DashboardTeacherHomeworkStudentResource(student dto.DashboardTeacherHomeworkStudent) *prot.DashboardTeacherHomeworkStudent {
	return &prot.DashboardTeacherHomeworkStudent{
		StudentId:          student.StudentID,
		StudentName:        student.StudentName,
		StudentAvatar:      utils.StaticURL(student.StudentAvatar.Path, models.Storage),
		TotalQuestions:     student.TotalQuestions,
		QuestionsCompleted: student.QuestionsCompleted,
	}
}

func DashboardTeacherHomeworkStudentCollection(students []dto.DashboardTeacherHomeworkStudent) []*prot.DashboardTeacherHomeworkStudent {
	var result []*prot.DashboardTeacherHomeworkStudent
	for _, student := range students {
		result = append(result, DashboardTeacherHomeworkStudentResource(student))
	}
	return result
}

func DashboardTeacherHomeworkStudentStatsResource(student dto.DashboardTeacherHomeworkStudentStats) *prot.DashboardTeacherHomeworkStudentStats {
	return &prot.DashboardTeacherHomeworkStudentStats{
		StudentId:              student.StudentID,
		StudentName:            student.StudentName,
		StudentAvatar:          utils.StaticURL(student.StudentAvatar.Path, models.Storage),
		TotalHomework:          student.TotalHomework,
		TotalAssignedHomeworks: student.TotalAssignedHomeworks,
		InProgressHomework:     student.InProgressHomework,
		CompletedHomework:      student.CompletedHomework,
		NotStartedHomework:     student.NotStartedHomework,
		AverageRatio:           student.AverageRatio,
	}
}

func DashboardTeacherHomeworkStudentStatsCollection(students []dto.DashboardTeacherHomeworkStudentStats) []*prot.DashboardTeacherHomeworkStudentStats {
	var result []*prot.DashboardTeacherHomeworkStudentStats
	for _, student := range students {
		result = append(result, DashboardTeacherHomeworkStudentStatsResource(student))
	}
	return result
}

func DashboardTeacherHomeworkOverviewResource(overview dto.DashboardTeacherHomeworkOverview) *prot.DashboardTeacherHomeworkOverview {
	return &prot.DashboardTeacherHomeworkOverview{
		NotStartedPercentage: overview.NotStartedPercentage,
		InProgressPercentage:  overview.InProgressPercentage,
		CompletedPercentage:   overview.CompletedPercentage,
		TotalStudents:         overview.TotalStudents,
		NotStartedCount:       overview.NotStartedCount,
		InProgressCount:       overview.InProgressCount,
		CompletedCount:        overview.CompletedCount,
	}
}
