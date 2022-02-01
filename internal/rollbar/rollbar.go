package rollbar

import (
	"errors"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/rollbar/rollbar-go"
)

func NewRollbar(conf *config.RollbarConfig) (*rollbar.Client, error) {
	if conf == nil {
		return nil, errors.New("rollbar setup error: no configuration provided")
	}

	client := rollbar.New(conf.Token, conf.Environment, conf.Version, conf.ServerHost, conf.ServerRoot)
	return client, nil
}