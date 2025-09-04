package sku

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
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

	BlobCopyResult struct {
		*Transacted        // may be nil
		interfaces.MarklId // may not be nil

		// -1: no remote blob store and the blob doesn't exist locally
		// -2: no remote blob store and the blob exists locally
		// -3: blob exists locally and remotely
		N int64
	}

	ImporterOptions struct {
		DedupingFormatId    string
		OverwriteSignatures bool
		BlobGenres          ids.Genre
		ExcludeObjects      bool
		RemoteBlobStore     interfaces.BlobStore
		PrintCopies         bool
		AllowMergeConflicts bool
		BlobCopierDelegate  interfaces.FuncIter[BlobCopyResult]
		ParentNegotiator    ParentNegotiator
		CheckedOutPrinter   interfaces.FuncIter[*CheckedOut]
	}

	Importer interface {
		GetCheckedOutPrinter() interfaces.FuncIter[*CheckedOut]

		SetCheckedOutPrinter(
			p interfaces.FuncIter[*CheckedOut],
		)

		ImportBlobIfNecessary(
			sk *Transacted,
		) (err error)

		Import(
			external *Transacted,
		) (co *CheckedOut, err error)
	}
)

func MakeBlobCopierDelegate(fd fd.Std) func(BlobCopyResult) error {
	return func(result BlobCopyResult) error {
		switch result.N {
		case -3:
			return fd.Printf(
				"Blob %s already exists",
				result.MarklId,
			)

		default:
			return fd.Printf(
				"copied Blob %s (%s)",
				result.MarklId,
				ui.GetHumanBytesString(uint64(result.N)),
			)
		}
	}
}
