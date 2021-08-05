package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/Hickar/gin-rush/docs"
	"github.com/Hickar/gin-rush/internal/api"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/rollbarinit"
	"github.com/Hickar/gin-rush/pkg/logging"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	user := router.Group("/api")
	{
		user.GET("user/:id", api.GetUser)
		user.POST("user", nil)
		user.PATCH("user", nil)
		user.DELETE("user/:id", nil)

		user.POST("/authorize/email/challenge/:code", nil)
		user.POST("/authorize", nil)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func main() {
	settings := config.New("./conf/config.json")

	if err := rollbarinit.Setup(); err != nil {
		log.Fatalf("Rollbar setup error: %s", err)
	}

	if err := logging.Setup("./logs/log.log", "%s_%s", "2006-01-02"); err != nil {
		log.Fatalf("logging setup error: %s", err)
	}

	if err := models.Setup(); err != nil {
		log.Fatalf("db setup error: %s", err)
	}

	gin.SetMode(settings.Server.Mode)
	router := setupRouter()

	if err := router.Run(fmt.Sprintf(":%d", settings.Server.Port)); err != nil {
		log.Fatalf("Cannot start GIN server: %s", err)
	}
}