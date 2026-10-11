package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/middleware"
)

var ErrUserNotFound = errors.New("user tidak ditemukan")

// Session merepresentasikan baris user_sessions (refresh token di-hash SHA-256).
type Session struct {
	ID               int64      `gorm:"primaryKey"`
	UserID           int64      `gorm:"column:user_id"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash"`
	UserAgent        string     `gorm:"column:user_agent"`
	IP               string     `gorm:"column:ip"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (Session) TableName() string { return "user_sessions" }

type userRow struct {
	ID           int64
	Username     string
	Email        string
	FullName     string
	RoleID       int64
	RoleName     string
	PasswordHash string
	EmployeeID   *int64
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

const activeUserSelect = `
SELECT u.id, u.username, u.email, u.full_name, u.role_id, r.name AS role_name, u.password_hash, u.employee_id
FROM users u
JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL AND u.is_active = true `

func (r *Repository) loadPermissions(ctx context.Context, roleID int64) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.code
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`, roleID).Scan(&codes).Error
	return codes, err
}

func (r *Repository) rowToAuthUser(ctx context.Context, row userRow) (*middleware.AuthUser, error) {
	perms, err := r.loadPermissions(ctx, row.RoleID)
	if err != nil {
		return nil, err
	}
	return &middleware.AuthUser{
		ID:          row.ID,
		Username:    row.Username,
		Email:       row.Email,
		FullName:    row.FullName,
		RoleID:      row.RoleID,
		RoleName:    row.RoleName,
		EmployeeID:  row.EmployeeID,
		Permissions: perms,
	}, nil
}

// LoadActiveUserByID mengimplementasikan middleware.UserLoader.
func (r *Repository) LoadActiveUserByID(ctx context.Context, id int64) (*middleware.AuthUser, error) {
	var row userRow
	err := r.db.WithContext(ctx).Raw(activeUserSelect+` AND u.id = ?`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, ErrUserNotFound
	}
	return r.rowToAuthUser(ctx, row)
}

// FindActiveByIdentifier mencari user aktif berdasarkan username ATAU email (untuk login).
func (r *Repository) FindActiveByIdentifier(ctx context.Context, identifier string) (*middleware.AuthUser, string, error) {
	var row userRow
	err := r.db.WithContext(ctx).Raw(
		activeUserSelect+` AND (u.username = ? OR u.email = ?)`, identifier, identifier,
	).Scan(&row).Error
	if err != nil {
		return nil, "", err
	}
	if row.ID == 0 {
		return nil, "", ErrUserNotFound
	}
	user, err := r.rowToAuthUser(ctx, row)
	if err != nil {
		return nil, "", err
	}
	return user, row.PasswordHash, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Exec(`UPDATE users SET last_login_at = now() WHERE id = ?`, userID).Error
}

func (r *Repository) CreateSession(ctx context.Context, s *Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

var ErrSessionNotFound = errors.New("sesi tidak ditemukan atau sudah berakhir")

// FindSessionByHash mencari sesi valid (belum dicabut & belum kadaluarsa) berdasarkan hash token.
func (r *Repository) FindSessionByHash(ctx context.Context, hash string) (*Session, error) {
	var s Session
	err := r.db.WithContext(ctx).Raw(`
		SELECT s.* FROM user_sessions s
		JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL AND u.is_active = true
		WHERE s.refresh_token_hash = ? AND s.revoked_at IS NULL AND s.expires_at > now()
	`, hash).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	if s.ID == 0 {
		return nil, ErrSessionNotFound
	}
	return &s, nil
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID int64) error {
	return r.db.WithContext(ctx).Exec(`UPDATE user_sessions SET revoked_at = now() WHERE id = ?`, sessionID).Error
}

func (r *Repository) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE user_sessions SET revoked_at = now() WHERE user_id = ? AND revoked_at IS NULL`, userID,
	).Error
}

func (r *Repository) FindPasswordHash(ctx context.Context, userID int64) (string, error) {
	var hash string
	err := r.db.WithContext(ctx).Raw(
		`SELECT password_hash FROM users WHERE id = ? AND deleted_at IS NULL`, userID,
	).Scan(&hash).Error
	if err != nil {
		return "", err
	}
	if hash == "" {
		return "", ErrUserNotFound
	}
	return hash, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	return r.db.WithContext(ctx).Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID).Error
}
