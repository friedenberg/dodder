package sku

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type UIStorePrinters struct {
	TransactedNew       interfaces.FuncIter[*Transacted]
	TransactedUpdated   interfaces.FuncIter[*Transacted]
	TransactedUnchanged interfaces.FuncIter[*Transacted]

	CheckedOutCheckedOut interfaces.FuncIter[SkuType]
	CheckedOutChanged    interfaces.FuncIter[SkuType]
}
