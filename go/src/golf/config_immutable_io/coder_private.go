package config_immutable_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type ConfigPrivateTypedBlob struct {
	ids.Type
	ImmutableConfig config_immutable.ConfigPrivate
}

type typeWithConfigLoadedPrivate = *triple_hyphen_io.TypedBlob[*ConfigPrivateTypedBlob]

var typedCodersPrivate = map[string]interfaces.CoderBufferedReadWriter[typeWithConfigLoadedPrivate]{
	builtin_types.ImmutableConfigV1: blobV1CoderPrivate{},
	"":                              blobV0CoderPrivate{},
}

var coderPrivate = triple_hyphen_io.CoderToTypedBlob[*ConfigPrivateTypedBlob]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*ConfigPrivateTypedBlob]{},
	Blob: triple_hyphen_io.CoderTypeMap[*ConfigPrivateTypedBlob](
		typedCodersPrivate,
	),
}

type CoderPrivate struct{}

func (coder CoderPrivate) DecodeFromFile(
	object *ConfigPrivateTypedBlob,
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
	subject *ConfigPrivateTypedBlob,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPrivate.DecodeFrom(
		&triple_hyphen_io.TypedBlob[*ConfigPrivateTypedBlob]{
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
	subject *ConfigPrivateTypedBlob,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPrivate.EncodeTo(
		&triple_hyphen_io.TypedBlob[*ConfigPrivateTypedBlob]{
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
