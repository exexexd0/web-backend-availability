package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/sirupsen/logrus"
)

const maxCreateRequestSize = 60 << 20

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
		ctx.Redirect(http.StatusFound, "/components")
		return
	}

	if ctx.Query("next") == "true" {
		next, nextErr := h.Repository.GetNextComponent(int(component.ID))
		if nextErr != nil {
			logrus.Error(nextErr)
			ctx.String(statusInternalServerError, "Не удалось загрузить следующий компонент")
			return
		}
		if next == nil {
			ctx.String(statusNotFound, "Опубликованные компоненты не найдены")
			return
		}
		component = next
	}

	likesCount, likesErr := h.Repository.CountLikes(component.ID)
	if likesErr != nil {
		logrus.Error(likesErr)
		ctx.String(statusInternalServerError, "Не удалось загрузить лайки")
		return
	}

	view := withResolvedMedia(*component)
	preview, truncated := truncateText(view.Description, descriptionPreviewLength)
	ctx.HTML(statusOK, "component_feed.html", pageData("feed", "/component/"+strconv.FormatUint(uint64(view.ID), 10), gin.H{
		"Component":            view,
		"LikesCount":           likesCount,
		"DescriptionPreview":   preview,
		"DescriptionTruncated": truncated,
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
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCreateRequestSize)

	name := strings.TrimSpace(ctx.PostForm("component_name"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	configType := ds.ConfigType(ctx.PostForm("config_type"))
	if name == "" || description == "" || configType == "" {
		ctx.String(statusBadRequest, "Заполните название, описание и тип конфигурации")
		return
	}
	if utf8.RuneCountInString(name) > 100 || utf8.RuneCountInString(description) > 1000 {
		ctx.String(statusBadRequest, "Название не длиннее 100 символов, описание не длиннее 1000 символов")
		return
	}
	if !configType.IsValid() {
		ctx.String(statusBadRequest, "Некорректный тип конфигурации")
		return
	}

	draft, err := h.Repository.GetDraftByCreator(repository.DefaultCreatorID)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить черновик")
		return
	}
	if draft != nil {
		ctx.Redirect(http.StatusFound, "/add_component")
		return
	}

	imageURL, ok := h.uploadFormFile(ctx, "component_image", "images", 5<<20, map[string]string{"image/png": ".png", "image/jpeg": ".jpg"})
	if !ok {
		return
	}
	videoURL, ok := h.uploadFormFile(ctx, "evaluation_video", "videos", 50<<20, map[string]string{"video/mp4": ".mp4", "video/webm": ".webm"})
	if !ok {
		return
	}

	_, err = h.Repository.CreateDraft(ds.Component{
		Name:        name,
		Description: description,
		ConfigType:  configType,
		ImageURL:    imageURL,
		VideoURL:    videoURL,
		CreatorID:   repository.DefaultCreatorID,
	})
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось создать черновик")
		return
	}

	ctx.Redirect(http.StatusFound, "/add_component")
}

func (h *Handler) uploadFormFile(ctx *gin.Context, field, folder string, maxSize int64, types map[string]string) (string, bool) {
	file, err := ctx.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", true
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(statusBadRequest, "Не удалось прочитать загруженный файл")
		return "", false
	}
	if file.Size > maxSize {
		ctx.String(statusBadRequest, "Файл "+file.Filename+" слишком большой")
		return "", false
	}

	src, err := file.Open()
	if err != nil {
		logrus.Error(err)
		ctx.String(statusBadRequest, "Не удалось прочитать загруженный файл")
		return "", false
	}
	head := make([]byte, 512)
	n, _ := src.Read(head)
	src.Close()

	contentType := http.DetectContentType(head[:n])
	ext, allowed := types[contentType]
	if !allowed {
		ctx.String(statusBadRequest, "Неподдерживаемый формат файла "+file.Filename)
		return "", false
	}

	url, err := h.Storage.Upload(ctx.Request.Context(), file, folder, contentType, ext)
	if err != nil {
		logrus.Error(err)
		ctx.String(statusInternalServerError, "Не удалось загрузить файл в хранилище")
		return "", false
	}
	return url, true
}

func (h *Handler) PublishComponent(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.String(statusBadRequest, "Некорректный идентификатор компонента")
		return
	}

	uptimeStr := ctx.PostForm("uptime_percent")
	impactStr := ctx.PostForm("system_impact_percent")
	if uptimeStr == "" || impactStr == "" {
		ctx.String(statusBadRequest, "Заполните расчётную доступность и влияние на систему")
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

	if err := h.Repository.PublishComponent(id, float32(uptime), float32(impact)); err != nil {
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
