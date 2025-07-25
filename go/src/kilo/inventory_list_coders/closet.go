package inventory_list_coders

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

type Closet struct {
	envRepo   env_repo.Env
	boxFormat *box_format.BoxTransacted

	coders map[string]sku.ListFormat

	objectCoders triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted]

	streamDecoders map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool]
}

func MakeCloset(
	envRepo env_repo.Env,
	box *box_format.BoxTransacted,
) Closet {
	store := Closet{
		envRepo:   envRepo,
		boxFormat: box,
	}

	store.coders = make(map[string]sku.ListFormat, len(coderConstructors))

	for tipe, coderConstructor := range coderConstructors {
		store.coders[tipe] = coderConstructor(envRepo, box)
	}

	{
		coders := make(
			map[string]interfaces.CoderBufferedReadWriter[*sku.Transacted],
			len(store.coders),
		)

		for tipe, coder := range store.coders {
			coders[tipe] = coder
		}

		store.objectCoders = triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted](
			coders,
		)
	}

	{
		coders := make(
			map[string]interfaces.DecoderFromBufferedReader[func(*sku.Transacted) bool],
			len(store.coders),
		)

		for tipe, coder := range store.coders {
			coders[tipe] = IterCoder{
				ctx:        envRepo,
				ListFormat: coder,
			}
		}

		store.streamDecoders = coders
	}

	return store
}

func (store Closet) GetBoxFormat() *box_format.BoxTransacted {
	return store.boxFormat
}

func (store Closet) WriteObjectToWriter(
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
func (store Closet) WriteBlobToWriter(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	list sku.Collection,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	format, ok := store.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return
	}

	if n, err = WriteInventoryList(
		ctx,
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

func (store Closet) StreamInventoryListBlobSkus(
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

func (store Closet) AllDecodedObjectsFromStream(
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

func (store Closet) IterInventoryListBlobSkusFromBlobStore(
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

func (store Closet) IterInventoryListBlobSkusFromReader(
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

func (store Closet) ReadInventoryListObject(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	reader *bufio.Reader,
) (out *sku.Transacted, err error) {
	format, ok := store.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return
	}

	iter := StreamInventoryList(ctx, format, reader)

	for object, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if out == nil {
			out = object.CloneTransacted()
		} else {
			err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
			return
		}
	}

	return
}

func (store Closet) ReadInventoryListBlob(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	reader *bufio.Reader,
) (list *sku.List, err error) {
	list = sku.MakeList()

	format, ok := store.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return
	}

	iter := StreamInventoryList(ctx, format, reader)

	for object, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if err = object.CalculateObjectDigests(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = list.Add(object); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
