package database

import (
	"errors"
	"fmt"

	"github.com/Hickar/gin-rush/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	*gorm.DB
}

func NewDatabase(conf *config.DatabaseConfig, gormConf *gorm.Config) (*Database, error) {
	if conf == nil {
		return nil, errors.New("error during db setup: no conf provided")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true", conf.User, conf.Password, conf.Host, conf.Name)

	db, err := gorm.Open(mysql.Open(dsn), gormConf)
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (db *Database) FindBy(model interface{}, field string, values interface{}) error {
	return db.Where(field + " = ?", values).First(model).Error
}

func (db *Database) Exists(model interface{}, key string, value interface{}) (bool, error) {
	err := db.FindBy(model, key, value)

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