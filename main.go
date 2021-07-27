package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Hickar/gin-rush/internal/config"
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
	fmt.Printf(settings.Server.Mode)
	router := setupRouter()

	if err := router.Run(fmt.Sprintf(":%d", settings.Server.Port)); err != nil {
		log.Fatalf("Cannot start GIN server: %s", err)
	}
}
