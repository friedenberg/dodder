package env_dir

import (
	"fmt"
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/foxtrot/config_mutable_cli"
)

const (
	EnvDir = "DIR_DODDER"
	EnvBin = "BIN_DODDER"
)

type Env interface {
	IsDryRun() bool
	GetCwd() string
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
	config config_mutable_cli.Config,
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
	do debug.Options,
	permitCwdXDGOverride bool,
) (s env) {
	s.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := s.beforeXDG.initialize(do); err != nil {
		s.CancelWithError(err)
	}

	addedPath := "dodder"
	pathCwdXDGOverride := filepath.Join(s.cwd, ".dodder")

	if permitCwdXDGOverride && files.Exists(pathCwdXDGOverride) {
		xdg.Home = pathCwdXDGOverride
		addedPath = ""
		if err := xdg.InitializeOverridden(addedPath); err != nil {
			s.CancelWithError(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			s.CancelWithError(err)
		}
	}

	if err := s.initializeXDG(xdg); err != nil {
		s.CancelWithError(err)
	}

	s.AfterWithContext(s.resetTempOnExit)

	return
}

func MakeWithHomeAndInitialize(
	context errors.Context,
	home string,
	do debug.Options,
	cwdXDGOverride bool,
) (s env) {
	s.Context = context

	xdg := xdg.XDG{
		Home: home,
	}

	if err := s.beforeXDG.initialize(do); err != nil {
		s.CancelWithError(err)
	}

	addedPath := "dodder"
	pathCwdXDGOverride := filepath.Join(s.cwd, ".dodder")

	if cwdXDGOverride {
		xdg.Home = pathCwdXDGOverride
		addedPath = ""
		if err := xdg.InitializeOverridden(addedPath); err != nil {
			s.CancelWithError(err)
		}
	} else {
		if err := xdg.InitializeStandardFromEnv(addedPath); err != nil {
			s.CancelWithError(err)
		}
	}

	if err := s.initializeXDG(xdg); err != nil {
		s.CancelWithError(err)
	}

	s.AfterWithContext(s.resetTempOnExit)

	return
}

func MakeWithXDG(
	context errors.Context,
	do debug.Options,
	xdg xdg.XDG,
) (s env) {
	s.Context = context

	if err := s.beforeXDG.initialize(do); err != nil {
		s.CancelWithError(err)
	}

	if err := s.initializeXDG(xdg); err != nil {
		s.CancelWithError(err)
	}

	return
}

func (layout *env) initializeXDG(xdg xdg.XDG) (err error) {
	layout.XDG = xdg

	layout.TempLocal.BasePath = filepath.Join(
		layout.Cache,
		fmt.Sprintf("tmp-%d", layout.GetPid()),
	)

	if err = layout.MakeDir(layout.GetTempLocal().BasePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (h env) GetDebug() debug.Options {
	return h.debug
}

func (h env) IsDryRun() bool {
	return h.dryRun
}

func (h env) GetPid() int {
	return h.pid
}

func (h env) GetExecPath() string {
	return h.execPath
}

func (h env) GetCwd() string {
	return h.cwd
}

func (h env) GetXDG() xdg.XDG {
	return h.XDG
}

func (h *env) SetXDG(x xdg.XDG) {
	h.XDG = x
}

func (h env) GetTempLocal() TemporaryFS {
	return h.TempLocal
}

func (s env) AbsFromCwdOrSame(p string) (p1 string) {
	var err error
	p1, err = filepath.Abs(p)
	if err != nil {
		p1 = p
	}

	return
}

func (s env) RelToCwdOrSame(p string) (p1 string) {
	var err error

	if p1, err = filepath.Rel(s.GetCwd(), p); err != nil {
		p1 = p
	}

	return
}

func (s env) Rel(
	p string,
) (out string) {
	out = p

	p1, _ := filepath.Rel(s.GetCwd(), p)

	if p1 != "" {
		out = p1
	}

	return
}

func (h env) MakeCommonEnv() map[string]string {
	return map[string]string{
		EnvBin: h.GetExecPath(),
		// TODO determine if EnvDir is kept
		// EnvDir: h.Dir(),
	}
}

func (s env) MakeDir(ds ...string) (err error) {
	return s.MakeDirPerms(0o755, ds...)
}

func (s env) MakeDirPerms(perms os.FileMode, ds ...string) (err error) {
	for _, d := range ds {
		if err = os.MkdirAll(d, os.ModeDir|perms); err != nil {
			err = errors.Wrapf(err, "Dir: %q", d)
			return
		}
	}

	return
}
