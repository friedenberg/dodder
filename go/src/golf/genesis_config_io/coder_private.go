package genesis_config_io

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
)

type PrivateTypedBlob = triple_hyphen_io2.TypedBlob[genesis_config.Private]

var CoderPrivate = triple_hyphen_io2.CoderToTypedBlob[genesis_config.Private]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[genesis_config.Private]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[genesis_config.Private](
		map[string]interfaces.CoderBufferedReadWriter[*genesis_config.Private]{
			ids.ImmutableConfigV2: triple_hyphen_io2.CoderToml[
				genesis_config.Private,
				*genesis_config.Private,
			]{
				Progenitor: func() genesis_config.Private {
					return &genesis_config.TomlV2Private{}
				},
			},
			ids.ImmutableConfigV1: triple_hyphen_io2.CoderToml[
				genesis_config.Private,
				*genesis_config.Private,
			]{
				Progenitor: func() genesis_config.Private {
					return &genesis_config.TomlV1Private{}
				},
			},
			"": triple_hyphen_io2.CoderGob[
				genesis_config.Private,
				*genesis_config.Private,
			]{
				Progenitor: func() genesis_config.Private {
					return &genesis_config.V0Private{}
				},
			},
		},
	),
}
