package database

import (
	"errors"
	"fmt"
	"log"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *Database

type Database struct {
	db *gorm.DB
}

func New(user, pass, host, name string) (*Database, error) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", user, pass, host, name)
	db, err := gorm.Open(mysql.Open(dsn))

	if err != nil {
		return nil, err
	}

	_db = &Database{db: db}
	return _db, nil
}

func NewMockDB() (*Database, sqlmock.Sqlmock) {
	dbm, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("Unexpected error occured during creation of db mock: %s", err)
	}

	db, err := gorm.Open(mysql.New(
		mysql.Config{
			Conn:                      dbm,
			SkipInitializeWithVersion: true}),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		log.Fatalf("Unexpected error occured during creation of db mock: %s", err)
	}

	_db = &Database{db: db}
	return _db, mock
}

func GetDB() *Database {
	return _db
}

func (d *Database) Close() error {
	db, err := d.db.DB()
	if err != nil {
		log.Fatalf("Unexpected error during db closing: %s", err)
	}

	return db.Close()
}

func (d *Database) AutoMigrate(models interface{}) error {
	return d.db.AutoMigrate(models)
}

func (d *Database) Create(model interface{}) error {
	return d.db.Create(model).Error
}

func (d *Database) FindBy(model interface{}, field string, values interface{}) error {
	return d.db.Where(field + " = ?", values).First(model).Error
}

func (d *Database) FindByID(model interface{}, id uint) error {
	return d.db.First(model, id).Error
}

func (d *Database) Update(model interface{}) error {
	return d.db.Save(model).Error
}

func (d *Database) Exists(model interface{}, key string, value interface{}) (bool, error) {
	err := d.FindBy(model, key, value)

	if err != nil {
		switch {
		case errors.As(err, &gorm.ErrRecordNotFound):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (d *Database) Delete(model interface{}) error {
	return d.db.Delete(model).Error
}