package triple_hyphen_io

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type TypedBlob[BLOB any] struct {
	// TODO determine why this needs to be a pointer
	Type *ids.Type
	Blob BLOB
}

func (typedBlob *TypedBlob[S]) GetType() *ids.Type {
	if typedBlob.Type == nil {
		typedBlob.Type = &ids.Type{}
	}

	return typedBlob.Type
}

type CoderTypeMap[BLOB any] map[string]interfaces.CoderBufferedReadWriter[*TypedBlob[BLOB]]

func (coderTypeMap CoderTypeMap[S]) DecodeFrom(
	typedBlob *TypedBlob[S],
	reader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetType()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.DecodeFrom(typedBlob, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coderTypeMap CoderTypeMap[S]) EncodeTo(
	typedBlob *TypedBlob[S],
	writer *bufio.Writer,
) (n int64, err error) {
	tipe := typedBlob.GetType()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.EncodeTo(typedBlob, writer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type DecoderTypeMapWithoutType[BLOB any] map[string]interfaces.DecoderFromBufferedReader[BLOB]

func (decoderTypeMap DecoderTypeMapWithoutType[S]) DecodeFrom(
	typedBlob *TypedBlob[S],
	reader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetType()
	decoder, ok := decoderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	bufferedReader := ohio.BufferedReader(reader)
	defer pool.GetBufioReader().Put(bufferedReader)

	if n, err = decoder.DecodeFrom(
		typedBlob.Blob,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type CoderTypeMapWithoutType[BLOB any] map[string]interfaces.CoderBufferedReadWriter[BLOB]

func (coderTypeMap CoderTypeMapWithoutType[S]) DecodeFrom(
	typedBlob *TypedBlob[S],
	reader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetType()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.DecodeFrom(typedBlob.Blob, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coderTypeMap CoderTypeMapWithoutType[S]) EncodeTo(
	typedBlob *TypedBlob[S],
	writer *bufio.Writer,
) (n int64, err error) {
	tipe := typedBlob.Type
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.EncodeTo(typedBlob.Blob, writer); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
