package errors

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"syscall"
	"time"
)

var (
	As     = errors.As
	Unwrap = errors.Unwrap
)

func IsAny(err error, funcTargets ...FuncIs) bool {
	for _, funcTarget := range funcTargets {
		if funcTarget(err) {
			return true
		}
	}

	return false
}

func IsSentinel(err, target error) bool {
	return err == target
}

func Is(err, target error) bool {
	if err == io.EOF || target == io.EOF {
		panic("checking for EOF via errors.Is")
	}

	if DebugBuild {
		return IsWithTimeout(err, target, time.Second)
	}

	return errors.Is(err, target)
}

func IsWithTimeout(err, target error, timeout time.Duration) bool {
	chBool := make(chan bool)
	defer close(chBool)

	go func() {
		chBool <- errors.Is(err, target)
	}()

	select {
	case result := <-chBool:
		return result

	case <-time.After(timeout):
		panic(fmt.Sprintf("calling errors.Is() timed out for %T", err))
	}
}

func IsNetTimeout(err error) (ok bool) {
	var netError net.Error

	if !As(err, &netError) {
		return
	}

	ok = netError.Timeout()

	return
}

func MakeIsErrno(targets ...syscall.Errno) FuncIs {
	return func(err error) bool {
		return IsErrno(err, targets...)
	}
}

func IsErrno(err error, targets ...syscall.Errno) (ok bool) {
	var errno syscall.Errno

	if !As(err, &errno) {
		return false
	}

	return slices.Contains(targets, errno)
}

func IsBrokenPipe(err error) bool {
	return IsErrno(err, syscall.EPIPE)
}

func IsTooManyOpenFiles(err error) bool {
	e := errors.Unwrap(err)
	return e.Error() == "too many open files"
}

// TODO remove
func IsNotNilAndNotEOF(err error) bool {
	if err == nil || err == io.EOF {
		return false
	}

	return true
}

func IsEOF(err error) bool {
	return err == io.EOF
}

func IsExist(err error) bool {
	return os.IsExist(errors.Unwrap(err))
}

func IsNotExist(err error) bool {
	return os.IsNotExist(errors.Unwrap(err))
}
