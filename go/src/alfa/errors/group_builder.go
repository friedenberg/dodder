package errors

import (
	"errors"
	"sync"
)

type GroupBuilder struct {
	lock  sync.Mutex
	group Group
}

func MakeGroupBuilder(errs ...error) (groupBuilder *GroupBuilder) {
	groupBuilder = &GroupBuilder{
		group: make([]error, 0, len(errs)),
	}

	for _, err := range errs {
		if err != nil {
			groupBuilder.Add(err)
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
