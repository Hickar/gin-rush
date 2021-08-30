package router

import (
	"net/http"

	"github.com/Hickar/gin-rush/internal/api"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/middlewares"
	"github.com/Hickar/gin-rush/internal/validators"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(conf *config.ServerConfig) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", validators.NotBlank)
		v.RegisterValidation("validemail", validators.ValidEmail)
		v.RegisterValidation("validpassword", validators.ValidPassword)
		v.RegisterValidation("validbirthdate", validators.ValidBirthDate)
	}

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	user := router.Group(conf.ApiUrl)
	{
		user.POST("user", api.CreateUser)
		user.POST("/authorize", api.AuthorizeUser)
		user.GET("/authorize/email/challenge/:code", api.EnableUser)
	}

	authUser := router.Group(conf.ApiUrl, middlewares.JWT())
	{
		authUser.GET("user/:id", api.GetUser)
		authUser.PATCH("user", api.UpdateUser)
		authUser.DELETE("user/:id", api.DeleteUser)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
