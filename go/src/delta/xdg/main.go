package xdg

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
)

type XDG struct {
	Home    env_vars.DirectoryLayoutBaseEnvVar
	Data    env_vars.DirectoryLayoutBaseEnvVar
	Config  env_vars.DirectoryLayoutBaseEnvVar
	State   env_vars.DirectoryLayoutBaseEnvVar
	Cache   env_vars.DirectoryLayoutBaseEnvVar
	Runtime env_vars.DirectoryLayoutBaseEnvVar

	AddedPath string // name of the utility
}

var _ interfaces.DirectoryLayout = XDG{}

func (exdg XDG) GetDirHome() interfaces.DirectoryLayoutBaseEnvVar { return exdg.Home }

func (exdg XDG) GetDirData() interfaces.DirectoryLayoutBaseEnvVar { return exdg.Data }

func (exdg XDG) GetDirConfig() interfaces.DirectoryLayoutBaseEnvVar { return exdg.Config }

func (exdg XDG) GetDirState() interfaces.DirectoryLayoutBaseEnvVar { return exdg.State }

func (exdg XDG) GetDirCache() interfaces.DirectoryLayoutBaseEnvVar { return exdg.Cache }

func (exdg XDG) GetDirRuntime() interfaces.DirectoryLayoutBaseEnvVar { return exdg.Runtime }

func (exdg XDG) GetXDGPaths() []string {
	return []string{
		exdg.Data.String(),
		exdg.Config.String(),
		exdg.State.String(),
		exdg.Cache.String(),
		exdg.Runtime.String(),
	}
}

func (exdg XDG) AddToEnvVars(envVars env_vars.EnvVars) {
	initElements := exdg.getInitElements()

	for _, element := range initElements {
		envVars[element.standard.Name] = *element.out
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
					return exdg.Home.String()

				default:
					return os.Getenv(v)
				}
			},
		)
	}

	if exdg.AddedPath != "" {
		out = filepath.Join(out, exdg.AddedPath)
	}

	return out, err
}

func (exdg *XDG) InitializeOverridden(
	addedPath string,
) (err error) {
	exdg.AddedPath = addedPath

	initElements := exdg.getInitElements()

	for _, ie := range initElements {
		// TODO validate this to prevent root xdg directories
		if *ie.out, err = exdg.setDefaultOrEnv(
			ie.standard.overridden,
			"",
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (exdg *XDG) InitializeStandardFromEnv(
	addedPath string,
) (err error) {
	exdg.AddedPath = addedPath

	toInitialize := exdg.getInitElements()

	for _, ie := range toInitialize {
		// TODO valid this to prevent root xdg directories
		if *ie.out, err = exdg.setDefaultOrEnv(
			ie.standard.DefaultValueTemplate,
			ie.standard.Name,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
