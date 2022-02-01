package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/Hickar/gin-rush/docs"
	"github.com/Hickar/gin-rush/internal/api"
	"github.com/Hickar/gin-rush/internal/broker"
	"github.com/Hickar/gin-rush/internal/cache"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/rollbar"
	"github.com/Hickar/gin-rush/internal/router"
	"github.com/Hickar/gin-rush/internal/usecase"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/logger"
	"github.com/gin-gonic/gin"
)

// @title Gin-Rush API
// @version 1.0
// @description Minimal API written on gin framework
// @termsOfService http://swagger.io/terms/

// @contact.name Hickar
// @contact.url https://hickar.space
// @contact.email hickar@icloud.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	if len(os.Args) < 1 {
		log.Fatalf("initialization error: no configuration file provided")
	}

	conf := config.NewConfig(os.Args[1])

	_, err := rollbar.NewRollbar(&conf.Rollbar);
	if err != nil {
		log.Fatalf("rollbar setup error: %s", err)
	}

	logger, err := logger.NewLogger("./logs/log.log", "%s_%s", "2006-01-02")
	if err != nil {
		log.Fatalf("logger setup error: %s", err)
	}

	db, err := database.NewDB(&conf.Database)
	if err != nil {
		log.Fatalf("database setup error: %s", err)
	}

	redis, err := cache.NewCache(context.Background(), &conf.Redis)
	if err != nil {
		log.Fatalf("redis setup error: %s", err)
	}

	br, err := broker.NewBroker(&conf.RabbitMQ)
	if err != nil {
		log.Fatalf("rabbitmq setup error: %s", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("models migration err: %s", err)
	}

	userUseCase, err := usecase.NewUserUseCase(db, conf, br, redis)
	if err != nil {
		log.Fatalf("cannot initialize UserUseCase type: %s", err)
	}

	userController := api.NewUserController(userUseCase)

	gin.SetMode(conf.Server.Mode)
	r := router.NewUserRouter(userController , conf)

	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		log.Fatalf("cannot start GIN server: %s", err)
	}
}
