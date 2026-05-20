package services

import (
	"be-Clever School/config"
	"be-Clever School/dto"
	"be-Clever School/i18n"
	"be-Clever School/models"
	"be-Clever School/prot"
	"be-Clever School/repositories"
	"be-Clever School/requests"
	"be-Clever School/resources"
	"be-Clever School/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAll(c *gin.Context) ([]models.User, int64, error)
	GetByID(c *gin.Context, id int) (*prot.User, error)
	Create(c *gin.Context, req *prot.User) (*models.User, error)
	Update(c *gin.Context, req *prot.User) (*models.User, error)
	UpdateProfile(c *gin.Context, req *prot.User) (*models.User, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.User, error)
	UsernameExits(c *gin.Context, username string) (bool, error)
	GetAllStudentsByParentID(parentID int) ([]*prot.User, error)
	CurrentClass(c *gin.Context) (*dto.Class, error)
	Export(c *gin.Context) (string, error)
	ExportUsersPDF(schoolID, classID int) ([]byte, string, error)
	Import(c *gin.Context, fileHeader *multipart.FileHeader) error
	UserActivated(c *gin.Context) ([]*models.User, int64, error)
	ResetPassword(c *gin.Context, id int) (error, string)
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAllStudentsByParentID(parentID int) ([]*prot.User, error) {
	users, err := s.repo.GetStudentsByParentID(parentID)
	if err != nil {
		return nil, err
	}

	var res []*prot.User
	for _, u := range users {
		res = append(res, modelToProtoUser(&u))
	}
	return res, nil
}

func (s *userService) GetAll(c *gin.Context) ([]models.User, int64, error) {
	allowedFilters := []string{
		"parent_id",
		// "school_id",
		"status",
		// "is_independent_student",
		// "is_failed_subject",
	}

	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	filter, _ = s.ApplyFilter(c, filter)

	// Qualify status filter with table name to avoid ambiguity
	if statusVal, exists := filter["status"]; exists {
		delete(filter, "status")
		filter["users.status"] = statusVal
	}

	s.repo.SetSearch(keyword, []string{
		"username",
		"name",
		"email",
		"phone_number",
		"code",
		"identifier",
		"id",
		// "user_classes:user_id:class_id:id:id:classes.name",
		// "user_address:user_id:id:user_address.address",
	})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	// s.repo.SetEnableRedis(true)
	s.repo.SetSort(sort)
	s.repo.SetContext(c)

	s.repo.SetPreload([]string{
		"Roles",
		"School",
		"UserAddress",
		"UserClasses.Class",
		"UserClasses.Class.School",
		"UserCourses.Course.Program",
		"UserClasses.Class.Faculty",
		"UserCourses.Course",
		"Subjects",
	})

	users, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return users, rows, nil
}

func (s *userService) GetByID(c *gin.Context, id int) (*prot.User, error) {
	s.repo.SetContext(c)
	s.repo.SetPreload([]string{
		"Roles",
		"School",
		"UserAddress",
		"Certificates",
		"Degrees",
		"Departments",
		"Positions",
		"UserClasses.Class",
		"UserClasses.Class.School",
		"UserCourses.Course",
		"UserCourses.Course.Program",
		"UserClasses.Class.Faculty",
		"Subjects",
	})

	user, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	userResource := resources.NewUserResource()
	formattedUser := userResource.FormatUser(user)

	return formattedUser, nil
}

func (s *userService) Create(c *gin.Context, req *prot.User) (*models.User, error) {
	key := req.Username

	if key == "" && req.Identifier != "" {
		key = req.Identifier
	} else if key == "" && req.Code != "" {
		key = req.Code
	}

	if key == "" {
		return nil, fmt.Errorf(i18n.Localize("messages.user_key_required"))
	}

	existingUser, _ := s.repo.UserExists(key, int64(0))

	if existingUser {
		return nil, fmt.Errorf(i18n.Localize("messages.account_exist_by_key", map[string]interface{}{"Key": key}))
	}

	userResource := resources.NewUserResource()
	user := userResource.FormatModelUser(req)

	if req.UserInfo == nil {
		if req.RoleId == 0 {
			return nil, fmt.Errorf(i18n.Localize("messages.role_invalid", map[string]interface{}{"role_id": 0}))
		}
	} else {
		if len(req.UserInfo.Roles) == 0 && (req.UserInfo.Role == nil || req.UserInfo.Role.Id == 0) && req.RoleId == 0 {
			return nil, fmt.Errorf(i18n.Localize("messages.role_invalid", map[string]interface{}{"role_id": 0}))
		}
	}

	password := req.Password
	if password == "" {
		password = req.RoleName
	}

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf(i18n.Localize("messages.create_data"))
		}
		user.Password = string(hashedPassword)
	}

	s.repo.SetContext(c)

	err := s.repo.Create(user)
	if err != nil {
		config.Log.Error(err)
		return nil, fmt.Errorf(i18n.Localize("messages.create_data"))
	}

	s.StoreUserClasses(int64(user.ID), req)
	s.StoreUserCourses(int64(user.ID), req)
	s.StoreUserAddress(int64(user.ID), req)
	s.StoreUserCertificates(int64(user.ID), req)
	s.StoreUserDegrees(int64(user.ID), req)
	s.StoreUserDepartments(int64(user.ID), req)
	s.StoreUserPositions(int64(user.ID), req)
	s.StoreUserSubjects(int64(user.ID), req)
	s.StoreUserRoles(int64(user.ID), req)

	id := int(user.ID)
	s.repo.SetPreload([]string{
		"Roles",
		"School",
		"UserAddress",
		"Certificates",
		"Degrees",
		"Departments",
		"Positions",
		"UserClasses.Class",
		"UserClasses.Class.School",
		"UserCourses.Course.Program",
		"UserClasses.Class.Faculty",
		"UserCourses.Course",
		"Subjects",
	})
	newUser, _ := s.repo.FindNewByID(id)

	return newUser, nil
}

func (s *userService) Update(c *gin.Context, req *prot.User) (*models.User, error) {
	id, _ := strconv.Atoi(c.Param("id"))
	req.Id = int64(id)
	return s.UpdateUser(c, req, true)
}

func (s *userService) UpdateProfile(c *gin.Context, req *prot.User) (*models.User, error) {
	return s.UpdateUser(c, req, false)
}

func (s *userService) Delete(c *gin.Context, id int) error {
	s.repo.SetContext(c)

	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	classes, _ := s.repo.GetClassesByUserId(int64(id))
	classRepo := repositories.NewClassRepository()

	for _, class := range classes {
		classRepo.UpdateCurrentStudentToClass(class.ID)
	}

	courses, _ := s.repo.GetCourses(int64(id))
	courseRepo := repositories.NewCourseRepository()

	for _, course := range courses {
		courseRepo.UpdateCurrentStudentToCourse(course.ID)
	}

	return nil
}

func (s *userService) Restore(c *gin.Context, id int) (*models.User, error) {
	s.repo.SetContext(c)
	user, err := s.repo.Restore(int(id))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) UsernameExits(c *gin.Context, username string) (bool, error) {
	existingUser, err := s.repo.UsernameExits(username, int64(0))
	return existingUser, err
}

func (s *userService) CurrentClass(c *gin.Context) (*dto.Class, error) {
	userId := utils.GetCurrentUserId(c)

	class, err := s.repo.GetClass(int64(userId))

	if err != nil {
		return nil, err
	}

	students, err := s.repo.GetStudentsByClassID(class.ID)

	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(students))
	for _, u := range students {
		userIDs = append(userIDs, u.ID)
	}

	// Sử dụng thời gian hiện tại để lấy activity logs
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	activityLogByUserIds, err := s.repo.GetActivityLogByUserIds(userIDs, startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	classRepo := resources.NewClassResource()
	dtoStudents := classRepo.MapUsersToDTOStudents(students, activityLogByUserIds, int64(userId))

	courses, _ := s.repo.GetCourses(int64(userId))
	dtoCourses := classRepo.MapCoursesToDTOCourses(courses)

	return &dto.Class{
		Class:    class,
		Students: dtoStudents,
		Courses:  dtoCourses,
	}, err
}

func modelToProtoUser(u *models.User) *prot.User {
	var roleName string
	var roleId int64

	if len(u.Roles) > 0 {
		roleName = u.Roles[0].Name
		roleId = int64(u.Roles[0].ID)
	}

	return &prot.User{
		Id:          int64(u.ID),
		Username:    u.Username,
		Name:        u.Name,
		RoleId:      roleId,
		Address:     u.Address,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Status:      u.Status,
		RoleName:    roleName,
	}
}

func (s *userService) StoreUserClasses(userID int64, user *prot.User) {
	userClasses := make([]models.UserClass, 0)

	allUserClasses, _ := s.repo.GetAllUserClassByUser(userID)

	if len(user.UserInfo.Classes) > 0 {
		for _, uc := range user.UserInfo.Classes {
			userClasses = append(userClasses, models.UserClass{
				UserId:    int64(userID),
				ClassId:   int64(uc.Id),
				IsCurrent: true,
			})
		}

		allUserClasses = append(allUserClasses, userClasses...)
	} else if user.UserInfo.Class != nil && user.UserInfo.Class.Id != 0 {
		userClasses = append(userClasses, models.UserClass{
			UserId:    int64(userID),
			ClassId:   int64(user.UserInfo.Class.Id),
			IsCurrent: true,
		})

		allUserClasses = append(allUserClasses, userClasses...)
	}

	if len(userClasses) == 0 {
		s.repo.DeleteUserClassesByUserID(userID)
	} else {
		classIDs := make(map[uint]bool)

		for _, uc := range userClasses {
			classIDs[uint(uc.ClassId)] = true
		}

		existingUserClasses, _ := s.repo.GetUserClassesByUserID(userID)

		for _, euc := range existingUserClasses {
			if !classIDs[uint(euc.ClassId)] {
				s.repo.DeleteUserClassesByUserAndClassID(userID, int64(euc.ClassId))
			}
		}

		for _, uc := range userClasses {
			exits := s.repo.ExitsUserClassesByUserAndClassID(userID, int64(uc.ClassId))

			if exits {
				s.repo.UpdateUserClass(uc)
			} else {
				s.repo.StoreUserClass(uc)
			}
		}
	}

	go func() {
		classRepo := repositories.NewClassRepository()

		for _, uc := range allUserClasses {
			classRepo.UpdateCurrentStudentToClass(uc.ClassId)
		}
	}()
}

func (s *userService) StoreUserCourses(userID int64, user *prot.User) {
	userCourses := make([]models.UserCourse, 0)

	alllUserCourses, _ := s.repo.GetAllUserCourses()

	for _, course := range user.UserInfo.Courses {
		userCourses = append(userCourses, models.UserCourse{
			UserId:    int64(userID),
			CourseId:  int64(course.Id),
			IsCurrent: true,
		})

		alllUserCourses = append(alllUserCourses, userCourses...)
	}

	if len(userCourses) == 0 {
		s.repo.DeleteUserCoursesByUserID(userID)
	} else {
		courseIDs := make(map[uint]bool)

		for _, uc := range userCourses {
			courseIDs[uint(uc.CourseId)] = true
		}

		existingUserCourses, _ := s.repo.GetUserCoursesByUserID(userID)

		for _, euc := range existingUserCourses {
			if !courseIDs[uint(euc.CourseId)] {
				s.repo.DeleteUserCoursesByUserAndCourseID(userID, int64(euc.CourseId))
			}
		}

		for _, uc := range userCourses {
			s.repo.UpdateOrCreateUserCourse(uc)
		}
	}

	go func() {
		courseRepo := repositories.NewCourseRepository()

		for _, course := range alllUserCourses {
			courseRepo.UpdateCurrentStudentToCourse(course.CourseId)
		}
	}()
}

func (s *userService) StoreUserSubjects(userID int64, user *prot.User) {
	userSubjects := make([]models.UserRefSubject, 0)

	for _, subject := range user.UserInfo.Subjects {
		userSubjects = append(userSubjects, models.UserRefSubject{
			UserID:    int64(userID),
			SubjectID: int64(subject.Id),
		})
	}

	if len(userSubjects) == 0 {
		s.repo.DeleteUserRefSubjectsByUserID(userID)
	} else {
		subjectIDs := make(map[uint]bool)

		for _, us := range userSubjects {
			subjectIDs[uint(us.SubjectID)] = true
		}

		existingUserSubjects, _ := s.repo.GetUserRefSubjectsByUserID(userID)

		for _, eus := range existingUserSubjects {
			if !subjectIDs[uint(eus.SubjectID)] {
				s.repo.DeleteUserRefSubjectsByUserAndSubjectID(userID, int64(eus.SubjectID))
			}
		}

		for _, us := range userSubjects {
			s.repo.UpdateOrCreateUserRefSubject(us)
		}
	}
}

func (s *userService) StoreUserRoles(userID int64, user *prot.User) {
	userRoles := make([]models.UserRefRole, 0)

	for _, role := range user.UserInfo.Roles {
		userRoles = append(userRoles, models.UserRefRole{
			UserId: int64(userID),
			RoleId: int64(role.Id),
		})
	}

	if len(userRoles) == 0 {
		if user.RoleId != 0 {
			userRoles = append(userRoles, models.UserRefRole{
				UserId: int64(userID),
				RoleId: int64(user.RoleId),
			})
		} else if user.UserInfo.Role.Id != 0 {
			userRoles = append(userRoles, models.UserRefRole{
				UserId: int64(userID),
				RoleId: int64(user.UserInfo.Role.Id),
			})
		}
	}

	if len(userRoles) == 0 {
		s.repo.DeleteUserRefRolesByUserID(userID)
	} else {
		roleIDs := make(map[uint]bool)

		for _, ur := range userRoles {
			roleIDs[uint(ur.RoleId)] = true
		}

		existingUserRoles, _ := s.repo.GetUserRefRolesByUserID(userID)

		for _, eur := range existingUserRoles {
			if !roleIDs[uint(eur.RoleId)] {
				s.repo.DeleteUserRefRolesByUserAndRoleID(userID, int64(eur.RoleId))
			}
		}

		for _, ur := range userRoles {
			s.repo.UpdateOrCreateUserRefRole(ur)
		}
	}
}

func (s *userService) StoreUserAddress(userID int64, user *prot.User) {
	if user.UserInfo.Address != nil {
		wardRepo := repositories.NewWardRepository()
		provinceRepo := repositories.NewProvinceRepository()

		var ward *models.Ward
		var province *models.Province
		var err error

		ward, err = wardRepo.GetByCode(user.UserInfo.Address.WardCode)
		if err != nil {
			ward = nil
		}

		if ward != nil {
			province, err = provinceRepo.GetByCode(ward.ProvinceCode)
			if err != nil {
				province = nil
			}
		} else {
			province, err = provinceRepo.GetByCode(user.UserInfo.Address.ProvinceCode)
			if err != nil {
				province = nil
			}
		}

		userAddress := models.UserAddress{
			ID:             int(user.UserInfo.Address.Id),
			UserId:         int(userID),
			ProvinceCode:   safeString(province, func(p *models.Province) string { return p.Code }),
			WardCode:       safeString(ward, func(w *models.Ward) string { return w.Code }),
			ProvinceName:   safeString(province, func(p *models.Province) string { return p.FullName }),
			WardName:       safeString(ward, func(w *models.Ward) string { return w.FullName }),
			ProvinceNameEn: safeString(province, func(p *models.Province) string { return p.FullNameEn }),
			WardNameEn:     safeString(ward, func(w *models.Ward) string { return w.FullNameEn }),
		}

		// Lưu địa chỉ theo ngôn ngữ người dùng truyền (X-Locale), hỗ trợ mở rộng nhiều field tái sử dụng
		addr := user.UserInfo.Address.Address
		_ = i18n.SetFieldsWithProviderFallback(&userAddress, map[string]string{
			"address": addr,
		}, i18n.GlobalLocaleProvider{})

		id, err := s.repo.UpdateOrCreateUserAddress(userAddress)

		if err != nil {
			fmt.Println(err)
		} else {
			s.repo.DeleteOrtherAddress(id, userID)
		}
	}
}

func (s *userService) StoreUserCertificates(userID int64, user *prot.User) {
	if user == nil || user.UserInfo == nil {
		return
	}

	userCertificates := make([]models.Certificate, 0)

	for _, certificate := range user.UserInfo.Certificates {
		var issuedDate, expiryDate time.Time
		var err error

		if certificate.IssuedDate != "" {
			issuedDate, err = time.Parse(time.RFC3339, certificate.IssuedDate)
			if err != nil {
				issuedDate = time.Time{}
			}
		}

		if certificate.ExpiryDate != "" {
			expiryDate, err = time.Parse(time.RFC3339, certificate.ExpiryDate)
			if err != nil {
				expiryDate = time.Time{}
			}
		}

		fileUrl := utils.StripDomain(certificate.FileUrl, models.Storage)
		mediaRepo := repositories.NewMediaRepository()
		fileInfo := mediaRepo.GetMediaInfo(fileUrl, models.Storage)

		userCertificates = append(userCertificates, models.Certificate{
			ID:              certificate.Id,
			UserId:          userID,
			Name:            certificate.Name,
			IssuedBy:        certificate.IssuedBy,
			IssuedDate:      issuedDate,
			ExpiryDate:      expiryDate,
			CertificateCode: certificate.CertificateCode,
			Description:     certificate.Description,
			FileInfo:        fileInfo,
			Status:          certificate.Status,
		})
	}

	if len(userCertificates) == 0 {
		_ = s.repo.DeleteCertificateByUserID(userID)
		return
	}

	certificateIDs := make([]int64, 0)

	for _, uc := range userCertificates {
		exits := s.repo.ExitsCertificate(uc)

		if exits {
			id, err := s.repo.UpdateCertificate(uc)

			if err == nil {
				certificateIDs = append(certificateIDs, id)
			} else {
				fmt.Println(err)
			}
		} else {
			id, err := s.repo.CreateCertificate(uc)

			if err == nil {
				certificateIDs = append(certificateIDs, id)
			} else {
				fmt.Println(err)
			}
		}
	}

	if len(certificateIDs) > 0 {
		_ = s.repo.DeleteNotExitCertificate(userID, certificateIDs)
	}
}

func (s *userService) StoreUserDegrees(userID int64, user *prot.User) {
	if user == nil || user.UserInfo == nil {
		return
	}

	userDegrees := make([]models.Degree, 0)

	for _, degree := range user.UserInfo.Degrees {
		var receivedDate time.Time
		var err error

		if degree.ReceivedDate != "" {
			receivedDate, err = time.Parse(time.RFC3339, degree.ReceivedDate)
			if err != nil {
				receivedDate = time.Time{}
			}
		}

		userDegrees = append(userDegrees, models.Degree{
			ID:             degree.Id,
			UserId:         userID,
			Name:           degree.Name,
			Major:          degree.Major,
			ReceivedDate:   receivedDate,
			Institution:    degree.Institution,
			GraduationYear: int(degree.GraduationYear),
			DegreeLevel:    degree.DegreeLevel,
			Note:           degree.Note,
		})
	}

	if len(userDegrees) == 0 {
		_ = s.repo.DeleteDegreeByUserID(userID)
		return
	}

	degreeIDs := make([]int64, 0)

	for _, ud := range userDegrees {
		exits := s.repo.ExitsDegree(ud)

		if exits {
			id, err := s.repo.UpdateDegree(ud)

			if err == nil {
				degreeIDs = append(degreeIDs, id)
			} else {
				fmt.Println(err)
			}
		} else {
			id, err := s.repo.CreateDegree(ud)

			if err == nil {
				degreeIDs = append(degreeIDs, id)
			} else {
				fmt.Println(err)
			}
		}
	}

	if len(degreeIDs) > 0 {
		_ = s.repo.DeleteNotExitDegree(userID, degreeIDs)
	}
}

func (s *userService) StoreUserDepartments(userID int64, user *prot.User) {
	if user == nil || user.UserInfo == nil {
		return
	}

	userDepartments := make([]models.UserDepartment, 0)

	for _, department := range user.UserInfo.Departments {
		if department.Id > 0 {
			userDepartments = append(userDepartments, models.UserDepartment{
				UserId:       userID,
				DepartmentId: department.Id,
			})
		}
	}

	if len(userDepartments) == 0 {
		_ = s.repo.DeleteUserDepartmentByUserID(userID)
		return
	}

	var userDepartmentIDs []int64

	for _, ud := range userDepartments {
		existId, exists := s.repo.ExistUserDepartment(ud)
		var id int64
		var err error

		if exists {
			ud.ID = existId
			id, err = s.repo.UpdateUserDepartment(ud)
		} else {
			id, err = s.repo.CreateUserDepartment(ud)
		}

		if err == nil {
			userDepartmentIDs = append(userDepartmentIDs, id)
		} else {
			fmt.Println("user_department error:", err)
		}
	}

	if len(userDepartmentIDs) > 0 {
		_ = s.repo.DeleteUserDepartmentNotIn(userID, userDepartmentIDs)
	}
}

func (s *userService) StoreUserPositions(userID int64, user *prot.User) {
	if user == nil || user.UserInfo == nil {
		return
	}

	userPositions := make([]models.UserPosition, 0)

	for _, position := range user.UserInfo.Departments {
		if position.Id > 0 {
			userPositions = append(userPositions, models.UserPosition{
				UserId:             userID,
				EmployeePositionId: position.Id,
			})
		}
	}

	if len(userPositions) == 0 {
		_ = s.repo.DeleteUserPositionByUserID(userID)
		return
	}

	var userPositionIDs []int64

	for _, up := range userPositions {
		existId, exists := s.repo.ExistUserPosition(up)
		var id int64
		var err error

		if exists {
			up.ID = existId
			id, err = s.repo.UpdateUserPositiont(up)
		} else {
			id, err = s.repo.CreateUserPosition(up)
		}

		if err == nil {
			userPositionIDs = append(userPositionIDs, id)
		} else {
			fmt.Println("user_department error:", err)
		}
	}

	if len(userPositionIDs) > 0 {
		_ = s.repo.DeleteUserPositionNotIn(userID, userPositionIDs)
	}
}

func (s *userService) UserActivated(c *gin.Context) ([]*models.User, int64, error) {
	var req requests.UserActivatedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, 0, err
	}

	now := time.Now()
	year := req.Year
	month := req.Month
	page := req.Page
	limit := req.Limit

	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 20
	}

	if year == 0 {
		year = now.Year()
	}

	if month == 0 {
		month = int(now.Month())
	}

	req.Month = month
	req.Year = year
	req.Page = page
	req.Limit = limit

	return s.repo.UserActivated(req)
}

func (s *userService) ResetPassword(c *gin.Context, id int) (error, string) {
	user, err := s.repo.FindByID(id)

	if err != nil || user.ID == 0 {
		return err, "messages.data_invalid"
	}

	password := user.Identifier

	if password == "" && user.Code != "" {
		password = user.Code
	} else if password == "" && user.Username != "" {
		password = user.Username
	}

	if password == "" {
		return errors.New(i18n.Localize("messages.data_invalid")), "messages.data_invalid"
	}

	err = s.repo.ResetPassword(id, password)

	if err != nil {
		return err, ""
	}

	return nil, ""
}

func (s *userService) UpdateUser(c *gin.Context, req *prot.User, updateRole bool) (*models.User, error) {
	userResource := resources.NewUserResource()
	user := userResource.FormatModelUser(req)

	if req.UserInfo == nil {
		if req.RoleId == 0 {
			return nil, fmt.Errorf(i18n.Localize("messages.role_invalid", map[string]interface{}{"role_id": 0}))
		}
	} else {
		if len(req.UserInfo.Roles) == 0 && (req.UserInfo.Role == nil || req.UserInfo.Role.Id == 0) && req.RoleId == 0 {
			return nil, fmt.Errorf(i18n.Localize("messages.role_invalid", map[string]interface{}{"role_id": 0}))
		}
	}

	key := req.Username

	if key == "" && req.Identifier != "" {
		key = req.Identifier
	} else if key == "" && req.Code != "" {
		key = req.Code
	}

	if key == "" {
		return nil, fmt.Errorf(i18n.Localize("messages.user_key_required"))
	}

	existingUser, _ := s.repo.UserExists(key, int64(user.ID))

	if existingUser {
		return nil, fmt.Errorf(i18n.Localize("messages.account_exist_by_key", map[string]interface{}{"Key": key}))
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf(i18n.Localize("messages.error_update_data"))
		}
		user.Password = string(hashedPassword)
	} else {
		existingUser, err := s.repo.FindByID(int(user.ID))
		if err != nil {
			return nil, fmt.Errorf(i18n.Localize("messages.error_update_data"))
		}
		user.Password = existingUser.Password
	}

	s.repo.SetContext(c)

	err := s.repo.Update(user)
	if err != nil {
		config.Log.Error(err)
		return nil, fmt.Errorf(i18n.Localize("messages.error_update_data"))
	}

	s.StoreUserClasses(int64(user.ID), req)
	s.StoreUserCourses(int64(user.ID), req)
	s.StoreUserAddress(int64(user.ID), req)
	s.StoreUserCertificates(int64(user.ID), req)
	s.StoreUserDegrees(int64(user.ID), req)
	s.StoreUserDepartments(int64(user.ID), req)
	s.StoreUserPositions(int64(user.ID), req)
	s.StoreUserSubjects(int64(user.ID), req)

	if updateRole {
		s.StoreUserRoles(int64(user.ID), req)
	}

	s.repo.SetPreload([]string{
		"Courses",
		"Classes",
		"Roles",
		"School",
		"UserAddress",
		"Certificates",
		"Degrees",
		"Departments",
		"Positions",
		"UserClasses.Class",
		"UserClasses.Class.School",
		"UserCourses.Course.Program",
		"UserClasses.Class.Faculty",
		"UserCourses.Course",
		"Subjects",
	})
	updateUser, err := s.repo.FindNewByID(int(user.ID))

	if err != nil {
		return nil, err
	}

	return updateUser, nil
}

func (s *userService) ApplyFilter(c *gin.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	if classIDStr := c.Query("class_id"); classIDStr != "" {
		filter["Classes.id"] = classIDStr + ":user_classes:users:classes:user_id:class_id:id:id"
	} else if facultyIDStr := c.Query("faculty_id"); facultyIDStr != "" {
		if facultyID, err := strconv.ParseInt(facultyIDStr, 10, 64); err == nil {
			classRepo := repositories.NewClassRepository()
			classRepo.SetContext(c)
			classRepo.SetFilter(map[string]interface{}{
				"faculty_id": facultyID,
			})
			classes, err := classRepo.GetAll()

			if err == nil && len(classes) > 0 {
				classIdStrs := make([]string, len(classes))
				for i, class := range classes {
					classIdStrs[i] = strconv.FormatInt(class.ID, 10)
				}
				filter["Classes.id"] = "in:" + strings.Join(classIdStrs, ",") + ":user_classes:users:classes:user_id:class_id:id:id"
			}
		}
	}

	if courseIDStr := c.Query("course_id"); courseIDStr != "" {
		filter["Courses.id"] = courseIDStr + ":user_courses:users:courses:user_id:course_id:id:id"
	} else if subjectIDStr := c.Query("subject_id"); subjectIDStr != "" {
		if subjectID, err := strconv.ParseInt(subjectIDStr, 10, 64); err == nil {
			courseRepo := repositories.NewCourseRepository()
			courseIds, err := courseRepo.GetCourseIdsBySubjectId(subjectID)

			if err == nil && len(courseIds) > 0 {
				courseIdStrs := make([]string, len(courseIds))
				for i, id := range courseIds {
					courseIdStrs[i] = strconv.FormatInt(id, 10)
				}
				filter["Courses.id"] = "in:" + strings.Join(courseIdStrs, ",") + ":user_courses:users:courses:user_id:course_id:id:id"
			}
		}
	} else if programIDStr := c.Query("program_id"); programIDStr != "" {
		if programID, err := strconv.ParseInt(programIDStr, 10, 64); err == nil {
			failTheSubjectStr := c.Query("fail_the_subject")
			notParticipatedStr := c.Query("not_participated")

			isFilterFailTheSubject := failTheSubjectStr != ""
			isFilterNotParticipated := notParticipatedStr != ""

			failTheSubject := failTheSubjectStr == "true"
			notParticipated := notParticipatedStr == "true"

			if isFilterNotParticipated || isFilterFailTheSubject {
				// Use repository method to get filtered user IDs
				userRepo := repositories.NewUserRepository()
				userIds, err := userRepo.GetIdsByProgramStatus(
					notParticipated,
					failTheSubject,
					isFilterNotParticipated,
					isFilterFailTheSubject,
					programID,
				)

				if err != nil {
					config.Log.Errorf("Error getting users by program status: %v", err)
					return filter, err
				}

				if isFilterNotParticipated && notParticipated {
					// Special case: not_participated = true
					// userIds contains participated users, we want to EXCLUDE them
					if len(userIds) > 0 {
						// Exclude these users using NOT IN
						userIdStrs := make([]string, len(userIds))
						for i, id := range userIds {
							userIdStrs[i] = strconv.FormatInt(id, 10)
						}
						filter["users.id"] = "not_in:" + strings.Join(userIdStrs, ",")
					}
					// If no participated users, all users are valid (no filter needed)
				} else {
					// Normal case: filter by user IDs
					if len(userIds) > 0 {
						userIdStrs := make([]string, len(userIds))
						for i, id := range userIds {
							userIdStrs[i] = strconv.FormatInt(id, 10)
						}
						filter["users.id"] = "in:" + strings.Join(userIdStrs, ",")
					} else {
						// No users match criteria, return empty result
						filter["users.id"] = "in:0"
					}
				}
			} else {
				// No specific filter, just get users in program courses
				courseRepo := repositories.NewCourseRepository()
				courseRepo.SetContext(c)
				courseRepo.SetFilter(map[string]interface{}{
					"program_id": programID,
				})
				courses, err := courseRepo.GetAll()
				if err == nil && len(courses) > 0 {
					courseIdStrs := make([]string, len(courses))
					for i, course := range courses {
						courseIdStrs[i] = strconv.FormatInt(course.ID, 10)
					}
					filter["Courses.id"] = "in:" + strings.Join(courseIdStrs, ",") + ":user_courses:users:courses:user_id:course_id:id:id"
				}
			}
		}
	}

	if wardCode := c.Query("ward_code"); wardCode != "" {
		filter["UserAddress.ward_code"] = wardCode + ":users:user_address:id:user_id:ward_code"
	} else if provinceCode := c.Query("province_code"); provinceCode != "" {
		filter["UserAddress.province_code"] = provinceCode + ":users:user_address:id:user_id:province_code"
	}

	if IsUsedParam := c.Query("is_used"); IsUsedParam != "" {
		if IsUsedParam == "true" || IsUsedParam == "1" || IsUsedParam == "yes" {
			filter["last_login_at"] = "!=:NULL"
		} else {
			filter["last_login_at"] = "==:NULL"
		}
	}

	if roleIDStr := c.Query("role_id"); roleIDStr != "" {
		filter["Roles.id"] = roleIDStr + ":user_ref_roles:users:roles:user_id:role_id:id:id"
	}

	filter = utils.ApplyFilterDate(c, filter)

	return filter, nil
}

func safeString[T any](model *T, getter func(*T) string) string {
	if model != nil {
		return getter(model)
	}
	return ""
}
