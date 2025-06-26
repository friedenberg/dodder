package quiter

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeSyncSerializer[T any](
	wf interfaces.FuncIter[T],
) interfaces.FuncIter[T] {
	l := &sync.Mutex{}

	return func(e T) (err error) {
		l.Lock()
		defer l.Unlock()

		if err = wf(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}
