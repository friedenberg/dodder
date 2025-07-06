package typed_blob_store

import (
	"bufio"
	"io"
	"iter"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_blobs"
)

type InventoryList struct {
	envRepo      env_repo.Env
	objectFormat object_inventory_format.Format
	boxFormat    *box_format.BoxTransacted
	v0           inventory_list_blobs.V0
	v1           inventory_list_blobs.V1
	v2           inventory_list_blobs.V2

	objectCoders   triple_hyphen_io.CoderTypeMapWithoutType[*sku.Transacted]
	streamDecoders map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]
}

func MakeInventoryStore(
	dirLayout env_repo.Env,
	objectFormat object_inventory_format.Format,
	boxFormat *box_format.BoxTransacted,
) InventoryList {
	objectOptions := object_inventory_format.Options{Tai: true}

	s := InventoryList{
		envRepo:      dirLayout,
		objectFormat: objectFormat,
		boxFormat:    boxFormat,
		v0: inventory_list_blobs.MakeV0(
			objectFormat,
			objectOptions,
		),
		v1: inventory_list_blobs.V1{
			V1ObjectCoder: inventory_list_blobs.V1ObjectCoder{
				Box: boxFormat,
			},
		},
		v2: inventory_list_blobs.V2{
			V2ObjectCoder: inventory_list_blobs.V2ObjectCoder{
				Box:                    boxFormat,
				ImmutableConfigPrivate: dirLayout.GetConfigPrivate().ImmutableConfig,
			},
		},
	}

	s.objectCoders = triple_hyphen_io.CoderTypeMapWithoutType[*sku.Transacted](
		map[string]interfaces.CoderBufferedReadWriter[*sku.Transacted]{
			"": inventory_list_blobs.V0ObjectCoder{
				V0: s.v0,
			},
			builtin_types.InventoryListTypeV1: s.v1.V1ObjectCoder,
			builtin_types.InventoryListTypeV2: s.v2.V2ObjectCoder,
		},
	)

	s.streamDecoders = map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]{
		"": inventory_list_blobs.V0IterDecoder{
			V0: s.v0,
		},
		builtin_types.InventoryListTypeV1: inventory_list_blobs.V1IterDecoder{
			V1: s.v1,
		},
		builtin_types.InventoryListTypeV2: inventory_list_blobs.V2IterDecoder{
			V2: s.v2,
		},
	}

	return s
}

func (a InventoryList) GetCommonStore() sku.BlobStore[*sku.List] {
	return a
}

func (a InventoryList) GetObjectFormat() object_inventory_format.Format {
	return a.objectFormat
}

func (a InventoryList) GetBoxFormat() *box_format.BoxTransacted {
	return a.boxFormat
}

func (a InventoryList) GetTransactedWithBlob(
	inventoryList sku.TransactedGetter,
) (objectAndBlob sku.TransactedWithBlob[*sku.List], n int64, err error) {
	objectAndBlob.Transacted = inventoryList.GetSku()
	blobSha := objectAndBlob.GetBlobSha()

	var readCloser interfaces.ShaReadCloser

	if readCloser, err = a.envRepo.BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader := ohio.BufferedReader(readCloser)
	defer pool.GetBufioReader().Put(bufferedReader)

	if n, err = a.GetTransactedWithBlobFromReader(
		&objectAndBlob,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a InventoryList) GetTransactedWithBlobFromReader(
	twb *sku.TransactedWithBlob[*sku.List],
	reader *bufio.Reader,
) (n int64, err error) {
	tipe := twb.GetType()
	twb.Blob = sku.MakeList()

	switch tipe.String() {
	case "", builtin_types.InventoryListTypeV0:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			a.v0,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV1:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			a.v1,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV2:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			a.v2,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (a InventoryList) WriteObjectToWriter(
	tipe ids.Type,
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	return a.objectCoders.EncodeTo(
		&triple_hyphen_io.TypedBlob[*sku.Transacted]{
			Type:   &tipe,
			Blob: object,
		},
		bufferedWriter,
	)
}

func (store InventoryList) WriteBlobToWriter(
	tipe ids.Type,
	list sku.Collection,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	switch tipe.String() {
	case "", builtin_types.InventoryListTypeV0:
		if n, err = store.v0.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV1:
		if n, err = store.v1.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV2:
		if n, err = store.v2.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (a InventoryList) PutTransactedWithBlob(
	twb sku.TransactedWithBlob[*sku.List],
) (err error) {
	tipe := twb.GetType()

	switch tipe.String() {
	case "", builtin_types.InventoryListTypeV0:
	case builtin_types.InventoryListTypeV1:
	}

	sku.GetTransactedPool().Put(twb.Transacted)

	return
}

type iterSku = func(*sku.Transacted) bool

func (a InventoryList) StreamInventoryListBlobSkus(
	transactedGetter sku.TransactedGetter,
) iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		sk := transactedGetter.GetSku()
		tipe := sk.GetType()
		blobSha := sk.GetBlobSha()

		var readCloser interfaces.ShaReadCloser

		{
			var err error

			if readCloser, err = a.envRepo.BlobReader(blobSha); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		iter := a.IterInventoryListBlobSkusFromReader(
			tipe,
			readCloser,
		)

		for sk, err := range iter {
			if !yield(sk, err) {
				return
			}
		}
	}
}

func (a InventoryList) AllDecodedObjectsFromStream(
	reader io.Reader,
) iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.Decoder[*triple_hyphen_io.TypedBlob[iterSku]]{
			Metadata: triple_hyphen_io.TypedMetadataCoder[iterSku]{},
			Blob: triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
				a.streamDecoders,
			),
		}

		bufferedReader := ohio.BufferedReader(reader)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: &ids.Type{},
				Blob: func(sk *sku.Transacted) bool {
					return yield(sk, nil)
				},
			},
			bufferedReader,
		); err != nil {
			yield(nil, err)
			return
		}
	}
}

func (a InventoryList) IterInventoryListBlobSkusFromBlobStore(
	tipe ids.Type,
	blobStore interfaces.BlobStore,
	blobSha interfaces.Sha,
) iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		var readCloser interfaces.ShaReadCloser

		{
			var err error

			if readCloser, err = blobStore.BlobReader(blobSha); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			a.streamDecoders,
		)

		bufferedReader := ohio.BufferedReader(readCloser)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: &tipe,
				Blob: func(sk *sku.Transacted) bool {
					return yield(sk, nil)
				},
			},
			bufferedReader,
		); err != nil {
			yield(nil, errors.Wrap(err))
			return
		}
	}
}

func (a InventoryList) IterInventoryListBlobSkusFromReader(
	tipe ids.Type,
	reader io.Reader,
) iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			a.streamDecoders,
		)

		bufferedReader := ohio.BufferedReader(reader)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: &tipe,
				Blob: func(sk *sku.Transacted) bool {
					return yield(sk, nil)
				},
			},
			bufferedReader,
		); err != nil {
			yield(nil, errors.Wrap(err))
			return
		}
	}
}

func (a InventoryList) ReadInventoryListObject(
	tipe ids.Type,
	reader *bufio.Reader,
) (out *sku.Transacted, err error) {
	switch tipe.String() {
	case "", builtin_types.InventoryListTypeV0:
		if _, out, err = a.v0.ReadInventoryListObject(
			reader,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV1:
		if err = a.v1.StreamInventoryListBlobSkus(
			reader,
			func(sk *sku.Transacted) (err error) {
				if out == nil {
					out = sk.CloneTransacted()
				} else {
					err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
					return
				}

				return
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case builtin_types.InventoryListTypeV2:
		if err = a.v2.StreamInventoryListBlobSkus(
			reader,
			func(sk *sku.Transacted) (err error) {
				if out == nil {
					out = sk.CloneTransacted()
				} else {
					err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
					return
				}

				return
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (a InventoryList) ReadInventoryListBlob(
	tipe ids.Type,
	reader *bufio.Reader,
) (list *sku.List, err error) {
	list = sku.MakeList()

	var listFormat sku.ListFormat

	switch tipe.String() {
	case "", builtin_types.InventoryListTypeV0:
		listFormat = a.v0

	case builtin_types.InventoryListTypeV1:
		listFormat = a.v1

	case builtin_types.InventoryListTypeV2:
		listFormat = a.v2
	}

	if err = listFormat.StreamInventoryListBlobSkus(
		reader,
		func(sk *sku.Transacted) (err error) {
			if err = sk.CalculateObjectShas(); err != nil {
				err = errors.Wrap(err)
				return
			}

			if err = list.Add(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
