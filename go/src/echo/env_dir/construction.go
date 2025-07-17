package env_dir

import (
	"fmt"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

// TODO separate read-only from write

func MakeDefault(
	context interfaces.Context,
	debugOptions debug.Options,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	return MakeWithHome(context, home, debugOptions, true, true)
}

func MakeDefaultNoInit(
	context interfaces.Context,
	debugOptions debug.Options,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	return MakeWithHome(context, home, debugOptions, true, false)
}

func MakeFromXDGDotenvPath(
	context interfaces.Context,
	config repo_config_cli.Config,
	xdgDotenvPath string,
) env {
	dotenv := xdg.Dotenv{
		XDG: &xdg.XDG{},
	}

	var file *os.File

	{
		var err error

		if file, err = os.Open(xdgDotenvPath); err != nil {
			context.Cancel(err)
		}
	}

	if _, err := dotenv.ReadFrom(file); err != nil {
		context.Cancel(err)
	}

	if err := file.Close(); err != nil {
		context.Cancel(err)
	}

	return MakeWithXDG(
		context,
		config.Debug,
		*dotenv.XDG,
	)
}

func MakeDefaultAndInitialize(
	context interfaces.Context,
	do debug.Options,
	overrideXDG bool,
) env {
	var home string

	{
		var err error
		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	return MakeWithHomeAndInitialize(
		context,
		home,
		do,
		overrideXDG,
	)
}

func MakeWithHome(
	context interfaces.Context,
	home string,
	debugOptions debug.Options,
	permitCwdXDGOverride bool,
	initialize bool,
) (env env) {
	env.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.Cancel(err)
	}

	if !initialize {
		return
	}

	addedPath := XDGUtilityName

	if addedPathOverride := os.Getenv(EnvXDGUtilityNameOverride); addedPathOverride != "" {
		addedPath = addedPathOverride
	}

	pathCwdXDGOverride := filepath.Join(env.cwd, fmt.Sprintf(".%s", addedPath))

	if permitCwdXDGOverride && files.Exists(pathCwdXDGOverride) {
		xdg.Home = pathCwdXDGOverride
		addedPath = ""
		if err := xdg.InitializeOverridden(addedPath); err != nil {
			env.Cancel(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			env.Cancel(err)
		}
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.Cancel(err)
	}

	env.After(env.resetTempOnExit)

	return
}

func MakeWithHomeAndInitialize(
	context interfaces.Context,
	home string,
	debugOptions debug.Options,
	cwdXDGOverride bool,
) (env env) {
	env.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.Cancel(err)
	}

	addedPath := "dodder"
	pathCwdXDGOverride := filepath.Join(env.cwd, ".dodder")

	if cwdXDGOverride {
		xdg.Home = pathCwdXDGOverride
		addedPath = ""
		if err := xdg.InitializeOverridden(addedPath); err != nil {
			env.Cancel(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			env.Cancel(err)
		}
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.Cancel(err)
	}

	env.After(env.resetTempOnExit)

	return
}

func MakeWithXDG(
	context interfaces.Context,
	debugOptions debug.Options,
	xdg xdg.XDG,
) (env env) {
	env.Context = context

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.Cancel(err)
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.Cancel(err)
	}

	return
}
