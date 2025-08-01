package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type GroupBuilder interface {
	Add(error)
	Empty() bool
	Reset()
	GetError() error
	Errors() []error
	interfaces.Lenner
}

// split into error groupBuilder builder and error groupBuilder
type groupBuilder struct {
	lock    sync.Locker
	chOnErr chan struct{}
	group
}

func MakeGroupBuilder(errs ...error) (em *groupBuilder) {
	em = &groupBuilder{
		lock:    &sync.Mutex{},
		chOnErr: make(chan struct{}),
		group:   make([]error, 0, len(errs)),
	}

	for _, err := range errs {
		if err != nil {
			em.Add(err)
		}
	}

	return
}

func (group *groupBuilder) GetError() error {
	group.lock.Lock()
	defer group.lock.Unlock()

	if len(group.group) > 0 {
		return group
	}

	return nil
}

func (group *groupBuilder) Reset() {
	group.group = group.group[:0]
}

func (group *groupBuilder) Len() int {
	group.lock.Lock()
	defer group.lock.Unlock()

	return len(group.group)
}

func (group *groupBuilder) Empty() (ok bool) {
	ok = group.Len() == 0
	return
}

func (group *groupBuilder) merge(err *groupBuilder) {
	group.lock.Lock()

	l := len(group.group)

	group.group = append(group.group, err.group...)

	if len(group.group) > l && l == 0 {
		close(group.chOnErr)
	}

	group.lock.Unlock()
}

func (e *groupBuilder) Add(err error) {
	if err == nil {
		return
	}

	if e == nil {
		panic("trying to add to nil multi error")
	}

	switch e1 := errors.Unwrap(err).(type) {
	case *groupBuilder:
		e.merge(e1)

	default:
		e.lock.Lock()

		l := len(e.group)

		e.group = append(e.group, err)

		if len(e.group) > l && l == 0 {
			close(e.chOnErr)
		}

		e.lock.Unlock()
	}
}

func (group *groupBuilder) Unwrap() []error {
	group.lock.Lock()
	defer group.lock.Unlock()

	out := make([]error, len(group.group))
	copy(out, group.group)

	return out
}

func (group *groupBuilder) Errors() (out []error) {
	group.lock.Lock()
	defer group.lock.Unlock()

	out = make([]error, len(group.group))
	copy(out, group.group)

	return
}

func (group *groupBuilder) Error() string {
	group.lock.Lock()
	defer group.lock.Unlock()

	switch len(group.group) {
	case 0:
		return "no errors!"

	case 1:
		return group.group[0].Error()

	default:
		sb := &strings.Builder{}

		fmt.Fprintf(sb, "# %d Errors", len(group.group))
		sb.WriteString("\n")

		for i, err := range group.group {
			fmt.Fprintf(sb, "Error %d:\n", i+1)
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}

		return sb.String()
	}
}
