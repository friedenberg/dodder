package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

var (
	poolExternal   interfaces.Pool[sku.Transacted, *sku.Transacted]
	poolCheckedOut interfaces.Pool[sku.CheckedOut, *sku.CheckedOut]
)

func init() {
	poolExternal = pool.Make[sku.Transacted](
		nil,
		nil,
	)

	poolCheckedOut = pool.Make[sku.CheckedOut](
		nil,
		nil,
	)
}

func GetExternalPool() interfaces.Pool[sku.Transacted, *sku.Transacted] {
	return poolExternal
}

func GetCheckedOutPool() interfaces.Pool[sku.CheckedOut, *sku.CheckedOut] {
	return poolCheckedOut
}
