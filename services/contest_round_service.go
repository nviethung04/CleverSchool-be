package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ContestRoundService interface {
	GetAll(c *gin.Context) ([]*models.ContestRound, int64, error)
	GetAllWithFilters(c *gin.Context, page string, limit string, filters map[string]interface{}) ([]*models.ContestRound, int64, error)
	GetByID(c *gin.Context, id int64) (*models.ContestRound, error)
	Create(c *gin.Context, req *prot.ContestRoundRequest) (*models.ContestRound, error)
	Update(c *gin.Context, id int64, req *prot.ContestRoundRequest) (*models.ContestRound, error)
	Delete(c *gin.Context, id int64) error
	Restore(c *gin.Context, id int64) error
	GetByContestId(c *gin.Context, contestId int64) ([]*models.ContestRound, error)
	GetContestRoundUsers(c *gin.Context, contestRoundId int64) ([]*models.User, error)
	GetContestRoundJoiners(c *gin.Context, contestRoundId int64) (interface{}, error)
	AddJoiner(c *gin.Context, contestRoundId int64, joinLevel string, joinerIds []int64) error
	RemoveJoiner(c *gin.Context, contestRoundId int64, joinLevel string, joinerId int64) error
	UpdateJoiner(c *gin.Context, contestRoundId int64, joinLevel string, oldJoinerId int64, newJoinerIds []int64) error
	GetContestRoundProvinces(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundSchools(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundClasses(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundStudents(c *gin.Context, contestRoundId int64) (interface{}, error)
	GetContestRoundEligibleUsers(c *gin.Context, contestRoundId int64) ([]*models.User, error)

	// New person joiner methods
	AddPersonJoiner(c *gin.Context, contestRoundId int64, userIds []int64) error
	GetPersonJoiners(c *gin.Context, contestRoundId int64) (interface{}, error)
	RemovePersonJoiner(c *gin.Context, contestRoundId int64, userId int64) error

	// Bulk remove methods
	BulkRemoveJoinerSchools(c *gin.Context, contestRoundId int64, schoolIds []int64) error
	BulkRemoveJoinerProvinces(c *gin.Context, contestRoundId int64, provinceIds []int64) error
	BulkRemoveJoinerClasses(c *gin.Context, contestRoundId int64, classIds []int64) error
	BulkRemoveJoinerPersons(c *gin.Context, contestRoundId int64, userIds []int64) error
}

type contestRoundService struct {
	repo repositories.ContestRoundRepository
}

func NewContestRoundService(repo repositories.ContestRoundRepository) ContestRoundService {
	return &contestRoundService{
		repo: repo,
	}
}

func (s *contestRoundService) GetAll(c *gin.Context) ([]*models.ContestRound, int64, error) {
	s.repo.SetContext(c)
	// Set sort by ID ascending
	s.repo.SetSort(map[string]string{"id": "asc"})
	rounds, total, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	// Convert slice of structs to slice of pointers
	result := make([]*models.ContestRound, len(rounds))
	for i := range rounds {
		result[i] = &rounds[i]
	}

	return result, total, nil
}

func (s *contestRoundService) GetAllWithFilters(c *gin.Context, page string, limit string, filters map[string]interface{}) ([]*models.ContestRound, int64, error) {
	s.repo.SetContext(c)

	// Convert page and limit to int
	pageInt := 1
	limitInt := 20
	if p, err := fmt.Sscanf(page, "%d", &pageInt); err == nil && p > 0 {
		// page parsed successfully
	}
	if l, err := fmt.Sscanf(limit, "%d", &limitInt); err == nil && l > 0 {
		// limit parsed successfully
	}

	// Set pagination
	s.repo.SetPage(pageInt)
	s.repo.SetLimit(limitInt)

	// Set sort by ID descending (newest first)
	s.repo.SetSort(map[string]string{"id": "desc"})

	// Apply filters
	filterMap := make(map[string]interface{})

	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		// Search in name and description
		s.repo.SetSearch(keyword, []string{"name", "description"})
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		filterMap["status"] = status
	}

	if contestId, ok := filters["contest_id"].(string); ok && contestId != "" {
		filterMap["contest_id"] = contestId
	}

	if len(filterMap) > 0 {
		s.repo.SetFilter(filterMap)
	}

	rounds, total, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	// Convert slice of structs to slice of pointers
	result := make([]*models.ContestRound, len(rounds))
	for i := range rounds {
		result[i] = &rounds[i]
	}

	return result, total, nil
}

func (s *contestRoundService) GetByID(c *gin.Context, id int64) (*models.ContestRound, error) {
	s.repo.SetContext(c)
	round, err := s.repo.FindByID(int(id))
	if err != nil {
		// Check if it's a "record not found" error
		errMsg := err.Error()
		if strings.Contains(errMsg, "no_records_found_by_id") ||
			strings.Contains(errMsg, "record not found") ||
			strings.Contains(errMsg, "Không tìm thấy bản ghi") ||
			strings.Contains(errMsg, "No record found") {
			return nil, fmt.Errorf("contest round with ID %d not found", id)
		}
		return nil, err
	}
	return round, nil
}

func (s *contestRoundService) Create(c *gin.Context, req *prot.ContestRoundRequest) (*models.ContestRound, error) {
	contestRoundResource := resources.NewContestRoundResource()
	round := contestRoundResource.FormatModelContestRound(req)
	round.CreatedBy = int64(utils.GetCurrentUserId(c))
	round.CreatedAt = time.Now()
	round.UpdatedAt = time.Now()

	s.repo.SetContext(c)
	err := s.repo.Create(round)
	if err != nil {
		return nil, err
	}
	return round, nil
}

func (s *contestRoundService) Update(c *gin.Context, id int64, req *prot.ContestRoundRequest) (*models.ContestRound, error) {
	// First, get the existing record
	s.repo.SetContext(c)
	existingRound, err := s.repo.FindByID(int(id))
	if err != nil {
		// Check if it's a "record not found" error
		errMsg := err.Error()
		if strings.Contains(errMsg, "no_records_found_by_id") ||
			strings.Contains(errMsg, "record not found") ||
			strings.Contains(errMsg, "Không tìm thấy bản ghi") ||
			strings.Contains(errMsg, "No record found") {
			return nil, fmt.Errorf("contest round with ID %d not found", id)
		}
		return nil, err
	}

	// Update only the fields that are provided in the request
	if req.Name != "" {
		existingRound.Name = req.Name
	}
	if req.Description != "" {
		existingRound.Description = req.Description
	}
	if req.ContestId > 0 {
		existingRound.ContestId = req.ContestId
	}
	if req.SortPosition > 0 {
		existingRound.SortPosition = req.SortPosition
	}
	if req.JoinLevel != "" {
		existingRound.JoinLevel = req.JoinLevel
	}
	if req.StartTime > 0 {
		startTime := time.Unix(req.StartTime, 0)
		existingRound.StartTime = &startTime
	}
	if req.EndTime > 0 {
		endTime := time.Unix(req.EndTime, 0)
		existingRound.EndTime = &endTime
	}

	// Set update fields
	existingRound.UpdatedBy = int64(utils.GetCurrentUserId(c))
	existingRound.UpdatedAt = time.Now()

	// Update the record
	err = s.repo.Update(existingRound)
	if err != nil {
		return nil, err
	}
	return existingRound, nil
}

func (s *contestRoundService) Delete(c *gin.Context, id int64) error {
	s.repo.SetContext(c)
	return s.repo.Delete(int(id))
}

func (s *contestRoundService) Restore(c *gin.Context, id int64) error {
	s.repo.SetContext(c)
	_, err := s.repo.Restore(int(id))
	return err
}

func (s *contestRoundService) GetByContestId(c *gin.Context, contestId int64) ([]*models.ContestRound, error) {
	s.repo.SetContext(c)
	rounds, err := s.repo.GetByContestId(contestId)
	if err != nil {
		return nil, err
	}

	// Convert slice of structs to slice of pointers
	result := make([]*models.ContestRound, len(rounds))
	for i := range rounds {
		result[i] = &rounds[i]
	}

	return result, nil
}

func (s *contestRoundService) GetContestRoundUsers(c *gin.Context, contestRoundId int64) ([]*models.User, error) {
	s.repo.SetContext(c)
	_, err := s.repo.GetContestRoundUsers(contestRoundId)
	if err != nil {
		return nil, err
	}

	// Convert to User models - this needs to be implemented properly
	// For now, return empty slice
	return []*models.User{}, nil
}

func (s *contestRoundService) GetContestRoundJoiners(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)

	// Get all joiners by level
	result := make(map[string]interface{})

	// Get school joiners and transform to flat structure
	schoolJoiners, err := s.repo.GetContestRoundJoinerSchools(contestRoundId)
	if err != nil {
		result["schools_error"] = err.Error()
		result["schools"] = []interface{}{}
	} else {
		// Transform to flat structure
		var schools []map[string]interface{}
		for _, joiner := range schoolJoiners {
			schools = append(schools, map[string]interface{}{
				"id":         joiner.SchoolId,
				"name":       joiner.School.Name,
				"address":    joiner.School.AddressVN,
				"created_at": joiner.CreatedAt,
			})
		}
		result["schools"] = schools
	}

	// Get province joiners and transform
	provinceJoiners, err := s.repo.GetContestRoundJoinerProvinces(contestRoundId)
	if err != nil {
		result["provinces_error"] = err.Error()
		result["provinces"] = []interface{}{}
	} else {
		var provinces []map[string]interface{}
		for _, joiner := range provinceJoiners {
			provinces = append(provinces, map[string]interface{}{
				"id":         joiner.ProvinceId,
				"name":       joiner.Province.Name,
				"created_at": joiner.CreatedAt,
			})
		}
		result["provinces"] = provinces
	}

	// Get person joiners (already flat structure from repository)
	personJoiners, err := s.repo.GetPersonJoiners(contestRoundId)
	if err != nil {
		result["persons_error"] = err.Error()
		result["persons"] = []interface{}{}
	} else {
		// Transform to consistent format
		if joinersList, ok := personJoiners.([]map[string]interface{}); ok {
			var persons []map[string]interface{}
			for _, joiner := range joinersList {
				persons = append(persons, map[string]interface{}{
					"id":    joiner["user_id"],
					"name":  joiner["user_name"],
					"email": "", // TODO: Add email if needed
					"user_info": map[string]interface{}{
						"school": joiner["school_name"],
					},
				})
			}
			result["persons"] = persons
		} else {
			result["persons"] = personJoiners
		}
	}

	// Get class joiners and transform
	classJoiners, err := s.repo.GetContestRoundJoinerClasses(contestRoundId)
	if err != nil {
		result["classes_error"] = err.Error()
		result["classes"] = []interface{}{}
	} else {
		var classes []map[string]interface{}
		for _, joiner := range classJoiners {
			classes = append(classes, map[string]interface{}{
				"id":         joiner.ClassId,
				"name":       joiner.Class.Name,
				"created_at": joiner.CreatedAt,
			})
		}
		result["classes"] = classes
	}

	return result, nil
}

func (s *contestRoundService) AddJoiner(c *gin.Context, contestRoundId int64, joinLevel string, joinerIds []int64) error {
	s.repo.SetContext(c)
	createdBy := int64(utils.GetCurrentUserId(c))

	switch joinLevel {
	case "school":
		for _, schoolId := range joinerIds {
			if err := s.repo.AddJoinerSchool(contestRoundId, schoolId, createdBy); err != nil {
				return err
			}
		}
	case "province":
		for _, provinceId := range joinerIds {
			if err := s.repo.AddJoinerProvince(contestRoundId, provinceId, createdBy); err != nil {
				return err
			}
		}
	case "class":
		for _, classId := range joinerIds {
			if err := s.repo.AddJoinerClass(contestRoundId, classId, createdBy); err != nil {
				return err
			}
		}
	case "person":
		for _, userId := range joinerIds {
			if err := s.repo.AddJoinerPerson(contestRoundId, userId, createdBy); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *contestRoundService) RemoveJoiner(c *gin.Context, contestRoundId int64, joinLevel string, joinerId int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))

	switch joinLevel {
	case "school":
		return s.repo.RemoveJoinerSchool(contestRoundId, joinerId, deletedBy)
	case "province":
		return s.repo.RemoveJoinerProvince(contestRoundId, joinerId, deletedBy)
	case "class":
		return s.repo.RemoveJoinerClass(contestRoundId, joinerId, deletedBy)
	case "person":
		return s.repo.RemoveJoinerPerson(contestRoundId, joinerId, deletedBy)
	}

	return nil
}

func (s *contestRoundService) UpdateJoiner(c *gin.Context, contestRoundId int64, joinLevel string, oldJoinerId int64, newJoinerIds []int64) error {
	// Disabled due to technical issues
	return fmt.Errorf("update joiner is temporarily disabled")
}

func (s *contestRoundService) GetContestRoundProvinces(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)
	return s.repo.GetContestRoundProvinces(contestRoundId)
}

func (s *contestRoundService) GetContestRoundSchools(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)
	return s.repo.GetContestRoundSchools(contestRoundId)
}

func (s *contestRoundService) GetContestRoundClasses(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)
	return s.repo.GetContestRoundClasses(contestRoundId)
}

func (s *contestRoundService) GetContestRoundStudents(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)
	return s.repo.GetContestRoundStudents(contestRoundId)
}

func (s *contestRoundService) GetContestRoundEligibleUsers(c *gin.Context, contestRoundId int64) ([]*models.User, error) {
	// Completely disabled to avoid table name issues
	return []*models.User{}, nil
}

// AddPersonJoiner - Add individual students to contest round
func (s *contestRoundService) AddPersonJoiner(c *gin.Context, contestRoundId int64, userIds []int64) error {
	s.repo.SetContext(c)
	createdBy := int64(utils.GetCurrentUserId(c))

	for _, userId := range userIds {
		if err := s.repo.AddPersonJoiner(contestRoundId, userId, createdBy); err != nil {
			return err
		}
	}
	return nil
}

// GetPersonJoiners - Get all person joiners for a contest round
func (s *contestRoundService) GetPersonJoiners(c *gin.Context, contestRoundId int64) (interface{}, error) {
	s.repo.SetContext(c)
	return s.repo.GetPersonJoiners(contestRoundId)
}

// RemovePersonJoiner - Remove a person joiner from contest round
func (s *contestRoundService) RemovePersonJoiner(c *gin.Context, contestRoundId int64, userId int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))
	return s.repo.RemovePersonJoiner(contestRoundId, userId, deletedBy)
}

// BulkRemoveJoinerSchools - Bulk remove school joiners
func (s *contestRoundService) BulkRemoveJoinerSchools(c *gin.Context, contestRoundId int64, schoolIds []int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))
	return s.repo.BulkRemoveJoinerSchools(contestRoundId, schoolIds, deletedBy)
}

// BulkRemoveJoinerProvinces - Bulk remove province joiners
func (s *contestRoundService) BulkRemoveJoinerProvinces(c *gin.Context, contestRoundId int64, provinceIds []int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))
	return s.repo.BulkRemoveJoinerProvinces(contestRoundId, provinceIds, deletedBy)
}

// BulkRemoveJoinerClasses - Bulk remove class joiners
func (s *contestRoundService) BulkRemoveJoinerClasses(c *gin.Context, contestRoundId int64, classIds []int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))
	return s.repo.BulkRemoveJoinerClasses(contestRoundId, classIds, deletedBy)
}

// BulkRemoveJoinerPersons - Bulk remove person joiners
func (s *contestRoundService) BulkRemoveJoinerPersons(c *gin.Context, contestRoundId int64, userIds []int64) error {
	s.repo.SetContext(c)
	deletedBy := int64(utils.GetCurrentUserId(c))
	return s.repo.BulkRemoveJoinerPersons(contestRoundId, userIds, deletedBy)
}
