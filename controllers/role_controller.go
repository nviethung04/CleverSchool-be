package controllers

import (
	"be-cleverschool/i18n"
	"be-cleverschool/models"
	"be-cleverschool/prot"
	"be-cleverschool/resources"
	"be-cleverschool/services"
	"be-cleverschool/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	*GenericController[models.Role, prot.Role, *prot.Role]
	service services.RoleService
}

func NewRoleController(service services.RoleService) *RoleController {
	roleResource := resources.NewRoleResource()
	roleResourceAdapter := NewRoleResourceAdapter(roleResource)

	genericController := NewGenericController(
		service,
		roleResourceAdapter,
		func() *prot.Role {
			return &prot.Role{}
		},
		func(roles []*prot.Role, totalCount uint64) interface{} {
			return &prot.RolesResponse{
				Roles:       roles,
				TotalCount: totalCount,
			}
		},
	)

	return &RoleController{
		GenericController: genericController,
		service:           service,
	}
}

func (rc *RoleController) GetPermissions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	permissions, err := rc.service.GetPermissions(c, id)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_records_found")), "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, permissions, err, "")
}

func (rc *RoleController) GetAllPermissions(c *gin.Context) {
	permissions, err := rc.service.GetAllPermissions(c)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.no_records_found")), "messages.no_records_found", http.StatusNotFound)
		return
	}

	utils.Respond(c, permissions, err, "")
}

func (rc *RoleController) AssignPermissions(c *gin.Context) {
	permissions, err := rc.service.AssignPermissions(c)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.assign_permissions_failed")), "messages.assign_permissions_failed")
		return
	}

	utils.Respond(c, permissions, err, "")
}

func (rc *RoleController) GetDisplayPermissions(c *gin.Context) {
	permissions, err := rc.service.GetDisplayPermissions(c)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.assign_permissions_failed")), "messages.assign_permissions_failed")
		return
	}

	utils.Respond(c, permissions, err, "")
}

func (rc *RoleController) UpdateDisplayPermissions(c *gin.Context) {
	permissions, err := rc.service.UpdateDisplayPermissions(c)
	if err != nil {
		utils.Respond(c, nil, fmt.Errorf(i18n.Localize("messages.assign_permissions_failed")), "messages.assign_permissions_failed")
		return
	}

	utils.Respond(c, permissions, err, "")
}

