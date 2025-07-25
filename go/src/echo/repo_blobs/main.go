package repo_blobs

import (
	"bufio"
	"crypto"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type Blob interface {
	GetRepoBlob() Blob
	GetPublicKey() repo_signing.PublicKey
	// TODO
	// GetSupportedConnectionTypes() []connection_type.Type
}

type BlobMutable interface {
	Blob
	SetPublicKey(crypto.PublicKey)
}

type TypedBlob = triple_hyphen_io.TypedBlob[*Blob]

var typedCoders = map[string]interfaces.CoderBufferedReadWriter[*TypedBlob]{
	ids.TypeTomlRepoLocalPath:   coderToml[TomlLocalPathV0]{},
	ids.TypeTomlRepoDotenvXdgV0: coderToml[TomlXDGV0]{},
	ids.TypeTomlRepoUri:         coderToml[TomlUriV0]{},
	"":                          coderToml[TomlUriV0]{},
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
			return
		}
	}

	blob := Blob(coder.Blob)
	subject.Blob = &blob

	return
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
			return
		}
	}

	return
}
