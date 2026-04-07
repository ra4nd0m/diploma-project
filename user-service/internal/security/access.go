package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	Sub   string `json:"sub"`
	Role  string `json:"role"`
	Name  string `json:"name"`
	Email string `json:"email"`

	jwt.RegisteredClaims
}

type AccessTokenManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewAccessTokenManager(secret string, ttl time.Duration, issuer string) *AccessTokenManager {
	return &AccessTokenManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}
}

func (m *AccessTokenManager) ParseAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *jwt.Token) (any, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token %w", err)
	}
	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid access token")
}
