package sku

import "code.linenisgreat.com/dodder/go/src/_/interfaces"

type UIStorePrinters struct {
	TransactedNew       interfaces.FuncIter[*Transacted]
	TransactedUpdated   interfaces.FuncIter[*Transacted]
	TransactedUnchanged interfaces.FuncIter[*Transacted]

	CheckedOut interfaces.FuncIter[SkuType] // for when objects are checked out
}
