package ids

import (
	"fmt"
	"strconv"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/alfa/quiter_collection"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/comments"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

var poolObjectId3 = pool.Make(
	nil,
	func(id *objectId3) {
		id.Reset()
	},
)

func getObjectIdPool3() interfaces.Pool[objectId3, *objectId3] {
	return poolObjectId3
}

// TODO add support for binary marshaling
// TODO make fields private
type objectId3 struct {
	Genre genres.Genre
	Seq   doddish.Seq
}

var _ Id = objectId3{}

func (id *objectId3) GetObjectId() *objectId3 {
	return id
}

func (id objectId3) GetGenre() interfaces.Genre {
	return id.Genre
}

func (id objectId3) IsEmpty() bool {
	return id.Seq.Len() == 0
}

func (id objectId3) Len() int {
	var count int

	for _, token := range id.Seq {
		count += len(token.Contents)
	}

	return count
}

func (id objectId3) String() string {
	return id.Seq.String()
}

func (id objectId3) Equals(other objectId3) bool {
	if id.Genre != other.Genre {
		return false
	}

	if !quiter_collection.Equals(id.Seq.GetSlice(), other.Seq.GetSlice()) {
		return false
	}

	return true
}

func (id *objectId3) SetWithId(otherId Id) (err error) {
	return id.SetWithSeq(otherId.ToSeq())
}

func (id *objectId3) Set(value string) (err error) {
	var seq doddish.Seq

	if seq, err = doddish.ScanExactlyOneSeqWithDotAllowedInIdenfierFromString(
		value,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return id.SetWithSeq(seq)
}

func (id *objectId3) SetWithSeq(seq doddish.Seq) (err error) {
	if id.Genre, err = ValidateSeqAndGetGenre(seq); err != nil {
		err = errors.Wrap(err)
		return
	}

	id.Seq = seq.Clone()

	return
}

func ValidateSeqAndGetGenre(
	seq doddish.Seq,
) (genre genres.Genre, err error) {
	switch {
	case seq.Len() == 0:
		return genres.None, doddish.ErrEmptySeq

		// tag
	case seq.MatchAll(doddish.TokenTypeIdentifier):

		if TokenIsConfig(seq.At(0)) {
			return genres.Config, nil
		} else {
			return genres.Tag, nil
		}

		// -tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpTagSeparator),
		doddish.TokenTypeIdentifier,
	):

		if TokenIsConfig(seq.At(1)) {
			return genres.Tag, errors.Errorf("config not allowed for tag")
		}

		return genres.Tag, nil

	// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		if TokenIsConfig(seq.At(1)) {
			return genres.Tag, errors.Errorf("unsupported seq: %q", seq)
		}

		return genres.Tag, nil

		// !type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		return genres.Type, nil

		// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		return genres.Tag, nil

		// /repo
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		return genres.Repo, nil

		// @digest-or-sig
	case seq.MatchAll(
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		return genres.Blob, nil

		// purpose@digest-or-sig
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		return genres.Blob, nil

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		return genres.Zettel, nil

		// sec.asec
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		comments.Performance("remove []byte->string conversion")
		if _, err = strconv.ParseInt(seq.At(0).String(), 10, 64); err != nil {
			return genres.InventoryList, errors.Wrapf(err, "failed to parse Sec time: %s", seq.At(0))
		}

		comments.Performance("remove []byte->string conversion")
		if _, err = strconv.ParseInt(seq.At(2).String(), 10, 64); err != nil {
			return genres.InventoryList, errors.Wrapf(err, "failed to parse Asec time: %s", seq.At(2))
		}

		return genres.InventoryList, nil

	default:
		return genres.None, errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq, For: "objectId3"})
	}
}

func (id *objectId3) Reset() {
	id.Genre = genres.None
	id.Seq.GetSliceMutable().Reset()
}

func (id *objectId3) ResetWith(other objectId3) {
	id.Genre = other.Genre

	comments.Performance("switch to reusing exising seq's data")
	id.Seq = other.Seq.Clone()
}

func (id *objectId3) ResetWithObjectId(other Id) {
	switch other := other.(type) {
	case TypeStruct:
		id.ResetWithType(other)

	case *TypeStruct:
		id.ResetWithType(*other)

	default:
		id.Genre = genres.Must(other.GetGenre())
		id.Seq = other.ToSeq()
	}
}

func (id *objectId3) ResetWithType(other TypeStruct) {
	id.Genre = genres.Type
	id.Seq = doddish.Seq{
		doddish.Token{
			Type:     doddish.TokenTypeOperator,
			Contents: []byte("!"),
		},
		doddish.Token{
			Type:     doddish.TokenTypeIdentifier,
			Contents: []byte(other.StringSansOp()),
		},
	}
}

func (id objectId3) ToType() TypeStruct {
	if id.IsEmpty() {
		return TypeStruct{}
	} else if !id.Genre.IsType() {
		panic(fmt.Sprintf("id is not a type, genre is %q: %q", id.Genre, id))
	}

	return TypeStruct{Value: id.Seq.At(1).String()}
}

func SeqIdCompare(left, right objectId3) cmp.Result {
	return doddish.SeqCompare(left.Seq, right.Seq)
}

func (id *objectId3) SetType(value string) (err error) {
	reader, repool := pool.GetStringReader(value)
	defer repool()

	boxScanner := doddish.MakeScanner(reader)

	var seq doddish.Seq

	if seq, err = boxScanner.ScanDotAllowedInIdentifiersOrError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	switch {
	default:
		err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq, For: "objectId3.SetType"})
		return err

	case seq.Len() == 0:
		err = doddish.ErrEmptySeq
		return err

		// type
	case seq.MatchAll(doddish.TokenTypeIdentifier):
		if TokenIsConfig(seq.At(0)) {
			err = errors.Errorf("cannot use konfig identifier: %q", seq)
			return err
		}

		id.Genre = genres.Type
		seq = doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte("!"),
			},
			seq.At(0),
		}

		// .type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		if TokenIsConfig(seq.At(1)) {
			err = errors.Errorf("cannot use konfig identifier: %q", seq)
			return err
		}

		id.Genre = genres.Type
		seq = doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte("!"),
			},
			seq.At(0),
		}

		// !type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		if TokenIsConfig(seq.At(1)) {
			err = errors.Errorf("cannot use konfig identifier: %q", seq)
			return err
		}

		id.Genre = genres.Type
	}

	id.Seq = seq.Clone()

	return err
}

func (id objectId3) ToSeq() doddish.Seq {
	return id.Seq
}

func (id *objectId3) ToSeqMutable() *doddish.Seq {
	return &id.Seq
}

func (id objectId3) MarshalBinary() ([]byte, error) {
	return id.AppendBinary(nil)
}

func (id objectId3) AppendBinary(bites []byte) ([]byte, error) {
	return id.ToSeq().GetBinaryMarshaler().AppendBinary(bites)
}

func (id *objectId3) UnmarshalBinary(bites []byte) (err error) {
	if err = id.ToSeqMutable().GetBinaryUnmarshaler().UnmarshalBinary(
		bites,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if id.Genre, err = ValidateSeqAndGetGenre(id.Seq); err != nil {
		if err == doddish.ErrEmptySeq {
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	return nil
}
