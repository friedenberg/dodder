package env_dir

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
)

type beforeXDG struct {
	cwd      string
	execPath string
	pid      int
	dryRun   bool
	debugOptions    debug.Options

	TempLocal, TempOS TemporaryFS
}

func (env *beforeXDG) initialize(debugOptions debug.Options) (err error) {
	env.debugOptions = debugOptions

	if env.cwd, err = os.Getwd(); err != nil {
		err = errors.Wrap(err)
		return
	}

	env.pid = os.Getpid()
	env.dryRun = debugOptions.DryRun

	if env.execPath, err = os.Executable(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO switch to useing MakeCommonEnv()
	{
		if err = os.Setenv(EnvBin, env.execPath); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
