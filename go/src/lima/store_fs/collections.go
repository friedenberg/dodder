package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type (
	CheckedOutSet        = interfaces.SetLike[*sku.CheckedOut]
	CheckedOutMutableSet = interfaces.MutableSetLike[*sku.CheckedOut]
)

func MakeCheckedOutMutableSet() CheckedOutMutableSet {
	return collections_value.MakeMutableValueSet[*sku.CheckedOut](
		nil,
	)
}
