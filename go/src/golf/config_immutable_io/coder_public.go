package config_immutable_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type typeWithConfigLoadedPublic = *triple_hyphen_io.TypedBlob[*ConfigPublicTypedBlob]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[typeWithConfigLoadedPublic]{
	builtin_types.ImmutableConfigV1: blobV1CoderPublic{},
	"":                              blobV0CoderPublic{},
}

var coderPublic = triple_hyphen_io.Coder[typeWithConfigLoadedPublic]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*ConfigPublicTypedBlob]{},
	Blob: triple_hyphen_io.CoderTypeMap[*ConfigPublicTypedBlob](
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
	subject *ConfigPublicTypedBlob,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPublic.DecodeFrom(
		&triple_hyphen_io.TypedBlob[*ConfigPublicTypedBlob]{
			Type: &subject.Type,
			Blob: subject,
		},
		reader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (CoderPublic) EncodeTo(
	subject *ConfigPublicTypedBlob,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPublic.EncodeTo(
		&triple_hyphen_io.TypedBlob[*ConfigPublicTypedBlob]{
			Type: &subject.Type,
			Blob: subject,
		},
		writer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
