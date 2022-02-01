package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/go-redis/redis/v8"
)

type UserRepository struct {
	db *database.Database
	cache *redis.Client
}

func NewUserRepository(db *database.Database, cache *redis.Client) *UserRepository {
	return &UserRepository{db: db, cache: cache}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) UserWithEmailExists(email string) (bool, error) {
	return r.db.Exists(&models.User{}, "email", email)
}

func (r *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	return &user, r.db.FindBy(&user, "email", email)
}

func (r *UserRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	return &user, r.db.FindBy(&user, "id", id)
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	err := r.db.Save(user).Error
	if err != nil {
		return errors.New("unable to update user record in database")
	}

	cacheKey := fmt.Sprintf("users:%d", user.ID)
	return r.cache.Del(context.Background(), cacheKey).Err()
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	return &user, r.db.FindBy(&user, "id", id)
}

func (r *UserRepository) EnableUserByCode(code string) (*models.User, error) {
	var user models.User

	err := r.db.FindBy(&user, "confirmation_code", code)
	if err != nil {
		return &user, err
	}

	if user.Enabled {
		return &user, nil
	}

	user.Enabled = true
	return &user, r.db.Save(&user).Error
}

func (r *UserRepository) DeleteUser(user *models.User) error {
	err := r.db.Delete(user).Error
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("users:%d", user.ID)
	return r.cache.Del(context.Background(), cacheKey).Err()
}

