package api

import (
	"net/http"

	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Name     string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Email    string `json:"email" binding:"required,validemail" maxLength:"128"`
	Password string `json:"password" binding:"required,validpassword" minLength:"6" maxLength:"64"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// CreateUser godoc
// @Summary Create new user
// @Description Create new user with credentials provided in request. Response contains user JWT.
// @ID get-string-by-int
// @Accept json
// @Produces json
// @Param new_user body CreateUserInput true "JSON with user credentials"
// @Success 201 {object} AuthResponse{token=string}
// @Failure 409
// @Failure 422
// @Router /user/ [post]
func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	userExists, _ := models.UserExistsByEmail(input.Email)
	if userExists {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	user, err := models.CreateUser(map[string]interface{}{"name": input.Name, "email": input.Email})
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	_, err = security.HashPassword(input.Password)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	jwtStr, err := security.GenerateJWT(int(user.ID))
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{"jwt": jwtStr},
	)
}
