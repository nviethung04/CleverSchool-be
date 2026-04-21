package docs

import (
	_ "embed"
	"strings"
)

//go:embed swagger/question_path.json
var questionPathRaw string

//go:embed swagger/lesson_plan.json
var lessonPlanPathRaw string

//go:embed swagger/lesson_plan_part.json
var lessonPlanPartRaw string

//go:embed swagger/exam.json
var examRaw string

//go:embed swagger/homework.json
var homeworkRaw string

//go:embed swagger/assessment.json
var assessmentRaw string

//go:embed swagger/assessment_criterion.json
var assessmentCriterionRaw string

//go:embed swagger/assessment_subcriterion.json
var assessmentSubcriterionRaw string

//go:embed swagger/assessment_criteria_group.json
var assessmentCriteriaGroupRaw string

//go:embed swagger/do_homework.json
var doHomeworkRaw string

//go:embed swagger/question_prot.json
var questionProtRaw string

//go:embed swagger/lesson_plan_prot.json
var lessonPlanProtRaw string

//go:embed swagger/lesson_path.json
var lessonPathRaw string

//go:embed swagger/lesson_prot.json
var lessonProtRaw string

//go:embed swagger/subject_prot.json
var subjectProtRaw string

//go:embed swagger/course_prot.json
var courseProtRaw string

//go:embed swagger/course_path.json
var coursePathRaw string

//go:embed swagger/course_schedule_prot.json
var courseScheduleProtRaw string

//go:embed swagger/course_schedule_path.json
var courseSchedulePathRaw string

//go:embed swagger/school_path.json
var schoolPathRaw string

//go:embed swagger/school_prot.json
var schoolProtRaw string

//go:embed swagger/subject_path.json
var subjectPathRaw string

//go:embed swagger/media_path.json
var mediaPathRaw string

//go:embed swagger/media_prot.json
var mediaProtRaw string

//go:embed swagger/chapter_prot.json
var chapterProtRaw string

//go:embed swagger/chapter_path.json
var chapterPathRaw string

//go:embed swagger/role_prot.json
var roleProtRaw string

//go:embed swagger/role_path.json
var rolePathRaw string

//go:embed swagger/permission_prot.json
var permissionProtRaw string

//go:embed swagger/permission_path.json
var permissionPathRaw string

//go:embed swagger/user_prot.json
var userProtRaw string

//go:embed swagger/user_path.json
var userPathRaw string

//go:embed swagger/source_question_prot.json
var sourceQuestionProtRaw string

//go:embed swagger/source_question_path.json
var sourceQuestionPathRaw string

//go:embed swagger/auth_prot.json
var authProtRaw string

//go:embed swagger/auth_path.json
var authPathRaw string

//go:embed swagger/class_prot.json
var classProtRaw string

//go:embed swagger/class_path.json
var classPathRaw string

//go:embed swagger/flashcard_prot.json
var flashcardProtRaw string

//go:embed swagger/flashcard_path.json
var flashcardPathRaw string

//go:embed swagger/employee_position_prot.json
var employeePositionProtRaw string

//go:embed swagger/employee_position_path.json
var employeePositionPathRaw string

//go:embed swagger/department_prot.json
var departmentProtRaw string

//go:embed swagger/department_path.json
var departmentPathRaw string

//go:embed swagger/degree_prot.json
var degreeProtRaw string

//go:embed swagger/degree_path.json
var degreePathRaw string

//go:embed swagger/certificate_prot.json
var certificateProtRaw string

//go:embed swagger/certificate_path.json
var certificatePathRaw string

//go:embed swagger/tag_prot.json
var tagProtRaw string

//go:embed swagger/tag_path.json
var tagPathRaw string

//go:embed swagger/topic_prot.json
var topicProtRaw string

//go:embed swagger/topic_path.json
var topicPathRaw string

//go:embed swagger/question_attribute_prot.json
var questionAttributeProtRaw string

//go:embed swagger/question_attribute_path.json
var questionAttributePathRaw string

//go:embed swagger/skill_prot.json
var skillProtRaw string

//go:embed swagger/skill_path.json
var skillPathRaw string

//go:embed swagger/dashboard_prot.json
var dashboardProtRaw string

//go:embed swagger/dashboard_path.json
var dashboardPathRaw string

//go:embed swagger/save_score.json
var saveScoreRaw string

//go:embed swagger/week_prot.json
var weekProtRaw string

//go:embed swagger/week_path.json
var weekPathRaw string

//go:embed swagger/study_shift_prot.json
var studyShiftProtRaw string

//go:embed swagger/study_shift_path.json
var studyShiftPathRaw string

//go:embed swagger/lesson_schedule_prot.json
var lessonScheduleProtRaw string

//go:embed swagger/lesson_schedule_path.json
var lessonSchedulePathRaw string

//go:embed swagger/school_dashboard_path.json
var schoolDashboardPathRaw string

//go:embed swagger/grade_prot.json
var gradeProtRaw string

//go:embed swagger/grade_path.json
var gradePathRaw string

//go:embed swagger/region_prot.json
var regionProtRaw string

//go:embed swagger/region_path.json
var regionPathRaw string

//go:embed swagger/h5p_path.json
var h5pPathRaw string

//go:embed swagger/h5p_prot.json
var h5pProtRaw string

//go:embed swagger/semester_path.json
var semesterPathRaw string

//go:embed swagger/semester_prot.json
var semesterProtRaw string

//go:embed swagger/holiday_path.json
var holidayPathRaw string

//go:embed swagger/holiday_prot.json
var holidayProtRaw string

//go:embed swagger/program_path.json
var programPathRaw string

//go:embed swagger/program_prot.json
var programProtRaw string

//go:embed swagger/upload_path.json
var uploadPathRaw string

//go:embed swagger/chat_path.json
var chatPathRaw string

//go:embed swagger/warning_path.json
var warningPathRaw string

//go:embed swagger/chat_prot.json
var chatProtRaw string

//go:embed swagger/warning_prot.json
var warningProtRaw string

//go:embed swagger/exam_prot.json
var examProtRaw string

//go:embed swagger/homework_prot.json
var homeworkProtRaw string

//go:embed swagger/assessment_prot.json
var assessmentProtRaw string

//go:embed swagger/assessment_criterion_prot.json
var assessmentCriterionProtRaw string

//go:embed swagger/assessment_subcriterion_prot.json
var assessmentSubcriterionProtRaw string

//go:embed swagger/dashboard_teacher_scoring_path.json
var dashboardTeacherScoringPathRaw string

//go:embed swagger/dashboard_teacher_scoring_prot.json
var dashboardTeacherScoringProtRaw string

//go:embed swagger/contest.json
var contestRaw string

//go:embed swagger/contest_scoring.json
var contestScoringRaw string

//go:embed swagger/internal_command_path.json
var internalCommandPathRaw string

//go:embed swagger/internal_command_prot.json
var internalCommandProtRaw string

//go:embed swagger/dashboard_report_path.json
var dashboardReportPathRaw string

//go:embed swagger/dashboard_report_prot.json
var dashboardReportProtRaw string

//go:embed swagger/dashboard_list_entity_prot.json
var dashboardListEntityProtRaw string

//go:embed swagger/dashboard_schools_path.json
var dashboardSchoolsPathRaw string

//go:embed swagger/dashboard_courses_path.json
var dashboardCoursesPathRaw string

//go:embed swagger/dashboard_list_entity_path.json
var dashboardListEntityPathRaw string

//go:embed swagger/submit_homework_prot.json
var submitHomeworkProtRaw string

//go:embed swagger/setting_path.json
var settingPathRaw string

//go:embed swagger/setting_prot.json
var settingProtRaw string

//go:embed swagger/dashboard_teacher_homework_list.json
var dashboardTeacherHomeworkListPathRaw string

//go:embed swagger/dashboard_teacher_homework_list_prot.json
var dashboardTeacherHomeworkListProtRaw string

//go:embed swagger/faculty_path.json
var facultyPathRaw string

//go:embed swagger/faculty_prot.json
var facultyProtRaw string

//go:embed swagger/study_report_path.json
var studyReportPathRaw string

//go:embed swagger/study_report_prot.json
var studyReportProtRaw string

//go:embed swagger/study_report_criteria_path.json
var studyReportCriteriaPathRaw string

//go:embed swagger/study_report_criteria_prot.json
var studyReportCriteriaProtRaw string

//go:embed swagger/heading_path.json
var headingPathRaw string

//go:embed swagger/heading_prot.json
var headingProtRaw string

//go:embed swagger/notice_path.json
var noticePathRaw string

//go:embed swagger/notice_prot.json
var noticeProtRaw string

//go:embed swagger/push_firebase_path.json
var pushFirebasePathRaw string

//go:embed swagger/push_firebase_prot.json
var pushFirebaseProtRaw string

//go:embed swagger/training_level_path.json
var trainingLevelPathRaw string

//go:embed swagger/training_level_prot.json
var trainingLevelProtRaw string

//go:embed swagger/teaching_plan_path.json
var teachingPlanPathRaw string

//go:embed swagger/teaching_plan_prot.json
var teachingPlanProtRaw string

var (
	questionPath                     = trim(questionPathRaw)
	lessonPlanPath                   = trim(lessonPlanPathRaw)
	questionProt                     = trim(questionProtRaw)
	lessonPlanProt                   = trim(lessonPlanProtRaw)
	lessonPlanPart                   = trim(lessonPlanPartRaw)
	exam                             = trim(examRaw)
	homework                         = trim(homeworkRaw)
	assessment                       = trim(assessmentRaw)
	assessmentCriterion              = trim(assessmentCriterionRaw)
	assessmentSubcriterion           = trim(assessmentSubcriterionRaw)
	assessmentCriteriaGroup          = trim(assessmentCriteriaGroupRaw)
	doHomework                       = trim(doHomeworkRaw)
	courseProt                       = trim(courseProtRaw)
	coursePath                       = trim(coursePathRaw)
	courseScheduleProt               = trim(courseScheduleProtRaw)
	courseSchedulePath               = trim(courseSchedulePathRaw)
	schoolPath                       = trim(schoolPathRaw)
	schoolProt                       = trim(schoolProtRaw)
	subjectProt                      = trim(subjectProtRaw)
	subjectPath                      = trim(subjectPathRaw)
	chapterProt                      = trim(chapterProtRaw)
	chapterPath                      = trim(chapterPathRaw)
	roleProt                         = trim(roleProtRaw)
	rolePath                         = trim(rolePathRaw)
	permissionProt                   = trim(permissionProtRaw)
	permissionPath                   = trim(permissionPathRaw)
	authProt                         = trim(authProtRaw)
	authPath                         = trim(authPathRaw)
	userProt                         = trim(userProtRaw)
	userPath                         = trim(userPathRaw)
	sourceQuestionPath               = trim(sourceQuestionPathRaw)
	sourceQuestionProt               = trim(sourceQuestionProtRaw)
	lessonPath                       = trim(lessonPathRaw)
	lessonProt                       = trim(lessonProtRaw)
	mediaPath                        = trim(mediaPathRaw)
	mediaProt                        = trim(mediaProtRaw)
	classPath                        = trim(classPathRaw)
	classProt                        = trim(classProtRaw)
	employeePositionPath             = trim(employeePositionPathRaw)
	employeePositionProt             = trim(employeePositionProtRaw)
	departmentPath                   = trim(departmentPathRaw)
	departmentsProt                  = trim(departmentProtRaw)
	degreePath                       = trim(degreePathRaw)
	degreeProt                       = trim(degreeProtRaw)
	certificatePath                  = trim(certificatePathRaw)
	certificateProt                  = trim(certificateProtRaw)
	tagPath                          = trim(tagPathRaw)
	tagProt                          = trim(tagProtRaw)
	topicPath                        = trim(topicPathRaw)
	topicProt                        = trim(topicProtRaw)
	questionAttributePath            = trim(questionAttributePathRaw)
	questionAttributeProt            = trim(questionAttributeProtRaw)
	skillPath                        = trim(skillPathRaw)
	skillProt                        = trim(skillProtRaw)
	dashboardPath                    = trim(dashboardPathRaw)
	dashboardProt                    = trim(dashboardProtRaw)
	saveScore                        = trim(saveScoreRaw)
	weekPath                         = trim(weekPathRaw)
	weekProt                         = trim(weekProtRaw)
	gradePath                        = trim(gradePathRaw)
	gradeProt                        = trim(gradeProtRaw)
	regionPath                       = trim(regionPathRaw)
	regionProt                       = trim(regionProtRaw)
	studyShiftPath                   = trim(studyShiftPathRaw)
	studyShiftProt                   = trim(studyShiftProtRaw)
	lessonSchedulePath               = trim(lessonSchedulePathRaw)
	lessonScheduleProt               = trim(lessonScheduleProtRaw)
	schoolDashboardPath              = trim(schoolDashboardPathRaw)
	h5pPath                          = trim(h5pPathRaw)
	h5pProt                          = trim(h5pProtRaw)
	semesterPath                     = trim(semesterPathRaw)
	semesterProt                     = trim(semesterProtRaw)
	holidayPath                      = trim(holidayPathRaw)
	holidayProt                      = trim(holidayProtRaw)
	programPath                      = trim(programPathRaw)
	programProt                      = trim(programProtRaw)
	uploadPath                       = trim(uploadPathRaw)
	chatPath                         = trim(chatPathRaw)
	chatProt                         = trim(chatProtRaw)
	warningPath                      = trim(warningPathRaw)
	warningProt                      = trim(warningProtRaw)
	flashcardPath                    = trim(flashcardPathRaw)
	flashcardProt                    = trim(flashcardProtRaw)
	examProt                         = trim(examProtRaw)
	homeworkProt                     = trim(homeworkProtRaw)
	assessmentProt                   = trim(assessmentProtRaw)
	assessmentCriterionProt          = trim(assessmentCriterionProtRaw)
	assessmentSubcriterionProt       = trim(assessmentSubcriterionProtRaw)
	dashboardTeacherScoringPath      = trim(dashboardTeacherScoringPathRaw)
	dashboardTeacherScoringProt      = trim(dashboardTeacherScoringProtRaw)
	internalCommandPath              = trim(internalCommandPathRaw)
	internalCommandProt              = trim(internalCommandProtRaw)
	dashboardReportPath              = trim(dashboardReportPathRaw)
	dashboardReportProt              = trim(dashboardReportProtRaw)
	dashboardListEntityProt          = trim(dashboardListEntityProtRaw)
	dashboardSchoolsPath             = trim(dashboardSchoolsPathRaw)
	dashboardCoursesPath             = trim(dashboardCoursesPathRaw)
	dashboardListEntityPath          = trim(dashboardListEntityPathRaw)
	contest                          = trim(contestRaw)
	contestScoring                   = trim(contestScoringRaw)
	submitHomeworkProt               = trim(submitHomeworkProtRaw)
	settingPath                      = trim(settingPathRaw)
	settingProt                      = trim(settingProtRaw)
	dashboardTeacherHomeworkListPath = trim(dashboardTeacherHomeworkListPathRaw)
	dashboardTeacherHomeworkListProt = trim(dashboardTeacherHomeworkListProtRaw)
	facultyPath                      = trim(facultyPathRaw)
	facultyProt                      = trim(facultyProtRaw)
	studyReportPath                  = trim(studyReportPathRaw)
	studyReportProt                  = trim(studyReportProtRaw)
	studyReportCriteriaPath          = trim(studyReportCriteriaPathRaw)
	studyReportCriteriaProt          = trim(studyReportCriteriaProtRaw)
	headingPath                      = trim(headingPathRaw)
	headingProt                      = trim(headingProtRaw)
	noticePath                       = trim(noticePathRaw)
	noticeProt                       = trim(noticeProtRaw)
	pushFirebasePath                 = trim(pushFirebasePathRaw)
	pushFirebaseProt                 = trim(pushFirebaseProtRaw)
	trainingLevelPath                = trim(trainingLevelPathRaw)
	trainingLevelProt                = trim(trainingLevelProtRaw)
	teachingPlanPath                 = trim(teachingPlanPathRaw)
	teachingPlanProt                 = trim(teachingPlanProtRaw)
)

// Hàm cắt {}
func trim(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '{' && s[len(s)-1] == '}' {
		return s[1 : len(s)-1]
	}
	return s
}
