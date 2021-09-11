package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Hickar/gin-rush/internal/broker"
	"github.com/Hickar/gin-rush/internal/cache"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/mailer"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/logger"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/response"
	"github.com/Hickar/gin-rush/pkg/utils"
	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @Summary Create new user
// @Description Create new user with credentials provided in request. Response contains user JWT.
// @Accept json
// @Produces json
// @Param new_user body request.CreateUserRequest true "JSON with user credentials"
// @Success 201 {object} response.AuthUserResponse{token=string}
// @Failure 409
// @Failure 422
// @Router /user [post]
func CreateUser(c *gin.Context) {
	var input request.CreateUserRequest
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Printf("err: %s", err.Error())
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.DB()

	exists, _ := db.Exists(&models.User{}, "email", input.Email)
	if exists {
		c.Status(http.StatusConflict)
		return
	}

	salt, _ := security.RandomBytes(16)
	hashedPassword, err := security.HashPassword(input.Password, salt)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	user.Name = input.Name
	user.Email = input.Email
	user.Password = hashedPassword
	user.ConfirmationCode = utils.RandomString(30)
	user.Salt = salt

	err = db.Create(&user)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusConflict)
		return
	}

	conf := config.GetConfig()
	token, err := security.GenerateJWT(user.ID, conf.Server.JWTSecret)
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	msg, _ := json.Marshal(&mailer.ConfirmationMessage{
		Username: user.Name,
		Email:    user.Email,
		Code:     user.ConfirmationCode,
	})

	err = broker.GetBroker().Publish("mailer_ex", "mailer", "text/plain", &msg)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(
		http.StatusCreated,
		response.AuthUserResponse{Token: token},
	)
}

// AuthorizeUser godoc
// @Summary Authorize user with username/password
// @Description Method for authorizing user with credentials, returning signed jwt in response
// @Accept json
// @Produces json
// @Param login_user body request.AuthUserRequest true "JSON with credentials"
// @Success 200 {object} response.AuthUserResponse{token=string}
// @Failure 404
// @Failure 422
// @Router /authorize [post]
func AuthorizeUser(c *gin.Context) {
	var input request.AuthUserRequest
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.DB()

	err := db.FindBy(&user, "email", input.Email)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	valid := security.VerifyPassword(input.Password, user.Password, user.Salt);
	if !valid {
		c.Status(http.StatusConflict)
		return
	}

	conf := config.GetConfig()
	token, err := security.GenerateJWT(user.ID, conf.Server.JWTSecret)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusConflict)
		return
	}

	c.JSON(
		http.StatusOK,
		response.AuthUserResponse{Token: token},
	)
}

// UpdateUser godoc
// @Summary Update user info
// @Description Method for updating user info: name, bio, avatar and birth date
// @Accept json
// @Produces json
// @Param update_user body request.UpdateUserRequest true "JSON with user info"
// @Success 204
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 422
// @Security ApiKeyAuth
// @Router /user [patch]
func UpdateUser(c *gin.Context) {
	var input request.UpdateUserRequest
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.DB()
	authUserID := c.GetUint("user_id")
	
	if err := db.FindByID(&user, authUserID); err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	if authUserID != user.ID {
		c.Status(http.StatusForbidden)
		return
	}

	formattedTime, err := time.Parse("2006-01-02", input.BirthDate)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	user.Name = input.Name
	user.Bio = sql.NullString{input.Bio, true}
	user.BirthDate = sql.NullTime{formattedTime, true}
	user.Avatar = sql.NullString{input.Avatar, true}

	if err = db.Update(&user); err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	cacheKey := fmt.Sprintf("users:%d", authUserID)
	err = cache.GetCache().Del(c, cacheKey).Err()
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUser godoc
// @Summary Get user
// @Description Get user by id
// @Accept json
// @Produces json
// @Param user_id path int true "User ID"
// @Success 200 {object} response.UpdateUserResponse{name=string,bio=string,avatar=string,birth_date=string}
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 422
// @Router /user/{id} [get]
// @Security ApiKeyAuth
func GetUser(c *gin.Context) {
	var user models.User
	var resp response.UpdateUserResponse

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	authUserID := c.GetUint("user_id")
	if uint(userID) != authUserID {
		c.Status(http.StatusForbidden)
		return
	}

	cacheKey := fmt.Sprintf("users:%d", authUserID)
	cached, _ := cache.GetCache().Get(context.Background(), cacheKey).Result()
	if cached != "" {
		err := json.Unmarshal([]byte(cached), &resp)
		if err != nil {
			logger.GetLogger().Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	} else {
		db := database.DB()
		err = db.FindByID(&user, uint(userID))
		if err != nil {
			logger.GetLogger().Error(err)
			c.Status(http.StatusNotFound)
			return
		}

		resp.Name = user.Name
		resp.Bio = user.Bio.String
		resp.Avatar = user.Avatar.String
		resp.BirthDate = user.BirthDate.Time.String()

		cacheData, err := json.Marshal(&resp)
		if err != nil {
			logger.GetLogger().Error(err)
		}

		_, err = cache.GetCache().Set(context.Background(), cacheKey, cacheData, time.Hour*168).Result()
		if err != nil {
			logger.GetLogger().Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, resp)
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
	var user models.User

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	authUserID := c.GetUint("user_id")
	if uint(userID) != authUserID {
		c.Status(http.StatusForbidden)
		return
	}

	db := database.DB()
	if err = db.FindByID(&user, uint(userID)); err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	if err := db.Delete(&user); err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	cacheKey := fmt.Sprintf("users:%d", authUserID)
	err = cache.GetCache().Del(context.Background(), cacheKey).Err()
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

// EnableUser godoc
// @Summary Enable user
// @Description Method for enabling user via verification message sent by email
// @Produces json
// @Param confirmation_code path string true "Confirmation code"
// @Success 200 {object} response.AuthUserResponse{token=string}
// @Failure 404
// @Failure 422
// @Router /authorize/email/challenge/{code} [get]
func EnableUser(c *gin.Context) {
	var user models.User
	code := c.Param("code")

	if len(code) != 30 {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	db := database.DB()

	if err := db.FindBy(&user, "confirmation_code", code); err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusNotFound)
		return
	}

	if !user.Enabled {
		user.Enabled = true
		err := db.Update(&user)
		fmt.Println(err)
	}

	conf := config.GetConfig()
	token, err := security.GenerateJWT(user.ID, conf.Server.JWTSecret)
	if err != nil {
		logger.GetLogger().Error(err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	c.JSON(
		http.StatusOK,
		response.AuthUserResponse{Token: token},
	)
}