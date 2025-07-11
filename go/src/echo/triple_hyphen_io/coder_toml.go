package triple_hyphen_io

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
)

type CoderToml[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] struct {
	Progenitor func() BLOB
}

func (coder CoderToml[BLOB, BLOB_PTR]) DecodeFrom(
	blob BLOB_PTR,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	tomlDecoder := toml.NewDecoder(bufferedReader)
	clone := coder.Progenitor()

	if err = tomlDecoder.Decode(clone); err != nil {
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

func (coder CoderToml[BLOB, BLOB_PTR]) EncodeTo(
	blob BLOB_PTR,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	tomlEncoder := toml.NewEncoder(bufferedWriter)

	if err = tomlEncoder.Encode(*blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
