package config_immutable_io

import (
	"bufio"
	"encoding/gob"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
)

type blobV1CoderPrivate struct{}

func (blobV1CoderPrivate) DecodeFrom(
	blob *config_immutable.ConfigPrivate,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &config_immutable.TomlV1Private{}
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
	blob *config_immutable.ConfigPrivate,
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
	blob *config_immutable.ConfigPrivate,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &config_immutable.V0Private{}

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
	blob *config_immutable.ConfigPrivate,
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
