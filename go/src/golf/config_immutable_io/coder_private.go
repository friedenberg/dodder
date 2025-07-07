package config_immutable_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type ConfigPrivateTypedBlob = triple_hyphen_io2.TypedBlob[config_immutable.Private]

var typedCodersPrivate = map[string]interfaces.CoderBufferedReadWriter[*config_immutable.Private]{
	builtin_types.ImmutableConfigV1: blobV1CoderPrivate{},
	"":                              blobV0CoderPrivate{},
}

var coderPrivate = triple_hyphen_io2.CoderToTypedBlob[config_immutable.Private]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[config_immutable.Private]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[config_immutable.Private](
		typedCodersPrivate,
	),
}

type CoderPrivate struct{}

func (coder CoderPrivate) DecodeFromFile(
	typedBlob *ConfigPrivateTypedBlob,
	path string,
) (err error) {
	var reader io.Reader

	{
		var file *os.File

		if file, err = files.OpenExclusiveReadOnly(path); err != nil {
			if errors.IsNotExist(err) {
				err = nil
				reader = bytes.NewBuffer(nil)
			} else {
				err = errors.Wrap(err)
				return
			}
		} else {
			defer errors.DeferredCloser(&err, file)

			reader = file
		}
	}

	if _, err = coder.DecodeFrom(typedBlob, reader); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (CoderPrivate) DecodeFrom(
	typedBlob *ConfigPrivateTypedBlob,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPrivate.DecodeFrom(
		typedBlob,
		reader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (CoderPrivate) EncodeTo(
	typedBlob *ConfigPrivateTypedBlob,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPrivate.EncodeTo(
		typedBlob,
		writer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
