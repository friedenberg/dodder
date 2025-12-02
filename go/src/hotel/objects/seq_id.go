package objects

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/alfa/quiter_collection"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/comments"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type SeqId struct {
	genre genres.Genre
	seq   doddish.Seq
}

var _ interfaces.ObjectId = SeqId{}

func (id SeqId) GetGenre() interfaces.Genre {
	return id.genre
}

func (id SeqId) IsEmpty() bool {
	return id.seq.Len() == 0
}

func (id SeqId) String() string {
	return id.seq.String()
}

func (id SeqId) Equals(other SeqId) bool {
	if id.genre != other.genre {
		return false
	}

	if !quiter_collection.Equals[doddish.Token](
		id.seq.GetSlice(),
		other.seq.GetSlice(),
	) {
		return false
	}

	return true
}

func (id *SeqId) Set(value string) (err error) {
	reader, repool := pool.GetStringReader(value)
	defer repool()

	boxScanner := doddish.MakeScanner(reader)

	var seq doddish.Seq

	if seq, err = boxScanner.ScanDotAllowedInIdentifiersOrError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	switch {
	case seq.Len() == 0:
		err = doddish.ErrEmptySeq
		return err

		// tag
	case seq.MatchAll(doddish.TokenTypeIdentifier):

		if ids.TokenIsConfig(seq.At(0)) {
			id.genre = genres.Config
		} else {
			id.genre = genres.Tag
		}

		// !type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Type

		// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Tag

		// /repo
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Repo

		// @digest-or-sig
	case seq.MatchAll(
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Blob

		// purpose@digest-or-sig
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Blob

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Zettel

		// sec.asec
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		comments.Performance("remove []byte->string conversion")
		if _, err = strconv.ParseInt(seq.At(0).String(), 10, 64); err != nil {
			err = errors.Wrapf(err, "failed to parse Sec time: %s", seq.At(0))
			return err
		}

		comments.Performance("remove []byte->string conversion")
		if _, err = strconv.ParseInt(seq.At(2).String(), 10, 64); err != nil {
			err = errors.Wrapf(err, "failed to parse Asec time: %s", seq.At(2))
			return err
		}

		id.genre = genres.InventoryList

	default:
		err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq})
		return err
	}

	id.seq = seq.Clone()

	return err
}

func (id *SeqId) Reset() {
	id.genre = genres.None
	id.seq.GetSliceMutable().Reset()
}

func (id *SeqId) ResetWith(other SeqId) {
	id.genre = other.genre

	comments.Performance("switch to reusing exising seq's data")
	id.seq = other.seq.Clone()
}
