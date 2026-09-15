package handler

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/sirupsen/logrus"
)

const (
	statusOK                  = 200
	statusBadRequest          = 400
	statusNotFound            = 404
	statusInternalServerError = 500
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

type componentCard struct {
	ID            int
	Name          string
	ImageURL      string
	UptimePercent float32
	LikesCount    int
}

func (h *Handler) GetComponents(ctx *gin.Context) {
	components, err := h.Repository.GetComponents()
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить компоненты")
		return
	}

	uptimeFilterStr := ctx.Query("uptime_percent")
	var uptimeFilter float64
	hasFilter := false
	if uptimeFilterStr != "" {
		parsed, parseErr := strconv.ParseFloat(uptimeFilterStr, 64)
		if parseErr != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < 0 || parsed > 100 {
			ctx.String(statusBadRequest, "Некорректное значение uptime_percent: укажите число от 0 до 100")
			return
		}
		uptimeFilter = parsed
		hasFilter = true
	}

	cards := make([]componentCard, 0, len(components))
	for _, component := range components {
		if component.Status != repository.StatusPublished {
			continue
		}
		if hasFilter && float64(component.UptimePercent) < uptimeFilter {
			continue
		}
		cards = append(cards, componentCard{
			ID:            component.ID,
			Name:          component.Name,
			ImageURL:      component.ImageURL,
			UptimePercent: component.UptimePercent,
			LikesCount:    len(component.Likes),
		})
	}

	ctx.HTML(statusOK, "component_grid.html", gin.H{
		"ActiveTab":    "grid",
		"Components":   cards,
		"UptimeFilter": uptimeFilterStr,
	})
}

func (h *Handler) GetComponent(ctx *gin.Context) {
	components, err := h.Repository.GetComponents()
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить компоненты")
		return
	}

	published := make([]repository.Component, 0)
	for _, c := range components {
		if c.Status == repository.StatusPublished {
			published = append(published, c)
		}
	}
	if len(published) == 0 {
		ctx.String(statusNotFound, "Опубликованные компоненты не найдены")
		return
	}

	idStr := ctx.Param("id")
	wantNext := ctx.Query("next") == "true"

	var current repository.Component
	if idStr == "" {
		current = published[0]
	} else {
		id, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			ctx.String(statusBadRequest, "Некорректный идентификатор компонента")
			return
		}
		idx := -1
		for i, c := range published {
			if c.ID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			ctx.String(statusNotFound, "Компонент не найден")
			return
		}
		if wantNext {
			idx = (idx + 1) % len(published)
		}
		current = published[idx]
	}

	ctx.HTML(statusOK, "component_feed.html", gin.H{
		"ActiveTab":  "feed",
		"Component":  current,
		"LikesCount": len(current.Likes),
	})
}

func (h *Handler) AddComponent(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftComponent()
	if err != nil {
		logrus.Error(err)
		ctx.String(statusNotFound, "Черновой компонент не найден")
		return
	}

	ctx.HTML(statusOK, "component_add.html", gin.H{
		"ActiveTab": "add",
		"Component": draft,
	})
}
