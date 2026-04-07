package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type InviteClaims struct {
	Type          string `json:"typ"`
	CohortID      string `json:"cohort_id"`
	InviteVersion int    `json:"invite_version"`

	jwt.RegisteredClaims
}

type InviteTokenManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewInviteTokenManager(secret string, ttl time.Duration, issuer string) *InviteTokenManager {
	return &InviteTokenManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}
}

func (m *InviteTokenManager) GenerateInviteToken(cohortID string) (string, error) {
	claims := &InviteClaims{
		Type:          "invite",
		CohortID:      cohortID,
		InviteVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign invite token %w", err)
	}
	return signed, nil
}

func (m *InviteTokenManager) ParseInviteToken(tokenStr string) (*InviteClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &InviteClaims{}, func(token *jwt.Token) (any, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse invite token %w", err)
	}

	claims, ok := token.Claims.(*InviteClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid invite token")
	}

	return claims, nil
}
