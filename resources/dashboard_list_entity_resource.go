package resources

import (
	"be-lms/dto"
	"be-lms/prot"
)

func DashboardSchoolListResource(school dto.DashboardSchool) *prot.DashboardSchoolList {
	return &prot.DashboardSchoolList{
		Id:              school.ID,
		SchoolName:      school.SchoolName,
		SchoolShortName: school.SchoolShortName,
		WardName:        school.WardName,
		ProvinceName:    school.ProvinceName,
	}
}

func DashboardSchoolListCollection(schools []dto.DashboardSchool) []*prot.DashboardSchoolList {
	var result []*prot.DashboardSchoolList
	for _, school := range schools {
		result = append(result, DashboardSchoolListResource(school))
	}
	return result
}

func DashboardCourseListResource(course dto.DashboardCourse) *prot.DashboardCourseList {
	return &prot.DashboardCourseList{
		Id:          course.ID,
		Name:        course.Name,
		ObjectTitle: course.ObjectTitle,
	}
}

func DashboardCourseListCollection(courses []dto.DashboardCourse) []*prot.DashboardCourseList {
	var result []*prot.DashboardCourseList
	for _, course := range courses {
		result = append(result, DashboardCourseListResource(course))
	}
	return result
}

func DashboardProgramListResource(program dto.DashboardProgram) *prot.DashboardProgramList {
	return &prot.DashboardProgramList{
		Id:   program.ID,
		Name: program.Name,
	}
}

func DashboardProgramListCollection(programs []dto.DashboardProgram) []*prot.DashboardProgramList {
	var result []*prot.DashboardProgramList
	for _, program := range programs {
		result = append(result, DashboardProgramListResource(program))
	}
	return result
}

func DashboardTeacherListResource(teacher dto.DashboardTeacher) *prot.DashboardTeacherList {
	return &prot.DashboardTeacherList{
		Id:   teacher.ID,
		Name: teacher.Name,
	}
}

func DashboardTeacherListCollection(teachers []dto.DashboardTeacher) []*prot.DashboardTeacherList {
	var result []*prot.DashboardTeacherList
	for _, teacher := range teachers {
		result = append(result, DashboardTeacherListResource(teacher))
	}
	return result
}

func DashboardSubjectListResource(subject dto.DashboardSubject) *prot.DashboardSubjectList {
	return &prot.DashboardSubjectList{
		Id:   subject.ID,
		Name: subject.Name,
	}
}

func DashboardSubjectListCollection(subjects []dto.DashboardSubject) []*prot.DashboardSubjectList {
	var result []*prot.DashboardSubjectList
	for _, subject := range subjects {
		result = append(result, DashboardSubjectListResource(subject))
	}
	return result
}

func DashboardExamListResource(exam dto.DashboardExam) *prot.DashboardExamList {
	return &prot.DashboardExamList{
		Id:        exam.ID,
		Name:      exam.Name,
		CourseId:  exam.CourseID,
		SubjectId: exam.SubjectID,
	}
}

func DashboardExamListCollection(exams []dto.DashboardExam) []*prot.DashboardExamList {
	var result []*prot.DashboardExamList
	for _, exam := range exams {
		result = append(result, DashboardExamListResource(exam))
	}
	return result
}

func DashboardHomeworkListResource(homework dto.DashboardHomework) *prot.DashboardHomeworkList {
	var assignedAt int64
	if homework.AssignedAt != nil {
		assignedAt = homework.AssignedAt.Unix()
	}
	
	return &prot.DashboardHomeworkList{
		Id:           homework.ID,
		Name:         homework.Name,
		CourseId:     homework.CourseID,
		SubjectId:    homework.SubjectID,
		CourseName:   homework.CourseName,
		SubjectName:  homework.SubjectName,
		LessonId:     homework.LessonID,
		LessonTitle:  homework.LessonTitle,
		AssignedAt:   assignedAt,
	}
}

func DashboardHomeworkListCollection(homeworks []dto.DashboardHomework) []*prot.DashboardHomeworkList {
	var result []*prot.DashboardHomeworkList
	for _, homework := range homeworks {
		result = append(result, DashboardHomeworkListResource(homework))
	}
	return result
}

func DashboardLessonListResource(lesson dto.DashboardLesson) *prot.DashboardLessonList {
	return &prot.DashboardLessonList{
		Id:          lesson.ID,
		Title:       lesson.Title,
		ChapterId:   lesson.ChapterID,
		ChapterName: lesson.ChapterName,
		CourseId:    lesson.CourseID,
		CourseName:  lesson.CourseName,
	}
}

func DashboardLessonListCollection(lessons []dto.DashboardLesson) []*prot.DashboardLessonList {
	var result []*prot.DashboardLessonList
	for _, lesson := range lessons {
		result = append(result, DashboardLessonListResource(lesson))
	}
	return result
} 