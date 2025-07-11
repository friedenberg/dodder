package blob_store_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

const (
	TypeV0       = "!toml-blob_store_config-v0"
	TypeVCurrent = TypeV0
)

var Coder = triple_hyphen_io.CoderToTypedBlob[Config]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Config]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Config](
		map[string]interfaces.CoderBufferedReadWriter[*Config]{
			TypeV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &TomlV0{}
				},
			},
		},
	),
}
