package pool_value

import (
	"sync"
)

type poolValue[SWIMMER any] struct {
	reset func(SWIMMER)
	inner *sync.Pool
}

func Make[SWIMMER any](
	construct func() SWIMMER,
	reset func(SWIMMER),
) poolValue[SWIMMER] {
	return poolValue[SWIMMER]{
		reset: reset,
		inner: &sync.Pool{
			New: func() any {
				swimmer := construct()
				return swimmer
			},
		},
	}
}

func (pool poolValue[SWIMMER]) Get() SWIMMER {
	return pool.inner.Get().(SWIMMER)
}

func (pool poolValue[SWIMMER]) Put(swimmer SWIMMER) (err error) {
	pool.reset(swimmer)
	pool.inner.Put(swimmer)

	return
}
