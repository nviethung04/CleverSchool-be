package routes

import (
	"be-lms/command"
	"be-lms/controllers"
	"be-lms/middleware"
	"be-lms/repositories"
	"be-lms/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ====== INIT ROUTES ======
func InitRoutes(router *gin.Engine) {
	// iOS Universal Links - apple-app-site-association (public, no auth)
	appleAASAController := controllers.NewAppleAppSiteAssociationController()
	router.GET("/.well-known/apple-app-site-association", appleAASAController.GetAppleAppSiteAssociation)
	router.GET("/apple-app-site-association", appleAASAController.GetAppleAppSiteAssociation)

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
	RoutePushNotification(router)
	RouteMessage(router)
	RegisterWebsocketRoutes(router)

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
	tusGroup.Any("", gin.WrapH(http.StripPrefix("/api/tus-uploads/", tusHandler)))
	tusGroup.Any("/*any", gin.WrapH(http.StripPrefix("/api/tus-uploads/", tusHandler)))

	api := router.Group("/api")
	authRepo := repositories.NewAuthRepository()

	publicMeetingJoinController := controllers.NewPublicMeetingJoinController(authRepo)
	api.GET("/meet/:code", publicMeetingJoinController.Join)
	// Timeout 30s for all /api/* routes
	api.Use(middleware.TimeoutWithSkip(30*time.Second, []string{
		"/api/manage/course-schedule/sync-all-family",
	}))

	// Public config endpoint (no auth required)
	appConfigController := NewAppConfigController()
	api.GET("/config", appConfigController.GetConfig)

	// iOS Universal Links - bản dưới /api để tránh 404 khi có reverse proxy (path không dấu chấm)
	api.GET("/well-known/apple-app-site-association", appleAASAController.GetAppleAppSiteAssociation)

	//api.POST("/refresh", authController.Register)

	userController := NewUserController()
	api.GET("/student-parent/:parent_id", userController.GetStudentsByParent)

	managementRouter := api.Group("/manage")
	managementRouter.Use(middleware.AuthMiddleware(authRepo))

	// OAuth callback routes - no auth required (called by OAuth providers)
	zoomAuthController := NewZoomAuthController()
	googleAuthController := NewGoogleAuthController()
	microsoftAuthController := NewMicrosoftAuthController()
	api.GET("/manage/zoom/auth/callback", zoomAuthController.Callback)
	api.GET("/manage/google/auth/callback", googleAuthController.Callback)
	api.GET("/manage/microsoft/auth/callback", microsoftAuthController.HandleCallback)
	{

		classController := NewClassController()
		questionController := NewQuestionController()
		lessonPlanController := NewLessonPlanController()
		lessonController := NewLessonController()
		lessonPlanPartController := NewLessonPlanPartController()
		examController := NewExamController()
		exerciseController := NewExerciseController()
		homeworkController := NewHomeworkController()
		assessmentController := NewAssessmentController()
		assessmentCriterionController := NewAssessmentCriterionController()
		assessmentSubcriterionController := NewAssessmentSubcriterionController()
		assessmentCriteriaGroupController := NewAssessmentCriteriaGroupController()
		programController := NewProgramController()
		courseController := NewCourseController()
		courseFamilyController := controllers.NewCourseFamilyController()
		courseScheduleController := NewCourseScheduleController()
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
		noticeController := NewNoticeController()

		studyShiftController := NewStudyShiftController()
		weekController := NewWeekController()
		lessonScheduleController := NewLessonScheduleController()

		gradeController := NewGradeController()
		facultyController := NewFacultyController()

		studyReportCriteriaController := NewStudyReportCriteriaController()
		studyReportController := NewStudyReportController()

		// Semester controller
		semesterController := NewSemesterController()
		holidayController := NewHolidayController()

		// Contest controllers
		contestController := NewContestController()
		contestRoundController := NewContestRoundController()
		contestQuestionController := NewContestQuestionController()
		contestResultController := NewContestResultController()

		settingController := NewSettingController()
		headingController := NewHeadingController()
		trainingLevelController := NewTrainingLevelController()

		teachingPlanController := NewTeachingPlanController()

		RegisterModuleRoute(managementRouter, "questions", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, questionController)
		RegisterModuleRoute(managementRouter, "classes", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, classController)
		RegisterModuleRoute(managementRouter, "lessons", []string{"index", "show", "store", "update", "destroy", "restore"}, lessonController)
		RegisterModuleRoute(managementRouter, "lesson-plans", []string{"index", "show", "store", "update", "destroy"}, lessonPlanController)
		RegisterModuleRoute(managementRouter, "lesson-plan-parts", []string{"index", "show", "store", "update", "destroy"}, lessonPlanPartController)
		RegisterModuleRoute(managementRouter, "exams", []string{"index", "show", "store", "update", "destroy"}, examController)
		RegisterModuleRoute(managementRouter, "exercises", []string{"index", "show", "store", "update", "destroy"}, exerciseController)
		RegisterModuleRoute(managementRouter, "homeworks", []string{"index", "show", "store", "update", "destroy"}, homeworkController)
		RegisterModuleRoute(managementRouter, "assessment-criteria", []string{"index", "show", "store", "update", "destroy"}, assessmentCriterionController)
		managementRouter.POST("/assessment-criteria/bulk", middleware.RoleMiddleware("assessment-criteria.store"), assessmentCriterionController.CreateBulk)
		RegisterModuleRoute(managementRouter, "assessment-subcriteria", []string{"index", "show", "store", "update", "destroy"}, assessmentSubcriterionController)
		RegisterModuleRoute(managementRouter, "assessment-criteria-groups", []string{"index", "show", "store", "update", "destroy"}, assessmentCriteriaGroupController)
		managementRouter.POST("/assessment-criteria-group/with-criteria", middleware.RoleMiddleware("assessment-criteria-groups.store"), assessmentCriteriaGroupController.CreateWithCriteria)
		managementRouter.PUT("/assessment-criteria-group/with-criteria/:id", middleware.RoleMiddleware("assessment-criteria-groups.update"), assessmentCriteriaGroupController.UpdateWithCriteria)

		managementRouter.POST("/assessments/ref-lesson", middleware.RoleMiddleware("assessments.store"), assessmentController.AssignRefLesson)
		managementRouter.POST("/assessments/with-criteria", middleware.RoleMiddleware("assessments.store"), assessmentController.CreateWithCriteria)
		managementRouter.POST("/assessments/save-score/bulk", middleware.RoleMiddleware("assessments.store"), assessmentController.SaveScoreBulk)
		assessmentExportController := NewAssessmentExportController()
		managementRouter.GET("/assessments/export-excel", middleware.RoleMiddleware("assessments.index"), assessmentExportController.ExportExcel)
		managementRouter.POST("/assessments/import-excel", middleware.RoleMiddleware("assessments.store"), assessmentExportController.ImportExcel)
		RegisterModuleRoute(managementRouter, "assessments", []string{"index", "show", "store", "update", "destroy"}, assessmentController)

		// Routes accessible without /manage prefix but still require auth/role
		assessmentsRouter := api.Group("/assessments")
		assessmentsRouter.Use(middleware.AuthMiddleware(authRepo))
		assessmentsRouter.POST("/save-score/bulk", middleware.RoleMiddleware("assessments.store"), assessmentController.SaveScoreBulk)
		RegisterModuleRoute(managementRouter, "contests", []string{"index", "show", "store", "update", "destroy", "restore"}, contestController)
		RegisterModuleRoute(managementRouter, "contest_rounds", []string{"index", "show", "store", "update", "destroy", "restore"}, contestRoundController)
		managementRouter.GET("/publish-assessments", middleware.RoleMiddleware("assessments.index"), assessmentController.GetPublish)
		managementRouter.PUT("/publish-assessments", middleware.RoleMiddleware("assessments.store"), assessmentController.UpdatePublish)
		// Contest custom routes
		managementRouter.GET("/contests/:id/rounds", contestController.GetContestRounds)
		managementRouter.GET("/contests/:id/with-rounds", contestController.GetContestWithRounds)

		// Contest Round custom routes
		managementRouter.GET("/contests/:id/contest_rounds", contestRoundController.GetByContestId)
		managementRouter.GET("/contest_rounds/:id/users", contestRoundController.GetContestRoundUsers)
		managementRouter.GET("/contest_rounds/:id/joiners", contestRoundController.GetContestRoundJoiners)
		managementRouter.POST("/contest_rounds/:id/joiners", contestRoundController.AddJoiner)
		managementRouter.DELETE("/contest_rounds/:id/joiners", contestRoundController.RemoveJoiner)
		managementRouter.DELETE("/contest_rounds/:id/joiners/bulk", contestRoundController.BulkRemoveJoiners)

		// Contest Round joiner list routes (deprecated - use /joiners instead)
		managementRouter.GET("/contest_rounds/:id/provinces", contestRoundController.GetContestRoundProvinces)
		managementRouter.GET("/contest_rounds/:id/schools", contestRoundController.GetContestRoundSchools)
		managementRouter.GET("/contest_rounds/:id/classes", contestRoundController.GetContestRoundClasses)
		managementRouter.GET("/contest_rounds/:id/students", contestRoundController.GetContestRoundStudents)

		// Contest Questions routes for Admin
		managementRouter.GET("/contest_rounds/:id/questions", middleware.RoleMiddleware("contest_rounds.show"), contestQuestionController.GetContestRoundQuestions)
		managementRouter.POST("/contest_rounds/:id/questions", middleware.RoleMiddleware("contest_rounds.update"), contestQuestionController.AssignQuestionsToContestRound)
		managementRouter.DELETE("/contest_rounds/:id/questions", middleware.RoleMiddleware("contest_rounds.update"), contestQuestionController.RemoveQuestionsFromContestRound)

		// Contest Results routes for Admin
		managementRouter.GET("/contest_rounds/:id/answers", middleware.RoleMiddleware("contest_rounds.show"), contestResultController.GetContestRoundAnswers)
		managementRouter.GET("/contest_rounds/:id/answers/student", middleware.RoleMiddleware("contest_rounds.show"), contestResultController.GetContestRoundAnswersByStudent)
		managementRouter.GET("/contest_rounds/:id/results", middleware.RoleMiddleware("contest_rounds.show"), contestResultController.GetContestRoundResults)

		RegisterModuleRoute(managementRouter, "programs", []string{"index", "show", "store", "update", "destroy", "restore"}, programController)
		managementRouter.GET("/programs/:id/export", middleware.RoleMiddleware("programs.export"), programController.Export)
		managementRouter.POST("/programs/import", middleware.RoleMiddleware("programs.import"), programController.Import)

		RegisterModuleRoute(managementRouter, "courses", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, courseController)
		managementRouter.GET("/courses/:id/export-users", middleware.RoleMiddleware("courses.export"), courseController.ExportUsers)
		managementRouter.POST("/courses/:id/import-users", middleware.RoleMiddleware("courses.import"), courseController.ImportUsers)
		managementRouter.GET("/list-courses-family", courseController.ListCoursesFamily)
		managementRouter.POST("/course-schedule/assign-parent", courseScheduleController.AssignParent)
		managementRouter.POST("/course-schedule/sync-all-family", courseScheduleController.SyncFamily)
		managementRouter.GET("/course-schedule/family", courseFamilyController.GetCourseFamily)
		RegisterModuleRoute(managementRouter, "schools", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, schoolController)
		RegisterModuleRoute(managementRouter, "subjects", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, subjectController)
		RegisterModuleRoute(managementRouter, "chapters", []string{"index", "show", "store", "update", "destroy", "restore"}, chapterController)
		RegisterModuleRoute(managementRouter, "headings", []string{"index", "show", "store", "update", "destroy"}, headingController)
		RegisterModuleRoute(managementRouter, "training-levels", []string{"index", "show", "store", "update", "destroy"}, trainingLevelController)
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

		// Faculty
		RegisterModuleRoute(managementRouter, "faculties", []string{"index", "show", "store", "update", "destroy", "restore", "export", "import"}, facultyController)

		RegisterModuleRoute(managementRouter, "study-report-criterias", []string{"index", "show", "store", "update", "destroy"}, studyReportCriteriaController)
		RegisterModuleRoute(managementRouter, "study-reports", []string{"index", "show", "store", "update", "destroy"}, studyReportController)
		managementRouter.GET("/study-reports/evaluates/:course_id", middleware.RoleMiddleware("study-reports.index"), studyReportController.Evaluate)
		managementRouter.GET("/publish-study-report", middleware.RoleMiddleware("study-reports.index"), studyReportController.GetPublish)
		managementRouter.PUT("/publish-study-report", middleware.RoleMiddleware("study-reports.store"), studyReportController.UpdatePublish)
		RegisterModuleRoute(managementRouter, "question-attributes", []string{"index", "show", "store", "update", "destroy", "restore"}, questionAttributeController)
		managementRouter.GET("/question-attributes/parent", middleware.RoleMiddleware("question-attributes.index"), questionAttributeController.GetParents)
		RegisterModuleRoute(managementRouter, "study-shifts", []string{"index", "show", "store", "update", "destroy", "restore"}, studyShiftController)
		RegisterModuleRoute(managementRouter, "holidays", []string{"index", "show", "store", "update", "destroy"}, holidayController)
		RegisterModuleRoute(managementRouter, "semesters", []string{"index", "show", "store", "update", "destroy"}, semesterController)

		// Notice routes
		RegisterModuleRoute(managementRouter, "notices", []string{"index", "show", "store", "update", "destroy"}, noticeController)

		RegisterModuleRoute(managementRouter, "settings", []string{"index", "show", "store", "update", "destroy"}, settingController)
		api.GET("/manage/settings/by-key/:key", settingController.GetByKey)
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

		// Zoom Integration Routes
		zoomAuthController := NewZoomAuthController()
		zoomMeetingController := NewZoomMeetingController()

		// Zoom Auth routes
		managementRouter.POST("/zoom/auth/connect", middleware.AuthMiddleware(authRepo), zoomAuthController.Connect)
		managementRouter.POST("/zoom/auth/disconnect", middleware.AuthMiddleware(authRepo), zoomAuthController.Disconnect)
		managementRouter.GET("/zoom/auth/status", middleware.AuthMiddleware(authRepo), zoomAuthController.Status)
		managementRouter.POST("/zoom/auth/refresh", middleware.AuthMiddleware(authRepo), zoomAuthController.RefreshToken)

		// Zoom Meeting routes
		managementRouter.POST("/zoom/meetings", middleware.AuthMiddleware(authRepo), zoomMeetingController.CreateMeeting)
		managementRouter.GET("/zoom/meetings/:id", middleware.AuthMiddleware(authRepo), zoomMeetingController.GetMeeting)
		managementRouter.GET("/zoom/meetings", middleware.AuthMiddleware(authRepo), zoomMeetingController.GetUserMeetings)
		managementRouter.PUT("/zoom/meetings/:id", middleware.AuthMiddleware(authRepo), zoomMeetingController.UpdateMeeting)
		managementRouter.DELETE("/zoom/meetings/:id", middleware.AuthMiddleware(authRepo), zoomMeetingController.DeleteMeeting)
		managementRouter.POST("/zoom/meetings/:id/start", middleware.AuthMiddleware(authRepo), zoomMeetingController.StartMeeting)

		// Course and Lesson integration routes
		managementRouter.GET("/courses/:id/zoom-meetings", middleware.AuthMiddleware(authRepo), zoomMeetingController.GetCourseMeetings)
		managementRouter.POST("/courses/:id/zoom-meeting", middleware.AuthMiddleware(authRepo), zoomMeetingController.CreateCourseMeeting)
		managementRouter.GET("/lessons/:id/zoom-meetings", middleware.AuthMiddleware(authRepo), zoomMeetingController.GetLessonMeetings)
		managementRouter.POST("/lessons/:id/zoom-meeting", middleware.AuthMiddleware(authRepo), zoomMeetingController.CreateLessonMeeting)

		// Google Meet Integration Routes
		googleAuthController := NewGoogleAuthController()
		googleMeetingController := NewGoogleMeetingController()

		// Google Auth routes
		managementRouter.POST("/google/auth/connect", middleware.AuthMiddleware(authRepo), googleAuthController.Connect)
		managementRouter.POST("/google/auth/disconnect", middleware.AuthMiddleware(authRepo), googleAuthController.Disconnect)
		managementRouter.GET("/google/auth/status", middleware.AuthMiddleware(authRepo), googleAuthController.Status)
		managementRouter.POST("/google/auth/refresh", middleware.AuthMiddleware(authRepo), googleAuthController.RefreshToken)

		// Google Meeting routes
		managementRouter.POST("/google/meetings", middleware.AuthMiddleware(authRepo), middleware.RoleMiddleware("meetings.store"), googleMeetingController.CreateMeeting)
		managementRouter.GET("/google/meetings/:id", middleware.AuthMiddleware(authRepo), googleMeetingController.GetMeeting)
		managementRouter.GET("/google/meetings", middleware.AuthMiddleware(authRepo), googleMeetingController.GetUserMeetings)
		managementRouter.PUT("/google/meetings/:id", middleware.AuthMiddleware(authRepo), googleMeetingController.UpdateMeeting)
		managementRouter.DELETE("/google/meetings/:id", middleware.AuthMiddleware(authRepo), googleMeetingController.DeleteMeeting)
		managementRouter.DELETE("/google/meetings/:id/force", middleware.AuthMiddleware(authRepo), googleMeetingController.ForceDeleteMeeting)
		managementRouter.POST("/google/meetings/:id/recording", middleware.AuthMiddleware(authRepo), googleMeetingController.UpdateRecording)
		managementRouter.GET("/google/meetings/upcoming", middleware.AuthMiddleware(authRepo), googleMeetingController.GetUpcomingMeetings)

		// Google Meet Course and Lesson integration routes
		managementRouter.GET("/google-courses/:id/meetings", middleware.AuthMiddleware(authRepo), googleMeetingController.GetCourseMeetings)
		managementRouter.POST("/google-courses/:id/meeting", middleware.AuthMiddleware(authRepo), middleware.RoleMiddleware("meetings.store"), googleMeetingController.CreateCourseIntegration)
		managementRouter.GET("/google-lessons/:id/meetings", middleware.AuthMiddleware(authRepo), googleMeetingController.GetLessonMeetings)
		managementRouter.POST("/google-lessons/:id/meeting", middleware.AuthMiddleware(authRepo), middleware.RoleMiddleware("meetings.store"), googleMeetingController.CreateLessonIntegration)

		// Google Meeting Attendance routes
		googleMeetingAttendanceController := controllers.NewGoogleMeetingAttendanceController()
		managementRouter.POST("/google/meetings/join/:code", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.JoinByShortCode)
		managementRouter.POST("/google/meetings/:id/join", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.JoinMeeting)
		managementRouter.POST("/google/meetings/:id/leave", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.LeaveMeeting)
		managementRouter.GET("/google/meetings/:id/attendances", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.GetMeetingAttendances)
		managementRouter.GET("/google/meetings/:id/attendances/export", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.ExportAttendancesExcel)
		managementRouter.GET("/google/meetings/:id/summary", middleware.AuthMiddleware(authRepo), googleMeetingAttendanceController.GetAttendanceSummary)

		microsoftAuthController := NewMicrosoftAuthController()
		microsoftMeetingController := NewMicrosoftMeetingController()

		managementRouter.POST("/microsoft/auth/connect", middleware.AuthMiddleware(authRepo), microsoftAuthController.ConnectAccount)
		managementRouter.GET("/microsoft/auth/status", middleware.AuthMiddleware(authRepo), microsoftAuthController.CheckConnection)
		managementRouter.POST("/microsoft/auth/disconnect", middleware.AuthMiddleware(authRepo), microsoftAuthController.DisconnectAccount)

		managementRouter.GET("/microsoft/meetings/permissions", middleware.AuthMiddleware(authRepo), microsoftMeetingController.CheckMeetingPermissions)
		managementRouter.POST("/microsoft/meetings", middleware.AuthMiddleware(authRepo), middleware.RoleMiddleware("meetings.store"), microsoftMeetingController.CreateMeeting)
		managementRouter.GET("/microsoft/meetings/:id", middleware.AuthMiddleware(authRepo), microsoftMeetingController.GetMeeting)
		managementRouter.PUT("/microsoft/meetings/:id", middleware.AuthMiddleware(authRepo), microsoftMeetingController.UpdateMeeting)
		managementRouter.DELETE("/microsoft/meetings/:id", middleware.AuthMiddleware(authRepo), microsoftMeetingController.DeleteMicrosoftMeeting)
		managementRouter.POST("/microsoft/meetings/:id/share-link", middleware.AuthMiddleware(authRepo), microsoftMeetingController.SetRecordingShareURL)
		managementRouter.GET("/microsoft/meetings/user", middleware.AuthMiddleware(authRepo), microsoftMeetingController.GetUserMeetings)
		managementRouter.GET("/microsoft/meetings/upcoming", middleware.AuthMiddleware(authRepo), microsoftMeetingController.GetUpcomingMeetings)
		managementRouter.GET("/microsoft/meetings/course/:course_id", middleware.AuthMiddleware(authRepo), microsoftMeetingController.GetCourseMeetings)
		managementRouter.GET("/microsoft/meetings/lesson/:lesson_id", middleware.AuthMiddleware(authRepo), microsoftMeetingController.GetLessonMeetings)
		managementRouter.POST("/microsoft/lessons/:lesson_id/meeting", middleware.AuthMiddleware(authRepo), middleware.RoleMiddleware("meetings.store"), microsoftMeetingController.CreateLessonIntegration)
		managementRouter.POST("/microsoft/meetings/:id/recording", middleware.AuthMiddleware(authRepo), microsoftMeetingController.FetchRecording)
		managementRouter.GET("/microsoft/meetings/:id/recording", middleware.AuthMiddleware(authRepo), microsoftMeetingController.FetchRecording)
		// Meeting Attendance routes (for online class)
		meetingAttendanceController := controllers.NewMeetingAttendanceController()
		managementRouter.POST("/microsoft/meetings/join/:code", middleware.AuthMiddleware(authRepo), meetingAttendanceController.JoinByShortCode)
		managementRouter.POST("/microsoft/meetings/:id/join", middleware.AuthMiddleware(authRepo), meetingAttendanceController.JoinMeeting)
		managementRouter.POST("/microsoft/meetings/:id/leave", middleware.AuthMiddleware(authRepo), meetingAttendanceController.LeaveMeeting)
		managementRouter.GET("/microsoft/meetings/:id/attendances", middleware.AuthMiddleware(authRepo), meetingAttendanceController.GetMeetingAttendances)
		managementRouter.GET("/microsoft/meetings/:id/attendances/export", middleware.AuthMiddleware(authRepo), meetingAttendanceController.ExportAttendancesToExcel)
		managementRouter.GET("/microsoft/meetings/:id/summary", middleware.AuthMiddleware(authRepo), meetingAttendanceController.GetAttendanceSummary)
		managementRouter.GET("/microsoft/meetings/my-attendance", middleware.AuthMiddleware(authRepo), meetingAttendanceController.GetUserAttendanceHistory)

		meetingNotificationController := controllers.NewMeetingNotificationController()
		managementRouter.GET("/microsoft/meeting-notifications", middleware.AuthMiddleware(authRepo), meetingNotificationController.ListMyNotifications)
		managementRouter.PUT("/microsoft/meeting-notifications/:id/read", middleware.AuthMiddleware(authRepo), meetingNotificationController.MarkRead)
		managementRouter.GET("/microsoft/meeting-notifications/unread-count", middleware.AuthMiddleware(authRepo), meetingNotificationController.CountUnread)

		// AI Grading Routes
		// aiGradingController := NewAIGradingController()
		// managementRouter.POST("/ai/grading/writing", aiGradingController.GradeWriting)
		// managementRouter.POST("/ai/grading/speaking", aiGradingController.GradeSpeaking)

		// Clone course
		// TODO: Re-enable these routes when methods are implemented
		// managementRouter.POST("/courses/:id/cloned", middleware.RoleMiddleware("courses.store"), courseController.Cloned)
		// managementRouter.DELETE("/courses/:id/cloned", middleware.RoleMiddleware("courses.store"), courseController.DeleteClone)
		// managementRouter.POST("/courses/:id/publish", middleware.RoleMiddleware("courses.store"), courseController.Publish)
		// managementRouter.POST("/courses/:id/sync", middleware.RoleMiddleware("courses.store"), courseController.Sync)
		// managementRouter.POST("/programs/:id/sync", middleware.RoleMiddleware("programs.store"), programController.Sync)
		// managementRouter.POST("/programs/create-by-course", middleware.RoleMiddleware("programs.store"), programController.CreateByCourse)
		// managementRouter.POST("/programs/:id/resync", middleware.RoleMiddleware("programs.store"), programController.ReSync)

		// Clone data
		managementRouter.POST("/homeworks/:id/cloned", middleware.RoleMiddleware("homeworks.store"), homeworkController.Cloned)
		managementRouter.POST("/exams/:id/cloned", middleware.RoleMiddleware("exams.store"), examController.Cloned)
		managementRouter.POST("/exercises/:id/cloned", middleware.RoleMiddleware("exercises.store"), exerciseController.Cloned)

		// Assigned data
		managementRouter.POST("/homeworks/:id/assigned", middleware.RoleMiddleware("homeworks.store"), homeworkController.Assigned)
		managementRouter.POST("/exams/:id/assigned", middleware.RoleMiddleware("exams.store"), examController.Assigned)
		managementRouter.POST("/exercises/:id/assigned", middleware.RoleMiddleware("exercises.store"), exerciseController.Assigned)

		// Assigned lesson
		managementRouter.GET("/homeworks/:id/assigned-lessons", middleware.RoleMiddleware("homeworks.index"), homeworkController.AssignedLesson)
		managementRouter.GET("/homeworks/students-doing", middleware.RoleMiddleware("homeworks.index"), homeworkController.StudentsDoing)
		managementRouter.GET("/exams/:id/assigned-lessons", middleware.RoleMiddleware("exams.index"), examController.AssignedLesson)
		managementRouter.GET("/exercises/:id/assigned-lessons", middleware.RoleMiddleware("exercises.index"), exerciseController.AssignedLesson)

		managementRouter.PUT("/lesson-plans/:id/complete", middleware.RoleMiddleware("lesson-plans.update"), lessonPlanController.Complete)
		managementRouter.PUT("/chapters/:id/sort-lessons", middleware.RoleMiddleware("chapters.update"), chapterController.SortLessons)
		managementRouter.PUT("/programs/:id/sort-chapters", middleware.RoleMiddleware("programs.update"), programController.SortChapters)
		managementRouter.PUT("/questions/sync-keywords", middleware.RoleMiddleware("questions.update"), questionController.SyncKeywords)

		managementRouter.GET("/weeks", weekController.GetAll)
		managementRouter.GET("/weeks/by-date", weekController.GetWeekByDate)
		managementRouter.GET("/lesson-schedules", lessonScheduleController.GetAll)
		managementRouter.GET("/lessons/schedules", middleware.RoleMiddleware("lessons.show"), lessonController.LessonSchedules)
		managementRouter.POST("/lessons/schedules", middleware.RoleMiddleware("lessons.update"), lessonController.StoreLessonSchedules)

		// Lesson Schedule Copy
		lessonScheduleCopyController := NewLessonScheduleCopyController()
		managementRouter.POST("/lesson-schedules/copy", middleware.RoleMiddleware("lesson_schedule.store"), lessonScheduleCopyController.CopyLessonSchedules)
		managementRouter.POST("/lesson-schedules/sync-all-program", middleware.RoleMiddleware("lesson_schedule.store"), lessonScheduleController.SyncAllProgram)

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

		managementRouter.GET("/classes/:id/users", middleware.RoleMiddleware("classes.show"), classController.GetUsers)
		managementRouter.POST("/classes/:id/users", middleware.RoleMiddleware("classes.update"), classController.StoreUsers)
		managementRouter.PUT("/classes/:id/users", middleware.RoleMiddleware("classes.update"), classController.AddUsers)
		schoolDashboardController := NewSchoolDashboardController()
		managementRouter.GET("/school-dashboard-starts", schoolDashboardController.GetSchoolSummary)

		managementRouter.PUT("/lessons/:id/completion", middleware.RoleMiddleware("lessons.show"), lessonController.Completion)
		managementRouter.PUT("/lessons/:id/studying", middleware.RoleMiddleware("lessons.show"), lessonController.Studying)

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
		managementRouter.PUT("/vocabulary-progress/:vocabularyId", flashcardController.UpdateVocabularyProgress)

		RegisterModuleRoute(managementRouter, "teaching-plans", []string{"index", "show", "store", "update", "destroy"}, teachingPlanController)
		managementRouter.PUT("/teaching-plan-approve", middleware.RoleMiddleware("teaching-plans.approve"), teachingPlanController.Approve)
	}

	// Feedback routes
	feedbackController := NewFeedbackController()
	managementRouter.GET("/feedbacks", feedbackController.GetAll)
	managementRouter.GET("/feedbacks/:id", feedbackController.GetByID)
	managementRouter.POST("/feedbacks", feedbackController.Create)
	managementRouter.PUT("/feedbacks/:id", feedbackController.Update)
	managementRouter.PATCH("/feedbacks/:id/status", feedbackController.UpdateStatus)
	managementRouter.DELETE("/feedbacks/:id", feedbackController.Delete)

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
		studyRouter.POST("/save-score/evaluate", saveScoreController.Evaluate)
		studyRouter.POST("/save-score/teacher-evaluate", saveScoreController.TeacherEvaluate)
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
		// exercise/homework comment
		exerciseCommentCtl := NewExerciseCommentController()
		homeworkCommentCtl := NewHomeworkCommentController()
		studyRouter.POST("/exercise-comment", exerciseCommentCtl.PostExerciseComment)
		studyRouter.POST("/homework-comment", homeworkCommentCtl.PostHomeworkComment)
		studyRouter.GET("/exercise-students", exerciseStudentController.GetExerciseStudents)

		examCourseController := NewExamCourseController()
		studyRouter.GET("/exam-courses", examCourseController.GetExamCourseDetail)

		homeworkStudentController := NewHomeworkStudentController()
		studyRouter.GET("/homework-students", homeworkStudentController.GetHomeworkStudents)
		studyRouter.GET("/teacher/homework-answers", homeworkAnswerController.GetTeacherHomeworkAnswer)
		studyRouter.GET("/student/homework-answers", homeworkAnswerController.GetStudentHomeworkAnswer)

		examCommentController := NewExamCommentController()
		studyRouter.POST("/exam-comment", examCommentController.PostExamComment)

		// Contest student routes
		contestStudentController := NewContestStudentController()
		studyRouter.GET("/contest-rounds-by-student", contestStudentController.GetContestRoundsByStudent)
		studyRouter.GET("/contest_rounds_by_student", contestStudentController.GetContestRoundsByStudent) // Backward compatibility

		// Contest Questions routes for Students
		contestQuestionController := NewContestQuestionController()
		studyRouter.GET("/contest_rounds/:id/questions", contestQuestionController.GetContestRoundQuestions)
	}

	dashboardRouter := api.Group("/dashboard")
	dashboardRouter.Use(middleware.AuthMiddleware(authRepo))
	{
		dashboardStudentExamController := NewDashboardStudentExamController()
		dashboardStudentHomeworkController := NewDashboardStudentHomeworkController()
		dashboardStudentExamListController := NewDashboardStudentExamListController()
		dashboardStudentHomeworkListController := NewDashboardStudentHomeworkListController()
		dashboardListEntityController := NewDashboardListEntityController()
		dashboardRouter.GET("/student/exam", middleware.AuthMiddleware(authRepo), dashboardStudentExamController.GetStudentExamStats)
		dashboardRouter.GET("/student/homework", middleware.AuthMiddleware(authRepo), dashboardStudentHomeworkController.GetStudentHomeworkStats)
		dashboardRouter.GET("/student/exam-list", dashboardStudentExamListController.GetStudentExamList)
		dashboardRouter.GET("/student/homework-list", dashboardStudentHomeworkListController.GetStudentHomeworkList)
		dashboardStudentAssessmentController := NewDashboardStudentAssessmentController()
		dashboardRouter.GET("/student/assessments", dashboardStudentAssessmentController.GetStudentAssessments)
		dashboardAssessmentReportExcelController := NewDashboardAssessmentReportExcelController()
		dashboardRouter.POST("/assessment/report-excel", dashboardAssessmentReportExcelController.ReportExcel)
		dashboardRouter.GET("/exam-list", dashboardListEntityController.GetExams)
		dashboardRouter.GET("/homework-list", dashboardListEntityController.GetHomeworks)
		dashboardRouter.GET("/school-list", dashboardListEntityController.GetSchools)
		dashboardRouter.GET("/course-list", dashboardListEntityController.GetCourses)
		dashboardRouter.GET("/class-list", dashboardListEntityController.GetClasses)
		dashboardRouter.GET("/class-main-list", dashboardListEntityController.GetClassMains)
		dashboardRouter.GET("/chapter-list", dashboardListEntityController.GetChapters)
		dashboardRouter.GET("/lesson-list", dashboardListEntityController.GetLessons)
		dashboardRouter.GET("/teacher-list", dashboardListEntityController.GetTeachers)
		dashboardRouter.GET("/subject-list", dashboardListEntityController.GetSubjects)

		// Dashboard Teacher Exam routes
		dashboardRouter.GET("/teacher/exam-scored", NewDashboardTeacherExamController().DashboardTeacherExamScored)
		dashboardRouter.GET("/teacher/exam-unscored", NewDashboardTeacherExamUnscoredController().DashboardTeacherExamUnscored)
		dashboardRouter.GET("/teacher/exam-overview", NewDashboardTeacherExamOverviewController().DashboardTeacherExamOverview)

		// Dashboard Teacher Homework routes
		dashboardRouter.GET("/teacher/homework/list-homeworks", NewDashboardTeacherHomeworkListController().GetHomeworkList)
		dashboardRouter.GET("/teacher/homework/export", NewDashboardTeacherHomeworkStudentController().ExportTeacherHomeworkStats)
		dashboardRouter.GET("/teacher/homework/:homework_id/export", NewDashboardTeacherHomeworkStudentController().ExportTeacherHomeworkStudent)
		dashboardRouter.GET("/teacher/homework/:homework_id", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkStudent)
		dashboardRouter.GET("/teacher/homework", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkStats)
		dashboardRouter.GET("/teacher/homework-overview", NewDashboardTeacherHomeworkStudentController().DashboardTeacherHomeworkOverview)
		dashboardRouter.GET("/teacher/homework-overview-grade", controllers.NewDashboardTeacherHomeworkOverviewController().DashboardTeacherHomeworkOverview)
		dashboardRouter.GET("/teacher/homework-unscored", NewDashboardTeacherHomeworkUnscoredController().DashboardTeacherHomeworkUnscored)
		dashboardRouter.GET("/teacher/homework-scored", NewDashboardTeacherHomeworkScoredController().DashboardTeacherHomeworkScored)

		// Dashboard Teacher Assessment routes
		dashboardRouter.GET("/teacher/assessment/students", NewAssessmentStudentController().GetStudentsWithAssessment)

		// Dashboard Assessment Report routes
		dashboardRouter.GET("/assessment/report", controllers.NewDashboardAssessmentReportController().GetAssessmentReport)

		// Dashboard Exam Ranking routes
		dashboardRouter.GET("/exam/ranking", NewDashboardExamRankingController().GetExamRanking)

		// Dashboard Contest Ranking routes (TODO: Implement controllers)
		// dashboardRouter.GET("/contest/ranking", NewDashboardContestRankingController().GetContestRanking)

		// Homework Ranking routes
		dashboardRouter.GET("/ranking/homework", controllers.NewHomeworkRankingController().GetHomeworkRanking)
		// Assessment Ranking routes
		dashboardRouter.GET("/ranking/assessment", controllers.NewAssessmentRankingController().GetAssessmentRanking)

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
	managementRouter.POST("/ai/grading/image", aiGradingController.GradeImage)
	internalController := controllers.NewInternalCommandController()
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

		// Recalculate total questions
		recalculateTotalQuestionsController := NewRecalculateTotalQuestionsController()
		internalRouter.POST("/recalculate-homework-total-questions", recalculateTotalQuestionsController.RecalculateHomeworkTotalQuestions)
		internalRouter.POST("/recalculate-exam-total-questions", recalculateTotalQuestionsController.RecalculateExamTotalQuestions)
		internalRouter.POST("/recalculate-exercise-total-questions", recalculateTotalQuestionsController.RecalculateExerciseTotalQuestions)
		internalRouter.POST("/recalculate-all-total-questions", recalculateTotalQuestionsController.RecalculateAllTotalQuestions)

		// Recalculate homework users data
		recalculateHomeworkUsersController := NewRecalculateHomeworkUsersController()
		internalRouter.POST("/recalculate-homework-users-data", recalculateHomeworkUsersController.RecalculateHomeworkUsersData)

		// Recalculate study reports star
		recalculateStudyReportsStarController := controllers.NewRecalculateStudyReportsStarController()
		internalRouter.POST("/recalculate-study-reports-star", recalculateStudyReportsStarController.RecalculateStudyReportsStar)

		internalRouter.POST("/sync-homework-user-questions", internalController.SyncHomeworkUserQuestions)

		// Recalculate homework users metrics (full logic như submit homework)
		internalRouter.POST("/recalculate-homework-users-metrics", internalController.RecalculateHomeworkUsersMetrics)

		// Sync teacher classes
		internalRouter.POST("/sync-teacher-classes", internalController.SyncTeacherClasses)
	}

	// Internal command routes
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
		internal.POST("/daily/all-statistic-courses", internalController.RunAllDailyCourseStatistics)
		internal.POST("/daily/all-statistic-schools", internalController.RunAllDailySchoolStatistics)
		internal.POST("/sync-homework-user-questions", internalController.SyncHomeworkUserQuestions)
		// Route để liệt kê tất cả jobs
		internal.GET("/jobs", internalController.ListAvailableJobs)

		// Routes để xóa dữ liệu
		internal.DELETE("/school-statistics/clear", internalController.ClearDashboardSchoolsData)
		internal.DELETE("/course-statistics/clear", internalController.ClearDashboardCoursesData)

		// Routes cho sync class main
		internal.POST("/sync-class-main", internalController.SyncClassMain)

		// Routes cho sync assessment ref lesson
		internal.POST("/sync-assessment-ref-lesson", internalController.SyncAssessmentRefLesson)
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

func NewContestQuestionController() *controllers.ContestQuestionController {
	return controllers.NewContestQuestionController()
}

func NewContestResultController() *controllers.ContestResultController {
	return controllers.NewContestResultController()
}
