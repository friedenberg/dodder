package file_lock

import (
	"io/fs"
	"os"
	"sync"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

type Lock struct {
	envUI       env_ui.Env
	path        string
	description string
	mutex       sync.Mutex
	file        *os.File
}

// TODO switch to using context
func New(
	envUI env_ui.Env,
	path string,
	description string,
) (l *Lock) {
	return &Lock{
		envUI:       envUI,
		path:        path,
		description: description,
	}
}

func (l *Lock) Path() string {
	return l.path
}

func (l *Lock) IsAcquired() (acquired bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	acquired = l.file != nil

	return
}

func (lock *Lock) Lock() (err error) {
	if !lock.mutex.TryLock() {
		err = errors.ErrorWithStackf("attempting concurrent locks")
		return
	}

	defer lock.mutex.Unlock()

	if lock.file != nil {
		err = errors.ErrorWithStackf("already locked")
		return
	}

	createLock := func(path string) (*os.File, error) {
		return files.TryOrTimeout(
			path,
			time.Second,
			func(path string) (*os.File, error) {
				return files.OpenFile(
					path,
					os.O_RDONLY|os.O_EXCL|os.O_CREATE,
					0o755,
				)
			},
			"acquiring lock",
		)
	}

	if lock.file, err = files.TryOrMakeDirIfNecessary(
		lock.Path(),
		createLock,
	); err != nil {
		if errors.Is(err, fs.ErrExist) {
			err = ErrUnableToAcquireLock{
				envUI:       lock.envUI,
				Path:        lock.Path(),
				description: lock.description,
			}
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (lock *Lock) Unlock() (err error) {
	if !lock.mutex.TryLock() {
		err = errors.ErrorWithStackf("attempting concurrent locks")
		return
	}

	defer lock.mutex.Unlock()

	if err = lock.file.Close(); err != nil {
		err = errors.Wrapf(err, "File: %v", lock.file)
		return
	}

	lock.file = nil

	if err = os.Remove(lock.Path()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
