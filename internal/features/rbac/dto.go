package rbac

import (
	"time"

	"penggajian/server/internal/platform/httpx"
)

type UserResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	RoleID      int64      `json:"role_id"`
	RoleName    string     `json:"role_name,omitempty"`
	EmployeeID  *int64     `json:"employee_id"`
	EmployeeName string    `json:"employee_name,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type UserListResult struct {
	Users []*UserResponse `json:"users"`
	Meta  *httpx.Meta     `json:"meta"`
}

type CreateUserRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	RoleID     int64  `json:"role_id"`
	EmployeeID *int64 `json:"employee_id"`
	Password   string `json:"password"`
	IsActive   *bool  `json:"is_active"`
}

type UpdateUserRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	RoleID     int64  `json:"role_id"`
	EmployeeID *int64 `json:"employee_id"`
	IsActive   *bool  `json:"is_active"`
}

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type RoleResponse struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	UserCount      int64   `json:"user_count"`
	PermissionIDs  []int64 `json:"permission_ids"`
	PermissionName string  `json:"permission_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type RoleListResult struct {
	Roles []*RoleResponse `json:"roles"`
	Meta  *httpx.Meta     `json:"meta"`
}

type SaveRoleRequest struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PermissionIDs []int64 `json:"permission_ids"`
}

type PermissionResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Module string `json:"module"`
}

func toUserResponse(u *User) *UserResponse {
	resp := &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		RoleID:      u.RoleID,
		EmployeeID:  u.EmployeeID,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
	if u.Role != nil {
		resp.RoleName = u.Role.Name
	}
	return resp
}

func toRoleResponse(r *Role, userCount int64) *RoleResponse {
	permIDs := make([]int64, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		permIDs = append(permIDs, p.ID)
	}
	return &RoleResponse{
		ID:            r.ID,
		Name:          r.Name,
		Description:   r.Description,
		UserCount:     userCount,
		PermissionIDs: permIDs,
		CreatedAt:     r.CreatedAt,
	}
}
