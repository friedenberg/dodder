package inventory_list_blobs

import (
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type funcListFormatConstructor func(
	env_repo.Env,
	*box_format.BoxTransacted,
) sku.ListFormat

var coderConstructors = map[string]funcListFormatConstructor{
	ids.TypeInventoryListV1: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		if box == nil {
			panic("empty box")
		}

		return V1{
			V1ObjectCoder: V1ObjectCoder{
				Box: box,
			},
		}
	},
	ids.TypeInventoryListV2: func(
		envRepo env_repo.Env,
		box *box_format.BoxTransacted,
	) sku.ListFormat {
		if box == nil {
			panic("empty box")
		}

		return V2{
			V2ObjectCoder: V2ObjectCoder{
				Box:                    box,
				ImmutableConfigPrivate: envRepo.GetConfigPrivate().Blob,
			},
		}
	},
}
