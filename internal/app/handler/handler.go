package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/one-compressive/web-backend-availability/internal/app/storage"
	"github.com/sirupsen/logrus"
)

const (
	statusOK                  = http.StatusOK
	statusCreated             = http.StatusCreated
	statusBadRequest          = http.StatusBadRequest
	statusNotFound            = http.StatusNotFound
	statusInternalServerError = http.StatusInternalServerError

	defaultImageURL = "/static/img/default.jpg"
	defaultVideoURL = "/static/img/default.mp4"

	descriptionPreviewLength = 30
)

type Handler struct {
	Repository *repository.Repository
	Storage    *storage.MediaStorage
}

func NewHandler(r *repository.Repository, s *storage.MediaStorage) *Handler {
	return &Handler{Repository: r, Storage: s}
}

func debugLog(hypothesisID, location, message string, data map[string]any) {
	// #region agent log
	f, err := os.OpenFile("/home/oaa/web-backend-availability/.cursor/debug-be3810.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	payload, err := json.Marshal(map[string]any{
		"sessionId":    "be3810",
		"runId":        "pre-fix",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	})
	if err != nil {
		return
	}
	_, _ = f.Write(append(payload, '\n'))
	// #endregion
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// #region agent log
	router.Use(func(ctx *gin.Context) {
		ctx.Next()
		debugLog("H2", "handler.go:request", "http request finished", map[string]any{
			"method": ctx.Request.Method,
			"path":   ctx.Request.URL.Path,
			"query":  ctx.Request.URL.RawQuery,
			"status": ctx.Writer.Status(),
		})
	})
	// #endregion
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusFound, "/components")
	})
	router.GET("/components", h.GetComponents)
	router.GET("/component/:id", h.GetComponent)
	router.GET("/add_component", h.GetAddComponent)
	router.POST("/components", h.CreateComponent)
	router.POST("/components/:id/publish", h.PublishComponent)
	router.POST("/components/:id/delete", h.DeleteComponent)

	api := router.Group("/api")
	api.GET("/components", h.GetComponentsAPI)
	api.GET("/components/draft", h.GetDraftAPI)
	api.GET("/components/feed", h.GetFeedAPI)
	api.GET("/components/feed/:id", h.GetFeedByIDAPI)
	api.POST("/components", h.AddComponentAPI)
	api.PUT("/components/:id", h.PublishComponentAPI)
	api.DELETE("/components/:id", h.DeleteComponentAPI)
	api.POST("/components/:id/like", h.LikeComponentAPI)
	api.POST("/users", h.RegisterAPI)
	api.POST("/users/login", h.LoginAPI)
	api.POST("/users/logout", h.LogoutAPI)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":  "fail",
		"message": err.Error(),
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.SetFuncMap(template.FuncMap{
		"percentValue": percentValue,
	})
	router.LoadHTMLGlob("./templates/*")
	router.Static("/static", "./static")
}

func (h *Handler) resolveMedia(imageURL, videoURL string) (string, string) {
	return h.resolveStoredMedia(imageURL, defaultImageURL), h.resolveStoredMedia(videoURL, defaultVideoURL)
}

func (h *Handler) resolveStoredMedia(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if strings.HasPrefix(value, "http") || strings.HasPrefix(value, "/static") {
		return value
	}
	return h.Storage.PublicObjectURL(value)
}

// sniffContentType распознаёт обычный mp4 по боксу ftyp.
// DetectContentType считает video/mp4 только часть файлов, у которых марка содержит буквы mp4.
func sniffContentType(head []byte) string {
	if len(head) >= 8 && string(head[4:8]) == "ftyp" {
		return "video/mp4"
	}
	return http.DetectContentType(head)
}

func truncateText(text string, limit int) (string, bool) {
	runes := []rune(text)
	if len(runes) <= limit {
		return text, false
	}
	return string(runes[:limit]), true
}

func percentValue(value *float32) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *value)
}

func (h *Handler) feedURL() string {
	id, err := h.Repository.GetFirstPublishedID()
	if err != nil || id == 0 {
		return "/components"
	}
	return "/component/" + strconv.FormatUint(uint64(id), 10)
}

type componentCard struct {
	ID            uint
	Name          string
	ImageURL      string
	UptimePercent *float32
	LikesCount    int64
}

func pageData(activeTab, feedURL string, extra gin.H) gin.H {
	data := gin.H{
		"ActiveTab": activeTab,
		"FeedURL":   feedURL,
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func (h *Handler) withResolvedMedia(component ds.Component) ds.Component {
	component.ImageURL, component.VideoURL = h.resolveMedia(component.ImageURL, component.VideoURL)
	return component
}

func parseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
