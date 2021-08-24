package repositories

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

type User struct {
	gorm.Model
	Name             string `gorm:"not null"`
	Email            string `gorm:"not null;unique"`
	Password         []byte `gorm:"type:varbinary(32);not null"`
	Salt             []byte `gorm:"type:varbinary(16);not null"`
	Bio              sql.NullString
	Avatar           sql.NullString
	BirthDate        sql.NullTime
	Enabled          bool   `gorm:"default:false"`
	ConfirmationCode string `gorm:"type:varchar(255);not null;unique"`
}

func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("DB must not be nil")
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}

	return &UserRepository{db}, nil
}

func (r *UserRepository) Create(user *User) (*User, error) {
	if err := r.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) FindBy(field string, values interface{}) (*User, error) {
	var user User

	if err := r.DB.Where(field + " = ?", values).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var user User
	err := r.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *UserRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.DB.First(&user, id).Error

	return &user, err
}

func (r *UserRepository) Update(user User) error {
	return r.DB.Save(&user).Error
}

func (r *UserRepository) Delete(user User) error {
	return r.DB.Delete(&user).Error
}
