package inventory_list_blobs

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type V0 struct {
	object_inventory_format.Format
	object_inventory_format.Options
}

func MakeV0(
	format object_inventory_format.Format,
	options object_inventory_format.Options,
) V0 {
	return V0{
		Format:  format,
		Options: options,
	}
}

func (v V0) GetListFormat() sku.ListFormat {
	return v
}

func (v V0) GetType() ids.Type {
	return ids.MustType(ids.TypeInventoryListV0)
}

func (format V0) WriteObjectToOpenList(
	object *sku.Transacted,
	list *sku.OpenList,
) (n int64, err error) {
	err = errors.Err405MethodNotAllowed
	return
}

func (s V0) WriteInventoryListObject(
	o *sku.Transacted,
	w *bufio.Writer,
) (n int64, err error) {
	if n, err = s.Format.FormatPersistentMetadata(
		w,
		o,
		object_inventory_format.Options{Tai: true},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s V0) WriteInventoryListBlob(
	o sku.Collection,
	w *bufio.Writer,
) (n int64, err error) {
	var n1 int64

	fo := makePrinter(
		w,
		s.Format,
		s.Options,
	)

	for sk := range o.All() {
		if sk.Metadata.GetSha().IsNull() {
			err = errors.ErrorWithStackf("empty sha: %s", sk)
			return
		}

		n1, err = fo.Print(sk)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (s V0) ReadInventoryListObject(
	r *bufio.Reader,
) (n int64, o *sku.Transacted, err error) {
	o = sku.GetTransactedPool().Get()

	if n, err = s.Format.ParsePersistentMetadata(
		catgut.MakeRingBuffer(r, 0),
		o,
		s.Options,
	); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

type V0StreamCoder struct {
	V0
}

func (coder V0StreamCoder) DecodeFrom(
	output interfaces.FuncIter[*sku.Transacted],
	reader *bufio.Reader,
) (n int64, err error) {
	dec := makeScanner(
		reader,
		coder.Format,
		coder.Options,
	)

	for dec.Scan() {
		sk := dec.GetTransacted()

		if err = output(sk); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sk)
			return
		}
	}

	if err = dec.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s V0) AllInventoryListBlobSkus(
	reader *bufio.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return interfaces.MakeSeqErrorWithError[*sku.Transacted](
		errors.Err405MethodNotAllowed,
	)
}

func (s V0) StreamInventoryListBlobSkus(
	r1 *bufio.Reader,
	f interfaces.FuncIter[*sku.Transacted],
) (err error) {
	dec := makeScanner(
		r1,
		s.Format,
		s.Options,
	)

	for dec.Scan() {
		sk := dec.GetTransacted()

		if err = f(sk); err != nil {
			err = errors.Wrapf(err, "Sku: %s", sk)
			return
		}
	}

	if err = dec.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type V0ObjectCoder struct {
	V0
}

func (s V0ObjectCoder) EncodeTo(
	o *sku.Transacted,
	w *bufio.Writer,
) (n int64, err error) {
	if n, err = s.Format.FormatPersistentMetadata(
		w,
		o,
		object_inventory_format.Options{Tai: true},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s V0ObjectCoder) DecodeFrom(
	o *sku.Transacted,
	r *bufio.Reader,
) (n int64, err error) {
	if n, err = s.Format.ParsePersistentMetadata(
		catgut.MakeRingBuffer(r, 0),
		o,
		s.Options,
	); err != nil {
		if errors.IsEOF(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

type V0IterDecoder struct {
	V0
}

func (coder V0IterDecoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	reader *bufio.Reader,
) (n int64, err error) {
	dec := makeScanner(
		reader,
		coder.Format,
		coder.Options,
	)

	for dec.Scan() {
		sk := dec.GetTransacted()

		if !yield(sk) {
			return
		}
	}

	if err = dec.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
