package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/dto"
	"be-cleverschool/models"
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// DashboardAssessmentReportRepository dùng riêng cho API /api/dashboard/assessment/report
type DashboardAssessmentReportRepository interface {
	GetStudentsByClass(classID int64, classMainID int64, studentIDs []int64, limit, offset int) ([]dto.AssessmentStudentItem, int64, error)
	GetClassInfoForReport(classID int64) (string, string, models.MediaInfo, error)
	GetClassMainInfoForReport(classMainID int64) (string, string, models.MediaInfo, error)
	GetTeachersByClass(classID int64, subjectID int64) ([]string, []string, error) // vn_teachers, fr_teachers
	GetTeachersByClassMain(classMainID int64, subjectID int64) ([]string, []string, error) // vn_teachers, fr_teachers
	GetProgramNamesByAssessmentAndClassMain(assessmentID, classMainID int64) (string, error)
	GetClassMainIDByClassID(classID int64, classMainID *int64) error
	// class_name, school_name, school_logo_info
}

type dashboardAssessmentReportRepository struct{}

func NewDashboardAssessmentReportRepository() DashboardAssessmentReportRepository {
	return &dashboardAssessmentReportRepository{}
}

func (r *dashboardAssessmentReportRepository) GetStudentsByClass(classID int64, classMainID int64, studentIDs []int64, limit, offset int) ([]dto.AssessmentStudentItem, int64, error) {
	var subQuery *gorm.DB

	// Nếu có class_main_id thì lấy học sinh từ tất cả classes có class_main_id này
	if classMainID > 0 {
		// Lấy tất cả class_id có class_main_id tương ứng
		classIDsSubQuery := db.ReplicaDB.Table("classes").
			Select("id").
			Where("class_main_id = ? AND deleted_at IS NULL", classMainID)

		// Lấy user_id từ các classes đó
		subQuery = db.ReplicaDB.Table("user_classes").
			Select("user_id").
			Where("class_id IN (?)", classIDsSubQuery)
	} else if classID > 0 {
		// Lấy học sinh từ class_id cụ thể
		subQuery = db.ReplicaDB.Table("user_classes").
			Select("user_id").
			Where("class_id = ?", classID)
	} else {
		// Trường hợp không có class_id và class_main_id:
		// Chỉ filter theo danh sách student_id (nếu có), không giới hạn theo lớp
		subQuery = db.ReplicaDB.Table("user_ref_roles").
			Select("user_id").
			Where("role_id = 3")
	}

	// Filter theo student_ids nếu có
	if len(studentIDs) > 0 {
		// Lọc các student_id hợp lệ (> 0)
		validStudentIDs := make([]int64, 0, len(studentIDs))
		for _, id := range studentIDs {
			if id > 0 {
				validStudentIDs = append(validStudentIDs, id)
			}
		}
		if len(validStudentIDs) > 0 {
			subQuery = subQuery.Where("user_id IN (?)", validStudentIDs)
		}
	}

	var students []dto.AssessmentStudentItem
	var total int64

	// Count total
	countQuery := db.ReplicaDB.Table("users u").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND u.status = true").
		Where("u.id IN (?)", subQuery)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch student list (không order ở đây, sẽ sắp xếp sau)
	query := db.ReplicaDB.Table("users u").
		Select("u.id AS student_id, u.name AS student_name").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("u.deleted_at IS NULL AND urr.role_id = 3 AND u.status = true").
		Where("u.id IN (?)", subQuery)

	if limit > 0 && offset >= 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&students).Error; err != nil {
		return nil, 0, err
	}

	// Sắp xếp theo tên (phần cuối cùng của name) tăng dần với hỗ trợ tiếng Việt
	c := collate.New(language.Vietnamese)
	sort.Slice(students, func(i, j int) bool {
		lastNameI := extractLastName(students[i].StudentName)
		lastNameJ := extractLastName(students[j].StudentName)
		return c.CompareString(lastNameI, lastNameJ) < 0
	})

	return students, total, nil
}

// GetClassInfoForReport dùng riêng cho API report, trả về class_name, school_name và school_logo_info
func (r *dashboardAssessmentReportRepository) GetClassInfoForReport(classID int64) (string, string, models.MediaInfo, error) {
	// Lấy class_name, school_name và logo_info từ classes + schools
	var info struct {
		Name       string
		SchoolName string
		LogoInfo   models.MediaInfo `gorm:"type:jsonb;column:logo_info"`
	}
	err := db.ReplicaDB.Table("classes c").
		Select("c.name, COALESCE(s.name, '') AS school_name, s.logo_info").
		Joins("LEFT JOIN schools s ON s.id = c.school_id AND s.deleted_at IS NULL").
		Where("c.id = ? AND c.deleted_at IS NULL", classID).
		First(&info).Error
	if err != nil {
		return "", "", models.MediaInfo{}, err
	}

	return info.Name, info.SchoolName, info.LogoInfo, nil
}

// GetClassMainInfoForReport lấy thông tin class_main, school_name và school_logo_info
func (r *dashboardAssessmentReportRepository) GetClassMainInfoForReport(classMainID int64) (string, string, models.MediaInfo, error) {
	var info struct {
		Name       string
		SchoolName string
		LogoInfo   models.MediaInfo `gorm:"type:jsonb;column:logo_info"`
	}
	err := db.ReplicaDB.Table("classes_main cm").
		Select("cm.name, COALESCE(s.name, '') AS school_name, s.logo_info").
		Joins("LEFT JOIN schools s ON s.id = cm.school_id AND s.deleted_at IS NULL").
		Where("cm.id = ? AND cm.deleted_at IS NULL", classMainID).
		First(&info).Error
	if err != nil {
		return "", "", models.MediaInfo{}, err
	}

	return info.Name, info.SchoolName, info.LogoInfo, nil
}

// GetTeachersByClass lấy danh sách giáo viên VN và FR theo class_id
// Trả về: vn_teachers (type_teacher = 1), fr_teachers (type_teacher = 2)
// Lọc giáo viên theo subject_id từ assessment thông qua user_courses -> courses -> programs
// Đồng thời khóa học đó phải có học sinh thuộc cả class và course
func (r *dashboardAssessmentReportRepository) GetTeachersByClass(classID int64, subjectID int64) ([]string, []string, error) {
	var vnTeachers, frTeachers []string

	// Subquery để lấy user_id của giáo viên có trong user_courses với subject_id tương ứng
	// VÀ khóa học đó phải có học sinh thuộc cả class và course
	// Logic: 
	// 1. user_courses -> courses -> programs, lấy programs.subject_id = subject_id
	// 2. Kiểm tra có học sinh (role_id = 3) thuộc cả class_id và course_id đó
	teacherSubQuery := db.ReplicaDB.Table("user_courses uc").
		Select("DISTINCT uc.user_id").
		Joins("JOIN courses co ON co.id = uc.course_id AND co.deleted_at IS NULL").
		Joins("JOIN programs p ON p.id = co.program_id AND p.deleted_at IS NULL").
		Joins("JOIN user_classes ucl ON ucl.class_id = ?", classID).
		Joins("JOIN user_ref_roles urr_stu ON urr_stu.user_id = ucl.user_id AND urr_stu.role_id = 3").
		Joins("JOIN user_courses uco_stu ON uco_stu.user_id = ucl.user_id AND uco_stu.course_id = co.id").
		Where("p.subject_id = ?", subjectID).
		Where("uc.user_id IN (SELECT user_id FROM user_ref_roles WHERE role_id = 2)")

	// Lấy giáo viên VN (type_teacher = 1)
	err := db.ReplicaDB.Table("user_classes uc").
		Select("DISTINCT u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("uc.class_id = ? AND urr.role_id = 2 AND u.type_teacher = 1 AND u.deleted_at IS NULL", classID).
		Where("u.id IN (?)", teacherSubQuery).
		Order("u.name ASC").
		Pluck("u.name", &vnTeachers).Error
	if err != nil {
		return nil, nil, err
	}

	// Lấy giáo viên FR (type_teacher = 2)
	err = db.ReplicaDB.Table("user_classes uc").
		Select("DISTINCT u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("uc.class_id = ? AND urr.role_id = 2 AND u.type_teacher = 2 AND u.deleted_at IS NULL", classID).
		Where("u.id IN (?)", teacherSubQuery).
		Order("u.name ASC").
		Pluck("u.name", &frTeachers).Error
	if err != nil {
		return nil, nil, err
	}

	return vnTeachers, frTeachers, nil
}

// GetTeachersByClassMain lấy danh sách giáo viên VN và FR theo class_main_id
// Trả về: vn_teachers (type_teacher = 1), fr_teachers (type_teacher = 2)
// Lọc giáo viên theo subject_id từ assessment thông qua user_courses -> courses -> programs
// Đồng thời khóa học đó phải có học sinh thuộc cả class và course
func (r *dashboardAssessmentReportRepository) GetTeachersByClassMain(classMainID int64, subjectID int64) ([]string, []string, error) {
	var vnTeachers, frTeachers []string

	// Lấy tất cả class_id có class_main_id này
	classIDsSubQuery := db.ReplicaDB.Table("classes").
		Select("id").
		Where("class_main_id = ? AND deleted_at IS NULL", classMainID)

	// Subquery để lấy user_id của giáo viên có trong user_courses với subject_id tương ứng
	// VÀ khóa học đó phải có học sinh thuộc cả class (trong classIDsSubQuery) và course
	// Logic: 
	// 1. user_courses -> courses -> programs, lấy programs.subject_id = subject_id
	// 2. Kiểm tra có học sinh (role_id = 3) thuộc cả class_id (trong classIDsSubQuery) và course_id đó
	teacherSubQuery := db.ReplicaDB.Table("user_courses uc").
		Select("DISTINCT uc.user_id").
		Joins("JOIN courses co ON co.id = uc.course_id AND co.deleted_at IS NULL").
		Joins("JOIN programs p ON p.id = co.program_id AND p.deleted_at IS NULL").
		Joins("JOIN user_classes ucl ON ucl.class_id IN (?)", classIDsSubQuery).
		Joins("JOIN user_ref_roles urr_stu ON urr_stu.user_id = ucl.user_id AND urr_stu.role_id = 3").
		Joins("JOIN user_courses uco_stu ON uco_stu.user_id = ucl.user_id AND uco_stu.course_id = co.id").
		Where("p.subject_id = ?", subjectID).
		Where("uc.user_id IN (SELECT user_id FROM user_ref_roles WHERE role_id = 2)")

	// Lấy giáo viên VN (type_teacher = 1) từ tất cả classes có class_main_id này
	err := db.ReplicaDB.Table("user_classes uc").
		Select("DISTINCT u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("uc.class_id IN (?) AND urr.role_id = 2 AND u.type_teacher = 1 AND u.deleted_at IS NULL", classIDsSubQuery).
		Where("u.id IN (?)", teacherSubQuery).
		Order("u.name ASC").
		Pluck("u.name", &vnTeachers).Error
	if err != nil {
		return nil, nil, err
	}

	// Lấy giáo viên FR (type_teacher = 2) từ tất cả classes có class_main_id này
	err = db.ReplicaDB.Table("user_classes uc").
		Select("DISTINCT u.name").
		Joins("JOIN users u ON u.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = u.id").
		Where("uc.class_id IN (?) AND urr.role_id = 2 AND u.type_teacher = 2 AND u.deleted_at IS NULL", classIDsSubQuery).
		Where("u.id IN (?)", teacherSubQuery).
		Order("u.name ASC").
		Pluck("u.name", &frTeachers).Error
	if err != nil {
		return nil, nil, err
	}

	return vnTeachers, frTeachers, nil
}

// GetProgramNamesByAssessmentAndClassMain lấy danh sách tên program từ assessment và class_main_id
// Logic:
// 1. Lấy program_id từ assessments → assessment_ref_lessons → lessons → chapters → programs
// 2. Lấy program_id từ class_main_id → user_classes → user_courses → courses → programs
// 3. Lấy các program có trong cả 2 danh sách (intersection)
// 4. Trả về string tên các program, nếu nhiều hơn 1 thì nối bằng dấu phẩy
func (r *dashboardAssessmentReportRepository) GetProgramNamesByAssessmentAndClassMain(assessmentID, classMainID int64) (string, error) {
	// 1. Lấy danh sách program_id từ assessment
	var programIDsFromAssessment []int64
	err := db.ReplicaDB.Table("assessments a").
		Select("DISTINCT p.id").
		Joins("JOIN assessment_ref_lessons arl ON arl.assessment_id = a.id").
		Joins("JOIN lessons l ON l.id = arl.lesson_id AND l.deleted_at IS NULL").
		Joins("JOIN chapters ch ON ch.id = l.chapter_id AND ch.deleted_at IS NULL").
		Joins("JOIN programs p ON p.id = ch.program_id AND p.deleted_at IS NULL").
		Where("a.id = ? AND a.deleted_at IS NULL", assessmentID).
		Pluck("p.id", &programIDsFromAssessment).Error
	if err != nil {
		return "", err
	}

	// 2. Lấy danh sách program_id từ class_main_id (chỉ lấy users có role_id = 3 - học sinh)
	var programIDsFromClassMain []int64
	err = db.ReplicaDB.Table("user_courses uc").
		Select("DISTINCT p.id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = uc.user_id AND urr.role_id = 3").
		Joins("JOIN user_classes ucl ON ucl.user_id = uc.user_id").
		Joins("JOIN classes c ON c.id = ucl.class_id AND c.deleted_at IS NULL").
		Joins("JOIN courses co ON co.id = uc.course_id AND co.deleted_at IS NULL").
		Joins("JOIN programs p ON p.id = co.program_id AND p.deleted_at IS NULL").
		Where("c.class_main_id = ?", classMainID).
		Pluck("p.id", &programIDsFromClassMain).Error
	if err != nil {
		return "", err
	}

	// 3. Tìm intersection (các program có trong cả 2 danh sách)
	programIDMap := make(map[int64]bool)
	for _, id := range programIDsFromAssessment {
		programIDMap[id] = true
	}

	var intersectionProgramIDs []int64
	for _, id := range programIDsFromClassMain {
		if programIDMap[id] {
			intersectionProgramIDs = append(intersectionProgramIDs, id)
		}
	}

	// 4. Nếu không có program nào, trả về rỗng
	if len(intersectionProgramIDs) == 0 {
		return "", nil
	}

	// 5. Lấy tên các program
	var programNames []string
	err = db.ReplicaDB.Table("programs").
		Select("name").
		Where("id IN (?) AND deleted_at IS NULL", intersectionProgramIDs).
		Order("name ASC").
		Pluck("name", &programNames).Error
	if err != nil {
		return "", err
	}

	// 6. Nối các tên bằng dấu phẩy
	result := ""
	for i, name := range programNames {
		if i > 0 {
			result += ", "
		}
		result += name
	}

	return result, nil
}

// GetClassMainIDByClassID lấy class_main_id từ class_id
func (r *dashboardAssessmentReportRepository) GetClassMainIDByClassID(classID int64, classMainID *int64) error {
	err := db.ReplicaDB.Table("classes").
		Select("class_main_id").
		Where("id = ? AND deleted_at IS NULL", classID).
		Scan(classMainID).Error
	return err
}

