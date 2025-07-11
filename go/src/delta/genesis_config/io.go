package genesis_config

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type PrivateTypedBlob = triple_hyphen_io.TypedBlob[Private]

var CoderPrivate = triple_hyphen_io.CoderToTypedBlob[Private]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Private]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Private](
		map[string]interfaces.CoderBufferedReadWriter[*Private]{
			ids.ImmutableConfigV2: triple_hyphen_io.CoderToml[
				Private,
				*Private,
			]{
				Progenitor: func() Private {
					return &TomlV2Private{}
				},
			},
			ids.ImmutableConfigV1: triple_hyphen_io.CoderToml[
				Private,
				*Private,
			]{
				Progenitor: func() Private {
					return &TomlV1Private{}
				},
			},
			"": triple_hyphen_io.CoderGob[
				Private,
				*Private,
			]{
				Progenitor: func() Private {
					return &V0Private{}
				},
			},
		},
	),
}

type PublicTypedBlob = triple_hyphen_io.TypedBlob[Public]

var CoderPublic = triple_hyphen_io.CoderToTypedBlob[Public]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Public]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Public](
		map[string]interfaces.CoderBufferedReadWriter[*Public]{
			ids.ImmutableConfigV2: triple_hyphen_io.CoderToml[
				Public,
				*Public,
			]{
				Progenitor: func() Public {
					return &TomlV2Public{}
				},
			},
			ids.ImmutableConfigV1: triple_hyphen_io.CoderToml[
				Public,
				*Public,
			]{
				Progenitor: func() Public {
					return &TomlV1Public{}
				},
			},
			"": triple_hyphen_io.CoderGob[
				Public,
				*Public,
			]{
				Progenitor: func() Public {
					return &V0Public{}
				},
			},
		},
	),
}
