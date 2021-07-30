package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/rollbarinit"
	"github.com/Hickar/gin-rush/pkg/logging"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	api := router.Group("/api")
	{
		api.GET("user/:id", nil)
		api.POST("user", nil)
		api.PATCH("user", nil)
		api.DELETE("user/:id", nil)

		api.POST("/authorize/email/challenge/:code", nil)
		api.POST("/authorize", nil)
	}

	return router
}

func main() {
	settings := config.New("./conf/config.json")

	if err := rollbarinit.Setup(); err != nil {
		log.Fatalf("Rollbar setup error: %s", err)
	}

	if err := logging.Setup("./logs/log.log", "%s_%s", "2006-01-02"); err != nil {
		log.Fatalf("Logging setup error: %s", err)
	}

	gin.SetMode(settings.Server.Mode)
	router := setupRouter()

	if err := router.Run(fmt.Sprintf(":%d", settings.Server.Port)); err != nil {
		log.Fatalf("Cannot start GIN server: %s", err)
	}
}
