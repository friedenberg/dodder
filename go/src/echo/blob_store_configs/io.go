package blob_store_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var Coder = triple_hyphen_io.CoderToTypedBlob[Config]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Config]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Config](
		map[string]interfaces.CoderBufferedReadWriter[*Config]{
			ids.TypeTomlBlobStoreConfigV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &TomlV0{}
				},
			},
			ids.TypeTomlBlobStoreConfigSftpExplicitV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &TomlSFTPV0{}
				},
			},
			ids.TypeTomlBlobStoreConfigSftpViaSSHConfigV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &TomlSFTPViaSSHConfigV0{}
				},
			},
		},
	),
}
