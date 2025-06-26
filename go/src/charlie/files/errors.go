package files

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyFileList = errors.New("empty file list")
	errNotDirectory  ErrNotDirectory
)

func IsErrNotDirectory(err error) bool {
	return errors.Is(errNotDirectory, err)
}

type ErrNotDirectory string

func (err ErrNotDirectory) Is(target error) bool {
	_, ok := target.(ErrNotDirectory)
	return ok
}

func (err ErrNotDirectory) Error() string {
	return fmt.Sprintf("%q is not a directory", string(err))
}
