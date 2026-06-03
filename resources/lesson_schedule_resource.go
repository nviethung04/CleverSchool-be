package resources

import (
	"be-lms/dto"
	"be-lms/models"
	"be-lms/prot"
	"sort"
	"time"
)

type LessonScheduleResource interface {
	FormatLessonSchedulesInfo(lessonScheduleDetail dto.LessonScheduleDetail) *prot.LessonSchedulesInfoResponse
	FormatLessonSchedulesByWeek(group dto.LessonSchedule) *prot.ScheduleByWeek
	FormatScheduleGroups(groups []dto.ScheduleGroup) []*prot.LessonScheduleGroup
}

type LessonScheduleResourceImpl struct{
	BeginTime *time.Time
	HolidayWeeks []models.Week
	AutoIndex bool
}

func NewLessonScheduleResource() LessonScheduleResource {
	return &LessonScheduleResourceImpl{
		BeginTime: nil,
		HolidayWeeks: []models.Week{},
		AutoIndex: false,
	}
}

func (r *LessonScheduleResourceImpl) FormatLessonSchedule(schedule *models.LessonSchedule) *prot.LessonScheduleDetail {
	if schedule == nil {
		return nil
	}

	// 🚫 Nếu tuần này là tuần nghỉ thì bỏ qua luôn
	for _, hw := range r.HolidayWeeks {
		if (schedule.ScheduledDate.Equal(hw.StartDate) || schedule.ScheduledDate.After(hw.StartDate)) &&
			(schedule.ScheduledDate.Equal(hw.EndDate) || schedule.ScheduledDate.Before(hw.EndDate)) {
			return nil
		}
	}

	var beginDate, scheduleDate time.Time
	var weekNumber int

	if r.BeginTime != nil {
		beginDate = *r.BeginTime
		scheduleDate = schedule.ScheduledDate
	} else if schedule.Course != nil {
		beginDate = schedule.Course.StartDate
		scheduleDate = schedule.ScheduledDate
	}

	if !beginDate.IsZero() {
		diffDays := int(scheduleDate.Sub(beginDate).Hours() / 24)
		weekNumber = (diffDays / 7) + 1
		if diffDays < 0 {
			weekNumber = 0
		}
	}

	// ✅ Trừ đi số tuần nghỉ trước ngày học
	holidayCount := 0
	for _, hw := range r.HolidayWeeks {
		if hw.EndDate.Before(schedule.ScheduledDate) || hw.EndDate.Equal(schedule.ScheduledDate) {
			holidayCount++
		}
	}
	weekNumber = weekNumber - holidayCount
	if weekNumber < 1 {
		weekNumber = 1
	}

	var lesson *prot.LessonBySchedule
	if schedule.Lesson.ID != 0 {
		lesson = &prot.LessonBySchedule{
			Id:          schedule.Lesson.ID,
			Title:       schedule.Lesson.Title,
			Description: schedule.Lesson.Description,
		}
	}

	var course *prot.CourseBySchedule
	if schedule.Course != nil {
		course = &prot.CourseBySchedule{
			Id:          schedule.Course.ID,
			Name:        schedule.Course.Name,
			Description: schedule.Course.Description,
			ObjectTitle: schedule.Course.ObjectTitle,
		}
	}


	var chapter *prot.ChapterBySchedule
	if schedule.Lesson.Chapter.ID != 0 {
		chapter = &prot.ChapterBySchedule{
			Id:   schedule.Lesson.Chapter.ID,
			Title: schedule.Lesson.Chapter.Title,
			Description: schedule.Lesson.Chapter.Description,
		}
	}

	return &prot.LessonScheduleDetail{
		Id:            int64(schedule.ID),
		ScheduledDate: schedule.ScheduledDate.Format("2006-01-02"),
		WeekNumber:    int32(weekNumber),
		Lesson:        lesson,
		Course:        course,
		Chapter:       chapter,
		SortPosition:  int32(schedule.SortPosition),
	}
}

func (r *LessonScheduleResourceImpl) FormatLessonSchedules(schedules []*models.LessonSchedule) []*prot.LessonScheduleDetail {
	result := make([]*prot.LessonScheduleDetail, 0, len(schedules))
	for _, s := range schedules {
		if detail := r.FormatLessonSchedule(s); detail != nil {
			result = append(result, detail)
		}
	}

	// Sắp xếp: SortPosition tăng dần, nếu bằng nhau thì theo Id tăng dần
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortPosition == result[j].SortPosition {
			return result[i].Id < result[j].Id
		}
		return result[i].SortPosition < result[j].SortPosition
	})

	return result
}

func (r *LessonScheduleResourceImpl) FormatScheduleGroups(groups []dto.ScheduleGroup) []*prot.LessonScheduleGroup {
	var result []*prot.LessonScheduleGroup

	for _, group := range groups {
		days := make(map[string]*prot.LessonScheduleList)

		for day, schedules := range group.Days {
			if len(schedules) == 0 {
				continue
			}

			schedulePtrs := make([]*models.LessonSchedule, 0, len(schedules))
			for i := range schedules {
				schedulePtrs = append(schedulePtrs, &schedules[i])
			}

			date := schedules[0].ScheduledDate.Format("2006-01-02")

			days[day] = &prot.LessonScheduleList{
				Date:      date,
				Schedules: r.FormatLessonSchedules(schedulePtrs),
			}
		}

		result = append(result, &prot.LessonScheduleGroup{
			WeekId:     int32(group.WeekId),
			WeekNumber: int32(group.WeekNumber),
			Year:       int32(group.Year),
			StartDate:  group.StartDate.Format("2006-01-02"),
			EndDate:    group.EndDate.Format("2006-01-02"),
			Days:       days,
		})
	}

	return result
}

func (r *LessonScheduleResourceImpl) FormatLessonSchedulesInfo(lessonScheduleDetail dto.LessonScheduleDetail) *prot.LessonSchedulesInfoResponse {
	var result []*prot.ScheduleByWeek

	uniqueHolidayWeeks := make([]models.Week, 0, len(r.HolidayWeeks))
	seen := make(map[int64]bool)

	for _, hw := range r.HolidayWeeks {
		if !seen[hw.ID] {
			seen[hw.ID] = true
			uniqueHolidayWeeks = append(uniqueHolidayWeeks, hw)
		}
	}

	r.HolidayWeeks = uniqueHolidayWeeks

	groups := lessonScheduleDetail.Groups

	weekNumberIndex := 0

	for _, group := range groups {
		weekNumberIndex += 1
		// ✅ Nếu tuần này là tuần nghỉ → bỏ qua, không show
		isHoliday := false
		for _, hw := range r.HolidayWeeks {
			if int(hw.ID) == group.WeekId {
				isHoliday = true
				break
			}
		}
		if isHoliday {
			continue
		}

		// gom con trỏ
		schedulePtrs := make([]*models.LessonSchedule, 0, len(group.Lessons))
		for i := range group.Lessons {
			schedulePtrs = append(schedulePtrs, &group.Lessons[i])
		}

		schedules := r.FormatLessonSchedules(schedulePtrs)

		var weekNumber int32
		if len(schedules) > 0 {
			weekNumber = schedules[0].WeekNumber
		} else {
			if r.BeginTime != nil {
				beginDate := *r.BeginTime
				startDate := group.StartDate

				diffDays := int(startDate.Sub(beginDate).Hours() / 24)
				weekNumber = int32(diffDays/7) + 1

				// ✅ Trừ số tuần nghỉ nằm trước tuần này
				holidayCount := 0
				for _, hw := range r.HolidayWeeks {
					if hw.ID < int64(group.WeekId) {
						holidayCount++
					}
					// if hw.EndDate.Before(startDate) || hw.EndDate.Equal(startDate) {
					// 	holidayCount++
					// }
				}
				weekNumber -= int32(holidayCount)
			} else {
				weekNumber = int32(group.WeekNumber)
			}
		}

		if weekNumber < 1 {
			weekNumber = 1
		}

		if r.AutoIndex {
			weekNumber = int32(weekNumberIndex)
		}

		result = append(result, &prot.ScheduleByWeek{
			WeekId:           int32(group.WeekId),
			WeekNumberInYear: int32(group.WeekNumberByYear),
			WeekNumber:       weekNumber,
			StartDate:        group.StartDate.Format("2006-01-02"),
			EndDate:          group.EndDate.Format("2006-01-02"),
			Schedules:        schedules,
		})
	}

	semesters := []*prot.ScheduleBySemester{}
	for _, s := range lessonScheduleDetail.Semesters {
		semester := &prot.ScheduleBySemester{
			Id:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			StartDate:   s.StartDate.Format("2006-01-02"),
			EndDate:     s.EndDate.Format("2006-01-02"),
		}

		for _, w := range result {
			startDate, _ := time.Parse("2006-01-02", w.StartDate)
			endDate, _ := time.Parse("2006-01-02", w.EndDate)

			if !(endDate.Before(s.StartDate) || startDate.After(s.EndDate)) {
				semester.Weeks = append(semester.Weeks, w)
			}
		}

		semesters = append(semesters, semester)
	}

	sort.Slice(semesters, func(i, j int) bool {
		t1, _ := time.Parse("2006-01-02", semesters[i].StartDate)
		t2, _ := time.Parse("2006-01-02", semesters[j].StartDate)
		return t1.Before(t2)
	})

	return &prot.LessonSchedulesInfoResponse{
		Weeks: result,
		Semesters: semesters,
	}
}

func (r *LessonScheduleResourceImpl) FormatLessonSchedulesByWeek(group dto.LessonSchedule) *prot.ScheduleByWeek {
	schedulePtrs := make([]*models.LessonSchedule, 0, len(group.Lessons))
	for i := range group.Lessons {
		schedulePtrs = append(schedulePtrs, &group.Lessons[i])
	}

	schedules := r.FormatLessonSchedules(schedulePtrs)

	weekNumber := int32(group.WeekNumber)

	if len(schedules) > 0 {
		weekNumber = schedules[0].WeekNumber
	} else {
		weekNumber = int32(group.WeekNumber)
	}

	return &prot.ScheduleByWeek{
		WeekId:           int32(group.WeekId),
		WeekNumberInYear: int32(group.WeekNumberByYear),
		WeekNumber:       weekNumber,
		StartDate:        group.StartDate.Format("2006-01-02"),
		EndDate:          group.EndDate.Format("2006-01-02"),
		Schedules:        schedules,
	}
}
