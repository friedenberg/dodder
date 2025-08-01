package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Multi interface {
	error
	Add(error)
	Empty() bool
	Reset()
	GetMultiError() Multi
	GetError() error
	Errors() []error
	interfaces.Lenner
}

// split into error group builder and error group
type multi struct {
	lock    sync.Locker
	chOnErr chan struct{}
	slice   []error
}

func MakeMulti(errs ...error) (em *multi) {
	em = &multi{
		lock:    &sync.Mutex{},
		chOnErr: make(chan struct{}),
		slice:   make([]error, 0, len(errs)),
	}

	for _, err := range errs {
		if err != nil {
			em.Add(err)
		}
	}

	return
}

func (err *multi) ChanOnErr() <-chan struct{} {
	return err.chOnErr
}

func (err *multi) GetError() error {
	err.lock.Lock()
	defer err.lock.Unlock()

	if len(err.slice) > 0 {
		return err
	}

	return nil
}

func (err *multi) GetMultiError() Multi {
	return err
}

func (err *multi) Reset() {
	err.slice = err.slice[:0]
}

func (err *multi) Len() int {
	err.lock.Lock()
	defer err.lock.Unlock()

	return len(err.slice)
}

func (err *multi) Empty() (ok bool) {
	ok = err.Len() == 0
	return
}

func (e *multi) merge(err *multi) {
	e.lock.Lock()

	l := len(e.slice)

	e.slice = append(e.slice, err.slice...)

	if len(e.slice) > l && l == 0 {
		close(e.chOnErr)
	}

	e.lock.Unlock()
}

func (e *multi) Add(err error) {
	if err == nil {
		return
	}

	if e == nil {
		panic("trying to add to nil multi error")
	}

	switch e1 := errors.Unwrap(err).(type) {
	case *multi:
		e.merge(e1)

	default:
		e.lock.Lock()

		l := len(e.slice)

		e.slice = append(e.slice, err)

		if len(e.slice) > l && l == 0 {
			close(e.chOnErr)
		}

		e.lock.Unlock()
	}
}

func (err *multi) Unwrap() []error {
	err.lock.Lock()
	defer err.lock.Unlock()

	out := make([]error, len(err.slice))
	copy(out, err.slice)

	return out
}

func (err *multi) Errors() (out []error) {
	err.lock.Lock()
	defer err.lock.Unlock()

	out = make([]error, len(err.slice))
	copy(out, err.slice)

	return
}

func (err *multi) Error() string {
	err.lock.Lock()
	defer err.lock.Unlock()

	switch len(err.slice) {
	case 0:
		return "no errors!"

	case 1:
		return err.slice[0].Error()

	default:
		sb := &strings.Builder{}

		fmt.Fprintf(sb, "# %d Errors", len(err.slice))
		sb.WriteString("\n")

		for i, err := range err.slice {
			fmt.Fprintf(sb, "Error %d:\n", i+1)
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}

		return sb.String()
	}
}
