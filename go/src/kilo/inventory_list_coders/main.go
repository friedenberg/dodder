package inventory_list_coders

import (
	"bufio"

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
		return DoddishV1{
			Box: box,
		}
	},
	ids.TypeInventoryListV2: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		return DoddishV2{
			Box:                    box,
			ImmutableConfigPrivate: envRepo.GetConfigPrivate().Blob,
		}
	},
}

var (
	_ sku.ListFormat = DoddishV1{}
	_ sku.ListFormat = DoddishV2{}
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
	format sku.ListFormat,
	skus interfaces.SeqError[*sku.Transacted],
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var n1 int64

	var object *sku.Transacted

	for object, err = range skus {
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
	listFormat sku.ListFormat,
	reader *bufio.Reader,
	list *sku.List,
) (err error) {
	iter := StreamInventoryList(listFormat, reader)

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

func StreamInventoryList(
	format sku.ListFormat,
	bufferedReader *bufio.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		for {
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

func WriteInventoryListObject(
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

type IterCoder struct {
	sku.ListFormat
}

func (coder IterCoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.ListFormat.DecodeFrom(object, bufferedReader); err != nil {
			if errors.IsEOF(err) {
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
