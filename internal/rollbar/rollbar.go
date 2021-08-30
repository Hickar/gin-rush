package rollbar

import (
	"errors"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/rollbar/rollbar-go"
)

func New(conf *config.RollbarConfig) error {
	if conf == nil {
		return errors.New("rollbar setup error: no configuration provided")
	}

	rollbar.SetToken(conf.Token)
	rollbar.SetEnvironment(conf.Environment)
	rollbar.SetCodeVersion("v2")
	rollbar.SetServerHost("web.1")
	rollbar.SetServerRoot(conf.ServerRoot)

	rollbar.Wait()

	return nil
}
