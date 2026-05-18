package repositories

import (
	"be-Clever School/database/db"
	"be-Clever School/models"
	"be-Clever School/repositories/base"
)

type RoleRepository interface {
	base.BaseRepositoryInterface[models.Role]
	FindAllWithPermissions() ([]models.Role, error)
	GetPermissions() ([]models.Permission, error)
	GetNewPermissions() ([]models.Permission, error)
	GetPermissionsByKeys(permissionKeys []string) ([]models.Permission, error)
	UpdatePermissions(id int, permissions []models.Permission) error
	UpdatePermission(permissionKey string, permission models.Permission) error
	FindByName(name string) (*models.Role, error)
	UpdatePermissionDisplayStatus(permission string, isDisplay bool) error
	AddUserRole(roleId int64, userIds []int64) error
}

type roleRepository struct {
	*base.BaseRepository[models.Role]
}

func NewRoleRepository() RoleRepository {
	return &roleRepository{
		BaseRepository: base.NewBaseRepository[models.Role](),
	}
}

func (r *roleRepository) FindAllWithPermissions() ([]models.Role, error) {
	var roles []models.Role
	err := db.ReplicaDB.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByID(id int) (*models.Role, error) {
	var role models.Role
	err := db.ReplicaDB.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByName(name string) (*models.Role, error) {
	var role models.Role
	err := db.ReplicaDB.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindNewByID(id int) (*models.Role, error) {
	var role models.Role
	err := db.MasterDB.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := db.ReplicaDB.Order("sort_position").Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) GetNewPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := db.MasterDB.Order("sort_position").Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) GetPermissionsByKeys(permissionKeys []string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := db.ReplicaDB.Where("permission IN ?", permissionKeys).Find(&permissions).Error
	return permissions, err
}

func (r *roleRepository) UpdatePermissions(id int, permissions []models.Permission) error {
	role, err := r.FindByID(id)
	if err != nil {
		return err
	}

	// Lấy permission keys
	var keys []string
	for _, p := range permissions {
		keys = append(keys, p.Permission)
	}

	// Lấy permissions đã tồn tại, có ID
	var existingPermissions []models.Permission
	if err := db.MasterDB.
		Select("id", "permission", "name", "group").
		Where("permission IN ?", keys).
		Find(&existingPermissions).Error; err != nil {
		return err
	}

	// Xóa các permission hiện có của role trong bảng liên kết (ví dụ tên bảng role_permissions)
	if err := db.MasterDB.Where("role_id = ?", role.ID).Delete(&models.RolePermission{}).Error; err != nil {
		return err
	}

	// Tạo slice cho chèn mới vào bảng liên kết
	var rolePermissions []models.RolePermission
	for _, p := range existingPermissions {
		rolePermissions = append(rolePermissions, models.RolePermission{
			RoleID:       role.ID,
			PermissionID: p.ID,
		})
	}

	// Chèn mới các bản ghi vào bảng liên kết
	if len(rolePermissions) > 0 {
		if err := db.MasterDB.Create(&rolePermissions).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *roleRepository) UpdatePermission(permissionKey string, permission models.Permission) error {
	err := db.MasterDB.Model(&models.Permission{}).
		Where("permission = ?", permissionKey).
		Updates(map[string]interface{}{
			"name":  permission.Name,
			"group": permission.Group,
		}).Error
	return err
}

func (r *roleRepository) UpdatePermissionDisplayStatus(permission string, isDisplay bool) error {
	return db.MasterDB.Model(&models.Permission{}).
		Where("permission = ?", permission).
		Update("is_display", isDisplay).Error
}

func (r *roleRepository) AddUserRole(roleId int64, userIds []int64) error {
	tx := db.MasterDB

	for _, uid := range userIds {
		q := tx.Model(&models.UserRefRole{}).
			Where("role_id = ? AND user_id = ?", roleId, uid)

		var count int64
		if err := q.Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			uc := models.UserRefRole{UserId: uid, RoleId: roleId}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
