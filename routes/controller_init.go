package routes

import (
	"be-cleverschool/controllers"
	"be-cleverschool/repositories"
	"be-cleverschool/services"
)

func NewAuthController() *controllers.AuthController {
	authRepo := repositories.NewAuthRepository()
	authService := services.NewAuthService(authRepo)
	return controllers.NewAuthController(authService)
}

func NewUserController() *controllers.UserController {
	UserRepo := repositories.NewUserRepository()
	UserService := services.NewUserService(UserRepo)
	return controllers.NewUserController(UserService)
}

func NewClassController() *controllers.ClassController {
	classRepo := repositories.NewClassRepository()
	classService := services.NewClassService(classRepo)
	return controllers.NewClassController(classService)
}

func NewQuestionController() *controllers.QuestionController {
	questionRepo := repositories.NewQuestionRepository()
	questionService := services.NewQuestionService(questionRepo)
	return controllers.NewQuestionController(questionService)
}

func NewLessonController() *controllers.LessonController {
	lessonRepo := repositories.NewLessonRepository()
	lessonService := services.NewLessonService(lessonRepo)
	return controllers.NewLessonController(lessonService)
}

func NewLessonPlanController() *controllers.LessonPlanController {
	lessonPlanRepo := repositories.NewLessonPlanRepository()
	lessonPlanService := services.NewLessonPlanService(lessonPlanRepo)
	return controllers.NewLessonPlanController(lessonPlanService)
}

func NewLessonPlanPartController() *controllers.LessonPlanPartController {
	lessonPlanPartRepo := repositories.NewLessonPlanPartRepository()
	lessonPlanPartService := services.NewLessonPlanPartService(lessonPlanPartRepo)
	return controllers.NewLessonPlanPartController(lessonPlanPartService)
}

func NewExamController() *controllers.ExamController {
	examRepo := repositories.NewExamRepository()
	examService := services.NewExamService(examRepo)
	return controllers.NewExamController(examService)
}

func NewExerciseController() *controllers.ExerciseController {
	exerciseRepo := repositories.NewExerciseRepository()
	exerciseService := services.NewExerciseService(exerciseRepo)
	return controllers.NewExerciseController(exerciseService)
}

func NewHomeworkController() *controllers.HomeworkController {
	homeworkRepo := repositories.NewHomeworkRepository()
	homeworkService := services.NewHomeworkService(homeworkRepo)
	return controllers.NewHomeworkController(homeworkService)
}

func NewAssessmentController() *controllers.AssessmentController {
	repo := repositories.NewAssessmentRepository()
	service := services.NewAssessmentService(repo)
	return controllers.NewAssessmentController(service)
}

func NewAssessmentExportController() *controllers.AssessmentExportController {
	return controllers.NewAssessmentExportController()
}

func NewAssessmentCriterionController() *controllers.AssessmentCriterionController {
	repo := repositories.NewAssessmentCriterionRepository()
	service := services.NewAssessmentCriterionService(repo)
	return controllers.NewAssessmentCriterionController(service)
}

func NewAssessmentSubcriterionController() *controllers.AssessmentSubcriterionController {
	repo := repositories.NewAssessmentSubcriterionRepository()
	service := services.NewAssessmentSubcriterionService(repo)
	return controllers.NewAssessmentSubcriterionController(service)
}

func NewAssessmentCriteriaGroupController() *controllers.AssessmentCriteriaGroupController {
	repo := repositories.NewAssessmentCriteriaGroupRepository()
	service := services.NewAssessmentCriteriaGroupService(repo)
	return controllers.NewAssessmentCriteriaGroupController(service)
}

func NewProgramController() *controllers.ProgramController {
	repo := repositories.NewProgramRepository()
	service := services.NewProgramService(repo)
	return controllers.NewProgramController(service)
}

func NewCourseController() *controllers.CourseController {
	courseRepo := repositories.NewCourseRepository()
	courseService := services.NewCourseService(courseRepo)
	return controllers.NewCourseController(courseService)
}

func NewCourseScheduleController() *controllers.CourseScheduleController {
	return controllers.NewCourseScheduleController()
}

func NewSubjectController() *controllers.SubjectController {
	subjectRepo := repositories.NewSubjectRepository()
	subjectService := services.NewSubjectService(subjectRepo)
	return controllers.NewSubjectController(subjectService)
}

func NewSchoolController() *controllers.SchoolController {
	schoolRepo := repositories.NewSchoolRepository()
	schoolService := services.NewSchoolService(schoolRepo)
	return controllers.NewSchoolController(schoolService)
}

func NewProvinceController() *controllers.ProvinceController {
	provinceRepo := repositories.NewProvinceRepository()
	provinceService := services.NewProvinceService(provinceRepo)
	return controllers.NewProvinceController(provinceService)
}

func NewChapterController() *controllers.ChapterController {
	chapterRepo := repositories.NewChapterRepository()
	chapterService := services.NewChapterService(chapterRepo)
	return controllers.NewChapterController(chapterService)
}

func NewRoleController() *controllers.RoleController {
	roleRepo := repositories.NewRoleRepository()
	roleService := services.NewRoleService(roleRepo)
	return controllers.NewRoleController(roleService)
}

func NewSourceQuestionController() *controllers.SourceQuestionController {
	sourceQuestionRepo := repositories.NewSourceQuestionRepository()
	questionRepo := repositories.NewQuestionRepository()
	sourceQuestionService := services.NewSourceQuestionService(sourceQuestionRepo, questionRepo)
	return controllers.NewSourceQuestionController(sourceQuestionService)
}

func NewQuestionRelationController() *controllers.QuestionRelationController {
	questionRelationRepo := repositories.NewQuestionRelationRepository()
	questionRelationService := services.NewQuestionRelationService(questionRelationRepo)
	return controllers.NewQuestionRelationController(questionRelationService)
}

func NewMediaController() *controllers.MediaController {
	mediaRepo := repositories.NewMediaRepository()
	mediaService := services.NewMediaService(mediaRepo)
	return controllers.NewMediaController(mediaService)
}

func NewH5pController() *controllers.H5pController {
	h5pRepo := repositories.NewH5pContentRepository()
	h5pService := services.NewH5pService(h5pRepo)
	return controllers.NewH5pController(h5pService)
}

func NewSaveScoreController() *controllers.SaveScoreController {
	clonedQuestionService := services.NewClonedQuestionService(repositories.NewClonedQuestionRepository())
	// Multiple Choice
	saveScoreRepo := repositories.NewSaveScoreMultipleChoiceRepository()
	saveScoreService := services.NewSaveScoreMultipleChoiceService(saveScoreRepo, clonedQuestionService)

	// Fill In Blank
	saveScoreFIBRepo := repositories.NewSaveScoreFillInBlankRepository()
	saveScoreFIBService := services.NewSaveScoreFillInBlankService(saveScoreFIBRepo, clonedQuestionService)

	// Position
	saveScorePRepo := repositories.NewSaveScorePositionRepository()
	saveScorePService := services.NewSaveScorePositionService(saveScorePRepo, clonedQuestionService)

	// Matching
	saveScoreMRepo := repositories.NewSaveScoreMatchingRepository()
	saveScoreMService := services.NewSaveScoreMatchingService(saveScoreMRepo, clonedQuestionService)

	// Labeling
	saveScoreLRepo := repositories.NewSaveScoreLabelingRepository()
	saveScoreLService := services.NewSaveScoreLabelingService(saveScoreLRepo, clonedQuestionService)

	// Group
	saveScoreGRepo := repositories.NewSaveScoreGroupRepository()
	saveScoreGService := services.NewSaveScoreGroupService(saveScoreGRepo, clonedQuestionService)

	// Homework User Service
	homeworkUserRepo := repositories.NewHomeworkUserRepository()
	manualScoringRepo := repositories.NewSaveScoreManualScoringRepository()
	homeworkRepo := repositories.NewHomeworkRepository()
	homeworkUserService := services.NewHomeworkUserService(homeworkUserRepo, homeworkRepo, clonedQuestionService, manualScoringRepo)

	// Homework User Question Service (dùng chung cho nhiều dạng câu hỏi, bao gồm manual scoring)
	homeworkUserQuestionRepo := repositories.NewHomeworkUserQuestionRepository()
	homeworkUserQuestionService := services.NewHomeworkUserQuestionService(homeworkUserQuestionRepo, clonedQuestionService)

	// Save Answer Manual Scoring
	saveAnswerMSRepo := repositories.NewManualScoringRepository()
	saveAnswerMSService := services.NewManualScoringService(saveAnswerMSRepo, homeworkUserQuestionService)

	// Exam User Service
	examUserRepo := repositories.NewExamUserRepository()
	examUserService := services.NewExamUserService(examUserRepo, clonedQuestionService)

	// Exercise User Service
	exerciseUserRepo := repositories.NewExerciseUserRepository()
	_ = services.NewExerciseUserService(exerciseUserRepo, clonedQuestionService)

	// Bulk
	saveScoreBService := services.NewSaveScoreBulkService(
		saveScoreService,
		saveScoreFIBService,
		saveScorePService,
		saveScoreMService,
		saveScoreLService,
		saveScoreGService,
		saveAnswerMSService,
		examUserService,
	)

	// Homework Skip Question Service
	homeworkSkipQuestionService := services.NewHomeworkSkipQuestionService()

	// Homework Calculation Service
	homeworkCalculationService := services.NewHomeworkCalculationService(homeworkUserService)

	// User Star Exp Service
	userStarExpService := services.NewUserStarExpService()

	// Save Score Manual Scoring
	saveScoreManualScoringRepo := repositories.NewSaveScoreManualScoringRepository()
	saveScoreManualScoringService := services.NewSaveScoreManualScoringService(saveScoreManualScoringRepo, examUserService, homeworkUserService, homeworkUserQuestionService, homeworkCalculationService, userStarExpService)

	// Controller
	return controllers.NewSaveScoreController(saveScoreService, saveScoreFIBService, saveScorePService, saveScoreMService, saveScoreLService, saveScoreGService, saveScoreBService, saveAnswerMSService, saveScoreManualScoringService, homeworkUserService, homeworkSkipQuestionService, homeworkCalculationService, userStarExpService)
}

func NewDepartmentController() *controllers.DepartmentController {
	departmentRepo := repositories.NewDepartmentRepository()
	departmentService := services.NewDepartmentService(departmentRepo)
	return controllers.NewDepartmentController(departmentService)
}

func NewEmployeePositionController() *controllers.EmployeePositionController {
	employeePositionRepo := repositories.NewEmployeePositionRepository()
	employeePositionService := services.NewEmployeePositionService(employeePositionRepo)
	return controllers.NewEmployeePositionController(employeePositionService)
}

func NewDegreeController() *controllers.DegreeController {
	degreeRepo := repositories.NewDegreeRepository()
	degreeService := services.NewDegreeService(degreeRepo)
	return controllers.NewDegreeController(degreeService)
}

func NewCertificateController() *controllers.CertificateController {
	certificateRepo := repositories.NewCertificateRepository()
	certificateService := services.NewCertificateService(certificateRepo)
	return controllers.NewCertificateController(certificateService)
}

func NewQuestionAttributeController() *controllers.QuestionAttributeController {
	questionAttributeRepo := repositories.NewQuestionAttributeRepository()
	questionAttributeService := services.NewQuestionAttributeService(questionAttributeRepo)
	return controllers.NewQuestionAttributeController(questionAttributeService)
}

func NewTagController() *controllers.TagController {
	tagRepo := repositories.NewTagRepository()
	tagService := services.NewTagService(tagRepo)
	return controllers.NewTagController(tagService)
}

func NewTopicController() *controllers.TopicController {
	topicRepo := repositories.NewTopicRepository()
	topicService := services.NewTopicService(topicRepo)
	return controllers.NewTopicController(topicService)
}

func NewSkillController() *controllers.SkillController {
	skillRepo := repositories.NewSkillRepository()
	skillService := services.NewSkillService(skillRepo)
	return controllers.NewSkillController(skillService)
}

func NewGetExamAnswersController() *controllers.GetExamAnswersController {
	// Khởi tạo repository
	repo := repositories.NewGetExamAnswersRepository()

	// Khởi tạo service
	service := services.NewGetExamAnswersService(repo)

	// Khởi tạo controller
	return controllers.NewGetExamAnswersController(service)
}

func NewGetExerciseAnswersController() *controllers.GetExerciseAnswersController {
	repo := repositories.NewGetExerciseAnswersRepository()
	service := services.NewGetExerciseAnswersService(repo)
	return controllers.NewGetExerciseAnswersController(service)
}

func NewHomeworkAnswerController() *controllers.HomeworkAnswerController {
	repo := repositories.NewHomeworkAnswerRepository()
	service := services.NewHomeworkAnswerService(repo)
	return controllers.NewHomeworkAnswerController(service)
}

func NewStudyShiftController() *controllers.StudyShiftController {
	studyShiftRepo := repositories.NewStudyShiftRepository()
	studyShiftService := services.NewStudyShiftService(studyShiftRepo)
	return controllers.NewStudyShiftController(studyShiftService)
}

func NewWeekController() *controllers.WeekController {
	weekRepo := repositories.NewWeekRepository()
	weekService := services.NewWeekService(weekRepo)
	return controllers.NewWeekController(weekService)
}

func NewLessonScheduleController() *controllers.LessonScheduleController {
	lessonScheduleRepo := repositories.NewLessonScheduleRepository()
	lessonScheduleService := services.NewLessonScheduleService(lessonScheduleRepo)
	return controllers.NewLessonScheduleController(lessonScheduleService)
}

func NewLessonScheduleCopyController() *controllers.LessonScheduleCopyController {
	return controllers.NewLessonScheduleCopyController()
}

func NewExamStudentController() *controllers.ExamStudentController {
	examStudentRepo := repositories.NewExamStudentRepository()
	examStudentService := services.NewExamStudentService(examStudentRepo)
	return controllers.NewExamStudentController(examStudentService)
}

func NewContestStudentController() *controllers.ContestStudentController {
	contestStudentRepo := repositories.NewContestStudentRepository()
	contestStudentService := services.NewContestStudentService(contestStudentRepo)
	return controllers.NewContestStudentController(contestStudentService)
}

func NewExerciseStudentController() *controllers.ExerciseStudentController {
	repo := repositories.NewExerciseStudentRepository()
	service := services.NewExerciseStudentService(repo)
	return controllers.NewExerciseStudentController(service)
}

func NewSchoolDashboardController() *controllers.SchoolDashboardController {
	schoolDashboardRepo := repositories.NewSchoolDashboardRepository()
	schoolDashboardService := services.NewSchoolDashboardService(schoolDashboardRepo)
	return controllers.NewSchoolDashboardController(schoolDashboardService)
}

func NewHomeworkStudentController() *controllers.HomeworkStudentController {
	repo := repositories.NewHomeworkStudentRepository()
	service := services.NewHomeworkStudentService(repo)
	return controllers.NewHomeworkStudentController(service)
}

func NewExamCourseController() *controllers.ExamCourseController {
	repo := repositories.NewExamCourseRepository()
	service := services.NewExamCourseService(repo)
	return controllers.NewExamCourseController(service)
}

func NewExamCommentController() *controllers.ExamCommentController {
	repo := repositories.NewExamCommentRepository()
	service := services.NewExamCommentService(repo)
	return controllers.NewExamCommentController(service)
}

func NewExerciseCommentController() *controllers.ExerciseCommentController {
	repo := repositories.NewExerciseCommentRepository()
	service := services.NewExerciseCommentService(repo)
	return controllers.NewExerciseCommentController(service)
}

func NewHomeworkCommentController() *controllers.HomeworkCommentController {
	repo := repositories.NewHomeworkCommentRepository()
	service := services.NewHomeworkCommentService(repo)
	return controllers.NewHomeworkCommentController(service)
}

func NewDashboardStudentExamController() *controllers.DashboardStudentExamController {
	repo := repositories.NewDashboardStudentExamRepository()
	service := services.NewDashboardStudentExamService(repo)
	return controllers.NewDashboardStudentExamController(service)
}

func NewDashboardStudentHomeworkController() *controllers.DashboardStudentHomeworkController {
	repo := repositories.NewDashboardStudentHomeworkRepository()
	service := services.NewDashboardStudentHomeworkService(repo)
	return controllers.NewDashboardStudentHomeworkController(service)
}

func NewDashboardStudentExamListController() *controllers.DashboardStudentExamListController {
	repo := repositories.NewDashboardStudentExamListRepository()
	service := services.NewDashboardStudentExamListService(repo)
	return controllers.NewDashboardStudentExamListController(service)
}

func NewDashboardStudentHomeworkListController() *controllers.DashboardStudentHomeworkListController {
	repo := repositories.NewDashboardStudentHomeworkListRepository()
	service := services.NewDashboardStudentHomeworkListService(repo)
	return controllers.NewDashboardStudentHomeworkListController(service)
}

func NewDashboardStudentAssessmentController() *controllers.DashboardStudentAssessmentController {
	service := services.NewDashboardStudentAssessmentService()
	return controllers.NewDashboardStudentAssessmentController(service)
}

func NewDashboardAssessmentReportExcelController() *controllers.DashboardAssessmentReportExcelController {
	return controllers.NewDashboardAssessmentReportExcelController()
}

func NewFeedbackController() *controllers.FeedbackController {
	repo := repositories.NewFeedbackRepository()
	service := services.NewFeedbackService(repo)
	return controllers.NewFeedbackController(service)
}

func NewGradeController() *controllers.GradeController {
	repo := repositories.NewGradeRepository()
	service := services.NewGradeService(repo)
	return controllers.NewGradeController(service)
}

func NewDashboardListEntityController() *controllers.DashboardListEntityController {
	return controllers.NewDashboardListEntityController()
}

func NewDashboardTeacherExamController() *controllers.DashboardTeacherExamController {
	return controllers.NewDashboardTeacherExamController()
}

func NewDashboardTeacherExamUnscoredController() *controllers.DashboardTeacherExamUnscoredController {
	return controllers.NewDashboardTeacherExamUnscoredController()
}

func NewDashboardTeacherExamOverviewController() *controllers.DashboardTeacherExamOverviewController {
	return controllers.NewDashboardTeacherExamOverviewController()
}

func NewDashboardTeacherHomeworkStudentController() *controllers.DashboardTeacherHomeworkStudentController {
	return controllers.NewDashboardTeacherHomeworkStudentController()
}

func NewDashboardTeacherHomeworkUnscoredController() *controllers.DashboardTeacherHomeworkUnscoredController {
	return controllers.NewDashboardTeacherHomeworkUnscoredController()
}

func NewDashboardTeacherHomeworkScoredController() *controllers.DashboardTeacherHomeworkScoredController {
	return controllers.NewDashboardTeacherHomeworkScoredController()
}

func NewDashboardTeacherHomeworkListController() *controllers.DashboardTeacherHomeworkListController {
	return controllers.NewDashboardTeacherHomeworkListController()
}

func NewAssessmentStudentController() *controllers.AssessmentStudentController {
	return controllers.NewAssessmentStudentController()
}

func NewClassUserRelationController() *controllers.ClassUserRelationController {
	return controllers.NewClassUserRelationController()
}

func NewDashboardExamRankingController() *controllers.DashboardExamRankingController {
	return controllers.NewDashboardExamRankingController()
}

func NewRecalculateTotalQuestionsController() *controllers.RecalculateTotalQuestionsController {
	return controllers.NewRecalculateTotalQuestionsController()
}

func NewRecalculateHomeworkUsersController() *controllers.RecalculateHomeworkUsersController {
	return controllers.NewRecalculateHomeworkUsersController()
}

func NewChatMessageController() *controllers.ChatMessageController {
	chatMessageRepo := repositories.NewChatMessageRepository()
	userRepo := repositories.NewUserRepository()
	courseRepo := repositories.NewCourseRepository()
	messageMediaRepo := repositories.NewMessageMediaRepository()
	mediaRepo := repositories.NewMediaRepository()
	chatMessageService := services.NewChatMessageService(chatMessageRepo, userRepo, courseRepo, messageMediaRepo, mediaRepo)
	return controllers.NewChatMessageController(chatMessageService)
}

func NewChatReplyController() *controllers.ChatReplyController {
	chatMessageRepo := repositories.NewChatMessageRepository()
	userRepo := repositories.NewUserRepository()
	courseRepo := repositories.NewCourseRepository()
	replyService := services.NewChatReplyService(chatMessageRepo, userRepo, courseRepo)
	return controllers.NewChatReplyController(replyService)
}

func NewSemesterController() *controllers.SemesterController {
	repo := repositories.NewSemesterRepository()
	service := services.NewSemesterService(repo)
	return controllers.NewSemesterController(service)
}

func NewHolidayController() *controllers.HolidayController {
	repo := repositories.NewHolidayRepository()
	service := services.NewHolidayService(repo)
	return controllers.NewHolidayController(service)
}

func NewWarningController() *controllers.WarningController {
	activityLogRepo := repositories.NewMonthlyActivityLogRepository()
	warningService := services.NewWarningService(activityLogRepo)
	return controllers.NewWarningController(warningService)
}

func NewZoomAuthController() *controllers.ZoomAuthController {
	return controllers.NewZoomAuthController()
}

func NewZoomMeetingController() *controllers.ZoomMeetingController {
	return controllers.NewZoomMeetingController()
}

// Google Meet Controllers
func NewGoogleAuthController() *controllers.GoogleAuthController {
	return controllers.NewGoogleAuthController()
}

func NewGoogleMeetingController() *controllers.GoogleMeetingController {
	return controllers.NewGoogleMeetingController()
}

// Microsoft Teams Controllers
func NewMicrosoftAuthController() *controllers.MicrosoftAuthController {
	return controllers.NewMicrosoftAuthController()
}

func NewMicrosoftMeetingController() *controllers.MicrosoftMeetingController {
	return controllers.NewMicrosoftMeetingController()
}

func NewAIGradingController() *controllers.AIGradingController {
	aiGradingService := services.NewAIGradingService()
	return controllers.NewAIGradingController(aiGradingService)
}

func NewSettingController() *controllers.SettingController {
	repo := repositories.NewSettingRepository()
	settingService := services.NewSettingService(repo)
	return controllers.NewSettingController(settingService)
}

func NewNoticeController() *controllers.NoticeController {
	repo := repositories.NewNoticeRepository()
	noticeService := services.NewNoticeService(repo)
	return controllers.NewNoticeController(noticeService)
}

func NewAppConfigController() *controllers.AppConfigController {
	appConfigService := services.NewAppConfigService()
	return controllers.NewAppConfigController(appConfigService)
}

func NewFacultyController() *controllers.FacultyController {
	repo := repositories.NewFacultyRepository()
	service := services.NewFacultyService(repo)
	return controllers.NewFacultyController(service)
}

func NewStudyReportCriteriaController() *controllers.StudyReportCriteriaController {
	repo := repositories.NewStudyReportCriteriaRepository()
	service := services.NewStudyReportCriteriaService(repo)
	return controllers.NewStudyReportCriteriaController(service)
}

func NewStudyReportController() *controllers.StudyReportController {
	repo := repositories.NewStudyReportRepository()
	service := services.NewStudyReportService(repo)
	return controllers.NewStudyReportController(service)
}

func NewHeadingController() *controllers.HeadingController {
	repo := repositories.NewHeadingRepository()
	service := services.NewHeadingService(repo)
	return controllers.NewHeadingController(service)
}

func NewTrainingLevelController() *controllers.TrainingLevelController {
	repo := repositories.NewTrainingLevelRepository()
	service := services.NewTrainingLevelService(repo)
	return controllers.NewTrainingLevelController(service)
}

func NewTeachingPlanController() *controllers.TeachingPlanController {
	repo := repositories.NewTeachingPlanRepository()
	service := services.NewTeachingPlanService(repo)
	return controllers.NewTeachingPlanController(service)
}

