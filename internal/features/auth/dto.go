package auth

import "penggajian/server/internal/platform/middleware"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserResponse struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	RoleID      int64    `json:"role_id"`
	RoleName    string   `json:"role_name"`
	EmployeeID  *int64   `json:"employee_id"`
	Permissions []string `json:"permissions"`
}

type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int64         `json:"expires_in"`
	User         *UserResponse `json:"user,omitempty"`
}

func toUserResponse(u *middleware.AuthUser) *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		RoleID:      u.RoleID,
		RoleName:    u.RoleName,
		EmployeeID:  u.EmployeeID,
		Permissions: u.Permissions,
	}
}
