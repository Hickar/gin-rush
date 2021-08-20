package models

import (
	"database/sql"
	"log"
	"time"

	"gorm.io/gorm"
)

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

func CreateUser(data map[string]interface{}) (*User, error) {
	user := &User{
		Name:             data["name"].(string),
		Email:            data["email"].(string),
		Password:         data["password"].([]byte),
		Salt:             data["salt"].([]byte),
		ConfirmationCode: data["confirmation_code"].(string),
	}

	log.Printf("len(code): %d\n", len(data["confirmation_code"].(string)))

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

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := DB.Where("email = ?", email).First(&user).Error

	return &user, err
}

func GetUserByID(id uint) (*User, error) {
	var user User
	err := DB.First(&user, id).Error

	return &user, err
}

func GetUserByConfirmationCode(code string) (*User, error)  {
	var user User
	err := DB.Where("confirmation_code = ?", code).First(&user).Error

	return &user, err
}

func UpdateUser(user User, data map[string]interface{}) error {
	err := DB.Model(&user).Updates(&User{
		Name:      data["name"].(string),
		Bio:       sql.NullString{String: data["bio"].(string), Valid: true},
		BirthDate: sql.NullTime{Time: data["birth_date"].(time.Time), Valid: true},
		Avatar:    sql.NullString{String: data["avatar"].(string), Valid: true},
	}).Error

	return err
}

func EnableUser(user User) error {
	err := DB.Model(&user).Updates(&User{
		Enabled: true,
	}).Error

	return err
}

func DeleteUser(user User) error {
	return DB.Delete(&user).Error
}
