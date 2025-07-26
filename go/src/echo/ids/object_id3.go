package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

// /repo/tag
// /repo/zettel-id
// /repo/!type
// /browser/one/uno
// /browser/bookmark-1
// /browser/!md
// /browser/!md
func (oid *objectId2) ReadFromSeq(
	seq doddish.Seq,
) (err error) {
	switch {
	case seq.Len() == 0:
		err = errors.ErrorWithStackf("empty seq")
		return

		// tag
	case seq.MatchAll(doddish.TokenTypeIdentifier):
		oid.g = genres.Tag
		oid.right.WriteLower(seq.At(0).Contents)

		if oid.right.EqualsBytes(configBytes) {
			oid.g = genres.Config
		}

		return

		// !type
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpType), doddish.TokenTypeIdentifier):
		oid.g = genres.Type
		oid.middle = doddish.OpType
		oid.right.Write(seq.At(1).Contents)
		return

		// %tag
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpVirtual), doddish.TokenTypeIdentifier):
		oid.g = genres.Tag
		oid.middle = doddish.OpVirtual
		oid.right.Write(seq.At(1).Contents)
		return

		// /repo
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpPathSeparator), doddish.TokenTypeIdentifier):
		oid.g = genres.Repo
		oid.middle = doddish.OpPathSeparator
		oid.right.Write(seq.At(1).Contents)
		return

		// @sha
	case seq.MatchAll(doddish.TokenMatcherOp('@'), doddish.TokenTypeIdentifier):
		oid.g = genres.Blob
		oid.middle = '@'
		oid.right.Write(seq.At(1).Contents)
		return

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		oid.g = genres.Zettel
		oid.left.Write(seq.At(0).Contents)
		oid.middle = doddish.OpPathSeparator
		oid.right.Write(seq.At(2).Contents)
		return

		// sec.asec
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		var t Tai

		if err = t.Set(seq.String()); err != nil {
			err = errors.ErrorWithStackf("unsupported seq: %q, %#v", seq, seq)
			return
		}

		oid.g = genres.InventoryList
		oid.left.Write(seq.At(0).Contents)
		oid.middle = doddish.OpSigilExternal
		oid.right.Write(seq.At(2).Contents)
		return

	default:
		err = errors.ErrorWithStackf("unsupported seq: %q, %#v", seq, seq)
		return
	}
}
