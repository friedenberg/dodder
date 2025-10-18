package env_dir

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

// TODO separate read-only from write

func MakeDefault(
	context interfaces.Context,
	utilityName string,
	debugOptions debug.Options,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	return MakeWithHome(
		context,
		home,
		utilityName,
		debugOptions,
		true,
		true,
	)
}

func MakeDefaultNoInit(
	context interfaces.Context,
	utilityName string,
	debugOptions debug.Options,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	return MakeWithHome(
		context,
		home,
		utilityName,
		debugOptions,
		true,
		false,
	)
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
	utilityName string,
	do debug.Options,
	overrideXDGWithCwd bool,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.Cancel(err)
		}
	}

	if overrideXDGWithCwd {
		var cwd string

		{
			var err error

			if cwd, err = os.Getwd(); err != nil {
				context.Cancel(err)
			}
		}

		return MakeWithXDGRootOverrideHomeAndInitialize(
			context,
			cwd,
			utilityName,
			do,
		)
	} else {
		return MakeWithHomeAndInitialize(
			context,
			home,
			do,
		)
	}
}

func MakeWithHome(
	context interfaces.Context,
	home string,
	utilityName string,
	debugOptions debug.Options,
	permitCwdXDGOverride bool,
	initialize bool,
) (env env) {
	env.Context = context

	if err := env.beforeXDG.initialize(debugOptions, utilityName); err != nil {
		env.Cancel(err)
		return env
	}

	if !initialize {
		return env
	}

	if permitCwdXDGOverride {
		if err := env.XDG.InitializeOverriddenIfNecessary(env.xdgInitArgs); err != nil {
			env.Cancel(err)
			return env
		}
	} else {
		if err := env.XDG.InitializeStandardFromEnv(env.xdgInitArgs); err != nil {
			env.Cancel(err)
			return env
		}
	}

	if err := env.initializeXDG(); err != nil {
		env.Cancel(err)
		return env
	}

	env.After(env.resetTempOnExit)

	return env
}

func MakeWithXDGRootOverrideHomeAndInitialize(
	context interfaces.Context,
	xdgRootOverride string,
	utilityName string,
	debugOptions debug.Options,
) (env env) {
	env.Context = context
	env.xdgInitArgs.Cwd = xdgRootOverride

	if err := env.beforeXDG.initialize(debugOptions, utilityName); err != nil {
		env.Cancel(err)
		return env
	}

	if err := env.XDG.InitializeOverridden(env.xdgInitArgs); err != nil {
		env.Cancel(err)
		return env
	}

	if err := env.initializeXDG(); err != nil {
		env.Cancel(err)
		return env
	}

	env.After(env.resetTempOnExit)

	return env
}

func MakeWithHomeAndInitialize(
	context interfaces.Context,
	home string,
	debugOptions debug.Options,
) (env env) {
	env.Context = context

	if err := env.beforeXDG.initialize(debugOptions, "dodder"); err != nil {
		env.Cancel(err)
	}

	if err := env.XDG.InitializeStandardFromEnv(env.xdgInitArgs); err != nil {
		env.Cancel(err)
		return env
	}

	if err := env.initializeXDG(); err != nil {
		env.Cancel(err)
		return env
	}

	env.After(env.resetTempOnExit)

	return env
}

func MakeWithXDG(
	context interfaces.Context,
	debugOptions debug.Options,
	xdg xdg.XDG,
) (env env) {
	env.Context = context
	env.XDG = xdg

	if err := env.beforeXDG.initialize(debugOptions, xdg.UtilityName); err != nil {
		env.Cancel(err)
		return env
	}

	if err := env.initializeXDG(); err != nil {
		env.Cancel(err)
		return env
	}

	return env
}
