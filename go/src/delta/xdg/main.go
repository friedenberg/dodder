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

func (exdg XDG) GetXDGPaths() []string {
	return []string{
		exdg.Data,
		exdg.Config,
		exdg.State,
		exdg.Cache,
		exdg.Runtime,
	}
}

func (exdg XDG) AddToEnvVars(envVars env_vars.EnvVars) {
	initElements := exdg.getInitElements()

	for _, element := range initElements {
		envVars[element.envKey] = *element.out
	}
}

// TODO simplify this and document it
func (exdg *XDG) setDefaultOrEnv(
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
					return exdg.Home

				default:
					return os.Getenv(v)
				}
			},
		)
	}

	if exdg.AddedPath != "" {
		out = filepath.Join(out, exdg.AddedPath)
	}

	return
}

func (exdg *XDG) InitializeOverridden(
	addedPath string,
) (err error) {
	exdg.AddedPath = addedPath

	toInitialize := exdg.getInitElements()

	for _, ie := range toInitialize {
		// TODO valid this to prevent root xdg directories
		if *ie.out, err = exdg.setDefaultOrEnv(
			ie.overridden,
			"",
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (exdg *XDG) InitializeStandardFromEnv(
	addedPath string,
) (err error) {
	exdg.AddedPath = addedPath

	toInitialize := exdg.getInitElements()

	for _, ie := range toInitialize {
		// TODO valid this to prevent root xdg directories
		if *ie.out, err = exdg.setDefaultOrEnv(
			ie.standard,
			ie.envKey,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
