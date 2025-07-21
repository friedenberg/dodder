package typed_blob_store

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
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

	objectCoders triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted]
	// TODO rewrite these as simple bufferedreader decoders and have a utility
	// function that turns them into a stream
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
				ImmutableConfigPrivate: dirLayout.GetConfigPrivate().Blob,
			},
		},
	}

	s.objectCoders = triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted](
		map[string]interfaces.CoderBufferedReadWriter[*sku.Transacted]{
			"": inventory_list_blobs.V0ObjectCoder{
				V0: s.v0,
			},
			ids.TypeInventoryListV1: s.v1.V1ObjectCoder,
			ids.TypeInventoryListV2: s.v2.V2ObjectCoder,
		},
	)

	s.streamDecoders = map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]{
		"": inventory_list_blobs.V0IterDecoder{
			V0: s.v0,
		},
		ids.TypeInventoryListV1: inventory_list_blobs.V1IterDecoder{
			V1: s.v1,
		},
		ids.TypeInventoryListV2: inventory_list_blobs.V2IterDecoder{
			V2: s.v2,
		},
	}

	return s
}

func (typedBlobStore InventoryList) GetCommonStore() sku.BlobStore[*sku.List] {
	return typedBlobStore
}

func (typedBlobStore InventoryList) GetObjectFormat() object_inventory_format.Format {
	return typedBlobStore.objectFormat
}

func (typedBlobStore InventoryList) GetBoxFormat() *box_format.BoxTransacted {
	return typedBlobStore.boxFormat
}

func (typedBlobStore InventoryList) GetTransactedWithBlob(
	inventoryList sku.TransactedGetter,
) (objectAndBlob sku.TransactedWithBlob[*sku.List], n int64, err error) {
	objectAndBlob.Transacted = inventoryList.GetSku()
	blobSha := objectAndBlob.GetBlobSha()

	var readCloser interfaces.ReadCloserDigester

	if readCloser, err = typedBlobStore.envRepo.GetDefaultBlobStore().BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader := ohio.BufferedReader(readCloser)
	defer pool.GetBufioReader().Put(bufferedReader)

	if n, err = typedBlobStore.GetTransactedWithBlobFromReader(
		&objectAndBlob,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (typedBlobStore InventoryList) GetTransactedWithBlobFromReader(
	twb *sku.TransactedWithBlob[*sku.List],
	reader *bufio.Reader,
) (n int64, err error) {
	tipe := twb.GetType()
	twb.Blob = sku.MakeList()

	switch tipe.String() {
	case "", ids.TypeInventoryListV0:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			typedBlobStore.v0,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeInventoryListV1:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			typedBlobStore.v1,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeInventoryListV2:
		if err = inventory_list_blobs.ReadInventoryListBlob(
			typedBlobStore.v2,
			reader,
			twb.Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (typedBlobStore InventoryList) WriteObjectToWriter(
	tipe ids.Type,
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	// Create TypedBlob and reset its Blob field directly from source
	typedBlob := &triple_hyphen_io.TypedBlob[sku.Transacted]{
		Type: tipe,
		// Blob field is zero-value sku.Transacted
	}
	sku.TransactedResetter.ResetWith(&typedBlob.Blob, object)

	return typedBlobStore.objectCoders.EncodeTo(typedBlob, bufferedWriter)
}

func (typedBlobStore InventoryList) WriteBlobToWriter(
	tipe ids.Type,
	list sku.Collection,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	switch tipe.String() {
	case "", ids.TypeInventoryListV0:
		if n, err = typedBlobStore.v0.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeInventoryListV1:
		if n, err = typedBlobStore.v1.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeInventoryListV2:
		if n, err = typedBlobStore.v2.WriteInventoryListBlob(
			list,
			bufferedWriter,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (typedBlobStore InventoryList) PutTransactedWithBlob(
	twb sku.TransactedWithBlob[*sku.List],
) (err error) {
	tipe := twb.GetType()

	switch tipe.String() {
	case "", ids.TypeInventoryListV0:
	case ids.TypeInventoryListV1:
	}

	sku.GetTransactedPool().Put(twb.Transacted)

	return
}

type iterSku = func(*sku.Transacted) bool

// TODO refactor all the below. Simplify the naming, and move away from the
// stream coders, instead use a utility function like in triple_hyphen_io

func (typedBlobStore InventoryList) StreamInventoryListBlobSkus(
	transactedGetter sku.TransactedGetter,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		sk := transactedGetter.GetSku()
		tipe := sk.GetType()
		blobSha := sk.GetBlobSha()

		var readCloser interfaces.ReadCloserDigester

		{
			var err error

			if readCloser, err = typedBlobStore.envRepo.GetDefaultBlobStore().BlobReader(blobSha); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		iter := typedBlobStore.IterInventoryListBlobSkusFromReader(
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

func (typedBlobStore InventoryList) AllDecodedObjectsFromStream(
	reader io.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.Decoder[*triple_hyphen_io.TypedBlob[iterSku]]{
			Metadata: triple_hyphen_io.TypedMetadataCoder[iterSku]{},
			Blob: triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
				typedBlobStore.streamDecoders,
			),
		}

		bufferedReader := ohio.BufferedReader(reader)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: ids.Type{},
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

func (typedBlobStore InventoryList) IterInventoryListBlobSkusFromBlobStore(
	tipe ids.Type,
	blobStore interfaces.BlobStore,
	blobSha interfaces.Digest,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		var readCloser interfaces.ReadCloserDigester

		{
			var err error

			if readCloser, err = blobStore.BlobReader(blobSha); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			typedBlobStore.streamDecoders,
		)

		bufferedReader := ohio.BufferedReader(readCloser)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: tipe,
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

func (typedBlobStore InventoryList) IterInventoryListBlobSkusFromReader(
	tipe ids.Type,
	reader io.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			typedBlobStore.streamDecoders,
		)

		bufferedReader := ohio.BufferedReader(reader)
		defer pool.GetBufioReader().Put(bufferedReader)

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[iterSku]{
				Type: tipe,
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

func (typedBlobStore InventoryList) ReadInventoryListObject(
	tipe ids.Type,
	reader *bufio.Reader,
) (out *sku.Transacted, err error) {
	switch tipe.String() {
	case "", ids.TypeInventoryListV0:
		if _, out, err = typedBlobStore.v0.ReadInventoryListObject(
			reader,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

	case ids.TypeInventoryListV1:
		if err = typedBlobStore.v1.StreamInventoryListBlobSkus(
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

	case ids.TypeInventoryListV2:
		if err = typedBlobStore.v2.StreamInventoryListBlobSkus(
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

func (typedBlobStore InventoryList) ReadInventoryListBlob(
	tipe ids.Type,
	reader *bufio.Reader,
) (list *sku.List, err error) {
	list = sku.MakeList()

	var listFormat sku.ListFormat

	switch tipe.String() {
	case "", ids.TypeInventoryListV0:
		listFormat = typedBlobStore.v0

	case ids.TypeInventoryListV1:
		listFormat = typedBlobStore.v1

	case ids.TypeInventoryListV2:
		listFormat = typedBlobStore.v2
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
