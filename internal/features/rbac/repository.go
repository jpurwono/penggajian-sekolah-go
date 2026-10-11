package rbac

import (
	"context"

	"gorm.io/gorm"
)

type ListUsersFilter struct {
	Page    int64
	PerPage int64
	Q       string
	RoleID  *int64
	Active  *bool
}

type ListRolesFilter struct {
	Page    int64
	PerPage int64
	Q       string
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListUsers(ctx context.Context, f ListUsersFilter) ([]*User, int64, error) {
	query := r.db.WithContext(ctx).Model(&User{}).Joins("JOIN roles ON roles.id = users.role_id")

	if f.Q != "" {
		like := "%" + f.Q + "%"
		query = query.Where("users.username ILIKE ? OR users.email ILIKE ? OR users.full_name ILIKE ?", like, like, like)
	}
	if f.RoleID != nil {
		query = query.Where("users.role_id = ?", *f.RoleID)
	}
	if f.Active != nil {
		query = query.Where("users.is_active = ?", *f.Active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*User
	err := query.
		Preload("Role").
		Order("users.id ASC").
		Offset(int((f.Page - 1) * f.PerPage)).
		Limit(int(f.PerPage)).
		Find(&users).Error
	return users, total, err
}

func (r *Repository) FindUser(ctx context.Context, id int64) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Preload("Role").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *Repository) UpdateUser(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Model(u).
		Select("username", "email", "full_name", "role_id", "employee_id", "is_active", "updated_at").
		Updates(u).Error
}

func (r *Repository) EmployeeExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("employees").
		Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) UpdateUserPassword(ctx context.Context, id int64, hash string) error {
	return r.db.WithContext(ctx).Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hash, id).Error
}

func (r *Repository) SoftDeleteUser(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *Repository) UsernameExists(ctx context.Context, username string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&User{}).
		Where("username = ? AND id <> ?", username, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) EmailExists(ctx context.Context, email string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&User{}).
		Where("email = ? AND id <> ?", email, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListRoles(ctx context.Context, f ListRolesFilter) ([]*Role, int64, error) {
	query := r.db.WithContext(ctx).Model(&Role{})
	if f.Q != "" {
		like := "%" + f.Q + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []*Role
	err := query.
		Preload("Permissions").
		Order("id ASC").
		Offset(int((f.Page - 1) * f.PerPage)).
		Limit(int(f.PerPage)).
		Find(&roles).Error
	return roles, total, err
}

func (r *Repository) FindRole(ctx context.Context, id int64) (*Role, error) {
	var role Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) CreateRole(ctx context.Context, role *Role, permissionIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var perms []Permission
		if len(permissionIDs) > 0 {
			if err := tx.Find(&perms, permissionIDs).Error; err != nil {
				return err
			}
		}
		role.Permissions = perms
		return tx.Create(role).Error
	})
}

func (r *Repository) UpdateRole(ctx context.Context, role *Role, permissionIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(role).
			Select("name", "description", "updated_at").
			Updates(role).Error; err != nil {
			return err
		}

		var perms []Permission
		if len(permissionIDs) > 0 {
			if err := tx.Find(&perms, permissionIDs).Error; err != nil {
				return err
			}
		}
		return tx.Model(role).Association("Permissions").Replace(perms)
	})
}

func (r *Repository) SoftDeleteRole(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Role{ID: id}).Association("Permissions").Clear(); err != nil {
			return err
		}
		return tx.Delete(&Role{}, id).Error
	})
}

func (r *Repository) CountUsersByRole(ctx context.Context, roleID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *Repository) RoleNameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&Role{}).
		Where("name = ? AND id <> ?", name, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) RoleExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Role{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListPermissions(ctx context.Context) ([]*Permission, error) {
	var perms []*Permission
	err := r.db.WithContext(ctx).Order("module ASC, code ASC").Find(&perms).Error
	return perms, err
}
