package genesis_config_io

import (
	"bufio"
	"encoding/gob"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
)

type blobV1CoderPublic struct{}

func (blobV1CoderPublic) DecodeFrom(
	blob *genesis_config.Public,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &genesis_config.TomlV1Public{}
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
	blob *genesis_config.Public,
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
	blob *genesis_config.Public,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	config := &genesis_config.V0Public{}

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
	blob *genesis_config.Public,
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
