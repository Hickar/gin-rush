package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/Hickar/gin-rush/docs"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/rollbarinit"
	"github.com/Hickar/gin-rush/internal/routes"
	"github.com/Hickar/gin-rush/pkg/config"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/logging"
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
	settings := config.New("./conf/config.json")

	if err := rollbarinit.Setup(); err != nil {
		log.Fatalf("rollbar setup error: %s", err)
	}

	if err := logging.Setup("./logs/log.log", "%s_%s", "2006-01-02"); err != nil {
		log.Fatalf("logging setup error: %s", err)
	}

	db, err := database.New(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatalf("database setup: %s", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("models migration err: %s", err)
	}

	if err := mailer.Setup(); err != nil {
		log.Fatalf("mailer setup error: %s", err)
	}

	gin.SetMode(settings.Server.Mode)
	router := routes.Setup()

	if err := router.Run(fmt.Sprintf(":%d", settings.Server.Port)); err != nil {
		log.Fatalf("cannot start GIN server: %s", err)
	}
}
