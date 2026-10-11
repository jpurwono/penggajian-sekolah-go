package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"penggajian/server/internal/platform/config"
	"penggajian/server/internal/platform/middleware"
	"penggajian/server/internal/platform/security"
)

var (
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrWrongPassword      = errors.New("password saat ini salah")
)

type Service struct {
	cfg  *config.Config
	repo *Repository
}

func NewService(cfg *config.Config, repo *Repository) *Service {
	return &Service{cfg: cfg, repo: repo}
}

func newRefreshToken() (plain, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(sum[:]), nil
}

func hashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

func (s *Service) issueTokens(ctx context.Context, user *middleware.AuthUser, ua, ip string) (*TokenResponse, error) {
	access, _, err := security.GenerateAccessToken(s.cfg.JWTSecret, user.ID, s.cfg.JWTAccessTTL)
	if err != nil {
		return nil, err
	}

	refreshPlain, refreshHash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        ua,
		IP:               ip,
		ExpiresAt:        time.Now().Add(s.cfg.JWTRefreshTTL),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refreshPlain,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.JWTAccessTTL.Seconds()),
		User:         toUserResponse(user),
	}, nil
}

func (s *Service) Login(ctx context.Context, identifier, password, userAgent, ip string) (*TokenResponse, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, hash, err := s.repo.FindActiveByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := security.VerifyPassword(password, hash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	result, err := s.issueTokens(ctx, user, userAgent, ip)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}
	return result, nil
}

// Refresh memutar sesi: token lama dicabut, diterbitkan pasangan token baru.
func (s *Service) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (*TokenResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrSessionNotFound
	}

	old, err := s.repo.FindSessionByHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return nil, err
	}

	user, err := s.repo.LoadActiveUserByID(ctx, old.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RevokeSession(ctx, old.ID); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, userAgent, ip)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil
	}
	old, err := s.repo.FindSessionByHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return s.repo.RevokeSession(ctx, old.ID)
}

func (s *Service) Me(ctx context.Context, userID int64) (*UserResponse, error) {
	user, err := s.repo.LoadActiveUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

// ChangePassword mengubah password sendiri dan mencabut semua sesi (wajib login ulang).
func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}

	hash, err := s.repo.FindPasswordHash(ctx, userID)
	if err != nil {
		return err
	}

	ok, err := security.VerifyPassword(currentPassword, hash)
	if err != nil || !ok {
		return ErrWrongPassword
	}

	newHash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}
	return s.repo.RevokeAllUserSessions(ctx, userID)
}
