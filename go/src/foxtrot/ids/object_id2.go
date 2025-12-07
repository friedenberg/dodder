package ids

import (
	"io"
	"math"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/doddish"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
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

// TODO replace with struct similar to markl.Id
type objectId2 struct {
	virtual     bool
	genre       genres.Genre
	middle      byte // remove and replace with virtual
	left, right catgut.String
	repoId      catgut.String
	// Domain
}

var _ Id = &objectId2{}

func (objectId *objectId2) GetObjectId() *objectId2 {
	return objectId
}

func (objectId *objectId2) Clone() (clone *objectId2) {
	clone = getObjectIdPool2().Get()
	clone.ResetWithIdLike(objectId)
	return clone
}

func (objectId *objectId2) WriteTo(w io.Writer) (n int64, err error) {
	if objectId.Len() > math.MaxUint8 {
		err = errors.ErrorWithStackf(
			"%q is greater than max uint8 (%d)",
			objectId.String(),
			math.MaxUint8,
		)

		return n, err
	}

	var n1 int64
	n1, err = objectId.genre.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	b := [2]uint8{uint8(objectId.Len()), uint8(objectId.left.Len())}

	var n2 int
	n2, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = objectId.left.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	bMid := [1]byte{objectId.middle}

	n2, err = ohio.WriteAllOrDieTrying(w, bMid[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = objectId.right.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (objectId *objectId2) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = objectId.genre.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	var b [2]uint8

	var n2 int
	n2, err = ohio.ReadAllOrDieTrying(r, b[:])
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

	if _, err = objectId.left.ReadNFrom(r, int(middlePos)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	var bMiddle [1]uint8

	n2, err = ohio.ReadAllOrDieTrying(r, bMiddle[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	objectId.middle = bMiddle[0]

	if _, err = objectId.right.ReadNFrom(r, int(contentLength-middlePos-1)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (objectId *objectId2) SetGenre(g interfaces.GenreGetter) {
	if g == nil {
		objectId.genre = genres.None
	} else {
		objectId.genre = genres.Must(g.GetGenre())
	}

	if objectId.genre == genres.Zettel {
		objectId.middle = '/'
	}
}

func (objectId *objectId2) IsEmpty() bool {
	switch objectId.genre {
	case genres.None:
		if objectId.left.String() == "/" && objectId.right.IsEmpty() {
			return true
		}

	case genres.Zettel, genres.Blob:
		if objectId.left.IsEmpty() && objectId.right.IsEmpty() {
			return true
		}
	}

	return objectId.left.Len() == 0 && objectId.middle == 0 &&
		objectId.right.Len() == 0
}

func (objectId *objectId2) Len() int {
	return objectId.left.Len() + 1 + objectId.right.Len()
}

func (objectId *objectId2) LenHeadAndTail() (int, int) {
	return objectId.left.Len(), objectId.right.Len()
}

// TODO perform validation
func (objectId *objectId2) SetRepoId(v string) (err error) {
	if err = objectId.repoId.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) StringSansRepo() string {
	var sb strings.Builder

	switch objectId.genre {
	case genres.Zettel:
		sb.Write(objectId.left.Bytes())

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		sb.Write(objectId.right.Bytes())

	case genres.Type:
		sb.WriteRune('!')
		sb.Write(objectId.right.Bytes())

	default:
		if objectId.left.Len() > 0 {
			sb.Write(objectId.left.Bytes())
		}

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		if objectId.right.Len() > 0 {
			sb.Write(objectId.right.Bytes())
		}
	}

	return sb.String()
}

func (objectId *objectId2) StringSansOp() string {
	var sb strings.Builder

	if objectId.repoId.Len() > 0 {
		sb.WriteRune('/')
		objectId.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch objectId.genre {
	case genres.Zettel:
		sb.Write(objectId.left.Bytes())

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		sb.Write(objectId.right.Bytes())

	case genres.Type:
		sb.Write(objectId.right.Bytes())

	default:
		if objectId.left.Len() > 0 {
			sb.Write(objectId.left.Bytes())
		}

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		if objectId.right.Len() > 0 {
			sb.Write(objectId.right.Bytes())
		}
	}

	return sb.String()
}

func (objectId *objectId2) String() string {
	var sb strings.Builder

	if objectId.repoId.Len() > 0 {
		sb.WriteRune('/')
		objectId.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch objectId.genre {
	case genres.Zettel:
		sb.Write(objectId.left.Bytes())

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		sb.Write(objectId.right.Bytes())

	case genres.Type:
		sb.WriteByte(doddish.OpType)
		sb.Write(objectId.right.Bytes())

	default:
		if objectId.left.Len() > 0 {
			sb.Write(objectId.left.Bytes())
		}

		if objectId.middle != '\x00' {
			sb.WriteByte(objectId.middle)
		}

		if objectId.right.Len() > 0 {
			sb.Write(objectId.right.Bytes())
		}
	}

	return sb.String()
}

func (objectId *objectId2) Reset() {
	objectId.genre = genres.None
	objectId.left.Reset()
	objectId.middle = 0
	objectId.right.Reset()
	objectId.repoId.Reset()
}

func (objectId *objectId2) ToSeq() doddish.Seq {
	switch objectId.genre {
	case genres.Blob:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{objectId.middle},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	case genres.Config:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	case genres.InventoryList:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.left.Bytes(),
			},
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{objectId.middle},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	case genres.Repo:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{objectId.middle},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	case genres.Tag:
		if objectId.virtual {
			return doddish.Seq{
				doddish.Token{
					TokenType: doddish.TokenTypeOperator,
					Contents:  []byte{'%'},
				},
				doddish.Token{
					TokenType: doddish.TokenTypeIdentifier,
					Contents:  objectId.right.Bytes(),
				},
			}
		} else {
			return doddish.Seq{
				doddish.Token{
					TokenType: doddish.TokenTypeIdentifier,
					Contents:  objectId.right.Bytes(),
				},
			}
		}

	case genres.Type:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{objectId.middle},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	case genres.Zettel:
		return doddish.Seq{
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.left.Bytes(),
			},
			doddish.Token{
				TokenType: doddish.TokenTypeOperator,
				Contents:  []byte{objectId.middle},
			},
			doddish.Token{
				TokenType: doddish.TokenTypeIdentifier,
				Contents:  objectId.right.Bytes(),
			},
		}

	default:
		panic(genres.MakeErrUnsupportedGenre(objectId.genre))
	}
}

func (objectId *objectId2) Parts() [3]string {
	var mid string

	if objectId.middle != 0 {
		mid = string([]byte{objectId.middle})
	}

	return [3]string{
		objectId.left.String(),
		mid,
		objectId.right.String(),
	}
}

func (objectId *objectId2) GetGenre() interfaces.Genre {
	return objectId.genre
}

func (objectId *objectId2) Expand(
	a Abbr,
) (err error) {
	ex := a.ExpanderFor(objectId.genre)

	if ex == nil {
		return err
	}

	v := objectId.String()

	if v, err = ex(v); err != nil {
		err = nil
		return err
	}

	if err = objectId.SetWithGenre(v, objectId.genre); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) Abbreviate(
	a Abbr,
) (err error) {
	return err
}

func (objectId *objectId2) SetWithIdLike(
	otherObjectId IdWithParts,
) (err error) {
	objectId.Reset()

	switch kt := otherObjectId.(type) {
	case *objectId2:
		objectId.ResetWith(kt)
		return err

	default:
		parts := otherObjectId.Parts()

		if err = objectId.left.Set(parts[0]); err != nil {
			err = errors.Wrap(err)
			return err
		}

		mid := []byte(parts[1])

		if len(mid) >= 1 {
			objectId.middle = mid[0]

			if objectId.middle == '%' {
				objectId.virtual = true
			}
		}

		if err = objectId.right.Set(parts[2]); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	objectId.SetGenre(otherObjectId)

	return err
}

func (objectId *objectId2) SetWithGenre(
	v string,
	g interfaces.GenreGetter,
) (err error) {
	objectId.genre = genres.Make(g.GetGenre())

	if err = objectId.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) SetLeft(v string) (err error) {
	objectId.genre = genres.Zettel

	if err = objectId.left.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) Set(v string) (err error) {
	if v == "/" {
		objectId.genre = genres.Zettel
		return err
	}

	var els []string

	if len(v) > 0 {
		els = strings.SplitAfterN(v[1:], "/", 2)
	}

	if strings.HasPrefix(v, "/") && len(els) == 2 {
		if err = objectId.SetRepoId(strings.TrimSuffix(els[0], "/")); err != nil {
			err = errors.Wrap(err)
			return err
		}

		v = els[1]
	}

	var k IdWithParts

	switch objectId.genre {
	case genres.None:
		if k, err = MakeObjectId(v); err != nil {
			objectId.genre = genres.Blob

			if err = objectId.left.Set(v); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		}

	case genres.Zettel:
		var h ZettelId
		err = h.Set(v)
		k = &h

	case genres.Tag:
		var h TagStruct
		err = h.Set(v)
		k = &h

	case genres.Type:
		var h TypeStruct
		err = h.Set(v)
		k = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(v)
		k = &h

	case genres.Config:
		err = Config.Set(v)
		k = Config

	case genres.InventoryList:
		var h Tai
		err = h.Set(v)
		k = &h

	case genres.Blob:
		if err = objectId.left.Set(v); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err

	default:
		err = genres.MakeErrUnrecognizedGenre(objectId.genre.String())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", v)
		return err
	}

	if err = objectId.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) SetOnlyNotUnknownGenre(v string) (err error) {
	if v == "/" {
		objectId.genre = genres.Zettel
		return err
	}

	if strings.HasPrefix(v, "/") {
		els := strings.SplitAfterN(v[1:], "/", 2)

		if len(els) != 2 {
			err = errors.ErrorWithStackf("invalid object id format: %q", v)
			return err
		}

		v = els[1]

		if err = objectId.SetRepoId(strings.TrimSuffix(els[0], "/")); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	var k IdWithParts

	switch objectId.genre {
	case genres.Zettel:
		var h ZettelId
		err = h.Set(v)
		k = &h

	case genres.Tag:
		var h TagStruct
		err = h.Set(v)
		k = &h

	case genres.Type:
		var h TypeStruct
		err = h.Set(v)
		k = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(v)
		k = &h

	case genres.Config:
		err = Config.Set(v)
		k = Config

	case genres.InventoryList:
		var h Tai
		err = h.Set(v)
		k = &h

	case genres.Blob:
		if err = objectId.left.Set(v); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err

	default:
		err = genres.MakeErrUnrecognizedGenre(objectId.genre.String())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", v)
		return err
	}

	if err = objectId.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) ResetWith(b *objectId2) {
	objectId.genre = b.genre
	b.left.CopyTo(&objectId.left)
	b.right.CopyTo(&objectId.right)
	objectId.middle = b.middle

	if objectId.middle == '%' {
		objectId.virtual = true
	}

	b.repoId.CopyTo(&objectId.repoId)
}

func (objectId *objectId2) SetObjectIdLike(b interfaces.ObjectId) (err error) {
	if b, ok := b.(*objectId2); ok {
		objectId.genre = b.genre
		b.left.CopyTo(&objectId.left)
		b.right.CopyTo(&objectId.right)
		objectId.middle = b.middle

		if objectId.middle == '%' {
			objectId.virtual = true
		}

		b.repoId.CopyTo(&objectId.repoId)

		return err
	}

	if objectId.SetWithGenre(b.String(), b.GetGenre()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) ResetWithIdLike(b IdWithParts) (err error) {
	return objectId.SetWithIdLike(b)
}

func (objectId *objectId2) MarshalText() (text []byte, err error) {
	text = []byte(FormattedString(objectId))
	return text, err
}

func (objectId *objectId2) UnmarshalText(text []byte) (err error) {
	if err = objectId.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId2) MarshalBinary() (text []byte, err error) {
	text = []byte(FormattedString(objectId))
	return text, err
}

func (objectId *objectId2) UnmarshalBinary(bs []byte) (err error) {
	text := string(bs)

	if err = objectId.Set(text); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

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
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpType),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Type
		objectId.middle = doddish.OpType
		objectId.right.Write(seq.At(1).Contents)
		return err

		// %tag
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpVirtual),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Tag
		objectId.middle = doddish.OpVirtual
		objectId.right.Write(seq.At(1).Contents)
		return err

		// /repo
	case seq.MatchAll(
		doddish.TokenMatcherOp(doddish.OpPathSeparator),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Repo
		objectId.middle = doddish.OpPathSeparator
		objectId.right.Write(seq.At(1).Contents)
		return err

		// @digest-or-sig
	case seq.MatchAll(
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Blob
		objectId.middle = '@'
		objectId.right.Write(seq.At(1).Contents)
		return err

		// purpose@digest-or-sig
	case seq.MatchAll(
		doddish.TokenTypeIdentifier,
		doddish.TokenMatcherOp('@'),
		doddish.TokenTypeIdentifier,
	):
		objectId.genre = genres.Blob
		objectId.left.Write(seq.At(0).Contents)
		objectId.middle = '@'
		objectId.right.Write(seq.At(2).Contents)
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

		// TODO don't use object for validation, validate directly
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

func (id objectId2) ToType() TypeStruct {
	errors.PanicIfError(genres.Type.AssertGenre(id.genre))

	return TypeStruct{
		Value: id.right.String(),
	}
}
