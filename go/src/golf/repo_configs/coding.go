package repo_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var Coder = triple_hyphen_io.CoderToTypedBlob[ConfigOverlay]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[ConfigOverlay]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[ConfigOverlay](
		map[string]interfaces.CoderBufferedReadWriter[*ConfigOverlay]{
			ids.TypeTomlConfigV0: triple_hyphen_io.CoderToml[
				ConfigOverlay,
				*ConfigOverlay,
			]{
				Progenitor: func() ConfigOverlay {
					return &V0{}
				},
			},
			ids.TypeTomlConfigV1: triple_hyphen_io.CoderToml[
				ConfigOverlay,
				*ConfigOverlay,
			]{
				Progenitor: func() ConfigOverlay {
					return &V1{}
				},
			},
			ids.TypeTomlConfigV2: triple_hyphen_io.CoderToml[
				ConfigOverlay,
				*ConfigOverlay,
			]{
				Progenitor: func() ConfigOverlay {
					return &V2{}
				},
			},
		},
	),
}
