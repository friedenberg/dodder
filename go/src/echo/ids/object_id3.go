package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/box"
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
	seq box.Seq,
) (err error) {
	switch {
	case seq.Len() == 0:
		err = errors.ErrorWithStackf("empty seq")
		return

		// tag
	case seq.MatchAll(box.TokenTypeIdentifier):
		oid.g = genres.Tag
		oid.right.WriteLower(seq.At(0).Contents)

		if oid.right.EqualsBytes(configBytes) {
			oid.g = genres.Config
		}

		return

		// !type
	case seq.MatchAll(box.TokenMatcherOp(box.OpType), box.TokenTypeIdentifier):
		oid.g = genres.Type
		oid.middle = box.OpType
		oid.right.Write(seq.At(1).Contents)
		return

		// %tag
	case seq.MatchAll(box.TokenMatcherOp(box.OpVirtual), box.TokenTypeIdentifier):
		oid.g = genres.Tag
		oid.middle = box.OpVirtual
		oid.right.Write(seq.At(1).Contents)
		return

		// /repo
	case seq.MatchAll(box.TokenMatcherOp(box.OpPathSeparator), box.TokenTypeIdentifier):
		oid.g = genres.Repo
		oid.middle = box.OpPathSeparator
		oid.right.Write(seq.At(1).Contents)
		return

		// @sha
	case seq.MatchAll(box.TokenMatcherOp('@'), box.TokenTypeIdentifier):
		oid.g = genres.Blob
		oid.middle = '@'
		oid.right.Write(seq.At(1).Contents)
		return

		// zettel/id
	case seq.MatchAll(
		box.TokenTypeIdentifier,
		box.TokenMatcherOp(box.OpPathSeparator),
		box.TokenTypeIdentifier,
	):
		oid.g = genres.Zettel
		oid.left.Write(seq.At(0).Contents)
		oid.middle = box.OpPathSeparator
		oid.right.Write(seq.At(2).Contents)
		return

		// sec.asec
	case seq.MatchAll(
		box.TokenTypeIdentifier,
		box.TokenMatcherOp(box.OpSigilExternal),
		box.TokenTypeIdentifier,
	):
		var t Tai

		if err = t.Set(seq.String()); err != nil {
			err = errors.ErrorWithStackf("unsupported seq: %q, %#v", seq, seq)
			return
		}

		oid.g = genres.InventoryList
		oid.left.Write(seq.At(0).Contents)
		oid.middle = box.OpSigilExternal
		oid.right.Write(seq.At(2).Contents)
		return

	default:
		err = errors.ErrorWithStackf("unsupported seq: %q, %#v", seq, seq)
		return
	}
}
