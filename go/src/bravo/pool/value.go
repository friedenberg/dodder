package pool

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type value[SWIMMER any] struct {
	inner *sync.Pool
	reset func(SWIMMER)
}

func MakeValue[SWIMMER any](
	New func() SWIMMER,
	Reset func(SWIMMER),
) *value[SWIMMER] {
	return &value[SWIMMER]{
		reset: Reset,
		inner: &sync.Pool{
			New: func() (swimmer any) {
				if New == nil {
					var element SWIMMER
					swimmer = element
				} else {
					swimmer = New()
				}

				return
			},
		},
	}
}

func (pool value[SWIMMER]) Get() SWIMMER {
	return pool.inner.Get().(SWIMMER)
}

func (pool value[SWIMMER]) PutMany(
	swimmers ...SWIMMER,
) (err error) {
	for _, i := range swimmers {
		if err = pool.Put(i); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (pool value[SWIMMER]) Put(swimmer SWIMMER) (err error) {
	if pool.reset != nil {
		pool.reset(swimmer)
	}

	pool.inner.Put(swimmer)

	return
}
