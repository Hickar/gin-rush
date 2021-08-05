package models

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Setup() error {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return errors.New("no environment variables provided")
	}

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName)
	db, err = gorm.Open(mysql.Open(dsn))

	if err != nil {
		log.Fatalf("DB Setup: %s", err)
	}

	return nil
}

func MakeMigrations() error {
	err := db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}