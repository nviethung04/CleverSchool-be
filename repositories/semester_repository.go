package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"errors"
	"time"

	"gorm.io/gorm"
)

type SemesterRepository interface {
	base.BaseRepositoryInterface[models.Semester]
	GetIdsByUserID(id int) ([]int, error)
	UpdateOrCreate(holiday models.Holiday) (*models.Holiday, error)
	GetAllHolidaysBySemesterID(semesterID int64) ([]models.Holiday, error)
	GetHolidayWeeksBySemesterID(semesterID int64) ([]models.Week, error)
	UpdateOrCreateHoliday(holiday models.SemesterRefHoliday) error
	DeleteOldHolidays(id int64, holidayIds []int64) error
}

type semesterRepository struct {
	*base.BaseRepository[models.Semester]
}

func NewSemesterRepository() SemesterRepository {
	return &semesterRepository{
		BaseRepository: base.NewBaseRepository[models.Semester](),
	}
}

// Override BaseRepository.Create: omit week FK columns when 0 (FK requires NULL or valid weeks.id).
func (r *semesterRepository) Create(entity *models.Semester) error {
	if entity == nil {
		return errors.New("entity is nil")
	}

	if err := r.BeforeCreate(entity); err != nil {
		return err
	}

	omit := []string{"author_id"}
	if entity.StartWeekId == 0 {
		omit = append(omit, "StartWeekId")
	}
	if entity.EndWeekId == 0 {
		omit = append(omit, "EndWeekId")
	}

	return db.MasterDB.Omit(omit...).Create(entity).Error
}

// Override BaseRepository.Update: avoid writing start_week_id/end_week_id=0 (would violate FK).
func (r *semesterRepository) Update(entity *models.Semester) error {
	if entity == nil {
		return errors.New("entity is nil")
	}

	if err := r.BeforeUpdate(entity); err != nil {
		return err
	}

	omit := []string{"created_at", "created_by", "author_id"}
	if entity.StartWeekId == 0 {
		omit = append(omit, "StartWeekId")
	}
	if entity.EndWeekId == 0 {
		omit = append(omit, "EndWeekId")
	}

	return db.MasterDB.Omit(omit...).Save(entity).Error
}

func (r *semesterRepository) GetIdsByUserID(userID int) ([]int, error) {
	var semesterIDs []int

	err := db.ReplicaDB.
		Table("semesters").
		Select("semesters.id").
		Joins("JOIN course_ref_semesters ON semesters.id = course_ref_semesters.semester_id").
		Joins("JOIN courses ON courses.id = course_ref_semesters.course_id").
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Where("user_courses.user_id = ?", userID).
		Where("semesters.deleted_at IS NULL").
		Where("courses.deleted_at IS NULL").
		Group("semesters.id").
		Pluck("semesters.id", &semesterIDs).Error

	return semesterIDs, err
}

func (r *semesterRepository) UpdateOrCreate(holiday models.Holiday) (*models.Holiday, error) {
	if holiday.ID > 0 {
		err := db.MasterDB.Model(&models.Holiday{}).
			Where("id = ?", holiday.ID).
			Updates(&holiday).Error
		if err != nil {
			return nil, err
		}
		return &holiday, nil
	}

	err := db.MasterDB.Create(&holiday).Error
	if err != nil {
		return nil, err
	}

	return &holiday, nil
}

func (r *semesterRepository) UpdateOrCreateHoliday(holiday models.SemesterRefHoliday) error {
	var existing models.SemesterRefHoliday
	err := db.MasterDB.
		Where("semester_id = ? AND holiday_id = ?", holiday.SemesterId, holiday.HolidayId).
		First(&existing).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.MasterDB.Create(&holiday).Error
	}
	return err
}

func (r *semesterRepository) DeleteOldHolidays(semesterID int64, holidayIDs []int64) error {
	q := db.MasterDB.Model(&models.SemesterRefHoliday{}).Where("semester_id = ?", semesterID)

	if len(holidayIDs) > 0 {
		q = q.Where("holiday_id NOT IN ?", holidayIDs)
	}
	return q.Delete(&models.SemesterRefHoliday{}).Error
}

func (r *semesterRepository) GetAllHolidaysBySemesterID(semesterID int64) ([]models.Holiday, error) {
	var holidays []models.Holiday
	currentID := semesterID

	for currentID != 0 {
		var semester models.Semester
		if err := db.MasterDB.First(&semester, currentID).Error; err != nil {
			return holidays, err
		}

		var holidayIDs []int64
		if err := db.MasterDB.
			Model(&models.SemesterRefHoliday{}).
			Where("semester_id = ?", semester.ID).
			Pluck("holiday_id", &holidayIDs).Error; err != nil {
			return holidays, err
		}

		if len(holidayIDs) > 0 {
			var hs []models.Holiday
			if err := db.MasterDB.Where("id IN ?", holidayIDs).Find(&hs).Error; err != nil {
				return holidays, err
			}
			holidays = append(holidays, hs...)
		}

		currentID = semester.PreviousSemesterId
	}

	return holidays, nil
}


func (r *semesterRepository) GetHolidayWeeksBySemesterID(semesterID int64) ([]models.Week, error) {
    holidays, err := r.GetAllHolidaysBySemesterID(semesterID)
    if err != nil {
        return nil, err
    }
    if len(holidays) == 0 {
        return []models.Week{}, nil
    }

    var semester models.Semester
    if err := db.MasterDB.First(&semester, semesterID).Error; err != nil {
        return nil, err
    }

    weeks, _ := r.GetWeeksForSemester(semester)

    var holidayWeeks []models.Week
    for _, week := range weeks {
        ws := dateOnly(week.StartDate)
        we := dateOnly(week.EndDate)

        for _, holiday := range holidays {
            hs := dateOnly(holiday.StartDate)
            he := dateOnly(holiday.EndDate)

            // Tuần nghỉ hợp lệ chỉ khi toàn bộ tuần nằm trong khoảng holiday
            if (hs.Before(ws) || hs.Equal(ws)) && (he.After(we) || he.Equal(we)) {
                holidayWeeks = append(holidayWeeks, week)
                break
            }
        }
    }

    return holidayWeeks, nil
}

// Helper: generate weeks from semester
func (r *semesterRepository) GetWeeksForSemester(semester models.Semester) ([]models.Week, error) {
    var weeks []models.Week
    if err := db.MasterDB.
        Where("start_date >= ? AND end_date <= ?", semester.BeginDate, semester.EndDate).
        Order("start_date ASC").
        Find(&weeks).Error; err != nil {
        return nil, err
    }

    return weeks, nil
}

func dateOnly(t time.Time) time.Time {
    y, m, d := t.Date()
    return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
