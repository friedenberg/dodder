package genesis_config_io

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_config"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
)

type PrivateTypedBlob = triple_hyphen_io2.TypedBlob[genesis_config.Private]

var typedCodersPrivate = map[string]interfaces.CoderBufferedReadWriter[*genesis_config.Private]{
	builtin_types.ImmutableConfigV1: blobV1CoderPrivate{},
	"":                              blobV0CoderPrivate{},
}

var CoderPrivate = triple_hyphen_io2.CoderToTypedBlob[genesis_config.Private]{
	Metadata: triple_hyphen_io2.TypedMetadataCoder[genesis_config.Private]{},
	Blob: triple_hyphen_io2.CoderTypeMapWithoutType[genesis_config.Private](
		typedCodersPrivate,
	),
}
