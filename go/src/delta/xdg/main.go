package xdg

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
)

type XDG struct {
	Home      string
	AddedPath string // name of the utility

	Data    string
	Config  string
	State   string
	Cache   string
	Runtime string
}

type xdgInitElement struct {
	standard   string
	overridden string
	envKey     string
	out        *string
}

func (x XDG) GetXDGPaths() []string {
	return []string{
		x.Data,
		x.Config,
		x.State,
		x.Cache,
		x.Runtime,
	}
}

func (xdg XDG) AddToEnvVars(envVars env_vars.EnvVars) {
	initElements := xdg.getInitElements()

	for _, element := range initElements {
		envVars[element.envKey] = *element.out
	}
}

func (x *XDG) setDefaultOrEnv(
	defaultValue string,
	envKey string,
) (out string, err error) {
	if v, ok := os.LookupEnv(envKey); envKey != "" && ok {
		out = v
	} else {
		out = os.Expand(
			defaultValue,
			func(v string) string {
				switch v {
				case "HOME":
					return x.Home

				default:
					return os.Getenv(v)
				}
			},
		)
	}

	if x.AddedPath != "" {
		out = filepath.Join(out, x.AddedPath)
	}

	return
}

func (x *XDG) InitializeOverridden(
	addedPath string,
) (err error) {
	x.AddedPath = addedPath

	toInitialize := x.getInitElements()

	for _, ie := range toInitialize {
		if *ie.out, err = x.setDefaultOrEnv(
			ie.overridden,
			"",
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (x *XDG) InitializeStandardFromEnv(
	addedPath string,
) (err error) {
	x.AddedPath = addedPath

	toInitialize := x.getInitElements()

	for _, ie := range toInitialize {
		if *ie.out, err = x.setDefaultOrEnv(
			ie.standard,
			ie.envKey,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
