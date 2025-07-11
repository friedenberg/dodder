package genesis_config_io

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
)

type PublicTypedBlob = triple_hyphen_io2.TypedBlob[genesis_config.Public]

var CoderPublic = triple_hyphen_io2.CoderToTypedBlob[genesis_config.Public]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[genesis_config.Public]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[genesis_config.Public](
		map[string]interfaces.CoderBufferedReadWriter[*genesis_config.Public]{
			ids.ImmutableConfigV2: triple_hyphen_io2.CoderToml[
				genesis_config.Public,
				*genesis_config.Public,
			]{
				Progenitor: func() genesis_config.Public {
					return &genesis_config.TomlV2Public{}
				},
			},
			ids.ImmutableConfigV1: triple_hyphen_io2.CoderToml[
				genesis_config.Public,
				*genesis_config.Public,
			]{
				Progenitor: func() genesis_config.Public {
					return &genesis_config.TomlV1Public{}
				},
			},
			"": triple_hyphen_io2.CoderGob[
				genesis_config.Public,
				*genesis_config.Public,
			]{
				Progenitor: func() genesis_config.Public {
					return &genesis_config.V0Public{}
				},
			},
		},
	),
}
