package services

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"errors"
	"fmt"
)

type CourseScheduleService interface {
	AssignParentAndCopySchedule(courseIDs []int64, parentCourseID int64) (*prot.CourseScheduleActionResponse, error)
	SyncFamilySchedules(courseID int64) (*prot.CourseFamilySyncResponse, error)
}

type courseScheduleService struct {
	copyService LessonScheduleCopySharedService
	familyRepo  repositories.CourseFamilyRepository
}

type courseScheduleInfo struct {
	ID             int64
	ProgramID      int64
	ParentCourseID int64
}

var errCourseNotFound = errors.New("course_not_found")

func NewCourseScheduleService() CourseScheduleService {
	return &courseScheduleService{
		copyService: NewLessonScheduleCopySharedService(),
		familyRepo:  repositories.NewCourseFamilyRepository(),
	}
}

func (s *courseScheduleService) AssignParentAndCopySchedule(courseIDs []int64, parentCourseID int64) (*prot.CourseScheduleActionResponse, error) {
	if parentCourseID <= 0 {
		return nil, errors.New("parent_course_id phải lớn hơn 0")
	}

	// Lấy thông tin course cha
	parentInfo, err := s.getCourseInfo(parentCourseID)
	if err != nil {
		if errors.Is(err, errCourseNotFound) {
			return nil, fmt.Errorf("không tìm thấy khóa cha %d", parentCourseID)
		}
		return nil, err
	}

	if parentInfo.ProgramID == 0 {
		return nil, errors.New("khóa cha phải có program_id")
	}

	// Đặt parent_course_id của course cha thành 0
	if err := db.MasterDB.Model(&models.Course{}).
		Where("id = ?", parentCourseID).
		Update("parent_course_id", 0).Error; err != nil {
		return nil, fmt.Errorf("không thể cập nhật parent_course_id của khóa cha: %w", err)
	}

	// Nếu course_ids rỗng, chỉ cập nhật course cha và trả về
	if len(courseIDs) == 0 {
		return &prot.CourseScheduleActionResponse{
			Success:         true,
			Message:         fmt.Sprintf("Đã đặt khóa học %d làm khóa cha (không có khóa con)", parentCourseID),
			CourseId:        0,
			ParentCourseId:  parentCourseID,
			CopiedSchedules:  0,
		}, nil
	}

	// Kiểm tra parent_course_id không nằm trong danh sách course_ids
	for _, courseID := range courseIDs {
		if courseID <= 0 {
			return nil, errors.New("course_ids phải lớn hơn 0")
		}
		if courseID == parentCourseID {
			return nil, errors.New("parent_course_id không được nằm trong course_ids")
		}
	}

	// Validate tất cả các course con cùng program_id với course cha
	var childInfos []courseScheduleInfo
	err = db.ReplicaDB.Table("courses").
		Select("id, program_id, parent_course_id").
		Where("id IN ? AND deleted_at IS NULL", courseIDs).
		Scan(&childInfos).Error
	if err != nil {
		return nil, fmt.Errorf("không thể lấy thông tin các khóa học: %w", err)
	}

	if len(childInfos) != len(courseIDs) {
		return nil, errors.New("một số khóa học không tồn tại hoặc đã bị xóa")
	}

	for _, childInfo := range childInfos {
		if childInfo.ProgramID == 0 || childInfo.ProgramID != parentInfo.ProgramID {
			return nil, fmt.Errorf("khóa học %d phải cùng program_id với khóa cha", childInfo.ID)
		}
	}

	// Đặt parent_course_id của các course con thành parentCourseID
	if err := db.MasterDB.Model(&models.Course{}).
		Where("id IN ?", courseIDs).
		Update("parent_course_id", parentCourseID).Error; err != nil {
		return nil, fmt.Errorf("không thể cập nhật parent_course_id của các khóa con: %w", err)
	}

	// Copy schedule từ course cha sang các course con
	totalCopiedSchedules := int32(0)
	var copiedCourseIDs []int64
	for _, courseID := range courseIDs {
		copyResult, err := s.copyService.CopyLessonSchedulesBetweenCourses(parentCourseID, courseID)
		if err != nil {
			// Log lỗi nhưng tiếp tục với các course khác
			continue
		}
		if copyResult.Success {
			totalCopiedSchedules += copyResult.CopiedSchedulesCount
			copiedCourseIDs = append(copiedCourseIDs, courseID)
		}
	}

	message := fmt.Sprintf("Đã gán %d khóa học làm con của khóa cha %d", len(courseIDs), parentCourseID)
	if len(copiedCourseIDs) > 0 {
		message += fmt.Sprintf(". Đã copy schedule cho %d khóa học", len(copiedCourseIDs))
	}

	return &prot.CourseScheduleActionResponse{
		Success:         true,
		Message:         message,
		CourseId:        courseIDs[0], // Giữ lại course_id đầu tiên để tương thích
		ParentCourseId:  parentCourseID,
		CopiedSchedules:  totalCopiedSchedules,
	}, nil
}

func (s *courseScheduleService) getCourseInfo(courseID int64) (*courseScheduleInfo, error) {
	var info courseScheduleInfo
	err := db.ReplicaDB.Table("courses").
		Select("id, program_id, parent_course_id").
		Where("id = ? AND deleted_at IS NULL", courseID).
		Scan(&info).Error

	if err != nil {
		return nil, err
	}

	if info.ID == 0 {
		return nil, errCourseNotFound
	}

	return &info, nil
}

func (s *courseScheduleService) SyncFamilySchedules(courseID int64) (*prot.CourseFamilySyncResponse, error) {
	if courseID <= 0 {
		return nil, errors.New("course_id phải lớn hơn 0")
	}

	sourceInfo, err := s.getCourseInfo(courseID)
	if err != nil {
		if errors.Is(err, errCourseNotFound) {
			return nil, fmt.Errorf("không tìm thấy khóa học %d", courseID)
		}
		return nil, err
	}

	rootCourseID := sourceInfo.ID
	if sourceInfo.ParentCourseID != 0 {
		rootCourseID = sourceInfo.ParentCourseID
	}

	children, err := s.familyRepo.GetChildren(rootCourseID)
	if err != nil {
		return nil, err
	}

	var targetCourseIDs []int64
	targetCourseIDs = append(targetCourseIDs, rootCourseID)
	for _, child := range children {
		if child.ID > 0 {
			targetCourseIDs = append(targetCourseIDs, child.ID)
		}
	}

	var syncedCourseIDs []int64
	for _, targetID := range targetCourseIDs {
		if targetID == courseID {
			continue
		}

		copyResult, err := s.copyService.CopyLessonSchedulesBetweenCourses(courseID, targetID)
		if err != nil {
			return nil, err
		}
		if !copyResult.Success {
			return nil, fmt.Errorf("không thể đồng bộ sang khóa %d: %s", targetID, copyResult.Message)
		}
		syncedCourseIDs = append(syncedCourseIDs, targetID)
	}

	if len(syncedCourseIDs) == 0 {
		return nil, errors.New("family không có khóa nào khác để đồng bộ")
	}

	return &prot.CourseFamilySyncResponse{
		Success:         true,
		Message:         fmt.Sprintf("Đồng bộ thành công tới %d khóa", len(syncedCourseIDs)),
		SourceCourseId:  courseID,
		SyncedCourseIds: syncedCourseIDs,
		SyncedCount:     int32(len(syncedCourseIDs)),
	}, nil
}

