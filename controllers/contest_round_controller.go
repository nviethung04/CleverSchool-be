package controllers

import (
	"be-lms/prot"
	"be-lms/resources"
	"be-lms/services"
	"be-lms/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContestRoundController struct {
	service  services.ContestRoundService
	resource resources.ContestRoundResource
}

func NewContestRoundController(service services.ContestRoundService) *ContestRoundController {
	return &ContestRoundController{
		service:  service,
		resource: resources.NewContestRoundResource(),
	}
}

func (crc *ContestRoundController) GetAll(c *gin.Context) {
	// Parse query parameters
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	keyword := c.Query("keyword")
	status := c.Query("status")
	contestId := c.Query("contest_id")

	// Create filters map
	filters := make(map[string]interface{})
	if keyword != "" {
		filters["keyword"] = keyword
	}
	if status != "" {
		filters["status"] = status
	}
	if contestId != "" {
		filters["contest_id"] = contestId
	}

	rounds, totalCount, err := crc.service.GetAllWithFilters(c, page, limit, filters)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	formattedRounds := crc.resource.FormatContestRounds(rounds)
	response := &prot.ContestRoundListResponse{
		ContestRounds: formattedRounds,
		TotalCount:    totalCount,
	}

	utils.Respond(c, response, nil, "")
}

func (crc *ContestRoundController) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	round, err := crc.service.GetByID(c, id)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			utils.Respond(c, nil, err, "messages.no_records_found_by_id", 404)
			return
		}
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	formattedRound := crc.resource.FormatContestRound(round)
	response := &prot.ContestRoundResponse{
		ContestRound: formattedRound,
		Message:      "Contest round retrieved successfully",
	}

	utils.Respond(c, response, nil, "")
}

func (crc *ContestRoundController) Create(c *gin.Context) {
	var req prot.ContestRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	round, err := crc.service.Create(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_create_data")
		return
	}

	formattedRound := crc.resource.FormatContestRound(round)
	response := &prot.ContestRoundResponse{
		ContestRound: formattedRound,
		Message:      "Contest round created successfully",
	}

	utils.Respond(c, response, nil, "")
}

func (crc *ContestRoundController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req prot.ContestRoundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	round, err := crc.service.Update(c, id, &req)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			utils.Respond(c, nil, err, "messages.no_records_found_by_id", 404)
			return
		}
		utils.Respond(c, nil, err, "messages.error_update_data")
		return
	}

	formattedRound := crc.resource.FormatContestRound(round)
	response := &prot.ContestRoundResponse{
		ContestRound: formattedRound,
		Message:      "Contest round updated successfully",
	}

	utils.Respond(c, response, nil, "")
}

func (crc *ContestRoundController) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	err = crc.service.Delete(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_delete_data")
		return
	}

	utils.Respond(c, gin.H{"message": "Contest round deleted successfully"}, nil, "")
}

func (crc *ContestRoundController) Restore(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	err = crc.service.Restore(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_restore_data")
		return
	}

	utils.Respond(c, gin.H{"message": "Contest round restored successfully"}, nil, "")
}

func (crc *ContestRoundController) GetByContestId(c *gin.Context) {
	contestId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	rounds, err := crc.service.GetByContestId(c, contestId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	formattedRounds := crc.resource.FormatContestRounds(rounds)
	response := &prot.ContestRoundListResponse{
		ContestRounds: formattedRounds,
		TotalCount:    int64(len(rounds)),
	}

	utils.Respond(c, response, nil, "")
}

func (crc *ContestRoundController) GetContestRoundUsers(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	users, err := crc.service.GetContestRoundUsers(c, contestRoundId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	utils.Respond(c, users, nil, "")
}

func (crc *ContestRoundController) GetContestRoundJoiners(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	// Get the contest round to check join_level
	round, err := crc.service.GetByID(c, contestRoundId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	// Get all joiners
	allJoiners, err := crc.service.GetContestRoundJoiners(c, contestRoundId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	// Extract joiners based on join_level
	var joiners interface{}
	joinersMap, ok := allJoiners.(map[string]interface{})
	if ok {
		switch round.JoinLevel {
		case "school":
			joiners = joinersMap["schools"]
		case "province":
			joiners = joinersMap["provinces"]
		case "class":
			joiners = joinersMap["classes"]
		case "student":
			joiners = joinersMap["persons"]
		default:
			joiners = []interface{}{}
		}
	} else {
		joiners = []interface{}{}
	}

	utils.Respond(c, joiners, nil, "")
}

// AddJoiner - Add joiners to contest round (supports student/class/school)
func (crc *ContestRoundController) AddJoiner(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req struct {
		JoinLevel string  `json:"join_level" binding:"required"`
		JoinerIds []int64 `json:"joiner_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	// Validate join_level
	if req.JoinLevel != "student" && req.JoinLevel != "class" && req.JoinLevel != "school" && req.JoinLevel != "province" {
		utils.Respond(c, nil, fmt.Errorf("invalid join_level: must be student, class, school, or province"), "messages.invalid_request")
		return
	}

	// Map student to person for service call
	joinLevel := req.JoinLevel
	if joinLevel == "student" {
		joinLevel = "person"
	}

	err = crc.service.AddJoiner(c, contestRoundId, joinLevel, req.JoinerIds)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_create_data")
		return
	}

	utils.Respond(c, gin.H{"message": fmt.Sprintf("%s joiners added successfully", req.JoinLevel)}, nil, "")
}

// AddPersonJoiner - Add individual students to contest round
func (crc *ContestRoundController) AddPersonJoiner(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req struct {
		UserIds []int64 `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	err = crc.service.AddPersonJoiner(c, contestRoundId, req.UserIds)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_create_data")
		return
	}

	utils.Respond(c, gin.H{"message": "Person joiners added successfully"}, nil, "")
}

// GetPersonJoiners - Get all person joiners for a contest round
func (crc *ContestRoundController) GetPersonJoiners(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	joiners, err := crc.service.GetPersonJoiners(c, contestRoundId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	utils.Respond(c, joiners, nil, "")
}

// RemovePersonJoiner - Remove a person joiner from contest round
func (crc *ContestRoundController) RemovePersonJoiner(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	userId, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	err = crc.service.RemovePersonJoiner(c, contestRoundId, userId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_delete_data")
		return
	}

	utils.Respond(c, gin.H{"message": "Person joiner removed successfully"}, nil, "")
}

func (crc *ContestRoundController) UpdateJoiner(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, nil, fmt.Errorf("update joiner is temporarily disabled"), "messages.error_update_data")
}

func (crc *ContestRoundController) RemoveJoiner(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, nil, fmt.Errorf("remove joiner is temporarily disabled"), "messages.error_delete_data")
}

func (crc *ContestRoundController) GetContestRoundProvinces(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, []interface{}{}, nil, "")
}

func (crc *ContestRoundController) GetContestRoundSchools(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, []interface{}{}, nil, "")
}

func (crc *ContestRoundController) GetContestRoundClasses(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, []interface{}{}, nil, "")
}

func (crc *ContestRoundController) GetContestRoundStudents(c *gin.Context) {
	contestRoundId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	students, err := crc.service.GetContestRoundStudents(c, contestRoundId)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	utils.Respond(c, students, nil, "")
}

func (crc *ContestRoundController) GetContestRoundEligibleUsers(c *gin.Context) {
	// Disabled due to technical issues
	utils.Respond(c, []interface{}{}, nil, "")
}
