package rbac

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
	"penggajian/server/internal/platform/security"
)

var (
	ErrNotFound       = errors.New("data tidak ditemukan")
	ErrRoleInUse      = errors.New("role masih dipakai user, tidak dapat dihapus")
	ErrCannotEditSelf = errors.New("tidak dapat mengubah role/status akun sendiri")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context, f ListUsersFilter) (*UserListResult, error) {
	users, total, err := s.repo.ListUsers(ctx, f)
	if err != nil {
		return nil, err
	}
	resp := make([]*UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toUserResponse(u))
	}
	return &UserListResult{Users: resp, Meta: httpx.NewMeta(f.Page, f.PerPage, total)}, nil
}

func (s *Service) GetUser(ctx context.Context, id int64) (*UserResponse, error) {
	u, err := s.repo.FindUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserResponse(u), nil
}

func (s *Service) resolveEmployee(ctx context.Context, employeeID *int64) (*int64, error) {
	if employeeID == nil || *employeeID <= 0 {
		return nil, nil
	}
	exists, err := s.repo.EmployeeExists(ctx, *employeeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("pegawai terkait tidak ditemukan")
	}
	return employeeID, nil
}

func (s *Service) validateUser(ctx context.Context, username, email string, roleID, excludeID int64) error {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	if len(username) < 3 {
		return errors.New("username minimal 3 karakter")
	}
	if !strings.Contains(email, "@") || len(email) < 5 {
		return errors.New("format email tidak valid")
	}
	if roleID <= 0 {
		return errors.New("role wajib dipilih")
	}

	if exists, err := s.repo.UsernameExists(ctx, username, excludeID); err != nil {
		return err
	} else if exists {
		return errors.New("username sudah digunakan")
	}
	if exists, err := s.repo.EmailExists(ctx, email, excludeID); err != nil {
		return err
	} else if exists {
		return errors.New("email sudah digunakan")
	}
	if exists, err := s.repo.RoleExists(ctx, roleID); err != nil {
		return err
	} else if !exists {
		return errors.New("role tidak ditemukan")
	}
	return nil
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {
	if len(req.Password) < 6 {
		return nil, errors.New("password minimal 6 karakter")
	}
	if err := s.validateUser(ctx, req.Username, req.Email, req.RoleID, 0); err != nil {
		return nil, err
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	employeeID, err := s.resolveEmployee(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(req.FullName),
		RoleID:       req.RoleID,
		EmployeeID:   employeeID,
		IsActive:     isActive,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return s.GetUser(ctx, user.ID)
}

func (s *Service) UpdateUser(ctx context.Context, id, actorID int64, req UpdateUserRequest) (*UserResponse, error) {
	if id == actorID && req.RoleID > 0 {
		existing, err := s.repo.FindUser(ctx, id)
		if err != nil {
			return nil, err
		}
		if req.RoleID != existing.RoleID || (req.IsActive != nil && *req.IsActive != existing.IsActive) {
			return nil, ErrCannotEditSelf
		}
	}
	if err := s.validateUser(ctx, req.Username, req.Email, req.RoleID, id); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	employeeID, err := s.resolveEmployee(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}

	user.Username = strings.TrimSpace(req.Username)
	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user.FullName = strings.TrimSpace(req.FullName)
	user.RoleID = req.RoleID
	user.EmployeeID = employeeID
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return s.GetUser(ctx, id)
}

func (s *Service) UpdateUserPassword(ctx context.Context, id int64, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password minimal 6 karakter")
	}
	if _, err := s.repo.FindUser(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdateUserPassword(ctx, id, hash)
}

func (s *Service) DeleteUser(ctx context.Context, id, actorID int64) error {
	if id == actorID {
		return errors.New("tidak dapat menghapus akun sendiri")
	}
	if _, err := s.repo.FindUser(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	return s.repo.SoftDeleteUser(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context, f ListRolesFilter) (*RoleListResult, error) {
	roles, total, err := s.repo.ListRoles(ctx, f)
	if err != nil {
		return nil, err
	}
	resp := make([]*RoleResponse, 0, len(roles))
	for _, role := range roles {
		count, err := s.repo.CountUsersByRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		resp = append(resp, toRoleResponse(role, count))
	}
	return &RoleListResult{Roles: resp, Meta: httpx.NewMeta(f.Page, f.PerPage, total)}, nil
}

func (s *Service) GetRole(ctx context.Context, id int64) (*RoleResponse, error) {
	role, err := s.repo.FindRole(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	count, err := s.repo.CountUsersByRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return toRoleResponse(role, count), nil
}

func (s *Service) validateRole(ctx context.Context, name string, id int64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nama role wajib diisi")
	}
	if exists, err := s.repo.RoleNameExists(ctx, name, id); err != nil {
		return err
	} else if exists {
		return errors.New("nama role sudah digunakan")
	}
	return nil
}

func (s *Service) CreateRole(ctx context.Context, req SaveRoleRequest) (*RoleResponse, error) {
	if err := s.validateRole(ctx, req.Name, 0); err != nil {
		return nil, err
	}

	role := &Role{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if err := s.repo.CreateRole(ctx, role, req.PermissionIDs); err != nil {
		return nil, err
	}
	return s.GetRole(ctx, role.ID)
}

func (s *Service) UpdateRole(ctx context.Context, id int64, req SaveRoleRequest) (*RoleResponse, error) {
	if _, err := s.repo.FindRole(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := s.validateRole(ctx, req.Name, id); err != nil {
		return nil, err
	}

	role, err := s.repo.FindRole(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Name = strings.TrimSpace(req.Name)
	role.Description = strings.TrimSpace(req.Description)
	if err := s.repo.UpdateRole(ctx, role, req.PermissionIDs); err != nil {
		return nil, err
	}
	return s.GetRole(ctx, id)
}

func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	if _, err := s.repo.FindRole(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	count, err := s.repo.CountUsersByRole(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrRoleInUse
	}
	return s.repo.SoftDeleteRole(ctx, id)
}

func (s *Service) ListPermissions(ctx context.Context) ([]*PermissionResponse, error) {
	perms, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]*PermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, &PermissionResponse{ID: p.ID, Code: p.Code, Name: p.Name, Module: p.Module})
	}
	return resp, nil
}
