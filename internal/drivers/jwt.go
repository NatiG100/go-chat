package drivers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secret string
}

func NewJWTManager(secret string) *JWTManager { return &JWTManager{secret: secret} }

func (j *JWTManager) Generate(userID uuid.UUID, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{"sub": userID.String(), "exp": time.Now().Add(exp).Unix()}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(j.secret))
}

func (j *JWTManager) Verify(tokenStr string) (uuid.UUID, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing")
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if !t.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	m, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims")
	}
	sub, ok := m["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid subject")
	}
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
