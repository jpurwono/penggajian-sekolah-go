package security

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("token tidak valid atau kadaluarsa")

// GenerateAccessToken membuat JWT HS256 berisi user ID sebagai subject.
func GenerateAccessToken(secret string, userID int64, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		Issuer:    "penggajian-sekolah",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// ParseAccessToken memverifikasi signature & masa berlaku, lalu mengembalikan user ID.
func ParseAccessToken(secret, tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return 0, ErrInvalidToken
	}
	userID, err := strconv.ParseInt(subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
