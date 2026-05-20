package repositories

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DataEntry struct {
	Id         int64
	CourseId   int64
	LessonId   int64
	LessonName string
	IsAssigned bool
	Date       time.Time
	Name       string
}

type DataNeedScoring struct {
	DataID   int64
	UserID   int64
	IsScored bool
	Date     time.Time
	Count    int32
}

type DataSubmitted struct {
	DataID int64
	UserID int64
	Date   time.Time
}

type DataNotSubmitted struct {
	DataID   int64
	CourseID int64
	UserID   int64
}

type CourseTeacher struct {
	CourseID int64
	UserID   int64
	Name     string
}

func (r *dashboardRepository) GetTeacherPerformance(filter dto.FilterDashboardAdmin) (*dto.TeacherPerformanceOverview, error) {
	var courseIds, userIds []int64
	var homeworkData, examData, exerciseData []DataEntry
	var courseTeachers []CourseTeacher
	type key struct {
		CourseID int64
		DataID   int64
	}
	var homeworkNeedScoring, examNeedScoring, exerciseNeedScoring []DataNeedScoring
	var homeworkSubmitted, examSubmitted, exerciseSubmitted []DataSubmitted
	var homeworkNotSubmitted, examNotSubmitted, exerciseNotSubmitted []DataNotSubmitted
	var homeworkIds, examIds, exerciseIds []int64
	var homeworkIdSet, examIdSet, exerciseIdSet map[int64]struct{}

	db := db.ReplicaDB
	filterByDate := filter.Month != 0 || filter.Year != 0 || filter.Quarter != 0
	hasFilter := filter.ProgramId > 0 || filter.CourseId > 0 || filter.SchoolId > 0 || filter.TeacherId > 0

	includeExam := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exam"
	includeHomework := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "homework"
	includeExercise := filter.ObjectType == "" || filter.ObjectType == "all" || filter.ObjectType == "exercise"

	if filter.TeacherId > 0 {
		db.Table("user_courses").
			Select("DISTINCT course_id").
			Where("user_id = ?", filter.TeacherId).
			Pluck("course_id", &courseIds)
	} else {
		courseQuery := db.Table("courses c").Select("DISTINCT c.id")
		if filter.ProgramId > 0 {
			courseQuery = courseQuery.Where("c.program_id = ?", filter.ProgramId)
		}
		if filter.CourseId > 0 {
			courseQuery = courseQuery.Where("c.id = ?", filter.CourseId)
		}
		if filter.SchoolId > 0 {
			courseQuery = courseQuery.Joins("JOIN course_schools cs ON cs.course_id = c.id").Where("cs.school_id = ?", filter.SchoolId)
		}
		if filter.ProgramId > 0 || filter.CourseId > 0 || filter.SchoolId > 0 {
			if err := courseQuery.Pluck("c.id", &courseIds).Error; err != nil {
				return nil, err
			}
		}
	}

	var mu sync.Mutex
	var firstErr error
	setErr := func(e error) {
		if e == nil {
			return
		}
		mu.Lock()
		if firstErr == nil {
			firstErr = e
		}
		mu.Unlock()
	}

	wg1 := sync.WaitGroup{}
	if hasFilter {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			userIdQuery := db.Table("users u").
				Select("DISTINCT u.id").
				Joins("JOIN user_courses ue ON ue.user_id = u.id").
				Joins("JOIN courses c ON c.id = ue.course_id").
				Joins("JOIN user_ref_roles ur ON ur.user_id = u.id").
				Where("ur.role_id = ?", models.StudentRoleId).
				Where("u.deleted_at IS NULL")
			if filter.SchoolId > 0 {
				userIdQuery = userIdQuery.Where("u.school_id = ?", filter.SchoolId)
			}
			if filter.ProgramId > 0 {
				userIdQuery = userIdQuery.Where("c.program_id = ?", filter.ProgramId)
			}
			if filter.CourseId > 0 {
				userIdQuery = userIdQuery.Where("c.id = ?", filter.CourseId)
			} else if len(courseIds) > 0 {
				userIdQuery = userIdQuery.Where("c.id IN (?)", courseIds)
			}
			setErr(userIdQuery.Pluck("id", &userIds).Error)
		}()
	}
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		if len(courseIds) > 0 {
			setErr(db.Table("users u").
				Select("DISTINCT uc.course_id AS course_id, u.id AS user_id, u.name").
				Joins("JOIN user_courses uc ON uc.user_id = u.id").
				Joins("JOIN user_ref_roles ur ON ur.user_id = u.id").
				Where("ur.role_id = ?", models.TeacherRoleId).
				Where("uc.course_id IN (?)", courseIds).
				Scan(&courseTeachers).Error)
		} else {
			setErr(db.Table("users u").
				Select("DISTINCT uc.course_id AS course_id, u.id AS user_id, u.name").
				Joins("JOIN user_courses uc ON uc.user_id = u.id").
				Joins("JOIN user_ref_roles ur ON ur.user_id = u.id").
				Where("ur.role_id = ?", models.TeacherRoleId).
				Scan(&courseTeachers).Error)
		}
	}()
	courseNameMap := make(map[int64]string)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		var rows []struct {
			Id   int64
			Name string
		}
		query := db.Table("courses").Select("id, name")
		if len(courseIds) > 0 {
			query = query.Where("id IN (?)", courseIds)
		}
		setErr(query.Scan(&rows).Error)
		for _, r := range rows {
			mu.Lock()
			courseNameMap[r.Id] = r.Name
			mu.Unlock()
		}
	}()
	wg1.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	type courseTeacherInfo struct {
		MinUserID int64
		Names     []string
	}
	courseTeacherMap := make(map[int64]courseTeacherInfo)
	examCourseMap := make(map[int64]int64)
	exerciseCourseMap := make(map[int64]int64)
	examMap := make(map[int64]DataEntry)
	exerciseMap := make(map[int64]DataEntry)
	for _, ct := range courseTeachers {
		info := courseTeacherMap[ct.CourseID]
		if info.MinUserID == 0 || ct.UserID < info.MinUserID {
			info.MinUserID = ct.UserID
		}
		info.Names = append(info.Names, ct.Name)
		courseTeacherMap[ct.CourseID] = info
	}

	studentMap := make(map[int64]string)

	wg2 := sync.WaitGroup{}
	if includeHomework {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			homeworkQuery := db.Table("homework_ref_lessons hrl").
				Select(`
				DISTINCT
				hrl.homework_id AS id,
				h.name AS name,
				COALESCE(NULLIF(hrl.course_id, 0), ls.course_id) AS course_id,
				hrl.lesson_id AS lesson_id,
				COALESCE(l.title, '') AS lesson_name,
				(hrl.assigned_at IS NOT NULL) AS is_assigned,
				ls.scheduled_date AS date`).
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = hrl.lesson_id").
				Joins("JOIN homeworks h ON h.id = hrl.homework_id").
				Joins("LEFT JOIN lessons l ON l.id = hrl.lesson_id").
				Joins("LEFT JOIN chapters c ON c.id = l.chapter_id")
			if filterByDate {
				homeworkQuery = homeworkQuery.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			} else {
				homeworkQuery = homeworkQuery.Where("ls.scheduled_date <= NOW()")
			}
			if len(courseIds) > 0 {
				placeholders := strings.Repeat("?,", len(courseIds))
				placeholders = placeholders[:len(placeholders)-1]
				args := make([]interface{}, 0, len(courseIds)*2)
				for _, id := range courseIds {
					args = append(args, id)
				}
				for _, id := range courseIds {
					args = append(args, id)
				}
				homeworkQuery = homeworkQuery.Where(
					fmt.Sprintf("(hrl.course_id IN (%s) OR (hrl.course_id = 0 AND ls.course_id IN (%s)))", placeholders, placeholders),
					args...,
				)
			}
			if filter.ProgramId > 0 {
				homeworkQuery = homeworkQuery.Where("c.program_id = ?", filter.ProgramId)
			}
			if filter.SchoolId > 0 {
				homeworkQuery = homeworkQuery.Where("EXISTS (SELECT 1 FROM course_schools cs WHERE cs.course_id = ls.course_id AND cs.school_id = ?)", filter.SchoolId)
			}
			setErr(homeworkQuery.Scan(&homeworkData).Error)
		}()
	}
	if includeExam {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			examQuery := db.Table("exam_ref_lessons erl").
				Select(`
				DISTINCT
				erl.exam_id AS id,
				e.name AS name,
				COALESCE(NULLIF(erl.course_id, 0), ls.course_id) AS course_id,
				(erl.assigned_at IS NOT NULL) AS is_assigned,
				ls.scheduled_date AS date`).
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = erl.lesson_id").
				Joins("JOIN exams e ON e.id = erl.exam_id").
				Joins("LEFT JOIN lessons l ON l.id = erl.lesson_id").
				Joins("LEFT JOIN chapters c ON c.id = l.chapter_id")
			if filterByDate {
				examQuery = examQuery.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			} else {
				examQuery = examQuery.Where("ls.scheduled_date <= NOW()")
			}
			if len(courseIds) > 0 {
				placeholders := strings.Repeat("?,", len(courseIds))
				placeholders = placeholders[:len(placeholders)-1]
				args := make([]interface{}, 0, len(courseIds)*2)
				for _, id := range courseIds {
					args = append(args, id)
				}
				for _, id := range courseIds {
					args = append(args, id)
				}
				examQuery = examQuery.Where(
					fmt.Sprintf("(erl.course_id IN (%s) OR (erl.course_id = 0 AND ls.course_id IN (%s)))", placeholders, placeholders),
					args...,
				)
			}
			if filter.ProgramId > 0 {
				examQuery = examQuery.Where("c.program_id = ?", filter.ProgramId)
			}
			if filter.SchoolId > 0 {
				examQuery = examQuery.Where("EXISTS (SELECT 1 FROM course_schools cs WHERE cs.course_id = ls.course_id AND cs.school_id = ?)", filter.SchoolId)
			}
			setErr(examQuery.Scan(&examData).Error)
		}()
	}
	if includeExercise {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			exerciseQuery := db.Table("exercise_ref_lessons erl").
				Select(`
				DISTINCT
				erl.exercise_id AS id,
				e.name AS name,
				COALESCE(NULLIF(erl.course_id, 0), ls.course_id) AS course_id,
				(erl.assigned_at IS NOT NULL) AS is_assigned,
				ls.scheduled_date AS date`).
				Joins("JOIN lesson_schedules ls ON ls.lesson_id = erl.lesson_id").
				Joins("JOIN exercises e ON e.id = erl.exercise_id").
				Joins("LEFT JOIN lessons l ON l.id = erl.lesson_id").
				Joins("LEFT JOIN chapters c ON c.id = l.chapter_id")
			if filterByDate {
				exerciseQuery = exerciseQuery.Where("ls.scheduled_date >= ? AND ls.scheduled_date <= ?", filter.StartTime, filter.EndTime)
			} else {
				exerciseQuery = exerciseQuery.Where("ls.scheduled_date <= NOW()")
			}
			if len(courseIds) > 0 {
				placeholders := strings.Repeat("?,", len(courseIds))
				placeholders = placeholders[:len(placeholders)-1]
				args := make([]interface{}, 0, len(courseIds)*2)
				for _, id := range courseIds {
					args = append(args, id)
				}
				for _, id := range courseIds {
					args = append(args, id)
				}
				exerciseQuery = exerciseQuery.Where(
					fmt.Sprintf("(erl.course_id IN (%s) OR (erl.course_id = 0 AND ls.course_id IN (%s)))", placeholders, placeholders),
					args...,
				)
			}
			if filter.ProgramId > 0 {
				exerciseQuery = exerciseQuery.Where("c.program_id = ?", filter.ProgramId)
			}
			if filter.SchoolId > 0 {
				exerciseQuery = exerciseQuery.Where("EXISTS (SELECT 1 FROM course_schools cs WHERE cs.course_id = ls.course_id AND cs.school_id = ?)", filter.SchoolId)
			}
			setErr(exerciseQuery.Scan(&exerciseData).Error)
		}()
	}
	wg2.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	if includeHomework {
		dedupMap := make(map[key]DataEntry)
		for _, h := range homeworkData {
			k := key{CourseID: h.CourseId, DataID: h.Id}
			if existing, ok := dedupMap[k]; ok {
				if existing.IsAssigned || !h.IsAssigned {
					continue
				}
			}
			dedupMap[k] = h
		}
		homeworkData = homeworkData[:0]
		for _, v := range dedupMap {
			homeworkData = append(homeworkData, v)
		}

		homeworkIds = homeworkIds[:0]
		homeworkIdSet = make(map[int64]struct{})
		for _, h := range homeworkData {
			if _, ok := homeworkIdSet[h.Id]; !ok {
				homeworkIdSet[h.Id] = struct{}{}
				homeworkIds = append(homeworkIds, h.Id)
			}
		}

	}

	if includeExam {
		dedupExam := make(map[key]DataEntry)
		for _, h := range examData {
			k := key{CourseID: h.CourseId, DataID: h.Id}
			if existing, ok := dedupExam[k]; ok {
				if existing.IsAssigned || !h.IsAssigned {
					continue
				}
			}
			dedupExam[k] = h
		}
		examData = examData[:0]
		for _, v := range dedupExam {
			examData = append(examData, v)
			examCourseMap[v.Id] = v.CourseId
			examMap[v.Id] = v
		}

		examIds = examIds[:0]
		examIdSet = make(map[int64]struct{})
		for _, h := range examData {
			if _, ok := examIdSet[h.Id]; !ok {
				examIdSet[h.Id] = struct{}{}
				examIds = append(examIds, h.Id)
			}
		}

	}

	if includeExercise {
		dedupExercise := make(map[key]DataEntry)
		for _, h := range exerciseData {
			k := key{CourseID: h.CourseId, DataID: h.Id}
			if existing, ok := dedupExercise[k]; ok {
				if existing.IsAssigned || !h.IsAssigned {
					continue
				}
			}
			dedupExercise[k] = h
		}
		exerciseData = exerciseData[:0]
		for _, v := range dedupExercise {
			exerciseData = append(exerciseData, v)
			exerciseCourseMap[v.Id] = v.CourseId
			exerciseMap[v.Id] = v
		}

		exerciseIds = exerciseIds[:0]
		exerciseIdSet = make(map[int64]struct{})
		for _, h := range exerciseData {
			if _, ok := exerciseIdSet[h.Id]; !ok {
				exerciseIdSet[h.Id] = struct{}{}
				exerciseIds = append(exerciseIds, h.Id)
			}
		}

	}

	// Phase 5: Chạy song song 6 query needScoring + submitted
	wg3 := sync.WaitGroup{}
	if includeHomework && len(homeworkIds) > 0 {
		wg3.Add(2)
		go func() {
			defer wg3.Done()
			q := db.Table("homework_question_user_manual_scoring hqums").
				Select(`hqums.homework_id AS data_id, hqums.user_id, BOOL_AND(COALESCE(hqums.is_scored, false)) AS is_scored, MIN(hqums.created_at) AS date, SUM(CASE WHEN COALESCE(hqums.is_scored, false) = FALSE THEN 1 ELSE 0 END)::int AS count`).
				Where("hqums.homework_id IN (?)", homeworkIds).Group("hqums.homework_id, hqums.user_id")
			if len(userIds) > 0 {
				q = q.Where("hqums.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&homeworkNeedScoring).Error)
		}()
		go func() {
			defer wg3.Done()
			q := db.Table("homework_users hu").Select("hu.homework_id AS data_id, hu.user_id, MIN(hu.created_at) AS date").Where("hu.homework_id IN (?)", homeworkIds).Group("hu.homework_id, hu.user_id")
			if len(userIds) > 0 {
				q = q.Where("hu.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&homeworkSubmitted).Error)
		}()
	}
	if includeExam && len(examIds) > 0 {
		wg3.Add(2)
		go func() {
			defer wg3.Done()
			q := db.Table("exam_question_user_manual_scoring equms").
				Select(`equms.exam_id AS data_id, equms.user_id, BOOL_AND(COALESCE(equms.is_scored, false)) AS is_scored, MIN(equms.created_at) AS date, SUM(CASE WHEN COALESCE(equms.is_scored, false) = FALSE THEN 1 ELSE 0 END)::int AS count`).
				Where("equms.exam_id IN (?)", examIds).Group("equms.exam_id, equms.user_id").Having("BOOL_AND(COALESCE(equms.is_scored, false)) = FALSE")
			if len(userIds) > 0 {
				q = q.Where("equms.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&examNeedScoring).Error)
		}()
		go func() {
			defer wg3.Done()
			q := db.Table("exam_users eu").Select("eu.exam_id AS data_id, eu.user_id, MIN(eu.created_at) AS date").Where("eu.exam_id IN (?)", examIds).Group("eu.exam_id, eu.user_id")
			if len(userIds) > 0 {
				q = q.Where("eu.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&examSubmitted).Error)
		}()
	}
	if includeExercise && len(exerciseIds) > 0 {
		wg3.Add(2)
		go func() {
			defer wg3.Done()
			q := db.Table("exercise_question_user_manual_scoring equms").
				Select(`equms.exercise_id AS data_id, equms.user_id, BOOL_AND(COALESCE(equms.is_scored, false)) AS is_scored, MIN(equms.created_at) AS date, SUM(CASE WHEN COALESCE(equms.is_scored, false) = FALSE THEN 1 ELSE 0 END)::int AS count`).
				Where("equms.exercise_id IN (?)", exerciseIds).Group("equms.exercise_id, equms.user_id").Having("BOOL_AND(COALESCE(equms.is_scored, false)) = FALSE")
			if len(userIds) > 0 {
				q = q.Where("equms.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&exerciseNeedScoring).Error)
		}()
		go func() {
			defer wg3.Done()
			q := db.Table("exercise_users eu").Select("eu.exercise_id AS data_id, eu.user_id, MIN(eu.created_at) AS date").Where("eu.exercise_id IN (?)", exerciseIds).Group("eu.exercise_id, eu.user_id")
			if len(userIds) > 0 {
				q = q.Where("eu.user_id IN (?)", userIds)
			}
			setErr(q.Scan(&exerciseSubmitted).Error)
		}()
	}
	wg3.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// Build NotSubmitted (cần homeworkSubmitted, examSubmitted, exerciseSubmitted đã có từ phase 5)
	if includeHomework && len(userIds) > 0 && len(homeworkIds) > 0 {
		submittedSet := make(map[int64]map[int64]struct{})
		for _, s := range homeworkSubmitted {
			if _, ok := submittedSet[s.DataID]; !ok {
				submittedSet[s.DataID] = make(map[int64]struct{})
			}
			submittedSet[s.DataID][s.UserID] = struct{}{}
		}
		for _, h := range homeworkData {
			if _, ok := homeworkIdSet[h.Id]; !ok {
				continue
			}
			for _, uid := range userIds {
				if _, ok := submittedSet[h.Id][uid]; !ok {
					homeworkNotSubmitted = append(homeworkNotSubmitted, DataNotSubmitted{DataID: h.Id, CourseID: h.CourseId, UserID: uid})
				}
			}
		}
	}
	if includeExam && len(userIds) > 0 && len(examIds) > 0 {
		submittedSet := make(map[int64]map[int64]struct{})
		for _, s := range examSubmitted {
			if _, ok := submittedSet[s.DataID]; !ok {
				submittedSet[s.DataID] = make(map[int64]struct{})
			}
			submittedSet[s.DataID][s.UserID] = struct{}{}
		}
		for _, h := range examData {
			if _, ok := examIdSet[h.Id]; !ok {
				continue
			}
			for _, uid := range userIds {
				if _, ok := submittedSet[h.Id][uid]; !ok {
					examNotSubmitted = append(examNotSubmitted, DataNotSubmitted{DataID: h.Id, CourseID: h.CourseId, UserID: uid})
				}
			}
		}
	}
	if includeExercise && len(userIds) > 0 && len(exerciseIds) > 0 {
		submittedSet := make(map[int64]map[int64]struct{})
		for _, s := range exerciseSubmitted {
			if _, ok := submittedSet[s.DataID]; !ok {
				submittedSet[s.DataID] = make(map[int64]struct{})
			}
			submittedSet[s.DataID][s.UserID] = struct{}{}
		}
		for _, h := range exerciseData {
			if _, ok := exerciseIdSet[h.Id]; !ok {
				continue
			}
			for _, uid := range userIds {
				if _, ok := submittedSet[h.Id][uid]; !ok {
					exerciseNotSubmitted = append(exerciseNotSubmitted, DataNotSubmitted{DataID: h.Id, CourseID: h.CourseId, UserID: uid})
				}
			}
		}
	}

	studentIdsSet := make(map[int64]struct{})
	for _, ng := range homeworkNeedScoring {
		studentIdsSet[ng.UserID] = struct{}{}
	}
	for _, ng := range examNeedScoring {
		studentIdsSet[ng.UserID] = struct{}{}
	}
	for _, ng := range exerciseNeedScoring {
		studentIdsSet[ng.UserID] = struct{}{}
	}
	studentIds := make([]int64, 0, len(studentIdsSet))
	for id := range studentIdsSet {
		studentIds = append(studentIds, id)
	}

	var allWeeks []struct {
		Id         int64
		StartDate  time.Time
		EndDate    time.Time
		WeekNumber int32
	}
	wg4 := sync.WaitGroup{}
	if len(studentIds) > 0 {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			var rows []struct {
				Id   int64
				Name string
			}
			setErr(db.Table("users").Select("id, name").Where("id IN (?)", studentIds).Scan(&rows).Error)
			for _, r := range rows {
				mu.Lock()
				studentMap[r.Id] = r.Name
				mu.Unlock()
			}
		}()
	}
	wg4.Add(1)
	go func() {
		defer wg4.Done()
		weekQuery := db.Table("weeks")
		if filterByDate {
			weekQuery = weekQuery.Where("end_date >= ? AND end_date <= ?", filter.StartTime, filter.EndTime)
		} else {
			weekQuery = weekQuery.Where("end_date <= NOW()")
		}
		if err := weekQuery.Order("id ASC").Limit(20).Select("id, start_date, end_date, week_number").Scan(&allWeeks).Error; err != nil {
			config.Log.Error("Failed to get all weeks: ", err)
			allWeeks = []struct {
				Id         int64
				StartDate  time.Time
				EndDate    time.Time
				WeekNumber int32
			}{}
		}
	}()
	wg4.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	beginAt, _ := time.Parse("2006-01-02", "2025-09-08")
	for i := range allWeeks {
		diff := allWeeks[i].StartDate.Sub(beginAt)
		weekNumber := int(diff.Hours()/(24*7)) + 1
		if weekNumber < 1 {
			weekNumber = 1
		}
		allWeeks[i].WeekNumber = int32(weekNumber)
	}

	hwDateMap := make(map[int64]time.Time)
	hwCourseMap := make(map[int64]int64)
	homeworkMap := make(map[int64]DataEntry)
	for _, h := range homeworkData {
		hwDateMap[h.Id] = h.Date
		hwCourseMap[h.Id] = h.CourseId
		homeworkMap[h.Id] = h
	}

	var weeklySubmitRates []dto.WeeklySubmitRate
	var weeklyMarkingRates []dto.WeeklyMarkingRate
	var weeklyAssignAssignmentRates []dto.WeeklyAssignAssignmentRate
	var weeklyPerformanceOverviewRates []dto.WeeklyPerformanceOverviewRate

	weeklySubmitRates = make([]dto.WeeklySubmitRate, 0, len(allWeeks))
	weeklyMarkingRates = make([]dto.WeeklyMarkingRate, 0, len(allWeeks))
	for _, weekRow := range allWeeks {
		total := int32(0)
		notGraded := int32(0)

		var notGradedExams, notGradedHomeworks, notGradedExercises []dto.NotGraded

		for _, ng := range homeworkNeedScoring {
			if date, ok := hwDateMap[ng.DataID]; ok {
				if !date.Before(weekRow.StartDate) && date.Before(weekRow.EndDate.AddDate(0, 0, 1)) {
					total++
					if !ng.IsScored {
						notGraded++
						courseID := hwCourseMap[ng.DataID]
						tInfo, ok := courseTeacherMap[courseID]
						notGradedItem := dto.NotGraded{
							NotGradedId:   ng.DataID,
							StudentId:     ng.UserID,
							SubmittedAt:   ng.Date.Format(time.RFC3339),
							QuestionCount: ng.Count,
							CourseId:      courseID,
						}
						if ok {
							notGradedItem.TeacherId = tInfo.MinUserID
							notGradedItem.TeacherName = strings.Join(tInfo.Names, ", ")
						}
						if hInfo, ok := homeworkMap[ng.DataID]; ok {
							notGradedItem.NotGradedName = hInfo.Name
							notGradedItem.LessonName = hInfo.LessonName
						}
						if cname, ok := courseNameMap[courseID]; ok {
							notGradedItem.CourseName = cname
						}
						if sName, ok := studentMap[ng.UserID]; ok {
							notGradedItem.StudentName = sName
						}
						notGradedHomeworks = append(notGradedHomeworks, notGradedItem)
					}
				}
			}
		}
		for _, ng := range examNeedScoring {
			if !ng.Date.Before(weekRow.StartDate) && ng.Date.Before(weekRow.EndDate.AddDate(0, 0, 1)) {
				total++
				if !ng.IsScored {
					notGraded++
					ngItem := dto.NotGraded{
						NotGradedId:   ng.DataID,
						StudentId:     ng.UserID,
						SubmittedAt:   ng.Date.Format(time.RFC3339),
						QuestionCount: ng.Count,
					}
					if cid, ok := examCourseMap[ng.DataID]; ok {
						if tInfo, ok2 := courseTeacherMap[cid]; ok2 {
							ngItem.TeacherId = tInfo.MinUserID
							ngItem.TeacherName = strings.Join(tInfo.Names, ", ")
						}
					}
					if eInfo, ok := examMap[ng.DataID]; ok {
						ngItem.NotGradedName = eInfo.Name
					}
					if sName, ok := studentMap[ng.UserID]; ok {
						ngItem.StudentName = sName
					}
					notGradedExams = append(notGradedExams, ngItem)
				}
			}
		}
		for _, ng := range exerciseNeedScoring {
			if !ng.Date.Before(weekRow.StartDate) && ng.Date.Before(weekRow.EndDate.AddDate(0, 0, 1)) {
				total++
				if !ng.IsScored {
					notGraded++
					ngItem := dto.NotGraded{
						NotGradedId:   ng.DataID,
						StudentId:     ng.UserID,
						SubmittedAt:   ng.Date.Format(time.RFC3339),
						QuestionCount: ng.Count,
					}
					if cid, ok := exerciseCourseMap[ng.DataID]; ok {
						if tInfo, ok2 := courseTeacherMap[cid]; ok2 {
							ngItem.TeacherId = tInfo.MinUserID
							ngItem.TeacherName = strings.Join(tInfo.Names, ", ")
						}
					}
					if exInfo, ok := exerciseMap[ng.DataID]; ok {
						ngItem.NotGradedName = exInfo.Name
					}
					if sName, ok := studentMap[ng.UserID]; ok {
						ngItem.StudentName = sName
					}
					notGradedExercises = append(notGradedExercises, ngItem)
				}
			}
		}

		weeklySubmitRates = append(weeklySubmitRates, dto.WeeklySubmitRate{
			Id:                 weekRow.Id,
			WeekNumber:         strconv.Itoa(int(weekRow.WeekNumber)),
			Total:              total,
			NotGraded:          notGraded,
			NotGradedExams:     notGradedExams,
			NotGradedHomeworks: notGradedHomeworks,
			NotGradedExercises: notGradedExercises,
		})

		markedCount := total - notGraded
		markingPercent := float32(0)
		if total > 0 {
			markingPercent = float32(markedCount) * 100 / float32(total)
		}

		ungradedSubmissions := []dto.UngradedSubmission{}
		for _, ng := range homeworkNeedScoring {
			if date, ok := hwDateMap[ng.DataID]; ok {
				if !date.Before(weekRow.StartDate) && date.Before(weekRow.EndDate.AddDate(0, 0, 1)) && !ng.IsScored {
					ungradedSubmissions = append(ungradedSubmissions, dto.UngradedSubmission{
						Type:      "homework",
						Id:        ng.DataID,
						StudentId: ng.UserID,
					})
				}
			}
		}

		weeklyMarkingRates = append(weeklyMarkingRates, dto.WeeklyMarkingRate{
			Id:                  weekRow.Id,
			WeekNumber:          strconv.Itoa(int(weekRow.WeekNumber)),
			MarkingPercent:      markingPercent,
			AssignedCount:       total,
			MarkedCount:         markedCount,
			UngradedSubmissions: ungradedSubmissions,
		})

		var totalAssigned, totalAssignAssignment int32

		for _, h := range homeworkData {
			if date, ok := hwDateMap[h.Id]; ok {
				if !date.Before(weekRow.StartDate) && date.Before(weekRow.EndDate.AddDate(0, 0, 1)) {
					totalAssignAssignment++
					if !h.IsAssigned {
						totalAssigned++
					}
				}
			}
		}

		weeklyAssignAssignmentRates = append(weeklyAssignAssignmentRates, dto.WeeklyAssignAssignmentRate{
			Id:                           weekRow.Id,
			WeekNumber:                   strconv.Itoa(int(weekRow.WeekNumber)),
			Total:                        totalAssignAssignment,
			AssignAssignment:             totalAssigned,
			NotAssignAssignmentExams:     []dto.NotAssignAssignment{},
			NotAssignAssignmentHomeworks: []dto.NotAssignAssignment{},
			NotAssignAssignmentExercises: []dto.NotAssignAssignment{},
		})

		assignPercent := float32(0)
		if totalAssignAssignment > 0 {
			assignPercent = float32(totalAssigned) * 100 / float32(totalAssignAssignment)
		}
		submitPercent := float32(0)
		if total > 0 {
			totalSubmitted := total + int32(len(homeworkNotSubmitted))
			submitPercent = float32(len(homeworkNotSubmitted)) * 100 / float32(totalSubmitted)
		}
		notGradedPercent := float32(0)
		if total > 0 {
			notGradedPercent = float32(notGraded) * 100 / float32(total)
		}

		weeklyPerformanceOverviewRates = append(weeklyPerformanceOverviewRates, dto.WeeklyPerformanceOverviewRate{
			Id:                      weekRow.Id,
			WeekNumber:              strconv.Itoa(int(weekRow.WeekNumber)),
			AssignAssignmentPercent: assignPercent,
			SubmitPercent:           submitPercent,
			NotGradedPercent:        notGradedPercent,
			TotalRequired:           total,
			TotalSubmitted:          total,
			NotSubmittedExams:       []dto.NotSubmittedExam{},
			NotSubmittedHomeworks:   []dto.NotSubmittedGrouped{},
			NotSubmittedExercises:   []dto.NotSubmittedGrouped{},
		})
	}

	var totalUngradedSubmissions, gradedSubmissions int32

	for _, h := range homeworkNeedScoring {
		if !h.IsScored {
			totalUngradedSubmissions++
		} else {
			gradedSubmissions++
		}
	}

	totalSubmissions := len(homeworkSubmitted)
	totalNeedScoring := len(homeworkNeedScoring)

	ungradedPercent := float32(0)
	if totalSubmissions > 0 {
		ungradedPercent = float32(totalUngradedSubmissions) * 100 / float32(totalNeedScoring)
	}

	gradingSummary := dto.GradingSummary{
		TotalSubmissions:    int32(totalSubmissions),
		GradedSubmissions:   int32(gradedSubmissions),
		UngradedSubmissions: int32(totalUngradedSubmissions),
		UngradedPercent:     ungradedPercent,
	}

	return &dto.TeacherPerformanceOverview{
		GradingSummary:                 gradingSummary,
		WeeklyMarkingRates:             weeklyMarkingRates,
		WeeklyAssignAssignmentRates:    weeklyAssignAssignmentRates,
		WeeklySubmitRates:              weeklySubmitRates,
		WeeklyPerformanceOverviewRates: weeklyPerformanceOverviewRates,
	}, nil
}

