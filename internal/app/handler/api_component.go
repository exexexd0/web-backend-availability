package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
)

var errSystemFields = errors.New("системные поля задаются на сервере и не принимаются из запроса")

var systemFormFields = []string{
	"id",
	"status",
	"creator_id",
	"moderator_id",
	"created_at",
	"formed_at",
	"completed_at",
	"image_url",
	"video_url",
}

func (h *Handler) currentUser(ctx *gin.Context) (*ds.User, bool) {
	user, err := h.Repository.CurrentUser()
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, err)
		return nil, false
	}
	return user, true
}

func (h *Handler) GetComponentsAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	uptimeFilterStr := ctx.Query("uptime_percent")
	var minUptime *float64
	if uptimeFilterStr != "" {
		parsed, parseErr := parsePercent(uptimeFilterStr)
		if parseErr != nil {
			h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректное значение uptime_percent: укажите число от 0 до 100"))
			return
		}
		minUptime = &parsed
	}

	components, err := h.Repository.GetPublishedComponents(minUptime)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить услуги"))
		return
	}

	items := make([]ds.ComponentListSerializer, 0, len(components))
	for _, component := range components {
		likesCount, likesErr := h.Repository.CountLikes(component.ID)
		if likesErr != nil {
			h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить лайки"))
			return
		}
		items = append(items, newListItem(component, likesCount, user.ID))
	}

	ctx.JSON(statusOK, gin.H{
		"status": "success",
		"data":   items,
	})
}

func (h *Handler) GetFeedAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	component, err := h.Repository.GetFirstPublishedComponent()
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить ленту"))
		return
	}
	if component == nil {
		h.errorHandler(ctx, statusNotFound, fmt.Errorf("опубликованные услуги не найдены"))
		return
	}

	h.writeFeed(ctx, component, user.ID)
}

func (h *Handler) GetFeedByIDAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	id, err := parseID(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректный идентификатор услуги"))
		return
	}

	var component *ds.Component
	if ctx.Query("next") == "true" {
		next, nextErr := h.Repository.GetNextComponent(int(id))
		if nextErr != nil {
			h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить ленту"))
			return
		}
		if next == nil {
			h.errorHandler(ctx, statusNotFound, fmt.Errorf("опубликованные услуги не найдены"))
			return
		}
		component, err = h.Repository.GetPublishedWithCreator(next.ID)
		if err != nil {
			h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить ленту"))
			return
		}
	} else {
		component, err = h.Repository.GetPublishedWithCreator(id)
		if err != nil {
			h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить ленту"))
			return
		}
	}
	if component == nil {
		h.errorHandler(ctx, statusNotFound, fmt.Errorf("опубликованная услуга не найдена"))
		return
	}

	h.writeFeed(ctx, component, user.ID)
}

func (h *Handler) GetDraftAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	draft, err := h.Repository.GetOwnDraft(user.ID)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить черновик"))
		return
	}
	if draft == nil {
		h.errorHandler(ctx, statusNotFound, fmt.Errorf("черновик не найден"))
		return
	}

	h.writeComponentDetail(ctx, draft)
}

func (h *Handler) AddComponentAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCreateRequestSize)
	if err := ctx.Request.ParseMultipartForm(maxCreateRequestSize); err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректная форма"))
		return
	}
	if field := forbiddenFormField(ctx); field != "" {
		h.errorHandler(ctx, statusBadRequest, errSystemFields)
		return
	}

	name := strings.TrimSpace(ctx.PostForm("name"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	configType := ds.ConfigType(ctx.PostForm("config_type"))
	if name == "" || description == "" || configType == "" {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("заполните название, описание и тип конфигурации"))
		return
	}
	if utf8.RuneCountInString(name) > 100 || utf8.RuneCountInString(description) > 1000 {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("название не длиннее 100 символов, описание не длиннее 1000 символов"))
		return
	}
	if !configType.IsValid() {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректный тип конфигурации"))
		return
	}

	existing, err := h.Repository.GetOwnDraft(user.ID)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить черновик"))
		return
	}
	if existing != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("у пользователя уже есть черновик"))
		return
	}

	imageName, ok := h.storeAPIFile(ctx, "image", "images", 5<<20, map[string]string{"image/png": ".png", "image/jpeg": ".jpg"})
	if !ok {
		return
	}
	videoName, ok := h.storeAPIFile(ctx, "video", "videos", 50<<20, map[string]string{"video/mp4": ".mp4", "video/webm": ".webm"})
	if !ok {
		return
	}

	draft, err := h.Repository.CreateDraft(ds.Component{
		Name:        name,
		Description: description,
		ConfigType:  configType,
		ImageURL:    imageName,
		VideoURL:    videoName,
		CreatorID:   user.ID,
	})
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось создать черновик"))
		return
	}

	ctx.JSON(statusCreated, gin.H{
		"status": "success",
		"data":   newDetail(*draft, 0),
	})
}

func (h *Handler) PublishComponentAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	id, err := parseID(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректный идентификатор услуги"))
		return
	}

	var body struct {
		UptimePercent *float64 `json:"uptime_percent"`
		SystemImpact  *float64 `json:"system_impact"`
	}
	if err := decodeStrictJSON(ctx, &body); err != nil {
		h.errorHandler(ctx, statusBadRequest, err)
		return
	}
	if body.UptimePercent == nil || body.SystemImpact == nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("заполните расчётную доступность и влияние на систему"))
		return
	}
	if !validPercent(*body.UptimePercent) {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректное значение доступности"))
		return
	}
	if !validPercent(*body.SystemImpact) {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректное значение влияния на систему"))
		return
	}

	if err := h.Repository.PublishOwnComponent(id, user.ID, float32(*body.UptimePercent), float32(*body.SystemImpact)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, statusNotFound, err)
			return
		}
		if errors.Is(err, repository.ErrNotDraft) {
			h.errorHandler(ctx, statusBadRequest, err)
			return
		}
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось опубликовать услугу"))
		return
	}

	updated, err := h.Repository.GetComponentWithCreator(id)
	if err != nil || updated == nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить услугу"))
		return
	}
	h.writeComponentDetail(ctx, updated)
}

func (h *Handler) DeleteComponentAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	id, err := parseID(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректный идентификатор услуги"))
		return
	}

	if err := h.Repository.SoftDeleteOwnComponent(id, user.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, statusNotFound, err)
			return
		}
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось удалить услугу"))
		return
	}

	ctx.JSON(statusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"id": id,
		},
	})
}

func (h *Handler) LikeComponentAPI(ctx *gin.Context) {
	user, ok := h.currentUser(ctx)
	if !ok {
		return
	}

	id, err := parseID(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("некорректный идентификатор услуги"))
		return
	}

	var body struct {
		Value *int `json:"value"`
	}
	if err := decodeStrictJSON(ctx, &body); err != nil {
		h.errorHandler(ctx, statusBadRequest, err)
		return
	}
	if body.Value == nil || (*body.Value != 0 && *body.Value != 1) {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("поле value должно быть 0 или 1"))
		return
	}

	if err := h.Repository.SetComponentLike(user.ID, id, *body.Value); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, statusNotFound, err)
			return
		}
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось сохранить лайк"))
		return
	}

	ctx.JSON(statusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"component_id": id,
			"value":        *body.Value,
		},
	})
}

func (h *Handler) writeComponentDetail(ctx *gin.Context, component *ds.Component) {
	likesCount, err := h.Repository.CountLikes(component.ID)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить лайки"))
		return
	}
	ctx.JSON(statusOK, gin.H{
		"status": "success",
		"data":   newDetail(*component, likesCount),
	})
}

func (h *Handler) writeFeed(ctx *gin.Context, component *ds.Component, userID uint) {
	likesCount, err := h.Repository.CountLikes(component.ID)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить лайки"))
		return
	}
	liked, err := h.Repository.UserLiked(userID, component.ID)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить лайки"))
		return
	}
	isLiked := 0
	if liked {
		isLiked = 1
	}
	ctx.JSON(statusOK, gin.H{
		"status": "success",
		"data":   newFeed(*component, likesCount, isLiked),
	})
}

func (h *Handler) storeAPIFile(ctx *gin.Context, field, folder string, maxSize int64, types map[string]string) (string, bool) {
	file, err := ctx.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", true
	}
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("не удалось прочитать загруженный файл"))
		return "", false
	}
	if file.Size > maxSize {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("файл слишком большой"))
		return "", false
	}

	src, err := file.Open()
	if err != nil {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("не удалось прочитать загруженный файл"))
		return "", false
	}
	head := make([]byte, 512)
	n, _ := src.Read(head)
	src.Close()

	contentType := sniffContentType(head[:n])
	ext, allowed := types[contentType]
	if !allowed {
		h.errorHandler(ctx, statusBadRequest, fmt.Errorf("неподдерживаемый формат файла"))
		return "", false
	}

	objectName, err := h.Storage.UploadObjectName(ctx.Request.Context(), file, folder, contentType, ext)
	if err != nil {
		h.errorHandler(ctx, statusInternalServerError, fmt.Errorf("не удалось загрузить файл в хранилище"))
		return "", false
	}
	return objectName, true
}

func newListItem(component ds.Component, likesCount int64, currentUserID uint) ds.ComponentListSerializer {
	isOwner := 0
	if component.CreatorID == currentUserID {
		isOwner = 1
	}
	return ds.ComponentListSerializer{
		ID:            component.ID,
		Name:          component.Name,
		Description:   component.Description,
		ImageURL:      component.ImageURL,
		VideoURL:      component.VideoURL,
		ConfigType:    component.ConfigType,
		UptimePercent: component.UptimePercent,
		SystemImpact:  component.SystemImpact,
		LikesCount:    likesCount,
		IsOwner:       isOwner,
	}
}

func newDetail(component ds.Component, likesCount int64) ds.ComponentDetailSerializer {
	return ds.ComponentDetailSerializer{
		Component:  component,
		LikesCount: likesCount,
	}
}

func newFeed(component ds.Component, likesCount int64, isLiked int) ds.ComponentFeedSerializer {
	return ds.ComponentFeedSerializer{
		ComponentDetailSerializer: newDetail(component, likesCount),
		IsLiked:                   isLiked,
	}
}

func forbiddenFormField(ctx *gin.Context) string {
	for _, field := range systemFormFields {
		if ctx.Request.PostForm.Has(field) {
			return field
		}
	}
	return ""
}

func decodeStrictJSON(ctx *gin.Context, dst any) error {
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return errSystemFields
		}
		return fmt.Errorf("некорректное тело запроса")
	}
	return nil
}

func parsePercent(value string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || !validPercent(parsed) {
		return 0, fmt.Errorf("некорректный процент")
	}
	return parsed, nil
}

func validPercent(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 100
}
