package pool_value

import (
	"sync"
)

type poolSlice[SWIMMER any, SWIMMER_SLICE ~[]SWIMMER] struct {
	inner *sync.Pool
}

func MakeSlice[SWIMMER any, SWIMMER_SLICE ~[]SWIMMER]() poolSlice[SWIMMER, SWIMMER_SLICE] {
	return poolSlice[SWIMMER, SWIMMER_SLICE]{
		inner: &sync.Pool{
			New: func() any {
				swimmer := make(SWIMMER_SLICE, 0)
				return swimmer
			},
		},
	}
}

func (pool poolSlice[_, SWIMMER_SLICE]) Get() SWIMMER_SLICE {
	return pool.inner.Get().(SWIMMER_SLICE)
}

func (pool poolSlice[_, SWIMMER_SLICE]) Put(
	swimmer SWIMMER_SLICE,
) (err error) {
	swimmer = swimmer[:0]
	pool.inner.Put(swimmer)
	return err
}
