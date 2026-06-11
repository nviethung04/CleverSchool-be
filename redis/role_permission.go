package redis

import (
	"be-lms/database/db"
	"be-lms/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	permissionsCacheKeyPrefix = "permissions:role"
	permissionsCacheTTL       = time.Hour * 24 * 1
)

type RolePermissionRedis struct {
	roleId int
}

func NewRoleRedis(roleId int) *RolePermissionRedis {
	return &RolePermissionRedis{roleId: roleId}
}

func (rr *RolePermissionRedis) getPermissionsCacheKey() string {
	return fmt.Sprintf("%s:%d", permissionsCacheKeyPrefix, rr.roleId)
}

func (rr *RolePermissionRedis) GetPermissions() ([]string, error) {
	key := rr.getPermissionsCacheKey()
	var permList []string
	ctx := context.Background()

	cached, err := db.RedisClient.Get(ctx, key).Result()

	if err == nil {
		if err := json.Unmarshal([]byte(cached), &permList); err != nil {
			fmt.Printf("Error unmarshalling cached permissions for role %d: %v\n", rr.roleId, err)
			return rr.SetPermissions()
		}
		return permList, nil
	} else if err == redis.Nil {
		return rr.SetPermissions()
	} else {
		return permList, fmt.Errorf("failed to get permissions from cache for role %d: %w", rr.roleId, err)
	}
}

func (rr *RolePermissionRedis) SetPermissions() ([]string, error) {
	var role models.Role
	var permList []string
	ctx := context.Background()
	key := rr.getPermissionsCacheKey()

	err := db.MasterDB.Preload("Permissions").First(&role, rr.roleId).Error
	if err != nil {
		return permList, err
	}

	if err != nil {
		return permList, fmt.Errorf("failed to fetch role from database for role %d: %w", rr.roleId, err)
	}

	if err := db.MasterDB.Preload("Permissions").First(&role, rr.roleId).Error; err != nil {
		return permList, fmt.Errorf("failed to fetch permissions from database for role %d: %w", rr.roleId, err)
	}

	for _, perm := range role.Permissions {
		if perm.IsDisplay != nil && !*perm.IsDisplay {
			continue
		}
		permList = append(permList, perm.Permission)
	}

	permJson, err := json.Marshal(permList)
	if err != nil {
		return permList, fmt.Errorf("failed to marshal permissions to JSON for role %d: %w", rr.roleId, err)
	}

	err = db.RedisClient.Set(ctx, key, permJson, permissionsCacheTTL).Err()
	if err != nil {
		return permList, fmt.Errorf("failed to set permissions in cache for role %d: %w", rr.roleId, err)
	}

	return permList, nil
}

func (rr *RolePermissionRedis) ClearRolePermissionsCache() error {
	key := rr.getPermissionsCacheKey()
	ctx := context.Background()

	deleted, err := db.RedisClient.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to clear permissions cache for role %d: %w", rr.roleId, err)
	}
	if deleted > 0 {
		fmt.Printf("Cleared permissions cache for role %d (key: %s)\n", rr.roleId, key)
	} else {
		fmt.Printf("No permissions cache found for role %d (key: %s)\n", rr.roleId, key)
	}

	rr.SetPermissions()

	return nil
}

func (rr *RolePermissionRedis) ClearCache() error {
	key := rr.getPermissionsCacheKey()
	ctx := context.Background()

	err := db.RedisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to clear cache for role %d: %w", rr.roleId, err)
	}

	return nil
}
