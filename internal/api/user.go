package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Name     string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Email    string `json:"email" binding:"required,validemail" maxLength:"128"`
	Password string `json:"password" binding:"required,validpassword" minLength:"6" maxLength:"64"`
}

type AuthUserInput struct {
	Email    string `json:"email" binding:"required,validemail" maxLength:"128"`
	Password string `json:"password" binding:"required,validpassword" minLength:"6" maxLength:"64"`
}

type UpdateUserInput struct {
	Name      string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Bio       string `json:"bio" binding:"max=512" maxLength:"512"`
	Avatar    string `json:"avatar"`
	BirthDate string `json:"birth_date" binding:"validbirthdate"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// CreateUser godoc
// @Summary Create new user
// @Description Create new user with credentials provided in request. Response contains user JWT.
// @Accept json
// @Produces json
// @Param new_user body CreateUserInput true "JSON with user credentials"
// @Success 201 {object} AuthResponse{token=string}
// @Failure 409
// @Failure 422
// @Router /user [post]
func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	if exists, _ := models.UserExistsByEmail(input.Email); exists {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	salt, _ := security.RandomBytes(16)
	hashedPassword, err := security.HashPassword(input.Password, salt)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	user, err := models.CreateUser(map[string]interface{}{
		"name":     input.Name,
		"email":    input.Email,
		"password": hashedPassword,
		"salt":     salt,
	})
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	jwtStr, err := security.GenerateJWT(user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{"jwt": jwtStr},
	)
}

// AuthorizeUser godoc
// @Summary Authorize user with username/password
// @Description Method for authorizing user with credentials, returning signed jwt in response
// @Accept json
// @Produces json
// @Param login_user body AuthUserInput true "JSON with credentials"
// @Success 200 {object} AuthResponse{token=string}
// @Failure 404
// @Failure 422
// @Router /authorize [post]
func AuthorizeUser(c *gin.Context) {
	var input AuthUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	if exists, err := models.UserExistsByEmail(input.Email); !exists || err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	if valid := security.VerifyPassword(input.Password, user.Password, user.Salt); valid {
		jwtStr, err := security.GenerateJWT(user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusConflict)
			return
		}

		c.JSON(
			http.StatusOK,
			gin.H{"jwt": jwtStr},
		)
	}
}

// UpdateUser godoc
// @Summary Update user info
// @Description Method for updating user info: name, bio, avatar and birth date
// @Accept json
// @Produces json
// @Param update_user body UpdateUserInput true "JSON with user info"
// @Success 204
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 422
// @Security ApiKeyAuth
// @Router /user [patch]
func UpdateUser(c *gin.Context) {
	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	rawToken := security.TrimJWTPrefix(c.GetHeader("AUTHORIZATION"))
	token, _ := security.ParseJWT(rawToken)

	user, err := models.GetUserByID(token.UserID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if token.UserID != user.ID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	formattedTime, err := time.Parse("2006-01-02", input.BirthDate)
	if err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	if err = models.UpdateUser(*user, map[string]interface{}{
		"name":       input.Name,
		"bio":        input.Bio,
		"birth_date": formattedTime,
		"avatar":     input.Avatar,
	}); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	c.AbortWithStatus(http.StatusOK)
}

// GetUser godoc
// @Summary Get user
// @Description Get user by id
// @Accept json
// @Produces json
// @Param user_id path int true "User ID"
// @Success 200 {object} UpdateUserInput{name=string,bio=string,avatar=string,birth_date=string}
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 422
// @Router /user/{id} [get]
// @Security ApiKeyAuth
func GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	rawToken := security.TrimJWTPrefix(c.GetHeader("AUTHORIZATION"))
	token, _ := security.ParseJWT(rawToken)

	user, err := models.GetUserByID(uint(userID))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if uint(userID) != token.UserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	bio, _ := user.Bio.Value()
	avatar, _ := user.Avatar.Value()
	birthDate, _ := user.BirthDate.Value()
	c.JSON(http.StatusOK, gin.H{
		"name":       user.Name,
		"bio":        bio,
		"avatar":     avatar,
		"birth_date": birthDate,
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user by id
// @Accept json
// @Produces json
// @Param user_id path int true "User ID"
// @Success 204
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 422
// @Security ApiKeyAuth
// @Router /user/{id} [delete]
func DeleteUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	rawToken := c.GetHeader("AUTHORIZATION")
	token, _ := security.ParseJWT(security.TrimJWTPrefix(rawToken))

	user, err := models.GetUserByID(uint(userID))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if uint(userID) != token.UserID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if err := models.DeleteUser(*user); err != nil {
		c.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}