package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/one-compressive/web-backend-availability/internal/app/config"
	"github.com/one-compressive/web-backend-availability/internal/app/handler"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	// #region agent log
	if f, err := os.OpenFile("/home/oaa/web-backend-availability/.cursor/debug-be3810.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		payload, _ := json.Marshal(map[string]any{
			"sessionId": "be3810", "runId": "pre-fix", "hypothesisId": "H1",
			"location": "app.go:RunApp", "message": "about to listen",
			"data":      map[string]any{"address": serverAddress},
			"timestamp": time.Now().UnixMilli(),
		})
		_, _ = f.Write(append(payload, '\n'))
		f.Close()
	}
	// #endregion
	if err := a.Router.Run(serverAddress); err != nil {
		// #region agent log
		if f, logErr := os.OpenFile("/home/oaa/web-backend-availability/.cursor/debug-be3810.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); logErr == nil {
			payload, _ := json.Marshal(map[string]any{
				"sessionId": "be3810", "runId": "pre-fix", "hypothesisId": "H5",
				"location": "app.go:RunApp", "message": "listen failed",
				"data":      map[string]any{"address": serverAddress, "error": err.Error()},
				"timestamp": time.Now().UnixMilli(),
			})
			_, _ = f.Write(append(payload, '\n'))
			f.Close()
		}
		// #endregion
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}
