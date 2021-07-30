package rollbarinit

import (
	"errors"
	"os"

	"github.com/rollbar/rollbar-go"
)

func Setup() error {
	token := os.Getenv("ROLLBAR_TOKEN")
	environment := os.Getenv("ROLLBAR_ENV_MODE")

	if token == "" || environment == "" {
		return errors.New("no rollbar token or env mode provided")
	}

	rollbar.SetToken(token)
	rollbar.SetEnvironment(environment)
	rollbar.SetCodeVersion("v2")
	rollbar.SetServerHost("web.1")
	rollbar.SetServerRoot("github.com/Hickar/gin-rush")

	rollbar.Wait()

	return nil
}
