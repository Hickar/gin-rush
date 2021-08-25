package models

import (
	"database/sql"

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