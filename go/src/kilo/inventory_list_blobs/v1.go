package inventory_list_blobs

import (
	"bufio"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type V1 struct {
	V1ObjectCoder
}

func (format V1) GetListFormat() sku.ListFormat {
	return format
}

func (format V1) GetType() ids.Type {
	return ids.MustType(ids.TypeInventoryListV1)
}

func (format V1) WriteObjectToOpenList(
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

	bufferedWriter := ohio.BufferedWriter(list.Mover)
	defer pool.GetBufioWriter().Put(bufferedWriter)

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

func (format V1) WriteInventoryListBlob(
	skus sku.Collection,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var n1 int64

	for sk := range skus.All() {
		n1, err = format.EncodeTo(sk, bufferedWriter)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (format V1) WriteInventoryListObject(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var n1 int64
	var n2 int

	n1, err = format.Box.EncodeStringTo(object, bufferedWriter)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = fmt.Fprintf(bufferedWriter, "\n")
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format V1) ReadInventoryListObject(
	reader *bufio.Reader,
) (n int64, object *sku.Transacted, err error) {
	object = sku.GetTransactedPool().Get()

	if n, err = format.DecodeFrom(object, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type V1StreamCoder struct {
	V1
}

func (coder V1StreamCoder) DecodeFrom(
	output interfaces.FuncIter[*sku.Transacted],
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		defer sku.GetTransactedPool().Put(object)

		if _, err = coder.V1ObjectCoder.DecodeFrom(object, bufferedReader); err != nil {
			if errors.IsEOF(err) {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return
			}
		}

		if err = output(object); err != nil {
			err = errors.Wrapf(err, "Object: %s", sku.String(object))
			return
		}
	}

	return
}

func (format V1) StreamInventoryListBlobSkus(
	bufferedReader *bufio.Reader,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = format.V1ObjectCoder.DecodeFrom(
			object,
			bufferedReader,
		); err != nil {
			if errors.IsEOF(err) {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return
			}
		}

		if err = output(object); err != nil {
			err = errors.Wrapf(err, "Object: %s", sku.String(object))
			return
		}
	}

	return
}

type V1ObjectCoder struct {
	Box *box_format.BoxTransacted
}

func (coder V1ObjectCoder) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.Sha().IsNull() {
		err = errors.ErrorWithStackf("empty sha: %q", sku.String(object))
		return
	}

	var n1 int64
	var n2 int

	n1, err = coder.Box.EncodeStringTo(object, bufferedWriter)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = fmt.Fprintf(bufferedWriter, "\n")
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coder V1ObjectCoder) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var isEOF bool

	if n, err = coder.Box.ReadStringFormat(object, bufferedReader); err != nil {
		if err == io.EOF {
			isEOF = true

			if n == 0 {
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	if err = object.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if isEOF {
		err = io.EOF
	}

	return
}

type V1IterDecoder struct {
	V1
}

func (coder V1IterDecoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.V1ObjectCoder.DecodeFrom(object, bufferedReader); err != nil {
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
