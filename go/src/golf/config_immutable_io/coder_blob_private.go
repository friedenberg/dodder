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
	subject typeWithConfigLoadedPrivate,
	r *bufio.Reader,
) (n int64, err error) {
	subject.Blob.ImmutableConfig = &config_immutable.TomlV1Private{}
	td := toml.NewDecoder(r)

	if err = td.Decode(subject.Blob.ImmutableConfig); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobV1CoderPrivate) EncodeTo(
	subject typeWithConfigLoadedPrivate,
	w *bufio.Writer,
) (n int64, err error) {
	te := toml.NewEncoder(w)

	if err = te.Encode(subject.Blob.ImmutableConfig); err != nil {
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
	subject typeWithConfigLoadedPrivate,
	r *bufio.Reader,
) (n int64, err error) {
	subject.Blob.ImmutableConfig = &config_immutable.V0Private{}

	dec := gob.NewDecoder(r)

	if err = dec.Decode(subject.Blob.ImmutableConfig); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobV0CoderPrivate) EncodeTo(
	subject typeWithConfigLoadedPrivate,
	w *bufio.Writer,
) (n int64, err error) {
	dec := gob.NewEncoder(w)

	if err = dec.Encode(subject.Blob.ImmutableConfig); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
