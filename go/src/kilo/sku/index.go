package sku

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	FuncReadOne = func(
		sh interfaces.MarklId,
		sk *Transacted,
	) (ok bool)

	ObjectProbeIndex interface {
		ReadOneObjectId(ids.ObjectIdLike, *Transacted) error
	}

	IndexPrimitives interface {
		ObjectExists(
			objectId *ids.ObjectId,
		) (err error)

		// ReadOneMarklId(
		// 	ctx interfaces.ActiveContext,
		// 	marklId interfaces.MarklId,
		// 	object *Transacted,
		// ) (ok bool)

		ReadOneMarklIdAdded(
			sh interfaces.MarklId,
			sk *Transacted,
		) (ok bool)

		ReadOneMarklId(
			sh interfaces.MarklId,
			sk *Transacted,
		) (ok bool)
	}

	Index interface {
		IndexPrimitives
		ObjectProbeIndex

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
	return ReadOneObjectIdBespoke(
		objectId,
		object,
		index.ReadOneMarklId,
	)
}

func ReadOneObjectIdBespoke(
	objectId interfaces.ObjectId,
	object *Transacted,
	funcs ...FuncReadOne,
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

	for _, funk := range funcs {
		if ok = funk(digest, object); ok {
			break
		}
	}

	return ok
}
