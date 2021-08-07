package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name      string `gorm:"not null"`
	Email     string `gorm:"not null;unique"`
	Bio       sql.NullString
	Avatar    sql.NullString
	BirthDate sql.NullTime
	Enabled   bool `gorm:"default:false"`
}

func CreateUser(data map[string]interface{}) (*User, error) {
	user := &User{
		Name:     data["name"].(string),
		Email:    data["email"].(string),
	}

	if err := DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func UserExistsByEmail(email string) (bool, error) {
	var user User
	err := DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		return false, err
	}

	return true, nil
}
