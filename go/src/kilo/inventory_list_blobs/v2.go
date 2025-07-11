package inventory_list_blobs

import (
	"bufio"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type V2 struct {
	V2ObjectCoder
}

func (format V2) GetListFormat() sku.ListFormat {
	return format
}

func (format V2) GetType() ids.Type {
	return ids.MustType(ids.TypeInventoryListV2)
}

func (format V2) WriteObjectToOpenList(
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

	if object.Metadata.RepoSig.IsEmpty() {
		err = errors.Errorf("repo sig empty")
		return
	}

	if len(object.Metadata.RepoPubKey) == 0 {
		err = errors.Errorf("repo pubkey empty")
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

func (format V2) WriteInventoryListBlob(
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

func (format V2) WriteInventoryListObject(
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

func (format V2) ReadInventoryListObject(
	reader *bufio.Reader,
) (n int64, object *sku.Transacted, err error) {
	object = sku.GetTransactedPool().Get()

	if n, err = format.DecodeFrom(object, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type V2StreamCoder struct {
	V2
}

func (coder V2StreamCoder) DecodeFrom(
	output interfaces.FuncIter[*sku.Transacted],
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		defer sku.GetTransactedPool().Put(object)

		if _, err = coder.V2ObjectCoder.DecodeFrom(object, bufferedReader); err != nil {
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

func (format V2) StreamInventoryListBlobSkus(
	bufferedReader *bufio.Reader,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = format.V2ObjectCoder.DecodeFrom(
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

type V2ObjectCoder struct {
	Box                    *box_format.BoxTransacted
	ImmutableConfigPrivate genesis_configs.Private
}

func (coder V2ObjectCoder) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.Sha().IsNull() {
		err = errors.ErrorWithStackf("empty sha: %q", sku.String(object))
		return
	}

	if object.Metadata.RepoSig.IsEmpty() {
		err = errors.ErrorWithStackf("no repo signature")
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

func (coder V2ObjectCoder) DecodeFrom(
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

	sh := sha.Make(object.GetTai().GetShaLike())
	defer sha.GetPool().Put(sh)

	if len(object.Metadata.RepoPubKey) == 0 {
		err = errors.ErrorWithStackf(
			"RepoPubkey missing for %s. Fields: %#v",
			sku.String(object),
			object.Metadata.Fields,
		)
		return
	}

	if object.Metadata.RepoSig.IsEmpty() {
		err = errors.ErrorWithStackf(
			"signature missing for %s. Fields: %#v",
			sku.String(object),
			object.Metadata.Fields,
		)
		return
	}

	if err = repo_signing.VerifySignature(
		object.Metadata.RepoPubKey,
		sh.GetShaBytes(),
		object.Metadata.RepoSig,
	); err != nil {
		err = errors.Wrap(err)
		return
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

type V2IterDecoder struct {
	V2
}

func (coder V2IterDecoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.V2ObjectCoder.DecodeFrom(object, bufferedReader); err != nil {
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
