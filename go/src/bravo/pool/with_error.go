package pool

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type poolWithError[SWIMMER any, SWIMMER_PTR interfaces.Ptr[SWIMMER]] struct {
	inner *sync.Pool
	reset func(SWIMMER_PTR)
}

var _ interfaces.PoolWithErrorsPtr[string, *string] = poolWithError[string, *string]{}

func MakeWithError[SWIMMER any, SWIMMER_PTR interfaces.Ptr[SWIMMER]](
	New func() (SWIMMER_PTR, error),
	Reset func(SWIMMER_PTR),
) *poolWithError[SWIMMER, SWIMMER_PTR] {
	return &poolWithError[SWIMMER, SWIMMER_PTR]{
		reset: Reset,
		inner: &sync.Pool{
			New: func() (swimmer any) {
				if New == nil {
					swimmer = new(SWIMMER)
				} else {
					var err error
					swimmer, err = New()
					if err != nil {
						panic(err)
					}
				}

				return swimmer
			},
		},
	}
}

func (pool poolWithError[SWIMMER, SWIMMER_PTR]) Get() (e SWIMMER_PTR, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch rt := r.(type) {
			case error:
				err = rt

			default:
				err = errors.ErrorWithStackf("panicked during pool new: %w", err)
			}
		}
	}()

	return pool.inner.Get().(SWIMMER_PTR), nil
}

func (pool poolWithError[SWIMMER, SWIMMER_PTR]) PutMany(
	swimmers ...SWIMMER_PTR,
) (err error) {
	for _, swimmer := range swimmers {
		if err = pool.Put(swimmer); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (pool poolWithError[SWIMMER, SWIMMER_PTR]) Put(
	swimmer SWIMMER_PTR,
) (err error) {
	if swimmer == nil {
		return err
	}

	if pool.reset != nil {
		pool.reset(swimmer)
	}

	pool.inner.Put(swimmer)

	return err
}
