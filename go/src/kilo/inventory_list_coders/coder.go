package inventory_list_coders

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	coderState struct {
		*int64
		eof bool
	}

	funcDecode func(coderState, *sku.Transacted, *bufio.Reader) error
	funcEncode func(coderState, *sku.Transacted, *bufio.Writer) error
)

type coder struct {
	encoders []funcEncode
	decoders []funcDecode
}

func makeCoder(
	encoderDecoder interfaces.CoderBufferedReadWriter[*sku.Transacted],
) coder {
	return coder{
		decoders: []funcDecode{
			func(state coderState, object *sku.Transacted, bufferedReader *bufio.Reader) (err error) {
				*state.int64, err = encoderDecoder.DecodeFrom(
					object,
					bufferedReader,
				)
				return
			},
		},
		encoders: []funcEncode{
			func(state coderState, object *sku.Transacted, bufferedWriter *bufio.Writer) (err error) {
				*state.int64, err = encoderDecoder.EncodeTo(
					object,
					bufferedWriter,
				)
				return
			},
		},
	}
}

func (coder coder) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	state := coderState{
		int64: &n,
	}

	for _, encoder := range coder.encoders {
		if err = encoder(state, object, bufferedWriter); err != nil {
			state.eof = err == io.EOF
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (coder coder) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	state := coderState{
		int64: &n,
	}

	for _, decoder := range coder.decoders {
		if err = decoder(state, object, bufferedReader); err == io.EOF {
			state.eof = true
		} else {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return
		}
	}

	return
}
