package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type objectWithSigil struct {
	*sku.Transacted
	ids.Sigil
}

type objectWithCursorAndSigil struct {
	objectWithSigil
	object_probe_index.Cursor
}

type objectMetaWithCursorAndSigil struct {
	ids.Sigil
	object_probe_index.Cursor
	Tai ids.Tai
}
