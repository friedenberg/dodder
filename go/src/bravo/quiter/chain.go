package quiter

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Chain[T any](e T, wfs ...interfaces.FuncIter[T]) (err error) {
	for _, w := range wfs {
		if w == nil {
			continue
		}

		err = w(e)

		switch {
		case err == nil:
			continue

		case errors.IsStopIteration(err):
			err = nil
			return

		default:
			return
		}
	}

	return
}

func MakeChainDebug[T any](wfs ...interfaces.FuncIter[T]) interfaces.FuncIter[T] {
	for i := range wfs {
		old := wfs[i]
		wfs[i] = func(e T) (err error) {
			if err = old(e); err != nil {
				panic(err)
			}

			return
		}
	}

	return MakeChain(wfs...)
}

func MakeChain[T any](wfs ...interfaces.FuncIter[T]) interfaces.FuncIter[T] {
	return func(e T) (err error) {
		for _, w := range wfs {
			if w == nil {
				continue
			}

			err = w(e)

			switch {
			case err == nil:
				continue

			case errors.IsStopIteration(err):
				err = nil
				return

			default:
				return
			}
		}

		return
	}
}

func Multiplex[T any](
	e interfaces.FuncIter[T],
	producers ...func(interfaces.FuncIter[T]) error,
) (err error) {
	ch := make(chan error, len(producers))
	wg := &sync.WaitGroup{}
	wg.Add(len(producers))

	for _, p := range producers {
		go func(p func(interfaces.FuncIter[T]) error, ch chan<- error) {
			var err error

			defer func() {
				ch <- err
				wg.Done()
			}()

			if err = p(e); err != nil {
				err = errors.Wrap(err)
				return
			}
		}(p, ch)
	}

	wg.Wait()
	close(ch)

	for e := range ch {
		if e != nil {
			err = errors.MakeMulti(err, e)
		}
	}

	return
}
