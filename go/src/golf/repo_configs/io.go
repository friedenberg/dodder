package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var Coder = triple_hyphen_io.CoderToTypedBlob[Config]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Config]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Config](
		map[string]interfaces.CoderBufferedReadWriter[*Config]{
			ids.TypeTomlConfigV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &V0{}
				},
			},
			ids.TypeTomlConfigV1: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &V1{}
				},
			},
		},
	),
}
