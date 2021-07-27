package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		api.GET("user/:id", nil)
		api.POST("user", nil)
		api.PATCH("user", nil)
		api.DELETE("user/:id", nil)

		api.POST("/authorize/email/challenge/:code", nil)
		api.POST("/authorize", nil)
	}

	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Cannot start GIN server: %s", err)
	}
}
