package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Hickar/gin-rush/internal/repositories"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/mailer"
	"github.com/Hickar/gin-rush/pkg/utils"
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
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)

	if user, _ := repo.FindBy("email", input.Email); user != nil {
		c.Status(http.StatusConflict)
		return
	}

	salt, _ := security.RandomBytes(16)
	hashedPassword, err := security.HashPassword(input.Password, salt)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	user, err := repo.Create(&repositories.User{
		Name:             input.Name,
		Email:            input.Email,
		Password:         hashedPassword,
		ConfirmationCode: utils.RandomString(30),
		Salt:             salt,
	})
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	jwtStr, err := security.GenerateJWT(user.ID)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	if err := mailer.SendConfirmationCode(user.Name, user.Email, user.ConfirmationCode); err != nil {
		c.Status(http.StatusUnprocessableEntity)
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
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)

	user, err := repo.FindBy("email", input.Email)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if valid := security.VerifyPassword(input.Password, user.Password, user.Salt); valid {
		jwtStr, err := security.GenerateJWT(user.ID)
		if err != nil {
			c.Status(http.StatusConflict)
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
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)
	authUserID, _ := c.MustGet("user_id").(uint)

	user, err := repo.FindByID(authUserID)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if authUserID != user.ID {
		c.Status(http.StatusForbidden)
		return
	}

	formattedTime, err := time.Parse("2006-01-02", input.BirthDate)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	user.Name = input.Name
	user.Bio = sql.NullString{input.Bio, true}
	user.BirthDate = sql.NullTime{formattedTime, true}
	user.Avatar = sql.NullString{input.Avatar, true}

	if err = repo.Update(*user); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	c.Status(http.StatusOK)
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
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)
	authUserID, _ := c.MustGet("user_id").(uint)

	user, err := repo.FindByID(uint(userID))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if uint(userID) != authUserID {
		c.Status(http.StatusForbidden)
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
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)
	authUserID, _ := c.MustGet("user_id").(uint)

	user, err := repo.FindByID(uint(userID))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if uint(userID) != authUserID {
		c.Status(http.StatusForbidden)
		return
	}

	if err := repo.Delete(*user); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	c.Status(http.StatusNoContent)
}

// EnableUser godoc
// @Summary Enable user
// @Description Method for enabling user via verification message sent by email
// @Produces json
// @Param confirmation_code path string true "Confirmation code"
// @Success 200 {object} AuthResponse{token=string}
// @Failure 404
// @Failure 422
// @Router /authorize/email/challenge/{code} [get]
func EnableUser(c *gin.Context) {
	code := c.Param("code")

	if len(code) != 30 {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.GetDB()
	repo, _ := repositories.NewUserRepository(db)

	user, err := repo.FindBy("confirmation_code", code)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	if !user.Enabled {
		user.Enabled = true
		repo.Update(*user)
	}

	token, err := security.GenerateJWT(user.ID)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
