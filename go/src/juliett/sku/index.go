package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	Index interface {
		ReadOneObjectIdTai(
			k interfaces.ObjectId,
			t ids.Tai,
		) (sk *Transacted, err error)

		ReadManyObjectId(
			id interfaces.ObjectId,
		) (skus []*Transacted, err error)

		ReadOneObjectId(
			oid interfaces.ObjectId,
			sk *Transacted,
		) (err error)

		ObjectExists(
			id *ids.ObjectId,
		) (err error)

		ReadManyMarklId(
			sh interfaces.MarklId,
		) (skus []*Transacted, err error)

		ReadOneMarklId(
			sh interfaces.MarklId,
			sk *Transacted,
		) (err error)
	}
)
