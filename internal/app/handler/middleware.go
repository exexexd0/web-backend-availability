package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/auth"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/one-compressive/web-backend-availability/internal/app/role"
	goredis "github.com/redis/go-redis/v9"
)

const (
	jwtPrefix      = "Bearer "
	userContextKey = "authUser"
)

// RequireAuth пропускает запрос только с действующим JWT нужной роли.
// Токен из чёрного списка Redis не принимается.
// Если роли не переданы, достаточно любой аутентифицированной роли.
func (h *Handler) RequireAuth(assignedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, ok := h.authenticate(ctx, true)
		if !ok {
			return
		}
		if len(assignedRoles) > 0 && !roleAllowed(claims.Role, assignedRoles) {
			h.errorHandler(ctx, http.StatusForbidden, errForbidden)
			ctx.Abort()
			return
		}
		user, err := h.Repository.GetUserByID(claims.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
				ctx.Abort()
				return
			}
			h.errorHandler(ctx, statusInternalServerError, err)
			ctx.Abort()
			return
		}
		ctx.Set(userContextKey, user)
		ctx.Next()
	}
}

// OptionalAuth читает JWT, если заголовок передан.
// Без заголовка запрос продолжается как гостевой: чтение списка и ленты от него не зависит.
func (h *Handler) OptionalAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if strings.TrimSpace(header) == "" {
			ctx.Next()
			return
		}
		claims, ok := h.authenticate(ctx, false)
		if !ok {
			return
		}
		user, err := h.Repository.GetUserByID(claims.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
				ctx.Abort()
				return
			}
			h.errorHandler(ctx, statusInternalServerError, err)
			ctx.Abort()
			return
		}
		ctx.Set(userContextKey, user)
		ctx.Next()
	}
}

func (h *Handler) authenticate(ctx *gin.Context, missingIsError bool) (*authClaims, bool) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		if missingIsError || strings.TrimSpace(jwtStr) != "" {
			h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
			ctx.Abort()
			return nil, false
		}
		return nil, false
	}

	jwtStr = jwtStr[len(jwtPrefix):]
	err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
	if err == nil {
		h.errorHandler(ctx, http.StatusForbidden, errTokenRevoked)
		ctx.Abort()
		return nil, false
	}
	if !errors.Is(err, goredis.Nil) {
		h.errorHandler(ctx, statusInternalServerError, err)
		ctx.Abort()
		return nil, false
	}

	claims, err := auth.Parse(h.JWT, jwtStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
		ctx.Abort()
		return nil, false
	}
	return &authClaims{UserID: claims.UserID, Role: claims.Role}, true
}

type authClaims struct {
	UserID uint
	Role   role.Role
}

func roleAllowed(actual role.Role, assigned []role.Role) bool {
	for _, candidate := range assigned {
		if actual == candidate {
			return true
		}
	}
	return false
}

var (
	errUnauthorized = errors.New("требуется авторизация")
	errForbidden    = errors.New("недостаточно прав")
	errTokenRevoked = errors.New("токен отозван")
)
