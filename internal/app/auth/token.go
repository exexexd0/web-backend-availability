package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/one-compressive/web-backend-availability/internal/app/config"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/role"
)

type Token struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
	ExpiresAt   time.Time
}

func Sign(cfg config.JWTConfig, user ds.User) (Token, error) {
	expiresAt := time.Now().Add(cfg.ExpiresIn)
	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
		UserID:      user.ID,
		Login:       user.Login,
		Role:        role.FromModerator(user.IsModerator),
		IsModerator: user.IsModerator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.Token))
	if err != nil {
		return Token{}, fmt.Errorf("не удалось подписать токен")
	}

	return Token{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int(cfg.ExpiresIn.Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}

func Parse(cfg config.JWTConfig, tokenStr string) (*ds.JWTClaims, error) {
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("неподдерживаемый алгоритм подписи")
		}
		return []byte(cfg.Token), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("токен недействителен")
	}
	return claims, nil
}
