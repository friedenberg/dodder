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

// var Coder = triple_hyphen_io.CoderToTypedBlob[Config]{
// 	Metadata: triple_hyphen_io.TypedMetadataCoder[Config]{},
// 	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Config](
// 		map[string]interfaces.CoderBufferedReadWriter[*Config]{
// 			ids.TypeTomlBlobStoreConfigV0: triple_hyphen_io.CoderToml[
// 				Config,
// 				*Config,
// 			]{
// 				Progenitor: func() Config {
// 					return &TomlV0{}
// 				},
// 			},
// 			ids.TypeTomlBlobStoreConfigV1: triple_hyphen_io.CoderToml[
// 				Config,
// 				*Config,
// 			]{
// 				Progenitor: func() Config {
// 					return &TomlV1{}
// 				},
// 			},
// 			ids.TypeTomlBlobStoreConfigV2: triple_hyphen_io.CoderToml[
// 				Config,
// 				*Config,
// 			]{
// 				Progenitor: func() Config {
// 					return &TomlV2{}
// 				},
// 			},
// 			ids.TypeTomlBlobStoreConfigSftpExplicitV0: triple_hyphen_io.CoderToml[
// 				Config,
// 				*Config,
// 			]{
// 				Progenitor: func() Config {
// 					return &TomlSFTPV0{}
// 				},
// 			},
// 			ids.TypeTomlBlobStoreConfigSftpViaSSHConfigV0: triple_hyphen_io.CoderToml[
// 				Config,
// 				*Config,
// 			]{
// 				Progenitor: func() Config {
// 					return &TomlSFTPViaSSHConfigV0{}
// 				},
// 			},
// 		},
// 	),
// }

type TypedBlob = triple_hyphen_io.TypedBlob[*Blob]

// TODO look into if this can be replaced with triple_hyphen_io.CoderTomls
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
