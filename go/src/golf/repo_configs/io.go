package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var Coder = triple_hyphen_io.CoderToTypedBlob[Blob]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Blob]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Blob](
		map[string]interfaces.CoderBufferedReadWriter[*Blob]{
			ids.TypeTomlConfigV0: triple_hyphen_io.CoderToml[
				Blob,
				*Blob,
			]{
				Progenitor: func() Blob {
					return &V0{}
				},
			},
			ids.TypeTomlConfigV1: triple_hyphen_io.CoderToml[
				Blob,
				*Blob,
			]{
				Progenitor: func() Blob {
					return &V1{}
				},
			},
		},
	),
}
