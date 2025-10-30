package repo_blobs

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type Blob interface {
	GetRepoBlob() Blob
	GetPublicKey() interfaces.MarklId
	// TODO
	// GetSupportedConnectionTypes() []connection_type.Type
}

type BlobMutable interface {
	Blob
	SetPublicKey(interfaces.MarklId)
}

type TypedBlob = triple_hyphen_io.TypedBlob[*Blob]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[*TypedBlob]{
	ids.TypeTomlRepoLocalOverridePath: coderToml[TomlLocalOverridePathV0]{},
	ids.TypeTomlRepoDotenvXdgV0:       coderToml[TomlXDGV0]{},
	ids.TypeTomlRepoUri:               coderToml[TomlUriV0]{},
	"":                                coderToml[TomlUriV0]{},
}

var Coder = interfaces.CoderBufferedReadWriter[*TypedBlob](
	triple_hyphen_io.CoderTypeMap[*Blob](typedCoders),
)

type coderToml[T Blob] struct {
	Blob T
}

func (coder coderToml[T]) DecodeFrom(
	subject *TypedBlob,
	reader *bufio.Reader,
) (n int64, err error) {
	decoder := toml.NewDecoder(reader)

	if err = decoder.Decode(&coder.Blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return n, err
		}
	}

	blob := Blob(coder.Blob)
	subject.Blob = &blob

	return n, err
}

func (coderToml[_]) EncodeTo(
	subject *TypedBlob,
	writer *bufio.Writer,
) (n int64, err error) {
	encoder := toml.NewEncoder(writer)

	if err = encoder.Encode(subject.Blob); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			err = errors.Wrap(err)
			return n, err
		}
	}

	return n, err
}
