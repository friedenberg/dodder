package genesis_configs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

var CoderPrivate = triple_hyphen_io.CoderToTypedBlob[BlobPrivate]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[BlobPrivate]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[BlobPrivate](
		map[string]interfaces.CoderBufferedReadWriter[*BlobPrivate]{
			ids.TypeTomlConfigImmutableV2: triple_hyphen_io.CoderToml[
				BlobPrivate,
				*BlobPrivate,
			]{
				Progenitor: func() BlobPrivate {
					return &TomlV2Private{}
				},
			},
			ids.TypeTomlConfigImmutableV1: triple_hyphen_io.CoderToml[
				BlobPrivate,
				*BlobPrivate,
			]{
				Progenitor: func() BlobPrivate {
					return &TomlV1Private{}
				},
			},
		},
	),
}

var CoderPublic = triple_hyphen_io.CoderToTypedBlob[BlobPublic]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[BlobPublic]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[BlobPublic](
		map[string]interfaces.CoderBufferedReadWriter[*BlobPublic]{
			ids.TypeTomlConfigImmutableV2: triple_hyphen_io.CoderToml[
				BlobPublic,
				*BlobPublic,
			]{
				Progenitor: func() BlobPublic {
					return &TomlV2Public{}
				},
			},
			ids.TypeTomlConfigImmutableV1: triple_hyphen_io.CoderToml[
				BlobPublic,
				*BlobPublic,
			]{
				Progenitor: func() BlobPublic {
					return &TomlV1Public{}
				},
			},
		},
	),
}
