package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type (
	IndexPrimitives interface {
		ObjectExists(
			objectId *ids.ObjectId,
		) (err error)

		ReadOneObjectId(
			objectId interfaces.ObjectId,
			object *Transacted,
		) (err error)
	}

	Index interface {
		IndexPrimitives

		ReadOneObjectIdTai(
			k interfaces.ObjectId,
			t ids.Tai,
		) (sk *Transacted, err error)

		ReadManyObjectId(
			id interfaces.ObjectId,
		) (skus []*Transacted, err error)

		ReadManyMarklId(
			sh interfaces.MarklId,
		) (skus []*Transacted, err error)

		ReadOneMarklId(
			sh interfaces.MarklId,
			sk *Transacted,
		) (err error)
	}

	IndexMutation interface {
		Add(
			object *Transacted,
			options CommitOptions,
		) (err error)
	}

	IndexMutable interface {
		Index
		IndexMutation
	}

	Reindexer interface {
		IndexPrimitives
		IndexMutation
	}
)
