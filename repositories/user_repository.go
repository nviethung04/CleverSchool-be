package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"be-lms/requests"
	"be-lms/table_manager"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetStudentsByParentID(parentID int) ([]models.User, error)
	base.BaseRepositoryInterface[models.User]
	base.BeforeQueryHook

	StoreUserClass(userClass models.UserClass) error
	UpdateUserClass(userClass models.UserClass) error
	DeleteUserClassesByUserID(userID int64) error
	GetUserClassesByUserID(userID int64) ([]models.UserClass, error)
	ExitsUserClassesByUserAndClassID(userID int64, classID int64) bool
	DeleteUserClassesByUserAndClassID(userID int64, classID int64) error

	UpdateOrCreateUserRefSubject(userSubject models.UserRefSubject) error
	DeleteUserRefSubjectsByUserID(userID int64) error
	GetUserRefSubjectsByUserID(userID int64) ([]models.UserRefSubject, error)
	DeleteUserRefSubjectsByUserAndSubjectID(userID int64, subjectID int64) error

	StoreUserCourse(userCourse models.UserCourse) error
	UpdateOrCreateUserCourse(userCourse models.UserCourse) error
	DeleteUserCoursesByUserID(userID int64) error
	GetUserCoursesByUserID(userID int64) ([]models.UserCourse, error)
	ExitsUserCoursesByUserAndCourseID(userID int64, courseID int64) bool
	DeleteUserCoursesByUserAndCourseID(userID int64, courseID int64) error

	FindByUsername(username string) (*models.User, error)

	UpdateOrCreateUserAddress(address models.UserAddress) (int64, error)
	DeleteOrtherAddress(addressID, userID int64) error

	DeleteCertificateByUserID(userID int64) error
	ExitsCertificate(certificate models.Certificate) bool
	CreateCertificate(certificate models.Certificate) (int64, error)
	UpdateCertificate(certificate models.Certificate) (int64, error)
	DeleteNotExitCertificate(userID int64, exitsIds []int64) error

	DeleteDegreeByUserID(userID int64) error
	ExitsDegree(certificate models.Degree) bool
	CreateDegree(certificate models.Degree) (int64, error)
	UpdateDegree(certificate models.Degree) (int64, error)
	DeleteNotExitDegree(userID int64, exitsIds []int64) error

	DeleteUserDepartmentByUserID(userID int64) error
	ExistUserDepartment(certificate models.UserDepartment) (int64, bool)
	CreateUserDepartment(certificate models.UserDepartment) (int64, error)
	UpdateUserDepartment(certificate models.UserDepartment) (int64, error)
	DeleteUserDepartmentNotIn(userID int64, exitsIds []int64) error

	DeleteUserPositionByUserID(userID int64) error
	ExistUserPosition(certificate models.UserPosition) (int64, bool)
	CreateUserPosition(certificate models.UserPosition) (int64, error)
	UpdateUserPositiont(certificate models.UserPosition) (int64, error)
	DeleteUserPositionNotIn(userID int64, exitsIds []int64) error

	GetAllUserClasses() ([]models.UserClass, error)
	GetAllUserCourses() ([]models.UserCourse, error)
	CreateOrUpdateUser(user *models.User) (int64, error)
	UsernameExits(username string, userID int64) (bool, error)
	UserExists(key string, id int64) (bool, error)

	GetAllUserClassByUser(id int64) ([]models.UserClass, error)

	GetClass(userID int64) (models.Class, error)
	GetStudentsByClassID(classID int64) ([]models.User, error)
	GetActivityLogByUserIds(userIDs []int64, start, end time.Time) (map[int64]models.ActivityLog, error)
	GetCourses(userID int64) ([]models.Course, error)
	GetClassesByUserId(userID int64) ([]models.Class, error)

	DeleteOldClassCourseRole(userIds []int64) error

	UserActivated(req requests.UserActivatedRequest) ([]*models.User, int64, error)
	ResetPassword(id int, password string) error

	UpdateOrCreateUserRefRole(userRole models.UserRefRole) error
	DeleteUserRefRolesByUserID(userID int64) error
	GetUserRefRolesByUserID(userID int64) ([]models.UserRefRole, error)
	DeleteUserRefRolesByUserAndRoleID(userID int64, roleID int64) error

	GetIdsByProgramStatus(notParticipated, failTheSubject, isFilterNotParticipated, isFilterFailTheSubject bool, programId int64) ([]int64, error)
}

type userRepository struct {
	*base.BaseRepository[models.User]
}

func NewUserRepository() *userRepository {
	repo := &userRepository{
		BaseRepository: base.NewBaseRepository[models.User](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *userRepository) GetStudentsByParentID(parentID int) ([]models.User, error) {
	var users []models.User
	err := db.ReplicaDB.Preload("Roles").Where("parent_id = ?", parentID).Find(&users).Error
	return users, err
}

func (r *userRepository) DeleteUserClassesByUserID(userID int64) error {
	if err := db.MasterDB.Exec("DELETE FROM user_classes WHERE user_id = ?", userID).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserClassesByUserID(userID int64) ([]models.UserClass, error) {
	var userClasses []models.UserClass
	err := db.MasterDB.Where("user_id = ?", userID).Find(&userClasses).Error
	return userClasses, err
}

func (r *userRepository) ExitsUserClassesByUserAndClassID(userID int64, classID int64) bool {
	var userClass models.UserClass
	result := db.MasterDB.Where("user_id = ? AND class_id = ?", userID, classID).First(&userClass)

	if result.Error == nil {
		return true
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false
	}
	return false
}

func (r *userRepository) StoreUserClass(userClass models.UserClass) error {
	if err := db.MasterDB.Create(&userClass).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateUserClass(userClass models.UserClass) error {
	if err := db.MasterDB.
		Where("user_id = ? AND class_id = ?", userClass.UserId, userClass.ClassId).
		Model(&models.UserClass{}).
		Updates(userClass).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) DeleteUserClassesByUserAndClassID(userID int64, classID int64) error {
	var userClasses []models.UserClass
	if err := db.MasterDB.Where("user_id = ? AND class_id = ?", userID, classID).
		Delete(&userClasses).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) DeleteUserCoursesByUserID(userID int64) error {
	if err := db.MasterDB.Exec("DELETE FROM user_courses WHERE user_id = ?", userID).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserCoursesByUserID(userID int64) ([]models.UserCourse, error) {
	var userCourses []models.UserCourse
	err := db.MasterDB.Where("user_id = ?", userID).Find(&userCourses).Error
	return userCourses, err
}

func (r *userRepository) DeleteUserCoursesByUserAndCourseID(userID int64, courseID int64) error {
	var userCourses []models.UserCourse
	if err := db.MasterDB.Where("user_id = ? AND course_id = ?", userID, courseID).
		Delete(&userCourses).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ExitsUserCoursesByUserAndCourseID(userID int64, courseID int64) bool {
	var userCourse models.UserCourse
	result := db.MasterDB.Where("user_id = ? AND course_id = ?", userID, courseID).First(&userCourse)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return true
	}
	return false
}

func (r *userRepository) StoreUserCourse(userCourse models.UserCourse) error {
	if err := db.MasterDB.Create(&userCourse).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateOrCreateUserCourse(userCourse models.UserCourse) error {
	var existing models.UserCourse

	err := db.MasterDB.
		Where("user_id = ? AND course_id = ?", userCourse.UserId, userCourse.CourseId).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(&userCourse).Error
		}
		return err
	}

	return nil
}

func (r *userRepository) DeleteUserRefSubjectsByUserID(userID int64) error {
	if err := db.MasterDB.Exec("DELETE FROM user_ref_subjects WHERE user_id = ?", userID).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserRefSubjectsByUserID(userID int64) ([]models.UserRefSubject, error) {
	var userSubjects []models.UserRefSubject
	err := db.MasterDB.Where("user_id = ?", userID).Find(&userSubjects).Error
	return userSubjects, err
}

func (r *userRepository) DeleteUserRefSubjectsByUserAndSubjectID(userID int64, subjectID int64) error {
	var userSubjects []models.UserRefSubject
	if err := db.MasterDB.Where("user_id = ? AND subject_id = ?", userID, subjectID).
		Delete(&userSubjects).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateOrCreateUserRefSubject(userSubject models.UserRefSubject) error {
	var existing models.UserRefSubject

	err := db.MasterDB.
		Where("user_id = ? AND subject_id = ?", userSubject.UserID, userSubject.SubjectID).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(&userSubject).Error
		}
		return err
	}

	return nil
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.MasterDB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UsernameExits(username string, id int64) (bool, error) {
	var user models.User
	query := db.MasterDB.Where("username = ?", username)

	if id > 0 {
		query = query.Where("id != ?", id)
	}

	err := query.First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *userRepository) UserExists(key string, excludeID int64) (bool, error) {
	var user models.User

	query := db.MasterDB.
		Where("(username = ? OR identifier = ? OR code = ?) AND deleted_at IS NULL", key, key, key)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *userRepository) UpdateOrCreateUserAddress(address models.UserAddress) (int64, error) {
	if address.ID > 0 {
		var existing models.UserAddress
		err := db.ReplicaDB.First(&existing, address.ID).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				result := db.MasterDB.Create(&address)
				return int64(address.ID), result.Error
			}
			return 0, err
		}

		result := db.MasterDB.Model(&existing).Updates(address)
		return int64(existing.ID), result.Error
	}

	result := db.MasterDB.Create(&address)
	return int64(address.ID), result.Error
}

func (r *userRepository) DeleteOrtherAddress(currentID, userID int64) error {
	return db.MasterDB.
		Where("user_id = ? AND id != ?", userID, currentID).
		Delete(&models.UserAddress{}).
		Error
}

func (r *userRepository) DeleteCertificateByUserID(userID int64) error {
	return db.MasterDB.Where("user_id = ?, userID").First(&models.Certificate{}).Error
}

func (r *userRepository) ExitsCertificate(certificate models.Certificate) bool {
	if certificate.ID == 0 {
		return false
	}
	var existing models.Certificate

	err := db.MasterDB.
		Where("id = ?", certificate.ID).
		First(&existing).Error

	return err == nil
}

func (r *userRepository) CreateCertificate(certificate models.Certificate) (int64, error) {
	certificate.CreatedAt = time.Now()
	certificate.UpdatedAt = time.Now()

	err := db.MasterDB.Create(&certificate).Error
	return certificate.ID, err
}

func (r *userRepository) UpdateCertificate(certificate models.Certificate) (int64, error) {
	certificate.UpdatedAt = time.Now()
	err := db.MasterDB.Updates(certificate).Error
	return certificate.ID, err
}

func (r *userRepository) DeleteNotExitCertificate(userID int64, exitsIds []int64) error {
	query := db.MasterDB.Unscoped().Where("user_id = ?", userID)

	if len(exitsIds) > 0 {
		query = query.Where("id NOT IN ?", exitsIds)
	}

	return query.Delete(&models.Certificate{}).Error
}

func (r *userRepository) DeleteDegreeByUserID(userID int64) error {
	return db.MasterDB.Where("user_id = ?", userID).First(&models.Degree{}).Error
}

func (r *userRepository) ExitsDegree(degree models.Degree) bool {
	if degree.ID == 0 {
		return false
	}
	var existing models.Degree

	err := db.MasterDB.
		Where("id = ?", degree.ID).
		First(&existing).Error

	return err == nil
}

func (r *userRepository) CreateDegree(degree models.Degree) (int64, error) {
	degree.CreatedAt = time.Now()
	degree.UpdatedAt = time.Now()

	err := db.MasterDB.Create(&degree).Error
	return degree.ID, err
}

func (r *userRepository) UpdateDegree(degree models.Degree) (int64, error) {
	degree.UpdatedAt = time.Now()
	err := db.MasterDB.Updates(degree).Error
	return degree.ID, err
}

func (r *userRepository) DeleteNotExitDegree(userID int64, exitsIds []int64) error {
	query := db.MasterDB.Unscoped().Where("user_id = ?", userID)

	if len(exitsIds) > 0 {
		query = query.Where("id NOT IN ?", exitsIds)
	}

	return query.Delete(&models.Degree{}).Error
}

func (r *userRepository) DeleteUserDepartmentByUserID(userID int64) error {
	return db.MasterDB.Where("user_id = ?", userID).First(&models.UserDepartment{}).Error
}

func (r *userRepository) ExistUserDepartment(department models.UserDepartment) (int64, bool) {
	var existing models.UserDepartment

	err := db.MasterDB.
		Where("user_id = ? AND department_id = ?", department.UserId, department.DepartmentId).
		First(&existing).Error

	return existing.ID, err == nil
}

func (r *userRepository) CreateUserDepartment(department models.UserDepartment) (int64, error) {
	err := db.MasterDB.Create(&department).Error
	return department.ID, err
}

func (r *userRepository) UpdateUserDepartment(department models.UserDepartment) (int64, error) {
	err := db.MasterDB.Updates(department).Error
	return department.ID, err
}

func (r *userRepository) DeleteUserDepartmentNotIn(userID int64, exitsIds []int64) error {
	query := db.MasterDB.Unscoped().Where("user_id = ?", userID)

	if len(exitsIds) > 0 {
		query = query.Where("id NOT IN ?", exitsIds)
	}

	return query.Delete(&models.UserDepartment{}).Error
}

func (r *userRepository) DeleteUserPositionByUserID(userID int64) error {
	return db.MasterDB.Where("user_id = ?", userID).First(&models.UserPosition{}).Error
}

func (r *userRepository) ExistUserPosition(position models.UserPosition) (int64, bool) {
	var existing models.UserPosition

	err := db.MasterDB.
		Where("user_id = ? AND employee_position_id = ?", position.UserId, position.EmployeePositionId).
		First(&existing).Error

	return existing.ID, err == nil
}

func (r *userRepository) CreateUserPosition(position models.UserPosition) (int64, error) {
	err := db.MasterDB.Create(&position).Error
	return position.ID, err
}

func (r *userRepository) UpdateUserPositiont(position models.UserPosition) (int64, error) {
	err := db.MasterDB.Updates(position).Error
	return position.ID, err
}

func (r *userRepository) DeleteUserPositionNotIn(userID int64, exitsIds []int64) error {
	query := db.MasterDB.Unscoped().Where("user_id = ?", userID)

	if len(exitsIds) > 0 {
		query = query.Where("id NOT IN ?", exitsIds)
	}

	return query.Delete(&models.UserPosition{}).Error
}

func (r *userRepository) GetAllUserClassByUser(id int64) ([]models.UserClass, error) {
	var users []models.UserClass
	err := db.ReplicaDB.Where("user_id = ?", id).Find(&users).Error
	return users, err
}

func (r *userRepository) GetAllUserClasses() ([]models.UserClass, error) {
	var users []models.UserClass
	err := db.ReplicaDB.Find(&users).Error
	return users, err
}

func (r *userRepository) GetAllUserCourses() ([]models.UserCourse, error) {
	var users []models.UserCourse
	err := db.ReplicaDB.Find(&users).Error
	return users, err
}

func (r *userRepository) CreateOrUpdateUser(user *models.User) (int64, error) {
	var existing models.User
	var err error

	if user.ID > 0 {
		err = db.MasterDB.First(&existing, user.ID).Error
	} else {
		query := db.MasterDB.Model(&models.User{})
		query = query.Where("username = ?", user.Username)

		if user.Identifier != "" {
			query = query.Or("identifier = ?", user.Identifier)
		}

		if user.Code != "" {
			query = query.Or("code = ?", user.Code)
		}

		err = query.First(&existing).Error
	}

	if err == nil {
		existing.Username = user.Username
		existing.Identifier = user.Identifier
		existing.Name = user.Name
		existing.Email = user.Email
		existing.PhoneNumber = user.PhoneNumber
		existing.AvatarInfo = user.AvatarInfo
		existing.Status = user.Status
		existing.SchoolID = user.SchoolID
		existing.Code = user.Code
		if user.Password != "" {
			existing.Password = user.Password
		}
		if saveErr := db.MasterDB.Save(&existing).Error; saveErr != nil {
			return 0, saveErr
		}
		return int64(existing.ID), nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if createErr := db.MasterDB.Create(user).Error; createErr != nil {
			return 0, createErr
		}
		return int64(user.ID), nil
	}

	return 0, err
}

func (r *userRepository) PermanentlyDeleteOldRecords(before time.Time) error {
	var model models.User

	var deletedUserIDs []int64

	if err := db.MasterDB.
		Model(&models.User{}).
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Pluck("id", &deletedUserIDs).Error; err != nil {
		return err
	}

	if len(deletedUserIDs) > 0 {
		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserAddress{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserCourse{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserClass{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.Degree{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.Certificate{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserPosition{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserDepartment{}).Error; err != nil {
			return err
		}

		if err := db.MasterDB.
			Where("user_id IN (?)", deletedUserIDs).
			Delete(&models.UserRefSubject{}).Error; err != nil {
			return err
		}
	}

	return db.MasterDB.
		Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Delete(&model).Error
}

func (r *userRepository) GetClass(userID int64) (models.Class, error) {
	var class models.Class

	err := db.ReplicaDB.
		Joins("JOIN user_classes ON user_classes.class_id = classes.id").
		Where("user_classes.user_id = ?", userID).
		First(&class).Error

	if err != nil {
		return models.Class{}, err
	}

	return class, nil
}

func (r *userRepository) GetCourses(userID int64) ([]models.Course, error) {
	var courses []models.Course

	err := db.ReplicaDB.
		Joins("JOIN user_courses ON user_courses.course_id = courses.id").
		Where("user_courses.user_id = ?", userID).
		Find(&courses).Error

	if err != nil {
		return nil, err
	}

	return courses, nil
}

func (r *userRepository) GetClassesByUserId(userID int64) ([]models.Class, error) {
	var classes []models.Class

	err := db.ReplicaDB.
		Joins("JOIN user_classes ON user_classes.class_id = classes.id").
		Where("user_classes.user_id = ?", userID).
		Find(&classes).Error

	if err != nil {
		return nil, err
	}

	return classes, nil
}

func (r *userRepository) GetStudentsByClassID(classID int64) ([]models.User, error) {
	var users []models.User

	err := db.ReplicaDB.
		Table("users").
		Joins("JOIN user_classes ON user_classes.user_id = users.id").
		Where("user_classes.class_id = ?", classID).
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Where("urr.role_id = ?", models.StudentRoleId).
		Distinct("users.*").
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetActivityLogByUserIds(userIDs []int64, start, end time.Time) (map[int64]models.ActivityLog, error) {
	if len(userIDs) == 0 {
		return map[int64]models.ActivityLog{}, nil
	}

	var logs []models.ActivityLog

	// Lấy danh sách bảng theo tháng
	tableNames := table_manager.GetTableNamesForDateRange("activity_logs", start, end)

	// Tạo UNION ALL cho các bảng tháng
	var unionParts []string
	for _, tableName := range tableNames {
		// Kiểm tra bảng có tồn tại không
		var exists bool
		err := db.ReplicaDB.Raw(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = ?
			)
		`, tableName).Scan(&exists).Error

		if err != nil || !exists {
			continue
		}

		unionParts = append(unionParts, fmt.Sprintf("SELECT * FROM %s", tableName))
	}

	// Nếu không có bảng nào tồn tại, trả về map rỗng
	if len(unionParts) == 0 {
		return map[int64]models.ActivityLog{}, nil
	}

	// Tạo virtual table từ UNION các bảng tháng
	allActivityLogs := fmt.Sprintf("(%s)", strings.Join(unionParts, " UNION ALL "))

	query := fmt.Sprintf(`
		SELECT DISTINCT ON (user_id) *
		FROM %s
		WHERE user_id IN ?
		ORDER BY user_id, created_at DESC
	`, allActivityLogs)

	err := db.ReplicaDB.Raw(query, userIDs).Scan(&logs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]models.ActivityLog)
	for _, log := range logs {
		if log.UserID != nil {
			result[int64(*log.UserID)] = log
		}
	}

	return result, nil
}

func (r *userRepository) BeforeQuery(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	// Handle nil context (when called from non-HTTP context like services)
	if ctx == nil || ctx.Request == nil {
		return query
	}

	reqSchoolIDStr := ctx.Query("school_id")
	var finalSchoolID int

	if reqSchoolIDStr != "" {
		if id, err := strconv.Atoi(reqSchoolIDStr); err == nil {
			finalSchoolID = id
		}
	} else {
		finalSchoolID = r.GetAdminSchoolId(ctx)
	}

	if finalSchoolID <= 0 {
		return query
	}

	// Filter theo school_id
	query = query.
		Joins(`LEFT JOIN user_classes uc ON uc.user_id = users.id`).
		Joins(`LEFT JOIN classes c ON c.id = uc.class_id`).
		Where(`users.school_id = ? OR c.school_id = ?`, finalSchoolID, finalSchoolID).
		Group("users.id")

	return query
}

func (r *userRepository) DeleteOldClassCourseRole(userIds []int64) error {
	db.MasterDB.
		Where("user_id IN (?)", userIds).
		Delete(&models.UserClass{})

	db.MasterDB.
		Where("user_id IN (?)", userIds).
		Delete(&models.UserCourse{})

	return nil
}

func (r *userRepository) UserActivated(req requests.UserActivatedRequest) ([]*models.User, int64, error) {
	date := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	var activityIds models.ActivityIds
	now := time.Now().UTC()

	if int(now.Month()) == req.Month && now.Year() == req.Year {
		historyRepo := NewHistoryUseRepository()
		activity, err := historyRepo.ActivityIds(date, now)

		if err != nil {
			return nil, 0, err
		}
		activityIds = activity
	} else {
		var historyUse models.HistoryUse
		err := db.ReplicaDB.
			Where("date = ? AND type = ?", date, "month").
			First(&historyUse).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				activityIds = models.ActivityIds{}
			} else {
				return nil, 0, err
			}
		} else {
			if err := json.Unmarshal(historyUse.ActivityIds, &activityIds); err != nil {
				return nil, 0, err
			}
		}
	}

	query := db.ReplicaDB.
		Table("users u").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.id IN ?", activityIds.Ids)

	if req.RoleId != 0 {
		query = query.Where("urr.role_id = ?", req.RoleId)
	}

	// Đếm tổng
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Lấy danh sách user có phân trang
	offset := (req.Page - 1) * req.Limit
	var users []*models.User
	if err := query.Limit(req.Limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) ResetPassword(id int, password string) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := db.MasterDB.
		Model(&models.User{}).
		Where("id = ?", id).
		Update("password", string(hashedPass)).
		Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepository) DeleteUserRefRolesByUserID(userID int64) error {
	if err := db.MasterDB.Exec("DELETE FROM user_ref_roles WHERE user_id = ?", userID).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserRefRolesByUserID(userID int64) ([]models.UserRefRole, error) {
	var userRoles []models.UserRefRole
	err := db.MasterDB.Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

func (r *userRepository) DeleteUserRefRolesByUserAndRoleID(userID int64, roleID int64) error {
	var userRoles []models.UserRefRole
	if err := db.MasterDB.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&userRoles).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) UpdateOrCreateUserRefRole(userRole models.UserRefRole) error {
	var existing models.UserRefRole

	err := db.MasterDB.
		Where("user_id = ? AND role_id = ?", userRole.UserId, userRole.RoleId).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.MasterDB.Create(&userRole).Error
		}
		return err
	}

	return nil
}

func (r *userRepository) GetIdsByProgramStatus(notParticipated, failTheSubject, isFilterNotParticipated, isFilterFailTheSubject bool, programId int64) ([]int64, error) {
	// Get all courses in the program
	var courseIDs []int64
	if err := db.ReplicaDB.Model(&models.Course{}).
		Where("program_id = ?", programId).
		Pluck("id", &courseIDs).Error; err != nil {
		return nil, err
	}

	if len(courseIDs) == 0 {
		return []int64{}, nil
	}

	var resultUserIds []int64

	// Case 1: Only filter by not_participated
	if isFilterNotParticipated && !isFilterFailTheSubject {
		if notParticipated {
			// Users NOT in any course of the program
			// Get all participated users first, then we'll exclude them
			var participatedUserIds []int64
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ?", courseIDs).
				Distinct("user_id").
				Pluck("user_id", &participatedUserIds)

			// Return participated IDs to be excluded (will be handled in service layer)
			// or return empty to indicate these should be excluded
			// For now, return the participated IDs with a marker
			return participatedUserIds, nil
		} else {
			// Users participated in at least one course of the program
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ?", courseIDs).
				Distinct("user_id").
				Pluck("user_id", &resultUserIds)
			return resultUserIds, nil
		}
	}

	// Case 2: Only filter by fail_the_subject
	if !isFilterNotParticipated && isFilterFailTheSubject {
		if failTheSubject {
			// Users who failed in the program
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ? AND is_failed = ?", courseIDs, true).
				Distinct("user_id").
				Pluck("user_id", &resultUserIds)
			return resultUserIds, nil
		} else {
			// Users participated but NOT failed
			// Get all participated users
			var participatedUserIds []int64
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ?", courseIDs).
				Distinct("user_id").
				Pluck("user_id", &participatedUserIds)

			// Get failed users
			var failedUserIds []int64
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ? AND is_failed = ?", courseIDs, true).
				Distinct("user_id").
				Pluck("user_id", &failedUserIds)

			// Filter out failed users from participated users
			failedMap := make(map[int64]bool)
			for _, id := range failedUserIds {
				failedMap[id] = true
			}

			for _, id := range participatedUserIds {
				if !failedMap[id] {
					resultUserIds = append(resultUserIds, id)
				}
			}
			return resultUserIds, nil
		}
	}

	// Case 3: Both filters are active
	if isFilterNotParticipated && isFilterFailTheSubject {
		if notParticipated {
			// Users NOT participated - fail_the_subject is irrelevant
			// Return participated users to be excluded
			var participatedUserIds []int64
			db.ReplicaDB.Table("user_courses").
				Where("course_id IN ?", courseIDs).
				Distinct("user_id").
				Pluck("user_id", &participatedUserIds)
			return participatedUserIds, nil
		} else {
			// Users participated
			if failTheSubject {
				// Users who failed
				db.ReplicaDB.Table("user_courses").
					Where("course_id IN ? AND is_failed = ?", courseIDs, true).
					Distinct("user_id").
					Pluck("user_id", &resultUserIds)
				return resultUserIds, nil
			} else {
				// Users participated but NOT failed
				var participatedUserIds []int64
				db.ReplicaDB.Table("user_courses").
					Where("course_id IN ?", courseIDs).
					Distinct("user_id").
					Pluck("user_id", &participatedUserIds)

				var failedUserIds []int64
				db.ReplicaDB.Table("user_courses").
					Where("course_id IN ? AND is_failed = ?", courseIDs, true).
					Distinct("user_id").
					Pluck("user_id", &failedUserIds)

				failedMap := make(map[int64]bool)
				for _, id := range failedUserIds {
					failedMap[id] = true
				}

				for _, id := range participatedUserIds {
					if !failedMap[id] {
						resultUserIds = append(resultUserIds, id)
					}
				}
				return resultUserIds, nil
			}
		}
	}

	// No filter applied
	return []int64{}, nil
}
