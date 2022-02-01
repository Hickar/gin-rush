package usecase

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Hickar/gin-rush/internal/broker"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/mailer"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/repository"
	"github.com/Hickar/gin-rush/pkg/logger"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/security"
	"github.com/Hickar/gin-rush/pkg/utils"
)

type UserUseCase struct {
	repo   *repository.UserRepository
	conf   *config.Config
	broker broker.Broker
	logger logger.Logger
}

func NewUserUseCase(repo *repository.UserRepository, conf *config.Config, broker broker.Broker, logger logger.Logger) (*UserUseCase, error) {
	if repo == nil {
		return nil, errors.New("user repository is nil")
	}

	if conf == nil {
		return nil, errors.New("config is nil")
	}

	if broker == nil {
		return nil, errors.New("broker is nil")
	}

	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	return &UserUseCase{repo: repo, conf: conf, broker: broker, logger: logger}, nil
}

func (uc *UserUseCase) CreateUser(email, name, pass string) (string, error) {
	var user models.User
	if exists, _ := uc.repo.UserWithEmailExists(email); exists {
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

	err = uc.repo.CreateUser(&user)
	if err != nil {
		uc.logger.Error(err)
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
	user, err := uc.repo.FindUserByEmail(email)
	if err != nil {
		uc.logger.Error(err)
		return "", ErrUserNotFound
	}

	valid := security.VerifyPassword(pass, user.Password, user.Salt)
	if !valid {
		return "", ErrInvalidPassword
	}

	token, err := security.GenerateJWT(user.ID, uc.conf.Server.JWTSecret)
	if err != nil {
		uc.logger.Error(err)
		return "", errors.New("can't generate jwt")
	}

	return token, nil
}

func (uc *UserUseCase) UpdateUser(newUserInfo request.UpdateUserRequest, authUserID uint) error {
	user, err := uc.repo.FindUserByID(authUserID)
	if err != nil {
		uc.logger.Error(err)
		return ErrUserNotFound
	}

	if authUserID != user.ID {
		return ErrUserForbidden
	}

	formattedTime, err := time.Parse("2006-01-02", newUserInfo.BirthDate)
	if err != nil {
		uc.logger.Error(err)
		return ErrUnprocessableEntity
	}

	user.Name = newUserInfo.Name
	user.Bio = sql.NullString{String: newUserInfo.Bio, Valid: true}
	user.BirthDate = sql.NullTime{Time: formattedTime, Valid: true}
	user.Avatar = sql.NullString{String: newUserInfo.Avatar, Valid: true}

	if err = uc.repo.UpdateUser(user); err != nil {
		uc.logger.Error(err)
		return ErrUnprocessableEntity
	}

	return nil
}

func (uc *UserUseCase) GetUser(userID uint) (*models.User, error) {
	return uc.repo.FindUserByID(userID)
}

func (uc *UserUseCase) DeleteUser(userID uint) error {
	user, err := uc.repo.FindUserByID(userID)
	if err != nil {
		uc.logger.Error(err)
		return ErrUserNotFound
	}

	if err = uc.repo.DeleteUser(user); err != nil {
		uc.logger.Error(err)
		return errors.New("can't delete user record in db")
	}

	return nil
}

func (uc *UserUseCase) EnableUser(code string) (string, error) {
	if len(code) != 30 {
		return "", ErrUnprocessableEntity
	}

	user, err := uc.repo.EnableUserByCode(code)
	if err != nil {
		return "", err
	}

	token, err := security.GenerateJWT(user.ID, uc.conf.Server.JWTSecret)
	if err != nil {
		uc.logger.Error(err)
		return "", errors.New("unable to generate jwt")
	}

	return token, nil
}
