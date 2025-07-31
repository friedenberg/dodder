package inventory_list_coders

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type funcListFormatConstructor func(
	env_repo.Env,
	*box_format.BoxTransacted,
) sku.ListFormat

var coderConstructors = map[string]funcListFormatConstructor{
	ids.TypeInventoryListV1: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		return doddishV1{
			Box: box,
		}
	},
	ids.TypeInventoryListV2: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		return doddishV2{
			Box:                    box,
			ImmutableConfigPrivate: envRepo.GetConfigPrivate().Blob,
		}
	},
	ids.TypeInventoryListJsonV0: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		return jsonV0{
			ImmutableConfigPrivate: envRepo.GetConfigPrivate().Blob,
		}
	},
}

var (
	_ sku.ListFormat = doddishV1{}
	_ sku.ListFormat = doddishV2{}
	_ sku.ListFormat = jsonV0{}
)

func WriteObjectToOpenList(
	format sku.ListFormat,
	object *sku.Transacted,
	list *sku.OpenList,
) (n int64, err error) {
	if !list.LastTai.Less(object.GetTai()) {
		err = errors.Errorf(
			"object order incorrect. Last: %s, current: %s",
			list.LastTai,
			object.GetTai(),
		)

		return
	}

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(list.Mover)
	defer repoolBufferedWriter()

	if n, err = format.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	list.LastTai = object.GetTai()
	list.Len += 1

	return
}

func WriteInventoryList(
	ctx interfaces.ActiveContext,
	format sku.ListFormat,
	skus interfaces.SeqError[*sku.Transacted],
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var n1 int64

	var object *sku.Transacted

	for object, err = range skus {
		errors.ContextContinueOrPanic(ctx)

		if err != nil {
			err = errors.Wrap(err)
			return
		}

		n1, err = format.EncodeTo(object, bufferedWriter)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

// TODO also return a repool func
func CollectSkuList(
	ctx interfaces.ActiveContext,
	listFormat sku.ListFormat,
	reader *bufio.Reader,
	list *sku.List,
) (err error) {
	iter := streamInventoryList(ctx, listFormat, reader)

	for sk, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if err = list.Add(sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func streamInventoryList(
	ctx interfaces.ActiveContext,
	format sku.ListFormat,
	bufferedReader *bufio.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		for {
			errors.ContextContinueOrPanic(ctx)

			object := sku.GetTransactedPool().Get()
			// TODO Fix upstream issues with repooling
			// defer sku.GetTransactedPool().Put(object)

			if _, err := format.DecodeFrom(
				object,
				bufferedReader,
			); err != nil {
				if errors.IsEOF(err) {
					err = nil
					break
				} else {
					if !yield(nil, err) {
						break
					}
				}
			}

			if !yield(object, nil) {
				break
			}
		}
	}
}

func writeInventoryListObject(
	format sku.ListFormat,
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if n, err = format.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type SeqCoder struct {
	ctx interfaces.ActiveContext
	sku.ListFormat
}

func (coder SeqCoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		errors.ContextContinueOrPanic(coder.ctx)

		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.ListFormat.DecodeFrom(object, bufferedReader); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return
			}
		}

		if !yield(object) {
			return
		}
	}

	return
}

type SeqErrorDecoder struct {
	ctx interfaces.ActiveContext
	sku.ListFormat
}

func (coder SeqErrorDecoder) DecodeFrom(
	yield func(*sku.Transacted, error) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		errors.ContextContinueOrPanic(coder.ctx)

		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.ListFormat.DecodeFrom(object, bufferedReader); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)

				if !yield(nil, err) {
					return
				}
			}
		}

		if !yield(object, nil) {
			return
		}
	}

	return
}

type SeqErrorEncoder struct {
	ctx interfaces.ActiveContext
	sku.ListFormat
}

func (coder SeqErrorDecoder) EncodeTo(
	seq sku.Seq,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	for object, iterErr := range seq {
		errors.ContextContinueOrPanic(coder.ctx)

		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if _, err = coder.ListFormat.EncodeTo(object, bufferedWriter); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return
			}
		}
	}

	return
}
