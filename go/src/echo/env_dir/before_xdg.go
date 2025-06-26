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
	debug    debug.Options

	TempLocal, TempOS TemporaryFS
}

func (layout *beforeXDG) initialize(do debug.Options) (err error) {
	if layout.cwd, err = os.Getwd(); err != nil {
		err = errors.Wrap(err)
		return
	}

	layout.pid = os.Getpid()
	layout.dryRun = do.DryRun

	if layout.execPath, err = os.Executable(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO switch to useing MakeCommonEnv()
	{
		if err = os.Setenv(EnvBin, layout.execPath); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
