package triple_hyphen_io2

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type TypedBlob[BLOB any] struct {
	// TODO determine why this needs to be a pointer
	Type *ids.Type
	Blob BLOB
}

func (typedBlob *TypedBlob[BLOB]) GetTypePtr() *ids.Type {
	if typedBlob.Type == nil {
		typedBlob.Type = &ids.Type{}
	}

	return typedBlob.Type
}

type TypedBlobEmpty = TypedBlob[struct{}]

type CoderTypeMap[BLOB any] map[string]interfaces.CoderBufferedReadWriter[*TypedBlob[BLOB]]

func (coderTypeMap CoderTypeMap[BLOB]) DecodeFrom(
	typedBlob *TypedBlob[BLOB],
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetTypePtr()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.DecodeFrom(typedBlob, bufferedReader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coderTypeMap CoderTypeMap[BLOB]) EncodeTo(
	typedBlob *TypedBlob[BLOB],
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	tipe := typedBlob.GetTypePtr()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.EncodeTo(typedBlob, bufferedWriter); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type DecoderTypeMapWithoutType[BLOB any] map[string]interfaces.DecoderFromBufferedReader[BLOB]

func (decoderTypeMap DecoderTypeMapWithoutType[BLOB]) DecodeFrom(
	typedBlob *TypedBlob[BLOB],
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetTypePtr()
	decoder, ok := decoderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = decoder.DecodeFrom(
		typedBlob.Blob,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type CoderTypeMapWithoutType[BLOB any] map[string]interfaces.CoderBufferedReadWriter[*BLOB]

func (coderTypeMap CoderTypeMapWithoutType[BLOB]) DecodeFrom(
	typedBlob *TypedBlob[BLOB],
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	tipe := typedBlob.GetTypePtr()
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.DecodeFrom(&typedBlob.Blob, bufferedReader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coderTypeMap CoderTypeMapWithoutType[BLOB]) EncodeTo(
	typedBlob *TypedBlob[BLOB],
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	tipe := typedBlob.Type
	coder, ok := coderTypeMap[tipe.String()]

	if !ok {
		err = errors.ErrorWithStackf("no coders available for type: %q", tipe)
		return
	}

	if n, err = coder.EncodeTo(&typedBlob.Blob, bufferedWriter); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
