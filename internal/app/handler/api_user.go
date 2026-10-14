package handler

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
)

func (h *Handler) RegisterAPI(ctx *gin.Context) {
	var body struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
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

func (h *Handler) LoginAPI(ctx *gin.Context) {
	ctx.JSON(statusOK, gin.H{
		"status":  "success",
		"message": "заглушка аутентификации",
	})
}

func (h *Handler) LogoutAPI(ctx *gin.Context) {
	ctx.JSON(statusOK, gin.H{
		"status":  "success",
		"message": "заглушка деавторизации",
	})
}
