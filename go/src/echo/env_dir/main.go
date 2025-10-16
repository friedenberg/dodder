package env_dir

import (
	"fmt"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
)

const (
	EnvDir                    = "DIR_DODDER" // TODO chang to dodder-prefixed
	EnvBin                    = "BIN_DODDER" // TODO change to dodder-prefixed
	EnvXDGUtilityNameOverride = "DODDER_XDG_UTILITY_OVERRIDE"
	XDGUtilityNameDodder      = "dodder"
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
	interfaces.Context
	beforeXDG
	xdg.XDG
}

// sets XDG and creates tmp local
func (env *env) initializeXDG(xdg xdg.XDG) (err error) {
	env.XDG = xdg

	env.TempLocal.BasePath = env.Cache.MakePath(
		fmt.Sprintf("tmp-%d", env.GetPid()),
	).String()

	if err = env.MakeDir(env.GetTempLocal().BasePath); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
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

	return p1
}

func (env env) RelToCwdOrSame(p string) (p1 string) {
	var err error

	if p1, err = filepath.Rel(env.GetCwd(), p); err != nil {
		p1 = p
	}

	return p1
}

func (env env) Rel(
	p string,
) (out string) {
	out = p

	p1, _ := filepath.Rel(env.GetCwd(), p)

	if p1 != "" {
		out = p1
	}

	return out
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
			return err
		}
	}

	return err
}
