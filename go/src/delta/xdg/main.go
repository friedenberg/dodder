package xdg

import (
	"fmt"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/xdg_defaults"
)

type XDG struct {
	Home env_vars.DirectoryLayoutBaseEnvVar
	Cwd  env_vars.DirectoryLayoutBaseEnvVar

	UtilityName string

	Data    env_vars.DirectoryLayoutBaseEnvVar
	Config  env_vars.DirectoryLayoutBaseEnvVar
	State   env_vars.DirectoryLayoutBaseEnvVar
	Cache   env_vars.DirectoryLayoutBaseEnvVar
	Runtime env_vars.DirectoryLayoutBaseEnvVar
}

var _ interfaces.DirectoryLayout = XDG{}

func (xdg XDG) GetDirHome() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Home }

func (xdg XDG) GetDirCwd() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Cwd }

func (xdg XDG) GetDirData() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Data }

func (xdg XDG) GetDirConfig() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Config }

func (xdg XDG) GetDirState() interfaces.DirectoryLayoutBaseEnvVar { return xdg.State }

func (xdg XDG) GetDirCache() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Cache }

func (xdg XDG) GetDirRuntime() interfaces.DirectoryLayoutBaseEnvVar { return xdg.Runtime }

func (xdg XDG) GetXDGEnvVars() []env_vars.DirectoryLayoutBaseEnvVar {
	return []env_vars.DirectoryLayoutBaseEnvVar{
		xdg.Data,
		xdg.Config,
		xdg.State,
		xdg.Cache,
		xdg.Runtime,
	}
}

func (xdg XDG) GetXDGPaths() []string {
	return []string{
		xdg.Data.String(),
		xdg.Config.String(),
		xdg.State.String(),
		xdg.Cache.String(),
		xdg.Runtime.String(),
	}
}

func (xdg XDG) AddToEnvVars(envVars interfaces.EnvVars) {
	initElements := xdg.getInitElements()

	for _, element := range initElements {
		envVars[element.defawlt.Name] = element.actual.ActualValue
	}
}

func (xdg *XDG) setInitArgs(initArgs InitArgs) (err error) {
	xdg.Home = xdg_defaults.Home.MakeBaseEnvVar(initArgs.Home)
	xdg.Cwd = xdg_defaults.Cwd.MakeBaseEnvVar(initArgs.Cwd)
	xdg.UtilityName = initArgs.UtilityName

	return err
}

func (xdg *XDG) InitializeOverriddenIfNecessary(
	initArgs InitArgs,
) (err error) {
	pathCwdXDGOverride := filepath.Join(
		initArgs.Cwd,
		fmt.Sprintf(".%s", initArgs.UtilityName),
	)

	if files.Exists(pathCwdXDGOverride) {
		if err = xdg.InitializeOverridden(initArgs); err != nil {
			err = errors.Wrap(err)
			return err
		}
	} else {
		if err = xdg.InitializeStandardFromEnv(initArgs); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (xdg *XDG) InitializeOverridden(
	initArgs InitArgs,
) (err error) {
	if err = xdg.setInitArgs(initArgs); err != nil {
		err = errors.Wrap(err)
		return err
	}

	getenv := xdg_defaults.MakeGetenv(
		os.Getenv,
		xdg.Cwd.ActualValue,
		xdg.UtilityName,
	)

	for _, initElement := range xdg.getInitElements() {
		initElement.actual.Name = initElement.defawlt.Name
		initElement.actual.DefaultValueTemplate = initElement.defawlt.TemplateOverride

		if err = initElement.actual.InitializeXDGTemplate(getenv); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (xdg *XDG) InitializeStandardFromEnv(
	initArgs InitArgs,
) (err error) {
	if err = xdg.setInitArgs(initArgs); err != nil {
		err = errors.Wrap(err)
		return err
	}

	getenv := xdg_defaults.MakeGetenv(
		os.Getenv,
		xdg.Cwd.ActualValue,
		xdg.UtilityName,
	)

	for _, initElement := range xdg.getInitElements() {
		initElement.actual.Name = initElement.defawlt.Name
		initElement.actual.DefaultValueTemplate = initElement.defawlt.TemplateDefault

		if err = initElement.actual.InitializeXDGEnvVarOrTemplate(
			xdg.UtilityName,
			getenv,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
