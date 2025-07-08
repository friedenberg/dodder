package genesis_config_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type PrivateTypedBlob = triple_hyphen_io2.TypedBlob[genesis_config.Private]

var typedCodersPrivate = map[string]interfaces.CoderBufferedReadWriter[*genesis_config.Private]{
	builtin_types.ImmutableConfigV1: blobV1CoderPrivate{},
	"":                              blobV0CoderPrivate{},
}

var coderPrivate = triple_hyphen_io2.CoderToTypedBlob[genesis_config.Private]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[genesis_config.Private]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[genesis_config.Private](
		typedCodersPrivate,
	),
}

type CoderPrivate struct{}

func (coder CoderPrivate) DecodeFromFile(
	typedBlob *PrivateTypedBlob,
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
	typedBlob *PrivateTypedBlob,
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
	typedBlob *PrivateTypedBlob,
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
