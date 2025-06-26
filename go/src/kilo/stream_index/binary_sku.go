package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type skuWithSigil struct {
	*sku.Transacted
	ids.Sigil
}

type skuWithRangeAndSigil struct {
	skuWithSigil
	object_probe_index.Range
}
