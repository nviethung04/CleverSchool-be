package repositories

import (
	"be-cleverschool/database/db"
	"be-cleverschool/models"
	"be-cleverschool/requests"
	"sort"
	"strings"
)

// HomeworkRankingRow kết quả từ DB, dùng trong service để map sang DTO (tránh import cycle utils -> repositories)
type HomeworkRankingRow struct {
	StudentID   int64            `gorm:"column:student_id"`
	StudentName string           `gorm:"column:student_name"`
	AvatarInfo  models.MediaInfo `gorm:"column:avatar_info"`
	Class       string           `gorm:"column:class"`
	Star        int64            `gorm:"column:star"`
	Exp         float64          `gorm:"column:exp"`
}

type HomeworkRankingRepository interface {
	GetHomeworkRanking(req *requests.HomeworkRankingRequest) ([]HomeworkRankingRow, int64, error)
}

type homeworkRankingRepository struct{}

func NewHomeworkRankingRepository() HomeworkRankingRepository {
	return &homeworkRankingRepository{}
}

func (r *homeworkRankingRepository) GetHomeworkRanking(req *requests.HomeworkRankingRequest) ([]HomeworkRankingRow, int64, error) {
	var rows []HomeworkRankingRow
	var totalCount int64

	// Query lấy tất cả học sinh trong course với star, exp từ user_star_exp và class từ user_classes + classes
	query := db.ReplicaDB.Table("users").
		Select(`
			users.id as student_id,
			users.name as student_name,
			users.avatar_info,
			(SELECT string_agg(c.name, ', ' ORDER BY c.name) FROM user_classes uc JOIN classes c ON c.id = uc.class_id AND c.deleted_at IS NULL WHERE uc.user_id = users.id AND uc.is_current = true) as class,
			COALESCE(use.total_star, 0) as star,
			COALESCE(use.total_exp, 0) as exp
		`).
		Joins("JOIN user_courses uc ON users.id = uc.user_id").
		Joins("JOIN user_ref_roles urr ON urr.user_id = users.id").
		Joins("LEFT JOIN user_star_exp use ON users.id = use.user_id AND use.is_current = true").
		Where("uc.course_id = ?", req.CourseID).
		Where("urr.role_id = 3").
		Where("users.deleted_at IS NULL")

	// Count total
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Lấy tất cả dữ liệu (không có pagination, không có sorting ở query level)
	if err := query.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	// Sắp xếp trong memory theo order_by, nếu bằng nhau thì sort theo giá trị còn lại
	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	sortOrder := strings.ToLower(strings.TrimSpace(req.Sort))
	if sortOrder == "" {
		sortOrder = "desc" // Mặc định giảm dần
	}

	sort.Slice(rows, func(i, j int) bool {
		if orderBy == "star" {
			// Sort theo star trước
			if rows[i].Star != rows[j].Star {
				if sortOrder == "asc" {
					return rows[i].Star < rows[j].Star
				} else {
					return rows[i].Star > rows[j].Star
				}
			}
			// Nếu star bằng nhau, sort theo exp
			if rows[i].Exp != rows[j].Exp {
				if sortOrder == "asc" {
					return rows[i].Exp < rows[j].Exp
				} else {
					return rows[i].Exp > rows[j].Exp
				}
			}
			// Nếu cả star và exp đều bằng nhau, sort theo student_id để có thứ tự ổn định
			return rows[i].StudentID < rows[j].StudentID
		} else {
			// Sort theo exp trước
			if rows[i].Exp != rows[j].Exp {
				if sortOrder == "asc" {
					return rows[i].Exp < rows[j].Exp
				} else {
					return rows[i].Exp > rows[j].Exp
				}
			}
			// Nếu exp bằng nhau, sort theo star
			if rows[i].Star != rows[j].Star {
				if sortOrder == "asc" {
					return rows[i].Star < rows[j].Star
				} else {
					return rows[i].Star > rows[j].Star
				}
			}
			// Nếu cả exp và star đều bằng nhau, sort theo student_id để có thứ tự ổn định
			return rows[i].StudentID < rows[j].StudentID
		}
	})

	return rows, totalCount, nil
}

