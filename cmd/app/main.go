package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/one-compressive/web-backend-availability/internal/app/config"
	"github.com/one-compressive/web-backend-availability/internal/app/dsn"
	"github.com/one-compressive/web-backend-availability/internal/app/handler"
	appredis "github.com/one-compressive/web-backend-availability/internal/app/redis"
	"github.com/one-compressive/web-backend-availability/internal/app/repository"
	"github.com/one-compressive/web-backend-availability/internal/app/storage"
	"github.com/one-compressive/web-backend-availability/internal/pkg"
	"github.com/sirupsen/logrus"

	_ "github.com/one-compressive/web-backend-availability/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Оценка отказоустойчивости системы
// @version         1.0
// @description     Веб-сервис компонентов отказоустойчивой системы. Чтение списка и ленты доступно без токена. Методы создателя требуют заголовок Authorization: Bearer {access_token}.

// @contact.name    API Support

// @license.name    AS IS (NO WARRANTY)

// @host            localhost:8081
// @BasePath        /
// @schemes         http

// @securityDefinitions.apikey BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите значение целиком: Bearer {access_token}

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

	redisClient, errRedis := appredis.New(context.Background(), conf.Redis)
	if errRedis != nil {
		logrus.Fatalf("error initializing redis: %v", errRedis)
	}
	defer redisClient.Close()

	media, errMedia := storage.NewFromEnv()
	if errMedia != nil {
		logrus.Fatalf("error initializing media storage: %v", errMedia)
	}

	router := gin.Default()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	hand := handler.NewHandler(rep, media, redisClient, conf.JWT)
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
