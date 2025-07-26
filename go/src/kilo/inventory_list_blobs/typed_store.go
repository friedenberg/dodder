package inventory_list_blobs

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type TypedStore struct {
	envRepo   env_repo.Env
	boxFormat *box_format.BoxTransacted

	v1 V1
	v2 V2

	objectCoders triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted]
	// TODO rewrite these as simple bufferedreader decoders and have a utility
	// function that turns them into a stream
	streamDecoders map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]
}

func MakeInventoryStore(
	envRepo env_repo.Env,
	boxFormat *box_format.BoxTransacted,
) TypedStore {
	store := TypedStore{
		envRepo:   envRepo,
		boxFormat: boxFormat,
		v1: V1{
			V1ObjectCoder: V1ObjectCoder{
				Box: boxFormat,
			},
		},
		v2: V2{
			V2ObjectCoder: V2ObjectCoder{
				Box:                    boxFormat,
				ImmutableConfigPrivate: envRepo.GetConfigPrivate().Blob,
			},
		},
	}

	store.objectCoders = triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted](
		map[string]interfaces.CoderBufferedReadWriter[*sku.Transacted]{
			ids.TypeInventoryListV1: store.v1.V1ObjectCoder,
			ids.TypeInventoryListV2: store.v2.V2ObjectCoder,
		},
	)

	store.streamDecoders = map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]{
		ids.TypeInventoryListV1: V1IterDecoder{
			V1: store.v1,
		},
		ids.TypeInventoryListV2: V2IterDecoder{
			V2: store.v2,
		},
	}

	return store
}

func (store TypedStore) GetBoxFormat() *box_format.BoxTransacted {
	return store.boxFormat
}

func (store TypedStore) WriteObjectToWriter(
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

	return store.objectCoders.EncodeTo(typedBlob, bufferedWriter)
}

// TODO consume interfaces.SeqError and expose as a coder instead
func (store TypedStore) WriteBlobToWriter(
	tipe ids.Type,
	list sku.Collection,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var format sku.ListFormat

	switch tipe.String() {
	case ids.TypeInventoryListV1:
		format = store.v1

	case ids.TypeInventoryListV2:
		format = store.v2

	default:
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return
	}

	if n, err = WriteInventoryList(
		format,
		quiter.MakeSeqErrorFromSeq(list.All()),
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type iterSku = func(*sku.Transacted) bool

// TODO refactor all the below. Simplify the naming, and move away from the
// stream coders, instead use a utility function like in triple_hyphen_io

func (store TypedStore) StreamInventoryListBlobSkus(
	transactedGetter sku.TransactedGetter,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		object := transactedGetter.GetSku()
		tipe := object.GetType()
		blobDigest := object.GetBlobSha()

		var readCloser interfaces.ReadCloseBlobIdGetter

		if blobDigest.IsNull() {
			return
		}

		{
			var err error

			if readCloser, err = store.envRepo.GetDefaultBlobStore().BlobReader(
				blobDigest,
			); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		iter := store.IterInventoryListBlobSkusFromReader(
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

func (store TypedStore) AllDecodedObjectsFromStream(
	reader io.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.Decoder[*triple_hyphen_io.TypedBlob[iterSku]]{
			Metadata: triple_hyphen_io.TypedMetadataCoder[iterSku]{},
			Blob: triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
				store.streamDecoders,
			),
		}

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(reader)
		defer repoolBufferedReader()

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

func (store TypedStore) IterInventoryListBlobSkusFromBlobStore(
	tipe ids.Type,
	blobStore interfaces.BlobStore,
	blobSha interfaces.BlobId,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		var readCloser interfaces.ReadCloseBlobIdGetter

		{
			var err error

			if readCloser, err = blobStore.BlobReader(blobSha); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			store.streamDecoders,
		)

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(
			readCloser,
		)
		defer repoolBufferedReader()

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

func (store TypedStore) IterInventoryListBlobSkusFromReader(
	tipe ids.Type,
	reader io.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[iterSku](
			store.streamDecoders,
		)

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(reader)
		defer repoolBufferedReader()

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

func (store TypedStore) ReadInventoryListObject(
	tipe ids.Type,
	reader *bufio.Reader,
) (out *sku.Transacted, err error) {
	switch tipe.String() {
	case ids.TypeInventoryListV1:
		iter := store.v1.StreamInventoryListBlobSkus(reader)
		for sk, iterErr := range iter {
			if iterErr != nil {
				err = errors.Wrap(iterErr)
				return
			}

			if out == nil {
				out = sk.CloneTransacted()
			} else {
				err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
				return
			}
		}

	case ids.TypeInventoryListV2:
		iter := store.v2.StreamInventoryListBlobSkus(reader)
		for sk, iterErr := range iter {
			if iterErr != nil {
				err = errors.Wrap(iterErr)
				return
			}

			if out == nil {
				out = sk.CloneTransacted()
			} else {
				err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
				return
			}
		}
	}

	return
}

func (store TypedStore) ReadInventoryListBlob(
	tipe ids.Type,
	reader *bufio.Reader,
) (list *sku.List, err error) {
	list = sku.MakeList()

	var listFormat sku.ListFormat

	switch tipe.String() {
	case ids.TypeInventoryListV1:
		listFormat = store.v1

	case ids.TypeInventoryListV2:
		listFormat = store.v2
	}

	iter := StreamInventoryListSkus(listFormat, reader)

	for sk, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if err = sk.CalculateObjectShas(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = list.Add(sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
