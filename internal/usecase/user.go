package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Hickar/gin-rush/internal/broker"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/mailer"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/logger"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/response"
	"github.com/Hickar/gin-rush/pkg/utils"
	"github.com/go-redis/redis/v8"
)

type UserUseCase struct {
	db     *database.Database
	cache  *redis.Client
	conf   *config.Config
	broker broker.Broker
}

func NewUserUseCase(db *database.Database, conf *config.Config, broker broker.Broker, cache *redis.Client) (*UserUseCase, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if conf == nil {
		return nil, errors.New("config is nil")
	}

	if broker == nil {
		return nil, errors.New("broker is nil")
	}

	if cache == nil {
		return nil, errors.New("redis client is nil")
	}

	return &UserUseCase{db: db, conf: conf, cache: cache, broker: broker}, nil
}

func (uc *UserUseCase) CreateUser(email, name, pass string) (string, error) {
	var user models.User
	if exists, _ := uc.db.Exists(&models.User{}, "email", email); exists {
		return "", ErrUserExists
	}

	salt, _ := security.RandomBytes(16)
	hashedPassword, err := security.HashPassword(pass, salt)
	if err != nil {
		return "", errors.New("unable to encrypt password")
	}

	user.Name = name
	user.Email = email
	user.Password = hashedPassword
	user.ConfirmationCode = utils.RandomString(30)
	user.Salt = salt

	err = uc.db.Create(&user)
	if err != nil {
		logger.GetLogger().Error(err)
		return "", errors.New("unable to create new user")
	}

	token, err := security.GenerateJWT(user.ID, uc.conf.Server.JWTSecret)
	if err != nil {
		return "", errors.New("can't generate jwt")
	}

	msg, err := json.Marshal(&mailer.ConfirmationMessage{
		Username: user.Name,
		Email:    user.Email,
		Code:     user.ConfirmationCode,
	})
	if err = uc.broker.Publish("mailer_ex", "mailer", "text/plain", &msg); err != nil {
		return "", errors.New("can't publish message to broker")
	}

	return token, nil
}

func (uc *UserUseCase) AuthorizeUser(email, pass string) (string, error) {
	var user models.User

	if err := uc.db.FindBy(&user, "email", email); err != nil {
		logger.GetLogger().Error(err)
		return "", ErrUserNotFound
	}

	valid := security.VerifyPassword(pass, user.Password, user.Salt)
	if !valid {
		return "", ErrInvalidPassword
	}

	token, err := security.GenerateJWT(user.ID, uc.conf.Server.JWTSecret)
	if err != nil {
		logger.GetLogger().Error(err)
		return "", errors.New("can't generate jwt")
	}

	return token, nil
}

func (uc *UserUseCase) UpdateUser(newUserInfo request.UpdateUserRequest, authUserID uint) error {
	var user models.User
	if err := uc.db.FindByID(&newUserInfo, authUserID); err != nil {
		logger.GetLogger().Error(err)
		return ErrUserNotFound
	}

	if authUserID != user.ID {
		return ErrUserForbidden
	}

	formattedTime, err := time.Parse("2006-01-02", newUserInfo.BirthDate)
	if err != nil {
		logger.GetLogger().Error(err)
		return ErrUnprocessableEntity
	}

	user.Name = newUserInfo.Name
	user.Bio = sql.NullString{String: newUserInfo.Bio, Valid: true}
	user.BirthDate = sql.NullTime{Time: formattedTime, Valid: true}
	user.Avatar = sql.NullString{String: newUserInfo.Avatar, Valid: true}

	if err = uc.db.Update(&newUserInfo); err != nil {
		logger.GetLogger().Error(err)
		return ErrUnprocessableEntity
	}

	cacheKey := fmt.Sprintf("users:%d", authUserID)
	err = uc.cache.Del(context.Background(), cacheKey).Err()
	if err != nil {
		logger.GetLogger().Error(err)
		return errors.New("can't delete cache record")
	}

	return nil
}

func (uc *UserUseCase) GetUser(userID uint) (*response.UpdateUserResponse, error) {
	var userResp response.UpdateUserResponse
	var user models.User

	cacheKey := fmt.Sprintf("users:%d", userID)
	cached, _ := uc.cache.Get(context.Background(), cacheKey).Result()
	if cached != "" {
		if err := json.Unmarshal([]byte(cached), &userResp); err != nil {
			logger.GetLogger().Error(err)
			return nil, errors.New("can't unmarshall user cached data")
		}

	} else {
		if err := uc.db.FindByID(&user, userID); err != nil {
			logger.GetLogger().Error(err)
			return nil, ErrUserNotFound
		}

		userResp.Name = user.Name
		userResp.Bio = user.Bio.String
		userResp.Avatar = user.Avatar.String
		userResp.BirthDate = user.BirthDate.Time.String()

		cacheData, err := json.Marshal(&userResp)
		if err != nil {
			logger.GetLogger().Error(err)
			return nil, errors.New("can't marshall user data for caching")
		}

		if _, err := uc.cache.Set(context.Background(), cacheKey, cacheData, time.Hour*168).Result(); err != nil {
			logger.GetLogger().Error(err)
			return nil, errors.New("can't set cache value")
		}
	}

	return &userResp, nil
}

func (uc *UserUseCase) DeleteUser(userID uint) error {
	var user models.User

	if err := uc.db.FindByID(&user, userID); err != nil {
		logger.GetLogger().Error(err)
		return ErrUserNotFound
	}

	if err := uc.db.Delete(&user); err != nil {
		logger.GetLogger().Error(err)
		return errors.New("can't delete user record in db")
	}

	cacheKey := fmt.Sprintf("users:%d", userID)
	if err := uc.cache.Del(context.Background(), cacheKey).Err(); err != nil {
		logger.GetLogger().Error(err)
		return errors.New("can't delete user record in cache")
	}

	return nil
}

func (uc *UserUseCase) EnableUser(code string) (string, error) {
	var user models.User

	if len(code) != 30 {
		return "", ErrUnprocessableEntity
	}

	if err := uc.db.FindBy(&user, "confirmation_code", code); err != nil {
		logger.GetLogger().Error(err)
		return "", ErrUserNotFound
	}

	if !user.Enabled {
		user.Enabled = true
		if err := uc.db.Update(&user); err != nil {
			return "", errors.New("can't update db user record")
		}
	}

	token, err := security.GenerateJWT(user.ID, uc.conf.Server.JWTSecret)
	if err != nil {
		logger.GetLogger().Error(err)
		return "", errors.New("unable to generate jwt")
	}

	return token, nil
}
