package format

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type writerTo[T any] struct {
	wf interfaces.FuncWriterElement[T]
	e  *T
}

func (wt *writerTo[T]) WriteTo(w io.Writer) (n int64, err error) {
	if n, err = wt.wf(w, wt.e); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeWriterTo2[T any](
	wf interfaces.FuncWriterElement[T],
	e *T,
) *writerTo[T] {
	return &writerTo[T]{
		wf: wf,
		e:  e,
	}
}

type writerToInterface[T any] struct {
	wf interfaces.FuncWriterElementInterface[T]
	e  T
}

func (wt writerToInterface[T]) WriteTo(w io.Writer) (n int64, err error) {
	if n, err = wt.wf(w, wt.e); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func MakeWriterToInterface[T any](
	wf interfaces.FuncWriterElementInterface[T],
	e T,
) writerToInterface[T] {
	return writerToInterface[T]{
		wf: wf,
		e:  e,
	}
}
