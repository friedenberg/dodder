package inventory_list_coders

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type (
	funcIterSeq      = func(*sku.Transacted) bool
	funcIterSeqError = func(*sku.Transacted, error) bool
)

type Closet struct {
	envRepo   env_repo.Env
	boxFormat *box_format.BoxTransacted

	coders map[string]coder

	objectCoders triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted]

	seqDecoders      map[string]interfaces.DecoderFromBufferedReader[funcIterSeq]
	seqErrorDecoders map[string]interfaces.DecoderFromBufferedReader[funcIterSeqError]
	seqEncoders      map[string]interfaces.EncoderToBufferedWriter[sku.Seq]
}

func MakeCloset(
	envRepo env_repo.Env,
	box *box_format.BoxTransacted,
) Closet {
	store := Closet{
		envRepo:   envRepo,
		boxFormat: box,
	}

	store.coders = make(map[string]coder, len(coderConstructors))

	for tipe, coderConstructor := range coderConstructors {
		store.coders[tipe] = coderConstructor(envRepo, box)
	}

	{
		coders := make(
			map[string]interfaces.CoderBufferedReadWriter[*sku.Transacted],
			len(store.coders),
		)

		for key, value := range store.coders {
			coders[key] = value
		}

		store.objectCoders = triple_hyphen_io.CoderTypeMapWithoutType[sku.Transacted](
			coders,
		)
	}

	{
		coders := make(
			map[string]interfaces.DecoderFromBufferedReader[funcIterSeq],
			len(store.coders),
		)

		for tipe, coder := range store.coders {
			coders[tipe] = SeqCoder{
				ctx:   envRepo,
				coder: coder,
			}
		}

		store.seqDecoders = coders
	}

	{
		coders := make(
			map[string]interfaces.DecoderFromBufferedReader[funcIterSeqError],
			len(store.coders),
		)

		for tipe, coder := range store.coders {
			coders[tipe] = SeqErrorDecoder{
				ctx:   envRepo,
				coder: coder,
			}
		}

		store.seqErrorDecoders = coders
	}

	{
		coders := make(
			map[string]interfaces.EncoderToBufferedWriter[sku.Seq],
			len(store.coders),
		)

		for tipe, coder := range store.coders {
			coders[tipe] = SeqErrorDecoder{
				ctx:   envRepo,
				coder: coder,
			}
		}

		store.seqEncoders = coders
	}

	return store
}

func (closet Closet) GetBoxFormat() *box_format.BoxTransacted {
	return closet.boxFormat
}

func (closet Closet) GetCoderForType(tipe ids.Type) sku.ListCoder {
	format, ok := closet.coders[tipe.String()]

	if !ok {
		panic(errors.Errorf("unsupported inventory list type: %q", tipe))
	}

	return format
}

func (closet Closet) WriteObjectToWriter(
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

	if n, err = closet.objectCoders.EncodeTo(typedBlob, bufferedWriter); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

// TODO consume interfaces.SeqError and expose as a coder instead
func (closet Closet) WriteBlobToWriter(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	seq sku.Seq,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	format, ok := closet.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return n, err
	}

	if n, err = WriteInventoryList(
		ctx,
		format,
		seq,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (closet Closet) WriteTypedBlobToWriter(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	seq sku.Seq,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	decoder := triple_hyphen_io.Encoder[*triple_hyphen_io.TypedBlob[sku.Seq]]{
		Metadata: triple_hyphen_io.TypedMetadataCoder[sku.Seq]{},
		Blob: triple_hyphen_io.EncoderTypeMapWithoutType[sku.Seq](
			closet.seqEncoders,
		),
	}

	if _, err = decoder.EncodeTo(
		&triple_hyphen_io.TypedBlob[sku.Seq]{
			Type: tipe,
			Blob: seq,
		},
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

// TODO refactor all the below. Simplify the naming, and move away from the
// stream coders, instead use a utility function like in triple_hyphen_io

func (closet Closet) StreamInventoryListBlobSkus(
	transactedGetter sku.TransactedGetter,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		object := transactedGetter.GetSku()
		tipe := object.GetType()
		blobDigest := object.GetBlobDigest()

		var readCloser interfaces.BlobReader

		if blobDigest.IsNull() {
			return
		}

		{
			var err error

			if readCloser, err = closet.envRepo.GetDefaultBlobStore().MakeBlobReader(
				blobDigest,
			); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		iter := closet.IterInventoryListBlobSkusFromReader(
			tipe,
			readCloser,
		)

		for object, err := range iter {
			if !yield(object, err) {
				return
			}
		}
	}
}

func (closet Closet) AllDecodedObjectsFromStream(
	reader io.Reader,
	afterDecoding func(*sku.Transacted) error,
) interfaces.SeqError[*sku.Transacted] {
	var coders map[string]interfaces.DecoderFromBufferedReader[funcIterSeqError]

	if afterDecoding == nil {
		coders = closet.seqErrorDecoders
	} else {
		coders = make(
			map[string]interfaces.DecoderFromBufferedReader[funcIterSeqError],
			len(closet.coders),
		)

		for tipe, coder := range closet.coders {
			coder.afterDecoding = afterDecoding
			coders[tipe] = SeqErrorDecoder{
				ctx:   closet.envRepo,
				coder: coder,
			}
		}
	}

	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.Decoder[*triple_hyphen_io.TypedBlob[funcIterSeqError]]{
			Metadata: triple_hyphen_io.TypedMetadataCoder[funcIterSeqError]{},
			Blob: triple_hyphen_io.DecoderTypeMapWithoutType[funcIterSeqError](
				coders,
			),
		}

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(reader)
		defer repoolBufferedReader()

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[funcIterSeqError]{
				Type: ids.Type{},
				Blob: func(object *sku.Transacted, err error) bool {
					return yield(object, err)
				},
			},
			bufferedReader,
		); err != nil {
			yield(nil, errors.Wrap(err))
			return
		}
	}
}

func (closet Closet) IterInventoryListBlobSkusFromBlobStore(
	tipe ids.Type,
	blobStore interfaces.BlobStore,
	blobId interfaces.MarklId,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		var readCloser interfaces.BlobReader

		{
			var err error

			if readCloser, err = blobStore.MakeBlobReader(blobId); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.DeferredYieldCloser(yield, readCloser)

		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[funcIterSeq](
			closet.seqDecoders,
		)

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(
			readCloser,
		)
		defer repoolBufferedReader()

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[funcIterSeq]{
				Type: tipe,
				Blob: func(sk *sku.Transacted) bool {
					return yield(sk, nil)
				},
			},
			bufferedReader,
		); err != nil {
			yield(nil, errors.Wrapf(err, "List Blob Id: %s", blobId))
			return
		}
	}
}

func (closet Closet) IterInventoryListBlobSkusFromReader(
	tipe ids.Type,
	reader io.Reader,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		decoder := triple_hyphen_io.DecoderTypeMapWithoutType[funcIterSeq](
			closet.seqDecoders,
		)

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(reader)
		defer repoolBufferedReader()

		if _, err := decoder.DecodeFrom(
			&triple_hyphen_io.TypedBlob[funcIterSeq]{
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

func (closet Closet) ReadInventoryListObject(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	reader *bufio.Reader,
) (out *sku.Transacted, err error) {
	format, ok := closet.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return out, err
	}

	iter := streamInventoryList(ctx, format, reader)

	for object, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return out, err
		}

		if out == nil {
			out = object.CloneTransacted()
		} else {
			err = errors.ErrorWithStackf("expected only one sku.Transacted, but read more than one")
			return out, err
		}
	}

	return out, err
}

func (closet Closet) ReadInventoryListBlob(
	ctx interfaces.ActiveContext,
	tipe ids.Type,
	reader *bufio.Reader,
) (list *sku.ListTransacted, err error) {
	list = sku.MakeListTransacted()

	format, ok := closet.coders[tipe.String()]

	if !ok {
		err = errors.Errorf("unsupported inventory list type: %q", tipe)
		return list, err
	}

	iter := streamInventoryList(ctx, format, reader)

	for object, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return list, err
		}

		if err = list.Add(object); err != nil {
			err = errors.Wrap(err)
			return list, err
		}
	}

	return list, err
}
