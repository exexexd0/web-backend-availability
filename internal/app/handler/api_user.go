package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/auth"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
)

type credentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginData struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type loginResponse struct {
	Status string    `json:"status"`
	Data   loginData `json:"data"`
}

type userResponse struct {
	Status string  `json:"status"`
	Data   ds.User `json:"data"`
}

type messageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type errorResponse struct {
	Status  string `json:"status" example:"fail"`
	Message string `json:"message"`
}

// RegisterAPI godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт пользователя. Поле is_moderator с клиента не принимается и всегда false. Пароль сохраняется как bcrypt-хеш и в ответ не входит.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        body body credentialsRequest true "Логин и пароль"
// @Success      201 {object} userResponse
// @Failure      400 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Router       /api/users [post]
func (h *Handler) RegisterAPI(ctx *gin.Context) {
	var body credentialsRequest
	if err := decodeStrictJSON(ctx, &body); err != nil {
		h.errorHandler(ctx, statusBadRequest, err)
		return
	}

	login := strings.TrimSpace(body.Login)
	password := body.Password
	if login == "" || password == "" {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("заполните логин и пароль"))
		return
	}
	if utf8.RuneCountInString(login) > 25 || utf8.RuneCountInString(password) > 100 {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("логин не длиннее 25 символов, пароль не длиннее 100 символов"))
		return
	}

	user, err := h.Repository.RegisterUser(login, password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginTaken) {
			h.errorHandler(ctx, statusBadRequest, err)
			return
		}
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось зарегистрировать пользователя"))
		return
	}

	ctx.JSON(statusCreated, gin.H{
		"status": "success",
		"data":   user,
	})
}

// LoginAPI godoc
// @Summary      Аутентификация
// @Description  Проверяет логин и пароль и возвращает JWT. Дальше токен передаётся в заголовке Authorization: Bearer {access_token}.
// @Tags         Пользователи
// @Accept       json
// @Produce      json
// @Param        body body credentialsRequest true "Логин и пароль"
// @Success      200 {object} loginResponse
// @Failure      400 {object} errorResponse
// @Failure      401 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Router       /api/users/login [post]
func (h *Handler) LoginAPI(ctx *gin.Context) {
	var body credentialsRequest
	if err := decodeStrictJSON(ctx, &body); err != nil {
		h.errorHandler(ctx, statusBadRequest, err)
		return
	}

	login := strings.TrimSpace(body.Login)
	password := body.Password
	if login == "" || password == "" {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("заполните логин и пароль"))
		return
	}

	user, err := h.Repository.GetUserByLogin(login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
			return
		}
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось выполнить вход"))
		return
	}
	if !auth.CheckPassword(user.Password, password) {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("неверный логин или пароль"))
		return
	}

	token, err := auth.Sign(h.JWT, *user)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, err)
		return
	}

	ctx.JSON(statusOK, loginResponse{
		Status: "success",
		Data: loginData{
			ExpiresIn:   token.ExpiresIn,
			AccessToken: token.AccessToken,
			TokenType:   token.TokenType,
		},
	})
}

// LogoutAPI godoc
// @Summary      Деавторизация
// @Description  Помещает текущий JWT в чёрный список Redis до истечения срока токена. Повторное использование этого токена даёт 403.
// @Tags         Пользователи
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} messageResponse
// @Failure      401 {object} errorResponse
// @Failure      403 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Router       /api/users/logout [post]
func (h *Handler) LogoutAPI(ctx *gin.Context) {
	jwtStr := strings.TrimPrefix(ctx.GetHeader("Authorization"), jwtPrefix)
	claims, err := auth.Parse(h.JWT, jwtStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
		return
	}

	if claims.ExpiresAt == nil {
		h.errorHandler(ctx, http.StatusUnauthorized, errUnauthorized)
		return
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, claims.Login, ttl); err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось завершить сессию"))
		return
	}

	ctx.JSON(statusOK, messageResponse{
		Status:  "success",
		Message: "выход выполнен",
	})
}
