package errors

import (
	"errors"
	"sync"
)

type GroupBuilder struct {
	lock sync.Mutex
	Group
}

func MakeGroupBuilder(errs ...error) (em *GroupBuilder) {
	em = &GroupBuilder{
		Group: make([]error, 0, len(errs)),
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

	if len(groupBuilder.Group) > 0 {
		return groupBuilder
	}

	return nil
}

func (groupBuilder *GroupBuilder) Reset() {
	groupBuilder.Group = groupBuilder.Group[:0]
}

func (groupBuilder *GroupBuilder) Len() int {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	return len(groupBuilder.Group)
}

func (groupBuilder *GroupBuilder) Empty() (ok bool) {
	ok = groupBuilder.Len() == 0
	return
}

func (groupBuilder *GroupBuilder) merge(err *GroupBuilder) {
	groupBuilder.lock.Lock()
	defer groupBuilder.lock.Unlock()

	groupBuilder.Group = append(groupBuilder.Group, err.Group...)
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
		groupBuilder.Group = append(groupBuilder.Group, err)
		groupBuilder.lock.Unlock()
	}
}
