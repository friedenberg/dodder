package ids

import (
	"io"
	"math"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

var poolObjectId2 interfaces.Pool[objectId2, *objectId2]

func init() {
	poolObjectId2 = pool.Make(
		nil,
		func(k *objectId2) {
			k.Reset()
		},
	)
}

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

func (objectId *objectId2) GetObjectId() *objectId2 {
	return objectId
}

func (objectId *objectId2) Clone() (clone *objectId2) {
	clone = getObjectIdPool2().Get()
	clone.ResetWithIdLike(objectId)
	return clone
}

func (objectId *objectId2) IsVirtual() bool {
	return objectId.virtual
}

func (objectId *objectId2) Equals(b *objectId2) bool {
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

func (objectId *objectId2) GetHeadAndTail() (head, tail string) {
	head = objectId.left.String()
	tail = objectId.right.String()

	return head, tail
}

func (objectId *objectId2) LenHeadAndTail() (int, int) {
	return objectId.left.Len(), objectId.right.Len()
}

func (objectId *objectId2) GetRepoId() string {
	return objectId.repoId.String()
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

func (objectId *objectId2) GetObjectIdString() string {
	return objectId.String()
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

func (objectId *objectId2) PartsStrings() IdParts {
	return IdParts{
		Left:   &objectId.left,
		Middle: objectId.middle,
		Right:  &objectId.right,
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
	k interfaces.ObjectId,
) (err error) {
	objectId.Reset()

	switch kt := k.(type) {
	case *objectId2:
		objectId.ResetWith(kt)
		return err

	default:
		p := k.Parts()

		if err = objectId.left.Set(p[0]); err != nil {
			err = errors.Wrap(err)
			return err
		}

		mid := []byte(p[1])

		if len(mid) >= 1 {
			objectId.middle = mid[0]

			if objectId.middle == '%' {
				objectId.virtual = true
			}
		}

		if err = objectId.right.Set(p[2]); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	objectId.SetGenre(k)

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

	var k interfaces.ObjectId

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
		var h Tag
		err = h.Set(v)
		k = &h

	case genres.Type:
		var h Type
		err = h.Set(v)
		k = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(v)
		k = &h

	case genres.Config:
		var h Config
		err = h.Set(v)
		k = &h

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
		err = genres.MakeErrUnrecognizedGenre(objectId.genre.GetGenreString())
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

	var k interfaces.ObjectId

	switch objectId.genre {
	case genres.Zettel:
		var h ZettelId
		err = h.Set(v)
		k = &h

	case genres.Tag:
		var h Tag
		err = h.Set(v)
		k = &h

	case genres.Type:
		var h Type
		err = h.Set(v)
		k = &h

	case genres.Repo:
		var h RepoId
		err = h.Set(v)
		k = &h

	case genres.Config:
		var h Config
		err = h.Set(v)
		k = &h

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
		err = genres.MakeErrUnrecognizedGenre(objectId.genre.GetGenreString())
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

func (objectId *objectId2) SetObjectIdLike(b ObjectIdLike) (err error) {
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

func (objectId *objectId2) ResetWithIdLike(b interfaces.ObjectId) (err error) {
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
