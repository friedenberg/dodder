package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type (
	Seq = interfaces.SeqError[*Transacted]

	InventoryListStore interface {
		WriteInventoryListObject(*Transacted) (err error)
		ReadLast() (max *Transacted, err error)
		AllInventoryListContents(interfaces.MarklId) Seq
		AllInventoryLists() Seq
	}
)
