package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/requests"
	"sort"
	"strings"
)

// AssessmentRankingRow kết quả từ DB, dùng trong service để map sang DTO
type AssessmentRankingRow struct {
	StudentID   int64            `gorm:"column:student_id"`
	StudentName string           `gorm:"column:student_name"`
	AvatarInfo  models.MediaInfo `gorm:"column:avatar_info"`
	Class       string           `gorm:"column:class"` // Tên lớp, nhiều lớp cách nhau bởi dấu phẩy
	Score       float64          `gorm:"column:score"`
	Star        int              `gorm:"column:star"`
}

type AssessmentRankingRepository interface {
	GetAssessmentRanking(req *requests.AssessmentRankingRequest) ([]AssessmentRankingRow, int64, error)
}

type assessmentRankingRepository struct{}

func NewAssessmentRankingRepository() AssessmentRankingRepository {
	return &assessmentRankingRepository{}
}

func (r *assessmentRankingRepository) GetAssessmentRanking(req *requests.AssessmentRankingRequest) ([]AssessmentRankingRow, int64, error) {
	var rows []AssessmentRankingRow

	// Học sinh trong course; JOIN assessments lấy study_report_criteria_id; score = assessment_scores; star = study_reports
	sql := `
SELECT u.id AS student_id,
       u.name AS student_name,
       u.avatar_info,
       (SELECT string_agg(c.name, ', ' ORDER BY c.name)
        FROM user_classes uc
        JOIN classes c ON c.id = uc.class_id AND c.deleted_at IS NULL
        WHERE uc.user_id = u.id AND uc.is_current = true) AS class,
       COALESCE(agg_score.total_score, 0)::numeric(8,2) AS score,
       COALESCE(MAX(sr.total_star), 0)::int AS star
FROM users u
JOIN user_courses uc ON uc.user_id = u.id AND uc.course_id = ?
JOIN user_ref_roles urr ON urr.user_id = u.id AND urr.role_id = 3
JOIN assessments a ON a.id = ? AND a.deleted_at IS NULL
LEFT JOIN (
  SELECT student_id, SUM(total_score) AS total_score
  FROM assessment_scores
  WHERE assessment_id = ? AND course_id = ? AND deleted_at IS NULL
  GROUP BY student_id
) agg_score ON agg_score.student_id = u.id
LEFT JOIN study_reports sr ON sr.student_id = u.id AND sr.course_id = ? AND sr.study_report_criteria_id = a.study_report_criteria_id AND sr.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.name, u.avatar_info, agg_score.total_score
`
	err := db.ReplicaDB.Raw(sql,
		req.CourseID, req.AssessmentID, req.AssessmentID, req.CourseID, req.CourseID,
	).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(rows))

	orderBy := strings.ToLower(strings.TrimSpace(req.OrderBy))
	sortOrder := strings.ToLower(strings.TrimSpace(req.Sort))
	if sortOrder == "" {
		sortOrder = "desc"
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
			// Nếu star bằng nhau, sort theo score
			if rows[i].Score != rows[j].Score {
				if sortOrder == "asc" {
					return rows[i].Score < rows[j].Score
				} else {
					return rows[i].Score > rows[j].Score
				}
			}
			// Nếu cả star và score đều bằng nhau, sort theo student_id để có thứ tự ổn định
			return rows[i].StudentID < rows[j].StudentID
		} else {
			// Sort theo score trước
			if rows[i].Score != rows[j].Score {
				if sortOrder == "asc" {
					return rows[i].Score < rows[j].Score
				} else {
					return rows[i].Score > rows[j].Score
				}
			}
			// Nếu score bằng nhau, sort theo star
			if rows[i].Star != rows[j].Star {
				if sortOrder == "asc" {
					return rows[i].Star < rows[j].Star
				} else {
					return rows[i].Star > rows[j].Star
				}
			}
			// Nếu cả score và star đều bằng nhau, sort theo student_id để có thứ tự ổn định
			return rows[i].StudentID < rows[j].StudentID
		}
	})

	return rows, total, nil
}
