package ids

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/doddish"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
)

type Genre byte

func MakeGenreAll() Genre {
	return MakeGenre(genres.All()...)
}

func MakeGenre(vs ...genres.Genre) (s Genre) {
	s.Add(vs...)
	return
}

func (genre Genre) IsEmpty() bool {
	return genre == 0
}

func (genre Genre) EqualsAny(b any) bool {
	return values.Equals(genre, b)
}

func (genre Genre) Equals(b Genre) bool {
	return genre == b
}

func (genre *Genre) Reset() {
	*genre = 0
}

func (genre *Genre) ResetWith(b Genre) {
	*genre = b
}

func (genre *Genre) Add(bs ...genres.Genre) {
	for _, b := range bs {
		*genre |= Genre(b.GetGenre().GetGenreBitInt())
	}
}

func (genre *Genre) Del(b interfaces.GenreGetter) {
	*genre &= ^Genre(b.GetGenre().GetGenreBitInt())
}

func (genre Genre) Contains(b interfaces.GenreGetter) bool {
	bg := Genre(b.GetGenre().GetGenreBitInt())
	return byte(genre&bg) == byte(bg)
}

func (genre Genre) ContainsOneOf(b interfaces.GenreGetter) bool {
	bg := Genre(b.GetGenre().GetGenreBitInt())
	return genre&bg != 0
}

func (genre Genre) Slice() []genres.Genre {
	tg := genres.All()
	out := make([]genres.Genre, 0, len(tg))

	for _, g := range tg {
		if !genre.ContainsOneOf(g) {
			continue
		}

		out = append(out, g)
	}

	return out
}

func (genre Genre) String() string {
	sb := strings.Builder{}

	first := true

	for _, g := range genres.All() {
		if !genre.ContainsOneOf(g) {
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

func (genre *Genre) AddString(v string) (err error) {
	var g genres.Genre

	if err = g.Set(v); err != nil {
		err = errors.Wrap(err)
		return
	}

	genre.Add(g)

	return
}

func (genre *Genre) Set(v string) (err error) {
	v = strings.TrimSpace(v)
	v = strings.ToLower(v)

	for _, g := range strings.Split(v, ",") {
		if err = genre.AddString(g); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (genre *Genre) ReadFromBoxScanner(
	scanner *doddish.Scanner,
) (err error) {
	for scanner.Scan() {
		seq := scanner.GetSeq()

		switch {
		case seq.MatchAll(doddish.TokenTypeIdentifier):
			// etikett type zettel kasten konfig
			if err = genre.AddString(string(seq.At(0).Contents)); err != nil {
				err = errors.Wrap(err)
				return
			}

		case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpOr)):
			// ,
			continue

		case seq.MatchAll(doddish.TokenMatcherOp(doddish.OpAnd)):
			// " "
			scanner.Unscan()
			return

		default:
			err = errors.ErrorWithStackf(
				"unsupported sequence: %q:%#v",
				seq,
				seq,
			)
			return
		}
	}

	if err = scanner.Error(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (genre Genre) Byte() byte {
	return byte(genre)
}

func (genre Genre) ReadByte() (byte, error) {
	return byte(genre), nil
}

func (genre *Genre) ReadFrom(r io.Reader) (n int64, err error) {
	var b [1]byte

	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b[:])
	n = int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	*genre = Genre(b[0])

	return
}

func (genre *Genre) WriteTo(w io.Writer) (n int64, err error) {
	var b byte

	if b, err = genre.ReadByte(); err != nil {
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
