package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var CoderPrivate = triple_hyphen_io.CoderToTypedBlob[Private]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Private]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Private](
		map[string]interfaces.CoderBufferedReadWriter[*Private]{
			ids.TypeTomlConfigImmutableV2: triple_hyphen_io.CoderToml[
				Private,
				*Private,
			]{
				Progenitor: func() Private {
					return &TomlV2Private{}
				},
			},
			ids.TypeTomlConfigImmutableV1: triple_hyphen_io.CoderToml[
				Private,
				*Private,
			]{
				Progenitor: func() Private {
					return &TomlV1Private{}
				},
			},
		},
	),
}

var CoderPublic = triple_hyphen_io.CoderToTypedBlob[Public]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Public]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Public](
		map[string]interfaces.CoderBufferedReadWriter[*Public]{
			ids.TypeTomlConfigImmutableV2: triple_hyphen_io.CoderToml[
				Public,
				*Public,
			]{
				Progenitor: func() Public {
					return &TomlV2Public{}
				},
			},
			ids.TypeTomlConfigImmutableV1: triple_hyphen_io.CoderToml[
				Public,
				*Public,
			]{
				Progenitor: func() Public {
					return &TomlV1Public{}
				},
			},
		},
	),
}
