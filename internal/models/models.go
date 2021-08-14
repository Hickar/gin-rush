package models

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Setup() error {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return errors.New("no environment variables provided")
	}

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", dbUser, dbPassword, dbHost, dbName)
	DB, err = gorm.Open(mysql.Open(dsn))

	if err != nil {
		log.Fatalf("DB Setup: %s", err)
	}

	if err := makeMigrations(); err != nil {
		log.Fatalf("Can't make migrations: %s", err)
	}

	return nil
}

func makeMigrations() error {
	err := DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&User{})
	if err != nil {
		return err
	}

	return nil
}