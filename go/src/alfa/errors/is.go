package errors

import (
	"errors"
	"io"
	"net"
	"os"
	"slices"
	"syscall"
)

var (
	As     = errors.As
	Unwrap = errors.Unwrap
)

type FuncIs func(error) bool

func IsAny(err error, ises ...FuncIs) bool {
	for _, is := range ises {
		if is(err) {
			return true
		}
	}

	return false
}

func Is(err, target error) bool {
	if err == target {
		return true
	}

	if errors.Is(err, target) {
		return true
	}

	switch u := err.(type) {
	case interface{ Unwrap() error }:
		if Is(u.Unwrap(), target) {
			return true
		}

	case interface{ Unwrap() []error }:

		for _, e := range u.Unwrap() {
			if Is(e, target) {
				return true
			}
		}

	default:
		if errors.Is(u, target) {
			return true
		}
	}

	return false
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

func IsNotNilAndNotEOF(err error) bool {
	if err == nil || err == io.EOF {
		return false
	}

	return true
}

func IsEOF(err error) bool {
	if err == nil {
		return false
	}

	return Is(err, io.EOF)
}

func IsExist(err error) bool {
	e := errors.Unwrap(err)
	return os.IsExist(e)
}

func IsNotExist(err error) bool {
	e := errors.Unwrap(err)
	return os.IsNotExist(e)
}

func IsAsNilOrWrapf(
	err error,
	target error,
	format string,
	values ...any,
) (out error) {
	if Is(err, target) {
		out = nil
	} else {
		out = Wrapf(err, format, values...)
	}

	return
}
