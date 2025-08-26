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
	"code.linenisgreat.com/dodder/go/src/delta/sha"
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

type objectId2 struct {
	virtual     bool
	g           genres.Genre
	middle      byte // remove and replace with virtual
	left, right catgut.String
	repoId      catgut.String
	sha         sha.Sha
	// Domain
}

func (a *objectId2) GetObjectId() *objectId2 {
	return a
}

func (a *objectId2) Clone() (b *objectId2) {
	b = getObjectIdPool2().Get()
	b.ResetWithIdLike(a)
	return
}

func (a *objectId2) IsVirtual() bool {
	return a.virtual
}

func (a *objectId2) Equals(b *objectId2) bool {
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

func (k2 *objectId2) WriteTo(w io.Writer) (n int64, err error) {
	if k2.Len() > math.MaxUint8 {
		err = errors.ErrorWithStackf(
			"%q is greater than max uint8 (%d)",
			k2.String(),
			math.MaxUint8,
		)

		return
	}

	var n1 int64
	n1, err = k2.g.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	b := [2]uint8{uint8(k2.Len()), uint8(k2.left.Len())}

	var n2 int
	n2, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = k2.left.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	bMid := [1]byte{k2.middle}

	n2, err = ohio.WriteAllOrDieTrying(w, bMid[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = k2.right.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (k2 *objectId2) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = k2.g.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	var b [2]uint8

	var n2 int
	n2, err = ohio.ReadAllOrDieTrying(r, b[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	contentLength := b[0]
	middlePos := b[1]

	if middlePos > contentLength-1 {
		err = errors.ErrorWithStackf(
			"middle position %d is greater than last index: %d",
			middlePos,
			contentLength,
		)
		return
	}

	if _, err = k2.left.ReadNFrom(r, int(middlePos)); err != nil {
		err = errors.Wrap(err)
		return
	}

	var bMiddle [1]uint8

	n2, err = ohio.ReadAllOrDieTrying(r, bMiddle[:])
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	k2.middle = bMiddle[0]

	if _, err = k2.right.ReadNFrom(r, int(contentLength-middlePos-1)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (k2 *objectId2) SetGenre(g interfaces.GenreGetter) {
	if g == nil {
		k2.g = genres.None
	} else {
		k2.g = genres.Must(g.GetGenre())
	}

	if k2.g == genres.Zettel {
		k2.middle = '/'
	}
}

func (oid *objectId2) IsEmpty() bool {
	switch oid.g {
	case genres.None:
		if oid.left.String() == "/" && oid.right.IsEmpty() {
			return true
		}

	case genres.Zettel, genres.Blob:
		if oid.left.IsEmpty() && oid.right.IsEmpty() {
			return true
		}
	}

	return oid.left.Len() == 0 && oid.middle == 0 && oid.right.Len() == 0
}

func (k2 *objectId2) Len() int {
	return k2.left.Len() + 1 + k2.right.Len()
}

func (k2 *objectId2) GetHeadAndTail() (head, tail string) {
	head = k2.left.String()
	tail = k2.right.String()

	return
}

func (k2 *objectId2) LenHeadAndTail() (int, int) {
	return k2.left.Len(), k2.right.Len()
}

func (k2 *objectId2) GetRepoId() string {
	return k2.repoId.String()
}

// TODO perform validation
func (k2 *objectId2) SetRepoId(v string) (err error) {
	if err = k2.repoId.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (oid *objectId2) StringSansRepo() string {
	var sb strings.Builder

	switch oid.g {
	case genres.Zettel:
		sb.Write(oid.left.Bytes())

		if oid.middle != '\x00' {
			sb.WriteByte(oid.middle)
		}

		sb.Write(oid.right.Bytes())

	case genres.Type:
		sb.WriteRune('!')
		sb.Write(oid.right.Bytes())

	default:
		if oid.left.Len() > 0 {
			sb.Write(oid.left.Bytes())
		}

		if oid.middle != '\x00' {
			sb.WriteByte(oid.middle)
		}

		if oid.right.Len() > 0 {
			sb.Write(oid.right.Bytes())
		}
	}

	return sb.String()
}

func (k2 *objectId2) GetObjectIdString() string {
	return k2.String()
}

func (k2 *objectId2) StringSansOp() string {
	var sb strings.Builder

	if k2.repoId.Len() > 0 {
		sb.WriteRune('/')
		k2.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch k2.g {
	case genres.Zettel:
		sb.Write(k2.left.Bytes())

		if k2.middle != '\x00' {
			sb.WriteByte(k2.middle)
		}

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

func (k2 *objectId2) String() string {
	var sb strings.Builder

	if k2.repoId.Len() > 0 {
		sb.WriteRune('/')
		k2.repoId.WriteTo(&sb)
		sb.WriteRune('/')
	}

	switch k2.g {
	case genres.Zettel:
		sb.Write(k2.left.Bytes())

		if k2.middle != '\x00' {
			sb.WriteByte(k2.middle)
		}

		sb.Write(k2.right.Bytes())

	case genres.Type:
		sb.WriteByte(doddish.OpType)
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

func (k2 *objectId2) Reset() {
	k2.g = genres.None
	k2.left.Reset()
	k2.middle = 0
	k2.right.Reset()
	k2.repoId.Reset()
}

func (k2 *objectId2) PartsStrings() IdParts {
	return IdParts{
		Left:   &k2.left,
		Middle: k2.middle,
		Right:  &k2.right,
	}
}

func (k2 *objectId2) Parts() [3]string {
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

func (k2 *objectId2) GetGenre() interfaces.Genre {
	return k2.g
}

func (k2 *objectId2) Expand(
	a Abbr,
) (err error) {
	ex := a.ExpanderFor(k2.g)

	if ex == nil {
		return
	}

	v := k2.String()

	if v, err = ex(v); err != nil {
		err = nil
		return
	}

	if err = k2.SetWithGenre(v, k2.g); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (k2 *objectId2) Abbreviate(
	a Abbr,
) (err error) {
	return
}

func (h *objectId2) SetWithIdLike(
	k interfaces.ObjectId,
) (err error) {
	h.Reset()

	switch kt := k.(type) {
	case *objectId2:
		h.ResetWith(kt)
		return

	default:
		p := k.Parts()

		if err = h.left.Set(p[0]); err != nil {
			err = errors.Wrap(err)
			return
		}

		mid := []byte(p[1])

		if len(mid) >= 1 {
			h.middle = mid[0]

			if h.middle == '%' {
				h.virtual = true
			}
		}

		if err = h.right.Set(p[2]); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	h.SetGenre(k)

	return
}

func (h *objectId2) SetWithGenre(
	v string,
	g interfaces.GenreGetter,
) (err error) {
	h.g = genres.Make(g.GetGenre())

	if err = h.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (h *objectId2) SetLeft(v string) (err error) {
	h.g = genres.Zettel

	if err = h.left.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (oid *objectId2) Set(v string) (err error) {
	if v == "/" {
		oid.g = genres.Zettel
		return
	}

	var els []string

	if len(v) > 0 {
		els = strings.SplitAfterN(v[1:], "/", 2)
	}

	if strings.HasPrefix(v, "/") && len(els) == 2 {
		if err = oid.SetRepoId(strings.TrimSuffix(els[0], "/")); err != nil {
			err = errors.Wrap(err)
			return
		}

		v = els[1]
	}

	var k interfaces.ObjectId

	switch oid.g {
	case genres.None:
		if k, err = MakeObjectId(v); err != nil {
			oid.g = genres.Blob

			if err = oid.left.Set(v); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
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
		if err = oid.left.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

		return

	default:
		err = genres.MakeErrUnrecognizedGenre(oid.g.GetGenreString())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", v)
		return
	}

	if err = oid.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (oid *objectId2) SetOnlyNotUnknownGenre(v string) (err error) {
	if v == "/" {
		oid.g = genres.Zettel
		return
	}

	if strings.HasPrefix(v, "/") {
		els := strings.SplitAfterN(v[1:], "/", 2)

		if len(els) != 2 {
			err = errors.ErrorWithStackf("invalid object id format: %q", v)
			return
		}

		v = els[1]

		if err = oid.SetRepoId(strings.TrimSuffix(els[0], "/")); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var k interfaces.ObjectId

	switch oid.g {
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
		if err = oid.left.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

		return

	default:
		err = genres.MakeErrUnrecognizedGenre(oid.g.GetGenreString())
	}

	if err != nil {
		err = errors.Wrapf(err, "String: %q", v)
		return
	}

	if err = oid.SetWithIdLike(k); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a *objectId2) ResetWith(b *objectId2) {
	a.g = b.g
	b.left.CopyTo(&a.left)
	b.right.CopyTo(&a.right)
	a.middle = b.middle

	if a.middle == '%' {
		a.virtual = true
	}

	b.repoId.CopyTo(&a.repoId)
}

func (a *objectId2) SetObjectIdLike(b ObjectIdLike) (err error) {
	if b, ok := b.(*objectId2); ok {
		a.g = b.g
		b.left.CopyTo(&a.left)
		b.right.CopyTo(&a.right)
		a.middle = b.middle

		if a.middle == '%' {
			a.virtual = true
		}

		b.repoId.CopyTo(&a.repoId)

		return
	}

	if a.SetWithGenre(b.String(), b.GetGenre()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (a *objectId2) ResetWithIdLike(b interfaces.ObjectId) (err error) {
	return a.SetWithIdLike(b)
}

func (t *objectId2) MarshalText() (text []byte, err error) {
	text = []byte(FormattedString(t))
	return
}

func (t *objectId2) UnmarshalText(text []byte) (err error) {
	if err = t.Set(string(text)); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (t *objectId2) MarshalBinary() (text []byte, err error) {
	text = []byte(FormattedString(t))
	return
}

func (t *objectId2) UnmarshalBinary(bs []byte) (err error) {
	text := string(bs)

	if err = t.Set(text); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
