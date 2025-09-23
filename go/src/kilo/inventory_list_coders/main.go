package inventory_list_coders

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type funcListFormatConstructor func(
	env_repo.Env,
	*box_format.BoxTransacted,
) coder

var coderConstructors = map[string]funcListFormatConstructor{
	ids.TypeInventoryListV1: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) coder {
		configGenesis := envRepo.GetConfigPublic().Blob

		doddishCoder := doddish{
			box: box,
		}

		return coder{
			listCoder: doddishCoder,
			afterDecoding: func(object *sku.Transacted) (err error) {
				object.Metadata.GetRepoPubKeyMutable().ResetWithMarklId(
					configGenesis.GetPublicKey(),
				)

				if store_version.LessOrEqual(
					envRepo.GetStoreVersion(),
					store_version.V8,
				) {
					if err = object.FinalizeWithoutPubKey(); err != nil {
						err = errors.Wrap(err)
						return err
					}
				} else {
					if err = object.FinalizeUsingObject(); err != nil {
						err = errors.Wrap(err)
						return err
					}
				}

				return err
			},
		}
	},
	ids.TypeInventoryListV2: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) coder {
		doddishCoder := doddish{
			box: box,
		}

		return coder{
			listCoder: doddishCoder,
			beforeEncoding: func(object *sku.Transacted) (err error) {
				if err = (*sku.Transacted).AssertObjectDigestAndObjectSigNotNull(object); err != nil {
					err = errors.Wrap(err)
					return err
				}

				return err
			},
			afterDecoding: (*sku.Transacted).FinalizeAndVerify,
		}
	},
	ids.TypeInventoryListJsonV0: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) coder {
		jsonCoder := jsonV0{
			genesisConfig: envRepo.GetConfigPrivate().Blob,
		}

		return coder{
			listCoder:      jsonCoder,
			beforeEncoding: (*sku.Transacted).Verify,
			afterDecoding:  (*sku.Transacted).FinalizeAndVerify,
		}
	},
}

var (
	_ sku.ListCoder = doddish{}
	_ sku.ListCoder = jsonV0{}
)

func WriteObjectToOpenList(
	format sku.ListCoder,
	object *sku.Transacted,
	list *sku.OpenList,
) (n int64, err error) {
	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(list.Mover)
	defer repoolBufferedWriter()

	if n, err = format.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	list.LastTai = object.GetTai()
	list.Len += 1

	return n, err
}

func WriteInventoryList(
	ctx interfaces.ActiveContext,
	format sku.ListCoder,
	skus interfaces.SeqError[*sku.Transacted],
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	var n1 int64

	var object *sku.Transacted

	for object, err = range skus {
		errors.ContextContinueOrPanic(ctx)

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}

		n1, err = format.EncodeTo(object, bufferedWriter)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}

// TODO also return a repool func
func CollectSkuList(
	ctx interfaces.ActiveContext,
	listFormat sku.ListCoder,
	reader *bufio.Reader,
	list *sku.ListTransacted,
) (err error) {
	iter := streamInventoryList(ctx, listFormat, reader)

	for sk, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return err
		}

		if err = list.Add(sk); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func streamInventoryList(
	ctx interfaces.ActiveContext,
	format sku.ListCoder,
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
	format sku.ListCoder,
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if n, err = format.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

type SeqCoder struct {
	ctx   interfaces.ActiveContext
	coder coder
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

		if _, err = coder.coder.DecodeFrom(object, bufferedReader); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return n, err
			}
		}

		if !yield(object) {
			return n, err
		}
	}

	return n, err
}

type SeqErrorDecoder struct {
	ctx   interfaces.ActiveContext
	coder coder
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

		if _, err = coder.coder.DecodeFrom(object, bufferedReader); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)

				if !yield(nil, err) {
					return n, err
				}
			}
		}

		if !yield(object, nil) {
			return n, err
		}
	}

	return n, err
}

type SeqErrorEncoder struct {
	ctx   interfaces.ActiveContext
	coder coder
}

func (coder SeqErrorDecoder) EncodeTo(
	seq sku.Seq,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	for object, iterErr := range seq {
		errors.ContextContinueOrPanic(coder.ctx)

		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return n, err
		}

		if _, err = coder.coder.EncodeTo(object, bufferedWriter); err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return n, err
			}
		}
	}

	return n, err
}
