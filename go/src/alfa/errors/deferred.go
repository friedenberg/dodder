package errors

import (
	"io"
	"os"
)

func DeferredFlusher(
	err *error,
	f Flusher,
) {
	if err1 := f.Flush(); err1 != nil {
		if err == nil {
			panic(err)
		} else {
			*err = Join(*err, err1)
		}
	}
}

func DeferredYieldCloser[T any](
	yield func(T, error) bool,
	c io.Closer,
) {
	if err := c.Close(); err != nil {
		var t T
		yield(t, err)
	}
}

func DeferredCloser(
	err *error,
	c io.Closer,
) {
	if err1 := c.Close(); err1 != nil {
		if err == nil {
			panic(err)
		} else {
			*err = Join(*err, err1)
		}
	}
}

func DeferredCloseAndRename(err *error, c io.Closer, oldpath, newpath string) {
	if err == nil {
		panic("deferred error interface is nil")
	}

	if err1 := c.Close(); err1 != nil {
		*err = Join(*err, err1)
		return
	}

	if err1 := os.Rename(oldpath, newpath); err1 != nil {
		*err = Join(*err, err1)
	}
}

func Deferred(
	err *error,
	ef func() error,
) {
	if err1 := ef(); err1 != nil {
		if err == nil {
			panic(err)
		} else {
			*err = Join(*err, err1)
		}
	}
}

func DeferredChanError(
	err *error,
	ch <-chan error,
) {
	var err1 error

	select {
	case err1 = <-ch:
	}

	if err1 != nil {
		*err = Join(*err, err1)
	}
}

func DeferredChan(
	ch chan<- error,
	f func() error,
) {
	if err := f(); err != nil {
		ch <- err
	}
}
