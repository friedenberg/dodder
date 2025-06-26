package pool

import (
	"sync"
)

type poolValue[T any] struct {
	reset func(T)
	inner *sync.Pool
}

func MakeValue[T any](
	construct func() T,
	reset func(T),
) poolValue[T] {
	return poolValue[T]{
		reset: reset,
		inner: &sync.Pool{
			New: func() interface{} {
				o := construct()

				return o
			},
		},
	}
}

func (ip poolValue[T]) Get() T {
	return ip.inner.Get().(T)
}

func (ip poolValue[T]) Put(i T) (err error) {
	ip.reset(i)
	ip.inner.Put(i)

	return
}
