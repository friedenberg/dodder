package ids

import (
	"io"
	"math"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ohio"
	"code.linenisgreat.com/dodder/go/src/charlie/catgut"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/genres"
)

var poolObjectId2 = pool.Make(
	nil,
	func(id *objectId2) {
		id.Reset()
	},
)

func getObjectIdPool2() interfaces.Pool[objectId2, *objectId2] {
	return poolObjectId2
}

type objectId2 struct {
	virtual     bool
	genre       genres.Genre
	middle      byte
	left, right catgut.String
	repoId      catgut.String
}

var _ Id = &objectId2{}

func (id *objectId2) GetObjectId() *objectId2 {
	return id
}

func (id *objectId2) Clone() (clone *objectId2) {
	clone = getObjectIdPool2().Get()
	clone.ResetWithObjectId(id)
	return clone
}

func (id *objectId2) WriteTo(w io.Writer) (n int64, err error) {
	if id.Len() > math.MaxUint8 {
		err = errors.ErrorWithStackf(
			"%q is greater than max uint8 (%d)",
			id.String(),
			math.MaxUint8,
		)

		return n, err
	}

	var n1 int64
	n1, err = id.genre.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	b := [2]uint8{uint8(id.Len()), uint8(id.left.Len())}

	var n2 int
	n2, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = id.left.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	bMid := [1]byte{id.middle}

	n2, err = ohio.WriteAllOrDieTrying(w, bMid[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = id.right.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (id *objectId2) ReadFrom(reader io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = id.genre.ReadFrom(reader)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	var b [2]uint8

	var n2 int
	n2, err = ohio.ReadAllOrDieTrying(reader, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	contentLength := b[0]
	middlePos := b[1]

	if middlePos > contentLength-1 {
		err = errors.ErrorWithStackf(
			"middle position %d is greater than last index: %d",
			middlePos,
			contentLength,
		)
		return n, err
	}

	if _, err = io.CopyN(&id.left, reader, int64(middlePos)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	var bMiddle [1]uint8

	n2, err = ohio.ReadAllOrDieTrying(reader, bMiddle[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	id.middle = bMiddle[0]

	if _, err = io.CopyN(&id.right, reader, int64(contentLength-middlePos-1)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (id *objectId2) SetGenre(genre interfaces.GenreGetter) {
	if genre == nil {
		id.genre = genres.Unknown
	} else {
		id.genre = genres.Must(genre.GetGenre())
	}

	if id.genre == genres.Zettel {
		id.middle = '/'
	}
}

func (id *objectId2) IsEmpty() bool {
	switch id.genre {
	case genres.Unknown:
		if id.left.String() == "/" && id.right.IsEmpty() {
			return true
		}

	case genres.Zettel, genres.Blob:
		if id.left.IsEmpty() && id.right.IsEmpty() {
			return true
		}
	}

	return id.left.Len() == 0 && id.middle == 0 &&
		id.right.Len() == 0
}

func (id *objectId2) Len() int {
	return id.left.Len() + 1 + id.right.Len()
}

// TODO perform validation
// func (id *objectId2) SetRepoId(v string) (err error) {
// 	if err = id.repoId.Set(v); err != nil {
// 		err = errors.Wrap(err)
// 		return err
// 	}

// 	return err
// }

func (id *objectId2) StringSansRepo() string {
	var sb strings.Builder

	switch id.genre {
	case genres.Zettel:
		sb.Write(id.left.Bytes())

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		sb.Write(id.right.Bytes())

	case genres.Type:
		sb.WriteRune('!')
		sb.Write(id.right.Bytes())

	default:
		if id.left.Len() > 0 {
			sb.Write(id.left.Bytes())
		}

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		if id.right.Len() > 0 {
			sb.Write(id.right.Bytes())
		}
	}

	return sb.String()
}

func (id *objectId2) StringSansOp() string {
	var sb strings.Builder

	if id.repoId.Len() > 0 {
		sb.WriteRune('/')
		id.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch id.genre {
	case genres.Zettel:
		sb.Write(id.left.Bytes())

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		sb.Write(id.right.Bytes())

	case genres.Type:
		sb.Write(id.right.Bytes())

	default:
		if id.left.Len() > 0 {
			sb.Write(id.left.Bytes())
		}

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		if id.right.Len() > 0 {
			sb.Write(id.right.Bytes())
		}
	}

	return sb.String()
}

func (id *objectId2) String() string {
	var sb strings.Builder

	if id.repoId.Len() > 0 {
		sb.WriteRune('/')
		id.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch id.genre {
	case genres.Zettel:
		sb.Write(id.left.Bytes())

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		sb.Write(id.right.Bytes())

	case genres.Type:
		sb.WriteRune(doddish.OpType.ToRune())
		sb.Write(id.right.Bytes())

	default:
		if id.left.Len() > 0 {
			sb.Write(id.left.Bytes())
		}

		if id.middle != '\x00' {
			sb.WriteByte(id.middle)
		}

		if id.right.Len() > 0 {
			sb.Write(id.right.Bytes())
		}
	}

	return sb.String()
}

func (id *objectId2) Reset() {
	id.genre = genres.Unknown
	id.left.Reset()
	id.middle = 0
	id.right.Reset()
	id.repoId.Reset()
}

func (id *objectId2) ToSeq() doddish.Seq {
	switch id.genre {
	case genres.Blob:
		return doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte{id.middle},
			},
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.right.Bytes(),
			},
		}

	case genres.Config:
		return doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.right.Bytes(),
			},
		}

	case genres.InventoryList:
		return doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.left.Bytes(),
			},
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte{id.middle},
			},
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.right.Bytes(),
			},
		}

	case genres.Repo:
		return doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte{id.middle},
			},
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.right.Bytes(),
			},
		}

	case genres.Tag:
		if id.virtual {
			if id.middle == '-' {
				return doddish.Seq{
					doddish.Token{
						Type:     doddish.TokenTypeOperator,
						Contents: []byte{'%'},
					},
					doddish.Token{
						Type:     doddish.TokenTypeOperator,
						Contents: []byte{'-'},
					},
					doddish.Token{
						Type:     doddish.TokenTypeIdentifier,
						Contents: id.right.Bytes(),
					},
				}
			} else {
				return doddish.Seq{
					doddish.Token{
						Type:     doddish.TokenTypeOperator,
						Contents: []byte{'%'},
					},
					doddish.Token{
						Type:     doddish.TokenTypeIdentifier,
						Contents: id.right.Bytes(),
					},
				}
			}
		} else {
			if id.middle == '-' {
				return doddish.Seq{
					doddish.Token{
						Type:     doddish.TokenTypeOperator,
						Contents: []byte{'-'},
					},
					doddish.Token{
						Type:     doddish.TokenTypeIdentifier,
						Contents: id.right.Bytes(),
					},
				}
			} else {
				return doddish.Seq{
					doddish.Token{
						Type:     doddish.TokenTypeIdentifier,
						Contents: id.right.Bytes(),
					},
				}
			}
		}

	case genres.Type:
		return doddish.Seq{
			doddish.Token{
				Type:     doddish.TokenTypeOperator,
				Contents: []byte{id.middle},
			},
			doddish.Token{
				Type:     doddish.TokenTypeIdentifier,
				Contents: id.right.Bytes(),
			},
		}

	case genres.Zettel:
		if id.IsEmpty() {
			return doddish.Seq{
				doddish.Token{
					Type:     doddish.TokenTypeOperator,
					Contents: doddish.OpPathSeparator.ToBytes(),
				},
			}
		} else {
			return doddish.Seq{
				doddish.Token{
					Type:     doddish.TokenTypeIdentifier,
					Contents: id.left.Bytes(),
				},
				doddish.Token{
					Type:     doddish.TokenTypeOperator,
					Contents: []byte{id.middle},
				},
				doddish.Token{
					Type:     doddish.TokenTypeIdentifier,
					Contents: id.right.Bytes(),
				},
			}
		}

	default:
		panic(genres.MakeErrUnsupportedGenre(id.genre))
	}
}

func (id *objectId2) GetGenre() interfaces.Genre {
	return id.genre
}

func (id *objectId2) SetWithId(
	otherObjectId Id,
) (err error) {
	id.Reset()

	switch kt := otherObjectId.(type) {
	case *objectId2:
		id.ResetWith(kt)
		return err

	default:
		seq := otherObjectId.ToSeq()

		if err = id.SetWithSeq(seq); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (id *objectId2) SetWithGenre(
	value string,
	genre interfaces.GenreGetter,
) (err error) {
	id.genre = genres.Make(genre.GetGenre())

	if err = id.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *objectId2) SetLeft(v string) (err error) {
	id.genre = genres.Zettel

	if err = id.left.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO switch to SetWithSeq
func (id *objectId2) Set(value string) (err error) {
	if value == "/" {
		id.genre = genres.Zettel
		return err
	}

	var parsedId Id

	switch id.genre {
	case genres.Unknown:
		if parsedId, err = MakeObjectId(value); err != nil {
			return id.SetBlob(value)
		}

	case genres.Zettel:
		var h ZettelId
		err = h.Set(value)
		parsedId = &h

	case genres.Tag:
		var h TagStruct
		err = h.Set(value)
		parsedId = &h

	case genres.Type:
		var h TypeStruct
		err = h.Set(value)
		parsedId = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(value)
		parsedId = &h

	case genres.Config:
		err = Config.Set(value)
		parsedId = Config

	case genres.InventoryList:
		var h Tai
		err = h.Set(value)
		parsedId = &h

	case genres.Blob:
		if err = id.left.Set(value); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err

	default:
		err = genres.MakeErrUnrecognizedGenre(id.genre.String())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", value)
		return err
	}

	if err = id.SetWithId(parsedId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *objectId2) ResetWith(other *objectId2) {
	id.genre = other.genre
	other.left.CopyTo(&id.left)
	other.right.CopyTo(&id.right)
	id.middle = other.middle

	if id.middle == '%' {
		id.virtual = true
	}

	other.repoId.CopyTo(&id.repoId)
}

func (id *objectId2) ResetWithObjectId(b Id) {
	errors.PanicIfError(id.SetWithId(b))
}

func (id *objectId2) MarshalText() (text []byte, err error) {
	text = []byte(FormattedString(id))
	return text, err
}

func (id *objectId2) UnmarshalText(text []byte) (err error) {
	if err = id.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *objectId2) MarshalBinary() (text []byte, err error) {
	return id.AppendBinary(nil)
}

func (id *objectId2) AppendBinary(text []byte) ([]byte, error) {
	return append(text, []byte(FormattedString(id))...), nil
}

func (id *objectId2) UnmarshalBinary(bites []byte) (err error) {
	text := string(bites)

	if err = id.Set(text); err != nil {
		err = errors.Wrapf(err, "Bytes: %x, Bytes as String: %q", bites, text)
		return err
	}

	return err
}

func (id *objectId2) SetWithSeq(
	seq doddish.Seq,
) (err error) {
	id.Reset()

	switch {
	case seq.Len() == 0:
		err = errors.Wrap(doddish.ErrEmptySeq)
		return err

	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
	):
		id.genre = genres.Zettel
		id.middle = '/'

		return err

		// tag
	case seq.MatchAll(doddish.TokenTypeIdentifier):
		id.genre = genres.Tag
		id.right.WriteLower(seq.At(0).Contents)

		if id.right.EqualsBytes(configBytes) {
			id.genre = genres.Config
		}

		return err

		// -tag
	case seq.MatchAll(
		doddish.TokenMatcherOp('-'),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Tag
		id.middle = '-'
		id.right.WriteLower(seq.At(1).Contents)

		if id.right.EqualsBytes(configBytes) {
			id.genre = genres.Config
		}

		return err

		// !type
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Type
		id.middle = doddish.OpType.ToByte()
		id.right.Write(seq.At(1).Contents)
		return err

		// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Tag
		id.middle = doddish.OpVirtual.ToByte()
		id.right.Write(seq.At(1).Contents)
		return err

		// /repo
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Repo
		id.middle = doddish.OpPathSeparator.ToByte()
		id.right.Write(seq.At(1).Contents)
		return err

		// @digest-or-sig
	case seq.MatchAll(
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Blob
		id.middle = '@'
		id.right.Write(seq.At(1).Contents)
		return err

		// purpose@digest-or-sig
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Blob
		id.left.Write(seq.At(0).Contents)
		id.middle = '@'
		id.right.Write(seq.At(2).Contents)
		return err

		// zettel/id
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		id.genre = genres.Zettel
		id.left.Write(seq.At(0).Contents)
		id.middle = doddish.OpPathSeparator.ToByte()
		id.right.Write(seq.At(2).Contents)
		return err

		// sec.asec
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp(doddish.OpSigilExternal),
		doddish.TokenTypeIdentifier,
	):
		var tai Tai

		// TODO don't use object for validation, validate directly
		if err = tai.Set(seq.String()); err != nil {
			err = errors.Wrap(doddish.ErrUnsupportedSeq{Seq: seq, For: "tai"})
			return err
		}

		id.genre = genres.InventoryList
		id.left.Write(seq.At(0).Contents)
		id.middle = doddish.OpSigilExternal.ToByte()
		id.right.Write(seq.At(2).Contents)

		return err

	default:
		err = doddish.ErrUnsupportedSeq{Seq: seq, For: "object id"}
		return err
	}
}

func (id *objectId2) SetBlob(value string) (err error) {
	id.genre = genres.Blob

	if err = id.left.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id *objectId2) ToType() TypeStruct {
	errors.PanicIfError(genres.Type.AssertGenre(id.genre))

	return TypeStruct{
		Value: id.right.String(),
	}
}
