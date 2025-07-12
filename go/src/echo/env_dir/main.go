package env_dir

import (
	"fmt"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

const (
	EnvDir                    = "DIR_DODDER" // TODO chang to dodder-prefixed
	EnvBin                    = "BIN_DODDER" // TODO change to dodder-prefixed
	EnvXDGUtilityNameOverride = "DODDER_XDG_UTILITY_OVERRIDE"
	XDGUtilityName            = "dodder"
)

type Env interface {
	IsDryRun() bool
	GetCwd() string
	AddToEnvVars(env_vars.EnvVars)
	GetXDG() xdg.XDG
	GetExecPath() string
	GetTempLocal() TemporaryFS
	MakeDir(ds ...string) (err error)
	MakeDirPerms(perms os.FileMode, ds ...string) (err error)
	Rel(p string) (out string)
	RelToCwdOrSame(p string) (p1 string)
	MakeCommonEnv() map[string]string
	MakeRelativePathStringFormatWriter() interfaces.StringEncoderTo[string]
	AbsFromCwdOrSame(p string) (p1 string)

	Delete(paths ...string) (err error)
}

type env struct {
	errors.Context
	beforeXDG
	xdg.XDG
}

func MakeFromXDGDotenvPath(
	context errors.Context,
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
			context.CancelWithError(err)
		}
	}

	if _, err := dotenv.ReadFrom(file); err != nil {
		context.CancelWithError(err)
	}

	if err := file.Close(); err != nil {
		context.CancelWithError(err)
	}

	return MakeWithXDG(
		context,
		config.Debug,
		*dotenv.XDG,
	)
}

func MakeDefault(
	context errors.Context,
	do debug.Options,
) env {
	var home string

	{
		var err error

		if home, err = os.UserHomeDir(); err != nil {
			context.CancelWithError(err)
		}
	}

	return MakeWithHome(context, home, do, true)
}

func MakeDefaultAndInitialize(
	context errors.Context,
	do debug.Options,
	overrideXDG bool,
) env {
	var home string

	{
		var err error
		if home, err = os.UserHomeDir(); err != nil {
			context.CancelWithError(err)
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
	context errors.Context,
	home string,
	debugOptions debug.Options,
	permitCwdXDGOverride bool,
) (env env) {
	env.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.CancelWithError(err)
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
			env.CancelWithError(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			env.CancelWithError(err)
		}
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.CancelWithError(err)
	}

	env.AfterWithContext(env.resetTempOnExit)

	return
}

func MakeWithHomeAndInitialize(
	context errors.Context,
	home string,
	debugOptions debug.Options,
	cwdXDGOverride bool,
) (env env) {
	env.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.CancelWithError(err)
	}

	addedPath := "dodder"
	pathCwdXDGOverride := filepath.Join(env.cwd, ".dodder")

	if cwdXDGOverride {
		xdg.Home = pathCwdXDGOverride
		addedPath = ""
		if err := xdg.InitializeOverridden(addedPath); err != nil {
			env.CancelWithError(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			env.CancelWithError(err)
		}
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.CancelWithError(err)
	}

	env.AfterWithContext(env.resetTempOnExit)

	return
}

func MakeWithXDG(
	context errors.Context,
	debugOptions debug.Options,
	xdg xdg.XDG,
) (env env) {
	env.Context = context

	if err := env.beforeXDG.initialize(debugOptions); err != nil {
		env.CancelWithError(err)
	}

	if err := env.initializeXDG(xdg); err != nil {
		env.CancelWithError(err)
	}

	return
}

func (env *env) initializeXDG(xdg xdg.XDG) (err error) {
	env.XDG = xdg

	env.TempLocal.BasePath = filepath.Join(
		env.Cache,
		fmt.Sprintf("tmp-%d", env.GetPid()),
	)

	if err = env.MakeDir(env.GetTempLocal().BasePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (env env) GetDebug() debug.Options {
	return env.debugOptions
}

func (env env) IsDryRun() bool {
	return env.dryRun
}

func (env env) GetPid() int {
	return env.pid
}

func (env env) AddToEnvVars(envVars env_vars.EnvVars) {
	envVars[EnvBin] = env.GetExecPath()
}

func (env env) GetExecPath() string {
	return env.execPath
}

func (env env) GetCwd() string {
	return env.cwd
}

func (env env) GetXDG() xdg.XDG {
	return env.XDG
}

func (env *env) SetXDG(x xdg.XDG) {
	env.XDG = x
}

func (env env) GetTempLocal() TemporaryFS {
	return env.TempLocal
}

func (env env) AbsFromCwdOrSame(p string) (p1 string) {
	var err error
	p1, err = filepath.Abs(p)
	if err != nil {
		p1 = p
	}

	return
}

func (env env) RelToCwdOrSame(p string) (p1 string) {
	var err error

	if p1, err = filepath.Rel(env.GetCwd(), p); err != nil {
		p1 = p
	}

	return
}

func (env env) Rel(
	p string,
) (out string) {
	out = p

	p1, _ := filepath.Rel(env.GetCwd(), p)

	if p1 != "" {
		out = p1
	}

	return
}

func (env env) MakeCommonEnv() map[string]string {
	return map[string]string{
		EnvBin: env.GetExecPath(),
		// TODO determine if EnvDir is kept
		// EnvDir: h.Dir(),
	}
}

func (env env) MakeDir(ds ...string) (err error) {
	return env.MakeDirPerms(0o755, ds...)
}

func (env env) MakeDirPerms(perms os.FileMode, ds ...string) (err error) {
	for _, d := range ds {
		if err = os.MkdirAll(d, os.ModeDir|perms); err != nil {
			err = errors.Wrapf(err, "Dir: %q", d)
			return
		}
	}

	return
}
