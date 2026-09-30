package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/one-compressive/web-backend-availability/internal/app/config"
	"github.com/one-compressive/web-backend-availability/internal/app/dsn"
	"github.com/one-compressive/web-backend-availability/internal/app/handler"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/one-compressive/web-backend-availability/internal/app/storage"
	"github.com/one-compressive/web-backend-availability/internal/pkg"
	"github.com/sirupsen/logrus"
)

func main() {
	_ = godotenv.Load()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	rep, errRep := repository.New(dsn.FromEnv())
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	media, errMedia := storage.NewFromEnv()
	if errMedia != nil {
		logrus.Fatalf("error initializing media storage: %v", errMedia)
	}

	router := gin.Default()
	hand := handler.NewHandler(rep, media)
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
