package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	IndexPrimitives interface {
		ObjectExists(
			objectId *ids.ObjectId,
		) (err error)

		// ReadOneMarklId(
		// 	ctx interfaces.ActiveContext,
		// 	marklId interfaces.MarklId,
		// 	object *Transacted,
		// ) (ok bool)

		ReadOneMarklId(
			sh interfaces.MarklId,
			sk *Transacted,
		) (err error)
	}

	Index interface {
		IndexPrimitives

		ReadOneObjectId(
			objectId interfaces.ObjectId,
			object *Transacted,
		) (err error)

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

func ReadOneObjectId(
	index IndexPrimitives,
	objectId interfaces.ObjectId,
	object *Transacted,
) (ok bool) {
	objectIdString := objectId.String()

	if objectIdString == "" {
		panic("empty object id")
	}

	// TODO don't hardcode hash format
	digest, repool := markl.FormatHashSha256.GetMarklIdForString(
		objectIdString,
	)
	defer repool()

	defer func() {
		r := recover()

		if r == nil {
			return
		}

		err, isErr := r.(error)

		if !isErr {
			panic(r)
		}

		if errors.IsNotExist(err) || collections.IsErrNotFound(err) {
			ok = false
		} else {
			panic(err)
		}
	}()

	if err := index.ReadOneMarklId(digest, object); err != nil {
		panic(err)
	} else {
		ok = true
	}

	return
}
