package services

import (
	"be-lms/models"
	"be-lms/prot"
	"be-lms/redis"
	"be-lms/repositories"
	"be-lms/resources"
	"be-lms/utils"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

type RoleService interface {
	GetAll(c *gin.Context) ([]models.Role, int64, error)
	GetByID(c *gin.Context, id int) (*prot.Role, error)
	Create(c *gin.Context, req *prot.Role) (*models.Role, error)
	Update(c *gin.Context, req *prot.Role) (*models.Role, error)
	Delete(c *gin.Context, id int) error
	Restore(c *gin.Context, id int) (*models.Role, error)
	GetPermissions(c *gin.Context, id int) (*prot.PermissionGroups, error)
	GetAllPermissions(c *gin.Context) (*prot.RolePermissionsList, error)
	AssignPermissions(c *gin.Context) (*prot.RolePermissionsList, error)
	GetDisplayPermissions(c *gin.Context) (*prot.RolePermissionsDisplayList, error)
	UpdateDisplayPermissions(c *gin.Context) (*prot.RolePermissionsDisplayList, error)
}

type roleService struct {
	repo repositories.RoleRepository
}

func NewRoleService(repo repositories.RoleRepository) RoleService {
	return &roleService{repo}
}

func (s *roleService) GetAll(c *gin.Context) ([]models.Role, int64, error) {
	allowedFilters := []string{"status"}
	filter, page, perPage, keyword, sort, err := utils.ParsePaginationParams(c, allowedFilters)
	if err != nil {
		return nil, 0, err
	}

	s.repo.SetSearch(keyword, []string{"name", "id"})
	s.repo.SetFilter(filter)
	s.repo.SetLimit(perPage)
	s.repo.SetPage(page)
	s.repo.SetSort(sort)

	roles, rows, err := s.repo.FindAll()
	if err != nil {
		return nil, 0, err
	}

	return roles, rows, nil
}

func (s *roleService) GetByID(c *gin.Context, id int) (*prot.Role, error) {
	role, err := s.repo.FindByID(id)

	if err != nil {
		return nil, err
	}

	roleResource := resources.NewRoleResource()
	formattedRole := roleResource.FormatRole(role)

	return formattedRole, nil
}

func (s *roleService) Create(c *gin.Context, req *prot.Role) (*models.Role, error) {
	roleResource := resources.NewRoleResource()
	role := roleResource.FormatModelRole(req)

	s.repo.SetContext(c)
	s.repo.SetEnableRedis(true)

	err := s.repo.Create(role)
	if err != nil {
		return nil, err
	}

	id := int(role.ID)
	newRole, _ := s.repo.FindNewByID(id)

	return newRole, nil
}

func (s *roleService) Update(c *gin.Context, req *prot.Role) (*models.Role, error) {
	roleResource := resources.NewRoleResource()
	role := roleResource.FormatModelRole(req)

	s.repo.SetContext(c)
	s.repo.SetEnableRedis(true)

	err := s.repo.Update(role)
	if err != nil {
		return nil, err
	}

	id := int(role.ID)
	updateRole, _ := s.repo.FindNewByID(id)

	return updateRole, nil
}

func (s *roleService) Delete(c *gin.Context, id int) error {
	s.repo.SetEnableRedis(true)
	err := s.repo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *roleService) Restore(c *gin.Context, id int) (*models.Role, error) {
	s.repo.SetContext(c)
	role, err := s.repo.Restore(int(id))
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleService) GetPermissions(c *gin.Context, id int) (*prot.PermissionGroups, error) {
	permissions, err := s.repo.GetPermissions()
	if err != nil {
		return nil, err
	}

	permissionGroups, err := redis.RememberCache[*prot.PermissionGroups](fmt.Sprintf("models.Role:%d:permissions", id), 24*time.Hour, func() (*prot.PermissionGroups, error) {
		role, err := s.repo.FindByID(id)
		if err != nil {
			return nil, err
		}

		permissionGroups := PermissionGroupsByRole(role, permissions)
		return permissionGroups, nil
	})

	return permissionGroups, err
}

func (s *roleService) GetAllPermissions(c *gin.Context) (*prot.RolePermissionsList, error) {
	permissions, err := s.repo.GetNewPermissions()
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.FindAllWithPermissions()
	if err != nil {
		return nil, err
	}

	rolePermissionsList := &prot.RolePermissionsList{
		RolePermissions: []*prot.RolePermission{},
	}

	for _, role := range roles {
		protoPermissions := PermissionGroupsByRole(&role, permissions)
		rolePermissionsMap := make(map[string]*prot.PermissionGroup)
		if protoPermissions != nil && protoPermissions.Permissions != nil {
			for _, entry := range protoPermissions.Permissions {
				rolePermissionsMap[entry.Key] = entry.Value
			}
		}

		rolePermissionsList.RolePermissions = append(rolePermissionsList.RolePermissions, &prot.RolePermission{
			Role: &prot.Role{
				Id:   int64(role.ID),
				Name: role.Name,
			},
			Permissions: rolePermissionsMap,
		})
	}

	rolePermissionsList.Permissions = GetPermissionGroupAction(permissions)

	return rolePermissionsList, nil
}

func (s *roleService) AssignPermissions(c *gin.Context) (*prot.RolePermissionsList, error) {
	var input prot.RolePermissionsList

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}

	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}

	if err := unmarshalOpts.Unmarshal(body, &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}

	for _, rolePermission := range input.RolePermissions {
		role, _ := s.repo.FindByID(int(rolePermission.Role.Id))

		var permissionKeys []string
		for resource, group := range rolePermission.Permissions {
			for action, enabled := range group.Actions {
				if enabled {
					permissionKeys = append(permissionKeys, fmt.Sprintf("%s.%s", resource, action))
				}
			}
		}

		permissions, err := s.repo.GetPermissionsByKeys(permissionKeys)
		if err != nil {
			return nil, err
		}

		if err := s.repo.UpdatePermissions(int(role.ID), permissions); err != nil {
			return nil, err
		}

		redis := redis.NewRoleRedis(int(role.ID))
		redis.ClearRolePermissionsCache()
	}

	for _, permission := range input.Permissions {
		for _, action := range permission.Actions {
			permissionKey := fmt.Sprintf("%s.%s", permission.Key, action.Key)

			s.repo.UpdatePermission(permissionKey, models.Permission{
				Permission: permissionKey,
				Name:       action.Name,
				Group:      permission.Name,
			})
		}
	}

	permissions, err := s.repo.GetNewPermissions()
	if err != nil {
		return nil, err
	}

	roles, err := s.repo.FindAllWithPermissions()
	if err != nil {
		return nil, err
	}

	rolePermissionsList := &prot.RolePermissionsList{
		RolePermissions: []*prot.RolePermission{},
	}

	for _, role := range roles {
		protoPermissions := PermissionGroupsByRole(&role, permissions)
		rolePermissionsMap := make(map[string]*prot.PermissionGroup)
		if protoPermissions != nil && protoPermissions.Permissions != nil {
			for _, entry := range protoPermissions.Permissions {
				rolePermissionsMap[entry.Key] = entry.Value
			}
		}
		rolePermissionsList.RolePermissions = append(rolePermissionsList.RolePermissions, &prot.RolePermission{
			Role:        &prot.Role{Id: int64(role.ID), Name: role.Name},
			Permissions: rolePermissionsMap,
		})
	}

	rolePermissionsList.Permissions = GetPermissionGroupAction(permissions)

	redis.ClearCacheByPrefix("models.Role")

	return rolePermissionsList, nil
}

func (s *roleService) GetDisplayPermissions(c *gin.Context) (*prot.RolePermissionsDisplayList, error) {
	permissions, err := s.repo.GetNewPermissions()
	if err != nil {
		return nil, err
	}

	groupMap := make(map[string]*prot.PermissionGroupDisplay)

	for _, p := range permissions {
		parts := strings.Split(p.Permission, ".")
		if len(parts) != 2 {
			continue
		}

		resource := parts[0]
		key := parts[1]

		groupKey := p.Group
		if _, ok := groupMap[groupKey]; !ok {
			groupMap[groupKey] = &prot.PermissionGroupDisplay{
				Key:          resource,
				Name:         p.Group,
				SortPosition: int32(p.SortPosition),
				Actions:      []*prot.RoleActionDisplay{},
			}
		}

		action := &prot.RoleActionDisplay{
			Key:       key,
			Name:      p.Name,
			IsDisplay: p.IsDisplay != nil && *p.IsDisplay,
		}

		groupMap[groupKey].Actions = append(groupMap[groupKey].Actions, action)
	}

	var groups []*prot.PermissionGroupDisplay
	for _, g := range groupMap {
		sort.Slice(g.Actions, func(i, j int) bool {
			return g.Actions[i].Key < g.Actions[j].Key
		})

		groups = append(groups, g)
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SortPosition < groups[j].SortPosition
	})

	return &prot.RolePermissionsDisplayList{
		Permissions: groups,
	}, nil
}

func (s *roleService) UpdateDisplayPermissions(c *gin.Context) (*prot.RolePermissionsDisplayList, error) {
	var req prot.RolePermissionsDisplayList
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	for _, group := range req.Permissions {
		resource := group.Key

		for _, action := range group.Actions {
			permKey := fmt.Sprintf("%s.%s", resource, action.Key)

			err := s.repo.UpdatePermissionDisplayStatus(permKey, action.IsDisplay)
			if err != nil {
				return nil, fmt.Errorf("failed to update %s: %w", permKey, err)
			}
		}
	}

	redis.ClearCacheByPrefix("models.Role")
	redis.ClearCacheByPrefix("permissions:role")

	return &req, nil
}

func PermissionGroupsByRole(role *models.Role, permissions []models.Permission) *prot.PermissionGroups {
	existing := make(map[string]bool)
	for _, p := range role.Permissions {
		existing[p.Permission] = true
	}

	permissionGroupsMap := make(map[string]*prot.PermissionGroup)

	for _, perm := range permissions {
		parts := strings.Split(perm.Permission, ".")
		if len(parts) != 2 {
			continue
		}
		resource := parts[0]
		action := parts[1]

		group, ok := permissionGroupsMap[resource]
		if !ok {
			group = &prot.PermissionGroup{
				Name:         resource,
				SortPosition: int32(perm.SortPosition),
				Actions:      make(map[string]bool),
			}
			permissionGroupsMap[resource] = group
		}

		group.Actions[action] = existing[perm.Permission]
	}

	protoGroups := &prot.PermissionGroups{}
	for key, val := range permissionGroupsMap {
		protoGroups.Permissions = append(protoGroups.Permissions, &prot.PermissionGroupEntry{
			Key:   key,
			Value: val,
		})
	}

	sort.Slice(protoGroups.Permissions, func(i, j int) bool {
		return protoGroups.Permissions[i].Value.SortPosition < protoGroups.Permissions[j].Value.SortPosition
	})

	return protoGroups
}

func GetPermissionGroupAction(permissions []models.Permission) []*prot.PermissionGroupAction {
	allPermissions := make([]*prot.PermissionGroupAction, 0, len(permissions))
	permissionGroupsMap := make(map[string]*prot.PermissionGroupAction)

	for _, permission := range permissions {
		parts := strings.Split(permission.Permission, ".")
		if len(parts) != 2 {
			continue
		}

		resource := parts[0]

		roleAction := &prot.RoleAction{
			Key:  parts[1],
			Name: permission.Name,
		}

		group, exists := permissionGroupsMap[resource]
		if !exists {
			group = &prot.PermissionGroupAction{
				Key:          resource,
				Name:         permission.Group,
				SortPosition: int32(permission.SortPosition),
				Actions:      []*prot.RoleAction{},
			}
			permissionGroupsMap[resource] = group
			allPermissions = append(allPermissions, group)
		}

		group.Actions = append(group.Actions, roleAction)
	}

	sort.Slice(allPermissions, func(i, j int) bool {
		return allPermissions[i].SortPosition < allPermissions[j].SortPosition
	})

	return allPermissions
}
