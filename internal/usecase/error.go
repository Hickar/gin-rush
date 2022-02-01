package usecase

import "errors"

var (
	ErrUserExists = errors.New("user with such email already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid user password")
	ErrUserForbidden       = errors.New("authenticated user doesn't allowed to update user")
	ErrUnprocessableEntity = errors.New("invalid data format")
)
