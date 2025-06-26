package files

import (
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func CreateExclusiveReadOnly(path string) (file *os.File, err error) {
	if file, err = os.OpenFile(
		path,
		os.O_RDONLY|os.O_CREATE|os.O_EXCL,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func CreateExclusiveWriteOnly(p string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		p,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func TryOrTimeout(
	path string,
	timeout time.Duration,
	apply func(string) (*os.File, error),
	explainApply string,
) (*os.File, error) {
	chSuccess := make(chan *os.File)
	chError := make(chan error)
	timer := time.NewTimer(timeout)

	go func() {
		<-timer.C
		chError <- errors.ErrorWithStackf("timeout while %s: %q", explainApply, path)
	}()

	go func() {
		file, err := apply(path)

		if err != nil {
			chError <- err
		} else {
			chSuccess <- file
		}
	}()

	defer func() {
		timer.Stop()
		close(chSuccess)
		close(chError)
	}()

	select {
	case file := <-chSuccess:
		return file, nil

	case err := <-chError:
		return nil, err
	}
}

func TryOrMakeDirIfNecessary(
	path string,
	apply func(string) (*os.File, error),
) (file *os.File, err error) {
	if file, err = apply(path); err != nil {
		if errors.IsNotExist(err) {
			dir := filepath.Dir(path)

			if err = os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
				err = errors.Wrap(err)
				return
			}

			return apply(path)
		}
	}

	return
}

func Create(s string) (f *os.File, err error) {
	if f, err = os.Create(s); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenFile(name string, flag int, perm os.FileMode) (f *os.File, err error) {
	if f, err = os.OpenFile(name, flag, perm); err != nil {
		err = errors.Wrapf(err, "Mode: %d, Perm: %d", flag, perm)
		return
	}

	return
}

func Open(s string) (f *os.File, err error) {
	if f, err = os.Open(s); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenReadWrite(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		s,
		os.O_RDWR|os.O_CREATE,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenCreate(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		s,
		os.O_RDWR|os.O_CREATE,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenCreateWriteOnlyTruncate(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		s,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenExclusiveWriteOnlyTruncate(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		s,
		os.O_WRONLY|os.O_EXCL|os.O_TRUNC,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenExclusiveWriteOnlyAppend(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(
		s,
		os.O_WRONLY|os.O_EXCL|os.O_APPEND,
		0o666,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func OpenReadOnly(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(s, os.O_RDONLY, 0o666); err != nil {
		err = errors.Wrapf(err, "Path: %q", s)
		return
	}

	return
}

func OpenExclusive(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(s, os.O_RDWR|os.O_EXCL, 0o666); err != nil {
		err = errors.Wrapf(err, "Path: %q", s)
		return
	}

	return
}

func OpenExclusiveReadOnly(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(s, os.O_RDONLY|os.O_EXCL, 0o666); err != nil {
		err = errors.Wrapf(err, "Path: %q", s)
		return
	}

	return
}

func OpenExclusiveWriteOnly(s string) (f *os.File, err error) {
	if f, err = os.OpenFile(s, os.O_WRONLY|os.O_EXCL, 0o666); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func Close(f *os.File) error {
	return f.Close()
}

func CombinedOutput(c *exec.Cmd) ([]byte, error) {
	return c.CombinedOutput()
}

func ReadAllString(s ...string) (o string, err error) {
	var f *os.File

	if f, err = Open(path.Join(s...)); err != nil {
		return
	}

	defer Close(f)

	var b []byte

	if b, err = io.ReadAll(f); err != nil {
		return
	}

	o = string(b)

	return
}
