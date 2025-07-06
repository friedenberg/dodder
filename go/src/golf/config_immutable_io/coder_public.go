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

type ConfigPublicTypedBlob = triple_hyphen_io2.TypedBlob[config_immutable.ConfigPublic]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[*config_immutable.ConfigPublic]{
	builtin_types.ImmutableConfigV1: blobV1CoderPublic{},
	"":                              blobV0CoderPublic{},
}

var coderPublic = triple_hyphen_io2.CoderToTypedBlob[config_immutable.ConfigPublic]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[config_immutable.ConfigPublic]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[config_immutable.ConfigPublic](
		typedCoders,
	),
}

type CoderPublic struct{}

func (coder CoderPublic) DecodeFromFile(
	object *ConfigPublicTypedBlob,
	p string,
) (err error) {
	var r io.Reader

	{
		var f *os.File

		if f, err = files.OpenExclusiveReadOnly(p); err != nil {
			if errors.IsNotExist(err) {
				err = nil
				r = bytes.NewBuffer(nil)
			} else {
				err = errors.Wrap(err)
				return
			}
		} else {
			defer errors.DeferredCloser(&err, f)

			r = f
		}
	}

	if _, err = coder.DecodeFrom(object, r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (CoderPublic) DecodeFrom(
	typedBlob *ConfigPublicTypedBlob,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPublic.DecodeFrom(
		typedBlob,
		reader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (CoderPublic) EncodeTo(
	typedBlob *ConfigPublicTypedBlob,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPublic.EncodeTo(
		typedBlob,
		writer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
