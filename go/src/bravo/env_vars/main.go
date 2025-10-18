package env_vars

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO add support for comments
type EnvVars map[string]string

type Adder = interfaces.EnvVarsAdder

type Getenv = func(string) string

func Make(adders ...interfaces.EnvVarsAdder) EnvVars {
	envVars := make(EnvVars)

	for _, adder := range adders {
		adder.AddToEnvVars(envVars)
	}

	return envVars
}

func (envVars EnvVars) Set() (err error) {
	for key, value := range envVars {
		if err = os.Setenv(key, value); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (envVars EnvVars) Unset() (err error) {
	for key := range envVars {
		if err = os.Unsetenv(key); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (envVars EnvVars) GetWithoutOSFallback(lookup string) string {
	value := envVars[lookup]
	return value
}

func (envVars EnvVars) GetWithOSFallback(lookup string) string {
	if value := envVars.GetWithoutOSFallback(lookup); value != "" {
		return value
	}

	return os.Getenv(lookup)
}
