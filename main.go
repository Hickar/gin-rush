package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/Hickar/gin-rush/docs"
	"github.com/Hickar/gin-rush/internal/cache"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/rollbar"
	"github.com/Hickar/gin-rush/internal/router"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/logger"
	"github.com/Hickar/gin-rush/pkg/mailer"
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
	conf := config.NewConfig("./conf/config.dev.json")

	if err := rollbar.NewRollbar(&conf.Rollbar); err != nil {
		log.Fatalf("rollbar setup error: %s", err)
	}

	_, err := logger.NewLogger("./logs/log.log", "%s_%s", "2006-01-02")
	if err != nil {
		log.Fatalf("logger setup error: %s", err)
	}

	db, err := database.NewDB(&conf.Database)
	if err != nil {
		log.Fatalf("database setup error: %s", err)
	}

	_, err = cache.NewCache(context.Background(), &conf.Redis)
	if err != nil {
		log.Fatalf("redis setup error: %s", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("models migration err: %s", err)
	}

	_, err = mailer.NewMailer("./conf/google_credentials.json")
	if err != nil {
		log.Fatalf("mailer setup error: %s", err)
	}

	gin.SetMode(conf.Server.Mode)
	r := router.NewRouter(&conf.Server)

	if err := r.Run(fmt.Sprintf(":%d", conf.Server.Port)); err != nil {
		log.Fatalf("cannot start GIN server: %s", err)
	}
}
