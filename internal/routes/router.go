package routes

import (
	"net/http"

	"github.com/Hickar/gin-rush/internal/api"
	"github.com/Hickar/gin-rush/internal/middlewares"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", security.NotBlank)
		v.RegisterValidation("validemail", security.ValidEmail)
		v.RegisterValidation("validpassword", security.ValidPassword)
		v.RegisterValidation("validbirthdate", security.ValidBirthDate)
	}

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	user := router.Group("/api")
	{
		user.POST("user", api.CreateUser)
		user.DELETE("user/:id", nil)

		user.POST("/authorize/email/challenge/:code", nil)
		user.POST("/authorize", api.AuthorizeUser)
	}

	authUser := router.Group("/api")
	{
		user.GET("user/:id", api.GetUser)
		user.PATCH("user", api.UpdateUser)
	}
	authUser.Use(middlewares.JWT())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
