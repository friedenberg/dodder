package env_dir

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO only call reset temp when actually not resetting temp
func (s env) resetTempOnExit(ctx errors.Context) (err error) {
	errIn := ctx.Cause()

	if errIn != nil || s.GetDebug().NoTempDirCleanup {
		// ui.Err().Printf("temp dir: %q", s.TempLocal.BasePath)
	} else {
		if err = os.RemoveAll(s.GetTempLocal().BasePath); err != nil {
			err = errors.Wrapf(err, "failed to remove temp local")
			return
		}
	}

	return
}

type TemporaryFS struct {
	BasePath string
}

func (s TemporaryFS) DirTemp() (d string, err error) {
	return s.DirTempWithTemplate("")
}

func (s TemporaryFS) DirTempWithTemplate(
	template string,
) (d string, err error) {
	if d, err = os.MkdirTemp(s.BasePath, template); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s TemporaryFS) FileTemp() (f *os.File, err error) {
	if f, err = s.FileTempWithTemplate(""); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s TemporaryFS) FileTempWithTemplate(t string) (f *os.File, err error) {
	if f, err = os.CreateTemp(s.BasePath, t); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
