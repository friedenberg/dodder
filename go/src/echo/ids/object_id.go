package ids

import (
	"io"
	"math"
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

var poolObjectId interfaces.Pool[objectId, *objectId]

func init() {
	poolObjectId = pool.Make(
		nil,
		func(k *objectId) {
			k.Reset()
		},
	)
}

func getObjectIdPool() interfaces.Pool[objectId, *objectId] {
	return poolObjectId
}

type objectId struct {
	g           genres.Genre
	middle      byte // remove and replace with virtual
	left, right catgut.String
}

func (a *objectId) Clone() (b *objectId) {
	b = getObjectIdPool().Get()
	b.ResetWithIdLike(a)
	return b
}

func (a *objectId) IsVirtual() bool {
	switch a.g {
	case genres.Zettel:
		return slices.Equal(a.left.Bytes(), []byte{'%'})

	case genres.Tag:
		return a.middle == '%' || slices.Equal(a.left.Bytes(), []byte{'%'})

	default:
		return false
	}
}

func (a *objectId) Equals(b *objectId) bool {
	if a.g != b.g {
		return false
	}

	if a.middle != b.middle {
		return false
	}

	if !a.left.Equals(&b.left) {
		return false
	}

	if !a.right.Equals(&b.right) {
		return false
	}

	return true
}

func (k2 *objectId) WriteTo(w io.Writer) (n int64, err error) {
	if k2.Len() > math.MaxUint8 {
		err = errors.ErrorWithStackf(
			"%q is greater than max uint8 (%d)",
			k2.String(),
			math.MaxUint8,
		)

		return n, err
	}

	var n1 int64
	n1, err = k2.g.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	b := [2]uint8{uint8(k2.Len()), uint8(k2.left.Len())}

	var n2 int
	n2, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = k2.left.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	bMid := [1]byte{k2.middle}

	n2, err = ohio.WriteAllOrDieTrying(w, bMid[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = k2.right.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (k2 *objectId) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = k2.g.ReadFrom(r)
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

	if _, err = k2.left.ReadNFrom(r, int(middlePos)); err != nil {
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

	k2.middle = bMiddle[0]

	if _, err = k2.right.ReadNFrom(r, int(contentLength-middlePos-1)); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}

func (k2 *objectId) SetGenre(g interfaces.GenreGetter) {
	if g == nil {
		k2.g = genres.None
	} else {
		k2.g = genres.Must(g.GetGenre())
	}

	if k2.g == genres.Zettel {
		k2.middle = '/'
	}
}

func (k2 *objectId) StringFromPtr() string {
	var sb strings.Builder

	switch k2.g {
	case genres.Zettel:
		sb.Write(k2.left.Bytes())
		sb.WriteByte(k2.middle)
		sb.Write(k2.right.Bytes())

	case genres.Type:
		sb.Write(k2.right.Bytes())

	default:
		if k2.left.Len() > 0 {
			sb.Write(k2.left.Bytes())
		}

		if k2.middle != '\x00' {
			sb.WriteByte(k2.middle)
		}

		if k2.right.Len() > 0 {
			sb.Write(k2.right.Bytes())
		}
	}

	return sb.String()
}

func (k2 *objectId) IsEmpty() bool {
	if k2.g == genres.Zettel {
		if k2.left.IsEmpty() && k2.right.IsEmpty() {
			return true
		}
	}

	return k2.left.Len() == 0 && k2.middle == 0 && k2.right.Len() == 0
}

func (k2 *objectId) Len() int {
	return k2.left.Len() + 1 + k2.right.Len()
}

func (k2 *objectId) GetHeadAndTail() (head, tail string) {
	head = k2.left.String()
	tail = k2.right.String()

	return head, tail
}

func (k2 *objectId) LenHeadAndTail() (int, int) {
	return k2.left.Len(), k2.right.Len()
}

func (k2 *objectId) GetObjectIdString() string {
	return k2.StringFromPtr()
}

func (k2 *objectId) String() string {
	return k2.StringFromPtr()
}

func (k2 *objectId) Reset() {
	k2.g = genres.None
	k2.left.Reset()
	k2.middle = 0
	k2.right.Reset()
}

func (k2 *objectId) PartsStrings() IdParts {
	return IdParts{
		Left:   &k2.left,
		Middle: k2.middle,
		Right:  &k2.right,
	}
}

func (k2 *objectId) Parts() [3]string {
	var mid string

	if k2.middle != 0 {
		mid = string([]byte{k2.middle})
	}

	return [3]string{
		k2.left.String(),
		mid,
		k2.right.String(),
	}
}

func (k2 *objectId) GetGenre() interfaces.Genre {
	return k2.g
}

func (k2 *objectId) Expand(
	a Abbr,
) (err error) {
	ex := a.ExpanderFor(k2.g)

	if ex == nil {
		return err
	}

	v := k2.String()

	if v, err = ex(v); err != nil {
		err = nil
		return err
	}

	if err = k2.SetWithGenre(v, k2.g); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (k2 *objectId) Abbreviate(
	a Abbr,
) (err error) {
	return err
}

func (h *objectId) SetWithIdLike(
	k interfaces.ObjectId,
) (err error) {
	switch kt := k.(type) {
	case *objectId:
		if err = kt.left.CopyTo(&h.left); err != nil {
			err = errors.Wrap(err)
			return err
		}

		h.middle = kt.middle

		if err = kt.right.CopyTo(&h.right); err != nil {
			err = errors.Wrap(err)
			return err
		}

	default:
		p := k.Parts()

		if err = h.left.Set(p[0]); err != nil {
			err = errors.Wrap(err)
			return err
		}

		mid := []byte(p[1])

		if len(mid) >= 1 {
			h.middle = mid[0]
		}

		if err = h.right.Set(p[2]); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	h.SetGenre(k)

	return err
}

func (h *objectId) SetWithGenre(
	v string,
	g interfaces.GenreGetter,
) (err error) {
	h.g = genres.Make(g.GetGenre())

	if err = h.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (h *objectId) TodoSetBytes(v *catgut.String) (err error) {
	return h.Set(v.String())
}

func (h *objectId) SetRaw(v string) (err error) {
	h.g = genres.None

	if err = h.left.Set(v); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (h *objectId) Set(v string) (err error) {
	var k interfaces.ObjectId

	switch h.g {
	case genres.None:
		k, err = MakeObjectId(v)

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

	default:
		err = genres.MakeErrUnrecognizedGenre(h.g.GetGenreString())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", v)
		return err
	}

	if err = h.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (a *objectId) ResetWith(b *objectId) {
	a.g = b.g
	b.left.CopyTo(&a.left)
	b.right.CopyTo(&a.right)
	a.middle = b.middle
}

func (a *objectId) ResetWithIdLike(b interfaces.ObjectId) (err error) {
	return a.SetWithIdLike(b)
}

func (t *objectId) MarshalText() (text []byte, err error) {
	text = []byte(FormattedString(t))
	return text, err
}

func (t *objectId) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (t *objectId) MarshalBinary() (text []byte, err error) {
	text = []byte(FormattedString(t))
	return text, err
}

func (t *objectId) UnmarshalBinary(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
