package config_immutable_io

import (
	"bufio"
	"encoding/gob"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
)

type blobV1CoderPublic struct{}

func (blobV1CoderPublic) DecodeFrom(
	blob *config_immutable.ConfigPublic,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &config_immutable.TomlV1Public{}
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

func (blobV1CoderPublic) EncodeTo(
	blob *config_immutable.ConfigPublic,
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

type blobV0CoderPublic struct{}

func (blobV0CoderPublic) DecodeFrom(
	blob *config_immutable.ConfigPublic,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &config_immutable.V0Public{}

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

func (blobV0CoderPublic) EncodeTo(
	blob *config_immutable.ConfigPublic,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	encoder := gob.NewEncoder(bufferedWriter)

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
