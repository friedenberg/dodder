package ids

import (
	"io"
	"math"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
)

var poolObjectId interfaces.Pool[objectId, *objectId]

func init() {
	poolObjectId = pool.Make(
		nil,
		func(objectId *objectId) {
			objectId.Reset()
		},
	)
}

func getObjectIdPool() interfaces.Pool[objectId, *objectId] {
	return poolObjectId
}

type objectId struct {
	genre       genres.Genre
	middle      byte // remove and replace with virtual
	left, right catgut.String
}

func (objectId *objectId) Clone() (clone *objectId) {
	clone = getObjectIdPool().Get()
	clone.ResetWithIdLike(objectId)
	return clone
}

func (objectId *objectId) IsVirtual() bool {
	switch objectId.genre {
	case genres.Zettel:
		return slices.Equal(objectId.left.Bytes(), []byte{'%'})

	case genres.Tag:
		return objectId.middle == '%' || slices.Equal(objectId.left.Bytes(), []byte{'%'})

	default:
		return false
	}
}

func (objectId *objectId) Equals(b *objectId) bool {
	if objectId.genre != b.genre {
		return false
	}

	if objectId.middle != b.middle {
		return false
	}

	if !objectId.left.Equals(&b.left) {
		return false
	}

	if !objectId.right.Equals(&b.right) {
		return false
	}

	return true
}

func (objectId *objectId) WriteTo(writer io.Writer) (n int64, err error) {
	if objectId.Len() > math.MaxUint8 {
		err = errors.ErrorWithStackf(
			"%q is greater than max uint8 (%d)",
			objectId.String(),
			math.MaxUint8,
		)

		return n, err
	}

	var n1 int64
	n1, err = objectId.genre.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	b := [2]uint8{uint8(objectId.Len()), uint8(objectId.left.Len())}

	var n2 int
	n2, err = ohio.WriteAllOrDieTrying(writer, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = objectId.left.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	bMid := [1]byte{objectId.middle}

	n2, err = ohio.WriteAllOrDieTrying(writer, bMid[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = objectId.right.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (objectId *objectId) ReadFrom(reader io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = objectId.genre.ReadFrom(reader)
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

	if _, err = objectId.left.ReadNFrom(reader, int(middlePos)); err != nil {
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

	objectId.middle = bMiddle[0]

	if _, err = objectId.right.ReadNFrom(reader, int(contentLength-middlePos-1)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (objectId *objectId) SetGenre(g interfaces.GenreGetter) {
	if g == nil {
		objectId.genre = genres.None
	} else {
		objectId.genre = genres.Must(g.GetGenre())
	}

	if objectId.genre == genres.Zettel {
		objectId.middle = '/'
	}
}

func (objectId *objectId) StringFromPtr() string {
	var sb strings.Builder

	switch objectId.genre {
	case genres.Zettel:
		sb.Write(objectId.left.Bytes())
		sb.WriteByte(objectId.middle)
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

func (objectId *objectId) IsEmpty() bool {
	if objectId.genre == genres.Zettel {
		if objectId.left.IsEmpty() && objectId.right.IsEmpty() {
			return true
		}
	}

	return objectId.left.Len() == 0 && objectId.middle == 0 && objectId.right.Len() == 0
}

func (objectId *objectId) Len() int {
	return objectId.left.Len() + 1 + objectId.right.Len()
}

func (objectId *objectId) GetHeadAndTail() (head, tail string) {
	head = objectId.left.String()
	tail = objectId.right.String()

	return head, tail
}

func (objectId *objectId) LenHeadAndTail() (int, int) {
	return objectId.left.Len(), objectId.right.Len()
}

func (objectId *objectId) GetObjectIdString() string {
	return objectId.StringFromPtr()
}

func (objectId *objectId) String() string {
	return objectId.StringFromPtr()
}

func (objectId *objectId) Reset() {
	objectId.genre = genres.None
	objectId.left.Reset()
	objectId.middle = 0
	objectId.right.Reset()
}

func (objectId *objectId) PartsStrings() IdParts {
	return IdParts{
		Left:   &objectId.left,
		Middle: objectId.middle,
		Right:  &objectId.right,
	}
}

func (objectId *objectId) Parts() [3]string {
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

func (objectId *objectId) GetGenre() interfaces.Genre {
	return objectId.genre
}

func (objectId *objectId) Expand(
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

func (objectId *objectId) Abbreviate(
	a Abbr,
) (err error) {
	return err
}

type objectIdPtr = *objectId

func (objectId *objectId) SetWithIdLike(
	otherObjectId interfaces.ObjectId,
) (err error) {
	switch kt := otherObjectId.(type) {
	case objectIdPtr:
		if err = kt.left.CopyTo(&objectId.left); err != nil {
			err = errors.Wrap(err)
			return err
		}

		objectId.middle = kt.middle

		if err = kt.right.CopyTo(&objectId.right); err != nil {
			err = errors.Wrap(err)
			return err
		}

	default:
		p := otherObjectId.Parts()

		if err = objectId.left.Set(p[0]); err != nil {
			err = errors.Wrap(err)
			return err
		}

		mid := []byte(p[1])

		if len(mid) >= 1 {
			objectId.middle = mid[0]
		}

		if err = objectId.right.Set(p[2]); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	objectId.SetGenre(otherObjectId)

	return err
}

func (objectId *objectId) SetWithGenre(
	value string,
	genre interfaces.GenreGetter,
) (err error) {
	objectId.genre = genres.Make(genre.GetGenre())

	if err = objectId.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId) TodoSetBytes(value *catgut.String) (err error) {
	return objectId.Set(value.String())
}

func (objectId *objectId) SetRaw(value string) (err error) {
	objectId.genre = genres.None

	if err = objectId.left.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId) Set(value string) (err error) {
	var parsingObjectId interfaces.ObjectId

	switch objectId.genre {
	case genres.None:
		parsingObjectId, err = MakeObjectId(value)

	case genres.Zettel:
		var h ZettelId
		err = h.Set(value)
		parsingObjectId = &h

	case genres.Tag:
		var h Tag
		err = h.Set(value)
		parsingObjectId = &h

	case genres.Type:
		var h Type
		err = h.Set(value)
		parsingObjectId = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(value)
		parsingObjectId = &h

	case genres.Config:
		err = Config.Set(value)
		parsingObjectId = Config

	case genres.InventoryList:
		var h Tai
		err = h.Set(value)
		parsingObjectId = &h

	default:
		err = genres.MakeErrUnrecognizedGenre(objectId.genre.GetGenreString())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", value)
		return err
	}

	if err = objectId.SetWithIdLike(parsingObjectId); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId) ResetWith(otherObjectId *objectId) {
	objectId.genre = otherObjectId.genre
	otherObjectId.left.CopyTo(&objectId.left)
	otherObjectId.right.CopyTo(&objectId.right)
	objectId.middle = otherObjectId.middle
}

func (objectId *objectId) ResetWithIdLike(otherObjectId interfaces.ObjectId) (err error) {
	return objectId.SetWithIdLike(otherObjectId)
}

func (objectId *objectId) MarshalText() (text []byte, err error) {
	text = []byte(FormattedString(objectId))
	return text, err
}

func (objectId *objectId) UnmarshalText(text []byte) (err error) {
	if err = objectId.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (objectId *objectId) MarshalBinary() (text []byte, err error) {
	text = []byte(FormattedString(objectId))
	return text, err
}

func (objectId *objectId) UnmarshalBinary(text []byte) (err error) {
	if err = objectId.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
