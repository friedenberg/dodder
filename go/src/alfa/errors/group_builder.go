package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type GroupBuilder struct {
	lock sync.Mutex
	group
}

func MakeGroupBuilder(errs ...error) (em *GroupBuilder) {
	em = &GroupBuilder{
		group: make([]error, 0, len(errs)),
	}

	for _, err := range errs {
		if err != nil {
			em.Add(err)
		}
	}

	return
}

func (groupBuilder *GroupBuilder) GetError() error {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	if len(groupBuilder.group) > 0 {
		return groupBuilder
	}

	return nil
}

func (groupBuilder *GroupBuilder) Reset() {
	groupBuilder.group = groupBuilder.group[:0]
}

func (groupBuilder *GroupBuilder) Len() int {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	return len(groupBuilder.group)
}

func (groupBuilder *GroupBuilder) Empty() (ok bool) {
	ok = groupBuilder.Len() == 0
	return
}

func (groupBuilder *GroupBuilder) merge(err *GroupBuilder) {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	groupBuilder.group = append(groupBuilder.group, err.group...)
}

func (groupBuilder *GroupBuilder) Add(err error) {
	if err == nil {
		return
	}

	if groupBuilder == nil {
		panic("trying to add to nil multi error")
	}

	switch e1 := errors.Unwrap(err).(type) {
	case *GroupBuilder:
		groupBuilder.merge(e1)

	default:
		groupBuilder.lock.Lock()
		groupBuilder.group = append(groupBuilder.group, err)
		groupBuilder.lock.Unlock()
	}
}

func (groupBuilder *GroupBuilder) Unwrap() []error {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	out := make([]error, len(groupBuilder.group))
	copy(out, groupBuilder.group)

	return out
}

func (groupBuilder *GroupBuilder) Errors() (out []error) {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	out = make([]error, len(groupBuilder.group))
	copy(out, groupBuilder.group)

	return
}

func (groupBuilder *GroupBuilder) Error() string {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	switch len(groupBuilder.group) {
	case 0:
		return "no errors!"

	case 1:
		return groupBuilder.group[0].Error()

	default:
		sb := &strings.Builder{}

		fmt.Fprintf(sb, "# %d Errors", len(groupBuilder.group))
		sb.WriteString("\n")

		for i, err := range groupBuilder.group {
			fmt.Fprintf(sb, "Error %d:\n", i+1)
			sb.WriteString(err.Error())
			sb.WriteString("\n")
		}

		return sb.String()
	}
}
