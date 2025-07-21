package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	IndexObject interface {
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

		ReadManySha(
			sh interfaces.Digest,
		) (skus []*Transacted, err error)

		ReadOneSha(
			sh interfaces.Digest,
			sk *Transacted,
		) (err error)
	}
)
