package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Hickar/gin-rush/internal/usecase"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserUseCase *usecase.UserUseCase
}

func NewUserController(useCase *usecase.UserUseCase) *UserController {
	return  &UserController{UserUseCase: useCase}
}

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
func (uc *UserController) CreateUser(c *gin.Context) {
	var input request.CreateUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Printf("err: %s", err.Error())
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	token, err := uc.UserUseCase.CreateUser(input.Email, input.Name, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserExists):
			c.Status(http.StatusConflict)
		default:
			c.Status(http.StatusInternalServerError)
		}
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
func (uc *UserController) AuthorizeUser(c *gin.Context) {
	var input request.AuthUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	token, err := uc.UserUseCase.AuthorizeUser(input.Email, input.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.Status(http.StatusNotFound)
		case errors.Is(err, usecase.ErrInvalidPassword):
			c.Status(http.StatusConflict)
		default:
			c.Status(http.StatusInternalServerError)
		}
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
func (uc *UserController) UpdateUser(c *gin.Context) {
	var input request.UpdateUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	authUserID := c.GetUint("user_id")
	if err := uc.UserUseCase.UpdateUser(input, authUserID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.Status(http.StatusNotFound)
		case errors.Is(err, usecase.ErrUnprocessableEntity):
			c.Status(http.StatusUnprocessableEntity)
		case errors.Is(err, usecase.ErrUserForbidden):
			c.Status(http.StatusForbidden)
		default:
			c.Status(http.StatusInternalServerError)
		}
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
func (uc *UserController) GetUser(c *gin.Context) {
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

	userResp, err := uc.UserUseCase.GetUser(authUserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, userResp)
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
func (uc *UserController) DeleteUser(c *gin.Context) {
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

	if err := uc.UserUseCase.DeleteUser(authUserID); err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusInternalServerError)
		}
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
func (uc *UserController) EnableUser(c *gin.Context) {
	code := c.Param("code")

	token, err := uc.UserUseCase.EnableUser(code)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.Status(http.StatusNotFound)
		case errors.Is(err, usecase.ErrUnprocessableEntity):
			c.Status(http.StatusUnprocessableEntity)
		}
		return
	}

	c.JSON(
		http.StatusOK,
		response.AuthUserResponse{Token: token},
	)
}