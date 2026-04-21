package repositories

import (
	"be-lms/database/db"
	"be-lms/models"
	"be-lms/repositories/base"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContestRepository interface {
	base.BaseRepositoryInterface[models.Contest]
	base.BeforeQueryHook
	GetContestRounds(contestId int64) ([]models.ContestRound, error)
	GetContestWithRounds(contestId int64) (*models.Contest, error)
	GetUserContests() ([]models.Contest, int64, error)
}

type contestRepository struct {
	*base.BaseRepository[models.Contest]
}

func NewContestRepository() ContestRepository {
	repo := &contestRepository{
		BaseRepository: base.NewBaseRepository[models.Contest](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *contestRepository) GetContestRounds(contestId int64) ([]models.ContestRound, error) {
	var rounds []models.ContestRound
	err := db.ReplicaDB.
		Where("contest_id = ?", contestId).
		Order("sort_position ASC").
		Find(&rounds).Error
	return rounds, err
}

func (r *contestRepository) GetContestWithRounds(contestId int64) (*models.Contest, error) {
	var contest models.Contest
	err := db.ReplicaDB.
		Preload("ContestRounds").
		Where("id = ?", contestId).
		First(&contest).Error
	return &contest, err
}

// BeforeQueryHook implementation
func (r *contestRepository) BeforeQuery(query *gorm.DB, c *gin.Context) *gorm.DB {
	// Preload Creator information for contests
	fmt.Printf("🔍 DEBUG: BeforeQuery called, preloading Creator\n")
	return query.Preload("Creator")
}

// GetUserContests implementation
func (r *contestRepository) GetUserContests() ([]models.Contest, int64, error) {
	var contests []models.Contest
	var total int64

	// For now, return all contests - you can implement user-specific logic later
	// This is a simplified implementation - you might need to adjust based on your business logic
	err := db.ReplicaDB.Model(&models.Contest{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.ReplicaDB.Find(&contests).Error
	return contests, total, err
}

type ContestRoundRepository interface {
	base.BaseRepositoryInterface[models.ContestRound]
	base.BeforeQueryHook
	GetByContestId(contestId int64) ([]models.ContestRound, error)
	GetContestRoundUsers(contestRoundId int64) ([]models.ContestRoundUser, error)
	GetContestRoundJoinerSchools(contestRoundId int64) ([]models.ContestRoundJoinerSchool, error)
	GetContestRoundJoinerProvinces(contestRoundId int64) ([]models.ContestRoundJoinerProvince, error)
	GetContestRoundJoinerPersons(contestRoundId int64) ([]models.ContestRoundJoinerPerson, error)
	GetContestRoundJoinerClasses(contestRoundId int64) ([]models.ContestRoundJoinerClass, error)
	AddJoinerSchool(contestRoundId int64, schoolId int64, createdBy int64) error
	AddJoinerProvince(contestRoundId int64, provinceId int64, createdBy int64) error
	AddJoinerPerson(contestRoundId int64, userId int64, createdBy int64) error
	AddJoinerClass(contestRoundId int64, classId int64, createdBy int64) error
	RemoveJoinerSchool(contestRoundId int64, schoolId int64, deletedBy int64) error
	RemoveJoinerProvince(contestRoundId int64, provinceId int64, deletedBy int64) error
	RemoveJoinerPerson(contestRoundId int64, userId int64, deletedBy int64) error

	// New person joiner methods
	AddPersonJoiner(contestRoundId int64, userId int64, createdBy int64) error
	GetPersonJoiners(contestRoundId int64) (interface{}, error)
	RemovePersonJoiner(contestRoundId int64, userId int64, deletedBy int64) error
	RemoveJoinerClass(contestRoundId int64, classId int64, deletedBy int64) error
	ReplaceJoinersByLevel(contestRoundId int64, joinLevel string, joinerIds []int64, createdBy int64) error

	// New bulk remove methods by entity IDs
	BulkRemoveJoinerSchools(contestRoundId int64, schoolIds []int64, deletedBy int64) error
	BulkRemoveJoinerProvinces(contestRoundId int64, provinceIds []int64, deletedBy int64) error
	BulkRemoveJoinerClasses(contestRoundId int64, classIds []int64, deletedBy int64) error
	BulkRemoveJoinerPersons(contestRoundId int64, userIds []int64, deletedBy int64) error

	GetContestRoundProvinces(contestRoundId int64) (interface{}, error)
	GetContestRoundSchools(contestRoundId int64) (interface{}, error)
	GetContestRoundClasses(contestRoundId int64) (interface{}, error)
	GetContestRoundStudents(contestRoundId int64) (interface{}, error)
	GetContestRoundEligibleUsers(contestRoundId int64) ([]models.User, error)
}

type contestRoundRepository struct {
	*base.BaseRepository[models.ContestRound]
}

func NewContestRoundRepository() ContestRoundRepository {
	repo := &contestRoundRepository{
		BaseRepository: base.NewBaseRepository[models.ContestRound](),
	}
	repo.BaseRepository.SetBeforeQueryHook(repo)
	return repo
}

func (r *contestRoundRepository) GetByContestId(contestId int64) ([]models.ContestRound, error) {
	var rounds []models.ContestRound
	err := db.ReplicaDB.
		Where("contest_id = ?", contestId).
		Order("sort_position ASC").
		Find(&rounds).Error
	return rounds, err
}

func (r *contestRoundRepository) GetContestRoundUsers(contestRoundId int64) ([]models.ContestRoundUser, error) {
	var users []models.ContestRoundUser
	err := db.ReplicaDB.
		Preload("User").
		Where("contest_round_id = ?", contestRoundId).
		Order("score DESC").
		Find(&users).Error
	return users, err
}

func (r *contestRoundRepository) GetContestRoundJoinerSchools(contestRoundId int64) ([]models.ContestRoundJoinerSchool, error) {
	var schools []models.ContestRoundJoinerSchool
	err := db.ReplicaDB.
		Preload("School").
		Preload("School.Ward").
		Preload("School.Ward.Province").
		Preload("ContestRound").
		Preload("ContestRound.Contest").
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&schools).Error
	return schools, err
}

func (r *contestRoundRepository) GetContestRoundJoinerProvinces(contestRoundId int64) ([]models.ContestRoundJoinerProvince, error) {
	var provinces []models.ContestRoundJoinerProvince
	err := db.ReplicaDB.
		Preload("Province").
		Preload("ContestRound").
		Preload("ContestRound.Contest").
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&provinces).Error
	return provinces, err
}

func (r *contestRoundRepository) GetContestRoundJoinerPersons(contestRoundId int64) ([]models.ContestRoundJoinerPerson, error) {
	var persons []models.ContestRoundJoinerPerson
	err := db.ReplicaDB.Raw(`
		SELECT id, contest_round_id, user_id, created_at, created_by, deleted_at, deleted_by
		FROM contest_round_joiner_persons 
		WHERE contest_round_id = ? AND deleted_at IS NULL
	`, contestRoundId).Scan(&persons).Error
	return persons, err
}

func (r *contestRoundRepository) GetContestRoundJoinerClasses(contestRoundId int64) ([]models.ContestRoundJoinerClass, error) {
	var classes []models.ContestRoundJoinerClass
	err := db.ReplicaDB.
		Preload("Class").
		Preload("ContestRound").
		Preload("ContestRound.Contest").
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&classes).Error
	return classes, err
}

func (r *contestRoundRepository) AddJoinerSchool(contestRoundId int64, schoolId int64, createdBy int64) error {
	// Check if joiner already exists (not deleted)
	var existingJoiner models.ContestRoundJoinerSchool
	err := db.MasterDB.Where("contest_round_id = ? AND school_id = ? AND deleted_at IS NULL", contestRoundId, schoolId).First(&existingJoiner).Error
	if err == nil {
		// Joiner already exists, no need to add
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return err
	}

	// Create new joiner
	joiner := models.ContestRoundJoinerSchool{
		ContestRoundId: contestRoundId,
		SchoolId:       schoolId,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}
	return db.MasterDB.Create(&joiner).Error
}

func (r *contestRoundRepository) AddJoinerProvince(contestRoundId int64, provinceId int64, createdBy int64) error {
	// Check if joiner already exists (not deleted)
	var existingJoiner models.ContestRoundJoinerProvince
	err := db.MasterDB.Where("contest_round_id = ? AND province_id = ? AND deleted_at IS NULL", contestRoundId, provinceId).First(&existingJoiner).Error
	if err == nil {
		// Joiner already exists, no need to add
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return err
	}

	// Create new joiner
	joiner := models.ContestRoundJoinerProvince{
		ContestRoundId: contestRoundId,
		ProvinceId:     provinceId,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}
	return db.MasterDB.Create(&joiner).Error
}

func (r *contestRoundRepository) AddJoinerPerson(contestRoundId int64, userId int64, createdBy int64) error {
	// Check if joiner already exists (not deleted)
	var count int64
	err := db.MasterDB.Table("contest_round_joiner_persons").Where("contest_round_id = ? AND user_id = ? AND deleted_at IS NULL", contestRoundId, userId).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		// Joiner already exists, no need to add
		return nil
	}

	// Create new joiner using raw SQL with proper PostgreSQL syntax
	now := time.Now()
	return db.MasterDB.Exec(`
		INSERT INTO contest_round_joiner_persons (contest_round_id, user_id, created_at, created_by, deleted_at, deleted_by) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, contestRoundId, userId, now, createdBy, nil, 0).Error
}

func (r *contestRoundRepository) AddJoinerClass(contestRoundId int64, classId int64, createdBy int64) error {
	joiner := models.ContestRoundJoinerClass{
		ContestRoundId: contestRoundId,
		ClassId:        classId,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}
	return db.MasterDB.Create(&joiner).Error
}

func (r *contestRoundRepository) RemoveJoinerSchool(contestRoundId int64, schoolId int64, deletedBy int64) error {
	return db.MasterDB.Model(&models.ContestRoundJoinerSchool{}).
		Where("contest_round_id = ? AND school_id = ?", contestRoundId, schoolId).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error
}

func (r *contestRoundRepository) RemoveJoinerProvince(contestRoundId int64, provinceId int64, deletedBy int64) error {
	return db.MasterDB.Model(&models.ContestRoundJoinerProvince{}).
		Where("contest_round_id = ? AND province_id = ?", contestRoundId, provinceId).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error
}

func (r *contestRoundRepository) RemoveJoinerPerson(contestRoundId int64, userId int64, deletedBy int64) error {
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_persons 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND user_id = $4
	`, time.Now(), deletedBy, contestRoundId, userId).Error
}

func (r *contestRoundRepository) RemoveJoinerClass(contestRoundId int64, classId int64, deletedBy int64) error {
	return db.MasterDB.Model(&models.ContestRoundJoinerClass{}).
		Where("contest_round_id = ? AND class_id = ?", contestRoundId, classId).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
		}).Error
}

// BeforeQueryHook implementation
func (r *contestRoundRepository) BeforeQuery(query *gorm.DB, c *gin.Context) *gorm.DB {
	// Add any common query modifications here
	return query
}

// GetContestRoundProvinces implementation
func (r *contestRoundRepository) GetContestRoundProvinces(contestRoundId int64) (interface{}, error) {
	var provinces []models.ContestRoundJoinerProvince
	err := db.ReplicaDB.
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&provinces).Error
	return provinces, err
}

// GetContestRoundSchools implementation
func (r *contestRoundRepository) GetContestRoundSchools(contestRoundId int64) (interface{}, error) {
	var schools []models.ContestRoundJoinerSchool
	err := db.ReplicaDB.
		Preload("School").
		Preload("ContestRound").
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&schools).Error

	return schools, err
}

// GetContestRoundClasses implementation
func (r *contestRoundRepository) GetContestRoundClasses(contestRoundId int64) (interface{}, error) {
	var classes []models.ContestRoundJoinerClass
	err := db.ReplicaDB.
		Where("contest_round_id = ? AND deleted_at IS NULL", contestRoundId).
		Find(&classes).Error
	return classes, err
}

// GetContestRoundStudents implementation
func (r *contestRoundRepository) GetContestRoundStudents(contestRoundId int64) (interface{}, error) {
	var students []models.ContestRoundJoinerPerson
	err := db.ReplicaDB.Raw(`
		SELECT id, contest_round_id, user_id, created_at, created_by, deleted_at, deleted_by
		FROM contest_round_joiner_persons 
		WHERE contest_round_id = ? AND deleted_at IS NULL
	`, contestRoundId).Scan(&students).Error
	return students, err
}

// ReplaceJoinersByLevel implementation - following course pattern
func (r *contestRoundRepository) ReplaceJoinersByLevel(contestRoundId int64, joinLevel string, joinerIds []int64, createdBy int64) error {
	if len(joinerIds) == 0 {
		joinerIds = []int64{0}
	}

	tx := db.MasterDB
	now := time.Now()

	// Find joiners to delete (exist in DB but not in new list)
	var joinerIdsToDelete []int64

	switch joinLevel {
	case "school":
		err := tx.Model(&models.ContestRoundJoinerSchool{}).
			Where("contest_round_id = ? AND deleted_at IS NULL AND school_id NOT IN ?", contestRoundId, joinerIds).
			Pluck("school_id", &joinerIdsToDelete).Error
		if err != nil {
			return err
		}

		// Delete old joiners
		if len(joinerIdsToDelete) > 0 {
			err = tx.Model(&models.ContestRoundJoinerSchool{}).
				Where("contest_round_id = ? AND school_id IN ?", contestRoundId, joinerIdsToDelete).
				Updates(map[string]interface{}{
					"deleted_at": now,
					"deleted_by": createdBy,
				}).Error
			if err != nil {
				return err
			}
		}

		// Add new joiners
		for _, joinerId := range joinerIds {
			if joinerId == 0 {
				continue
			}
			var count int64
			err = tx.Model(&models.ContestRoundJoinerSchool{}).
				Where("contest_round_id = ? AND school_id = ? AND deleted_at IS NULL", contestRoundId, joinerId).
				Count(&count).Error
			if err != nil {
				return err
			}

			if count == 0 {
				joiner := models.ContestRoundJoinerSchool{
					ContestRoundId: contestRoundId,
					SchoolId:       joinerId,
					CreatedAt:      now,
					CreatedBy:      createdBy,
				}
				if err = tx.Create(&joiner).Error; err != nil {
					return err
				}
			}
		}

	case "province":
		err := tx.Model(&models.ContestRoundJoinerProvince{}).
			Where("contest_round_id = ? AND deleted_at IS NULL AND province_id NOT IN ?", contestRoundId, joinerIds).
			Pluck("province_id", &joinerIdsToDelete).Error
		if err != nil {
			return err
		}

		if len(joinerIdsToDelete) > 0 {
			err = tx.Model(&models.ContestRoundJoinerProvince{}).
				Where("contest_round_id = ? AND province_id IN ?", contestRoundId, joinerIdsToDelete).
				Updates(map[string]interface{}{
					"deleted_at": now,
					"deleted_by": createdBy,
				}).Error
			if err != nil {
				return err
			}
		}

		for _, joinerId := range joinerIds {
			if joinerId == 0 {
				continue
			}
			var count int64
			err = tx.Model(&models.ContestRoundJoinerProvince{}).
				Where("contest_round_id = ? AND province_id = ? AND deleted_at IS NULL", contestRoundId, joinerId).
				Count(&count).Error
			if err != nil {
				return err
			}

			if count == 0 {
				joiner := models.ContestRoundJoinerProvince{
					ContestRoundId: contestRoundId,
					ProvinceId:     joinerId,
					CreatedAt:      now,
					CreatedBy:      createdBy,
				}
				if err = tx.Create(&joiner).Error; err != nil {
					return err
				}
			}
		}

	case "person":
		// Get existing joiner IDs to delete
		var existingJoinerIds []int64
		err := tx.Raw(`
			SELECT user_id FROM contest_round_joiner_persons 
			WHERE contest_round_id = ? AND deleted_at IS NULL AND user_id NOT IN ?
		`, contestRoundId, joinerIds).Scan(&existingJoinerIds).Error
		if err != nil {
			return err
		}

		// Soft delete existing joiners not in new list
		if len(existingJoinerIds) > 0 {
			err = tx.Exec(`
				UPDATE contest_round_joiner_persons 
				SET deleted_at = $1, deleted_by = $2 
				WHERE contest_round_id = $3 AND user_id = ANY($4)
			`, now, createdBy, contestRoundId, existingJoinerIds).Error
			if err != nil {
				return err
			}
		}

		// Add new joiners
		for _, joinerId := range joinerIds {
			if joinerId == 0 {
				continue
			}
			var count int64
			err = tx.Raw(`
				SELECT COUNT(*) FROM contest_round_joiner_persons 
				WHERE contest_round_id = ? AND user_id = ? AND deleted_at IS NULL
			`, contestRoundId, joinerId).Scan(&count).Error
			if err != nil {
				return err
			}

			if count == 0 {
				err = tx.Exec(`
					INSERT INTO contest_round_joiner_persons (contest_round_id, user_id, created_at, created_by, deleted_at, deleted_by) 
					VALUES ($1, $2, $3, $4, $5, $6)
				`, contestRoundId, joinerId, now, createdBy, nil, 0).Error
				if err != nil {
					return err
				}
			}
		}

	case "class":
		err := tx.Model(&models.ContestRoundJoinerClass{}).
			Where("contest_round_id = ? AND deleted_at IS NULL AND class_id NOT IN ?", contestRoundId, joinerIds).
			Pluck("class_id", &joinerIdsToDelete).Error
		if err != nil {
			return err
		}

		if len(joinerIdsToDelete) > 0 {
			err = tx.Model(&models.ContestRoundJoinerClass{}).
				Where("contest_round_id = ? AND class_id IN ?", contestRoundId, joinerIdsToDelete).
				Updates(map[string]interface{}{
					"deleted_at": now,
					"deleted_by": createdBy,
				}).Error
			if err != nil {
				return err
			}
		}

		for _, joinerId := range joinerIds {
			if joinerId == 0 {
				continue
			}
			var count int64
			err = tx.Model(&models.ContestRoundJoinerClass{}).
				Where("contest_round_id = ? AND class_id = ? AND deleted_at IS NULL", contestRoundId, joinerId).
				Count(&count).Error
			if err != nil {
				return err
			}

			if count == 0 {
				joiner := models.ContestRoundJoinerClass{
					ContestRoundId: contestRoundId,
					ClassId:        joinerId,
					CreatedAt:      now,
					CreatedBy:      createdBy,
				}
				if err = tx.Create(&joiner).Error; err != nil {
					return err
				}
			}
		}

	default:
		return fmt.Errorf("invalid join level: %s", joinLevel)
	}

	return nil
}

// GetContestRoundEligibleUsers implementation - DISABLED
func (r *contestRoundRepository) GetContestRoundEligibleUsers(contestRoundId int64) ([]models.User, error) {
	// Completely disabled to avoid table name issues
	return []models.User{}, nil
}

// AddPersonJoiner - Add a person joiner to contest round
func (r *contestRoundRepository) AddPersonJoiner(contestRoundId int64, userId int64, createdBy int64) error {
	// Check if joiner already exists (not deleted)
	var count int64
	err := db.MasterDB.Raw(`
		SELECT COUNT(*) FROM contest_round_joiner_persons 
		WHERE contest_round_id = ? AND user_id = ? AND deleted_at IS NULL
	`, contestRoundId, userId).Scan(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		// Joiner already exists, no need to add
		return nil
	}

	// Create new joiner using raw SQL
	now := time.Now()
	return db.MasterDB.Exec(`
		INSERT INTO contest_round_joiner_persons (contest_round_id, user_id, created_at, created_by, deleted_at, deleted_by) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, contestRoundId, userId, now, createdBy, nil, 0).Error
}

// GetPersonJoiners - Get all person joiners for a contest round with basic user info
func (r *contestRoundRepository) GetPersonJoiners(contestRoundId int64) (interface{}, error) {
	var joiners []struct {
		UserId     int64  `json:"user_id"`
		UserName   string `json:"user_name"`
		SchoolName string `json:"school_name"`
	}

	err := db.ReplicaDB.Raw(`
		SELECT 
			crjp.user_id,
			u.name as user_name,
			COALESCE(s.name, '') as school_name
		FROM contest_round_joiner_persons crjp
		LEFT JOIN users u ON crjp.user_id = u.id
		LEFT JOIN schools s ON u.school_id = s.id AND s.deleted_at IS NULL
		WHERE crjp.contest_round_id = ? AND crjp.deleted_at IS NULL
		ORDER BY crjp.created_at DESC
	`, contestRoundId).Scan(&joiners).Error

	if err != nil {
		return nil, err
	}

	// Transform to array of user info objects
	var result []map[string]interface{}
	for _, joiner := range joiners {
		result = append(result, map[string]interface{}{
			"user_id":     joiner.UserId,
			"user_name":   joiner.UserName,
			"school_name": joiner.SchoolName,
		})
	}

	return result, nil
}

// RemovePersonJoiner - Remove a person joiner from contest round
func (r *contestRoundRepository) RemovePersonJoiner(contestRoundId int64, userId int64, deletedBy int64) error {
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_persons 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND user_id = $4
	`, time.Now(), deletedBy, contestRoundId, userId).Error
}

// BulkRemoveJoinerSchools - Xóa nhiều trường học khỏi joiner
func (r *contestRoundRepository) BulkRemoveJoinerSchools(contestRoundId int64, schoolIds []int64, deletedBy int64) error {
	if len(schoolIds) == 0 {
		return nil
	}
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_schools 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND school_id = ANY($4)
	`, time.Now(), deletedBy, contestRoundId, schoolIds).Error
}

// BulkRemoveJoinerProvinces - Xóa nhiều tỉnh thành khỏi joiner
func (r *contestRoundRepository) BulkRemoveJoinerProvinces(contestRoundId int64, provinceIds []int64, deletedBy int64) error {
	if len(provinceIds) == 0 {
		return nil
	}
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_provinces 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND province_id = ANY($4)
	`, time.Now(), deletedBy, contestRoundId, provinceIds).Error
}

// BulkRemoveJoinerClasses - Xóa nhiều lớp học khỏi joiner
func (r *contestRoundRepository) BulkRemoveJoinerClasses(contestRoundId int64, classIds []int64, deletedBy int64) error {
	if len(classIds) == 0 {
		return nil
	}
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_classes 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND class_id = ANY($4)
	`, time.Now(), deletedBy, contestRoundId, classIds).Error
}

// BulkRemoveJoinerPersons - Xóa nhiều học sinh khỏi joiner
func (r *contestRoundRepository) BulkRemoveJoinerPersons(contestRoundId int64, userIds []int64, deletedBy int64) error {
	if len(userIds) == 0 {
		return nil
	}
	return db.MasterDB.Exec(`
		UPDATE contest_round_joiner_persons 
		SET deleted_at = $1, deleted_by = $2 
		WHERE contest_round_id = $3 AND user_id = ANY($4)
	`, time.Now(), deletedBy, contestRoundId, userIds).Error
}
