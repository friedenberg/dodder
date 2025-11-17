package ids

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

// /repo/tag
// /repo/zettel-id
// /repo/!type
// /browser/one/uno
// /browser/bookmark-1
// /browser/!md
// /browser/!md
func (objectId *objectId2) ReadFromSeq(
	seq doddish.Seq,
) (err error) {
	switch {
	case seq.Len() == 0:
		err = errors.Wrap(doddish.ErrEmptySeq)
		return err

		// tag
	case seq.MatchAll(doddish.TokenTypeIdentifier):
		objectId.genre = genres.Tag
		objectId.right.WriteLower(seq.At(0).Contents)

		if objectId.right.EqualsBytes(configBytes) {
			objectId.genre = genres.Config
		}

		return err

		// !type
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpType), doddish.TokenTypeIdentifier):
		objectId.genre = genres.Type
		objectId.middle = doddish.OpType
		objectId.right.Write(seq.At(1).Contents)
		return err

		// %tag
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpVirtual), doddish.TokenTypeIdentifier):
		objectId.genre = genres.Tag
		objectId.middle = doddish.OpVirtual
		objectId.right.Write(seq.At(1).Contents)
		return err

		// /repo
	case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpPathSeparator), doddish.TokenTypeIdentifier):
		objectId.genre = genres.Repo
		objectId.middle = doddish.OpPathSeparator
		objectId.right.Write(seq.At(1).Contents)
		return err

		// @sha
	case seq.MatchAll(doddish.TokenMatcherOp('@'), doddish.TokenTypeIdentifier):
		objectId.genre = genres.Blob
		objectId.middle = '@'
		objectId.right.Write(seq.At(1).Contents)
		return err

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Zettel
		objectId.left.Write(seq.At(0).Contents)
		objectId.middle = doddish.OpPathSeparator
		objectId.right.Write(seq.At(2).Contents)
		return err

		// sec.asec
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		var tai Tai

		if err = tai.Set(seq.String()); err != nil {
			err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq})
			return err
		}

		objectId.genre = genres.InventoryList
		objectId.left.Write(seq.At(0).Contents)
		objectId.middle = doddish.OpSigilExternal
		objectId.right.Write(seq.At(2).Contents)
		return err

	default:
		err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq})
		return err
	}
}
