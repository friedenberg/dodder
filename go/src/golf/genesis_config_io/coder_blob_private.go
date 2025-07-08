package genesis_config_io

import (
	"bufio"
	"encoding/gob"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
)

type blobV1CoderPrivate struct{}

func (blobV1CoderPrivate) DecodeFrom(
	blob *genesis_config.Private,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &genesis_config.TomlV1Private{}
	td := toml.NewDecoder(bufferedReader)

	if err = td.Decode(config); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	*blob = config

	return
}

func (blobV1CoderPrivate) EncodeTo(
	blob *genesis_config.Private,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	encoder := toml.NewEncoder(bufferedWriter)

	if err = encoder.Encode(*blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

type blobV0CoderPrivate struct{}

func (blobV0CoderPrivate) DecodeFrom(
	blob *genesis_config.Private,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &genesis_config.V0Private{}

	dec := gob.NewDecoder(bufferedReader)

	if err = dec.Decode(config); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	*blob = config

	return
}

func (blobV0CoderPrivate) EncodeTo(
	blob *genesis_config.Private,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	dec := gob.NewEncoder(bufferedWriter)

	if err = dec.Encode(*blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
