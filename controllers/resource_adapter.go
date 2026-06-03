package controllers

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/resources"
)

// Topic
type TopicResourceAdapter struct {
	resource resources.TopicResource
}

func NewTopicResourceAdapter(resource resources.TopicResource) *TopicResourceAdapter {
	return &TopicResourceAdapter{resource: resource}
}

func (adapter *TopicResourceAdapter) FormatItems(items []*models.Topic) []*prot.Topic {
	return adapter.resource.FormatTopics(items)
}

func (adapter *TopicResourceAdapter) FormatItem(item *models.Topic) *prot.Topic {
	return adapter.resource.FormatTopic(item)
}

// Tag
type TagResourceAdapter struct {
	resource resources.TagResource
}

func NewTagResourceAdapter(resource resources.TagResource) *TagResourceAdapter {
	return &TagResourceAdapter{resource: resource}
}

func (adapter *TagResourceAdapter) FormatItems(items []*models.Tag) []*prot.Tag {
	return adapter.resource.FormatTags(items)
}

func (adapter *TagResourceAdapter) FormatItem(item *models.Tag) *prot.Tag {
	return adapter.resource.FormatTag(item)
}

// Subject
type SubjectResourceAdapter struct {
	resource resources.SubjectResource
}

func NewSubjectResourceAdapter(resource resources.SubjectResource) *SubjectResourceAdapter {
	return &SubjectResourceAdapter{resource: resource}
}

func (adapter *SubjectResourceAdapter) FormatItems(items []*models.Subject) []*prot.Subject {
	return adapter.resource.FormatSubjects(items)
}

func (adapter *SubjectResourceAdapter) FormatItem(item *models.Subject) *prot.Subject {
	return adapter.resource.FormatSubject(item)
}

// Skill
type SkillResourceAdapter struct {
	resource resources.SkillResource
}

func NewSkillResourceAdapter(resource resources.SkillResource) *SkillResourceAdapter {
	return &SkillResourceAdapter{resource: resource}
}

func (adapter *SkillResourceAdapter) FormatItems(items []*models.Skill) []*prot.Skill {
	return adapter.resource.FormatSkills(items)
}

func (adapter *SkillResourceAdapter) FormatItem(item *models.Skill) *prot.Skill {
	return adapter.resource.FormatSkill(item)
}

// Degree
type DegreeResourceAdapter struct {
	resource resources.DegreeResource
}

func NewDegreeResourceAdapter(resource resources.DegreeResource) *DegreeResourceAdapter {
	return &DegreeResourceAdapter{resource: resource}
}

func (adapter *DegreeResourceAdapter) FormatItems(items []*models.Degree) []*prot.Degree {
	return adapter.resource.FormatDegrees(items)
}

func (adapter *DegreeResourceAdapter) FormatItem(item *models.Degree) *prot.Degree {
	return adapter.resource.FormatDegree(item)
}

// Department
type DepartmentResourceAdapter struct {
	resource resources.DepartmentResource
}

func NewDepartmentResourceAdapter(resource resources.DepartmentResource) *DepartmentResourceAdapter {
	return &DepartmentResourceAdapter{resource: resource}
}

func (adapter *DepartmentResourceAdapter) FormatItems(items []*models.Department) []*prot.Department {
	return adapter.resource.FormatDepartments(items)
}

func (adapter *DepartmentResourceAdapter) FormatItem(item *models.Department) *prot.Department {
	return adapter.resource.FormatDepartment(item)
}

// EmployeePosition
type EmployeePositionResourceAdapter struct {
	resource resources.EmployeePositionResource
}

func NewEmployeePositionResourceAdapter(resource resources.EmployeePositionResource) *EmployeePositionResourceAdapter {
	return &EmployeePositionResourceAdapter{resource: resource}
}

func (adapter *EmployeePositionResourceAdapter) FormatItems(items []*models.EmployeePosition) []*prot.EmployeePosition {
	return adapter.resource.FormatEmployeePositions(items)
}

func (adapter *EmployeePositionResourceAdapter) FormatItem(item *models.EmployeePosition) *prot.EmployeePosition {
	return adapter.resource.FormatEmployeePosition(item)
}

// Certificate
type CertificateResourceAdapter struct {
	resource resources.CertificateResource
}

func NewCertificateResourceAdapter(resource resources.CertificateResource) *CertificateResourceAdapter {
	return &CertificateResourceAdapter{resource: resource}
}

func (adapter *CertificateResourceAdapter) FormatItems(items []*models.Certificate) []*prot.Certificate {
	return adapter.resource.FormatCertificates(items)
}

func (adapter *CertificateResourceAdapter) FormatItem(item *models.Certificate) *prot.Certificate {
	return adapter.resource.FormatCertificate(item)
}

// Class
type ClassResourceAdapter struct {
	resource resources.ClassResource
}

func NewClassResourceAdapter(resource resources.ClassResource) *ClassResourceAdapter {
	return &ClassResourceAdapter{resource: resource}
}

func (adapter *ClassResourceAdapter) FormatItems(items []*models.Class) []*prot.ClassResponse {
	return adapter.resource.FormatClasses(items)
}

func (adapter *ClassResourceAdapter) FormatItem(item *models.Class) *prot.ClassResponse {
	return adapter.resource.FormatClass(item)
}

// Course
type CourseResourceAdapter struct {
	resource resources.CourseResource
}

func NewCourseResourceAdapter(resource resources.CourseResource) *CourseResourceAdapter {
	return &CourseResourceAdapter{resource: resource}
}

func (adapter *CourseResourceAdapter) FormatItems(items []*models.Course) []*prot.Course {
	return adapter.resource.FormatCourses(items)
}

func (adapter *CourseResourceAdapter) FormatItem(item *models.Course) *prot.Course {
	return adapter.resource.FormatCourse(item)
}

// Chapter
type ChapterResourceAdapter struct {
	resource resources.ChapterResource
}

func NewChapterResourceAdapter(resource resources.ChapterResource) *ChapterResourceAdapter {
	return &ChapterResourceAdapter{resource: resource}
}

func (adapter *ChapterResourceAdapter) FormatItems(items []*models.Chapter) []*prot.Chapter {
	return adapter.resource.FormatChapters(items)
}

func (adapter *ChapterResourceAdapter) FormatItem(item *models.Chapter) *prot.Chapter {
	return adapter.resource.FormatChapter(item)
}

// Lesson
type LessonResourceAdapter struct {
	resource resources.LessonResource
}

func NewLessonResourceAdapter(resource resources.LessonResource) *LessonResourceAdapter {
	return &LessonResourceAdapter{resource: resource}
}

func (adapter *LessonResourceAdapter) FormatItems(items []*models.Lesson) []*prot.Lesson {
	return adapter.resource.FormatLessons(items)
}

func (adapter *LessonResourceAdapter) FormatItem(item *models.Lesson) *prot.Lesson {
	return adapter.resource.FormatLesson(item)
}

// School
type SchoolResourceAdapter struct {
	resource resources.SchoolResource
}

func NewSchoolResourceAdapter(resource resources.SchoolResource) *SchoolResourceAdapter {
	return &SchoolResourceAdapter{resource: resource}
}

func (adapter *SchoolResourceAdapter) FormatItems(items []*models.School) []*prot.School {
	return adapter.resource.FormatSchools(items)
}

func (adapter *SchoolResourceAdapter) FormatItem(item *models.School) *prot.School {
	return adapter.resource.FormatSchool(item)
}

// Grade
type GradeResourceAdapter struct {
	resource resources.GradeResource
}

func NewGradeResourceAdapter(resource resources.GradeResource) *GradeResourceAdapter {
	return &GradeResourceAdapter{resource: resource}
}

func (adapter *GradeResourceAdapter) FormatItems(items []*models.Grade) []*prot.Grade {
	return adapter.resource.FormatGrades(items)
}

func (adapter *GradeResourceAdapter) FormatItem(item *models.Grade) *prot.Grade {
	return adapter.resource.FormatGrade(item)
}

// QuestionAttribute
type QuestionAttributeResourceAdapter struct {
	resource resources.QuestionAttributeResource
}

func NewQuestionAttributeResourceAdapter(resource resources.QuestionAttributeResource) *QuestionAttributeResourceAdapter {
	return &QuestionAttributeResourceAdapter{resource: resource}
}

func (adapter *QuestionAttributeResourceAdapter) FormatItems(items []*models.QuestionAttribute) []*prot.QuestionAttribute {
	return adapter.resource.FormatQuestionAttributes(items)
}

func (adapter *QuestionAttributeResourceAdapter) FormatItem(item *models.QuestionAttribute) *prot.QuestionAttribute {
	return adapter.resource.FormatQuestionAttribute(item)
}

// Role
type RoleResourceAdapter struct {
	resource resources.RoleResource
}

func NewRoleResourceAdapter(resource resources.RoleResource) *RoleResourceAdapter {
	return &RoleResourceAdapter{resource: resource}
}

func (adapter *RoleResourceAdapter) FormatItems(items []*models.Role) []*prot.Role {
	return adapter.resource.FormatRoles(items)
}

func (adapter *RoleResourceAdapter) FormatItem(item *models.Role) *prot.Role {
	return adapter.resource.FormatRole(item)
}

// StudyShift
type StudyShiftResourceAdapter struct {
	resource resources.StudyShiftResource
}

func NewStudyShiftResourceAdapter(resource resources.StudyShiftResource) *StudyShiftResourceAdapter {
	return &StudyShiftResourceAdapter{resource: resource}
}

func (adapter *StudyShiftResourceAdapter) FormatItems(items []*models.StudyShift) []*prot.StudyShift {
	return adapter.resource.FormatStudyShifts(items)
}

func (adapter *StudyShiftResourceAdapter) FormatItem(item *models.StudyShift) *prot.StudyShift {
	return adapter.resource.FormatStudyShift(item)
}

// Exam
type ExamResourceAdapter struct {
	resource resources.ExamResource
}

func NewExamResourceAdapter(resource resources.ExamResource) *ExamResourceAdapter {
	return &ExamResourceAdapter{resource: resource}
}

func (adapter *ExamResourceAdapter) FormatItems(items []*models.Exam) []*prot.Exam {
	return adapter.resource.FormatExams(items)
}

func (adapter *ExamResourceAdapter) FormatItem(item *models.Exam) *prot.Exam {
	return adapter.resource.FormatExam(item)
}

// Exercise
type ExerciseResourceAdapter struct {
	resource resources.ExerciseResource
}

func NewExerciseResourceAdapter(resource resources.ExerciseResource) *ExerciseResourceAdapter {
	return &ExerciseResourceAdapter{resource: resource}
}

func (adapter *ExerciseResourceAdapter) FormatItems(items []*models.Exercise) []*prot.Exercise {
	return adapter.resource.FormatExercises(items)
}

func (adapter *ExerciseResourceAdapter) FormatItem(item *models.Exercise) *prot.Exercise {
	return adapter.resource.FormatExercise(item)
}

// Homework
type HomeworkResourceAdapter struct {
	resource resources.HomeworkResource
}

func NewHomeworkResourceAdapter(resource resources.HomeworkResource) *HomeworkResourceAdapter {
	return &HomeworkResourceAdapter{resource: resource}
}

func (adapter *HomeworkResourceAdapter) FormatItems(items []*models.Homework) []*prot.Homework {
	return adapter.resource.FormatHomeworks(items)
}

func (adapter *HomeworkResourceAdapter) FormatItem(item *models.Homework) *prot.Homework {
	return adapter.resource.FormatHomework(item)
}

// Feedback
type FeedbackResourceAdapter struct {
	resource resources.FeedbackResource
}

func NewFeedbackResourceAdapter(resource resources.FeedbackResource) *FeedbackResourceAdapter {
	return &FeedbackResourceAdapter{resource: resource}
}

func (adapter *FeedbackResourceAdapter) FormatItems(items []*dto.FeedbackResponse) []*prot.Feedback {
	return adapter.resource.FormatFeedbacks(items)
}

func (adapter *FeedbackResourceAdapter) FormatItem(item *dto.FeedbackResponse) *prot.Feedback {
	return adapter.resource.FormatFeedback(item)
}

// User
type UserResourceAdapter struct {
	resource resources.UserResource
}

func NewUserResourceAdapter(resource resources.UserResource) *UserResourceAdapter {
	return &UserResourceAdapter{resource: resource}
}

func (adapter *UserResourceAdapter) FormatItems(items []*models.User) []*prot.User {
	return adapter.resource.FormatUsers(items)
}

func (adapter *UserResourceAdapter) FormatItem(item *models.User) *prot.User {
	return adapter.resource.FormatUser(item)
}

// Contest
type ContestResourceAdapter struct {
	resource resources.ContestResource
}

func NewContestResourceAdapter(resource resources.ContestResource) *ContestResourceAdapter {
	return &ContestResourceAdapter{resource: resource}
}

func (adapter *ContestResourceAdapter) FormatItems(items []*models.Contest) []*prot.Contest {
	return adapter.resource.FormatContests(items)
}

func (adapter *ContestResourceAdapter) FormatItem(item *models.Contest) *prot.Contest {
	return adapter.resource.FormatContest(item)
}

// ContestRound
type ContestRoundResourceAdapter struct {
	resource resources.ContestRoundResource
}

func NewContestRoundResourceAdapter(resource resources.ContestRoundResource) *ContestRoundResourceAdapter {
	return &ContestRoundResourceAdapter{resource: resource}
}

func (adapter *ContestRoundResourceAdapter) FormatItems(items []*models.ContestRound) []*prot.ContestRound {
	return adapter.resource.FormatContestRounds(items)
}

func (adapter *ContestRoundResourceAdapter) FormatItem(item *models.ContestRound) *prot.ContestRound {
	return adapter.resource.FormatContestRound(item)
}

// LessonPlan
type LessonPlanResourceAdapter struct {
	resource resources.LessonPlanResource
}

func NewLessonPlanResourceAdapter(resource resources.LessonPlanResource) *LessonPlanResourceAdapter {
	return &LessonPlanResourceAdapter{resource: resource}
}

func (adapter *LessonPlanResourceAdapter) FormatItems(items []*models.LessonPlan) []*prot.LessonPlan {
	return adapter.resource.FormatLessonPlans(items)
}

func (adapter *LessonPlanResourceAdapter) FormatItem(item *models.LessonPlan) *prot.LessonPlan {
	return adapter.resource.FormatLessonPlan(item)
}

// LessonPlanPart
type LessonPlanPartResourceAdapter struct {
	resource resources.LessonPlanPartResource
}

func NewLessonPlanPartResourceAdapter(resource resources.LessonPlanPartResource) *LessonPlanPartResourceAdapter {
	return &LessonPlanPartResourceAdapter{resource: resource}
}

func (adapter *LessonPlanPartResourceAdapter) FormatItems(items []*models.LessonPlanPart) []*prot.LessonPlanPart {
	return adapter.resource.FormatLessonPlanParts(items)
}

func (adapter *LessonPlanPartResourceAdapter) FormatItem(item *models.LessonPlanPart) *prot.LessonPlanPart {
	return adapter.resource.FormatLessonPlanPart(item)
}

// Question
type QuestionResourceAdapter struct {
	resource resources.QuestionResource
}

func NewQuestionResourceAdapter(resource resources.QuestionResource) *QuestionResourceAdapter {
	return &QuestionResourceAdapter{resource: resource}
}

func (adapter *QuestionResourceAdapter) FormatItems(items []*models.Question) []*prot.Question {
	return adapter.resource.FormatQuestions(items)
}

func (adapter *QuestionResourceAdapter) FormatItem(item *models.Question) *prot.Question {
	return adapter.resource.FormatQuestion(item)
}

// Holiday
type HolidayResourceAdapter struct {
	resource resources.HolidayResource
}

func NewHolidayResourceAdapter(resource resources.HolidayResource) *HolidayResourceAdapter {
	return &HolidayResourceAdapter{resource: resource}
}

func (adapter *HolidayResourceAdapter) FormatItems(items []*models.Holiday) []*prot.Holiday {
	return adapter.resource.FormatHolidays(items)
}

func (adapter *HolidayResourceAdapter) FormatItem(item *models.Holiday) *prot.Holiday {
	return adapter.resource.FormatHoliday(item)
}

// Semester
type SemesterResourceAdapter struct {
	resource resources.SemesterResource
}

func NewSemesterResourceAdapter(resource resources.SemesterResource) *SemesterResourceAdapter {
	return &SemesterResourceAdapter{resource: resource}
}

func (adapter *SemesterResourceAdapter) FormatItems(items []*models.Semester) []*prot.Semester {
	return adapter.resource.FormatSemesters(items)
}

func (adapter *SemesterResourceAdapter) FormatItem(item *models.Semester) *prot.Semester {
	return adapter.resource.FormatSemester(item)
}

// Program
type ProgramResourceAdapter struct {
	resource resources.ProgramResource
}

func NewProgramResourceAdapter(resource resources.ProgramResource) *ProgramResourceAdapter {
	return &ProgramResourceAdapter{resource: resource}
}

func (adapter *ProgramResourceAdapter) FormatItems(items []*models.Program) []*prot.Program {
	return adapter.resource.FormatPrograms(items)
}

func (adapter *ProgramResourceAdapter) FormatItem(item *models.Program) *prot.Program {
	return adapter.resource.FormatProgram(item)
}
