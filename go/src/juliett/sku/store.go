package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	IdAbbrIndexPresenceGeneric[_ any] interface {
		Exists([3]string) error
	}

	IdAbbrIndexGeneric[ID any, ID_PTR interfaces.Ptr[ID]] interface {
		IdAbbrIndexPresenceGeneric[ID]
		ExpandStringString(string) (string, error)
		ExpandString(string) (ID_PTR, error)
		Expand(ID_PTR) (ID_PTR, error)
		Abbreviate(ids.Abbreviatable) (string, error)
	}

	IdIndex interface {
		GetSeenIds() map[genres.Genre]interfaces.Collection[string]
		GetZettelIds() IdAbbrIndexGeneric[ids.ZettelId, *ids.ZettelId]
		GetBlobIds() IdAbbrIndexGeneric[markl.Id, *markl.Id]

		AddObjectToIdIndex(*Transacted) error
		GetAbbr() ids.Abbr

		errors.Flusher
	}

	// TODO rename to RepoStore
	RepoStore interface {
		Commit(ExternalLike, CommitOptions) (err error)
		ReadOneInto(interfaces.ObjectId, *Transacted) (err error)
		ReadPrimitiveQuery(
			qg PrimitiveQueryGroup,
			w interfaces.FuncIter[*Transacted],
		) (err error)
	}

	ExternalObjectId       = ids.ExternalObjectIdLike
	ExternalObjectIdGetter = ids.ExternalObjectIdGetter

	FuncReadOneInto = func(
		k1 interfaces.ObjectId,
		out *Transacted,
	) (err error)

	ExternalStoreUpdateTransacted interface {
		UpdateTransacted(z *Transacted) (err error)
	}

	ExternalStoreReadExternalLikeFromObjectIdLike interface {
		ReadExternalLikeFromObjectIdLike(
			o CommitOptions,
			oid interfaces.Stringer,
			t *Transacted,
		) (e ExternalLike, err error)
	}
)
