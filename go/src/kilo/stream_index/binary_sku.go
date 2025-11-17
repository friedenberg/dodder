package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type objectWithSigil struct {
	*sku.Transacted
	ids.Sigil
}

type objectWithCursorAndSigil struct {
	objectWithSigil
	ohio.Cursor
}

type objectMetaWithCursorAndSigil struct {
	ids.Sigil
	ohio.Cursor
	Tai ids.Tai
}
