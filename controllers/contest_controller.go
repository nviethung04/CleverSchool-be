package controllers

import (
	"be-Clever School/prot"
	"be-Clever School/resources"
	"be-Clever School/services"
	"be-Clever School/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ContestController struct {
	service  services.ContestService
	resource resources.ContestResource
}

func NewContestController(service services.ContestService) *ContestController {
	return &ContestController{
		service:  service,
		resource: resources.NewContestResource(),
	}
}

// GetAll handles GET /api/manage/contests
func (cc *ContestController) GetAll(c *gin.Context) {
	contests, totalCount, err := cc.service.GetAll(c)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	formattedContests := cc.resource.FormatContests(contests)
	response := &prot.ContestListResponse{
		Contests:   formattedContests,
		TotalCount: totalCount,
	}

	utils.Respond(c, response, nil, "")
}

// GetByID handles GET /api/manage/contests/:id
func (cc *ContestController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	contest, err := cc.service.GetByID(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	formattedContest := cc.resource.FormatContest(contest)
	response := &prot.ContestResponse{
		Contest: formattedContest,
		Message: "Success",
	}

	utils.Respond(c, response, nil, "")
}

// Create handles POST /api/manage/contests
func (cc *ContestController) Create(c *gin.Context) {
	var req prot.ContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	contest, err := cc.service.Create(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_create_data")
		return
	}

	formattedContest := cc.resource.FormatContest(contest)
	response := &prot.ContestResponse{
		Contest: formattedContest,
		Message: "Contest created successfully",
	}

	utils.Respond(c, response, nil, "")
}

// Update handles PUT /api/manage/contests/:id
func (cc *ContestController) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	var req prot.ContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Respond(c, nil, err, "messages.invalid_request")
		return
	}

	req.Id = int64(id)
	contest, err := cc.service.Update(c, &req)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_update_data")
		return
	}

	formattedContest := cc.resource.FormatContest(contest)
	response := &prot.ContestResponse{
		Contest: formattedContest,
		Message: "Contest updated successfully",
	}

	utils.Respond(c, response, nil, "")
}

// Delete handles DELETE /api/manage/contests/:id
func (cc *ContestController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	err = cc.service.Delete(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_delete_data")
		return
	}

	utils.Respond(c, map[string]interface{}{
		"id":      id,
		"message": "Contest deleted successfully",
	}, nil, "")
}

// Restore handles POST /api/manage/contests/:id/restore
func (cc *ContestController) Restore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	contest, err := cc.service.Restore(c, id)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_restore_data")
		return
	}

	formattedContest := cc.resource.FormatContest(contest)
	response := &prot.ContestResponse{
		Contest: formattedContest,
		Message: "Contest restored successfully",
	}

	utils.Respond(c, response, nil, "")
}

// GetContestRounds handles GET /api/manage/contests/:id/rounds
func (cc *ContestController) GetContestRounds(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	rounds, err := cc.service.GetContestRounds(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	// Create a temporary contest round resource for formatting
	contestRoundResource := resources.NewContestRoundResource()
	formattedRounds := contestRoundResource.FormatContestRounds(rounds)
	response := &prot.ContestRoundListResponse{
		ContestRounds: formattedRounds,
		TotalCount:    int64(len(rounds)),
	}

	utils.Respond(c, response, nil, "")
}

// GetContestWithRounds handles GET /api/manage/contests/:id/with-rounds
func (cc *ContestController) GetContestWithRounds(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Respond(c, nil, err, "messages.invalid_id")
		return
	}

	contest, err := cc.service.GetContestWithRounds(c, int64(id))
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_data")
		return
	}

	formattedContest := cc.resource.FormatContest(contest)
	response := &prot.ContestResponse{
		Contest: formattedContest,
		Message: "Success",
	}

	utils.Respond(c, response, nil, "")
}

// GetUserContests handles GET /api/manage/contests/user
func (cc *ContestController) GetUserContests(c *gin.Context) {
	contests, totalCount, err := cc.service.GetUserContests(c)
	if err != nil {
		utils.Respond(c, nil, err, "messages.error_get_list_data")
		return
	}

	formattedContests := cc.resource.FormatContests(contests)
	response := &prot.ContestListResponse{
		Contests:   formattedContests,
		TotalCount: totalCount,
	}

	utils.Respond(c, response, nil, "")
}
