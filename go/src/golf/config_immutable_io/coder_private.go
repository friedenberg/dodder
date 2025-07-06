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

type typeWithConfigLoadedPrivate = *triple_hyphen_io.TypedBlob[*ConfigPrivatedTypedBlob]

var typedCodersPrivate = map[string]interfaces.CoderBufferedReadWriter[typeWithConfigLoadedPrivate]{
	builtin_types.ImmutableConfigV1: blobV1CoderPrivate{},
	"":                              blobV0CoderPrivate{},
}

var coderPrivate = triple_hyphen_io.Coder[typeWithConfigLoadedPrivate]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*ConfigPrivatedTypedBlob]{},
	Blob: triple_hyphen_io.CoderTypeMap[*ConfigPrivatedTypedBlob](
		typedCodersPrivate,
	),
}

type CoderPrivate struct{}

func (coder CoderPrivate) DecodeFromFile(
	object *ConfigPrivatedTypedBlob,
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

func (CoderPrivate) DecodeFrom(
	subject *ConfigPrivatedTypedBlob,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPrivate.DecodeFrom(
		&triple_hyphen_io.TypedBlob[*ConfigPrivatedTypedBlob]{
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

func (CoderPrivate) EncodeTo(
	subject *ConfigPrivatedTypedBlob,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPrivate.EncodeTo(
		subject,
		writer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
