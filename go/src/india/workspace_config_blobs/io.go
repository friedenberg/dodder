package workspace_config_blobs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
)

var Coder = triple_hyphen_io.CoderToTypedBlob[Config]{
	Metadata: triple_hyphen_io.TypedMetadataCoder[Config]{},
	Blob: triple_hyphen_io.CoderTypeMapWithoutType[Config](
		map[string]interfaces.CoderBufferedReadWriter[*Config]{
			ids.TypeTomlWorkspaceConfigV0: triple_hyphen_io.CoderToml[
				Config,
				*Config,
			]{
				Progenitor: func() Config {
					return &V0{}
				},
			},
		},
	),
}
