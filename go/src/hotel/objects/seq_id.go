package objects

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/alfa/quiter_collection"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/comments"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type SeqId struct {
	Genre genres.Genre
	Seq   doddish.Seq
}

var _ interfaces.ObjectId = SeqId{}

func (id SeqId) GetGenre() interfaces.Genre {
	return id.Genre
}

func (id SeqId) IsEmpty() bool {
	return id.Seq.Len() == 0
}

func (id SeqId) String() string {
	return id.Seq.String()
}

func (id SeqId) Equals(other SeqId) bool {
	if id.Genre != other.Genre {
		return false
	}

	if !quiter_collection.Equals(id.Seq.GetSlice(), other.Seq.GetSlice()) {
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
			id.Genre = genres.Config
		} else {
			id.Genre = genres.Tag
		}

		// !type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Type

		// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Tag

		// /repo
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Repo

		// @digest-or-sig
	case seq.MatchAll(
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Blob

		// purpose@digest-or-sig
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Blob

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.Genre = genres.Zettel

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

		id.Genre = genres.InventoryList

	default:
		err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq})
		return err
	}

	id.Seq = seq.Clone()

	return err
}

func (id *SeqId) Reset() {
	id.Genre = genres.None
	id.Seq.GetSliceMutable().Reset()
}

func (id *SeqId) ResetWith(other SeqId) {
	id.Genre = other.Genre

	comments.Performance("switch to reusing exising seq's data")
	id.Seq = other.Seq.Clone()
}

func SeqIdCompare(left, right SeqId) cmp.Result {
	return doddish.SeqCompare(left.Seq, right.Seq)
}
