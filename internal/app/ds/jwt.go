package ds

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/one-compressive/web-backend-availability/internal/app/role"
)

// JWTClaims — полезная нагрузка токена: стандартные поля RFC 7519 и данные пользователя.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint      `json:"user_id"`
	Login       string    `json:"login"`
	Role        role.Role `json:"role"`
	IsModerator bool      `json:"is_moderator"`
}
