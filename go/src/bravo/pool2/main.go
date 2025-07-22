package pool2

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type pool[T any, TPtr interfaces.Ptr[T]] struct {
	inner *sync.Pool
	reset func(TPtr)
}

func MakePool[T any, TPtr interfaces.Ptr[T]](
	New func() (TPtr, error),
	Reset func(TPtr),
) *pool[T, TPtr] {
	return &pool[T, TPtr]{
		reset: Reset,
		inner: &sync.Pool{
			New: func() (o interface{}) {
				if New == nil {
					o = new(T)
				} else {
					var err error
					o, err = New()

					if err != nil {
						panic(err)
					}
				}

				return
			},
		},
	}
}

func (ip pool[T, TPtr]) GetPool() interfaces.PoolWithErrorsPtr[T, TPtr] {
	return ip
}

func (ip pool[T, TPtr]) Get() (e TPtr, err error) {
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

	return ip.inner.Get().(TPtr), nil
}

func (ip pool[T, TPtr]) PutMany(is ...TPtr) (err error) {
	for _, i := range is {
		if err = ip.Put(i); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (ip pool[T, TPtr]) Put(i TPtr) (err error) {
	if i == nil {
		return
	}

	if ip.reset != nil {
		ip.reset(i)
	}

	ip.inner.Put(i)

	return
}
