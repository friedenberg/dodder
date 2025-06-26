package ids

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/box"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type Genre byte

func MakeGenreAll() Genre {
	return MakeGenre(genres.All()...)
}

func MakeGenre(vs ...genres.Genre) (s Genre) {
	s.Add(vs...)
	return
}

func (a Genre) IsEmpty() bool {
	return a == 0
}

func (a Genre) EqualsAny(b any) bool {
	return values.Equals(a, b)
}

func (a Genre) Equals(b Genre) bool {
	return a == b
}

func (a *Genre) Reset() {
	*a = 0
}

func (a *Genre) ResetWith(b Genre) {
	*a = b
}

func (a *Genre) Add(bs ...genres.Genre) {
	for _, b := range bs {
		*a |= Genre(b.GetGenre().GetGenreBitInt())
	}
}

func (a *Genre) Del(b interfaces.GenreGetter) {
	*a &= ^Genre(b.GetGenre().GetGenreBitInt())
}

func (a Genre) Contains(b interfaces.GenreGetter) bool {
	bg := Genre(b.GetGenre().GetGenreBitInt())
	return byte(a&bg) == byte(bg)
}

func (a Genre) ContainsOneOf(b interfaces.GenreGetter) bool {
	bg := Genre(b.GetGenre().GetGenreBitInt())
	return a&bg != 0
}

func (a Genre) Slice() []genres.Genre {
	tg := genres.All()
	out := make([]genres.Genre, 0, len(tg))

	for _, g := range tg {
		if !a.ContainsOneOf(g) {
			continue
		}

		out = append(out, g)
	}

	return out
}

func (a Genre) String() string {
	sb := strings.Builder{}

	first := true

	for _, g := range genres.All() {
		if !a.ContainsOneOf(g) {
			continue
		}

		if !first {
			sb.WriteRune(',')
		}

		sb.WriteString(g.GetGenreString())
		first = false
	}

	return sb.String()
}

func (i *Genre) AddString(v string) (err error) {
	var g genres.Genre

	if err = g.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.Add(g)

	return
}

func (gs *Genre) Set(v string) (err error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	for _, g := range strings.Split(v, ",") {
		if err = gs.AddString(g); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (g *Genre) ReadFromBoxScanner(
	scanner *box.Scanner,
) (err error) {
	for scanner.Scan() {
		seq := scanner.GetSeq()

		switch {
		case seq.MatchAll(box.TokenTypeIdentifier):
			// etikett type zettel kasten konfig
			if err = g.AddString(string(seq.At(0).Contents)); err != nil {
				err = errors.Wrap(err)
				return
			}

		case seq.MatchAll(box.TokenMatcherOp(box.OpOr)):
			// ,
			continue

		case seq.MatchAll(box.TokenMatcherOp(box.OpAnd)):
			// " "
			scanner.Unscan()
			return

		default:
			err = errors.ErrorWithStackf("unsupported sequence: %q:%#v", seq, seq)
			return
		}
	}

	if err = scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i Genre) GetSha() *sha.Sha {
	return sha.FromStringContent(i.String())
}

func (i Genre) Byte() byte {
	return byte(i)
}

func (i Genre) ReadByte() (byte, error) {
	return byte(i), nil
}

func (i *Genre) ReadFrom(r io.Reader) (n int64, err error) {
	var b [1]byte

	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b[:])
	n = int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	*i = Genre(b[0])

	return
}

func (i *Genre) WriteTo(w io.Writer) (n int64, err error) {
	var b byte

	if b, err = i.ReadByte(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, []byte{b})
	n = int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
