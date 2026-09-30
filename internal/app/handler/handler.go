package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/ds"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/one-compressive/web-backend-availability/internal/app/storage"
)

const (
	statusOK                  = http.StatusOK
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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/components", h.GetComponents)
	router.GET("/component/:id", h.GetComponent)
	router.GET("/add_component", h.GetAddComponent)
	router.POST("/components", h.CreateComponent)
	router.POST("/components/:id/publish", h.PublishComponent)
	router.POST("/components/:id/delete", h.DeleteComponent)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.SetFuncMap(template.FuncMap{
		"percentValue": percentValue,
	})
	router.LoadHTMLGlob("./templates/*")
	router.Static("/static", "./static")
}

func resolveMedia(imageURL, videoURL string) (string, string) {
	if strings.TrimSpace(imageURL) == "" {
		imageURL = defaultImageURL
	}
	if strings.TrimSpace(videoURL) == "" {
		videoURL = defaultVideoURL
	}
	return imageURL, videoURL
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

func withResolvedMedia(component ds.Component) ds.Component {
	component.ImageURL, component.VideoURL = resolveMedia(component.ImageURL, component.VideoURL)
	return component
}

func parseID(idStr string) (uint, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
