package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
)

var (
	poolTransacted = pool.Make(
		nil,
		TransactedResetter.Reset,
	)

	poolCheckedOut = pool.Make(
		nil,
		CheckedOutResetter.Reset,
	)
)

func GetTransactedPool() interfaces.Pool[Transacted, *Transacted] {
	return poolTransacted
}

func GetCheckedOutPool() interfaces.Pool[CheckedOut, *CheckedOut] {
	return poolCheckedOut
}
