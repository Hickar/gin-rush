package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id int
	Name string
	Email string
	Password string
	Bio string
	Avatar string
	BirthDate time.Time
}