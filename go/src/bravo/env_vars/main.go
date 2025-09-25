package env_vars

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO add support for comments
type EnvVars map[string]string

type Adder interface {
	AddToEnvVars(EnvVars)
}

func Make(adders ...Adder) EnvVars {
	envVars := make(EnvVars)

	for _, adder := range adders {
		adder.AddToEnvVars(envVars)
	}

	return envVars
}

func (envVars EnvVars) Setenv() (err error) {
	for key, value := range envVars {
		if err = os.Setenv(key, value); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
