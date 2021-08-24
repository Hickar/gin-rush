package database

import (
	"fmt"
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Setup(user, pass, host, name string) (*gorm.DB, error) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, name)
	db, err = gorm.Open(mysql.Open(dsn))

	if err != nil {
		return nil, err
	}

	return db, nil
}

func SetupMockDB() (*gorm.DB, sqlmock.Sqlmock) {
	dbm, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("Unexpected error occured during creation of db mock: %s", err)
	}

	db, err = gorm.Open(mysql.New(
		mysql.Config{
			Conn:                      dbm,
			SkipInitializeWithVersion: true}),
		&gorm.Config{},
	)
	if err != nil {
		log.Fatalf("Unexpected error occured during creation of db mock: %s", err)
	}

	return db, mock
}

func GetDB() *gorm.DB {
	return db
}

func Close() error {
	db, err := db.DB()
	if err != nil {
		log.Fatalf("Unexpected error during db closing: %s", err)
	}

	return db.Close()
}