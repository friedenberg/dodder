package config_immutable_io

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type typeWithConfigLoadedPublic = *triple_hyphen_io.TypedStruct[*ConfigLoadedPublic]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[typeWithConfigLoadedPublic]{
	builtin_types.ImmutableConfigV1: blobV1CoderPublic{},
	"":                              blobV0CoderPublic{},
}

var coderPublic = triple_hyphen_io.Coder[typeWithConfigLoadedPublic]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[*ConfigLoadedPublic]{},
	Blob:     triple_hyphen_io.CoderTypeMap[*ConfigLoadedPublic](typedCoders),
}

type CoderPublic struct{}

func (coder CoderPublic) DecodeFromFile(
	object *ConfigLoadedPublic,
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
	subject *ConfigLoadedPublic,
	reader io.Reader,
) (n int64, err error) {
	if n, err = coderPublic.DecodeFrom(
		&triple_hyphen_io.TypedStruct[*ConfigLoadedPublic]{
			Type:   &subject.Type,
			Struct: subject,
		},
		reader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	subject.BlobStoreImmutableConfig = env_dir.MakeConfigFromImmutableBlobConfig(
		subject.ImmutableConfig.GetBlobStoreConfigImmutable(),
	)

	return
}

func (CoderPublic) EncodeTo(
	subject *ConfigLoadedPublic,
	writer io.Writer,
) (n int64, err error) {
	if n, err = coderPublic.EncodeTo(
		&triple_hyphen_io.TypedStruct[*ConfigLoadedPublic]{
			Type:   &subject.Type,
			Struct: subject,
		},
		writer,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
