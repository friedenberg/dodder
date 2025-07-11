package triple_hyphen_io

import (
	"bufio"
	"encoding/gob"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type CoderGob[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] struct {
	Progenitor func() BLOB
}

func (coder CoderGob[BLOB, BLOB_PTR]) DecodeFrom(
	blob BLOB_PTR,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	gobDecoder := gob.NewDecoder(bufferedReader)
	clone := coder.Progenitor()

	if err = gobDecoder.Decode(clone); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	*blob = clone

	return
}

func (coder CoderGob[BLOB, BLOB_PTR]) EncodeTo(
	blob BLOB_PTR,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	gobEncoder := gob.NewEncoder(bufferedWriter)

	if err = gobEncoder.Encode(*blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
