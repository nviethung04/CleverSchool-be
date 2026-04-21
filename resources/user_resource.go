package resources

import (
	"be-lms/i18n"
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/utils"
	"sort"
	"time"
)

type UserResource interface {
	FormatUser(user *models.User) *prot.User
	FormatUsers(users []*models.User) []*prot.User
	FormatModelUser(user *prot.User) *models.User
}

type UserResourceImpl struct {
	CourseId       int64
	FailedUserIds  []int64
	localeProvider i18n.LocaleProvider
}

func NewUserResource() UserResource {
	return &UserResourceImpl{
		CourseId:      0,
		FailedUserIds: []int64{},
	}
}

func (r *UserResourceImpl) FormatUser(user *models.User) *prot.User {
	if user == nil {
		return nil
	}

	userInfo := &prot.UserInfo{}
	var schools []*prot.SchoolInfo
	var classes []*prot.ClassInfo

	if user.UserClasses != nil && len(user.UserClasses) > 0 {

		for _, uc := range user.UserClasses {
			if uc.IsCurrent {
				if uc.Class != nil {
					school := &prot.SchoolInfo{}

					if uc.Class.School != nil {
						school.Id = uc.Class.School.ID
						school.Name = uc.Class.School.Name

						schools = append(schools, school)
					}

					class := &prot.ClassInfo{
						Id:     uc.Class.ID,
						Name:   uc.Class.Name,
						School: school,
					}

					classes = append(classes, class)

					if uc.Class.Faculty != nil && userInfo.Faculty == nil {
						userInfo.Faculty = &prot.FacultyInfo{
							Id:   uc.Class.Faculty.ID,
							Name: uc.Class.Faculty.Name,
						}
					}
				}
			}
		}
	}

	userInfo.Classes = classes
	userInfo.Schools = schools

	if len(classes) > 0 {
		userInfo.Class = classes[0]
		userInfo.School = classes[0].School
	}

	if user.UserAddress.ID != 0 {
		var tr models.UserAddressTranslation
		_ = i18n.FillTranslationWithProvider(user.UserAddress, &tr, r.localeProvider)
		userInfo.Address = &prot.AddressInfo{
			Id:           int64(user.UserAddress.ID),
			ProvinceCode: user.UserAddress.ProvinceCode,
			WardCode:     user.UserAddress.WardCode,
			ProvinceName: tr.ProvinceName,
			WardName:     tr.WardName,
			Address:      tr.Address,
		}
	}

	var mainTeacher bool

	if user.UserCourses != nil {
		var courses []*prot.CourseInfo
		for _, uc := range user.UserCourses {
			if uc.Course != nil {
				var program *prot.UserProgramInfo
				if uc.Course.Program.ID > 0 {
					program = &prot.UserProgramInfo{
						Id:   uc.Course.Program.ID,
						Name: uc.Course.Program.Name,
					}
				}
				courses = append(courses, &prot.CourseInfo{
					Id:      uc.Course.ID,
					Name:    uc.Course.Name,
					Program: program,
				})
			}

			if uc.MainTeacher && r.CourseId == uc.CourseId {
				mainTeacher = true
			}
		}
		userInfo.Courses = courses
	}

	if user.Subjects != nil {
		var subjects []*prot.SubjectInfo
		for _, s := range user.Subjects {
			subjects = append(subjects, &prot.SubjectInfo{
				Id:   s.ID,
				Name: s.Name,
			})
		}
		userInfo.Subjects = subjects
	}

	if user.Certificates != nil {
		var certificates []*prot.Certificate
		for _, uc := range user.Certificates {
			var issuedDateStr string
			if !uc.IssuedDate.IsZero() {
				issuedDateStr = uc.IssuedDate.Format("2006-01-02")
			}

			var expiryDateStr string
			if !uc.ExpiryDate.IsZero() {
				expiryDateStr = uc.ExpiryDate.Format("2006-01-02")
			}

			certificates = append(certificates, &prot.Certificate{
				Id:              uc.ID,
				UserId:          int64(user.ID),
				Name:            uc.Name,
				IssuedBy:        uc.IssuedBy,
				IssuedDate:      issuedDateStr,
				ExpiryDate:      expiryDateStr,
				CertificateCode: uc.CertificateCode,
				Status:          uc.Status,
				Description:     uc.Description,
				FileUrl:         utils.StaticURL(uc.FileInfo.Path, models.Storage),
			})
		}
		userInfo.Certificates = certificates
	}

	if user.Degrees != nil {
		var degrees []*prot.Degree
		for _, ud := range user.Degrees {
			var receivedDateStr string
			if !ud.ReceivedDate.IsZero() {
				receivedDateStr = ud.ReceivedDate.Format("2006-01-02")
			}

			degrees = append(degrees, &prot.Degree{
				Id:             ud.ID,
				Name:           ud.Name,
				Major:          ud.Major,
				Institution:    ud.Institution,
				GraduationYear: int32(ud.GraduationYear),
				DegreeLevel:    ud.DegreeLevel,
				ReceivedDate:   receivedDateStr,
				Note:           ud.Note,
				DegreeCode:     ud.DegreeCode,
				Status:         ud.Status,
				FileUrl:        utils.StaticURL(ud.FileInfo.Path, models.Storage),
			})
		}
		userInfo.Degrees = degrees
	}

	if user.Departments != nil {
		var departments []*prot.Department
		for _, ud := range user.Departments {

			departments = append(departments, &prot.Department{
				Id:          ud.ID,
				Name:        ud.Name,
				Description: ud.Description,
			})
		}
		userInfo.Departments = departments
	}

	if user.Positions != nil {
		var positions []*prot.EmployeePosition
		for _, up := range user.Positions {

			positions = append(positions, &prot.EmployeePosition{
				Id:          up.ID,
				Name:        up.Name,
				Level:       up.Level,
				Description: up.Description,
			})
		}
		userInfo.Positions = positions
	}

	var roleName string
	var roleId int64

	if len(user.Roles) > 0 {
		roles := make([]*prot.UserRole, 0, len(user.Roles))

		for _, role := range user.Roles {
			roles = append(roles, &prot.UserRole{
				Id:   int64(role.ID),
				Name: role.Name,
			})
		}

		userInfo.Roles = roles

		sort.Slice(roles, func(i, j int) bool {
			return roles[i].Id < roles[j].Id
		})

		userInfo.Role = roles[0]
		roleName = user.Roles[0].Name
		roleId = user.Roles[0].ID
	}

	memberType := models.MemberTypeInternal

	if user.MemberType == models.MemberTypeExternal {
		memberType = models.MemberTypeExternal
	}

	var isFailed bool
	for _, fid := range r.FailedUserIds {
		if fid == int64(user.ID) {
			isFailed = true
			break
		}
	}

	return &prot.User{
		Id:          int64(user.ID),
		Email:       user.Email,
		Identifier:  user.Identifier,
		Code:        user.Code,
		Username:    user.Username,
		Name:        user.Name,
		Address:     user.Address,
		Status:      user.Status,
		PhoneNumber: user.PhoneNumber,
		ParentId:    int64(user.ParentID),
		RoleId:      roleId,
		RoleName:    roleName,
		UserInfo:    userInfo,
		Description: user.Description,
		Avatar:      utils.StaticURL(user.AvatarInfo.Path, models.Storage),
		MainTeacher: mainTeacher,
		TypeTeacher: int32(user.TypeTeacher),
		DateOfBirth: user.DateOfBirth.Format("2006-01-02"),
		MemberType:  memberType,
		IsFailed:    isFailed,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (r *UserResourceImpl) FormatUsers(users []*models.User) []*prot.User {
	result := make([]*prot.User, 0, len(users))
	for _, u := range users {
		if formatted := r.FormatUser(u); formatted != nil {
			result = append(result, formatted)
		}
	}
	return result
}

func (r *UserResourceImpl) FormatModelUser(user *prot.User) *models.User {
	if user == nil {
		return nil
	}

	schoolId := 0
	if user.UserInfo != nil && user.UserInfo.School != nil {
		schoolId = int(user.UserInfo.School.Id)
	}

	if schoolId == 0 && user.UserInfo != nil && user.UserInfo.Class != nil && user.UserInfo.Class.Id != 0 {
		classRepo := repositories.NewClassRepository()
		class, err := classRepo.FindByID(int(user.UserInfo.Class.Id))
		if err == nil && class != nil {
			schoolId = int(class.SchoolId)
		}
	}

	dateOfBirth, err := time.Parse("2006-01-02", user.DateOfBirth)
	if err != nil {
		dateOfBirth = time.Time{}
	}

	avatar := utils.StripDomain(user.Avatar, models.Storage)

	mediaRepo := repositories.NewMediaRepository()
	avatarInfo := mediaRepo.GetMediaInfo(avatar, models.Storage)

	memberType := models.MemberTypeInternal
	if user.MemberType == models.MemberTypeExternal {
		memberType = models.MemberTypeExternal
	}

	return &models.User{
		ID:                   int64(user.Id),
		Identifier:           user.Identifier,
		Username:             user.Username,
		Code:                 user.Code,
		Name:                 user.Name,
		Address:              user.Address,
		Status:               user.Status,
		PhoneNumber:          user.PhoneNumber,
		ParentID:             int(user.ParentId),
		Email:                user.Email,
		SchoolID:             schoolId,
		Description:          user.Description,
		AvatarInfo:           avatarInfo,
		DateOfBirth:          dateOfBirth,
		TypeTeacher:          int16(user.TypeTeacher),
		MemberType:           memberType,
		IsIndependentStudent: user.IsIndependentStudent,
		IsFailedSubject:      user.IsFailedSubject,
	}
}
