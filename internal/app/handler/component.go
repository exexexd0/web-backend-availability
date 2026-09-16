package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetComponents(ctx *gin.Context) {
	uptimeFilterStr := ctx.Query("uptime_percent")
	var minUptime *float64
	if uptimeFilterStr != "" {
		parsed, parseErr := strconv.ParseFloat(uptimeFilterStr, 64)
		if parseErr != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < 0 || parsed > 100 {
			ctx.String(statusBadRequest, "Некорректное значение uptime_percent: укажите число от 0 до 100")
			return
		}
		minUptime = &parsed
	}

	components, err := h.Repository.GetPublishedComponents(minUptime)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить компоненты")
		return
	}

	cards := make([]componentCard, 0, len(components))
	for _, component := range components {
		likesCount, likesErr := h.Repository.CountLikes(component.ID)
		if likesErr != nil {
			logrus.Error(likesErr)
			ctx.String(statusInternalServerError, "Не удалось загрузить лайки")
			return
		}
		imageURL, _ := resolveMedia(component.ImageURL, component.VideoURL)
		cards = append(cards, componentCard{
			ID:            component.ID,
			Name:          component.Name,
			ImageURL:      imageURL,
			UptimePercent: component.UptimePercent,
			LikesCount:    likesCount,
		})
	}

	ctx.HTML(statusOK, "component_grid.html", pageData("grid", h.feedURL(), gin.H{
		"Components":   cards,
		"UptimeFilter": uptimeFilterStr,
	}))
}

func (h *Handler) GetComponent(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.String(statusBadRequest, "Некорректный идентификатор компонента")
		return
	}

	component, err := h.Repository.GetComponentByID(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить компонент")
		return
	}
	if component == nil {
		ctx.String(statusNotFound, "Удалённую услугу просматривать нельзя")
		return
	}

	wantNext := ctx.Query("next") == "true"
	if wantNext {
		published, listErr := h.Repository.GetPublishedComponents(nil)
		if listErr != nil {
			logrus.Error(listErr)
			ctx.String(statusInternalServerError, "Не удалось загрузить компоненты")
			return
		}
		if len(published) == 0 {
			ctx.String(statusNotFound, "Опубликованные компоненты не найдены")
			return
		}
		idx := 0
		for i, item := range published {
			if item.ID == component.ID {
				idx = i
				break
			}
		}
		next := published[(idx+1)%len(published)]
		component = &next
	}

	likesCount, likesErr := h.Repository.CountLikes(component.ID)
	if likesErr != nil {
		logrus.Error(likesErr)
		ctx.String(statusInternalServerError, "Не удалось загрузить лайки")
		return
	}

	view := withResolvedMedia(*component)
	ctx.HTML(statusOK, "component_feed.html", pageData("feed", "/component/"+strconv.FormatUint(uint64(view.ID), 10), gin.H{
		"Component":  view,
		"LikesCount": likesCount,
	}))
}

func (h *Handler) GetAddComponent(ctx *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(repository.DefaultCreatorID)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить черновик")
		return
	}

	var view *ds.Component
	if draft != nil {
		resolved := withResolvedMedia(*draft)
		view = &resolved
	}

	ctx.HTML(statusOK, "component_add.html", pageData("add", h.feedURL(), gin.H{
		"HasDraft":  draft != nil,
		"Component": view,
	}))
}

func (h *Handler) CreateComponent(ctx *gin.Context) {
	name := ctx.PostForm("component_name")
	if name == "" {
		ctx.String(statusBadRequest, "Укажите название компонента")
		return
	}

	_, err := h.Repository.CreateDraft(name, repository.DefaultCreatorID)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось создать черновик")
		return
	}

	ctx.Redirect(http.StatusFound, "/add_component")
}

func (h *Handler) PublishComponent(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.String(statusBadRequest, "Некорректный идентификатор компонента")
		return
	}

	shortDescription := ctx.PostForm("short_description")
	description := ctx.PostForm("description")
	configType := ds.ConfigType(ctx.PostForm("config_type"))
	uptimeStr := ctx.PostForm("uptime_percent")
	impactStr := ctx.PostForm("system_impact_percent")
	if shortDescription == "" || description == "" || configType == "" || uptimeStr == "" || impactStr == "" {
		ctx.String(statusBadRequest, "Заполните краткую информацию, описание, тип конфигурации, доступность и влияние на систему")
		return
	}

	uptime, parseErr := strconv.ParseFloat(uptimeStr, 32)
	if parseErr != nil || math.IsNaN(uptime) || math.IsInf(uptime, 0) || uptime < 0 || uptime > 100 {
		ctx.String(statusBadRequest, "Некорректное значение доступности")
		return
	}
	impact, impactErr := strconv.ParseFloat(impactStr, 32)
	if impactErr != nil || math.IsNaN(impact) || math.IsInf(impact, 0) || impact < 0 || impact > 100 {
		ctx.String(statusBadRequest, "Некорректное значение влияния на систему")
		return
	}

	if err := h.Repository.PublishComponent(id, shortDescription, description, configType, float32(uptime), float32(impact)); err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось опубликовать компонент")
		return
	}

	ctx.Redirect(http.StatusFound, "/component/"+strconv.FormatUint(uint64(id), 10))
}

func (h *Handler) DeleteComponent(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.String(statusBadRequest, "Некорректный идентификатор компонента")
		return
	}

	if err := h.Repository.DeleteComponent(id); err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось удалить компонент")
		return
	}

	ctx.Redirect(http.StatusFound, "/components")
}
