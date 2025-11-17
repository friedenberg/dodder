package pool

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type pool[SWIMMER any, SWIMMER_PTR interfaces.Ptr[SWIMMER]] struct {
	inner *sync.Pool
	reset func(SWIMMER_PTR)
}

var _ interfaces.Pool[string, *string] = pool[string, *string]{}

func MakeWithResetable[SWIMMER any, SWIMMER_PTR interfaces.ResetablePtr[SWIMMER]]() *pool[SWIMMER, SWIMMER_PTR] {
	return Make(nil, func(swimmer SWIMMER_PTR) {
		swimmer.Reset()
	})
}

func Make[SWIMMER any, SWIMMER_PTR interfaces.Ptr[SWIMMER]](
	New func() SWIMMER_PTR,
	Reset func(SWIMMER_PTR),
) *pool[SWIMMER, SWIMMER_PTR] {
	return &pool[SWIMMER, SWIMMER_PTR]{
		reset: Reset,
		inner: &sync.Pool{
			New: func() (swimmer any) {
				if New == nil {
					swimmer = new(SWIMMER)
				} else {
					swimmer = New()
				}

				return swimmer
			},
		},
	}
}

func (pool pool[SWIMMER, SWIMMER_PTR]) Get() SWIMMER_PTR {
	return pool.inner.Get().(SWIMMER_PTR)
}

func (pool pool[SWIMMER, SWIMMER_PTR]) GetWithRepool() (SWIMMER_PTR, interfaces.FuncRepool) {
	element := pool.Get()

	return element, func() {
		pool.Put(element)
	}
}

func (pool pool[SWIMMER, SWIMMER_PTR]) Put(swimmer SWIMMER_PTR) (err error) {
	if swimmer == nil {
		return err
	}

	if pool.reset != nil {
		pool.reset(swimmer)
	}

	pool.inner.Put(swimmer)

	return err
}
