package routes

import (
	"be-lms/command"
	"be-lms/controllers"
	"be-lms/middleware"
	"be-lms/repositories"
	"be-lms/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ====== INIT ROUTES ======
func InitRoutes(router *gin.Engine) {
	RegisterCliRoutes(router)
	RegisterLogRoute(router)
	RouteMedia(router)
	RouteAuth(router)
	RouteH5p(router)
	RoutePowerPoint(router)
	RouteDashboard(router)
	InitStatic(router)
	RegisterScormRoutes(router)
	RegisterMigrationRoutes(router)

	// Upload routes (no auth required)
	uploadController := controllers.NewUploadController()
	router.GET("/upload", uploadController.ShowUploadPage)
	router.GET("/api/upload/stats", uploadController.GetUploadStats)

	// Upload to s3
	uploadS3Controller := controllers.NewUploadS3Controller()
	router.GET("/upload/s3", uploadS3Controller.ShowUploadPage)

	// tusd resumable uploads endpoint (no auth)
	tusHandler := services.GetTusHandler()
	tusGroup := router.Group("/api/tus-uploads")
	tusGroup.Use(middleware.RoleMiddleware("medias.update"))
	{
		tusGroup.Any("", gin.WrapH(http.StripPrefix("/api/tus-uploads", tusHandler)))
		tusGroup.Any("/*any", gin.WrapH(http.StripPrefix("/api/tus-uploads", tusHandler)))
	}

	api := router.Group("/api")
	authRepo := repositories.NewAuthRepository()

	apiAssessmentScoringController := NewAssessmentScoringController()
	api.POST("/assessments/save-score/bulk", middleware.AuthMiddleware(authRepo), apiAssessmentScoringController.SaveScoresBulk)

	//api.POST("/refresh", authController.Register)

	userController := NewUserController()
	api.GET("/student-parent/:parent_id", userController.GetStudentsByParent)

	settingController := NewSettingController()
	api.GET("/manage/settings/by-key/:key", settingController.GetByKey)

	managementRouter := api.Group("/manage")
	managementRouter.Use(middleware.AuthMiddleware(authRepo))
	{

		classController := NewClassController()
		questionController := NewQuestionController()
		lessonPlanController := NewLessonPlanController()
		lessonController := NewLessonController()
		lessonPlanPartController := NewLessonPlanPartController()
		examController := NewExamController()
		exerciseController := NewExerciseController()
		homeworkController := NewHomeworkController()
		programController := NewProgramController()
		courseController := NewCourseController()
		schoolController := NewSchoolController()
		subjectController := NewSubjectController()
		chapterController := NewChapterController()
		roleController := NewRoleController()
		sourceQuestionController := NewSourceQuestionController()
		questionRelationController := NewQuestionRelationController()
		provinceController := NewProvinceController()

		departmentController := NewDepartmentController()
		employeePositionController := NewEmployeePositionController()
		degreeController := NewDegreeController()
		certificateController := NewCertificateController()

		tagController := NewTagController()
		topicController := NewTopicController()
		skillController := NewSkillController()
		questionAttributeController := NewQuestionAttributeController()

		studyShiftController := NewStudyShiftController()
		weekController := NewWeekController()
		lessonScheduleController := NewLessonScheduleController()

		gradeController := NewGradeController()
		assessmentController := NewAssessmentController()
		settingController := NewSettingController()

		// Semester controller
		semesterController := NewSemesterController()
		holidayController := NewHolidayController()

		// Chat controller
		chatMessageController := NewChatMessageController()

		chatReactionController := NewChatMessageReactionController()

		// Chat reply controller
		chatReplyController := NewChatReplyController()

		// Contest controllers
		contestController := NewContestController()
		contestRoundController := NewContestRoundController()

		RegisterModuleRoute(managementRouter, "questions", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, questionController)
		RegisterModuleRoute(managementRouter, "classes", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, classController)
		RegisterModuleRoute(managementRouter, "lessons", []string{"index", "show", "store", "update", "destroy", "restore"}, lessonController)
		RegisterModuleRoute(managementRouter, "lesson-plans", []string{"index", "show", "store", "update", "destroy"}, lessonPlanController)
		RegisterModuleRoute(managementRouter, "lesson-plan-parts", []string{"index", "show", "store", "update", "destroy"}, lessonPlanPartController)
		RegisterModuleRoute(managementRouter, "exams", []string{"index", "show", "store", "update", "destroy"}, examController)
		RegisterModuleRoute(managementRouter, "exercises", []string{"index", "show", "store", "update", "destroy"}, exerciseController)
		RegisterModuleRoute(managementRouter, "homeworks", []string{"index", "show", "store", "update", "destroy"}, homeworkController)
		RegisterModuleRoute(managementRouter, "assessments", []string{"index", "show", "store", "update", "destroy"}, assessmentController)
		RegisterModuleRoute(managementRouter, "settings", []string{"index", "show", "store", "update", "destroy"}, settingController)

		assessmentScoringController := NewAssessmentScoringController()
		managementRouter.GET("/publish-assessments", assessmentScoringController.GetPublish)
		managementRouter.PUT("/publish-assessments", assessmentScoringController.SetPublish)
		managementRouter.PUT("/publish-study-report", assessmentScoringController.SetStudyReportPublish)
		managementRouter.POST("/assessment-criteria-group/with-criteria", assessmentScoringController.CreateCriteriaGroup)
		managementRouter.PUT("/assessment-criteria-group/with-criteria/:id", assessmentScoringController.UpdateCriteriaGroup)
		managementRouter.GET("/assessment-criteria-groups", assessmentScoringController.ListCriteriaGroups)
		managementRouter.GET("/assessment-criteria-groups/:id", assessmentScoringController.GetCriteriaGroup)
		managementRouter.DELETE("/assessment-criteria-groups/:id", assessmentScoringController.DeleteCriteriaGroup)
		managementRouter.GET("/study-report-criterias", assessmentScoringController.ListStudyReportCriterias)
		managementRouter.GET("/study-report-criterias/:id", assessmentScoringController.GetStudyReportCriteria)
		managementRouter.POST("/study-report-criterias", assessmentScoringController.CreateStudyReportCriteria)
		managementRouter.PUT("/study-report-criterias/:id", assessmentScoringController.UpdateStudyReportCriteria)
		managementRouter.DELETE("/study-report-criterias/:id", assessmentScoringController.DeleteStudyReportCriteria)
		managementRouter.GET("/study-reports/evaluates/:course_id", assessmentScoringController.GetStudyReportEvaluates)
		managementRouter.GET("/study-reports", assessmentScoringController.ListStudyReports)
		managementRouter.GET("/study-reports/:id", assessmentScoringController.GetStudyReportDetail)
		managementRouter.POST("/study-reports", assessmentScoringController.CreateStudyReport)
		managementRouter.PUT("/study-reports/:id", assessmentScoringController.UpdateStudyReport)
		RegisterModuleRoute(managementRouter, "contests", []string{"index", "show", "store", "update", "destroy", "restore"}, contestController)
		RegisterModuleRoute(managementRouter, "contest-rounds", []string{"index", "show", "store", "update", "destroy", "restore"}, contestRoundController)

		// Contest custom routes
		managementRouter.GET("/contests/:id/rounds", contestController.GetContestRounds)
		managementRouter.GET("/contests/:id/with-rounds", contestController.GetContestWithRounds)

		// Contest Round custom routes
		managementRouter.GET("/contests/:id/contest-rounds", contestRoundController.GetByContestId)
		managementRouter.GET("/contest-rounds/:id/users", contestRoundController.GetContestRoundUsers)
		managementRouter.GET("/contest-rounds/:id/joiners", contestRoundController.GetContestRoundJoiners)
		managementRouter.POST("/contest-rounds/:id/joiners", contestRoundController.AddJoiner)
		managementRouter.DELETE("/contest-rounds/:id/joiners", contestRoundController.RemoveJoiner)
		RegisterModuleRoute(managementRouter, "programs", []string{"index", "show", "store", "update", "destroy", "restore"}, programController)
		RegisterModuleRoute(managementRouter, "courses", []string{"index", "show", "store", "update", "destroy", "restore"}, courseController)
		RegisterModuleRoute(managementRouter, "schools", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, schoolController)
		RegisterModuleRoute(managementRouter, "subjects", []string{"index", "show", "store", "update", "destroy", "restore"}, subjectController)
		RegisterModuleRoute(managementRouter, "chapters", []string{"index", "show", "store", "update", "destroy", "restore"}, chapterController)
		RegisterModuleRoute(managementRouter, "roles", []string{"index", "show", "store", "update", "destroy"}, roleController)
		RegisterModuleRoute(managementRouter, "users", []string{"index", "show", "store", "update", "destroy", "export", "import", "restore"}, userController)
		RegisterModuleRoute(managementRouter, "source-questions", []string{"index", "show", "store", "update", "destroy", "restore"}, sourceQuestionController)

		RegisterModuleRoute(managementRouter, "departments", []string{"index", "show", "store", "update", "destroy"}, departmentController)
		RegisterModuleRoute(managementRouter, "employee-positions", []string{"index", "show", "store", "update", "destroy"}, employeePositionController)
		RegisterModuleRoute(managementRouter, "degrees", []string{"index", "show", "store", "update", "destroy"}, degreeController)
		RegisterModuleRoute(managementRouter, "certificates", []string{"index", "show", "store", "update", "destroy"}, certificateController)

		RegisterModuleRoute(managementRouter, "tags", []string{"index", "show", "store", "update", "destroy", "restore"}, tagController)
		RegisterModuleRoute(managementRouter, "topics", []string{"index", "show", "store", "update", "destroy", "restore"}, topicController)
		RegisterModuleRoute(managementRouter, "skills", []string{"index", "show", "store", "update", "destroy", "restore"}, skillController)

		RegisterModuleRoute(managementRouter, "grades", []string{"index", "show", "store", "update", "destroy", "restore"}, gradeController)

		RegisterModuleRoute(managementRouter, "question-attributes", []string{"index", "show", "store", "update", "destroy", "restore"}, questionAttributeController)
		managementRouter.GET("/question-attributes/parent", middleware.RoleMiddleware("question-attributes.index"), questionAttributeController.GetParents)
		RegisterModuleRoute(managementRouter, "study-shifts", []string{"index", "show", "store", "update", "destroy", "restore"}, studyShiftController)

		// Lesson for course
		managementRouter.GET("/lessons/lesson-plans/:lesson_id/:course_id", middleware.RoleMiddleware("homeworks.index"), lessonController.GetLessonPlanByCourse)
		managementRouter.GET("/lessons/exams/:lesson_id/:course_id", middleware.RoleMiddleware("homeworks.index"), lessonController.GetExamByCourse)
		managementRouter.GET("/lessons/homeworks/:lesson_id/:course_id", middleware.RoleMiddleware("homeworks.index"), lessonController.GetHomeworkByCourse)
		managementRouter.GET("/lessons/exercises/:lesson_id/:course_id", middleware.RoleMiddleware("homeworks.index"), lessonController.GetExerciseByCourse)
		managementRouter.POST("/lessons/lesson-plans/:lesson_id/:course_id", middleware.RoleMiddleware("lesson-plans.update"), lessonController.StoreLessonPlanByCourse)
		managementRouter.POST("/lessons/exams/:lesson_id/:course_id", middleware.RoleMiddleware("chapters.update"), lessonController.StoreExamByCourse)
		managementRouter.POST("/lessons/homeworks/:lesson_id/:course_id", middleware.RoleMiddleware("questions.update"), lessonController.StoreHomeworkByCourse)
		managementRouter.POST("/lessons/exercises/:lesson_id/:course_id", middleware.RoleMiddleware("questions.update"), lessonController.StoreExerciseByCourse)

		// Route xuất PDF danh sách tài khoản
		managementRouter.GET("/users/export-pdf", middleware.RoleMiddleware("users.export"), userController.ExportUsersPDF)

		managementRouter.GET("/users/activated", middleware.RoleMiddleware("users.index"), userController.UserActivated)
		managementRouter.PUT("/users/:id/reset-password", middleware.RoleMiddleware("users.update"), userController.ResetPassword)

		// Warning System Routes
		warningController := NewWarningController()
		managementRouter.GET("/warnings/active-users-count", middleware.RoleMiddleware("users.index"), warningController.GetActiveUsersCount)
		managementRouter.GET("/warnings/active-users", middleware.RoleMiddleware("users.index"), warningController.GetActiveUsers)
		managementRouter.GET("/warnings/failed-logins-count", middleware.RoleMiddleware("users.index"), warningController.GetFailedLoginsCount)
		managementRouter.GET("/warnings/failed-logins", middleware.RoleMiddleware("users.index"), warningController.GetFailedLogins)

		// Clone data
		managementRouter.POST("/homeworks/:id/cloned", middleware.RoleMiddleware("homeworks.store"), homeworkController.Cloned)
		managementRouter.POST("/exams/:id/cloned", middleware.RoleMiddleware("exams.store"), examController.Cloned)
		managementRouter.POST("/exercises/:id/cloned", middleware.RoleMiddleware("exercises.store"), exerciseController.Cloned)

		// Assigned data
		managementRouter.POST("/homeworks/:id/assigned", middleware.RoleMiddleware("homeworks.store"), homeworkController.Assigned)
		managementRouter.POST("/exams/:id/assigned", middleware.RoleMiddleware("exams.store"), examController.Assigned)
		managementRouter.POST("/exercises/:id/assigned", middleware.RoleMiddleware("exercises.store"), exerciseController.Assigned)
		managementRouter.POST("/assessments/:id/assigned", middleware.RoleMiddleware("assessments.store"), assessmentController.Assigned)

		// Assigned lesson
		managementRouter.GET("/homeworks/:id/assigned-lessons", middleware.RoleMiddleware("homeworks.index"), homeworkController.AssignedLesson)
		managementRouter.GET("/exams/:id/assigned-lessons", middleware.RoleMiddleware("exams.index"), examController.AssignedLesson)
		managementRouter.GET("/exercises/:id/assigned-lessons", middleware.RoleMiddleware("exercises.index"), exerciseController.AssignedLesson)
		managementRouter.GET("/assessments/:id/assigned-lessons", middleware.RoleMiddleware("assessments.index"), assessmentController.AssignedLesson)

		managementRouter.PUT("/lesson-plans/:id/complete", middleware.RoleMiddleware("lesson-plans.update"), lessonPlanController.Complete)
		managementRouter.PUT("/chapters/:id/sort-lessons", middleware.RoleMiddleware("chapters.update"), chapterController.SortLessons)
		managementRouter.PUT("/questions/sync-keywords", middleware.RoleMiddleware("questions.update"), questionController.SyncKeywords)

		// Semester
		managementRouter.GET("/semesters", semesterController.GetAll)
		managementRouter.GET("/semesters/:id", semesterController.GetByID)
		managementRouter.POST("/semesters", semesterController.Create)
		managementRouter.PUT("/semesters/:id", semesterController.Update)
		managementRouter.DELETE("/semesters/:id", semesterController.Delete)

		// Holiday
		managementRouter.GET("/holidays", holidayController.GetAll)
		managementRouter.GET("/holidays/:id", holidayController.GetByID)
		managementRouter.POST("/holidays", holidayController.Create)
		managementRouter.PUT("/holidays/:id", holidayController.Update)
		managementRouter.DELETE("/holidays/:id", holidayController.Delete)

		managementRouter.GET("/weeks", weekController.GetAll)
		managementRouter.GET("/weeks/by-date", weekController.GetWeekByDate)
		managementRouter.GET("/lesson-schedules", lessonScheduleController.GetAll)
		managementRouter.GET("/lessons/schedules", middleware.RoleMiddleware("lessons.show"), lessonController.LessonSchedules)
		managementRouter.POST("/lessons/schedules", middleware.RoleMiddleware("lessons.update"), lessonController.StoreLessonSchedules)

		// Lesson Schedule Copy
		lessonScheduleCopyController := NewLessonScheduleCopyController()
		managementRouter.POST("/lesson-schedules/copy", middleware.RoleMiddleware("lesson_schedule.store"), lessonScheduleCopyController.CopyLessonSchedules)

		managementRouter.PUT("/profile", middleware.AuthMiddleware(authRepo), userController.UpdateProfile)
		managementRouter.GET("/users/permissions", middleware.RoleMiddleware("users.show"), userController.GetPermissions)
		managementRouter.GET("/users/username-exists/:username", userController.UsernameExits)
		managementRouter.GET("/users/class", userController.CurrentClass)

		managementRouter.POST("/question-relation", questionRelationController.Create)
		managementRouter.GET("/profile", userController.GetMyProfile)

		managementRouter.GET("/roles/:id/permissions", middleware.RoleMiddleware("roles.show"), roleController.GetPermissions)
		managementRouter.GET("/permissions", middleware.RoleMiddleware("permissions.show"), roleController.GetAllPermissions)
		managementRouter.POST("/permissions", middleware.RoleMiddleware("permissions.update"), roleController.AssignPermissions)
		managementRouter.GET("/permissions/display", middleware.RoleMiddleware("permissions.show"), roleController.GetDisplayPermissions)
		managementRouter.POST("/permissions/display", middleware.RoleMiddleware("permissions.update"), roleController.UpdateDisplayPermissions)

		managementRouter.GET("/provinces", provinceController.GetAll)
		managementRouter.GET("/provinces/:code", provinceController.GetByID)
		managementRouter.GET("/provinces/:code/wards", provinceController.Wards)

		managementRouter.GET("/courses/:id/users", middleware.RoleMiddleware("courses.show"), courseController.GetUsers)

		managementRouter.POST("/courses/:id/users", middleware.RoleMiddleware("courses.update"), courseController.StoreUsers)
		managementRouter.PUT("/courses/:id/users", middleware.RoleMiddleware("courses.update"), courseController.AddUsers)
		managementRouter.GET("/courses/:id/score", middleware.RoleMiddleware("courses.show"), courseController.GetScore)
		managementRouter.POST("/courses/resync-schedules", middleware.RoleMiddleware("courses.update"), courseController.ResyncSchedule)
		managementRouter.GET("/course-schedule/family", middleware.RoleMiddleware("courses.show"), courseController.GetCourseFamily)
		managementRouter.POST("/course-schedule/sync-all-family", middleware.RoleMiddleware("courses.update"), courseController.SyncAllFamilyCourses)

		managementRouter.GET("/classes/:id/users", middleware.RoleMiddleware("classes.show"), classController.GetUsers)
		managementRouter.POST("/classes/:id/users", middleware.RoleMiddleware("classes.update"), classController.StoreUsers)
		managementRouter.PUT("/classes/:id/users", middleware.RoleMiddleware("classes.update"), classController.AddUsers)
		schoolDashboardController := NewSchoolDashboardController()
		managementRouter.GET("/school-dashboard-starts", schoolDashboardController.GetSchoolSummary)

		managementRouter.PUT("/lessons/:id/completion", middleware.RoleMiddleware("lessons.show"), lessonController.Completion)

		managementRouter.PUT("/class/user-relation", NewClassUserRelationController().UpdateUserClassRelation)

		// Flashcard Routes
		flashcardController := controllers.NewFlashcardController()

		// Vocabulary routes
		managementRouter.POST("/flashcard/vocabularies", flashcardController.CreateVocabulary)
		managementRouter.GET("/flashcard/vocabularies", flashcardController.GetAllVocabularies)
		managementRouter.GET("/flashcard/vocabularies/:vocabularyId", flashcardController.GetVocabulary)

		// Lesson vocabulary routes
		managementRouter.GET("/flashcard/lessons/:lessonId/vocabularies", flashcardController.GetLessonFlashcards)
		managementRouter.POST("/flashcard/lessons/:lessonId/vocabularies", flashcardController.AddVocabulariesToLesson)
		managementRouter.PUT("/flashcard/lessons/:lessonId/vocabularies", flashcardController.SyncLessonVocabularies)

		// Study session routes
		managementRouter.POST("/flashcard/sessions", flashcardController.StartFlashcardSession)
		managementRouter.POST("/flashcard/sessions/:sessionId/activities", flashcardController.RecordFlashcardActivity)
		managementRouter.PUT("/flashcard/sessions/:sessionId/complete", flashcardController.CompleteFlashcardSession)
		managementRouter.POST("/flashcard/sessions/:sessionId/resume", flashcardController.ResumeFlashcardSession)

		// Progress and session management routes
		managementRouter.GET("/flashcard/lessons/:lessonId/progress", flashcardController.GetLessonProgress)
		managementRouter.GET("/flashcard/lessons/:lessonId/active-session", flashcardController.GetActiveSession)

		// Student progress routes
		managementRouter.PUT("/flashcard/vocabulary-progress/:vocabularyId", flashcardController.UpdateVocabularyProgress)
		managementRouter.POST("/flashcard-sessions", flashcardController.StartFlashcardSession)
		managementRouter.POST("/flashcard-sessions/:sessionId/activities", flashcardController.RecordFlashcardActivity)

		// Student progress routes
		managementRouter.PUT("/vocabulary-progress/:vocabularyId", flashcardController.UpdateVocabularyProgress) // Chat Message Routes
		managementRouter.POST("/courses/:id/chat/messages", chatMessageController.SendMessageWithMedias)
		managementRouter.POST("/courses/:id/chat/messages/upload-files-to-medias", chatMessageController.UploadFilesToMedias)
		managementRouter.GET("/courses/:id/chat/messages", chatMessageController.GetMessages)
		managementRouter.DELETE("/courses/:id/chat/messages/:messageId", chatMessageController.DeleteMessage)
		managementRouter.POST("/courses/:id/chat/messages/:messageId/pin", chatMessageController.TogglePinMessage)
		managementRouter.GET("/courses/:id/chat/messages/pinned", chatMessageController.GetPinnedMessages)
		managementRouter.GET("/courses/:id/chat/messages/count", chatMessageController.GetMessageCount)
		managementRouter.GET("/courses/:id/chat/messages/recent", chatMessageController.GetRecentMessages)

		// Chat Message Reaction Routes
		managementRouter.POST("/courses/:id/chat/messages/:messageId/reactions", chatReactionController.AddReaction)
		managementRouter.DELETE("/courses/:id/chat/messages/:messageId/reactions", chatReactionController.RemoveReaction)
		managementRouter.GET("/courses/:id/chat/messages/:messageId/reactions", chatReactionController.GetMessageReactions)
		managementRouter.DELETE("/courses/:id/chat/messages/:messageId/reactions/all", chatReactionController.DeleteAllReactions)

		// Chat Message Reply Routes
		managementRouter.POST("/courses/:id/chat/messages/:messageId/replies", chatReplyController.SendReply)
		// managementRouter.GET("/courses/:id/chat/messages/:messageId/replies", chatReplyController.GetReplies) // Tạm thời comment - chưa cần thiết
		managementRouter.GET("/courses/:id/chat/messages/:messageId/with-replies", chatReplyController.GetMessageWithReplies)
		// managementRouter.DELETE("/courses/:id/chat/messages/:messageId/replies/:replyId", chatReplyController.DeleteReply) // Không cần - đã có cascade delete trong DeleteMessage

	}

	// Chat SSE (Server-Sent Events) Routes for real-time notifications - no auth middleware needed (self-authenticated)
	chatSSEController := controllers.NewChatSSEController()
	api.GET("/manage/courses/:id/chat/events", chatSSEController.StreamCourseEvents)
	api.GET("/manage/courses/events", chatSSEController.StreamMultipleCourseEvents)
	api.GET("/manage/users/events", chatSSEController.StreamUserEvents)
	managementRouter.GET("/chat/channels", chatSSEController.GetActiveChannels)
	managementRouter.GET("/chat/channels/stats", chatSSEController.GetChannelStats)

	// Feedback routes
	feedbackController := NewFeedbackController()
	RegisterModuleRoute(managementRouter, "feedbacks", []string{"index", "show", "store", "update", "destroy"}, feedbackController)
	managementRouter.PATCH("/feedbacks/:id/status", middleware.RoleMiddleware("feedbacks.update"), feedbackController.UpdateStatus)

	studyRouter := api.Group("/study")
	studyRouter.Use(middleware.AuthMiddleware(authRepo))
	{
		saveScoreController := NewSaveScoreController()
		studyRouter.POST("/save-score/multiple-choice", saveScoreController.SaveScoreMultipleChoice)
		studyRouter.POST("/save-score/fill-in-blank", saveScoreController.SaveScoreFillInBlank)
		studyRouter.POST("/save-score/ordering-and-dragdrop", saveScoreController.SaveScorePosition)
		studyRouter.POST("/save-score/matching", saveScoreController.SaveScoreMatching)
		studyRouter.POST("/save-score/labeling", saveScoreController.SaveScoreLabeling)
		studyRouter.POST("/save-score/category", saveScoreController.SaveScoreGroup)
		studyRouter.POST("/save-answer/manual-scoring", saveScoreController.SaveManualScoring)
		studyRouter.POST("/save-score/manual-scoring", saveScoreController.SaveScoreManualScoring)
		studyRouter.POST("/save-score/bulk", saveScoreController.SaveScoreBulk)
		studyRouter.POST("/save-score/submit-homework", saveScoreController.SubmitHomework)

		// Skip question
		skipQuestionController := controllers.NewSkipQuestionController()
		studyRouter.POST("/save-score/skip-question", skipQuestionController.SkipQuestion)
		studyRouter.GET("/save-score/check-submit-homework", skipQuestionController.CheckSubmitHomework)

		// Get exam answers
		examAnswersController := NewGetExamAnswersController()
		exerciseAnswersController := NewGetExerciseAnswersController()
		homeworkAnswerController := NewHomeworkAnswerController()
		examStudentController := NewExamStudentController()

		studyRouter.GET("/teacher/exam-answers", examAnswersController.GetTeacherExamAnswers)
		studyRouter.GET("/teacher/exercise-answers", exerciseAnswersController.GetTeacherExerciseAnswers)
		studyRouter.GET("/student/exam-answers", examAnswersController.GetStudentExamAnswers)
		studyRouter.GET("/student/exercise-answers", exerciseAnswersController.GetStudentExerciseAnswers)
		studyRouter.GET("/student/exams-by-week", examStudentController.GetExamByStudent)

		// Get exam/exercise students
		studyRouter.GET("/exam-students", examStudentController.GetExamStudents)
		exerciseStudentController := NewExerciseStudentController()
		exerciseControllerForStudy := NewExerciseController()
		// HS làm bài luyện tập: Auth + kiểm tra ghi danh/khóa (không cần exercises.show trên /manage)
		studyRouter.GET("/student/exercises/:id", exerciseControllerForStudy.GetByID)
		// exercise/homework comment
		exerciseCommentCtl := NewExerciseCommentController()
		homeworkCommentCtl := NewHomeworkCommentController()
		studyRouter.POST("/exercise-comment", exerciseCommentCtl.PostExerciseComment)
		studyRouter.POST("/homework-comment", homeworkCommentCtl.PostHomeworkComment)
		studyRouter.GET("/exercise-students", exerciseStudentController.GetExerciseStudents)
		studyRouter.POST("/exercise-reset", exerciseStudentController.ResetExerciseAttempt)

		examCourseController := NewExamCourseController()
		studyRouter.GET("/exam-courses", examCourseController.GetExamCourseDetail)

		homeworkStudentController := NewHomeworkStudentController()
		studyRouter.GET("/homework-students", homeworkStudentController.GetHomeworkStudents)
		studyRouter.GET("/teacher/homework-answers", homeworkAnswerController.GetTeacherHomeworkAnswer)
		studyRouter.GET("/student/homework-answers", homeworkAnswerController.GetStudentHomeworkAnswer)

		examCommentController := NewExamCommentController()
		studyRouter.POST("/exam-comment", examCommentController.PostExamComment)

		assessmentScoringStudyController := NewAssessmentScoringController()
		studyRouter.POST("/assessment-submit", assessmentScoringStudyController.StudentSubmit)
	}

	dashboardRouter := api.Group("/dashboard")
	dashboardRouter.Use(middleware.AuthMiddleware(authRepo))
	{
		dashboardStudentExamController := NewDashboardStudentExamController()
		dashboardStudentHomeworkController := NewDashboardStudentHomeworkController()
		dashboardStudentExerciseController := NewDashboardStudentExerciseController()
		dashboardStudentExamListController := NewDashboardStudentExamListController()
		dashboardStudentHomeworkListController := NewDashboardStudentHomeworkListController()
		dashboardStudentExerciseListController := NewDashboardStudentExerciseListController()
		dashboardStudentAssessmentController := NewDashboardStudentAssessmentController()
		dashboardListEntityController := NewDashboardListEntityController()
		dashboardAssessmentScoringController := NewAssessmentScoringController()
		dashboardRouter.GET("/student/exam", middleware.AuthMiddleware(authRepo), dashboardStudentExamController.GetStudentExamStats)
		dashboardRouter.GET("/student/homework", middleware.AuthMiddleware(authRepo), dashboardStudentHomeworkController.GetStudentHomeworkStats)
		dashboardRouter.GET("/student/exercise", middleware.AuthMiddleware(authRepo), dashboardStudentExerciseController.GetStudentExerciseStats)
		dashboardRouter.GET("/student/exam-list", dashboardStudentExamListController.GetStudentExamList)
		dashboardRouter.GET("/student/homework-list", dashboardStudentHomeworkListController.GetStudentHomeworkList)
		dashboardRouter.GET("/student/exercise-list", dashboardStudentExerciseListController.GetStudentExerciseList)
		dashboardRouter.GET("/student/assessment-list", dashboardStudentAssessmentController.GetStudentAssessmentList)
		dashboardRouter.GET("/student/assessments", dashboardStudentAssessmentController.GetStudentAssessments)
		dashboardRouter.GET("/teacher/assessment/students", dashboardAssessmentScoringController.GetTeacherStudents)
		dashboardRouter.GET("/exam-list", dashboardListEntityController.GetExams)
		dashboardRouter.GET("/homework-list", dashboardListEntityController.GetHomeworks)
		dashboardRouter.GET("/exercise-list", dashboardListEntityController.GetExercises)
		dashboardRouter.GET("/school-list", dashboardListEntityController.GetSchools)
		dashboardRouter.GET("/program-list", dashboardListEntityController.GetPrograms)
		dashboardRouter.GET("/course-list", dashboardListEntityController.GetCourses)
		dashboardRouter.GET("/lesson-list", dashboardListEntityController.GetLessons)
		dashboardRouter.GET("/teacher-list", dashboardListEntityController.GetTeachers)
		dashboardRouter.GET("/subject-list", dashboardListEntityController.GetSubjects)

		// Dashboard Teacher Exam routes
		dashboardRouter.GET("/teacher/exam-scored", NewDashboardTeacherExamController().DashboardTeacherExamScored)
		dashboardRouter.GET("/teacher/exam-unscored", NewDashboardTeacherExamUnscoredController().DashboardTeacherExamUnscored)
		dashboardRouter.GET("/teacher/exam-overview", NewDashboardTeacherExamOverviewController().DashboardTeacherExamOverview)

		// Dashboard Teacher Homework routes — static paths before /:homework_id
		dashboardRouter.GET("/teacher/homework/list-homeworks", controllers.NewDashboardTeacherHomeworkListController().GetStudentHomeworkList)
		dashboardRouter.GET("/teacher/homework-overview", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkOverview)
		dashboardRouter.GET("/teacher/homework-overview-grade", controllers.NewDashboardTeacherHomeworkOverviewController().DashboardTeacherHomeworkOverview)
		dashboardRouter.GET("/teacher/homework-unscored", NewDashboardTeacherHomeworkUnscoredController().DashboardTeacherHomeworkUnscored)
		dashboardRouter.GET("/teacher/homework-scored", NewDashboardTeacherHomeworkScoredController().DashboardTeacherHomeworkScored)
		dashboardRouter.GET("/teacher/homework/:homework_id", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkStudent)
		dashboardRouter.GET("/teacher/homework", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkStats)

		// Dashboard Teacher Exercise routes
		dashboardRouter.GET("/teacher/exercise/list-exercises", controllers.NewDashboardTeacherExerciseListController().GetStudentExerciseList)
		dashboardRouter.GET("/teacher/exercise", controllers.NewDashboardTeacherExerciseStudentController().GetStudentStats)

		// Dashboard Exam Ranking routes
		dashboardRouter.GET("/exam/ranking", NewDashboardExamRankingController().GetExamRanking)

		// Dashboard Contest Ranking routes (TODO: Implement controllers)
		// dashboardRouter.GET("/contest/ranking", NewDashboardContestRankingController().GetContestRanking)

		// Dashboard Contest routes (TODO: Implement controllers)
		// dashboardRouter.GET("/contests/stats", dashboardContestController.GetContestStats)
		// dashboardRouter.GET("/contests/list", dashboardContestController.GetContestList)
		// dashboardRouter.GET("/contest-round/:id/stats", dashboardContestController.GetContestRoundStats)
		// dashboardRouter.GET("/contest-round/:id/list", dashboardContestController.GetContestRoundList)
	}

	homeworkController := controllers.NewHomeworkController(services.NewHomeworkService(repositories.NewHomeworkRepository()))
	router.POST("/api/homework/update-all-total-questions", homeworkController.UpdateAllHomeworkTotalQuestions)

	// AI Grading Routes
	aiGradingController := NewAIGradingController()
	managementRouter.POST("/ai/grading/writing", aiGradingController.GradeWriting)
	managementRouter.POST("/ai/grading/speaking", aiGradingController.GradeSpeaking)

	// Internal command routes
	internalRouter := api.Group("/internal/command")
	internalRouter.Use(middleware.AuthMiddleware(authRepo))
	internalRouter.Use(middleware.RoleMiddleware("internal.command"))
	{
		copyHomeworkToExerciseCmd := command.NewCopyHomeworkToExerciseCommand()
		internalRouter.POST("/copy-homework-to-exercise", copyHomeworkToExerciseCmd.Execute)
		clearExerciseCmd := command.NewClearExerciseCommand()
		internalRouter.POST("/clear-exercise", clearExerciseCmd.Execute)

		// Update homework lesson IDs
		updateHomeworkLessonIdsCmd := command.NewUpdateHomeworkLessonIdsCommand()
		internalRouter.POST("/update-homework-lesson-ids", updateHomeworkLessonIdsCmd.Execute)

		// Homework status scoring sync
		homeworkStatusScoringController := controllers.NewHomeworkStatusScoringController()
		internalRouter.POST("/sync-homework-status-scoring", homeworkStatusScoringController.SyncHomeworkStatusScoring)
	}

	// Internal command routes
	internalController := controllers.NewInternalCommandController()
	internal := api.Group("/internal")
	internal.Use(middleware.AuthMiddleware(authRepo))
	internal.Use(middleware.RoleMiddleware("internal.command"))
	{
		// Routes cho school statistics
		internal.POST("/school-statistics/generate", internalController.GenerateSchoolStatistics)
		internal.POST("/school-statistics/generate-last-week", internalController.GenerateSchoolStatisticsLastWeek)
		internal.GET("/school-statistics", internalController.GetSchoolStatistics)

		// Routes cho course statistics
		internal.POST("/course-statistics/generate", internalController.GenerateCourseStatistics)
		internal.POST("/course-statistics/generate-last-week", internalController.GenerateCourseStatisticsLastWeek)
		internal.GET("/course-statistics", internalController.GetCourseStatistics)

		// Routes cho daily statistics
		internal.POST("/daily/school-statistics", internalController.RunDailySchoolStatistics)
		internal.POST("/daily/course-statistics", internalController.RunDailyCourseStatistics)
		internal.POST("/daily/all-statistics", internalController.RunAllDailyStatistics)

		// Route để liệt kê tất cả jobs
		internal.GET("/jobs", internalController.ListAvailableJobs)

		// Routes để xóa dữ liệu
		internal.DELETE("/school-statistics/clear", internalController.ClearDashboardSchoolsData)
		internal.DELETE("/course-statistics/clear", internalController.ClearDashboardCoursesData)
	}
}

// Contest Controller Factory Functions
func NewContestController() *controllers.ContestController {
	contestRepo := repositories.NewContestRepository()
	contestService := services.NewContestService(contestRepo)
	return controllers.NewContestController(contestService)
}

func NewContestRoundController() *controllers.ContestRoundController {
	contestRoundRepo := repositories.NewContestRoundRepository()
	contestRoundService := services.NewContestRoundService(contestRoundRepo)
	return controllers.NewContestRoundController(contestRoundService)
}
